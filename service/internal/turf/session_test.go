package turf

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/ownership"
)

// Tests for the review session's Turf requests: `FR-028` (#23) and `FR-029`
// (#24) end to end, through the real SessionClient and the process's one
// budget, with every request COUNTED WHERE THE API WOULD COUNT IT — on arrival
// at the server. A count of calls to a fake port would pass an adapter that
// issued two HTTP requests per call; this one does not.
//
// The domain package is imported by this test file alone, as budget_test.go
// imports the sync: the adapter's code does not know its consumer.

var (
	usersPath   = "/" + APIVersion + "/users"
	regionsPath = "/" + APIVersion + "/regions"
)

// sessionRequest is one request as the stand-in saw it.
type sessionRequest struct {
	method, path, contentType, body string
}

// turfStandIn answers the two session endpoints with fixed bodies and records
// every request it receives, on arrival.
type turfStandIn struct {
	users, regions string

	mu   sync.Mutex
	seen []sessionRequest
}

func (s *turfStandIn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	s.mu.Lock()
	s.seen = append(s.seen, sessionRequest{r.Method, r.URL.Path, r.Header.Get("Content-Type"), string(body)})
	s.mu.Unlock()

	switch r.URL.Path {
	case usersPath:
		_, _ = w.Write([]byte(s.users))
	case regionsPath:
		_, _ = w.Write([]byte(s.regions))
	default:
		http.NotFound(w, r)
	}
}

func (s *turfStandIn) requests() []sessionRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	return slices.Clone(s.seen)
}

func (s *turfStandIn) count(path string) int {
	n := 0

	for _, r := range s.requests() {
		if r.path == path {
			n++
		}
	}

	return n
}

func sessionFor(t *testing.T, stand *turfStandIn) *SessionClient {
	t.Helper()

	srv := httptest.NewServer(stand)
	t.Cleanup(srv.Close)

	c, err := NewSessionClient(SessionEndpoints{Users: srv.URL + usersPath, Regions: srv.URL + regionsPath}, 1<<20, testLimits)
	if err != nil {
		t.Fatalf("building the session client: %v", err)
	}

	return c
}

// The planning user of these tests, and the player-data response for them:
// they hold zones 101 and 103 and no other.
const (
	planner      = "planner"
	plannerID    = 4242
	plannerUsers = `[{"id":4242,"name":"planner","rank":12,"blocktime":3600,"zones":[101,103],"region":{"id":141,"name":"Stockholm"}}]`
)

// TestEveryZoneIsDecidedFromTheOneHeldZoneList is #23 AC1: a route of held and
// unheld zones, each decided by membership, and no request after the one
// player-data call.
func TestEveryZoneIsDecidedFromTheOneHeldZoneList(t *testing.T) {
	t.Parallel()

	stand := &turfStandIn{users: plannerUsers}
	client := sessionFor(t, stand)

	player, err := ownership.LoadPlayer(t.Context(), client.FetchPlayer, planner)
	if err != nil {
		t.Fatalf("the player-data call: %v", err)
	}

	after := stand.requests()
	if len(after) != 1 || after[0].method != http.MethodPost || after[0].path != usersPath {
		t.Fatalf("the session opened with requests %+v, want exactly one POST to %s", after, usersPath)
	}

	if after[0].body != `[{"name":"planner"}]` || after[0].contentType != "application/json" {
		t.Errorf("the player-data request carried %q as %q, want [{\"name\":\"planner\"}] as application/json, the form the API documents", after[0].body, after[0].contentType)
	}

	route := []struct {
		zone int32
		held bool
	}{{101, true}, {102, false}, {103, true}, {104, false}}

	for _, stop := range route {
		if got := player.Holds(stop.zone); got != stop.held {
			t.Errorf("zone %d decided held=%v, want %v: the held-zone list is [101,103]", stop.zone, got, stop.held)
		}
	}

	if n := len(stand.requests()); n != 1 {
		t.Errorf("reviewing the route made %d Turf requests in all, want 1: FR-028 decides every zone from the one held-zone list and issues no further request", n)
	}
}

// TestReviewingMoreZonesCostsNoMoreRequests is #23 AC2: once the first zone is
// decided, deciding further ones leaves the session's request count unchanged.
func TestReviewingMoreZonesCostsNoMoreRequests(t *testing.T) {
	t.Parallel()

	stand := &turfStandIn{users: plannerUsers}
	client := sessionFor(t, stand)

	player, err := ownership.LoadPlayer(t.Context(), client.FetchPlayer, planner)
	if err != nil {
		t.Fatalf("the player-data call: %v", err)
	}

	player.Holds(101)

	before := len(stand.requests())

	for zone := int32(1); zone <= 200; zone++ {
		player.Holds(zone)
	}

	if after := len(stand.requests()); after != before {
		t.Errorf("reviewing 200 further zones took the session from %d Turf requests to %d, want it unchanged", before, after)
	}
}

// TestTheSessionDrawsOnTheOneBudget holds `NFR-001` for the session's requests:
// a SessionClient and the sync's Client against one API are spaced by the
// general limit between them, whichever sends first.
func TestTheSessionDrawsOnTheOneBudget(t *testing.T) {
	t.Parallel()

	limits := Limits{Request: 30 * time.Millisecond, AllZones: time.Millisecond}

	rec := &recorder{}
	srv := httptest.NewServer(rec)
	defer srv.Close()

	syncClient, err := NewClient(srv.URL+currentPath, 1<<20, limits)
	if err != nil {
		t.Fatalf("building the sync's client: %v", err)
	}

	session, err := NewSessionClient(SessionEndpoints{Users: srv.URL + usersPath, Regions: srv.URL + regionsPath}, 1<<20, limits)
	if err != nil {
		t.Fatalf("building the session client: %v", err)
	}

	var wg sync.WaitGroup

	for range 3 {
		wg.Go(func() { _, _, _ = syncClient.FetchAllZones(t.Context()) })
		wg.Go(func() { _, _ = session.FetchPlayer(t.Context(), planner) })
	}

	wg.Wait()

	arrivals := rec.seen()
	if len(arrivals) != 6 {
		t.Fatalf("the server saw %d requests, want 6", len(arrivals))
	}

	if gap, a, b := closestPair(arrivals); gap < limits.Request {
		t.Errorf("%s and %s arrived %s apart, inside the %s limit: the session's requests are not drawn from the one budget", a.path, b.path, gap, limits.Request)
	}
}

// TestASessionRedirectIsRefused: a redirect is answered with an error naming
// it, and the hop is never requested. A 302 would otherwise have turned the
// player-data POST into a GET.
func TestASessionRedirectIsRefused(t *testing.T) {
	t.Parallel()

	var hops atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc(usersPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+APIVersion+"/elsewhere", http.StatusFound)
	})
	mux.HandleFunc("/"+APIVersion+"/elsewhere", func(w http.ResponseWriter, _ *http.Request) {
		hops.Add(1)
		_, _ = w.Write([]byte(`[]`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewSessionClient(SessionEndpoints{Users: srv.URL + usersPath, Regions: srv.URL + regionsPath}, 1<<20, testLimits)
	if err != nil {
		t.Fatalf("building the session client: %v", err)
	}

	if _, err := c.FetchPlayer(t.Context(), planner); !errors.Is(err, errRedirect) {
		t.Errorf("a redirected player-data request returned %v, want the redirect refused", err)
	}

	if n := hops.Load(); n != 0 {
		t.Errorf("the redirect target was requested %d times, want 0", n)
	}
}

// TestASessionResponseOtherThan200IsAnError: a refusal is never handed on as
// a body to decide from.
func TestASessionResponseOtherThan200IsAnError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"errorMessage":"Only one request per second allowed"}`, http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c, err := NewSessionClient(SessionEndpoints{Users: srv.URL + usersPath, Regions: srv.URL + regionsPath}, 1<<20, testLimits)
	if err != nil {
		t.Fatalf("building the session client: %v", err)
	}

	if body, err := c.FetchPlayer(context.Background(), planner); err == nil || body != nil || !strings.Contains(err.Error(), "429") {
		t.Errorf("a 429 returned body %q and error %v, want no body and an error naming the status", body, err)
	}
}

// TestASessionEndpointMustAddressTheCurrentVersion holds `FR-019` for both
// endpoints, and refuses a client with no limit to draw on.
func TestASessionEndpointMustAddressTheCurrentVersion(t *testing.T) {
	t.Parallel()

	good := "https://example.test" + usersPath
	stale := "https://example.test/users" // addresses no version: FR-019 refuses it

	for name, c := range map[string]struct {
		endpoints SessionEndpoints
		limits    Limits
	}{
		"stale users":   {SessionEndpoints{Users: stale, Regions: good}, testLimits},
		"stale regions": {SessionEndpoints{Users: good, Regions: stale}, testLimits},
		"no users":      {SessionEndpoints{Regions: good}, testLimits},
		"no limit":      {SessionEndpoints{Users: good, Regions: good}, Limits{}},
	} {
		if _, err := NewSessionClient(c.endpoints, 1<<20, c.limits); err == nil {
			t.Errorf("%s: NewSessionClient accepted it, want it refused", name)
		}
	}
}
