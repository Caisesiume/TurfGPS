package accessload

import (
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/osmpbf"
)

// extractPath is the committed extract the routing tiles are built from
// (`internal/valhalla/testdata/README.md`).
const extractPath = "../valhalla/testdata/goteborg-inom-vallgraven.osm.pbf"

func value(v Values, key string) (string, bool) {
	for i, k := range Keys {
		if k == key && v[i] != nil {
			return *v[i], true
		}
	}

	return "", false
}

// TestTheReaderTakesTheTwelveKeysVerbatim reads the committed extract and finds
// the features #255's criteria name, each with exactly the values the extract
// holds, and no `:conditional` form read as a key (the Owner's third tag
// decision on #258).
func TestTheReaderTakesTheTwelveKeysVerbatim(t *testing.T) {
	t.Parallel()

	ex, err := ReadFile(extractPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if ex.WayFeatures != int64(len(ex.Ways)) || ex.BarrierFeatures != int64(len(ex.Barriers)) {
		t.Fatalf("read %d way features into %d rows and %d barrier features into %d rows, want one row each", ex.WayFeatures, len(ex.Ways), ex.BarrierFeatures, len(ex.Barriers))
	}

	ways := map[int64]Way{}
	for _, w := range ex.Ways {
		ways[w.ID] = w
	}

	barriers := map[int64]Barrier{}
	for _, b := range ex.Barriers {
		barriers[b.ID] = b
	}

	for _, tc := range []struct {
		way  int64
		tags map[string]string
	}{
		{25683224, map[string]string{"access": "military"}},
		{5179080, map[string]string{"motor_vehicle": "private", "access": "yes"}},
		{107135272, map[string]string{"access": "no", "motor_vehicle": "yes"}},
		{32709123, map[string]string{"vehicle:backward": "no"}},
	} {
		w, ok := ways[tc.way]
		if !ok {
			t.Errorf("way %d is not read, want it among the ways carrying an access key", tc.way)

			continue
		}

		for k, want := range tc.tags {
			if got, ok := value(w.Values, k); !ok || got != want {
				t.Errorf("way %d's %s is %q (%v), want %q", tc.way, k, got, ok, want)
			}
		}

		if len(w.Lons) != len(w.Nodes) || len(w.Lats) != len(w.Nodes) || len(w.Nodes) < 2 {
			t.Errorf("way %d has %d nodes and %d/%d coordinates", tc.way, len(w.Nodes), len(w.Lons), len(w.Lats))
		}
	}

	gate, ok := barriers[8691377189]
	if !ok || gate.Barrier == nil || *gate.Barrier != "lift_gate" {
		t.Fatalf("node 8691377189 is read as %+v, want the lift_gate", gate)
	}

	if got, _ := value(gate.Values, "access"); got != "residents" {
		t.Errorf("the lift_gate's access is %q, want residents", got)
	}

	// A node carrying the keys without a `barrier` key is a barrier too (the
	// Owner's first tag decision on #258).
	if entrance, ok := barriers[1705784512]; !ok || entrance.Barrier != nil {
		t.Errorf("node 1705784512 (access=private, entrance=home) is read as %+v, want a barrier with no barrier key", entrance)
	}

	// Östra Larmgatan carries only motor_vehicle:conditional.
	if w, ok := ways[175304439]; ok {
		t.Errorf("way 175304439 carries only a :conditional form and was read as %+v, want it not taken", w)
	}

	var paired bool

	for _, p := range ex.BarrierWays {
		if p.Node == 8691377189 && p.Way == 444469850 {
			paired = true
		}
	}

	if !paired {
		t.Error("the lift_gate is not paired with way 444469850, which references it")
	}

	if len(ex.SHA256) != 64 {
		t.Errorf("the SHA-256 is %q", ex.SHA256)
	}
}

// TestAConditionalFormIsNoKey holds the Owner's third tag decision on #258
// (`RD-064`) at the reader: a `:conditional` form of an access key, plain or
// directional, is read as none of the twelve keys. A `:backward:conditional`
// permission is therefore not a more specific directional form, and lifts
// nothing a `:backward` key marks (N16 on #259); a conditional permission
// lifts no plain marking either (`FR-125`'s fourteenth criterion).
func TestAConditionalFormIsNoKey(t *testing.T) {
	t.Parallel()

	const when = "yes @ (Mo-Fr 08:00-18:00)"

	for _, tc := range []struct {
		name string
		tags map[string]string
		want map[string]string
	}{
		{
			"a :backward:conditional permission beside a :backward marking",
			map[string]string{"motor_vehicle:backward": "no", "motorcar:backward:conditional": when},
			map[string]string{"motor_vehicle:backward": "no"},
		},
		{
			"a :forward:conditional permission beside a :forward marking",
			map[string]string{"access:forward": "private", "motorcar:forward:conditional": when, "motor_vehicle:forward:conditional": when},
			map[string]string{"access:forward": "private"},
		},
		{
			"a plain conditional permission beside a plain marking",
			map[string]string{"access": "private", "motor_vehicle:conditional": when},
			map[string]string{"access": "private"},
		},
	} {
		got, carrying := valuesOf(tc.tags)
		if !carrying {
			t.Errorf("%s: %v is read as carrying no access key", tc.name, tc.tags)

			continue
		}

		for i, k := range Keys {
			want, named := tc.want[k]

			switch {
			case named && (got[i] == nil || *got[i] != want):
				t.Errorf("%s: %s is read as %v, want %q", tc.name, k, deref(got[i]), want)
			case !named && got[i] != nil:
				t.Errorf("%s: %s is read as %q from %v, want it absent: a :conditional form is not that key", tc.name, k, *got[i], tc.tags)
			}
		}
	}

	if _, carrying := valuesOf(map[string]string{"motorcar:backward:conditional": when}); carrying {
		t.Error("a feature carrying only motorcar:backward:conditional is read as carrying an access key")
	}
}

func deref(s *string) any {
	if s == nil {
		return nil
	}

	return *s
}

// TestAnUnsortedFileIsRefused: both passes rely on every node preceding every
// way, so a file in another order is refused rather than read with barriers
// missing from their ways.
func TestAnUnsortedFileIsRefused(t *testing.T) {
	t.Parallel()

	for name, h := range map[string]osmpbf.Handler{
		"the ways pass":  &waysPass{nodesOf: map[int64]bool{}},
		"the nodes pass": &nodesPass{wayNodes: map[int64]bool{}, coords: map[int64][2]float64{}, barrierIDs: map[int64]bool{}},
	} {
		if err := h.Way(osmpbf.Way{ID: 1}); err != nil {
			t.Fatalf("%s: a first way: %v", name, err)
		}

		if err := h.Node(osmpbf.Node{ID: 2}); err == nil {
			t.Errorf("%s: a node after a way was accepted", name)
		}
	}
}

// The synthetic area cases' ids (`testdata/README.md`, #253's cases).
const synthetic = 9_000_000_000_000_000_000

func synArea(c int) int64     { return synthetic + 3000 + int64(c)*20 + 10 }
func synRelation(c int) int64 { return synthetic + 100 + int64(c) }

// TestTheReaderTakesTheAreaFeatures reads the committed extract and finds the
// area features #253's criteria and the box name, each with the values the
// extract holds and all its linework: a closed way, a multipolygon relation
// from OpenStreetMap, and the synthetic relation whose outer ring is two ways
// and whose inner ring is a third. An area feature carrying only a
// `:conditional` form is not taken (`RD-064`).
func TestTheReaderTakesTheAreaFeatures(t *testing.T) {
	t.Parallel()

	ex, err := ReadFile(extractPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if ex.AreaFeatures != int64(len(ex.Areas)) || ex.AreaFeatures == 0 {
		t.Fatalf("read %d area features into %d rows, want one row each, and some", ex.AreaFeatures, len(ex.Areas))
	}

	areas := map[[2]int64]Area{}
	for _, a := range ex.Areas {
		areas[[2]int64{int64(a.Type), a.ID}] = a
	}

	for _, tc := range []struct {
		typ   byte
		id    int64
		tags  map[string]string
		lines []int64
	}{
		// A customers-only car park in the box, drawn as a closed way.
		{'w', 1088367799, map[string]string{"access": "customers"}, []int64{1088367799}},
		// A multi-storey car park in the box, a multipolygon of four ways.
		{'r', 10476740, map[string]string{"access": "yes"}, []int64{16239958, 758459162, 758459164, 758459161}},
		{'w', synArea(0), map[string]string{"access": "customers"}, []int64{synArea(0)}},
		{'r', synRelation(12), map[string]string{"access": "private"}, []int64{synArea(12), synArea(12) + 1, synArea(12) + 2}},
		// S1 on PR #296: a type=boundary relation, as Vårdsätra naturpark is.
		{'r', synRelation(21), map[string]string{"access": "no"}, []int64{synArea(21), synArea(21) + 1}},
		// S2 on PR #296: closed highway=rest_area and highway=services ways
		// without area=yes.
		{'w', synArea(22), map[string]string{"access": "private"}, []int64{synArea(22)}},
		{'w', synArea(23), map[string]string{"access": "private"}, []int64{synArea(23)}},
	} {
		a, ok := areas[[2]int64{int64(tc.typ), tc.id}]
		if !ok {
			t.Errorf("%c%d is not read, want it among the areas carrying an access key", tc.typ, tc.id)

			continue
		}

		for k, want := range tc.tags {
			if got, ok := value(a.Values, k); !ok || got != want {
				t.Errorf("%c%d's %s is %q (%v), want %q", tc.typ, tc.id, k, got, ok, want)
			}
		}

		if len(a.Lines) != len(tc.lines) {
			t.Errorf("%c%d is read with %d lines, want %d", tc.typ, tc.id, len(a.Lines), len(tc.lines))

			continue
		}

		for i, l := range a.Lines {
			if l.Way != tc.lines[i] || len(l.Lons) < 2 || len(l.Lons) != len(l.Lats) {
				t.Errorf("%c%d's line %d is way %d with %d/%d coordinates, want way %d", tc.typ, tc.id, i, l.Way, len(l.Lons), len(l.Lats), tc.lines[i])
			}
		}
	}

	if a, ok := areas[[2]int64{'w', synArea(19)}]; ok {
		t.Errorf("the area way %d carries only access:conditional and was read as %+v, want it not taken", synArea(19), a)
	}
}

// TestWhichWaysAreAreas holds the reader to the area features `Architecture.md
// § What the tables hold` names: a closed way without a `highway` key, a closed
// way tagged `area=yes`, and a closed way whose `highway` value is only ever an
// area (S2 on PR #296); another closed `highway` way without `area=yes` is a
// way, and an open way is never an area.
func TestWhichWaysAreAreas(t *testing.T) {
	t.Parallel()

	ring, open := []int64{1, 2, 3, 1}, []int64{1, 2, 3}

	for _, tc := range []struct {
		name  string
		nodes []int64
		tags  map[string]string
		want  bool
	}{
		{"a closed way without a highway key", ring, map[string]string{"amenity": "parking"}, true},
		{"a closed way with no tag but the access key", ring, map[string]string{"access": "private"}, true},
		{"a closed highway way tagged area=yes", ring, map[string]string{"highway": "pedestrian", "area": "yes"}, true},
		{"a closed highway way, a loop road", ring, map[string]string{"highway": "service"}, false},
		{"a closed highway way tagged area=no", ring, map[string]string{"highway": "service", "area": "no"}, false},
		{"a closed highway=rest_area way without area=yes", ring, map[string]string{"highway": "rest_area", "access": "private"}, true},
		{"a closed highway=services way without area=yes", ring, map[string]string{"highway": "services"}, true},
		{"a closed highway=trailhead way without area=yes", ring, map[string]string{"highway": "trailhead"}, true},
		{"a closed highway=turntable way without area=yes", ring, map[string]string{"highway": "turntable"}, true},
		{"a closed highway=rest_area way tagged area=no", ring, map[string]string{"highway": "rest_area", "area": "no"}, true},
		{"an open highway=rest_area way", open, map[string]string{"highway": "rest_area"}, false},
		{"an open way without a highway key", open, map[string]string{"amenity": "parking"}, false},
	} {
		if got := isArea(osmpbf.Way{ID: 1, Nodes: tc.nodes, Tags: tc.tags}); got != tc.want {
			t.Errorf("%s: isArea is %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestAnAreaThatCannotBeTakenWholeIsRefused: a marked area read as smaller
// than it is reads its missing part as unmarked, so every area the reader
// cannot take whole fails the read.
func TestAnAreaThatCannotBeTakenWholeIsRefused(t *testing.T) {
	t.Parallel()

	private := "private"
	values := Values{3: &private}

	nodes := &nodesPass{coords: map[int64][2]float64{
		1: {11.70, 57.00}, 2: {11.71, 57.00}, 3: {11.71, 57.01}, 4: {11.70, 57.01},
	}}

	members := map[int64][]int64{
		10: {1, 2, 3},
		11: {3, 4, 1},
		12: {3, 4},
		13: {3, 5, 1},
		14: {1},
	}

	if _, err := nodes.relationArea(areaRelation{id: 7, values: values, ways: []int64{10, 11}}, members); err != nil {
		t.Fatalf("the premise does not hold: two ways closing one ring are refused: %v", err)
	}

	for name, ways := range map[string][]int64{
		"a member way the file does not hold":    {10, 99},
		"linework that does not close":           {10, 12},
		"a member way whose node the file lacks": {10, 13},
		"a member way of one node":               {10, 11, 14},
	} {
		if _, err := nodes.relationArea(areaRelation{id: 7, values: values, ways: ways}, members); err == nil {
			t.Errorf("a multipolygon with %s was taken", name)
		}
	}

	for name, a := range map[string]areaWay{
		"a closed way of three nodes":        {id: 8, values: values, nodes: []int64{1, 2, 1}},
		"a closed way whose node is missing": {id: 8, values: values, nodes: []int64{1, 2, 6, 1}},
	} {
		if _, err := nodes.wayArea(a); err == nil {
			t.Errorf("an area that is %s was taken", name)
		}
	}

	if _, err := nodes.wayArea(areaWay{id: 8, values: values, nodes: []int64{1, 2, 3, 1}}); err != nil {
		t.Errorf("a closed way of three distinct nodes was refused: %v", err)
	}

	rels := &relationsPass{members: map[int64]bool{}}

	for name, r := range map[string]osmpbf.Relation{
		"a boundary naming a relation, a subarea": {ID: 7, Tags: map[string]string{"type": "boundary", "access": "no"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberWay, Ref: 10}, {Type: osmpbf.MemberRelation, Ref: 70, Role: "subarea"}}},
		"a boundary naming no way": {ID: 7, Tags: map[string]string{"type": "boundary", "vehicle": "no"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberNode, Ref: 1, Role: "label"}}},
		"a relation member": {ID: 7, Tags: map[string]string{"type": "multipolygon", "access": "private"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberWay, Ref: 10}, {Type: osmpbf.MemberRelation, Ref: 70}}},
		"no way member": {ID: 7, Tags: map[string]string{"type": "multipolygon", "access": "private"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberNode, Ref: 1}}},
	} {
		if err := rels.Relation(r); err == nil {
			t.Errorf("an area relation carrying an access key with %s was taken", name)
		}
	}

	for name, r := range map[string]osmpbf.Relation{
		"a multipolygon carrying no access key": {ID: 8, Tags: map[string]string{"type": "multipolygon", "access:conditional": "private @ (Mo-Fr 08:00-18:00)"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberRelation, Ref: 70}}},
		"a route relation carrying one": {ID: 9, Tags: map[string]string{"type": "route", "access": "private"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberRelation, Ref: 70}}},
		"a site relation carrying one": {ID: 10, Tags: map[string]string{"type": "site", "access": "customers"},
			Members: []osmpbf.Member{{Type: osmpbf.MemberRelation, Ref: 70}}},
	} {
		if err := rels.Relation(r); err != nil {
			t.Errorf("%s was refused, want it passed over: %v", name, err)
		}
	}

	if len(rels.areas) != 0 {
		t.Errorf("the relations taken are %+v, want none", rels.areas)
	}
}

// TestWhichRelationsAreAreas holds the reader to the area relations
// `Architecture.md § What the tables hold` names (S1 on PR #296): a
// type=multipolygon and a type=boundary relation carrying a key are each taken,
// with every way it names, and a node member, a boundary's label, passed over.
func TestWhichRelationsAreAreas(t *testing.T) {
	t.Parallel()

	members := []osmpbf.Member{
		{Type: osmpbf.MemberWay, Ref: 10, Role: "outer"},
		{Type: osmpbf.MemberNode, Ref: 1, Role: "label"},
		{Type: osmpbf.MemberWay, Ref: 11, Role: "outer"},
	}

	for _, kind := range []string{"multipolygon", "boundary"} {
		rels := &relationsPass{members: map[int64]bool{}}

		r := osmpbf.Relation{ID: 7, Tags: map[string]string{"type": kind, "boundary": "protected_area", "access": "no"}, Members: members}
		if err := rels.Relation(r); err != nil {
			t.Fatalf("a type=%s relation carrying access=no: %v", kind, err)
		}

		if len(rels.areas) != 1 || rels.areas[0].id != 7 || len(rels.areas[0].ways) != 2 || rels.areas[0].ways[0] != 10 || rels.areas[0].ways[1] != 11 {
			t.Errorf("a type=%s relation carrying access=no is taken as %+v, want relation 7 drawn from ways 10 and 11", kind, rels.areas)

			continue
		}

		if got, _ := value(rels.areas[0].values, "access"); got != "no" {
			t.Errorf("a type=%s relation's access is read as %q, want no", kind, got)
		}

		if !rels.members[10] || !rels.members[11] {
			t.Errorf("a type=%s relation's member ways are not kept for the ways pass: %v", kind, rels.members)
		}
	}
}
