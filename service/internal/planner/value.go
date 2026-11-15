package planner

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

// A JOURNEY ALTERNATIVE'S TURF VALUE IS DERIVED FROM THE ZONES IT CAPTURES,
// NEVER FROM THE ROUTE IT TRAVELS (`FR-007`, #9). It reads the zone each of
// its Turf stops captures and nothing else the alternative holds: not its
// journey, its legs, their ways or their driving time, not where it stops, not
// the time spent there, and not its base. Those are cost, which is cost.go's;
// once one of them reaches value, the same minutes count as a benefit on one
// side and a penalty on the other (`FR-007`'s `Risk`), and difficulty stays on
// the cost side under `CalculationSpecification.md § Proposed form: value per
// minute`.
//
// What one captured zone is worth is not decided here. The weight it carries
// comes from the user's stated preferences (`FR-011`, #11), translated by
// `CalculationSpecification.md § Zone value`; this file takes it as a
// ZoneValue and states no weight, no curve and no base value of its own.

// ErrInvalidZoneValue is the error TurfValue wraps where a ZoneValue answers a
// captured zone with a value that is not a finite number. NaN equals nothing,
// itself included, so a NaN summed in would make two alternatives capturing
// the same zones compare unequal; an infinity summed against its opposite is
// NaN again.
var ErrInvalidZoneValue = errors.New("a captured zone's value is not a finite number")

// ZoneValue answers what the zone with the given id is worth to the user the
// journey is planned for, or an error where it cannot say. It is supplied by
// the caller, never computed in this package.
type ZoneValue func(zone int32) (float64, error)

// Captured returns the ids of the zones the alternative captures: the zone of
// each of its Turf stops, each once, in ascending order of id. An alternative
// holding no Turf stop captures none.
//
// It is a set, and its order is the ids' rather than the stops': two
// alternatives capturing the same zones over different roads may visit them in
// different orders, and stop at one of them twice.
func (a Alternative) Captured() []int32 {
	zones := make([]int32, 0, len(a.Stops))
	for _, s := range a.Stops {
		zones = append(zones, s.Zone)
	}

	slices.Sort(zones)

	return slices.Compact(zones)
}

// TurfValue returns the alternative's Turf value (`FR-007`): the sum, over the
// zones it captures, of what value answers each is worth. Each zone is counted
// once, and value is asked about each once. An alternative capturing no zone
// is worth nothing, and value is not asked at all.
//
// The sum is taken in Captured's order, never the order the stops are visited
// in. Floating-point addition is not associative, so summed in visiting order
// two alternatives capturing the same zones in different orders could differ
// in their last bit and compare unequal (`FR-007`'s acceptance).
//
// An error value returns for a zone is returned wrapped, naming the zone, and
// a value that is not a finite number with an error wrapping
// ErrInvalidZoneValue; neither is summed as though it were a worth.
func (a Alternative) TurfValue(value ZoneValue) (float64, error) {
	var total float64

	for _, stop := range a.Stops {
		zone := stop.Zone
		v, err := value(zone)
		if err != nil {
			return 0, fmt.Errorf("the value of captured zone %d: %w", zone, err)
		}

		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, fmt.Errorf("%w: zone %d is valued at %v", ErrInvalidZoneValue, zone, v)
		}

		total += v
	}

	return total, nil
}
