package planner

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// The end tolerance at the planner (#250). The criterion is held against the
// real engine and the committed extract in `internal/valhalla`'s
// `valhalla_integration_test.go`; these hold the reading itself, on every
// answer the planner reads.

// endsAnywhere is the end tolerance of the tests that are not about it: the
// most endDistance can return, half a great circle of radius
// wgs84PolarRadiusOfCurvatureMetres, so no answer is refused for its ends.
var endsAnywhere = EndsWithin(wgs84PolarRadiusOfCurvatureMetres * math.Pi)

// placed answers every request with a straight leg along ways 1 and 2 from
// start(from) to end(to): where the engine placed the ends.
type placed struct {
	start, end func(Point) Point
}

func (p placed) Route(_ context.Context, from, to Point, _ []Point) (Route, error) {
	a, b := p.start(from), p.end(to)
	mid := Point{Latitude: (a.Latitude + b.Latitude) / 2, Longitude: (a.Longitude + b.Longitude) / 2}

	return Route{Edges: []Edge{{WayID: 1, Shape: []Point{a, mid}}, {WayID: 2, Shape: []Point{mid, b}}}}, nil
}

func at(p Point) Point { return p }

// northBy returns where p moved north by metres of endDistance lies.
func northBy(metres float64) func(Point) Point {
	return func(p Point) Point {
		return Point{Latitude: p.Latitude + metres/wgs84PolarRadiusOfCurvatureMetres*180/math.Pi, Longitude: p.Longitude}
	}
}

// TestNoLegIsServedWithoutAnEndTolerance (#250, the Owner's ruling: until the
// tolerance is set, nothing plans): a planner built without EndsWithin refuses
// a leg the engine placed exactly at its locations, through Plan and PlanLeg.
func TestNoLegIsServedWithoutAnEndTolerance(t *testing.T) {
	t.Parallel()

	p, err := New(placed{start: at, end: at}, copyHeld, clearAccess{}, clearAccess{})
	if err != nil {
		t.Fatal(err)
	}

	journey, err := p.Plan(t.Context(), Request{Origin: home, Destination: shops})
	if journey.Legs != nil || !errors.Is(err, ErrNoEndTolerance) {
		t.Errorf("Plan without an end tolerance gave %d leg(s) and %v; want no journey and an error wrapping ErrNoEndTolerance", len(journey.Legs), err)
	}

	leg, err := p.PlanLeg(t.Context(), LegEnd{At: home, EnteredByUser: true}, LegEnd{At: shops})
	if leg.Shape != nil || !errors.Is(err, ErrNoEndTolerance) {
		t.Errorf("PlanLeg without an end tolerance gave a leg of %d points and %v; want no leg and an error wrapping ErrNoEndTolerance", len(leg.Shape), err)
	}
}

// TestALegWhoseEndDriftsBeyondTheToleranceIsRefused (#250 AC2): an answer
// whose first vertex lies beyond the tolerance from the origin, or whose last
// lies beyond it from the destination, is refused and never served; one with
// both ends within it, or exactly at it, is served.
func TestALegWhoseEndDriftsBeyondTheToleranceIsRefused(t *testing.T) {
	t.Parallel()

	const tolerance = 50.0

	beyond := math.Nextafter(tolerance, math.Inf(1)) * (1 + 1e-9)

	for _, tc := range []struct {
		name       string
		start, end func(Point) Point
		served     bool
	}{
		{"both ends at their locations", at, at, true},
		{"both ends within the tolerance", northBy(tolerance / 2), northBy(tolerance / 2), true},
		{"the first vertex beyond it", northBy(beyond), at, false},
		{"the last vertex beyond it", at, northBy(beyond), false},
		{"the first vertex as far as #246 measured", northBy(32053), at, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := New(placed{start: tc.start, end: tc.end}, copyHeld, clearAccess{}, clearAccess{}, EndsWithin(tolerance))
			if err != nil {
				t.Fatal(err)
			}

			journey, err := p.Plan(t.Context(), Request{Origin: home, Destination: shops})

			switch {
			case tc.served && err != nil:
				t.Fatalf("Plan refused a leg with its ends within %v m: %v", tolerance, err)
			case !tc.served && (journey.Legs != nil || !errors.Is(err, ErrEndBeyondTolerance)):
				t.Fatalf("Plan gave %d leg(s) and %v; want the leg refused with an error wrapping ErrEndBeyondTolerance", len(journey.Legs), err)
			}
		})
	}

	// Exactly at the tolerance is within it, and the next value down is not.
	end := northBy(30)(home)
	d := endDistance(end, home)

	for tolerance, served := range map[float64]bool{d: true, math.Nextafter(d, 0): false} {
		p, err := New(placed{start: func(Point) Point { return end }, end: at}, copyHeld, clearAccess{}, clearAccess{}, EndsWithin(tolerance))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := p.Plan(t.Context(), Request{Origin: home, Destination: shops}); (err == nil) != served {
			t.Errorf("a first vertex %v m from the origin, under a tolerance of %v m: Plan gave %v, want served=%v", d, tolerance, err, served)
		}
	}
}

// TestARefusalForAnEndSaysWhichAndHowFar: the refusal names the end, its
// distance from the location and the tolerance, so a deployment measuring
// the tolerance can read the drift off it.
func TestARefusalForAnEndSaysWhichAndHowFar(t *testing.T) {
	t.Parallel()

	p, err := New(placed{start: at, end: northBy(120)}, copyHeld, clearAccess{}, clearAccess{}, EndsWithin(50))
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Plan(t.Context(), Request{Origin: home, Destination: shops})
	if err == nil {
		t.Fatal("Plan served a leg whose last vertex lies 120 m from the destination under a 50 m tolerance")
	}

	for _, want := range []string{"last vertex", "120.0 m", "50 m"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not say %q", err, want)
		}
	}
}

// shifted answers as its provider does, every point of its answers from the
// from'th request on moved north by metres: an engine whose placement of the
// leg's ends moves once it is asked to keep off places. It says what it keeps
// off as its provider does.
type shifted struct {
	RoutingProvider
	from   int
	metres float64
	n      int
}

func (s *shifted) Route(ctx context.Context, from, to Point, avoid []Point) (Route, error) {
	s.n++

	r, err := s.RoutingProvider.Route(ctx, from, to, avoid)
	if err != nil || s.n < s.from {
		return r, err
	}

	move := northBy(s.metres)

	for i, e := range r.Edges {
		shape := make([]Point, len(e.Shape))
		for j, pt := range e.Shape {
			shape[j] = move(pt)
		}

		r.Edges[i].Shape = shape
	}

	return r, nil
}

func (s *shifted) Excluded(ctx context.Context, avoid []Point) ([][]int64, error) {
	return s.RoutingProvider.(Exclusions).Excluded(ctx, avoid)
}

// TestAnAvoidedAnswerIsHeldToTheTolerance (#250, the lead's routing: every
// answer checked, avoided ones included): a leg whose first answer crosses a
// marking and is asked for again is refused where the avoided answer's ends
// drift beyond the tolerance, though it crosses nothing; the same answer
// placed at the locations is served.
func TestAnAvoidedAnswerIsHeldToTheTolerance(t *testing.T) {
	t.Parallel()

	for _, drift := range []float64{0, 1000} {
		routing := &shifted{RoutingProvider: &scripted{answers: [][]int64{{1, 9, 2}, {1, 3, 2}}}, from: 2, metres: drift}
		table := &tableFake{marked: map[int64]bool{9: true}}

		journey, err := plannerOn(t, routing, table, AvoidUpTo(3), EndsWithin(50)).Plan(t.Context(), Request{Origin: home, Destination: shops})

		if drift == 0 {
			if err != nil || len(journey.Legs) != 1 || !slices.Equal(waysOf(journey.Legs)[0], []int64{1, 3, 2}) {
				t.Fatalf("the avoided answer at the locations: Plan gave %v and %v, want the leg along 1, 3, 2", waysOf(journey.Legs), err)
			}

			continue
		}

		if journey.Legs != nil || !errors.Is(err, ErrEndBeyondTolerance) {
			t.Fatalf("the avoided answer %v m off: Plan gave %v and %v; want the leg refused with an error wrapping ErrEndBeyondTolerance", drift, waysOf(journey.Legs), err)
		}

		if routing.n != 2 {
			t.Errorf("the provider was asked %d time(s), want the first answer and the avoided one", routing.n)
		}
	}
}

// TestAConnectionTheWarnedLegsSearchFindsIsHeldToTheTolerance (#250, every
// answer checked): where avoidance ends in no route between two locations the
// user entered, the connection the search for a less marked leg finds is held
// to the tolerance too — served at the locations, refused where its ends
// drift beyond it, never taken as the warned leg's.
func TestAConnectionTheWarnedLegsSearchFindsIsHeldToTheTolerance(t *testing.T) {
	t.Parallel()

	for _, drift := range []float64{0, 1000} {
		table := &tableFake{marked: map[int64]bool{7: true}, barriers: barrierOn(9)}
		routing := &shifted{RoutingProvider: &stepped{table: table, steps: [][]int64{{1, 9, 7, 2}, nil, {1, 7, 3, 2}, nil}}, from: 3, metres: drift}

		journey, err := plannerOn(t, routing, table, AvoidUpTo(3), EndsWithin(50)).Plan(t.Context(), Request{Origin: home, Destination: shops})

		if drift == 0 {
			if err != nil || len(journey.Legs) != 1 || !journey.Legs[0].Warned || !slices.Equal(waysOf(journey.Legs)[0], []int64{1, 7, 3, 2}) {
				t.Fatalf("the search's connection at the locations: Plan gave %v and %v, want the warned leg along 1, 7, 3, 2", waysOf(journey.Legs), err)
			}

			continue
		}

		if journey.Legs != nil || !errors.Is(err, ErrEndBeyondTolerance) {
			t.Fatalf("the search's connection %v m off: Plan gave %v and %v; want the leg refused with an error wrapping ErrEndBeyondTolerance", drift, waysOf(journey.Legs), err)
		}

		if routing.n != 3 {
			t.Errorf("the provider was asked %d time(s), want the first answer, the avoided one and the search's", routing.n)
		}
	}
}

// TestAnEndToleranceThatIsNoDistanceIsRefused: EndsWithin takes a finite
// number of metres at or above zero, zero included, and refuses the rest.
func TestAnEndToleranceThatIsNoDistanceIsRefused(t *testing.T) {
	t.Parallel()

	for _, m := range []float64{-1, math.Nextafter(0, -1), math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := New(straightLegs(), copyHeld, clearAccess{}, clearAccess{}, EndsWithin(m)); err == nil {
			t.Errorf("New with an end tolerance of %v m returned no error, want a refusal", m)
		}
	}

	if _, err := New(straightLegs(), copyHeld, clearAccess{}, clearAccess{}, EndsWithin(0)); err != nil {
		t.Errorf("New with an end tolerance of 0 m: %v; want it taken, as the strictest", err)
	}
}

// TestTheEndDistanceNeverUnderstatesTheGeodesic: endDistance is never less
// than the WGS 84 geodesic, so a leg whose end lies beyond the tolerance on
// the spheroid is refused. It is held against the two quadrants, whose
// geodesics are the meridian and the equator, and against the
// ellipsoid's own length element over short distances at every latitude, where
// that element is the geodesic to far better than the margin asserted; and
// it overstates by no more than the ratio its derivation gives.
func TestTheEndDistanceNeverUnderstatesTheGeodesic(t *testing.T) {
	t.Parallel()

	f := 1 / wgs84InverseFlattening
	e2 := f * (2 - f)
	ratio := 1 / (1 - e2) / (1 - f) // c/M(0), the most it overstates by

	radii := func(phi float64) (meridian, primeVertical float64) {
		w := 1 - e2*math.Sin(phi)*math.Sin(phi)

		return wgs84SemiMajorAxisMetres * (1 - e2) / math.Pow(w, 1.5), wgs84SemiMajorAxisMetres / math.Sqrt(w)
	}

	// The meridian quadrant, equator to pole, is the integral of M(φ) over
	// [0, π/2], here by Simpson's rule; the equatorial quadrant is a·π/2.
	const steps = 100_000

	quadrant := 0.0

	for k := 0; k <= steps; k++ {
		m, _ := radii(math.Pi / 2 * float64(k) / steps)

		switch {
		case k == 0 || k == steps:
			quadrant += m
		case k%2 == 1:
			quadrant += 4 * m
		default:
			quadrant += 2 * m
		}
	}

	quadrant *= math.Pi / 2 / steps / 3

	for name, geodesic := range map[string]struct {
		a, b Point
		m    float64
	}{
		"the meridian quadrant":   {Point{0, 0}, Point{90, 0}, quadrant},
		"the equatorial quadrant": {Point{0, 0}, Point{0, 90}, wgs84SemiMajorAxisMetres * math.Pi / 2},
	} {
		if d := endDistance(geodesic.a, geodesic.b); d < geodesic.m {
			t.Errorf("%s: endDistance gave %v m, below the geodesic of %v m", name, d, geodesic.m)
		}
	}

	r := rand.New(rand.NewPCG(250, 1))

	for range 100_000 {
		a := Point{Latitude: (r.Float64()*2 - 1) * 89.9, Longitude: (r.Float64()*2 - 1) * 180}
		b := Point{Latitude: a.Latitude + (r.Float64()*2-1)*1e-3, Longitude: a.Longitude + (r.Float64()*2-1)*1e-3}

		phi := (a.Latitude + b.Latitude) / 2 * math.Pi / 180
		meridian, primeVertical := radii(phi)

		dPhi := (b.Latitude - a.Latitude) * math.Pi / 180
		dLambda := (b.Longitude - a.Longitude) * math.Pi / 180
		element := math.Hypot(meridian*dPhi, primeVertical*math.Cos(phi)*dLambda)

		d := endDistance(a, b)
		if d < element*(1-1e-6) || d > element*ratio*(1+1e-6) {
			t.Fatalf("endDistance(%v, %v) gave %v m against the ellipsoid's %v m; want it at or above, and at most %v times it", a, b, d, element, ratio)
		}
	}
}
