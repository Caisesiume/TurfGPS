-- 0002_fixture.sql — a second throwaway migration (#153).
--
-- A fixture: see 0001_fixture.sql. It indexes 0001's table, so it fails unless
-- the runner applies 0001 first — which is what makes the order observable.

BEGIN;

CREATE INDEX IF NOT EXISTS fixture_point_geom_gist ON public.fixture_point USING gist (geom);

COMMIT;
