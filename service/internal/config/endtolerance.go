package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// The end tolerance (#250): how far a routed leg's ends may lie from the
// locations it was asked for — `CalculationSpecification.md § Enforcement
// constants that do not yet exist`, where it is defined, and
// `Architecture.md § What is not decided`.
//
// IT HAS NO DEFAULT, AND NONE MAY BE ADDED. It is an enforcement constant whose
// value no document gives, and the Owner ruled on #250 that it is set, with its
// origin, at deployment, as the barrier tolerance is: a figure compiled in here
// would be the unexplained literal `CalculationSpecification.md § Conventions`
// forbids, on a safety path. While it is unset, no leg is served, so nothing
// is planned (`internal/planner`, EndsWithin).
//
// ITS STRICT DIRECTION IS DOWNWARD. A smaller tolerance refuses more legs; a
// larger one serves legs whose ends the engine placed farther from the
// locations, which is what #246 measured at 32,053 m with no tolerance at all.
// Zero is admitted, as the strictest value: it serves only a leg the engine
// placed exactly at its locations.
const (
	EnvEndTolerance       = "TURFGPS_LEG_END_TOLERANCE_METRES"
	EnvEndToleranceOrigin = "TURFGPS_LEG_END_TOLERANCE_ORIGIN"
)

// EndTolerance is the end tolerance and where it comes from.
type EndTolerance struct {
	// Metres is the end tolerance: finite and at or above zero.
	Metres float64

	// Origin is where the value comes from — the measurement, or the record
	// that set it — as the deployment states it.
	Origin string
}

// LoadEndTolerance reads the end tolerance.
//
// It returns (nil, nil) when neither variable is set, which is a service that
// serves no leg. A tolerance without its origin, or an origin without a
// tolerance, is refused as a partial configuration; so is a tolerance that is
// not a finite number of metres at or above zero. A refusal stops the service
// at start.
func LoadEndTolerance(lookup LookupFunc) (*EndTolerance, error) {
	raw, haveValue := lookup(EnvEndTolerance)
	origin, haveOrigin := lookup(EnvEndToleranceOrigin)

	raw, origin = strings.TrimSpace(raw), strings.TrimSpace(origin)
	haveValue, haveOrigin = haveValue && raw != "", haveOrigin && origin != ""

	switch {
	case !haveValue && !haveOrigin:
		return nil, nil
	case !haveValue:
		return nil, fmt.Errorf("the end tolerance is partially configured: %s is set and %s is not — the origin names a value, so set both, or neither to serve no leg",
			EnvEndToleranceOrigin, EnvEndTolerance)
	case !haveOrigin:
		return nil, fmt.Errorf("%s is set without %s: an enforcement constant is taken with its documented origin, never without one",
			EnvEndTolerance, EnvEndToleranceOrigin)
	}

	metres, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(metres) || math.IsInf(metres, 0) || metres < 0 {
		return nil, fmt.Errorf("%s is %q, want a finite number of metres at or above zero", EnvEndTolerance, raw)
	}

	return &EndTolerance{Metres: metres, Origin: origin}, nil
}
