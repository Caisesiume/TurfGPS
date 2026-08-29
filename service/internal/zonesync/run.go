// Package zonesync refreshes the service's local synced copy of the zone set
// from the all-zones endpoint recorded under `Architecture.md § Retrieving
// zones`, on a schedule, from a background job.
//
// # It is off the request path, including when it fails
//
// `FR-022` is one obligation and not two: a job that runs on a schedule and can
// also be triggered by a request is not off the request path. This package is
// built so the second half survives the code that has not been written yet.
//
//   - It exports no way to run a single refresh. The only exported entry that
//     can reach the endpoint is Scheduler.Run, a loop the process owns for its
//     whole life, and everything below it is unexported. A handler cannot
//     "just call the refresh" because there is nothing to call.
//   - No package in this module may import it transitively except the
//     composition root and `internal/syncstore`, which implements the ports
//     declared here and so must name them; and nothing reachable from the
//     request surface imports either of those two. That is checked rather than
//     intended — `offrequestpath_test.go`, which fails closed, and which is
//     where the exemption is argued rather than here.
//   - Even a caller that defeated both cannot spend the endpoint's allowance:
//     the attempt is gated on the last attempt recorded in `sync_run` and taken
//     under an exclusive lock, so the gate holds across ticks, across processes
//     and across restarts rather than across one process's uptime.
//
// # What a request sees while a refresh is running, and after one failed
//
// A request never waits on this package: it runs in its own goroutine, on the
// process's context, and no request-path code path reaches it.
//
// MID-REFRESH, a reader sees either the state before the merge or the state
// after it, never a mixture. That is a property of the merge being one
// transaction and never batched, argued in `Architecture.md § The sync write
// path`; the merge takes ROW EXCLUSIVE, which does not block readers. Staleness
// of up to an interval remains and is a product fact forced by the rate limit,
// but partial state is not something a query has to defend against.
//
// AFTER A FAILED REFRESH, a reader sees one whole state and never a mixture. The
// merge is all or nothing TO EVERY READER, which is what MVCC guarantees over
// one unbatched transaction and is what `Architecture.md § The sync write path`
// argues; it is the property a query actually has to be able to rely on.
//
// IT IS NOT THE STRONGER CLAIM THAT A FAILED RUN WRITES NOTHING TO `zone`, which
// this comment made until it was narrowed to what MVCC will support. A commit
// can be IN DOUBT: the server makes the merge durable and the acknowledgement
// does not arrive, so this run sees an error, records a terminal failure, and
// the merged rows are there anyway — carrying a last_changed_at that no `ok` run
// accounts for. Nothing available to this package can tell that case from a
// merge that never committed, which is why the invariant is stated over what a
// reader sees rather than over what a failed run wrote.
//
// CURRENCY DOES NOT MOVE EITHER WAY, because it is the completion instant of the
// latest `ok` run and a failed run has none — so the in-doubt case is reported
// as a copy one interval staler than it is, never as one fresher. A store that
// has never completed a run reports that distinctly rather than reporting a zero
// instant, so "never synced" cannot be read as "synced long ago". What is done
// with that answer is `FR-024`'s and `FR-033`'s, not this package's.
package zonesync

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Outcome is what one run writes to `sync_run.outcome`.
//
// The vocabulary's home is `Architecture.md § The sync write path`, which gives
// each value its meaning and says which of them are terminal, and it is
// enforced by a CHECK constraint in `migrations/0001_zone_store.sql`. These
// constants are the binding of that vocabulary into Go and add nothing to it,
// which is a claim this file has to keep true rather than merely make.
//
// FOUR OF THE FIVE ARE REACHABLE FROM THIS FILE AND THE FIFTH IS NOT, and the
// distinction is worth the line because the file said "every one of them" and
// that was not true. The four terminal values are each reached from real code by
// a real failure, and run_test.go asserts each one. `OutcomeRunning` is written
// only by `internal/syncstore`'s BeginRun, in SQL that has never been sent to a
// server — see that package's NEVER EXECUTED marker — and it appears in no test
// in this package. It is in the vocabulary for the reason
// `Architecture.md § The sync write path` gives, which is an argument and not a
// demonstration, and nothing here should be read as having demonstrated it.
type Outcome string

// The whole vocabulary, and nothing about it: the meanings are the section
// cited on Outcome above. They were restated here once, near-verbatim, which
// gave them a second home — one that would have gone stale with nothing in this
// file moving to notice it.
const (
	OutcomeRunning         Outcome = "running"
	OutcomeOK              Outcome = "ok"
	OutcomeHTTPError       Outcome = "http_error"
	OutcomeAssertionFailed Outcome = "assertion_failed"
	OutcomeAborted         Outcome = "aborted"
)

// FetchFunc fetches the complete all-zones response.
//
// It returns the body, the HTTP status the response carried, and an error. A
// status of zero means no response was received at all, which is what separates
// a refused request from a body that would not parse once both are recorded.
//
// It is a func rather than an interface so the adapter that implements it needs
// no knowledge of this package: `Architecture.md § Ports and adapters` puts the
// zone sync behind TurfClient, and a port whose type the adapter must import is
// a port the adapter is coupled to.
type FetchFunc func(ctx context.Context) (body []byte, status int, err error)

// Merged is what one merge did.
type Merged struct {
	Inserted int
	Updated  int

	// CompletedAt is when the merge committed, read from the SERVER inside the
	// merge's own transaction. It is both the instant every row the merge
	// touched carries in last_changed_at and the instant the run records as
	// completed_at, so the two cannot disagree — and neither is a reading of the
	// clock of whichever host happened to be running the sync.
	CompletedAt time.Time

	// AbsentIDs are ids held in `zone` and missing from the response. They are
	// recorded and never acted on — `Architecture.md § Absence is recorded and
	// never acted on` is the argument, and nothing in this package deletes a
	// zone on the strength of one.
	AbsentIDs []int32
}

// Result is everything one run learned, and the second of its two `sync_run`
// writes.
//
// The optional fields are pointers because their columns are nullable and the
// distinction is load-bearing: `ZonesReceived` of zero is a response that
// carried no zones, and a nil `ZonesReceived` is a run that never got far enough
// to count any. Only started_at and outcome are NOT NULL, which is what makes a
// run that failed before it received anything recordable at all.
type Result struct {
	Outcome       Outcome
	CompletedAt   *time.Time
	HTTPStatus    *int
	ResponseBytes *int64
	ZonesReceived *int
	RowsInserted  *int
	RowsUpdated   *int
	RowsUnchanged *int
	AbsentCount   *int
	AbsentIDs     []int32
}

// Store is what a run needs of the zone store.
//
// The interface is declared here, by the consumer, so the PostGIS adapter that
// satisfies it is substitutable and so this package's own behaviour — which
// outcome each failure reaches, and that every path out writes the row — is
// testable without a database.
type Store interface {
	// SinceLastAttempt returns how long ago the most recent run of ANY outcome
	// started, and whether there has been one. It is the rate limit's gate, and
	// it is the last attempt rather than the last success because the endpoint
	// meters requests and not results.
	//
	// IT IS AN ELAPSED TIME AND NOT AN INSTANT, which is what makes the gate
	// hold across hosts. The question is whether a whole interval has passed,
	// and answering it takes two instants; put a process clock on either side of
	// that subtraction and a host running fast subtracts an honest instant from
	// its own early one, decides the interval is up before it is, and spends an
	// allowance the endpoint meters per interval and not per process. Both
	// instants are the server's and the server subtracts them, so what comes
	// back is a duration no clock in this process can skew.
	SinceLastAttempt(ctx context.Context) (time.Duration, bool, error)

	// BeginRun inserts the run's row carrying `running` and the server's current
	// instant, and returns its id and that instant.
	//
	// The instant comes from the store for the reason above — it is the value
	// the gate subtracts — and comes back because the run is dated by it
	// everywhere else too: Inspect bounds date_created by it.
	BeginRun(ctx context.Context) (id int64, startedAt time.Time, err error)

	// FinishRun updates that row once, with the terminal outcome and whatever
	// else the run learned.
	FinishRun(ctx context.Context, id int64, result Result) error

	// Stage truncates the staging table and loads the response into it.
	Stage(ctx context.Context, zones []Zone) error

	// Inspect counts the staged rows for the assertions, bounding date_created
	// by the run's own start instant.
	Inspect(ctx context.Context, startedAt time.Time) (Staged, error)

	// Merge merges the staged rows into `zone` in ONE transaction, never
	// batched, stamping the server's instant into last_changed_at and returning
	// it on Merged.
	Merge(ctx context.Context) (Merged, error)
}

// Locker is the sync's exclusive lock.
//
// `Architecture.md § Migrating against a running sync` requires the sync and any
// later migration to serialise on a well-known key rather than hope the table is
// idle. The same lock is what makes "two ticks never produce two concurrent
// all-zones requests" hold between processes and not merely between iterations
// of one loop.
type Locker interface {
	// Acquire takes the lock without waiting for it. It reports false when
	// another holder has it, which is a skip and not an error. release is
	// non-nil exactly when acquired is true, and takes its own context so a
	// cancelled run can still let the lock go.
	Acquire(ctx context.Context) (release func(context.Context), acquired bool, err error)
}

// runner performs one refresh. It is unexported, and so is every method on it,
// because an exported single-refresh entry point is the seam a convenience
// trigger is added through.
type runner struct {
	fetch FetchFunc
	store Store
	log   *slog.Logger

	fetchTimeout    time.Duration
	databaseTimeout time.Duration
	mergeTimeout    time.Duration
	minZoneRatio    float64
}

// runOnce performs one refresh and returns the outcome it recorded.
//
// THE ROW IS WRITTEN ON EVERY PATH OUT OF THIS FUNCTION, failure included, and
// that is what the deferred finalise below is for. A run that returns through an
// error, a cancellation or a panic still updates the row it inserted, and it
// updates it through a context detached from ctx: at shutdown ctx is already
// cancelled, and a terminal write derived from it would fail exactly when the
// thing worth recording happened, leaving the row at `running` for a run whose
// end is known.
func (r *runner) runOnce(ctx context.Context) Outcome {
	beginCtx, cancelBegin := context.WithTimeout(ctx, r.databaseTimeout)
	id, startedAt, err := r.store.BeginRun(beginCtx)

	cancelBegin()

	if err != nil {
		// There is no row to finish. A store that cannot insert cannot record,
		// and inventing a row later would date the run wrongly; the log is the
		// only place this can be said.
		r.log.Error("the zone sync could not record the start of a run, so it did not start one", "error", err)

		return OutcomeAborted
	}

	// EVERY LINE THIS RUN WRITES CARRIES THE ID OF THE ROW IT DESCRIBES, and
	// that is not decoration. `sync_run` records THAT the staging assertions
	// refused a response; WHICH assertion refused it is in the log and nowhere
	// else, as are the store error and the status behind every other outcome. An
	// operator holding a failed row and a log that named the run once had to
	// match the two on a timestamp — across however many lines the interval put
	// between them.
	log := r.log.With("run", id)

	result := Result{Outcome: OutcomeAborted}

	defer func() {
		recovered := recover()
		if recovered != nil {
			result.Outcome = OutcomeAborted
			log.Error("the zone sync run panicked", "panic", recovered)
		}

		finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.databaseTimeout)
		defer cancel()

		if err := r.store.FinishRun(finishCtx, id, result); err != nil {
			log.Error("the zone sync could not record the end of a run, which will be left at running", "outcome", result.Outcome, "error", err)
		}

		if recovered != nil {
			panic(recovered)
		}
	}()

	r.execute(ctx, log, startedAt, &result)

	return result.Outcome
}

// execute is the run itself. It reports by filling result rather than by
// returning, so the deferred finalise records what was learned before a failure
// as well as the failure.
func (r *runner) execute(ctx context.Context, log *slog.Logger, startedAt time.Time, result *Result) {
	zones, ok := r.receive(ctx, log, result)
	if !ok {
		return
	}

	received := len(zones)
	result.ZonesReceived = &received

	stageCtx, cancelStage := context.WithTimeout(ctx, r.databaseTimeout)
	err := r.store.Stage(stageCtx, zones)

	cancelStage()

	if err != nil {
		result.Outcome = OutcomeAborted

		log.Error("the response could not be staged", "error", err)

		return
	}

	inspectCtx, cancelInspect := context.WithTimeout(ctx, r.databaseTimeout)
	staged, err := r.store.Inspect(inspectCtx, startedAt)

	cancelInspect()

	if err != nil {
		result.Outcome = OutcomeAborted

		log.Error("the staged response could not be inspected", "error", err)

		return
	}

	if reasons := staged.Verify(r.minZoneRatio); len(reasons) > 0 {
		result.Outcome = OutcomeAssertionFailed

		log.Error("the staged response was refused and nothing was merged", "reasons", reasons)

		return
	}

	r.merge(ctx, log, received, result)
}

// receive performs the one external call, bounded by its own deadline, and maps
// what came back. It records what the response said about itself whether or not
// it turned out to be usable.
//
// The body is local to this function on purpose. It is the larger of the two
// representations of the same response, and letting it fall out of scope as the
// mapped rows are returned keeps the peak to one of them plus the staging load
// rather than to both across the whole run.
func (r *runner) receive(ctx context.Context, log *slog.Logger, result *Result) ([]Zone, bool) {
	fetchCtx, cancel := context.WithTimeout(ctx, r.fetchTimeout)
	body, status, err := r.fetch(fetchCtx)

	cancel()

	if status != 0 {
		result.HTTPStatus = &status
	}

	if body != nil {
		size := int64(len(body))
		result.ResponseBytes = &size
	}

	if err != nil {
		// A run the process cancelled is aborted rather than an HTTP failure:
		// the endpoint did nothing wrong, the service stopped. Testing the
		// PARENT context rather than the error distinguishes this from the
		// fetch's own deadline expiring, which is an unusable response like any
		// other.
		if ctx.Err() != nil {
			result.Outcome = OutcomeAborted

			log.Info("the zone sync run was cancelled during the fetch", "error", err)

			return nil, false
		}

		result.Outcome = OutcomeHTTPError

		log.Error("the all-zones fetch failed", "status", status, "error", err)

		return nil, false
	}

	zones, err := ParseAllZones(body)
	if err != nil {
		// A body that would not parse is an unusable response, which
		// `Architecture.md § The sync write path` puts under http_error
		// alongside a refused request. response_bytes is already recorded and
		// is what separates the two.
		result.Outcome = OutcomeHTTPError

		log.Error("the all-zones response could not be read", "error", err)

		return nil, false
	}

	return zones, true
}

// merge runs the one transaction and records what it did.
func (r *runner) merge(ctx context.Context, log *slog.Logger, received int, result *Result) {
	mergeCtx, cancel := context.WithTimeout(ctx, r.mergeTimeout)
	merged, err := r.store.Merge(mergeCtx)

	cancel()

	if err != nil {
		result.Outcome = OutcomeAborted

		log.Error("the merge did not commit and nothing was changed", "error", err)

		return
	}

	unchanged := received - merged.Inserted - merged.Updated
	absent := len(merged.AbsentIDs)
	completedAt := merged.CompletedAt

	result.Outcome = OutcomeOK
	result.CompletedAt = &completedAt
	result.RowsInserted = &merged.Inserted
	result.RowsUpdated = &merged.Updated
	result.RowsUnchanged = &unchanged
	result.AbsentCount = &absent
	result.AbsentIDs = merged.AbsentIDs
}

// newRunner validates what a run cannot proceed without.
func newRunner(cfg Config) (*runner, error) {
	if cfg.Fetch == nil {
		return nil, errors.New("no fetch")
	}

	if cfg.Store == nil {
		return nil, errors.New("no store")
	}

	// Ordered rather than ranged over a map, so the same misconfiguration
	// always names the same field.
	budgets := []struct {
		name string
		d    time.Duration
	}{
		{"the fetch timeout", cfg.FetchTimeout},
		{"the database timeout", cfg.DatabaseTimeout},
		{"the merge timeout", cfg.MergeTimeout},
	}

	for _, b := range budgets {
		if b.d <= 0 {
			return nil, fmt.Errorf("%s is not positive", b.name)
		}
	}

	if cfg.MinZoneRatio <= 0 || cfg.MinZoneRatio > 1 {
		return nil, fmt.Errorf("the minimum zone ratio %v is outside (0, 1]", cfg.MinZoneRatio)
	}

	return &runner{
		fetch:           cfg.Fetch,
		store:           cfg.Store,
		log:             cfg.logger(),
		fetchTimeout:    cfg.FetchTimeout,
		databaseTimeout: cfg.DatabaseTimeout,
		mergeTimeout:    cfg.MergeTimeout,
		minZoneRatio:    cfg.MinZoneRatio,
	}, nil
}
