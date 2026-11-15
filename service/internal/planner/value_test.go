package planner

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
	"time"
)

// A journey alternative's Turf value (`FR-007`, #9). No figure here is taken
// from `CalculationSpecification.md § Zone value`: each zone is worth an
// arbitrary value of the test's choosing, and each test asserts the relation
// the criteria state between values TurfValue computed — which holds whatever
// weights that section, or `FR-011`'s preferences, give a zone.

// randomRoads answers every leg along a random road of its own: through
// between none and seven random points, along ways numbered from firstWay,
// taking a random time of up to a day.
type randomRoads struct {
	r        *rand.Rand
	firstWay int64
}

func (f *randomRoads) Route(_ context.Context, from, to Point, _ []Point) (Route, error) {
	shape := []Point{from}
	for range f.r.IntN(8) {
		shape = append(shape, randomPoint(f.r))
	}

	edges := edgesOf(append(shape, to))
	for i := range edges {
		edges[i].WayID += f.firstWay
	}

	return Route{Edges: edges, Duration: time.Duration(f.r.Int64N(int64(24 * time.Hour)))}, nil
}

// someZones is one to six distinct zone ids, each worth a random value of up
// to a thousand — a fraction, so that the order the values are summed in can
// change the last bit of their sum.
func someZones(r *rand.Rand) map[int32]float64 {
	zones := map[int32]float64{}
	for len(zones) < 1+r.IntN(6) {
		zones[r.Int32()] = r.Float64() * 1000
	}

	return zones
}

// valuedFrom is a ZoneValue answering each zone with its value in zones, and
// failing the test for a zone it does not hold.
func valuedFrom(t *testing.T, zones map[int32]float64) ZoneValue {
	t.Helper()

	return func(zone int32) (float64, error) {
		v, ok := zones[zone]
		if !ok {
			t.Fatalf("the value of zone %d was asked, which no stop captures", zone)
		}

		return v, nil
	}
}

// stopsCapturing is a stop at each of zones, at a random place, carrying a
// random time, in a random order, and one in three times a second stop at one
// of them.
func stopsCapturing(r *rand.Rand, zones map[int32]float64) []Stop {
	stops := make([]Stop, 0, len(zones)+1)
	for zone := range zones {
		stops = append(stops, Stop{At: randomPoint(r), Zone: zone, Time: someStopTime(r)})
	}

	if r.IntN(3) == 0 {
		again := stops[r.IntN(len(stops))]
		stops = append(stops, Stop{At: randomPoint(r), Zone: again.Zone, Time: someStopTime(r)})
	}

	r.Shuffle(len(stops), func(i, j int) { stops[i], stops[j] = stops[j], stops[i] })

	return stops
}

// TestTheSameCapturedZonesAreWorthTheSameOverDifferentRoads is AC1: two
// alternatives capturing the same set of zones while travelling different
// roads — other ways, other driving times, other places to stop, other times
// spent stopping, the zones visited in another order and one perhaps stopped
// at twice — are given Turf values that compare equal.
func TestTheSameCapturedZonesAreWorthTheSameOverDifferentRoads(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(9, 1))

	for range 200 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		zones := someZones(r)
		value := valuedFrom(t, zones)

		one := plannedThrough(t, &randomRoads{r: r, firstWay: 1}, req)
		other := plannedThrough(t, &randomRoads{r: r, firstWay: 1_000_000}, req)

		if one.same(other) {
			t.Fatalf("the two journeys travel the same roads: the premise is different ones")
		}

		a := Alternative{Journey: one, Stops: stopsCapturing(r, zones), Base: one}
		b := Alternative{Journey: other, Stops: stopsCapturing(r, zones), Base: other}

		if !slices.Equal(a.Captured(), b.Captured()) {
			t.Fatalf("the alternatives capture %v and %v: the premise is the same set", a.Captured(), b.Captured())
		}

		va, err := a.TurfValue(value)
		if err != nil {
			t.Fatalf("TurfValue of an alternative capturing %d zones: %v", len(zones), err)
		}

		vb, err := b.TurfValue(value)
		if err != nil {
			t.Fatalf("TurfValue of an alternative capturing %d zones: %v", len(zones), err)
		}

		if va != vb {
			t.Fatalf("two alternatives capturing zones %v, visited as %v and as %v over different roads, are worth %v and %v; want them equal",
				a.Captured(), zonesOf(a.Stops), zonesOf(b.Stops), va, vb)
		}
	}
}

// TestDrivingTimeAloneDoesNotChangeValue is AC2: an alternative whose driving
// time changes — each of its legs driven for longer or shorter, along the
// same ways — while the zones it captures do not, is worth what it was worth
// before.
func TestDrivingTimeAloneDoesNotChangeValue(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(9, 2))

	for range 200 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		zones := someZones(r)
		value := valuedFrom(t, zones)
		stops := stopsCapturing(r, zones)

		recorded := randomTimes(r)
		before := plannedThrough(t, recorded, req)

		// The same legs, each driven for up to an hour longer than recorded,
		// or, where every leg was recorded as taking long enough, for up to
		// half the shortest leg's time less — never less than no time.
		added := 1 + time.Duration(r.Int64N(int64(time.Hour)))
		if shortest := slices.Min(recorded.answered); shortest >= 2 && r.IntN(2) == 0 {
			added = -(1 + time.Duration(r.Int64N(int64(shortest/2))))
		}

		after := plannedThrough(t, &replayedTimes{recorded: recorded.answered, added: added}, req)

		if after.DrivingTime() == before.DrivingTime() {
			t.Fatalf("the journey drives %v before and after: the premise is a change in driving time", before.DrivingTime())
		}

		was, err := Alternative{Journey: before, Stops: stops, Base: before}.TurfValue(value)
		if err != nil {
			t.Fatalf("TurfValue before the change: %v", err)
		}

		is, err := Alternative{Journey: after, Stops: stops, Base: before}.TurfValue(value)
		if err != nil {
			t.Fatalf("TurfValue after the change: %v", err)
		}

		if is != was {
			t.Fatalf("an alternative capturing zones %v was worth %v driving %v and is worth %v driving %v; want the value unchanged",
				zonesOf(stops), was, before.DrivingTime(), is, after.DrivingTime())
		}
	}
}

// TestTheValueIsWhatTheCapturedZonesAreWorthEachOnce: the value is derived
// from the zones and is not merely blind to the route — it is the sum of what
// each zone the alternative captures is worth, each counted once however many
// stops capture it, with each zone asked about once. Every value here is a
// whole number, so the sum is exact in any order and the test's own sum is
// the one to compare against.
func TestTheValueIsWhatTheCapturedZonesAreWorthEachOnce(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(9, 3))
	journey := plannedThrough(t, randomTimes(r), Request{Origin: home, Destination: shops})

	for range 200 {
		zones := map[int32]float64{}
		for len(zones) < 1+r.IntN(6) {
			zones[r.Int32()] = float64(r.IntN(1000))
		}

		var want float64
		for _, v := range zones {
			want += v
		}

		asked := map[int32]int{}
		value := func(zone int32) (float64, error) {
			asked[zone]++

			return valuedFrom(t, zones)(zone)
		}

		stops := stopsCapturing(r, zones)

		got, err := Alternative{Journey: journey, Stops: stops, Base: journey}.TurfValue(value)
		if err != nil {
			t.Fatalf("TurfValue: %v", err)
		}

		if got != want {
			t.Fatalf("an alternative stopping at zones %v, worth %v, is worth %v; want %v", zonesOf(stops), zones, got, want)
		}

		for zone := range zones {
			if asked[zone] != 1 {
				t.Fatalf("zone %d, captured, was asked about %d times; want once", zone, asked[zone])
			}
		}
	}

	got, err := Alternative{Journey: journey, Base: journey}.TurfValue(func(zone int32) (float64, error) {
		t.Fatalf("the value of zone %d was asked of an alternative holding no Turf stop", zone)

		return 0, nil
	})
	if err != nil || got != 0 {
		t.Fatalf("an alternative holding no Turf stop is worth %v, %v; want 0, nil", got, err)
	}
}

// TestAZoneThatCannotBeValuedIsNotSummed: a zone the ZoneValue cannot value,
// or values at a number that is not finite, gives an error and no value — an
// error naming the zone and wrapping the ZoneValue's own, or one wrapping
// ErrInvalidZoneValue — rather than a sum that, holding a NaN, would equal
// nothing.
func TestAZoneThatCannotBeValuedIsNotSummed(t *testing.T) {
	t.Parallel()

	stops := []Stop{{At: home, Zone: 71, Time: time.Minute}, {At: shops, Zone: 72, Time: time.Minute}}
	alt := Alternative{Stops: stops}

	unknown := errors.New("no such zone")

	got, err := alt.TurfValue(func(zone int32) (float64, error) {
		if zone == 72 {
			return 0, unknown
		}

		return 1, nil
	})
	if !errors.Is(err, unknown) || got != 0 {
		t.Errorf("a zone that cannot be valued gave %v, %v; want 0 and an error wrapping the value's own", got, err)
	}

	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		got, err := alt.TurfValue(func(int32) (float64, error) { return bad, nil })
		if !errors.Is(err, ErrInvalidZoneValue) || got != 0 {
			t.Errorf("zones valued at %v gave %v, %v; want 0 and an error wrapping ErrInvalidZoneValue", bad, got, err)
		}
	}
}

// zonesOf is the zone each stop captures, in the order visited.
func zonesOf(stops []Stop) string {
	zones := make([]int32, len(stops))
	for i, s := range stops {
		zones[i] = s.Zone
	}

	return fmt.Sprint(zones)
}
