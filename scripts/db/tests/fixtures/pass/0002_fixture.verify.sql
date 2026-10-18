-- 0002_fixture.verify.sql — the assertions over 0002_fixture.sql (#153).
--
-- A fixture: see 0001_fixture.sql.

SET search_path = pg_catalog, public, pg_temp;

-- === Part A — the index exists, on public.fixture_point, as GiST ============

DO $$
DECLARE
    method text;
BEGIN
    SELECT am.amname INTO method
      FROM pg_index i
      JOIN pg_class ic ON ic.oid = i.indexrelid
      JOIN pg_namespace n ON n.oid = ic.relnamespace
      JOIN pg_am am ON am.oid = ic.relam
     WHERE n.nspname = 'public' AND ic.relname = 'fixture_point_geom_gist'
       AND i.indrelid = 'public.fixture_point'::regclass;

    IF method IS DISTINCT FROM 'gist' THEN
        RAISE EXCEPTION 'fixture_point_geom_gist on public.fixture_point is %, expected a gist index.',
            coalesce(method, 'absent');
    END IF;
END
$$;

-- === Part B — EXPLAIN (ANALYZE, BUFFERS), for the job to capture ============
--
-- This part asserts nothing, and it says so here so that its `pass` is read
-- for what it is. It exists so the job's capture of EXPLAIN output is
-- exercised: the table is empty, so the plan is evidence of nothing about
-- index use, which needs a loaded copy.

EXPLAIN (ANALYZE, BUFFERS)
SELECT id
FROM   public.fixture_point
WHERE  ST_DWithin(geom, ST_GeogFromText('SRID=4326;POINT(0 0)'), 1000);
