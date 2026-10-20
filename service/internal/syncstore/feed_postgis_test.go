package syncstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// ---------------------------------------------------------------------------
// THIS FILE SENDS SQL TO A SERVER, AND ONLY WHEN ONE IS NAMED — the database
// postgis_test.go's envMigratedCopy names, with migrations/0001 and 0002
// applied. CI's `migrations` job is that database and fails on a skip.
//
// What it measures is the zone feed's write (#232): a batch decided by
// zonesync.ParseZoneFeed from a feed body, written by ApplyFeed and read back —
// a zone new to the copy and a held one left unchanged (AC3), an item that cannot be
// stored leaving no row and a record (AC4), the latest feed time recorded,
// read back by a new reader and never moved back (AC2), and nothing written at
// all while the sync holds its key.
// ---------------------------------------------------------------------------

// feedCopy opens the migrated copy and refuses one whose feed tables hold
// anything, since every case empties them.
func feedCopy(t *testing.T) (*pgxpool.Pool, *Store) {
	t.Helper()

	pool := migratedCopy(t)

	var held int64

	if err := pool.QueryRow(t.Context(),
		`SELECT (SELECT count(*) FROM zone_feed_state) + (SELECT count(*) FROM zone_feed_refusal)`).Scan(&held); err != nil {
		t.Fatalf("reading the feed tables — has 0002 been applied? %v", err)
	}

	if held != 0 {
		t.Fatalf("the feed tables of the database %s names hold %d row(s); this test empties them and refuses to run where that would destroy anything", envMigratedCopy, held)
	}

	store, err := New(pool, nil)
	if err != nil {
		t.Fatalf("building the store: %v", err)
	}

	return pool, store
}

func feedEmptiedAfter(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `TRUNCATE zone, zone_feed_state, zone_feed_refusal`); err != nil {
			t.Errorf("emptying the copy after the case: %v", err)
		}
	})
}

// feedItem renders one zone-feed item. zone is the zone object's members.
func feedItem(itemTime, zone string) string {
	return `{"type":"zone","time":"` + itemTime + `","zone":{` + zone + `,"dateLastTaken":"2026-10-19T22:00:00+0000"}}`
}

// poll decides body as a poll starting now would, and writes it.
func poll(t *testing.T, store *Store, items ...string) zonesync.FeedApplied {
	t.Helper()

	_, now, err := store.FeedCursor(t.Context())
	if err != nil {
		t.Fatalf("reading the latest feed time: %v", err)
	}

	batch, err := zonesync.ParseZoneFeed([]byte("["+strings.Join(items, ",")+"]"), now)
	if err != nil {
		t.Fatalf("the feed was refused: %v", err)
	}

	applied, err := store.ApplyFeed(t.Context(), batch)
	if err != nil {
		t.Fatalf("the batch was not written: %v", err)
	}

	return applied
}

// storedZone reads every ingest column of id back as a zonesync.Zone, or
// reports that there is no such row.
func storedZone(t *testing.T, pool *pgxpool.Pool, id int32) (zonesync.Zone, bool) {
	t.Helper()

	var z zonesync.Zone

	err := pool.QueryRow(t.Context(), `
SELECT id, name, latitude, longitude, date_created, total_takeovers, takeover_points,
       points_per_hour, type_id, region_id, region_name, region_country, area_id, area_name
FROM   zone WHERE id = $1`, id).Scan(
		&z.ID, &z.Name, &z.Latitude, &z.Longitude, &z.DateCreated, &z.TotalTakeovers, &z.TakeoverPoints,
		&z.PointsPerHour, &z.TypeID, &z.RegionID, &z.RegionName, &z.RegionCountry, &z.AreaID, &z.AreaName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zonesync.Zone{}, false
		}

		t.Fatalf("reading zone %d: %v", id, err)
	}

	z.DateCreated = z.DateCreated.UTC()

	return z, true
}

type storedRefusal struct {
	zoneID   *int32
	itemTime *time.Time
	reason   string
	subject  string
}

func storedRefusals(t *testing.T, pool *pgxpool.Pool) []storedRefusal {
	t.Helper()

	rows, err := pool.Query(t.Context(), `SELECT zone_id, item_time, reason, subject FROM zone_feed_refusal ORDER BY id`)
	if err != nil {
		t.Fatalf("reading the refusals: %v", err)
	}
	defer rows.Close()

	var out []storedRefusal

	for rows.Next() {
		var r storedRefusal

		if err := rows.Scan(&r.zoneID, &r.itemTime, &r.reason, &r.subject); err != nil {
			t.Fatalf("reading a refusal: %v", err)
		}

		if r.itemTime != nil {
			utc := r.itemTime.UTC()
			r.itemTime = &utc
		}

		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("reading the refusals: %v", err)
	}

	return out
}

func i32(v int32) *int32 { return &v }
func i16(v int16) *int16 { return &v }
func str(v string) *string {
	return &v
}

func TestTheZoneFeedWriteAgainstAMigratedCopy(t *testing.T) {
	pool, store := feedCopy(t)

	const (
		itemTime = "2025-06-01T12:00:00+0000"
		alpha    = `"id":831667,"name":"NotSaneNow","latitude":44.753365,"longitude":-92.801862,` +
			`"dateCreated":"2025-06-01T12:00:00+0000","totalTakeovers":0,"takeoverPoints":185,"pointsPerHour":1,` +
			`"region":{"area":{"name":"Pierce County","id":4720},"country":"us","name":"Wisconsin","id":412}`
	)

	// A live item's values, dated well before any instant this test can run at:
	// the poll's start is the server's clock, and a dateCreated after it is
	// refused by the assertion it is held to.
	itemAt := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	alphaZone := zonesync.Zone{
		ID: 831667, Name: "NotSaneNow", Latitude: 44.753365, Longitude: -92.801862, DateCreated: itemAt,
		TotalTakeovers: 0, TakeoverPoints: 185, PointsPerHour: 1, RegionID: 412, RegionName: "Wisconsin",
		RegionCountry: str("us"), AreaID: i32(4720), AreaName: str("Pierce County"),
	}

	t.Run("AC3: a complete item for an id the copy lacks is stored with its values", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		applied := poll(t, store, feedItem(itemTime, alpha))

		got, ok := storedZone(t, pool, 831667)
		if !ok {
			t.Fatal("a complete feed item for an id the copy lacked left no row")
		}

		if !reflect.DeepEqual(got, alphaZone) {
			t.Errorf("the copy holds\n%+v\nwant the item's values\n%+v", got, alphaZone)
		}

		if applied.Inserted != 1 || applied.Held != 0 {
			t.Errorf("the write reports %d inserted and %d held, want 1 and 0", applied.Inserted, applied.Held)
		}
	})

	// Insert-only (#238): a feed item is the zone as it was created, so the
	// copy's values for an id it holds — coordinates included — may be newer,
	// and the feed leaves them as they are.
	t.Run("AC3: a complete item for a held id with other values leaves the held row unchanged", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		hold(t, pool, fixtureZone{id: 831667, created: time.Date(2019, 5, 6, 7, 8, 9, 0, time.UTC), typeID: i16(3)})

		before, _ := storedZone(t, pool, 831667)
		if reflect.DeepEqual(before, alphaZone) {
			t.Fatal("the held fixture already carries the item's values, so this case could not tell an update from none")
		}

		var changedBefore, changedAfter time.Time

		if err := pool.QueryRow(t.Context(), `SELECT last_changed_at FROM zone WHERE id = 831667`).Scan(&changedBefore); err != nil {
			t.Fatalf("reading the held zone's last_changed_at: %v", err)
		}

		applied := poll(t, store, feedItem(itemTime, alpha))

		if after, _ := storedZone(t, pool, 831667); !reflect.DeepEqual(after, before) {
			t.Errorf("a held zone fed an item with other values changed from\n%+v\nto\n%+v\nwant it unchanged", before, after)
		}

		if err := pool.QueryRow(t.Context(), `SELECT last_changed_at FROM zone WHERE id = 831667`).Scan(&changedAfter); err != nil {
			t.Fatalf("reading the held zone's last_changed_at after the feed: %v", err)
		}

		if !changedAfter.Equal(changedBefore) {
			t.Errorf("the held zone's last_changed_at moved from %s to %s, want it unchanged", changedBefore, changedAfter)
		}

		if applied.Inserted != 0 || applied.Held != 1 {
			t.Errorf("the write reports %d inserted and %d held, want 0 and 1", applied.Inserted, applied.Held)
		}
	})

	t.Run("AC4: an item lacking a required field changes no row and is recorded", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		held := fixtureZone{id: 831667, created: time.Date(2019, 5, 6, 7, 8, 9, 0, time.UTC)}
		hold(t, pool, held)

		before, _ := storedZone(t, pool, held.id)

		noLatitude := strings.Replace(alpha, `"latitude":44.753365,`, "", 1)
		orphan := strings.Replace(strings.Replace(alpha, `"id":831667,`, `"id":900001,`, 1), `"name":"NotSaneNow",`, "", 1)

		poll(t, store, feedItem(itemTime, noLatitude), feedItem(itemTime, orphan))

		if after, _ := storedZone(t, pool, held.id); !reflect.DeepEqual(after, before) {
			t.Errorf("a held zone fed an item lacking its latitude changed from\n%+v\nto\n%+v", before, after)
		}

		if _, ok := storedZone(t, pool, 900001); ok {
			t.Error("an item lacking its name wrote a row for zone 900001")
		}

		want := []storedRefusal{
			{zoneID: i32(831667), itemTime: &itemAt, reason: "missing_field", subject: "zone.latitude"},
			{zoneID: i32(900001), itemTime: &itemAt, reason: "missing_field", subject: "zone.name"},
		}

		if got := storedRefusals(t, pool); !reflect.DeepEqual(got, want) {
			t.Errorf("the refusals recorded are %s, want %s", refusalsText(got), refusalsText(want))
		}
	})

	t.Run("AC4: an item an assertion refuses writes no row and is recorded by the assertion", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		poll(t, store, feedItem(itemTime, strings.Replace(alpha, `"latitude":44.753365`, `"latitude":95`, 1)))

		if _, ok := storedZone(t, pool, 831667); ok {
			t.Error("an item whose latitude is outside its range wrote a row")
		}

		want := []storedRefusal{{zoneID: i32(831667), itemTime: &itemAt, reason: "assertion", subject: zonesync.AssertLatitudeRange}}

		if got := storedRefusals(t, pool); !reflect.DeepEqual(got, want) {
			t.Errorf("the refusals recorded are %s, want %s", refusalsText(got), refusalsText(want))
		}
	})

	t.Run("AC4: an item lacking only optional fields is stored and nothing recorded", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		bare := strings.Replace(alpha, `{"area":{"name":"Pierce County","id":4720},"country":"us",`, `{`, 1)

		poll(t, store, feedItem(itemTime, bare))

		got, ok := storedZone(t, pool, 831667)
		if !ok || got.RegionCountry != nil || got.AreaID != nil || got.TypeID != nil {
			t.Errorf("an item with no type, country or area was stored as %+v (stored: %v), want it held with those three absent", got, ok)
		}

		if refusals := storedRefusals(t, pool); len(refusals) != 0 {
			t.Errorf("an item lacking only optional fields was recorded: %s", refusalsText(refusals))
		}
	})

	t.Run("AC2: the latest feed time is recorded, read back, and never moved back", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		if latest, _, err := store.FeedCursor(t.Context()); err != nil || latest != nil {
			t.Fatalf("with no feed time recorded FeedCursor gave %v, %v; want none", latest, err)
		}

		poll(t, store, feedItem(itemTime, alpha))

		// A second store over a second pool: what a restarted process reads.
		second, err := pgxpool.New(t.Context(), os.Getenv(envMigratedCopy))
		if err != nil {
			t.Fatalf("opening a second pool: %v", err)
		}

		t.Cleanup(second.Close)

		restarted, err := New(second, nil)
		if err != nil {
			t.Fatalf("building the second store: %v", err)
		}

		latest, _, err := restarted.FeedCursor(t.Context())
		if err != nil || latest == nil || !latest.Equal(itemAt) {
			t.Fatalf("after a batch whose newest item is %s, a new reader got %v, %v", itemAt, latest, err)
		}

		earlier := itemAt.Add(-time.Hour)

		if _, err := store.ApplyFeed(t.Context(), zonesync.FeedBatch{Latest: &earlier}); err != nil {
			t.Fatalf("writing an earlier latest feed time: %v", err)
		}

		if latest, _, _ := store.FeedCursor(t.Context()); latest == nil || !latest.Equal(itemAt) {
			t.Errorf("an earlier feed time moved the recorded one back to %v, want it left at %s", latest, itemAt)
		}
	})

	t.Run("nothing is written while the sync holds its key", func(t *testing.T) {
		feedEmptiedAfter(t, pool)

		release, acquired, err := store.Acquire(t.Context())
		if err != nil || !acquired {
			t.Fatalf("taking the sync's lock: acquired %v, %v", acquired, err)
		}

		_, now, err := store.FeedCursor(t.Context())
		if err != nil {
			t.Fatalf("reading the server's instant: %v", err)
		}

		batch, err := zonesync.ParseZoneFeed([]byte("["+feedItem(itemTime, alpha)+"]"), now)
		if err != nil {
			t.Fatalf("refused: %v", err)
		}

		_, err = store.ApplyFeed(t.Context(), batch)

		release(context.Background())

		if !errors.Is(err, zonesync.ErrSyncLockHeld) {
			t.Errorf("a feed batch written while the sync held its key gave %v, want ErrSyncLockHeld", err)
		}

		if _, ok := storedZone(t, pool, 831667); ok {
			t.Error("a feed batch wrote a zone while the sync held its key")
		}

		if latest, _, _ := store.FeedCursor(t.Context()); latest != nil {
			t.Errorf("a feed batch refused under the sync's key recorded the latest feed time %s", latest)
		}
	})
}

func refusalsText(rs []storedRefusal) string {
	parts := make([]string, 0, len(rs))

	for _, r := range rs {
		id, when := "none", "none"
		if r.zoneID != nil {
			id = fmt.Sprint(*r.zoneID)
		}

		if r.itemTime != nil {
			when = r.itemTime.Format(time.RFC3339)
		}

		parts = append(parts, fmt.Sprintf("{zone %s, time %s, %s, %s}", id, when, r.reason, r.subject))
	}

	return "[" + strings.Join(parts, " ") + "]"
}
