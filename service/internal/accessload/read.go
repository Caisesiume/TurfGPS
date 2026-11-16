// Package accessload fills the access-marking tables of `Architecture.md § The
// access-marking tables` from the `.osm.pbf` file the routing tiles are built
// from (#255, `§ Loading from the extract the tiles are built from`).
//
// IT COMPUTES NOTHING ABOUT ACCESS. It reads the twelve access keys verbatim
// and stores them; whether a feature is marked is the tables' generated
// column, so the rule has one home, the migration, and a reviewer reads it
// there.
//
// A FEATURE LOST ON LOAD FAILS THE LOAD. The reader counts every feature it
// takes as it reads it, and the writer records that count beside the rows each
// table received; the load record's CHECK refuses a row where the two differ,
// and a way whose geometry cannot be built from the file fails the read before
// anything is written. A failed load writes nothing: the previous load stays
// current, or none is.
//
// IT IS NEVER ON A REQUEST PATH. `cmd/accessload` runs it; nothing the request
// surface imports reaches it.
//
// THE AREA FEATURES (`access_area`, #253, `FR-128`) are the set `Architecture.md
// § What the tables hold` names: every closed way without a `highway` key, every
// closed way tagged `area=yes`, every closed way whose `highway` value is one
// OpenStreetMap documents on areas and never on ways (areaOnlyHighway), and
// every relation of a type OpenStreetMap draws as an area (areaRelationTypes),
// carrying any of the twelve keys. The reader takes each one's linework — the
// closed way, or every way the relation names as a member, whatever its role —
// and PostGIS draws the area from it as the rows enter the table (write.go). An area this reader cannot take whole fails
// the read: a member way or node the file does not hold, a relation naming a
// relation, linework whose ends do not close into rings, or a closed way with
// too few nodes to enclose anything. Node members are not linework and are
// passed over: they enclose nothing under any reading.
package accessload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/osmpbf"
)

// Keys are the twelve access keys of `§ What the tables hold`, in the order of
// the tables' columns. A `:conditional` form of any of them is not among them:
// the Owner's third tag decision on #258 is that the loader reads none.
var Keys = [12]string{
	"motorcar", "motor_vehicle", "vehicle", "access",
	"motorcar:forward", "motor_vehicle:forward", "vehicle:forward", "access:forward",
	"motorcar:backward", "motor_vehicle:backward", "vehicle:backward", "access:backward",
}

// Values are a feature's values for the twelve keys, nil where it does not
// carry one.
type Values [12]*string

// valuesOf returns tags' values for the twelve keys, and whether it carries
// any.
func valuesOf(tags map[string]string) (Values, bool) {
	var (
		v        Values
		carrying bool
	)

	for i, k := range Keys {
		if val, ok := tags[k]; ok {
			v[i] = &val
			carrying = true
		}
	}

	return v, carrying
}

// Way is a row of access_way.
type Way struct {
	ID     int64
	Values Values
	Nodes  []int64

	// Lons and Lats are the way's nodes' coordinates, in node order.
	Lons, Lats []float64
}

// Barrier is a row of access_barrier.
type Barrier struct {
	ID      int64
	Barrier *string
	Values  Values
	Lon     float64
	Lat     float64
}

// Pair is a way and one of its nodes.
type Pair struct{ Way, Node int64 }

// Area is a row of access_area, before PostGIS draws it: the feature, its
// values, and its linework.
type Area struct {
	// Type is 'w' for a closed way and 'r' for a relation, as
	// access_area.osm_type spells them.
	Type byte
	ID   int64

	Values Values

	// Lines is the area's linework: the closed way itself, or each way the
	// relation names, in member order.
	Lines []Line
}

// Line is one way of an area's linework, its coordinates in node order.
type Line struct {
	Way        int64
	Lons, Lats []float64
}

// Extract is everything the loader takes from one file.
type Extract struct {
	// SHA256 is the file's SHA-256, in lower-case hex.
	SHA256 string

	// Timestamp is the file's own timestamp, where its header carries one.
	Timestamp time.Time

	Ways     []Way
	Barriers []Barrier

	// BarrierWays pairs each barrier with every way referencing it.
	BarrierWays []Pair

	// Junctions pairs each `highway` way with each of its nodes that a way in
	// Ways also has, so the writer can find where an unmarked way meets a
	// marked one without computing the marking here.
	Junctions []Pair

	// Areas are the area features carrying any of the keys.
	Areas []Area

	// WayFeatures, BarrierFeatures and AreaFeatures are the counts of
	// features in the file carrying any of the keys, counted as they were
	// read.
	WayFeatures, BarrierFeatures, AreaFeatures int64
}

// ReadFile reads the extract at path. The file is read three times — the
// area relations decide which member ways' nodes are kept, and the
// ways the tables take decide which node coordinates are — and every read
// must hash to the same SHA-256, so a file replaced between them fails the
// read.
func ReadFile(path string) (*Extract, error) {
	rels := &relationsPass{members: map[int64]bool{}}

	sha, err := readHashed(path, rels)
	if err != nil {
		return nil, fmt.Errorf("reading %s for its relations: %w", path, err)
	}

	first := &waysPass{nodesOf: map[int64]bool{}, members: rels.members, memberNodes: map[int64][]int64{}}

	if err := readAgain(path, sha, "ways", first); err != nil {
		return nil, err
	}

	second := &nodesPass{
		ways:       first.ways,
		wayNodes:   first.nodesOf,
		coords:     map[int64][2]float64{},
		barrierIDs: map[int64]bool{},
	}

	if err := readAgain(path, sha, "nodes", second); err != nil {
		return nil, err
	}

	ex := &Extract{
		SHA256:          sha,
		Timestamp:       second.timestamp,
		Barriers:        second.barriers,
		BarrierWays:     second.barrierWays,
		Junctions:       second.junctions,
		WayFeatures:     int64(len(first.ways)),
		BarrierFeatures: second.barrierFeatures,
		AreaFeatures:    int64(len(first.areaWays) + len(rels.areas)),
	}

	for _, w := range first.ways {
		lons, lats, err := second.line(w.Nodes)
		if err != nil {
			return nil, fmt.Errorf("way %d carries an access key and %w, so its geometry cannot be built", w.ID, err)
		}

		if len(w.Nodes) < 2 {
			return nil, fmt.Errorf("way %d carries an access key and has %d node(s), so its geometry cannot be built", w.ID, len(w.Nodes))
		}

		w.Lons, w.Lats = lons, lats
		ex.Ways = append(ex.Ways, w)
	}

	for _, a := range first.areaWays {
		area, err := second.wayArea(a)
		if err != nil {
			return nil, fmt.Errorf("way %d is an area carrying an access key and %w, so its area cannot be drawn", a.id, err)
		}

		ex.Areas = append(ex.Areas, area)
	}

	for _, r := range rels.areas {
		area, err := second.relationArea(r, first.memberNodes)
		if err != nil {
			return nil, fmt.Errorf("relation %d is a type=%s relation carrying an access key and %w, so its area cannot be drawn", r.id, r.kind, err)
		}

		ex.Areas = append(ex.Areas, area)
	}

	return ex, nil
}

// readAgain reads the file at path into h and fails unless it hashes to sha,
// the SHA-256 of the first read.
func readAgain(path, sha, what string, h osmpbf.Handler) error {
	again, err := readHashed(path, h)
	if err != nil {
		return fmt.Errorf("reading %s for its %s: %w", path, what, err)
	}

	if again != sha {
		return fmt.Errorf("%s changed while it was read: its SHA-256 was %s and then %s", path, sha, again)
	}

	return nil
}

// readHashed reads the file at path into h and returns its SHA-256.
func readHashed(path string, h osmpbf.Handler) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	sum := sha256.New()

	if err := osmpbf.Read(io.TeeReader(f, sum), h); err != nil {
		return "", err
	}

	// The reader stops at the last block; anything after it is part of the
	// file and of its identity.
	if _, err := io.Copy(sum, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(sum.Sum(nil)), nil
}

// order holds a pass to the file order the format's Sort.Type_then_ID gives —
// every node before every way, every way before every relation — which both
// passes rely on: a barrier is known before the ways referencing it are read.
// A file in another order is refused rather than read partly.
type order struct{ seen int }

func (o *order) at(kind int, what string) error {
	if kind < o.seen {
		return fmt.Errorf("a %s after a later kind of object: the file is not sorted nodes, then ways, then relations", what)
	}

	o.seen = kind

	return nil
}

// areaRelationTypes are the relation types OpenStreetMap defines as areas: a
// multipolygon (wiki `Relation:multipolygon`) and a boundary, which the wiki's
// `Relation:boundary` says is "defined in a similar manner as multipolygons:
// they must contain at least one outer way". A relation of another type — a
// route, a site, a restriction — is not one, and is passed over. Narrowing
// this set is a change to a safety rule (`Architecture.md § What the tables
// hold`).
var areaRelationTypes = map[string]bool{"multipolygon": true, "boundary": true}

// relationsPass collects the area relations carrying any of the keys, and the
// ways they name.
type relationsPass struct {
	order
	areas   []areaRelation
	members map[int64]bool
}

// areaRelation is an area relation carrying any of the keys: its type, its
// values and the ways it names, in member order.
type areaRelation struct {
	id     int64
	kind   string
	values Values
	ways   []int64
}

func (p *relationsPass) Header(osmpbf.Header) error { return nil }

func (p *relationsPass) Node(osmpbf.Node) error { return p.at(0, "node") }

func (p *relationsPass) Way(osmpbf.Way) error { return p.at(1, "way") }

func (p *relationsPass) Relation(r osmpbf.Relation) error {
	if err := p.at(2, "relation"); err != nil {
		return err
	}

	kind := r.Tags["type"]
	if !areaRelationTypes[kind] {
		return nil
	}

	v, carrying := valuesOf(r.Tags)
	if !carrying {
		return nil
	}

	a := areaRelation{id: r.ID, kind: kind, values: v}

	for _, m := range r.Members {
		switch m.Type {
		case osmpbf.MemberWay:
			a.ways = append(a.ways, m.Ref)
			p.members[m.Ref] = true
		case osmpbf.MemberRelation:
			return fmt.Errorf("relation %d is a type=%s relation carrying an access key and names relation %d as a member, whose linework this reader does not take, so its area cannot be drawn whole", r.ID, kind, m.Ref)
		case osmpbf.MemberNode:
			// A node encloses nothing: it adds no area under any reading.
		}
	}

	if len(a.ways) == 0 {
		return fmt.Errorf("relation %d is a type=%s relation carrying an access key and names no way, so its area cannot be drawn", r.ID, kind)
	}

	p.areas = append(p.areas, a)

	return nil
}

// waysPass collects the ways carrying any of the keys, the closed ways that
// are areas carrying them, the nodes of the area relations' member ways, and
// the nodes of all three.
type waysPass struct {
	order
	ways    []Way
	nodesOf map[int64]bool

	areaWays    []areaWay
	members     map[int64]bool
	memberNodes map[int64][]int64
}

// areaWay is a closed way that is an area carrying any of the keys.
type areaWay struct {
	id     int64
	values Values
	nodes  []int64
}

// areaOnlyHighway are the `highway` values OpenStreetMap documents on areas
// and never on ways: each value's wiki page (`Tag:highway=<value>`) states
// `onWay = no` and `onArea = yes`. Read on 16 November 2026 over every `highway`
// value taginfo lists among the 300 most used with a wiki page; PR #296 records
// the query. A closed way carrying one is an area whether or not it is tagged
// `area=yes`, as `highway=rest_area` and `highway=services` are found without it.
// Narrowing this set is a change to a safety rule (`Architecture.md § What the
// tables hold`).
var areaOnlyHighway = map[string]bool{"rest_area": true, "services": true, "trailhead": true, "turntable": true}

// isArea reports whether a way is one of the area features access_area takes
// (`Architecture.md § What the tables hold`): closed, and either without a
// `highway` key, tagged `area=yes`, or carrying a `highway` value that is
// only ever an area.
func isArea(w osmpbf.Way) bool {
	if len(w.Nodes) < 2 || w.Nodes[0] != w.Nodes[len(w.Nodes)-1] {
		return false
	}

	highway, ok := w.Tags["highway"]

	return !ok || w.Tags["area"] == "yes" || areaOnlyHighway[highway]
}

func (p *waysPass) Header(osmpbf.Header) error { return nil }

func (p *waysPass) Node(osmpbf.Node) error { return p.at(0, "node") }

func (p *waysPass) Way(w osmpbf.Way) error {
	if err := p.at(1, "way"); err != nil {
		return err
	}

	if p.members[w.ID] {
		p.memberNodes[w.ID] = slices.Clone(w.Nodes)
		p.keep(w.Nodes)
	}

	v, carrying := valuesOf(w.Tags)
	if !carrying {
		return nil
	}

	p.keep(w.Nodes)

	p.ways = append(p.ways, Way{ID: w.ID, Values: v, Nodes: slices.Clone(w.Nodes)})

	if isArea(w) {
		p.areaWays = append(p.areaWays, areaWay{id: w.ID, values: v, nodes: slices.Clone(w.Nodes)})
	}

	return nil
}

func (p *waysPass) keep(nodes []int64) {
	for _, n := range nodes {
		p.nodesOf[n] = true
	}
}

func (p *waysPass) Relation(osmpbf.Relation) error { return p.at(2, "relation") }

// nodesPass collects the barriers, the coordinates of the first pass's ways'
// nodes, the barriers' referencing ways and the junctions.
type nodesPass struct {
	order
	ways      []Way
	wayNodes  map[int64]bool
	timestamp time.Time

	coords          map[int64][2]float64
	barriers        []Barrier
	barrierIDs      map[int64]bool
	barrierFeatures int64
	barrierWays     []Pair
	junctions       []Pair
}

func (p *nodesPass) Header(h osmpbf.Header) error {
	p.timestamp = h.ReplicationTimestamp

	return nil
}

func (p *nodesPass) Node(n osmpbf.Node) error {
	if err := p.at(0, "node"); err != nil {
		return err
	}

	if p.wayNodes[n.ID] {
		p.coords[n.ID] = [2]float64{n.Longitude, n.Latitude}
	}

	v, carrying := valuesOf(n.Tags)
	if !carrying {
		return nil
	}

	p.barrierFeatures++
	p.barrierIDs[n.ID] = true

	b := Barrier{ID: n.ID, Values: v, Lon: n.Longitude, Lat: n.Latitude}
	if kind, ok := n.Tags["barrier"]; ok {
		b.Barrier = &kind
	}

	p.barriers = append(p.barriers, b)

	return nil
}

func (p *nodesPass) Way(w osmpbf.Way) error {
	if err := p.at(1, "way"); err != nil {
		return err
	}

	_, highway := w.Tags["highway"]

	seenBarrier, seenJunction := map[int64]bool{}, map[int64]bool{}

	for _, n := range w.Nodes {
		if p.barrierIDs[n] && !seenBarrier[n] {
			seenBarrier[n] = true
			p.barrierWays = append(p.barrierWays, Pair{Way: w.ID, Node: n})
		}

		if highway && p.wayNodes[n] && !seenJunction[n] {
			seenJunction[n] = true
			p.junctions = append(p.junctions, Pair{Way: w.ID, Node: n})
		}
	}

	return nil
}

func (p *nodesPass) Relation(osmpbf.Relation) error { return p.at(2, "relation") }

// line returns the coordinates of the nodes, in their order.
func (p *nodesPass) line(nodes []int64) (lons, lats []float64, err error) {
	lons, lats = make([]float64, len(nodes)), make([]float64, len(nodes))

	for i, n := range nodes {
		c, ok := p.coords[n]
		if !ok {
			return nil, nil, fmt.Errorf("references node %d, which the file does not hold", n)
		}

		lons[i], lats[i] = c[0], c[1]
	}

	return lons, lats, nil
}

// wayArea is a closed way's area, refused where its nodes are not all in the
// file or are too few to enclose anything: a ring of fewer than three distinct
// nodes encloses nothing.
func (p *nodesPass) wayArea(a areaWay) (Area, error) {
	if len(a.nodes) < 4 {
		return Area{}, fmt.Errorf("is closed with %d node(s), too few to enclose an area", len(a.nodes))
	}

	lons, lats, err := p.line(a.nodes)
	if err != nil {
		return Area{}, err
	}

	return Area{Type: 'w', ID: a.id, Values: a.values, Lines: []Line{{Way: a.id, Lons: lons, Lats: lats}}}, nil
}

// relationArea gathers an area relation's linework, and refuses it unless every
// member way is in the file with all its nodes and the linework closes: each
// node at which a member way ends must be an end of an even number of them,
// counting both ends of a way, so every way lies on a closed ring and PostGIS
// can draw an area from all of it, none left dangling. Roles are not read:
// every member way's linework is taken, an inner ring's included (write.go).
func (p *nodesPass) relationArea(r areaRelation, memberNodes map[int64][]int64) (Area, error) {
	area := Area{Type: 'r', ID: r.id, Values: r.values}
	ends := map[int64]int{}

	for _, id := range r.ways {
		nodes, ok := memberNodes[id]
		if !ok {
			return Area{}, fmt.Errorf("names way %d, which the file does not hold", id)
		}

		if len(nodes) < 2 {
			return Area{}, fmt.Errorf("names way %d, which has %d node(s)", id, len(nodes))
		}

		lons, lats, err := p.line(nodes)
		if err != nil {
			return Area{}, fmt.Errorf("names way %d, which %w", id, err)
		}

		ends[nodes[0]]++
		ends[nodes[len(nodes)-1]]++

		area.Lines = append(area.Lines, Line{Way: id, Lons: lons, Lats: lats})
	}

	for _, n := range sortedIDs(ends) {
		if ends[n]%2 != 0 {
			return Area{}, fmt.Errorf("its member ways do not close into rings: %d of their ends lie at node %d", ends[n], n)
		}
	}

	return area, nil
}

// sortedIDs is m's keys in order, so an error names the same node each run.
func sortedIDs(m map[int64]int) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	slices.Sort(out)

	return out
}
