package planner

import (
	"cmp"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

// The tests here are #3's (`FR-004`). The planner produces one alternative
// today and builds it from the request, so it cannot drop or replace a
// waypoint; the alternatives that try are a stub optimizer's, which #3's Notes
// allow, passed through the check every alternative leaves the planner by.

// produce is the system producing alternatives: whatever an optimizer offers
// leaves the planner only through requireEnteredWaypoints, so where the check
// refuses, nothing is produced.
func produce(req Request, offered []Journey) ([]Journey, error) {
	if err := requireEnteredWaypoints(req, offered); err != nil {
		return nil, err
	}

	return offered, nil
}

// asSet returns points ordered by latitude then longitude, so two collections
// holding the same locations, each as many times, compare equal with
// slices.Equal. It is the tests' own reading of set identity, written apart
// from the implementation's.
func asSet(points []Point) []Point {
	s := slices.Clone(points)
	slices.SortFunc(s, func(a, b Point) int {
		return cmp.Or(cmp.Compare(a.Latitude, b.Latitude), cmp.Compare(a.Longitude, b.Longitude))
	})

	return s
}

// waypointsOf returns a journey's mandatory waypoints as the tests read them:
// its locations between the first and the last.
func waypointsOf(t *testing.T, j Journey) []Point {
	t.Helper()

	if len(j.Locations) < 2 {
		t.Fatalf("an alternative has %d locations, want an origin and a destination", len(j.Locations))
	}

	return j.Locations[1 : len(j.Locations)-1]
}

// assertEachComparesEqual is the criterion both of #3's share: each produced
// alternative's mandatory-waypoint set compares equal to the entered set.
func assertEachComparesEqual(t *testing.T, req Request, produced []Journey) {
	t.Helper()

	for i, alt := range produced {
		if got, want := asSet(waypointsOf(t, alt)), asSet(req.Intermediates); !slices.Equal(got, want) {
			t.Fatalf("alternative %d has the mandatory-waypoint set %+v, want the entered set %+v", i, got, want)
		}
	}
}

// journeyTime drives straight lines at a constant speed, so a journey's time is
// its length, as in TestAFasterOrderDoesNotOverrideTheEnteredOrder.
func journeyTime(locations []Point) float64 {
	var total float64
	for i := range len(locations) - 1 {
		total += math.Hypot(locations[i+1].Latitude-locations[i].Latitude, locations[i+1].Longitude-locations[i].Longitude)
	}

	return total
}

var (
	wpStockholm = Point{Latitude: 59.3293, Longitude: 18.0686}
	wpJonkoping = Point{Latitude: 57.7826, Longitude: 14.1618}
	wpVisby     = Point{Latitude: 57.6348, Longitude: 18.2948}
	wpGoteborg  = Point{Latitude: 57.7089, Longitude: 11.9746}
)

// TestACostlyMandatoryWaypointIsNotDropped is #3 AC1 (`FR-004`): given an
// entered mandatory waypoint that raises the journey's cost relative to routing
// without it, each alternative's mandatory-waypoint set compares equal to the
// entered set, the costly waypoint having not been dropped.
//
// Visby is on Gotland, east of a journey from Stockholm through Jönköping to
// Göteborg, and the test first proves that leaving it out is cheaper, so the
// property is not passed vacuously.
func TestACostlyMandatoryWaypointIsNotDropped(t *testing.T) {
	t.Parallel()

	req := Request{Origin: wpStockholm, Intermediates: []Point{wpJonkoping, wpVisby}, Destination: wpGoteborg}
	withoutVisby := []Point{wpStockholm, wpJonkoping, wpGoteborg}

	if with, without := journeyTime(requiredLocations(req)), journeyTime(withoutVisby); with <= without {
		t.Fatalf("the fixture is wrong: the journey takes %.3f with Visby and %.3f without it, want Visby to raise its cost", with, without)
	}

	t.Run("the planner", func(t *testing.T) {
		t.Parallel()

		routing := straightLegs()

		p, err := New(routing, copyHeld, clearAccess{}, clearAccess{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("Plan refused the journey: %v", err)
		}

		assertEachComparesEqual(t, req, []Journey{journey})

		// Kept is visited: a leg was routed to Visby and one on from it.
		if !slices.Contains(routing.asked, [2]Point{wpJonkoping, wpVisby}) || !slices.Contains(routing.asked, [2]Point{wpVisby, wpGoteborg}) {
			t.Fatalf("the provider was asked %+v, want legs routed to Visby and on from it", routing.asked)
		}
	})

	t.Run("an optimizer dropping it", func(t *testing.T) {
		t.Parallel()

		kept := Journey{Locations: requiredLocations(req)}
		dropped := Journey{Locations: withoutVisby}

		produced, err := produce(req, []Journey{kept, dropped})
		if !errors.Is(err, ErrWaypointsNotPreserved) {
			t.Fatalf("an alternative dropping the costly waypoint was offered and the error is %v, want one wrapping %q", err, ErrWaypointsNotPreserved)
		}

		assertEachComparesEqual(t, req, produced)

		// Not vacuous: an alternative keeping it is produced.
		produced, err = produce(req, []Journey{kept})
		if err != nil || len(produced) != 1 {
			t.Fatalf("an alternative keeping every entered waypoint produced %d alternatives and the error %v, want it produced", len(produced), err)
		}

		assertEachComparesEqual(t, req, produced)
	})
}

// TestABetterScoringNearbyLocationIsNotSubstituted is #3 AC2 (`FR-004`): given
// an entered mandatory waypoint that has a nearby location the optimizer scores
// more favourably, each alternative's mandatory-waypoint set compares equal to
// the entered set, the nearby location having not been substituted for it.
//
// The nearby location is about a hundred metres from Jönköping, and the stub
// optimizer scores it above the entered waypoint. Two substitutions are
// offered: the nearby location in the waypoint's place, and the nearby location
// beside it — the second is the one a check of inclusion would pass (#3 Notes),
// and the test proves that it would before holding it to set identity.
func TestABetterScoringNearbyLocationIsNotSubstituted(t *testing.T) {
	t.Parallel()

	nearby := Point{Latitude: wpJonkoping.Latitude + 0.0009, Longitude: wpJonkoping.Longitude}
	req := Request{Origin: wpStockholm, Intermediates: []Point{wpJonkoping}, Destination: wpGoteborg}

	score := map[Point]float64{wpJonkoping: 1, nearby: 3}
	if score[nearby] <= score[wpJonkoping] {
		t.Fatal("the fixture is wrong: the stub optimizer must score the nearby location above the entered waypoint")
	}

	t.Run("the planner", func(t *testing.T) {
		t.Parallel()

		p, err := New(straightLegs(), copyHeld, clearAccess{}, clearAccess{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("Plan refused the journey: %v", err)
		}

		assertEachComparesEqual(t, req, []Journey{journey})

		if slices.Contains(journey.Locations, nearby) {
			t.Fatalf("the journey visits %+v, which holds the nearby location %+v", journey.Locations, nearby)
		}
	})

	faithful := Journey{Locations: requiredLocations(req)}
	inPlace := Journey{Locations: []Point{wpStockholm, nearby, wpGoteborg}}
	beside := Journey{Locations: []Point{wpStockholm, wpJonkoping, nearby, wpGoteborg}}

	// The trap is set: every entered waypoint is present in the alternative
	// holding the nearby location beside it.
	for _, w := range req.Intermediates {
		if !slices.Contains(waypointsOf(t, beside), w) {
			t.Fatalf("the fixture is wrong: the alternative %+v does not include the entered waypoint %+v", beside.Locations, w)
		}
	}

	for _, tc := range []struct {
		name       string
		substitute Journey
	}{
		{"in its place", inPlace},
		{"beside it", beside},
	} {
		t.Run("an optimizer substituting it "+tc.name, func(t *testing.T) {
			t.Parallel()

			produced, err := produce(req, []Journey{faithful, tc.substitute})
			if !errors.Is(err, ErrWaypointsNotPreserved) {
				t.Fatalf("the alternative %+v was offered and the error is %v, want one wrapping %q", tc.substitute.Locations, err, ErrWaypointsNotPreserved)
			}

			assertEachComparesEqual(t, req, produced)
		})
	}

	t.Run("an optimizer keeping it", func(t *testing.T) {
		t.Parallel()

		produced, err := produce(req, []Journey{faithful})
		if err != nil || len(produced) != 1 {
			t.Fatalf("the faithful alternative produced %d alternatives and the error %v, want it produced", len(produced), err)
		}

		assertEachComparesEqual(t, req, produced)
	})
}

// TestOnlyAlternativesHoldingTheEnteredSetAreProduced is both of #3's criteria
// as a property over requests. For each, the alternative the planner plans is
// produced, and an alternative that drops one waypoint, replaces one with a
// location beside it, adds a location beside one, or holds one entered
// waypoint in place of another is refused. randomRequest repeats an entered
// location one time in eight, so the last of those is drawn against a
// waypoint entered twice — which a comparison of plain sets would pass — and the
// test requires that it was.
func TestOnlyAlternativesHoldingTheEnteredSetAreProduced(t *testing.T) {
	t.Parallel()

	const seed, cases = 3, 2000

	r := rand.New(rand.NewPCG(seed, seed))

	// nudge returns a location beside p that is not p, and is a coordinate.
	nudge := func(p Point) Point {
		if p.Latitude > 0 {
			return Point{Latitude: p.Latitude - 1e-6, Longitude: p.Longitude}
		}

		return Point{Latitude: p.Latitude + 1e-6, Longitude: p.Longitude}
	}

	withEnds := func(req Request, waypoints []Point) Journey {
		return Journey{Locations: slices.Concat([]Point{req.Origin}, waypoints, []Point{req.Destination})}
	}

	repeatsSwapped := 0

	for c := range cases {
		req := randomRequest(r, 1, 6)

		p, err := New(randomShapes(r), copyHeld, clearAccess{}, clearAccess{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		planned, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("seed %d case %d: Plan(%+v) refused the journey: %v", seed, c, req, err)
		}

		if _, err := produce(req, []Journey{planned}); err != nil {
			t.Fatalf("seed %d case %d: the planned journey %+v was refused: %v", seed, c, planned.Locations, err)
		}

		entered := req.Intermediates
		i := r.IntN(len(entered))

		swapped := slices.Clone(entered)
		swapped[i] = entered[(i+1)%len(entered)]

		offers := map[string][]Point{
			"dropping one":           slices.Delete(slices.Clone(entered), i, i+1),
			"replacing one":          slices.Concat(entered[:i], []Point{nudge(entered[i])}, entered[i+1:]),
			"adding one beside one":  slices.Insert(slices.Clone(entered), i+1, nudge(entered[i])),
			"one in another's place": swapped,
		}

		for name, waypoints := range offers {
			if slices.Equal(asSet(waypoints), asSet(entered)) {
				// One waypoint in place of an equal one is no change.
				continue
			}

			if name == "one in another's place" && slices.Equal(slices.Compact(asSet(waypoints)), slices.Compact(asSet(entered))) {
				repeatsSwapped++
			}

			offered := withEnds(req, waypoints)

			produced, err := produce(req, []Journey{planned, offered})
			if !errors.Is(err, ErrWaypointsNotPreserved) || produced != nil {
				t.Fatalf("seed %d case %d: an alternative %s, %+v, was offered against the entered %+v, and %d alternatives were produced with the error %v; want none produced and an error wrapping %q",
					seed, c, name, waypoints, entered, len(produced), err, ErrWaypointsNotPreserved)
			}
		}
	}

	if repeatsSwapped == 0 {
		t.Fatalf("seed %d drew no alternative holding an entered waypoint in place of another that a comparison of plain sets would pass, want at least one", seed)
	}
}

// TestAnAlternativeWithoutEndsIsRefused: a journey with no origin and
// destination has no mandatory waypoints to compare, and is never produced.
func TestAnAlternativeWithoutEndsIsRefused(t *testing.T) {
	t.Parallel()

	req := Request{Origin: wpStockholm, Destination: wpGoteborg}

	for _, locations := range [][]Point{nil, {wpStockholm}} {
		if _, err := produce(req, []Journey{{Locations: locations}}); !errors.Is(err, ErrWaypointsNotPreserved) {
			t.Fatalf("an alternative with the locations %+v was offered and the error is %v, want one wrapping %q", locations, err, ErrWaypointsNotPreserved)
		}
	}
}

// TestEveryJourneyPlanReturnsPassesTheWaypointCheck holds Plan's wiring of
// `FR-004` to the source, as checkedleg_test.go holds #255 AC4: Plan cannot
// fail the check today, so no behavioural test notices the call removed, or a
// later return that skips it. In Plan, every return either returns the empty
// Journey{} — no journey — or returns a variable that requireEnteredWaypoints
// was handed, in a call that comes before the return.
//
// It refuses to pass having found no Plan, no call, or no return of a journey.
func TestEveryJourneyPlanReturnsPassesTheWaypointCheck(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "planner.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	var plan *ast.FuncDecl

	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "Plan" && fn.Recv != nil {
			plan = fn
		}
	}

	if plan == nil {
		t.Fatal("planner.go declares no method Plan")
	}

	// checked holds each variable handed to requireEnteredWaypoints, with
	// where the call is.
	checked := map[string]token.Pos{}

	ast.Inspect(plan.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "requireEnteredWaypoints" || len(call.Args) != 2 {
			return true
		}

		ast.Inspect(call.Args[1], func(m ast.Node) bool {
			if id, ok := m.(*ast.Ident); ok {
				checked[id.Name] = call.Pos()
			}

			return true
		})

		return true
	})

	if len(checked) == 0 {
		t.Fatal("Plan never calls requireEnteredWaypoints")
	}

	journeys := 0

	ast.Inspect(plan.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) == 0 {
			return true
		}

		switch first := ret.Results[0].(type) {
		case *ast.CompositeLit:
			if len(first.Elts) == 0 {
				return true
			}
		case *ast.Ident:
			if at, ok := checked[first.Name]; ok && at < ret.Pos() {
				journeys++

				return true
			}
		}

		t.Errorf("%s: Plan returns a journey that requireEnteredWaypoints was not handed before it", fset.Position(ret.Pos()))

		return true
	})

	if journeys == 0 {
		t.Fatal("Plan returns no journey that requireEnteredWaypoints was handed")
	}
}
