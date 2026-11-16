package planner

import (
	"context"
	"errors"
	"math"
	"slices"
)

// The run at a location the user entered, and transit everywhere else on a leg
// (#248, `FR-125`, `FR-126`, `Architecture.md § The route check`, step 4).
//
// THE READING. At an end of the leg that is a location the user entered, the
// run is the leg's edges from that end along marked ways, up to the first edge
// along a way not marked. It is admitted only where all of this holds:
//
//   - the leg meets an unmarked way after it: a leg that never leaves marked
//     ways has no run, since the run joins the location to the first unmarked
//     way (`FR-126`, `RD-057`);
//   - no edge of it crosses a marked area, which is #253's to read (`FR-128`)
//     and so refuses until it does;
//   - its marked length, measured along the geometry the tables hold, is no
//     more than the least that joins the location to an unmarked way, measured
//     the same way (AccessMarkings.Run). A least that is undefined admits
//     nothing.
//
// An admitted run admits the marked ways it runs along and the marked barriers
// on them, and one barrier more: the one at the node the run ends at, where it
// meets the unmarked way (RunMeasure.Exit) — a gate at the end of a private
// road — which the run's last edge and the first edge after it are both found
// passing. It is the same barrier, passed once, on the run's own way. A barrier
// the first edge after the run passes anywhere else is in transit, though it
// lie on the run's way and within the tolerance of its stretch: a way that
// meets the road twice puts its second meeting there. Every other crossing is
// in transit and refuses the leg, and so does every crossing of a run not
// admitted.

// legEnds says which ends of a leg are locations the user entered. A stopping
// position the system chose is not one (`RD-052`, `RD-061`), and #256 passes
// false for it.
type legEnds struct {
	fromEntered bool
	toEntered   bool
}

// refusing returns the crossings that refuse the leg, the places to ask the
// provider to avoid in a new answer, and whether any refused crossing lies on
// the run at a location the user entered, which was then not admitted. No
// crossing refusing is a leg admitted. An error means the leg could not be
// read and is not served.
func (p *Planner) refusing(ctx context.Context, edges []Edge, crossings []Crossing, ends legEnds) ([]Crossing, []place, bool, error) {
	if len(crossings) == 0 {
		return nil, nil, false, nil
	}

	admitted := make([]bool, len(crossings))

	// An end's run is the edges from it; at the leg's far end that is its
	// edges reversed, so one reading serves both.
	var (
		places []place
		run    = [2][]int{}
	)

	for side, entered := range []bool{ends.fromEntered, ends.toEntered} {
		if !entered {
			continue
		}

		order := edgeOrder(len(edges), side == 1)

		k := 0
		for k < len(order) && hasKind(crossings, order[k], "way") {
			k++
		}

		if k == 0 {
			continue
		}

		run[side] = order[:k]

		if k == len(order) || slices.ContainsFunc(order[:k], func(e int) bool { return hasKind(crossings, e, "area") }) {
			continue
		}

		measure, err := p.markings.Run(ctx, travelled(edges, order[:k], side == 1))

		switch {
		case errors.Is(err, ErrNoLeastRun):
			continue
		case err != nil:
			return nil, nil, false, err
		case !measurable(measure.Marked) || !measurable(measure.Least):
			// A length that is not one is no bound.
			continue
		case measure.Marked > measure.Least:
			// Ask again keeping off the ways of the run the least does not
			// use; never the location's own way, where every run starts.
			for _, e := range order[1:k] {
				if w := edges[e].WayID; w != edges[order[0]].WayID && !slices.Contains(measure.LeastWays, w) {
					places = append(places, place{at: onEdge(edges[e]), ways: []int64{w}})
				}
			}

			continue
		}

		last, after := order[k-1], order[k]

		for j, c := range crossings {
			switch {
			case slices.Contains(order[:k], c.Edge) && c.Kind != "area":
				admitted[j] = true
			case c.Edge == after && c.Kind == "barrier" && measure.Exit != 0 && c.OSMID == measure.Exit && passes(crossings, last, c.OSMID):
				admitted[j] = true
			}
		}
	}

	var (
		refused    []Crossing
		runRefused bool
	)

	for j, c := range crossings {
		if admitted[j] {
			continue
		}

		refused = append(refused, c)

		// The stretch a run was not admitted along is never avoided whole: its
		// own ways are the location's way out, and the longer-than-least case
		// above names what to avoid on it.
		if slices.Contains(run[0], c.Edge) || slices.Contains(run[1], c.Edge) {
			runRefused = true

			continue
		}

		switch c.Kind {
		case "way":
			places = append(places, place{at: onEdge(edges[c.Edge]), ways: []int64{edges[c.Edge].WayID}})
		case "barrier":
			places = append(places, place{at: c.At, ways: barrierWays(edges, crossings, c.OSMID)})
		}
	}

	return refused, places, runRefused, nil
}

// edgeOrder is the leg's edge indices from one end: from its start, or from
// its far end backwards.
func edgeOrder(n int, fromFarEnd bool) []int {
	order := make([]int, n)
	for i := range order {
		if fromFarEnd {
			order[i] = n - 1 - i
		} else {
			order[i] = i
		}
	}

	return order
}

// travelled is the run's edges in the order travelled away from the location:
// at the far end, each stretch reversed.
func travelled(edges []Edge, order []int, fromFarEnd bool) []Edge {
	out := make([]Edge, len(order))

	for i, e := range order {
		shape := slices.Clone(edges[e].Shape)
		if fromFarEnd {
			slices.Reverse(shape)
		}

		out[i] = Edge{WayID: edges[e].WayID, Shape: shape}
	}

	return out
}

// measurable reports whether m is a length: finite and not negative.
func measurable(m float64) bool { return !math.IsNaN(m) && !math.IsInf(m, 0) && m >= 0 }

func hasKind(crossings []Crossing, edge int, kind string) bool {
	return slices.ContainsFunc(crossings, func(c Crossing) bool { return c.Edge == edge && c.Kind == kind })
}

// passes reports whether edge is found passing the barrier node.
func passes(crossings []Crossing, edge int, node int64) bool {
	return slices.ContainsFunc(crossings, func(c Crossing) bool {
		return c.Edge == edge && c.Kind == "barrier" && c.OSMID == node
	})
}

// onEdge is the point the provider is asked to avoid for an edge: the middle
// of its longest segment, which lies on the edge's own way and as far from the
// segment's ends as any point of it. It is not thereby at no junction. A
// provider may map a point near a junction to the junction and keep a route
// off every road there — Valhalla does within its node snap tolerance, so on
// an edge whose longest segment is shorter than twice that, the point is
// mapped to an end of the edge and the roads at that end are kept off with it
// (`internal/valhalla`'s Excluded cites where). What a provider keeps off for
// the point is therefore never assumed: a no-route answer is believed only
// where the provider names the edge's way alone (noRouteShown). The middle is
// the mean of the segment's ends, which lies on the segment to within far less
// than any road's width at a segment's length.
func onEdge(e Edge) Point {
	best, at := -1.0, 0

	for i := range len(e.Shape) - 1 {
		a, b := e.Shape[i], e.Shape[i+1]
		if d := (a.Latitude-b.Latitude)*(a.Latitude-b.Latitude) + (a.Longitude-b.Longitude)*(a.Longitude-b.Longitude); d > best {
			best, at = d, i
		}
	}

	a, b := e.Shape[at], e.Shape[at+1]

	return Point{Latitude: (a.Latitude + b.Latitude) / 2, Longitude: (a.Longitude + b.Longitude) / 2}
}
