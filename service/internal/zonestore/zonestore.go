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
// So the two halves are two packages. This one answers questions; `zonesyncstore`
// writes. A handler may import this and could not usefully import that.
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

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to the store: %w", err)
	}

	return pool, nil
}
