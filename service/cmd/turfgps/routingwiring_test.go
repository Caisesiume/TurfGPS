package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

const journeyBody = `{"origin":{"latitude":57.7036,"longitude":11.9611},"destination":{"latitude":57.7052,"longitude":11.9714}}`

func postJourney(t *testing.T, handler http.Handler) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/journeys", strings.NewReader(journeyBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

// heldZoneCopy reports a zone copy held, so a plan reaches its routing: the
// copy a test injects where main passes unwiredZoneCopy.
type heldZoneCopy struct{}

func (heldZoneCopy) Held(context.Context) (bool, error) { return true, nil }

// clearAccess is access markings that find nothing on any leg and routing
// data from the same extract: what a test injects where main passes
// unwiredAccessMarkings and unknownRoutingData.
type clearAccess struct{}

func (clearAccess) ExtractSHA256(context.Context) (string, error) {
	return strings.Repeat("0", 64), nil
}

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

// TestStartupRefusesAMalformedRoutingEngineURL: a routing engine configured
// wrongly stops the service at start, with an error naming the setting that
// main logs before anything listens — not a service answering every journey
// with 502.
func TestStartupRefusesAMalformedRoutingEngineURL(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"127.0.0.1:8002", "tcp://127.0.0.1:8002", "http://127.0.0.1:8002/?json="} {
		plans, err := newPlanner(lookupOnly(map[string]string{config.EnvValhallaURL: raw}), heldZoneCopy{}, clearAccess{}, clearAccess{})
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

	plans, err := newPlanner(lookupOnly(nil), heldZoneCopy{}, clearAccess{}, clearAccess{})
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

// TestJourneysAreRoutedForACarThroughTheConfiguredEngine binds `FR-006` AC1 to
// the served route: with an engine configured, `POST /journeys` stops
// answering 503, and the leg it returns was asked of that engine with the car
// costing — the adapter, not a fake planner, is what the composition root
// registered. The zone copy is injected held, as main's is not until #249, so
// that the plan reaches the engine.
func TestJourneysAreRoutedForACarThroughTheConfiguredEngine(t *testing.T) {
	t.Parallel()

	var (
		mu      sync.Mutex
		costing []any
		paths   []string
	)

	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)

		var body map[string]any
		_ = json.Unmarshal(raw, &body)

		mu.Lock()
		costing = append(costing, body["costing"])
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		// Two vertices, (57.7036, 11.9611) and (57.7052, 11.9714), encoded at
		// six digits; the route API reference documents the shape, and the
		// map-matching reference the trace's one edge along it.
		w.Header().Set("Content-Type", "application/json")

		const shape = "_j}`mBwo`yU_cBwbS"

		if r.URL.Path == "/trace_attributes" {
			_, _ = io.WriteString(w, `{"shape":"`+shape+`","edges":[{"way_id":1,"begin_shape_index":0,"end_shape_index":1}]}`)

			return
		}

		_, _ = io.WriteString(w, `{"trip":{"status":0,"legs":[{"shape":"`+shape+`","maneuvers":[{"travel_mode":"drive","travel_type":"car"}]}]}}`)
	}))
	t.Cleanup(engine.Close)

	lookup := lookupOnly(map[string]string{config.EnvValhallaURL: engine.URL})

	plans, err := newPlanner(lookup, heldZoneCopy{}, clearAccess{}, clearAccess{})
	if err != nil || plans == nil {
		t.Fatalf("newPlanner with %s=%q gave %v, %v; want a planner", config.EnvValhallaURL, engine.URL, plans, err)
	}

	handler, err := newHandler(lookup, httpapi.WithPlanner(plans))
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

	mu.Lock()
	defer mu.Unlock()

	if len(costing) != 2 || costing[0] != "auto" || paths[0] != "/route" || paths[1] != "/trace_attributes" {
		t.Errorf("the engine was asked %v at %v, want the auto costing once at /route and once at /trace_attributes", costing, paths)
	}
}

// deadlineRouting answers every leg at once, recording the deadline it was
// asked under and whether it had one.
type deadlineRouting struct {
	mu        sync.Mutex
	deadlines []time.Time
	bounded   []bool
}

func (d *deadlineRouting) Route(ctx context.Context, from, _ planner.Point, _ []planner.Point) ([]planner.Edge, error) {
	deadline, ok := ctx.Deadline()

	d.mu.Lock()
	d.deadlines = append(d.deadlines, deadline)
	d.bounded = append(d.bounded, ok)
	d.mu.Unlock()

	return []planner.Edge{{WayID: 1, Shape: []planner.Point{from, from}}}, nil
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

// TestTheServedPlannerFailsClosedUntilTheZoneStoreIsWired (#6, the lead's
// ruling on #246): with an engine configured, the copy main plans from refuses
// every plan with 503 and the store's message, and no leg is asked of the
// engine — until #249 gives the request surface a pool on the zone store.
func TestTheServedPlannerFailsClosedUntilTheZoneStoreIsWired(t *testing.T) {
	t.Parallel()

	var (
		mu    sync.Mutex
		asked int
	)

	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		asked++
		mu.Unlock()

		http.Error(w, "not expected", http.StatusInternalServerError)
	}))
	t.Cleanup(engine.Close)

	lookup := lookupOnly(map[string]string{config.EnvValhallaURL: engine.URL})

	plans, err := newPlanner(lookup, unwiredZoneCopy{}, unwiredAccessMarkings{}, unknownRoutingData{})
	if err != nil || plans == nil {
		t.Fatalf("newPlanner with %s=%q gave %v, %v; want a planner", config.EnvValhallaURL, engine.URL, plans, err)
	}

	handler, err := newHandler(lookup, httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	rec := postJourney(t, handler)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "could not be read") {
		t.Errorf("POST /journeys answered %d %q, want 503 saying the zone copy could not be read", rec.Code, rec.Body)
	}

	mu.Lock()
	defer mu.Unlock()

	if asked != 0 {
		t.Errorf("the engine was asked %d times, want none before the zone copy is read", asked)
	}
}

// TestTheServedPlannerPlansNothingUntilTheAccessCheckIsWired (#255): even with
// a zone copy held, the access markings and routing data main registers
// refuse every plan before a leg is routed — no leg is served until the
// request surface has a pool on the access-marking tables (#249) and the tile
// set's identity can be read (#25).
func TestTheServedPlannerPlansNothingUntilTheAccessCheckIsWired(t *testing.T) {
	t.Parallel()

	var (
		mu    sync.Mutex
		asked int
	)

	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		asked++
		mu.Unlock()

		http.Error(w, "not expected", http.StatusInternalServerError)
	}))
	t.Cleanup(engine.Close)

	lookup := lookupOnly(map[string]string{config.EnvValhallaURL: engine.URL})

	plans, err := newPlanner(lookup, heldZoneCopy{}, unwiredAccessMarkings{}, unknownRoutingData{})
	if err != nil || plans == nil {
		t.Fatalf("newPlanner gave %v, %v; want a planner", plans, err)
	}

	journey, err := plans.Plan(t.Context(), planner.Request{
		Origin:      planner.Point{Latitude: 57.7036, Longitude: 11.9611},
		Destination: planner.Point{Latitude: 57.7052, Longitude: 11.9714},
	})
	if !errors.Is(err, planner.ErrAccessUnchecked) || journey.Legs != nil {
		t.Fatalf("Plan gave %d legs and %v, want no journey and an error wrapping ErrAccessUnchecked", len(journey.Legs), err)
	}

	handler, err := newHandler(lookup, httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	if rec := postJourney(t, handler); rec.Code == http.StatusOK {
		t.Errorf("POST /journeys answered 200 %s, want a refusal", rec.Body)
	}

	mu.Lock()
	defer mu.Unlock()

	if asked != 0 {
		t.Errorf("the engine was asked %d times, want none before the access markings are read", asked)
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

		if plans, err := newPlanner(lookupOnly(env), heldZoneCopy{}, unwiredAccessMarkings{}, unknownRoutingData{}); err == nil || plans != nil {
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

		if plans, err := newPlanner(lookupOnly(env), heldZoneCopy{}, unwiredAccessMarkings{}, unknownRoutingData{}); err == nil || plans != nil {
			t.Errorf("newPlanner with %v gave %v, %v; want a refusal", env, plans, err)
		}
	}

	env := map[string]string{
		config.EnvValhallaURL:                   "http://127.0.0.1:8002",
		config.EnvAccessAvoidanceAttempts:       "3",
		config.EnvAccessAvoidanceAttemptsOrigin: "chosen",
	}

	if plans, err := newPlanner(lookupOnly(env), heldZoneCopy{}, unwiredAccessMarkings{}, unknownRoutingData{}); err != nil || plans == nil {
		t.Errorf("newPlanner with a bound and its origin gave %v, %v; want a planner", plans, err)
	}
}
