-- 0004_access_run.verify.sql — the assertions that make 0004 falsifiable
--
-- Run against a migrated copy, never against production:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0004_access_run.verify.sql
--
-- Every assertion raises with a message naming what it found; a silent run is
-- a pass. It writes nothing that survives: each part inserts synthetic rows
-- inside a transaction it rolls back, under ids above 9e15 that no OSM object
-- carries.
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY, part by part, by CI's `migrations` job
-- (`scripts/db/migrate-verify.sh`, #153), after 0001 to 0004 are applied.
--
-- Part A  the network: which ways it takes, their segments, their exits, and the coordinate guard
-- Part B  the points: the vertex and the segment a point lies nearest, and the lengths either side
-- Part C  EXPLAIN (ANALYZE, BUFFERS) of both over a synthetic load: the run's query, which
--         0003's part F left to #248 (`Architecture.md § The access-marking indexes`)

-- A small network, in a line east along 57° N from 11.000° E, every node 0.001°
-- of longitude (60.77 m at this latitude) from the next:
--
--   w1 marked     n1 — n2 — n3      the location's way
--   w2 marked               n3 — n4 its continuation, meeting w3 at n4
--   w3 not marked                n4 — n5
--   w5 marked                         n5 — n8   reachable only across w3
--   w4 marked     n6 — n7 at 57.01° N, joined to nothing
--
-- and one exit: (w2, n4), where w3 meets w2.
CREATE FUNCTION pg_temp.small_network() RETURNS void
LANGUAGE sql AS $fn$
    INSERT INTO public.access_way (way_id, access, geom, nodes) VALUES
        (9000000000000001, 'private',
         ST_GeogFromText('SRID=4326;LINESTRING(11.000 57.0, 11.001 57.0, 11.002 57.0)'),
         ARRAY[9000000000000101, 9000000000000102, 9000000000000103]),
        (9000000000000002, 'destination',
         ST_GeogFromText('SRID=4326;LINESTRING(11.002 57.0, 11.003 57.0)'),
         ARRAY[9000000000000103, 9000000000000104]),
        (9000000000000003, 'yes',
         ST_GeogFromText('SRID=4326;LINESTRING(11.003 57.0, 11.004 57.0)'),
         ARRAY[9000000000000104, 9000000000000105]),
        (9000000000000004, 'private',
         ST_GeogFromText('SRID=4326;LINESTRING(11.000 57.01, 11.001 57.01)'),
         ARRAY[9000000000000106, 9000000000000107]),
        (9000000000000005, 'private',
         ST_GeogFromText('SRID=4326;LINESTRING(11.004 57.0, 11.005 57.0)'),
         ARRAY[9000000000000105, 9000000000000108]);
    INSERT INTO public.access_way_exit (way_id, node_id) VALUES (9000000000000002, 9000000000000104);
$fn$;

-- === Part A — the network ===================================================
--
-- From w1: w1 and w2, which share n3, and neither w3, which is not marked, nor
-- w5, which only w3 reaches, nor w4, which nothing reaches. Each segment is
-- 60.77 m, which is the coordinate guard: with the axes swapped the same
-- numbers are 0.001° of latitude at 11° N, 110.6 m. From a way not marked,
-- and from one the tables do not hold, nothing: the least run is undefined.

BEGIN;

SELECT pg_temp.small_network();

DO $$
DECLARE
    got record;
    ways bigint[];
BEGIN
    SELECT array_agg(network_way ORDER BY network_way) INTO ways FROM public.access_run_network(9000000000000001);
    IF ways IS DISTINCT FROM ARRAY[9000000000000001, 9000000000000002]::bigint[] THEN
        RAISE EXCEPTION 'part A: the network from w1 is %, want w1 and w2 alone — marked ways sharing a node, and no other', ways;
    END IF;

    FOR got IN SELECT * FROM public.access_run_network(9000000000000001) LOOP
        IF got.network_way = 9000000000000001 THEN
            IF got.network_nodes IS DISTINCT FROM ARRAY[9000000000000101, 9000000000000102, 9000000000000103]::bigint[] THEN
                RAISE EXCEPTION 'part A: w1''s nodes are %, want n1, n2, n3 in order', got.network_nodes;
            END IF;
            IF cardinality(got.segment_metres) <> 2
               OR NOT (got.segment_metres[1] BETWEEN 60.5 AND 61.0 AND got.segment_metres[2] BETWEEN 60.5 AND 61.0) THEN
                RAISE EXCEPTION 'part A: w1''s segments are % m, want two of 60.77 m — the coordinate guard: 110.6 m is the axes swapped', got.segment_metres;
            END IF;
            IF cardinality(got.exit_nodes) <> 0 THEN
                RAISE EXCEPTION 'part A: w1 has the exits %, want none', got.exit_nodes;
            END IF;
        ELSE
            IF got.exit_nodes IS DISTINCT FROM ARRAY[9000000000000104]::bigint[] THEN
                RAISE EXCEPTION 'part A: w2 has the exits %, want n4, where w3 meets it', got.exit_nodes;
            END IF;
            IF cardinality(got.segment_metres) <> 1 OR NOT (got.segment_metres[1] BETWEEN 60.5 AND 61.0) THEN
                RAISE EXCEPTION 'part A: w2''s segments are % m, want one of 60.77 m', got.segment_metres;
            END IF;
        END IF;
    END LOOP;

    IF EXISTS (SELECT 1 FROM public.access_run_network(9000000000000003)) THEN
        RAISE EXCEPTION 'part A: a network is returned from w3, which is not marked; want none, so the least run is undefined';
    END IF;

    IF EXISTS (SELECT 1 FROM public.access_run_network(9000000000000999)) THEN
        RAISE EXCEPTION 'part A: a network is returned from a way the tables do not hold; want none';
    END IF;

    IF EXISTS (SELECT 1 FROM public.access_run_network(NULL)) THEN
        RAISE EXCEPTION 'part A: a network is returned from no way at all; want none';
    END IF;
END
$$;

ROLLBACK;

-- === Part B — the points ====================================================
--
-- On w1, a point 0.0004° east of n1: nearest n1 (vertex 1), on segment 1,
-- 24.31 m from n1 and 36.46 m from n2. A point at n2: vertex 2, and on both
-- segments at once, so the lowest, segment 1, 60.77 m from n1 and none from
-- n2. On w2, a point 0.0006° east of n3: nearest n4 (vertex 2). A point on a
-- way the tables do not hold, and one that is not a coordinate, give no row.

BEGIN;

SELECT pg_temp.small_network();

DO $$
DECLARE
    got record;
    n   integer;
BEGIN
    SELECT count(*) INTO n
      FROM public.access_run_points(
               ARRAY[9000000000000001, 9000000000000001, 9000000000000002, 9000000000000999, 9000000000000001]::bigint[],
               ARRAY[11.0004, 11.001, 11.0026, 11.0, 'NaN']::double precision[],
               ARRAY[57.0, 57.0, 57.0, 57.0, 57.0]::double precision[]);
    IF n <> 3 THEN
        RAISE EXCEPTION 'part B: % rows for five points, want three — none for the way not held, none for the point that is not a coordinate', n;
    END IF;

    FOR got IN
        SELECT *
          FROM public.access_run_points(
                   ARRAY[9000000000000001, 9000000000000001, 9000000000000002]::bigint[],
                   ARRAY[11.0004, 11.001, 11.0026]::double precision[],
                   ARRAY[57.0, 57.0, 57.0]::double precision[])
    LOOP
        CASE got.point_index
        WHEN 1 THEN
            IF got.vertex <> 1 OR got.segment <> 1
               OR NOT (got.metres_back BETWEEN 24.2 AND 24.4) OR NOT (got.metres_ahead BETWEEN 36.3 AND 36.6) THEN
                RAISE EXCEPTION 'part B: the point east of n1 gave vertex %, segment %, % m back and % m ahead; want 1, 1, 24.31 and 36.46',
                    got.vertex, got.segment, got.metres_back, got.metres_ahead;
            END IF;
        WHEN 2 THEN
            IF got.vertex <> 2 OR got.segment <> 1 OR NOT (got.metres_back BETWEEN 60.5 AND 61.0) OR got.metres_ahead <> 0 THEN
                RAISE EXCEPTION 'part B: the point at n2 gave vertex %, segment %, % m back and % m ahead; want 2, 1, 60.77 and 0',
                    got.vertex, got.segment, got.metres_back, got.metres_ahead;
            END IF;
        WHEN 3 THEN
            IF got.vertex <> 2 OR got.segment <> 1 THEN
                RAISE EXCEPTION 'part B: the point on w2 gave vertex % and segment %, want vertex 2, n4, and segment 1', got.vertex, got.segment;
            END IF;
        ELSE
            RAISE EXCEPTION 'part B: a row for point %, which was not asked for', got.point_index;
        END CASE;
    END LOOP;
END
$$;

ROLLBACK;

-- === Part C — EXPLAIN, which is the only thing that proves an index is used =
--
-- A synthetic load of 200,000 ways in chains of ten sharing nodes, the chains
-- joined to nothing, one chain in ten marked and given an exit at its end, is
-- written inside a transaction rolled back at the end of this part and
-- ANALYZEd. The network from a marked chain's first way must name
-- access_way_pkey, access_way_nodes_gin and access_way_exit_pkey, and the
-- points on it access_way_pkey; neither may scan access_way or
-- access_way_exit sequentially, nor show its function unexpanded: a
-- `Function Scan on access_run_` means it was not inlined, and the plan
-- proves nothing.

CREATE FUNCTION pg_temp.explain_run(sql text, wanted text[]) RETURNS SETOF text
LANGUAGE plpgsql AS $fn$
DECLARE
    line text;
    plan text[] := ARRAY[]::text[];
    idx  text;
BEGIN
    FOR line IN EXECUTE 'EXPLAIN (ANALYZE, BUFFERS) ' || sql LOOP
        plan := plan || line;
    END LOOP;

    IF EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ 'Function Scan on access_run_') THEN
        RAISE EXCEPTION E'part C: % was not inlined, so its plan is not shown. Plan:\n%', sql, array_to_string(plan, E'\n');
    END IF;

    IF EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ 'Seq Scan on (public\.)?access_way(_exit)?( |$)') THEN
        RAISE EXCEPTION E'part C: % scans a feature table sequentially. Plan:\n%', sql, array_to_string(plan, E'\n');
    END IF;

    FOREACH idx IN ARRAY wanted LOOP
        IF NOT EXISTS (SELECT 1 FROM unnest(plan) AS l WHERE l ~ idx) THEN
            RAISE EXCEPTION E'part C: the plan of % never names %. Plan:\n%', sql, idx, array_to_string(plan, E'\n');
        END IF;
    END LOOP;

    FOREACH line IN ARRAY plan LOOP
        RETURN NEXT line;
    END LOOP;
END
$fn$;

BEGIN;

-- Way i is the (i - 1) % 10'th of chain (i - 1) / 10, from node c*20 + k to
-- the next, so the ten ways of a chain share nodes and no two chains do.
INSERT INTO public.access_way (way_id, access, geom, nodes)
SELECT 9000000000000000 + i,
       CASE WHEN ((i - 1) / 10) % 10 = 0 THEN 'private' ELSE 'yes' END,
       ST_SetSRID(ST_MakeLine(
           ST_MakePoint(11.0 + (((i - 1) / 10) % 1000) * 0.01 + ((i - 1) % 10) * 0.0005, 57.0 + ((i - 1) / 10000) * 0.001),
           ST_MakePoint(11.0 + (((i - 1) / 10) % 1000) * 0.01 + ((i - 1) % 10 + 1) * 0.0005, 57.0 + ((i - 1) / 10000) * 0.001)), 4326)::geography,
       ARRAY[9000000000000000 + ((i - 1) / 10) * 20 + (i - 1) % 10, 9000000000000000 + ((i - 1) / 10) * 20 + (i - 1) % 10 + 1]
  FROM generate_series(1, 200000) AS i;

INSERT INTO public.access_way_exit (way_id, node_id)
SELECT 9000000000000000 + c * 10 + 10, 9000000000000000 + c * 20 + 10
  FROM generate_series(0, 19999) AS c
 WHERE c % 10 = 0;

ANALYZE public.access_way;
ANALYZE public.access_way_exit;

DO $$
DECLARE
    ways bigint[];
BEGIN
    SELECT array_agg(network_way ORDER BY network_way) INTO ways FROM public.access_run_network(9000000000000001);
    IF cardinality(ways) IS DISTINCT FROM 10 THEN
        RAISE EXCEPTION 'part C: the network from the first marked chain holds % ways, want its ten — the load is not what this part EXPLAINs', cardinality(ways);
    END IF;
END
$$;

SELECT plan_line AS "QUERY PLAN (access_run_network, a marked chain of ten ways in 200,000)"
  FROM pg_temp.explain_run('SELECT * FROM public.access_run_network(9000000000000001)',
                           ARRAY['access_way_pkey', 'access_way_nodes_gin', 'access_way_exit_pkey']) AS plan_line;

SELECT plan_line AS "QUERY PLAN (access_run_points, a run's four points)"
  FROM pg_temp.explain_run(
           'SELECT * FROM public.access_run_points(ARRAY[9000000000000001, 9000000000000001, 9000000000000002, 9000000000000002]::bigint[], '
           || 'ARRAY[11.0001, 11.0005, 11.0005, 11.001]::double precision[], ARRAY[57.0, 57.0, 57.0, 57.0]::double precision[])',
           ARRAY['access_way_pkey']) AS plan_line;

ROLLBACK;
