//go:build valhalla

package accessstore_test

// #248's acceptance criteria, against the committed extract: the tables loaded
// from it, the tiles a real Valhalla serves built from the same file, and the
// extract read independently as OPL, as integration_test.go (#255) does, whose
// helpers these share.
//
// EACH TEST ESTABLISHES ITS PREMISE FROM THE EXTRACT FIRST: that the marked way
// or barrier is there with its marking and no other access key, that the ends
// lie on ways the reading leaves unmarked, that the avoiding connection the
// case names is in the file, and — through the engine itself, asked with
// `ignore_access`, so that what a marking costs or forbids is set aside — that
// the quickest drivable connection between the ends runs along the marked way
// or past the marked node. A run against an extract that no longer holds the
// case fails there rather than passing vacuously.
//
// Most cases are synthetic (`testdata/README.md` § The synthetic cases): each
// criterion names a road layout — a marking with a way around it, a home on a
// private road with a second way out — that the extract held no instance of,
// so each is a small road network of its own in the Kattegat, joined to
// nothing.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// testAvoidance is this test's avoidance bound, not the service's, which the
// deployment configures (`config.LoadAccessAvoidance`). No case here needs
// more than one or two answers avoided; the bound only stops a provider that
// ignores what it is told.
const testAvoidance = 4

const syntheticNote248 = "synthetic, not from OpenStreetMap: TurfGPS #248 test case"

// The synthetic cases' ids and places, as the generator wrote them: case c's
// nodes are syntheticBase + 10000 + 100c + k and its ways syntheticBase + 1000
// + 20c + k, and it lies along 57.050 + 0.006c ° N.
func synNode(c, k int) int64 { return syntheticBase + 10000 + int64(c)*100 + int64(k) }
func synWay(c, k int) int64  { return syntheticBase + 1000 + int64(c)*20 + int64(k) }
func synLat(c int) float64   { return 57.050 + float64(c)*0.006 }

// transitCase is a road of three ways east — the from-way, the case way, the
// to-way — between points on the first and the last, and, where detour is
// set, an unmarked way round the case way to the north. The marking is on the
// case way, or on the node in its middle.
type transitCase struct {
	c        int
	name     string
	wayTags  map[string]string
	nodeTags map[string]string
	backward bool // the case way's nodes run east to west, so the route drives it backward
	detour   bool
}

func (tc transitCase) from() planner.Point {
	return planner.Point{Latitude: synLat(tc.c), Longitude: 11.7010}
}
func (tc transitCase) to() planner.Point {
	return planner.Point{Latitude: synLat(tc.c), Longitude: 11.7080}
}
func (tc transitCase) caseWay() int64 { return synWay(tc.c, 2) }
func (tc transitCase) middle() int64  { return synNode(tc.c, 3) }

var (
	ac2Cases = []transitCase{
		{c: 0, name: "access=private", wayTags: map[string]string{"access": "private"}, detour: true},
		{c: 1, name: "access=destination", wayTags: map[string]string{"access": "destination"}, detour: true},
		{c: 2, name: "access=customers", wayTags: map[string]string{"access": "customers"}, detour: true},
		{c: 3, name: "access=delivery", wayTags: map[string]string{"access": "delivery"}, detour: true},
		{c: 4, name: "access=permit", wayTags: map[string]string{"access": "permit"}, detour: true},
		{c: 5, name: "access=residents", wayTags: map[string]string{"access": "residents"}, detour: true},
		{c: 6, name: "access=restricted", wayTags: map[string]string{"access": "restricted"}, detour: true},
		{c: 7, name: "access=military, a value no table lists", wayTags: map[string]string{"access": "military"}, detour: true},
		{c: 8, name: "motor_vehicle=unknown, a value no table lists", wayTags: map[string]string{"motor_vehicle": "unknown"}, detour: true},
		{c: 9, name: "motorcar=private with access=yes", wayTags: map[string]string{"motorcar": "private", "access": "yes"}, detour: true},
		{c: 10, name: "motor_vehicle=private with access=yes", wayTags: map[string]string{"motor_vehicle": "private", "access": "yes"}, detour: true},
		{c: 11, name: "vehicle=private with access=yes", wayTags: map[string]string{"vehicle": "private", "access": "yes"}, detour: true},
	}
	ac3Cases = []transitCase{
		{c: 12, name: "a lift_gate marked residents", nodeTags: map[string]string{"barrier": "lift_gate", "access": "residents"}, detour: true},
		{c: 13, name: "a gate marked on motor_vehicle", nodeTags: map[string]string{"barrier": "gate", "motor_vehicle": "private"}, detour: true},
	}
	ac14Barrier = transitCase{c: 14, name: "a gate carrying access=private with motorcar=yes", nodeTags: map[string]string{"barrier": "gate", "access": "private", "motorcar": "yes"}}
	ac15Cases   = []transitCase{
		{c: 15, name: "access=public", wayTags: map[string]string{"access": "public"}},
		{c: 16, name: "motor_vehicle=allowed", wayTags: map[string]string{"motor_vehicle": "allowed"}},
	}
	ac16Case  = transitCase{c: 17, name: "access=no and no more specific key", wayTags: map[string]string{"access": "no"}, detour: true}
	ac17Cases = []transitCase{
		{c: 18, name: "a way carrying access=private with bicycle=yes and foot=yes", wayTags: map[string]string{"access": "private", "bicycle": "yes", "foot": "yes"}, detour: true},
		{c: 19, name: "a bus gate: a way carrying motor_vehicle=private with psv=yes", wayTags: map[string]string{"motor_vehicle": "private", "psv": "yes"}, detour: true},
		{c: 20, name: "a way carrying motor_vehicle=destination with motorcycle=yes", wayTags: map[string]string{"motor_vehicle": "destination", "motorcycle": "yes"}, detour: true},
		{c: 21, name: "a gate carrying access=private with bicycle=yes and foot=yes", nodeTags: map[string]string{"barrier": "gate", "access": "private", "bicycle": "yes", "foot": "yes"}, detour: true},
	}
	ac18Cases = []transitCase{
		{c: 22, name: "only motor_vehicle:backward=no, driven forward", wayTags: map[string]string{"motor_vehicle:backward": "no"}, detour: true},
		{c: 23, name: "only access:forward=private, driven backward", wayTags: map[string]string{"access:forward": "private"}, backward: true, detour: true},
	}
	ac19Cases = []transitCase{
		{c: 24, name: "a node carrying access=private and no barrier key", nodeTags: map[string]string{"access": "private"}, detour: true},
		{c: 25, name: "a node carrying motor_vehicle=destination and no barrier key", nodeTags: map[string]string{"motor_vehicle": "destination"}, detour: true},
	}
	ac20Cases = []transitCase{
		{c: 26, name: "motorcar=yes with motor_vehicle:backward=no, driven forward", wayTags: map[string]string{"motorcar": "yes", "motor_vehicle:backward": "no"}, detour: true},
		{c: 27, name: "motor_vehicle=yes with access:forward=private, driven backward", wayTags: map[string]string{"motor_vehicle": "yes", "access:forward": "private"}, backward: true, detour: true},
	}
	ac22Case = transitCase{c: 28, name: "access=private with motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00)",
		wayTags: map[string]string{"access": "private", "motor_vehicle:conditional": "yes @ (Mo-Fr 08:00-18:00)"}, detour: true}
)

// --- premises ---------------------------------------------------------------

// synthetic248 fails the test unless the way or node is one of #248's
// synthetic cases with every tag named and no access key beyond them.
func synthetic248(t *testing.T, set map[int64]feature, kind string, id int64, tags map[string]string) feature {
	t.Helper()

	ft := premise(t, set, kind, id, tags)

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}

	onlyKeys(t, kind, id, ft, keys...)

	if id < syntheticBase || (kind == "way" && ft.tags["note"] != syntheticNote248) {
		t.Fatalf("the premise does not hold: %s %d (note %q) is not one of #248's synthetic cases", kind, id, ft.tags["note"])
	}

	return ft
}

// plainWay fails the test unless the synthetic way carries no access key and
// runs between the two nodes, in that order.
func (fx *fixture) plainWay(t *testing.T, id, from, to int64) {
	t.Helper()

	ft := synthetic248(t, fx.opl.ways, "way", id, nil)
	if carries(ft.tags) || len(ft.nodes) < 2 || ft.nodes[0] != from || ft.nodes[len(ft.nodes)-1] != to {
		t.Fatalf("the premise does not hold: way %d carries %v and runs %v, want no access key and %d to %d", id, ft.tags, ft.nodes, from, to)
	}
}

// casePremise establishes a transit case from the extract and the engine.
func (fx *fixture) casePremise(t *testing.T, tc transitCase, wantMarked bool) {
	t.Helper()

	b, c := synNode(tc.c, 2), synNode(tc.c, 4)

	fx.plainWay(t, synWay(tc.c, 1), synNode(tc.c, 1), b)
	fx.plainWay(t, synWay(tc.c, 3), c, synNode(tc.c, 5))

	if tc.detour {
		p, q := synNode(tc.c, 6), synNode(tc.c, 7)
		fx.plainWay(t, synWay(tc.c, 4), b, p)
		fx.plainWay(t, synWay(tc.c, 5), p, q)
		fx.plainWay(t, synWay(tc.c, 6), q, c)
	} else if _, ok := fx.opl.ways[synWay(tc.c, 4)]; ok {
		t.Fatalf("the premise does not hold: case %d has a way round its case way, want none", tc.c)
	}

	order := []int64{b, tc.middle(), c}
	if tc.backward {
		slices.Reverse(order)
	}

	way := synthetic248(t, fx.opl.ways, "way", tc.caseWay(), tc.wayTags)
	if !slices.Equal(way.nodes, order) {
		t.Fatalf("the premise does not hold: the case way runs %v, want %v", way.nodes, order)
	}

	tags := way.tags

	if tc.nodeTags != nil {
		if carries(way.tags) {
			t.Fatalf("the premise does not hold: the way carrying the case's node is itself tagged %v", way.tags)
		}

		tags = synthetic248(t, fx.opl.nodes, "node", tc.middle(), tc.nodeTags).tags
	}

	if marked(tags) != wantMarked {
		t.Fatalf("the premise does not hold: the independent reading gives marked=%v for %v, want %v", marked(tags), tags, wantMarked)
	}

	n := fx.opl.nodes[tc.middle()]
	if !fx.quickestPasses(t, tc.from(), tc.to(), n.lat, n.lon) {
		t.Fatalf("the premise does not hold: the quickest drivable connection between the ends, access set aside, does not pass node %d of the case way", tc.middle())
	}
}

// quickestPasses reports whether the engine's quickest connection between two
// points, access set aside (`ignore_access`, the `auto` costing's documented
// option at 3.9.1), passes within a shape vertex of the position. It is the
// premise "whose quickest drivable connection runs along" a marking, asked of
// the engine without the marking's cost; the service never sends it.
func (fx *fixture) quickestPasses(t *testing.T, from, to planner.Point, lat, lon float64) bool {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"locations":       []map[string]float64{{"lat": from.Latitude, "lon": from.Longitude}, {"lat": to.Latitude, "lon": to.Longitude}},
		"costing":         "auto",
		"costing_options": map[string]any{"auto": map[string]any{"ignore_access": true}},
		"directions_type": "maneuvers",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(input(t, envValhalla), "/")+"/route", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("the premise could not be asked: the engine answered %d, %s, %v", resp.StatusCode, raw, err)
	}

	var r struct {
		Trip struct {
			Legs []struct {
				Shape string `json:"shape"`
			} `json:"legs"`
		} `json:"trip"`
	}

	if err := json.Unmarshal(raw, &r); err != nil || len(r.Trip.Legs) != 1 {
		t.Fatalf("the premise could not be read: %s, %v", raw, err)
	}

	for _, p := range decodePolyline6(t, r.Trip.Legs[0].Shape) {
		// Six decimal places round a vertex by at most half a millionth of a
		// degree on each axis.
		if math.Abs(p.Latitude-lat) <= 1e-6 && math.Abs(p.Longitude-lon) <= 1e-6 {
			return true
		}
	}

	return false
}

// decodePolyline6 decodes an encoded polyline at six decimal places
// (docs/docs/api/decoding.md at 3.9.1), written again here rather than taken
// from the adapter it tests.
func decodePolyline6(t *testing.T, s string) []planner.Point {
	t.Helper()

	var (
		out      []planner.Point
		lat, lon int64
	)

	next := func(i *int) int64 {
		var result, shift int64

		for {
			if *i >= len(s) {
				t.Fatalf("the shape %q ends inside a value", s)
			}

			b := int64(s[*i]) - 63
			*i++
			result |= (b & 0x1f) << shift
			shift += 5

			if b < 0x20 {
				break
			}
		}

		if result&1 != 0 {
			return ^(result >> 1)
		}

		return result >> 1
	}

	for i := 0; i < len(s); {
		lat += next(&i)
		lon += next(&i)
		out = append(out, planner.Point{Latitude: float64(lat) / 1e6, Longitude: float64(lon) / 1e6})
	}

	return out
}

// --- what is measured -------------------------------------------------------

// counting is the engine with every request it is asked recorded.
type counting struct {
	planner.RoutingProvider

	mu     sync.Mutex
	avoids [][]planner.Point
}

func (c *counting) Route(ctx context.Context, from, to planner.Point, avoid []planner.Point) ([]planner.Edge, error) {
	c.mu.Lock()
	c.avoids = append(c.avoids, slices.Clone(avoid))
	c.mu.Unlock()

	return c.RoutingProvider.Route(ctx, from, to, avoid)
}

// avoiding is the planner #248 builds: every leg checked, and a refused one
// asked for again up to testAvoidance times.
func (fx *fixture) avoiding(t *testing.T) (*planner.Planner, *counting) {
	t.Helper()

	routing := &counting{RoutingProvider: fx.routing}

	p, err := planner.New(routing, heldCopy{}, fx.store, tilesFrom(fx.tiles), planner.AvoidUpTo(testAvoidance))
	if err != nil {
		t.Fatal(err)
	}

	return p, routing
}

// markedOn lists what the independent reading marks on the leg's edges: each
// way it marks, and each node it marks on an edge's way that the edge's
// stretch passes within a vertex of.
func (fx *fixture) markedOn(edges []planner.Edge) []string {
	var out []string

	for i, e := range edges {
		ft, ok := fx.opl.ways[e.WayID]
		if !ok {
			out = append(out, fmt.Sprintf("edge %d along way %d, which the reading does not hold", i, e.WayID))

			continue
		}

		if marked(ft.tags) {
			out = append(out, fmt.Sprintf("edge %d along the marked way %d", i, e.WayID))
		}

		for _, id := range ft.nodes {
			n := fx.opl.nodes[id]
			if !marked(n.tags) {
				continue
			}

			for _, p := range e.Shape {
				if math.Abs(p.Latitude-n.lat) <= 1e-6 && math.Abs(p.Longitude-n.lon) <= 1e-6 {
					out = append(out, fmt.Sprintf("edge %d past the marked node %d", i, id))

					break
				}
			}
		}
	}

	return out
}

// avoided plans the case's leg with the planner #248 builds and fails unless
// it is produced, runs along no way and past no node the independent reading
// marks, keeps off the case way, and the check finds nothing on it.
func (fx *fixture) avoided(t *testing.T, tc transitCase) {
	t.Helper()

	fx.casePremise(t, tc, true)

	p, routing := fx.avoiding(t)

	journey, err := p.Plan(t.Context(), planner.Request{Origin: tc.from(), Destination: tc.to()})
	if err != nil {
		t.Fatalf("Plan produced no leg though a connection avoiding the marking exists: %v (asked to avoid %v)", err, routing.avoids)
	}

	if len(journey.Legs) != 1 {
		t.Fatalf("the journey has %d legs, want 1", len(journey.Legs))
	}

	leg := journey.Legs[0]

	if found := fx.markedOn(leg.Edges); len(found) > 0 {
		t.Errorf("the leg crosses what the extract marks: %v", found)
	}

	if slices.Contains(wayIDs(leg.Edges), tc.caseWay()) {
		t.Errorf("the leg runs along the case way %d, want it kept off", tc.caseWay())
	}

	if crossings, err := fx.store.Crossings(t.Context(), leg.Edges); err != nil || len(crossings) > 0 {
		t.Errorf("the check finds %v, %v on the leg produced, want nothing", crossings, err)
	}

	t.Logf("the leg runs along %v; the engine was asked %d time(s), avoiding %v", wayIDs(leg.Edges), len(routing.avoids), routing.avoids)
}

// driven plans the leg with the planner #248 builds and fails unless it is
// produced along the way named, crossing nothing the reading or the check
// marks.
func (fx *fixture) driven(t *testing.T, from, to planner.Point, way int64) {
	t.Helper()

	p, _ := fx.avoiding(t)

	journey, err := p.Plan(t.Context(), planner.Request{Origin: from, Destination: to})
	if err != nil {
		t.Fatalf("Plan refused a leg along no marking: %v", err)
	}

	leg := journey.Legs[0]

	if !slices.Contains(wayIDs(leg.Edges), way) {
		t.Fatalf("the leg runs along %v, not along way %d", wayIDs(leg.Edges), way)
	}

	if found := fx.markedOn(leg.Edges); len(found) > 0 {
		t.Errorf("the leg crosses what the extract marks: %v", found)
	}

	if crossings, err := fx.store.Crossings(t.Context(), leg.Edges); err != nil || len(crossings) > 0 {
		t.Errorf("the check finds %v, %v on the leg, want nothing", crossings, err)
	}

	t.Logf("the leg runs along %v", wayIDs(leg.Edges))
}

// --- the criteria: transit --------------------------------------------------

// TestAC1AWayMarkedPrivateIsNotDrivenInTransit is #248 AC1 (`FR-125`, first
// criterion).
func TestAC1AWayMarkedPrivateIsNotDrivenInTransit(t *testing.T) {
	fx := setUp(t)
	fx.avoided(t, ac2Cases[0])
}

// TestAC2EveryValueOutsideThePermissiveSetIsExcluded is #248 AC2 (`FR-125`,
// first criterion; `RD-059`, `RD-062`): each listed value, two values no table
// lists, and a marking on each key more specific than access, beside
// access=yes.
func TestAC2EveryValueOutsideThePermissiveSetIsExcluded(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac2Cases {
		t.Run(tc.name, func(t *testing.T) { fx.avoided(t, tc) })
	}
}

// TestAC3AMarkedBarrierIsNotPassedInTransit is #248 AC3 (`FR-125`, first
// criterion; `RD-059`).
func TestAC3AMarkedBarrierIsNotPassedInTransit(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac3Cases {
		t.Run(tc.name, func(t *testing.T) { fx.avoided(t, tc) })
	}
}

// TestAC10WhereNothingAvoidsThemNoLegRunsThroughThem is #248 AC10 (`FR-125`,
// third criterion): between points every drivable connection between which
// runs along a marked way, or past a marked barrier, the planner that asks
// again avoiding them produces no leg. The cases are #255's: Frönäs sjöväg's
// lift_gate, on the only road between its ends, and the synthetic road whose
// middle way carries motorcar=yes with motor_vehicle:backward=no, the only way
// between its outer two.
func TestAC10WhereNothingAvoidsThemNoLegRunsThroughThem(t *testing.T) {
	fx := setUp(t)

	premise(t, fx.opl.nodes, "node", 8691377189, map[string]string{"barrier": "lift_gate", "access": "residents"})
	synthetic(t, fx.opl, synMotorcarBackward, map[string]string{"motorcar": "yes", "motor_vehicle:backward": "no"})

	for _, tc := range []legCase{crossingCases[1], directionalCases[len(directionalCases)-1]} {
		t.Run(tc.name, func(t *testing.T) {
			unmarkedEnd(t, fx.opl, tc.fromWay)
			unmarkedEnd(t, fx.opl, tc.toWay)

			p, routing := fx.avoiding(t)

			journey, err := p.Plan(t.Context(), planner.Request{Origin: tc.from, Destination: tc.to})

			var refusal *planner.CrossingError
			if journey.Legs != nil || !errors.As(err, &refusal) {
				t.Fatalf("Plan gave %d legs and %v, want none and a CrossingError", len(journey.Legs), err)
			}

			if refusal.Avoided == 0 || len(routing.avoids) < 2 {
				t.Errorf("the leg was refused after %d re-asks (%d requests), want it asked for again avoiding the marking first", refusal.Avoided, len(routing.avoids))
			}

			if !slices.ContainsFunc(refusal.Crossings, func(c planner.Crossing) bool { return c.OSMID == tc.feature }) {
				t.Errorf("the refusal names %v, want the %s %d", refusal.Crossings, tc.kind, tc.feature)
			}

			t.Logf("refused: %v; asked to avoid %v", err, routing.avoids)
		})
	}
}

// TestAC14AMoreSpecificKeysPermissionIsDriven is #248 AC14 (`FR-125`, fifth
// criterion; `RD-063`): Ringvägen, access=no with motor_vehicle=yes (#255's
// AC3 case), and a gate carrying access=private with motorcar=yes.
func TestAC14AMoreSpecificKeysPermissionIsDriven(t *testing.T) {
	fx := setUp(t)

	t.Run("a way carrying access=no with motor_vehicle=yes", func(t *testing.T) {
		ringvagen := premise(t, fx.opl.ways, "way", 107135272, map[string]string{"access": "no", "motor_vehicle": "yes"})
		onlyKeys(t, "way", 107135272, ringvagen, "access", "motor_vehicle")

		if marked(ringvagen.tags) {
			t.Fatalf("the premise does not hold: the independent reading marks Ringvägen (%v)", ringvagen.tags)
		}

		fx.driven(t, ringvagenFrom, ringvagenTo, 107135272)
	})

	t.Run(ac14Barrier.name, func(t *testing.T) {
		fx.casePremise(t, ac14Barrier, false)
		fx.driven(t, ac14Barrier.from(), ac14Barrier.to(), ac14Barrier.caseWay())
	})
}

// TestAC15PublicAndAllowedAreDriven is #248 AC15 (`FR-125`, sixth criterion;
// `RD-063`).
func TestAC15PublicAndAllowedAreDriven(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac15Cases {
		t.Run(tc.name, func(t *testing.T) {
			fx.casePremise(t, tc, false)
			fx.driven(t, tc.from(), tc.to(), tc.caseWay())
		})
	}
}

// TestAC16AccessNoWithNoMoreSpecificKeyStaysExcluded is #248 AC16 (`FR-125`,
// seventh criterion; `RD-063`).
func TestAC16AccessNoWithNoMoreSpecificKeyStaysExcluded(t *testing.T) {
	fx := setUp(t)
	fx.avoided(t, ac16Case)
}

// TestAC17APermissionOffTheCarChainLiftsNoMarking is #248 AC17 (`FR-125`,
// eighth criterion; `RD-063`): bicycle, foot, psv and motorcycle permissions
// beside a marking on the car chain, on ways and on a gate. The psv case is a
// bus gate's tagging.
func TestAC17APermissionOffTheCarChainLiftsNoMarking(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac17Cases {
		t.Run(tc.name, func(t *testing.T) { fx.avoided(t, tc) })
	}
}

// TestAC18AWayMarkedInOneDirectionIsNotDrivenInTheOther is #248 AC18
// (`FR-125`, ninth criterion; `RD-064`).
func TestAC18AWayMarkedInOneDirectionIsNotDrivenInTheOther(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac18Cases {
		t.Run(tc.name, func(t *testing.T) { fx.avoided(t, tc) })
	}
}

// TestAC19ANodeSoMarkedWithNoBarrierKeyIsNotPassed is #248 AC19 (`FR-125`,
// tenth criterion; `RD-064`).
func TestAC19ANodeSoMarkedWithNoBarrierKeyIsNotPassed(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac19Cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := fx.opl.nodes[tc.middle()].tags["barrier"]; ok {
				t.Fatalf("the premise does not hold: node %d carries a barrier key", tc.middle())
			}

			fx.avoided(t, tc)
		})
	}
}

// TestAC20ADirectionalMarkingStandsAgainstAPlainMoreSpecificPermission is #248
// AC20 (`FR-125`, eleventh criterion; `RD-065`).
func TestAC20ADirectionalMarkingStandsAgainstAPlainMoreSpecificPermission(t *testing.T) {
	fx := setUp(t)

	for _, tc := range ac20Cases {
		t.Run(tc.name, func(t *testing.T) { fx.avoided(t, tc) })
	}
}

// TestAC21AConditionalValueAloneIsDriven is #248 AC21 (`FR-125`, twelfth
// criterion; `RD-064`), on #255's AC8 case: Mornington Road, carrying only
// motor_vehicle:conditional=no @ (Mo-Fr 08:00-18:00), between Anglesea Place
// and the rest of Mornington Road.
func TestAC21AConditionalValueAloneIsDriven(t *testing.T) {
	fx := setUp(t)

	conditional := premise(t, fx.opl.ways, "way", 4313847, map[string]string{"motor_vehicle:conditional": "no @ (Mo-Fr 08:00-18:00)"})
	onlyKeys(t, "way", 4313847, conditional)
	unmarkedEnd(t, fx.opl, 4315572)
	unmarkedEnd(t, fx.opl, 462835349)

	fx.driven(t,
		planner.Point{Latitude: 51.4684353, Longitude: -2.6163190},
		planner.Point{Latitude: 51.4686804, Longitude: -2.6146257},
		4313847)
}

// TestAC22AConditionalPermissionLiftsNoMarking is #248 AC22 (`FR-125`, first
// criterion, as `RD-064` ignores a conditional form).
func TestAC22AConditionalPermissionLiftsNoMarking(t *testing.T) {
	fx := setUp(t)
	fx.avoided(t, ac22Case)
}

// --- the criteria: the run --------------------------------------------------

// The run cases' places, as the generator wrote them.
const (
	runCase    = 29 // AC4
	farCase    = 30 // AC6
	beyondCase = 31 // AC12, with a way round the gate
	walledCase = 32 // AC12, with none
)

// runOf is a leg's run at its start, or at its end: the edges from that end
// along ways the reading marks, and the edge after them.
func (fx *fixture) runOf(leg planner.Leg, atEnd bool) (run []int, after int) {
	order := make([]int, len(leg.Edges))
	for i := range order {
		order[i] = i
		if atEnd {
			order[i] = len(leg.Edges) - 1 - i
		}
	}

	for _, i := range order {
		if !marked(fx.opl.ways[leg.Edges[i].WayID].tags) {
			return run, i
		}

		run = append(run, i)
	}

	return run, -1
}

// distance is the geodesic distance between two points, by PostGIS.
func (fx *fixture) distance(t *testing.T, a, b planner.Point) float64 {
	t.Helper()

	var m float64
	if err := fx.pool.QueryRow(t.Context(),
		`SELECT ST_Distance(ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, ST_SetSRID(ST_MakePoint($3, $4), 4326)::geography)`,
		a.Longitude, a.Latitude, b.Longitude, b.Latitude).Scan(&m); err != nil {
		t.Fatal(err)
	}

	return m
}

// atLocation fails the test unless the leg starts — or ends — at the
// location, its travel there along marked ways is one run along the ways
// named that ends at the first unmarked way, nothing past that run is marked,
// and the check finds nothing past it either.
func (fx *fixture) atLocation(t *testing.T, leg planner.Leg, loc planner.Point, atEnd bool, runWays []int64) {
	t.Helper()

	end := leg.Shape[0]
	if atEnd {
		end = leg.Shape[len(leg.Shape)-1]
	}

	// The location lies on its way, so the engine's end of the leg is at it;
	// how far an end may lie from a location is #250's to bound.
	if d := fx.distance(t, end, loc); d > 1 {
		t.Errorf("the leg's end lies %.2f m from the location, want it at the location", d)
	}

	run, after := fx.runOf(leg, atEnd)
	if len(run) == 0 || after < 0 {
		t.Fatalf("the leg's travel at the location along marked ways is %v, then edge %d; want a run ending at an unmarked way", run, after)
	}

	for _, i := range run {
		if !slices.Contains(runWays, leg.Edges[i].WayID) {
			t.Errorf("the run at the location runs along way %d, want only %v", leg.Edges[i].WayID, runWays)
		}
	}

	var beyond []planner.Edge

	for i, e := range leg.Edges {
		if !slices.Contains(run, i) {
			beyond = append(beyond, e)
		}
	}

	if found := fx.markedOn(beyond); len(found) > 0 {
		t.Errorf("beyond the run the leg crosses what the extract marks: %v", found)
	}

	crossings, err := fx.store.Crossings(t.Context(), leg.Edges)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range crossings {
		if !slices.Contains(run, c.Edge) {
			t.Errorf("the check finds %v beyond the run", c)
		}
	}
}

// TestAC4EveryLocationTheUserEnteredIsReachedAlongTheRun is #248 AC4
// (`FR-126`, first criterion): a home on a private road, and a customers-only
// car park whose two ways are marked customers, each reached by a car only
// along marked ways, as the origin, the destination, and an intermediate
// destination. A stopping position the user chose is a location the user
// entered too (`RD-061`); the request has no stopping position yet (#78,
// #103, #105), so the car park as an intermediate destination is the case a
// stop the user chose there would be: both legs at it, held to the same run.
func TestAC4EveryLocationTheUserEnteredIsReachedAlongTheRun(t *testing.T) {
	fx := setUp(t)

	lat := synLat(runCase)
	home := planner.Point{Latitude: lat + 0.0015, Longitude: 11.703}
	park := planner.Point{Latitude: lat + 0.001, Longitude: 11.7066}
	west := planner.Point{Latitude: lat, Longitude: 11.7005}
	east := planner.Point{Latitude: lat, Longitude: 11.7085}

	homeWay := synthetic248(t, fx.opl.ways, "way", synWay(runCase, 4), map[string]string{"access": "private"})
	park1 := synthetic248(t, fx.opl.ways, "way", synWay(runCase, 5), map[string]string{"access": "customers"})
	park2 := synthetic248(t, fx.opl.ways, "way", synWay(runCase, 6), map[string]string{"access": "customers"})

	// A car reaches each only along marked ways: the home's road meets the
	// public road at its one end, the car park's at its first way's.
	if homeWay.nodes[0] != synNode(runCase, 2) || park1.nodes[0] != synNode(runCase, 3) || park2.nodes[0] != park1.nodes[1] {
		t.Fatalf("the premise does not hold: the marked ways run %v, %v, %v", homeWay.nodes, park1.nodes, park2.nodes)
	}

	for k := 1; k <= 3; k++ {
		fx.plainWay(t, synWay(runCase, k), synNode(runCase, k), synNode(runCase, k+1))
	}

	homeRun := []int64{synWay(runCase, 4)}
	parkRun := []int64{synWay(runCase, 5), synWay(runCase, 6)}

	type at struct {
		leg   int
		atEnd bool
		loc   planner.Point
		ways  []int64
	}

	for _, tc := range []struct {
		name string
		req  planner.Request
		at   []at
	}{
		{"the origin, a home on a private road", planner.Request{Origin: home, Destination: east}, []at{{0, false, home, homeRun}}},
		{"the destination, a home on a private road", planner.Request{Origin: west, Destination: home}, []at{{0, true, home, homeRun}}},
		{"an intermediate destination, a home on a private road", planner.Request{Origin: west, Intermediates: []planner.Point{home}, Destination: east},
			[]at{{0, true, home, homeRun}, {1, false, home, homeRun}}},
		{"the origin, a customers-only car park", planner.Request{Origin: park, Destination: west}, []at{{0, false, park, parkRun}}},
		{"the destination, a customers-only car park", planner.Request{Origin: west, Destination: park}, []at{{0, true, park, parkRun}}},
		{"a stop in a customers-only car park, as an intermediate destination", planner.Request{Origin: west, Intermediates: []planner.Point{park}, Destination: east},
			[]at{{0, true, park, parkRun}, {1, false, park, parkRun}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := fx.avoiding(t)

			journey, err := p.Plan(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("Plan refused the journey: %v", err)
			}

			for _, a := range tc.at {
				fx.atLocation(t, journey.Legs[a.leg], a.loc, a.atEnd, a.ways)
				t.Logf("leg %d runs along %v", a.leg, wayIDs(journey.Legs[a.leg].Edges))
			}
		})
	}
}

// TestAC6TheRunCarriesNoMoreMarkedTravelThanTheLeast is #248 AC6 (`FR-126`,
// third criterion): a home on private way A, which meets the north road 89 m
// from it; A continues as private way B through an estate to the south road,
// on which the destination lies just past B's end, so the quickest leg leaves
// by B. The leg produced leaves by A to the north road.
func TestAC6TheRunCarriesNoMoreMarkedTravelThanTheLeast(t *testing.T) {
	fx := setUp(t)

	lat := synLat(farCase)
	home := planner.Point{Latitude: lat + 0.0032, Longitude: 11.703}
	dest := planner.Point{Latitude: lat, Longitude: 11.7035}
	wayA, wayB := synWay(farCase, 3), synWay(farCase, 4)
	x1, x2 := synNode(farCase, 2), synNode(farCase, 6)

	a := synthetic248(t, fx.opl.ways, "way", wayA, map[string]string{"access": "private"})
	b := synthetic248(t, fx.opl.ways, "way", wayB, map[string]string{"access": "private"})

	if a.nodes[0] != x1 || a.nodes[1] != b.nodes[0] || b.nodes[1] != x2 {
		t.Fatalf("the premise does not hold: A runs %v and B %v, want A from the north road to B, and B to the south road", a.nodes, b.nodes)
	}

	fx.plainWay(t, synWay(farCase, 6), x2, synNode(farCase, 7))

	// The quickest leg leaves by the far side.
	far := fx.opl.nodes[x2]
	if !fx.quickestPasses(t, home, dest, far.lat, far.lon) {
		t.Fatal("the premise does not hold: the quickest drivable connection from the home to the destination does not leave by B")
	}

	if edges, err := fx.routing.Route(t.Context(), home, dest, nil); err == nil {
		t.Logf("the engine's own route, nothing avoided, runs along %v", wayIDs(edges))
	}

	near := fx.opl.nodes[x1]
	least := fx.distance(t, home, planner.Point{Latitude: near.lat, Longitude: near.lon})

	for _, tc := range []struct {
		name  string
		req   planner.Request
		atEnd bool
		loc   planner.Point
	}{
		{"the leg leaving the home", planner.Request{Origin: home, Destination: dest}, false, home},
		{"the leg reaching the home", planner.Request{Origin: dest, Destination: home}, true, home},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, routing := fx.avoiding(t)

			journey, err := p.Plan(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("Plan refused the journey: %v (asked to avoid %v)", err, routing.avoids)
			}

			leg := journey.Legs[0]
			fx.atLocation(t, leg, tc.loc, tc.atEnd, []int64{wayA})

			run, _ := fx.runOf(leg, tc.atEnd)

			ordered := make([]planner.Edge, len(run))
			for i, e := range run {
				ordered[i] = leg.Edges[e]
				if tc.atEnd {
					ordered[i].Shape = slices.Clone(ordered[i].Shape)
					slices.Reverse(ordered[i].Shape)
				}
			}

			measure, err := fx.store.Run(t.Context(), ordered)
			if err != nil {
				t.Fatalf("the run could not be measured: %v", err)
			}

			if measure.Marked > measure.Least {
				t.Errorf("the run carries %.3f m of marked way, more than the least, %.3f m", measure.Marked, measure.Least)
			}

			if math.Abs(measure.Least-least) > 1 {
				t.Errorf("the least run measures %.3f m, want the %.3f m from the home to the north road", measure.Least, least)
			}

			t.Logf("the leg runs along %v; its run carries %.3f m against a least of %.3f m; asked to avoid %v", wayIDs(leg.Edges), measure.Marked, measure.Least, routing.avoids)
		})
	}
}

// TestAC12BeyondTheRunALegIsHeldToTheRuleInTransit is #248 AC12 (`FR-125`,
// fourth criterion): from a home on a private road, the quickest connection to
// the destination passes a lift_gate marked residents beyond the run. With a
// way round the gate the leg is produced on it, its run at the home admitted
// and nothing beyond it marked; with none, no leg is produced.
func TestAC12BeyondTheRunALegIsHeldToTheRuleInTransit(t *testing.T) {
	fx := setUp(t)

	for _, c := range []int{beyondCase, walledCase} {
		lat := synLat(c)
		home := planner.Point{Latitude: lat + 0.003, Longitude: 11.703}
		dest := planner.Point{Latitude: lat + 0.002, Longitude: 11.710}

		synthetic248(t, fx.opl.ways, "way", synWay(c, 1), map[string]string{"access": "private"})
		gate := synthetic248(t, fx.opl.nodes, "node", synNode(c, 5), map[string]string{"barrier": "lift_gate", "access": "residents"})

		if !fx.quickestPasses(t, home, dest, gate.lat, gate.lon) {
			t.Fatalf("the premise does not hold: the quickest drivable connection in case %d does not pass the lift_gate", c)
		}

		_, bypass := fx.opl.ways[synWay(c, 6)]
		if bypass != (c == beyondCase) {
			t.Fatalf("the premise does not hold: case %d has a way round the gate: %v", c, bypass)
		}
	}

	t.Run("with a way round the gate", func(t *testing.T) {
		lat := synLat(beyondCase)
		home := planner.Point{Latitude: lat + 0.003, Longitude: 11.703}
		dest := planner.Point{Latitude: lat + 0.002, Longitude: 11.710}

		for _, req := range []planner.Request{{Origin: home, Destination: dest}, {Origin: dest, Destination: home}} {
			p, routing := fx.avoiding(t)

			journey, err := p.Plan(t.Context(), req)
			if err != nil {
				t.Fatalf("Plan refused the journey: %v (asked to avoid %v)", err, routing.avoids)
			}

			leg := journey.Legs[0]
			fx.atLocation(t, leg, home, req.Destination == home, []int64{synWay(beyondCase, 1)})

			if slices.Contains(wayIDs(leg.Edges), synWay(beyondCase, 4)) {
				t.Errorf("the leg runs along the gate's way, %v", wayIDs(leg.Edges))
			}

			t.Logf("the leg runs along %v; asked to avoid %v", wayIDs(leg.Edges), routing.avoids)
		}
	})

	t.Run("with nothing round the gate", func(t *testing.T) {
		lat := synLat(walledCase)
		home := planner.Point{Latitude: lat + 0.003, Longitude: 11.703}
		dest := planner.Point{Latitude: lat + 0.002, Longitude: 11.710}

		p, routing := fx.avoiding(t)

		journey, err := p.Plan(t.Context(), planner.Request{Origin: home, Destination: dest})

		var refusal *planner.CrossingError
		if journey.Legs != nil || !errors.As(err, &refusal) {
			t.Fatalf("Plan gave %d legs and %v, want none and a CrossingError", len(journey.Legs), err)
		}

		if !slices.ContainsFunc(refusal.Crossings, func(c planner.Crossing) bool { return c.Kind == "barrier" && c.OSMID == synNode(walledCase, 5) }) {
			t.Errorf("the refusal names %v, want the lift_gate", refusal.Crossings)
		}

		t.Logf("refused: %v; asked to avoid %v", err, routing.avoids)
	})
}
