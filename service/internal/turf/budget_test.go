package turf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// Tests for `NFR-001`, the system-wide hold on outbound Turf requests (#28).
//
// ---------------------------------------------------------------------------
// WHY THESE DRIVE THE REAL SYNC. Read before simplifying.
//
// #28's notes warn that a limiter correct per caller and wrong in aggregate
// passes a careless test written from the metric alone, and that the
// concurrency in AC1's condition is what makes that reading fail. So AC1 here is
// several callers at once, each holding a Client OF ITS OWN — the shape a
// per-instance limiter satisfies one caller at a time — together with the
// scheduled sync of #20, which is `zonesync.Scheduler` itself and not a stand-in.
// A harness stands in for the solve sessions, which the ticket permits; the sync
// may not be stood in for, and is not.
//
// The scheduler is imported by this test file alone. The adapter's code does
// not know its consumer, and `go list` — which the off-request-path check reads
// — does not count a test's imports.
//
// NO FIGURE FROM EITHER LIMIT APPEARS HERE. The intervals below are this file's
// own and are deliberately unlike the API's, whose homes are named on Limits:
// what is under test is that whatever the configured interval is, it holds.
//
// TIMINGS ARE READ WHERE THE API WOULD READ THEM — on arrival at the server —
// and compared with no tolerance. The budget measures each interval from the
// moment the previous request's response arrived, which is after the server saw
// that request, so the gap on arrival is at least the interval whatever the
// scheduler does with either goroutine. "At least" and not "more than" because
// every reading here is one process's monotonic clock, which on some hosts
// ticks coarsely enough for two readings an interval apart to differ by exactly
// the interval. A tolerance here would be room for the defect.
// ---------------------------------------------------------------------------

// arrival is one request as the server saw it.
type arrival struct {
	at   time.Time
	path string
}

// recorder serves every path and records each request on arrival, before it
// answers, so that a request's arrival precedes its response.
type recorder struct {
	mu       sync.Mutex
	arrivals []arrival
}

func (r *recorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	r.arrivals = append(r.arrivals, arrival{at: time.Now(), path: req.URL.Path})
	r.mu.Unlock()

	_, _ = w.Write([]byte(`[]`))
}

func (r *recorder) seen() []arrival {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.arrivals)
}

// closestPair returns the shortest gap between successive arrivals and the
// pair that has it.
func closestPair(arrivals []arrival) (time.Duration, arrival, arrival) {
	sorted := slices.Clone(arrivals)
	slices.SortFunc(sorted, func(a, b arrival) int { return a.at.Compare(b.at) })

	closest := time.Duration(math.MaxInt64)

	var first, second arrival

	for i := 1; i < len(sorted); i++ {
		if gap := sorted[i].at.Sub(sorted[i-1].at); gap < closest {
			closest, first, second = gap, sorted[i-1], sorted[i]
		}
	}

	return closest, first, second
}

// waitFor polls cond until it holds, or fails the test after a bound no healthy
// run comes near.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)

	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("gave up waiting for %s", what)
		}

		time.Sleep(time.Millisecond)
	}
}

// standInStore is the sync's store, held in memory and dated by this process's
// clock. It refuses to stage, so every run ends after its fetch: the fetch is
// the one thing these tests measure, and what follows it is #20's.
type standInStore struct {
	mu        sync.Mutex
	lastBegun time.Time
	runs      int64
}

func (s *standInStore) SinceLastAttempt(context.Context) (time.Duration, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastBegun.IsZero() {
		return 0, false, nil
	}

	return time.Since(s.lastBegun), true, nil
}

func (s *standInStore) BeginRun(context.Context) (int64, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastBegun = time.Now()
	s.runs++

	return s.runs, s.lastBegun, nil
}

func (s *standInStore) FinishRun(context.Context, int64, zonesync.Result) error { return nil }

func (s *standInStore) Stage(context.Context, []zonesync.Zone) error {
	return errors.New("the stand-in store stages nothing")
}

func (s *standInStore) Inspect(context.Context, time.Time) (zonesync.Staged, error) {
	return zonesync.Staged{}, errors.New("the stand-in store inspects nothing")
}

func (s *standInStore) Merge(context.Context) (zonesync.Merged, error) {
	return zonesync.Merged{}, errors.New("the stand-in store merges nothing")
}

func (s *standInStore) Acquire(context.Context) (func(context.Context), bool, error) {
	return func(context.Context) {}, true, nil
}

// startSync runs the scheduled sync of #20 against srv through its own Client,
// and returns the function that stops it and waits for it to stop.
func startSync(t *testing.T, srv *httptest.Server, limits Limits, schedule, fetchTimeout time.Duration) func() {
	t.Helper()

	client, err := NewClient(srv.URL+currentPath, 1<<20, limits)
	if err != nil {
		t.Fatalf("building the sync's client: %v", err)
	}

	store := &standInStore{}

	scheduler, err := zonesync.NewScheduler(zonesync.Config{
		Fetch:           client.FetchAllZones,
		Store:           store,
		Locker:          store,
		Interval:        schedule,
		FetchTimeout:    fetchTimeout,
		DatabaseTimeout: time.Second,
		MergeTimeout:    time.Second,
		MinZoneRatio:    0.5,
	})
	if err != nil {
		t.Fatalf("building the scheduler: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)

		if err := scheduler.Run(ctx); err != nil {
			t.Errorf("the sync stopped with %v", err)
		}
	}()

	return func() {
		cancel()
		<-stopped
	}
}

// TestTheRequestLimitHoldsAcrossEveryCallerAtOnce is AC1: the per-resource
// limit, held across the whole system, every concurrent caller counted, the
// scheduled sync included.
func TestTheRequestLimitHoldsAcrossEveryCallerAtOnce(t *testing.T) {
	t.Parallel()

	const sessions = 4

	limits := Limits{Request: 20 * time.Millisecond, AllZones: 200 * time.Millisecond}

	rec := &recorder{}
	srv := httptest.NewServer(rec)
	defer srv.Close()

	countPath := func(path string) int {
		n := 0

		for _, a := range rec.seen() {
			if a.path == path {
				n++
			}
		}

		return n
	}

	// The solve sessions: each its own Client, each asking for its own
	// resource, back to back until told to stop.
	stop := make(chan struct{})

	var (
		running sync.WaitGroup
		failed  atomic.Int32
	)

	for n := range sessions {
		client, err := NewClient(srv.URL+currentPath, 1<<20, limits)
		if err != nil {
			t.Fatalf("building session %d's client: %v", n, err)
		}

		resource := fmt.Sprintf("%s/%s/session-stand-in-%d", srv.URL, APIVersion, n)

		running.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}

				req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, resource, nil)
				if err != nil {
					failed.Add(1)

					return
				}

				resp, err := client.http.Do(req)
				if err != nil {
					failed.Add(1)

					return
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
		})
	}

	// The sync starts once the sessions are already drawing on the budget, and
	// the sessions run on well past its request, so the two are measured
	// together rather than one after the other.
	waitFor(t, "the sessions' first request", func() bool { return len(rec.seen()) > 0 })

	stopSync := startSync(t, srv, limits, time.Hour, 10*time.Second)

	waitFor(t, "the sync's request", func() bool { return countPath(currentPath) > 0 })

	after := len(rec.seen())
	waitFor(t, "the sessions to run on past the sync's request", func() bool { return len(rec.seen()) >= after+2*sessions })

	close(stop)
	running.Wait()
	stopSync()

	if n := failed.Load(); n > 0 {
		t.Fatalf("%d session request(s) failed, so the sessions did not run for the whole measurement", n)
	}

	arrivals := rec.seen()

	if syncs := countPath(currentPath); syncs != 1 {
		t.Fatalf("the sync's endpoint was requested %d times, want exactly 1: the schedule permits one attempt in this test", syncs)
	}

	// The sync's request must sit among the sessions', or the two were never
	// concurrent and nothing below measures the aggregate.
	var sessionsBefore, sessionsAfter int

	var syncAt time.Time

	for _, a := range arrivals {
		if a.path == currentPath {
			syncAt = a.at
		}
	}

	for _, a := range arrivals {
		switch {
		case a.path == currentPath:
		case a.at.Before(syncAt):
			sessionsBefore++
		default:
			sessionsAfter++
		}
	}

	if sessionsBefore == 0 || sessionsAfter == 0 {
		t.Fatalf("the sync's request arrived with %d session request(s) before it and %d after, want some on each side: the sync and the sessions did not run together", sessionsBefore, sessionsAfter)
	}

	if gap, first, second := closestPair(arrivals); gap < limits.Request {
		t.Errorf("two of %d requests from %d sessions and the sync reached the API %s apart (%s, then %s), want at least the request interval %s between every two: each caller kept its own allowance, and together they spent more than the API's",
			len(arrivals), sessions, gap, first.path, second.path, limits.Request)
	}
}

// TestTheAllZonesIntervalHoldsUnderAScheduleThatAsksMoreOften is AC2: the
// all-zones interval, under a sync schedule far shorter than it.
//
// The configuration refuses such a schedule (`internal/config`), and that is
// the first guard. This is the second: a scheduler built with one anyway — the
// configuration bypassed, as a test or a later composition can — still does not
// make the requests. Both ways a request can be held are run: waited out, when
// the fetch's deadline is long enough to wait, and refused unsent, when it is
// not.
func TestTheAllZonesIntervalHoldsUnderAScheduleThatAsksMoreOften(t *testing.T) {
	t.Parallel()

	const (
		schedule = 5 * time.Millisecond
		requests = 4
	)

	limits := Limits{Request: time.Millisecond, AllZones: 150 * time.Millisecond}

	for name, fetchTimeout := range map[string]time.Duration{
		"a fetch that can wait is held until the interval has passed": 5 * time.Second,
		"a fetch that cannot wait is refused unsent":                  20 * time.Millisecond,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := &recorder{}
			srv := httptest.NewServer(rec)
			defer srv.Close()

			stopSync := startSync(t, srv, limits, schedule, fetchTimeout)

			waitFor(t, fmt.Sprintf("%d all-zones requests", requests), func() bool { return len(rec.seen()) >= requests })
			stopSync()

			arrivals := rec.seen()

			if gap, _, _ := closestPair(arrivals); gap < limits.AllZones {
				t.Errorf("the sync, scheduled every %s, reached the all-zones endpoint twice %s apart among %d requests, want at least its interval %s between every two: the schedule decided the rate, not the endpoint's limit",
					schedule, gap, len(arrivals), limits.AllZones)
			}
		})
	}
}

// TestARequestTheLimitCannotPermitInTimeIsNotSent pins the refusal itself: no
// request leaves, and the error says why rather than reading as a slow API.
func TestARequestTheLimitCannotPermitInTimeIsNotSent(t *testing.T) {
	t.Parallel()

	rec := &recorder{}
	srv := httptest.NewServer(rec)
	defer srv.Close()

	client, err := NewClient(srv.URL+currentPath, 1<<20, Limits{Request: time.Millisecond, AllZones: 500 * time.Millisecond})
	if err != nil {
		t.Fatalf("building the client: %v", err)
	}

	if _, _, err := client.FetchAllZones(t.Context()); err != nil {
		t.Fatalf("the first fetch: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, status, err := client.FetchAllZones(ctx)
	if !errors.Is(err, errNotPermitted) {
		t.Fatalf("a second fetch inside the all-zones interval returned %v, want it refused as not permitted", err)
	}

	if status != 0 {
		t.Errorf("the refused fetch reported status %d, want 0: no response was received, which is what tells a refused request from an answered one on the run's row", status)
	}

	if n := len(rec.seen()); n != 1 {
		t.Errorf("the API saw %d requests, want 1: the refused one was sent", n)
	}
}

// TestTheLimitsAreRequired refuses a client with no limit to hold, rather than
// one that falls back on a figure this module does not own.
func TestTheLimitsAreRequired(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name    string
		limits  Limits
		refusal string
		crossed string
	}{
		{"no request interval", Limits{AllZones: time.Minute}, "request interval", "all-zones"},
		{"no all-zones interval", Limits{Request: time.Minute}, "all-zones", "request interval"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewClient("https://example.test"+currentPath, 1<<20, c.limits)
			if err == nil {
				t.Fatalf("a client was built with %+v, want it refused", c.limits)
			}

			if !strings.Contains(err.Error(), c.refusal) || strings.Contains(err.Error(), c.crossed) {
				t.Errorf("the refusal reads %q, want it to name the %s and not the %s", err, c.refusal, c.crossed)
			}
		})
	}
}
