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
//
// ---------------------------------------------------------------------------
// NEVER EXECUTED: NOT ONE SQL STATEMENT BELOW HAS RUN.
//
// Nothing in this file has been sent to a PostgreSQL server, no method on this
// type is called anywhere in the suite, and no database was provisioned to
// author it — so the transaction shape, the staging load, the assertions, the
// advisory lock, the conflict target and the counts written to `sync_run` are
// all argued and none of them is measured.
//
// THE MARKER COVERS THE SQL HALF ONLY, AND IT CLAIMED MORE THAN THAT UNTIL
// columns_test.go LANDED BESIDE IT. It also read "and this package carries no
// tests", with "there is no test file in this directory" beneath it — two
// sentences this package's own test file falsified while leaving every line
// above them true. What that file measures is decided inside this process
// before any statement is sent: which value `zoneColumns` binds to which column
// name, and the text buildMergeSQL derives from it. It opens no pool, fakes
// none, and sends nothing, so it reaches none of what is unverified here.
//
// READ THAT AGAINST `internal/zonestore`, WHICH STILL CARRIES BOTH HALVES. Its
// marker says it has no test file, and that is still true, so the two no longer
// read alike — deliberately, and this line is where that is said rather than
// left to be noticed. A marker outliving the condition it describes is worse
// than no marker: a reader cannot tell one that was measured from one that was
// merely never revisited, and for one cycle this package was the counter-example
// that made the other unreadable.
//
// THE MARKER IS HERE BECAUSE THE PROSE ABOVE AND BELOW READS EXACTLY LIKE THE
// PROSE IN THE PACKAGES WHOSE TESTS RUN. Every comment in this file states what
// the code does in the same settled voice as `internal/zonesync`, whose claims
// are held up by assertions that go red when they stop being true. Nothing
// distinguishes the two on the page, which is the whole failure this line
// closes: a reader has no way to tell an implemented claim from a verified one,
// and will reasonably assume the stronger of the two. It is the Go half of the
// condition `migrations/0001_zone_store.sql` states under NEVER EXECUTED, whose
// DDL these statements are written against; that file is the argument in full
// for why neither half can be exercised here.
//
// A FAKE OVER `pgxpool` WOULD NOT DISCHARGE THIS, and is refused rather than
// owed. What is unverified is what PostgreSQL does with these statements —
// isolation, the conflict target, the lock's scope, the command tags — and a
// fake answers every one of those from whatever the author believed while
// writing it. Such a test would report green over precisely the beliefs that
// are in doubt, and it would replace an honest marker with a false one.
//
// WHAT DISCHARGES IT IS A DATABASE. `Architecture.md § What is unproven` item 1
// names the evidence owed against this schema; the run that produces it is also
// the first run that can put a test under the statements below. Tests over what
// this process decides before they are sent need no database and are already
// here, which is exactly the line the narrowing above draws.
// ---------------------------------------------------------------------------
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
// Swap two columns between them and the compiler accepts it, pgx accepts it,
// PostgreSQL accepts it, and the first thing to notice is a zone whose area name
// is its region name. Which crossings the compiler accepts, and what these
// columns' Go types do and do not decide about that, is stated in the header of
// `columns_test.go` and is not restated here. Naming the column beside its
// accessor makes that swap unwriteable; deriving the other five lists from this
// one slice makes them unable to disagree at all.
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
	// beginRunSQL dates the run by the SERVER and hands that instant back.
	//
	// `now()` rather than a parameter because started_at is the rate limit's
	// gate and sinceLastAttemptSQL below subtracts it from the server's own
	// clock. A row dated by whichever host ran the sync would put one host's
	// clock on one side of that subtraction and the server's on the other, and a
	// host running fast would decide the interval was up before it was. It is
	// returned because the run is dated by it everywhere else too — inspectSQL
	// bounds date_created by it.
	beginRunSQL = `
INSERT INTO sync_run (started_at, outcome)
VALUES (now(), $1)
RETURNING id, started_at`

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

	// sinceLastAttemptSQL is the rate limit's gate. It is max(started_at) over
	// every outcome, including `running`: a run in flight has already spent the
	// request, and a run that died holds the allowance until the interval
	// elapses just as a completed one does.
	//
	// The subtraction is done here so that both of its operands are the
	// server's — see beginRunSQL. Epoch seconds rather than an interval because
	// a float8 becomes a Duration without a driver-specific interval type in
	// between, and the gate is measured in minutes, where float64 seconds have
	// precision to spare. An empty table gives one row carrying NULL, not no
	// rows.
	sinceLastAttemptSQL = `SELECT extract(epoch FROM now() - max(started_at)) FROM sync_run`

	// serverNowSQL reads the instant the merge stamps and the run records.
	//
	// IT IS WHEN THE MERGE BEGAN AND NOT WHEN IT COMMITTED. now() is
	// transaction_timestamp(), fixed for the whole of the transaction it is
	// read in, so what completed_at carries is the start of the merge and every
	// second the merge then takes is on the far side of it. That is what the
	// column is worth to a reader and it is a bound rather than a reading: the
	// merge committed at some instant after this one.
	//
	// It is taken here anyway, and being constant is the reason. This value is
	// stamped on every row the merge touches AND recorded as the run's
	// completed_at, and a commit instant is knowable to neither — the rows are
	// written before it exists and the run's own writer is a second transaction.
	// One instant both can carry is worth more here than a closer one only one
	// of them could.
	serverNowSQL = `SELECT now()`

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

// SinceLastAttempt returns how long ago the most recent run of any outcome
// started, measured by the server.
func (s *Store) SinceLastAttempt(ctx context.Context) (time.Duration, bool, error) {
	// The subtraction over an empty table is one row carrying NULL, not no rows,
	// so the scan target is nullable and the emptiness is read off it.
	var seconds *float64

	if err := s.pool.QueryRow(ctx, sinceLastAttemptSQL).Scan(&seconds); err != nil {
		return 0, false, fmt.Errorf("reading how long ago the last sync attempt was: %w", err)
	}

	if seconds == nil {
		return 0, false, nil
	}

	return time.Duration(*seconds * float64(time.Second)), true, nil
}

// BeginRun writes the first of the run's two rows-worth of bookkeeping, dated by
// the server, and returns the date it was given.
func (s *Store) BeginRun(ctx context.Context) (int64, time.Time, error) {
	var (
		id        int64
		startedAt time.Time
	)

	if err := s.pool.QueryRow(ctx, beginRunSQL, string(zonesync.OutcomeRunning)).Scan(&id, &startedAt); err != nil {
		return 0, time.Time{}, fmt.Errorf("recording the start of a sync run: %w", err)
	}

	return id, startedAt.UTC(), nil
}

// FinishRun writes the second, whatever end the run reached.
func (s *Store) FinishRun(ctx context.Context, id int64, result zonesync.Result) error {
	tag, err := s.pool.Exec(ctx, finishRunSQL,
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

	// AN UPDATE THAT MATCHED NOTHING IS NOT AN ERROR THE DRIVER REPORTS. The
	// statement was valid and the WHERE found no row, which is success as far as
	// PostgreSQL and pgx are concerned — so with the command tag thrown away, a
	// row that was never inserted, or was deleted under this run, is
	// indistinguishable from one updated exactly as intended. `sync_run` is the
	// only durable record this worker keeps, and this is the write that closes
	// it: reporting success for having written nowhere is the one failure it
	// cannot afford.
	if n := tag.RowsAffected(); n != 1 {
		return fmt.Errorf("recording the end of sync run %d updated %d rows and not 1: the row this run inserted is not there to finish, so the outcome %q is recorded nowhere",
			id, n, result.Outcome)
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
// splitting it into batches takes that guarantee away.
//
// WHAT MAKES THE TWO COUNTS COMPARABLE IS THE ADVISORY LOCK AND NOT THIS
// TRANSACTION. The default isolation level is READ COMMITTED, under which every
// statement takes its own snapshot: `before` and `after` are therefore read from
// two different views of the table, and a writer committing between them makes
// their difference describe its work as well as this merge's. That difference is
// subtracted from the rows the merge affected, so it can drive Updated below
// zero and into `sync_run.rows_updated`, a column carrying no CHECK to refuse
// it. This comment previously credited the transaction with a guarantee only
// REPEATABLE READ or stricter would have given it.
//
// The lock is what actually holds. Acquire takes the sync's advisory key at
// session level for the whole run, nothing else in this service writes `zone`,
// and `Architecture.md § Migrating against a running sync` requires any
// migration touching it to take that same key — so this run is the only writer
// while the merge is open.
//
// That is an argument rather than an enforcement, which is why the figures are
// tested below BEFORE anything is committed. A count that cannot be true is
// evidence the premise failed, and a run that cannot say what it did does not
// leave it behind: it costs one interval of freshness, which is what every other
// failure in this package costs, against an audit row asserting something
// impossible.
func (s *Store) Merge(ctx context.Context) (zonesync.Merged, error) {
	var merged zonesync.Merged

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return merged, fmt.Errorf("opening the merge transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	// The instant this merge stamps every row it touches with, and the instant
	// the run records as completed_at. Read here rather than taken from the
	// process clock, for the reason beginRunSQL gives about started_at: a run
	// whose two instants come from different clocks can record an end before its
	// own beginning, and currency is read off this one.
	var completedAt time.Time

	if err := tx.QueryRow(ctx, serverNowSQL).Scan(&completedAt); err != nil {
		return zonesync.Merged{}, fmt.Errorf("reading the instant to stamp the merge with: %w", err)
	}

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

	// Inserted and updated are separated by the change in the table's size
	// rather than by inspecting each affected row. The alternative is a
	// RETURNING clause testing the system column xmax, which reads whether a row
	// version was superseded — an internal detail this file would then depend
	// on to report a statistic.
	inserted := after - before
	updated := int(tag.RowsAffected()) - inserted

	// Neither can be negative while this run is the only writer, so either one
	// being negative says it was not. See the note on the lock above.
	if inserted < 0 || updated < 0 {
		return zonesync.Merged{}, fmt.Errorf(
			"the merge is not committing: it affected %d rows while `zone` went from %d to %d, giving %d inserted and %d updated — figures that cannot both be true unless something else wrote `zone` while the sync held its lock, and what this run merged is therefore not known",
			tag.RowsAffected(), before, after, inserted, updated)
	}

	if err := tx.Commit(ctx); err != nil {
		return zonesync.Merged{}, fmt.Errorf("committing the merge: %w", err)
	}

	merged.Inserted = inserted
	merged.Updated = updated
	merged.CompletedAt = completedAt.UTC()

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
//
// A FALSE ANSWER IS A DIFFERENT THING AND IS NOT TREATED AS THAT ONE.
// pg_advisory_unlock returns false when the session did not hold the lock at
// all, which the statement here could not once report because it was run for
// its effect with its answer discarded. A session that holds nothing cannot
// hand anything to the next borrower, so the connection is safe to return —
// what is not safe is the silence, because it means this run believed it held
// the lock for its whole duration and did not.
func (s *Store) release(ctx context.Context, conn *pgxpool.Conn) {
	var unlocked bool

	if err := conn.QueryRow(ctx, "SELECT pg_advisory_unlock($1)", syncAdvisoryLockKey).Scan(&unlocked); err != nil {
		s.log.Error("the sync lock could not be released, so its connection is being discarded", "error", err)

		if hijacked := conn.Hijack(); hijacked != nil {
			_ = hijacked.Close(ctx)
		}

		return
	}

	if !unlocked {
		s.log.Error("the sync lock was not held by the session asked to release it, so this run did not hold what it believed it held",
			"key", syncAdvisoryLockKey)
	}

	conn.Release()
}
