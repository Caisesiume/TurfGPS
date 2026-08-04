---
name: postgis-migration-protocol
description: The protocol for TurfGPS database migrations on self-hosted PostgreSQL/PostGIS — live-state-over-files, idempotent DDL, precondition audits against real data, branch or copy testing, and the human-gated production apply. Use for ANY schema/DDL work.
---

# PostGIS Migration Protocol

Database: **self-hosted PostgreSQL with PostGIS**, per `Architecture.md § D4` — one store holding synced zones, OSM-derived feature data, and stored plans.

> **Status: awaiting the schema.** The schema behind `PlanStore` and the synced zone table are listed under `Architecture.md § Still owed by this document`. Until it exists there is nothing to migrate, and the first migration is itself a reviewable item. The protocol below is in force from that first migration onward; it is transferable law learned on another platform, not a TurfGPS observation, and the parts marked as such should be confirmed against this stack rather than assumed.

## The sequence (no step is optional)

1. **Read LIVE state, not migration files.** Migration files drift from reality. Triggers: `SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE ...` — never trust the file that created them. Constraints and columns: query `information_schema` / `pg_catalog`.
2. **Author idempotent DDL** — `IF NOT EXISTS`, `ON CONFLICT`, guarded `DO $$` blocks. Document the rollback; call out irreversible DDL loudly.
3. **Precondition audit against live data** — before adding a constraint, prove the data satisfies it (count violating rows = 0). Record the audit query and its result.
4. **Test on a copy, never on production.** A restored dump or a dedicated test database. The production database is never where you find out. Spatial indexes in particular must be verified to be *used* — check `EXPLAIN` output, not merely that `CREATE INDEX` succeeded.
5. **⚠️ Trigger-probe gotcha:** `DO`-block probes **never commit**, so DEFERRED constraint triggers silently do not fire in them — a probe can "pass" against a trigger that would reject a real commit. Run `SET CONSTRAINTS ALL IMMEDIATE;` inside the probe first, or the probe proves nothing.
6. **Production apply is HUMAN-GATED.** Only after PR approval AND explicit human authorization. Then apply, re-run the audit, and verify.
7. **Record evidence** in the PR: the test-copy run, the precondition audit, the apply output, and the post-apply verification.

## Standing schema facts

- Migrations are **not** auto-applied at boot; application is deliberate.
- Column names follow the migration, not the Go code — fix Go to match SQL, never the reverse.
- **Coordinate order is a known trap on this project.** The deleted prototype indexed zone coordinates as `[latitude, longitude]` under a `2dsphere` index, which GeoJSON specifies as `[longitude, latitude]` — every spatial query it could have served would have been wrong, silently. PostGIS `geography`/`geometry` with an explicit SRID and a test asserting a known distance is the guard.
- The zone sync writes to this database on a schedule and is **never on a request path**. A migration that locks the zone table must account for a sync possibly running.
- `PG NUMERIC` accepts `NaN`, and `NaN >= 0` is TRUE — a CHECK constraint cannot catch NaN. Application-level guards stay primary wherever a numeric column must be finite.
