-- 0001_fixture.down.sql — the rollback of 0001_fixture.sql (#153).
--
-- A fixture: see 0001_fixture.sql. Like a real rollback written with IF EXISTS,
-- it exits 0 whether it dropped the table or found nothing, which is what the
-- runner's second run of it observes. It leaves the PostGIS extension in place.

BEGIN;

DROP TABLE IF EXISTS public.fixture_point;

COMMIT;
