-- 0003_access_markings.verify.sql — the assertions that make 0003 falsifiable
--
-- Run against a migrated copy, never against production:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0003_access_markings.verify.sql
--
-- Every assertion raises with a message naming what it found; a silent run is
-- a pass. It writes nothing that survives: parts C to F insert synthetic rows
-- inside a transaction each rolls back, under ids above 9e15 that no OSM
-- object carries, which consumes values from the load record's identity and
-- leaves no row behind.
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY, part by part, by CI's `migrations` job
-- (`scripts/db/migrate-verify.sh`, #153), after 0001 to 0003 are applied.
--
-- `Architecture.md § The evidence CI owes`, `migrations`, is what this file
-- answers:
--
-- Part A  the column sets of the six tables, in both directions, and the geometry types
-- Part B  the rule against every value the criteria name, each spelled as they spell it
-- Part C  the coordinate guard, on a way's length and on the check finding a barrier
-- Part D  the check: what it finds, what it does not, and every input it refuses
-- Part E  the load record: one current, and a feature lost refuses the row
-- Part F  EXPLAIN (ANALYZE, BUFFERS) of the check over a synthetic load

-- The rule over a feature given as JSON, its keys exactly as OSM spells them,
-- so a case below reads as the tags it is about.
CREATE FUNCTION pg_temp.feature_marked(tags jsonb) RETURNS boolean
LANGUAGE sql IMMUTABLE AS $fn$
    SELECT public.access_feature_marked(
        tags->>'motorcar', tags->>'motor_vehicle', tags->>'vehicle', tags->>'access',
        tags->>'motorcar:forward', tags->>'motor_vehicle:forward', tags->>'vehicle:forward', tags->>'access:forward',
        tags->>'motorcar:backward', tags->>'motor_vehicle:backward', tags->>'vehicle:backward', tags->>'access:backward')
$fn$;

-- === Part A — the column sets, in both directions ===========================
--
-- A column added to a feature table is the event worth catching: a second
-- marking column, or a key read the rule does not name, is a second home for
-- the rule.

DO $$
DECLARE
    want    record;
    actual  text[];
    extra   text[];
    missing text[];
    keys    text[] := ARRAY['access', 'access_backward', 'access_forward', 'motor_vehicle', 'motor_vehicle_backward',
                            'motor_vehicle_forward', 'motorcar', 'motorcar_backward', 'motorcar_forward', 'vehicle',
                            'vehicle_backward', 'vehicle_forward'];
BEGIN
    FOR want IN
        SELECT * FROM (VALUES
            ('access_way'::text,       keys || ARRAY['geom', 'marked', 'nodes', 'way_id']),
            ('access_barrier',         keys || ARRAY['barrier', 'geom', 'marked', 'node_id']),
            ('access_barrier_way',     ARRAY['node_id', 'way_id']),
            ('access_way_exit',        ARRAY['node_id', 'way_id']),
            ('access_area',            keys || ARRAY['geom', 'marked', 'osm_id', 'osm_type']),
            ('osm_extract_load',       ARRAY['area_features', 'area_rows', 'barrier_features', 'barrier_rows',
                                             'barrier_way_rows', 'current', 'extract_sha256', 'extract_timestamp',
                                             'id', 'loaded_at', 'way_exit_rows', 'way_features', 'way_rows'])
        ) AS w(tbl, cols)
    LOOP
        SELECT array_agg(column_name::text ORDER BY column_name) INTO actual
          FROM information_schema.columns
         WHERE table_schema = 'public' AND table_name = want.tbl;

        IF actual IS NULL THEN
            RAISE EXCEPTION 'table public.% does not exist — 0003 has not been applied here', want.tbl;
        END IF;

        SELECT array_agg(c ORDER BY c) INTO extra FROM unnest(actual) AS c WHERE c <> ALL (want.cols);
        SELECT array_agg(c ORDER BY c) INTO missing FROM unnest(want.cols) AS c WHERE c <> ALL (actual);

        IF extra IS NOT NULL OR missing IS NOT NULL THEN
            RAISE EXCEPTION '% has the extra column(s) % and lacks %, against 0003', want.tbl, extra, missing;
        END IF;
    END LOOP;
END
$$;

DO $$
DECLARE
    want record;
    typ  text;
BEGIN
    FOR want IN
        SELECT * FROM (VALUES
            ('access_way'::text,  'geography(LineString,4326)'::text),
            ('access_barrier',    'geography(Point,4326)'),
            ('access_area',       'geography(MultiPolygon,4326)')
        ) AS w(tbl, typ)
    LOOP
        SELECT format_type(a.atttypid, a.atttypmod) INTO typ
          FROM pg_attribute a
         WHERE a.attrelid = ('public.' || want.tbl)::regclass AND a.attname = 'geom';

        IF typ IS DISTINCT FROM want.typ THEN
            RAISE EXCEPTION '%.geom is %, not %', want.tbl, typ, want.typ;
        END IF;
    END LOOP;
END
$$;

-- === Part B — the rule, against every value the criteria name ===============
--
-- Each case is a feature's tags, spelled exactly as `FR-125`'s criteria, #248's
-- AC2 and AC14 to AC17, and the Owner's decisions on #258 spell them, and the
-- marking the records give it. Keys off the car chain (`bicycle`, `foot`,
-- `psv`, `motorcycle`) are written in the cases that name them though the rule
-- has no column for them: that they change nothing is the point of AC17.

DO $$
DECLARE
    c      record;
    got    boolean;
    wrong  text[] := ARRAY[]::text[];
BEGIN
    FOR c IN
        SELECT * FROM (VALUES
            -- #248 AC2: every value outside the permissive set, on access
            ('AC2 access=private',        '{"access":"private"}'::jsonb,        true),
            ('AC2 access=destination',    '{"access":"destination"}',           true),
            ('AC2 access=customers',      '{"access":"customers"}',             true),
            ('AC2 access=delivery',       '{"access":"delivery"}',              true),
            ('AC2 access=permit',         '{"access":"permit"}',                true),
            ('AC2 access=residents',      '{"access":"residents"}',             true),
            ('AC2 access=restricted',     '{"access":"restricted"}',            true),
            ('AC2 access=military',       '{"access":"military"}',              true),
            ('AC2 motor_vehicle=unknown', '{"motor_vehicle":"unknown"}',        true),
            -- #248 AC2: marked on each key other than access while access=yes
            ('AC2 motor_vehicle=private, access=yes', '{"motor_vehicle":"private","access":"yes"}', true),
            ('AC2 motorcar=private, access=yes',      '{"motorcar":"private","access":"yes"}',      true),
            ('AC2 vehicle=private, access=yes',       '{"vehicle":"private","access":"yes"}',       true),
            -- #248 AC14: a more specific key's permission is driven
            ('AC14 access=no, motor_vehicle=yes',     '{"access":"no","motor_vehicle":"yes"}',      false),
            ('AC14 access=private, motorcar=yes',     '{"access":"private","motorcar":"yes"}',      false),
            -- #248 AC15: public and allowed are driven
            ('AC15 access=public',                    '{"access":"public"}',                        false),
            ('AC15 motor_vehicle=allowed',            '{"motor_vehicle":"allowed"}',                false),
            -- #248 AC16: access=no with no more specific key stays excluded
            ('AC16 access=no',                        '{"access":"no"}',                            true),
            -- #248 AC17: a permission off the car chain lifts no marking
            ('AC17 access=private, bicycle=yes, foot=yes',      '{"access":"private","bicycle":"yes","foot":"yes"}',   true),
            ('AC17 motor_vehicle=private, psv=yes',             '{"motor_vehicle":"private","psv":"yes"}',             true),
            ('AC17 motor_vehicle=destination, motorcycle=yes',  '{"motor_vehicle":"destination","motorcycle":"yes"}',  true),
            -- RD-063: the permissive set, each value alone
            ('access=yes',                            '{"access":"yes"}',                           false),
            ('access=permissive',                     '{"access":"permissive"}',                    false),
            ('access=designated',                     '{"access":"designated"}',                    false),
            ('access=public',                         '{"access":"public"}',                        false),
            ('access=allowed',                        '{"access":"allowed"}',                       false),
            ('access=no, motor_vehicle=designated',   '{"access":"no","motor_vehicle":"designated"}', false),
            -- RD-062: the comparison is exact, and fails closed
            ('access=Yes',                            '{"access":"Yes"}',                           true),
            ('access=yes;private',                    '{"access":"yes;private"}',                   true),
            ('access="yes "',                         '{"access":"yes "}',                          true),
            ('access=""',                             '{"access":""}',                              true),
            -- no key at all: unmarked by definition
            ('no access key',                         '{}',                                         false),
            ('a key off the chain only',              '{"bicycle":"no"}',                           false),
            -- the Owner's second decision on #258: a directional key governs its direction
            ('motor_vehicle:backward=no only',        '{"motor_vehicle:backward":"no"}',            true),
            ('vehicle:backward=no on a oneway',       '{"vehicle:backward":"no"}',                  true),
            ('access:forward=yes, access:backward=yes, access=no', '{"access":"no","access:forward":"yes","access:backward":"yes"}', false),
            ('motorcar:backward=yes, access=no',      '{"access":"no","motorcar:backward":"yes"}',  true),
            ('motor_vehicle:forward=yes, access=private', '{"access":"private","motor_vehicle:forward":"yes"}', true),
            -- N8 on #258, N9 of the dispatch: direction specificity beats vehicle specificity
            ('N9 motorcar=yes, motor_vehicle:backward=no', '{"motorcar":"yes","motor_vehicle:backward":"no"}', true),
            ('N9 motorcar=yes, motor_vehicle:forward=no',  '{"motorcar":"yes","motor_vehicle:forward":"no"}',  true),
            ('N9 motorcar=yes, access:backward=private',   '{"motorcar":"yes","access:backward":"private"}',   true),
            -- among the directional forms the first non-null still governs
            ('motorcar:forward=yes, motor_vehicle:forward=no', '{"motorcar:forward":"yes","motor_vehicle:forward":"no"}', false)
        ) AS t(name, tags, want)
    LOOP
        got := pg_temp.feature_marked(c.tags);

        IF got IS DISTINCT FROM c.want THEN
            wrong := wrong || format('%s: %s, want %s', c.name, got, c.want);
        END IF;
    END LOOP;

    IF cardinality(wrong) > 0 THEN
        RAISE EXCEPTION E'the rule disagrees with its records on % case(s):\n%', cardinality(wrong), array_to_string(wrong, E'\n');
    END IF;

    IF public.access_marked(NULL, NULL, NULL, NULL) IS DISTINCT FROM false THEN
        RAISE EXCEPTION 'access_marked with no key is %, want false', public.access_marked(NULL, NULL, NULL, NULL);
    END IF;
END
$$;

-- === Part C — the coordinate guard =========================================
--
-- A way of 0.01 degrees of longitude at 57.70 N is about 596 m; read with its
-- axes swapped it is 0.01 degrees of latitude, about 1,113 m. A barrier on the
-- way is found by the check only if the check reads the same axes the table
-- holds.

BEGIN;

DO $$
DECLARE
    metres double precision;
    found  bigint;
BEGIN
    INSERT INTO public.access_way (way_id, access, geom, nodes)
    VALUES (9000000000000001, 'yes', ST_GeogFromText('SRID=4326;LINESTRING(11.96 57.70, 11.97 57.70)'), ARRAY[9000000000000101, 9000000000000102]);

    SELECT ST_Length(geom) INTO metres FROM public.access_way WHERE way_id = 9000000000000001;

    IF metres NOT BETWEEN 590 AND 600 THEN
        RAISE EXCEPTION 'the way along 57.70 N from 11.96 to 11.97 E is % m long, want about 596 m: the axes are not longitude then latitude', metres;
    END IF;

    INSERT INTO public.access_barrier (node_id, barrier, access, geom)
    VALUES (9000000000000201, 'gate', 'private', ST_SetSRID(ST_MakePoint(11.965, 57.70), 4326)::geography);
    INSERT INTO public.access_barrier_way (way_id, node_id) VALUES (9000000000000001, 9000000000000201);

    SELECT count(*) INTO found
      FROM public.access_crossings(ARRAY[9000000000000001]::bigint[], ARRAY[1, 1], ARRAY[11.96, 11.97], ARRAY[57.70, 57.70], 1.0)
     WHERE kind = 'barrier' AND osm_id = 9000000000000201;

    IF found <> 1 THEN
        RAISE EXCEPTION 'the check found the barrier at (57.70 N, 11.965 E) on the leg along it % time(s), want once', found;
    END IF;
END
$$;

ROLLBACK;

-- === Part D — the check ====================================================
--
-- A leg of three edges: along a marked way, along an unmarked way carrying
-- barriers, and along a way the tables do not hold that crosses a marked area.
-- With a tolerance of 1 m the check names the marked way at edge 1, the marked
-- barrier on the stretch at edge 2 and the area at edge 3, and nothing else:
-- not the unmarked barrier, not the marked barrier 11 m off the stretch, and
-- not the marked barrier on the stretch whose way is another. At 20 m it names
-- the barrier 11 m off too. Every input it cannot read whole returns a row of
-- kind `unchecked`.

BEGIN;

INSERT INTO public.access_way (way_id, access, geom, nodes) VALUES
    (9000000000000001, 'private', ST_GeogFromText('SRID=4326;LINESTRING(11.96 57.70, 11.97 57.70)'), ARRAY[9000000000000101, 9000000000000102]),
    (9000000000000002, 'yes',     ST_GeogFromText('SRID=4326;LINESTRING(11.97 57.70, 11.98 57.70)'), ARRAY[9000000000000102, 9000000000000103]);

INSERT INTO public.access_barrier (node_id, barrier, access, geom) VALUES
    (9000000000000211, 'gate',    'private', ST_SetSRID(ST_MakePoint(11.975, 57.70),   4326)::geography),
    (9000000000000212, 'gate',    'yes',     ST_SetSRID(ST_MakePoint(11.976, 57.70),   4326)::geography),
    (9000000000000213, 'bollard', 'private', ST_SetSRID(ST_MakePoint(11.975, 57.7001), 4326)::geography),
    (9000000000000214, NULL,      'private', ST_SetSRID(ST_MakePoint(11.978, 57.70),   4326)::geography);

INSERT INTO public.access_barrier_way (way_id, node_id) VALUES
    (9000000000000002, 9000000000000211),
    (9000000000000002, 9000000000000212),
    (9000000000000002, 9000000000000213),
    (9000000000000009, 9000000000000214);

INSERT INTO public.access_area (osm_type, osm_id, access, geom) VALUES
    ('w', 9000000000000301, 'private',
     ST_GeogFromText('SRID=4326;MULTIPOLYGON(((11.984 57.6995, 11.986 57.6995, 11.986 57.7005, 11.984 57.7005, 11.984 57.6995)))'));

DO $$
DECLARE
    ways   bigint[] := ARRAY[9000000000000001, 9000000000000002, 9000000000000005];
    edges  integer[] := ARRAY[1, 1, 2, 2, 3, 3];
    lons   double precision[] := ARRAY[11.96, 11.97, 11.97, 11.98, 11.98, 11.99];
    lats   double precision[] := ARRAY[57.70, 57.70, 57.70, 57.70, 57.70, 57.70];
    got    text;
    c      record;
BEGIN
    SELECT string_agg(format('%s:%s:%s%s', edge_index, kind, osm_type, osm_id), ' ' ORDER BY edge_index, kind, osm_id) INTO got
      FROM public.access_crossings(ways, edges, lons, lats, 1.0);

    IF got IS DISTINCT FROM '1:way:w9000000000000001 2:barrier:n9000000000000211 3:area:w9000000000000301' THEN
        RAISE EXCEPTION 'at 1 m the check found [%], want the marked way at edge 1, the marked barrier on the stretch at edge 2 and the area at edge 3', got;
    END IF;

    SELECT string_agg(format('%s:%s:%s%s', edge_index, kind, osm_type, osm_id), ' ' ORDER BY edge_index, kind, osm_id) INTO got
      FROM public.access_crossings(ways, edges, lons, lats, 20.0);

    IF got IS DISTINCT FROM '1:way:w9000000000000001 2:barrier:n9000000000000211 2:barrier:n9000000000000213 3:area:w9000000000000301' THEN
        RAISE EXCEPTION 'at 20 m the check found [%], want the barrier 11 m off the stretch as well', got;
    END IF;

    -- A barrier's position comes back with it, longitude and latitude by name.
    SELECT longitude, latitude INTO c FROM public.access_crossings(ways, edges, lons, lats, 1.0) WHERE kind = 'barrier';

    IF abs(c.longitude - 11.975) > 1e-9 OR abs(c.latitude - 57.70) > 1e-9 THEN
        RAISE EXCEPTION 'the barrier is reported at (%, %), want longitude 11.975 and latitude 57.70', c.longitude, c.latitude;
    END IF;

    FOR c IN
        SELECT * FROM (VALUES
            ('a null tolerance',          ways, edges, lons, lats, NULL::double precision),
            ('a negative tolerance',      ways, edges, lons, lats, -1.0),
            ('a NaN tolerance',           ways, edges, lons, lats, 'NaN'::double precision),
            ('an infinite tolerance',     ways, edges, lons, lats, 'Infinity'::double precision),
            ('no edges',                  ARRAY[]::bigint[], ARRAY[]::integer[], ARRAY[]::double precision[], ARRAY[]::double precision[], 1.0),
            ('a point without a latitude', ways, edges, lons, lats[1:5], 1.0),
            ('an edge without points',    ways || 9000000000000006::bigint, edges, lons, lats, 1.0),
            ('an edge of one point',      ways, ARRAY[1, 1, 2, 2, 3, 2], lons, lats, 1.0),
            ('an edge without a way',     ARRAY[9000000000000001, NULL, 9000000000000005]::bigint[], edges, lons, lats, 1.0),
            ('a point of no edge',        ways, ARRAY[1, 1, 2, 2, 3, 4], lons, lats, 1.0)
        ) AS t(name, w, e, x, y, tol)
    LOOP
        IF NOT EXISTS (SELECT 1 FROM public.access_crossings(c.w, c.e, c.x, c.y, c.tol) WHERE kind = 'unchecked') THEN
            RAISE EXCEPTION 'the check given % returned no unchecked row: it would pass a leg it could not read whole', c.name;
        END IF;
    END LOOP;
END
$$;

-- The Owner's first and third tag decisions on #258, at the table: a node
-- carrying access=private and no barrier key is a marked barrier, and a
-- feature carrying none of the twelve keys — one tagged only
-- motor_vehicle:conditional, which the loader does not read — is not stored.
DO $$
BEGIN
    IF NOT (SELECT marked FROM public.access_barrier WHERE node_id = 9000000000000214) THEN
        RAISE EXCEPTION 'a node tagged access=private with no barrier key is not a marked barrier';
    END IF;

    BEGIN
        INSERT INTO public.access_way (way_id, geom, nodes)
        VALUES (9000000000000007, ST_GeogFromText('SRID=4326;LINESTRING(11.96 57.71, 11.97 57.71)'), ARRAY[1, 2]);
        RAISE EXCEPTION 'a way carrying no access key was stored';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END
$$;

ROLLBACK;

-- === Part E — the load record ==============================================

BEGIN;

DO $$
BEGIN
    INSERT INTO public.osm_extract_load (extract_sha256, way_features, way_rows, barrier_features, barrier_rows, barrier_way_rows, way_exit_rows, current)
    VALUES (repeat('a', 64), 3, 3, 2, 2, 2, 1, true);

    BEGIN
        INSERT INTO public.osm_extract_load (extract_sha256, way_features, way_rows, barrier_features, barrier_rows, barrier_way_rows, way_exit_rows, current)
        VALUES (repeat('b', 64), 3, 3, 2, 2, 2, 1, true);
        RAISE EXCEPTION 'a second load was recorded as current';
    EXCEPTION WHEN unique_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO public.osm_extract_load (extract_sha256, way_features, way_rows, barrier_features, barrier_rows, barrier_way_rows, way_exit_rows)
        VALUES (repeat('c', 64), 3, 2, 2, 2, 2, 1);
        RAISE EXCEPTION 'a load that lost a way was recorded';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;

    BEGIN
        INSERT INTO public.osm_extract_load (extract_sha256, way_features, way_rows, barrier_features, barrier_rows, barrier_way_rows, way_exit_rows)
        VALUES ('ABC', 0, 0, 0, 0, 0, 0);
        RAISE EXCEPTION 'a load was recorded under an extract identity that is not a SHA-256 in lower-case hex';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END
$$;

ROLLBACK;

-- === Part F — EXPLAIN, which is the only thing that proves an index is used =
--
-- `§ The access-marking indexes`: none of them is proved until EXPLAIN
-- (ANALYZE, BUFFERS) on step 3's statement shows each used. A synthetic load —
-- 200,000 ways on a grid, a tenth marked, a barrier on every fourth, 5,000
-- marked areas — is written inside a transaction rolled back at the end of
-- this part and ANALYZEd, and the check is EXPLAINed for a leg of twenty
-- edges along twenty of the ways. The plan must name access_way_pkey,
-- access_barrier_way_pkey, access_barrier_pkey and access_area_geom_gist,
-- scan none of the four tables sequentially, and show the statement itself:
-- a `Function Scan on access_crossings` means the function was not inlined and
-- the plan proves nothing.
--
-- The leg's arrays are written into the statement as literals, so the planner
-- sees what the service's parameters give it. The run's query and the
-- stopping-position queries that access_way_exit_pkey, access_way_nodes_gin
-- and access_way_geom_gist serve are #248's and #256's to EXPLAIN.

CREATE FUNCTION pg_temp.explain_crossings() RETURNS SETOF text
LANGUAGE plpgsql AS $fn$
DECLARE
    ways  bigint[];
    edges integer[];
    lons  double precision[];
    lats  double precision[];
    sql   text;
    line  text;
    plan  text[] := ARRAY[]::text[];
    idx   text;
BEGIN
    SELECT array_agg(9000000000000000 + i ORDER BY i) INTO ways FROM generate_series(500, 519) AS i;

    SELECT array_agg(e ORDER BY e, k), array_agg(11.0 + (e % 1000) * 0.001 + k * 0.0008 ORDER BY e, k),
           array_agg(57.0 + (e / 1000) * 0.001 ORDER BY e, k)
      INTO edges, lons, lats
      FROM generate_series(500, 519) AS e, generate_series(0, 1) AS k;

    -- The edges are numbered from 1 in the statement.
    SELECT array_agg(x - 499 ORDER BY o) INTO edges FROM unnest(edges) WITH ORDINALITY AS u(x, o);

    sql := format('SELECT * FROM public.access_crossings(%L::bigint[], %L::integer[], %L::double precision[], %L::double precision[], 1.0)',
                  ways, edges, lons, lats);

    FOR line IN EXECUTE 'EXPLAIN (ANALYZE, BUFFERS) ' || sql LOOP
        plan := plan || line;
    END LOOP;

    IF EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ 'Function Scan on access_crossings') THEN
        RAISE EXCEPTION E'part F: the check was not inlined, so its plan is not shown. Plan:\n%', array_to_string(plan, E'\n');
    END IF;

    IF EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ 'Seq Scan on (public\.)?access_(way|barrier|barrier_way|area)( |$)') THEN
        RAISE EXCEPTION E'part F: the check scans a feature table sequentially. Plan:\n%', array_to_string(plan, E'\n');
    END IF;

    FOREACH idx IN ARRAY ARRAY['access_way_pkey', 'access_barrier_way_pkey', 'access_barrier_pkey', 'access_area_geom_gist'] LOOP
        IF NOT EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ idx) THEN
            RAISE EXCEPTION E'part F: the plan never names %. Plan:\n%', idx, array_to_string(plan, E'\n');
        END IF;
    END LOOP;

    FOREACH line IN ARRAY plan LOOP
        RETURN NEXT line;
    END LOOP;
END
$fn$;

BEGIN;

INSERT INTO public.access_way (way_id, access, geom, nodes)
SELECT 9000000000000000 + i,
       CASE WHEN i % 10 = 0 THEN 'private' ELSE 'yes' END,
       ST_SetSRID(ST_MakeLine(ST_MakePoint(11.0 + (i % 1000) * 0.001, 57.0 + (i / 1000) * 0.001),
                              ST_MakePoint(11.0 + (i % 1000) * 0.001 + 0.0008, 57.0 + (i / 1000) * 0.001)), 4326)::geography,
       ARRAY[9000000000000000 + 2 * i, 9000000000000000 + 2 * i + 1]
  FROM generate_series(1, 200000) AS i;

INSERT INTO public.access_barrier (node_id, barrier, access, geom)
SELECT 9000000000000000 + 2 * i, 'gate', CASE WHEN i % 8 = 0 THEN 'private' ELSE 'yes' END,
       ST_SetSRID(ST_MakePoint(11.0 + (i % 1000) * 0.001, 57.0 + (i / 1000) * 0.001), 4326)::geography
  FROM generate_series(4, 200000, 4) AS i;

INSERT INTO public.access_barrier_way (way_id, node_id)
SELECT 9000000000000000 + i, 9000000000000000 + 2 * i FROM generate_series(4, 200000, 4) AS i;

INSERT INTO public.access_area (osm_type, osm_id, access, geom)
SELECT 'w', 9000000000000000 + i, 'private',
       ST_Multi(ST_Expand(ST_SetSRID(ST_MakePoint(11.0 + (i * 37 % 1000) * 0.001, 57.0 + (i * 53 % 200) * 0.001), 4326), 0.0002))::geography
  FROM generate_series(1, 5000) AS i;

ANALYZE public.access_way;
ANALYZE public.access_barrier;
ANALYZE public.access_barrier_way;
ANALYZE public.access_area;

SELECT plan_line AS "QUERY PLAN (access_crossings, a leg of twenty edges)" FROM pg_temp.explain_crossings() AS plan_line;

ROLLBACK;
