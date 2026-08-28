-- 0001_zone_store.sql — the synced zone store
--
-- Creates `zone`, `zone_incoming`, `sync_run`, and the one index on `zone`.
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
-- IDEMPOTENCY AND ITS LIMIT. Every statement is `IF NOT EXISTS` or a guarded
-- `DO` block, so re-applying over a partially applied state completes instead
-- of aborting. What that does NOT do is repair an object that exists in a
-- different shape: `CREATE TABLE IF NOT EXISTS` on an existing table adds no
-- missing column and no missing constraint, silently. Detecting that case is
-- the verify script's job, not this file's.
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

BEGIN;

-- --- Preconditions ----------------------------------------------------------
--
-- Two version floors. Neither chooses the PostGIS version — that is open under
-- `DEPLOYMENT.md § Still owed by this document` — and each is read off a floor
-- an existing section already states. The migration refuses a stack it is
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

CREATE EXTENSION IF NOT EXISTS postgis;

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

CREATE TABLE IF NOT EXISTS zone (
    id               integer          PRIMARY KEY,
    name             text             NOT NULL,
    latitude         double precision NOT NULL,
    longitude        double precision NOT NULL,

    -- The axis order appears exactly once in this system: here, in DDL.
    -- ST_MakePoint takes X then Y — longitude then latitude — which is the
    -- exact inversion the deleted prototype shipped. Because the column is
    -- GENERATED, the write path never supplies the point and therefore cannot
    -- invert it. The two scalars above are kept precisely so that the derived
    -- point has something to be checked against; see
    -- `Architecture.md § Geometry, SRID, and the coordinate guard`.
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
    -- The real guard is the generated column above and the three assertions in
    -- `0001_zone_store.verify.sql`.
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
-- scalars, so the staging table has nothing to derive and nothing to invert.

CREATE UNLOGGED TABLE IF NOT EXISTS zone_incoming (
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

CREATE TABLE IF NOT EXISTS sync_run (
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

-- --- The index --------------------------------------------------------------
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

CREATE INDEX IF NOT EXISTS zone_geom_gist ON zone USING gist (geom);

COMMIT;
