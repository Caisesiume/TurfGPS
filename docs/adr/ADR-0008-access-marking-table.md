# ADR-0008 — Access markings are read from a PostGIS table, and every car leg is checked against it

**Status:** accepted — 2026-11-14, by the Owner on #248 ([#issuecomment-6879855672](https://github.com/Caisesiume/TurfGPS/issues/248#issuecomment-6879855672)); in force from the merge of the PR that adds it.
**Source:** #248 — the lead's question with both options, [#issuecomment-6879463662](https://github.com/Caisesiume/TurfGPS/issues/248#issuecomment-6879463662), which carries backend-dev's blocked finding against Valhalla 3.9.1's source, and the Owner's decision on it, [#issuecomment-6879855672](https://github.com/Caisesiume/TurfGPS/issues/248#issuecomment-6879855672). Where this record and those comments differ, the comments are the Owner's and this record is the defect.
**Relation:** adds `Architecture.md § The access-marking tables`, the access-marking part of the OSM-derived feature tables that `Architecture.md § Still owed by this document` has owed since `§ D4`; the rest of those tables stays owed. Adds a port beside the others in `Architecture.md § Ports and adapters`. Leaves `§ D3` standing: Valhalla stays the default engine, unmodified. Changes no requirement: `FR-125` to `FR-128` and `DECISIONS.md § RD-059` to `RD-063` say what is marked and what may be driven; this record says only where the system reads it.

## Context

`FR-125`, `FR-126` and `FR-127` (#248) and `FR-128` (#253) forbid a driving leg along a way, past a barrier, or through an area that the map data marks private or access-restricted, outside one bounded run at a location the user entered. `RD-062` and `RD-063` make the marking an allow-list: the most specific of `motorcar`, `motor_vehicle`, `vehicle` and `access` that a feature carries governs, and every value on it but five is a marking. Every story that lets a client call `POST /journeys` waits on #248 and #253.

The engine cannot carry that rule. Read at the version CI pins (3.9.1, `.github/workflows/gates.yml`, `VALHALLA_IMAGE`):

- **Most marked values are open.** `lua/graph.lua`'s `access` table (lines 81–103) maps the marked values `private`, `destination`, `customers`, `delivery`, `permit`, `residents`, `restricted` and `official` to `"true"`, open, exactly as it maps the permissive `yes`, `permissive`, `designated`, `public` and `allowed` (`RD-063`); its `private` table (lines 105–112) flags six of them as private access, and `restricted` and `official` are in neither that flag nor any refusal. A value the tables do not list, such as `access=military`, matches nothing and leaves car access at the road's default, as the safety sentinel's S7 on #251 found (`#issuecomment-6878724583`).
- **A flagged way is a cost, not a prohibition.** `src/thor/route_action.cc` disables destination-only edges on bidirectional A*'s first pass (line 314) and allows them on the relaxed second pass (line 346), so a route that finds nothing else uses them; and where they are allowed, the `auto` costing prices them at a penalty, as the safety review of #246 read `lua/graph.lua` and `src/mjolnir/pbfgraphparser.cc` (`service/internal/valhalla/valhalla.go`'s package doc records it).
- **The route answer does not say which ways it used.** The route API reference at 3.9.1 (`docs/docs/api/route/api-reference.md`) documents no way identifier in its output; its one `way_id` is a location input, line 37, "not currently implemented". Nor does it carry a way's private or destination flag.

A hard, checkable exclusion therefore needs the markings held somewhere the service can read them, and the route's ways named so they can be checked. The map-matching service does name them: `trace_attributes` returns, for each edge it walks, `way_id`, the "Way identifier of the OpenStreetMap base data", and the edge's `begin_shape_index` and `end_shape_index` into the shape (`docs/docs/api/map-matching.md` at 3.9.1, § Edge items, lines 228 and 208), and `shape_match=edge_walk` matches a shape "from a prior Valhalla route" (§ Shape-matching parameters, line 38).

## Decision

The Owner chose option (a). Verbatim, from [#issuecomment-6879855672](https://github.com/Caisesiume/TurfGPS/issues/248#issuecomment-6879855672):

> **(a).** The service holds an OSM-derived access-marking table in PostGIS. Every route Valhalla returns is checked against it, and a leg that crosses a marking outside the allowed run is refused; Valhalla stays stock.

What is now true:

- **D1 — The markings live in PostGIS**, in the single store of `Architecture.md § D4`: one table each for ways, barriers and area features carrying any of the four keys, with whether each is marked computed in the database by one function from the keys it carries. The design is `Architecture.md § The access-marking tables`.
- **D2 — The tables and Valhalla's tiles are built from the same extract file**, and the service plans nothing while the two record different files.
- **D3 — The planner checks every car leg** against the tables, by the leg's OpenStreetMap way identifiers and its shape, and refuses one that runs along a marked way, passes a marked barrier or crosses a marked area outside the run `FR-126` admits. The check is the guarantee; nothing the engine is told replaces it.
- **D4 — The planner may tell Valhalla to avoid what the check found**, through the route request's documented exclusions, and checks the new answer the same way. Avoidance is how a leg avoiding a marking is produced where one exists; it is never what makes a leg safe.
- **D5 — Valhalla is stock.** No custom tile-build script, no change to its tables, no costing option.

## Options rejected

The reasons are the ones the question recorded, quoted from [#issuecomment-6879463662](https://github.com/Caisesiume/TurfGPS/issues/248#issuecomment-6879463662).

- **(b) A custom Valhalla tile-build script that writes the markings into the graph.** Against it: *"It changes the deployment's tile build (`DEPLOYMENT.md`, #25), misuses engine flags, and a marked gate can only be fully blocked (which breaks FR-126's run) or invisible."*

## Consequences

**Obliged:**

- A migration creating the tables of `Architecture.md § The access-marking tables`, under `postgis-migration-protocol`, with its rollback, its verify file and `EXPLAIN` evidence, applied by nobody without the Owner's authorisation. It is the marking-table story the analyst files on #248.
- A loader that fills them from the extract file, and a record of which file it loaded. The deployment's tile build (`DEPLOYMENT.md`, #25) builds the tiles from the same file in the same act and records it the same way.
- The `RoutingProvider` port's contract gains what the check reads: a car leg comes with the way identifier and the stretch of shape of every edge it runs along. A provider that cannot supply them cannot serve a car leg. The openrouteservice adapter inherits the obligation, and whether that engine can meet it is not established.
- One more engine call per car leg, `trace_attributes`, and one more database statement. `Architecture.md § The call budget` must count them.
- CI's `valhalla` job loads the committed test extract into a PostGIS beside the tiles it builds from the same file, so #248's and #253's criteria run against both.

**Given up, stated rather than implied:**

- A second copy of part of the map. The tables and the tiles are two derivations of one file, and the guard in D2 is what keeps them one; a rebuild of either alone stops planning until the other follows.
- The engine's own choice of route where a marking lies on it. An avoided leg is the engine's best answer under the exclusions the check added, which may be slower than an unconstrained one; where the exclusions cannot express the avoidance within the engine's limits, the leg is not produced though a connection avoiding the marking exists. That is a lost leg, never an unsafe one.
- Matrix figures. `sources_to_targets` returns times and distances, not paths, so a matrix figure may be for a connection through a marking. Only a leg that is built is checked, so only a built, checked leg's figures enter a stop's cost, its ranking, the journey's stated time and the ceiling check, and the ceiling is checked again on a leg that avoidance replaced (`Architecture.md § The route check`). A matrix figure may propose which legs to build; how the optimizer does so is not decided here.

**Reversibility: medium.** The tables, the loader and the check are additive and can be dropped by a migration's rollback, but every car leg's safety rests on them from the merge of the story that builds them; replacing them means option (b) or an engine that enforces the allow-list itself, either of which is a new decision.
