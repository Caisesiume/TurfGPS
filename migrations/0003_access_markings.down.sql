-- 0003_access_markings.down.sql — the documented rollback of 0003_access_markings.sql
--
-- ----------------------------------------------------------------------------
-- DESTRUCTIVE, AND NARROWLY. It drops the access-marking tables, the load
-- record, the rule and the check. The markings are a derivation of the
-- extract and the loader rebuilds them after a re-apply; the load history is
-- lost. `zone` and everything 0001 and 0002 created are untouched.
--
-- WITH THE TABLES GONE THE SERVICE PLANS NO JOURNEY: the check can read no
-- extract identity, and an identity it cannot read refuses every journey
-- (`Architecture.md § Loading from the extract the tiles are built from`).
-- That is the safe state, and the reason this rollback is not a way to serve
-- legs unchecked.
--
-- Applying this is a human's decision, taken with the same explicit
-- authorisation as the apply, per `postgis-migration-protocol`. No agent
-- applies it.
-- ----------------------------------------------------------------------------
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY. CI's `migrations` job runs it twice
-- after the verify file (`scripts/db/migrate-verify.sh`, #153).
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0003_access_markings.down.sql
--
-- The tables go first, because their generated columns depend on the rule's
-- functions. `IF EXISTS` lets it finish a rollback that stopped part way, and
-- for the reason `0001_zone_store.down.sql` gives it exits 0 whether it
-- dropped everything or nothing.

BEGIN;

SET LOCAL search_path = pg_catalog, public, pg_temp;

DROP FUNCTION IF EXISTS public.access_crossings(bigint[], integer[], double precision[], double precision[], double precision);

DROP TABLE IF EXISTS public.access_way_exit;
DROP TABLE IF EXISTS public.access_barrier_way;
DROP TABLE IF EXISTS public.access_area;
DROP TABLE IF EXISTS public.access_barrier;
DROP TABLE IF EXISTS public.access_way;
DROP TABLE IF EXISTS public.osm_extract_load;

DROP FUNCTION IF EXISTS public.access_feature_marked(text, text, text, text, text, text, text, text, text, text, text, text);
DROP FUNCTION IF EXISTS public.access_marked(text, text, text, text);

COMMIT;
