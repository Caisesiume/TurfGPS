package zonesync

import (
	"context"
	"errors"
	"sync"
	"time"
)

// errStore is the failure a fake store returns when a test asks it to fail.
var errStore = errors.New("the store said no")

// fakeStore is an in-memory Store. Every field is guarded because the scheduler
// runs the store from its own goroutine and these tests run under -race
// wherever a C compiler is available.
type fakeStore struct {
	mu sync.Mutex

	// clock is the store's own clock, the one `now()` reads in SQL; nil is the
	// process's. A test that drives the scheduler's clock hands it here too,
	// because the real store's instants are the server's and move with time.
	clock func() time.Time

	// lastBegun and lastCompleted are the two instants `sync_run` answers the
	// scheduler from: the start of the latest run of any outcome, and the
	// completion of the latest one recorded `ok`. A test seeds them to stand
	// for runs a process that is now gone recorded; the fake moves them the way
	// the real store's writes do.
	lastBegun        time.Time
	lastCompleted    time.Time
	lastAttemptErr   error
	lastCompletedErr error

	beginErr error
	nextID   int64
	begun    []time.Time

	finishErr error
	finished  []Result

	stageErr error
	staged   [][]Zone

	inspect    Staged
	inspectErr error

	merged   Merged
	mergeErr error
	merges   int
}

// now reads the store's clock. The caller holds mu.
func (f *fakeStore) now() time.Time {
	if f.clock != nil {
		return f.clock()
	}

	return time.Now()
}

func (f *fakeStore) setClock(clock func() time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.clock = clock
}

func (f *fakeStore) SinceLastAttempt(context.Context) (time.Duration, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.lastAttemptErr != nil {
		return 0, false, f.lastAttemptErr
	}

	if f.lastBegun.IsZero() {
		return 0, false, nil
	}

	return f.now().Sub(f.lastBegun), true, nil
}

func (f *fakeStore) LastCompleted(context.Context) (time.Time, time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.lastCompletedErr != nil {
		return time.Time{}, time.Time{}, f.lastCompletedErr
	}

	return f.now(), f.lastCompleted, nil
}

func (f *fakeStore) BeginRun(context.Context) (int64, time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.beginErr != nil {
		return 0, time.Time{}, f.beginErr
	}

	// The store dates the run, so the fake does too rather than being handed an
	// instant it would only echo.
	startedAt := f.now()

	f.begun = append(f.begun, startedAt)
	f.nextID++

	// A started run is the recorded attempt the rate limit is gated on, so the
	// fake records it the way the real store's INSERT does.
	f.lastBegun = startedAt

	return f.nextID, startedAt, nil
}

func (f *fakeStore) FinishRun(_ context.Context, _ int64, result Result) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.finishErr != nil {
		return f.finishErr
	}

	f.finished = append(f.finished, result)

	// What `ok` writes is what the monthly gate reads back, as the real
	// store's max(completed_at) over `ok` rows does.
	if result.Outcome == OutcomeOK && result.CompletedAt != nil && result.CompletedAt.After(f.lastCompleted) {
		f.lastCompleted = *result.CompletedAt
	}

	return nil
}

func (f *fakeStore) Stage(_ context.Context, zones []Zone) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.stageErr != nil {
		return f.stageErr
	}

	f.staged = append(f.staged, zones)

	return nil
}

func (f *fakeStore) Inspect(context.Context, time.Time) (Staged, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.inspectErr != nil {
		return Staged{}, f.inspectErr
	}

	return f.inspect, nil
}

func (f *fakeStore) Merge(context.Context) (Merged, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.merges++

	if f.mergeErr != nil {
		return Merged{}, f.mergeErr
	}

	merged := f.merged

	// The real store always answers with the instant it stamped, so a fixture
	// that did not set one is completed by the fake rather than modelling a
	// merge that stamped the zero instant.
	if merged.CompletedAt.IsZero() {
		merged.CompletedAt = f.now()
	}

	return merged, nil
}

func (f *fakeStore) outcomes() []Outcome {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]Outcome, 0, len(f.finished))
	for _, r := range f.finished {
		out = append(out, r.Outcome)
	}

	return out
}

func (f *fakeStore) results() []Result {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]Result(nil), f.finished...)
}

// starts is when each run the store recorded began, by the store's clock.
func (f *fakeStore) starts() []time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]time.Time(nil), f.begun...)
}

func (f *fakeStore) mergeCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.merges
}

// fakeLocker is a Locker that also reports whether it was ever held twice at
// once, which is the property the sync's own tests care about.
type fakeLocker struct {
	mu sync.Mutex

	refuse bool
	err    error

	held     bool
	overlaps int
	grants   int
}

func (l *fakeLocker) Acquire(context.Context) (func(context.Context), bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.err != nil {
		return nil, false, l.err
	}

	if l.refuse {
		return nil, false, nil
	}

	if l.held {
		l.overlaps++
	}

	l.held = true
	l.grants++

	return l.release, true, nil
}

func (l *fakeLocker) release(context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.held = false
}

func (l *fakeLocker) stats() (grants, overlaps int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.grants, l.overlaps
}

// fetching returns a FetchFunc answering with body and status, counting calls.
type fakeFetch struct {
	mu sync.Mutex

	body   []byte
	status int
	err    error

	calls int
}

func (f *fakeFetch) fetch(ctx context.Context) ([]byte, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++

	if ctx.Err() != nil {
		return nil, 0, ctx.Err()
	}

	return f.body, f.status, f.err
}

func (f *fakeFetch) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}
