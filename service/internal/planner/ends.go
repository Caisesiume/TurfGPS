package planner

import (
	"errors"
	"fmt"
	"math"
)

// The end tolerance (#250, `CalculationSpecification.md § Enforcement
// constants that do not yet exist`, `Architecture.md § The route check`).
//
// WHY THE ENDS ARE CHECKED. The provider places a leg's ends on its own graph,
// so a leg's first and last vertices are generally near its locations and not
// at them. Nothing bounded how near: #246 measured an answer whose first
// vertex lay 32,053 m from the origin it was asked for, and the run at a
// location the user entered (`run.go`, #248) reads the location's way from
// that placement. So every answer is held to the end tolerance before
// anything else is read of it — the first answer, each answer to avoidance,
// and each connection the search for a warned leg asks for, since every one
// reaches the planner through checkedLeg's ask and is read by read. An answer
// whose first vertex lies farther from the location it leaves, or whose last
// lies farther from the location it reaches, than the tolerance refuses the
// leg: it is never served, warned or not.
//
// WHILE NO TOLERANCE IS SET, NO LEG IS SERVED. The tolerance is the
// deployment's, taken from configuration with its origin
// (`config.LoadEndTolerance`), by the Owner's ruling on #250; a planner built
// without EndsWithin refuses every answer, so nothing is planned.
//
// THE DISTANCE NEVER UNDERSTATES. It is measured by endDistance, which bounds
// the WGS 84 geodesic from above, so a leg whose end lies beyond the tolerance
// on the spheroid PostGIS measures on is refused: the strict direction. The
// bound has slack of up to about 1% away from the poles, far above the
// rounding of the arithmetic, and is tight only at them.

// ErrNoEndTolerance is the error a leg is refused with while no end tolerance
// is configured: its ends were not held to one, so it is not served.
var ErrNoEndTolerance = errors.New("no end tolerance is configured, so no leg's ends are checked (TURFGPS_LEG_END_TOLERANCE_METRES)")

// ErrEndBeyondTolerance is the error a leg is refused with where the provider
// placed an end of it farther from its location than the end tolerance.
var ErrEndBeyondTolerance = errors.New("a leg's end lies farther from its location than the end tolerance")

// EndsWithin sets the end tolerance, in metres: how far an answer's first
// vertex may lie from the location the leg leaves, and its last from the
// location it reaches. Without it every leg is refused with ErrNoEndTolerance.
// The tolerance is an enforcement constant whose strict direction is downward;
// zero, the strictest, serves only a leg placed exactly at its locations. A
// value that is negative or not a finite number is refused.
func EndsWithin(metres float64) Option {
	return func(p *Planner) error {
		if math.IsNaN(metres) || math.IsInf(metres, 0) || metres < 0 {
			return fmt.Errorf("an end tolerance of %v m is not a finite number of metres at or above zero", metres)
		}

		p.endTolerance = &metres

		return nil
	}
}

// endsHeld refuses an answer whose shape does not start and end within the
// end tolerance of the locations from and to, or any answer while no
// tolerance is set.
func (p *Planner) endsHeld(i int, shape []Point, from, to Point) error {
	if p.endTolerance == nil {
		return fmt.Errorf("leg %d: %w", i, ErrNoEndTolerance)
	}

	tolerance := *p.endTolerance

	for _, end := range []struct {
		name     string
		at, want Point
	}{
		{"first vertex", shape[0], from},
		{"last vertex", shape[len(shape)-1], to},
	} {
		if d := endDistance(end.at, end.want); false && !(d <= tolerance) {
			return fmt.Errorf("leg %d: %w: its %s lies %.1f m from (%v, %v), beyond the end tolerance of %v m",
				i, ErrEndBeyondTolerance, end.name, d, end.want.Latitude, end.want.Longitude, tolerance)
		}
	}

	return nil
}

// endDistance returns a distance in metres between a and b that is never less
// than the WGS 84 geodesic between them.
//
// It is the central angle between them on the unit sphere, taking their
// geodetic latitude and longitude as spherical ones (the haversine formula),
// times the ellipsoid's polar radius of curvature c = a/(1−f). On the
// ellipsoid the length element is √(M(φ)²·dφ² + (N(φ)·cos φ)²·dλ²), and the
// meridian radius M(φ) and the prime-vertical radius N(φ) are each at most c
// at every latitude, reaching it at the poles; the unit sphere's element at
// the same (φ, λ) is √(dφ² + cos²φ·dλ²). So the great circle between them,
// carried onto the ellipsoid at the same coordinates, is no longer than c
// times its angle, and the geodesic is no longer than that curve. It
// overstates the geodesic by at most c/M(0), about 1%, the strict direction
// for a tolerance whose strict direction is downward. The same bound is
// argued for the barrier tolerance's floor in
// `service/internal/config/accessmarkings.go`.
func endDistance(a, b Point) float64 {
	const radians = math.Pi / 180

	phi1, phi2 := a.Latitude*radians, b.Latitude*radians
	dPhi, dLambda := phi2-phi1, (b.Longitude-a.Longitude)*radians

	h := math.Sin(dPhi/2)*math.Sin(dPhi/2) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)
	h = math.Min(1, math.Max(0, h))

	return wgs84PolarRadiusOfCurvatureMetres * 2 * math.Asin(math.Sqrt(h))
}

// The WGS 84 ellipsoid's defining semi-major axis and inverse flattening
// (NGA, "World Geodetic System 1984", NGA.STND.0036_1.0.0_WGS84, Table 3.1):
// the spheroid of SRID 4326, on which PostGIS measures geography, and the one
// `service/internal/config/accessmarkings.go` derives the barrier tolerance's
// floor on.
const (
	wgs84SemiMajorAxisMetres          = 6378137.0
	wgs84InverseFlattening            = 298.257223563
	wgs84PolarRadiusOfCurvatureMetres = wgs84SemiMajorAxisMetres / (1 - 1/wgs84InverseFlattening)
)
