//go:build valhalla

package accessstore_test

// #271's acceptance criteria, AC1 to AC7, against the committed extract: the
// tables loaded from it, the tiles a real Valhalla serves built from the same
// file, and the extract read independently as OPL, as integration_test.go
// (#255) and exclusion_integration_test.go (#248) do, whose helpers these
// share.
//
// EACH TEST ESTABLISHES ITS PREMISE FROM THE EXTRACT FIRST: the markings with
// their tags, the ends on ways the reading leaves unmarked, and — for a
// synthetic layout — that the ways named are every way at its nodes, so the
// connections between its ends are exactly those the layout draws. Where the
// premise is "every connection passes the marking" on a real road, the case is
// #255's Frönäs sjöväg, the only road between its ends; there the engine's own
// answer to a request keeping off the gate is logged, and is no premise.
//
// AC1 is asked through `POST /journeys`, as it states; the identification of
// every marked stretch, which the wire cannot be held to without the leg's
// ways, is held on the planner's answer to the same request.

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

const syntheticNote271 = "synthetic, not from OpenStreetMap: TurfGPS #271 test case"

// #271's layouts, as the generator wrote them: layout j's nodes are
// syntheticBase + 271000 + 100j + k and its ways syntheticBase + 27100 + 20j +
// k, along 57.034 + 0.005j ° N (`testdata/README.md` § #271's cases).
func wNode(j, k int) int64 { return syntheticBase + 271000 + int64(j)*100 + int64(k) }
func wWay(j, k int) int64  { return syntheticBase + 27100 + int64(j)*20 + int64(k) }
func wLat(j int) float64   { return 57.034 + 0.005*float64(j) }

const (
	avoidableCase = 0 // AC4
	eitherCase    = 1 // AC5
	bothCase      = 2 // AC6
)

var (
	gateTags     = map[string]string{"barrier": "lift_gate", "access": "residents"}
	backwardTags = map[string]string{"motorcar": "yes", "motor_vehicle:backward": "no"}
)

// --- premises ---------------------------------------------------------------

// synthetic271 fails the test unless the way or node is one of #271's
// synthetic cases with every tag named and no access key beyond them.
func synthetic271(t *testing.T, set map[int64]feature, kind string, id int64, tags map[string]string) feature {
	t.Helper()

	ft := premise(t, set, kind, id, tags)

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}

	onlyKeys(t, kind, id, ft, keys...)

	if id < syntheticBase || (kind == "way" && ft.tags["note"] != syntheticNote271) {
		t.Fatalf("the premise does not hold: %s %d (note %q) is not one of #271's synthetic cases", kind, id, ft.tags["note"])
	}

	return ft
}

// isolated fails the test unless the ways named are every way in the extract
// at any of their nodes: a car reaches nothing from them but along them.
func (fx *fixture) isolated(t *testing.T, ways ...int64) {
	t.Helper()

	at := map[int64]bool{}

	for _, id := range ways {
		ft, ok := fx.opl.ways[id]
		if !ok {
			t.Fatalf("the premise does not hold: the extract holds no way %d", id)
		}

		for _, n := range ft.nodes {
			at[n] = true
		}
	}

	for id, ft := range fx.opl.ways {
		if slices.Contains(ways, id) {
			continue
		}

		if slices.ContainsFunc(ft.nodes, func(n int64) bool { return at[n] }) {
			t.Fatalf("the premise does not hold: way %d meets the layout %v, which is then not every connection", id, ways)
		}
	}
}

// on fails the test unless the way runs along the nodes given, in order.
func (fx *fixture) on(t *testing.T, way int64, nodes ...int64) {
	t.Helper()

	if got := fx.opl.ways[way].nodes; !slices.Equal(got, nodes) {
		t.Fatalf("the premise does not hold: way %d runs %v, want %v", way, got, nodes)
	}
}

// --- what is measured -------------------------------------------------------

// marking is one marking the independent reading finds on a leg's edge.
type marking struct {
	edge int
	kind string
	id   int64
}

// markingsOn is markedOn, structured: each way the reading marks that an edge
// runs along, and each node it marks on an edge's way that the edge's stretch
// passes within a vertex of.
func (fx *fixture) markingsOn(edges []planner.Edge) []marking {
	var out []marking

	for i, e := range edges {
		ft := fx.opl.ways[e.WayID]
		if marked(ft.tags) {
			out = append(out, marking{i, "way", e.WayID})
		}

		for _, id := range ft.nodes {
			if n := fx.opl.nodes[id]; marked(n.tags) && passesVertex(e.Shape, n) {
				out = append(out, marking{i, "barrier", id})
			}
		}
	}

	return out
}

func passesVertex(shape []planner.Point, n feature) bool {
	return slices.ContainsFunc(shape, func(p planner.Point) bool {
		return math.Abs(p.Latitude-n.lat) <= 1e-6 && math.Abs(p.Longitude-n.lon) <= 1e-6
	})
}

// starts is where each edge's stretch begins in the leg's shape, which joins
// the stretches each sharing its first point with the last of the one before
// (planner.Leg.Shape), and, last, where the shape ends.
func starts(leg planner.Leg) []int {
	out := make([]int, len(leg.Edges)+1)
	for k, e := range leg.Edges {
		out[k+1] = out[k] + len(e.Shape) - 1
	}

	return out
}

// namesEveryMarking fails the test unless the leg is warned and names every
// marking on it — each the independent reading finds and each the check finds
// — with a stretch of the leg's shape that holds the edge crossing it: a way's
// stretch along that way alone, a barrier's at its point (`FR-131`).
func (fx *fixture) namesEveryMarking(t *testing.T, leg planner.Leg) {
	t.Helper()

	if !leg.Warned {
		t.Fatalf("the leg along %v is not identified as warned", wayIDs(leg.Edges))
	}

	start := starts(leg)

	covers := func(kind string, id int64, edge int) bool {
		return slices.ContainsFunc(leg.Marked, func(m planner.Marked) bool {
			return m.Kind == kind && m.OSMID == id && m.From <= start[edge] && start[edge+1] <= m.To
		})
	}

	for _, m := range fx.markingsOn(leg.Edges) {
		if !covers(m.kind, m.id, m.edge) {
			t.Errorf("the reading marks the %s %d on edge %d, which the leg names nowhere there (it names %+v)", m.kind, m.id, m.edge, leg.Marked)
		}
	}

	crossings, err := fx.store.Crossings(t.Context(), leg.Edges)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range crossings {
		if !covers(c.Kind, c.OSMID, c.Edge) {
			t.Errorf("the check finds %v, which the leg names nowhere there", c)
		}
	}

	for _, m := range leg.Marked {
		if m.From < 0 || m.To >= len(leg.Shape) || m.From >= m.To || !slices.Contains(start, m.From) || !slices.Contains(start, m.To) {
			t.Errorf("the stretch %+v is not a span of whole edges of the leg's %d-point shape", m, len(leg.Shape))

			continue
		}

		switch m.Kind {
		case "way":
			for k := range leg.Edges {
				if start[k] >= m.From && start[k+1] <= m.To && leg.Edges[k].WayID != m.OSMID {
					t.Errorf("the stretch of way %d spans edge %d, along way %d", m.OSMID, k, leg.Edges[k].WayID)
				}
			}
		case "barrier":
			n := fx.opl.nodes[m.OSMID]
			if math.Abs(m.At.Latitude-n.lat) > 1e-6 || math.Abs(m.At.Longitude-n.lon) > 1e-6 {
				t.Errorf("the barrier %d is named at %v, and lies at (%.7f, %.7f)", m.OSMID, m.At, n.lat, n.lon)
			}

			if !passesVertex(leg.Shape[m.From:m.To+1], n) {
				t.Errorf("the barrier %d's stretch, shape %d to %d, does not pass it", m.OSMID, m.From, m.To)
			}
		default:
			t.Errorf("the leg names a %s", m.Kind)
		}
	}

	t.Logf("the warned leg runs along %v and names %+v", wayIDs(leg.Edges), leg.Marked)
}

// names lists the features a warned leg names, "kind id".
func names(leg planner.Leg) []string {
	var out []string

	for _, m := range leg.Marked {
		if s := fmt.Sprintf("%s %d", m.Kind, m.OSMID); !slices.Contains(out, s) {
			out = append(out, s)
		}
	}

	return out
}

// noRouteKeepingOff logs what the engine answers asked for the leg between
// from and to keeping off the place, so a run shows the answer a warned leg
// was produced on.
func (fx *fixture) noRouteKeepingOff(t *testing.T, from, to, place planner.Point) {
	t.Helper()

	_, err := fx.routing.Route(t.Context(), from, to, []planner.Point{place})
	t.Logf("asked to keep off %v, the engine answers %v (ErrNoRoute: %v)", place, err, errors.Is(err, planner.ErrNoRoute))
}

// wireJourney is the part of `POST /journeys`'s answer these tests read.
type wireJourney struct {
	Legs []struct {
		Warned bool         `json:"warned"`
		Marked []wireMarked `json:"marked"`
	} `json:"legs"`
}

type wireMarked struct {
	Kind  string `json:"kind"`
	OSMID int64  `json:"osm_id"`
	From  int    `json:"from"`
	To    int    `json:"to"`
}

func body(req planner.Request) string {
	loc := func(p planner.Point) string {
		return fmt.Sprintf(`{"latitude":%.7f,"longitude":%.7f}`, p.Latitude, p.Longitude)
	}

	via := make([]string, len(req.Intermediates))
	for i, p := range req.Intermediates {
		via[i] = loc(p)
	}

	return fmt.Sprintf(`{"origin":%s,"intermediates":[%s],"destination":%s}`, loc(req.Origin), strings.Join(via, ","), loc(req.Destination))
}

// posted plans req through `POST /journeys` and fails the test unless it is
// answered 200 with a journey.
func (fx *fixture) posted(t *testing.T, req planner.Request) wireJourney {
	t.Helper()

	p, _ := fx.avoiding(t)

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/journeys", strings.NewReader(body(req)))
	r.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	httpapi.NewHandler(nil, httpapi.WithPlanner(p)).ServeHTTP(rec, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /journeys answered %d %q, want 200 and the journey", rec.Code, rec.Body.String())
	}

	var got wireJourney
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("the answer is not a journey: %v; %q", err, rec.Body.String())
	}

	return got
}

// planned plans req with the planner #248 builds and fails the test on an
// error.
func (fx *fixture) planned(t *testing.T, req planner.Request) planner.Journey {
	t.Helper()

	p, routing := fx.avoiding(t)

	journey, err := p.Plan(t.Context(), req)
	if err != nil {
		t.Fatalf("Plan refused the journey: %v (asked to avoid %v)", err, routing.avoids)
	}

	return journey
}

// --- the cases of AC1, AC2 and AC7 -------------------------------------------

// warnedCase is a journey every connection of each of whose legs crosses a
// marking beyond the run: the legs, the feature each must name.
type warnedCase struct {
	name string
	req  planner.Request
	// feature is the marking each leg passes in transit.
	kind    string
	feature int64
	// place is where the engine is asked to keep off it, for the log.
	place planner.Point
}

// warnedCases establishes AC1's premises from the extract and returns its
// cases: #248 AC10's layout — Frönäs sjöväg's lift_gate, and the synthetic
// road whose middle way carries motorcar=yes with motor_vehicle:backward=no —
// #248 AC12's second case, a home on a private road whose every connection
// beyond its run passes a lift_gate marked residents, and that home as an
// intermediate destination, both legs at it in that position.
func (fx *fixture) warnedCases(t *testing.T) []warnedCase {
	t.Helper()

	frönäs, backward := crossingCases[1], directionalCases[len(directionalCases)-1]

	premise(t, fx.opl.nodes, "node", frönäs.feature, gateTags)

	for _, tc := range []legCase{frönäs, backward} {
		unmarkedEnd(t, fx.opl, tc.fromWay)
		unmarkedEnd(t, fx.opl, tc.toWay)
	}

	synthetic(t, fx.opl, synMotorcarBackward, backwardTags)
	fx.isolated(t, syntheticBase+1, synMotorcarBackward, syntheticBase+3)

	lat := synLat(walledCase)
	home := planner.Point{Latitude: lat + 0.003, Longitude: 11.703}
	dest := planner.Point{Latitude: lat + 0.002, Longitude: 11.710}

	synthetic248(t, fx.opl.ways, "way", synWay(walledCase, 1), map[string]string{"access": "private"})
	synthetic248(t, fx.opl.nodes, "node", synNode(walledCase, 5), gateTags)

	for _, k := range []int{2, 3, 4, 5} {
		synthetic248(t, fx.opl.ways, "way", synWay(walledCase, k), nil)
	}

	fx.isolated(t, synWay(walledCase, 1), synWay(walledCase, 2), synWay(walledCase, 3), synWay(walledCase, 4), synWay(walledCase, 5))

	gate := fx.opl.nodes[synNode(walledCase, 5)]
	walledGate := planner.Point{Latitude: gate.lat, Longitude: gate.lon}
	fg := fx.opl.nodes[frönäs.feature]

	return []warnedCase{
		{"Frönäs sjöväg's lift_gate, on the only road between the ends", planner.Request{Origin: frönäs.from, Destination: frönäs.to},
			"barrier", frönäs.feature, planner.Point{Latitude: fg.lat, Longitude: fg.lon}},
		{"the only way between the ends, carrying motorcar=yes with motor_vehicle:backward=no", planner.Request{Origin: backward.from, Destination: backward.to},
			"way", synMotorcarBackward, planner.Point{Latitude: 57.0, Longitude: 11.70375}},
		{"a home on a private road, every connection beyond its run past a lift_gate", planner.Request{Origin: home, Destination: dest},
			"barrier", synNode(walledCase, 5), walledGate},
		{"that home as an intermediate destination, both legs at it", planner.Request{Origin: dest, Intermediates: []planner.Point{home}, Destination: dest},
			"barrier", synNode(walledCase, 5), walledGate},
	}
}

// TestAC1WhereNothingAvoidsAMarkingTheLegIsAnsweredWarned is #271 AC1
// (`FR-129`, first criterion; `FR-125`, third and fourth criteria as `RD-067`
// amends them): `POST /journeys` answers each case with the journey, every leg
// identified as warned and naming the marking it passes; and every marking on
// every leg of the journey is named where it lies.
func TestAC1WhereNothingAvoidsAMarkingTheLegIsAnsweredWarned(t *testing.T) {
	fx := setUp(t)

	for _, tc := range fx.warnedCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			next := tc.req.Destination
			if len(tc.req.Intermediates) > 0 {
				next = tc.req.Intermediates[0]
			}

			fx.noRouteKeepingOff(t, tc.req.Origin, next, tc.place)

			got := fx.posted(t, tc.req)
			if want := len(tc.req.Intermediates) + 1; len(got.Legs) != want {
				t.Fatalf("the journey has %d legs, want %d", len(got.Legs), want)
			}

			for i, leg := range got.Legs {
				if !leg.Warned || !slices.ContainsFunc(leg.Marked, func(m wireMarked) bool { return m.Kind == tc.kind && m.OSMID == tc.feature }) {
					t.Errorf("leg %d is answered warned=%v naming %+v, want it warned naming the %s %d", i, leg.Warned, leg.Marked, tc.kind, tc.feature)
				}
			}

			for i, leg := range fx.planned(t, tc.req).Legs {
				t.Run(fmt.Sprintf("leg %d", i), func(t *testing.T) { fx.namesEveryMarking(t, leg) })
			}
		})
	}
}

// TestAC2ALegWithAStopTheSystemChoseAtOneEndIsStillRefused is #271 AC2
// (`FR-129`, second criterion; `RD-068`): AC1's legs, one end a stopping
// position the system chose, told to the planner directly, at either end in
// turn, give no leg, warned or not.
func TestAC2ALegWithAStopTheSystemChoseAtOneEndIsStillRefused(t *testing.T) {
	fx := setUp(t)

	for _, tc := range fx.warnedCases(t)[:3] {
		for _, systemAt := range []string{"from", "to"} {
			t.Run(tc.name+", the system's stop at its "+systemAt+" end", func(t *testing.T) {
				p, routing := fx.avoiding(t)

				from := planner.LegEnd{At: tc.req.Origin, EnteredByUser: systemAt != "from"}
				to := planner.LegEnd{At: tc.req.Destination, EnteredByUser: systemAt != "to"}

				leg, err := p.PlanLeg(t.Context(), from, to)

				var refusal *planner.CrossingError
				if leg.Edges != nil || leg.Warned || !errors.As(err, &refusal) {
					t.Fatalf("PlanLeg gave a leg along %v, warned=%v, and %v; want no leg and a CrossingError", wayIDs(leg.Edges), leg.Warned, err)
				}

				t.Logf("refused: %v; asked to avoid %v", err, routing.avoids)
			})
		}
	}
}

// TestAC3ALegAConnectionAvoidingEveryMarkingJoinsIsNotWarned is #271 AC3
// (`FR-129`, fourth criterion): #248 AC1's layout, a way marked private with
// an unmarked way round it, both ends locations the user entered. The leg is
// produced on the way round and is not warned; avoided holds every case of
// #248 with a way round its marking to the same.
func TestAC3ALegAConnectionAvoidingEveryMarkingJoinsIsNotWarned(t *testing.T) {
	fx := setUp(t)
	fx.avoided(t, ac2Cases[0])
}

// --- AC4 to AC7 -------------------------------------------------------------

// TestAC4AWarnedLegCarriesNoMarkingAConnectionAvoids is #271 AC4 (`FR-130`,
// first criterion): every connection passes the lift_gate on way 2, and the
// quickest, along marked way 3, is shorter than the way round it, which the
// warned leg takes, passing the gate alone.
func TestAC4AWarnedLegCarriesNoMarkingAConnectionAvoids(t *testing.T) {
	fx := setUp(t)

	j := avoidableCase
	gate, way3 := wNode(j, 3), wWay(j, 3)

	synthetic271(t, fx.opl.nodes, "node", gate, gateTags)
	synthetic271(t, fx.opl.ways, "way", way3, backwardTags)

	for _, k := range []int{1, 2, 4, 5, 6, 7} {
		synthetic271(t, fx.opl.ways, "way", wWay(j, k), nil)
	}

	fx.on(t, wWay(j, 2), wNode(j, 2), gate, wNode(j, 4))
	fx.on(t, way3, wNode(j, 4), wNode(j, 5))
	fx.isolated(t, wWay(j, 1), wWay(j, 2), way3, wWay(j, 4), wWay(j, 5), wWay(j, 6), wWay(j, 7))
	fx.shorter(t, []int64{way3}, []int64{wWay(j, 5), wWay(j, 6), wWay(j, 7)})

	from := planner.Point{Latitude: wLat(j), Longitude: 11.7010}
	to := planner.Point{Latitude: wLat(j), Longitude: 11.7100}

	if route, err := fx.routing.Route(t.Context(), from, to, nil); err == nil {
		t.Logf("the engine's own route, nothing avoided, runs along %v", wayIDs(route.Edges))
	}

	leg := fx.planned(t, planner.Request{Origin: from, Destination: to}).Legs[0]
	fx.namesEveryMarking(t, leg)

	if slices.Contains(wayIDs(leg.Edges), way3) {
		t.Errorf("the warned leg runs along the marked way %d, which the way round avoids", way3)
	}

	if got := names(leg); !slices.Equal(got, []string{fmt.Sprintf("barrier %d", gate)}) {
		t.Errorf("the warned leg names %v, want the gate %d alone", got, gate)
	}
}

// TestAC5OfTwoMarkingsEitherOfWhichJoinsTheWarnedLegCrossesOne is #271 AC5
// (`FR-130`, second criterion): two connections, each through a lift_gate of
// its own; the warned leg passes one of them and not both.
func TestAC5OfTwoMarkingsEitherOfWhichJoinsTheWarnedLegCrossesOne(t *testing.T) {
	fx := setUp(t)

	j := eitherCase
	g1, g2 := wNode(j, 3), wNode(j, 7)

	synthetic271(t, fx.opl.nodes, "node", g1, gateTags)
	synthetic271(t, fx.opl.nodes, "node", g2, gateTags)

	for k := 1; k <= 6; k++ {
		synthetic271(t, fx.opl.ways, "way", wWay(j, k), nil)
	}

	fx.on(t, wWay(j, 2), wNode(j, 2), g1, wNode(j, 4))
	fx.on(t, wWay(j, 5), wNode(j, 6), g2, wNode(j, 8))
	fx.isolated(t, wWay(j, 1), wWay(j, 2), wWay(j, 3), wWay(j, 4), wWay(j, 5), wWay(j, 6))

	from := planner.Point{Latitude: wLat(j), Longitude: 11.7010}
	to := planner.Point{Latitude: wLat(j), Longitude: 11.7080}

	leg := fx.planned(t, planner.Request{Origin: from, Destination: to}).Legs[0]
	fx.namesEveryMarking(t, leg)

	got := names(leg)
	if len(got) != 1 || (got[0] != fmt.Sprintf("barrier %d", g1) && got[0] != fmt.Sprintf("barrier %d", g2)) {
		t.Errorf("the warned leg names %v, want exactly one of the gates %d and %d", got, g1, g2)
	}

	var passed []int64

	for _, m := range fx.markingsOn(leg.Edges) {
		if m.kind == "barrier" && !slices.Contains(passed, m.id) {
			passed = append(passed, m.id)
		}
	}

	if len(passed) != 1 {
		t.Errorf("the reading finds the leg passing the gates %v, want one and not both", passed)
	}
}

// TestAC6EveryMarkedStretchOfAWarnedLegIsNamedWhereItLies is #271 AC6
// (`FR-131`, first criterion): a warned leg along the marked way 2 and past the
// lift_gate on way 3, nothing round either, names the way's stretch and the
// gate's point, each locating its own part of the leg's shape.
func TestAC6EveryMarkedStretchOfAWarnedLegIsNamedWhereItLies(t *testing.T) {
	fx := setUp(t)

	j := bothCase
	gate, way2 := wNode(j, 4), wWay(j, 2)

	synthetic271(t, fx.opl.nodes, "node", gate, gateTags)
	synthetic271(t, fx.opl.ways, "way", way2, backwardTags)

	for _, k := range []int{1, 3, 4} {
		synthetic271(t, fx.opl.ways, "way", wWay(j, k), nil)
	}

	fx.on(t, way2, wNode(j, 2), wNode(j, 3))
	fx.on(t, wWay(j, 3), wNode(j, 3), gate, wNode(j, 5))
	fx.isolated(t, wWay(j, 1), way2, wWay(j, 3), wWay(j, 4))

	from := planner.Point{Latitude: wLat(j), Longitude: 11.7010}
	to := planner.Point{Latitude: wLat(j), Longitude: 11.7100}

	leg := fx.planned(t, planner.Request{Origin: from, Destination: to}).Legs[0]
	fx.namesEveryMarking(t, leg)

	want := []string{fmt.Sprintf("way %d", way2), fmt.Sprintf("barrier %d", gate)}
	if got := names(leg); !slices.Equal(got, want) {
		t.Errorf("the warned leg names %v, want %v in the order the leg reaches them", got, want)
	}
}

// TestAC7TheRunOnAWarnedLegIsNamedToo is #271 AC7 (`FR-131`, second
// criterion): AC1's home on a private road, the warned leg leaving it, and
// the leg reaching it, name the run's stretch along the private road, at the
// home's end of the shape.
func TestAC7TheRunOnAWarnedLegIsNamedToo(t *testing.T) {
	fx := setUp(t)

	cases := fx.warnedCases(t)
	leaving := cases[2].req
	road := synWay(walledCase, 1)

	for _, tc := range []struct {
		name  string
		req   planner.Request
		atEnd bool
	}{
		{"the leg leaving the home", leaving, false},
		{"the leg reaching the home", planner.Request{Origin: leaving.Destination, Destination: leaving.Origin}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			leg := fx.planned(t, tc.req).Legs[0]
			fx.namesEveryMarking(t, leg)

			run, _ := fx.runOf(leg, tc.atEnd)
			if len(run) == 0 {
				t.Fatalf("the leg along %v has no run at the home", wayIDs(leg.Edges))
			}

			end := 0
			if tc.atEnd {
				end = len(leg.Shape) - 1
			}

			if !slices.ContainsFunc(leg.Marked, func(m planner.Marked) bool {
				return m.Kind == "way" && m.OSMID == road && m.From <= end && end <= m.To
			}) {
				t.Errorf("the warned leg names %+v, want the run along the private road %d named from the home's end of the shape", leg.Marked, road)
			}
		})
	}
}
