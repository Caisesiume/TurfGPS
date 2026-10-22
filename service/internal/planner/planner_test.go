package planner

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

// fakeRouting answers every route with a shape of its own choosing and records
// what it was asked and what it answered.
type fakeRouting struct {
	shape    func(from, to Point) []Point
	err      error
	asked    [][2]Point
	answered [][]Point
}

func (f *fakeRouting) Route(_ context.Context, from, to Point) ([]Point, error) {
	f.asked = append(f.asked, [2]Point{from, to})
	if f.err != nil {
		return nil, f.err
	}

	shape := f.shape(from, to)
	f.answered = append(f.answered, shape)

	return shape, nil
}

// randomShapes is a fake whose every route is a random shape, which never
// begins or ends at the locations asked for.
func randomShapes(r *rand.Rand) *fakeRouting {
	return &fakeRouting{shape: func(_, _ Point) []Point {
		shape := make([]Point, 1+r.IntN(16))
		for i := range shape {
			shape[i] = randomPoint(r)
		}

		return shape
	}}
}

// randomPoint returns a coordinate, with the bounds of each axis drawn often
// enough that a check which mishandles them is found.
func randomPoint(r *rand.Rand) Point {
	axis := func(limit float64) float64 {
		switch r.IntN(8) {
		case 0:
			return -limit
		case 1:
			return limit
		default:
			return (r.Float64()*2 - 1) * limit
		}
	}

	return Point{Latitude: axis(90), Longitude: axis(180)}
}

// randomRequest returns a request naming between minIntermediates and
// maxIntermediates intermediate destinations. One in eight intermediates
// repeats a location already entered, so a planner that drops or merges a
// repeated location is found as well as one that reorders.
func randomRequest(r *rand.Rand, minIntermediates, maxIntermediates int) Request {
	req := Request{Origin: randomPoint(r), Destination: randomPoint(r)}

	for range minIntermediates + r.IntN(maxIntermediates-minIntermediates+1) {
		next := randomPoint(r)
		if r.IntN(8) == 0 {
			entered := append([]Point{req.Origin}, req.Intermediates...)
			next = entered[r.IntN(len(entered))]
		}

		req.Intermediates = append(req.Intermediates, next)
	}

	return req
}

// requiredLocations returns the locations req names, in the order entered.
func requiredLocations(req Request) []Point {
	return slices.Concat([]Point{req.Origin}, req.Intermediates, []Point{req.Destination})
}

// TestEveryJourneyBeginsAtTheOriginAndEndsAtTheDestination is `FR-001`'s
// verification, both of its cases, as a property over requests: whatever origin
// and destination a request names, whether or not it also names intermediate
// destinations (#1 AC1, #2 AC4), and whatever geometry the routing provider
// answers with, the journey planned begins at the origin and ends at the
// destination.
//
// The provider's shapes are random and never begin or end at the locations
// asked for, so a planner that read the journey's ends back off the route — as
// a real engine's snapped geometry would make it do — fails here.
//
// The seed is fixed so that a failure is reproduced by rerunning the test, and
// is printed with the failing case.
func TestEveryJourneyBeginsAtTheOriginAndEndsAtTheDestination(t *testing.T) {
	t.Parallel()

	const seed, cases = 1, 2000

	r := rand.New(rand.NewPCG(seed, seed))

	var withIntermediates, without int

	for c := range cases {
		routing := randomShapes(r)

		p, err := New(routing, copyHeld)
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := randomRequest(r, 0, 5)
		if len(req.Intermediates) == 0 {
			without++
		} else {
			withIntermediates++
		}

		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("seed %d case %d: Plan(%+v) refused a request naming only coordinates: %v", seed, c, req, err)
		}

		if len(journey.Locations) < 2 {
			t.Fatalf("seed %d case %d: the journey has %d locations, want at least an origin and a destination", seed, c, len(journey.Locations))
		}

		if got := journey.Locations[0]; got != req.Origin {
			t.Fatalf("seed %d case %d: the journey begins at %+v, not at the entered origin %+v", seed, c, got, req.Origin)
		}

		if got := journey.Locations[len(journey.Locations)-1]; got != req.Destination {
			t.Fatalf("seed %d case %d: the journey ends at %+v, not at the entered destination %+v", seed, c, got, req.Destination)
		}

		// The ends are not merely echoed: every leg was routed, between the
		// locations it joins.
		if len(journey.Legs) != len(journey.Locations)-1 {
			t.Fatalf("seed %d case %d: %d legs for %d locations, want one per consecutive pair", seed, c, len(journey.Legs), len(journey.Locations))
		}

		for i := range journey.Legs {
			want := [2]Point{journey.Locations[i], journey.Locations[i+1]}
			if i >= len(routing.asked) || routing.asked[i] != want {
				t.Fatalf("seed %d case %d: leg %d was not routed from %+v to %+v; the provider was asked %+v", seed, c, i, want[0], want[1], routing.asked)
			}
		}
	}

	// Both of `FR-001`'s cases were drawn, so neither passes vacuously.
	if withIntermediates == 0 || without == 0 {
		t.Fatalf("seed %d drew %d requests with intermediate destinations and %d without, want both", seed, withIntermediates, without)
	}
}

// TestEveryJourneyPassesThroughItsIntermediateDestinationsInTheEnteredOrder is
// #2 AC1 and AC2 (`FR-002`, `FR-003`), as a property over requests: whatever
// intermediate destinations a request names, in whatever order, and whatever
// geometry the routing provider answers with, the request is planned rather
// than refused, and the journey's locations are exactly the required locations
// in the order entered, each leg routed from one to the next.
func TestEveryJourneyPassesThroughItsIntermediateDestinationsInTheEnteredOrder(t *testing.T) {
	t.Parallel()

	const seed, cases = 2, 2000

	r := rand.New(rand.NewPCG(seed, seed))

	for c := range cases {
		routing := randomShapes(r)

		p, err := New(routing, copyHeld)
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := randomRequest(r, 1, 6)
		entered := slices.Clone(req.Intermediates)

		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("seed %d case %d: Plan(%+v) refused a request naming intermediate destinations: %v", seed, c, req, err)
		}

		if !slices.Equal(req.Intermediates, entered) {
			t.Fatalf("seed %d case %d: planning changed the request's intermediate destinations from %+v to %+v", seed, c, entered, req.Intermediates)
		}

		want := requiredLocations(req)
		if !slices.Equal(journey.Locations, want) {
			t.Fatalf("seed %d case %d: the journey visits %+v, want the required locations in the entered order %+v", seed, c, journey.Locations, want)
		}

		// Passing through a location is being routed to it and on from it.
		if len(routing.asked) != len(want)-1 || len(journey.Legs) != len(want)-1 {
			t.Fatalf("seed %d case %d: %d legs routed and %d returned for %d required locations, want one per consecutive pair", seed, c, len(routing.asked), len(journey.Legs), len(want))
		}

		for i := range journey.Legs {
			if pair := [2]Point{want[i], want[i+1]}; routing.asked[i] != pair {
				t.Fatalf("seed %d case %d: leg %d was routed from %+v to %+v, want from %+v to %+v", seed, c, i, routing.asked[i][0], routing.asked[i][1], pair[0], pair[1])
			}

			if !slices.Equal(journey.Legs[i].Shape, routing.answered[i]) {
				t.Fatalf("seed %d case %d: leg %d has the shape %+v, want the route the provider answered for it, %+v", seed, c, i, journey.Legs[i].Shape, routing.answered[i])
			}
		}
	}
}

// TestAFasterOrderDoesNotOverrideTheEnteredOrder is #2 AC3 (`FR-003`): a journey
// whose mandatory waypoints would be reached in less time in another order is
// still planned in the order entered.
//
// The fake drives in straight lines at a constant speed, so a leg's time is its
// length. The waypoints are entered in an order that doubles back — Stockholm to
// Malmö, back north to Jönköping, then Göteborg — and the test first proves that
// a faster order exists, so the property is not passed vacuously.
func TestAFasterOrderDoesNotOverrideTheEnteredOrder(t *testing.T) {
	t.Parallel()

	req := Request{
		Origin: Point{Latitude: 59.3293, Longitude: 18.0686}, // Stockholm
		Intermediates: []Point{
			{Latitude: 55.6050, Longitude: 13.0038}, // Malmö
			{Latitude: 57.7826, Longitude: 14.1618}, // Jönköping
		},
		Destination: Point{Latitude: 57.7089, Longitude: 11.9746}, // Göteborg
	}

	legTime := func(from, to Point) float64 {
		return math.Hypot(to.Latitude-from.Latitude, to.Longitude-from.Longitude)
	}

	journeyTime := func(locations []Point) float64 {
		var total float64
		for i := range len(locations) - 1 {
			total += legTime(locations[i], locations[i+1])
		}

		return total
	}

	entered := requiredLocations(req)
	other := requiredLocations(Request{Origin: req.Origin, Intermediates: []Point{req.Intermediates[1], req.Intermediates[0]}, Destination: req.Destination})

	if journeyTime(other) >= journeyTime(entered) {
		t.Fatalf("the fixture is wrong: the other order takes %.3f and the entered order %.3f, want the other order faster", journeyTime(other), journeyTime(entered))
	}

	routing := &fakeRouting{shape: func(from, to Point) []Point { return []Point{from, to} }}

	p, err := New(routing, copyHeld)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	journey, err := p.Plan(t.Context(), req)
	if err != nil {
		t.Fatalf("Plan refused the journey: %v", err)
	}

	if !slices.Equal(journey.Locations, entered) {
		t.Fatalf("the journey visits %+v (time %.3f), want the entered order %+v (time %.3f) even though another order is faster", journey.Locations, journeyTime(journey.Locations), entered, journeyTime(entered))
	}

	if len(routing.asked) != len(entered)-1 {
		t.Fatalf("%d legs were routed, want %d", len(routing.asked), len(entered)-1)
	}

	for i, pair := range routing.asked {
		if want := [2]Point{entered[i], entered[i+1]}; pair != want {
			t.Fatalf("leg %d was routed from %+v to %+v, want from %+v to %+v", i, pair[0], pair[1], want[0], want[1])
		}
	}
}

func TestAPlanThatCannotBeRoutedIsRefused(t *testing.T) {
	t.Parallel()

	stockholm := Point{Latitude: 59.3293, Longitude: 18.0686}
	goteborg := Point{Latitude: 57.7089, Longitude: 11.9746}
	errEngine := errors.New("engine unavailable")

	for _, tc := range []struct {
		name     string
		req      Request
		routing  *fakeRouting
		wantErr  error
		wantNoIO bool
	}{
		{
			name:     "an origin that is not a coordinate",
			req:      Request{Origin: Point{Latitude: 90.5, Longitude: 0}, Destination: goteborg},
			wantErr:  ErrNotACoordinate,
			wantNoIO: true,
		},
		{
			name:     "a destination that is not a coordinate",
			req:      Request{Origin: stockholm, Destination: Point{Latitude: 0, Longitude: math.NaN()}},
			wantErr:  ErrNotACoordinate,
			wantNoIO: true,
		},
		{
			name:    "a provider that fails",
			req:     Request{Origin: stockholm, Destination: goteborg},
			routing: &fakeRouting{err: errEngine},
			wantErr: errEngine,
		},
		{
			name:    "a provider that returns no geometry",
			req:     Request{Origin: stockholm, Destination: goteborg},
			routing: &fakeRouting{shape: func(_, _ Point) []Point { return nil }},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := tc.routing
			if routing == nil {
				routing = &fakeRouting{shape: func(from, to Point) []Point { return []Point{from, to} }}
			}

			p, err := New(routing, copyHeld)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			journey, err := p.Plan(t.Context(), tc.req)
			if err == nil {
				t.Fatalf("Plan returned %+v and no error, want a refusal", journey)
			}

			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("Plan error = %v, want one wrapping %v", err, tc.wantErr)
			}

			if tc.wantNoIO && len(routing.asked) != 0 {
				t.Errorf("the provider was asked %+v, want no routing for a request that is refused", routing.asked)
			}

			if journey.Locations != nil || journey.Legs != nil {
				t.Errorf("a refused plan returned the journey %+v, want none", journey)
			}
		})
	}
}

func TestNewRefusesNoProvider(t *testing.T) {
	t.Parallel()

	if _, err := New(nil, copyHeld); err == nil {
		t.Error("New(nil, copyHeld) returned no error, want a refusal")
	}
}
