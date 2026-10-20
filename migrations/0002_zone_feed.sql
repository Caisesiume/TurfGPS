-- 0002_zone_feed.sql — the zone feed's own state (#232)
--
-- Creates `zone_feed_state`, which holds the latest feed time processed, and
-- `zone_feed_refusal`, the durable record of a feed item the copy could not
-- store. Neither touches `zone`: the ticket keeps both outside it, so the
-- column-set equality in part A of `0001_zone_store.verify.sql` is untouched,
-- and this file takes no lock on any table 0001 created.
--
-- The source is the Owner's zone-sync decision on #20
-- (`#issuecomment-6279336314`) and the draft records `FR-119` … `FR-122`, until
-- `Architecture.md § Retrieving zones` is amended to carry the zone feed. What
-- the feed writes into `zone` goes through the merge of
-- `Architecture.md § The sync write path`, unchanged; only its bookkeeping is
-- new, and it is here.
--
-- ----------------------------------------------------------------------------
-- EXECUTED ON THROWAWAY INSTANCES ONLY, as 0001 is: CI's `migrations` job
-- (`scripts/db/migrate-verify.sh`, #153) applies it after 0001, runs
-- `0002_zone_feed.verify.sql` part by part, and runs the rollback twice. It has
-- not been applied to any database that outlives the job.
-- ----------------------------------------------------------------------------
--
-- APPLY. By an authorised operator, deliberately, after 0001, never at boot:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0002_zone_feed.sql
--
-- The feed worker reads and writes both tables, so it fails on every poll —
-- logged, and never on a request — until this has been applied.
--
-- ROLLBACK. `0002_zone_feed.down.sql`. It drops the latest feed time, so the
-- next feed request after a re-apply carries no lower bound (`FR-120`'s second
-- criterion), and it drops every refusal recorded so far. Nothing in `zone` is
-- touched by it.
--
-- IRREVERSIBLE STEPS. None.
--
-- DECISIONS TAKEN HERE:
--
--   1. The latest feed time is ONE ROW, enforced. `zone_feed_state` is keyed on
--      a boolean that a CHECK holds to true, so a second row is refused by the
--      primary key rather than left for a reader to pick between. Absence of
--      the row is "no feed time recorded", which is the condition `FR-120`
--      sends no lower bound under.
--
--   2. A refusal names its item by what the item carried, and nothing it did
--      not. `zone_id` and `item_time` are nullable because an item can lack
--      either (`FR-122`: "the item's zone id where present"). `subject` names
--      the field that was missing or unreadable, or the assertion that refused
--      the value; one row per subject, so a query for one field is a WHERE
--      and not a string search. No raw item is stored: a feed item carries
--      `dateLastTaken`, and `FR-026` forbids that reaching anything the copy
--      keeps.
--
--   3. No index beyond the primary keys. The refusal table is read by an
--      operator, rarely, and written only for the unusual item; the state table
--      has one row.
--
--   4. `search_path` is pinned and every relation is schema-qualified, for the
--      reason decision 5 of `0001_zone_store.sql` gives.

BEGIN;

SET LOCAL search_path = pg_catalog, public, pg_temp;

-- --- zone_feed_state ----------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.zone_feed_state (
    singleton        boolean     PRIMARY KEY DEFAULT true,
    latest_item_time timestamptz NOT NULL,

    CONSTRAINT zone_feed_state_singleton CHECK (singleton)
);

-- --- zone_feed_refusal --------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.zone_feed_refusal (
    id          bigserial   PRIMARY KEY,
    recorded_at timestamptz NOT NULL,
    zone_id     integer,
    item_time   timestamptz,
    reason      text        NOT NULL,
    subject     text        NOT NULL,

    -- The vocabulary is `internal/zonesync`'s FeedRefusalReason, and the three
    -- values are the three ways `FR-122` names an item unstorable: a field the
    -- copy requires is absent, present in a form its column cannot hold, or
    -- carries a value an assertion refuses.
    CONSTRAINT zone_feed_refusal_reason_known CHECK (
        reason IN ('missing_field', 'unreadable_field', 'assertion')
    )
);

-- --- The divergence check ---------------------------------------------------
--
-- `CREATE TABLE IF NOT EXISTS` over a same-named table of another shape adds
-- nothing and says nothing, so the shape is asserted against the catalogue
-- before commit, in the direction 0001's decision 6 argues: expected ⊆ actual.

DO $$
DECLARE
    divergences text;
BEGIN
    WITH expected(tbl, col, typ, not_null) AS (
        VALUES
            ('zone_feed_state'::text, 'singleton'::text,        'boolean'::text,            true),
            ('zone_feed_state',       'latest_item_time',       'timestamp with time zone', true),

            ('zone_feed_refusal',     'id',                     'bigint',                   true),
            ('zone_feed_refusal',     'recorded_at',            'timestamp with time zone', true),
            ('zone_feed_refusal',     'zone_id',                'integer',                  false),
            ('zone_feed_refusal',     'item_time',              'timestamp with time zone', false),
            ('zone_feed_refusal',     'reason',                 'text',                     true),
            ('zone_feed_refusal',     'subject',                'text',                     true)
    ),
    actual(tbl, col, typ, not_null) AS (
        SELECT c.relname::text,
               a.attname::text,
               format_type(a.atttypid, a.atttypmod),
               a.attnotnull
          FROM pg_attribute a
          JOIN pg_class     c ON c.oid = a.attrelid
         WHERE a.attrelid IN ('public.zone_feed_state'::regclass,
                              'public.zone_feed_refusal'::regclass)
           AND a.attnum > 0
           AND NOT a.attisdropped
    )
    SELECT string_agg(
               format('  %s.%s: expected %s%s, found %s',
                      e.tbl, e.col, e.typ,
                      CASE WHEN e.not_null THEN ' NOT NULL' ELSE '' END,
                      CASE
                          WHEN a.col IS NULL THEN 'no such column'
                          ELSE a.typ || CASE WHEN a.not_null THEN ' NOT NULL' ELSE '' END
                      END),
               E'\n' ORDER BY e.tbl, e.col)
      INTO divergences
      FROM expected e
      LEFT JOIN actual a ON a.tbl = e.tbl AND a.col = e.col
     WHERE a.col IS NULL
        OR (a.typ, a.not_null) IS DISTINCT FROM (e.typ, e.not_null);

    IF divergences IS NOT NULL THEN
        RAISE EXCEPTION
            E'tables this migration would have created already exist in a different shape:\n%',
            divergences
            USING HINT = 'CREATE TABLE IF NOT EXISTS repaired none of this. Reconcile the existing tables deliberately, or roll them back with 0002_zone_feed.down.sql, before applying 0002.';
    END IF;

    SELECT string_agg(format('  %s on %s', x.name, x.tbl), E'\n' ORDER BY x.name)
      INTO divergences
      FROM (VALUES ('zone_feed_state_singleton'::text,      'public.zone_feed_state'::text),
                   ('zone_feed_refusal_reason_known',       'public.zone_feed_refusal')) AS x(name, tbl)
     WHERE NOT EXISTS (SELECT 1
                         FROM pg_constraint k
                        WHERE k.conrelid = x.tbl::regclass
                          AND k.conname  = x.name
                          AND k.contype  = 'c'
                          AND k.convalidated);

    IF divergences IS NOT NULL THEN
        RAISE EXCEPTION
            E'CHECK constraints this migration promises are missing or NOT VALID:\n%',
            divergences;
    END IF;

    SELECT string_agg(format('  %s: primary key is (%s), expected (%s)',
                             x.tbl, COALESCE(pk.cols, 'none'), x.cols),
                      E'\n' ORDER BY x.tbl)
      INTO divergences
      FROM (VALUES ('public.zone_feed_state'::text, 'singleton'::text),
                   ('public.zone_feed_refusal',     'id')) AS x(tbl, cols)
      LEFT JOIN LATERAL (
               SELECT string_agg(a.attname::text, ',' ORDER BY u.ord) AS cols
                 FROM pg_constraint c
                 CROSS JOIN LATERAL unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
                 JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = u.attnum
                WHERE c.conrelid = x.tbl::regclass
                  AND c.contype  = 'p') pk ON true
     WHERE pk.cols IS DISTINCT FROM x.cols;

    IF divergences IS NOT NULL THEN
        RAISE EXCEPTION
            E'primary keys are not what this migration declares:\n%',
            divergences;
    END IF;
END
$$;

COMMIT;
