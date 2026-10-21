package turf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// Tests for the zone feed's request (#232), through the real FeedClient and the
// process's one budget. Counted where the API would count it, on arrival.

var feedPath = "/" + APIVersion + "/feeds/zone"

func feedClientFor(t *testing.T, srv *httptest.Server, limits Limits) *FeedClient {
	t.Helper()

	c, err := NewFeedClient(srv.URL+feedPath, 1<<20, limits)
	if err != nil {
		t.Fatalf("building the feed client: %v", err)
	}

	return c
}

// TestTheFeedRequestCarriesTheLowerBoundItIsGiven is AC2's wire half
// (`FR-120`): no lower bound when none is given, and the given one in the form
// and encoding the API index documents — `afterDate=2013-08-27T12%3A11%3A14%2B0000`.
func TestTheFeedRequestCarriesTheLowerBoundItIsGiven(t *testing.T) {
	t.Parallel()

	var (
		mu      sync.Mutex
		queries []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		queries = append(queries, r.URL.Path+"?"+r.URL.RawQuery)
		mu.Unlock()

		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	client := feedClientFor(t, srv, testLimits)

	// Given in another zone on purpose: the bound is an instant, and the
	// request names it in UTC.
	after := time.Date(2026, 10, 19, 23, 46, 7, 0, time.FixedZone("CEST", 2*60*60))

	for _, a := range []*time.Time{nil, &after} {
		if _, err := client.FetchZoneFeed(t.Context(), a); err != nil {
			t.Fatalf("the feed request failed: %v", err)
		}
	}

	want := []string{feedPath + "?", feedPath + "?afterDate=2026-10-19T21%3A46%3A07%2B0000"}

	mu.Lock()
	defer mu.Unlock()

	if len(queries) != 2 || queries[0] != want[0] || queries[1] != want[1] {
		t.Errorf("the feed was requested as %q, want %q", queries, want)
	}
}

// feedStandInStore is the feed's store held in memory: no feed time recorded,
// and every batch accepted.
type feedStandInStore struct{}

func (feedStandInStore) FeedCursor(context.Context) (*time.Time, time.Time, error) {
	return nil, time.Now(), nil
}

func (feedStandInStore) ApplyFeed(context.Context, zonesync.FeedBatch) (zonesync.FeedApplied, error) {
	return zonesync.FeedApplied{}, nil
}

// TestTheFeedJobDrawsOnTheProcessBudget is AC1's budget half (`FR-119`,
// `NFR-001`): the feed job's requests, issued by the real FeedPoller on a
// schedule far shorter than the limit, and review-session stand-ins each
// holding a Client of their own, reach the API no closer together than the
// general limit allows — one budget across all of them, as
// TestTheRequestLimitHoldsAcrossEveryCallerAtOnce measures for the sync.
func TestTheFeedJobDrawsOnTheProcessBudget(t *testing.T) {
	t.Parallel()

	const sessions = 3

	limits := Limits{Request: 20 * time.Millisecond, AllZones: time.Hour}

	rec := &recorder{}
	srv := httptest.NewServer(rec)
	defer srv.Close()

	feeds := func() int {
		n := 0

		for _, a := range rec.seen() {
			if a.path == feedPath {
				n++
			}
		}

		return n
	}

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

	waitFor(t, "the sessions' first request", func() bool { return len(rec.seen()) > 0 })

	poller, err := zonesync.NewFeedPoller(zonesync.FeedConfig{
		Fetch:           feedClientFor(t, srv, limits).FetchZoneFeed,
		Store:           feedStandInStore{},
		Interval:        time.Millisecond,
		FetchTimeout:    10 * time.Second,
		DatabaseTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("building the feed poller: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	polled := make(chan struct{})

	go func() {
		defer close(polled)

		_ = poller.Run(ctx)
	}()

	waitFor(t, "three feed requests", func() bool { return feeds() >= 3 })

	cancel()
	<-polled
	close(stop)
	running.Wait()

	if n := failed.Load(); n > 0 {
		t.Fatalf("%d session request(s) failed, so the sessions did not run for the whole measurement", n)
	}

	arrivals := rec.seen()

	if feeds() == len(arrivals) {
		t.Fatalf("only feed requests arrived, so the feed and the sessions were never measured together")
	}

	if gap, first, second := closestPair(arrivals); gap < limits.Request {
		t.Errorf("two of %d requests reached the API %s apart (%s, then %s), want at least the request interval %s: the feed job kept an allowance of its own",
			len(arrivals), gap, first.path, second.path, limits.Request)
	}
}

func TestAFeedResponseTheClientCannotUseIsAnError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{"a refusal", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"errorMessage":"nope"}`, http.StatusServiceUnavailable)
		}, "503"},
		{"an oversized body", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Repeat(" ", 2<<20)))
		}, "ceiling"},
		{"a redirect", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/"+APIVersion+"/feeds/other", http.StatusFound)
		}, "redirect"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			body, err := feedClientFor(t, srv, testLimits).FetchZoneFeed(t.Context(), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("%s gave body %d bytes and error %v, want an error naming %q", tc.name, len(body), err, tc.want)
			}

			if tc.name == "a redirect" && !errors.Is(err, errFeedRedirect) {
				t.Errorf("a redirect gave %v, want errFeedRedirect", err)
			}
		})
	}
}

func TestTheFeedEndpointMustAddressTheCurrentVersion(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{"", "https://api.example/v4/feeds/zone", "https://api.example/" + APIVersion + "/../v4/feeds/zone", "://"} {
		if _, err := NewFeedClient(bad, 1<<20, testLimits); err == nil {
			t.Errorf("the zone-feed endpoint %q was accepted", bad)
		}
	}

	if _, err := NewFeedClient("https://api.example"+feedPath, 1<<20, Limits{}); err == nil {
		t.Error("a feed client with no request interval was built")
	}

	if _, err := NewFeedClient("https://api.example"+feedPath, 0, testLimits); err == nil {
		t.Error("a feed client with no response ceiling was built")
	}
}
