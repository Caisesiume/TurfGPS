package zonesync

import (
	"context"
	"sync"
	"testing"
	"time"
)

// testInterval is the schedule these tests run on.
//
// IT IS NOT THE INTERVAL THE ENDPOINT PERMITS, and it deliberately looks nothing
// like it. That figure's one home is `Architecture.md § Retrieving zones`; a
// fixture carrying it would be a second home that goes wrong silently, which is
// what #20 forbids in code, comments, fixtures and defaults alike. What these
// tests measure is that the gate is applied to whatever interval it was given.
const testInterval = 100 * time.Second

// scheduleHarness drives Scheduler.Run without waiting on a clock.
//
// Time is a value the harness holds and sleeping advances it, so a test can
// assert exactly what a given elapsed interval permits. Real sleeps would
// measure the host rather than the schedule.
type scheduleHarness struct {
	mu sync.Mutex

	now    time.Time
	sleeps []time.Duration
	budget int

	cancel context.CancelFunc
}

func (h *scheduleHarness) clock() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.now
}

// sleep advances the clock and stops the loop once the test's budget of sleeps
// is spent, so Run always terminates.
func (h *scheduleHarness) sleep(ctx context.Context, d time.Duration) bool {
	h.mu.Lock()

	h.sleeps = append(h.sleeps, d)
	h.now = h.now.Add(d)
	h.budget--

	spent := h.budget <= 0

	h.mu.Unlock()

	if spent {
		h.cancel()

		return false
	}

	return ctx.Err() == nil
}

func (h *scheduleHarness) waited() []time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]time.Duration(nil), h.sleeps...)
}

// runSchedule builds a scheduler over store, locker and fetch, runs it until the
// sleep budget is spent, and returns the harness.
func runSchedule(t *testing.T, store Store, locker Locker, fetch FetchFunc, start time.Time, budget int) *scheduleHarness {
	t.Helper()

	s, err := NewScheduler(Config{
		Fetch:           fetch,
		Store:           store,
		Locker:          locker,
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

	harness := &scheduleHarness{now: start, budget: budget, cancel: cancel}

	s.now = harness.clock
	s.sleep = harness.sleep

	if err := s.Run(ctx); err != nil {
		t.Fatalf("the scheduler returned %v, want nil", err)
	}

	return harness
}

// TestTheRefreshRunsWhenTheScheduleFires is `FR-022` AC1: given a copy due for
// refresh and no request in flight, when the schedule fires, the refresh runs
// and updates the copy.
//
// Nothing in this test issues a request, and nothing in the loop under it can
// observe one — that is what "from its schedule" means here, and AC2 below is
// where the other half is bound.
func TestTheRefreshRunsWhenTheScheduleFires(t *testing.T) {
	t.Parallel()

	store := okStore()
	locker := &fakeLocker{}
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	runSchedule(t, store, locker, fetch.fetch, time.Now(), 1)

	if calls := fetch.count(); calls != 1 {
		t.Errorf("the schedule fired and the endpoint was called %d times, want 1", calls)
	}

	if merges := store.mergeCount(); merges != 1 {
		t.Errorf("the schedule fired and the copy was merged %d times, want 1 — a refresh that fetches and does not write has not updated the copy", merges)
	}

	if got := store.outcomes(); len(got) != 1 || got[0] != OutcomeOK {
		t.Errorf("the run recorded %v, want one %q", got, OutcomeOK)
	}
}

// TestTheNextFetchIsGatedOnTheRECORDEDLastAttempt is the rate limit surviving a
// restart.
//
// A process that has just started has no memory of the last run, so a gate on
// its own uptime would fetch immediately — spending an allowance the endpoint
// meters per interval, not per process. The store's record is what the gate
// reads, and the store outlives the process.
func TestTheNextFetchIsGatedOnTheRECORDEDLastAttempt(t *testing.T) {
	t.Parallel()

	start := time.Now()

	store := okStore()
	// A run recorded a moment ago by a process that is now gone.
	store.everAttempted = true
	store.sinceLastAttempt = testInterval / 4

	locker := &fakeLocker{}
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	harness := runSchedule(t, store, locker, fetch.fetch, start, 1)

	if calls := fetch.count(); calls != 0 {
		t.Errorf("a freshly started process called the endpoint %d times within the interval of a recorded attempt, want 0", calls)
	}

	waited := harness.waited()
	if len(waited) != 1 {
		t.Fatalf("the loop waited %d times, want 1", len(waited))
	}

	if want := testInterval - testInterval/4; waited[0] != want {
		t.Errorf("the loop waited %s, want %s — the remainder of the interval measured from the recorded attempt, not from this process's start", waited[0], want)
	}
}

// TestAnAttemptThatCannotTakeTheLockDoesNotFetch is the between-process half of
// "two ticks never produce two concurrent all-zones requests".
//
// Within one process the loop is sequential and cannot overlap itself. Between
// processes only the lock decides, and a tick that cannot take it must do
// nothing rather than fetch anyway.
func TestAnAttemptThatCannotTakeTheLockDoesNotFetch(t *testing.T) {
	t.Parallel()

	store := okStore()
	locker := &fakeLocker{refuse: true}
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	runSchedule(t, store, locker, fetch.fetch, time.Now(), 2)

	if calls := fetch.count(); calls != 0 {
		t.Errorf("the endpoint was called %d times while another holder had the lock, want 0", calls)
	}

	if got := store.outcomes(); len(got) != 0 {
		t.Errorf("a skipped attempt recorded %v, want nothing — a tick that did not run is not a run that failed, and recording one would move the gate the next attempt reads", got)
	}
}

// TestTheLockIsNeverHeldTwiceAtOnce watches the lock itself across a loop that
// runs repeatedly, which is the property the previous test's refusal stands in
// for.
func TestTheLockIsNeverHeldTwiceAtOnce(t *testing.T) {
	t.Parallel()

	store := okStore()
	locker := &fakeLocker{}
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	runSchedule(t, store, locker, fetch.fetch, time.Now(), 3)

	grants, overlaps := locker.stats()

	if grants == 0 {
		t.Fatal("the lock was never taken, so this test measured nothing")
	}

	if overlaps != 0 {
		t.Errorf("the lock was granted while already held %d times, want 0", overlaps)
	}
}

// TestAStoreThatCannotBeReadCostsFreshnessAndNotTheRateLimit binds the answer
// the loop gives when the gate itself is unreadable.
//
// There is no answer that both refreshes promptly and cannot exceed the limit.
// Waiting a whole interval costs one interval of freshness, and the cost of the
// other choice — fetching against an unknown last attempt — is spending the
// endpoint's allowance and having the damage outlive the attempt, which is the
// risk `FR-022` names.
func TestAStoreThatCannotBeReadCostsFreshnessAndNotTheRateLimit(t *testing.T) {
	t.Parallel()

	store := okStore()
	store.lastAttemptErr = errStore

	locker := &fakeLocker{}
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	harness := runSchedule(t, store, locker, fetch.fetch, time.Now(), 1)

	if calls := fetch.count(); calls != 0 {
		t.Errorf("the endpoint was called %d times with the gate unreadable, want 0", calls)
	}

	waited := harness.waited()
	if len(waited) != 1 || waited[0] != testInterval {
		t.Errorf("the loop waited %v, want one full interval of %s", waited, testInterval)
	}
}

// TestAStoreThatCannotRecordAStartDoesNotSpin is the loop's own protection.
//
// The gate is the recorded attempt, so a store that refuses the opening insert
// leaves it reading "no attempt ever" for ever. Without a second, local record
// the loop would attempt, fail, re-read the same empty gate, and attempt again
// as fast as the store can refuse — against the very endpoint whose allowance is
// the reason the gate exists.
func TestAStoreThatCannotRecordAStartDoesNotSpin(t *testing.T) {
	t.Parallel()

	store := okStore()
	store.beginErr = errStore

	locker := &fakeLocker{}
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	harness := runSchedule(t, store, locker, fetch.fetch, time.Now(), 1)

	waited := harness.waited()
	if len(waited) != 1 || waited[0] != testInterval {
		t.Errorf("after a start that could not be recorded the loop waited %v, want one full interval of %s rather than retrying immediately", waited, testInterval)
	}

	if calls := fetch.count(); calls != 0 {
		t.Errorf("the endpoint was called %d times by runs that could not record themselves, want 0", calls)
	}
}

// TestRunReturnsOnCancellation binds the shutdown path: the loop is owned by the
// process and lets go of it.
func TestRunReturnsOnCancellation(t *testing.T) {
	t.Parallel()

	s, err := NewScheduler(Config{
		Fetch:           (&fakeFetch{body: []byte(oneZoneResponse), status: 200}).fetch,
		Store:           okStore(),
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
	cancel()

	if err := s.Run(ctx); err != nil {
		t.Errorf("Run returned %v on a cancelled context, want nil", err)
	}
}
