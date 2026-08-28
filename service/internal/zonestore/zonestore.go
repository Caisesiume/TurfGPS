// Package zonestore is the read side of the synced zone store — the half a
// request may reach.
//
// IT IS A SEPARATE PACKAGE FROM THE SYNC'S WRITE PATH, and the split is the
// point rather than a filing convention. `FR-022` AC2 obliges a planning request
// to trigger no refresh and wait on none, and the invariant that enforces it is
// checked over the import graph: no package reachable from the request surface
// may reach `internal/zonesync`. A single store package holding both halves
// would put the sync worker in the transitive imports of every handler that ever
// reads a zone, and the check would have to be weakened until it said nothing.
//
// So the two halves are two packages. This one answers questions; `syncstore`
// writes. A handler may import this and could not usefully import that.
//
// ---------------------------------------------------------------------------
// NEVER EXECUTED, AND THIS PACKAGE CARRIES NO TESTS.
//
// The currency query below has not been sent to a PostgreSQL server, the pool
// this file opens has not opened, and there is no test file in this directory.
// `internal/syncstore` carries the write half of the same condition and the
// argument in full — why a fake over `pgxpool` is refused rather than owed, and
// what discharges the marker. This comment does not restate it.
//
// ONE UNVERIFIED CLAIM HERE IS MORE EXPOSED THAN THE REST, because it is the
// one a request reaches. `Currency` reads pgx.ErrNoRows as "no run has ever
// succeeded" and returns it as an answer rather than a failure — which is
// right, and is byte-for-byte the same answer the query would give while
// resolving against a `sync_run` this service has never written to.
// `poolSearchPath` below is what is supposed to keep those two apart, and it
// has never resolved anything. A copy reported as never-synced for a reason
// nobody can see is the asymmetry `Architecture.md § Absence is recorded and
// never acted on` weighs, and neither branch has been observed here.
// ---------------------------------------------------------------------------
package zonestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// currencySQL reads the completion instant of the latest successful run.
//
// IT ASKS `sync_run` WHAT HAPPENED AND NEVER WHAT WAS SCHEDULED. A schedule that
// fired is not evidence that a copy was refreshed — the run behind it may have
// been refused by the endpoint, rejected by the staging assertions, or killed
// before it merged — so currency derived from a timer would report a fresh copy
// during exactly the outage that made it stale. `completed_at` is written only
// by a merge that committed, and only `ok` carries one.
const currencySQL = `
SELECT completed_at
FROM   sync_run
WHERE  outcome = 'ok'
  AND  completed_at IS NOT NULL
ORDER  BY completed_at DESC
LIMIT  1`

// poolSearchPath is the schema resolution order every connection from this
// pool carries.
//
// IT IS THE POOL HALF OF A PINNING THE MIGRATION ONLY MADE FOR ITSELF.
// `migrations/0001_zone_store.sql` decision 5 is the argument in full — why an
// unpinned path is the CVE-2018-1058 class, why it is silent by construction,
// and why the three entries are in this order — and this constant is not a
// second statement of it. What that decision could not reach is the session
// that later READS what it built: it pinned the path the tables were created
// in and left the queries below resolving through whatever path the connecting
// role happened to carry, which is exactly the two questions it names, asked of
// tables this migration never built.
//
// The value matches the migration's byte for byte, ordering included. A second
// choice of path would be the same as no pinning at all.
const poolSearchPath = "pg_catalog, public, pg_temp"

// minPoolConns is the floor this service holds its connection pool to.
//
// IT IS A FLOOR AND NOT A SETTING. An operator who sizes the pool up through
// the DSN's pool_max_conns keeps their number; one who sizes it below this, or
// leaves it unset, is raised to this. Unset is the case that matters, because
// pgx's own default is max(4, NumCPU) — a property of the host the process
// landed on, never a decision about this workload, and two on a small one.
//
// The floor exists because two connections are spoken for while a sync runs and
// neither is available to a request. `internal/syncstore` takes the sync's
// advisory lock at SESSION level, which pins the connection holding it for the
// whole run, and the run needs a second for whatever statement it is on. A pool
// of two therefore has nothing left at all for the half a request may reach —
// the currency read below — for the duration of every run.
//
// Eight is this package's operational choice in the sense `internal/config`
// uses for the figures it picks: no measurement of the request path exists to
// derive one from, because that path is one handler today. It leaves six while
// a run holds two, and the DSN is where an operator raises it.
const minPoolConns int32 = 8

// Currency is how current the local synced copy is.
//
// EverSucceeded is a field rather than a zero-value convention because "never
// synced" and "synced at the zero instant" must not be the same answer. A
// consumer that read a zero time as an age would compute a copy two thousand
// years stale for a store that has simply never run, and one that read it as
// "now" would report a fresh copy for a store that holds nothing at all.
//
// What is done with the answer — what bound is tolerated, what is disclosed to
// the user — is `FR-024`, `FR-033` and `FR-034`, and none of it is decided here.
// This type reports the observation and stops.
type Currency struct {
	// LastSuccessAt is when the latest successful run completed. It is
	// meaningful only when EverSucceeded is true.
	LastSuccessAt time.Time

	// EverSucceeded reports whether any run has ever merged.
	EverSucceeded bool
}

// Reader answers what the request path may ask about the synced copy.
type Reader struct {
	pool *pgxpool.Pool
}

// NewReader wraps a pool.
func NewReader(pool *pgxpool.Pool) (*Reader, error) {
	if pool == nil {
		return nil, errors.New("no connection pool")
	}

	return &Reader{pool: pool}, nil
}

// Currency reports how current the local copy is, from what the system observed.
//
// A store that has never completed a run is not an error: it is an answer, and a
// distinguishable one.
func (r *Reader) Currency(ctx context.Context) (Currency, error) {
	var completedAt time.Time

	err := r.pool.QueryRow(ctx, currencySQL).Scan(&completedAt)
	if err != nil {
		// No row is the answer "no run has ever succeeded", not a failure.
		if errors.Is(err, pgx.ErrNoRows) {
			return Currency{}, nil
		}

		return Currency{}, fmt.Errorf("reading the zone copy's currency: %w", err)
	}

	return Currency{LastSuccessAt: completedAt.UTC(), EverSucceeded: true}, nil
}

// Open connects to the store of `Architecture.md § D4`.
//
// It does not apply migrations and must never be made to. `migrations/README.md`
// records that application is a deliberate act by an authorised operator, never
// something this repository's tooling does and never something that happens at
// boot.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("the database URL could not be parsed: %w", err)
	}

	// Names this service in pg_stat_activity, so a lock or a long transaction
	// is attributable without guessing which client held it.
	cfg.ConnConfig.RuntimeParams["application_name"] = "turfgps"

	// See poolSearchPath. Set as a startup parameter rather than by a statement
	// on each checkout, so there is no window in which a connection has been
	// handed out and not yet pinned.
	cfg.ConnConfig.RuntimeParams["search_path"] = poolSearchPath

	// See minPoolConns. Compared rather than assigned, so a DSN that asked for
	// more keeps what it asked for.
	if cfg.MaxConns < minPoolConns {
		cfg.MaxConns = minPoolConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to the store: %w", err)
	}

	return pool, nil
}
