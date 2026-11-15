package planner

import (
	"context"
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
)

// The run at a location the user entered, and avoidance, at the planner
// (#248). The fakes here say what the tables would: which ways are marked,
// which barriers lie where, and what a run measures. The criteria are held
// against the real tables and engine in `internal/accessstore`'s
// `exclusion_integration_test.go`; these hold the reading itself.

// tableFake marks the ways in marked, reports each barrier on the edges
// along the way it is keyed by, and measures a run with run.
type tableFake struct {
	mu       sync.Mutex
	marked   map[int64]bool
	barriers map[int64][]int64 // way → barrier nodes on it
	areas    map[int64]bool    // way whose stretch crosses a marked area
	run      func([]Edge) (RunMeasure, error)
	runs     [][]Edge
}

func (f *tableFake) ExtractSHA256(context.Context) (string, error) { return testExtract, nil }

func (f *tableFake) Crossings(_ context.Context, edges []Edge) ([]Crossing, error) {
	var out []Crossing

	for i, e := range edges {
		if f.marked[e.WayID] {
			out = append(out, Crossing{Edge: i, Kind: "way", OSMType: "w", OSMID: e.WayID})
		}

		for _, n := range f.barriers[e.WayID] {
			out = append(out, Crossing{Edge: i, Kind: "barrier", OSMType: "n", OSMID: n, At: barrierAt(n)})
		}

		if f.areas[e.WayID] {
			out = append(out, Crossing{Edge: i, Kind: "area", OSMType: "w", OSMID: 777})
		}
	}

	return out, nil
}

func (f *tableFake) Run(_ context.Context, run []Edge) (RunMeasure, error) {
	f.mu.Lock()
	f.runs = append(f.runs, slices.Clone(run))
	f.mu.Unlock()

	if f.run == nil {
		return RunMeasure{}, ErrNoLeastRun
	}

	return f.run(run)
}

func barrierAt(n int64) Point { return Point{Latitude: 57, Longitude: 11 + float64(n)/1000} }

// scripted answers the n'th request with answers[n], the last repeated, and
// records what each request asked to avoid. A nil answer is an error.
type scripted struct {
	answers [][]int64 // way ids, one edge each, on a line east
	asked   [][]Point
}

func (s *scripted) Route(_ context.Context, from, to Point, avoid []Point) (Route, error) {
	s.asked = append(s.asked, slices.Clone(avoid))

	ways := s.answers[min(len(s.asked)-1, len(s.answers)-1)]
	if ways == nil {
		return Route{}, errors.New("no route keeps off the places")
	}

	return Route{Edges: lineOf(from, to, ways)}, nil
}

// lineOf is a leg of one edge per way, from from to to in equal steps.
func lineOf(from, to Point, ways []int64) []Edge {
	at := func(i int) Point {
		f := float64(i) / float64(len(ways))
		return Point{Latitude: from.Latitude + f*(to.Latitude-from.Latitude), Longitude: from.Longitude + f*(to.Longitude-from.Longitude)}
	}

	edges := make([]Edge, len(ways))
	for i, w := range ways {
		edges[i] = Edge{WayID: w, Shape: []Point{at(i), at(i + 1)}}
	}

	return edges
}

var (
	home  = Point{Latitude: 57.0, Longitude: 11.0}
	shops = Point{Latitude: 57.0, Longitude: 11.1}
)

func plannerOn(t *testing.T, routing RoutingProvider, table *tableFake, opts ...Option) *Planner {
	t.Helper()

	p, err := New(routing, copyHeld, table, routingData{id: testExtract}, opts...)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func admitted(least float64, ways ...int64) func([]Edge) (RunMeasure, error) {
	return func(run []Edge) (RunMeasure, error) {
		return RunMeasure{Marked: least, Least: least, LeastWays: ways}, nil
	}
}

// exitingAt is admitted, the run ending at the exit node given.
func exitingAt(exit int64, least float64, ways ...int64) func([]Edge) (RunMeasure, error) {
	return func(run []Edge) (RunMeasure, error) {
		return RunMeasure{Marked: least, Least: least, LeastWays: ways, Exit: exit}, nil
	}
}

func waysOf(legs []Leg) [][]int64 {
	out := make([][]int64, len(legs))
	for i, l := range legs {
		out[i] = wayIDs(l.Edges)
	}

	return out
}

func wayIDs(edges []Edge) []int64 {
	out := make([]int64, len(edges))
	for i, e := range edges {
		out[i] = e.WayID
	}

	return out
}

// TestALegAvoidingAMarkingIsAskedForAndProduced is AC1 at the planner: a leg
// whose first answer runs along a marked way in transit is asked for again
// avoiding a point on that way, and the answer that avoids it is the leg.
func TestALegAvoidingAMarkingIsAskedForAndProduced(t *testing.T) {
	t.Parallel()

	routing := &scripted{answers: [][]int64{{1, 9, 2}, {1, 3, 4, 2}}}
	table := &tableFake{marked: map[int64]bool{9: true}}

	journey, err := plannerOn(t, routing, table, AvoidUpTo(3)).Plan(t.Context(), Request{Origin: home, Destination: shops})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if got := waysOf(journey.Legs); !slices.Equal(got[0], []int64{1, 3, 4, 2}) {
		t.Fatalf("the leg runs along %v, want the answer avoiding way 9", got)
	}

	if len(routing.asked) != 2 || len(routing.asked[0]) != 0 || len(routing.asked[1]) != 1 {
		t.Fatalf("the provider was asked to avoid %v, want nothing and then one place", routing.asked)
	}

	marked := lineOf(home, shops, []int64{1, 9, 2})[1]
	if p := routing.asked[1][0]; !onSegment(p, marked.Shape[0], marked.Shape[1]) {
		t.Errorf("the place avoided, %v, is not on the marked edge %v", p, marked.Shape)
	}
}

func onSegment(p, a, b Point) bool {
	cross := (b.Latitude-a.Latitude)*(p.Longitude-a.Longitude) - (b.Longitude-a.Longitude)*(p.Latitude-a.Latitude)

	return math.Abs(cross) < 1e-12 &&
		p.Longitude >= min(a.Longitude, b.Longitude) && p.Longitude <= max(a.Longitude, b.Longitude) &&
		p.Latitude >= min(a.Latitude, b.Latitude) && p.Latitude <= max(a.Latitude, b.Latitude) &&
		p != a && p != b
}

// TestAvoidanceIsBoundedAndNeverServesACrossing is AC10 at the planner: an
// engine that keeps answering along the marking, one that finds nothing once
// told to avoid it, and a planner told never to ask again, each produce no leg,
// and the refusal is a CrossingError naming the marking.
func TestAvoidanceIsBoundedAndNeverServesACrossing(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		answers [][]int64
		bound   int
		asked   int
	}{
		{"an engine ignoring what it is told", [][]int64{{1, 9, 2}, {1, 9, 2}}, 4, 2},
		{"an engine ignoring it on a new way each time", [][]int64{{1, 9, 2}, {1, 3, 10, 2}, {1, 3, 4, 11, 2}, {1, 3, 4, 5, 12, 2}}, 2, 3},
		{"an engine that finds nothing avoiding it", [][]int64{{1, 9, 2}, nil}, 4, 2},
		{"no avoidance configured", [][]int64{{1, 9, 2}, {1, 3, 2}}, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := &scripted{answers: tc.answers}
			table := &tableFake{marked: map[int64]bool{9: true, 10: true, 11: true, 12: true}}

			journey, err := plannerOn(t, routing, table, AvoidUpTo(tc.bound)).Plan(t.Context(), Request{Origin: home, Destination: shops})

			var refusal *CrossingError
			if journey.Legs != nil || !errors.As(err, &refusal) {
				t.Fatalf("Plan gave %v and %v, want no leg and a CrossingError", waysOf(journey.Legs), err)
			}

			if len(routing.asked) != tc.asked {
				t.Errorf("the provider was asked %d times, want %d", len(routing.asked), tc.asked)
			}

			if len(refusal.Crossings) == 0 || refusal.Crossings[0].Kind != "way" {
				t.Errorf("the refusal names %v, want the marked way", refusal.Crossings)
			}
		})
	}
}

// TestTheRunAtALocationTheUserEnteredIsAdmitted is AC4 at the planner: the leg
// leaving an origin on a marked way, the leg reaching a destination on one, and
// both legs at an intermediate destination on one, are produced where the run
// measures no more than the least; the run handed to the measure is the leg's
// marked edges from that location, in the order travelled away from it.
func TestTheRunAtALocationTheUserEnteredIsAdmitted(t *testing.T) {
	t.Parallel()

	stop := Point{Latitude: 57.0, Longitude: 11.05}

	for _, tc := range []struct {
		name    string
		req     Request
		answers [][]int64
		runs    [][]int64
	}{
		{"the origin", Request{Origin: home, Destination: shops}, [][]int64{{8, 9, 1, 2}}, [][]int64{{8, 9}}},
		{"the destination", Request{Origin: shops, Destination: home}, [][]int64{{2, 1, 9, 8}}, [][]int64{{8, 9}}},
		{"an intermediate destination", Request{Origin: home, Intermediates: []Point{stop}, Destination: shops},
			[][]int64{{1, 9, 8}, {8, 9, 2}}, [][]int64{{8, 9}, {8, 9}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := &routeEach{answers: tc.answers}
			table := &tableFake{marked: map[int64]bool{8: true, 9: true}, run: admitted(100, 8, 9)}

			journey, err := plannerOn(t, routing, table).Plan(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("Plan refused the run: %v", err)
			}

			if len(journey.Legs) != len(tc.answers) {
				t.Fatalf("the journey has %d legs, want %d", len(journey.Legs), len(tc.answers))
			}

			if len(table.runs) != len(tc.runs) {
				t.Fatalf("%d runs were measured, want %d", len(table.runs), len(tc.runs))
			}

			for i, run := range table.runs {
				if !slices.Equal(wayIDs(run), tc.runs[i]) {
					t.Errorf("run %d was measured along %v, want %v from the location", i, wayIDs(run), tc.runs[i])
				}
			}

			// The run starts where the leg's shape meets the location.
			for i, run := range table.runs {
				var want Point

				for _, l := range journey.Legs {
					if l.Edges[0].WayID == 8 {
						want = l.Shape[0]
					}

					if l.Edges[len(l.Edges)-1].WayID == 8 {
						want = l.Shape[len(l.Shape)-1]
					}

					if run[0].Shape[0] == want {
						break
					}
				}

				if run[0].Shape[0] != want {
					t.Errorf("run %d starts at %v, which is no end of a leg at the location", i, run[0].Shape[0])
				}
			}
		})
	}
}

// routeEach answers the i'th request with answers[i], between the points it
// is asked for.
type routeEach struct {
	answers [][]int64
	n       int
}

func (r *routeEach) Route(_ context.Context, from, to Point, _ []Point) (Route, error) {
	ways := r.answers[min(r.n, len(r.answers)-1)]
	r.n++

	return Route{Edges: lineOf(from, to, ways)}, nil
}

// TestARunLongerThanTheLeastIsAvoidedOffItsOtherWays is AC6 at the planner: a
// run that measures more than the least is not admitted; the leg is asked for
// again avoiding the run's ways the least does not use, and never the least's;
// and the answer along the least is the leg.
func TestARunLongerThanTheLeastIsAvoidedOffItsOtherWays(t *testing.T) {
	t.Parallel()

	routing := &scripted{answers: [][]int64{{8, 7, 2}, {8, 1, 2}}}
	table := &tableFake{marked: map[int64]bool{8: true, 7: true}, run: func(run []Edge) (RunMeasure, error) {
		if wayIDs(run)[len(run)-1] == 7 {
			return RunMeasure{Marked: 300, Least: 100, LeastWays: []int64{8}}, nil
		}

		return RunMeasure{Marked: 100, Least: 100, LeastWays: []int64{8}}, nil
	}}

	journey, err := plannerOn(t, routing, table, AvoidUpTo(2)).Plan(t.Context(), Request{Origin: home, Destination: shops})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if got := waysOf(journey.Legs)[0]; !slices.Equal(got, []int64{8, 1, 2}) {
		t.Fatalf("the leg runs along %v, want the near exit, 8 then 1", got)
	}

	far := lineOf(home, shops, []int64{8, 7, 2})

	if len(routing.asked[1]) != 1 || !onSegment(routing.asked[1][0], far[1].Shape[0], far[1].Shape[1]) {
		t.Errorf("the second request avoided %v, want one point on way 7 alone", routing.asked[1])
	}

	// Refused without avoidance, the longer run names its ways.
	_, err = plannerOn(t, &scripted{answers: [][]int64{{8, 7, 2}}}, table).Plan(t.Context(), Request{Origin: home, Destination: shops})

	var refusal *CrossingError
	if !errors.As(err, &refusal) || len(refusal.Crossings) != 2 {
		t.Fatalf("a run longer than the least gave %v, want a CrossingError naming its two ways", err)
	}
}

// TestNoRunIsAdmittedThatIsNotBounded: a least that is undefined, a run that
// never reaches an unmarked way, one crossing a marked area, and one measuring
// what is not a number, each refuse the leg, and none is avoided whole.
func TestNoRunIsAdmittedThatIsNotBounded(t *testing.T) {
	t.Parallel()

	nan := func([]Edge) (RunMeasure, error) { return RunMeasure{Marked: math.NaN(), Least: 100}, nil }

	for _, tc := range []struct {
		name  string
		ways  []int64
		table *tableFake
		ends  legEnds
	}{
		{"a least that is undefined", []int64{8, 1}, &tableFake{marked: map[int64]bool{8: true}}, legEnds{true, true}},
		{"a leg along marked ways alone", []int64{8, 9}, &tableFake{marked: map[int64]bool{8: true, 9: true}, run: admitted(1)}, legEnds{true, true}},
		{"a run crossing a marked area", []int64{8, 1}, &tableFake{marked: map[int64]bool{8: true}, areas: map[int64]bool{8: true}, run: admitted(1)}, legEnds{true, true}},
		{"a run measuring NaN", []int64{8, 1}, &tableFake{marked: map[int64]bool{8: true}, run: nan}, legEnds{true, true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := &scripted{answers: [][]int64{tc.ways}}

			_, err := plannerOn(t, routing, tc.table, AvoidUpTo(3)).checkedLeg(t.Context(), 0, home, shops, tc.ends)

			var refusal *CrossingError
			if !errors.As(err, &refusal) {
				t.Fatalf("checkedLeg gave %v, want a CrossingError", err)
			}

			if len(routing.asked) != 1 {
				t.Errorf("the leg was asked for %d times, want once: the location's way out is never avoided", len(routing.asked))
			}
		})
	}
}

// TestAnEndTheUserDidNotEnterHasNoRun: at an end that is not a location the
// user entered — a stopping position the system chose, #256's — the leg's
// marked edges are in transit, so they refuse the leg and are what it is asked
// for again avoiding.
func TestAnEndTheUserDidNotEnterHasNoRun(t *testing.T) {
	t.Parallel()

	routing := &scripted{answers: [][]int64{{8, 1}, {3, 1}}}
	table := &tableFake{marked: map[int64]bool{8: true}, run: admitted(1, 8)}

	leg, err := plannerOn(t, routing, table, AvoidUpTo(1)).checkedLeg(t.Context(), 0, home, shops, legEnds{fromEntered: false, toEntered: true})
	if err != nil {
		t.Fatalf("checkedLeg: %v", err)
	}

	if !slices.Equal(wayIDs(leg.Edges), []int64{3, 1}) || len(table.runs) != 0 {
		t.Errorf("the leg runs along %v with %d runs measured, want the answer avoiding way 8 and none", wayIDs(leg.Edges), len(table.runs))
	}
}

// TestBeyondTheRunALegIsInTransit is AC12 at the planner: past an admitted
// run, a marked way or a marked barrier refuses the leg, and is what the leg is
// asked for again avoiding; the barrier at the run's own exit node, passed
// once by the run's last edge and the first edge after it, is the run's.
func TestBeyondTheRunALegIsInTransit(t *testing.T) {
	t.Parallel()

	t.Run("a barrier further on", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 3, 2}, {8, 1, 4, 2}}}
		table := &tableFake{marked: map[int64]bool{8: true}, barriers: map[int64][]int64{3: {55}}, run: admitted(10, 8)}

		journey, err := plannerOn(t, routing, table, AvoidUpTo(1)).Plan(t.Context(), Request{Origin: home, Destination: shops})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		if got := waysOf(journey.Legs)[0]; !slices.Equal(got, []int64{8, 1, 4, 2}) {
			t.Errorf("the leg runs along %v, want the one avoiding the barrier on way 3", got)
		}

		if !slices.Equal(routing.asked[1], []Point{barrierAt(55)}) {
			t.Errorf("the leg was asked for again avoiding %v, want the barrier's position", routing.asked[1])
		}
	})

	t.Run("a barrier further on, with nothing around it", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 3, 2}, nil}}
		table := &tableFake{marked: map[int64]bool{8: true}, barriers: map[int64][]int64{3: {55}}, run: admitted(10, 8)}

		_, err := plannerOn(t, routing, table, AvoidUpTo(1)).Plan(t.Context(), Request{Origin: home, Destination: shops})

		var refusal *CrossingError
		if !errors.As(err, &refusal) || len(refusal.Crossings) != 1 || refusal.Crossings[0].OSMID != 55 {
			t.Fatalf("Plan gave %v, want a CrossingError naming the barrier alone", err)
		}
	})

	t.Run("a marked way further on", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 9, 2}}}
		table := &tableFake{marked: map[int64]bool{8: true, 9: true}, run: admitted(10, 8)}

		_, err := plannerOn(t, routing, table).Plan(t.Context(), Request{Origin: home, Destination: shops})

		var refusal *CrossingError
		if !errors.As(err, &refusal) || len(refusal.Crossings) != 1 || refusal.Crossings[0].OSMID != 9 {
			t.Fatalf("Plan gave %v, want a CrossingError naming way 9 alone", err)
		}
	})

	t.Run("the gate at the run's exit", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 2}}}
		table := &tableFake{marked: map[int64]bool{8: true}, barriers: map[int64][]int64{8: {66}, 1: {66}}, run: exitingAt(66, 10, 8)}

		if _, err := plannerOn(t, routing, table).Plan(t.Context(), Request{Origin: home, Destination: shops}); err != nil {
			t.Fatalf("the gate where the run meets the unmarked way refused the leg: %v", err)
		}
	})

	// The safety verdict's S1 on #270: a customers-only way 8, not closed,
	// meets the public road 1 at node 65 and again at node 77, 15 m on, where
	// a private lift_gate stands across the road. The run ends at 65 and the
	// leg drives on along the road through 77. The gate lies on way 8 and
	// within the tolerance of the run's stretch, so the run's last edge is
	// found passing it; it is in transit all the same.
	t.Run("a gate where the run's way meets the road a second time", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 2}}}
		table := &tableFake{marked: map[int64]bool{8: true}, barriers: map[int64][]int64{8: {77}, 1: {77}}, run: exitingAt(65, 10, 8)}

		_, err := plannerOn(t, routing, table).Plan(t.Context(), Request{Origin: home, Destination: shops})

		var refusal *CrossingError
		if !errors.As(err, &refusal) || len(refusal.Crossings) != 1 || refusal.Crossings[0].OSMID != 77 || refusal.Crossings[0].Edge != 1 {
			t.Fatalf("Plan gave %v, want a CrossingError naming the gate at node 77 on the road alone", err)
		}
	})

	t.Run("a run with no exit node admits no gate", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 2}}}
		table := &tableFake{marked: map[int64]bool{8: true}, barriers: map[int64][]int64{8: {66}, 1: {66}}, run: admitted(10, 8)}

		if _, err := plannerOn(t, routing, table).Plan(t.Context(), Request{Origin: home, Destination: shops}); !errors.Is(err, ErrCrossesMarking) {
			t.Fatalf("Plan gave %v, want the gate refused where the measure names no exit", err)
		}
	})

	t.Run("a barrier after the run that the run does not pass", func(t *testing.T) {
		t.Parallel()

		routing := &scripted{answers: [][]int64{{8, 1, 2}}}
		table := &tableFake{marked: map[int64]bool{8: true}, barriers: map[int64][]int64{1: {67}}, run: admitted(10, 8)}

		_, err := plannerOn(t, routing, table).Plan(t.Context(), Request{Origin: home, Destination: shops})
		if !errors.Is(err, ErrCrossesMarking) {
			t.Fatalf("Plan gave %v, want the barrier on the first unmarked way refused", err)
		}
	})
}

// TestARunNotMeasuredIsNotServed: a measure that fails for any reason but an
// undefined least leaves the leg unchecked.
func TestARunNotMeasuredIsNotServed(t *testing.T) {
	t.Parallel()

	table := &tableFake{marked: map[int64]bool{8: true}, run: func([]Edge) (RunMeasure, error) {
		return RunMeasure{}, errors.New("the database is gone")
	}}

	journey, err := plannerOn(t, &scripted{answers: [][]int64{{8, 1}}}, table).Plan(t.Context(), Request{Origin: home, Destination: shops})
	if !errors.Is(err, ErrAccessUnchecked) || journey.Legs != nil {
		t.Fatalf("Plan gave %d legs and %v, want none and ErrAccessUnchecked", len(journey.Legs), err)
	}
}

// TestAvoidUpToRefusesANegativeBound is New's half of the bound.
func TestAvoidUpToRefusesANegativeBound(t *testing.T) {
	t.Parallel()

	if _, err := New(&scripted{}, copyHeld, &tableFake{}, routingData{id: testExtract}, AvoidUpTo(-1)); err == nil {
		t.Fatal("New accepted a negative avoidance bound")
	}
}
