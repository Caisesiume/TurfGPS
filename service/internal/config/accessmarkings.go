package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// The access-marking check's configuration (#255): the barrier tolerance of
// `Architecture.md § The route check`, steps 3 and 5 — how near a marked
// barrier must lie to a leg's stretch of shape to count as passed.
//
// IT HAS NO DEFAULT, AND NONE MAY BE ADDED. It is an enforcement constant
// (`CalculationSpecification.md § Conventions`) whose value no document gives
// (`Architecture.md § What is not decided`): a figure compiled in here would be
// the unexplained literal those sections forbid, on a safety path. While it is
// unset, the service serves no car leg: the store that checks legs refuses
// every one (`internal/accessstore`).
//
// ITS STRICT DIRECTION IS UPWARD, and this block does not enforce one. A larger
// tolerance counts more barriers as passed and refuses more legs; a smaller one
// misses barriers, and zero misses every barrier a shape does not pass exactly
// through. Which value is safe is the measurement that section describes, not
// a bound this package could hold; what it holds is that a value arrives with
// the origin that justifies it, so a figure in the environment is never one
// nobody can account for.
const (
	EnvAccessBarrierTolerance       = "TURFGPS_ACCESS_BARRIER_TOLERANCE_METRES"
	EnvAccessBarrierToleranceOrigin = "TURFGPS_ACCESS_BARRIER_TOLERANCE_ORIGIN"
)

// AccessMarkings is the configuration the access-marking check needs.
type AccessMarkings struct {
	// BarrierToleranceMetres is the barrier tolerance, in metres: finite and
	// not negative.
	BarrierToleranceMetres float64

	// BarrierToleranceOrigin is where the value comes from — the measurement,
	// or the record that set it — as the deployment states it.
	BarrierToleranceOrigin string
}

// LoadAccessMarkings reads the access-marking check's configuration.
//
// It returns (nil, nil) when neither variable is set, which is a service that
// checks no leg and so serves none. A tolerance without its origin, or an
// origin without a tolerance, is refused as a partial configuration, and so is
// a tolerance that is not a finite number of metres at or above zero.
func LoadAccessMarkings(lookup LookupFunc) (*AccessMarkings, error) {
	raw, haveValue := lookup(EnvAccessBarrierTolerance)
	origin, haveOrigin := lookup(EnvAccessBarrierToleranceOrigin)

	raw, origin = strings.TrimSpace(raw), strings.TrimSpace(origin)
	haveValue, haveOrigin = haveValue && raw != "", haveOrigin && origin != ""

	switch {
	case !haveValue && !haveOrigin:
		return nil, nil
	case !haveValue:
		return nil, fmt.Errorf("the barrier tolerance is partially configured: %s is set and %s is not — the origin names a value, so set both, or neither to serve no car leg",
			EnvAccessBarrierToleranceOrigin, EnvAccessBarrierTolerance)
	case !haveOrigin:
		return nil, fmt.Errorf("%s is set without %s: an enforcement constant is taken with its documented origin, never without one",
			EnvAccessBarrierTolerance, EnvAccessBarrierToleranceOrigin)
	}

	metres, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(metres) || math.IsInf(metres, 0) || metres < 0 {
		return nil, fmt.Errorf("%s is %q, want a finite number of metres at or above zero", EnvAccessBarrierTolerance, raw)
	}

	return &AccessMarkings{BarrierToleranceMetres: metres, BarrierToleranceOrigin: origin}, nil
}
