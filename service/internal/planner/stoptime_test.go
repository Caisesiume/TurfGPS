package planner

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"
	"time"
)

// The time spent at a Turf stop is charged to the journey (`FR-009`, #8). No
// figure here is taken from `CalculationSpecification.md § Stop time`: each
// stop carries an arbitrary time, and each test asserts the relation the
// criteria state between totals the planner computed — which holds whatever
// values that section carries.

// replayedTimes answers each leg asked for, in order, as taking the time the
// n'th answer recorded took, plus added: the same journey's legs, each driven
// for longer by added.
type replayedTimes struct {
	recorded []time.Duration
	added    time.Duration
	n        int
}

func (r *replayedTimes) Route(_ context.Context, from, to Point, _ []Point) (Route, error) {
	d := r.recorded[r.n] + r.added
	r.n++

	return Route{Edges: []Edge{{WayID: 1, Shape: []Point{from, to}}}, Duration: d}, nil
}

// TestAStopOnTheRoadAlreadyDrivenIsChargedItsTime is AC1: an alternative
// whose only stop is a zone on the road its base already drives — its journey
// is that base, so its driving time is unchanged — is charged more additional
// time than the same journey planned without the stop, by the stop's time.
func TestAStopOnTheRoadAlreadyDrivenIsChargedItsTime(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(8, 1))

	for range 200 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		base := plannedThrough(t, randomTimes(r), req)
		stop := Stop{At: base.Legs[r.IntN(len(base.Legs))].Shape[0], Time: someStopTime(r)}

		withStop := Alternative{Journey: base, Stops: []Stop{stop}, Base: base}
		withoutStop := Alternative{Journey: base, Base: base}

		if withStop.Journey.DrivingTime() != withoutStop.Journey.DrivingTime() {
			t.Fatalf("the journey with the stop drives %v, without it %v: the premise is no deviation", withStop.Journey.DrivingTime(), withoutStop.Journey.DrivingTime())
		}

		charged, err := withStop.AdditionalTime()
		if err != nil {
			t.Fatalf("AdditionalTime with a stop of %v on the road driven: %v", stop.Time, err)
		}

		chargedWithout, err := withoutStop.AdditionalTime()
		if err != nil {
			t.Fatalf("AdditionalTime of the same journey without the stop: %v", err)
		}

		if charged <= chargedWithout || charged-chargedWithout != stop.Time {
			t.Fatalf("a stop of %v on the road already driven, %v of driving, was charged %v against %v without it; want it to exceed that by the stop's time, %v",
				stop.Time, base.DrivingTime(), charged, chargedWithout, stop.Time)
		}
	}
}

// TestAStopRequiringADeviationIsChargedBothComponents is AC2: an alternative
// whose stop requires a driving deviation — each of its base's legs driven for
// longer — is charged the changed driving time and the stop's time, both:
// neither alone, and nothing else.
func TestAStopRequiringADeviationIsChargedBothComponents(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(8, 2))

	for range 200 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		baseRouting := randomTimes(r)
		base := plannedThrough(t, baseRouting, req)

		deviated := plannedThrough(t, &replayedTimes{recorded: baseRouting.answered, added: 1 + time.Duration(r.Int64N(int64(time.Hour)))}, req)
		stops := someStops(r)
		alt := Alternative{Journey: deviated, Stops: stops, Base: base}

		changedDriving := deviated.DrivingTime() - base.DrivingTime()
		if changedDriving <= 0 {
			t.Fatalf("the deviated journey drives %v against its base's %v: the premise is a deviation", deviated.DrivingTime(), base.DrivingTime())
		}

		got, err := alt.AdditionalTime()
		if err != nil {
			t.Fatalf("AdditionalTime with %d stops on a deviation: %v", len(stops), err)
		}

		var stopTime time.Duration
		for _, s := range stops {
			stopTime += s.Time
		}

		if want := changedDriving + stopTime; got != want {
			t.Fatalf("an alternative deviating by %v of driving, with %d stops carrying %v, was charged %v; want both, %v",
				changedDriving, len(stops), stopTime, got, want)
		}
	}
}

// TestAStopCarryingNoTimeIsRefused: a stop carrying no time spent at it — the
// Time a Stop nobody priced holds, or less — is refused, never charged as
// though stopping were free: every stop model takes the zone, so none takes
// no time, and a stop costing nothing is the roadside stop `FR-009` forbids.
func TestAStopCarryingNoTimeIsRefused(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(8, 3))

	req := Request{Origin: home, Destination: shops}
	base := plannedThrough(t, randomTimes(r), req)

	for name, stops := range map[string][]Stop{
		"one stop, unpriced":            {{At: shops}},
		"one stop, less than no time":   {{At: shops, Time: -time.Second}},
		"a priced stop, then unpriced":  {{At: home, Time: someStopTime(r)}, {At: shops}},
		"an unpriced stop, then priced": {{At: home}, {At: shops, Time: someStopTime(r)}},
	} {
		got, err := Alternative{Journey: base, Stops: stops, Base: base}.AdditionalTime()
		if !errors.Is(err, ErrUnpricedStop) {
			t.Errorf("%s: gave %v, %v; want an error wrapping ErrUnpricedStop", name, got, err)
		}
	}
}
