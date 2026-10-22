package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
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

// zoneCopy is a planner.ZoneCopy reporting a copy held or not, or failing with
// err.
type zoneCopy struct {
	held bool
	err  error
}

func (z zoneCopy) Held(context.Context) (bool, error) { return z.held, z.err }

// withFakePlanner plans through routing from a zone copy that is held, which is
// every copy but none (#33).
func withFakePlanner(t *testing.T, routing planner.RoutingProvider) Option {
	t.Helper()

	return withPlannerOver(t, routing, zoneCopy{held: true})
}

func withPlannerOver(t *testing.T, routing planner.RoutingProvider, zones planner.ZoneCopy) Option {
	t.Helper()

	p, err := planner.New(routing, zones)
	if err != nil {
		t.Fatalf("planner.New: %v", err)
	}

	return WithPlanner(p)
}

// TestNoJourneyIsPlannedWhereNoZoneCopyIsHeld is #33 AC4 (`FR-033`) through
// the request surface: with no local copy of the zone set, `POST /journeys`
// returns no journey and says why. An unreadable store is refused too, and
// says that instead, so it is never reported as an empty one.
func TestNoJourneyIsPlannedWhereNoZoneCopyIsHeld(t *testing.T) {
	t.Parallel()

	const valid = `{"origin":{"latitude":59.3293,"longitude":18.0686},"destination":{"latitude":57.7089,"longitude":11.9746}}`

	for _, tc := range []struct {
		name  string
		zones zoneCopy
		want  string
	}{
		{name: "no copy held", zones: zoneCopy{held: false}, want: "holds no local copy of the zone set"},
		{name: "a store that cannot be read", zones: zoneCopy{err: errors.New("store unreachable")}, want: "could not be read"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := postJourney(t, valid, withPlannerOver(t, straightRouting{}, tc.zones))

			if rec.Code != http.StatusServiceUnavailable {
				t.Errorf("status = %d, want %d; body %q", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
			}

			if body := rec.Body.String(); strings.Contains(body, `"locations"`) || !strings.Contains(body, tc.want) {
				t.Errorf("body %q, want no journey and a message saying it %s", body, tc.want)
			}
		})
	}
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

// TestAPlanningRequestReturnsItsIntermediateDestinationsInTheEnteredOrder is #2
// AC1, AC2 and AC4 through the request surface: the journey `POST /journeys`
// returns passes through the intermediate destinations the request names, in
// the order it names them — here an order that doubles back, Stockholm to Malmö
// and back north to Jönköping — and still begins at the origin and ends at the
// destination. A request naming none, by omitting the field, sending null or
// sending an empty list, is the two-location journey of #1.
func TestAPlanningRequestReturnsItsIntermediateDestinationsInTheEnteredOrder(t *testing.T) {
	t.Parallel()

	const (
		stockholm = `{"latitude":59.3293,"longitude":18.0686}`
		malmo     = `{"latitude":55.605,"longitude":13.0038}`
		jonkoping = `{"latitude":57.7826,"longitude":14.1618}`
		goteborg  = `{"latitude":57.7089,"longitude":11.9746}`
	)

	origin, destination := wirePoint{59.3293, 18.0686}, wirePoint{57.7089, 11.9746}

	for _, tc := range []struct {
		name          string
		intermediates string
		want          []wirePoint
	}{
		{name: "two intermediate destinations", intermediates: `,"intermediates":[` + malmo + `,` + jonkoping + `]`, want: []wirePoint{origin, {55.605, 13.0038}, {57.7826, 14.1618}, destination}},
		{name: "the same two entered the other way round", intermediates: `,"intermediates":[` + jonkoping + `,` + malmo + `]`, want: []wirePoint{origin, {57.7826, 14.1618}, {55.605, 13.0038}, destination}},
		{name: "no intermediates field", want: []wirePoint{origin, destination}},
		{name: "null intermediates", intermediates: `,"intermediates":null`, want: []wirePoint{origin, destination}},
		{name: "an empty list of intermediates", intermediates: `,"intermediates":[]`, want: []wirePoint{origin, destination}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := postJourney(t,
				`{"origin":`+stockholm+tc.intermediates+`,"destination":`+goteborg+`}`,
				withFakePlanner(t, straightRouting{}))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got journeyResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("the body is not a journey: %v; body %q", err, rec.Body.String())
			}

			if !slices.Equal(got.Locations, tc.want) {
				t.Errorf("locations = %+v, want the required locations in the entered order %+v", got.Locations, tc.want)
			}

			if len(got.Legs) != len(tc.want)-1 {
				t.Errorf("%d legs, want one per consecutive pair of the %d required locations", len(got.Legs), len(tc.want))
			}
		})
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
		{name: "an intermediate destination with no latitude", body: `{"origin":{"latitude":59.3,"longitude":18.0},"intermediates":[{"longitude":13.0}],"destination":{"latitude":57.7,"longitude":11.9}}`, want: http.StatusBadRequest},
		{name: "an intermediate destination with no longitude", body: `{"origin":{"latitude":59.3,"longitude":18.0},"intermediates":[{"latitude":55.6,"longitude":13.0},{"latitude":57.8}],"destination":{"latitude":57.7,"longitude":11.9}}`, want: http.StatusBadRequest},
		{name: "a null intermediate destination", body: `{"origin":{"latitude":59.3,"longitude":18.0},"intermediates":[null],"destination":{"latitude":57.7,"longitude":11.9}}`, want: http.StatusBadRequest},
		{name: "an intermediate destination that is not a coordinate", body: `{"origin":{"latitude":59.3,"longitude":18.0},"intermediates":[{"latitude":55.6,"longitude":181}],"destination":{"latitude":57.7,"longitude":11.9}}`, want: http.StatusBadRequest},
		{name: "intermediates that are not a list", body: `{"origin":{"latitude":59.3,"longitude":18.0},"intermediates":{"latitude":55.6,"longitude":13.0},"destination":{"latitude":57.7,"longitude":11.9}}`, want: http.StatusBadRequest},
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
