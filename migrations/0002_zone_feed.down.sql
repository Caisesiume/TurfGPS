-- 0002_zone_feed.down.sql — the documented rollback of 0002_zone_feed.sql
--
-- ----------------------------------------------------------------------------
-- DESTRUCTIVE, AND NARROWLY. It drops the latest feed time and every refusal
-- recorded. The first costs one feed request with no lower bound after a
-- re-apply, which re-reads items the copy already holds and changes nothing
-- (`FR-120`'s Rationale); the second loses the record of every item the copy
-- could not store. `zone` and everything 0001 created are untouched.
--
-- Applying this is a human's decision, taken with the same explicit
-- authorisation as the apply, per `postgis-migration-protocol`. No agent
-- applies it. Stop the feed worker first, or it fails on every poll until 0002
-- is applied again — logged, and never on a request.
-- ----------------------------------------------------------------------------
--
-- EXECUTED ON THROWAWAY INSTANCES ONLY. CI's `migrations` job runs it twice
-- after the verify file (`scripts/db/migrate-verify.sh`, #153).
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0002_zone_feed.down.sql
--
-- `IF EXISTS` lets it finish a rollback that stopped part way, and for the
-- reason `0001_zone_store.down.sql` gives it exits 0 whether it dropped both
-- tables or neither.

BEGIN;

SET LOCAL search_path = pg_catalog, public, pg_temp;

DROP TABLE IF EXISTS public.zone_feed_refusal;
DROP TABLE IF EXISTS public.zone_feed_state;

COMMIT;
