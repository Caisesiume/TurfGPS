-- 0004_access_run.down.sql — the documented rollback of 0004_access_run.sql
--
-- It drops the two statements the run is measured by, and nothing else: no
-- table, no row, nothing 0001 to 0003 created.
--
-- WITH THEM GONE NO RUN IS MEASURED, so the planner refuses every leg that runs
-- along a marked way at a location the user entered, as it did before #248:
-- the stricter state, and not a way to admit a run unmeasured.
--
-- Applying this is a human's decision, taken with the same explicit
-- authorisation as the apply, per `postgis-migration-protocol`. No agent
-- applies it.
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY. CI's `migrations` job runs it twice
-- after the verify file (`scripts/db/migrate-verify.sh`, #153).
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0004_access_run.down.sql

BEGIN;

SET LOCAL search_path = pg_catalog, public, pg_temp;

DROP FUNCTION IF EXISTS public.access_run_points(bigint[], double precision[], double precision[]);
DROP FUNCTION IF EXISTS public.access_run_network(bigint);

COMMIT;
