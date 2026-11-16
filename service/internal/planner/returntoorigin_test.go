package planner

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

// TestAReturnToTheOriginWithNowhereBetweenIsRefused is #5 AC1 and AC2
// (`FR-012`): a request naming one location as both its origin and its
// destination plans no journey when it names no intermediate destination, and
// none when every intermediate destination it names is that location. AC2 is
// the case a check for merely naming an intermediate admits — A → A → A.
//
// The refusal reads only what the request names, so nothing is routed and the
// zone copy is not read: the copy here is one that is held, so a refusal from
// the gate behind it cannot stand in for this one.
func TestAReturnToTheOriginWithNowhereBetweenIsRefused(t *testing.T) {
	t.Parallel()

	home := Point{Latitude: 59.3293, Longitude: 18.0686}
	equator := Point{Latitude: 0, Longitude: 0}

	for _, tc := range []struct {
		name string
		req  Request
	}{
		// AC1.
		{name: "no intermediate destination", req: Request{Origin: home, Destination: home}},
		{name: "an empty list of intermediate destinations", req: Request{Origin: home, Intermediates: []Point{}, Destination: home}},
		// AC2.
		{name: "one intermediate destination, the origin again", req: Request{Origin: home, Intermediates: []Point{home}, Destination: home}},
		{name: "three intermediate destinations, each the origin again", req: Request{Origin: home, Intermediates: []Point{home, home, home}, Destination: home}},
		// Negative zero is the same coordinate as zero, so the same location.
		{name: "the origin again written with a negative zero", req: Request{Origin: equator, Intermediates: []Point{{Latitude: math.Copysign(0, -1), Longitude: 0}}, Destination: equator}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := straightLegs()

			p, err := New(routing, copyHeld, clearAccess{}, clearAccess{}, endsAnywhere)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			journey, err := p.Plan(t.Context(), tc.req)
			if !errors.Is(err, ErrReturnToOrigin) {
				t.Fatalf("Plan(%+v) = %+v, %v; want no journey and an error wrapping %v", tc.req, journey, err, ErrReturnToOrigin)
			}

			if journey.Locations != nil || journey.Legs != nil {
				t.Errorf("a refused return to the origin returned the journey %+v, want none", journey)
			}

			if len(routing.asked) != 0 {
				t.Errorf("a refused return to the origin routed %+v, want nothing routed", routing.asked)
			}
		})
	}
}

// TestAReturnToTheOriginThroughADistinctIntermediateIsNotRefusedForIt is #5
// AC3 (`FR-012`): a request naming one location as both its origin and its
// destination and at least one intermediate destination distinct from it is
// not refused for having named its origin as its destination. That it is
// planned is `FR-002`'s, and is not asserted here; that it went on to be
// routed is, so the test cannot pass on a request refused for another reason
// before the provider was asked.
//
// Distinct is exact identity, never proximity, so an intermediate a ten
// millionth of a degree from the origin is distinct.
func TestAReturnToTheOriginThroughADistinctIntermediateIsNotRefusedForIt(t *testing.T) {
	t.Parallel()

	home := Point{Latitude: 59.3293, Longitude: 18.0686}
	shop := Point{Latitude: 59.8586, Longitude: 17.6389}
	nextDoor := Point{Latitude: home.Latitude + 1e-7, Longitude: home.Longitude}

	for _, tc := range []struct {
		name string
		req  Request
	}{
		{name: "one distinct intermediate destination", req: Request{Origin: home, Intermediates: []Point{shop}, Destination: home}},
		{name: "a distinct one after the origin again", req: Request{Origin: home, Intermediates: []Point{home, shop}, Destination: home}},
		{name: "a distinct one before the origin again", req: Request{Origin: home, Intermediates: []Point{shop, home}, Destination: home}},
		{name: "a distinct one between the origin twice", req: Request{Origin: home, Intermediates: []Point{home, shop, home}, Destination: home}},
		{name: "one a ten millionth of a degree away", req: Request{Origin: home, Intermediates: []Point{nextDoor}, Destination: home}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := straightLegs()

			p, err := New(routing, copyHeld, clearAccess{}, clearAccess{}, endsAnywhere)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			_, err = p.Plan(t.Context(), tc.req)
			if errors.Is(err, ErrReturnToOrigin) {
				t.Fatalf("Plan(%+v) refused a return to the origin through a distinct intermediate destination for naming its origin as its destination: %v", tc.req, err)
			}

			if len(routing.asked) == 0 {
				t.Errorf("Plan(%+v) routed nothing and returned %v, want the request to go on to be routed", tc.req, err)
			}
		})
	}
}

// TestOnlyAReturnWithNowhereBetweenIsRefused holds `FR-012`'s three criteria
// together as a property: whatever a request names, it is refused as a return
// to the origin exactly where its destination is its origin and no
// intermediate destination differs from that location — so the refusal never
// reaches a journey with a distinct destination, and never misses A → … → A.
func TestOnlyAReturnWithNowhereBetweenIsRefused(t *testing.T) {
	t.Parallel()

	const seed, cases = 5, 2000

	r := rand.New(rand.NewPCG(seed, seed))

	var refused, admitted int

	for c := range cases {
		req := randomRequest(r, 0, 4)

		// Make most requests returns, and most of their intermediates the
		// origin again, so both outcomes are drawn often.
		if r.IntN(4) != 0 {
			req.Destination = req.Origin
		}

		for i := range req.Intermediates {
			if r.IntN(3) != 0 {
				req.Intermediates[i] = req.Origin
			}
		}

		want := req.Destination == req.Origin
		for _, via := range req.Intermediates {
			want = want && via == req.Origin
		}

		p, err := New(straightLegs(), copyHeld, clearAccess{}, clearAccess{}, endsAnywhere)
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		_, err = p.Plan(t.Context(), req)

		if got := errors.Is(err, ErrReturnToOrigin); got != want {
			t.Fatalf("seed %d case %d: Plan(%+v) refused as a return to the origin = %v, want %v; error %v", seed, c, req, got, want, err)
		}

		if want {
			refused++
		} else {
			admitted++
		}
	}

	if refused == 0 || admitted == 0 {
		t.Fatalf("seed %d drew %d requests to refuse and %d to admit, want both", seed, refused, admitted)
	}
}
