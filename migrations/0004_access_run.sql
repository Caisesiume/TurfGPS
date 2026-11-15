-- 0004_access_run.sql — the run at a location the user entered
--
-- Creates the two statements the run of `FR-126` is measured by (#248):
-- `access_run_network`, the marked network a run at a location can travel, and
-- `access_run_points`, where a leg's points lie on the ways the tables hold.
-- It creates no table, no index and no rule, and alters nothing 0001 to 0003
-- created.
--
-- The design is `Architecture.md § What the tables hold` and `§ The route
-- check`, step 4, on `docs/adr/ADR-0008-access-marking-table.md`: the least
-- run is a shortest path by length over `access_way` rows sharing nodes, from
-- the point where the leg's shape meets the location's way to any node of
-- `access_way_exit`, measured as the leg's run is — both along the geometry
-- PostGIS holds, both as `ST_Length` on `geography`, and over every marked
-- way sharing a node whatever its direction of travel. These two statements
-- give `internal/accessstore` the lengths, one per segment of each way, and
-- it takes the shortest path and the leg's run from the same numbers in the
-- same order, so a leg along the least run measures exactly the least.
--
-- ----------------------------------------------------------------------------
-- A SAFETY PATH. The least run is the bound a leg's marked travel at a location
-- the user entered is held to. A change to what the network takes — which
-- ways, which exits, how a segment is measured — is a change to that bound.
-- ----------------------------------------------------------------------------
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY. CI's `migrations` job applies it after
-- 0001 to 0003, runs `0004_access_run.verify.sql` part by part and the rollback
-- twice; CI's `valhalla` job applies it to the PostGIS beside the engine. It
-- has not been applied to any database that outlives either job.
--
-- APPLY. By an authorised operator, deliberately, never at boot, per
-- `postgis-migration-protocol`:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0004_access_run.sql
--
-- ROLLBACK. `0004_access_run.down.sql`, which drops the two functions. With
-- them gone the run cannot be measured, so every leg running along a marked
-- way at a location the user entered is refused (`internal/accessstore`
-- returns the database's error, and the planner refuses an unmeasured run):
-- the stricter state.
--
-- IRREVERSIBLE STEPS. None.
--
-- IDEMPOTENCY. Both statements are `CREATE OR REPLACE`.

BEGIN;

SET LOCAL search_path = pg_catalog, public, pg_temp;

-- --- the network ------------------------------------------------------------
--
-- Every marked way reachable from start_way over marked ways sharing a node,
-- start_way among them, each with its node ids in order, the length in metres
-- of each of its segments — segment i joins node i to node i + 1 — and the
-- nodes at which an unmarked `highway` way meets it, which is where a run ends
-- (`access_way_exit`). Nothing where start_way is not a marked way the tables
-- hold: the least run is then undefined, and the caller refuses the leg.
--
-- Direction is not read: a way marked in either direction is marked (`RD-064`),
-- and the network is taken whichever way its ways may be driven, so the least
-- run over it is no longer than any run a leg can drive. A segment is
-- measured as a two-point line on `geography`, which is how `ST_Length` on a
-- way's `geography` measures it.
--
-- A single-SELECT SQL function, neither STRICT nor carrying a SET clause, with
-- its tables schema-qualified, so the planner inlines it and part C of the
-- verify file EXPLAINs the statement itself, as `access_crossings` is.

CREATE OR REPLACE FUNCTION public.access_run_network(start_way bigint)
RETURNS TABLE (
    network_way    bigint,
    network_nodes  bigint[],
    segment_metres double precision[],
    exit_nodes     bigint[]
)
LANGUAGE sql STABLE PARALLEL SAFE
AS $$
    WITH RECURSIVE component(way_id) AS (
        SELECT s.way_id
          FROM public.access_way s
         WHERE s.way_id = start_way
           AND s.marked
        UNION
        SELECT n.way_id
          FROM component c
          JOIN public.access_way w ON w.way_id = c.way_id
          JOIN public.access_way n ON n.nodes && w.nodes
         WHERE n.marked
    )
    SELECT w.way_id,
           w.nodes,
           ARRAY(SELECT ST_Length(ST_MakeLine(ST_PointN(w.geom::geometry, g.i), ST_PointN(w.geom::geometry, g.i + 1))::geography)
                   FROM generate_series(1, ST_NPoints(w.geom::geometry) - 1) AS g(i)
                  ORDER BY g.i),
           ARRAY(SELECT x.node_id
                   FROM public.access_way_exit x
                  WHERE x.way_id = w.way_id
                  ORDER BY x.node_id)
      FROM component c
      JOIN public.access_way w ON w.way_id = c.way_id
     ORDER BY w.way_id
$$;

-- --- the points -------------------------------------------------------------
--
-- For point i (1-based) of the arrays, on way point_ways[i]: the vertex of the
-- way nearest it (1-based, the lowest where two are equally near), and the
-- segment of the way nearest it (segment s joins vertex s to vertex s + 1, the
-- lowest where two are equally near), with the length in metres from the
-- segment's first vertex to the point of the segment nearest the point, and
-- from that point to its second vertex. No row for a point whose way the
-- tables do not hold, or that is not a coordinate; the caller refuses a run
-- with a point missing.
--
-- The vertex places a leg's edge ends, which the engine puts at the nodes of
-- its graph: OpenStreetMap nodes, whose positions the engine's shape rounds
-- to six decimal places. The segment places where the leg's shape meets the
-- location's way, the run's first point, which the engine snapped anywhere
-- along it.

CREATE OR REPLACE FUNCTION public.access_run_points(
    point_ways bigint[],
    point_lons double precision[],
    point_lats double precision[]
) RETURNS TABLE (
    point_index  integer,
    vertex       integer,
    segment      integer,
    metres_back  double precision,
    metres_ahead double precision
)
LANGUAGE sql STABLE PARALLEL SAFE
AS $$
    WITH p AS (
        SELECT q.ord::integer AS ord, q.way_id, ST_SetSRID(ST_MakePoint(q.lon, q.lat), 4326) AS pt
          FROM unnest(point_ways, point_lons, point_lats) WITH ORDINALITY AS q(way_id, lon, lat, ord)
         WHERE q.lon BETWEEN -180 AND 180
           AND q.lat BETWEEN -90 AND 90
    )
    SELECT p.ord, v.vertex, s.segment, s.back, s.ahead
      FROM p
      JOIN public.access_way w ON w.way_id = p.way_id
     CROSS JOIN LATERAL (
         SELECT (d.path)[1] AS vertex
           FROM ST_DumpPoints(w.geom::geometry) AS d
          ORDER BY ST_Distance(d.geom::geography, p.pt::geography), (d.path)[1]
          LIMIT 1) v
     CROSS JOIN LATERAL (
         SELECT g.i AS segment,
                ST_Length(ST_MakeLine(g.a, ST_ClosestPoint(g.line, p.pt))::geography) AS back,
                ST_Length(ST_MakeLine(ST_ClosestPoint(g.line, p.pt), g.b)::geography) AS ahead
           FROM (SELECT k.i,
                        ST_PointN(w.geom::geometry, k.i) AS a,
                        ST_PointN(w.geom::geometry, k.i + 1) AS b,
                        ST_MakeLine(ST_PointN(w.geom::geometry, k.i), ST_PointN(w.geom::geometry, k.i + 1)) AS line
                   FROM generate_series(1, ST_NPoints(w.geom::geometry) - 1) AS k(i)) g
          ORDER BY ST_Distance(g.line::geography, p.pt::geography), g.i
          LIMIT 1) s
     ORDER BY p.ord
$$;

COMMIT;
