-- 0001_zone_store.down.sql — the documented rollback of 0001_zone_store.sql
--
-- ----------------------------------------------------------------------------
-- DESTRUCTIVE. This drops `zone` and every row in it.
--
-- The zone table is rebuildable from the all-zones endpoint recorded under
-- `Architecture.md § Retrieving zones`, so the data is not unrecoverable — but
-- the rebuild is a fresh sync, subject to that endpoint's own limit, and until
-- it completes the corridor query has nothing to resolve against. Two columns
-- do not survive the round trip at all: `first_seen_at` is written on insert
-- and never again, so a rebuilt table reports every zone as first seen at the
-- rebuild, and `sync_run` takes the whole history of what the sync has done
-- with it, including every `absent_ids` array recorded so far.
--
-- Applying this is a human's decision, taken with the same explicit
-- authorisation as the apply, per `postgis-migration-protocol`. No agent
-- applies it.
-- ----------------------------------------------------------------------------
--
-- NEVER EXECUTED. Like its forward migration, this file has not been run
-- against any PostgreSQL. It is the rollback the protocol requires to be
-- written and reviewed, not a rollback that has been demonstrated.
--
--     psql -v ON_ERROR_STOP=1 -f migrations/0001_zone_store.down.sql
--
-- ONE THING IT DELIBERATELY DOES NOT DO. It does not drop the PostGIS
-- extension. The extension is a database-level object that outlives this
-- migration: the OSM-derived feature tables owed under
-- `Architecture.md § Still owed by this document` will need it, and dropping
-- it here would take with it any spatial object created outside 0001 —
-- including, with CASCADE, ones this migration never saw.
--
-- NO INVALID INDEX TO CLEAN UP. 0001 builds its index inside the transaction
-- rather than concurrently, so there is no half-built index for this file to
-- drop. A later migration that uses CREATE INDEX CONCURRENTLY must carry that
-- drop in its own rollback, per
-- `Architecture.md § Migrating against a running sync`.

BEGIN;

-- --- The schema this rollback resolves in -----------------------------------
--
-- The forward migration's own line, byte for byte, for the reason decision 5 of
-- `0001_zone_store.sql` gives at length. The reason applies harder here.
--
-- Left unpinned, `DROP TABLE IF EXISTS zone` resolves through whatever path the
-- applying session happens to carry, and both ways it can go wrong are silent.
-- Against a role owning a same-named schema ahead of `public`, it drops a table
-- this migration never built. Where it finds nothing at all, `IF EXISTS`
-- downgrades that to a notice and the transaction commits — reporting a
-- rollback that removed nothing exactly as it reports one that removed
-- everything. A DESTRUCTIVE file is the last place to accept an exit status
-- that means either.
--
-- Every DROP below names `public`, which is where 0001 creates. `SET LOCAL` is
-- scoped to this transaction, so the session's own path returns at COMMIT or
-- ROLLBACK.

SET LOCAL search_path = pg_catalog, public, pg_temp;

DROP INDEX IF EXISTS public.zone_geom_gist;

DROP TABLE IF EXISTS public.sync_run;
DROP TABLE IF EXISTS public.zone_incoming;
DROP TABLE IF EXISTS public.zone;

COMMIT;
