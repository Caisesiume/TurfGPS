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
// Areas (`access_area`) are #253's to take; this loader records that it did
// not take them, with null area counts, and leaves the table empty.
package accessload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
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

	for i, k := range Keys[:4] { // RED RUN: no directional form
		if val, ok := tags[k]; ok {
			v[i] = &val
			carrying = true
		} else if val, ok := tags[k+":conditional"]; ok { // RED RUN: a conditional form read as its key
			val, _, _ = strings.Cut(val, " @")
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

	// WayFeatures and BarrierFeatures are the counts of features in the file
	// carrying any of the keys, counted as they were read.
	WayFeatures, BarrierFeatures int64
}

// ReadFile reads the extract at path. The file is read twice — the ways the
// tables take decide which node coordinates are kept — and both reads must
// hash to the same SHA-256, so a file replaced between them fails the read.
func ReadFile(path string) (*Extract, error) {
	first := &waysPass{nodesOf: map[int64]bool{}}

	sha, err := readHashed(path, first)
	if err != nil {
		return nil, fmt.Errorf("reading %s for its ways: %w", path, err)
	}

	second := &nodesPass{
		ways:       first.ways,
		wayNodes:   first.nodesOf,
		coords:     map[int64][2]float64{},
		barrierIDs: map[int64]bool{},
	}

	again, err := readHashed(path, second)
	if err != nil {
		return nil, fmt.Errorf("reading %s for its nodes: %w", path, err)
	}

	if again != sha {
		return nil, fmt.Errorf("%s changed while it was read: its SHA-256 was %s and then %s", path, sha, again)
	}

	ex := &Extract{
		SHA256:          sha,
		Timestamp:       second.timestamp,
		Barriers:        second.barriers,
		BarrierWays:     second.barrierWays,
		Junctions:       second.junctions,
		WayFeatures:     int64(len(first.ways)),
		BarrierFeatures: second.barrierFeatures,
	}

	for _, w := range first.ways {
		lons, lats := make([]float64, len(w.Nodes)), make([]float64, len(w.Nodes))

		for i, n := range w.Nodes {
			c, ok := second.coords[n]
			if !ok {
				return nil, fmt.Errorf("way %d carries an access key and references node %d, which the file does not hold, so its geometry cannot be built", w.ID, n)
			}

			lons[i], lats[i] = c[0], c[1]
		}

		if len(w.Nodes) < 2 {
			return nil, fmt.Errorf("way %d carries an access key and has %d node(s), so its geometry cannot be built", w.ID, len(w.Nodes))
		}

		w.Lons, w.Lats = lons, lats
		ex.Ways = append(ex.Ways, w)
	}

	return ex, nil
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

// waysPass collects the ways carrying any of the keys, and their nodes.
type waysPass struct {
	order
	ways    []Way
	nodesOf map[int64]bool
}

func (p *waysPass) Header(osmpbf.Header) error { return nil }

func (p *waysPass) Node(osmpbf.Node) error { return p.at(0, "node") }

func (p *waysPass) Way(w osmpbf.Way) error {
	if err := p.at(1, "way"); err != nil {
		return err
	}

	v, carrying := valuesOf(w.Tags)
	if !carrying {
		return nil
	}

	for _, n := range w.Nodes {
		p.nodesOf[n] = true
	}

	p.ways = append(p.ways, Way{ID: w.ID, Values: v, Nodes: slices.Clone(w.Nodes)})

	return nil
}

func (p *waysPass) Relation(int64) error { return p.at(2, "relation") }

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
	if _, barrier := n.Tags["barrier"]; !carrying || !barrier { // RED RUN
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

func (p *nodesPass) Relation(int64) error { return p.at(2, "relation") }
