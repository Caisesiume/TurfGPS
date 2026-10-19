-- 0001_zone_store.verify.sql — the assertions that make 0001 falsifiable
--
-- Run against a migrated copy, never against production:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0001_zone_store.verify.sql
--
-- Every assertion below raises with a message naming what it found. A silent
-- run is a pass. It writes nothing that survives: it creates temporary objects,
-- one probe in part B inserts into `sync_run` inside subtransactions it then
-- discards, which consumes a few values from that table's id sequence and
-- leaves no row behind, and parts D and E fill an EMPTY `zone` with the
-- synthetic copy described above part A, inside a transaction each rolls back.
-- Nothing is ever written to a `zone` that holds a row, or to `zone_incoming`.
--
-- ----------------------------------------------------------------------------
-- EXECUTED ON THROWAWAY INSTANCES ONLY. CI's `migrations` job runs this file
-- part by part through `scripts/db/migrate-verify.sh` (#153), after applying
-- 0001 to an ephemeral PostGIS, on every pull request.
--
-- `Architecture.md § What is unproven` item 1 makes EXPLAIN (ANALYZE, BUFFERS)
-- against a loaded copy an acceptance condition of the first migration, for Q1
-- and Q2; Q3's is owed by the migration that creates `plan`. Part E prints
-- both plans and fails on either one that scans `zone` sequentially. It
-- carries no `sync_run` query at all, so the two indexes on `sync_run` are
-- still a prediction on the far side of it.
-- ----------------------------------------------------------------------------
--
-- Part A  the column set, in both directions
-- Part B  the structural checks IF NOT EXISTS cannot make
-- Part C  the coordinate guard, three assertions
-- Part D  the axis order over every row actually stored
-- Part E  EXPLAIN, which is the only thing that proves an index is used
--
-- A note on schema qualification, correcting the note this file used to carry.
--
-- That note said part A followed `current_schema()` because 0001 creates its
-- tables unqualified, so they land wherever the applying session's path puts
-- them. The premise is no longer true, and it was made untrue in this same
-- branch: 0001 pins `search_path` for its transaction and every CREATE in it
-- names `public`. Following the session's path here would now look somewhere
-- 0001 never wrote, and it would fail in the direction that says nothing — a
-- `zone` sitting earlier on the path answers every assertion below about a
-- table this migration did not build, and part C clones that decoy through
-- `LIKE` and pronounces its generation expression correct.
--
-- So this file pins the path too, and names `public` on every relation it
-- reads, writes or clones — which is also what
-- `Architecture.md § The two absences, and the test that keeps them absent`
-- writes part A's query against. The path is 0001's own, for the reasons
-- decision 5 of that file gives: `pg_catalog` first so no built-in can be
-- shadowed, `public` next because that is where PostGIS and this schema live,
-- `pg_temp` LAST rather than left implicit.
--
-- It is `SET` and not `SET LOCAL`, which is the one difference from 0001 and is
-- deliberate. This file is not a transaction — it is a sequence of statements
-- run under `ON_ERROR_STOP=1`, each committing on its own — and `SET LOCAL`
-- outside a transaction block raises a warning and then governs nothing. The
-- path is set for the psql session and no further than it.
--
-- One consequence, in part C. With `pg_temp` named last, an unqualified
-- reference to the temp clone would find an ordinary `public.zone_guard` ahead
-- of it if a database had one, so every reference to the clone is written
-- `pg_temp.zone_guard`. Its CREATE needs no such help: TEMP puts the table in
-- the temp schema whatever the path says.

SET search_path = pg_catalog, public, pg_temp;

-- --- The loaded copy, for parts D and E -------------------------------------
--
-- At sample size the planner seq-scans whatever indexes exist, so an empty
-- migrated copy proves nothing about index use. `load_zone_copy` fills an
-- EMPTY `zone` with 154,845 synthetic zones, the table's size in
-- `Architecture.md § Volume`, placed where real zones are dense. That is the
-- Owner's decision on #142 (Q2 (a), #issuecomment-6256273644), which replaced
-- the even spread over the dump's bounding box: an even spread puts no zone in
-- a corridor, so Q1 returned 0 rows and Q2's lateral side never executed.
--
-- THE WEIGHTS are `0001_zone_store.density.csv`, beside this file: one row per
-- 0.1-degree cell holding at least one zone of
-- `data/zones-dump-2025-01-04.json`, keyed by the cell's south-west corner,
-- with that cell's zone count. A cell is floor(latitude * 10) by
-- floor(longitude * 10), in decimal arithmetic on the dump's own literals. It
-- has 11,615 rows summing to the dump's 138,830 zones, and it holds no zone id
-- and no name. 0.1 degrees is about 11 km by 6 km at Stockholm's latitude,
-- and it is the finest grid the decision allows (0.1 to 1 degree). Only that
-- grid keeps a city corridor near its real count. By a spherical estimate,
-- part E's probe corridor holds 156 real zones and 179 synthetic ones on this
-- grid. On a 0.25-degree grid it would hold 83, and on a 0.5-degree grid 34.
--
-- THE LOAD scales the counts to 154,845 by rounding each cell's running total,
-- so the per-cell numbers sum to exactly that, and gives every cell at least
-- one. Ids run from 1 in cell order (south, then west). Inside a cell the
-- points follow the R2 low-discrepancy sequence, computed in numeric, so every
-- server produces the same rows.
--
-- A `zone` that already holds a row is used as it is and nothing is added, so
-- the load never meets a real row. Parts D and E call it inside a transaction
-- they roll back, so it never outlives the part. The table and the functions
-- are temporary and are defined ahead of every part. The functions are called
-- schema-qualified because `pg_temp` is never searched for functions. The
-- `\copy` path is relative to the repository root, which is where CI's
-- `migrations` job runs psql. From any other directory the `\copy` fails, and
-- every part fails with it.

CREATE TEMP TABLE zone_density (
    south numeric NOT NULL,
    west  numeric NOT NULL,
    zones integer NOT NULL CHECK (zones > 0),
    PRIMARY KEY (south, west)
);

\copy pg_temp.zone_density (south, west, zones) FROM 'migrations/0001_zone_store.density.csv' WITH (FORMAT csv, HEADER true)

CREATE FUNCTION pg_temp.load_zone_copy() RETURNS bigint
LANGUAGE plpgsql AS $fn$
DECLARE
    synthetic_zones constant bigint  := 154845;
    cell            constant numeric := 0.1;
    cells      bigint;
    dump_zones bigint;
    n          bigint;
BEGIN
    SELECT count(*) INTO n FROM public.zone;
    IF n > 0 THEN
        RAISE NOTICE 'loaded copy: zone already holds % row(s); used as it is, nothing synthetic added.', n;
        RETURN n;
    END IF;

    SELECT count(*), sum(zones) INTO cells, dump_zones FROM pg_temp.zone_density;
    IF cells = 0 THEN
        RAISE EXCEPTION 'loaded copy: the density table is empty, so there is nothing to weight the synthetic zones by.';
    END IF;

    -- sum() over an integer is a bigint, so both ends are rounded by integer
    -- division: floor((2 * running * synthetic + dump) / (2 * dump)).
    INSERT INTO public.zone (id, name, latitude, longitude, date_created,
                             total_takeovers, takeover_points, points_per_hour,
                             region_id, region_name, first_seen_at, last_changed_at)
    SELECT c.before + k, 'synthetic-' || (c.before + k),
           c.south + cell * ((0.5 + k * 0.5698402909980532) - floor(0.5 + k * 0.5698402909980532)),
           c.west  + cell * ((0.5 + k * 0.7548776662466927) - floor(0.5 + k * 0.7548776662466927)),
           timestamptz '2000-01-01T00:00:00Z', 0, 0, 0, 0, 'synthetic', now(), now()
      FROM (SELECT south, west,
                   (2 * (running - zones) * synthetic_zones + dump_zones) / (2 * dump_zones) AS before,
                   (2 * running           * synthetic_zones + dump_zones) / (2 * dump_zones) AS through
              FROM (SELECT south, west, zones,
                           sum(zones) OVER (ORDER BY south, west) AS running
                      FROM pg_temp.zone_density) AS d) AS c
     CROSS JOIN LATERAL generate_series(1, c.through - c.before) AS k;

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> synthetic_zones THEN
        RAISE EXCEPTION 'loaded copy: % synthetic zones written, not %.', n, synthetic_zones;
    END IF;
    RAISE NOTICE 'loaded copy: % synthetic zones over % cells of % degrees, weighted by the % zones of data/zones-dump-2025-01-04.json.',
        n, cells, cell, dump_zones;
    RETURN n;
END
$fn$;

-- EXPLAIN (ANALYZE, BUFFERS) of one query, returned line by line so the plan is
-- the statement's printed result, and asserted on that same output: a
-- sequential scan of `zone`, or a plan that never names `zone_geom_gist`,
-- raises with the plan in the message. Below 1000 rows the planner is not
-- answering the real question, so it reports unreached instead of judging.
CREATE FUNCTION pg_temp.explain_asserting_gist(shape text, sql_text text)
RETURNS SETOF text
LANGUAGE plpgsql AS $fn$
DECLARE
    line       text;
    plan       text[] := ARRAY[]::text[];
    rows_total bigint;
BEGIN
    SELECT count(*) INTO rows_total FROM public.zone;

    FOR line IN EXECUTE 'EXPLAIN (ANALYZE, BUFFERS) ' || sql_text LOOP
        plan := plan || line;
    END LOOP;

    IF rows_total < 1000 THEN
        RAISE NOTICE 'part E, %: zone holds % row(s), too few for the planner to be answering the real question. Index use is UNPROVEN. This is not a pass.', shape, rows_total;
    ELSIF EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ 'Seq Scan on (public\.)?zone( |$)') THEN
        RAISE EXCEPTION E'part E, %: the plan reads zone by a sequential scan over % rows, not by zone_geom_gist. Plan:\n%',
            shape, rows_total, array_to_string(plan, E'\n');
    ELSIF NOT EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ 'zone_geom_gist') THEN
        RAISE EXCEPTION E'part E, %: the plan never names zone_geom_gist over % rows. Plan:\n%',
            shape, rows_total, array_to_string(plan, E'\n');
    END IF;

    FOREACH line IN ARRAY plan LOOP
        RETURN NEXT line;
    END LOOP;
END
$fn$;

-- === Part A — the column set, in both directions =============================
--
-- The list below is not documentation of the table. It IS the assertion, and
-- the table is checked against it. Equality is asserted in both directions,
-- and the unusual direction is the one that matters: a test asserting that
-- these columns exist cannot fail when a column is ADDED, and a column being
-- added is precisely the event being guarded against.
--
-- The event is `current_owner` or `date_last_taken` appearing, months from
-- now, added in good faith by someone who has not read
-- `Architecture.md § The two absences, and the test that keeps them absent`.
-- Both are round-scoped. A synced table that acquires one acquires a rollover
-- problem that today it simply does not have.

DO $$
DECLARE
    expected text[] := ARRAY[
        'area_id', 'area_name', 'date_created', 'first_seen_at', 'geom', 'id',
        'last_changed_at', 'latitude', 'longitude', 'name', 'points_per_hour',
        'region_country', 'region_id', 'region_name', 'takeover_points',
        'total_takeovers', 'type_id'
    ];
    actual  text[];
    extra   text[];
    missing text[];
BEGIN
    SELECT array_agg(column_name::text ORDER BY column_name)
      INTO actual
      FROM information_schema.columns
     WHERE table_schema = 'public'
       AND table_name   = 'zone';

    IF actual IS NULL THEN
        RAISE EXCEPTION
            'table public.zone does not exist — 0001 has not been applied here, whatever else this database may hold under that name';
    END IF;

    SELECT array_agg(c ORDER BY c) INTO extra
      FROM unnest(actual) AS c WHERE c <> ALL (expected);

    SELECT array_agg(c ORDER BY c) INTO missing
      FROM unnest(expected) AS c WHERE c <> ALL (actual);

    IF extra IS NOT NULL THEN
        RAISE EXCEPTION
            'zone has gained the column(s) %, which the design does not admit. If one of them is round-scoped, the synced zone table has just acquired a round-rollover problem it did not have.',
            extra;
    END IF;

    IF missing IS NOT NULL THEN
        RAISE EXCEPTION
            'zone is missing the column(s) %, which the sync writes and the corridor query reads.',
            missing;
    END IF;
END
$$;

-- === Part B — the structural checks IF NOT EXISTS cannot make ================
--
-- `CREATE TABLE IF NOT EXISTS` on a table that already exists in a different
-- shape adds nothing and says nothing. These are the checks that notice.

DO $$
DECLARE
    gen  char;
    typ  text;
BEGIN
    SELECT a.attgenerated, format_type(a.atttypid, a.atttypmod)
      INTO gen, typ
      FROM pg_attribute a
     WHERE a.attrelid = 'public.zone'::regclass AND a.attname = 'geom';

    IF gen IS DISTINCT FROM 's' THEN
        RAISE EXCEPTION
            'zone.geom is not a STORED generated column (attgenerated=%). The POINT''S axis order — ST_MakePoint''s X-then-Y — is no longer declared once in DDL: whatever writes the row now supplies the point itself, per row, and can invert it there.',
            coalesce(gen::text, 'none')
            USING HINT = 'That is the point''s CONSTRUCTION, and it is the only half of the question this check reaches. Nothing in this file closes the write-path axis question: which value is bound to the `latitude` column and which to `longitude` is decided in the write path whether `geom` is generated or not, and is pinned in Go by `TestEveryColumnIsLoadedFromItsOwnField` in `service/internal/syncstore/columns_test.go`. See the note above part D and the header of that test''s file.';
    END IF;

    IF typ <> 'geography(Point,4326)' THEN
        RAISE EXCEPTION
            'zone.geom is % rather than geography(Point,4326). A geometry in 4326 measures in degrees, which are not a unit of distance, and an unstated SRID is one defaulted somewhere a reader cannot see.',
            typ;
    END IF;
END
$$;

-- Each name is paired with the relation 0001 declares it on, as 0001's own
-- divergence check pairs them. A set membership would let `zone_lat_range`
-- answer from `sync_run`, which is a constraint 0001 never wrote reported as
-- the one it did.
DO $$
DECLARE
    lost text[];
BEGIN
    SELECT array_agg(format('%s on %s', w.name, w.tbl) ORDER BY w.name) INTO lost
      FROM (VALUES ('zone_lat_range'::text,   'public.zone'::text),
                   ('zone_lon_range',         'public.zone'),
                   ('zone_takeovers_nonneg',  'public.zone'),
                   ('sync_run_outcome_known', 'public.sync_run')) AS w(name, tbl)
     WHERE NOT EXISTS (
             SELECT 1
               FROM pg_constraint k
              WHERE k.conname  = w.name
                AND k.contype  = 'c'
                AND k.conrelid = w.tbl::regclass);

    IF lost IS NOT NULL THEN
        RAISE EXCEPTION 'constraint(s) % are absent; 0001 is applied only in part.', lost;
    END IF;
END
$$;

DO $$
DECLARE
    valid  boolean;
    method text;
BEGIN
    -- `indrelid` is pinned as well as the index's own schema, as 0001's
    -- divergence check pins it. Name and namespace identify the index; they do
    -- not say what it indexes, so a `zone_geom_gist` sitting on some other
    -- table in `public` would answer here for the one `zone` does not have.
    SELECT i.indisvalid, am.amname
      INTO valid, method
      FROM pg_index i
      JOIN pg_class     c  ON c.oid  = i.indexrelid
      JOIN pg_am        am ON am.oid = c.relam
     WHERE c.relname      = 'zone_geom_gist'
       AND c.relnamespace = 'public'::regnamespace
       AND i.indrelid     = 'public.zone'::regclass;

    IF valid IS NULL THEN
        RAISE EXCEPTION
            'no index named zone_geom_gist exists on public.zone. The corridor query and the nearest-neighbour query both fall back to a sequential scan of the whole zone table, which is the difference between a product and a timeout.'
            USING HINT = 'An index of that name on some other relation is reported here as absent, because that is what it is to the two queries it was created to serve. 0001''s own divergence check says the same of one on the wrong column.';
    END IF;

    IF NOT valid THEN
        RAISE EXCEPTION
            'zone_geom_gist exists but is INVALID — a concurrent build failed and left it behind. The planner will not use it. It must be dropped and rebuilt.';
    END IF;

    IF method <> 'gist' THEN
        RAISE EXCEPTION 'zone_geom_gist uses the % access method rather than gist.', method;
    END IF;
END
$$;

DO $$
DECLARE
    persistence char;
BEGIN
    SELECT relpersistence INTO persistence
      FROM pg_class
     WHERE relname      = 'zone_incoming'
       AND relnamespace = 'public'::regnamespace;

    IF persistence IS NULL THEN
        RAISE EXCEPTION 'the staging table zone_incoming does not exist; the sync has nowhere to stage and nothing to assert against before it merges.';
    END IF;

    IF persistence <> 'u' THEN
        RAISE EXCEPTION
            'zone_incoming is not UNLOGGED (relpersistence=%). Every bulk load into it now writes WAL, for a table that is truncated and rebuilt on the next run.',
            persistence;
    END IF;
END
$$;

-- --- sync_run must be able to record a run that went wrong ------------------
--
-- This is the one probe that writes, and it writes nothing that survives: each
-- insert happens inside a subtransaction that is rolled back deliberately. It
-- consumes a handful of values from the sync_run id sequence, which is a gap in
-- a bigserial and nothing else.
--
-- What it proves is the property that makes sync_run worth having. A run that
-- failed before it received a single byte knows two things: when it started and
-- how it ended. If the schema demands more than that, the run cannot be
-- recorded — and a run that cannot be recorded is indistinguishable from a run
-- that never happened, which is exactly the observation an operator needs.

DO $$
DECLARE
    v text;
BEGIN
    FOREACH v IN ARRAY ARRAY['running', 'ok', 'http_error', 'assertion_failed', 'aborted']
    LOOP
        BEGIN
            INSERT INTO public.sync_run (started_at, outcome) VALUES (now(), v);
            RAISE EXCEPTION USING ERRCODE = 'ZZ001', MESSAGE = 'probe_rollback';
        EXCEPTION
            WHEN sqlstate 'ZZ001' THEN
                NULL;   -- intended; the subtransaction is discarded
            WHEN others THEN
                RAISE EXCEPTION
                    'sync_run cannot record a run whose only known facts are its start instant and outcome=%. Underlying error: %',
                    v, SQLERRM;
        END;
    END LOOP;
END
$$;

DO $$
BEGIN
    BEGIN
        INSERT INTO public.sync_run (started_at, outcome)
        VALUES (now(), 'not_a_known_outcome');
        RAISE EXCEPTION USING ERRCODE = 'ZZ002', MESSAGE = 'vocabulary_not_enforced';
    EXCEPTION
        WHEN check_violation THEN
            NULL;   -- expected: the CHECK rejected it, and the probe rolls back
        WHEN sqlstate 'ZZ002' THEN
            RAISE EXCEPTION
                'sync_run accepted an outcome value outside its vocabulary. This column is the only durable record of what a run did, so an unanticipated value is a run nobody can count.';
    END;
END
$$;

-- === Part C — the coordinate guard, three assertions =========================
--
-- The prototype this project deleted indexed zone coordinates as
-- [latitude, longitude] under a convention specifying [longitude, latitude].
-- Every spatial query it could have served would have been wrong, silently,
-- and no test it had would have said so. The full argument, the fixture, and
-- the independently computed expected values are in
-- `Architecture.md § Geometry, SRID, and the coordinate guard`.
--
-- The clone below is taken with INCLUDING GENERATED, so it carries the real
-- generation expression off the real table rather than a copy of it written
-- here. That is the whole point: what is under test is the DDL that shipped.
-- Nothing is written to `zone`, and the two fixture ids exist in live data.
--
-- WHAT PART C PROVES, AND ONLY THAT. It proves the DDL. The clone carries
-- `public.zone`'s own generation expression, the fixture hands that expression
-- a latitude and a longitude under their own column names, and the three
-- assertions measure what came back out. It proves nothing about any row the
-- sync wrote — the clone starts empty and the only two rows in it were written
-- here, by hand, this file choosing which value went to which column. Stored
-- rows are part D, and the gap that both parts leave open is stated there.

CREATE TEMP TABLE zone_guard (LIKE public.zone INCLUDING GENERATED INCLUDING CONSTRAINTS);

-- Only the coordinates are the fixture. Every other value here is filler
-- chosen to satisfy NOT NULL and to look like nothing.
INSERT INTO pg_temp.zone_guard (id, name, latitude, longitude, date_created,
                                total_takeovers, takeover_points, points_per_hour,
                                region_id, region_name, first_seen_at, last_changed_at)
VALUES
    (8240,   'VonScheeles', 59.346932, 18.021527,
     '2000-01-01T00:00:00Z', 0, 0, 0, 0, 'fixture', now(), now()),
    (119704, 'StGravkoret', 59.354872, 18.029727,
     '2000-01-01T00:00:00Z', 0, 0, 0, 0, 'fixture', now(), now());

-- --- Assertion 1 — a known distance -----------------------------------------
--
-- A millimetre tolerance also pins the spheroid at no extra cost: the same
-- pair measured on a sphere comes out several metres short, which this
-- tolerance excludes without a second assertion.

DO $$
DECLARE
    measured double precision;
    expected constant double precision := 1000.0006;
    tolerance constant double precision := 0.001;
BEGIN
    SELECT ST_Distance(a.geom, b.geom) INTO measured
      FROM pg_temp.zone_guard a, pg_temp.zone_guard b
     WHERE a.id = 8240 AND b.id = 119704;

    IF measured IS NULL THEN
        RAISE EXCEPTION 'the coordinate fixture produced no distance at all; one of its two rows is missing.';
    END IF;

    IF abs(measured - expected) > tolerance THEN
        RAISE EXCEPTION
            'the known distance between zones 8240 and 119704 measured % m, expected % m (tolerance % m). The stored geometry is not what these two coordinates mean.',
            measured, expected, tolerance;
    END IF;
END
$$;

-- --- Assertion 2 — the axis order, named ------------------------------------
--
-- This is what keeping the raw scalars alongside the derived point buys: the
-- failure message says WHICH axis, instead of reporting a distance that is
-- wrong by some factor and leaving the reader to work out why.

DO $$
DECLARE
    bad_y int;
    bad_x int;
BEGIN
    SELECT count(*) INTO bad_y FROM pg_temp.zone_guard WHERE ST_Y(geom::geometry) <> latitude;
    SELECT count(*) INTO bad_x FROM pg_temp.zone_guard WHERE ST_X(geom::geometry) <> longitude;

    IF bad_y > 0 THEN
        RAISE EXCEPTION
            'zone.geom carries a Y ordinate that is not the latitude, on % fixture row(s). Y is latitude; a swapped pair puts longitude there.',
            bad_y;
    END IF;

    IF bad_x > 0 THEN
        RAISE EXCEPTION
            'zone.geom carries an X ordinate that is not the longitude, on % fixture row(s). X is longitude; ST_MakePoint takes X first, which is the inversion this guard exists for.',
            bad_x;
    END IF;
END
$$;

-- --- Assertion 3 — the fixture is capable of failing -------------------------
--
-- Without this, the guard can be silently defanged by someone substituting a
-- better-looking pair: the corpus contains zones whose latitude and longitude
-- are within one degree of each other, and assertions 1 and 2 both pass on
-- such a pair even under a swap. A test that does not check that it can fail
-- is a test that reports success either way.
--
-- The swapped figure is also the danger stated as one number. It is not an
-- absurdity — it is a plausible distance, roughly a quarter too large, and
-- nothing in any log would look wrong.

DO $$
DECLARE
    swapped double precision;
    upright constant double precision := 1000.0006;
    expected_swapped constant double precision := 1237.1695;
    tolerance constant double precision := 0.001;
BEGIN
    -- ST_MakePoint(latitude, longitude) — deliberately inverted.
    SELECT ST_Distance(
             ST_SetSRID(ST_MakePoint(59.346932, 18.021527), 4326)::geography,
             ST_SetSRID(ST_MakePoint(59.354872, 18.029727), 4326)::geography)
      INTO swapped;

    IF abs(swapped - upright) <= tolerance THEN
        RAISE EXCEPTION
            'this fixture cannot detect a coordinate swap: swapping its axes moves the distance from % m to % m, inside the % m tolerance. Assertions 1 and 2 above are passing either way and are proving nothing.',
            upright, swapped, tolerance;
    END IF;

    IF abs(swapped - expected_swapped) > tolerance THEN
        RAISE EXCEPTION
            'the swapped fixture measured % m, expected % m. The fixture still detects a swap, but it is no longer the pair whose values were computed independently, so assertion 1 is now checking an unverified number.',
            swapped, expected_swapped;
    END IF;
END
$$;

DROP TABLE pg_temp.zone_guard;

-- === Part D — the axis order over every row actually stored ==================
--
-- Part C proves the DDL. This proves the table: on a loaded copy it is the
-- assertion that notices a row whose point does not match the scalars it was
-- derived from. An empty migrated copy is first given the synthetic load above
-- part A, inside a transaction rolled back at the end of this part, so there
-- it shows the generation expression over 154,845 rows and not the sync's.
--
-- WHAT IT CANNOT NOTICE, stated so the gap is chosen rather than assumed
-- closed. A stored point that disagrees with its own scalars is a crossed
-- generation EXPRESSION, and that is the whole of what this comparison
-- detects. It is worth having for exactly that reason: neither this part nor
-- 0001's divergence check ever reads the expression text, so behaviour is the
-- only evidence either of them holds about what builds the point.
--
-- A crossed WRITE PATH is invisible to it, and this is the direction that
-- matters most. If the longitude value is bound to the `latitude` column,
-- `geom` is generated FROM that column and moves with it: `ST_Y(geom) =
-- latitude` holds exactly as it held before, over a zone now in another
-- country. In that direction this part is tautological, and neither pinning
-- this file's `search_path` nor qualifying its relations narrows it by
-- anything at all. The binding is decided in Go, before a statement reaches a
-- server, and it is pinned there — by `TestEveryColumnIsLoadedFromItsOwnField`
-- in `service/internal/syncstore/columns_test.go`.

BEGIN;

SELECT pg_temp.load_zone_copy() AS zones_in_copy;

DO $$
DECLARE
    rows_total bigint;
    bad        bigint;
BEGIN
    SELECT count(*) INTO rows_total FROM public.zone;

    IF rows_total = 0 THEN
        RAISE NOTICE 'part D: zone is empty, so nothing was checked. This is not a pass.';
        RETURN;
    END IF;

    SELECT count(*) INTO bad
      FROM public.zone
     WHERE ST_Y(geom::geometry) <> latitude
        OR ST_X(geom::geometry) <> longitude;

    IF bad > 0 THEN
        RAISE EXCEPTION
            '% of % stored zones carry a point that disagrees with their own latitude and longitude columns.',
            bad, rows_total;
    END IF;

    RAISE NOTICE 'part D: % stored zones, all axis-consistent.', rows_total;
END
$$;

ROLLBACK;

-- The rolled-back load leaves dead tuples and a stale row estimate behind;
-- this returns the heap to the state the part found it in, so part E's plans
-- are not costed against them.
VACUUM public.zone;

-- === Part E — EXPLAIN, which is the only thing that proves an index is used ==
--
-- `CREATE INDEX` succeeding proves nothing. A corridor query falling back to a
-- sequential scan of the whole zone table is the difference between a usable
-- product and a timeout, and it fails no other check in this file.
--
-- `Architecture.md § What is unproven` item 1 makes this part an acceptance
-- condition of the first migration, against a LOADED copy, for Q1 and Q2. Q3
-- reads `plan`, which 0001 does not create; its EXPLAIN is owed by the
-- migration that creates `plan` (the same item, as amended on #142), and this
-- part reports it so rather than as a pass.
--
-- An empty copy is given the synthetic load above part A, inside a transaction
-- rolled back at the end of this part, and ANALYZEd, so the planner costs the
-- real row count rather than its default guess for a table it never measured.
-- Each plan is EXPLAIN (ANALYZE, BUFFERS), printed as the statement's result,
-- and that same output is asserted: a sequential scan of `zone` raises and
-- turns the run red, with the plan in the message.
--
-- The probe corridor starts in the densest cell of the density table, the cell
-- at (59.3, 18.0) in Stockholm with 687 real zones, and ends on that cell's
-- eastern edge. Q1 therefore returns rows, inside the 156 to 8,874 that
-- `Architecture.md § The queries the schema exists to serve` measured for Q1
-- on real routes, and Q2's lateral side runs once for each of its five
-- candidates. The row counts appear in the printed plans.
--
-- The probe distances below are probe distances. The corridor half-width the
-- product uses is chosen nowhere yet — item 7 of the same section holds it
-- open — and nothing here settles it.
--
-- WKT is longitude first, exactly as ST_MakePoint is.

BEGIN;

SELECT pg_temp.load_zone_copy() AS zones_in_copy;

ANALYZE public.zone;

SELECT plan_line AS "QUERY PLAN (Q1, corridor containment)"
  FROM pg_temp.explain_asserting_gist('Q1', $q$
        SELECT id, name, latitude, longitude, date_created, total_takeovers,
               takeover_points, points_per_hour, type_id
        FROM   public.zone
        WHERE  ST_DWithin(geom,
                          ST_GeogFromText('SRID=4326;LINESTRING(18.02 59.33, 18.10 59.37)'),
                          1000)
  $q$) AS plan_line;

SELECT plan_line AS "QUERY PLAN (Q2, activity baseline)"
  FROM pg_temp.explain_asserting_gist('Q2', $q$
        WITH candidate AS (
            SELECT id, geom
            FROM   public.zone
            WHERE  ST_DWithin(geom,
                              ST_GeogFromText('SRID=4326;LINESTRING(18.02 59.33, 18.10 59.37)'),
                              1000)
            LIMIT  5
        )
        SELECT c.id, n.id AS neighbour_id, n.total_takeovers, n.date_created
        FROM   candidate c
        CROSS  JOIN LATERAL (
                 SELECT z.id, z.total_takeovers, z.date_created
                 FROM   public.zone z
                 WHERE  ST_DWithin(z.geom, c.geom, 25000)
                   AND  z.id <> c.id
                 ORDER  BY z.geom <-> c.geom
                 LIMIT  100
               ) n
  $q$) AS plan_line;

SELECT 'owed by the migration that creates plan, per Architecture.md § What is unproven, item 1' AS "Q3, plan lookup by code";

ROLLBACK;

-- As at the end of part D: the heap returns to the state the part found it in.
VACUUM public.zone;
