// Package syncstore is the write side of the synced zone store: the PostGIS
// adapter behind the ports `internal/zonesync` declares.
//
// It is separate from `internal/zonestore` so that reading how current the copy
// is does not drag the sync worker into the import graph of everything that
// reads it — the boundary `FR-022` AC2 is checked over. See that package's doc.
//
// IT IS NAMED AWAY FROM `internal/zonestore` DELIBERATELY. The two were once
// `zonestore` and `zonesyncstore`, which differ by four letters in the middle of
// a word — on the one boundary this design calls load-bearing, where the whole
// question a reader is asking is which of the two they are looking at. A name
// that has to be read to the end to be told apart from its opposite is a name
// that will be misread on the import line.
//
// EVERY STATEMENT HERE IMPLEMENTS `Architecture.md § The sync write path` rather
// than inventing anything. Where this file departs from the shape written there
// it says so and why, at the statement.
package syncstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// syncAdvisoryLockKey is the well-known key `Architecture.md § Migrating against
// a running sync` requires the sync and any later migration to take, so that a
// migration serialises against a run instead of hoping the table is idle.
//
// THE VALUE IS ARBITRARY AND FIXED. Nothing derives it and nothing may change
// it: what it buys is that two parties agree, so a second choice of key is the
// same as no lock at all. Any migration that touches `zone` takes
// pg_advisory_lock on this number.
const syncAdvisoryLockKey int64 = 20260820

// zoneColumn is one column of the ingest set, bound to the value a staged zone
// supplies for it.
//
// THE BINDING ON ONE LINE IS THE POINT OF THIS TYPE. The fourteen columns were
// written out six times in this file — the staging column list, the COPY row,
// and the four lists inside the merge — and exactly one of those pairings had to
// agree by POSITION rather than by name: the staging list is what COPY is given,
// and the row builder supplied values in the order it hoped that list was in.
// Swap two columns of the same type between them and the compiler accepts it,
// pgx accepts it, PostgreSQL accepts it, and the first thing to notice is a zone
// whose area name is its region name. Naming the column beside its accessor
// makes that swap unwriteable; deriving the other five lists from this one slice
// makes them unable to disagree at all.
type zoneColumn struct {
	name  string
	value func(zonesync.Zone) any
}

const (
	// The two stamps the merge writes that no staged row supplies. They are
	// named because the merge treats them differently from each other and from
	// everything in zoneColumns: first_seen_at is written on insert and never
	// updated, last_changed_at on both.
	firstSeenColumn   = "first_seen_at"
	lastChangedColumn = "last_changed_at"

	// zoneKeyColumn is the merge's conflict target, and so the one ingest column
	// it neither assigns nor tests for change.
	zoneKeyColumn = "id"
)

// zoneColumns is `zone_incoming` in its declared order.
//
// THE ORDER IS THE MIGRATION'S AND MUST STAY THE MIGRATION'S. COPY is positional
// against the column list it is handed, that list is derived from this one, and
// so this slice is the single place the order is decided.
var zoneColumns = []zoneColumn{
	{"id", func(z zonesync.Zone) any { return z.ID }},
	{"name", func(z zonesync.Zone) any { return z.Name }},
	{"latitude", func(z zonesync.Zone) any { return z.Latitude }},
	{"longitude", func(z zonesync.Zone) any { return z.Longitude }},
	{"date_created", func(z zonesync.Zone) any { return z.DateCreated }},
	{"total_takeovers", func(z zonesync.Zone) any { return z.TotalTakeovers }},
	{"takeover_points", func(z zonesync.Zone) any { return z.TakeoverPoints }},
	{"points_per_hour", func(z zonesync.Zone) any { return z.PointsPerHour }},
	{"type_id", func(z zonesync.Zone) any { return z.TypeID }},
	{"region_id", func(z zonesync.Zone) any { return z.RegionID }},
	{"region_name", func(z zonesync.Zone) any { return z.RegionName }},
	{"region_country", func(z zonesync.Zone) any { return z.RegionCountry }},
	{"area_id", func(z zonesync.Zone) any { return z.AreaID }},
	{"area_name", func(z zonesync.Zone) any { return z.AreaName }},
}

// stagingColumns is the name half of zoneColumns, in the same order, which is
// the form pgx's CopyFrom takes.
var stagingColumns = func() []string {
	names := make([]string, len(zoneColumns))
	for i, c := range zoneColumns {
		names[i] = c.name
	}

	return names
}()

// stagingValues is the value half, for one zone, in that same order — the order
// being a fact of zoneColumns rather than of this function.
func stagingValues(z zonesync.Zone) []any {
	values := make([]any, len(zoneColumns))
	for i, c := range zoneColumns {
		values[i] = c.value(z)
	}

	return values
}

// mergeSQL is the statement of `Architecture.md § The sync write path`, written
// from zoneColumns.
//
// ONE DEPARTURE FROM THE FORM WRITTEN THERE, and it changes nothing the
// statement does: the select list names the staging columns instead of `i.*`.
// The expansion of `i.*` is positional, so the two column lists agreeing would
// be a property of the migration's declaration order rather than of anything
// visible here, and a column inserted into `zone_incoming` in the middle would
// silently shift fourteen values one place to the right.
//
// The four load-bearing properties are unchanged and are argued in that section:
// first_seen_at is absent from the SET list, every other field is refreshed
// including the ones that look constant, the change test is IS DISTINCT FROM
// rather than <> because four columns are nullable, and the WHERE is what keeps
// a sync writing a thousand rows instead of all of them.
var mergeSQL = buildMergeSQL()

func buildMergeSQL() string {
	width := len(lastChangedColumn)

	for _, c := range zoneColumns {
		if len(c.name) > width {
			width = len(c.name)
		}
	}

	names := make([]string, 0, len(zoneColumns))
	selected := make([]string, 0, len(zoneColumns))
	assignments := make([]string, 0, len(zoneColumns))
	changed := make([]string, 0, len(zoneColumns))

	for _, c := range zoneColumns {
		names = append(names, c.name)
		selected = append(selected, "i."+c.name)

		if c.name == zoneKeyColumn {
			continue
		}

		assignments = append(assignments, fmt.Sprintf("%-*s = excluded.%s", width, c.name, c.name))
		changed = append(changed, fmt.Sprintf("zone.%-*s IS DISTINCT FROM excluded.%s", width, c.name, c.name))
	}

	assignments = append(assignments,
		fmt.Sprintf("%-*s = excluded.%s", width, lastChangedColumn, lastChangedColumn))

	return "\n" +
		"INSERT INTO zone (" + strings.Join(names, ",\n                  ") + ",\n" +
		"                  " + firstSeenColumn + ", " + lastChangedColumn + ")\n" +
		"SELECT " + strings.Join(selected, ",\n       ") + ",\n" +
		"       $1, $1\n" +
		"FROM   zone_incoming i\n" +
		"ON CONFLICT (" + zoneKeyColumn + ") DO UPDATE SET\n" +
		"       " + strings.Join(assignments, ",\n       ") + "\n" +
		"WHERE  " + strings.Join(changed, "\n   OR  ")
}

const (
	beginRunSQL = `
INSERT INTO sync_run (started_at, outcome)
VALUES ($1, $2)
RETURNING id`

	finishRunSQL = `
UPDATE sync_run
SET    outcome        = $2,
       completed_at   = $3,
       http_status    = $4,
       response_bytes = $5,
       zones_received = $6,
       rows_inserted  = $7,
       rows_updated   = $8,
       rows_unchanged = $9,
       absent_count   = $10,
       absent_ids     = $11
WHERE  id = $1`

	// lastAttemptSQL is the rate limit's gate. It is max(started_at) over every
	// outcome, including `running`: a run in flight has already spent the
	// request, and a run that died holds the allowance until the interval
	// elapses just as a completed one does.
	lastAttemptSQL = `SELECT max(started_at) FROM sync_run`

	// inspectSQL counts everything the staging assertions of
	// `Architecture.md § The sync write path` decide on, in one round trip. The
	// deciding is `zonesync.Staged.Verify`, which is where the counts are
	// turned into a verdict; this statement only produces them.
	inspectSQL = `
SELECT (SELECT count(*) FROM zone_incoming),
       (SELECT count(DISTINCT id) FROM zone_incoming WHERE id IS NOT NULL),
       (SELECT count(*) FROM zone_incoming WHERE id IS NULL),
       (SELECT count(*) FROM zone_incoming
         WHERE latitude  IS NULL OR longitude IS NULL
            OR latitude  NOT BETWEEN  -90 AND  90
            OR longitude NOT BETWEEN -180 AND 180),
       (SELECT count(*) FROM zone_incoming
         WHERE date_created IS NULL OR date_created > $1),
       (SELECT count(*) FROM zone)`

	countZonesSQL = `SELECT count(*) FROM zone`

	// absentSQL is the ids held in `zone` and missing from the response.
	// Computed on every merged run, recorded on the run, and acted on by
	// nothing: `Architecture.md § Absence is recorded and never acted on`.
	absentSQL = `
SELECT coalesce(array_agg(z.id ORDER BY z.id), '{}'::integer[])
FROM   zone z
WHERE  NOT EXISTS (SELECT 1 FROM zone_incoming i WHERE i.id = z.id)`
)

// Store is the PostGIS adapter for the zone sync's write path. It satisfies
// both zonesync.Store and zonesync.Locker.
type Store struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// New wraps a pool.
func New(pool *pgxpool.Pool, log *slog.Logger) (*Store, error) {
	if pool == nil {
		return nil, errors.New("no connection pool")
	}

	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &Store{pool: pool, log: log}, nil
}

// LastAttempt returns the start instant of the most recent run of any outcome.
func (s *Store) LastAttempt(ctx context.Context) (time.Time, bool, error) {
	// max() over an empty table is one row carrying NULL, not no rows, so the
	// scan target is nullable and the emptiness is read off it.
	var startedAt *time.Time

	if err := s.pool.QueryRow(ctx, lastAttemptSQL).Scan(&startedAt); err != nil {
		return time.Time{}, false, fmt.Errorf("reading the last sync attempt: %w", err)
	}

	if startedAt == nil {
		return time.Time{}, false, nil
	}

	return startedAt.UTC(), true, nil
}

// BeginRun writes the first of the run's two rows-worth of bookkeeping.
func (s *Store) BeginRun(ctx context.Context, startedAt time.Time) (int64, error) {
	var id int64

	if err := s.pool.QueryRow(ctx, beginRunSQL, startedAt, string(zonesync.OutcomeRunning)).Scan(&id); err != nil {
		return 0, fmt.Errorf("recording the start of a sync run: %w", err)
	}

	return id, nil
}

// FinishRun writes the second, whatever end the run reached.
func (s *Store) FinishRun(ctx context.Context, id int64, result zonesync.Result) error {
	_, err := s.pool.Exec(ctx, finishRunSQL,
		id,
		string(result.Outcome),
		result.CompletedAt,
		result.HTTPStatus,
		result.ResponseBytes,
		result.ZonesReceived,
		result.RowsInserted,
		result.RowsUpdated,
		result.RowsUnchanged,
		result.AbsentCount,
		result.AbsentIDs,
	)
	if err != nil {
		return fmt.Errorf("recording the end of sync run %d: %w", id, err)
	}

	return nil
}

// Stage truncates the staging table and loads the response into it by binary
// COPY, which is what `zone_incoming` is UNLOGGED for.
func (s *Store) Stage(ctx context.Context, zones []zonesync.Zone) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("opening the staging transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "TRUNCATE TABLE zone_incoming"); err != nil {
		return fmt.Errorf("truncating the staging table: %w", err)
	}

	copied, err := tx.CopyFrom(ctx, pgx.Identifier{"zone_incoming"}, stagingColumns,
		pgx.CopyFromSlice(len(zones), func(i int) ([]any, error) {
			return stagingValues(zones[i]), nil
		}))
	if err != nil {
		return fmt.Errorf("staging the response: %w", err)
	}

	if copied != int64(len(zones)) {
		return fmt.Errorf("staged %d of %d zones", copied, len(zones))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing the staging transaction: %w", err)
	}

	return nil
}

// Inspect counts the staged rows for the assertions.
func (s *Store) Inspect(ctx context.Context, startedAt time.Time) (zonesync.Staged, error) {
	var staged zonesync.Staged

	err := s.pool.QueryRow(ctx, inspectSQL, startedAt).Scan(
		&staged.Rows,
		&staged.DistinctIDs,
		&staged.NullIDs,
		&staged.OutOfRange,
		&staged.FutureCreated,
		&staged.CurrentZoneRows,
	)
	if err != nil {
		return zonesync.Staged{}, fmt.Errorf("inspecting the staged response: %w", err)
	}

	return staged, nil
}

// Merge merges the staged rows into `zone`.
//
// ONE TRANSACTION, NEVER BATCHED. That is not tidiness: it is the whole answer
// to a request served mid-refresh, because under MVCC no reader can observe a
// partial merge, and `Architecture.md § The sync write path` states plainly that
// splitting it into batches takes that guarantee away. Everything below happens
// inside the one transaction, including the two counts the row figures are
// derived from, so those figures describe the state the merge actually committed
// rather than a table something else may have touched between statements.
func (s *Store) Merge(ctx context.Context, completedAt time.Time) (zonesync.Merged, error) {
	var merged zonesync.Merged

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return merged, fmt.Errorf("opening the merge transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var before int

	if err := tx.QueryRow(ctx, countZonesSQL).Scan(&before); err != nil {
		return merged, fmt.Errorf("counting the zones held: %w", err)
	}

	// Computed before the merge, which is when it can still be asked. It is
	// unchanged by the merge — nothing here deletes — but asking first keeps
	// the meaning obvious rather than resting on that.
	if err := tx.QueryRow(ctx, absentSQL).Scan(&merged.AbsentIDs); err != nil {
		return merged, fmt.Errorf("computing the absent ids: %w", err)
	}

	tag, err := tx.Exec(ctx, mergeSQL, completedAt)
	if err != nil {
		return zonesync.Merged{}, fmt.Errorf("merging the staged zones: %w", err)
	}

	var after int

	if err := tx.QueryRow(ctx, countZonesSQL).Scan(&after); err != nil {
		return zonesync.Merged{}, fmt.Errorf("counting the zones held after the merge: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return zonesync.Merged{}, fmt.Errorf("committing the merge: %w", err)
	}

	// Inserted and updated are separated by the change in the table's size
	// rather than by inspecting each affected row. The alternative is a
	// RETURNING clause testing the system column xmax, which reads whether a row
	// version was superseded — an internal detail this file would then depend
	// on to report a statistic.
	merged.Inserted = after - before
	merged.Updated = int(tag.RowsAffected()) - merged.Inserted

	return merged, nil
}

// Acquire takes the sync's advisory lock without waiting for it.
//
// The lock is session-level, so it is held on ONE connection for the run's
// duration and the connection is kept out of the pool until it is released.
func (s *Store) Acquire(ctx context.Context) (func(context.Context), bool, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("taking a connection for the sync lock: %w", err)
	}

	var locked bool

	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", syncAdvisoryLockKey).Scan(&locked); err != nil {
		conn.Release()

		return nil, false, fmt.Errorf("taking the sync lock: %w", err)
	}

	if !locked {
		conn.Release()

		return nil, false, nil
	}

	return func(releaseCtx context.Context) { s.release(releaseCtx, conn) }, true, nil
}

// release unlocks and returns the connection to the pool.
//
// A CONNECTION WHOSE UNLOCK FAILED IS DESTROYED RATHER THAN RETURNED. A session
// lock outlives the pool's handle on it, so putting that connection back would
// hand the next borrower a session still holding the sync's lock — and the next
// attempt, and every attempt after it, would read that as "another holder is
// running" and quietly stop syncing. Losing one connection is the cheaper
// failure by a wide margin.
func (s *Store) release(ctx context.Context, conn *pgxpool.Conn) {
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", syncAdvisoryLockKey); err != nil {
		s.log.Error("the sync lock could not be released, so its connection is being discarded", "error", err)

		if hijacked := conn.Hijack(); hijacked != nil {
			_ = hijacked.Close(ctx)
		}

		return
	}

	conn.Release()
}
