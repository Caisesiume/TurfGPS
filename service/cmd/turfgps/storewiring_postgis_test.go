//go:build postgis

package main

// The planner the composition root builds, planning through a job's pool on a
// migrated store (#249 AC2, AC4, AC7): the zone copy, the access-marking
// tables and the current load are read from PostGIS, and the routing engine
// and its tile identity are a fake answering as Valhalla does.
//
// IT IS BUILT ONLY UNDER THE `postgis` TAG, AND THAT IS NOT A SKIP. CI's
// `migrations` job builds it against the scratch database it has just
// migrated; without one named it fails rather than skips. It refuses any
// database but an empty one on loopback, writes the rows each case needs, and
// deletes them.

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/accessstore"
	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
	"github.com/Caisesiume/TurfGPS/service/internal/planner"
	"github.com/Caisesiume/TurfGPS/service/internal/zonestore"
)

// undefinedTable is PostgreSQL's SQLSTATE for a relation that does not exist.
const undefinedTable = "42P01"

// testDatabaseEnv names the scratch database. It is a test's input and never
// the service's configuration.
const testDatabaseEnv = "TURFGPS_TEST_DATABASE_URL"

// The rows a case writes. The zone is the one `held_postgis_test.go` writes;
// way 1 is the way the fake engine's leg runs along.
const (
	insertZone = `
		INSERT INTO zone (id, name, latitude, longitude, date_created, total_takeovers,
		                  takeover_points, points_per_hour, region_id, region_name,
		                  first_seen_at, last_changed_at)
		VALUES (9000249, 'root', 57.7044, 11.9662, now(), 0, 65, 0, 141, 'Göteborg', now(), now())`

	insertLoad = `
		INSERT INTO osm_extract_load (extract_sha256, way_features, way_rows, barrier_features, barrier_rows,
		                              barrier_way_rows, way_exit_rows, current)
		VALUES ($1, 1, 1, 0, 0, 0, 0, true)`

	// insertWay writes a way marked private along the line $2 gives.
	insertWay = `
		INSERT INTO access_way (way_id, access, geom, nodes)
		VALUES ($1, 'private', ST_GeogFromText('SRID=4326;LINESTRING(' || $2::text || ')'), '{1,2}')`

	// alongTheLeg is the fake engine's leg; awayFromIt is a line in Stockholm.
	alongTheLeg = "11.9611 57.7036, 11.9714 57.7052"
	awayFromIt  = "18.0600 59.3300, 18.0700 59.3310"
)

// emptied is every table a case writes, in an order the foreign keys allow.
var emptied = []string{"access_way_exit", "access_barrier_way", "access_barrier", "access_area", "access_way", "osm_extract_load", "zone", "sync_run"}

// scratchURL is the scratch database's URL, refused unless it is on loopback.
func scratchURL(t *testing.T) string {
	t.Helper()

	raw := os.Getenv(testDatabaseEnv)
	if raw == "" {
		t.Fatalf("%s is not set: this file is built only under the postgis tag, for a scratch database with migrations/ applied, and it fails rather than skips without one", testDatabaseEnv)
	}

	cfg, err := pgxpool.ParseConfig(raw)
	if err != nil {
		t.Fatalf("%s does not parse: %v", testDatabaseEnv, err)
	}

	switch cfg.ConnConfig.Host {
	case "127.0.0.1", "localhost", "::1":
	default:
		t.Fatalf("%s names host %q, not loopback; this test writes rows and runs only against a throwaway local instance", testDatabaseEnv, cfg.ConnConfig.Host)
	}

	return raw
}

// rootOnStore is the root's planner on the sync's pool over the database at
// dsn, as main builds it, with the fake engine and the tile record for
// tileSHA, and the pool it reads through.
func rootOnStore(t *testing.T, dsn, tileSHA string) (*planner.Planner, *pgxpool.Pool, *fakeEngine, config.LookupFunc) {
	t.Helper()

	syncCfg := wireable()
	syncCfg.DatabaseURL = dsn

	pools, err := openJobPools(t.Context(), &syncCfg, nil)
	if err != nil {
		t.Fatalf("openJobPools: %v", err)
	}
	t.Cleanup(pools.close)

	engine := newFakeEngine(t, tileStamp(t, tileSHA))
	lookup := lookupOnly(map[string]string{
		config.EnvValhallaURL:                  engine.URL,
		config.EnvTileIdentityFile:             writeTileRecord(t, tileSHA),
		config.EnvAccessBarrierTolerance:       "2",
		config.EnvAccessBarrierToleranceOrigin: "this test's own figure",
		config.EnvEndTolerance:                 endsAnywhere,
		config.EnvEndToleranceOrigin:           endsAnywhereOrigin,
	})

	plans, err := newPlanner(lookup, pools.request())
	if err != nil || plans == nil {
		t.Fatalf("newPlanner on the sync's pool gave %v, %v; want a planner", plans, err)
	}

	return plans, pools.request(), engine, lookup
}

// emptyStore refuses a store holding anything a case writes, and empties what
// the test wrote when it ends.
func emptyStore(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	for _, table := range emptied {
		var n int64
		if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM `+table).Scan(&n); err != nil {
			t.Fatalf("counting %s: %v", table, err)
		}

		if n != 0 {
			t.Fatalf("%s holds %d row(s); this test writes and deletes rows and refuses anything but an empty, freshly migrated scratch copy", table, n)
		}
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for _, table := range emptied {
			if _, err := pool.Exec(ctx, `DELETE FROM `+table); err != nil {
				t.Errorf("emptying %s: %v", table, err)
			}
		}
	})
}

func execSQL(t *testing.T, pool *pgxpool.Pool, stmt string, args ...any) {
	t.Helper()

	if _, err := pool.Exec(t.Context(), stmt, args...); err != nil {
		t.Fatalf("%s: %v", strings.TrimSpace(stmt), err)
	}
}

// TestTheRootPlansThroughTheStoreOnAJobsPool is #249 AC2, AC4 and AC7 over a
// migrated store, one state after another:
//
//   - no zone held: 503, the copy none (AC2, empty);
//   - a zone held and no load current, then a load current with no feature:
//     no journey (AC4, empty);
//   - a load whose SHA-256 is not the tile identity read: no journey,
//     ErrExtractMismatch (AC7);
//   - the leg's way marked private in the store: no journey, the crossing
//     named (AC4, the root reading the store's markings);
//   - the store filled and the load the tiles': the journey plans (AC2,
//     filled; AC7, equal).
func TestTheRootPlansThroughTheStoreOnAJobsPool(t *testing.T) {
	dsn := scratchURL(t)
	tiles := clearSHA256
	other := aboveFloat

	plans, pool, engine, lookup := rootOnStore(t, dsn, tiles)
	emptyStore(t, pool)

	handler, err := newHandler(lookup, httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	// AC2, empty: the no-copy 503, before the engine is asked.
	if rec := postJourney(t, handler); rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "holds no local copy") {
		t.Errorf("with no zone held POST /journeys answered %d %q, want 503 saying no copy is held", rec.Code, rec.Body)
	}

	if got := engine.asked(); got != "" {
		t.Errorf("with no zone held the engine was asked at %s, want nothing", got)
	}

	execSQL(t, pool, insertZone)

	// AC4, empty: no load current.
	if journey, err := plans.Plan(t.Context(), journeyRequest); journey.Legs != nil || !errors.Is(err, accessstore.ErrNoCurrentLoad) {
		t.Errorf("with no load current Plan gave %d leg(s) and %v; want no journey and ErrNoCurrentLoad", len(journey.Legs), err)
	}

	// AC4, empty: a load current and no feature.
	execSQL(t, pool, insertLoad, tiles)

	if journey, err := plans.Plan(t.Context(), journeyRequest); journey.Legs != nil || !errors.Is(err, accessstore.ErrEmpty) {
		t.Errorf("with a load current and no feature Plan gave %d leg(s) and %v; want no journey and ErrEmpty", len(journey.Legs), err)
	}

	// AC7: a load from another extract than the tiles'.
	execSQL(t, pool, insertWay, 2, awayFromIt)
	execSQL(t, pool, `UPDATE osm_extract_load SET extract_sha256 = $1 WHERE current`, other)

	if journey, err := plans.Plan(t.Context(), journeyRequest); journey.Legs != nil || !errors.Is(err, planner.ErrExtractMismatch) {
		t.Errorf("with the tiles from %s and the load from %s Plan gave %d leg(s) and %v; want no journey and ErrExtractMismatch", tiles, other, len(journey.Legs), err)
	}

	if n := engine.routed(); n != 0 {
		t.Errorf("before the store vouched for a leg the engine was asked %d leg(s), want none", n)
	}

	execSQL(t, pool, `UPDATE osm_extract_load SET extract_sha256 = $1 WHERE current`, tiles)

	// AC4: the leg's own way marked in the store refuses it.
	execSQL(t, pool, insertWay, 1, alongTheLeg)

	journey, err := plans.Plan(t.Context(), journeyRequest)

	var crossing *planner.CrossingError
	if journey.Legs != nil || !errors.As(err, &crossing) || len(crossing.Crossings) != 1 || crossing.Crossings[0].OSMID != 1 {
		t.Errorf("with the leg's way 1 marked private in the store Plan gave %d leg(s) and %v; want no journey and the crossing of way 1", len(journey.Legs), err)
	}

	execSQL(t, pool, `DELETE FROM access_way WHERE way_id = 1`)

	// AC2, filled; AC7, equal.
	rec := postJourney(t, handler)
	if rec.Code != http.StatusOK {
		t.Errorf("with the store filled and the load the tiles' POST /journeys answered %d %q, want 200", rec.Code, rec.Body)
	}
}

// TestNoJourneyWhereTheAccessTablesAreAbsent is #249 AC4's other half: a store
// migrated for the zone copy and not for the access markings — the tables
// absent — holds a copy and plans no journey. It migrates a database of its
// own with `0001` and `0002` alone, and drops it.
func TestNoJourneyWhereTheAccessTablesAreAbsent(t *testing.T) {
	dsn := scratchURL(t)

	admin, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatalf("connecting to the scratch database: %v", err)
	}
	t.Cleanup(admin.Close)

	const db = "turfgps_root_no_access_tables"

	execSQL(t, admin, `DROP DATABASE IF EXISTS `+db+` WITH (FORCE)`)
	execSQL(t, admin, `CREATE DATABASE `+db+` TEMPLATE template0`)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if _, err := admin.Exec(ctx, `DROP DATABASE IF EXISTS `+db+` WITH (FORCE)`); err != nil {
			t.Errorf("dropping %s: %v", db, err)
		}
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("%s does not parse as a URL: %v", testDatabaseEnv, err)
	}

	u.Path = "/" + db

	zoneOnly, err := zonestore.Open(t.Context(), u.String())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(zoneOnly.Close)

	for _, name := range []string{"0001_zone_store.sql", "0002_zone_feed.sql"} {
		sql, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}

		execSQL(t, zoneOnly, string(sql))
	}

	execSQL(t, zoneOnly, insertZone)

	plans, _, engine, _ := rootOnStore(t, u.String(), clearSHA256)

	journey, err := plans.Plan(t.Context(), journeyRequest)
	var absent *pgconn.PgError
	if journey.Legs != nil || !errors.Is(err, planner.ErrAccessUnchecked) || !errors.As(err, &absent) || absent.Code != undefinedTable {
		t.Errorf("with the access-marking tables absent Plan gave %d leg(s) and %v; want no journey and the store's undefined-table error", len(journey.Legs), err)
	}

	if n := engine.routed(); n != 0 {
		t.Errorf("with the access-marking tables absent the engine was asked %d leg(s), want none", n)
	}
}
