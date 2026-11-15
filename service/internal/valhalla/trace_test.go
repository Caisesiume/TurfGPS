package valhalla

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// The trace half of Route (#255, `Architecture.md § The route check`, step 2):
// every leg's shape is sent back to the engine's map-matching service, and the
// leg is returned as the ways the answer names. That a real engine names them
// for a real route is the integration test's to show.

var (
	traceA = planner.Point{Latitude: 57.700000, Longitude: 11.960000}
	traceB = planner.Point{Latitude: 57.700500, Longitude: 11.961000}
	traceC = planner.Point{Latitude: 57.701000, Longitude: 11.962000}
	traceD = planner.Point{Latitude: 57.701500, Longitude: 11.963000}
)

// startFake serves engine and returns a Client for it.
func startFake(t *testing.T, engine *fakeEngine) *Client {
	t.Helper()

	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	c, err := New(srv.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

// routeOf answers every route request with one car leg along shape.
func routeOf(shape []planner.Point) func(map[string]any) (int, string) {
	return func(map[string]any) (int, string) {
		return http.StatusOK, answer(leg{Shape: encodeShape(shape), Summary: aMinuteAndAHalf, Maneuvers: []maneuver{byCar}})
	}
}

// TestTheTraceIsAskedForTheRoutesOwnShape: the trace request carries the
// route's encoded shape verbatim — re-encoding would move it — walked edge by
// edge for a car, asking for the four attributes the check reads and nothing
// else, once per leg and after the route.
func TestTheTraceIsAskedForTheRoutesOwnShape(t *testing.T) {
	t.Parallel()

	shape := []planner.Point{traceA, traceB, traceC}
	engine, c := startEngine(t, routeOf(shape))

	if _, err := c.Route(t.Context(), traceA, traceC, nil); err != nil {
		t.Fatalf("Route: %v", err)
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	if len(engine.asked) != 1 || len(engine.traced) != 1 {
		t.Fatalf("the engine was asked %d routes and %d traces, want one of each", len(engine.asked), len(engine.traced))
	}

	got := engine.traced[0]
	if got.Method != http.MethodPost || got.Path != "/trace_attributes" {
		t.Errorf("the trace was %s %s, want POST /trace_attributes", got.Method, got.Path)
	}

	b, err := json.Marshal(got.Body)
	if err != nil {
		t.Fatal(err)
	}

	want, err := json.Marshal(map[string]any{
		"encoded_polyline": encodeShape(shape),
		"costing":          "auto",
		"shape_match":      "edge_walk",
		"filters": map[string]any{
			"attributes": []string{"edge.way_id", "edge.begin_shape_index", "edge.end_shape_index", "shape"},
			"action":     "include",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != string(want) {
		t.Errorf("the trace request is\n%s\nwant\n%s", b, want)
	}
}

// TestALegIsItsTracedEdges: the edges returned are the trace's, each its way
// id and its own stretch of the matched shape, adjacent stretches sharing their
// joining point, and an edge over a single point repeating it — so the
// planner joins them back into the matched shape exactly.
func TestALegIsItsTracedEdges(t *testing.T) {
	t.Parallel()

	matched := []planner.Point{traceA, traceB, traceC, traceD}

	c := startFake(t, &fakeEngine{respond: routeOf([]planner.Point{traceA, traceD}), trace: func(map[string]any) (int, string) {
		return http.StatusOK, traceAnswer(encodeShape(matched),
			traceEdge{WayID: 11, BeginShapeIndex: 0, EndShapeIndex: 2},
			traceEdge{WayID: 22, BeginShapeIndex: 2, EndShapeIndex: 2},
			traceEdge{WayID: 33, BeginShapeIndex: 2, EndShapeIndex: 3})
	}})

	route, err := c.Route(t.Context(), traceA, traceD, nil)
	edges := route.Edges
	if err != nil {
		t.Fatalf("Route: %v", err)
	}

	want := []planner.Edge{
		{WayID: 11, Shape: []planner.Point{traceA, traceB, traceC}},
		{WayID: 22, Shape: []planner.Point{traceC, traceC}},
		{WayID: 33, Shape: []planner.Point{traceC, traceD}},
	}

	if len(edges) != len(want) {
		t.Fatalf("Route returned %d edges, want %d: %+v", len(edges), len(want), edges)
	}

	for i := range want {
		if edges[i].WayID != want[i].WayID || !slices.Equal(edges[i].Shape, want[i].Shape) {
			t.Errorf("edge %d is %+v, want %+v", i, edges[i], want[i])
		}
	}

	if got := joined(edges); !slices.Equal(slices.Compact(got), matched) {
		t.Errorf("the edges join to %v, want the matched shape %v", got, matched)
	}
}

// TestALegWhoseWaysAreNotNamedIsRefused: a trace the engine cannot walk, or an
// answer that does not name every way of the leg end to end, refuses the leg
// with ErrWaysUnnamed and returns no route — never a leg with part of it
// unnamed, which the check could not read.
func TestALegWhoseWaysAreNotNamedIsRefused(t *testing.T) {
	t.Parallel()

	matched := encodeShape([]planner.Point{traceA, traceB, traceC, traceD})

	for name, respond := range map[string]func(map[string]any) (int, string){
		"a walk that fails": func(map[string]any) (int, string) {
			return http.StatusBadRequest, `{"error_code":443,"error":"Exact route match algorithm failed to find path","status_code":400}`
		},
		"not JSON": func(map[string]any) (int, string) { return http.StatusOK, `<html>` },
		"no shape": func(map[string]any) (int, string) {
			return http.StatusOK, `{"edges":[{"way_id":1,"begin_shape_index":0,"end_shape_index":3}]}`
		},
		"a broken shape": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer("_p~iF~ps|U_", traceEdge{WayID: 1, EndShapeIndex: 1})
		},
		"no edges": func(map[string]any) (int, string) { return http.StatusOK, traceAnswer(matched) },
		"an edge without a way id": func(map[string]any) (int, string) {
			return http.StatusOK, `{"shape":"` + matched + `","edges":[{"begin_shape_index":0,"end_shape_index":3}]}`
		},
		"an edge on way 0": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer(matched, traceEdge{WayID: 0, BeginShapeIndex: 0, EndShapeIndex: 3})
		},
		"an edge without its stretch": func(map[string]any) (int, string) {
			return http.StatusOK, `{"shape":"` + matched + `","edges":[{"way_id":1}]}`
		},
		"a first edge starting late": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer(matched, traceEdge{WayID: 1, BeginShapeIndex: 1, EndShapeIndex: 3})
		},
		"a gap between edges": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer(matched, traceEdge{WayID: 1, BeginShapeIndex: 0, EndShapeIndex: 1}, traceEdge{WayID: 2, BeginShapeIndex: 2, EndShapeIndex: 3})
		},
		"edges running backwards": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer(matched, traceEdge{WayID: 1, BeginShapeIndex: 0, EndShapeIndex: 2}, traceEdge{WayID: 2, BeginShapeIndex: 2, EndShapeIndex: 1})
		},
		"edges short of the end": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer(matched, traceEdge{WayID: 1, BeginShapeIndex: 0, EndShapeIndex: 2})
		},
		"an edge past the shape": func(map[string]any) (int, string) {
			return http.StatusOK, traceAnswer(matched, traceEdge{WayID: 1, BeginShapeIndex: 0, EndShapeIndex: 4})
		},
	} {
		c := startFake(t, &fakeEngine{respond: routeOf([]planner.Point{traceA, traceD}), trace: respond})

		got, err := c.Route(t.Context(), traceA, traceD, nil)
		if !errors.Is(err, ErrWaysUnnamed) || got.Edges != nil {
			t.Errorf("%s: Route gave %v, %v; want no route and an error wrapping ErrWaysUnnamed", name, got, err)
		}
	}
}
