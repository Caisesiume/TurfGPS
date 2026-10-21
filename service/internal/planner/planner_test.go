package planner

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

// fakeRouting answers every route with a shape of its own choosing and records
// what it was asked.
type fakeRouting struct {
	shape func(from, to Point) []Point
	err   error
	asked [][2]Point
}

func (f *fakeRouting) Route(_ context.Context, from, to Point) ([]Point, error) {
	f.asked = append(f.asked, [2]Point{from, to})
	if f.err != nil {
		return nil, f.err
	}

	return f.shape(from, to), nil
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

// TestEveryJourneyBeginsAtTheOriginAndEndsAtTheDestination is `FR-001` AC1, as
// a property over requests: whatever origin and destination a request names,
// and whatever geometry the routing provider answers with, the journey planned
// begins at the origin and ends at the destination.
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

	for c := range cases {
		routing := &fakeRouting{shape: func(_, _ Point) []Point {
			shape := make([]Point, 1+r.IntN(16))
			for i := range shape {
				shape[i] = randomPoint(r)
			}

			return shape
		}}

		p, err := New(routing)
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := Request{Origin: randomPoint(r), Destination: randomPoint(r)}

		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("seed %d case %d: Plan(%+v) refused a request naming two coordinates: %v", seed, c, req, err)
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

			p, err := New(routing)
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

	if _, err := New(nil); err == nil {
		t.Error("New(nil) returned no error, want a refusal")
	}
}
