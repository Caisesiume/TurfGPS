package accessstore

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// The run at a location the user entered (#248, `FR-126`, `Architecture.md §
// What the tables hold`): the least marked length that joins the location to
// an unmarked way, and the marked length the leg's run travels, measured the
// same way.
//
// THE LENGTHS ARE POSTGIS'S, THE SUMS ARE HERE, AND BOTH SUMS ARE ONE. The
// migration's `access_run_network` gives every marked way the run could travel
// with the length of each of its segments, as `ST_Length` on `geography`, and
// `access_run_points` says where the run's points lie on its ways
// (`migrations/0004_access_run.sql`). The least run is the shortest path over
// those segments from the run's first point to an exit node; the leg's run is
// the same segments, in the order the leg travels them, summed in that order.
// A leg along a least run therefore adds the same numbers in the same order as
// the shortest path did, and measures exactly the least, so no tolerance
// stands between the two: a run is admitted where its length is no more than
// the least, and a run longer by any amount is not.
//
// IT REFUSES WHAT IT CANNOT PLACE, with planner.ErrNoLeastRun. The leg's run is
// placed on the tables' geometry by its edges' ends, which the engine puts at
// the nodes of its graph — OpenStreetMap nodes — and each is held to the one
// beside it: an edge starts at the node the edge before it ended at, on its own
// way's nodes, and the run's last node is an exit of its way. A run whose
// ends disagree, that turns back along a way, or that runs along a way passing
// one node twice (a closed way, whose direction its nodes cannot tell) is not
// placed, and so not admitted.

// runNetworkQuery and runPointsQuery are the migration's two statements.
const (
	runNetworkQuery = `
SELECT network_way, network_nodes, segment_metres, exit_nodes
  FROM public.access_run_network($1::bigint)`

	runPointsQuery = `
SELECT point_index, vertex, segment, metres_back, metres_ahead
  FROM public.access_run_points($1::bigint[], $2::double precision[], $3::double precision[])`
)

// runWay is one marked way of the network a run may travel.
type runWay struct {
	nodes    []int64
	segments []float64 // segments[i] joins nodes[i] to nodes[i+1], in metres
	exits    []int64
}

// runPoint is where one point lies on its way: the nearest vertex, and the
// nearest segment with the lengths from its ends. Indices are zero-based.
type runPoint struct {
	vertex      int
	segment     int
	back, ahead float64
}

// Run measures a leg's run at a location the user entered.
func (s *Store) Run(ctx context.Context, run []planner.Edge) (planner.RunMeasure, error) {
	if len(run) == 0 {
		return planner.RunMeasure{}, fmt.Errorf("%w: a run of no edges", planner.ErrNoLeastRun)
	}

	ways := make([]int64, 0, 2*len(run))
	lons := make([]float64, 0, 2*len(run))
	lats := make([]float64, 0, 2*len(run))

	for i, e := range run {
		if len(e.Shape) < 2 {
			return planner.RunMeasure{}, fmt.Errorf("%w: edge %d of the run has %d points", planner.ErrNoLeastRun, i, len(e.Shape))
		}

		for _, p := range []planner.Point{e.Shape[0], e.Shape[len(e.Shape)-1]} {
			ways = append(ways, e.WayID)
			lons = append(lons, p.Longitude)
			lats = append(lats, p.Latitude)
		}
	}

	points, err := s.runPoints(ctx, ways, lons, lats)
	if err != nil {
		return planner.RunMeasure{}, err
	}

	network, err := s.runNetwork(ctx, run[0].WayID)
	if err != nil {
		return planner.RunMeasure{}, err
	}

	return measure(run, points, network)
}

func (s *Store) runPoints(ctx context.Context, ways []int64, lons, lats []float64) ([]runPoint, error) {
	rows, err := s.pool.Query(ctx, runPointsQuery, ways, lons, lats)
	if err != nil {
		return nil, fmt.Errorf("placing the run on the access-marking tables: %w", err)
	}
	defer rows.Close()

	points := make([]runPoint, len(ways))
	seen := make([]bool, len(ways))

	for rows.Next() {
		var (
			index, vertex, segment int32
			back, ahead            float64
		)

		if err := rows.Scan(&index, &vertex, &segment, &back, &ahead); err != nil {
			return nil, fmt.Errorf("reading where the run lies: %w", err)
		}

		if index < 1 || int(index) > len(ways) || seen[index-1] || vertex < 1 || segment < 1 {
			return nil, fmt.Errorf("%w: the tables placed point %d at vertex %d, segment %d", ErrUnchecked, index, vertex, segment)
		}

		seen[index-1] = true
		points[index-1] = runPoint{vertex: int(vertex) - 1, segment: int(segment) - 1, back: back, ahead: ahead}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("placing the run on the access-marking tables: %w", err)
	}

	if i := slices.Index(seen, false); i >= 0 {
		return nil, fmt.Errorf("%w: point %d of the run lies on way %d, which the tables do not hold", planner.ErrNoLeastRun, i, ways[i])
	}

	return points, nil
}

func (s *Store) runNetwork(ctx context.Context, start int64) (map[int64]runWay, error) {
	rows, err := s.pool.Query(ctx, runNetworkQuery, start)
	if err != nil {
		return nil, fmt.Errorf("reading the marked network at the location: %w", err)
	}
	defer rows.Close()

	network := map[int64]runWay{}

	for rows.Next() {
		var (
			id int64
			w  runWay
		)

		if err := rows.Scan(&id, &w.nodes, &w.segments, &w.exits); err != nil {
			return nil, fmt.Errorf("reading a way of the marked network: %w", err)
		}

		network[id] = w
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the marked network at the location: %w", err)
	}

	if _, ok := network[start]; !ok {
		return nil, fmt.Errorf("%w: the location's way %d is not a marked way the tables hold", planner.ErrNoLeastRun, start)
	}

	return network, nil
}

// source is the run's first point in the shortest-path search. OpenStreetMap
// node ids are positive, so no node is it.
const source int64 = -1

type arc struct {
	to  int64
	way int64
	m   float64
}

// measure takes the least run over the network from the run's first point,
// and the run's own length over the same segments. points holds two per edge
// of the run, its first point and its last.
func measure(run []planner.Edge, points []runPoint, network map[int64]runWay) (planner.RunMeasure, error) {
	if len(points) != 2*len(run) {
		return planner.RunMeasure{}, fmt.Errorf("%w: %d points placed for a run of %d edges", ErrUnchecked, len(points), len(run))
	}

	arcs := map[int64][]arc{}
	exits := map[int64]bool{}

	// In the ways' order, so the search breaks a tie the same way every time.
	ids := make([]int64, 0, len(network))
	for id := range network {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	for _, id := range ids {
		w := network[id]
		if len(w.nodes) < 2 || len(w.segments) != len(w.nodes)-1 {
			return planner.RunMeasure{}, fmt.Errorf("%w: way %d holds %d nodes and %d segments", ErrUnchecked, id, len(w.nodes), len(w.segments))
		}

		for i, m := range w.segments {
			if math.IsNaN(m) || math.IsInf(m, 0) || m < 0 {
				return planner.RunMeasure{}, fmt.Errorf("%w: segment %d of way %d measures %v m", ErrUnchecked, i, id, m)
			}

			a, b := w.nodes[i], w.nodes[i+1]
			arcs[a] = append(arcs[a], arc{to: b, way: id, m: m})
			arcs[b] = append(arcs[b], arc{to: a, way: id, m: m})
		}

		for _, n := range w.exits {
			exits[n] = true
		}
	}

	start := network[run[0].WayID]
	first := points[0]

	if first.segment < 0 || first.segment >= len(start.segments) {
		return planner.RunMeasure{}, fmt.Errorf("%w: the run starts on segment %d of a way of %d", ErrUnchecked, first.segment, len(start.segments))
	}

	for _, m := range []float64{first.back, first.ahead} {
		if math.IsNaN(m) || math.IsInf(m, 0) || m < 0 {
			return planner.RunMeasure{}, fmt.Errorf("%w: the run's first point measures %v m from a vertex", ErrUnchecked, m)
		}
	}

	arcs[source] = []arc{
		{to: start.nodes[first.segment], way: run[0].WayID, m: first.back},
		{to: start.nodes[first.segment+1], way: run[0].WayID, m: first.ahead},
	}

	least, leastWays, ok := shortestToExit(arcs, exits)
	if !ok {
		return planner.RunMeasure{}, fmt.Errorf("%w: no unmarked way is reachable over marked ways from way %d", planner.ErrNoLeastRun, run[0].WayID)
	}

	marked, exit, err := travelled(run, points, network, exits)
	if err != nil {
		return planner.RunMeasure{}, err
	}

	return planner.RunMeasure{Marked: marked, Least: least, LeastWays: leastWays, Exit: exit}, nil
}

// travelled is the run's own marked length: the segments it travels, in its
// order, summed in that order from the length its first point lies from the
// vertex it travels to; and the exit node it ends at.
func travelled(run []planner.Edge, points []runPoint, network map[int64]runWay, exits map[int64]bool) (float64, int64, error) {
	var (
		metres  float64
		node    int64 // the node the edge before ended at
		prevWay int64
		prevDir int
	)

	for i, e := range run {
		w, ok := network[e.WayID]
		if !ok {
			return 0, 0, fmt.Errorf("%w: edge %d of the run is along way %d, which the marked network at the location does not hold", planner.ErrNoLeastRun, i, e.WayID)
		}

		if repeats(w.nodes) {
			return 0, 0, fmt.Errorf("%w: edge %d of the run is along way %d, which passes a node twice", planner.ErrNoLeastRun, i, e.WayID)
		}

		end := points[2*i+1].vertex
		if end < 0 || end >= len(w.nodes) {
			return 0, 0, fmt.Errorf("%w: edge %d ends at vertex %d of %d", ErrUnchecked, i, end, len(w.nodes))
		}

		var from, dir int

		if i == 0 {
			// From the first point, along its segment to the vertex the edge
			// travels towards: the shortest path's first step, by the same
			// number.
			seg := points[0].segment

			switch {
			case end > seg:
				metres, from, dir = points[0].ahead, seg+1, 1
			default:
				metres, from, dir = points[0].back, seg, -1
			}
		} else {
			from = points[2*i].vertex
			if from < 0 || from >= len(w.nodes) || w.nodes[from] != node {
				return 0, 0, fmt.Errorf("%w: edge %d does not start at the node edge %d ended at", planner.ErrNoLeastRun, i, i-1)
			}

			switch {
			case end > from:
				dir = 1
			case end < from:
				dir = -1
			}
		}

		if e.WayID == prevWay && dir != 0 && prevDir != 0 && dir != prevDir {
			return 0, 0, fmt.Errorf("%w: the run turns back along way %d at edge %d", planner.ErrNoLeastRun, e.WayID, i)
		}

		for v := from; v != end; v += dir {
			if dir > 0 {
				metres += w.segments[v]
			} else {
				metres += w.segments[v-1]
			}
		}

		if e.WayID != prevWay {
			prevDir = 0
		}

		if dir != 0 {
			prevDir = dir
		}

		node, prevWay = w.nodes[end], e.WayID
	}

	if !exits[node] || !slices.Contains(network[prevWay].exits, node) {
		return 0, 0, fmt.Errorf("%w: the run ends at node %d, which is no exit of way %d to an unmarked way", planner.ErrNoLeastRun, node, prevWay)
	}

	return metres, node, nil
}

func repeats(nodes []int64) bool {
	seen := make(map[int64]bool, len(nodes))
	for _, n := range nodes {
		if seen[n] {
			return true
		}

		seen[n] = true
	}

	return false
}

// shortestToExit is Dijkstra's search from source over arcs, stopping at the
// first exit settled, which is the nearest. It returns that exit's distance
// and the ways of the path to it.
func shortestToExit(arcs map[int64][]arc, exits map[int64]bool) (float64, []int64, bool) {
	type step struct {
		from int64
		way  int64
	}

	dist := map[int64]float64{source: 0}
	prev := map[int64]step{}
	done := map[int64]bool{}
	q := &frontier{{node: source, m: 0}}

	for q.Len() > 0 {
		cur := heap.Pop(q).(reached)
		if done[cur.node] {
			continue
		}

		done[cur.node] = true

		if exits[cur.node] {
			var ways []int64

			for n := cur.node; n != source; n = prev[n].from {
				if w := prev[n].way; !slices.Contains(ways, w) {
					ways = append(ways, w)
				}
			}

			slices.Reverse(ways)

			return cur.m, ways, true
		}

		for _, a := range arcs[cur.node] {
			m := dist[cur.node] + a.m
			if d, seen := dist[a.to]; !seen || m < d {
				dist[a.to] = m
				prev[a.to] = step{from: cur.node, way: a.way}
				heap.Push(q, reached{node: a.to, m: m})
			}
		}
	}

	return 0, nil, false
}

type reached struct {
	node int64
	m    float64
}

// frontier is the search's queue, nearest first; nodes tie on their id, so
// the search is the same on every run.
type frontier []reached

func (f frontier) Len() int { return len(f) }
func (f frontier) Less(i, j int) bool {
	if f[i].m != f[j].m {
		return f[i].m < f[j].m
	}

	return f[i].node < f[j].node
}
func (f frontier) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f *frontier) Push(x any)   { *f = append(*f, x.(reached)) }
func (f *frontier) Pop() any {
	old := *f
	x := old[len(old)-1]
	*f = old[:len(old)-1]

	return x
}
