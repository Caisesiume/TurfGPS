package osmpbf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"os"
	"testing"
	"time"
)

// A protocol-buffer encoder for building test files, the reader's inverse.

func pbVarint(v uint64) []byte {
	var b []byte
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}

	return append(b, byte(v))
}

func pbZig(v int64) uint64 { return uint64((v << 1) ^ (v >> 63)) }

func pbField(num int, wire int, payload []byte) []byte {
	return append(pbVarint(uint64(num<<3|wire)), payload...)
}

func pbBytes(num int, b []byte) []byte {
	return pbField(num, wireBytes, append(pbVarint(uint64(len(b))), b...))
}

func pbUint(num int, v uint64) []byte { return pbField(num, wireVarint, pbVarint(v)) }

func pbPacked(num int, vs ...uint64) []byte {
	var b []byte
	for _, v := range vs {
		b = append(b, pbVarint(v)...)
	}

	return pbBytes(num, b)
}

// fileBlock frames one block, its Blob raw or zlib-compressed, or carrying the
// field of another compression method.
func fileBlock(typ string, data []byte, compression string) []byte {
	var blob []byte

	switch compression {
	case "raw":
		blob = pbBytes(1, data)
	case "zlib":
		var z bytes.Buffer
		w := zlib.NewWriter(&z)
		_, _ = w.Write(data)
		_ = w.Close()
		blob = append(pbUint(2, uint64(len(data))), pbBytes(3, z.Bytes())...)
	case "lzma":
		blob = append(pbUint(2, uint64(len(data))), pbBytes(4, data)...)
	}

	header := append(pbBytes(1, []byte(typ)), pbUint(3, uint64(len(blob)))...)

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(header)))

	return append(append(size[:], header...), blob...)
}

func headerBlock(features ...string) []byte {
	var b []byte
	for _, f := range features {
		b = append(b, pbBytes(4, []byte(f))...)
	}

	return append(b, pbUint(32, 1_700_000_000)...)
}

// dataBlock is a PrimitiveBlock of one group holding two dense nodes — the
// second tagged access=private — and one way through both, tagged
// highway=service.
func dataBlock() []byte {
	strs := pbBytes(1, nil)
	for _, s := range []string{"access", "private", "highway", "service"} {
		strs = append(strs, pbBytes(1, []byte(s))...)
	}

	dense := append(pbPacked(1, pbZig(10), pbZig(1)), // ids 10, 11
		pbPacked(8, pbZig(577_000_000), pbZig(1_000))...) // granularity 100: 57.7, 57.7001
	dense = append(dense, pbPacked(9, pbZig(119_600_000), pbZig(2_000))...) // 11.96, 11.9602
	dense = append(dense, pbPacked(10, 0, 1, 2, 0)...)                      // none; access=private

	way := append(pbUint(1, 99), pbPacked(2, 3)...)
	way = append(way, pbPacked(3, 4)...)
	way = append(way, pbPacked(8, pbZig(10), pbZig(1))...)

	group := append(pbBytes(2, dense), pbBytes(3, way)...)

	return append(pbBytes(1, strs), pbBytes(2, group)...)
}

type collect struct {
	header Header
	nodes  []Node
	ways   []Way
	rels   []int64
}

func (c *collect) Header(h Header) error   { c.header = h; return nil }
func (c *collect) Node(n Node) error       { c.nodes = append(c.nodes, n); return nil }
func (c *collect) Way(w Way) error         { c.ways = append(c.ways, w); return nil }
func (c *collect) Relation(id int64) error { c.rels = append(c.rels, id); return nil }

// TestAFileIsReadAsWritten: the header's timestamp, dense nodes with their
// coordinates and tags, and a way's delta-coded node ids, from raw and from
// zlib blobs alike.
func TestAFileIsReadAsWritten(t *testing.T) {
	t.Parallel()

	for _, compression := range []string{"raw", "zlib"} {
		file := append(fileBlock("OSMHeader", headerBlock("OsmSchema-V0.6", "DenseNodes"), compression), fileBlock("OSMData", dataBlock(), compression)...)

		var c collect
		if err := Read(bytes.NewReader(file), &c); err != nil {
			t.Fatalf("%s: Read: %v", compression, err)
		}

		if !c.header.ReplicationTimestamp.Equal(time.Unix(1_700_000_000, 0)) {
			t.Errorf("%s: the timestamp is %v", compression, c.header.ReplicationTimestamp)
		}

		if len(c.nodes) != 2 || c.nodes[0].ID != 10 || c.nodes[1].ID != 11 {
			t.Fatalf("%s: the nodes are %+v", compression, c.nodes)
		}

		if n := c.nodes[1]; n.Latitude != 57.7001 || n.Longitude != 11.9602 || n.Tags["access"] != "private" || len(c.nodes[0].Tags) != 0 {
			t.Errorf("%s: node 11 is %+v, want (57.7001, 11.9602) tagged access=private", compression, n)
		}

		if len(c.ways) != 1 || c.ways[0].ID != 99 || c.ways[0].Tags["highway"] != "service" || len(c.ways[0].Nodes) != 2 || c.ways[0].Nodes[1] != 11 {
			t.Errorf("%s: the ways are %+v", compression, c.ways)
		}
	}
}

// TestAFileThatCannotBeReadWholeIsRefused: a partial read would load tables
// missing features, so every file this reader cannot read whole is an error.
func TestAFileThatCannotBeReadWholeIsRefused(t *testing.T) {
	t.Parallel()

	good := append(fileBlock("OSMHeader", headerBlock("OsmSchema-V0.6"), "raw"), fileBlock("OSMData", dataBlock(), "raw")...)

	for name, file := range map[string][]byte{
		"a history file":           fileBlock("OSMHeader", headerBlock("OsmSchema-V0.6", "HistoricalInformation"), "raw"),
		"another compression":      append(fileBlock("OSMHeader", headerBlock(), "raw"), fileBlock("OSMData", dataBlock(), "lzma")...),
		"no header":                fileBlock("OSMData", dataBlock(), "raw"),
		"an unknown block":         append(fileBlock("OSMHeader", headerBlock(), "raw"), fileBlock("OSMIndex", nil, "raw")...),
		"a file cut short":         good[:len(good)-3],
		"an empty file":            nil,
		"a header block cut short": good[:6],
		"a second header":          append(good, fileBlock("OSMHeader", headerBlock(), "raw")...),
	} {
		if err := Read(bytes.NewReader(file), &collect{}); err == nil {
			t.Errorf("%s was read without an error", name)
		}
	}
}

// TestAHandlerErrorStopsTheRead: an error from the handler is returned, and
// nothing after it is delivered.
func TestAHandlerErrorStopsTheRead(t *testing.T) {
	t.Parallel()

	file := append(fileBlock("OSMHeader", headerBlock(), "raw"), fileBlock("OSMData", dataBlock(), "raw")...)
	stop := errors.New("stop")

	h := &stopAt{err: stop}
	if err := Read(bytes.NewReader(file), h); !errors.Is(err, stop) || h.ways != 0 {
		t.Errorf("Read gave %v after %d ways, want the handler's error before any way", err, h.ways)
	}
}

type stopAt struct {
	collect
	err  error
	ways int
}

func (s *stopAt) Node(Node) error { return s.err }
func (s *stopAt) Way(Way) error   { s.ways++; return nil }

// TestTheCommittedExtractReadsWhole: the extract the tiles are built from,
// read to its end, holds the objects its README counts.
func TestTheCommittedExtractReadsWhole(t *testing.T) {
	t.Parallel()

	f, err := os.Open("../valhalla/testdata/goteborg-inom-vallgraven.osm.pbf")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var c collect
	if err := Read(f, &c); err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got := []int{len(c.nodes), len(c.ways), len(c.rels)}; got[0] != 13786 || got[1] != 4078 || got[2] != 57 {
		t.Errorf("the extract holds %v nodes, ways and relations, want [13786 4078 57] as its README records", got)
	}
}
