-- 0002_fixture.down.sql — the rollback of 0002_fixture.sql (#153).
--
-- A fixture: see 0001_fixture.sql.

BEGIN;

DROP INDEX IF EXISTS public.fixture_point_geom_gist;

COMMIT;
