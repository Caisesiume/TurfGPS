package planner

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// A JOURNEY ALTERNATIVE'S COST IS THE TIME IT ADDS TO THE JOURNEY WITHOUT ITS
// TURF STOPS (`FR-008`, #7). The arithmetic has one home,
// `CalculationSpecification.md § Additional journey time`, and is not restated
// here; what this file holds is which two durations it is taken between.
//
// The journey without Turf stops is the general route the alternative is built
// on, never another — not the fastest conventional route, and not any other
// alternative's — because `SPECIFICATION.md § General route alternatives`
// makes each general route the baseline its Turf-enhanced variants are
// evaluated against. Alternative carries that route beside it as Base, so the
// baseline travels with the alternative rather than being looked up. The
// whole additional time the stated limit and the ceiling test, which adds the
// general route's own deviation from the fastest one, is `FR-061`'s quantity,
// built from this one, and is not computed here.
//
// Only the figures of legs that passed the access check enter it: every Leg
// Plan returns carries the estimate of the answer checkedLeg accepted, so an
// avoided leg is costed at the avoided answer's time, never the refused
// answer's.

// ErrNotCorresponding is the error AdditionalTime wraps where an alternative's
// Base is not a journey it can be built on: Base does not visit the
// alternative's required locations, in its order, or is not a planned journey
// with a leg between each consecutive pair of them.
var ErrNotCorresponding = errors.New("the base is not the journey without Turf stops that the alternative corresponds to")

// Stop is a Turf stop a journey alternative holds: a place it stops at to
// capture a zone. It names where, and nothing yet of what stopping there
// costs, which is `FR-009`'s (#8), or of how the place is chosen, which is
// #256's.
type Stop struct {
	At Point
}

// Alternative is a journey alternative: the journey as it would be driven,
// the Turf stops it holds, and the general route it is built on.
type Alternative struct {
	// Journey is the alternative as it would be driven, through its Turf
	// stops. Its Locations are the required locations the user entered, as
	// Base's are; a stop is not one of them.
	Journey Journey

	// Stops are the Turf stops the alternative holds, in the order visited.
	// A general route offered as it is holds none.
	Stops []Stop

	// Base is the corresponding journey without Turf stops: the general route
	// the alternative is built on.
	Base Journey
}

// DrivingTime returns the time a car is estimated to take to drive the
// journey's legs: the sum of each leg's Duration.
func (j Journey) DrivingTime() time.Duration {
	var total time.Duration

	for _, l := range j.Legs {
		total += l.Duration
	}

	return total
}

// Duration returns the alternative's estimated duration. Today that is the
// driving time of its journey alone: the time spent at each stop, which a
// routing provider does not return, is the other term of the duration
// `CalculationSpecification.md § Additional journey time` composes, and it
// lands with `FR-009` (#8) here, so that AdditionalTime charges it with no
// change of its own.
func (a Alternative) Duration() time.Duration {
	return a.Journey.DrivingTime()
}

// AdditionalTime returns the alternative's cost (`FR-008`): its estimated
// duration less that of the corresponding journey without Turf stops, its
// Base. An alternative holding no Turf stop is charged nothing, whatever its
// journey's driving time and its Base's are: whatever separates them is not
// time added by a Turf stop.
//
// An alternative whose Base does not correspond to it is refused with an error
// wrapping ErrNotCorresponding, rather than costed against the wrong origin,
// which no later check could detect (`FR-008`'s `Risk`).
func (a Alternative) AdditionalTime() (time.Duration, error) {
	if err := a.corresponds(); err != nil {
		return 0, err
	}

	return a.Duration() - a.Base.DrivingTime(), nil
}

// corresponds reports whether Base is a journey the alternative can be built
// on: a planned journey, with a leg between each consecutive pair of its
// locations, through the alternative's own required locations in its order.
// A Base with no legs would cost every minute driven as added.
func (a Alternative) corresponds() error {
	base := a.Base

	if len(base.Locations) < 2 || len(base.Legs) != len(base.Locations)-1 {
		return fmt.Errorf("%w: it has %d legs between %d locations", ErrNotCorresponding, len(base.Legs), len(base.Locations))
	}

	if !slices.Equal(base.Locations, a.Journey.Locations) {
		return fmt.Errorf("%w: it visits other required locations than the alternative", ErrNotCorresponding)
	}

	return nil
}
