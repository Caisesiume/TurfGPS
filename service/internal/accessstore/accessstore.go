// Package accessstore is the PostGIS implementation of the planner's
// AccessMarkings port (#255): it reads which extract the access-marking tables
// were loaded from, and which markings a car leg's edges cross, from the
// tables of `Architecture.md § The access-marking tables`
// (`migrations/0003_access_markings.sql`), and measures the run at a location
// the user entered against the least that joins it to an unmarked way (#248,
// `run.go`, on `migrations/0004_access_run.sql`).
//
// It reads and never writes. The tables are written by the loader,
// `internal/accessload`, which nothing on the request path imports, for the
// reason `internal/zonestore` and `internal/syncstore` are two packages.
//
// THE CHECK IS ONE STATEMENT, AND IT LIVES IN THE MIGRATION. Crossings calls
// `access_crossings`, the SQL function that is step 3 of `§ The route check`,
// so the statement EXPLAINed by the migration's verify file is the one served.
// Nothing here says what is marked: that is `access_feature_marked`, the
// tables' generated column, and this package reads the column.
//
// IT FAILS CLOSED, at every step it can fail:
//
//   - With no barrier tolerance configured, Crossings checks nothing and
//     returns an error, so no car leg is served (`Architecture.md § What is
//     not decided`; `config.LoadAccessMarkings`).
//   - A row of kind `unchecked` — an input the statement could not check whole
//     — is an error, never a pass.
//   - ExtractSHA256 is an error where no load is current, and where the
//     tables hold no feature at all: a table emptied after its load would
//     otherwise read every way as unmarked, the vacuous pass `FR-125`'s
//     `Risk` describes.
//   - Any error from the database, a table absent included, is returned as
//     one; the planner refuses the leg or the journey on it.
package accessstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// ErrNoTolerance is what Crossings returns while no barrier tolerance is
// configured: no leg is checked, so none is served.
var ErrNoTolerance = errors.New("no barrier tolerance is configured, so no car leg is checked (TURFGPS_ACCESS_BARRIER_TOLERANCE_METRES)")

// ErrNoCurrentLoad is what ExtractSHA256 returns where no load of the tables
// is current.
var ErrNoCurrentLoad = errors.New("no load of the access-marking tables is current")

// ErrEmpty is what ExtractSHA256 returns where a load is current and the
// tables hold no way and no barrier.
var ErrEmpty = errors.New("the access-marking tables hold no feature")

// ErrUnchecked is what Crossings returns where the statement could not check
// the edges it was given whole.
var ErrUnchecked = errors.New("the access-marking check could not read the leg whole")

// Store reads the access-marking tables.
type Store struct {
	pool      *pgxpool.Pool
	tolerance *float64
}

var _ planner.AccessMarkings = (*Store)(nil)

// New returns a Store reading through pool, counting a marked barrier as
// passed within barrierToleranceMetres of a stretch. A nil tolerance is the
// unconfigured constant: the Store then checks no leg.
func New(pool *pgxpool.Pool, barrierToleranceMetres *float64) (*Store, error) {
	if pool == nil {
		return nil, errors.New("no pool on the store")
	}

	var tolerance *float64

	if barrierToleranceMetres != nil {
		v := *barrierToleranceMetres
		tolerance = &v
	}

	return &Store{pool: pool, tolerance: tolerance}, nil
}

// extractQuery reads the current load's extract, and whether the tables hold
// any feature.
const extractQuery = `
SELECT l.extract_sha256,
       EXISTS (SELECT 1 FROM public.access_way) OR EXISTS (SELECT 1 FROM public.access_barrier)
  FROM public.osm_extract_load l
 WHERE l.current`

// ExtractSHA256 returns the SHA-256 of the extract the current load read.
func (s *Store) ExtractSHA256(ctx context.Context) (string, error) {
	var (
		sha  string
		held bool
	)

	err := s.pool.QueryRow(ctx, extractQuery).Scan(&sha, &held)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoCurrentLoad
	}

	if err != nil {
		return "", fmt.Errorf("reading the current load of the access-marking tables: %w", err)
	}

	if !held {
		return "", ErrEmpty
	}

	return sha, nil
}

// crossingsQuery is step 3 of `§ The route check`, as the migration defines
// it.
const crossingsQuery = `
SELECT edge_index, kind, osm_type, osm_id, longitude, latitude
  FROM public.access_crossings($1::bigint[], $2::integer[], $3::double precision[], $4::double precision[], $5::double precision)`

// Crossings returns every marking the edges cross.
func (s *Store) Crossings(ctx context.Context, edges []planner.Edge) ([]planner.Crossing, error) {
	if s.tolerance == nil {
		return nil, ErrNoTolerance
	}

	ways := make([]int64, len(edges))

	var (
		pointEdges []int32
		lons, lats []float64
	)

	for i, e := range edges {
		ways[i] = e.WayID

		for _, p := range e.Shape {
			// The statement's edges are numbered from 1, in order.
			pointEdges = append(pointEdges, int32(i+1))
			lons = append(lons, p.Longitude)
			lats = append(lats, p.Latitude)
		}
	}

	rows, err := s.pool.Query(ctx, crossingsQuery, ways, pointEdges, lons, lats, *s.tolerance)
	if err != nil {
		return nil, fmt.Errorf("checking the leg against the access markings: %w", err)
	}
	defer rows.Close()

	var out []planner.Crossing

	for rows.Next() {
		var (
			edge     int32
			kind     string
			osmType  *string
			osmID    *int64
			lon, lat *float64
		)

		if err := rows.Scan(&edge, &kind, &osmType, &osmID, &lon, &lat); err != nil {
			return nil, fmt.Errorf("reading a crossing: %w", err)
		}

		if kind == "unchecked" {
			return nil, ErrUnchecked
		}

		if osmType == nil || osmID == nil || edge < 1 || int(edge) > len(edges) {
			return nil, fmt.Errorf("%w: a %s crossing names no feature or no edge", ErrUnchecked, kind)
		}

		c := planner.Crossing{Edge: int(edge) - 1, Kind: kind, OSMType: *osmType, OSMID: *osmID}
		if lon != nil && lat != nil {
			c.At = planner.Point{Latitude: *lat, Longitude: *lon}
		}

		out = append(out, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("checking the leg against the access markings: %w", err)
	}

	return out, nil
}
