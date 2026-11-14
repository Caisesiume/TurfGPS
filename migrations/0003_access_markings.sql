-- 0003_access_markings.sql — the access-marking tables
--
-- Creates the rule `access_marked` and its per-feature form
-- `access_feature_marked`, the feature tables `access_way`, `access_barrier`,
-- `access_barrier_way`, `access_way_exit` and `access_area`, the load record
-- `osm_extract_load`, the route check `access_crossings`, and the indexes of
-- `Architecture.md § The access-marking indexes` (#255).
--
-- The design is `Architecture.md § The access-marking tables`, on
-- `docs/adr/ADR-0008-access-marking-table.md`. What is marked is not owned
-- here either: it is `FR-125` to `FR-128` with `DECISIONS.md § RD-059`,
-- `RD-060`, `RD-062` and `RD-063`, and the Owner's three tag decisions on #258
-- recorded in that section. Where this file and those records differ, the
-- records govern and this file is the defect.
--
-- ----------------------------------------------------------------------------
-- A SAFETY PATH. `access_marked` and `access_feature_marked` are the rule every
-- car leg is checked against, and `access_crossings` is the check. A change to
-- the permissive set, to the order of the keys, or to how the directional
-- forms combine is a change to a safety rule and goes to the Owner.
-- ----------------------------------------------------------------------------
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY. CI's `migrations` job
-- (`scripts/db/migrate-verify.sh`, #153) applies it after 0001 and 0002, runs
-- `0003_access_markings.verify.sql` part by part and the rollback twice; CI's
-- `valhalla` job applies it to the PostGIS beside the engine and loads the
-- committed test extract into it. It has not been applied to any database that
-- outlives either job.
--
-- APPLY. By an authorised operator, deliberately, never at boot, per
-- `postgis-migration-protocol`:
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0003_access_markings.sql
--
-- ROLLBACK. `0003_access_markings.down.sql`. DESTRUCTIVE of the loaded
-- markings only, which the loader rebuilds from the extract; with the tables
-- gone the service plans no journey (`Architecture.md § Loading from the
-- extract the tiles are built from`), which is the safe state.
--
-- IRREVERSIBLE STEPS. None.
--
-- IDEMPOTENCY. Every statement is `IF NOT EXISTS` or `CREATE OR REPLACE`, and
-- the check at the end asserts the postcondition `CREATE TABLE IF NOT EXISTS`
-- cannot: that `marked` is generated on each feature table, so a table left in
-- another shape fails the apply rather than passing silently.

BEGIN;

SET LOCAL search_path = pg_catalog, public, pg_temp;

CREATE EXTENSION IF NOT EXISTS postgis;

-- --- the rule ---------------------------------------------------------------
--
-- `RD-063`, spelled once: the first non-null of the four keys, in this order,
-- governs, and the feature is marked unless that value is one of the five. A
-- feature carrying none of the four is unmarked. The comparison is exact text
-- equality: `Yes`, `yes;private` and `yes ` are values outside the set, and a
-- value the rule does not name is a marking (`RD-062`).

CREATE OR REPLACE FUNCTION public.access_marked(
    motorcar      text,
    motor_vehicle text,
    vehicle       text,
    access        text
) RETURNS boolean
LANGUAGE sql IMMUTABLE PARALLEL SAFE
RETURN coalesce(
    coalesce(motorcar, motor_vehicle, vehicle, access)
        NOT IN ('yes', 'permissive', 'designated', 'public', 'allowed'),
    false);

-- The rule over a feature's twelve access keys, as `§ What the tables hold`
-- composes it: per direction, once with each key's directional form where the
-- feature carries one and its plain form otherwise, and once with the
-- directional forms alone — the Owner's N8 on #258, direction specificity
-- beats vehicle specificity — and marked where any of the four calls is.
-- Until a leg's direction along a way is known, a feature marked in either
-- direction is marked in both. It is one function so the three tables' columns
-- cannot drift apart.

CREATE OR REPLACE FUNCTION public.access_feature_marked(
    motorcar               text, motor_vehicle          text, vehicle          text, access          text,
    motorcar_forward       text, motor_vehicle_forward  text, vehicle_forward  text, access_forward  text,
    motorcar_backward      text, motor_vehicle_backward text, vehicle_backward text, access_backward text
) RETURNS boolean
LANGUAGE sql IMMUTABLE PARALLEL SAFE
RETURN public.access_marked(coalesce(motorcar_forward, motorcar), coalesce(motor_vehicle_forward, motor_vehicle),
                            coalesce(vehicle_forward, vehicle), coalesce(access_forward, access))
    OR public.access_marked(motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward)
    OR public.access_marked(coalesce(motorcar_backward, motorcar), coalesce(motor_vehicle_backward, motor_vehicle),
                            coalesce(vehicle_backward, vehicle), coalesce(access_backward, access))
    OR public.access_marked(motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward);

-- --- the feature tables -----------------------------------------------------
--
-- Each holds a feature carrying any of the twelve access keys, whatever their
-- values, each value exactly as the extract holds it; the CHECK refuses a row
-- carrying none, which is unmarked by definition and not stored. `marked` is
-- generated so the loader never computes it and cannot compute it wrongly.
-- Geometry is geography with SRID 4326 written explicitly; the coordinate guard
-- is part C of the verify file.

CREATE TABLE IF NOT EXISTS public.access_way (
    way_id                 bigint PRIMARY KEY CHECK (way_id > 0),
    motorcar               text, motor_vehicle          text, vehicle          text, access          text,
    motorcar_forward       text, motor_vehicle_forward  text, vehicle_forward  text, access_forward  text,
    motorcar_backward      text, motor_vehicle_backward text, vehicle_backward text, access_backward text,
    geom                   geography(LineString, 4326) NOT NULL,
    nodes                  bigint[] NOT NULL CHECK (cardinality(nodes) >= 2),
    marked                 boolean NOT NULL GENERATED ALWAYS AS (public.access_feature_marked(
                               motorcar, motor_vehicle, vehicle, access,
                               motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
                               motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward)) STORED,
    CHECK (num_nonnulls(motorcar, motor_vehicle, vehicle, access,
                        motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
                        motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward) > 0)
);

-- Every node carrying the keys, with a `barrier` key of any value or none: the
-- Owner's first tag decision on #258 counts a node carrying a non-permissive
-- value without a `barrier` key as a barrier, so `marked` does not read it.
CREATE TABLE IF NOT EXISTS public.access_barrier (
    node_id                bigint PRIMARY KEY CHECK (node_id > 0),
    barrier                text,
    motorcar               text, motor_vehicle          text, vehicle          text, access          text,
    motorcar_forward       text, motor_vehicle_forward  text, vehicle_forward  text, access_forward  text,
    motorcar_backward      text, motor_vehicle_backward text, vehicle_backward text, access_backward text,
    geom                   geography(Point, 4326) NOT NULL,
    marked                 boolean NOT NULL GENERATED ALWAYS AS (public.access_feature_marked(
                               motorcar, motor_vehicle, vehicle, access,
                               motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
                               motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward)) STORED,
    CHECK (num_nonnulls(motorcar, motor_vehicle, vehicle, access,
                        motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
                        motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward) > 0)
);

-- Each pair of such a barrier and a way referencing it, whether or not the way
-- is in `access_way`: the route check finds the barriers on a traversed way
-- through it.
CREATE TABLE IF NOT EXISTS public.access_barrier_way (
    way_id                 bigint NOT NULL CHECK (way_id > 0),
    node_id                bigint NOT NULL REFERENCES public.access_barrier (node_id),
    PRIMARY KEY (way_id, node_id)
);

-- Each node of a marked way where a `highway` way not marked meets it: where
-- `FR-126`'s run can end (#248 reads it).
CREATE TABLE IF NOT EXISTS public.access_way_exit (
    way_id                 bigint NOT NULL REFERENCES public.access_way (way_id),
    node_id                bigint NOT NULL CHECK (node_id > 0),
    PRIMARY KEY (way_id, node_id)
);

-- Area features carrying the keys. Which features it takes, and their loading,
-- are #253's (`FR-128`, `RD-060`); this migration creates it so the route check
-- reads it from the start, and the loader of #255 leaves it empty and records
-- that it did not load it.
CREATE TABLE IF NOT EXISTS public.access_area (
    osm_type               char(1) NOT NULL CHECK (osm_type IN ('w', 'r')),
    osm_id                 bigint NOT NULL CHECK (osm_id > 0),
    motorcar               text, motor_vehicle          text, vehicle          text, access          text,
    motorcar_forward       text, motor_vehicle_forward  text, vehicle_forward  text, access_forward  text,
    motorcar_backward      text, motor_vehicle_backward text, vehicle_backward text, access_backward text,
    geom                   geography(MultiPolygon, 4326) NOT NULL,
    marked                 boolean NOT NULL GENERATED ALWAYS AS (public.access_feature_marked(
                               motorcar, motor_vehicle, vehicle, access,
                               motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
                               motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward)) STORED,
    PRIMARY KEY (osm_type, osm_id),
    CHECK (num_nonnulls(motorcar, motor_vehicle, vehicle, access,
                        motorcar_forward, motor_vehicle_forward, vehicle_forward, access_forward,
                        motorcar_backward, motor_vehicle_backward, vehicle_backward, access_backward) > 0)
);

-- --- the load record --------------------------------------------------------
--
-- One row per load, `current` on at most one. The service plans nothing while
-- no row is current, or while its SHA-256 differs from the tile set's
-- (`§ Loading from the extract the tiles are built from`). A feature lost on
-- load fails the load, so a row whose counts disagree cannot be written at
-- all. The area counts are null where the load did not take areas — every load
-- until #253's loader does.

CREATE TABLE IF NOT EXISTS public.osm_extract_load (
    id                     bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    extract_sha256         text NOT NULL CHECK (extract_sha256 ~ '^[0-9a-f]{64}$'),
    extract_timestamp      timestamptz,
    loaded_at              timestamptz NOT NULL DEFAULT now(),
    way_features           bigint NOT NULL CHECK (way_features >= 0),
    way_rows               bigint NOT NULL,
    barrier_features       bigint NOT NULL CHECK (barrier_features >= 0),
    barrier_rows           bigint NOT NULL,
    barrier_way_rows       bigint NOT NULL CHECK (barrier_way_rows >= 0),
    way_exit_rows          bigint NOT NULL CHECK (way_exit_rows >= 0),
    area_features          bigint CHECK (area_features >= 0),
    area_rows              bigint,
    current                boolean NOT NULL DEFAULT false,
    CHECK (way_rows = way_features),
    CHECK (barrier_rows = barrier_features),
    CHECK ((area_features IS NULL) = (area_rows IS NULL)),
    CHECK (area_rows = area_features)
);

CREATE UNIQUE INDEX IF NOT EXISTS osm_extract_load_one_current
    ON public.osm_extract_load (current) WHERE current;

-- --- the indexes ------------------------------------------------------------
--
-- The primary keys above are four of `§ The access-marking indexes`. Part E of
-- the verify file EXPLAINs `access_crossings` and fails on a sequential scan of
-- any table it reads; `access_way_nodes_gin` and `access_way_geom_gist` serve
-- the run and the stopping-position checks of #248 and #256, and are unproven
-- until those stories' queries are EXPLAINed, as are `access_way_exit_pkey`.

CREATE INDEX IF NOT EXISTS access_area_geom_gist ON public.access_area USING gist (geom);
CREATE INDEX IF NOT EXISTS access_way_nodes_gin  ON public.access_way  USING gin (nodes);
CREATE INDEX IF NOT EXISTS access_way_geom_gist  ON public.access_way  USING gist (geom);

-- --- the route check --------------------------------------------------------
--
-- Step 3 of `§ The route check`, the one statement: every marking a car leg's
-- edges cross. Edge i (1-based) runs along OSM way edge_way_ids[i], and its
-- stretch of shape is the points whose point_edges entry is i, in order. A row
-- is returned for each marked way an edge runs along, each marked barrier on
-- an edge's way lying within tolerance_m metres of its stretch, and each
-- marked area its stretch intersects, so the caller can say what was found and
-- where on the leg (AC2).
--
-- IT FAILS CLOSED. An input it cannot check in full returns a row of kind
-- `unchecked` rather than fewer rows: a tolerance that is null, negative, NaN
-- or infinite (ST_DWithin with a null distance would drop every barrier
-- silently), arrays of unequal length, an edge without a way id, an edge with
-- fewer than two points, or a point naming an edge that does not exist. Any
-- row refuses the leg.
--
-- It is a single-SELECT SQL function, neither STRICT nor carrying a SET
-- clause, so the planner inlines it and EXPLAIN shows the plan of the
-- statement itself (part E). The table names are schema-qualified for the same
-- reason: a SET search_path would stop the inlining.

CREATE OR REPLACE FUNCTION public.access_crossings(
    edge_way_ids bigint[],
    point_edges  integer[],
    point_lons   double precision[],
    point_lats   double precision[],
    tolerance_m  double precision
) RETURNS TABLE (
    edge_index integer,
    kind       text,
    osm_type   char(1),
    osm_id     bigint,
    longitude  double precision,
    latitude   double precision
)
LANGUAGE sql STABLE PARALLEL SAFE
AS $$
    WITH leg_edge AS (
        SELECT e.ord::integer AS edge_index, e.way_id
          FROM unnest(edge_way_ids) WITH ORDINALITY AS e(way_id, ord)
    ), leg_point AS (
        SELECT p.edge_index, p.lon, p.lat, p.ord
          FROM unnest(point_edges, point_lons, point_lats) WITH ORDINALITY AS p(edge_index, lon, lat, ord)
    ), leg_stretch AS (
        SELECT p.edge_index,
               count(*) AS points,
               ST_MakeLine(ST_SetSRID(ST_MakePoint(p.lon, p.lat), 4326) ORDER BY p.ord)::geography AS geom
          FROM leg_point p
         WHERE p.lon IS NOT NULL AND p.lat IS NOT NULL
         GROUP BY p.edge_index
        HAVING count(*) >= 2
    )
    SELECT e.edge_index, 'way'::text, 'w'::char(1), w.way_id, NULL::double precision, NULL::double precision
      FROM leg_edge e
      JOIN public.access_way w ON w.way_id = e.way_id
     WHERE w.marked
    UNION ALL
    SELECT e.edge_index, 'barrier', 'n', b.node_id, ST_X(b.geom::geometry), ST_Y(b.geom::geometry)
      FROM leg_edge e
      JOIN leg_stretch s ON s.edge_index = e.edge_index
      JOIN public.access_barrier_way bw ON bw.way_id = e.way_id
      JOIN public.access_barrier b ON b.node_id = bw.node_id
     WHERE b.marked
       AND ST_DWithin(b.geom, s.geom, tolerance_m)
    UNION ALL
    SELECT s.edge_index, 'area', a.osm_type, a.osm_id, NULL, NULL
      FROM leg_stretch s
      JOIN public.access_area a ON ST_Intersects(a.geom, s.geom)
     WHERE a.marked
    UNION ALL
    SELECT 0, 'unchecked', NULL, NULL, NULL, NULL
     WHERE tolerance_m IS NULL OR tolerance_m < 0 OR tolerance_m = 'NaN'::double precision
        OR tolerance_m = 'Infinity'::double precision
        OR edge_way_ids IS NULL OR cardinality(edge_way_ids) = 0
        OR cardinality(point_edges) IS DISTINCT FROM cardinality(point_lons)
        OR cardinality(point_edges) IS DISTINCT FROM cardinality(point_lats)
        OR EXISTS (SELECT 1 FROM leg_edge e WHERE e.way_id IS NULL OR e.way_id <= 0)
        OR EXISTS (SELECT 1 FROM leg_edge e WHERE NOT EXISTS (SELECT 1 FROM leg_stretch s WHERE s.edge_index = e.edge_index))
        OR EXISTS (SELECT 1 FROM leg_point p WHERE p.edge_index IS NULL OR p.lon IS NULL OR p.lat IS NULL
                                               OR p.edge_index < 1 OR p.edge_index > cardinality(edge_way_ids))
    ORDER BY 1, 2, 4
$$;

-- --- the postcondition ------------------------------------------------------
--
-- `CREATE TABLE IF NOT EXISTS` over a table already present in another shape
-- adds nothing and says nothing. The one property every check rests on is that
-- `marked` is the generated column above, so it is asserted before COMMIT.

DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['access_way', 'access_barrier', 'access_area'] LOOP
        IF NOT EXISTS (
            SELECT 1
              FROM pg_attribute a
             WHERE a.attrelid = ('public.' || t)::regclass
               AND a.attname = 'marked'
               AND a.attgenerated = 's'
               AND pg_get_expr((SELECT d.adbin FROM pg_attrdef d WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum), a.attrelid)
                   LIKE '%access_feature_marked(%'
        ) THEN
            RAISE EXCEPTION 'public.%.marked is not the stored column generated by access_feature_marked — the table exists in another shape', t;
        END IF;
    END LOOP;
END
$$;

COMMIT;
