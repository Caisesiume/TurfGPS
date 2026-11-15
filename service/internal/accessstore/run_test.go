package accessstore

import (
	"errors"
	"slices"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// measure's reading of a run, against networks written out here. The SQL that
// gives it the numbers is `migrations/0004_access_run.sql`, held by its
// verify file and, with the engine, by `exclusion_integration_test.go`.

// A private road w1 (nodes 1, 2, 3) continuing as w2 (3, 4) to an exit at 4,
// and, the other way, w0 (5, 1) to an exit at 5. The run's first point lies
// on w1's first segment, 0.1 m from node 1 and 0.7 m from node 2. The lengths
// are chosen so no sum of them is exact in binary.
func network() map[int64]runWay {
	return map[int64]runWay{
		1: {nodes: []int64{1, 2, 3}, segments: []float64{0.8, 0.3}},
		2: {nodes: []int64{3, 4}, segments: []float64{0.7}, exits: []int64{4}},
		0: {nodes: []int64{5, 1}, segments: []float64{2.9}, exits: []int64{5}},
	}
}

var start = runPoint{segment: 0, back: 0.1, ahead: 0.7}

// edges is a run along the ways given, two points each; only the way ids and
// the edge count matter to measure, which reads the points' placement.
func edges(ways ...int64) []planner.Edge {
	out := make([]planner.Edge, len(ways))
	for i, w := range ways {
		out[i] = planner.Edge{WayID: w, Shape: []planner.Point{{}, {}}}
	}

	return out
}

func at(v int) runPoint { return runPoint{vertex: v} }

func TestALegAlongTheLeastRunMeasuresExactlyTheLeast(t *testing.T) {
	t.Parallel()

	// Out by w1 and w2 to node 4: 0.7 + 0.3 + 0.7, the least, since the way
	// back by w0 is 0.1 + 2.9.
	got, err := measure(edges(1, 2), []runPoint{start, at(2), at(0), at(1)}, network())
	if err != nil {
		t.Fatalf("measure: %v", err)
	}

	if got.Marked != got.Least {
		t.Fatalf("the run along the least measures %.17g against a least of %.17g, want exactly equal", got.Marked, got.Least)
	}

	// Summed at run time, in path order: a constant expression is exact.
	a, b, c := 0.7, 0.3, 0.7
	if want := a + b + c; got.Least != want {
		t.Errorf("the least is %.17g, want %.17g", got.Least, want)
	}

	if !slices.Equal(got.LeastWays, []int64{1, 2}) {
		t.Errorf("the least runs along %v, want w1 then w2", got.LeastWays)
	}

	if got.Exit != 4 {
		t.Errorf("the run's exit is node %d, want node 4, where it ends", got.Exit)
	}
}

// TestTheRunsExitIsTheNodeItEndsAt: a way that meets unmarked ways at both its
// ends, nodes 1 and 3, gives as the run's exit the one the run ends at, never
// the other (#270's S1).
func TestTheRunsExitIsTheNodeItEndsAt(t *testing.T) {
	t.Parallel()

	loop := map[int64]runWay{1: {nodes: []int64{1, 2, 3}, segments: []float64{0.8, 0.3}, exits: []int64{1, 3}}}

	for _, tc := range []struct {
		end  int
		want int64
	}{{0, 1}, {2, 3}} {
		got, err := measure(edges(1), []runPoint{start, at(tc.end)}, loop)
		if err != nil {
			t.Fatalf("measure: %v", err)
		}

		if got.Exit != tc.want {
			t.Errorf("the run ending at vertex %d gave exit node %d, want %d", tc.end, got.Exit, tc.want)
		}
	}
}

func TestARunPastTheNearestExitMeasuresMore(t *testing.T) {
	t.Parallel()

	n := network()
	n[0] = runWay{nodes: []int64{5, 1}, segments: []float64{0.2}, exits: []int64{5}}

	got, err := measure(edges(1, 2), []runPoint{start, at(2), at(0), at(1)}, n)
	if err != nil {
		t.Fatalf("measure: %v", err)
	}

	back1, w0 := 0.1, 0.2
	if !(got.Marked > got.Least) || got.Least != back1+w0 || !slices.Equal(got.LeastWays, []int64{1, 0}) {
		t.Errorf("measured %v, want the far exit longer than the least of 0.3 m by w1 and w0", got)
	}

	// Backward from the first point, the near way out measures the least.
	back, err := measure(edges(1, 0), []runPoint{start, at(0), at(1), at(0)}, n)
	if err != nil || back.Marked != back.Least {
		t.Errorf("the run by the near exit measured %v, %v; want it equal to the least", back, err)
	}
}

func TestARunThatCannotBePlacedIsNotMeasured(t *testing.T) {
	t.Parallel()

	closed := network()
	closed[1] = runWay{nodes: []int64{1, 2, 3, 1}, segments: []float64{0.8, 0.3, 0.9}}

	for _, tc := range []struct {
		name    string
		run     []planner.Edge
		points  []runPoint
		network map[int64]runWay
	}{
		{"turning back along a way", edges(1, 1, 2), []runPoint{start, at(2), at(2), at(1), at(1), at(1)}, network()},
		{"starting where the edge before did not end", edges(1, 2), []runPoint{start, at(2), at(1), at(0)}, network()},
		{"ending at no exit", edges(1), []runPoint{start, at(2)}, network()},
		{"along a closed way", edges(1, 2), []runPoint{start, at(2), at(0), at(1)}, closed},
		{"along a way outside the network", edges(1, 9), []runPoint{start, at(2), at(0), at(1)}, network()},
		{"with no exit reachable", edges(1), []runPoint{start, at(2)}, map[int64]runWay{1: network()[1]}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := measure(tc.run, tc.points, tc.network)
			if !errors.Is(err, planner.ErrNoLeastRun) {
				t.Fatalf("measure gave %v, %v; want ErrNoLeastRun", got, err)
			}
		})
	}
}

func TestANetworkThatIsNotOneIsUnchecked(t *testing.T) {
	t.Parallel()

	bad := network()
	bad[2] = runWay{nodes: []int64{3, 4}, segments: []float64{-1}, exits: []int64{4}}

	if _, err := measure(edges(1, 2), []runPoint{start, at(2), at(0), at(1)}, bad); !errors.Is(err, ErrUnchecked) {
		t.Errorf("a negative segment gave %v, want ErrUnchecked", err)
	}

	if _, err := measure(edges(1, 2), []runPoint{start, at(2)}, network()); !errors.Is(err, ErrUnchecked) {
		t.Errorf("too few points gave %v, want ErrUnchecked", err)
	}
}
