package valhalla

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// These hold the adapter to the half of the port #7 adds: every car leg comes
// with the engine's estimate of its time, the leg's `summary.time` in seconds,
// and a leg with no usable estimate is refused rather than given a zero.

var (
	durationFrom = planner.Point{Latitude: 57.7036, Longitude: 11.9611}
	durationTo   = planner.Point{Latitude: 57.7052, Longitude: 11.9714}
)

// carLegTaking answers every route with one car leg between the two points,
// whose summary is the one given.
func carLegTaking(summary *legSummary) func(map[string]any) (int, string) {
	return func(map[string]any) (int, string) {
		return http.StatusOK, answer(leg{
			Shape:     encodeShape([]planner.Point{durationFrom, durationTo}),
			Summary:   summary,
			Maneuvers: []maneuver{byCar},
		})
	}
}

// TestALegCarriesTheEnginesTime: the route Route returns, and the leg a plan
// through it carries, take their duration from the leg's `summary.time`, read
// as seconds, fractions included.
func TestALegCarriesTheEnginesTime(t *testing.T) {
	t.Parallel()

	for _, secs := range []float64{0, 0.25, 437.5, 5400} {
		want := time.Duration(secs * float64(time.Second))

		_, c := startEngine(t, carLegTaking(&legSummary{Time: secs}))

		route, err := c.Route(t.Context(), durationFrom, durationTo, nil)
		if err != nil {
			t.Fatalf("Route with a summary time of %v s: %v", secs, err)
		}

		if route.Duration != want {
			t.Errorf("a leg whose summary time is %v s was returned taking %v, want %v", secs, route.Duration, want)
		}

		p, err := planner.New(c, heldCopy{}, clearAccess{}, clearAccess{})
		if err != nil {
			t.Fatalf("planner.New: %v", err)
		}

		journey, err := p.Plan(t.Context(), planner.Request{Origin: durationFrom, Destination: durationTo})
		if err != nil {
			t.Fatalf("Plan with a summary time of %v s: %v", secs, err)
		}

		if got := journey.Legs[0].Duration; got != want {
			t.Errorf("a planned leg whose summary time is %v s carries %v, want %v", secs, got, want)
		}
	}
}

// TestALegWithNoUsableTimeIsRefused: an answer whose leg holds no summary, a
// summary with no time, a negative time, or one past what a duration holds is
// an error and never a route — in particular never one taking no time.
func TestALegWithNoUsableTimeIsRefused(t *testing.T) {
	t.Parallel()

	shape := encodeShape([]planner.Point{durationFrom, durationTo})

	for name, body := range map[string]string{
		"no summary":             answer(leg{Shape: shape, Maneuvers: []maneuver{byCar}}),
		"a summary with no time": `{"trip":{"legs":[{"shape":` + quoted(shape) + `,"summary":{"length":1.2},"maneuvers":[{"travel_mode":"drive","travel_type":"car"}]}]}}`,
		"a null time":            `{"trip":{"legs":[{"shape":` + quoted(shape) + `,"summary":{"time":null},"maneuvers":[{"travel_mode":"drive","travel_type":"car"}]}]}}`,
		"a negative time":        answer(leg{Shape: shape, Summary: &legSummary{Time: -1}, Maneuvers: []maneuver{byCar}}),
		"a time past a duration": answer(leg{Shape: shape, Summary: &legSummary{Time: 1e300}, Maneuvers: []maneuver{byCar}}),
	} {
		_, c := startEngine(t, func(map[string]any) (int, string) { return http.StatusOK, body })

		got, err := c.Route(t.Context(), durationFrom, durationTo, nil)
		if err == nil || got.Edges != nil || got.Duration != 0 {
			t.Errorf("an answer with %s gave %v, %v; want an error and no route", name, got, err)
		}
	}
}

// quoted is s as a JSON string; an encoded shape holds backslashes.
func quoted(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(b)
}
