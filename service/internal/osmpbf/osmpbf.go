// Package osmpbf reads an OpenStreetMap `.osm.pbf` file: its header, and its
// nodes, ways and relations with their tags, exactly as the file holds them.
//
// It exists for the access-marking loader (`Architecture.md § The
// access-marking tables`, § Loading from the extract the tiles are built
// from), which reads the same file `valhalla_build_tiles` reads. That section
// proposes `osm2pgsql` and names a Go reader as the alternative; this is the
// alternative, chosen because the loader's own feature count is what catches a
// feature lost on load, and a reader the loader owns reports every feature it
// read. It depends on nothing outside the standard library.
//
// IT FAILS CLOSED. A file this package cannot read whole is an error, never a
// partial read: a blob compressed by a method other than none or zlib, a
// required feature it does not implement (history files among them), a
// malformed message, or a block of a type it does not know. A partial read
// would load a table missing features, and a feature missing from the table
// reads as unmarked.
//
// A relation is read with its tags and its members, each member's type, id
// and role as the file holds them; assembling a multipolygon from them is the
// loader's (#253), not this package's.
//
// The format is https://wiki.openstreetmap.org/wiki/PBF_Format, and the
// messages are `osmformat.proto` and `fileformat.proto` of the OSM-binary
// project; the field numbers below are theirs.
package osmpbf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Size bounds from the format's specification: a BlobHeader is under 64 KiB
// and a Blob's uncompressed data under 32 MiB.
const (
	maxHeaderBytes = 64 << 10
	maxBlobBytes   = 32 << 20
)

// The required features this reader implements. Any other a file requires is
// refused, as the format specifies for a parser that does not implement one.
var implemented = map[string]bool{"OsmSchema-V0.6": true, "DenseNodes": true}

// Header is what a file's header block says about it.
type Header struct {
	// ReplicationTimestamp is the file's own timestamp, where its header
	// carries one (`osmosis_replication_timestamp`), and zero where it does
	// not.
	ReplicationTimestamp time.Time
}

// Node is an OSM node. Its coordinate is in WGS 84 degrees.
type Node struct {
	ID        int64
	Latitude  float64
	Longitude float64
	Tags      map[string]string
}

// Way is an OSM way: its node ids in order, and its tags.
type Way struct {
	ID    int64
	Nodes []int64
	Tags  map[string]string
}

// MemberType is the type of the object a relation member references.
type MemberType int

// The member types, numbered as `osmformat.proto`'s Relation.MemberType.
const (
	MemberNode     MemberType = 0
	MemberWay      MemberType = 1
	MemberRelation MemberType = 2
)

// Member is one member of a relation: what it references, and its role.
type Member struct {
	Type MemberType
	Ref  int64
	Role string
}

// Relation is an OSM relation: its members in order, and its tags.
type Relation struct {
	ID      int64
	Members []Member
	Tags    map[string]string
}

// Handler receives a file's contents in file order: the header first, then
// each node, way and relation as it is read. An error from any of them stops
// the read and is returned.
type Handler interface {
	Header(Header) error
	Node(Node) error
	Way(Way) error
	Relation(Relation) error
}

// Read reads the whole file r holds into h.
func Read(r io.Reader, h Handler) error {
	sawHeader := false

	for n := 0; ; n++ {
		typ, data, err := readBlob(r)
		if errors.Is(err, io.EOF) {
			if !sawHeader {
				return errors.New("the file holds no OSMHeader block")
			}

			return nil
		}

		if err != nil {
			return fmt.Errorf("block %d: %w", n, err)
		}

		switch typ {
		case "OSMHeader":
			if sawHeader {
				return fmt.Errorf("block %d: a second OSMHeader block", n)
			}

			sawHeader = true

			hdr, err := decodeHeader(data)
			if err != nil {
				return fmt.Errorf("block %d, the header: %w", n, err)
			}

			if err := h.Header(hdr); err != nil {
				return err
			}
		case "OSMData":
			if !sawHeader {
				return fmt.Errorf("block %d: an OSMData block before the OSMHeader", n)
			}

			if err := decodeBlock(data, h); err != nil {
				return fmt.Errorf("block %d: %w", n, err)
			}
		default:
			return fmt.Errorf("block %d has the type %q, which this reader does not know", n, typ)
		}
	}
}

// readBlob reads one BlobHeader and its Blob, returning the block's type and
// its uncompressed data. io.EOF is returned only at a clean end of file.
func readBlob(r io.Reader) (string, []byte, error) {
	var size [4]byte
	if _, err := io.ReadFull(r, size[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return "", nil, io.EOF
		}

		return "", nil, fmt.Errorf("the blob header's length: %w", err)
	}

	n := binary.BigEndian.Uint32(size[:])
	if n == 0 || n > maxHeaderBytes {
		return "", nil, fmt.Errorf("a blob header of %d bytes", n)
	}

	raw := make([]byte, n)
	if _, err := io.ReadFull(r, raw); err != nil {
		return "", nil, fmt.Errorf("the blob header: %w", err)
	}

	var (
		typ      string
		dataSize uint64
		haveSize bool
	)

	m := message{buf: raw}
	for {
		f, ok, err := m.next()
		if err != nil {
			return "", nil, fmt.Errorf("the blob header: %w", err)
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			if err := wantWire(f, wireBytes, "BlobHeader.type"); err != nil {
				return "", nil, err
			}

			typ = string(f.bytes)
		case 3:
			if err := wantWire(f, wireVarint, "BlobHeader.datasize"); err != nil {
				return "", nil, err
			}

			dataSize, haveSize = f.value, true
		}
	}

	if !haveSize || dataSize == 0 || dataSize > maxBlobBytes {
		return "", nil, fmt.Errorf("a blob of %d bytes", dataSize)
	}

	blob := make([]byte, dataSize)
	if _, err := io.ReadFull(r, blob); err != nil {
		return "", nil, fmt.Errorf("the blob: %w", err)
	}

	data, err := decodeBlob(blob)
	if err != nil {
		return "", nil, fmt.Errorf("the %s blob: %w", typ, err)
	}

	return typ, data, nil
}

// decodeBlob returns a Blob's data, uncompressed.
func decodeBlob(blob []byte) ([]byte, error) {
	var (
		raw, zdata []byte
		rawSize    uint64
		other      int
	)

	m := message{buf: blob}
	for {
		f, ok, err := m.next()
		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			raw = f.bytes
		case 2:
			rawSize = f.value
		case 3:
			zdata = f.bytes
		case 4, 5, 6, 7:
			other = f.num
		}
	}

	switch {
	case other != 0:
		return nil, fmt.Errorf("compressed by the method of Blob field %d, which this reader does not implement (it reads none and zlib)", other)
	case raw != nil:
		return raw, nil
	case zdata != nil:
		if rawSize == 0 || rawSize > maxBlobBytes {
			return nil, fmt.Errorf("an uncompressed size of %d bytes", rawSize)
		}

		zr, err := zlib.NewReader(bytes.NewReader(zdata))
		if err != nil {
			return nil, err
		}

		out := make([]byte, rawSize)
		if _, err := io.ReadFull(zr, out); err != nil {
			return nil, fmt.Errorf("inflating: %w", err)
		}

		var extra [1]byte
		if n, _ := zr.Read(extra[:]); n != 0 {
			return nil, errors.New("inflates to more than its stated size")
		}

		return out, nil
	default:
		return nil, errors.New("carries no data")
	}
}

// decodeHeader reads a HeaderBlock.
func decodeHeader(data []byte) (Header, error) {
	var hdr Header

	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return Header{}, err
		}

		if !ok {
			return hdr, nil
		}

		switch f.num {
		case 4: // required_features
			if err := wantWire(f, wireBytes, "HeaderBlock.required_features"); err != nil {
				return Header{}, err
			}

			if feature := string(f.bytes); !implemented[feature] {
				return Header{}, fmt.Errorf("the file requires the feature %q, which this reader does not implement", feature)
			}
		case 32: // osmosis_replication_timestamp, seconds since the epoch
			if err := wantWire(f, wireVarint, "HeaderBlock.osmosis_replication_timestamp"); err != nil {
				return Header{}, err
			}

			hdr.ReplicationTimestamp = time.Unix(int64(f.value), 0).UTC()
		}
	}
}

// block is the decoding context of one PrimitiveBlock.
type block struct {
	strings     [][]byte
	granularity int64
	latOffset   int64
	lonOffset   int64
}

// decodeBlock reads one PrimitiveBlock into h.
func decodeBlock(data []byte, h Handler) error {
	b := block{granularity: 100}

	var groups [][]byte

	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return err
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			if err := wantWire(f, wireBytes, "PrimitiveBlock.stringtable"); err != nil {
				return err
			}

			st := message{buf: f.bytes}
			for {
				s, ok, err := st.next()
				if err != nil {
					return fmt.Errorf("the string table: %w", err)
				}

				if !ok {
					break
				}

				if s.num == 1 {
					b.strings = append(b.strings, s.bytes)
				}
			}
		case 2:
			if err := wantWire(f, wireBytes, "PrimitiveBlock.primitivegroup"); err != nil {
				return err
			}

			groups = append(groups, f.bytes)
		case 17:
			b.granularity = int64(f.value)
		case 19:
			b.latOffset = int64(f.value)
		case 20:
			b.lonOffset = int64(f.value)
		}
	}

	if b.granularity <= 0 {
		return fmt.Errorf("a granularity of %d", b.granularity)
	}

	for _, g := range groups {
		if err := b.group(g, h); err != nil {
			return err
		}
	}

	return nil
}

func (b *block) str(i uint64) (string, error) {
	if i >= uint64(len(b.strings)) {
		return "", fmt.Errorf("string %d of a table of %d", i, len(b.strings))
	}

	return string(b.strings[i]), nil
}

// degrees converts a coordinate in the block's units to degrees. It divides
// by 1e9 rather than multiplying by 1e-9: the division is correctly rounded,
// so 57.7001000 in the file is the float64 57.7001, which the multiplication
// misses by an ulp.
func (b *block) degrees(offset, v int64) float64 {
	return float64(offset+b.granularity*v) / 1e9
}

// tags pairs a message's keys and values.
func (b *block) tags(keys, vals []uint64) (map[string]string, error) {
	if len(keys) != len(vals) {
		return nil, fmt.Errorf("%d keys and %d values", len(keys), len(vals))
	}

	if len(keys) == 0 {
		return nil, nil
	}

	out := make(map[string]string, len(keys))

	for i := range keys {
		k, err := b.str(keys[i])
		if err != nil {
			return nil, err
		}

		v, err := b.str(vals[i])
		if err != nil {
			return nil, err
		}

		out[k] = v
	}

	return out, nil
}

// group reads one PrimitiveGroup.
func (b *block) group(data []byte, h Handler) error {
	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		if f.wire != wireBytes {
			return fmt.Errorf("PrimitiveGroup field %d has wire type %d", f.num, f.wire)
		}

		switch f.num {
		case 1:
			err = b.node(f.bytes, h)
		case 2:
			err = b.dense(f.bytes, h)
		case 3:
			err = b.way(f.bytes, h)
		case 4:
			err = b.relation(f.bytes, h)
		default:
			err = fmt.Errorf("PrimitiveGroup field %d, which this reader does not know", f.num)
		}

		if err != nil {
			return err
		}
	}
}

func (b *block) node(data []byte, h Handler) error {
	var (
		n              Node
		lat, lon       int64
		keys, vals     []uint64
		hasLat, hasLon bool
	)

	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return fmt.Errorf("a node: %w", err)
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			n.ID = zigzag(f.value)
		case 2:
			keys, err = packed(f, keys)
		case 3:
			vals, err = packed(f, vals)
		case 8:
			lat, hasLat = zigzag(f.value), true
		case 9:
			lon, hasLon = zigzag(f.value), true
		}

		if err != nil {
			return fmt.Errorf("node %d: %w", n.ID, err)
		}
	}

	if !hasLat || !hasLon {
		return fmt.Errorf("node %d carries no coordinate", n.ID)
	}

	tags, err := b.tags(keys, vals)
	if err != nil {
		return fmt.Errorf("node %d: %w", n.ID, err)
	}

	n.Latitude, n.Longitude, n.Tags = b.degrees(b.latOffset, lat), b.degrees(b.lonOffset, lon), tags

	return h.Node(n)
}

func (b *block) dense(data []byte, h Handler) error {
	var (
		ids, lats, lons []int64
		kv              []uint64
	)

	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return fmt.Errorf("dense nodes: %w", err)
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			ids, err = packedSigned(f, ids)
		case 8:
			lats, err = packedSigned(f, lats)
		case 9:
			lons, err = packedSigned(f, lons)
		case 10:
			kv, err = packed(f, kv)
		}

		if err != nil {
			return fmt.Errorf("dense nodes: %w", err)
		}
	}

	if len(lats) != len(ids) || len(lons) != len(ids) {
		return fmt.Errorf("dense nodes: %d ids, %d latitudes, %d longitudes", len(ids), len(lats), len(lons))
	}

	var id, lat, lon int64

	k := 0

	for i := range ids {
		id, lat, lon = id+ids[i], lat+lats[i], lon+lons[i]

		var tags map[string]string

		// keys_vals is empty where no node of the group carries a tag;
		// otherwise every node's pairs end with a 0.
		if len(kv) > 0 {
			for {
				if k >= len(kv) {
					return fmt.Errorf("dense node %d: its tags run past the end of keys_vals", id)
				}

				if kv[k] == 0 {
					k++

					break
				}

				if k+1 >= len(kv) {
					return fmt.Errorf("dense node %d: a key without a value", id)
				}

				key, err := b.str(kv[k])
				if err != nil {
					return fmt.Errorf("dense node %d: %w", id, err)
				}

				val, err := b.str(kv[k+1])
				if err != nil {
					return fmt.Errorf("dense node %d: %w", id, err)
				}

				if tags == nil {
					tags = map[string]string{}
				}

				tags[key] = val
				k += 2
			}
		}

		if err := h.Node(Node{ID: id, Latitude: b.degrees(b.latOffset, lat), Longitude: b.degrees(b.lonOffset, lon), Tags: tags}); err != nil {
			return err
		}
	}

	if len(kv) > 0 && k != len(kv) {
		return fmt.Errorf("dense nodes: %d entries of keys_vals left over", len(kv)-k)
	}

	return nil
}

func (b *block) way(data []byte, h Handler) error {
	var (
		w          Way
		keys, vals []uint64
		refs       []int64
	)

	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return fmt.Errorf("a way: %w", err)
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			w.ID = int64(f.value)
		case 2:
			keys, err = packed(f, keys)
		case 3:
			vals, err = packed(f, vals)
		case 8:
			refs, err = packedSigned(f, refs)
		}

		if err != nil {
			return fmt.Errorf("way %d: %w", w.ID, err)
		}
	}

	tags, err := b.tags(keys, vals)
	if err != nil {
		return fmt.Errorf("way %d: %w", w.ID, err)
	}

	w.Tags = tags
	w.Nodes = make([]int64, len(refs))

	var id int64
	for i, d := range refs {
		id += d
		w.Nodes[i] = id
	}

	return h.Way(w)
}

func (b *block) relation(data []byte, h Handler) error {
	var (
		r          Relation
		hasID      bool
		keys, vals []uint64
		roles      []uint64
		refs       []int64
		types      []uint64
	)

	m := message{buf: data}
	for {
		f, ok, err := m.next()
		if err != nil {
			return fmt.Errorf("a relation: %w", err)
		}

		if !ok {
			break
		}

		switch f.num {
		case 1:
			r.ID, hasID = int64(f.value), true
		case 2:
			keys, err = packed(f, keys)
		case 3:
			vals, err = packed(f, vals)
		case 8:
			roles, err = packed(f, roles)
		case 9:
			refs, err = packedSigned(f, refs)
		case 10:
			types, err = packed(f, types)
		}

		if err != nil {
			return fmt.Errorf("relation %d: %w", r.ID, err)
		}
	}

	if !hasID {
		return errors.New("a relation without an id")
	}

	// The three member fields are parallel arrays; members a writer left
	// unequal cannot be paired, and a relation read with fewer members than
	// it has would assemble a smaller area than it draws.
	if len(roles) != len(refs) || len(types) != len(refs) {
		return fmt.Errorf("relation %d: %d member roles, %d member ids and %d member types", r.ID, len(roles), len(refs), len(types))
	}

	tags, err := b.tags(keys, vals)
	if err != nil {
		return fmt.Errorf("relation %d: %w", r.ID, err)
	}

	r.Tags = tags
	r.Members = make([]Member, len(refs))

	var id int64

	for i, d := range refs {
		id += d

		if types[i] > uint64(MemberRelation) {
			return fmt.Errorf("relation %d: member %d has the type %d, which the format does not define", r.ID, i, types[i])
		}

		role, err := b.str(roles[i])
		if err != nil {
			return fmt.Errorf("relation %d, member %d's role: %w", r.ID, i, err)
		}

		r.Members[i] = Member{Type: MemberType(types[i]), Ref: id, Role: role}
	}

	return h.Relation(r)
}
