package syncstore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// The zone feed's half of the write side (#232): the adapter behind
// zonesync.FeedStore. Its tables are `migrations/0002_zone_feed.sql`'s; what
// it writes into `zone` goes through the INSERT … SELECT mergeSQL begins with,
// built from the same zoneColumns, so the feed binds every value to the column
// the all-zones path binds it to. It never updates a held id: see
// feedInsertSQL.

var _ zonesync.FeedStore = (*Store)(nil)

// feedStagingTable is where one feed batch is staged: a temporary table of
// zone_incoming's shape, private to the session and dropped at commit.
//
// NOT zone_incoming ITSELF. A full refresh refused by its staging assertions
// leaves the refused response staged there, which is the one place it can be
// inspected; a feed poll twenty minutes later would overwrite it with a batch
// of a hundred zones.
const feedStagingTable = "pg_temp.zone_feed_incoming"

// feedInsertSQL is mergeSQL's INSERT … SELECT reading from feedStagingTable,
// and it leaves an id `zone` already holds untouched: ON CONFLICT DO NOTHING
// where mergeSQL updates. columns_test.go asserts both halves.
//
// INSERT-ONLY BECAUSE A FEED ITEM IS A CREATION SNAPSHOT. Read live on
// 20 October 2026, zone 830403's feed item carried 0 takeovers and the
// placeholder dateLastTaken while `POST /v5/zones` already gave 1 takeover
// (#238, the safety review's S1). Updating a held id from such an item would
// write older values over newer ones — coordinates included, reverting a move
// a full sync recorded until the next one. The Owner's decision on #20 takes
// *new* zones from the feed; a held zone's values come from the monthly full
// sync alone. This supersedes `RD-041`.
var feedInsertSQL = buildInsertSQL(feedStagingTable) + "ON CONFLICT (" + zoneKeyColumn + ") DO NOTHING"

// refusalColumns are the columns of `zone_feed_refusal` a refusal fills; `id`
// is the sequence's.
var refusalColumns = []string{"recorded_at", "zone_id", "item_time", "reason", "subject"}

const (
	// feedCursorSQL reads the latest feed time and the server's instant in one
	// round trip. No row is no time recorded, and the scalar subquery reads it
	// as NULL.
	feedCursorSQL = `SELECT (SELECT latest_item_time FROM zone_feed_state), now()`

	// feedLockSQL takes the sync's own key for the length of the feed's
	// transaction, without waiting. See ApplyFeed.
	feedLockSQL = `SELECT pg_try_advisory_xact_lock($1)`

	feedStageSQL = `CREATE TEMP TABLE zone_feed_incoming (LIKE zone_incoming) ON COMMIT DROP`

	// advanceFeedSQL moves the latest feed time forward and never back: the
	// WHERE leaves a recorded time that is already later where it is, so two
	// polls that overlap cannot regress it.
	advanceFeedSQL = `
INSERT INTO zone_feed_state (singleton, latest_item_time)
VALUES (true, $1)
ON CONFLICT (singleton) DO UPDATE
SET    latest_item_time = excluded.latest_item_time
WHERE  zone_feed_state.latest_item_time < excluded.latest_item_time`
)

// FeedCursor returns the latest feed time recorded, nil when none is, and the
// server's current instant.
func (s *Store) FeedCursor(ctx context.Context) (*time.Time, time.Time, error) {
	var (
		latest *time.Time
		now    time.Time
	)

	if err := s.pool.QueryRow(ctx, feedCursorSQL).Scan(&latest, &now); err != nil {
		return nil, time.Time{}, fmt.Errorf("reading the latest zone-feed time: %w", err)
	}

	if latest != nil {
		utc := latest.UTC()
		latest = &utc
	}

	return latest, now.UTC(), nil
}

// ApplyFeed writes one feed batch in ONE transaction: the zones, the refusals
// and the latest feed time commit together or not at all, so the time is never
// recorded past an item whose zone or refusal was not.
//
// IT TAKES THE SYNC'S ADVISORY KEY, AND ONLY FOR THE TRANSACTION. The key is
// what `Architecture.md § Migrating against a running sync` has every writer of
// `zone` serialise on, and Merge's inserted and updated counts are sound only
// while one writer holds it. Taken at transaction level and without waiting, it
// is held for the milliseconds of this write rather than across the feed's
// request: a full refresh in progress holds it for its whole run, and a poll
// that meets it writes nothing and returns zonesync.ErrSyncLockHeld.
func (s *Store) ApplyFeed(ctx context.Context, batch zonesync.FeedBatch) (zonesync.FeedApplied, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("opening the zone-feed transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var locked bool

	if err := tx.QueryRow(ctx, feedLockSQL, syncAdvisoryLockKey).Scan(&locked); err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("taking the sync lock for the zone feed: %w", err)
	}

	if !locked {
		return zonesync.FeedApplied{}, zonesync.ErrSyncLockHeld
	}

	// One instant for every row this batch touches and every refusal it
	// records, read from the server for the reason Merge reads its own.
	var stamp time.Time

	if err := tx.QueryRow(ctx, serverNowSQL).Scan(&stamp); err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("reading the instant to stamp the zone feed with: %w", err)
	}

	var applied zonesync.FeedApplied

	if len(batch.Zones) > 0 {
		if applied, err = insertFeedZones(ctx, tx, batch.Zones, stamp); err != nil {
			return zonesync.FeedApplied{}, err
		}
	}

	if len(batch.Refusals) > 0 {
		copied, err := tx.CopyFrom(ctx, pgx.Identifier{"zone_feed_refusal"}, refusalColumns,
			pgx.CopyFromSlice(len(batch.Refusals), func(i int) ([]any, error) {
				r := batch.Refusals[i]

				return []any{stamp, r.ZoneID, r.ItemTime, string(r.Reason), r.Subject}, nil
			}))
		if err != nil {
			return zonesync.FeedApplied{}, fmt.Errorf("recording the refused zone-feed items: %w", err)
		}

		if copied != int64(len(batch.Refusals)) {
			return zonesync.FeedApplied{}, fmt.Errorf("recorded %d of %d refused zone-feed items", copied, len(batch.Refusals))
		}
	}

	if batch.Latest != nil {
		if _, err := tx.Exec(ctx, advanceFeedSQL, *batch.Latest); err != nil {
			return zonesync.FeedApplied{}, fmt.Errorf("recording the latest zone-feed time: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("committing the zone-feed batch: %w", err)
	}

	return applied, nil
}

// insertFeedZones stages the batch's zones and inserts those whose id `zone`
// does not hold; a held id is left as it is (feedInsertSQL).
//
// The batch holds each id once — zonesync.ParseZoneFeed keeps one item per id,
// the later feed time winning — so which of two items for one new id is stored
// is decided there, not by the order the rows reach this statement.
//
// ON CONFLICT DO NOTHING makes the command tag count the rows inserted and
// nothing else, so every staged zone it did not insert is one the copy held.
func insertFeedZones(ctx context.Context, tx pgx.Tx, zones []zonesync.Zone, stamp time.Time) (zonesync.FeedApplied, error) {
	if _, err := tx.Exec(ctx, feedStageSQL); err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("creating the zone-feed staging table: %w", err)
	}

	copied, err := tx.CopyFrom(ctx, pgx.Identifier{"pg_temp", "zone_feed_incoming"}, stagingColumns,
		pgx.CopyFromSlice(len(zones), func(i int) ([]any, error) {
			return stagingValues(zones[i]), nil
		}))
	if err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("staging the zone-feed zones: %w", err)
	}

	if copied != int64(len(zones)) {
		return zonesync.FeedApplied{}, fmt.Errorf("staged %d of %d zone-feed zones", copied, len(zones))
	}

	tag, err := tx.Exec(ctx, feedInsertSQL, stamp)
	if err != nil {
		return zonesync.FeedApplied{}, fmt.Errorf("inserting the zone-feed zones: %w", err)
	}

	inserted := int(tag.RowsAffected())

	return zonesync.FeedApplied{Inserted: inserted, Held: len(zones) - inserted}, nil
}
