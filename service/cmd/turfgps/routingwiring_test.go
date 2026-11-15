package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

const journeyBody = `{"origin":{"latitude":57.7036,"longitude":11.9611},"destination":{"latitude":57.7052,"longitude":11.9714}}`

// journeyRequest is journeyBody as the planner takes it.
var journeyRequest = planner.Request{
	Origin:      planner.Point{Latitude: 57.7036, Longitude: 11.9611},
	Destination: planner.Point{Latitude: 57.7052, Longitude: 11.9714},
}

func postJourney(t *testing.T, handler http.Handler) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/journeys", strings.NewReader(journeyBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

// heldZoneCopy reports a zone copy held, so a plan reaches its routing: the
// copy a test injects where the root reads `internal/zonestore`'s.
type heldZoneCopy struct{}

func (heldZoneCopy) Held(context.Context) (bool, error) { return true, nil }

// clearSHA256 is the extract clearAccess reports, and the one a test's tile
// record names where the two are to agree.
var clearSHA256 = strings.Repeat("0", 15) + "1" + strings.Repeat("c", 48)

// clearAccess is access markings that find nothing on any leg, loaded from
// clearSHA256: what a test injects where the root reads
// `internal/accessstore`'s. It also stands as routing data from the same
// extract where a test builds a planner itself.
type clearAccess struct{}

func (clearAccess) ExtractSHA256(context.Context) (string, error) { return clearSHA256, nil }

func (clearAccess) Crossings(context.Context, []planner.Edge) ([]planner.Crossing, error) {
	return nil, nil
}

func (clearAccess) Run(context.Context, []planner.Edge) (planner.RunMeasure, error) {
	return planner.RunMeasure{}, planner.ErrNoLeastRun
}

func lookupOnly(vars map[string]string) config.LookupFunc {
	return func(name string) (string, bool) {
		v, ok := vars[name]

		return v, ok
	}
}

// tileStamp is the tile stamp of sha, as `DEPLOYMENT.md § The tile build`
// step 1 derives it, written in decimal as the engine writes osm_changeset.
func tileStamp(t *testing.T, sha string) string {
	t.Helper()

	v, err := strconv.ParseUint(sha[:16], 16, 64)
	if err != nil {
		t.Fatalf("the stamp of %s: %v", sha, err)
	}

	return strconv.FormatUint(v, 10)
}

// writeTileRecord writes the tile record for sha, as step 4 writes it, and
// returns its path.
func writeTileRecord(t *testing.T, sha string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "tiles.sha256")
	if err := os.WriteFile(path, []byte(sha+"\n"), 0o600); err != nil {
		t.Fatalf("writing the tile record: %v", err)
	}

	return path
}

// fakeEngine answers as Valhalla does: `/status` with `osm_changeset` written
// verbatim where it is asked verbose and changeset is set, and one leg of two
// vertices — (57.7036, 11.9611) to (57.7052, 11.9714), encoded at six digits —
// along way 1, from `/route` and `/trace_attributes`. It records what it was
// asked.
type fakeEngine struct {
	URL string

	mu        sync.Mutex
	changeset string
	paths     []string
	costing   []any
}

func newFakeEngine(t *testing.T, changeset string) *fakeEngine {
	t.Helper()

	e := &fakeEngine{changeset: changeset}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)

		var body map[string]any
		_ = json.Unmarshal(raw, &body)

		e.mu.Lock()
		defer e.mu.Unlock()

		e.paths = append(e.paths, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		const shape = "_j}`mBwo`yU_cBwbS"

		switch r.URL.Path {
		case "/status":
			// Without verbose the engine answers before osm_changeset, as
			// it does where the verbose status is not allowed.
			if body["verbose"] != true || e.changeset == "" {
				_, _ = io.WriteString(w, `{"version":"3.9.1","tileset_last_modified":1}`)

				return
			}

			_, _ = io.WriteString(w, `{"version":"3.9.1","tileset_last_modified":1,"has_tiles":true,"osm_changeset":`+e.changeset+`}`)
		case "/trace_attributes":
			e.costing = append(e.costing, body["costing"])
			_, _ = io.WriteString(w, `{"shape":"`+shape+`","edges":[{"way_id":1,"begin_shape_index":0,"end_shape_index":1}]}`)
		case "/route":
			e.costing = append(e.costing, body["costing"])
			_, _ = io.WriteString(w, `{"trip":{"status":0,"legs":[{"shape":"`+shape+`","summary":{"time":90},"maneuvers":[{"travel_mode":"drive","travel_type":"car"}]}]}}`)
		default:
			http.Error(w, "not expected", http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	e.URL = srv.URL

	return e
}

// asked is the paths the engine was asked at, in order.
func (e *fakeEngine) asked() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return strings.Join(e.paths, " ")
}

// routed is how many legs the engine was asked to route.
func (e *fakeEngine) routed() int {
	return strings.Count(e.asked(), "/route")
}

// rootPlanner is the planner the composition root builds from lookup, over
// zones and markings in place of the store's — so its routing adapter and its
// tile identity are the root's own.
func rootPlanner(t *testing.T, lookup config.LookupFunc, zones planner.ZoneCopy, markings planner.AccessMarkings) *planner.Planner {
	t.Helper()

	cfg, err := loadPlanning(lookup)
	if err != nil || cfg == nil {
		t.Fatalf("loadPlanning gave %v, %v; want the planner's configuration", cfg, err)
	}

	plans, err := cfg.build(zones, markings)
	if err != nil || plans == nil {
		t.Fatalf("build gave %v, %v; want a planner", plans, err)
	}

	return plans
}

// TestStartupRefusesAMalformedRoutingEngineURL: a routing engine configured
// wrongly stops the service at start, with an error naming the setting that
// main logs before anything listens — not a service answering every journey
// with 502.
func TestStartupRefusesAMalformedRoutingEngineURL(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"127.0.0.1:8002", "tcp://127.0.0.1:8002", "http://127.0.0.1:8002/?json="} {
		plans, err := newPlanner(lookupOnly(map[string]string{config.EnvValhallaURL: raw}), unreachedPool())
		if err == nil {
			t.Errorf("newPlanner accepted %s=%q, want a refusal", config.EnvValhallaURL, raw)

			continue
		}

		if !strings.Contains(err.Error(), config.EnvValhallaURL) {
			t.Errorf("the refusal %q does not name %s", err, config.EnvValhallaURL)
		}

		if plans != nil {
			t.Errorf("newPlanner returned a planner beside its refusal of %q, want none", raw)
		}
	}
}

// TestWithoutARoutingEngineJourneysAnswer503: a service with no engine
// configured starts, and says on the route itself that it cannot plan.
func TestWithoutARoutingEngineJourneysAnswer503(t *testing.T) {
	t.Parallel()

	plans, err := newPlanner(lookupOnly(nil), unreachedPool())
	if err != nil || plans != nil {
		t.Fatalf("with no engine configured newPlanner gave %v, %v; want no planner and no error", plans, err)
	}

	handler, err := newHandler(lookupOnly(nil), httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	if rec := postJourney(t, handler); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("POST /journeys with no engine answered %d, want 503", rec.Code)
	}
}

// TestWithoutAStoreJourneysAnswer503 (#249): an engine configured with no
// background job, so no pool to read the store through, starts the service
// and plans nothing — there is no copy to plan from and no table to check a
// leg against — and the engine is never asked.
func TestWithoutAStoreJourneysAnswer503(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))
	lookup := lookupOnly(map[string]string{config.EnvValhallaURL: engine.URL})

	plans, err := newPlanner(lookup, nil)
	if err != nil || plans != nil {
		t.Fatalf("with no store newPlanner gave %v, %v; want no planner and no error", plans, err)
	}

	handler, err := newHandler(lookup, httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	if rec := postJourney(t, handler); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("POST /journeys with no store answered %d, want 503", rec.Code)
	}

	if got := engine.asked(); got != "" {
		t.Errorf("the engine was asked at %s, want nothing asked", got)
	}
}

// TestJourneysAreRoutedForACarThroughTheConfiguredEngine binds `FR-006` AC1 to
// the served route: with an engine configured, `POST /journeys` stops
// answering 503, and the leg it returns was asked of that engine with the car
// costing — the adapter, not a fake planner, is what the composition root
// registered. The zone copy and the markings are injected; the tile identity
// is the root's, read from the same engine before and after the leg.
func TestJourneysAreRoutedForACarThroughTheConfiguredEngine(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))
	lookup := lookupOnly(map[string]string{
		config.EnvValhallaURL:      engine.URL,
		config.EnvTileIdentityFile: writeTileRecord(t, clearSHA256),
	})

	handler, err := newHandler(lookup, httpapi.WithPlanner(rootPlanner(t, lookup, heldZoneCopy{}, clearAccess{})))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	rec := postJourney(t, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /journeys answered %d: %s", rec.Code, rec.Body)
	}

	var journey struct {
		Legs []struct {
			Shape []struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"shape"`
		} `json:"legs"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &journey); err != nil {
		t.Fatalf("the journey does not decode: %v: %s", err, rec.Body)
	}

	if len(journey.Legs) != 1 || len(journey.Legs[0].Shape) != 2 || journey.Legs[0].Shape[1].Longitude != 11.9714 {
		t.Errorf("the journey is %s, want the one leg the engine answered", rec.Body)
	}

	if got := engine.asked(); got != "/status /route /trace_attributes /status" {
		t.Errorf("the engine was asked at %s, want /status /route /trace_attributes /status", got)
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	if len(engine.costing) != 2 || engine.costing[0] != "auto" || engine.costing[1] != "auto" {
		t.Errorf("the engine was asked the costings %v, want auto at /route and at /trace_attributes", engine.costing)
	}
}

// deadlineRouting answers every leg at once, recording the deadline it was
// asked under and whether it had one.
type deadlineRouting struct {
	mu        sync.Mutex
	deadlines []time.Time
	bounded   []bool
}

func (d *deadlineRouting) Route(ctx context.Context, from, _ planner.Point, _ []planner.Point) (planner.Route, error) {
	deadline, ok := ctx.Deadline()

	d.mu.Lock()
	d.deadlines = append(d.deadlines, deadline)
	d.bounded = append(d.bounded, ok)
	d.mu.Unlock()

	return planner.Route{Edges: []planner.Edge{{WayID: 1, Shape: []planner.Point{from, from}}}}, nil
}

// TestEveryPlanIsBoundedBelowTheWriteTimeout (#6): the request surface the
// composition root builds plans every journey under planTimeout, measured from
// the handler's start, and planTimeout leaves the server's write budget room to
// write the refusal. Without it a plan routed one leg at a time outlasts
// writeTimeout and the client is cut off with nothing written.
func TestEveryPlanIsBoundedBelowTheWriteTimeout(t *testing.T) {
	t.Parallel()

	if planTimeout <= 0 || planTimeout >= writeTimeout {
		t.Fatalf("planTimeout is %v and writeTimeout %v, want a positive plan deadline below the write budget", planTimeout, writeTimeout)
	}

	routing := &deadlineRouting{}

	plans, err := planner.New(routing, heldZoneCopy{}, clearAccess{}, clearAccess{})
	if err != nil {
		t.Fatalf("planner.New: %v", err)
	}

	handler, err := newHandler(lookupOnly(nil), httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	before := time.Now()

	if rec := postJourney(t, handler); rec.Code != http.StatusOK {
		t.Fatalf("POST /journeys answered %d: %s", rec.Code, rec.Body)
	}

	after := time.Now()

	routing.mu.Lock()
	defer routing.mu.Unlock()

	if len(routing.bounded) != 1 {
		t.Fatalf("%d legs were routed, want the one the request names", len(routing.bounded))
	}

	if !routing.bounded[0] {
		t.Fatalf("the leg was routed with no deadline, want one planTimeout (%v) after the handler's start", planTimeout)
	}

	if got := routing.deadlines[0]; got.Before(before.Add(planTimeout)) || got.After(after.Add(planTimeout)) {
		t.Errorf("the leg's deadline is %v after the request, want planTimeout (%v)", got.Sub(before), planTimeout)
	}
}

// TestStartupRefusesAMalformedBarrierTolerance (#255): a barrier tolerance set
// without its origin, or not a number of metres, stops the service at start.
func TestStartupRefusesAMalformedBarrierTolerance(t *testing.T) {
	t.Parallel()

	for _, env := range []map[string]string{
		{config.EnvAccessBarrierTolerance: "2"},
		{config.EnvAccessBarrierTolerance: "-2", config.EnvAccessBarrierToleranceOrigin: "measured"},
	} {
		env[config.EnvValhallaURL] = "http://127.0.0.1:8002"

		if plans, err := newPlanner(lookupOnly(env), unreachedPool()); err == nil || plans != nil {
			t.Errorf("newPlanner with %v gave %v, %v; want a refusal", env, plans, err)
		}
	}
}

// TestStartupRefusesAMalformedAvoidanceBound (#248): an avoidance bound set
// without its origin, or not a whole number, stops the service at start; one
// set with its origin starts it.
func TestStartupRefusesAMalformedAvoidanceBound(t *testing.T) {
	t.Parallel()

	for _, env := range []map[string]string{
		{config.EnvAccessAvoidanceAttempts: "3"},
		{config.EnvAccessAvoidanceAttempts: "-1", config.EnvAccessAvoidanceAttemptsOrigin: "chosen"},
	} {
		env[config.EnvValhallaURL] = "http://127.0.0.1:8002"

		if plans, err := newPlanner(lookupOnly(env), unreachedPool()); err == nil || plans != nil {
			t.Errorf("newPlanner with %v gave %v, %v; want a refusal", env, plans, err)
		}
	}

	env := map[string]string{
		config.EnvValhallaURL:                   "http://127.0.0.1:8002",
		config.EnvAccessAvoidanceAttempts:       "3",
		config.EnvAccessAvoidanceAttemptsOrigin: "chosen",
	}

	if plans, err := newPlanner(lookupOnly(env), unreachedPool()); err != nil || plans == nil {
		t.Errorf("newPlanner with a bound and its origin gave %v, %v; want a planner", plans, err)
	}
}
