package zonesync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
)

// This file is the SECOND HALF of `FR-022` AC2: a request served while the copy
// is due for refresh "does not wait on the refresh job".
//
// ---------------------------------------------------------------------------
// WHY THE IMPORT INVARIANT DOES NOT ALREADY COVER THIS. Read before editing.
//
// AC2 is two obligations in one sentence — a request ISSUES NO all-zones
// request, and it DOES NOT WAIT ON the refresh job. `offrequestpath_test.go`
// closes the first, structurally and for every package that will ever exist:
// nothing but the composition root may reach `internal/zonesync`, so nothing on
// the request path can call the thing that fetches.
//
// The second does not follow from it. Waiting does not require an import.
// A request that never names this package can still be made to wait on a
// refresh by anything the two share below the import graph — a mutex over an
// in-memory zone cache that the merge write-locks, a connection pool the merge
// has exhausted, a channel some future readiness check blocks on, a lock in the
// store the merge holds for the length of its transaction. An import-graph
// check cannot see any of those, because there is nothing in the graph to see.
//
// So this test measures the runtime property directly, and it measures it
// against a refresh that is GENUINELY IN FLIGHT rather than one that has been
// arranged to look busy. The store's merge — the longest and most exclusive
// step of a run, and the one holding a transaction open — is held open by the
// test, and the request is served while it is held. If serving the request
// touched anything the merge is holding, the request could not be answered
// until this test let the merge go, and this test does not let it go until the
// request has been answered.
//
// WHAT IT MEASURES IS NARROWER THAN THAT LIST, AND THE LIST IS NOT A CLAIM OF
// COVERAGE. The four mechanisms above are named to show what an import-graph
// check is blind to. They are not four things this file guards, and reading
// them as such is the misreading this paragraph exists to close. The
// arrangement below is entirely in-process: fakes from `fakes_test.go` and a
// mux built inside the timed section. It holds no zone cache, no readiness
// channel and NO CONNECTION POOL — so the pool case is not merely unexercised
// here, it is unreachable, because exhausting a pool requires a pool and these
// fakes own none.
//
// What it does measure is the whole of what the arrangement actually shares: a
// request answered while a refresh sits inside its merge, over every coupling
// that exists in this process. Today that set is empty, which is why the test
// passes; it stops passing on the day a handler takes a lock or a channel that
// the merge is also holding. NOT A POOLED CONNECTION, which the paragraph above
// has just finished retracting: these fakes own no pool, so that day cannot
// arrive here however this test grows, and naming it in the future tense eight
// lines below the retraction put it straight back. That is a real guard over a
// growing surface, and it is a different and smaller claim than guarding the
// four mechanisms named above.
//
// THE POOL VERSION IS A SEPARATE TEST AND IS DELIBERATELY NOT WRITTEN HERE. It
// needs one `pgxpool` genuinely shared between the write side and the read
// side, with the merge holding connections while a read is served; faked, it
// would measure the fake and report green over the arrangement it exists to
// refuse. It is recorded as FW-06 on #20's review record, and it is gated on
// there being a database to run it against — which
// `migrations/0001_zone_store.sql` records this repository as not having.
//
// It is not `TestTheLockIsNeverHeldTwiceAtOnce`, which watches sync against
// sync. This is request against sync, and the two questions have no answer in
// common.
//
// WHAT IT ASSERTS AND WHY EACH ONE FAILS CLOSED. The failure mode this file is
// most exposed to is its own: an arrangement in which nothing is in flight, or
// nothing on the request surface is exercised, reports green while measuring
// nothing — the defect `FR-019`'s Rationale names, a criterion satisfied
// vacuously wherever its artefact is absent. Every assertion below is therefore
// a claim that an empty world falsifies rather than satisfies. The refresh must
// have reached its merge; it must still be there, unfinished, when the request
// is answered; the request must reach a handler rather than a 404; and exactly
// one all-zones request must exist across the whole episode — zero meaning the
// refresh never ran, more than one meaning something else spent the endpoint's
// allowance while a refresh was already in flight.
//
// WHY IT LIVES IN THIS PACKAGE AND IMPORTS THE REQUEST SURFACE, WHICH IS THE
// DIRECTION THAT IS ALLOWED. The harness it needs — a store whose merge can be
// held, a locker, a fetch that counts — is this package's, and unexported. The
// import runs from a test of the worker to the request surface, which is the
// safe direction: `go list` reports a package's test imports separately from
// its imports, so nothing here enters `internal/zonesync`'s Deps and the
// invariant next door is untouched by it. The reverse — the request surface's
// tests importing this package — would be the shape AC2 forbids and must not be
// written.
// ---------------------------------------------------------------------------

// harnessBudget bounds the harness, not the service.
//
// It is generous by orders of magnitude rather than tuned. Every step it bounds
// completes in microseconds when it completes at all, so reaching it means the
// step did not happen rather than that the host was slow.
//
// IT IS NOT A SCHEDULE AND IT IS NOT THE REFRESH INTERVAL, whose one home is
// `Architecture.md § Retrieving zones`, for the reason `scheduler_test.go`
// gives at testInterval.
const harnessBudget = 10 * time.Second

// mergeHeldStore is a store whose merge stops and waits until the test releases
// it, so a refresh can be held at its longest, most exclusive step for as long
// as a test needs to do something else alongside it.
//
// Everything but Merge is the ordinary fake, so the run reaches the merge the
// way a real one does rather than being parked somewhere convenient.
type mergeHeldStore struct {
	*fakeStore

	// entered receives once, when a run reaches the merge.
	entered chan struct{}

	release chan struct{}

	releaseOnce func()
}

func newMergeHeldStore() *mergeHeldStore {
	s := &mergeHeldStore{
		fakeStore: okStore(),
		entered:   make(chan struct{}, 1),
		release:   make(chan struct{}),
	}

	// One-shot, so a test can release it defensively against its own early exit
	// and still release it at the point it means to. An unreleased merge would
	// strand the scheduler's goroutine on a test that has already failed.
	s.releaseOnce = sync.OnceFunc(func() { close(s.release) })

	return s
}

// Merge signals that a run has reached the merge and then holds it there.
//
// The signal is a non-blocking send for the reason the one-shot release exists:
// the number of runs is not the test's to control, and a second run must not
// wedge on a signal nobody is listening for.
func (s *mergeHeldStore) Merge(ctx context.Context) (Merged, error) {
	select {
	case s.entered <- struct{}{}:
	default:
	}

	<-s.release

	return s.fakeStore.Merge(ctx)
}

func (s *mergeHeldStore) letTheMergeFinish() { s.releaseOnce() }

// TestARequestIsAnsweredWhileARefreshIsHeldInItsMerge is `FR-022` AC2's second
// half: a request handled while the copy is being refreshed does not wait on
// the refresh job, and issues no all-zones request of its own.
func TestARequestIsAnsweredWhileARefreshIsHeldInItsMerge(t *testing.T) {
	t.Parallel()

	store := newMergeHeldStore()

	// Released here as well as below, so a failure anywhere in this test still
	// lets the scheduler's goroutine finish.
	t.Cleanup(store.letTheMergeFinish)

	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	scheduler, err := NewScheduler(Config{
		Fetch:           fetch.fetch,
		Store:           store,
		Locker:          &fakeLocker{},
		Interval:        testInterval,
		FetchTimeout:    time.Second,
		DatabaseTimeout: time.Second,
		MergeTimeout:    time.Second,
		MinZoneRatio:    0.9,
	})
	if err != nil {
		t.Fatalf("building the scheduler: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stopped := make(chan error, 1)

	go func() { stopped <- scheduler.Run(ctx) }()

	select {
	case <-store.entered:
	case <-time.After(harnessBudget):
		t.Fatalf("no refresh reached its merge within %s, so nothing was in flight and everything below would have measured an idle system rather than a request racing a refresh",
			harnessBudget)
	}

	// The refresh is inside the merge and cannot leave it until this test says
	// so. A terminal outcome recorded here would mean it had already left.
	if finished := store.results(); len(finished) != 0 {
		t.Fatalf("the refresh recorded %d terminal outcome(s) before the request was served, so it was not in flight when the request was served and this test measured nothing",
			len(finished))
	}

	// The request carries the test's own context and not the scheduler's. The
	// two must be independent, and one derived from the scheduler's would be
	// answering a different question.
	requestCtx := t.Context()

	served := make(chan *httptest.ResponseRecorder, 1)

	go func() {
		rec := httptest.NewRecorder()

		// The mux is built inside the timed section on purpose: wiring that
		// blocks at construction is a way of waiting on the refresh too, and it
		// would be invisible to a test that built the surface in advance.
		httpapi.NewMux().ServeHTTP(rec, httptest.NewRequestWithContext(requestCtx, http.MethodGet, "/", nil))

		served <- rec
	}()

	var rec *httptest.ResponseRecorder

	select {
	case rec = <-served:
	case <-time.After(harnessBudget):
		t.Fatalf("the request was still unanswered %s after a refresh entered its merge, and the merge is still being held, so the request is waiting on the refresh job: FR-022 AC2 requires a request served while the local copy is due for refresh not to block on it",
			harnessBudget)
	}

	// A 404 is answered by the mux before any handler runs, so a route that has
	// moved would leave this test asserting that a miss does not block. It
	// refuses rather than passing, for the reason the invariant next door
	// refuses a missing package.
	if rec.Code == http.StatusNotFound {
		t.Fatalf("the request reached no handler (status %d), so nothing on the request surface was exercised: the route this test serves has moved and this test has to move with it rather than report green",
			rec.Code)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("the request answered %d while a refresh was in flight, want %d", rec.Code, http.StatusOK)
	}

	if calls := fetch.count(); calls != 1 {
		t.Errorf("%d all-zones requests had been issued by the time the request was answered, want exactly 1 — the refresh's own. Zero means the refresh never reached the endpoint and this test measured nothing; more than one means a second was issued while a refresh was already in flight, which is the allowance FR-022's Risk says a request path can spend on a single page load",
			calls)
	}

	// Re-read after the request rather than only before it. The merge is still
	// held, so a run that has recorded an outcome by now finished some other
	// way, and the request was not raced against anything.
	if finished := store.results(); len(finished) != 0 {
		t.Errorf("the refresh recorded %d terminal outcome(s) while it was still held inside its merge, so it was not in flight for the whole of the request and this test measured nothing",
			len(finished))
	}

	store.letTheMergeFinish()
	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("the scheduler returned %v, want nil", err)
		}

		// The other direction of the same property: the request must neither
		// wait on the refresh nor disturb it.
		if got := store.outcomes(); len(got) != 1 || got[0] != OutcomeOK {
			t.Errorf("the refresh that ran alongside the request recorded %v, want one %q", got, OutcomeOK)
		}
	case <-time.After(harnessBudget):
		t.Errorf("the scheduler did not return within %s of cancellation", harnessBudget)
	}
}
