package accessload

import (
	"encoding/binary"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// #301's cases are constructed extracts, written by each test into its own
// directory and read by ReadFile, the loader's first act, as `turfgps
// load-access-markings` reads the file it is given. They are never merged into
// the committed extract: a case that fails the read would fail the load CI
// runs on it (`internal/valhalla/testdata/README.md`). They follow the rules of
// the synthetic cases there: every id is 9×10¹⁸ or above, in a range named for
// the story — nodes 9×10¹⁸ + 301,000 + 100c + k, ways 9×10¹⁸ + 30,100 + 20c + k,
// relations 9×10¹⁸ + 3,010 + c — and every way carries the note below. A
// relation carries one only where its case allows: AC1's carries `type` alone.

const note301 = "synthetic, not from OpenStreetMap: TurfGPS #301 test case"

func node301(c, k int) int64 { return synthetic + 301_000 + int64(c)*100 + int64(k) }
func way301(c, k int) int64  { return synthetic + 30_100 + int64(c)*20 + int64(k) }
func rel301(c int) int64     { return synthetic + 3_010 + int64(c) }

// A protocol-buffer encoder for constructed extracts, the reader's inverse, as
// `internal/osmpbf`'s own tests build theirs.

func pbVarint(v uint64) []byte {
	var b []byte
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}

	return append(b, byte(v))
}

func pbZig(v int64) uint64 { return uint64((v << 1) ^ (v >> 63)) }

func pbUint(num int, v uint64) []byte { return append(pbVarint(uint64(num<<3)), pbVarint(v)...) }

func pbBytes(num int, b []byte) []byte {
	return append(append(pbVarint(uint64(num<<3|2)), pbVarint(uint64(len(b)))...), b...)
}

func pbPacked(num int, vs ...uint64) []byte {
	var b []byte
	for _, v := range vs {
		b = append(b, pbVarint(v)...)
	}

	return pbBytes(num, b)
}

func pbFrame(typ string, data []byte) []byte {
	blob := pbBytes(1, data)
	header := append(pbBytes(1, []byte(typ)), pbUint(3, uint64(len(blob)))...)

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(header)))

	return append(append(size[:], header...), blob...)
}

type osmNode struct {
	id       int64
	lat, lon float64
}

type osmWay struct {
	id    int64
	nodes []int64
	tags  map[string]string
}

type osmRelation struct {
	id        int64
	tags      map[string]string
	ways      []int64
	relations []int64
}

// constructed is an extract built in a test: its nodes, ways and relations,
// written in that order, as a sorted file holds them.
type constructed struct {
	nodes []osmNode
	ways  []osmWay
	rels  []osmRelation
}

// write writes the extract as an .osm.pbf file of one header block and one
// data block, and returns its path.
func (e constructed) write(t *testing.T) string {
	t.Helper()

	strs := []string{""}
	index := map[string]uint64{"": 0}

	str := func(s string) uint64 {
		if i, ok := index[s]; ok {
			return i
		}

		index[s] = uint64(len(strs))
		strs = append(strs, s)

		return index[s]
	}

	tagFields := func(tags map[string]string) []byte {
		var keys, vals []uint64

		for _, k := range slices.Sorted(maps.Keys(tags)) {
			keys, vals = append(keys, str(k)), append(vals, str(tags[k]))
		}

		return append(pbPacked(2, keys...), pbPacked(3, vals...)...)
	}

	var nodes, ways, rels []byte

	for _, n := range e.nodes {
		msg := append(pbUint(1, pbZig(n.id)), pbUint(8, pbZig(int64(math.Round(n.lat*1e7))))...)
		nodes = append(nodes, pbBytes(1, append(msg, pbUint(9, pbZig(int64(math.Round(n.lon*1e7))))...))...)
	}

	for _, w := range e.ways {
		var refs []uint64

		var prev int64
		for _, n := range w.nodes {
			refs, prev = append(refs, pbZig(n-prev)), n
		}

		msg := append(append(pbUint(1, uint64(w.id)), tagFields(w.tags)...), pbPacked(8, refs...)...)
		ways = append(ways, pbBytes(3, msg)...)
	}

	for _, r := range e.rels {
		var roles, refs, types []uint64

		var prev int64
		for _, w := range r.ways {
			roles, refs, types, prev = append(roles, str("outer")), append(refs, pbZig(w-prev)), append(types, 1), w
		}

		for _, m := range r.relations {
			roles, refs, types, prev = append(roles, str("")), append(refs, pbZig(m-prev)), append(types, 2), m
		}

		msg := append(pbUint(1, uint64(r.id)), tagFields(r.tags)...)
		msg = append(append(append(msg, pbPacked(8, roles...)...), pbPacked(9, refs...)...), pbPacked(10, types...)...)
		rels = append(rels, pbBytes(4, msg)...)
	}

	var table []byte
	for _, s := range strs {
		table = append(table, pbBytes(1, []byte(s))...)
	}

	block := pbBytes(1, table)
	for _, g := range [][]byte{nodes, ways, rels} {
		if len(g) != 0 {
			block = append(block, pbBytes(2, g)...)
		}
	}

	file := append(pbFrame("OSMHeader", pbBytes(4, []byte("OsmSchema-V0.6"))), pbFrame("OSMData", block)...)

	path := filepath.Join(t.TempDir(), "constructed.osm.pbf")
	if err := os.WriteFile(path, file, 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// ringCase is case c: four nodes at the corners of a square, two open ways
// that together close it — way 1 through nodes 1, 2, 3 and way 2 through 3, 4,
// 1 — and a closed way 3 round all four, and the relation, naming the ways
// members names. Each way's tags are the note and those given.
func ringCase(c int, rel map[string]string, members []int, wayTags map[int]map[string]string) constructed {
	lat := 57.100 + 0.002*float64(c)

	e := constructed{nodes: []osmNode{
		{node301(c, 1), lat, 11.700}, {node301(c, 2), lat, 11.701},
		{node301(c, 3), lat + 0.001, 11.701}, {node301(c, 4), lat + 0.001, 11.700},
	}}

	for k, nodes := range map[int][]int{1: {1, 2, 3}, 2: {3, 4, 1}, 3: {1, 2, 3, 4, 1}} {
		tags := map[string]string{"note": note301}
		for key, v := range wayTags[k] {
			tags[key] = v
		}

		w := osmWay{id: way301(c, k), tags: tags}
		for _, n := range nodes {
			w.nodes = append(w.nodes, node301(c, n))
		}

		e.ways = append(e.ways, w)
	}

	slices.SortFunc(e.ways, func(a, b osmWay) int { return int(a.id - b.id) })

	if rel != nil {
		r := osmRelation{id: rel301(c), tags: rel}
		for _, k := range members {
			r.ways = append(r.ways, way301(c, k))
		}

		e.rels = []osmRelation{r}
	}

	return e
}

// TestAnOldStyleAreaRelationFailsTheRead holds #301's AC1 and AC2 (`RD-087`):
// an area relation carrying no key that defines a feature, one of whose member
// ways carries an access key and is no area itself, fails the read with a
// message naming both; the N1 cases are the sentinel's on PR #306. ReadFile is the
// loader's first act and returns no extract, so nothing is staged or swapped,
// and the previous load stays current, or none is: `turfgps
// load-access-markings` reads the file before it opens the store. Each case's
// extract read without its relation is read whole, with the way taken as a
// way, so the failure is the relation's and nothing else in the file's.
func TestAnOldStyleAreaRelationFailsTheRead(t *testing.T) {
	t.Parallel()

	private := map[string]string{"access": "private"}

	for _, tc := range []struct {
		name string
		c    int
		rel  map[string]string
		way1 map[string]string
	}{
		{"AC1: a multipolygon carrying only type, its marked member no highway way",
			0, map[string]string{"type": "multipolygon"}, private},
		{"AC2: a boundary carrying only type",
			1, map[string]string{"type": "boundary"}, private},
		{"AC2: a multipolygon also carrying name, note, source:geometry and fixme",
			2, map[string]string{"type": "multipolygon", "name": "Kundparkering", "note": "outer ring tagged", "source:geometry": "survey", "fixme": "move tags to the relation"}, private},
		{"AC2: a multipolygon whose marked member is an open highway=service way",
			3, map[string]string{"type": "multipolygon"}, map[string]string{"highway": "service", "access": "private"}},
		// The sentinel's N1 on PR #306: a key that defines no feature hides
		// no marking.
		{"N1: a multipolygon also carrying area=yes",
			7, map[string]string{"type": "multipolygon", "area": "yes"}, private},
		{"N1: a multipolygon also carrying layer=-1",
			8, map[string]string{"type": "multipolygon", "layer": "-1"}, private},
		{"N1: a multipolygon also carrying alt_name",
			9, map[string]string{"type": "multipolygon", "alt_name": "Kundparkering"}, private},
		{"N1: a multipolygon also carrying old_name",
			10, map[string]string{"type": "multipolygon", "old_name": "Kundparkering"}, private},
		{"N1: a multipolygon also carrying wikidata",
			11, map[string]string{"type": "multipolygon", "wikidata": "Q1"}, private},
	} {
		e := ringCase(tc.c, tc.rel, []int{1, 2}, map[int]map[string]string{1: tc.way1})

		alone := e
		alone.rels = nil

		premise, err := ReadFile(alone.write(t))
		if err != nil {
			t.Fatalf("%s: the premise does not hold: the extract without its relation is refused: %v", tc.name, err)
		}

		if !slices.ContainsFunc(premise.Ways, func(w Way) bool { return w.ID == way301(tc.c, 1) }) {
			t.Fatalf("%s: the premise does not hold: way %d is not read as a way carrying an access key", tc.name, way301(tc.c, 1))
		}

		ex, err := ReadFile(e.write(t))
		if err == nil {
			t.Errorf("%s: the extract is read, with %d areas and %d ways, want it refused", tc.name, len(ex.Areas), len(ex.Ways))

			continue
		}

		t.Logf("%s: %v", tc.name, err)

		if ex != nil {
			t.Errorf("%s: a refused read returned an extract, which a load could write", tc.name)
		}

		for _, want := range []string{fmt.Sprintf("relation %d ", rel301(tc.c)), fmt.Sprintf("member way %d ", way301(tc.c, 1))} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the read fails with %q, which does not name %q", tc.name, err, strings.TrimSpace(want))
			}
		}
	}
}

// TestAnAreaRelationThatIsNotOldStyleIsRead holds #301's AC3: a relation
// describing a feature of its own is not old-style, so its marked open member
// is read as a way and the relation, carrying no access key, draws no area;
// and a relation whose only marked member is a closed way without a `highway`
// key reads, the closed way taken as an area of its own.
func TestAnAreaRelationThatIsNotOldStyleIsRead(t *testing.T) {
	t.Parallel()

	t.Run("a multipolygon carrying landuse=residential", func(t *testing.T) {
		t.Parallel()

		e := ringCase(4, map[string]string{"type": "multipolygon", "landuse": "residential"}, []int{1, 2},
			map[int]map[string]string{1: {"highway": "service", "access": "private"}})

		ex, err := ReadFile(e.write(t))
		if err != nil {
			t.Fatalf("the extract is refused: %v", err)
		}

		if !slices.ContainsFunc(ex.Ways, func(w Way) bool { return w.ID == way301(4, 1) }) {
			t.Errorf("way %d is not read among the ways carrying an access key, want it in access_way", way301(4, 1))
		}

		if slices.ContainsFunc(ex.Areas, func(a Area) bool { return a.Type == 'r' }) || ex.AreaFeatures != int64(len(ex.Areas)) {
			t.Errorf("the areas read are %+v from %d features, want no relation among them", ex.Areas, ex.AreaFeatures)
		}
	})

	t.Run("a multipolygon carrying area:highway, a top-level key of a dependent element", func(t *testing.T) {
		t.Parallel()

		e := ringCase(14, map[string]string{"type": "multipolygon", "area:highway": "pedestrian"}, []int{1, 2},
			map[int]map[string]string{1: {"highway": "residential", "vehicle": "no"}})

		ex, err := ReadFile(e.write(t))
		if err != nil {
			t.Fatalf("the extract is refused: %v", err)
		}

		if !slices.ContainsFunc(ex.Ways, func(w Way) bool { return w.ID == way301(14, 1) }) {
			t.Errorf("way %d is not read among the ways carrying an access key, want it in access_way", way301(14, 1))
		}
	})

	t.Run("a multipolygon carrying only type, its marked member a closed way", func(t *testing.T) {
		t.Parallel()

		e := ringCase(5, map[string]string{"type": "multipolygon"}, []int{3},
			map[int]map[string]string{3: {"access": "private"}})

		ex, err := ReadFile(e.write(t))
		if err != nil {
			t.Fatalf("the extract is refused: %v", err)
		}

		if !slices.ContainsFunc(ex.Areas, func(a Area) bool { return a.Type == 'w' && a.ID == way301(5, 3) }) {
			t.Errorf("the closed way %d is not read as an area, want it in access_area; the areas are %+v", way301(5, 3), ex.Areas)
		}

		if slices.ContainsFunc(ex.Areas, func(a Area) bool { return a.Type == 'r' }) || ex.AreaFeatures != int64(len(ex.Areas)) {
			t.Errorf("the areas read are %+v from %d features, want no relation among them", ex.Areas, ex.AreaFeatures)
		}
	})
}

// TestAnOldStyleAreaRelationMissingAMemberFailsTheRead holds the lead's
// ruling on PR #306: an old-style area relation naming a way the file does not
// hold, as one clipped at the extract's edge may, fails the read, naming the
// relation and the way. The missing way cannot be checked for a marking and may
// be the one carrying it, so the read fails closed, as it does on a missing
// member of a relation carrying a key. The same extract holding the way, which
// carries no key, is read whole, so the failure is the missing way's.
func TestAnOldStyleAreaRelationMissingAMemberFailsTheRead(t *testing.T) {
	t.Parallel()

	whole := ringCase(6, map[string]string{"type": "multipolygon"}, []int{1, 2}, nil)

	if _, err := ReadFile(whole.write(t)); err != nil {
		t.Fatalf("the premise does not hold: the extract holding every member way is refused: %v", err)
	}

	clipped := whole
	clipped.ways = slices.DeleteFunc(slices.Clone(whole.ways), func(w osmWay) bool { return w.id == way301(6, 2) })

	ex, err := ReadFile(clipped.write(t))
	if err == nil {
		t.Fatalf("the extract missing member way %d of old-style relation %d is read, with %d areas and %d ways, want it refused",
			way301(6, 2), rel301(6), len(ex.Areas), len(ex.Ways))
	}

	t.Logf("%v", err)

	if ex != nil {
		t.Errorf("a refused read returned an extract, which a load could write")
	}

	for _, want := range []string{fmt.Sprintf("relation %d ", rel301(6)), fmt.Sprintf("names way %d, which the file does not hold", way301(6, 2))} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the read fails with %q, which does not name %q", err, strings.TrimSpace(want))
		}
	}

	if strings.Contains(err.Error(), fmt.Sprintf("way %d,", way301(6, 1))) {
		t.Errorf("the read fails with %q, which names member way %d, which the file holds", err, way301(6, 1))
	}
}

// TestAnOldStyleAreaRelationNamingARelationFailsTheRead holds the sentinel's
// N2 on PR #306 and the lead's ruling on it: an old-style area relation naming
// a relation as a member fails the read, naming both, as a relation carrying a
// key does. The guard checks the ways an old-style relation names, not a
// member relation's, and a marking on those may be the area's. Case 12's
// relation names relation 9×10¹⁸ + 3,013, a `type=collection` naming case 12's
// ways, so the marking sits on a way of the member relation's; no case 13 is
// built. The extract without case 12's relation, and with it describing a
// feature of its own, are each read, so the failure is the old-style rule's.
func TestAnOldStyleAreaRelationNamingARelationFailsTheRead(t *testing.T) {
	t.Parallel()

	sub := osmRelation{id: rel301(13), tags: map[string]string{"type": "collection"}, ways: []int64{way301(12, 1), way301(12, 2)}}

	with := func(tags map[string]string) constructed {
		e := ringCase(12, nil, nil, map[int]map[string]string{1: {"access": "private"}})
		if tags != nil {
			e.rels = append(e.rels, osmRelation{id: rel301(12), tags: tags, relations: []int64{rel301(13)}})
		}

		e.rels = append(e.rels, sub)

		return e
	}

	alone, err := ReadFile(with(nil).write(t))
	if err != nil {
		t.Fatalf("the premise does not hold: the extract without relation %d is refused: %v", rel301(12), err)
	}

	if !slices.ContainsFunc(alone.Ways, func(w Way) bool { return w.ID == way301(12, 1) }) {
		t.Fatalf("the premise does not hold: way %d is not read as a way carrying an access key", way301(12, 1))
	}

	if _, err := ReadFile(with(map[string]string{"type": "multipolygon", "landuse": "residential"}).write(t)); err != nil {
		t.Fatalf("the premise does not hold: the extract with relation %d describing a feature is refused: %v", rel301(12), err)
	}

	ex, err := ReadFile(with(map[string]string{"type": "multipolygon"}).write(t))
	if err == nil {
		t.Fatalf("the extract whose old-style relation %d names relation %d is read, with %d areas and %d ways, want it refused",
			rel301(12), rel301(13), len(ex.Areas), len(ex.Ways))
	}

	t.Logf("%v", err)

	if ex != nil {
		t.Errorf("a refused read returned an extract, which a load could write")
	}

	for _, want := range []string{fmt.Sprintf("relation %d ", rel301(12)), fmt.Sprintf("names relation %d as a member", rel301(13))} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the read fails with %q, which does not name %q", err, strings.TrimSpace(want))
		}
	}
}

// TestWhichRelationsAreOldStyle holds the test the lead's ruling on PR #306
// sets: an area relation is old-style unless it carries a key of featureKeys,
// the top-level keys of the wiki's `Top-level tag`. Every other key, names,
// housekeeping, `area`, `layer`, `wikidata` and a lifecycle prefix among them,
// leaves it old-style. A relation carrying an access key never reaches the
// test: it is read as an area of its own.
func TestWhichRelationsAreOldStyle(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		tags map[string]string
		want bool
	}{
		{map[string]string{"type": "multipolygon"}, true},
		{map[string]string{"type": "multipolygon", "name": "x", "name:en": "x", "created_by": "JOSM", "source": "survey", "source:date": "2020"}, true},
		{map[string]string{"type": "boundary", "note": "x", "note:sv": "x", "fixme": "x", "FIXME": "x"}, true},
		{map[string]string{"type": "multipolygon", "area": "yes", "layer": "-1", "alt_name": "x", "old_name": "x", "wikidata": "Q1"}, true},
		{map[string]string{"type": "multipolygon", "name_1": "x", "notes": "x", "wikipedia": "sv:x", "ref": "1"}, true},
		{map[string]string{"type": "multipolygon", "disused:amenity": "parking", "surface": "asphalt"}, true},
		{map[string]string{"type": "multipolygon", "snowmobile": "no"}, true},
		{map[string]string{"type": "multipolygon", "landuse": "residential"}, false},
		{map[string]string{"type": "multipolygon", "amenity": "parking", "area": "yes"}, false},
		{map[string]string{"type": "boundary", "boundary": "administrative", "admin_level": "8"}, false},
		{map[string]string{"type": "multipolygon", "area:highway": "pedestrian"}, false},
		{map[string]string{"type": "multipolygon", "building:part": "yes"}, false},
		{map[string]string{"type": "multipolygon", "piste:type": "nordic"}, false},
	} {
		if got := oldStyleTags(tc.tags); got != tc.want {
			t.Errorf("oldStyleTags(%v) is %v, want %v", tc.tags, got, tc.want)
		}
	}

	if len(featureKeys) != 50 {
		t.Errorf("featureKeys holds %d keys, want the 50 the wiki's Top-level tag lists at revision 2875567: 33 for standalone elements and 17 for dependent ones", len(featureKeys))
	}
}
