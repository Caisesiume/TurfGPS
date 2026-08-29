-- 0001_zone_store.sql — the synced zone store
--
-- Creates `zone`, `zone_incoming`, `sync_run`, one index on `zone` and two on
-- `sync_run`.
-- This is the first migration in the repository.
--
-- The design is settled in `Architecture.md § The schema` and its subsections.
-- That document is the home of every measurement behind the choices here; this
-- file cites it and deliberately does not restate a figure from it.
--
-- ----------------------------------------------------------------------------
-- NEVER EXECUTED.
--
-- No database exists in this repository and none was provisioned to author
-- this file. It has not been applied, not been rolled back, and not been
-- EXPLAINed against a running PostgreSQL. `postgis-migration-protocol` step 1
-- (read live state) and step 4 (test on a copy) could not be performed because
-- there is no live state and no copy — not because they were skipped.
--
-- `Architecture.md § What is unproven` item 1 states the acceptance condition
-- this migration therefore does not yet meet: EXPLAIN (ANALYZE, BUFFERS)
-- output for the real query shapes against a loaded copy.
-- `0001_zone_store.verify.sql` is the instrument that produces it, and a first
-- run of that script is owed before this file is applied anywhere that matters.
--
-- The divergence check at the end of this file is covered by the same
-- sentence. It has never been parsed by a PostgreSQL server, and until it has
-- been run once against a copy that was deliberately diverged first — so that
-- it is seen to RAISE, and not merely seen to pass — it is a guard that is
-- written rather than a guard that is known to fire.
-- ----------------------------------------------------------------------------
--
-- APPLY. By an authorised operator, deliberately, never at boot and never by
-- this repository's own tooling, per `postgis-migration-protocol`:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0001_zone_store.sql
--
-- No psql meta-command appears in this file or in its siblings, so a CRLF
-- checkout cannot break the invocation.
--
-- ROLLBACK. `0001_zone_store.down.sql`, in this directory. It is DESTRUCTIVE —
-- it drops the zone table and everything in it — and applying it is a human's
-- decision, taken with the same authorisation as the apply.
--
-- IRREVERSIBLE STEPS. None. Every object created here is dropped by the
-- rollback, except the PostGIS extension, which the rollback deliberately
-- leaves in place.
--
-- IDEMPOTENCY, AND THE CHECK THAT MAKES IT SAFE. Every statement is
-- `IF NOT EXISTS` or a guarded `DO` block, so re-applying over a partially
-- applied state completes instead of aborting. On its own that is not enough,
-- and the gap is why this file ends in a divergence check: `CREATE TABLE IF
-- NOT EXISTS` over a table that already exists in a DIFFERENT shape adds no
-- missing column and no missing constraint, and says nothing while it does
-- so. The migration then reports success over a store it did not build, and
-- every later statement — the merge, the rate-limit gate, the audit record —
-- runs against that store believing otherwise.
--
-- So this migration asserts its own postcondition before it commits, against
-- the live catalogue rather than against this file: every column, type,
-- nullability and generated-ness it promises, every named constraint, both
-- primary keys, all three indexes, `zone_incoming`'s UNLOGGED persistence, and
-- `zone.geom`'s type, SRID and generation. It runs inside the same
-- transaction as the DDL, so a divergence raises and rolls the whole
-- migration back rather than leaving a half-truth behind — and detection does
-- not depend on anyone remembering to run the verify script afterwards.
-- `0001_zone_store.verify.sql` remains the deeper instrument, and it asserts
-- things this check cannot; what it stops being is the only thing standing
-- between a divergent table and a green apply.
--
-- DECISIONS TAKEN HERE, and why, since a store outlives the reasoning that
-- shaped it:
--
--   1. Plain `CREATE INDEX`, not `CREATE INDEX CONCURRENTLY`. The rule in
--      `Architecture.md § Migrating against a running sync` is scoped by that
--      section's own first sentence to a table already live with a sync
--      writing to it. Here the table is created three statements earlier and
--      can have neither reader nor writer. CONCURRENTLY would buy nothing and
--      cost the two things that matter most in a first migration: it cannot
--      run inside a transaction block, so 0001 could not be atomic, and a
--      failed build would leave an INVALID index behind to be found and
--      dropped by hand. Later migrations against the live table use
--      CONCURRENTLY and carry that drop in their rollback.
--
--   2. `sync_run.outcome` carries a non-terminal value, `running`, and the
--      vocabulary is enforced by a CHECK rather than left to convention. The
--      values and the meaning of each are recorded in
--      `Architecture.md § The sync write path`. The reason a non-terminal
--      value has to exist is that a worker killed mid-run writes nothing, so
--      without a row inserted at the start of the run a crashed run and a run
--      that never happened are the same observation.
--
--   3. No `fillfactor` on `zone`. `Architecture.md § What is unproven` item 8
--      records that one below 100 is wanted and that none has been chosen, and
--      choosing one here would settle by fiat a question that section holds
--      open for measurement. It is a table-level reloption, so setting it
--      later is cheap and rewrites nothing.
--
--   4. No constraint on `zone` beyond the four that
--      `Architecture.md § What becomes a constraint` enforces. Six further
--      properties are as well measured there and deliberately not enforced,
--      on the argument that a constraint's only power over upstream data is to
--      convert Turf changing something into a local outage.
--
--   5. `search_path` is pinned for the transaction, and every relation this
--      file creates or reads is schema-qualified. Left unpinned, unqualified
--      `sync_run` resolves through whatever path the applying session happens
--      to carry, and the two questions this store exists to answer — has
--      enough time passed to fetch again, and what did the last run do —
--      could be answered by a table this migration never built. That is the
--      CVE-2018-1058 class, and it is silent by construction. `pg_catalog`
--      leads the pinned path so no built-in can be shadowed, `public` follows
--      because that is where PostGIS and this schema live, and `pg_temp` is
--      named LAST rather than left implicit, which is the only way to stop a
--      temporary relation being searched ahead of both.
--
--      TWO CONSEQUENCES FOR ANYONE EDITING THIS FILE. Because `pg_catalog`
--      leads, an UNQUALIFIED `CREATE` here would target `pg_catalog`: every
--      `CREATE` below names `public` explicitly and any statement added later
--      must too. And because PostGIS functions are reached through `public`,
--      the preconditions refuse an installation that put the extension
--      somewhere else, rather than resolving `ST_MakePoint` through a path
--      this file did not choose.
--
--   6. The divergence check asserts `expected ⊆ actual`, not equality. A
--      column this file does not know about is not evidence that this
--      migration failed, and refusing one would make 0001 unre-runnable the
--      moment 0002 adds one. The direction that matters here is the other
--      one: something this file promised, absent or differently shaped.
--      Asserting the ABSENCE of a column is a live requirement — it is
--      `current_owner` and `date_last_taken` specifically, it is about what a
--      later migration may add rather than about what 0001 built, and part A
--      of `0001_zone_store.verify.sql` is its home.
--
--      ONE LIMIT, STATED RATHER THAN LEFT TO BE DISCOVERED. Constraints are
--      checked by name, type and validity, not by expression text.
--      PostgreSQL normalises a CHECK expression when it stores it, so a
--      string comparison against `pg_get_constraintdef` output would be
--      brittle across server versions rather than strict — it would fail on
--      correct databases, which is the fastest way to get a guard deleted. A
--      constraint that is present under the right name and carries a
--      different expression is therefore not caught here.
--
--   7. A second index, on `sync_run`, for the currency read. `zonestore` asks
--      for the `completed_at` of the latest successful run — which is the
--      instant that run's merge BEGAN and not the instant it finished, because
--      the column carries `now()`, and `now()` is `transaction_timestamp()`,
--      fixed for the whole of the transaction the merge runs in. It is a lower
--      bound on the commit, and `zonestore.Currency.LastSuccessAt` is where
--      what that is worth to a reader is stated. It is also the one question
--      about this store that a request may reach. Against
--      the primary key alone it is a sequential scan and a sort over every
--      run ever recorded — a table that grows by roughly 17,500 rows a year
--      and is never pruned. `sync_run_completed_at_ok` is partial on
--      `outcome = 'ok'` and descending on `completed_at`, which is the
--      query's own predicate and its own ordering, so the answer is the first
--      entry the scan reads. It costs the sync two index entries per run,
--      forty-eight times a day. This is free to take now only because 0001
--      has not been applied anywhere; after an apply it would be a second
--      migration. `Architecture.md § The indexes` counts this index and
--      decision 8's: the line that recorded one on `sync_run` was owed two
--      more, and it was corrected to three in this same change.
--
--   8. A third index, also on `sync_run`, for the rate limit's own gate.
--      `syncstore.sinceLastAttemptSQL` asks for `max(started_at)` over every
--      outcome, and `zonesync.Scheduler` reads it twice per interval, both
--      through `untilDue`: once by the loop deciding whether to sleep, and once
--      after the lock is acquired, where it decides whether another holder
--      refreshed the copy while this attempt waited for it. Neither existing
--      index can answer that query. The primary key is on `id`, which leaves a
--      sequential scan over a table that grows by roughly 17,500 rows a year
--      and is never pruned; `sync_run_completed_at_ok` is partial on
--      `outcome = 'ok'` and keyed on `completed_at`, where this query filters
--      on no outcome at all and reads a different column.
--
--      So `sync_run_started_at` is a plain btree on `started_at DESC`, with no
--      partial predicate because the query excludes no row — a run still in
--      flight and a run that died have both spent the request, which is the
--      whole reason that statement looks at every outcome. PostgreSQL rewrites
--      `max()` over an indexed column into a one-row scan of that index, and
--      `DESC` is written out so that scan reads forward. `started_at` is NOT
--      NULL, so the NULLS FIRST that DESC implies orders nothing.
--
--      It is taken here for the same reason decision 7's index is, and the
--      reasoning is not new: while 0001 is unapplied an index costs one line,
--      and afterwards it is CREATE INDEX CONCURRENTLY in a migration of its
--      own, which by decision 1 cannot share a transaction with anything else.

BEGIN;

-- --- The schema this migration resolves in ----------------------------------
--
-- Decision 5 in the header is the argument; this is the statement. `SET LOCAL`
-- is scoped to this transaction and the session's own path is restored at
-- COMMIT or ROLLBACK, so applying this file changes nothing about the session
-- that applied it.
--
-- `pg_catalog` leads, so an unqualified CREATE below would land in it. Every
-- CREATE names `public`.

SET LOCAL search_path = pg_catalog, public, pg_temp;

-- --- Preconditions ----------------------------------------------------------
--
-- Two version floors. Neither chooses the PostGIS version — that is open under
-- `DEPLOYMENT.md § Open questions owned by this document`, which owns the target
-- host distribution and the PostGIS version it implies — and each is read off a
-- floor an existing section already states. The migration refuses a stack it is
-- known to be wrong on rather than applying and being quietly wrong on it.

DO $$
DECLARE
    server_num int := current_setting('server_version_num')::int;
BEGIN
    IF server_num < 120000 THEN
        RAISE EXCEPTION
            'zone.geom is a STORED generated column, which PostgreSQL 12 introduced; this server reports server_version_num=%',
            server_num;
    END IF;
END
$$;

CREATE EXTENSION IF NOT EXISTS postgis SCHEMA public;

-- A third precondition, and it exists because the statement above is
-- `IF NOT EXISTS`: an extension that already exists somewhere else is skipped
-- silently, `SCHEMA public` and all. Every PostGIS function this file calls —
-- `postgis_lib_version` here, `ST_MakePoint` and `ST_SetSRID` in the generated
-- column, `postgis_typmod_type` and `postgis_typmod_srid` in the divergence
-- check — is reached through `public` on the pinned path. If the extension is
-- not there, this migration refuses rather than resolving them through a path
-- it did not choose.

DO $$
DECLARE
    ext_schema text;
BEGIN
    SELECT n.nspname
      INTO ext_schema
      FROM pg_extension e
      JOIN pg_namespace n ON n.oid = e.extnamespace
     WHERE e.extname = 'postgis';

    IF ext_schema IS DISTINCT FROM 'public' THEN
        RAISE EXCEPTION
            'PostGIS is installed in schema %, and this migration resolves its functions through public: CREATE EXTENSION IF NOT EXISTS skips an extension that already exists, so the SCHEMA clause above did not move it.',
            COALESCE(ext_schema, '(not installed)')
            USING HINT = 'Install PostGIS in public, or amend this migration deliberately to name the schema it is in.';
    END IF;
END
$$;

DO $$
DECLARE
    lib   text   := postgis_lib_version();
    parts text[] := regexp_match(lib, '^(\d+)\.(\d+)');
BEGIN
    IF parts IS NULL THEN
        RAISE EXCEPTION
            'could not read a major.minor version from postgis_lib_version() = %', lib;
    END IF;

    IF (parts[1]::int, parts[2]::int) < (2, 2) THEN
        RAISE EXCEPTION
            'PostGIS % is below the 2.2 floor: before 2.2 the <-> operator orders geography by bounding box, which makes the nearest-neighbour query approximate without saying so. See Architecture.md, What is unproven, item 3.',
            lib;
    END IF;
END
$$;

-- --- zone -------------------------------------------------------------------
--
-- Seventeen columns, per `Architecture.md § The zone table`. Every column width
-- and every nullability decision below follows a measurement recorded there and
-- in `Architecture.md § Retrieving zones`. The ones a schema written from
-- intuition gets wrong are `total_takeovers`, which does not fit `smallint`,
-- and `type_id`, `region_country`, `area_id` and `area_name`, which are absent
-- often enough that NOT NULL would reject real responses.
--
-- There are deliberately NO columns for `currentOwner` or `dateLastTaken`.
-- They are round-scoped, they are absent from every record the sync reads, and
-- a synced table that acquires one acquires a rollover problem it does not
-- currently have. The argument is in
-- `Architecture.md § The two absences, and the test that keeps them absent`;
-- the verify script is its enforcement.

CREATE TABLE IF NOT EXISTS public.zone (
    id               integer          PRIMARY KEY,
    name             text             NOT NULL,
    latitude         double precision NOT NULL,
    longitude        double precision NOT NULL,

    -- The POINT'S axis order appears exactly once in this system: here, in DDL.
    -- ST_MakePoint takes X then Y — longitude then latitude — which is the
    -- exact inversion the deleted prototype shipped. Because the column is
    -- GENERATED, the write path never supplies the point and therefore cannot
    -- invert it.
    --
    -- WHAT IS DECIDED AGAIN, ELSEWHERE, IS WHICH VALUE REACHES WHICH COLUMN.
    -- This expression is faithful to the two columns it reads, and that is the
    -- limit of what it promises. The binding of a value to an axis lives in
    -- `syncstore.zoneColumns` and is pinned in Go by
    -- `TestEveryColumnIsLoadedFromItsOwnField`. Why nothing in this file can
    -- detect a crossed binding — including the ST_X/ST_Y assertion, which
    -- passes under one — is `Architecture.md § What the DDL cannot reach`, and
    -- is not restated here.
    --
    -- The two scalars above are kept precisely so that the derived
    -- point has something to be checked against; see
    -- `Architecture.md § Geometry, SRID, and the coordinate guard` and
    -- `Architecture.md § What the DDL cannot reach`.
    geom             geography(Point, 4326)
                     GENERATED ALWAYS AS (
                       ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography
                     ) STORED,

    date_created     timestamptz      NOT NULL,
    total_takeovers  integer          NOT NULL,
    takeover_points  smallint         NOT NULL,
    points_per_hour  smallint         NOT NULL,
    type_id          smallint,
    region_id        smallint         NOT NULL,
    region_name      text             NOT NULL,
    region_country   text,
    area_id          integer,
    area_name        text,

    -- Bookkeeping, not Turf's. `first_seen_at` is written on insert and never
    -- updated; `last_changed_at` is stamped with the run's `completed_at`, so
    -- every changed row is attributable to a row of `sync_run`.
    first_seen_at    timestamptz      NOT NULL,
    last_changed_at  timestamptz      NOT NULL,

    -- These range checks are a weak guard against a coordinate swap and are
    -- kept for what they catch, not for what they prove:
    -- `Architecture.md § Geometry, SRID, and the coordinate guard` measures how
    -- much of the corpus a table-wide swap leaves inside both ranges, and the
    -- answer is nearly all of it — including every one of the primary markets.
    -- What guards the DDL is the generated column above and the three
    -- assertions in `0001_zone_store.verify.sql` — and that is the half of the
    -- question SQL can reach. A swap in the WRITE PATH, the longitude value
    -- bound to the `latitude` column, leaves the generated point agreeing with
    -- the two columns it was generated from and passes every one of those
    -- assertions. That binding is pinned in Go, by
    -- `TestEveryColumnIsLoadedFromItsOwnField` in
    -- `service/internal/syncstore/columns_test.go`; see the note on `geom`
    -- above and `Architecture.md § What the DDL cannot reach`.
    CONSTRAINT zone_lat_range        CHECK (latitude  BETWEEN  -90 AND  90),
    CONSTRAINT zone_lon_range        CHECK (longitude BETWEEN -180 AND 180),
    CONSTRAINT zone_takeovers_nonneg CHECK (total_takeovers >= 0)
);

-- --- zone_incoming ----------------------------------------------------------
--
-- The sync's staging table, per `Architecture.md § The sync write path`.
-- UNLOGGED because it is truncated and rebuilt from the response on every run:
-- a bulk COPY into it generates no WAL, and losing it to a crash costs nothing
-- the next run does not rebuild.
--
-- It carries no constraints, no primary key and no index, and that is the
-- design rather than an omission. The properties a constraint would enforce
-- are checked as assertions over the staged rows before the merge — where a
-- violation can be reported and the run recorded, instead of aborting a
-- transaction and leaving the table stale. Which assertions those are is
-- recorded in the section cited above.
--
-- It carries no `geom` either: `zone.geom` is generated from the merged
-- scalars, so the staging table has no point to derive.
--
-- THAT IS THE DERIVATION, AND IT IS NOT THE AXIS QUESTION. This is in fact the
-- one relation where a crossed WRITE PATH physically lands. The sync COPYs into
-- the column list derived from `syncstore.zoneColumns`, positionally, so which
-- value arrives in the `latitude` column declared below and which in
-- `longitude` is decided in Go before a byte reaches this table; the merge then
-- carries that pair into `zone` unchanged, and `zone.geom` is generated from it
-- faithfully. Nothing here detects that — not the absent constraints, and not
-- the absent generated column, neither of which would catch it if it were
-- present. That binding is pinned by `TestEveryColumnIsLoadedFromItsOwnField`
-- in `service/internal/syncstore/columns_test.go`; see the note on `geom` above
-- and `Architecture.md § What the DDL cannot reach`.

CREATE UNLOGGED TABLE IF NOT EXISTS public.zone_incoming (
    id               integer,
    name             text,
    latitude         double precision,
    longitude        double precision,
    date_created     timestamptz,
    total_takeovers  integer,
    takeover_points  smallint,
    points_per_hour  smallint,
    type_id          smallint,
    region_id        smallint,
    region_name      text,
    region_country   text,
    area_id          integer,
    area_name        text
);

-- --- sync_run ---------------------------------------------------------------
--
-- Sync bookkeeping, per `Architecture.md § The sync write path`. A table rather
-- than log lines, because the questions asked of it are queries.
--
-- `started_at` and `outcome` are the only NOT NULL columns, and that is what
-- makes a failed or partial run recordable: a run that never reached the
-- endpoint records a row carrying those two and nothing else. A run that
-- failed is therefore distinguishable from a run that never happened, the
-- latter having no row at all.
--
-- `absent_ids` records ids present in the table and missing from the response.
-- Nothing is ever deleted on the strength of it —
-- `Architecture.md § Absence is recorded and never acted on` is the argument,
-- and the asymmetry it turns on is that deletion has never been observed while
-- a truncated response has.

CREATE TABLE IF NOT EXISTS public.sync_run (
    id             bigserial   PRIMARY KEY,
    started_at     timestamptz NOT NULL,
    completed_at   timestamptz,
    outcome        text        NOT NULL,
    http_status    smallint,
    response_bytes bigint,
    zones_received integer,
    rows_inserted  integer,
    rows_updated   integer,
    rows_unchanged integer,
    absent_count   integer,
    absent_ids     integer[],

    -- The vocabulary and the meaning of each value live in
    -- `Architecture.md § The sync write path`. It is enforced here rather than
    -- left to convention because this column is the only durable record of
    -- what a run did, and a value nobody anticipated is a run that cannot be
    -- counted.
    CONSTRAINT sync_run_outcome_known CHECK (
        outcome IN ('running', 'ok', 'http_error', 'assertion_failed', 'aborted')
    )
);

-- --- The indexes -----------------------------------------------------------
--
-- One index on `zone`, besides the primary key's own.
-- `Architecture.md § The indexes` records the four that are deliberately not
-- created, and why each of them looks obviously useful and is not.
--
-- This one serves both spatial query shapes — containment within a corridor
-- and nearest-neighbour ordering — off a single index, which is a property of
-- GiST over geography rather than a coincidence to rely on silently. It also
-- has no cell size, no precision and no bits parameter: there is nothing to
-- tune and therefore nothing to tune wrongly, which on this surface is a
-- selection criterion rather than a convenience.
--
-- Creating it proves nothing. Only EXPLAIN on the real query shape counts, and
-- there is no database here to run one against.

CREATE INDEX IF NOT EXISTS zone_geom_gist ON public.zone USING gist (geom);

-- The second index, on `sync_run`, for the currency read. Decision 7 in the
-- header is the argument. The query it serves is `zonestore.currencySQL`:
--
--     SELECT completed_at
--     FROM   sync_run
--     WHERE  outcome = 'ok' AND completed_at IS NOT NULL
--     ORDER  BY completed_at DESC
--     LIMIT  1
--
-- Two choices below are deliberate and both could be made differently.
--
-- The partial predicate is `outcome = 'ok'` ALONE, omitting the query's
-- `completed_at IS NOT NULL`. The query's predicate implies this one, so the
-- planner may still use the index, and the narrower form would buy nothing: a
-- row at `ok` without a `completed_at` is not a state the write path can
-- produce. Omitting it keeps the index serving a currency read written without
-- that clause.
--
-- `DESC` is written out rather than left to a backward scan of an ascending
-- index, so the index's ordering — descending, NULLS FIRST by implication —
-- matches the query's own exactly and the LIMIT 1 needs no sort above it.
--
-- Creating it proves nothing here either, for the reason the paragraph above
-- gives about the GiST index — and nothing in this repository has EXPLAINed
-- it. Part E of `0001_zone_store.verify.sql` probes `zone_geom_gist` and
-- carries no `sync_run` query at all, so this index and the one below it stay
-- unproven even after part E has run against a loaded copy.

CREATE INDEX IF NOT EXISTS sync_run_completed_at_ok
    ON public.sync_run (completed_at DESC)
    WHERE outcome = 'ok';

-- The third index, also on `sync_run`, for the rate limit's gate. Decision 8 in
-- the header is the argument. The query it serves is
-- `syncstore.sinceLastAttemptSQL`:
--
--     SELECT extract(epoch FROM now() - max(started_at)) FROM sync_run
--
-- The aggregate is what the index is for: with no GROUP BY and a btree on the
-- aggregated column, the planner replaces `max(started_at)` with a scan of this
-- index returning one row, and the arithmetic above it reads that one value.
-- Without it the same statement reads every row ever recorded, twice per
-- interval, to compute a maximum.
--
-- Unlike the index above it this one is NOT partial. `sinceLastAttemptSQL`
-- deliberately spans every outcome — a run in flight and a run that died have
-- both spent the request — so a predicate here would exclude rows the gate
-- exists to count.
--
-- Creating it proves nothing, for the reason both of its neighbours give above.
-- Nothing in this repository has EXPLAINed it.

CREATE INDEX IF NOT EXISTS sync_run_started_at
    ON public.sync_run (started_at DESC);

-- --- The divergence check ---------------------------------------------------
--
-- This migration's postcondition, asserted against the live catalogue before
-- it commits. The header paragraph IDEMPOTENCY, AND THE CHECK THAT MAKES IT
-- SAFE is the argument for its existence; decision 6 states its one direction
-- and its one limit.
--
-- Everything below reads `pg_catalog` unqualified, which is exact because
-- `pg_catalog` leads the path pinned at the top of this transaction, and names
-- every relation it examines as `public.<name>` for the same reason the CREATE
-- statements do.

-- Columns: name, type, nullability, and whether the column is generated.
-- `zone.geom` is not in this list; it is checked on its own below, where the
-- SRID can be read as a number instead of as part of a rendered type name.
DO $$
DECLARE
    divergences text;
BEGIN
    WITH expected(tbl, col, typ, not_null) AS (
        VALUES
            ('zone'::text,    'id'::text,        'integer'::text,            true),
            ('zone',          'name',            'text',                     true),
            ('zone',          'latitude',        'double precision',         true),
            ('zone',          'longitude',       'double precision',         true),
            ('zone',          'date_created',    'timestamp with time zone', true),
            ('zone',          'total_takeovers', 'integer',                  true),
            ('zone',          'takeover_points', 'smallint',                 true),
            ('zone',          'points_per_hour', 'smallint',                 true),
            ('zone',          'type_id',         'smallint',                 false),
            ('zone',          'region_id',       'smallint',                 true),
            ('zone',          'region_name',     'text',                     true),
            ('zone',          'region_country',  'text',                     false),
            ('zone',          'area_id',         'integer',                  false),
            ('zone',          'area_name',       'text',                     false),
            ('zone',          'first_seen_at',   'timestamp with time zone', true),
            ('zone',          'last_changed_at', 'timestamp with time zone', true),

            ('zone_incoming', 'id',              'integer',                  false),
            ('zone_incoming', 'name',            'text',                     false),
            ('zone_incoming', 'latitude',        'double precision',         false),
            ('zone_incoming', 'longitude',       'double precision',         false),
            ('zone_incoming', 'date_created',    'timestamp with time zone', false),
            ('zone_incoming', 'total_takeovers', 'integer',                  false),
            ('zone_incoming', 'takeover_points', 'smallint',                 false),
            ('zone_incoming', 'points_per_hour', 'smallint',                 false),
            ('zone_incoming', 'type_id',         'smallint',                 false),
            ('zone_incoming', 'region_id',       'smallint',                 false),
            ('zone_incoming', 'region_name',     'text',                     false),
            ('zone_incoming', 'region_country',  'text',                     false),
            ('zone_incoming', 'area_id',         'integer',                  false),
            ('zone_incoming', 'area_name',       'text',                     false),

            ('sync_run',      'id',              'bigint',                   true),
            ('sync_run',      'started_at',      'timestamp with time zone', true),
            ('sync_run',      'completed_at',    'timestamp with time zone', false),
            ('sync_run',      'outcome',         'text',                     true),
            ('sync_run',      'http_status',     'smallint',                 false),
            ('sync_run',      'response_bytes',  'bigint',                   false),
            ('sync_run',      'zones_received',  'integer',                  false),
            ('sync_run',      'rows_inserted',   'integer',                  false),
            ('sync_run',      'rows_updated',    'integer',                  false),
            ('sync_run',      'rows_unchanged',  'integer',                  false),
            ('sync_run',      'absent_count',    'integer',                  false),
            ('sync_run',      'absent_ids',      'integer[]',                false)
    ),
    actual(tbl, col, typ, not_null, is_generated) AS (
        SELECT c.relname::text,
               a.attname::text,
               format_type(a.atttypid, a.atttypmod),
               a.attnotnull,
               a.attgenerated <> ''
          FROM pg_attribute a
          JOIN pg_class     c ON c.oid = a.attrelid
         WHERE a.attrelid IN ('public.zone'::regclass,
                              'public.zone_incoming'::regclass,
                              'public.sync_run'::regclass)
           AND a.attnum > 0
           AND NOT a.attisdropped
    )
    SELECT string_agg(
               format('  %s.%s: expected %s%s, found %s',
                      e.tbl, e.col, e.typ,
                      CASE WHEN e.not_null THEN ' NOT NULL' ELSE '' END,
                      CASE
                          WHEN a.col IS NULL THEN 'no such column'
                          ELSE a.typ
                               || CASE WHEN a.not_null     THEN ' NOT NULL'  ELSE '' END
                               || CASE WHEN a.is_generated THEN ' GENERATED' ELSE '' END
                      END),
               E'\n' ORDER BY e.tbl, e.col)
      INTO divergences
      FROM expected e
      LEFT JOIN actual a ON a.tbl = e.tbl AND a.col = e.col
     WHERE a.col IS NULL
        OR (a.typ, a.not_null, a.is_generated) IS DISTINCT FROM (e.typ, e.not_null, false);

    IF divergences IS NOT NULL THEN
        RAISE EXCEPTION
            E'tables this migration would have created already exist in a different shape:\n%',
            divergences
            USING HINT = 'CREATE TABLE IF NOT EXISTS repaired none of this. Reconcile the existing tables deliberately, or roll them back with 0001_zone_store.down.sql, before applying 0001.';
    END IF;
END
$$;

-- Constraints, primary keys, indexes, and the staging table's persistence.
DO $$
DECLARE
    problems text;
BEGIN
    SELECT string_agg(format('  %s on %s', x.name, x.tbl), E'\n' ORDER BY x.name)
      INTO problems
      FROM (VALUES ('zone_lat_range'::text,   'public.zone'::text),
                   ('zone_lon_range',         'public.zone'),
                   ('zone_takeovers_nonneg',  'public.zone'),
                   ('sync_run_outcome_known', 'public.sync_run')) AS x(name, tbl)
     WHERE NOT EXISTS (SELECT 1
                         FROM pg_constraint k
                        WHERE k.conrelid = x.tbl::regclass
                          AND k.conname  = x.name
                          AND k.contype  = 'c'
                          AND k.convalidated);

    IF problems IS NOT NULL THEN
        RAISE EXCEPTION
            E'CHECK constraints this migration promises are missing or NOT VALID:\n%',
            problems;
    END IF;

    SELECT string_agg(format('  %s: primary key is (%s), expected (%s)',
                             x.tbl, COALESCE(pk.cols, 'none'), x.cols),
                      E'\n' ORDER BY x.tbl)
      INTO problems
      FROM (VALUES ('public.zone'::text, 'id'::text),
                   ('public.sync_run',   'id')) AS x(tbl, cols)
      LEFT JOIN LATERAL (
               SELECT string_agg(a.attname::text, ',' ORDER BY u.ord) AS cols
                 FROM pg_constraint c
                 CROSS JOIN LATERAL unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
                 JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = u.attnum
                WHERE c.conrelid = x.tbl::regclass
                  AND c.contype  = 'p') pk ON true
     WHERE pk.cols IS DISTINCT FROM x.cols;

    IF problems IS NOT NULL THEN
        RAISE EXCEPTION
            E'primary keys are not what this migration declares:\n%',
            problems;
    END IF;

    SELECT string_agg(format('  %s on %s using %s (%s)', x.name, x.tbl, x.am, x.col),
                      E'\n' ORDER BY x.name)
      INTO problems
      FROM (VALUES ('zone_geom_gist'::text,     'public.zone'::text, 'gist'::text, 'geom'::text),
                   ('sync_run_completed_at_ok', 'public.sync_run',   'btree',      'completed_at'),
                   ('sync_run_started_at',      'public.sync_run',   'btree',      'started_at')) AS x(name, tbl, am, col)
     WHERE NOT EXISTS (SELECT 1
                         FROM pg_index     i
                         JOIN pg_class     ic ON ic.oid = i.indexrelid
                         JOIN pg_am        m  ON m.oid  = ic.relam
                         JOIN pg_attribute a  ON a.attrelid = i.indrelid
                                             AND a.attnum   = i.indkey[0]
                        WHERE i.indrelid = x.tbl::regclass
                          AND ic.relname = x.name
                          AND m.amname   = x.am
                          AND a.attname  = x.col
                          AND i.indisvalid
                          AND i.indisready);

    IF problems IS NOT NULL THEN
        RAISE EXCEPTION
            E'indexes this migration promises are absent, invalid, or built on something else:\n%',
            problems
            USING HINT = 'An index carrying the right name on the wrong column or the wrong access method is reported here as missing, because that is what it is to the queries it was created to serve.';
    END IF;

    IF (SELECT c.relpersistence
          FROM pg_class c
         WHERE c.oid = 'public.zone_incoming'::regclass) <> 'u' THEN
        RAISE EXCEPTION
            'zone_incoming exists and is not UNLOGGED'
            USING HINT = 'The staging table is truncated and rebuilt from the response on every run. A LOGGED one writes WAL for every staged row, to protect data the next run reconstructs anyway.';
    END IF;
END
$$;

-- `zone.geom`, on its own, because this is the column the deleted prototype got
-- wrong. Type, geometry type and SRID are read as values rather than matched
-- against a rendered type name, which keeps the assertion exact across PostGIS
-- versions instead of exact against one of them. The generation check is the
-- load-bearing one: a `geom` that is an ordinary column rather than GENERATED
-- is NULL on every row, because the write path never supplies a point — that
-- being the whole reason the column is generated.
--
-- ONE THING THIS CHECK DOES NOT READ, stated so the gap is chosen rather than
-- assumed closed: the generation EXPRESSION. `attgenerated` records that the
-- column is computed, not what computes it, so an expression handing
-- ST_MakePoint its two arguments the other way round satisfies every assertion
-- below — the limit decision 6 states for CHECK constraints, on the one column
-- where it costs most. What would catch it is measurement, not catalogue text:
-- part C of `0001_zone_store.verify.sql` inserts a pair of known coordinates by
-- name into a clone taken INCLUDING GENERATED and asserts the distance between
-- them. That file has never been run.
--
-- AND IT DOES NOT CLOSE THE OTHER AXIS QUESTION, which nothing in SQL can. If
-- the WRITE PATH crosses the pair — the longitude value bound to the `latitude`
-- column — `geom` is generated from those columns and goes on agreeing with
-- them exactly as it did before, while the zone sits in another country. That
-- binding is decided in Go before a statement is sent, and it is pinned by
-- `TestEveryColumnIsLoadedFromItsOwnField` in
-- `service/internal/syncstore/columns_test.go`.
DO $$
DECLARE
    geom_type   regtype;
    geom_typmod integer;
    geom_gen    "char";
BEGIN
    SELECT a.atttypid, a.atttypmod, a.attgenerated
      INTO geom_type, geom_typmod, geom_gen
      FROM pg_attribute a
     WHERE a.attrelid = 'public.zone'::regclass
       AND a.attname  = 'geom'
       AND a.attnum   > 0
       AND NOT a.attisdropped;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'zone exists without a geom column';
    END IF;

    IF geom_type <> 'geography'::regtype THEN
        RAISE EXCEPTION
            'zone.geom is %, not geography', geom_type
            USING HINT = 'geometry is planar. It would make every distance in the corridor and neighbourhood queries a projected one, and would not raise while doing it.';
    END IF;

    IF upper(postgis_typmod_type(geom_typmod)) IS DISTINCT FROM 'POINT'
       OR postgis_typmod_srid(geom_typmod) IS DISTINCT FROM 4326 THEN
        RAISE EXCEPTION
            'zone.geom is geography(%, %), not geography(Point, 4326)',
            COALESCE(postgis_typmod_type(geom_typmod), 'unconstrained'),
            postgis_typmod_srid(geom_typmod);
    END IF;

    IF geom_gen <> 's' THEN
        RAISE EXCEPTION
            'zone.geom is not a STORED generated column'
            USING HINT = 'The write path supplies latitude and longitude and never a point. An ordinary geom column is NULL on every row, and every spatial query then returns nothing rather than failing.';
    END IF;
END
$$;

COMMIT;
