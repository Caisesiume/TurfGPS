# Codebase map

A directory-level map of what is on disk: what each directory and package is
for, and where the boundary between two of them lies.

**It states no model that has a home elsewhere.** The schema, the sync write
path, the ports, the refresh schedule and the API version are cited here and
repeated nowhere, per `docs/README.md § Conventions`. Which document answers
which question is `docs/README.md § The documents`, and is not restated either.

**Accurate as of 15 November 2026**, against the tree that lands the scheduled
zone sync, corridor resolution against the local copy, the review
session's ownership decisions, the zone feed (#232), the planning entry
point (#1), the planning gate on the local zone copy (#33), the car
routing adapter (#6), the access-marking tables and route check (#255), the
exclusion of marked ways from driving legs (#248), and the root's wiring of the
store and the tile set's identity to the planner (#249), the refusal of a
return to the origin with nowhere between (#5), the check that every
journey holds the mandatory waypoints entered (#3), and a journey
alternative's cost as the time it adds over the general route it is built
on (#7), its Turf value as what the zones it captures are worth (#9), and
the warned leg where nothing avoids a marking between two locations the user
entered (#271). It
is maintained by the work that changes the structure:
a PR that
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
| `.github/` | the CI workflow that runs `scripts/team/gates.sh` on every pull request, and the jobs that run tests against a real PostGIS and a real Valhalla — `TEAM.md § Scripts` |
| `.claude/` | the agent library — the agent definitions and the skills |
| `Makefile` | the gate recipes, which `scripts/team/gates.sh` runs directly where there is no `make` |

`bin/` appears on a machine that has built the service and is not in the
repository — `.gitignore` covers it.

## The Go module

Everything under `service/` is one module. `cmd/turfgps` is the executable and
the composition root: it reads configuration, opens a connection pool for each
background job configured, registers the routing adapter when a routing
engine is configured (#6) — planning through one of those pools, the sync's or
else the feed's, and none of its own (#249) — starts the HTTP
server, the sync worker and the zone
feed's poller (`cmd/turfgps/zonefeed.go`), and drains what is in flight on a
signal.
The server binds loopback unless `TURFGPS_LISTEN_ADDR` names another address
(#125). The same executable, run as `turfgps load-access-markings -extract FILE`
(`cmd/turfgps/accessload.go`), fills the access-marking tables from the extract
the tiles are built from and returns, so the clean host still needs one file
(`NFR-003`, #255). The planner the root builds reads the zone copy and the
access-marking tables through that pool, and the identity of the tiles the
engine serves from that engine's verbose status and the tile record
`TURFGPS_TILE_IDENTITY_FILE` names, as `DEPLOYMENT.md § The tile build` defines
(#249).

| Package | What it is for |
|---|---|
| `internal/config` | reads the runtime configuration from the process environment, the access-marking check's barrier tolerance (#255) and avoidance bound (#248) among it, neither with a default, and the tile record's path (#249) |
| `internal/corridor` | resolves the zones within a route corridor, through the corridor half of the `ZoneRepository` port it declares |
| `internal/httpapi` | the request surface — every HTTP handler is registered here or under here, behind the cross-origin grant to the origins `TURFGPS_CORS_ALLOWED_ORIGINS` names and to no other (#224); `POST /journeys` (`journeys.go`, #1) plans through `internal/planner`, in a wire shape that is provisional, answers 503 when no routing engine or no store is configured, or when no zone copy is held or it cannot be read, and 504 when a plan outlives the deadline the composition root sets below the server's write timeout (#6); a journey holding a warned leg is answered 200, each leg saying whether it is warned and a warned one naming its marked stretches as spans of its shape (#271) |
| `internal/planner` | plans a journey from an origin to a destination, through any intermediate destinations in the order entered (#2), via the `RoutingProvider` port it declares, never a concrete provider (#1); the port's contract is a route a car can drive (#6), returned as the OpenStreetMap ways it runs along, and every leg is checked against the `AccessMarkings` port it declares and refused where it crosses a marking in transit, while the markings and the `RoutingData` port report one extract (#255); the run at a location the user entered is read in `run.go`, and a refused leg is asked for again avoiding what it crossed, up to a configured bound (#248); where the provider then finds no route keeping off the places (`ErrNoRoute`) and both ends are locations the user entered, and the provider says it kept a route off no way but the ones the places were found on (`Exclusions`), the leg is produced warned instead, carrying no marking a connection avoids and naming every marked stretch on its shape, the run's included, in `warned.go` (`FR-129` to `FR-131`, #271) — never where an end is a stopping position the system chose, which `PlanLeg` is told directly until the stories choosing one land; it plans from whatever local zone copy is held and refuses only where none is, through the `ZoneCopy` port it declares (#33); it refuses a request whose destination is its origin unless it names an intermediate destination distinct from it (`FR-012`, #5); and every journey alternative leaves it through the check in `waypoints.go` that its mandatory-waypoint set is the set entered, none dropped and none substituted (`FR-004`, #3); every leg carries the provider's estimate of its driving time, from the answer that passed the check, and `cost.go` charges a journey alternative holding a Turf stop the time it adds over the general route it is built on, and one holding none — that route as it is — nothing, refusing one whose journey is not that route (`FR-008`, #7); the time added holds the time each stop carries as spent at it beside the driving, so a stop on the road already driven is not free, and a stop carrying none is refused (`FR-009`, #8) — what that time is, by access class, is priced before a stop reaches it, not there; and `value.go` derives an alternative's Turf value from the zones its stops capture, each once, and from nothing its route holds, taking what one zone is worth from its caller rather than stating a weight (`FR-007`, #9) |
| `internal/valhalla` | the Valhalla adapter behind `internal/planner`'s port, routing every leg for a car (#6), naming its ways through the engine's `trace_attributes` (#255), sending the places a leg is to avoid as `exclude_locations` (#248) and reporting the engine's no-path answer to them, error 442 alone, as the planner's `ErrNoRoute` (#271), and naming through `/locate` the ways it keeps a route off for each place, the planner's `Exclusions`, without which that answer is not believed (PR #297), reading each leg's estimated time from its `summary.time` (#7), and reading the identity of the tiles it serves as `internal/planner`'s `RoutingData` (`tileidentity.go`, #249); its integration tests run under the `valhalla` build tag against an engine CI builds from the OpenStreetMap extract in `testdata/`, which is ODbL-licensed data and not code — its README says so, and how to regenerate it |
| `internal/accessstore` | the PostGIS adapter behind `internal/planner`'s `AccessMarkings`: reads the current load's extract and calls the migration's route check (#255), and measures the run at a location the user entered against the least from the lengths `0004_access_run.sql` gives (`run.go`, #248); its integration tests run under the `valhalla` tag, against the engine and the tables loaded from the same extract |
| `internal/accessload` | the loader that fills the access-marking tables from an extract in one transaction and records the load (#255); nothing the request surface imports reaches it |
| `internal/osmpbf` | reads an OpenStreetMap `.osm.pbf` file whole, or refuses it — the loader's reader (#255) |
| `internal/ownership` | the review session's ownership decisions — whether the user holds a zone, whether they hold a region lordship anywhere, and a zone record's ownership fields carried as absent where the API omits them — through the two reads it declares as its ports, never the Turf adapter (#22, #23, #24) |
| `internal/turf` | the adapter for the Turf API, behind the `TurfClient` port of `Architecture.md § Ports and adapters`: `Client` for the sync's all-zones fetch, `SessionClient` for a review session's player-data and region requests, `FeedClient` for the zone feed (#232), all drawing on the process's one request budget |
| `internal/zonestore` | the read side of the synced zone store — the adapter behind `internal/corridor`'s port and `internal/planner`'s `ZoneCopy` (#33) — and where each connection pool is opened |
| `internal/zonesync` | the scheduled workers that write the local copy — the full refresh, and the zone feed's poller with its per-item ingest decision (`feed.go`, #232) — and the ports they declare |
| `internal/syncstore` | the PostGIS adapter behind those ports — the write side |

Four constraints across that set are not visible from the package names.

**`internal/config` deliberately has no default for the refresh schedule or the
all-zones endpoint** — the monthly instant's time of day and time zone (#233),
and the endpoint. They arrive from the environment or the sync does not run, so
that this module is a second home for none of them; their home is
`Architecture.md § Retrieving zones`. `internal/turf` writes no API path for
the same reason, and takes the endpoint as a resolved URL. The zone feed's
interval and endpoint are held the same way, in a configuration block of their
own (`internal/config/zonefeed.go`), with the same home.

**`internal/zonesync` exports no way to run a single refresh.** The only
exported entries that can reach an endpoint are the two loops — the
scheduler's and the zone feed's — which the process owns for its whole life,
so there is nothing for a handler to call.

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
reports how current the local copy is and whether one is held at all, and
resolves a corridor's zones from it —
and a handler may import it. Neither it nor `internal/corridor` may reach
`internal/turf` or `net/http`, which `service/internal/corridor/noturf_test.go`
checks.
`internal/syncstore` writes, imports `internal/zonesync` for its port
types, and no handler could usefully import it.

**A pool is opened per background job, and the package that opens them does not
hold them.** The sync and the zone feed each get one, so neither job's wiring
depends on the other's being configured. `internal/zonestore` carries the
constructor that opens each and pins the settings the service holds it to; the
composition root holds the only references, and closes each on every path but
one. Which settings, and why each is pinned, is on those constants in
`service/internal/zonestore/zonestore.go` and is not repeated here.

**The exception is the stop-timeout branch**, where `service/cmd/turfgps/main.go`
returns without closing — deliberately, to avoid blocking on the one connection
that has not come back, and `zonefeed.go`'s supervision of the feed does the
same. It is not a leak to be tidied up; the reason in full is
on that branch, and what it leaves behind for an operator reading `sync_run` is
`Architecture.md § The sync write path`.

**One of those pools reaches the read half too.** The root hands each to
`internal/syncstore`, and the sync's — or, where no sync is configured, the
feed's — to the planner as well, which reads `Reader.Held` and
`internal/accessstore` through it (#249); the request path opens no pool of its
own. One pool serving both halves is what the split is *for*. `Reader.Currency`,
`Reader.ZonesInCorridor` and `corridor.Resolve` still have **no callers**
outside their tests, and `internal/httpapi` imports neither store: the planner
reaches them through the ports it declares.

## The database directory

`migrations/` is a root peer and not a subdirectory of `service/`. Why it sits
there rather than inside the module is the opening of `migrations/README.md`,
which is the directory's front door and carries the rest of what a reader
needs: the file convention, `migrations/README.md § Applying one`, the ingest
field mapping the sync writes against, what the worker owes the run record,
the zone feed's own state (`0002_zone_feed.sql`), the access-marking tables
and route check (`0003_access_markings.sql`, #255), and the statements the run
at a location the user entered is measured by (`0004_access_run.sql`, #248).

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
