package zonestore

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Caisesiume/TurfGPS/service/internal/corridor"
)

// Reader is the local copy's half of the corridor port: `FR-023` resolves a
// corridor here and nowhere that reaches Turf.
var _ corridor.ZoneRepository = (*Reader)(nil)

// corridorSQL is Q1 of `Architecture.md § The queries the schema exists to
// serve`, with the route passed as EWKT.
//
// IT IS THE SHAPE PART E OF `migrations/0001_zone_store.verify.sql` EXPLAINS:
// ST_DWithin over `geom` against ST_GeogFromText, which CI's `migrations` job
// asserts reads `zone` through `zone_geom_gist`. A rewrite that leaves the
// containment test as anything else — a distance compared after the fact, a
// buffer intersected — has left the evidence behind with it.
//
// It returns no ownership column because `zone` has none (`FR-026`).
const corridorSQL = `
SELECT id, name, latitude, longitude, date_created, total_takeovers,
       takeover_points, points_per_hour, type_id
FROM   zone
WHERE  ST_DWithin(geom, ST_GeogFromText($1), $2)`

// ZonesInCorridor returns every zone of the local copy lying within
// halfWidthMetres of the line through route. It reads PostGIS and nothing else.
func (r *Reader) ZonesInCorridor(ctx context.Context, route []corridor.Point, halfWidthMetres float64) ([]corridor.Zone, error) {
	rows, err := r.pool.Query(ctx, corridorSQL, routeEWKT(route), halfWidthMetres)
	if err != nil {
		return nil, fmt.Errorf("querying the corridor: %w", err)
	}
	defer rows.Close()

	var zones []corridor.Zone

	for rows.Next() {
		var z corridor.Zone

		// In corridorSQL's column order, one field per column.
		if err := rows.Scan(&z.ID, &z.Name, &z.Latitude, &z.Longitude, &z.DateCreated,
			&z.TotalTakeovers, &z.TakeoverPoints, &z.PointsPerHour, &z.TypeID); err != nil {
			return nil, fmt.Errorf("reading a corridor row: %w", err)
		}

		z.DateCreated = z.DateCreated.UTC()
		zones = append(zones, z)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the corridor: %w", err)
	}

	return zones, nil
}

// routeEWKT renders a route as an SRID-4326 LINESTRING.
//
// THIS IS WHERE A ROUTE'S AXIS ORDER IS WRITTEN, and the only place: WKT is
// longitude first, as ST_MakePoint is in the generated `geom` column. Each figure
// is the shortest decimal that parses back to the same float64, in plain
// notation, so no precision is lost and no exponent reaches the WKT parser.
func routeEWKT(route []corridor.Point) string {
	var b strings.Builder

	b.WriteString("SRID=4326;LINESTRING(")

	for i, p := range route {
		if i > 0 {
			b.WriteByte(',')
		}

		b.WriteString(strconv.FormatFloat(p.Longitude, 'f', -1, 64))
		b.WriteByte(' ')
		b.WriteString(strconv.FormatFloat(p.Latitude, 'f', -1, 64))
	}

	b.WriteByte(')')

	return b.String()
}
