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
//
// The duration it is taken from holds the time spent at each Turf stop as well
// as the time driven (`FR-009`, #8): a routing provider returns driving time
// only, so a cost read off its answers alone prices a zone on the road already
// driven at nothing. What a stop's time is, by its access class, is
// `CalculationSpecification.md § Stop time`'s, and is priced before a Stop
// reaches an Alternative; this file charges it and does not price it.

// ErrNotCorresponding is the error AdditionalTime wraps where an alternative's
// Base is not a journey it can be built on: Base does not visit the
// alternative's required locations, in its order, or is not a planned journey
// with a leg between each consecutive pair of them — or the alternative holds
// no Turf stop and its Journey is not its Base.
var ErrNotCorresponding = errors.New("the base is not the journey without Turf stops that the alternative corresponds to")

// ErrUnpricedStop is the error AdditionalTime wraps where a Turf stop the
// alternative holds carries no time spent at it: a Time of zero, which is
// what a Stop nobody priced holds, or less. Every model under
// `CalculationSpecification.md § Stop time` takes the zone, and takeover time
// is a component of each of them, so no stop takes no time; charged as it
// stands, such a stop would be the free roadside stop `FR-009` forbids.
var ErrUnpricedStop = errors.New("a Turf stop carries no time spent at it")

// Stop is a Turf stop a journey alternative holds: a place it stops at to
// capture a zone, the zone it captures, and the time spent there. How the
// place is chosen is #256's.
type Stop struct {
	At Point

	// Zone is the id of the zone the stop captures, as the local zone copy
	// holds it. It is what the alternative's Turf value is derived from
	// (`FR-007`, value.go); the place and the time are not.
	Zone int32

	// Time is the time spent at the stop, beyond the driving its journey's
	// legs carry: the stop time `CalculationSpecification.md § Stop time`
	// prices by the stop's access class, for this journey in its direction.
	// It is priced before the Stop is built, never here.
	Time time.Duration
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

// StopTime returns the time the alternative spends at its Turf stops: the
// sum of each stop's Time.
func (a Alternative) StopTime() time.Duration {
	var total time.Duration

	for _, s := range a.Stops {
		total += s.Time
	}

	return total
}

// Duration returns the alternative's estimated duration: the driving time of
// its journey and the time spent at its Turf stops, the two terms
// `CalculationSpecification.md § Additional journey time` composes it from.
// The second is the one a routing provider does not return (`FR-009`).
func (a Alternative) Duration() time.Duration {
	return a.Journey.DrivingTime() + a.StopTime()
}

// AdditionalTime returns the alternative's cost (`FR-008`): its estimated
// duration, the time spent at its Turf stops included, less that of the
// corresponding journey without Turf stops, its Base. A stop on the road the
// Base already drives changes no driving time and is still charged its time
// (`FR-009`). An alternative holding no Turf stop is the general route offered
// as it is, so its Journey is its Base and it is charged nothing.
//
// An alternative whose Base does not correspond to it is refused with an error
// wrapping ErrNotCorresponding, rather than costed against the wrong origin,
// which no later check could detect (`FR-008`'s `Risk`). So is one holding no
// Turf stop whose Journey is not its Base — re-planned apart from it, along
// other ways or taking another time: charged nothing, the whole additional
// time would read the Base while the journey served is another, and costed
// the difference, it would be charged for a Turf stop it does not hold. And so
// is one holding a stop that carries no time spent at it, with an error
// wrapping ErrUnpricedStop, rather than charged as though stopping were free.
func (a Alternative) AdditionalTime() (time.Duration, error) {
	if err := a.corresponds(); err != nil {
		return 0, err
	}

	if len(a.Stops) == 0 {
		if !a.Journey.same(a.Base) {
			return 0, fmt.Errorf("%w: it holds no Turf stop, so it is the general route as it is, and its journey is not that route", ErrNotCorresponding)
		}

		return 0, nil
	}

	for i, s := range a.Stops {
		if s.Time <= 0 {
			return 0, fmt.Errorf("%w: stop %d of %d carries %v", ErrUnpricedStop, i+1, len(a.Stops), s.Time)
		}
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

// same reports whether j and o are one journey: the same required locations,
// and legs of the same shape, along the same ways, estimated as taking the
// same time.
func (j Journey) same(o Journey) bool {
	return slices.Equal(j.Locations, o.Locations) && slices.EqualFunc(j.Legs, o.Legs, Leg.same)
}

func (l Leg) same(o Leg) bool {
	return l.Duration == o.Duration && slices.Equal(l.Shape, o.Shape) &&
		slices.EqualFunc(l.Edges, o.Edges, func(e, f Edge) bool { return e.WayID == f.WayID && slices.Equal(e.Shape, f.Shape) })
}
