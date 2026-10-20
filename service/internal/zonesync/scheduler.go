package zonesync

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"
)

// Config is everything a Scheduler needs. Every field is required; there are no
// defaults here, because the values that would most invite one — the monthly
// instant, the all-zones limit and the endpoint — have their homes outside this
// module and reach it through `internal/config`.
type Config struct {
	Fetch  FetchFunc
	Store  Store
	Locker Locker
	Logger *slog.Logger

	// Schedule is when the month's refresh falls due (#233).
	Schedule Monthly

	// RetryInterval is how long must pass between one attempt and the next,
	// counted from the recorded start of the last. It paces the retry of a
	// month's refresh that did not complete, and it is the all-zones limit of
	// `NFR-001` as `config.ShortestInterval` derives it rather than a figure of
	// its own (`DECISIONS.md § RD-044`).
	RetryInterval time.Duration

	FetchTimeout    time.Duration
	DatabaseTimeout time.Duration
	MergeTimeout    time.Duration
	MinZoneRatio    float64
}

func (c Config) logger() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}

	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Scheduler is the background job of `FR-022`. Run is the only exported entry
// point in this package that can reach the endpoint, and it is a loop rather
// than a refresh: there is deliberately nothing here for a request handler to
// call.
type Scheduler struct {
	runner   *runner
	store    Store
	locker   Locker
	log      *slog.Logger
	schedule Monthly
	retry    time.Duration

	databaseTimeout time.Duration

	// now and sleep are fields so a test can drive the loop without waiting on
	// a clock. They are unexported and have no setters: substituting them is
	// available to this package's own tests and to nothing else.
	now   func() time.Time
	sleep func(ctx context.Context, d time.Duration) bool
}

// NewScheduler builds the job.
func NewScheduler(cfg Config) (*Scheduler, error) {
	run, err := newRunner(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Locker == nil {
		return nil, errors.New("no locker")
	}

	if err := cfg.Schedule.validate(); err != nil {
		return nil, err
	}

	if cfg.RetryInterval <= 0 {
		return nil, errors.New("the retry interval is not positive")
	}

	return &Scheduler{
		runner:          run,
		store:           cfg.Store,
		locker:          cfg.Locker,
		log:             cfg.logger(),
		schedule:        cfg.Schedule,
		retry:           cfg.RetryInterval,
		databaseTimeout: cfg.DatabaseTimeout,
		now:             time.Now,
		sleep:           sleep,
	}, nil
}

// Run refreshes the local copy on the schedule until ctx is cancelled, and then
// returns nil.
//
// IT NEVER RETURNS ON A FAILURE. A store that is down, an endpoint that refuses,
// a response that is refused by the assertions — each of those costs one retry
// interval of freshness and nothing else, and a loop that exited on any of them
// would turn a transient fault into a copy that is never refreshed again until
// somebody restarts the process. Nothing on the request path waits on this
// either way, so there is nothing a failure here can block.
func (s *Scheduler) Run(ctx context.Context) error {
	// lastLocal is this process's own record of when it last tried, and it
	// exists for the case where the recorded one cannot be written or read.
	// Without it, a store that refuses the opening `sync_run` insert leaves the
	// gate reading "no attempt ever" on every iteration, and the loop spins
	// against a broken store as fast as it can fail.
	var lastLocal time.Time

	for {
		if ctx.Err() != nil {
			return nil
		}

		wait := s.untilDue(ctx, lastLocal)
		if wait > 0 {
			if !s.sleep(ctx, wait) {
				return nil
			}

			continue
		}

		lastLocal = s.now()

		s.attempt(ctx)
	}
}

// untilDue is how long is left before the next attempt is permitted.
//
// AN ATTEMPT IS DUE WHEN THE MONTH'S REFRESH IS NOT DONE AND `NFR-001` PERMITS
// ONE (`DECISIONS.md § RD-044`). The month's refresh is done when a run recorded
// `ok` completed at or after the most recent monthly instant; until then the
// attempts are paced by the retry interval, counted from the last recorded
// attempt of any outcome, because the endpoint meters requests and a failed run
// spends the allowance exactly as a successful one does. Once it is done,
// nothing is due before the next instant.
//
// BOTH GATES ARE THE STORE'S RECORD, not the process's uptime, which is what
// makes a restart neither skip nor double the month's run: a process that starts
// after the month's refresh completed waits for the next instant, and one that
// starts after an instant it was down through finds no completion since it and
// runs.
//
// A store it cannot read yields a whole retry interval rather than an error.
// There is no answer that both refreshes promptly and cannot exceed the limit,
// and of the two the one that costs freshness is the one whose damage does not
// outlive it.
func (s *Scheduler) untilDue(ctx context.Context, lastLocal time.Time) time.Duration {
	readCtx, cancel := context.WithTimeout(ctx, s.databaseTimeout)
	defer cancel()

	wait, err := s.recordedWait(readCtx)
	if err != nil {
		s.log.Error("the zone sync could not read its record of past runs, so it is waiting out a whole retry interval", "error", err)

		return s.retry
	}

	// The local fallback is the other way round, and stays that way: both of its
	// instants are this process's own clock, which is consistent with itself
	// whatever it reads, and it is only ever asked about this process's own last
	// attempt.
	if !lastLocal.IsZero() {
		if local := s.retry - s.now().Sub(lastLocal); local > wait {
			wait = local
		}
	}

	return wait
}

// recordedWait is the wait the store's record decides, every instant in it read
// from the store.
func (s *Scheduler) recordedWait(ctx context.Context) (time.Duration, error) {
	now, completedAt, err := s.store.LastCompleted(ctx)
	if err != nil {
		return 0, err
	}

	if !completedAt.IsZero() && !completedAt.Before(s.schedule.latest(now)) {
		next := s.schedule.next(now)

		s.log.Info("the month's all-zones refresh has completed, so none is due before the next monthly instant",
			"completed_at", completedAt, "next", next)

		return next.Sub(now), nil
	}

	// The recorded gate is the server's elapsed time, already subtracted there,
	// so nothing in this process is on either side of it.
	since, ever, err := s.store.SinceLastAttempt(ctx)
	if err != nil || !ever {
		return 0, err
	}

	return s.retry - since, nil
}

// attempt takes the lock and, if it is still due under it, performs one run.
//
// TWO TICKS CANNOT PRODUCE TWO CONCURRENT REQUESTS, and it takes both halves
// here to say so. Within one process the loop is sequential, so an overrunning
// run simply delays the next iteration. Between processes — a restart overlapping
// a shutdown, an operator running a second copy — only the lock decides, and the
// due-gate is re-read under it because the holder that just released it may have
// refreshed the copy while this one was waiting to be let in.
func (s *Scheduler) attempt(ctx context.Context) {
	lockCtx, cancelLock := context.WithTimeout(ctx, s.databaseTimeout)
	release, acquired, err := s.locker.Acquire(lockCtx)

	cancelLock()

	if err != nil {
		s.log.Error("the zone sync could not take its lock", "error", err)

		return
	}

	if !acquired {
		s.log.Info("another holder has the zone sync lock, so this attempt does nothing")

		return
	}

	defer func() {
		// Detached from ctx: at shutdown ctx is already cancelled, and a
		// release that inherited it would leave the lock held until the
		// connection went away.
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.databaseTimeout)
		defer cancel()

		release(releaseCtx)
	}()

	if wait := s.untilDue(ctx, time.Time{}); wait > 0 {
		s.log.Info("the copy was refreshed while this attempt waited for the lock", "due_in", wait)

		return
	}

	outcome := s.runner.runOnce(ctx)

	s.log.Info("a zone sync run finished", "outcome", string(outcome))
}

// sleep waits d, or returns false as soon as ctx is done.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
