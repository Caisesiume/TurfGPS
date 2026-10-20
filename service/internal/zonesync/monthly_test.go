package zonesync

import (
	"errors"
	"testing"
	"time"
)

// Tests for #233: the month's all-zones refresh is neither skipped (`FR-123`)
// nor doubled (`FR-124`), whatever the process does around the instant.
//
// Every case runs the real loop on a clock the harness holds, with the store's
// clock bound to it, and a restart is what it is in production: a new scheduler
// over the same store, remembering nothing but what the store recorded.

// marchInstant and aprilInstant are two consecutive instants of testSchedule,
// on either side of the change to summer time, so a wait from one to the other
// is a month less the hour the clocks went forward.
var (
	marchInstant = utc(2026, time.March, 1, 6, 45)
	aprilInstant = utc(2026, time.April, 5, 5, 45)
)

// completedLastMonth is a store whose record is one refresh, completed after
// February's instant by a process that is now gone.
func completedLastMonth() *fakeStore {
	store := okStore()
	store.lastBegun = utc(2026, time.February, 1, 7, 0)
	store.lastCompleted = store.lastBegun.Add(5 * time.Minute)

	return store
}

// failingFetch answers every request with a refusal, so no attempt completes.
func failingFetch() *fakeFetch {
	return &fakeFetch{status: 503, err: errors.New("the endpoint refused")}
}

// TestTheMonthsRefreshRunsAtTheInstantWhenTheProcessRunsThroughIt is AC1's
// first case: the process is running when the instant passes. It sleeps until
// the instant exactly, and then refreshes.
func TestTheMonthsRefreshRunsAtTheInstantWhenTheProcessRunsThroughIt(t *testing.T) {
	t.Parallel()

	store := completedLastMonth()
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	harness := runSchedule(t, store, &fakeLocker{}, fetch.fetch, marchInstant.Add(-2*time.Hour), 2)

	if calls := fetch.count(); calls != 1 {
		t.Fatalf("the process ran through the month's instant and called the endpoint %d times, want 1", calls)
	}

	if waited := harness.waited(); len(waited) == 0 || waited[0] != 2*time.Hour {
		t.Errorf("two hours before the instant the loop waited %v first, want 2h0m0s: until the instant, and no longer", waited)
	}

	if got := store.starts()[0]; !got.Equal(marchInstant) {
		t.Errorf("the month's refresh started at %s, want the instant %s", got.UTC().Format(time.RFC3339), marchInstant.Format(time.RFC3339))
	}
}

// TestAProcessStartingAfterTheInstantRefreshesAtOnce is AC1's second case, and
// a restart before the month's refresh ran: the process was down through the
// instant, and it refreshes when it starts rather than waiting a month for the
// next one. A second restart after that refresh then issues nothing (AC2).
func TestAProcessStartingAfterTheInstantRefreshesAtOnce(t *testing.T) {
	t.Parallel()

	store := completedLastMonth()
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	started := marchInstant.Add(3 * 24 * time.Hour)

	runSchedule(t, store, &fakeLocker{}, fetch.fetch, started, 1)

	if calls := fetch.count(); calls != 1 {
		t.Fatalf("a process started three days after the month's instant, with no refresh completed since it, called the endpoint %d times, want 1: the month's refresh was skipped", calls)
	}

	if got := store.starts()[0]; !got.Equal(started) {
		t.Errorf("the refresh started at %s, want at once, at %s", got.UTC().Format(time.RFC3339), started.Format(time.RFC3339))
	}

	runSchedule(t, store, &fakeLocker{}, fetch.fetch, started.Add(time.Hour), 1)

	if calls := fetch.count(); calls != 1 {
		t.Errorf("a process restarted an hour after the month's refresh completed called the endpoint again (%d calls in all), want 1: the month's refresh was doubled", calls)
	}
}

// TestAnEmptyCopyIsRefreshedAtOnceWhateverTheDate is AC1's third case: no run
// has ever completed, so the most recent instant has no completion after it.
func TestAnEmptyCopyIsRefreshedAtOnceWhateverTheDate(t *testing.T) {
	t.Parallel()

	for _, started := range []time.Time{marchInstant.Add(time.Minute), aprilInstant.Add(-time.Minute)} {
		store := okStore()
		fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

		runSchedule(t, store, &fakeLocker{}, fetch.fetch, started, 1)

		if calls := fetch.count(); calls != 1 {
			t.Errorf("against an empty copy at %s the endpoint was called %d times, want 1 at once", started.Format(time.RFC3339), calls)
		}
	}
}

// TestAnAttemptThatDidNotCompleteIsRetriedAsTheLimitPermits is AC1's fourth
// case. A failed attempt is not the month's refresh, so another follows — after
// the retry interval and not a month later — and so does one after an earlier
// process's attempt that never finished.
func TestAnAttemptThatDidNotCompleteIsRetriedAsTheLimitPermits(t *testing.T) {
	t.Parallel()

	t.Run("in this process", func(t *testing.T) {
		t.Parallel()

		store := completedLastMonth()
		fetch := failingFetch()

		harness := runSchedule(t, store, &fakeLocker{}, fetch.fetch, marchInstant, 2)

		if calls := fetch.count(); calls != 2 {
			t.Errorf("after a failed attempt at the instant the endpoint was called %d times within two retry intervals, want 2", calls)
		}

		if waited := harness.waited(); len(waited) < 2 || waited[0] != testRetryInterval || waited[1] != testRetryInterval {
			t.Errorf("after failed attempts the loop waited %v, want the retry interval %s each time", waited, testRetryInterval)
		}
	})

	t.Run("by an earlier process", func(t *testing.T) {
		t.Parallel()

		started := marchInstant.Add(time.Hour)

		store := completedLastMonth()
		store.lastBegun = started.Add(-testRetryInterval / 4)

		fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

		harness := runSchedule(t, store, &fakeLocker{}, fetch.fetch, started, 2)

		if calls := fetch.count(); calls != 1 {
			t.Errorf("after an earlier process's attempt that did not complete, the endpoint was called %d times, want 1", calls)
		}

		if waited := harness.waited(); len(waited) == 0 || waited[0] != testRetryInterval-testRetryInterval/4 {
			t.Errorf("the loop waited %v first, want the rest of the retry interval, %s", waited, testRetryInterval-testRetryInterval/4)
		}
	})
}

// TestARestartAfterTheMonthsRefreshIssuesNothingBeforeTheNextInstant is AC2's
// first case, at restarts spread across the month: each one waits exactly until
// the next instant and calls nothing.
func TestARestartAfterTheMonthsRefreshIssuesNothingBeforeTheNextInstant(t *testing.T) {
	t.Parallel()

	completed := marchInstant.Add(10 * time.Minute)

	for _, restart := range []time.Time{
		completed,
		completed.Add(time.Nanosecond),
		marchInstant.Add(14 * 24 * time.Hour),
		utc(2026, time.March, 29, 12, 0), // the day summer time began
		aprilInstant.Add(-time.Second),
	} {
		store := okStore()
		store.lastBegun = marchInstant
		store.lastCompleted = completed

		fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

		harness := runSchedule(t, store, &fakeLocker{}, fetch.fetch, restart, 1)

		if calls := fetch.count(); calls != 0 {
			t.Errorf("a process restarted at %s, after the month's refresh completed at %s, called the endpoint %d times, want 0",
				restart.Format(time.RFC3339Nano), completed.Format(time.RFC3339), calls)
		}

		if waited, want := harness.waited(), aprilInstant.Sub(restart); len(waited) != 1 || waited[0] != want {
			t.Errorf("a process restarted at %s waited %v, want %s: until the next instant %s",
				restart.Format(time.RFC3339Nano), waited, want, aprilInstant.Format(time.RFC3339))
		}
	}
}

// TestTheClockAdvancingIssuesNothingUntilTheNextInstant is AC2's second case.
// The process refreshes at the instant and runs on; the next request is at the
// next instant exactly, across the change to summer time, and none between.
func TestTheClockAdvancingIssuesNothingUntilTheNextInstant(t *testing.T) {
	t.Parallel()

	store := completedLastMonth()
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	harness := runSchedule(t, store, &fakeLocker{}, fetch.fetch, marchInstant, 2)

	if calls := fetch.count(); calls != 2 {
		t.Fatalf("over one month and one instant the endpoint was called %d times, want 2: one at each instant", calls)
	}

	if begun := store.starts(); !begun[0].Equal(marchInstant) || !begun[1].Equal(aprilInstant) {
		t.Errorf("the refreshes started at %s and %s, want %s and %s: nothing between the two instants, and the second at 07:45 local time after the clocks went forward",
			begun[0].UTC().Format(time.RFC3339), begun[1].UTC().Format(time.RFC3339),
			marchInstant.Format(time.RFC3339), aprilInstant.Format(time.RFC3339))
	}

	if waited := harness.waited(); len(waited) == 0 || waited[0] != aprilInstant.Sub(marchInstant) {
		t.Errorf("after the month's refresh the loop waited %v first, want %s, from one instant to the next", waited, aprilInstant.Sub(marchInstant))
	}
}

// TestAStoreThatCannotSayWhetherTheMonthIsDoneDoesNotFetch is AC2 against a
// store that cannot be read: not knowing whether the month's refresh completed
// is not grounds to run it again.
func TestAStoreThatCannotSayWhetherTheMonthIsDoneDoesNotFetch(t *testing.T) {
	t.Parallel()

	store := completedLastMonth()
	store.lastCompletedErr = errStore

	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	harness := runSchedule(t, store, &fakeLocker{}, fetch.fetch, marchInstant.Add(time.Hour), 1)

	if calls := fetch.count(); calls != 0 {
		t.Errorf("the endpoint was called %d times while the store could not say whether the month's refresh completed, want 0", calls)
	}

	if waited := harness.waited(); len(waited) != 1 || waited[0] != testRetryInterval {
		t.Errorf("the loop waited %v, want one retry interval of %s", waited, testRetryInterval)
	}
}
