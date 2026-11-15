package config

import (
	"fmt"
	"strconv"
	"strings"
)

// The avoidance bound (#248): how many times the planner asks the routing
// engine again for a leg refused for crossing an access marking, avoiding what
// it crossed (`Architecture.md § Avoiding, not only refusing`).
//
// IT HAS NO DEFAULT. No document gives its value (`Architecture.md § What is
// not decided`: "how many re-requests the planner makes, are not chosen"), so
// a figure compiled in here would be an unexplained literal. While it is unset
// a refused leg is never asked for again: every leg that crosses a marking in
// transit is refused, which loses legs and serves none that cross.
//
// IT IS NOT A SAFETY CONSTANT IN EITHER DIRECTION. Every answer is checked
// again, so no value of it lets a crossing leg through; a larger one finds
// more avoiding legs at the cost of more engine calls per leg (`Architecture.md
// § The call budget`), a smaller one loses more. It arrives with the origin
// that accounts for it all the same, as every configured constant does.
const (
	EnvAccessAvoidanceAttempts       = "TURFGPS_ACCESS_AVOIDANCE_ATTEMPTS"
	EnvAccessAvoidanceAttemptsOrigin = "TURFGPS_ACCESS_AVOIDANCE_ATTEMPTS_ORIGIN"
)

// AccessAvoidance is the avoidance bound and where it comes from.
type AccessAvoidance struct {
	// Attempts is how many times a refused leg is asked for again: zero or
	// more.
	Attempts int

	// Origin is where the value comes from, as the deployment states it.
	Origin string
}

// LoadAccessAvoidance reads the avoidance bound.
//
// It returns (nil, nil) when neither variable is set: no leg is asked for
// again. A value without its origin, or an origin without a value, is refused
// as a partial configuration, and so is a value that is not a whole number at
// or above zero.
func LoadAccessAvoidance(lookup LookupFunc) (*AccessAvoidance, error) {
	raw, haveValue := lookup(EnvAccessAvoidanceAttempts)
	origin, haveOrigin := lookup(EnvAccessAvoidanceAttemptsOrigin)

	raw, origin = strings.TrimSpace(raw), strings.TrimSpace(origin)
	haveValue, haveOrigin = haveValue && raw != "", haveOrigin && origin != ""

	switch {
	case !haveValue && !haveOrigin:
		return nil, nil
	case !haveValue:
		return nil, fmt.Errorf("the avoidance bound is partially configured: %s is set and %s is not — set both, or neither to ask no leg again",
			EnvAccessAvoidanceAttemptsOrigin, EnvAccessAvoidanceAttempts)
	case !haveOrigin:
		return nil, fmt.Errorf("%s is set without %s: a configured constant is taken with its documented origin, never without one",
			EnvAccessAvoidanceAttempts, EnvAccessAvoidanceAttemptsOrigin)
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return nil, fmt.Errorf("%s is %q, want a whole number of attempts at or above zero", EnvAccessAvoidanceAttempts, raw)
	}

	return &AccessAvoidance{Attempts: n, Origin: origin}, nil
}
