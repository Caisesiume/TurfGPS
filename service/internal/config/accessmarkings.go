package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// The access-marking check's configuration (#255): the barrier tolerance of
// `Architecture.md § The route check`, steps 3 and 5 — how near a marked
// barrier must lie to a leg's stretch of shape to count as passed.
//
// IT HAS NO DEFAULT, AND NONE MAY BE ADDED. It is an enforcement constant
// (`CalculationSpecification.md § Conventions`) whose value no document gives
// (`Architecture.md § What is not decided`): a figure compiled in here would be
// the unexplained literal those sections forbid, on a safety path. While it is
// unset, the service serves no car leg: the store that checks legs refuses
// every one (`internal/accessstore`).
//
// ITS STRICT DIRECTION IS UPWARD, AND IT IS HELD AT A FLOOR (#265). A larger
// tolerance counts more barriers as passed and refuses more legs; a smaller one
// misses barriers, and zero misses every barrier a shape does not pass exactly
// through. Which value is safe is the measurement that section describes, and
// the Owner's to set; what this block holds is that a value arrives with the
// origin that justifies it, and that no value below BarrierToleranceFloorMetres
// is taken, since below it a barrier the leg runs through can lie farther from
// the shape than the tolerance by the engine's coordinate rounding alone.
const (
	EnvAccessBarrierTolerance       = "TURFGPS_ACCESS_BARRIER_TOLERANCE_METRES"
	EnvAccessBarrierToleranceOrigin = "TURFGPS_ACCESS_BARRIER_TOLERANCE_ORIGIN"
)

// BarrierToleranceFloorMetres is the least barrier tolerance the service takes
// (#265): the most a node of the extract can lie from the shape the engine
// returns through it, by the precisions at which the engine stores and returns
// that node's coordinate, at every latitude a coordinate can take. It is
// derived here, from the precisions cited below and the spheroid PostGIS
// measures on, and from nothing else. It is a bound, not a tolerance: it
// chooses no value, and a larger bound may raise it, never a smaller lower it.
//
// THE NODE'S PATH, AT VALHALLA 3.9.1 (the version CI pins,
// `.github/workflows/gates.yml`, `VALHALLA_IMAGE`). The extract's node is the
// coordinate the `.osm.pbf` file gives, as `internal/osmpbf` reads it, to
// within float64 rounding. The leg's stretch of shape is cut from the `shape`
// of the `trace_attributes` answer (`internal/valhalla`, `tracedEdges`). A node
// of a way the leg runs along reaches that shape through four roundings, each
// bounded on each axis, in degrees. Paths are Valhalla's at tag 3.9.1 unless
// libosmium's are named.
//
//  1. libosmium reads it into a Location of fixed precision 1e-7
//     (`coordinate_precision`, libosmium include/osmium/osm/location.hpp:68, at
//     a44ae7be, the commit 3.9.1 vendors as third_party/libosmium, which it
//     builds with unless PREFER_EXTERNAL_DEPS is set, CMakeLists.txt:54, :157
//     and :162, and docker/Dockerfile:46 does not set it) by integer
//     division of the file's nanodegrees by 100, which truncates (libosmium
//     include/osmium/io/detail/pbf.hpp:72-74, pbf_decoder.hpp:316-322): less
//     than one osmiumQuantumDegrees.
//  2. `OSMNode::set_latlng` rounds it to 1e-7 (valhalla/mjolnir/osmnode.h:91-97,
//     called at src/mjolnir/pbfgraphparser.cc:2106), and the edge's shape is
//     read back from those integers (valhalla/mjolnir/osmdata.h:29-40,
//     src/mjolnir/graphbuilder.cc:189 and :582): half of
//     osmNodeQuantumDegrees. Every node of the way enters the edge's shape
//     there, one point each (graphbuilder.cc:575-583), so the node is a vertex
//     of it.
//  3. The tile stores the edge's shape with `encode7` at ENCODE_PRECISION,
//     1e6 — "we store 6 digits of precision in the tiles"
//     (valhalla/midgard/encoded.h:10-20; src/mjolnir/edgeinfobuilder.cc:83,
//     from src/mjolnir/graphtilebuilder.cc:781 and graphbuilder.cc:1005-1008)
//     — rounding each point's own coordinate, not its delta
//     (encoded.h:317-318): half of tileShapeQuantumDegrees.
//  4. The trace answer's leg is built from those shapes, decoded into doubles
//     (src/baldr/edgeinfo.cc:450; `PointLL` is `GeoPoint<double>`,
//     valhalla/midgard/pointll.h:297): an edge whole, or trimmed at a leg's end
//     with every vertex between its trim points kept
//     (src/thor/triplegbuilder.cc:2055-2133; src/midgard/util.cc:230-260). For
//     `edge_walk` it is built there (src/thor/trace_attributes_action.cc:46-48,
//     src/thor/trace_route_action.cc:160), then encoded with `encode` at
//     ENCODE_PRECISION (triplegbuilder.cc:2315; encoded.h:250 and :274-275),
//     which the answer carries as it is (src/tyr/trace_serializer.cc:574): half
//     of legShapeQuantumDegrees.
//
// Each rounding is counted whole, though step 4 re-rounds a value step 3 left
// on its own grid: crediting that would rest on floating-point reasoning no
// source states, and leaving it uncredited errs upward, the strict direction.
// The most a node is moved on either axis is their sum, nodeDisplacementDegrees.
//
// FROM DEGREES TO METRES, AT EVERY LATITUDE. PostGIS measures the tolerance on
// geography (`access_crossings`, `migrations/0003_access_markings.sql`,
// ST_DWithin), on the WGS84 spheroid. On it, the path that moves a point by Δφ
// in latitude and Δλ in longitude, both in radians, along a straight line in
// (φ, λ) is no longer than c·√(Δφ² + Δλ²), where c = a/(1−f) is the polar
// radius of curvature: its element is √(M(φ)²·dφ² + (N(φ)·cos φ)²·dλ²), and
// the meridian radius M(φ) and the prime-vertical radius N(φ), which bounds
// N(φ)·cos φ, are each at most c at every latitude, reaching it at the poles.
// The geodesic is no longer than that path, and the distance from the node to
// a shape through its vertex is no longer than the distance to the vertex. A
// node moved by at most d degrees on each axis therefore lies within
// √2·c·d·π/180 metres of the shape, at any latitude and longitude, and that is
// the floor.
const BarrierToleranceFloorMetres = math.Sqrt2 * wgs84PolarRadiusOfCurvatureMetres * nodeDisplacementDegrees * math.Pi / 180

// The precisions of the floor's derivation, each a quantum in degrees, cited
// on BarrierToleranceFloorMetres. A deployment of another Valhalla version
// re-reads each citation first.
const (
	osmiumQuantumDegrees    = 1e-7 // step 1: truncated, so less than one quantum
	osmNodeQuantumDegrees   = 1e-7 // step 2: rounded, so at most half of one
	tileShapeQuantumDegrees = 1e-6 // step 3: rounded
	legShapeQuantumDegrees  = 1e-6 // step 4: rounded

	nodeDisplacementDegrees = osmiumQuantumDegrees + osmNodeQuantumDegrees/2 +
		tileShapeQuantumDegrees/2 + legShapeQuantumDegrees/2
)

// The WGS84 ellipsoid's defining semi-major axis and inverse flattening (NGA,
// "World Geodetic System 1984", NGA.STND.0036_1.0.0_WGS84, Table 3.1): the
// spheroid of SRID 4326, on which PostGIS measures geography.
const (
	wgs84SemiMajorAxisMetres          = 6378137.0
	wgs84InverseFlattening            = 298.257223563
	wgs84PolarRadiusOfCurvatureMetres = wgs84SemiMajorAxisMetres / (1 - 1/wgs84InverseFlattening)
)

// AccessMarkings is the configuration the access-marking check needs.
type AccessMarkings struct {
	// BarrierToleranceMetres is the barrier tolerance, in metres: finite and
	// at or above BarrierToleranceFloorMetres.
	BarrierToleranceMetres float64

	// BarrierToleranceOrigin is where the value comes from — the measurement,
	// or the record that set it — as the deployment states it.
	BarrierToleranceOrigin string
}

// LoadAccessMarkings reads the access-marking check's configuration.
//
// It returns (nil, nil) when neither variable is set, which is a service that
// checks no leg and so serves none. A tolerance without its origin, or an
// origin without a tolerance, is refused as a partial configuration; so is a
// tolerance that is not a finite number of metres, and one below
// BarrierToleranceFloorMetres, zero included. A refusal stops the service at
// start, so it serves no car leg, as with an unset tolerance.
func LoadAccessMarkings(lookup LookupFunc) (*AccessMarkings, error) {
	raw, haveValue := lookup(EnvAccessBarrierTolerance)
	origin, haveOrigin := lookup(EnvAccessBarrierToleranceOrigin)

	raw, origin = strings.TrimSpace(raw), strings.TrimSpace(origin)
	haveValue, haveOrigin = haveValue && raw != "", haveOrigin && origin != ""

	switch {
	case !haveValue && !haveOrigin:
		return nil, nil
	case !haveValue:
		return nil, fmt.Errorf("the barrier tolerance is partially configured: %s is set and %s is not — the origin names a value, so set both, or neither to serve no car leg",
			EnvAccessBarrierToleranceOrigin, EnvAccessBarrierTolerance)
	case !haveOrigin:
		return nil, fmt.Errorf("%s is set without %s: an enforcement constant is taken with its documented origin, never without one",
			EnvAccessBarrierTolerance, EnvAccessBarrierToleranceOrigin)
	}

	metres, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(metres) || math.IsInf(metres, 0) {
		return nil, fmt.Errorf("%s is %q, want a finite number of metres at or above the floor, config.BarrierToleranceFloorMetres (%g m)",
			EnvAccessBarrierTolerance, raw, BarrierToleranceFloorMetres)
	}

	if metres < BarrierToleranceFloorMetres {
		return nil, fmt.Errorf("%s is %q, below the floor, config.BarrierToleranceFloorMetres (%g m, derived in service/internal/config/accessmarkings.go): the most a node of the extract can lie from the shape Valhalla 3.9.1 returns through it, by the coordinate precisions cited there, so a smaller tolerance can miss a barrier the leg runs through",
			EnvAccessBarrierTolerance, raw, BarrierToleranceFloorMetres)
	}

	return &AccessMarkings{BarrierToleranceMetres: metres, BarrierToleranceOrigin: origin}, nil
}
