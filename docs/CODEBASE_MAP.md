# Codebase map

A directory-level map of what is on disk: what each directory and package is
for, and where the boundary between two of them lies.

**It states no model that has a home elsewhere.** The schema, the sync write
path, the ports, the refresh interval and the API version are cited here and
repeated nowhere, per `docs/README.md § Conventions`. Which document answers
which question is `docs/README.md § The documents`, and is not restated either.

**Accurate as of 19 October 2026**, against the tree that lands the scheduled
zone sync and corridor resolution against the local copy. It is maintained by the work that changes the structure: a PR that
adds a package, moves a responsibility, or adds a root directory updates this
file in the same diff, per `codebase-map § When code arrives`. A map drifted
from disk is worse than no map, because the reader who trusts it opens the
wrong file and has no way to discover that from the file they opened.

## Repository layout

`Architecture.md § D8` decides the shape: there is no module at the repository
root, and the Go service is one peer directory among several.

| Directory | What it holds |
|---|---|
| `docs/` | the documentation set, which is this repository's primary artifact |
| `service/` | the Go service — one module, `github.com/Caisesiume/TurfGPS/service` |
| `web/` | the client of `Architecture.md § D2` — a Vite + React build to static files |
| `migrations/` | the DDL for the PostgreSQL/PostGIS store of `Architecture.md § D4` |
| `scripts/` | shell utilities for the team and the agent loop, and the gates — `TEAM.md § Scripts` |
| `.github/` | the CI workflow that runs `scripts/team/gates.sh` on every pull request — `TEAM.md § Scripts` |
| `.claude/` | the agent library — the agent definitions and the skills |
| `Makefile` | the gate recipes, which `scripts/team/gates.sh` runs directly where there is no `make` |

`bin/` appears on a machine that has built the service and is not in the
repository — `.gitignore` covers it.

## The Go module

Everything under `service/` is one module. `cmd/turfgps` is the executable and
the composition root: it reads configuration, opens the connection pool, starts
the HTTP server and the sync worker, and drains what is in flight on a signal.

| Package | What it is for |
|---|---|
| `internal/config` | reads the runtime configuration from the process environment |
| `internal/corridor` | resolves the zones within a route corridor, through the corridor half of the `ZoneRepository` port it declares |
| `internal/httpapi` | the request surface — every HTTP handler is registered here or under here, behind the cross-origin grant to the origins `TURFGPS_CORS_ALLOWED_ORIGINS` names and to no other (#224) |
| `internal/turf` | the adapter for the Turf API, behind the `TurfClient` port of `Architecture.md § Ports and adapters` |
| `internal/zonestore` | the read side of the synced zone store — the adapter behind `internal/corridor`'s port — and where the connection pool is opened |
| `internal/zonesync` | the scheduled worker that refreshes the local copy, and the ports it declares |
| `internal/syncstore` | the PostGIS adapter behind those ports — the write side |

Four constraints across that set are not visible from the package names.

**`internal/config` deliberately has no default for two values** — the refresh
interval and the all-zones endpoint. Both arrive from the environment or the
sync does not run, so that this module is a second home for neither; their home
is `Architecture.md § Retrieving zones`. `internal/turf` writes no API path for
the same reason, and takes the endpoint as a resolved URL.

**`internal/zonesync` exports no way to run a single refresh.** The only
exported entry that can reach the endpoint is the scheduler's loop, which the
process owns for its whole life, so there is nothing for a handler to call.

**No package may reach `internal/zonesync` transitively except the composition
root and `internal/syncstore`**, which is exempt because it implements the
ports the worker declares and must name them. Nothing reachable from the
request surface imports either. That is checked rather than intended, by a test
that refuses rather than passes when the packages it names are missing —
`service/internal/zonesync/offrequestpath_test.go`, which is also where the
argument for guarding this structurally lives.

## The client

`web/` builds to static files, deployed on Vercel and calling the service
cross-origin, per `docs/adr/ADR-0006-deployment-topology.md`. `src/api.ts` is
the one call path to the service; `src/apiBaseUrl.ts` validates the service's
address, which `vite.config.ts` fixes into the bundle at build time.
`test/nfr-005-static-serving.test.ts` serves the built output from a plain
file server and is `NFR-005`'s acceptance test.

## Why the zone store is two packages

This is the boundary most easily mistaken for a filing convention, so it is
stated here rather than left to be inferred from the names.

The split is what keeps the import-graph invariant above checkable: one store
package holding both halves would put the sync worker in the transitive imports
of every handler that ever reads a zone. The argument in full is on
`service/internal/zonestore/zonestore.go`'s package doc, beside the code it
constrains, and this file does not repeat it.

The boundary itself: `internal/zonestore` answers questions — it opens the pool,
reports how current the local copy is, and resolves a corridor's zones from it —
and a handler may import it. Neither it nor `internal/corridor` may reach
`internal/turf` or `net/http`, which `service/internal/corridor/noturf_test.go`
checks.
`internal/syncstore` writes, imports `internal/zonesync` for its port
types, and no handler could usefully import it.

**One pool is opened, and the package that opens it does not hold it.**
`internal/zonestore` carries the constructor that opens it and pins the settings
the service holds it to; the composition root holds the only reference, and
closes it on every path but one. Which settings, and why each is pinned, is on
those constants in `service/internal/zonestore/zonestore.go` and is not repeated
here.

**The exception is the stop-timeout branch**, where `service/cmd/turfgps/main.go`
returns without closing — deliberately, to avoid blocking on the one connection
that has not come back. It is not a leak to be tidied up; the reason in full is
on that branch, and what it leaves behind for an operator reading `sync_run` is
`Architecture.md § The sync write path`.

**Today that pool reaches one half.** The root hands it to `internal/syncstore`
and to nothing else: `zonestore.NewReader`, `Reader.Currency`,
`Reader.ZonesInCorridor` and `corridor.Resolve` are written and have **no
callers** outside their tests, and `internal/httpapi` imports neither store, so
the read half is built and unreached. One pool serving both is what the split is *for* —
the handler that eventually reads the currency takes the same reference rather
than opening a second pool — but that is owed work rather than the built
topology, and this file states it as owed.

## The database directory

`migrations/` is a root peer and not a subdirectory of `service/`. Why it sits
there rather than inside the module is the opening of `migrations/README.md`,
which is the directory's front door and carries the rest of what a reader
needs: the file convention, `migrations/README.md § Applying one`, the ingest
field mapping the sync writes against, and what the worker owes the run record.

The schema and the write path themselves are `Architecture.md § The schema` and
`Architecture.md § The sync write path`. Neither this file nor
`migrations/README.md` repeats them.

## Keeping this file true

Two properties are worth checking whenever this file is edited, because both
have failed elsewhere in this repository.

- **A directory or package listed here exists on disk.** This map describes
  what is, never what is planned; a planned package belongs on the board.
- **A description here does not restate a model.** Where a package's behaviour
  is decided by a document, this file names the document and stops. The
  reasoning is `docs/README.md § Conventions`.
