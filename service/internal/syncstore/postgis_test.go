package syncstore

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// ---------------------------------------------------------------------------
// THIS FILE SENDS SQL TO A SERVER, AND ONLY WHEN ONE IS NAMED.
//
// envMigratedCopy names a PostgreSQL database that
// migrations/0001_zone_store.sql has been applied to. CI's `migrations` job
// creates one, sets the variable, and fails if this test skips or does not run.
// Everywhere else it is unset and the test skips, so a green `go test ./...` on
// a host without a database says nothing about what is asserted here.
//
// IT REFUSES A DATABASE THAT HOLDS DATA. Every case empties `zone` and
// `zone_incoming` when it ends, so the test begins by requiring `zone` and
// `sync_run` to be empty: pointed at a copy holding a single zone or a single
// run, it fails before writing anything rather than truncating them.
// ---------------------------------------------------------------------------

const envMigratedCopy = "TURFGPS_TEST_DATABASE_URL"

// verifyRatio is the floor the staging assertions are applied with here. It is
// the strictest (0, 1] admits rather than the configured one: every refresh in
// this file carries at least as many zones as the copy holds, so no floor
// configuration can set refuses what this one admits.
const verifyRatio = 1.0

// fixtureZone is one zone, rendered both as a record of the all-zones response
// and, for a zone the copy is to hold already, as a row of `zone`. Every field
// but these three is derived from the id, so a held row and the record a
// refresh returns for it differ in exactly the fields a case chooses.
type fixtureZone struct {
	id      int32
	created time.Time
	typeID  *int16
}

// record is the zone as the response carries it. `type` is omitted when typeID
// is nil, which is how the response carries a zone without one.
func (z fixtureZone) record() string {
	typ := ""
	if z.typeID != nil {
		typ = fmt.Sprintf(`"type":{"id":%d,"name":"a type"},`, *z.typeID)
	}

	return fmt.Sprintf(`{"id":%d,"name":"zone %d","latitude":59.3,"longitude":18.06,`+
		`"dateCreated":%q,"totalTakeovers":%d,"takeoverPoints":120,"pointsPerHour":3,%s`+
		`"region":{"id":141,"name":"a region"}}`,
		z.id, z.id, z.created.Format(time.RFC3339), z.id%97, typ)
}

// migratedCopy opens the database envMigratedCopy names, or skips.
func migratedCopy(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(envMigratedCopy)
	if dsn == "" {
		t.Skipf("%s is not set, so there is no migrated copy to send these statements to and nothing here is measured", envMigratedCopy)
	}

	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatalf("opening the copy %s names: %v", envMigratedCopy, err)
	}

	t.Cleanup(pool.Close)

	var held int64

	if err := pool.QueryRow(t.Context(),
		`SELECT (SELECT count(*) FROM zone) + (SELECT count(*) FROM sync_run)`).Scan(&held); err != nil {
		t.Fatalf("reading the copy %s names — has 0001 been applied to it? %v", envMigratedCopy, err)
	}

	if held != 0 {
		t.Fatalf("the database %s names holds %d zone or sync_run rows; this test empties `zone` and refuses to run where that would destroy anything", envMigratedCopy, held)
	}

	return pool
}

// emptiedAfter empties what a case wrote when it ends, so each case starts
// from an empty copy whatever the previous one left.
func emptiedAfter(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `TRUNCATE zone, zone_incoming`); err != nil {
			t.Errorf("emptying the copy after the case: %v", err)
		}
	})
}

// hold puts z in the copy as a row it already holds, written directly rather
// than through a refresh, so the case that reads it depends on no merge
// branch but the one it is about.
func hold(t *testing.T, pool *pgxpool.Pool, z fixtureZone) {
	t.Helper()

	if _, err := pool.Exec(t.Context(), `
INSERT INTO zone (id, name, latitude, longitude, date_created, total_takeovers,
                  takeover_points, points_per_hour, type_id, region_id,
                  region_name, first_seen_at, last_changed_at)
VALUES ($1, $2, 59.3, 18.06, $3, $4, 120, 3, $5, 141, 'a region', now(), now())`,
		z.id, fmt.Sprintf("zone %d", z.id), z.created, z.id%97, z.typeID); err != nil {
		t.Fatalf("seeding zone %d as held: %v", z.id, err)
	}
}

// refresh performs one refresh of the copy from a response carrying zones: the
// ingest mapping, then the store's calls in the order zonesync's runner makes
// them — stage, inspect, the staging assertions, merge. The runner itself is
// unexported by design (see zonesync's package doc), so this replays it.
//
// Any refusal on the way is a failure of the case, named by the step.
func refresh(t *testing.T, store *Store, zones ...fixtureZone) zonesync.Merged {
	t.Helper()

	records := make([]string, len(zones))
	for i, z := range zones {
		records[i] = z.record()
	}

	parsed, err := zonesync.ParseAllZones([]byte("[" + strings.Join(records, ",") + "]"))
	if err != nil {
		t.Fatalf("the refresh was refused by the ingest mapping: %v", err)
	}

	if err := store.Stage(t.Context(), parsed); err != nil {
		t.Fatalf("the refresh was refused at the staging load: %v", err)
	}

	staged, err := store.Inspect(t.Context(), time.Now())
	if err != nil {
		t.Fatalf("the refresh was refused at the inspection: %v", err)
	}

	if reasons := staged.Verify(verifyRatio); len(reasons) > 0 {
		t.Fatalf("the refresh was refused by the staging assertions: %v", reasons)
	}

	merged, err := store.Merge(t.Context())
	if err != nil {
		t.Fatalf("the refresh was refused at the merge: %v", err)
	}

	return merged
}

// storedDateCreated reads what the copy holds for id.
func storedDateCreated(t *testing.T, pool *pgxpool.Pool, id int32) time.Time {
	t.Helper()

	var stored time.Time

	if err := pool.QueryRow(t.Context(), `SELECT date_created FROM zone WHERE id = $1`, id).Scan(&stored); err != nil {
		t.Fatalf("reading zone %d's stored date_created: %v", id, err)
	}

	return stored
}

// TestTheSyncWritePathAgainstAMigratedCopy sends the refresh's statements to
// PostgreSQL and reads back what they stored. Its cases share one copy, so
// they run in order and each empties it when it ends.
func TestTheSyncWritePathAgainstAMigratedCopy(t *testing.T) {
	pool := migratedCopy(t)

	store, err := New(pool, nil)
	if err != nil {
		t.Fatalf("building the store: %v", err)
	}

	// `FR-035` (#38). The two cases are one obligation under two conditions,
	// and the first is the one a write path gets wrong while reading as right:
	// a refresh that only ever meets ids the copy does not hold passes against
	// a merge that never updates anything. So the held row is built on purpose,
	// identical to the returned record in every column but date_created —
	// which also makes date_created alone what the merge's change test has to
	// see. Both carry a type, so neither case turns on how a record without
	// one is stored: that is #39's, below.
	typed := int16(3)

	t.Run("a held id's stored dateCreated is replaced by the returned one", func(t *testing.T) {
		emptiedAfter(t, pool)

		held := fixtureZone{id: 101, created: time.Date(2019, 5, 6, 7, 8, 9, 0, time.UTC), typeID: &typed}
		returned := held
		returned.created = time.Date(2021, 3, 4, 5, 6, 7, 0, time.UTC)

		hold(t, pool, held)

		merged := refresh(t, store, returned)

		if stored := storedDateCreated(t, pool, held.id); !stored.Equal(returned.created) {
			t.Errorf("zone %d was held with date_created %s and the refresh returned %s, and the copy still holds %s — a held id's date frozen at first sight is the failure `FR-035` exists against",
				held.id, held.created.Format(time.RFC3339), returned.created.Format(time.RFC3339), stored.UTC().Format(time.RFC3339))
		}

		if merged.Updated != 1 || merged.Inserted != 0 {
			t.Errorf("the merge reports %d inserted and %d updated for one held zone whose date the refresh revised, want 0 and 1", merged.Inserted, merged.Updated)
		}
	})

	t.Run("an id new to the copy is stored with the returned dateCreated", func(t *testing.T) {
		emptiedAfter(t, pool)

		returned := fixtureZone{id: 102, created: time.Date(2022, 9, 10, 11, 12, 13, 0, time.UTC), typeID: &typed}

		merged := refresh(t, store, returned)

		if stored := storedDateCreated(t, pool, returned.id); !stored.Equal(returned.created) {
			t.Errorf("zone %d was new to the copy and the refresh returned date_created %s, and the copy holds %s",
				returned.id, returned.created.Format(time.RFC3339), stored.UTC().Format(time.RFC3339))
		}

		if merged.Inserted != 1 || merged.Updated != 0 {
			t.Errorf("the merge reports %d inserted and %d updated for one zone new to the copy, want 1 and 0", merged.Inserted, merged.Updated)
		}
	})

	// `FR-036` (#39), the store's half: a set in which some records carry no
	// `type` is staged, asserted and merged as one in which every record
	// carries one, and every record is held afterwards. A NOT NULL on the
	// column, or an assertion counting the absence, is the form this refuses.
	// The run's half — the mapping, and what the run raises — is
	// `internal/zonesync`'s absenttype_test.go.
	for _, set := range []struct {
		name  string
		zones []fixtureZone
	}{
		{"every record carries a type", []fixtureZone{
			{id: 201, created: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), typeID: &typed},
			{id: 202, created: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), typeID: &typed},
			{id: 203, created: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC), typeID: &typed},
			{id: 204, created: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC), typeID: &typed},
		}},
		{"some records carry none", []fixtureZone{
			{id: 201, created: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), typeID: &typed},
			{id: 202, created: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)},
			{id: 203, created: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC), typeID: &typed},
			{id: 204, created: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC)},
		}},
	} {
		t.Run("a zone set in which "+set.name+" is held whole", func(t *testing.T) {
			emptiedAfter(t, pool)

			merged := refresh(t, store, set.zones...)

			var held int

			if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM zone`).Scan(&held); err != nil {
				t.Fatalf("counting the zones held: %v", err)
			}

			if held != len(set.zones) || merged.Inserted != len(set.zones) {
				t.Errorf("a refresh of %d records left %d held and reports %d inserted, want every record retained", len(set.zones), held, merged.Inserted)
			}
		})
	}
}
