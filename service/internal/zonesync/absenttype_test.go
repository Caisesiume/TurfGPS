package zonesync

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// recordingHandler keeps every record a run logs, so a test can ask what the
// run raised as well as what it returned.
//
// It shares one slice across WithAttrs and WithGroup, because the runner logs
// through a logger derived with the run's id and those records are the ones
// that matter.
type recordingHandler struct {
	mu      sync.Mutex
	records []string
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if r.Level >= slog.LevelWarn {
		h.records = append(h.records, fmt.Sprintf("%s %q", r.Level, r.Message))
	}

	return nil
}

func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *recordingHandler) WithGroup(string) slog.Handler { return h }

func (h *recordingHandler) raised() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]string(nil), h.records...)
}

// TestAZoneSetWithoutTypeIsIngestedWholeAndRaisesNothing is `FR-036` (#39) over
// the run: a set in which some records carry no `type` is ingested exactly as
// a set in which every record carries one, every record reaches the store, and
// nothing is raised at warn or above on account of the absence.
//
// THE MIXED SET IS THE FIXTURE, AND THE ALL-TYPED SET IS ITS CONTROL. The
// assumption that `type` is always present can sit in a parser guard, a
// validator, or a log line written as good hygiene, and one set carrying
// records without the field fails every one of those forms at once. The
// control is what makes "as it does for a set in which every record carries
// one" an assertion rather than a phrase. What PostgreSQL then stores is
// `internal/syncstore`'s postgis_test.go.
func TestAZoneSetWithoutTypeIsIngestedWholeAndRaisesNothing(t *testing.T) {
	t.Parallel()

	const record = `{"id":%d,"name":"zone %d","latitude":59.3,"longitude":18.06,` +
		`"dateCreated":"2020-01-02T03:04:05Z","totalTakeovers":4,"takeoverPoints":120,` +
		`"pointsPerHour":3,%s"region":{"id":141,"name":"a region"}}`

	for name, typed := range map[string][]bool{
		"every record carries a type": {true, true, true, true},
		"some records carry none":     {true, false, true, false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			records := make([]string, len(typed))
			untyped := 0

			for i, hasType := range typed {
				typ := `"type":{"id":3,"name":"a type"},`
				if !hasType {
					typ = ""
					untyped++
				}

				records[i] = fmt.Sprintf(record, i+1, i+1, typ)
			}

			n := len(records)
			store := &fakeStore{
				inspect: Staged{Rows: n, DistinctIDs: n, CurrentZoneRows: n},
				merged:  Merged{Inserted: n},
			}
			fetch := &fakeFetch{body: []byte("[" + strings.Join(records, ",") + "]"), status: 200}
			logs := &recordingHandler{}

			r, err := newRunner(Config{
				Fetch:           fetch.fetch,
				Store:           store,
				Logger:          slog.New(logs),
				FetchTimeout:    time.Second,
				DatabaseTimeout: time.Second,
				MergeTimeout:    time.Second,
				MinZoneRatio:    1,
			})
			if err != nil {
				t.Fatalf("building the runner: %v", err)
			}

			if got := r.runOnce(t.Context()); got != OutcomeOK {
				t.Errorf("the run reached %q, want %q — a set of %d records, %d of them without a type, is ingested as a set in which every record carries one", got, OutcomeOK, n, untyped)
			}

			store.mu.Lock()
			staged := store.staged
			store.mu.Unlock()

			if len(staged) != 1 || len(staged[0]) != n {
				t.Fatalf("the store was handed %v, want one staging load of all %d records", stagedSizes(staged), n)
			}

			// What the absent field becomes once staged is deliberately not
			// asserted. Whether it may be substituted is a duty #39 records as
			// owed elsewhere and binds nothing here.
			if raised := logs.raised(); len(raised) > 0 {
				t.Errorf("the run raised %v at warn or above over a set the run accepted — a signal on the ordinary case fires on most of the corpus and is soon unread", raised)
			}
		})
	}
}

func stagedSizes(staged [][]Zone) []int {
	sizes := make([]int, len(staged))
	for i, s := range staged {
		sizes[i] = len(s)
	}

	return sizes
}
