package zonesync

import (
	"context"
	"errors"
	"testing"
	"time"
)

// oneZoneResponse is the smallest well-formed all-zones response. It carries the
// nine fields measured on every record and omits the optional ones, which is the
// ordinary shape rather than a degenerate one.
const oneZoneResponse = `[{"id":1,"name":"Alpha","latitude":57.7,"longitude":11.97,
  "dateCreated":"2024-04-28T10:00:00Z","totalTakeovers":7,"takeoverPoints":80,
  "pointsPerHour":3,"region":{"id":142,"name":"Göteborg","country":"SE"}}]`

func testRunner(t *testing.T, store Store, fetch FetchFunc) *runner {
	t.Helper()

	r, err := newRunner(Config{
		Fetch:           fetch,
		Store:           store,
		FetchTimeout:    time.Second,
		DatabaseTimeout: time.Second,
		MergeTimeout:    time.Second,
		MinZoneRatio:    0.9,
	})
	if err != nil {
		t.Fatalf("building the runner: %v", err)
	}

	return r
}

// okStore is a store that lets a run reach `ok`.
func okStore() *fakeStore {
	return &fakeStore{
		inspect: Staged{Rows: 1, DistinctIDs: 1, CurrentZoneRows: 1},
		merged:  Merged{Inserted: 1, Updated: 0},
	}
}

// TestEveryOutcomeIsReachableAndTheRowIsAlwaysWritten binds the instrument
// itself: `sync_run` is the only durable record of what the sync has done, and a
// record that cannot describe a failure describes nothing worth having.
//
// It asserts two things at once because they are one property. Each terminal
// value in the vocabulary of `Architecture.md § The sync write path` is reached
// from real code by a real failure — a value nobody can reach is a value the
// table cannot report — and each of those paths writes the row exactly once,
// which is the second of the two writes `migrations/README.md § Recording a run`
// puts on the worker.
func TestEveryOutcomeIsReachableAndTheRowIsAlwaysWritten(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		store func() *fakeStore
		fetch *fakeFetch
		want  Outcome
	}{
		{
			name:  "a merge that commits",
			store: okStore,
			fetch: &fakeFetch{body: []byte(oneZoneResponse), status: 200},
			want:  OutcomeOK,
		},
		{
			name:  "an endpoint that refuses the request",
			store: okStore,
			fetch: &fakeFetch{body: []byte(`{"error":"too many requests"}`), status: 429, err: errors.New("429")},
			want:  OutcomeHTTPError,
		},
		{
			name:  "a body that will not parse",
			store: okStore,
			fetch: &fakeFetch{body: []byte(`{"zones": "not an array"}`), status: 200},
			want:  OutcomeHTTPError,
		},
		{
			name: "a response the staging assertions reject",
			store: func() *fakeStore {
				s := okStore()
				// A response carrying one zone against a table holding a
				// thousand is the truncated response the floor exists for.
				s.inspect = Staged{Rows: 1, DistinctIDs: 1, CurrentZoneRows: 1000}

				return s
			},
			fetch: &fakeFetch{body: []byte(oneZoneResponse), status: 200},
			want:  OutcomeAssertionFailed,
		},
		{
			name: "a database error during the merge",
			store: func() *fakeStore {
				s := okStore()
				s.mergeErr = errStore

				return s
			},
			fetch: &fakeFetch{body: []byte(oneZoneResponse), status: 200},
			want:  OutcomeAborted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := tc.store()

			got := testRunner(t, store, tc.fetch.fetch).runOnce(t.Context())
			if got != tc.want {
				t.Errorf("the run reached outcome %q, want %q", got, tc.want)
			}

			recorded := store.outcomes()
			if len(recorded) != 1 {
				t.Fatalf("the run wrote %d terminal rows (%v), want exactly one — every path out of a run records itself", len(recorded), recorded)
			}

			if recorded[0] != tc.want {
				t.Errorf("sync_run recorded outcome %q, want %q", recorded[0], tc.want)
			}
		})
	}
}

// TestACancelledRunIsAbortedAndStillRecorded binds the fifth terminal value and
// the hardest of the write paths.
//
// A run cancelled at shutdown is `aborted` rather than `http_error`: the
// endpoint did nothing wrong. And the terminal write still has to land, which it
// can only do on a context detached from the cancelled one — a finalise
// inheriting ctx would fail exactly when there was something worth recording,
// and would leave the row at `running` for a run whose end is known.
func TestACancelledRunIsAbortedAndStillRecorded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	store := okStore()
	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	if got := testRunner(t, store, fetch.fetch).runOnce(ctx); got != OutcomeAborted {
		t.Errorf("a cancelled run reached outcome %q, want %q", got, OutcomeAborted)
	}

	recorded := store.outcomes()
	if len(recorded) != 1 || recorded[0] != OutcomeAborted {
		t.Fatalf("a cancelled run recorded %v, want exactly one %q — a run that died at shutdown must still be distinguishable from one that never happened", recorded, OutcomeAborted)
	}
}

// TestAPanickingRunIsStillRecorded is the path nobody plans for, which is the
// one `migrations/README.md § Recording a run` names specifically: the second
// write happens on every path out, including the ones the worker did not plan.
func TestAPanickingRunIsStillRecorded(t *testing.T) {
	t.Parallel()

	store := okStore()

	panicking := func(context.Context) ([]byte, int, error) { panic("the adapter exploded") }

	func() {
		defer func() {
			if recovered := recover(); recovered == nil {
				t.Error("the panic was swallowed, want it re-raised after the run was recorded")
			}
		}()

		_ = testRunner(t, store, panicking).runOnce(t.Context())
	}()

	recorded := store.outcomes()
	if len(recorded) != 1 || recorded[0] != OutcomeAborted {
		t.Fatalf("a panicking run recorded %v, want exactly one %q", recorded, OutcomeAborted)
	}
}

// TestARefusedResponseIsNeverMerged binds the staging assertions to their
// consequence rather than to their verdict. `Architecture.md § Absence is
// recorded and never acted on` argues the cost of merging a truncated response
// is unbounded in one direction, so the assertion returning reasons is not the
// property that matters — the merge not happening is.
func TestARefusedResponseIsNeverMerged(t *testing.T) {
	t.Parallel()

	store := okStore()
	store.inspect = Staged{Rows: 1, DistinctIDs: 1, CurrentZoneRows: 1000}

	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	_ = testRunner(t, store, fetch.fetch).runOnce(t.Context())

	if merges := store.mergeCount(); merges != 0 {
		t.Errorf("the merge ran %d times against a refused response, want 0", merges)
	}
}

// TestASuccessfulRunRecordsWhatItDid checks the figures `Architecture.md § The
// sync write path` says are worth watching from the first day, and the one they
// are derived from.
func TestASuccessfulRunRecordsWhatItDid(t *testing.T) {
	t.Parallel()

	store := okStore()
	store.inspect = Staged{Rows: 1, DistinctIDs: 1, CurrentZoneRows: 1}
	store.merged = Merged{Inserted: 0, Updated: 0, AbsentIDs: []int32{9, 12}}

	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	if got := testRunner(t, store, fetch.fetch).runOnce(t.Context()); got != OutcomeOK {
		t.Fatalf("the run reached %q, want %q", got, OutcomeOK)
	}

	results := store.results()
	if len(results) != 1 {
		t.Fatalf("the run recorded %d rows, want 1", len(results))
	}

	got := results[0]

	if got.CompletedAt == nil {
		t.Error("completed_at was not recorded, and it is what every changed row's last_changed_at is stamped with")
	}

	if got.ZonesReceived == nil || *got.ZonesReceived != 1 {
		t.Errorf("zones_received = %v, want 1", got.ZonesReceived)
	}

	// One zone received, none inserted, none updated: the remaining one was
	// unchanged, which is the figure the change-detection WHERE exists to make
	// large.
	if got.RowsUnchanged == nil || *got.RowsUnchanged != 1 {
		t.Errorf("rows_unchanged = %v, want 1", got.RowsUnchanged)
	}

	if got.AbsentCount == nil || *got.AbsentCount != 2 {
		t.Errorf("absent_count = %v, want 2 — absence is recorded even though nothing acts on it", got.AbsentCount)
	}
}

// TestAnUnknowableFigureIsRecordedAsUnknownRatherThanZero binds the reason the
// optional columns are nullable at all. A run that never received a response has
// not received zero zones, and a store that wrote 0 into every counter would make
// a failed fetch indistinguishable from an empty corpus.
func TestAnUnknowableFigureIsRecordedAsUnknownRatherThanZero(t *testing.T) {
	t.Parallel()

	store := okStore()
	fetch := &fakeFetch{err: errors.New("no route to host")}

	_ = testRunner(t, store, fetch.fetch).runOnce(t.Context())

	results := store.results()
	if len(results) != 1 {
		t.Fatalf("the run recorded %d rows, want 1", len(results))
	}

	got := results[0]

	if got.ZonesReceived != nil {
		t.Errorf("zones_received = %d for a run that received no response, want it left unrecorded", *got.ZonesReceived)
	}

	if got.HTTPStatus != nil {
		t.Errorf("http_status = %d for a request that got no response, want it left unrecorded — that absence is what separates a refused request from a body that would not parse", *got.HTTPStatus)
	}

	if got.CompletedAt != nil {
		t.Error("completed_at was recorded for a run that never merged, want it left unrecorded so currency cannot advance on a failure")
	}
}

// TestARunThatCannotRecordItsStartDoesNotFetch binds the ordering of the two
// writes. The row is inserted first, and if that insert fails there is no run:
// fetching anyway would spend the endpoint's allowance on an attempt nothing can
// record, and the next attempt would see no recorded attempt and spend it again.
func TestARunThatCannotRecordItsStartDoesNotFetch(t *testing.T) {
	t.Parallel()

	store := okStore()
	store.beginErr = errStore

	fetch := &fakeFetch{body: []byte(oneZoneResponse), status: 200}

	if got := testRunner(t, store, fetch.fetch).runOnce(t.Context()); got != OutcomeAborted {
		t.Errorf("the run reached %q, want %q", got, OutcomeAborted)
	}

	if calls := fetch.count(); calls != 0 {
		t.Errorf("the endpoint was called %d times by a run that could not record itself, want 0", calls)
	}
}
