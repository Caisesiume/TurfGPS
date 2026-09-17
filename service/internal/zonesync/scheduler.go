package zonesync

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"
)

// Config is everything a Scheduler needs. Every field is required; there are no
// defaults here, because the two values that would most invite one — the
// interval and the endpoint — have their homes outside this module and reach it
// through `internal/config`.
type Config struct {
	Fetch  FetchFunc
	Store  Store
	Locker Locker
	Logger *slog.Logger

	// Interval is how long must pass between one attempt and the next. Its home
	// is `Architecture.md § Retrieving zones`.
	Interval time.Duration

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
	interval time.Duration

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

	if cfg.Interval <= 0 {
		return nil, errors.New("the refresh interval is not positive")
	}

	return &Scheduler{
		runner:          run,
		store:           cfg.Store,
		locker:          cfg.Locker,
		log:             cfg.logger(),
		interval:        cfg.Interval,
		databaseTimeout: cfg.DatabaseTimeout,
		now:             time.Now,
		sleep:           sleep,
	}, nil
}

// Run refreshes the local copy on the schedule until ctx is cancelled, and then
// returns nil.
//
// IT NEVER RETURNS ON A FAILURE. A store that is down, an endpoint that refuses,
// a response that is refused by the assertions — each of those loses one
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
// THE GATE IS THE RECORDED LAST ATTEMPT, not the process's uptime, which is what
// makes the endpoint's allowance survive a restart: a service that starts a
// second after a run finished waits out the rest of the interval rather than
// fetching because its own timer is new. It is the last ATTEMPT rather than the
// last success because the endpoint meters requests, so a failed run spends the
// allowance exactly as a successful one does.
//
// A store it cannot read yields a whole interval rather than an error. There is
// no answer that both refreshes promptly and cannot exceed the limit, and of the
// two the one that costs freshness is the one whose damage does not outlive it.
func (s *Scheduler) untilDue(ctx context.Context, lastLocal time.Time) time.Duration {
	readCtx, cancel := context.WithTimeout(ctx, s.databaseTimeout)
	defer cancel()

	since, ever, err := s.store.SinceLastAttempt(readCtx)
	if err != nil {
		s.log.Error("the zone sync could not read when it last ran, so it is waiting out a whole interval", "error", err)

		return s.interval
	}

	var wait time.Duration

	// The recorded gate is the server's elapsed time, already subtracted there,
	// so nothing in this process is on either side of it.
	if ever {
		wait = s.interval - since
	}

	// The local fallback is the other way round, and stays that way: both of its
	// instants are this process's own clock, which is consistent with itself
	// whatever it reads, and it is only ever asked about this process's own last
	// attempt.
	if !lastLocal.IsZero() {
		if local := s.interval - s.now().Sub(lastLocal); local > wait {
			wait = local
		}
	}

	return wait
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
