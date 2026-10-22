package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// straightRouting routes every leg as a line through a midpoint of its own, so
// the route's vertices are never the locations a journey is planned between.
type straightRouting struct{}

func (straightRouting) Route(_ context.Context, from, to planner.Point) ([]planner.Point, error) {
	mid := planner.Point{Latitude: (from.Latitude + to.Latitude) / 2, Longitude: (from.Longitude + to.Longitude) / 2}

	return []planner.Point{mid}, nil
}

type failingRouting struct{}

func (failingRouting) Route(context.Context, planner.Point, planner.Point) ([]planner.Point, error) {
	return nil, errors.New("engine unavailable")
}

func postJourney(t *testing.T, body string, opts ...Option) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/journeys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	NewHandler(nil, opts...).ServeHTTP(rec, req)

	return rec
}

func withFakePlanner(t *testing.T, routing planner.RoutingProvider) Option {
	t.Helper()

	p, err := planner.New(routing)
	if err != nil {
		t.Fatalf("planner.New: %v", err)
	}

	return WithPlanner(p)
}

// TestAPlanningRequestReturnsAJourneyFromTheOriginToTheDestination is `FR-001`
// AC1 through the request surface: the journey `POST /journeys` returns begins
// at the origin the request names and ends at its destination.
func TestAPlanningRequestReturnsAJourneyFromTheOriginToTheDestination(t *testing.T) {
	t.Parallel()

	rec := postJourney(t,
		`{"origin":{"latitude":59.3293,"longitude":18.0686},"destination":{"latitude":57.7089,"longitude":11.9746}}`,
		withFakePlanner(t, straightRouting{}))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusOK, rec.Body.String())
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var got journeyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("the body is not a journey: %v; body %q", err, rec.Body.String())
	}

	origin, destination := wirePoint{59.3293, 18.0686}, wirePoint{57.7089, 11.9746}

	if len(got.Locations) < 2 || got.Locations[0] != origin || got.Locations[len(got.Locations)-1] != destination {
		t.Errorf("locations = %+v, want a journey beginning at %+v and ending at %+v", got.Locations, origin, destination)
	}

	if len(got.Legs) != len(got.Locations)-1 || len(got.Legs) == 0 || len(got.Legs[0].Shape) == 0 {
		t.Errorf("legs = %+v, want one routed leg per consecutive pair of locations", got.Legs)
	}
}

func TestAPlanningRequestThatCannotBePlannedIsRefused(t *testing.T) {
	t.Parallel()

	const valid = `{"origin":{"latitude":59.3293,"longitude":18.0686},"destination":{"latitude":57.7089,"longitude":11.9746}}`

	for _, tc := range []struct {
		name   string
		body   string
		noPlan bool
		fail   bool
		want   int
	}{
		{name: "no routing provider is configured", body: valid, noPlan: true, want: http.StatusServiceUnavailable},
		{name: "a provider that fails", body: valid, fail: true, want: http.StatusBadGateway},
		{name: "a body that is not JSON", body: `origin=here`, want: http.StatusBadRequest},
		{name: "an absent destination", body: `{"origin":{"latitude":59.3293,"longitude":18.0686}}`, want: http.StatusBadRequest},
		{name: "an absent origin", body: `{"destination":{"latitude":57.7089,"longitude":11.9746}}`, want: http.StatusBadRequest},
		{name: "a field this shape cannot express", body: `{"origin":{"latitude":59.3,"longitude":18.0},"destination":{"latitude":57.7,"longitude":11.9},"via":[]}`, want: http.StatusBadRequest},
		{name: "two requests in one body", body: valid + valid, want: http.StatusBadRequest},
		{name: "a location that is not a coordinate", body: `{"origin":{"latitude":91,"longitude":18.0},"destination":{"latitude":57.7,"longitude":11.9}}`, want: http.StatusBadRequest},
		// An absent or null axis would decode as 0 and plan from a coordinate
		// nobody entered.
		{name: "an origin with no latitude", body: `{"origin":{"longitude":18.0686},"destination":{"latitude":57.7089,"longitude":11.9746}}`, want: http.StatusBadRequest},
		{name: "an origin with no longitude", body: `{"origin":{"latitude":59.3293},"destination":{"latitude":57.7089,"longitude":11.9746}}`, want: http.StatusBadRequest},
		{name: "a destination with no latitude", body: `{"origin":{"latitude":59.3293,"longitude":18.0686},"destination":{"longitude":11.9746}}`, want: http.StatusBadRequest},
		{name: "a destination with no longitude", body: `{"origin":{"latitude":59.3293,"longitude":18.0686},"destination":{"latitude":57.7089}}`, want: http.StatusBadRequest},
		{name: "a null latitude", body: `{"origin":{"latitude":null,"longitude":18.0686},"destination":{"latitude":57.7089,"longitude":11.9746}}`, want: http.StatusBadRequest},
		{name: "an empty location", body: `{"origin":{},"destination":{"latitude":57.7089,"longitude":11.9746}}`, want: http.StatusBadRequest},
		{name: "a body over the bound", body: `{"origin":` + strings.Repeat(" ", maxJourneyRequestBytes) + `null}`, want: http.StatusRequestEntityTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts []Option

			switch {
			case tc.fail:
				opts = append(opts, withFakePlanner(t, failingRouting{}))
			case !tc.noPlan:
				opts = append(opts, withFakePlanner(t, straightRouting{}))
			}

			rec := postJourney(t, tc.body, opts...)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d; body %q", rec.Code, tc.want, rec.Body.String())
			}

			if strings.Contains(rec.Body.String(), `"locations"`) {
				t.Errorf("a refused request returned a journey: %q", rec.Body.String())
			}
		})
	}
}
