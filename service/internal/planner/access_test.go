package planner

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"testing"
)

// markedWay is the way id the fakes below treat as marked.
const markedWay = 4242

// markingsOn reports a crossing for every edge along markedWay, records every
// set of edges it was asked to check, and reports its extract from ids, one per
// read, repeating the last.
type markingsOn struct {
	mu      sync.Mutex
	checked [][]Edge
	ids     []string
	reads   int
	idErr   error
	err     error
}

// Run finds no least run: a run along markedWay is never admitted here, which
// run_test.go's fakes vary.
func (m *markingsOn) Run(context.Context, []Edge) (RunMeasure, error) {
	return RunMeasure{}, ErrNoLeastRun
}

func (m *markingsOn) ExtractSHA256(context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.idErr != nil {
		return "", m.idErr
	}

	if len(m.ids) == 0 {
		return testExtract, nil
	}

	id := m.ids[min(m.reads, len(m.ids)-1)]
	m.reads++

	return id, nil
}

func (m *markingsOn) Crossings(_ context.Context, edges []Edge) ([]Crossing, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checked = append(m.checked, slices.Clone(edges))

	if m.err != nil {
		return nil, m.err
	}

	var out []Crossing

	for i, e := range edges {
		if e.WayID == markedWay {
			out = append(out, Crossing{Edge: i, Kind: "way", OSMType: "w", OSMID: markedWay})
		}
	}

	return out, nil
}

// routingData reports a fixed identity, or an error.
type routingData struct {
	id  string
	err error
}

func (d routingData) ExtractSHA256(context.Context) (string, error) { return d.id, d.err }

// alongMarkedWay answers every request with a route whose middle edge runs
// along markedWay, between two edges along ways not marked.
type alongMarkedWay struct{ asked int }

func (a *alongMarkedWay) Route(_ context.Context, from, to Point, _ []Point) ([]Edge, error) {
	a.asked++

	mid := Point{Latitude: (from.Latitude + to.Latitude) / 2, Longitude: (from.Longitude + to.Longitude) / 2}
	mid2 := Point{Latitude: mid.Latitude, Longitude: (mid.Longitude + to.Longitude) / 2}

	return []Edge{
		{WayID: 1, Shape: []Point{from, mid}},
		{WayID: markedWay, Shape: []Point{mid, mid2}},
		{WayID: 2, Shape: []Point{mid2, to}},
	}, nil
}

// TestNoRouteReachesAJourneyUnchecked is #255 AC4 at the planner: with a
// routing provider whose every route runs along a marked way, no journey of
// any shape — no intermediate destination, or up to five — is planned, and
// every route the provider answered was handed to the check before anything
// else was done with it.
func TestNoRouteReachesAJourneyUnchecked(t *testing.T) {
	t.Parallel()

	const seed, cases = 3, 500

	r := rand.New(rand.NewPCG(seed, seed))

	for c := range cases {
		routing := &alongMarkedWay{}
		markings := &markingsOn{}

		p, err := New(routing, copyHeld, markings, routingData{id: testExtract})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := randomRequest(r, 0, 5)

		journey, err := p.Plan(t.Context(), req)
		if err == nil || journey.Legs != nil {
			t.Fatalf("seed %d case %d: Plan returned %d legs and %v, want no leg from a provider whose every route runs along a marked way", seed, c, len(journey.Legs), err)
		}

		var crossing *CrossingError
		if !errors.As(err, &crossing) || !errors.Is(err, ErrCrossesMarking) {
			t.Fatalf("seed %d case %d: Plan refused with %v, want a CrossingError", seed, c, err)
		}

		if routing.asked == 0 || len(markings.checked) != routing.asked {
			t.Fatalf("seed %d case %d: the provider answered %d routes and %d were checked, want every route answered checked", seed, c, routing.asked, len(markings.checked))
		}
	}
}

// markedOnLastLeg answers every leg along ways not marked except the last
// request it is told of, which runs along markedWay.
type markedOnLastLeg struct {
	last  Point
	asked int
}

func (m *markedOnLastLeg) Route(_ context.Context, from, to Point, _ []Point) ([]Edge, error) {
	m.asked++

	way := int64(m.asked)
	if to == m.last {
		way = markedWay
	}

	return []Edge{{WayID: way, Shape: []Point{from, to}}}, nil
}

// TestEveryLegIsCheckedNotOnlyTheFirst is AC4's other half: a marking on the
// journey's last leg is found too, after every earlier leg was checked, so a
// planner that checked only some legs fails here.
func TestEveryLegIsCheckedNotOnlyTheFirst(t *testing.T) {
	t.Parallel()

	req := Request{
		Origin:        Point{Latitude: 57.70, Longitude: 11.96},
		Intermediates: []Point{{Latitude: 57.71, Longitude: 11.97}, {Latitude: 57.72, Longitude: 11.98}},
		Destination:   Point{Latitude: 57.73, Longitude: 11.99},
	}

	routing := &markedOnLastLeg{last: req.Destination}
	markings := &markingsOn{}

	p, err := New(routing, copyHeld, markings, routingData{id: testExtract})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	journey, err := p.Plan(t.Context(), req)

	var crossing *CrossingError
	if !errors.As(err, &crossing) || crossing.Leg != 2 || journey.Legs != nil {
		t.Fatalf("Plan gave %d legs and %v, want the third leg refused and no journey", len(journey.Legs), err)
	}

	if routing.asked != 3 || len(markings.checked) != 3 {
		t.Fatalf("%d legs routed and %d checked, want all three of each", routing.asked, len(markings.checked))
	}
}

// TestARefusalNamesWhatItFoundAndWhere is AC2's naming half at the planner: a
// refused leg's error carries the edges and, for each marking, the edge it lies
// on and the feature, so #248 can tell a run at a location from transit.
func TestARefusalNamesWhatItFoundAndWhere(t *testing.T) {
	t.Parallel()

	p, err := New(&alongMarkedWay{}, copyHeld, &markingsOn{}, routingData{id: testExtract})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = p.Plan(t.Context(), Request{Origin: Point{Latitude: 57.70, Longitude: 11.96}, Destination: Point{Latitude: 57.71, Longitude: 11.97}})

	var crossing *CrossingError
	if !errors.As(err, &crossing) {
		t.Fatalf("Plan refused with %v, want a CrossingError", err)
	}

	if len(crossing.Crossings) != 1 || crossing.Crossings[0].Edge != 1 || crossing.Crossings[0].OSMID != markedWay || len(crossing.Edges) != 3 {
		t.Fatalf("the refusal carries %+v over %d edges, want way %d at edge 1 of 3", crossing.Crossings, len(crossing.Edges), markedWay)
	}

	if msg := err.Error(); !strings.Contains(msg, "way w4242 at edge 1") {
		t.Errorf("the refusal says %q, want it to name the way and the edge", msg)
	}
}

// TestNoJourneyWithoutOneExtract is #255 AC5 at the planner: markings or
// routing data whose identity cannot be read, is empty, differs from the
// other's, or changes while the journey is planned, plan nothing. Where the
// identity is known to be wrong before any leg, no leg is routed either.
func TestNoJourneyWithoutOneExtract(t *testing.T) {
	t.Parallel()

	other := strings.Repeat("e", 64)
	errUnread := errors.New("unreadable")

	for _, tc := range []struct {
		name       string
		markings   *markingsOn
		data       routingData
		want       error
		wantRouted bool
	}{
		{name: "markings with no current load", markings: &markingsOn{idErr: errUnread}, data: routingData{id: testExtract}, want: ErrAccessUnchecked},
		{name: "routing data whose extract is unknown", markings: &markingsOn{}, data: routingData{err: errUnread}, want: ErrAccessUnchecked},
		{name: "an empty identity", markings: &markingsOn{ids: []string{""}}, data: routingData{id: ""}, want: ErrAccessUnchecked},
		{name: "markings from another extract", markings: &markingsOn{ids: []string{other}}, data: routingData{id: testExtract}, want: ErrExtractMismatch},
		{name: "markings reloaded while planning", markings: &markingsOn{ids: []string{testExtract, other}}, data: routingData{id: testExtract}, want: ErrExtractMismatch, wantRouted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			routing := straightLegs()

			p, err := New(routing, copyHeld, tc.markings, tc.data)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			journey, err := p.Plan(t.Context(), Request{Origin: gateOrigin, Intermediates: []Point{gateVia}, Destination: gateDestination})
			if !errors.Is(err, tc.want) || journey.Legs != nil {
				t.Fatalf("Plan gave %d legs and %v, want no journey and an error wrapping %v", len(journey.Legs), err, tc.want)
			}

			if routed := len(routing.asked) > 0; routed != tc.wantRouted {
				t.Errorf("legs routed: %v, want %v", routing.asked, tc.wantRouted)
			}
		})
	}
}

// edgesRouting answers every leg with the same edges.
type edgesRouting struct{ edges []Edge }

func (e edgesRouting) Route(context.Context, Point, Point, []Point) ([]Edge, error) {
	return e.edges, nil
}

// TestALegTheCheckCannotReadIsNotProduced holds every provider to the contract
// RoutingProvider states: a leg with no edges, an edge with no way, an edge of
// one point, a gap between edges, or a check that fails, is refused as
// unchecked rather than produced.
func TestALegTheCheckCannotReadIsNotProduced(t *testing.T) {
	t.Parallel()

	a, b, c := Point{Latitude: 57.70, Longitude: 11.96}, Point{Latitude: 57.71, Longitude: 11.97}, Point{Latitude: 57.72, Longitude: 11.98}

	for _, tc := range []struct {
		name     string
		edges    []Edge
		checkErr error
	}{
		{name: "no edges"},
		{name: "an edge with no way", edges: []Edge{{WayID: 0, Shape: []Point{a, b}}}},
		{name: "an edge of one point", edges: []Edge{{WayID: 1, Shape: []Point{a}}}},
		{name: "a gap between edges", edges: []Edge{{WayID: 1, Shape: []Point{a, b}}, {WayID: 2, Shape: []Point{c, a}}}},
		{name: "an edge off the globe", edges: []Edge{{WayID: 1, Shape: []Point{a, {Latitude: 91}}}}},
		{name: "a check that fails", edges: []Edge{{WayID: 1, Shape: []Point{a, b}}}, checkErr: errors.New("store unreachable")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := New(edgesRouting{edges: tc.edges}, copyHeld, &markingsOn{err: tc.checkErr}, routingData{id: testExtract})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			journey, err := p.Plan(t.Context(), Request{Origin: a, Destination: c})
			if !errors.Is(err, ErrAccessUnchecked) || journey.Legs != nil {
				t.Fatalf("Plan gave %d legs and %v, want no journey and an error wrapping ErrAccessUnchecked", len(journey.Legs), err)
			}
		})
	}
}

// TestNewRefusesNoMarkingsAndNoRoutingData: a planner that cannot ask whether
// a leg crosses a marking, or which extract the routing data is from, is not
// built.
func TestNewRefusesNoMarkingsAndNoRoutingData(t *testing.T) {
	t.Parallel()

	if _, err := New(straightLegs(), copyHeld, nil, routingData{}); err == nil {
		t.Error("New with no access markings returned no error")
	}

	if _, err := New(straightLegs(), copyHeld, &markingsOn{}, nil); err == nil {
		t.Error("New with no routing data returned no error")
	}
}
