package planner

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
	"time"
)

// A journey alternative's cost (`FR-008`, #7). No figure here is taken from
// `CalculationSpecification.md § Additional journey time`, and no duration is
// derived from its formula: the fakes answer arbitrary times, and each test
// asserts a relation between the durations the planner computed from them —
// the cost against the alternative's duration and its base's, as the
// criteria state it.

// timedRouting answers every leg as a straight line along one way, taking a
// time of took's choosing, and records the time of each answer.
type timedRouting struct {
	took     func(from, to Point) time.Duration
	answered []time.Duration
}

func (r *timedRouting) Route(_ context.Context, from, to Point, _ []Point) (Route, error) {
	d := r.took(from, to)
	r.answered = append(r.answered, d)

	return Route{Edges: []Edge{{WayID: 1, Shape: []Point{from, to}}}, Duration: d}, nil
}

// sum is the total of the times answered.
func sum(ds []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range ds {
		total += d
	}

	return total
}

// randomTimes answers every leg taking a random time, of up to a day.
func randomTimes(r *rand.Rand) *timedRouting {
	return &timedRouting{took: func(_, _ Point) time.Duration {
		return time.Duration(r.Int64N(int64(24 * time.Hour)))
	}}
}

// plannedThrough plans req through routing, with nothing marked, and fails the
// test where it cannot.
func plannedThrough(t *testing.T, routing RoutingProvider, req Request) Journey {
	t.Helper()

	p, err := New(routing, copyHeld, clearAccess{}, clearAccess{})
	if err != nil {
		t.Fatal(err)
	}

	j, err := p.Plan(t.Context(), req)
	if err != nil {
		t.Fatalf("Plan(%+v): %v", req, err)
	}

	return j
}

// randomJourneyRequest is a request of one to four legs between distinct
// random points.
func randomJourneyRequest(r *rand.Rand) Request {
	req := Request{Origin: randomPoint(r), Destination: randomPoint(r)}
	for range r.IntN(4) {
		req.Intermediates = append(req.Intermediates, randomPoint(r))
	}

	return req
}

func someStops(r *rand.Rand) []Stop {
	stops := make([]Stop, 1+r.IntN(3))
	for i := range stops {
		stops[i] = Stop{At: randomPoint(r)}
	}

	return stops
}

// TestAnAlternativeIsChargedItsDurationLessItsBase is AC1: an alternative
// holding at least one Turf stop costs its estimated duration less the
// estimated duration of the journey it is built on, each the sum of the
// times the routing provider answered for the legs that journey was planned
// from — whichever of the two is the longer, and whatever the journey's
// shape.
func TestAnAlternativeIsChargedItsDurationLessItsBase(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(7, 8))

	for range 200 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		baseRouting, altRouting := randomTimes(r), randomTimes(r)
		base := plannedThrough(t, baseRouting, req)
		alt := Alternative{Journey: plannedThrough(t, altRouting, req), Stops: someStops(r), Base: base}

		got, err := alt.AdditionalTime()
		if err != nil {
			t.Fatalf("AdditionalTime of an alternative with %d stops: %v", len(alt.Stops), err)
		}

		if want := sum(altRouting.answered) - sum(baseRouting.answered); got != want {
			t.Fatalf("an alternative whose legs were answered as taking %v in all, built on a journey whose legs took %v, was charged %v; want the difference, %v",
				sum(altRouting.answered), sum(baseRouting.answered), got, want)
		}

		if want := alt.Duration() - base.DrivingTime(); got != want {
			t.Fatalf("an alternative of duration %v on a base of %v was charged %v, want %v", alt.Duration(), base.DrivingTime(), got, want)
		}
	}
}

// TestTheCostMovesWithTheBaseAndNothingElse is AC1 from the other side: two
// alternatives identical but for the journey each is built on are charged
// differently by exactly the difference between those journeys' durations, so
// the cost is measured from the base and from no other origin.
func TestTheCostMovesWithTheBaseAndNothingElse(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(7, 9))

	for range 100 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		journey := plannedThrough(t, randomTimes(r), req)
		stops := someStops(r)
		oneBase, otherBase := plannedThrough(t, randomTimes(r), req), plannedThrough(t, randomTimes(r), req)

		onOne, err := Alternative{Journey: journey, Stops: stops, Base: oneBase}.AdditionalTime()
		if err != nil {
			t.Fatal(err)
		}

		onOther, err := Alternative{Journey: journey, Stops: stops, Base: otherBase}.AdditionalTime()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := onOne-onOther, otherBase.DrivingTime()-oneBase.DrivingTime(); got != want {
			t.Fatalf("one alternative on bases of %v and %v was charged %v and %v, which differ by %v; want them to differ by the bases', %v",
				oneBase.DrivingTime(), otherBase.DrivingTime(), onOne, onOther, got, want)
		}
	}
}

// TestAnAlternativeWithNoTurfStopIsChargedNothing is AC2: an alternative
// holding no Turf stop — the general route offered as it is, so its journey is
// its base — is attributed no additional time.
func TestAnAlternativeWithNoTurfStopIsChargedNothing(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(7, 10))

	for range 100 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		base := plannedThrough(t, randomTimes(r), req)

		for name, alt := range map[string]Alternative{
			"its own base, no stops":          {Journey: base, Base: base},
			"its own base, an empty stop set": {Journey: base, Stops: []Stop{}, Base: base},
		} {
			got, err := alt.AdditionalTime()
			if err != nil {
				t.Fatalf("%s: AdditionalTime: %v", name, err)
			}

			if got != 0 {
				t.Fatalf("%s: an alternative holding no Turf stop, of duration %v on a base of %v, was charged %v; want no additional time",
					name, alt.Duration(), base.DrivingTime(), got)
			}
		}
	}
}

// TestAStoplessAlternativeApartFromItsBaseIsRefused (safety verdict on #289):
// an alternative holding no Turf stop whose journey is not its base — planned
// apart from it, taking another time, or along other ways in the same time, as
// a re-plan around a closed way would be — is refused, not charged nothing:
// charged nothing, the whole additional time would read the base while the
// journey served is another, which no later check could detect.
func TestAStoplessAlternativeApartFromItsBaseIsRefused(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(7, 12))

	for range 100 {
		req := randomJourneyRequest(r)
		if returnsWithNowhereBetween(req) {
			continue
		}

		base := plannedThrough(t, randomTimes(r), req)
		apart := plannedThrough(t, randomTimes(r), req)

		if apart.DrivingTime() == base.DrivingTime() {
			t.Fatalf("the journey planned apart took %v, as its base did: the premise is two times", apart.DrivingTime())
		}

		otherWays := Journey{Locations: base.Locations, Legs: slices.Clone(base.Legs)}
		last := &otherWays.Legs[len(otherWays.Legs)-1]
		last.Edges = []Edge{{WayID: last.Edges[0].WayID + 1, Shape: last.Edges[0].Shape}}

		if otherWays.DrivingTime() != base.DrivingTime() {
			t.Fatalf("the journey along other ways took %v, its base %v: the premise is one time", otherWays.DrivingTime(), base.DrivingTime())
		}

		for name, journey := range map[string]Journey{
			"planned apart, taking another time": apart,
			"along other ways, in the same time": otherWays,
		} {
			for _, stops := range [][]Stop{nil, {}} {
				got, err := Alternative{Journey: journey, Stops: stops, Base: base}.AdditionalTime()
				if !errors.Is(err, ErrNotCorresponding) {
					t.Fatalf("%s: an alternative holding no Turf stop, of duration %v on a base of %v that is not its journey, gave %v, %v; want an error wrapping ErrNotCorresponding",
						name, journey.DrivingTime(), base.DrivingTime(), got, err)
				}
			}
		}
	}
}

// TestAnAlternativeIsCostedOnlyAgainstTheJourneyItIsBuiltOn: a base through
// other required locations, or one with no leg between them, is not the
// corresponding journey without Turf stops, and the alternative is refused
// rather than costed against it — stops or none.
func TestAnAlternativeIsCostedOnlyAgainstTheJourneyItIsBuiltOn(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(7, 11))

	a, b, c := Point{Latitude: 57.70, Longitude: 11.96}, Point{Latitude: 57.80, Longitude: 12.00}, Point{Latitude: 57.75, Longitude: 11.90}

	journey := plannedThrough(t, randomTimes(r), Request{Origin: a, Destination: b})
	elsewhere := plannedThrough(t, randomTimes(r), Request{Origin: a, Destination: c})
	through := plannedThrough(t, randomTimes(r), Request{Origin: a, Intermediates: []Point{c}, Destination: b})
	reversed := plannedThrough(t, randomTimes(r), Request{Origin: b, Destination: a})

	for name, base := range map[string]Journey{
		"to another destination":              elsewhere,
		"through an intermediate it does not": through,
		"the other way":                       reversed,
		"with its locations and no legs":      {Locations: journey.Locations},
		"with nothing at all":                 {},
		"with a leg fewer than its locations": {Locations: through.Locations, Legs: through.Legs[:1]},
	} {
		for _, stops := range [][]Stop{nil, {{At: c}}} {
			got, err := Alternative{Journey: journey, Stops: stops, Base: base}.AdditionalTime()
			if !errors.Is(err, ErrNotCorresponding) {
				t.Errorf("a base %s, with %d stops, gave %v, %v; want an error wrapping ErrNotCorresponding", name, len(stops), got, err)
			}
		}
	}
}

// avoidedScript answers the n'th request for a leg with answers[n], the last
// repeated: a straight line along the ways named, taking the time given.
type avoidedScript struct {
	answers []timedWays
	n       int
}

type timedWays struct {
	ways []int64
	took time.Duration
}

func (s *avoidedScript) Route(_ context.Context, from, to Point, _ []Point) (Route, error) {
	a := s.answers[min(s.n, len(s.answers)-1)]
	s.n++

	return Route{Edges: lineOf(from, to, a.ways), Duration: a.took}, nil
}

// TestAnAvoidedLegIsCostedAtTheAnswerThatPassed is the lead's routing note on
// #7 (sentinel N4 on #248): a leg asked for again around a marking it crossed
// in transit carries the time of the answer that passed the check, never the
// time of the one refused before it — here the refused answer is the quicker
// — so a cost, and the ceiling when it lands (#44), read the leg that is
// served.
func TestAnAvoidedLegIsCostedAtTheAnswerThatPassed(t *testing.T) {
	t.Parallel()

	const marked = 9

	refusedTook, passedTook := 7*time.Minute, 19*time.Minute

	routing := &avoidedScript{answers: []timedWays{
		{ways: []int64{1, marked, 2}, took: refusedTook},
		{ways: []int64{1, 3, 2}, took: passedTook},
	}}

	p := plannerOn(t, routing, &tableFake{marked: map[int64]bool{marked: true}}, AvoidUpTo(1))

	req := Request{Origin: home, Destination: shops}

	journey, err := p.Plan(t.Context(), req)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if routing.n != 2 {
		t.Fatalf("the leg was asked for %d times, want twice: the premise is a leg avoided once", routing.n)
	}

	if got := journey.Legs[0].Duration; got != passedTook {
		t.Fatalf("the avoided leg carries %v; want the time of the answer that passed, %v, not the refused answer's, %v", got, passedTook, refusedTook)
	}

	base := plannedThrough(t, &timedRouting{took: func(_, _ Point) time.Duration { return refusedTook }}, req)

	got, err := Alternative{Journey: journey, Stops: []Stop{{At: shops}}, Base: base}.AdditionalTime()
	if err != nil {
		t.Fatal(err)
	}

	if want := journey.DrivingTime() - base.DrivingTime(); got != want || got == 0 {
		t.Fatalf("an alternative carrying the avoided leg, on a base answered as the refused leg's time, was charged %v; want %v, the avoided leg's time less the base's", got, want)
	}
}

// TestANegativeEstimateIsRefused: a provider answering a leg as taking less
// than no time breaks the port's contract, and the leg is not produced — it
// would subtract from every cost it entered.
func TestANegativeEstimateIsRefused(t *testing.T) {
	t.Parallel()

	p, err := New(&timedRouting{took: func(_, _ Point) time.Duration { return -time.Second }}, copyHeld, clearAccess{}, clearAccess{})
	if err != nil {
		t.Fatal(err)
	}

	if j, err := p.Plan(t.Context(), Request{Origin: home, Destination: shops}); err == nil {
		t.Errorf("a leg answered as taking -1s was planned as %+v, want it refused", j)
	}
}
