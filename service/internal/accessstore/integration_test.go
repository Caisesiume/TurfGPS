//go:build valhalla

package accessstore_test

// #255's acceptance criteria, against the committed extract
// (`internal/valhalla/testdata/`): the access-marking tables loaded from it
// into a PostGIS, the tiles a real Valhalla serves built from the same file,
// and the extract read a second, independent way — by libosmium, as OPL text
// `osmium cat` wrote — so the table is compared with a reading that shares
// neither this module's PBF reader nor the migration's rule.
//
// BUILT ONLY UNDER THE `valhalla` TAG, as `internal/valhalla`'s integration
// test is, and run by CI's `valhalla` job with every input named; without one
// it fails rather than skips. The tests share one database and run one after
// another; each that changes the tables restores them.
//
// EACH TEST ESTABLISHES ITS PREMISE FROM THE EXTRACT FIRST, from the OPL
// reading: that the way or barrier is there with the tags the case names, and
// that the ways a route's ends lie on carry no marking. An extract that no
// longer holds a case fails here rather than passing vacuously.

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/accessload"
	"github.com/Caisesiume/TurfGPS/service/internal/accessstore"
	"github.com/Caisesiume/TurfGPS/service/internal/planner"
	"github.com/Caisesiume/TurfGPS/service/internal/valhalla"
	"github.com/Caisesiume/TurfGPS/service/internal/zonestore"
)

// The test's inputs, never the service's configuration.
const (
	envDatabase = "TURFGPS_TEST_DATABASE_URL"
	envValhalla = "TURFGPS_TEST_VALHALLA_URL"
	envOPL      = "TURFGPS_TEST_EXTRACT_OPL"
	envTiles    = "TURFGPS_TEST_TILES_SHA256"
)

const extractPath = "../valhalla/testdata/goteborg-inom-vallgraven.osm.pbf"

// testTolerance is this test's barrier tolerance, not the service's, which no
// document gives and the service leaves unset. It sits above what the shape's
// six-digit encoding can move a vertex (under 0.1 m at this latitude), so a
// barrier the engine's shape passes through is counted; each barrier test logs
// the distance it measured, which is the measurement `Architecture.md § What
// is not decided` says the value may be taken from.
var testTolerance = 1.0

func input(t *testing.T, name string) string {
	t.Helper()

	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("%s is not set: this file is built only under the valhalla tag, for CI's valhalla job, and it fails rather than skips without its inputs", name)
	}

	return v
}

// --- the independent reading ------------------------------------------------

type feature struct {
	tags     map[string]string
	nodes    []int64
	lon, lat float64 // a node's
}

type oplExtract struct {
	nodes map[int64]feature
	ways  map[int64]feature
}

var oplEscape = regexp.MustCompile(`%([0-9a-fA-F]+)%`)

func unescape(s string) string {
	return oplEscape.ReplaceAllStringFunc(s, func(m string) string {
		r, err := strconv.ParseUint(m[1:len(m)-1], 16, 32)
		if err != nil {
			return m
		}

		return string(rune(r))
	})
}

// readOPL reads the OPL text `osmium cat -f opl` wrote of the extract
// (https://osmcode.org/opl-file-format/): one object a line, its fields
// separated by spaces, its tags in the field opening with T, a way's nodes in
// the field opening with N, and a node's longitude and latitude in the fields
// opening with x and y.
func readOPL(t *testing.T) *oplExtract {
	t.Helper()

	f, err := os.Open(input(t, envOPL))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	ex := &oplExtract{nodes: map[int64]feature{}, ways: map[int64]feature{}}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)

	for sc.Scan() {
		fields := strings.Split(sc.Text(), " ")
		if len(fields) == 0 || len(fields[0]) < 2 {
			continue
		}

		id, err := strconv.ParseInt(fields[0][1:], 10, 64)
		if err != nil {
			t.Fatalf("an OPL line opens with %q", fields[0])
		}

		ft := feature{tags: map[string]string{}}

		for _, fl := range fields[1:] {
			switch {
			case strings.HasPrefix(fl, "T") && len(fl) > 1:
				for _, kv := range strings.Split(fl[1:], ",") {
					k, v, _ := strings.Cut(kv, "=")
					ft.tags[unescape(k)] = unescape(v)
				}
			case strings.HasPrefix(fl, "N") && len(fl) > 1:
				for _, n := range strings.Split(fl[1:], ",") {
					ref, err := strconv.ParseInt(strings.TrimPrefix(n, "n"), 10, 64)
					if err != nil {
						t.Fatalf("way %d's node %q", id, n)
					}

					ft.nodes = append(ft.nodes, ref)
				}
			case strings.HasPrefix(fl, "x") && len(fl) > 1:
				if ft.lon, err = strconv.ParseFloat(fl[1:], 64); err != nil {
					t.Fatalf("node %d's longitude %q", id, fl)
				}
			case strings.HasPrefix(fl, "y") && len(fl) > 1:
				if ft.lat, err = strconv.ParseFloat(fl[1:], 64); err != nil {
					t.Fatalf("node %d's latitude %q", id, fl)
				}
			}
		}

		switch fields[0][0] {
		case 'n':
			ex.nodes[id] = ft
		case 'w':
			ex.ways[id] = ft
		}
	}

	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	if len(ex.nodes) == 0 || len(ex.ways) == 0 {
		t.Fatalf("the OPL reading holds %d nodes and %d ways", len(ex.nodes), len(ex.ways))
	}

	return ex
}

// The rule, written again from its records rather than from the migration:
// `DECISIONS.md § RD-063` — the first of motorcar, motor_vehicle, vehicle and
// access a feature carries governs, and every value but the five permissive
// ones is a marking — read on the plain keys with every directional form set
// aside, and again on each direction's directional forms alone (`RD-065`); a
// directional form only adds markings (`RD-066`), a feature marked in either
// direction is marked, and no `:conditional` form counts (`RD-064`).
var (
	chain      = []string{"motorcar", "motor_vehicle", "vehicle", "access"}
	permissive = map[string]bool{"yes": true, "permissive": true, "designated": true, "public": true, "allowed": true}
)

func governs(values []*string) bool {
	for _, v := range values {
		if v != nil {
			return !permissive[*v]
		}
	}

	return false
}

func carries(tags map[string]string) bool {
	for _, k := range chain {
		for _, form := range []string{k, k + ":forward", k + ":backward"} {
			if _, ok := tags[form]; ok {
				return true
			}
		}
	}

	return false
}

func marked(tags map[string]string) bool {
	get := func(k string) *string {
		if v, ok := tags[k]; ok {
			return &v
		}

		return nil
	}

	for _, suffix := range []string{"", ":forward", ":backward"} {
		values := make([]*string, len(chain))
		for i, k := range chain {
			values[i] = get(k + suffix)
		}

		if governs(values) {
			return true
		}
	}

	return false
}

// --- the store under test ---------------------------------------------------

type fixture struct {
	pool    *pgxpool.Pool
	store   *accessstore.Store
	routing *valhalla.Client
	tiles   string
	opl     *oplExtract
}

func setUp(t *testing.T) *fixture {
	t.Helper()

	ctx := t.Context()

	pool, err := zonestore.Open(ctx, input(t, envDatabase))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	store, err := accessstore.New(pool, &testTolerance)
	if err != nil {
		t.Fatal(err)
	}

	routing, err := valhalla.New(input(t, envValhalla), 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	fx := &fixture{pool: pool, store: store, routing: routing, tiles: input(t, envTiles), opl: readOPL(t)}
	fx.load(t)

	return fx
}

// load fills the tables from the committed extract, as the loader does, and
// holds the load to the file the tiles were built from.
func (fx *fixture) load(t *testing.T) {
	t.Helper()

	ex, err := accessload.ReadFile(extractPath)
	if err != nil {
		t.Fatalf("reading the extract: %v", err)
	}

	rec, err := accessload.Load(context.WithoutCancel(t.Context()), fx.pool, ex)
	if err != nil {
		t.Fatalf("loading the extract: %v", err)
	}

	if rec.SHA256 != fx.tiles {
		t.Fatalf("the tables were loaded from %s and the tiles built from %s, want one file", rec.SHA256, fx.tiles)
	}
}

// tilesFrom is routing data reporting the extract the CI job built the tiles
// from — the read `DEPLOYMENT.md` owes the service (#25), done here by the job.
type tilesFrom string

func (s tilesFrom) ExtractSHA256(context.Context) (string, error) { return string(s), nil }

type heldCopy struct{}

func (heldCopy) Held(context.Context) (bool, error) { return true, nil }

func (fx *fixture) planner(t *testing.T, routing planner.RoutingProvider, markings planner.AccessMarkings, tiles string) *planner.Planner {
	t.Helper()

	p, err := planner.New(routing, heldCopy{}, markings, tilesFrom(tiles))
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func (fx *fixture) ids(t *testing.T, sql string) []int64 {
	t.Helper()

	rows, err := fx.pool.Query(t.Context(), sql)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var out []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}

		out = append(out, id)
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	slices.Sort(out)

	return out
}

// premise fails the test unless the OPL reading has the feature with every tag
// named, and none of the twelve access keys beyond those named.
func premise(t *testing.T, set map[int64]feature, kind string, id int64, tags map[string]string) feature {
	t.Helper()

	ft, ok := set[id]
	if !ok {
		t.Fatalf("the premise does not hold: the extract holds no %s %d, so this case would measure nothing", kind, id)
	}

	for k, want := range tags {
		if got, ok := ft.tags[k]; !ok || got != want {
			t.Fatalf("the premise does not hold: %s %d carries %s=%q, want %q", kind, id, k, got, want)
		}
	}

	return ft
}

// unmarkedEnd fails the test unless the way a route's end lies on carries no
// access key at all.
func unmarkedEnd(t *testing.T, opl *oplExtract, way int64) {
	t.Helper()

	ft := premise(t, opl.ways, "way", way, nil)
	if carries(ft.tags) {
		t.Fatalf("the premise does not hold: way %d, which a route's end lies on, carries an access key (%v)", way, ft.tags)
	}
}

func sortedKeys(m map[int64]bool) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	slices.Sort(out)

	return out
}

func diff(a, b []int64) []int64 {
	var out []int64

	for _, x := range a {
		if _, found := slices.BinarySearch(b, x); !found {
			out = append(out, x)
		}
	}

	return out
}

// --- the criteria -----------------------------------------------------------

// TestAC1TheTableHoldsExactlyTheExtractsMarkedWaysAndBarriers is #255 AC1
// (`FR-125`, first and fifth to eighth criteria, at the table).
func TestAC1TheTableHoldsExactlyTheExtractsMarkedWaysAndBarriers(t *testing.T) {
	fx := setUp(t)

	// The premises: each case the criterion names is in the extract.
	premise(t, fx.opl.ways, "way", 976307072, map[string]string{"access": "private"})
	premise(t, fx.opl.ways, "way", 25683224, map[string]string{"access": "military"})
	premise(t, fx.opl.ways, "way", 5179080, map[string]string{"motor_vehicle": "private", "access": "yes"})
	premise(t, fx.opl.nodes, "node", 8691377189, map[string]string{"barrier": "lift_gate", "access": "residents"})
	premise(t, fx.opl.ways, "way", 107135272, map[string]string{"access": "no", "motor_vehicle": "yes"})

	wantWays, wantBarriers := map[int64]bool{}, map[int64]bool{}
	storedWays, storedNodes := map[int64]bool{}, map[int64]bool{}

	for id, ft := range fx.opl.ways {
		if carries(ft.tags) {
			storedWays[id] = true
		}

		if marked(ft.tags) {
			wantWays[id] = true
		}
	}

	for id, ft := range fx.opl.nodes {
		if carries(ft.tags) {
			storedNodes[id] = true
		}

		if marked(ft.tags) {
			wantBarriers[id] = true
		}
	}

	for _, tc := range []struct {
		what string
		sql  string
		want []int64
	}{
		{"marked ways", `SELECT way_id FROM public.access_way WHERE marked`, sortedKeys(wantWays)},
		{"marked barriers", `SELECT node_id FROM public.access_barrier WHERE marked`, sortedKeys(wantBarriers)},
		{"ways stored", `SELECT way_id FROM public.access_way`, sortedKeys(storedWays)},
		{"barriers stored", `SELECT node_id FROM public.access_barrier`, sortedKeys(storedNodes)},
	} {
		got := fx.ids(t, tc.sql)
		if missing, extra := diff(tc.want, got), diff(got, tc.want); len(missing) > 0 || len(extra) > 0 {
			t.Errorf("the table's %s differ from the extract's: missing %v, extra %v", tc.what, missing, extra)
		}

		t.Logf("%s: %d in the table, %d in the independent reading", tc.what, len(got), len(tc.want))
	}

	ways := fx.ids(t, `SELECT way_id FROM public.access_way WHERE marked`)
	barriers := fx.ids(t, `SELECT node_id FROM public.access_barrier WHERE marked`)

	for _, id := range []int64{976307072, 25683224, 5179080} {
		if _, ok := slices.BinarySearch(ways, id); !ok {
			t.Errorf("way %d is not held as marked", id)
		}
	}

	if _, ok := slices.BinarySearch(barriers, 8691377189); !ok {
		t.Error("the lift_gate 8691377189 (access=residents) is not held as marked")
	}

	if _, ok := slices.BinarySearch(ways, 107135272); ok {
		t.Error("way 107135272 (access=no, motor_vehicle=yes) is held as marked, want it driven: the more specific key permits")
	}
}

// legCase is a route between two points on unmarked ways and the marking it
// is expected to cross.
type legCase struct {
	name       string
	from, to   planner.Point
	fromWay    int64
	toWay      int64
	kind       string
	feature    int64
	featureWay int64
	forward    [2]int64 // where set, two of featureWay's nodes, in its order, the route must run between
}

// The cases. Viktoriagatan's block 32709123 carries vehicle:backward=no and so
// is marked in both directions until direction is read (the Owner's decision
// on #258); the leg runs along it northbound between the blocks either side.
// Frönäs sjöväg 444469850 is the only road between Skärlinge skogsväg and its
// southern end, and carries the lift_gate 8691377189, access=residents, which
// the engine passes at a penalty.
var crossingCases = []legCase{
	{
		name: "a leg along a marked way",
		from: planner.Point{Latitude: 57.6995038, Longitude: 11.9647739}, fromWay: 491969022,
		to: planner.Point{Latitude: 57.7002529, Longitude: 11.9643691}, toWay: 1216125508,
		kind: "way", feature: 32709123, featureWay: 32709123,
	},
	{
		name: "a leg past a marked barrier",
		from: planner.Point{Latitude: 58.8966924, Longitude: 17.7973019}, fromWay: 590450309, // node 8691379001
		to: planner.Point{Latitude: 58.8866392, Longitude: 17.8024903}, toWay: 431248825, // node 4304890893
		kind: "barrier", feature: 8691377189, featureWay: 444469850,
	},
}

// TestAC2ARouteCrossingAMarkingYieldsNoLeg is #255 AC2 (`FR-125`, first
// criterion): a route the engine returns between two points on unmarked ways,
// along a marked way, and separately past a marked barrier, produces no leg,
// and the refusal names the feature and the edge of the route it lies on.
func TestAC2ARouteCrossingAMarkingYieldsNoLeg(t *testing.T) {
	fx := setUp(t)

	premise(t, fx.opl.ways, "way", 32709123, map[string]string{"vehicle:backward": "no"})
	premise(t, fx.opl.nodes, "node", 8691377189, map[string]string{"barrier": "lift_gate", "access": "residents"})

	if gate := premise(t, fx.opl.ways, "way", 444469850, nil); !slices.Contains(gate.nodes, 8691377189) || carries(gate.tags) {
		t.Fatalf("the premise does not hold: way 444469850 does not carry the lift_gate, or is itself marked (%v)", gate.tags)
	}

	for _, tc := range crossingCases {
		t.Run(tc.name, func(t *testing.T) { fx.refused(t, tc) })
	}
}

// refused plans the case's leg and fails unless the engine's own route runs
// along the case's way, no leg is produced from it, and the refusal names the
// case's feature at an edge along that way.
func (fx *fixture) refused(t *testing.T, tc legCase) {
	t.Helper()

	unmarkedEnd(t, fx.opl, tc.fromWay)
	unmarkedEnd(t, fx.opl, tc.toWay)

	// The engine's route runs where the case says, before the check.
	edges, err := fx.routing.Route(t.Context(), tc.from, tc.to, nil)
	if err != nil {
		t.Fatalf("the engine routed no leg: %v", err)
	}

	if !slices.ContainsFunc(edges, func(e planner.Edge) bool { return e.WayID == tc.featureWay }) {
		t.Fatalf("the premise does not hold: the engine's route runs along the ways %v, not along way %d", wayIDs(edges), tc.featureWay)
	}

	if tc.forward != [2]int64{} {
		fx.runsForward(t, edges, tc.featureWay, tc.forward)
	}

	journey, err := fx.planner(t, fx.routing, fx.store, fx.tiles).Plan(t.Context(), planner.Request{Origin: tc.from, Destination: tc.to})
	if journey.Legs != nil {
		t.Fatalf("the journey has %d legs, want none", len(journey.Legs))
	}

	var refusal *planner.CrossingError
	if !errors.As(err, &refusal) {
		t.Fatalf("Plan refused with %v, want a CrossingError", err)
	}

	found := false

	for _, c := range refusal.Crossings {
		if c.Kind != tc.kind || c.OSMID != tc.feature {
			continue
		}

		found = true

		if c.Edge < 0 || c.Edge >= len(refusal.Edges) || refusal.Edges[c.Edge].WayID != tc.featureWay {
			t.Errorf("the %s %d is named at edge %d, want an edge along way %d", tc.kind, tc.feature, c.Edge, tc.featureWay)
		}

		if tc.kind == "barrier" {
			fx.logDistance(t, tc.feature, refusal.Edges[c.Edge])
		}
	}

	if !found {
		t.Errorf("the refusal names %v, want the %s %d among them", refusal.Crossings, tc.kind, tc.feature)
	}

	t.Logf("refused: %v", err)
}

// runsForward fails the test unless the route runs along the way forward,
// from the first of the two nodes to the second: each point of the shape the
// route has along the way, in the route's order, lies nearest a node of the
// way no earlier in the way's order than the point before it, and the points
// span the two nodes. The engine cuts a way into edges at every junction on
// it, so it is the points, not the edges' ends, that are compared.
func (fx *fixture) runsForward(t *testing.T, edges []planner.Edge, way int64, nodes [2]int64) {
	t.Helper()

	ft := premise(t, fx.opl.ways, "way", way, nil)

	from, to := slices.Index(ft.nodes, nodes[0]), slices.Index(ft.nodes, nodes[1])
	if from < 0 || to <= from {
		t.Fatalf("the premise does not hold: way %d's nodes %v do not run from %d to %d", way, ft.nodes, nodes[0], nodes[1])
	}

	nearest := func(p planner.Point) int {
		best, at := 0.0, -1

		for i, id := range ft.nodes {
			n := fx.opl.nodes[id]
			if d := (p.Latitude-n.lat)*(p.Latitude-n.lat) + (p.Longitude-n.lon)*(p.Longitude-n.lon); at < 0 || d < best {
				best, at = d, i
			}
		}

		return at
	}

	var order []int

	for _, e := range edges {
		if e.WayID != way {
			continue
		}

		for _, p := range e.Shape {
			order = append(order, nearest(p))
		}
	}

	if len(order) < 2 || !slices.IsSorted(order) || order[0] > from || order[len(order)-1] < to {
		t.Fatalf("the premise does not hold: the engine's route runs along way %d through its nodes at %v, not forward from node %d (at %d) to node %d (at %d)", way, order, nodes[0], from, nodes[1], to)
	}
}

// logDistance logs how far the barrier lies from the edge's stretch: the
// measurement the barrier tolerance's value may be taken from.
func (fx *fixture) logDistance(t *testing.T, node int64, e planner.Edge) {
	t.Helper()

	points := make([]string, len(e.Shape))
	for i, p := range e.Shape {
		points[i] = fmt.Sprintf("%.7f %.7f", p.Longitude, p.Latitude)
	}

	var metres float64

	err := fx.pool.QueryRow(t.Context(),
		`SELECT ST_Distance(b.geom, ST_GeogFromText('SRID=4326;LINESTRING(' || $2 || ')')) FROM public.access_barrier b WHERE b.node_id = $1`,
		node, strings.Join(points, ",")).Scan(&metres)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("the barrier %d lies %.3f m from the stretch of the edge along way %d", node, metres, e.WayID)
}

// AC3's ends: nodes 7853731846 and 1231715095 of Ringvägen itself, neither a
// junction, so the leg between them runs along it. (Ends on the service roads
// either side of it gave a route along the road north of it instead, on the
// first CI run of #255, and the premise below failed rather than pass.)
var (
	ringvagenFrom = planner.Point{Latitude: 57.7164555, Longitude: 11.8293466}
	ringvagenTo   = planner.Point{Latitude: 57.7169564, Longitude: 11.8259839}
)

// wayIDs lists the ways a leg runs along, for a failure's message.
func wayIDs(edges []planner.Edge) []int64 {
	out := make([]int64, len(edges))
	for i, e := range edges {
		out[i] = e.WayID
	}

	return out
}

// TestAC3ARouteCrossingNoMarkingPassesTheCheck is #255 AC3 (`FR-125`, fifth
// criterion): a route along unmarked ways only, passing no marked barrier, one
// way on it carrying access=no with motor_vehicle=yes, is refused nothing.
func TestAC3ARouteCrossingNoMarkingPassesTheCheck(t *testing.T) {
	fx := setUp(t)

	// Ringvägen, in the extract's Torslanda part: access=no, motor_vehicle=yes,
	// which the independent reading leaves unmarked: the more specific key
	// permits.
	ringvagen := premise(t, fx.opl.ways, "way", 107135272, map[string]string{"access": "no", "motor_vehicle": "yes"})
	if marked(ringvagen.tags) {
		t.Fatalf("the premise does not hold: the independent reading marks Ringvägen (%v)", ringvagen.tags)
	}

	fx.passes(t, ringvagenFrom, ringvagenTo, 107135272, "access=no with motor_vehicle=yes")
}

// passes plans the leg between from and to and fails unless it is produced,
// runs along the case's way, runs along no way and past no node the
// independent reading marks, and the check finds nothing on it.
func (fx *fixture) passes(t *testing.T, from, to planner.Point, caseWay int64, carrying string) {
	t.Helper()

	journey, err := fx.planner(t, fx.routing, fx.store, fx.tiles).Plan(t.Context(), planner.Request{Origin: from, Destination: to})
	if err != nil {
		t.Fatalf("Plan refused a route along unmarked ways: %v", err)
	}

	if len(journey.Legs) != 1 {
		t.Fatalf("the journey has %d legs, want 1", len(journey.Legs))
	}

	leg := journey.Legs[0]

	// The premise on the route itself: it runs along the case's way, and every
	// way it runs along is one the independent reading leaves unmarked.
	if !slices.ContainsFunc(leg.Edges, func(e planner.Edge) bool { return e.WayID == caseWay }) {
		t.Fatalf("the premise does not hold: the route runs along the ways %v, not along %d, so it carries no %s", wayIDs(leg.Edges), caseWay, carrying)
	}

	t.Logf("the leg runs along the ways %v", wayIDs(leg.Edges))

	for _, e := range leg.Edges {
		if ft, ok := fx.opl.ways[e.WayID]; !ok || marked(ft.tags) {
			t.Errorf("the route runs along way %d, which the independent reading %s", e.WayID, map[bool]string{true: "marks", false: "does not hold"}[ok])
		}

		for _, n := range fx.opl.ways[e.WayID].nodes {
			if marked(fx.opl.nodes[n].tags) {
				t.Errorf("way %d of the route carries the marked node %d, so the case is not a route past no marked barrier", e.WayID, n)
			}
		}
	}

	crossings, err := fx.store.Crossings(t.Context(), leg.Edges)
	if err != nil || len(crossings) != 0 {
		t.Errorf("the check found %v, %v on the leg, want nothing", crossings, err)
	}
}

// everyLegAlong is a routing provider whose every route runs along the given
// way, by its own geometry.
type everyLegAlong struct {
	way   int64
	shape []planner.Point
	asked int
}

func (e *everyLegAlong) Route(context.Context, planner.Point, planner.Point, []planner.Point) ([]planner.Edge, error) {
	e.asked++

	return []planner.Edge{{WayID: e.way, Shape: e.shape}}, nil
}

// TestAC4NoRouteReachesAJourneyUnchecked is #255 AC4 (`FR-125`, first
// criterion) against the real table: with a routing provider whose every
// route runs along a way marked access=private, journeys of one leg and of
// several, through intermediate destinations, produce no leg.
func TestAC4NoRouteReachesAJourneyUnchecked(t *testing.T) {
	fx := setUp(t)

	const way = 976307072

	premise(t, fx.opl.ways, "way", way, map[string]string{"access": "private"})

	var lons, lats []float64
	if err := fx.pool.QueryRow(t.Context(),
		`SELECT array_agg(ST_X(p.geom) ORDER BY p.path), array_agg(ST_Y(p.geom) ORDER BY p.path)
		   FROM public.access_way w, ST_DumpPoints(w.geom::geometry) p WHERE w.way_id = $1`, way).Scan(&lons, &lats); err != nil {
		t.Fatal(err)
	}

	shape := make([]planner.Point, len(lons))
	for i := range lons {
		shape[i] = planner.Point{Latitude: lats[i], Longitude: lons[i]}
	}

	a := planner.Point{Latitude: 57.7059, Longitude: 11.9814}
	b := planner.Point{Latitude: 57.7047, Longitude: 11.9813}

	for _, req := range []planner.Request{
		{Origin: a, Destination: b},
		{Origin: a, Intermediates: []planner.Point{b, a}, Destination: b},
	} {
		routing := &everyLegAlong{way: way, shape: shape}

		journey, err := fx.planner(t, routing, fx.store, fx.tiles).Plan(t.Context(), req)
		if !errors.Is(err, planner.ErrCrossesMarking) || journey.Legs != nil {
			t.Errorf("a journey of %d legs gave %d legs and %v, want none and a crossing", len(req.Intermediates)+1, len(journey.Legs), err)
		}

		if routing.asked == 0 {
			t.Error("no route was obtained, so nothing was checked")
		}
	}
}

// TestAC5NoTableOrTheWrongOnePlansNoJourney is #255 AC5 (`FR-125`,
// statement): a table that is absent, empty, or loaded from an extract other
// than the one the routing data was built from, plans no journey.
func TestAC5NoTableOrTheWrongOnePlansNoJourney(t *testing.T) {
	fx := setUp(t)

	// A request whose route crosses no marking (AC3's), so a refusal here is
	// the table's and not the route's.
	req := planner.Request{Origin: ringvagenFrom, Destination: ringvagenTo}

	if _, err := fx.planner(t, fx.routing, fx.store, fx.tiles).Plan(t.Context(), req); err != nil {
		t.Fatalf("the premise does not hold: with the table loaded the journey is refused: %v", err)
	}

	plansNothing := func(t *testing.T, p *planner.Planner, want error) {
		t.Helper()

		journey, err := p.Plan(t.Context(), req)
		if !errors.Is(err, want) || journey.Legs != nil {
			t.Errorf("Plan gave %d legs and %v, want none and an error wrapping %v", len(journey.Legs), err, want)
		}
	}

	t.Run("loaded from another extract", func(t *testing.T) {
		// The extract this one was extended from, recorded in its README.
		plansNothing(t, fx.planner(t, fx.routing, fx.store, "0b878c989152be85784eea44c01e67d363140937629f27813caa0a3ac4f9aab9"), planner.ErrExtractMismatch)
	})

	t.Run("empty", func(t *testing.T) {
		t.Cleanup(func() { fx.load(t) })

		if _, err := fx.pool.Exec(t.Context(), `TRUNCATE public.access_way_exit, public.access_barrier_way, public.access_area, public.access_barrier, public.access_way`); err != nil {
			t.Fatal(err)
		}

		plansNothing(t, fx.planner(t, fx.routing, fx.store, fx.tiles), planner.ErrAccessUnchecked)
	})

	t.Run("no load current", func(t *testing.T) {
		t.Cleanup(func() { fx.load(t) })

		if _, err := fx.pool.Exec(t.Context(), `UPDATE public.osm_extract_load SET current = false WHERE current`); err != nil {
			t.Fatal(err)
		}

		plansNothing(t, fx.planner(t, fx.routing, fx.store, fx.tiles), planner.ErrAccessUnchecked)
	})

	t.Run("absent", func(t *testing.T) {
		// A database the migration was never applied to.
		const scratch = "turfgps_access_absent"

		if _, err := fx.pool.Exec(t.Context(), `CREATE DATABASE `+scratch); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_, _ = fx.pool.Exec(context.WithoutCancel(t.Context()), `DROP DATABASE IF EXISTS `+scratch+` WITH (FORCE)`)
		})

		cfg, err := pgxpool.ParseConfig(input(t, envDatabase))
		if err != nil {
			t.Fatal(err)
		}

		cfg.ConnConfig.Database = scratch

		bare, err := pgxpool.NewWithConfig(t.Context(), cfg)
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(bare.Close)

		store, err := accessstore.New(bare, &testTolerance)
		if err != nil {
			t.Fatal(err)
		}

		plansNothing(t, fx.planner(t, fx.routing, store, fx.tiles), planner.ErrAccessUnchecked)
	})

	t.Run("no barrier tolerance", func(t *testing.T) {
		store, err := accessstore.New(fx.pool, nil)
		if err != nil {
			t.Fatal(err)
		}

		plansNothing(t, fx.planner(t, fx.routing, store, fx.tiles), planner.ErrAccessUnchecked)
	})
}

// onlyKeys fails the test unless the feature carries, of the twelve access
// keys, exactly those named: the "only" of AC6 to AC8.
func onlyKeys(t *testing.T, kind string, id int64, ft feature, keys ...string) {
	t.Helper()

	for _, k := range chain {
		for _, form := range []string{k, k + ":forward", k + ":backward"} {
			if _, ok := ft.tags[form]; ok && !slices.Contains(keys, form) {
				t.Fatalf("the premise does not hold: %s %d carries %s=%q beside %v", kind, id, form, ft.tags[form], keys)
			}
		}
	}
}

// The synthetic cases (`testdata/README.md` § The synthetic cases): ways and
// nodes not from OpenStreetMap, which holds no instance of these tag sets
// (#261), written into the extract by hand under ids from syntheticBase up, a
// range OpenStreetMap's ids cannot reach. Each lies in the Kattegat, joined to
// nothing the extract holds from OpenStreetMap.
const (
	syntheticBase int64 = 9_000_000_000_000_000_000
	syntheticNote       = "synthetic, not from OpenStreetMap: TurfGPS #255 test case"

	// motorcar=yes, motor_vehicle:backward=no: AC6's third clause, AC7's third case.
	synMotorcarBackward = syntheticBase + 2
	// motor_vehicle=no, motor_vehicle:forward=yes, motor_vehicle:backward=yes: AC9's route.
	synMotorVehicleBoth = syntheticBase + 12
	// motorcar=no, motorcar:forward=yes, motorcar:backward=yes: AC9.
	synMotorcarBoth = syntheticBase + 14
	// motor_vehicle=no, motorcar:forward=yes, motorcar:backward=yes: AC9.
	synMotorVehicleMotorcarBoth = syntheticBase + 15
	// access=private, motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00): AC6's fifth clause.
	synPrivateConditional = syntheticBase + 21
)

// synthetic is the premise for a synthetic case: the way is in the extract
// with every tag named, carries no access key beyond those named, and is
// marked as not from OpenStreetMap, by its id's range and by its note.
func synthetic(t *testing.T, opl *oplExtract, id int64, tags map[string]string) {
	t.Helper()

	ft := premise(t, opl.ways, "way", id, tags)

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}

	onlyKeys(t, "way", id, ft, keys...)

	if id < syntheticBase || ft.tags["note"] != syntheticNote {
		t.Fatalf("the premise does not hold: way %d (note %q) is not one of the extract's synthetic cases", id, ft.tags["note"])
	}
}

// TestAC6TheTableReadsTheDirectionalKeylessAndConditionalForms is #255 AC6
// (`FR-125`, ninth to twelfth criteria, at the table): a way carrying only
// motor_vehicle:backward=no is held as marked, a node carrying access=private
// and no barrier key is held as a marked barrier, a way carrying motorcar=yes
// with motor_vehicle:backward=no is held as marked, a way carrying only
// motor_vehicle:conditional=no @ (Mo-Fr 08:00-18:00) is not held as marked,
// and a way carrying access=private with motor_vehicle:conditional=yes @
// (Mo-Fr 08:00-18:00) is held as marked. The third and fifth are synthetic:
// OpenStreetMap holds no way tagged as either (#261).
func TestAC6TheTableReadsTheDirectionalKeylessAndConditionalForms(t *testing.T) {
	fx := setUp(t)

	backward := premise(t, fx.opl.ways, "way", 14175890, map[string]string{"motor_vehicle:backward": "no"})
	onlyKeys(t, "way", 14175890, backward, "motor_vehicle:backward")

	keyless := premise(t, fx.opl.nodes, "node", 312500257, map[string]string{"access": "private"})
	if b, ok := keyless.tags["barrier"]; ok {
		t.Fatalf("the premise does not hold: node 312500257 carries barrier=%q", b)
	}

	conditional := premise(t, fx.opl.ways, "way", 4313847, map[string]string{"motor_vehicle:conditional": "no @ (Mo-Fr 08:00-18:00)"})
	onlyKeys(t, "way", 4313847, conditional)

	synthetic(t, fx.opl, synMotorcarBackward, map[string]string{"motorcar": "yes", "motor_vehicle:backward": "no"})
	synthetic(t, fx.opl, synPrivateConditional, map[string]string{"access": "private", "motor_vehicle:conditional": "yes @ (Mo-Fr 08:00-18:00)"})

	ways := fx.ids(t, `SELECT way_id FROM public.access_way WHERE marked`)
	if _, ok := slices.BinarySearch(ways, 14175890); !ok {
		t.Error("way 14175890 (motor_vehicle:backward=no only) is not held as marked")
	}

	if _, ok := slices.BinarySearch(ways, synMotorcarBackward); !ok {
		t.Errorf("way %d (motorcar=yes, motor_vehicle:backward=no) is not held as marked: the backward form's marking is lost to the plain permission", synMotorcarBackward)
	}

	if _, ok := slices.BinarySearch(ways, synPrivateConditional); !ok {
		t.Errorf("way %d (access=private, motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00)) is not held as marked: a :conditional permission lifted the plain marking", synPrivateConditional)
	}

	if _, ok := slices.BinarySearch(ways, 4313847); ok {
		t.Error("way 4313847 (motor_vehicle:conditional=no @ (Mo-Fr 08:00-18:00) only) is held as marked, want it driven: a :conditional form is ignored")
	}

	var (
		barrier *string
		held    bool
	)

	err := fx.pool.QueryRow(t.Context(), `SELECT barrier, marked FROM public.access_barrier WHERE node_id = 312500257`).Scan(&barrier, &held)
	if err != nil {
		t.Fatalf("node 312500257 (access=private, no barrier key) is not in access_barrier: %v", err)
	}

	if !held || barrier != nil {
		t.Errorf("node 312500257 is held with barrier %v and marked %t, want no barrier key and marked", barrier, held)
	}
}

// AC7's cases. Falkenbergsgatan 14175890 runs from node 135472415, where
// Ebbe Lieberathsgatan crosses it, to node 135473091, where the rest of
// Falkenbergsgatan goes on west; the leg runs along it forward, the direction
// motor_vehicle:backward=no does not name. Node 312500257, access=private and
// no barrier key, lies on way 28443588 between that way's dead end and
// Krusegatan; the leg runs from the dead end's stretch to Krusegatan. The
// synthetic way synMotorcarBackward runs east between two unmarked synthetic
// ways, and the leg runs along it east, forward: the direction the engine,
// reading motor_vehicle:backward=no as a one-way, routes it.
var directionalCases = []legCase{
	{
		name: "a leg along a way marked only by motor_vehicle:backward=no, run forward",
		from: planner.Point{Latitude: 57.6810412, Longitude: 11.9998148}, fromWay: 733746795, // between nodes 135472415 and 1311282119
		to: planner.Point{Latitude: 57.6807610, Longitude: 11.9985051}, toWay: 228006263, // between nodes 135473091 and 3530862679
		kind: "way", feature: 14175890, featureWay: 14175890, forward: [2]int64{135472415, 135473091},
	},
	{
		name: "a leg past a node carrying access=private and no barrier key",
		from: planner.Point{Latitude: 55.6226584, Longitude: 13.0530011}, fromWay: 28443588, // between nodes 312500255 and 312500257
		to: planner.Point{Latitude: 55.6192836, Longitude: 13.0495794}, toWay: 28288767, // between nodes 312500253 and 10188608433
		kind: "barrier", feature: 312500257, featureWay: 28443588,
	},
	{
		name: "a leg along a way carrying motorcar=yes with motor_vehicle:backward=no",
		from: planner.Point{Latitude: 57.0000000, Longitude: 11.7015000}, fromWay: syntheticBase + 1, // between nodes +101 and +102
		to: planner.Point{Latitude: 57.0000000, Longitude: 11.7075000}, toWay: syntheticBase + 3, // between nodes +104 and +105
		kind: "way", feature: synMotorcarBackward, featureWay: synMotorcarBackward, forward: [2]int64{syntheticBase + 102, syntheticBase + 104},
	},
}

// TestAC7ARouteCrossingSuchAMarkingYieldsNoLeg is #255 AC7 (`FR-125`, ninth
// to eleventh criteria): a route the engine returns between two points on
// unmarked ways, along a way carrying only motor_vehicle:backward=no in its
// forward direction, separately past a node carrying access=private and no
// barrier key, and separately along a way carrying motorcar=yes with
// motor_vehicle:backward=no, produces no leg, and the refusal names the
// feature and the edge it lies on. The third is synthetic: OpenStreetMap holds
// no way tagged so (#261).
func TestAC7ARouteCrossingSuchAMarkingYieldsNoLeg(t *testing.T) {
	fx := setUp(t)

	backward := premise(t, fx.opl.ways, "way", 14175890, map[string]string{"motor_vehicle:backward": "no"})
	onlyKeys(t, "way", 14175890, backward, "motor_vehicle:backward")

	keyless := premise(t, fx.opl.nodes, "node", 312500257, map[string]string{"access": "private"})
	if b, ok := keyless.tags["barrier"]; ok {
		t.Fatalf("the premise does not hold: node 312500257 carries barrier=%q", b)
	}

	if road := premise(t, fx.opl.ways, "way", 28443588, nil); !slices.Contains(road.nodes, 312500257) || carries(road.tags) {
		t.Fatalf("the premise does not hold: way 28443588 does not carry node 312500257, or is itself marked (%v)", road.tags)
	}

	synthetic(t, fx.opl, synMotorcarBackward, map[string]string{"motorcar": "yes", "motor_vehicle:backward": "no"})

	for _, tc := range directionalCases {
		t.Run(tc.name, func(t *testing.T) { fx.refused(t, tc) })
	}
}

// TestAC8AConditionalValueAloneRefusesNothing is #255 AC8 (`FR-125`, twelfth
// criterion): a route along unmarked ways only, passing no marked barrier, one
// way on it carrying only motor_vehicle:conditional=no @ (Mo-Fr 08:00-18:00),
// is refused nothing. Mornington Road's ends are Anglesea Place, which meets
// it at node 26149484, and the rest of Mornington Road, from node 26149485.
func TestAC8AConditionalValueAloneRefusesNothing(t *testing.T) {
	fx := setUp(t)

	conditional := premise(t, fx.opl.ways, "way", 4313847, map[string]string{"motor_vehicle:conditional": "no @ (Mo-Fr 08:00-18:00)"})
	onlyKeys(t, "way", 4313847, conditional)

	unmarkedEnd(t, fx.opl, 4315572)
	unmarkedEnd(t, fx.opl, 462835349)

	fx.passes(t,
		planner.Point{Latitude: 51.4684353, Longitude: -2.6163190}, // Anglesea Place, between nodes 26163964 and 26149484
		planner.Point{Latitude: 51.4686804, Longitude: -2.6146257}, // Mornington Road 462835349, between nodes 26149485 and 26149486
		4313847, "motor_vehicle:conditional=no @ (Mo-Fr 08:00-18:00)")
}

// TestAC9ADirectionalPermissionLiftsNoPlainMarking is #255 AC9 (`FR-125`,
// thirteenth criterion; `RD-066`): the ways carrying motor_vehicle=no with
// motor_vehicle:forward=yes and motor_vehicle:backward=yes, motorcar=no with
// motorcar:forward=yes and motorcar:backward=yes, and motor_vehicle=no with
// motorcar:forward=yes and motorcar:backward=yes are held as marked, and a
// route the engine returns along the first, between two points on unmarked
// ways, produces no leg. All three are synthetic: OpenStreetMap holds no way
// tagged as any of them (#261). The engine routes the first as a road a car
// may use (Valhalla 3.9.1's graph.lua lets motor_vehicle:forward and
// :backward override motor_vehicle), so the check is what refuses it.
func TestAC9ADirectionalPermissionLiftsNoPlainMarking(t *testing.T) {
	fx := setUp(t)

	cases := []struct {
		id   int64
		tags map[string]string
	}{
		{synMotorVehicleBoth, map[string]string{"motor_vehicle": "no", "motor_vehicle:forward": "yes", "motor_vehicle:backward": "yes"}},
		{synMotorcarBoth, map[string]string{"motorcar": "no", "motorcar:forward": "yes", "motorcar:backward": "yes"}},
		{synMotorVehicleMotorcarBoth, map[string]string{"motor_vehicle": "no", "motorcar:forward": "yes", "motorcar:backward": "yes"}},
	}

	for _, c := range cases {
		synthetic(t, fx.opl, c.id, c.tags)
	}

	ways := fx.ids(t, `SELECT way_id FROM public.access_way WHERE marked`)
	for _, c := range cases {
		if _, ok := slices.BinarySearch(ways, c.id); !ok {
			t.Errorf("way %d (%v) is not held as marked: a directional permission lifted the plain marking", c.id, c.tags)
		}
	}

	// The synthetic way runs east between two unmarked synthetic ways.
	fx.refused(t, legCase{
		name: "a leg along a way carrying motor_vehicle=no with motor_vehicle:forward=yes and motor_vehicle:backward=yes",
		from: planner.Point{Latitude: 57.0100000, Longitude: 11.7015000}, fromWay: syntheticBase + 11, // between nodes +201 and +202
		to: planner.Point{Latitude: 57.0100000, Longitude: 11.7075000}, toWay: syntheticBase + 13, // between nodes +204 and +205
		kind: "way", feature: synMotorVehicleBoth, featureWay: synMotorVehicleBoth,
	})
}
