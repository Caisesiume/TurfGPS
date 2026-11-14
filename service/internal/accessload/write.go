package accessload

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Record is what one load wrote.
type Record struct {
	ID                           int64
	SHA256                       string
	WayRows, BarrierRows         int64
	BarrierWayRows, WayExitRows  int64
	WayFeatures, BarrierFeatures int64
}

// keyColumns are the twelve access keys' columns, in Keys' order.
const keyColumns = `motorcar, motor_vehicle, vehicle, access,
	motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
	motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward`

// The staging tables, session-private and dropped at commit. They hold what
// the file says, and nothing derived from it: the geometry is built, and the
// marking computed, as the rows enter the feature tables.
var staging = []string{
	`CREATE TEMP TABLE access_load_way (
		way_id bigint NOT NULL, ` + typedKeys() + `,
		nodes bigint[] NOT NULL, lons double precision[] NOT NULL, lats double precision[] NOT NULL
	) ON COMMIT DROP`,
	`CREATE TEMP TABLE access_load_barrier (
		node_id bigint NOT NULL, barrier text, ` + typedKeys() + `,
		lon double precision NOT NULL, lat double precision NOT NULL
	) ON COMMIT DROP`,
	`CREATE TEMP TABLE access_load_barrier_way (way_id bigint NOT NULL, node_id bigint NOT NULL) ON COMMIT DROP`,
	`CREATE TEMP TABLE access_load_junction (way_id bigint NOT NULL, node_id bigint NOT NULL) ON COMMIT DROP`,
}

// typedKeys declares each key's column as text.
func typedKeys() string {
	out := ""

	for _, k := range Keys {
		if out != "" {
			out += ", "
		}

		out += column(k) + " text"
	}

	return out
}

// column is the column a key is stored in: its name with the colon of a
// directional form as an underscore.
func column(key string) string {
	b := []byte(key)
	for i := range b {
		if b[i] == ':' {
			b[i] = '_'
		}
	}

	return string(b)
}

// The swap: the feature tables replaced in the transaction that records the
// load as current. The axis order is named: ST_MakePoint takes longitude, then
// latitude.
const (
	truncateFeatures = `TRUNCATE public.access_way_exit, public.access_barrier_way, public.access_area, public.access_barrier, public.access_way`

	insertWays = `
INSERT INTO public.access_way (way_id, ` + keyColumns + `, geom, nodes)
SELECT s.way_id, ` + keyColumns + `,
       (SELECT ST_MakeLine(ST_SetSRID(ST_MakePoint(p.lon, p.lat), 4326) ORDER BY p.ord)
          FROM unnest(s.lons, s.lats) WITH ORDINALITY AS p(lon, lat, ord))::geography,
       s.nodes
  FROM pg_temp.access_load_way s`

	insertBarriers = `
INSERT INTO public.access_barrier (node_id, barrier, ` + keyColumns + `, geom)
SELECT node_id, barrier, ` + keyColumns + `, ST_SetSRID(ST_MakePoint(lon, lat), 4326)::geography
  FROM pg_temp.access_load_barrier`

	insertBarrierWays = `
INSERT INTO public.access_barrier_way (way_id, node_id)
SELECT way_id, node_id FROM pg_temp.access_load_barrier_way`

	// A node of a marked way where a `highway` way not marked meets it: the
	// meeting way is in the junctions because it is a highway way, and it is
	// not marked where access_way does not hold it or holds it unmarked.
	insertExits = `
INSERT INTO public.access_way_exit (way_id, node_id)
SELECT DISTINCT m.way_id, j.node_id
  FROM public.access_way m
 CROSS JOIN LATERAL unnest(m.nodes) AS n(node_id)
  JOIN pg_temp.access_load_junction j ON j.node_id = n.node_id AND j.way_id <> m.way_id
  LEFT JOIN public.access_way o ON o.way_id = j.way_id
 WHERE m.marked
   AND (o.way_id IS NULL OR NOT o.marked)`

	retireCurrent = `UPDATE public.osm_extract_load SET current = false WHERE current`

	insertRecord = `
INSERT INTO public.osm_extract_load
       (extract_sha256, extract_timestamp, way_features, way_rows, barrier_features, barrier_rows,
        barrier_way_rows, way_exit_rows, area_features, area_rows, current)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL, NULL, true)
RETURNING id`
)

// Load replaces the access-marking tables with ex in one transaction and
// records it as the current load. Any failure rolls the whole load back.
func Load(ctx context.Context, pool *pgxpool.Pool, ex *Extract) (Record, error) {
	if ex == nil {
		return Record{}, errors.New("no extract to load")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Record{}, fmt.Errorf("opening the load's transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, stmt := range staging {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return Record{}, fmt.Errorf("creating a staging table: %w", err)
		}
	}

	if err := stage(ctx, tx, ex); err != nil {
		return Record{}, err
	}

	rec := Record{SHA256: ex.SHA256, WayFeatures: ex.WayFeatures, BarrierFeatures: ex.BarrierFeatures}

	if _, err := tx.Exec(ctx, truncateFeatures); err != nil {
		return Record{}, fmt.Errorf("emptying the feature tables: %w", err)
	}

	for _, step := range []struct {
		what string
		sql  string
		rows *int64
	}{
		{"the ways", insertWays, &rec.WayRows},
		{"the barriers", insertBarriers, &rec.BarrierRows},
		{"the barriers' ways", insertBarrierWays, &rec.BarrierWayRows},
		{"the ways' exits", insertExits, &rec.WayExitRows},
	} {
		tag, err := tx.Exec(ctx, step.sql)
		if err != nil {
			return Record{}, fmt.Errorf("writing %s: %w", step.what, err)
		}

		*step.rows = tag.RowsAffected()
	}

	// The record's own CHECK refuses counts that differ; this says which.
	if rec.WayRows != ex.WayFeatures || rec.BarrierRows != ex.BarrierFeatures || rec.BarrierWayRows != int64(len(ex.BarrierWays)) {
		return Record{}, fmt.Errorf("a feature was lost on load: %d ways read and %d written, %d barriers read and %d written, %d barrier-way pairs read and %d written",
			ex.WayFeatures, rec.WayRows, ex.BarrierFeatures, rec.BarrierRows, len(ex.BarrierWays), rec.BarrierWayRows)
	}

	if _, err := tx.Exec(ctx, retireCurrent); err != nil {
		return Record{}, fmt.Errorf("retiring the current load: %w", err)
	}

	var timestamp any
	if !ex.Timestamp.IsZero() {
		timestamp = ex.Timestamp
	}

	if err := tx.QueryRow(ctx, insertRecord, ex.SHA256, timestamp, ex.WayFeatures, rec.WayRows, ex.BarrierFeatures,
		rec.BarrierRows, rec.BarrierWayRows, rec.WayExitRows).Scan(&rec.ID); err != nil {
		return Record{}, fmt.Errorf("recording the load: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Record{}, fmt.Errorf("committing the load: %w", err)
	}

	return rec, nil
}

// stage copies ex into the staging tables.
func stage(ctx context.Context, tx pgx.Tx, ex *Extract) error {
	keyCols := make([]string, len(Keys))
	for i, k := range Keys {
		keyCols[i] = column(k)
	}

	values := func(v Values) []any {
		out := make([]any, len(v))
		for i := range v {
			out[i] = v[i]
		}

		return out
	}

	wayCols := append(append([]string{"way_id"}, keyCols...), "nodes", "lons", "lats")
	ways := make([][]any, len(ex.Ways))

	for i, w := range ex.Ways {
		ways[i] = append(append([]any{w.ID}, values(w.Values)...), w.Nodes, w.Lons, w.Lats)
	}

	barrierCols := append(append([]string{"node_id", "barrier"}, keyCols...), "lon", "lat")
	barriers := make([][]any, len(ex.Barriers))

	for i, b := range ex.Barriers {
		barriers[i] = append(append([]any{b.ID, b.Barrier}, values(b.Values)...), b.Lon, b.Lat)
	}

	pairs := func(ps []Pair) [][]any {
		out := make([][]any, len(ps))
		for i, p := range ps {
			out[i] = []any{p.Way, p.Node}
		}

		return out
	}

	for _, c := range []struct {
		table string
		cols  []string
		rows  [][]any
	}{
		{"access_load_way", wayCols, ways},
		{"access_load_barrier", barrierCols, barriers},
		{"access_load_barrier_way", []string{"way_id", "node_id"}, pairs(ex.BarrierWays)},
		{"access_load_junction", []string{"way_id", "node_id"}, pairs(ex.Junctions)},
	} {
		n, err := tx.CopyFrom(ctx, pgx.Identifier{"pg_temp", c.table}, c.cols, pgx.CopyFromRows(c.rows))
		if err != nil {
			return fmt.Errorf("staging %s: %w", c.table, err)
		}

		if n != int64(len(c.rows)) {
			return fmt.Errorf("staging %s: %d of %d rows copied", c.table, n, len(c.rows))
		}
	}

	return nil
}
