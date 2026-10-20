// Package corridor resolves the zones lying within a route corridor — `FR-023`.
//
// IT DECLARES THE PORT AND HOLDS NO ADAPTER. The zones come through
// ZoneRepository, the corridor half of the `ZoneRepository` port in
// `Architecture.md § Ports and adapters`, which `internal/zonestore` implements
// against the local synced copy with Q1 of `Architecture.md § The queries the
// schema exists to serve`. Nothing here, and nothing the adapter imports, can
// issue a Turf request: `noturf_test.go` checks that over the import graph
// rather than trusting this sentence.
//
// IT CHOOSES NO HALF-WIDTH. The rule that derives one for a route is decided in
// `CalculationSpecification.md § Bounding the candidate set`, which `FR-063`
// carries (#20); `Architecture.md § What is unproven` item 7 now holds open only
// the row counts it yields at its cap. The caller passes it in metres; this
// package only refuses one that would resolve a corridor nobody asked for.
package corridor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// Point is one vertex of a route.
//
// THE AXES ARE NAMED, NEVER POSITIONAL. Which of the two a geometry writes first
// is the trap `postgis-migration-protocol § Standing schema facts` records this
// project having shipped once; a named field leaves no order to get wrong here,
// and the one place an order is written is the adapter's encoding.
type Point struct {
	Latitude  float64
	Longitude float64
}

// Zone is one zone of the local copy lying within a corridor: Q1's columns and
// no others.
//
// IT CARRIES NO OWNERSHIP, and must not acquire any. `FR-026` forbids deriving
// a statement of ownership from the local copy, and the copy holds none to
// derive one from.
type Zone struct {
	ID             int32
	Name           string
	Latitude       float64
	Longitude      float64
	DateCreated    time.Time
	TotalTakeovers int32
	TakeoverPoints int16
	PointsPerHour  int16

	// TypeID is nil where the zone has no type, which is the ordinary case
	// (`Architecture.md § Retrieving zones`) and never an anomaly.
	TypeID *int16
}

// ZoneRepository is what corridor resolution needs of the zone store.
//
// It is declared here, by its consumer, as `internal/zonesync` declares the
// sync's ports, so the adapter is substitutable and this package depends on no
// provider.
type ZoneRepository interface {
	// ZonesInCorridor returns every zone of the local copy lying within
	// halfWidthMetres of the line through route, in no particular order.
	ZonesInCorridor(ctx context.Context, route []Point, halfWidthMetres float64) ([]Zone, error)
}

// Resolve returns the zones of the local copy lying within halfWidthMetres of
// route.
//
// A route is at least two points, each a coordinate the copy could hold, and the
// half-width is a positive, finite distance. Anything else is refused before the
// repository is asked: a corridor of zero or negative width is an empty answer
// that reads exactly like an empty stretch of map.
func Resolve(ctx context.Context, repo ZoneRepository, route []Point, halfWidthMetres float64) ([]Zone, error) {
	if repo == nil {
		return nil, errors.New("no zone repository")
	}

	if err := validate(route, halfWidthMetres); err != nil {
		return nil, err
	}

	zones, err := repo.ZonesInCorridor(ctx, route, halfWidthMetres)
	if err != nil {
		return nil, fmt.Errorf("resolving the corridor against the local zone copy: %w", err)
	}

	return zones, nil
}

func validate(route []Point, halfWidthMetres float64) error {
	if math.IsNaN(halfWidthMetres) || math.IsInf(halfWidthMetres, 0) || halfWidthMetres <= 0 {
		return fmt.Errorf("the corridor half-width %v m is not a positive, finite distance", halfWidthMetres)
	}

	if len(route) < 2 {
		return fmt.Errorf("a route needs at least two points to have a corridor, and this one has %d", len(route))
	}

	for i, p := range route {
		if !within(p.Latitude, -90, 90) || !within(p.Longitude, -180, 180) {
			return fmt.Errorf("route point %d (latitude %v, longitude %v) is not a coordinate", i, p.Latitude, p.Longitude)
		}
	}

	return nil
}

// within reports whether v lies in [lo, hi]. NaN fails both comparisons, so it
// is never within anything.
func within(v, lo, hi float64) bool {
	return v >= lo && v <= hi
}
