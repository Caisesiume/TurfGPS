# Migrations

DDL for the single PostgreSQL/PostGIS store of `Architecture.md § D4`. One
numbered migration per reviewable change, each with a rollback beside it.

`migrations/` is a peer directory alongside `service/`, `web/` and `docs/`, in
the sense `Architecture.md § D8` gives that word. It is deliberately not inside
`service/`: the database is a data-plane component in its own right per
`DEPLOYMENT.md`, migrations are never applied by the service and never at boot,
and nothing here is compiled into the Go module.

## Files

| File | What it is |
|---|---|
| `NNNN_name.sql` | the forward migration; one transaction unless a statement forbids it |
| `NNNN_name.down.sql` | its rollback, written at the same time and reviewed with it |
| `NNNN_name.verify.sql` | the assertions that make the migration falsifiable |

`0001_zone_store.sql` creates the synced zone store: `zone`, the `zone_incoming`
staging table, `sync_run`, and the one index on `zone`.

## Applying one

**No agent applies a migration.** Application is a deliberate act by an
authorised operator after PR approval and explicit human authorisation, per
`postgis-migration-protocol`. Migrations are not applied at boot and are not
applied by this repository's own tooling.

```
psql -v ON_ERROR_STOP=1 -f migrations/0001_zone_store.sql
psql -v ON_ERROR_STOP=1 -f migrations/0001_zone_store.verify.sql
```

The verify script runs second and is not optional. `CREATE INDEX` succeeding
proves nothing about whether the planner uses the index, and `CREATE TABLE IF
NOT EXISTS` succeeding proves nothing about the shape of a table that already
existed.

## State of proof

**Nothing in this directory has been executed.** There is no database in this
repository, none was provisioned, and no PostgreSQL was available to the work
that wrote these files. The forward migration has not been applied, the
rollback has not been demonstrated, and no `EXPLAIN` output exists for any
query shape.

**Two conditions in `Architecture.md § What is unproven` bear on this migration
and neither is met.** Item 1 makes `EXPLAIN` evidence against a loaded copy an
acceptance condition of the first migration; `0001_zone_store.verify.sql` part E
is the instrument that would meet it. Item 2 asks whether PostgreSQL accepts this
table's generated column as `STORED` at all, and only applying this file answers
that one — there is no instrument for it, because the migration either applies or
it does not. What a rejection would cost is that item's to state and is not
restated here. Both are owed before anything in this directory is trusted, and
naming only the first read as though the second had been settled.

## The zone ingest field mapping

What the sync worker reads out of one record of the all-zones response recorded
under `Architecture.md § Retrieving zones`, and where it puts it. Every row of
`zone_incoming` is one record of that response; the merge into `zone` is the
statement in `Architecture.md § The sync write path`.

**Nullability here is the schema's fact and is enforced by the DDL. Why a field
is nullable is a measurement, and its home is
`Architecture.md § Retrieving zones` and
`Architecture.md § The region hierarchy is not a tree`** — the short
version is that a field absent from a substantial fraction of real records
cannot be `NOT NULL` without the sync rejecting responses that are perfectly
ordinary.

| Response field | Column | Type | Null | Note |
|---|---|---|:--:|---|
| `id` | `id` | `integer` | no | primary key, and the conflict target of the merge |
| `name` | `name` | `text` | no | **not** unique-constrained; the id is the key and the name is a label |
| `latitude` | `latitude` | `double precision` | no | stored as sent; also the Y ordinate of `geom` |
| `longitude` | `longitude` | `double precision` | no | stored as sent; also the X ordinate of `geom` |
| `dateCreated` | `date_created` | `timestamptz` | no | mutable upstream; refreshed on every run, never insert-only |
| `totalTakeovers` | `total_takeovers` | `integer` | no | **not** `smallint` — the observed maximum does not fit |
| `takeoverPoints` | `takeover_points` | `smallint` | no | |
| `pointsPerHour` | `points_per_hour` | `smallint` | no | |
| `type.id` | `type_id` | `smallint` | **yes** | absent on most records; absence is the ordinary case, never an anomaly |
| `type.name` | — | — | — | **not stored.** One name per id, no disagreement anywhere in the corpus; the id carries everything |
| `region.id` | `region_id` | `smallint` | no | |
| `region.name` | `region_name` | `text` | no | for a country that Turf has not subdivided, this holds the country's own name |
| `region.country` | `region_country` | `text` | **yes** | a two-letter code, and only for the countries Turf subdivides |
| `region.area.id` | `area_id` | `integer` | **yes** | absent for two unrelated reasons; see the section cited above |
| `region.area.name` | `area_name` | `text` | **yes** | |

The rows above flatten the `region` and `type` objects; the response's own field
list, and which of them are objects, is in `Architecture.md § Retrieving zones`.
Fourteen of the flattened fields map to the fourteen columns of `zone_incoming`,
and `type.name` maps to nothing. The remaining three columns of `zone` come from
nowhere in the response:

| Column | Written by | When |
|---|---|---|
| `geom` | PostgreSQL | never by the worker — it is `GENERATED ALWAYS ... STORED` |
| `first_seen_at` | the merge | on insert only, and never updated afterwards |
| `last_changed_at` | the merge | on insert and on every update, from the run's `completed_at` |

### Four things the mapping fixes that a reader might otherwise decide

**`geom` is never written by the worker.** It is a generated column, so an
`INSERT` naming it is an error rather than an override. The axis order lives in
DDL, in one line, and the write path has no opportunity to invert it — which is
the whole reason it is a generated column and not a trigger or an application
concern. `Architecture.md § Geometry, SRID, and the coordinate guard`.

**`region` and `type` are flattened, not normalised.** There is no `region`
table, no `area` table and no `country` dimension, and there cannot be one built
from this feed: `Architecture.md § The region hierarchy is not a tree` records
the three measured facts that kill the obvious normalisation, one of which
fails its precondition audit against real data today. The columns are
denormalised onto the zone row, and what would have been foreign keys are
assertions the sync counts and reports rather than constraints that abort it.

**Every field is refreshed on every run, including the ones that look
constant.** A write path that skips fields it assumes constant carries silent
staleness, and for the coordinates it carries silent spatial error. The merge in
`Architecture.md § The sync write path` refreshes all fourteen and compares with
`IS DISTINCT FROM` rather than `<>`, because four of them are nullable and `<>`
yields NULL rather than true when either side is — so an update that should fire
silently does not.

**There is no column for ownership, and adding one is the failure this schema
guards against.** `currentOwner` and `dateLastTaken` are absent from every
record the sync reads, and they are round-scoped: a synced table that acquired
one would acquire a round-rollover problem it does not currently have. Part A of
the verify script asserts the column set in **both** directions for exactly this
reason — a test that asserts these columns exist cannot fail when a column is
added, and a column being added is the event.
`Architecture.md § The two absences, and the test that keeps them absent`.

**What this store may *not* carry is not settled here.** `FR-020` and `FR-026`
are another story's, and they are inspected against this artefact rather than
decided by it.

## Recording a run

`sync_run` is the only durable record of what the sync has done. Both the
`outcome` vocabulary and the two-write sequence a run is recorded by — the row
inserted at the start, updated once at whatever end the run reaches — are
`Architecture.md § The sync write path`'s, and the columns the second write fills
are the table's own in `0001_zone_store.sql`. The vocabulary is enforced by a
`CHECK` there. **None of that is restated here**, and it was: this section
carried its own copy of the sequence, which is a second home that would have gone
stale with nothing in this file moving to notice it.

**What this file adds is an obligation on the worker rather than a description of
the sequence.** The worker must actually make both writes, and must make the
second one on **every** path out — including the ones it did not plan for. That
is the half a schema cannot enforce: nothing in the DDL can require a row to be
updated, so a run that ends without its second write leaves a record the table
happily accepts and no reader can tell from a run that died.

`0001_zone_store.verify.sql` is written to prove the schema's half of that, and
**has not proved it**: it inserts a row for each outcome carrying nothing but a
start instant and rolls each one back, and it has never been run, per
`§ State of proof` above. If any of those inserts is refused, a
run that failed before it received anything cannot be recorded, and the assertion
says so by name — which is what the script would establish and has not.

**Absence is recorded and never acted on.** Ids present in `zone` and missing
from the response go into `absent_ids`, and nothing is deleted on the strength
of them. `Architecture.md § Absence is recorded and never acted on` carries the
asymmetry that decides it: zone deletion has never been observed, while a
truncated response has, and a delete-on-absence rule turns the second into a
half-empty zone table sitting on the request path.

**Serialise against a migration, not merely against another run.** Both the
sync and any later migration take a `pg_advisory_lock` on a well-known key, per
`Architecture.md § Migrating against a running sync`. That turns *hope the table
is idle* into a guarantee, and it costs one line in each.
