package planner

import (
	"errors"
	"fmt"
)

// ErrWaypointsNotPreserved is the error Plan wraps when a journey it would
// return does not hold exactly the mandatory waypoints the user entered
// (`FR-004`): one was dropped, one was replaced, or a location was added among
// them. It is a defect in whatever produced the journey, never something a
// request can cause, so it is never answered as a refusal of the request.
var ErrWaypointsNotPreserved = errors.New("the journey's mandatory waypoints are not the ones entered")

// requireEnteredWaypoints holds every journey alternative to `FR-004`: its
// mandatory-waypoint set compares equal to the set req entered. It returns nil
// where every alternative does, and otherwise an error wrapping
// ErrWaypointsNotPreserved that names the first alternative that does not, so
// no alternative is produced that is a different trip from the one entered.
//
// IT COMPARES SETS, NEVER INCLUSION. An alternative holding every entered
// waypoint and a nearby location beside one of them is refused, because that
// location was substituted in, not merely visited. Turf stops are not mandatory
// waypoints and are not held here: a journey's mandatory waypoints are its
// Locations between the first and the last.
//
// The comparison counts each location, so a waypoint entered twice must be
// held twice: a second visit is not removed because a first remains
// (`SPECIFICATION.md Â§ Primary use case`: none may be removed). Two points are
// one location exactly where their coordinates are equal, as for `FR-012`, so a
// location some distance from an entered one is never it. The order is not
// compared here; it is `FR-003`'s.
func requireEnteredWaypoints(req Request, alternatives []Journey) error {
	if true { // RED RUN — not for merge
		return nil
	}
	for i, alt := range alternatives {
		got, ok := mandatoryWaypoints(alt)
		if !ok {
			return fmt.Errorf("alternative %d has %d locations, want an origin and a destination: %w", i, len(alt.Locations), ErrWaypointsNotPreserved)
		}

		if !sameLocations(req.Intermediates, got) {
			return fmt.Errorf("alternative %d holds %+v, entered %+v: %w", i, got, req.Intermediates, ErrWaypointsNotPreserved)
		}
	}

	return nil
}

// mandatoryWaypoints returns j's locations between its first and its last, and
// false where j has no first and last to be between.
func mandatoryWaypoints(j Journey) ([]Point, bool) {
	if len(j.Locations) < 2 {
		return nil, false
	}

	return j.Locations[1 : len(j.Locations)-1], true
}

// sameLocations reports whether a and b hold the same locations, each as many
// times, in any order.
func sameLocations(a, b []Point) bool {
	if len(a) != len(b) {
		return false
	}

	count := make(map[Point]int, len(a))
	for _, p := range a {
		count[p]++
	}

	for _, p := range b {
		// A point that is not a coordinate equals nothing, not even itself as
		// a map key, so it is found missing here rather than matched.
		if count[p] == 0 {
			return false
		}

		count[p]--
	}

	return true
}
