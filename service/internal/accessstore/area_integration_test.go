//go:build valhalla

package accessstore_test

// #253's acceptance criteria in the phase the lead's ruling on #253 accepted:
// the area features loaded into access_area from closed ways and multipolygon
// relations, and AC1, AC2, AC5, AC7, AC8 and AC9 against the committed
// extract, the tiles a real Valhalla serves built from the same file, and the
// extract read independently as OPL, as integration_test.go (#255) does, whose
// helpers these share.
//
// AC1 IS READ FAIL-CLOSED, as the ruling reads it: a leg whose answer crosses
// such an area is refused, so no leg runs inside the area. The planner does not
// yet ask for a leg again avoiding an area (`planner.go`, #253's second phase),
// so where a connection round the area exists the leg is still not produced:
// a lost leg, never an unsafe one. AC6, and AC1 answered with a produced leg,
// are that phase's; AC3 and AC4, the run and a stop inside such an area, the
// third's.
//
// EACH TEST ESTABLISHES ITS PREMISE FROM THE EXTRACT FIRST: that the area
// feature is there with its marking and no other access key, that the ways it
// encloses carry no marking of their own, that the leg's ends lie outside it,
// that the way inside it lies on the quickest connection between them, and,
// where the case names one, that a connection round it is in the file, outside
// it and longer. A run against an extract that no longer holds the case fails
// there rather than passing vacuously.
//
// The cases are synthetic (`testdata/README.md`, #253's cases): each criterion
// names a road layout through an area, which a real one found by search would
// bring with a neighbourhood of roads whose quickest connection nothing here
// controls, so each is a small road network of its own in the Kattegat,
// joined to nothing.

import (
	"errors"
	"math"
	"slices"
	"strconv"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

const syntheticNote253 = "synthetic, not from OpenStreetMap: TurfGPS #253 test case"

// The synthetic area cases' ids and places, as the generator wrote them: case
// c's nodes are syntheticBase + 20000 + 100c + k, its ways syntheticBase + 3000
// + 20c + k, its relation, where it has one, syntheticBase + 100 + c, and it
// lies along 56.800 + 0.006c ° N. Cases 21 to 23 are S1's and S2's on PR #296.
func areaNode(c, k int) int64 { return syntheticBase + 20000 + int64(c)*100 + int64(k) }
func areaWay(c, k int) int64  { return syntheticBase + 3000 + int64(c)*20 + int64(k) }
func areaRel(c int) int64     { return syntheticBase + 100 + int64(c) }
func areaLat(c int) float64   { return 56.800 + float64(c)*0.006 }

// areaCase is a road of five ways east — two outside the area, one entering
// it, the way inside it, one leaving it, and one outside again — between
// points on the first and the last, and, where detour is set, three ways round
// the area to the north. The area is a closed way or a relation, of type
// relType, multipolygon where it is empty, enclosing the middle way; a car park
// unless its tags say what else.
type areaCase struct {
	c        int
	name     string
	tags     map[string]string
	detour   bool
	relation bool
	relType  string
}

func (ac areaCase) relationType() string {
	if ac.relType == "" {
		return "multipolygon"
	}

	return ac.relType
}

func (ac areaCase) from() planner.Point {
	return planner.Point{Latitude: areaLat(ac.c), Longitude: 11.7010}
}
func (ac areaCase) to() planner.Point {
	return planner.Point{Latitude: areaLat(ac.c), Longitude: 11.7080}
}

// inside is the way the area encloses, which the quickest connection runs along.
func (ac areaCase) inside() int64 { return areaWay(ac.c, 3) }

// feature is the area feature as access_area names it.
func (ac areaCase) feature() (string, int64) {
	if ac.relation {
		return "r", areaRel(ac.c)
	}

	return "w", areaWay(ac.c, 10)
}

const when = "@ (Mo-Fr 08:00-18:00)"

var (
	ac1Area  = areaCase{c: 0, name: "a customers-only car park drawn as an area, its aisle unmarked", tags: map[string]string{"access": "customers"}, detour: true}
	ac2Areas = []areaCase{
		ac1Area,
		{c: 1, name: "access=private", tags: map[string]string{"access": "private"}, detour: true},
		{c: 2, name: "access=destination", tags: map[string]string{"access": "destination"}, detour: true},
		{c: 3, name: "access=delivery", tags: map[string]string{"access": "delivery"}, detour: true},
		{c: 4, name: "access=permit", tags: map[string]string{"access": "permit"}, detour: true},
		{c: 5, name: "access=residents", tags: map[string]string{"access": "residents"}, detour: true},
		{c: 6, name: "access=restricted", tags: map[string]string{"access": "restricted"}, detour: true},
		{c: 7, name: "access=no", tags: map[string]string{"access": "no"}, detour: true},
		{c: 8, name: "access=staff, a value no table lists", tags: map[string]string{"access": "staff"}, detour: true},
		{c: 9, name: "motor_vehicle=private with access=yes", tags: map[string]string{"motor_vehicle": "private", "access": "yes"}, detour: true},
		{c: 13, name: "motorcar=private with access=yes", tags: map[string]string{"motorcar": "private", "access": "yes"}, detour: true},
		{c: 14, name: "vehicle=private with access=yes", tags: map[string]string{"vehicle": "private", "access": "yes"}, detour: true},
		{c: 10, name: "access=private with bicycle=yes and foot=yes", tags: map[string]string{"access": "private", "bicycle": "yes", "foot": "yes"}, detour: true},
		{c: 11, name: "motor_vehicle=private with psv=yes", tags: map[string]string{"motor_vehicle": "private", "psv": "yes"}, detour: true},
		{c: 12, name: "access=private on a multipolygon relation", tags: map[string]string{"access": "private"}, detour: true, relation: true},
		{c: 21, name: "access=no on a type=boundary relation, a nature reserve", tags: map[string]string{"access": "no", "boundary": "protected_area", "leisure": "nature_reserve"},
			detour: true, relation: true, relType: "boundary"},
		{c: 22, name: "access=private on a closed highway=rest_area way without area=yes", tags: map[string]string{"access": "private", "highway": "rest_area"}, detour: true},
		{c: 23, name: "access=private on a closed highway=services way without area=yes", tags: map[string]string{"access": "private", "highway": "services"}, detour: true},
	}
	permittedArea = areaCase{c: 15, name: "access=no with motor_vehicle=yes", tags: map[string]string{"access": "no", "motor_vehicle": "yes"}}
	ac5Area       = areaCase{c: 16, name: "access=private, with no connection round it", tags: map[string]string{"access": "private"}}
	ac7Areas      = []areaCase{
		{c: 17, name: "only motor_vehicle:backward=no", tags: map[string]string{"motor_vehicle:backward": "no"}, detour: true},
		{c: 18, name: "motorcar=yes with motor_vehicle:backward=no", tags: map[string]string{"motorcar": "yes", "motor_vehicle:backward": "no"}, detour: true},
	}
	ac8Area = areaCase{c: 19, name: "only access:conditional=private " + when, tags: map[string]string{"access:conditional": "private " + when}}
	ac9Area = areaCase{c: 20, name: "access=private with motor_vehicle:conditional=yes " + when,
		tags: map[string]string{"access": "private", "motor_vehicle:conditional": "yes " + when}, detour: true}
)

// --- the independent reading of areas --------------------------------------

type areaKey struct {
	kind string
	id   int64
}

// oplAreas is the area features the extract holds carrying any of the twelve
// keys, read from the OPL as `Architecture.md § What the tables hold` names
// them, and each one's linework: every closed way without a `highway` key,
// tagged `area=yes`, or whose `highway` value the OpenStreetMap wiki documents
// on areas and never on ways, and every `type=multipolygon` and
// `type=boundary` relation, whatever its members' roles.
func oplAreas(opl *oplExtract) map[areaKey]feature {
	out := map[areaKey]feature{}

	areaOnly := map[string]bool{"rest_area": true, "services": true, "trailhead": true, "turntable": true}

	for id, ft := range opl.ways {
		highway, ok := ft.tags["highway"]
		closed := len(ft.nodes) >= 2 && ft.nodes[0] == ft.nodes[len(ft.nodes)-1]

		if closed && carries(ft.tags) && (!ok || ft.tags["area"] == "yes" || areaOnly[highway]) {
			out[areaKey{"w", id}] = ft
		}
	}

	for id, ft := range opl.relations {
		if (ft.tags["type"] == "multipolygon" || ft.tags["type"] == "boundary") && carries(ft.tags) {
			out[areaKey{"r", id}] = ft
		}
	}

	return out
}

// linework is the ways an area is drawn from.
func linework(k areaKey, ft feature) []int64 {
	if k.kind == "w" {
		return []int64{k.id}
	}

	var ways []int64

	for _, m := range ft.members {
		if m.kind == 'w' {
			ways = append(ways, m.ref)
		}
	}

	return ways
}

// --- premises ---------------------------------------------------------------

// synthetic253 fails the test unless the feature is one of #253's synthetic
// cases with every tag named and no access key beyond them.
func synthetic253(t *testing.T, set map[int64]feature, kind string, id int64, tags map[string]string) feature {
	t.Helper()

	ft := premise(t, set, kind, id, tags)

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}

	onlyKeys(t, kind, id, ft, keys...)

	if id < syntheticBase || ft.tags["note"] != syntheticNote253 {
		t.Fatalf("the premise does not hold: %s %d (note %q) is not one of #253's synthetic cases", kind, id, ft.tags["note"])
	}

	return ft
}

// box is a rectangle in degrees.
type box struct{ west, south, east, north float64 }

func (b box) holds(lon, lat float64) bool {
	return lon > b.west && lon < b.east && lat > b.south && lat < b.north
}

// ring fails the test unless the nodes form the rectangle of their bounding
// box — every one at a corner of it, and every corner taken — and returns the
// box, inside which a point is then inside the ring.
func (fx *fixture) ring(t *testing.T, nodes []int64) box {
	t.Helper()

	b := box{west: math.Inf(1), south: math.Inf(1), east: math.Inf(-1), north: math.Inf(-1)}

	for _, id := range nodes {
		n, ok := fx.opl.nodes[id]
		if !ok {
			t.Fatalf("the premise does not hold: the area's node %d is not in the extract", id)
		}

		b.west, b.east = math.Min(b.west, n.lon), math.Max(b.east, n.lon)
		b.south, b.north = math.Min(b.south, n.lat), math.Max(b.north, n.lat)
	}

	corners := map[[2]float64]bool{}

	for _, id := range nodes {
		n := fx.opl.nodes[id]
		if (n.lon != b.west && n.lon != b.east) || (n.lat != b.south && n.lat != b.north) {
			t.Fatalf("the premise does not hold: the area's node %d (%.7f, %.7f) is not a corner of its box %+v", id, n.lat, n.lon, b)
		}

		corners[[2]float64{n.lon, n.lat}] = true
	}

	if len(corners) != 4 || !(b.west < b.east && b.south < b.north) {
		t.Fatalf("the premise does not hold: the area's ring %v is not a rectangle", nodes)
	}

	return b
}

// road fails the test unless the synthetic way carries no access key and runs
// between the two nodes, in that order, and returns it.
func (fx *fixture) road(t *testing.T, id, from, to int64) feature {
	t.Helper()

	ft := synthetic253(t, fx.opl.ways, "way", id, nil)
	if carries(ft.tags) || len(ft.nodes) < 2 || ft.nodes[0] != from || ft.nodes[len(ft.nodes)-1] != to {
		t.Fatalf("the premise does not hold: way %d carries %v and runs %v, want no access key and %d to %d", id, ft.tags, ft.nodes, from, to)
	}

	return ft
}

// areaPremise establishes an area case from the extract: the area feature with
// its tags, marked as wantMarked; the way it encloses, and the road in and out,
// carrying no access key; the leg's ends outside it; any connection round it
// outside it and longer, along ways of the same class.
func (fx *fixture) areaPremise(t *testing.T, ac areaCase, wantMarked bool) {
	t.Helper()

	kind, id := ac.feature()

	var (
		ft    feature
		outer []int64
	)

	if ac.relation {
		ft = synthetic253(t, fx.opl.relations, "relation", id, ac.tags)
		if ft.tags["type"] != ac.relationType() {
			t.Fatalf("the premise does not hold: relation %d is type=%q, want %s", id, ft.tags["type"], ac.relationType())
		}

		for _, m := range ft.members {
			if m.kind == 'w' && m.role == "outer" {
				outer = append(outer, fx.opl.ways[m.ref].nodes...)
			}
		}
	} else {
		ft = synthetic253(t, fx.opl.ways, "way", id, ac.tags)
		if len(ft.nodes) < 4 || ft.nodes[0] != ft.nodes[len(ft.nodes)-1] {
			t.Fatalf("the premise does not hold: way %d runs %v, want a closed way", id, ft.nodes)
		}

		if got, ok := ft.tags["highway"]; ok && got != ac.tags["highway"] {
			t.Fatalf("the premise does not hold: the area way %d carries highway=%s, want %q", id, got, ac.tags["highway"])
		}

		if _, ok := ft.tags["area"]; ok {
			t.Fatalf("the premise does not hold: the area way %d carries area=%s, want no area key", id, ft.tags["area"])
		}

		outer = ft.nodes
	}

	if marked(ft.tags) != wantMarked {
		t.Fatalf("the premise does not hold: the independent reading gives marked=%v for %v, want %v", marked(ft.tags), ft.tags, wantMarked)
	}

	area := fx.ring(t, outer)

	in := func(n int64) bool { return area.holds(fx.opl.nodes[n].lon, fx.opl.nodes[n].lat) }

	fx.road(t, areaWay(ac.c, 1), areaNode(ac.c, 1), areaNode(ac.c, 2))
	fx.road(t, areaWay(ac.c, 2), areaNode(ac.c, 2), areaNode(ac.c, 3))
	fx.road(t, areaWay(ac.c, 4), areaNode(ac.c, 5), areaNode(ac.c, 6))
	fx.road(t, areaWay(ac.c, 5), areaNode(ac.c, 6), areaNode(ac.c, 7))

	for _, n := range fx.road(t, ac.inside(), areaNode(ac.c, 3), areaNode(ac.c, 5)).nodes {
		if !in(n) {
			t.Fatalf("the premise does not hold: node %d of the way %d lies outside the area %+v", n, ac.inside(), area)
		}
	}

	for _, n := range []int64{areaNode(ac.c, 1), areaNode(ac.c, 2), areaNode(ac.c, 6), areaNode(ac.c, 7)} {
		if in(n) {
			t.Fatalf("the premise does not hold: node %d, on the road outside the area, lies inside it", n)
		}
	}

	for _, p := range []planner.Point{ac.from(), ac.to()} {
		if area.holds(p.Longitude, p.Latitude) {
			t.Fatalf("the premise does not hold: the leg's end %+v lies inside the area", p)
		}
	}

	if ac.detour {
		fx.road(t, areaWay(ac.c, 6), areaNode(ac.c, 2), areaNode(ac.c, 8))
		fx.road(t, areaWay(ac.c, 7), areaNode(ac.c, 8), areaNode(ac.c, 9))
		fx.road(t, areaWay(ac.c, 8), areaNode(ac.c, 9), areaNode(ac.c, 6))

		for _, n := range []int64{areaNode(ac.c, 8), areaNode(ac.c, 9)} {
			if in(n) {
				t.Fatalf("the premise does not hold: node %d, on the connection round the area, lies inside it", n)
			}
		}

		fx.shorter(t, []int64{areaWay(ac.c, 2), ac.inside(), areaWay(ac.c, 4)}, []int64{areaWay(ac.c, 6), areaWay(ac.c, 7), areaWay(ac.c, 8)})
	} else if _, ok := fx.opl.ways[areaWay(ac.c, 6)]; ok {
		t.Fatalf("the premise does not hold: case %d has a way round its area, want none", ac.c)
	}

	t.Logf("the premise holds: %s%d %v encloses way %d", kind, id, ac.tags, ac.inside())
}

// --- what is measured -------------------------------------------------------

// heldAs reports how access_area holds the case's area: whether it holds a row,
// and whether that row is marked.
func (fx *fixture) heldAs(t *testing.T, ac areaCase) (held, isMarked bool) {
	t.Helper()

	kind, id := ac.feature()

	rows, err := fx.pool.Query(t.Context(), `SELECT marked FROM public.access_area WHERE osm_type = $1 AND osm_id = $2`, kind, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		held = true

		if err := rows.Scan(&isMarked); err != nil {
			t.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	return held, isMarked
}

// refusedForTheArea plans the case's leg and fails unless no leg is produced,
// the answer refused ran along the way inside the area — the case's premise,
// on the engine's own route — and the refusal names the case's area and
// nothing else.
func (fx *fixture) refusedForTheArea(t *testing.T, ac areaCase) {
	t.Helper()

	fx.areaPremise(t, ac, true)

	if held, isMarked := fx.heldAs(t, ac); !held || !isMarked {
		t.Errorf("access_area holds the case's area as held=%v marked=%v, want a marked row", held, isMarked)
	}

	p, routing := fx.avoiding(t)

	journey, err := p.Plan(t.Context(), planner.Request{Origin: ac.from(), Destination: ac.to()})
	if journey.Legs != nil {
		t.Fatalf("the journey has %d leg(s), running along %v, want none: a leg through the area was produced", len(journey.Legs), wayIDs(journey.Legs[0].Edges))
	}

	var refusal *planner.CrossingError
	if !errors.As(err, &refusal) {
		t.Fatalf("Plan refused with %v, want a CrossingError", err)
	}

	if !slices.Contains(wayIDs(refusal.Edges), ac.inside()) {
		t.Fatalf("the premise does not hold: the engine's route runs along %v, not along the way %d inside the area", wayIDs(refusal.Edges), ac.inside())
	}

	kind, id := ac.feature()
	found := false

	for _, c := range refusal.Crossings {
		if c.Kind == "area" && c.OSMType == kind && c.OSMID == id {
			found = true

			continue
		}

		t.Errorf("the refusal names %v, which is not the case's area %s%d: the case holds no other marking", c, kind, id)
	}

	if !found {
		t.Errorf("the refusal names %v, want the area %s%d among them", refusal.Crossings, kind, id)
	}

	t.Logf("refused: %v; the engine was asked %d time(s)", err, len(routing.avoids))
}

// drivenThrough plans the case's leg and fails unless it is produced along the
// way inside the area feature, crossing nothing the reading or the check
// marks.
func (fx *fixture) drivenThrough(t *testing.T, ac areaCase) {
	t.Helper()

	fx.areaPremise(t, ac, false)

	if _, isMarked := fx.heldAs(t, ac); isMarked {
		t.Errorf("access_area holds the case's area as marked, want it unmarked or not held")
	}

	fx.driven(t, ac.from(), ac.to(), ac.inside())
}

// --- the criteria -----------------------------------------------------------

// TestTheAreaTableHoldsExactlyTheExtractsAreaFeatures is the loader's half of
// #253 (`FR-128`, first criterion, at the table): access_area holds every area
// feature the independent reading takes from the extract, and nothing else,
// each marked as that reading marks it, each drawn through every node of its
// linework, with the load recording as many areas drawn as read. A
// multipolygon's inner ring is inside its area: the loader fills it, which
// refuses more legs and never fewer. Case 0's car park, a rectangle, is the
// coordinate guard: its area is that of the rectangle, which it would not be
// with the axes swapped.
func TestTheAreaTableHoldsExactlyTheExtractsAreaFeatures(t *testing.T) {
	fx := setUp(t)

	// The premises: the cases the criteria name are in the extract, among
	// them a closed way and a multipolygon relation from OpenStreetMap.
	premise(t, fx.opl.ways, "way", 1088367799, map[string]string{"access": "customers", "amenity": "parking"})
	premise(t, fx.opl.relations, "relation", 10476740, map[string]string{"type": "multipolygon", "access": "yes"})
	synthetic253(t, fx.opl.relations, "relation", areaRel(12), map[string]string{"type": "multipolygon", "access": "private"})
	synthetic253(t, fx.opl.relations, "relation", areaRel(21), map[string]string{"type": "boundary", "access": "no"})
	synthetic253(t, fx.opl.ways, "way", areaWay(22, 10), map[string]string{"highway": "rest_area", "access": "private"})
	synthetic253(t, fx.opl.ways, "way", areaWay(23, 10), map[string]string{"highway": "services", "access": "private"})
	synthetic253(t, fx.opl.ways, "way", areaWay(ac8Area.c, 10), ac8Area.tags)

	areas := oplAreas(fx.opl)

	var stored, wantMarked []string

	for k, ft := range areas {
		stored = append(stored, k.kind+itoa(k.id))
		if marked(ft.tags) {
			wantMarked = append(wantMarked, k.kind+itoa(k.id))
		}
	}

	slices.Sort(stored)
	slices.Sort(wantMarked)

	for _, tc := range []struct {
		what string
		sql  string
		want []string
	}{
		{"areas stored", `SELECT osm_type || osm_id FROM public.access_area ORDER BY 1`, stored},
		{"marked areas", `SELECT osm_type || osm_id FROM public.access_area WHERE marked ORDER BY 1`, wantMarked},
	} {
		got := fx.strings(t, tc.sql)
		if missing, extra := diffStrings(tc.want, got), diffStrings(got, tc.want); len(missing) > 0 || len(extra) > 0 {
			t.Errorf("the table's %s differ from the extract's: missing %v, extra %v", tc.what, missing, extra)
		}

		t.Logf("%s: %d in the table, %d in the independent reading", tc.what, len(got), len(tc.want))
	}

	var features, rows *int64
	if err := fx.pool.QueryRow(t.Context(), `SELECT area_features, area_rows FROM public.osm_extract_load WHERE current`).Scan(&features, &rows); err != nil {
		t.Fatal(err)
	}

	if features == nil || rows == nil || *features != int64(len(areas)) || *rows != int64(len(areas)) {
		t.Errorf("the load records %v areas read and %v drawn, want %d of each", deref(features), deref(rows), len(areas))
	}

	// Every node of every area's linework lies on the area drawn.
	var types []string

	var (
		ids        []int64
		lons, lats []float64
	)

	for k, ft := range areas {
		for _, w := range linework(k, ft) {
			way, ok := fx.opl.ways[w]
			if !ok {
				t.Fatalf("the independent reading holds no way %d, which %s%d names", w, k.kind, k.id)
			}

			for _, n := range way.nodes {
				types, ids = append(types, k.kind), append(ids, k.id)
				lons, lats = append(lons, fx.opl.nodes[n].lon), append(lats, fx.opl.nodes[n].lat)
			}
		}
	}

	off := fx.strings(t, `
SELECT p.osm_type || p.osm_id || ' at ' || p.lat || ' ' || p.lon
  FROM unnest($1::text[], $2::bigint[], $3::double precision[], $4::double precision[]) AS p(osm_type, osm_id, lon, lat)
  LEFT JOIN public.access_area a ON a.osm_type = p.osm_type AND a.osm_id = p.osm_id
 WHERE a.geom IS NULL OR NOT ST_DWithin(a.geom, ST_SetSRID(ST_MakePoint(p.lon, p.lat), 4326)::geography, 0.05)
 ORDER BY 1 LIMIT 20`, types, ids, lons, lats)
	if len(off) > 0 {
		t.Errorf("nodes of the areas' linework lie more than 5 cm from the area drawn: %v", off)
	}

	t.Logf("%d nodes of the areas' linework checked against the areas drawn", len(ids))

	// The hole of case 12's multipolygon is inside its area.
	hole := planner.Point{Latitude: areaLat(12) - 0.00015, Longitude: 11.7037}

	var covered bool
	if err := fx.pool.QueryRow(t.Context(),
		`SELECT ST_Covers(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) FROM public.access_area WHERE osm_type = 'r' AND osm_id = $3`,
		hole.Longitude, hole.Latitude, areaRel(12)).Scan(&covered); err != nil {
		t.Fatal(err)
	}

	if !covered {
		t.Error("the point inside case 12's inner ring is not inside its area, want the inner ring filled")
	}

	// The coordinate guard: case 0's car park spans 0.004° of longitude and
	// 0.0012° of latitude at 56.8° N.
	var m2 float64
	if err := fx.pool.QueryRow(t.Context(), `SELECT ST_Area(geom) FROM public.access_area WHERE osm_type = 'w' AND osm_id = $1`, areaWay(0, 10)).Scan(&m2); err != nil {
		t.Fatal(err)
	}

	lat := areaLat(0) * math.Pi / 180
	want := (0.004 * math.Pi / 180 * 6371008.8 * math.Cos(lat)) * (0.0012 * math.Pi / 180 * 6371008.8)

	if math.Abs(m2-want) > 0.02*want {
		t.Errorf("case 0's car park is drawn with an area of %.0f m², want %.0f m² within 2%%: its coordinates are not where the extract puts them", m2, want)
	}
}

// TestAC1AWayInsideAMarkedAreaIsNotDrivenInTransit is #253 AC1 (`FR-128`,
// first criterion) in its fail-closed reading: between two points outside a
// customers-only car park drawn as an area, whose quickest connection runs
// along an aisle inside it that carries no marking of its own, and between
// which a connection round it exists, no leg runs along the aisle: the leg
// crossing the area is refused.
func TestAC1AWayInsideAMarkedAreaIsNotDrivenInTransit(t *testing.T) {
	fx := setUp(t)
	fx.refusedForTheArea(t, ac1Area)
}

// TestAC2EveryValueOutsideThePermissiveSetMarksTheArea is #253 AC2 (`FR-128`,
// first criterion; `RD-060`, `RD-062`, `RD-063`): each listed value, a value no
// table lists, a marking on each key more specific than access beside
// access=yes, a marking beside a permissive value on keys that are none of the
// four, and a marking on a multipolygon relation rather than a closed way.
func TestAC2EveryValueOutsideThePermissiveSetMarksTheArea(t *testing.T) {
	fx := setUp(t)

	for _, ac := range ac2Areas {
		t.Run(ac.name, func(t *testing.T) { fx.refusedForTheArea(t, ac) })
	}
}

// TestAnAreaWhoseMostSpecificKeyPermitsIsNotSuchAnArea holds #253's
// definition of such an area (`RD-063`): keys less specific than the most
// specific one an area carries are not read, so an area carrying access=no
// with motor_vehicle=yes is driven through.
func TestAnAreaWhoseMostSpecificKeyPermitsIsNotSuchAnArea(t *testing.T) {
	fx := setUp(t)

	if held, _ := fx.heldAs(t, permittedArea); !held {
		t.Error("access_area does not hold the area carrying access=no with motor_vehicle=yes, want it held unmarked")
	}

	fx.drivenThrough(t, permittedArea)
}

// TestAC5WhereNothingAvoidsTheAreaNoLegRunsThroughIt is #253 AC5 (`FR-128`,
// first criterion): between two points outside an area marked
// access=private, every connection between which runs along a way inside it,
// no leg is produced.
func TestAC5WhereNothingAvoidsTheAreaNoLegRunsThroughIt(t *testing.T) {
	fx := setUp(t)
	fx.refusedForTheArea(t, ac5Area)
}

// TestAC7ADirectionalFormMarksAnArea is #253 AC7 (`FR-128`, first criterion;
// `RD-064`, `RD-065`): an area carrying only motor_vehicle:backward=no, and
// one carrying motorcar=yes with motor_vehicle:backward=no.
func TestAC7ADirectionalFormMarksAnArea(t *testing.T) {
	fx := setUp(t)

	for _, ac := range ac7Areas {
		t.Run(ac.name, func(t *testing.T) { fx.refusedForTheArea(t, ac) })
	}
}

// TestAC8AConditionalValueAloneMarksNoArea is #253 AC8 (`FR-128`, first
// criterion; `RD-064`): a way inside an area feature carrying only
// access:conditional=private @ (Mo-Fr 08:00-18:00) is driven along, and the
// table holds no row for the feature, which carries none of the twelve keys.
func TestAC8AConditionalValueAloneMarksNoArea(t *testing.T) {
	fx := setUp(t)

	if held, _ := fx.heldAs(t, ac8Area); held {
		t.Error("access_area holds the area carrying only access:conditional, want it not held: it carries none of the twelve keys")
	}

	fx.drivenThrough(t, ac8Area)
}

// TestAC9AConditionalPermissionLiftsNoAreaMarking is #253 AC9 (`FR-128`, first
// criterion; `RD-064`): an area carrying access=private with
// motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00).
func TestAC9AConditionalPermissionLiftsNoAreaMarking(t *testing.T) {
	fx := setUp(t)
	fx.refusedForTheArea(t, ac9Area)
}

// --- helpers ----------------------------------------------------------------

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// strings runs a query returning one text column, sorted as Go sorts, so a
// binary search over it does not depend on the database's collation.
func (fx *fixture) strings(t *testing.T, sql string, args ...any) []string {
	t.Helper()

	rows, err := fx.pool.Query(t.Context(), sql, args...)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var out []string

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}

		out = append(out, s)
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	slices.Sort(out)

	return out
}

func diffStrings(a, b []string) []string {
	var out []string

	for _, x := range a {
		if _, found := slices.BinarySearch(b, x); !found {
			out = append(out, x)
		}
	}

	return out
}

func deref(p *int64) any {
	if p == nil {
		return nil
	}

	return *p
}
