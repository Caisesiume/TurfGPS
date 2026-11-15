package osmpbf

import (
	"errors"
	"fmt"
)

// The protocol-buffer wire format, the subset the OSM PBF messages use
// (https://protobuf.dev/programming-guides/encoding/). Four wire types occur in
// them; any other is a malformed message and is refused, never skipped.
const (
	wireVarint  = 0
	wireFixed64 = 1
	wireBytes   = 2
	wireFixed32 = 5
)

var errTruncated = errors.New("truncated protocol-buffer message")

// message walks the fields of one encoded protocol-buffer message.
type message struct {
	buf []byte
	pos int
}

// field is one field of a message: its number, its wire type, and either its
// integer value or its bytes.
type field struct {
	num   int
	wire  int
	value uint64
	bytes []byte
}

// next returns the message's next field, false at the end, or an error where
// the message is malformed.
func (m *message) next() (field, bool, error) {
	if m.pos >= len(m.buf) {
		return field{}, false, nil
	}

	key, err := m.varint()
	if err != nil {
		return field{}, false, err
	}

	f := field{num: int(key >> 3), wire: int(key & 7)}
	if f.num <= 0 {
		return field{}, false, fmt.Errorf("field number %d", f.num)
	}

	switch f.wire {
	case wireVarint:
		f.value, err = m.varint()
	case wireFixed64:
		f.value, err = m.fixed(8)
	case wireFixed32:
		f.value, err = m.fixed(4)
	case wireBytes:
		var n uint64

		n, err = m.varint()
		if err == nil {
			if n > uint64(len(m.buf)-m.pos) {
				return field{}, false, errTruncated
			}

			f.bytes = m.buf[m.pos : m.pos+int(n)]
			m.pos += int(n)
		}
	default:
		return field{}, false, fmt.Errorf("field %d has wire type %d, which no OSM PBF message uses", f.num, f.wire)
	}

	if err != nil {
		return field{}, false, err
	}

	return f, true, nil
}

func (m *message) varint() (uint64, error) {
	var v uint64

	for shift := uint(0); shift < 64; shift += 7 {
		if m.pos >= len(m.buf) {
			return 0, errTruncated
		}

		b := m.buf[m.pos]
		m.pos++
		v |= uint64(b&0x7f) << shift

		if b < 0x80 {
			return v, nil
		}
	}

	return 0, errors.New("a varint longer than 64 bits")
}

func (m *message) fixed(n int) (uint64, error) {
	if len(m.buf)-m.pos < n {
		return 0, errTruncated
	}

	var v uint64
	for i := range n {
		v |= uint64(m.buf[m.pos+i]) << (8 * i)
	}

	m.pos += n

	return v, nil
}

// zigzag decodes a sint64.
func zigzag(v uint64) int64 {
	return int64(v>>1) ^ -int64(v&1)
}

// packed decodes a packed repeated varint field, or a single unpacked value
// where the writer did not pack it.
func packed(f field, out []uint64) ([]uint64, error) {
	switch f.wire {
	case wireVarint:
		return append(out, f.value), nil
	case wireBytes:
		m := message{buf: f.bytes}
		for m.pos < len(m.buf) {
			v, err := m.varint()
			if err != nil {
				return nil, err
			}

			out = append(out, v)
		}

		return out, nil
	default:
		return nil, fmt.Errorf("field %d: a repeated integer with wire type %d", f.num, f.wire)
	}
}

// packedSigned decodes a packed repeated sint64 field.
func packedSigned(f field, out []int64) ([]int64, error) {
	raw, err := packed(f, nil)
	if err != nil {
		return nil, err
	}

	for _, v := range raw {
		out = append(out, zigzag(v))
	}

	return out, nil
}

// wantWire refuses a field of a wire type its message does not give it.
func wantWire(f field, wire int, what string) error {
	if f.wire != wire {
		return fmt.Errorf("%s (field %d) has wire type %d, want %d", what, f.num, f.wire, wire)
	}

	return nil
}
