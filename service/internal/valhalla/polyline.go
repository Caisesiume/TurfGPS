package valhalla

import (
	"errors"
	"fmt"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// shapePrecision is the factor a leg's shape is encoded at. Valhalla encodes
// with six decimal digits where the Google algorithm it follows uses five, and
// its decoding page says so in as many words (docs/docs/api/decoding.md at
// 3.9.1): a shape decoded at five lands ten times too far from the origin.
const shapePrecision = 1e6

// maxShapeChunks bounds how many five-bit chunks one encoded value may span. A
// delta of a full turn of longitude at shapePrecision is 3.6e8, which needs
// thirty bits with the sign bit — six chunks. Seven leaves one of slack, and
// refusing an eighth refuses a value that cannot be a coordinate before its
// shift can overflow.
const maxShapeChunks = 7

var errShapeTruncated = errors.New("the encoded shape ends inside a value")

// decodeShape decodes an encoded polyline at shapePrecision, the algorithm of
// the decoding page cited on shapePrecision. Latitude is encoded first in each
// pair, and every value after the first pair is a delta on the one before.
func decodeShape(encoded string) ([]planner.Point, error) {
	var (
		points   []planner.Point
		lat, lon int64
	)

	for i := 0; i < len(encoded); {
		dLat, next, err := decodeValue(encoded, i)
		if err != nil {
			return nil, err
		}

		dLon, next, err := decodeValue(encoded, next)
		if err != nil {
			return nil, err
		}

		i = next
		lat += dLat
		lon += dLon

		p := planner.Point{Latitude: float64(lat) / shapePrecision, Longitude: float64(lon) / shapePrecision}
		if p.Latitude < -90 || p.Latitude > 90 || p.Longitude < -180 || p.Longitude > 180 {
			return nil, fmt.Errorf("vertex %d decodes to (%v, %v), which is not a coordinate", len(points), p.Latitude, p.Longitude)
		}

		points = append(points, p)
	}

	return points, nil
}

// decodeValue decodes the signed value that starts at encoded[i] and returns
// it with the index just past it.
func decodeValue(encoded string, i int) (int64, int, error) {
	var result int64

	for chunk := 0; ; chunk++ {
		if i >= len(encoded) {
			return 0, 0, errShapeTruncated
		}

		if chunk == maxShapeChunks {
			return 0, 0, fmt.Errorf("a value at byte %d of the encoded shape runs past %d chunks", i, maxShapeChunks)
		}

		b := int64(encoded[i]) - 63
		i++

		if b < 0 || b > 63 {
			return 0, 0, fmt.Errorf("byte %d of the encoded shape is %q, outside the encoding's alphabet", i-1, encoded[i-1])
		}

		result |= (b & 0x1f) << (5 * chunk)

		if b < 0x20 {
			break
		}
	}

	if result&1 != 0 {
		return ^(result >> 1), i, nil
	}

	return result >> 1, i, nil
}
