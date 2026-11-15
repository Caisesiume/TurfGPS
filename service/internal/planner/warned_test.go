package planner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
)

// The warned leg at the planner (#271). The fakes say what the tables and the
// engine would; the criteria are held against the real tables and engine in
// `internal/accessstore`'s `warned_integration_test.go`.

// network is an engine over a fixed set of connections between any two
// points, quickest first. Asked to keep off a place, it keeps off the whole
// way at it, as Valhalla's exclude_locations keeps off the road a place is
// mapped to: a point on an edge it answered names that edge's way, and a
// barrier's position names every way the table puts the barrier on. Where no
// connection keeps off every way so named it answers with ErrNoRoute, or with
// fail where that is set.
type network struct {
	table       *tableFake
	connections [][]int64
	fail        error

	mu    sync.Mutex
	at    map[Point]int64
	asked [][]Point
}

func (n *network) Route(_ context.Context, from, to Point, avoid []Point) (Route, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.asked = append(n.asked, slices.Clone(avoid))

	if n.at == nil {
		n.at = map[Point]int64{}
	}

	off := map[int64]bool{}

	for _, pl := range avoid {
		if w, ok := n.at[pl]; ok {
			off[w] = true
		}

		for way, nodes := range n.table.barriers {
			for _, b := range nodes {
				if barrierAt(b) == pl {
					off[way] = true
				}
			}
		}
	}

	for _, ways := range n.connections {
		if slices.ContainsFunc(ways, func(w int64) bool { return off[w] }) {
			continue
		}

		edges := lineOf(from, to, ways)
		for _, e := range edges {
			n.at[onEdge(e)] = e.WayID
		}

		return Route{Edges: edges}, nil
	}

	if n.fail != nil {
		return Route{}, n.fail
	}

	return Route{}, fmt.Errorf("the engine found no path: %w", ErrNoRoute)
}

// stepped answers the n'th request with steps[n], the last repeated: a way
// list, or nil for ErrNoRoute.
type stepped struct {
	steps [][]int64
	asked [][]Point
}

func (s *stepped) Route(_ context.Context, from, to Point, avoid []Point) (Route, error) {
	s.asked = append(s.asked, slices.Clone(avoid))

	ways := s.steps[min(len(s.asked)-1, len(s.steps)-1)]
	if ways == nil {
		return Route{}, ErrNoRoute
	}

	return Route{Edges: lineOf(from, to, ways)}, nil
}

func barrierOn(ways ...int64) map[int64][]int64 {
	out := map[int64][]int64{}
	for i, w := range ways {
		out[w] = []int64{int64(50 + 10*i)}
	}

	return out
}

// named lists a warned leg's stretches as "kind id from-to".
func named(leg Leg) []string {
	out := make([]string, len(leg.Marked))
	for i, m := range leg.Marked {
		out[i] = fmt.Sprintf("%s %d %d-%d", m.Kind, m.OSMID, m.From, m.To)
	}

	return out
}

// TestALegNothingAvoidsBetweenLocationsTheUserEnteredIsWarned is #271 AC1 at
// the planner (`FR-129`, first criterion): a gate on every connection between
// two locations the user entered, which the engine finds no route keeping off,
// gives a warned leg naming it, not a refusal — for a journey's one leg and for
// both legs at an intermediate destination.
func TestALegNothingAvoidsBetweenLocationsTheUserEnteredIsWarned(t *testing.T) {
	t.Parallel()

	table := &tableFake{barriers: barrierOn(9)}
	via := Point{Latitude: 57.0, Longitude: 11.2}

	for _, req := range []Request{
		{Origin: home, Destination: shops},
		{Origin: home, Intermediates: []Point{via}, Destination: shops},
	} {
		routing := &network{table: table, connections: [][]int64{{1, 9, 2}}}

		journey, err := plannerOn(t, routing, table, AvoidUpTo(3)).Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("Plan refused the journey: %v", err)
		}

		for i, leg := range journey.Legs {
			if !leg.Warned || !slices.Equal(named(leg), []string{"barrier 50 1-2"}) {
				t.Errorf("leg %d is warned=%v naming %v, want warned naming the gate on its second edge", i, leg.Warned, named(leg))
			}

			if leg.Marked[0].At != barrierAt(50) {
				t.Errorf("leg %d names the gate at %v, want %v", i, leg.Marked[0].At, barrierAt(50))
			}
		}
	}
}

// TestOnlyANoRouteAnswerWarnsALeg: the warned leg replaces a refusal only
// where the engine has said nothing keeps off the marking. An engine failing
// otherwise, and a planner told never to ask again, leave the leg refused.
func TestOnlyANoRouteAnswerWarnsALeg(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		table   *tableFake
		routing func(*tableFake) RoutingProvider
		bound   int
	}{
		{"an engine that fails otherwise", &tableFake{barriers: barrierOn(9)}, func(tb *tableFake) RoutingProvider {
			return &network{table: tb, connections: [][]int64{{1, 9, 2}}, fail: errors.New("the engine timed out")}
		}, 3},
		{"no avoidance configured", &tableFake{barriers: barrierOn(9)}, func(tb *tableFake) RoutingProvider {
			return &network{table: tb, connections: [][]int64{{1, 9, 2}}}
		}, 0},
		{"an engine failing otherwise on the search for less", &tableFake{marked: map[int64]bool{7: true}, barriers: barrierOn(9)}, func(*tableFake) RoutingProvider {
			return &failingAfter{RoutingProvider: &stepped{steps: [][]int64{{1, 9, 7, 2}, nil}}, after: 2}
		}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			journey, err := plannerOn(t, tc.routing(tc.table), tc.table, AvoidUpTo(tc.bound)).Plan(t.Context(), Request{Origin: home, Destination: shops})

			var refusal *CrossingError
			if journey.Legs != nil || !errors.As(err, &refusal) {
				t.Fatalf("Plan gave %v and %v, want no leg and a CrossingError", waysOf(journey.Legs), err)
			}
		})
	}
}

// failingAfter answers as its provider does for the first after requests and
// fails with a plain error from then on.
type failingAfter struct {
	RoutingProvider
	after int
	n     int
}

func (f *failingAfter) Route(ctx context.Context, from, to Point, avoid []Point) (Route, error) {
	f.n++
	if f.n > f.after {
		return Route{}, errors.New("the engine is gone")
	}

	return f.RoutingProvider.Route(ctx, from, to, avoid)
}

// TestALegToAStopTheSystemChoseIsNeverWarned is #271 AC2 at the planner
// (`FR-129`, second criterion; `RD-068`): the same gate, one end of the leg a
// stopping position the system chose, told to the planner directly, gives no
// leg, warned or not, at either end.
func TestALegToAStopTheSystemChoseIsNeverWarned(t *testing.T) {
	t.Parallel()

	table := &tableFake{barriers: barrierOn(9)}

	for _, ends := range [][2]LegEnd{
		{{At: home, EnteredByUser: true}, {At: shops}},
		{{At: home}, {At: shops, EnteredByUser: true}},
	} {
		routing := &network{table: table, connections: [][]int64{{1, 9, 2}}}

		leg, err := plannerOn(t, routing, table, AvoidUpTo(3)).PlanLeg(t.Context(), ends[0], ends[1])

		var refusal *CrossingError
		if leg.Edges != nil || leg.Warned || !errors.As(err, &refusal) {
			t.Errorf("PlanLeg from %+v gave %v warned=%v and %v, want no leg and a CrossingError", ends, wayIDs(leg.Edges), leg.Warned, err)
		}
	}

	// Told as two locations the user entered, the same leg is warned, so the
	// refusal above is the end's and not the case's.
	routing := &network{table: table, connections: [][]int64{{1, 9, 2}}}
	if leg, err := plannerOn(t, routing, table, AvoidUpTo(3)).PlanLeg(t.Context(), LegEnd{At: home, EnteredByUser: true}, LegEnd{At: shops, EnteredByUser: true}); err != nil || !leg.Warned {
		t.Errorf("between two locations the user entered PlanLeg gave warned=%v and %v, want a warned leg", leg.Warned, err)
	}
}

// TestALegAConnectionAvoidingEveryMarkingJoinsIsNotWarned is #271 AC3 at the
// planner (`FR-129`, fourth criterion), and the case where the engine first
// says nothing keeps off the marking and then finds a connection that does:
// the leg is that connection, not warned.
func TestALegAConnectionAvoidingEveryMarkingJoinsIsNotWarned(t *testing.T) {
	t.Parallel()

	table := &tableFake{barriers: barrierOn(9)}

	for _, tc := range []struct {
		name    string
		routing RoutingProvider
	}{
		{"a way round the gate", &network{table: table, connections: [][]int64{{1, 9, 2}, {1, 3, 4, 2}}}},
		{"an engine that finds it only on the search for less", &stepped{steps: [][]int64{{1, 9, 2}, nil, {1, 3, 4, 2}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			journey, err := plannerOn(t, tc.routing, table, AvoidUpTo(3)).Plan(t.Context(), Request{Origin: home, Destination: shops})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			leg := journey.Legs[0]
			if leg.Warned || leg.Marked != nil || !slices.Equal(wayIDs(leg.Edges), []int64{1, 3, 4, 2}) {
				t.Errorf("the leg runs along %v warned=%v naming %v, want the way round, not warned", wayIDs(leg.Edges), leg.Warned, named(leg))
			}
		})
	}
}

// TestAWarnedLegCarriesNoMarkingAConnectionAvoids is #271 AC4 at the planner
// (`FR-130`, first criterion): every connection passes the gate on way 9, and
// the quickest also runs along marked way 7, which a connection passing the
// gate alone avoids. The warned leg passes the gate and runs along no way 7.
func TestAWarnedLegCarriesNoMarkingAConnectionAvoids(t *testing.T) {
	t.Parallel()

	table := &tableFake{marked: map[int64]bool{7: true}, barriers: barrierOn(9)}
	routing := &network{table: table, connections: [][]int64{{1, 9, 7, 2}, {1, 9, 3, 4, 2}}}

	journey, err := plannerOn(t, routing, table, AvoidUpTo(3)).Plan(t.Context(), Request{Origin: home, Destination: shops})
	if err != nil {
		t.Fatalf("Plan: %v (asked to avoid %v)", err, routing.asked)
	}

	leg := journey.Legs[0]
	if !leg.Warned || !slices.Equal(wayIDs(leg.Edges), []int64{1, 9, 3, 4, 2}) || !slices.Equal(named(leg), []string{"barrier 50 1-2"}) {
		t.Errorf("the leg runs along %v warned=%v naming %v, want the gate alone, warned", wayIDs(leg.Edges), leg.Warned, named(leg))
	}
}

// TestOfTwoMarkingsEitherOfWhichJoinsTheWarnedLegCrossesOne is #271 AC5 at the
// planner (`FR-130`, second criterion): two connections, each through a gate
// of its own, whichever the engine answers first, and with a quickest third
// passing both, which the warned leg is never.
func TestOfTwoMarkingsEitherOfWhichJoinsTheWarnedLegCrossesOne(t *testing.T) {
	t.Parallel()

	table := &tableFake{barriers: barrierOn(9, 8)}

	for _, order := range [][][]int64{{{1, 9, 2}, {1, 8, 2}}, {{1, 8, 2}, {1, 9, 2}}, {{1, 9, 8, 2}, {1, 9, 2}, {1, 8, 2}}} {
		routing := &network{table: table, connections: order}

		journey, err := plannerOn(t, routing, table, AvoidUpTo(3)).Plan(t.Context(), Request{Origin: home, Destination: shops})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		leg := journey.Legs[0]
		if got := named(leg); !leg.Warned || len(got) != 1 {
			t.Errorf("the leg runs along %v warned=%v naming %v, want one gate, warned", wayIDs(leg.Edges), leg.Warned, got)
		}
	}
}

// TestEveryMarkedStretchOfAWarnedLegIsNamedWhereItLies is #271 AC6 and AC7 at
// the planner (`FR-131`): a warned leg along marked way 7 and past the gate
// on way 9 names each, each locating its own stretch of the shape, and one
// leaving a home on marked way 8 along its run names the run too. A way the
// leg runs along over two edges is one stretch.
func TestEveryMarkedStretchOfAWarnedLegIsNamedWhereItLies(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		table *tableFake
		ways  []int64
		want  []string
	}{
		{"a marked way and a gate", &tableFake{marked: map[int64]bool{7: true}, barriers: barrierOn(9)}, []int64{1, 7, 9, 2}, []string{"way 7 1-2", "barrier 50 2-3"}},
		{"a marked way over two edges", &tableFake{marked: map[int64]bool{7: true}}, []int64{1, 7, 7, 2}, []string{"way 7 1-3"}},
		{"the run at a home on a private road", &tableFake{marked: map[int64]bool{8: true}, barriers: barrierOn(9), run: admitted(1, 8)}, []int64{8, 1, 9, 2}, []string{"way 8 0-1", "barrier 50 2-3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			routing := &network{table: tc.table, connections: [][]int64{tc.ways}}

			journey, err := plannerOn(t, routing, tc.table, AvoidUpTo(3)).Plan(t.Context(), Request{Origin: home, Destination: shops})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			leg := journey.Legs[0]
			if !leg.Warned || !slices.Equal(named(leg), tc.want) {
				t.Fatalf("the leg is warned=%v naming %v, want %v", leg.Warned, named(leg), tc.want)
			}

			// Each stretch is the shape of the edges along its feature.
			for _, m := range leg.Marked {
				for k := m.From; k < m.To; k++ {
					if w := leg.Edges[k].WayID; m.Kind == "way" && w != m.OSMID {
						t.Errorf("the stretch of way %d spans edge %d, along way %d", m.OSMID, k, w)
					}
				}
			}
		})
	}
}

// TestAWarnedLegItCannotVouchForIsRefused: a candidate whose run is not
// admitted, or which crosses a marked area, which nothing can be asked to keep
// off, is refused and not warned.
func TestAWarnedLegItCannotVouchForIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		table *tableFake
		ways  []int64
	}{
		{"a run with no least, before a gate", &tableFake{marked: map[int64]bool{8: true}, barriers: barrierOn(9)}, []int64{8, 1, 9, 2}},
		{"a leg crossing a marked area at a gate", &tableFake{areas: map[int64]bool{9: true}, barriers: barrierOn(9)}, []int64{1, 9, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			routing := &network{table: tc.table, connections: [][]int64{tc.ways}}

			journey, err := plannerOn(t, routing, tc.table, AvoidUpTo(3)).Plan(t.Context(), Request{Origin: home, Destination: shops})

			var refusal *CrossingError
			if journey.Legs != nil || !errors.As(err, &refusal) {
				t.Fatalf("Plan gave %v and %v, want no leg and a CrossingError", waysOf(journey.Legs), err)
			}
		})
	}
}
