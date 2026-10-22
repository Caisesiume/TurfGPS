//go:build postgis

package zonestore

// #33 (`FR-024`, `FR-033`) against PostGIS: Held, the one input of the gate in
// front of planning, reports a copy held in every state a copy can be planned
// from, and none only where the copy holds no zone.
//
// Built only under the `postgis` tag, against the scratch database CI's
// `migrations` job has just migrated, and refusing any other, through
// openScratch in corridor_postgis_test.go; the reasons are that file's.
//
// NO STALENESS BOUND APPEARS HERE. The ages below are none recorded, now, and
// the Unix epoch: none is the bound of `Architecture.md § Retrieving zones`,
// none needs to be, and Held reads no age to compare one against. A copy as
// old as the bound lies between the second and the third.

import (
	"context"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/corridor"
)

// heldDeadline bounds each Held read below. A read that waited on the open
// refresh would wait for as long as that transaction stays open, which is the
// rest of the test, so overrunning this is how waiting shows up.
const heldDeadline = 5 * time.Second

var heldZone = corridor.Zone{
	ID: 9_000_201, Name: "held", Latitude: 59.33, Longitude: 18.06,
	DateCreated:    time.Date(2018, 5, 6, 0, 0, 0, 0, time.UTC),
	TotalTakeovers: 4, TakeoverPoints: 95, PointsPerHour: 1,
}

func TestAZoneCopyIsHeldWhateverItsAgeOrRefreshState(t *testing.T) {
	ctx := t.Context()
	pool := openScratch(t)

	reader, err := NewReader(pool)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	exec := func(stmt string, args ...any) {
		t.Helper()

		if _, err := pool.Exec(ctx, stmt, args...); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}

	held := func(state string, want bool) {
		t.Helper()

		readCtx, cancel := context.WithTimeout(ctx, heldDeadline)
		defer cancel()

		got, err := reader.Held(readCtx)
		if err != nil {
			t.Fatalf("%s: Held failed: %v; it must answer without waiting on a refresh", state, err)
		}

		if got != want {
			t.Errorf("%s: Held = %v, want %v", state, got, want)
		}
	}

	// AC4's Given: no local copy of the zone set.
	held("an empty copy with no run on record", false)

	// A first fill in flight is no copy yet: the row it writes is not
	// committed, and the read neither sees it nor waits on it.
	exec(`INSERT INTO sync_run (started_at, outcome) VALUES (now(), 'running')`)

	fill, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("beginning the first fill: %v", err)
	}

	defer func() { _ = fill.Rollback(context.WithoutCancel(ctx)) }()

	if _, err := fill.Exec(ctx, `
		INSERT INTO zone (id, name, latitude, longitude, date_created, total_takeovers,
		                  takeover_points, points_per_hour, region_id, region_name,
		                  first_seen_at, last_changed_at)
		VALUES (9000299, 'mid-fill', 59.33, 18.06, now(), 0, 65, 0, 141, 'Stockholm', now(), now())`); err != nil {
		t.Fatalf("writing the first fill's row: %v", err)
	}

	held("an empty copy whose first fill has not committed", false)

	if err := fill.Rollback(ctx); err != nil {
		t.Fatalf("ending the first fill: %v", err)
	}

	exec(`DELETE FROM sync_run`)

	// AC3's Given, with none recorded: zones held and no refresh of any kind on
	// record — no completed all-zones run, and no zone-feed run, which this
	// schema does not record at all yet.
	loadZones(t, pool, heldZone)

	held("a copy with no refresh on record", true)

	exec(`INSERT INTO sync_run (started_at, outcome) VALUES (now(), 'http_error')`)
	held("a copy whose only run failed", true)

	exec(`INSERT INTO sync_run (started_at, completed_at, outcome) VALUES ($1, $1, 'ok')`, time.Unix(0, 0))
	held("a copy last refreshed at the Unix epoch", true)

	exec(`INSERT INTO sync_run (started_at, completed_at, outcome) VALUES (now(), now(), 'ok')`)
	held("a copy refreshed just now", true)

	// AC1's Given: a refresh in progress, recorded as running, with a merge
	// open on `zone` that holds the lock the real one takes and a row lock on
	// every row it holds.
	exec(`INSERT INTO sync_run (started_at, outcome) VALUES (now(), 'running')`)

	merge, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("beginning the merge: %v", err)
	}

	defer func() { _ = merge.Rollback(context.WithoutCancel(ctx)) }()

	if _, err := merge.Exec(ctx, `LOCK TABLE zone IN ROW EXCLUSIVE MODE`); err != nil {
		t.Fatalf("taking the merge's lock: %v", err)
	}

	if _, err := merge.Exec(ctx, `UPDATE zone SET total_takeovers = total_takeovers + 1`); err != nil {
		t.Fatalf("locking every row as the merge would: %v", err)
	}

	held("a copy with a refresh in progress", true)
}
