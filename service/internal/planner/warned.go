package planner

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// The warned leg (#271, `FR-129`, `FR-130`, `FR-131`, `RD-067`, `RD-068`).
//
// WHEN A LEG IS WARNED. checkedLeg asks for a leg avoiding what each refused
// answer crossed, as it always has. Where the provider then answers that no
// route keeps off the places (ErrNoRoute), and both ends of the leg are
// locations the user entered, every connection the provider can find crosses
// a marking beyond the run at either end, and the leg is produced warned
// instead of refused (`FR-129`, first criterion). Nothing else opens it: an
// avoidance bound spent, a provider failing otherwise or ignoring what it was
// told, or an end that is a stopping position the system chose leaves the leg
// refused as before (`RD-068`'s first edge, #271 AC2).
//
// WHICH CONNECTION IT TAKES. A warned leg carries no marking a connection
// avoids that crosses no marking the leg does not (`FR-130`, `RD-068`'s second
// edge, the Owner's "least private"). The answers refused on the way there are
// candidates, and the first whose run, at each end the user entered, was
// admitted is taken. Then, for each marking it crosses beyond the run in turn,
// the provider is asked for a connection keeping off that marking and crossing
// nothing the candidate does not:
//
//   - an answer crossing only markings the candidate crosses, less that one,
//     replaces the candidate, and the markings are tried again from the first;
//   - an answer crossing a marking the candidate does not is asked for again,
//     keeping off that marking too, as avoidance does, within the same bound;
//   - ErrNoRoute says no such connection exists, and the marking stays.
//
// Each replacement crosses strictly fewer markings, so the search ends; where
// it ends with none, the leg is produced as a leg not warned (`FR-129`, fourth
// criterion). Anything that proves nothing — the bound spent, an answer
// adding no place to avoid, an answer whose run is not admitted, a marking
// with no place to avoid it by (a marked area, #253), a provider failing
// otherwise — leaves the leg refused: a lost leg, never a warned leg carrying
// a marking something avoids.
//
// WHAT IT NAMES. Every crossing the check finds on the leg taken, the run's
// included (`FR-131`, second criterion), each located on the leg's shape by
// the edges that cross it; see stretches.

// errUnproven is the cause of a search that proved nothing about a marking.
var errUnproven = errors.New("whether the marking can be avoided was not shown")

// warnedLeg produces the warned leg from the answers checkedLeg refused, seen,
// or returns refusal where it cannot vouch for one.
func (p *Planner) warnedLeg(ctx context.Context, ask func([]Point) (answer, error), seen []answer, refusal *CrossingError) (Leg, error) {
	i := slices.IndexFunc(seen, warnable)
	if i < 0 {
		return Leg{}, refusal
	}

	best := seen[i]

	for {
		smaller, err := p.lessMarked(ctx, ask, best)

		switch {
		case errors.Is(err, errUnproven):
			return Leg{}, refusal
		case err != nil:
			return Leg{}, err
		case smaller == nil:
			leg := best.leg()
			leg.Warned = true
			leg.Marked = stretches(best)

			return leg, nil
		}

		best = *smaller

		if len(best.refused) == 0 {
			return best.leg(), nil
		}
	}
}

// warnable reports whether a refused answer may be a warned leg: every
// crossing refusing it lies beyond an admitted run, and each has a place the
// provider can be asked to keep off.
func warnable(a answer) bool {
	return !a.runRefused && !slices.ContainsFunc(a.refused, func(c Crossing) bool { return c.Kind == "area" })
}

// feature names a marking, whichever edges cross it.
type feature struct {
	kind    string
	osmType string
	id      int64
}

func featureOf(c Crossing) feature { return feature{c.Kind, c.OSMType, c.OSMID} }

// marks are the markings refusing a, in the order the leg reaches them.
func marks(a answer) []feature {
	var out []feature

	for _, c := range a.refused {
		if f := featureOf(c); !slices.Contains(out, f) {
			out = append(out, f)
		}
	}

	return out
}

// lessMarked returns an answer crossing a strict subset of the markings best
// crosses beyond its run, or nil where the provider shows that none keeps off
// any of them. An error wrapping errUnproven means the search proved nothing.
func (p *Planner) lessMarked(ctx context.Context, ask func([]Point) (answer, error), best answer) (*answer, error) {
	held := marks(best)[:0] // RED RUN AC4

	for _, x := range held {
		allowed := slices.DeleteFunc(slices.Clone(held), func(f feature) bool { return f == x })

		avoid, err := placesOf(best, func(f feature) bool { return f == x })
		if err != nil {
			return nil, err
		}

		found, err := p.keepingOff(ctx, ask, avoid, allowed)
		if err != nil {
			return nil, fmt.Errorf("%s %s%d: %w", x.kind, x.osmType, x.id, err)
		}

		if found != nil {
			return found, nil
		}
	}

	return nil, nil
}

// keepingOff asks for a connection keeping off avoid and crossing, beyond its
// run, no marking outside allowed. It returns the answer found, or nil where
// the provider answers that no route keeps off the places.
func (p *Planner) keepingOff(ctx context.Context, ask func([]Point) (answer, error), avoid []Point, allowed []feature) (*answer, error) {
	for attempt := 0; attempt <= p.avoidUpTo; attempt++ {
		a, err := ask(avoid)

		var failed *routeFailure

		switch {
		case errors.As(err, &failed):
			if ctx.Err() != nil {
				return nil, failed.err
			}

			if errors.Is(failed.err, ErrNoRoute) {
				return nil, nil
			}

			return nil, fmt.Errorf("%w: the provider answered %w", errUnproven, failed.err)
		case err != nil:
			return nil, err
		}

		if a.runRefused {
			return nil, fmt.Errorf("%w: an answer's run at a location the user entered was not admitted", errUnproven)
		}

		outside := func(f feature) bool { return !slices.Contains(allowed, f) }

		if !slices.ContainsFunc(a.refused, func(c Crossing) bool { return outside(featureOf(c)) }) {
			return &a, nil
		}

		more, err := placesOf(a, outside)
		if err != nil {
			return nil, err
		}

		added := 0

		for _, pl := range more {
			if !slices.Contains(avoid, pl) {
				avoid = append(avoid, pl)
				added++
			}
		}

		if added == 0 {
			return nil, fmt.Errorf("%w: the provider crossed again what it was asked to keep off", errUnproven)
		}
	}

	return nil, fmt.Errorf("%w: the avoidance bound of %d was spent", errUnproven, p.avoidUpTo)
}

// placesOf is where to ask the provider to keep off the markings of a's
// refused crossings that which picks: a point on each edge along a marked way,
// as avoidance asks for it, and each marked barrier's position. A marked area
// has no such place, and asking to keep off one proves nothing.
func placesOf(a answer, which func(feature) bool) ([]Point, error) {
	var out []Point

	for _, c := range a.refused {
		if !which(featureOf(c)) {
			continue
		}

		var pl Point

		switch c.Kind {
		case "way":
			pl = onEdge(a.edges[c.Edge])
		case "barrier":
			pl = c.At
		default:
			return nil, fmt.Errorf("%w: a marked %s has no place to keep off", errUnproven, c.Kind)
		}

		if !slices.Contains(out, pl) {
			out = append(out, pl)
		}
	}

	return out, nil
}

// stretches names every marking the check finds on the leg, located on its
// shape (`FR-131`). Each is the span of the consecutive edges crossing one
// feature, as indices into the shape joinEdges builds, where edge k begins
// at the point edge k-1 ends at; a feature the leg crosses again after
// leaving it is named again where it does.
func stretches(a answer) []Marked {
	start := make([]int, len(a.edges)+1)
	for k, e := range a.edges {
		start[k+1] = start[k] + len(e.Shape) - 1
	}

	var out []Marked

	for k := range a.edges {
		for _, c := range a.refused { // RED RUN AC6, AC7
			if c.Edge != k || c.Kind == "barrier" { // RED RUN AC6
				continue
			}

			f := featureOf(c)
			same := func(m Marked) bool { return feature{m.Kind, m.OSMType, m.OSMID} == f }

			switch {
			// Found twice on this edge: named once.
			case slices.ContainsFunc(out, func(m Marked) bool { return same(m) && m.To == start[k+1] }):
			// Found on the edge before too, and named there: the stretch runs
			// on over this one.
			case slices.ContainsFunc(out, func(m Marked) bool { return same(m) && m.To == start[k] }):
				j := slices.IndexFunc(out, func(m Marked) bool { return same(m) && m.To == start[k] })
				out[j].To = start[k+1]
			default:
				out = append(out, Marked{Kind: c.Kind, OSMType: c.OSMType, OSMID: c.OSMID, From: start[k], To: start[k+1], At: c.At})
			}
		}
	}

	return out
}
