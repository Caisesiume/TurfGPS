# Architecture

How TurfGPS will feasibly satisfy the behaviour defined in `Concept.md`.

This document owns *how the system is built*. It does not restate what the system does, and it does not restate a formula or model that `Concept.md` defines — several sections there are explicitly the single statement of a model, and copying one breaks the anti-duplication rule the documentation depends on. Where behaviour is referenced below, it is referenced by section name.

**Status: technology decisions recorded, structure incomplete.** The decisions in *Technology decisions* were taken on 31 July 2026 and are binding until revised. The sections marked *Awaiting the split* are placeholders for content that must be **moved** out of `Concept.md`, per `docs/HANDOFF.md` §6.

---

## System context

TurfGPS is a single-tenant, account-free web application over a self-hosted geospatial data plane.

It consumes exactly one third-party service at runtime — the **Turf API** (`https://api.turfgame.com/v5`), for zone data, player rank, held zones, and region lordship. Everything else it needs is derived from open datasets it hosts itself: road and path geometry, elevation, and map tiles.

It integrates outward with **Google Maps** for stop-by-stop hand-off, per *The waypoint limit problem*. It provides no navigation of its own.

---

## Technology decisions

Each decision below records what was chosen, why, and what it costs. A decision marked **Proposed** follows the repository convention in `docs/README.md`: it is a concrete position to argue against, not a measured result.

### D1 — Go for the backend pipeline

**Decided.** The optimizer, access analysis, solve-session server, and zone-sync worker are written in Go.

The determining requirement is not raw compute but **retained state**. *Consequences for the optimizer* requires the candidate set, access classifications, and computed costs to survive the initial solve, and *Response time and progressive results* requires re-solve to feel immediate by reusing that state rather than recomputing it. That is a long-lived stateful process, and it rules out any serverless deployment target regardless of language.

Given that shape, Go earns its place on three properties: a natural fit for a stateful service holding many concurrent solve sessions; bounded worker pools with `context` cancellation over the candidate fan-out, which is directly how *progressive results* gets implemented rather than simulated; and a single static binary, which keeps deployment simple as `DEPLOYMENT.md` will require.

**What it costs.** Go has the thinnest geospatial ecosystem of the candidates considered. The mitigation is deliberate and load-bearing: **push geometry into PostGIS** — corridor buffers, proximity filtering, and the nearest-neighbour query behind the activity baseline in *Activity clusters* are all SQL — and keep in-process geometry to the light work that `orb` covers. Raster sampling is the one area where Go is genuinely weaker, and it is addressed in D6.

Python was recommended and not chosen. The reasoning above is why the choice is defensible rather than merely accepted.

### D2 — Vite + React SPA for the frontend

**Decided.** The client is a static single-page application built with Vite and React, served as files, talking to the Go service over HTTP.

The product has no SEO surface and no server-rendering benefit: it is an authenticated-by-nothing, map-heavy, interactive planning tool. Server-side rendering would add a build and deployment layer serving no requirement.

It also removes a hazard. A framework whose default deployment target is serverless creates continuous pressure toward a topology that D1 has already established cannot work.

The zone-by-zone review under *Reviewing zones one at a time* is a map-and-single-card interaction and must be built mobile-first, per *Platform and mobile-first design*. A wide table is a specification violation, not a styling choice.

### D3 — Valhalla as the default routing engine; openrouteservice as a registered adapter

**Decided.** Valhalla serves both car and pedestrian routing. openrouteservice is implemented behind the same port as a selectable alternative.

Valhalla is the default for four reasons that map directly onto stated requirements:

* **One instance serves both costing models.** Car and pedestrian routing come from the same tiles, where OSRM requires a separate process and graph per profile.
* **Grade-aware pedestrian costing is native**, not an add-on. *Accessibility scope for the first release* ships elevation-aware walking in full, so this is a first-release requirement.
* **`sources_to_targets` batches the matrix**, which is the mechanism by which *External call budget* is satisfied — see *The call budget* below.
* **`locate` returns edge attributes** — road class, speed limit, drivability, `layer` — which is exactly what *Enforceable exclusions* must test against when validating a stopping position.

**Why not split engines by need.** Using one engine for car routing and another for walking was considered and rejected. The stop model in *Elevation-aware walking time* chains car → walk → car through a single stopping point. Two engines snap that point to their own graphs, so the two halves of one stop's cost would be computed against **different geometry** — and the failure is silent, because every stop still yields a plausible number. Against a product whose measure of success is that *no zone is classified confidently and wrongly*, a silent geometry mismatch is the worst available failure mode. One engine owns any geometry that must agree with itself.

**On openrouteservice specifically.** Its hosted free tier is not viable here and must not be treated as a fallback. Published quotas are on the order of thousands of directions and hundreds of matrix requests per day; the candidate counts in *Bounding the candidate set* mean a single journey can exhaust a day's allowance. This is precisely the failure *External call budget* exists to prevent. Self-hosted openrouteservice is a legitimate alternative and is why the adapter exists; the hosted free tier is not.

The adapter seam makes this decision **measurable rather than permanent**. Once real corridors exist, both engines can be benchmarked on the same journeys and the default revisited on evidence.

### D4 — PostgreSQL with PostGIS as the single stateful store

**Decided.** One database holds synced zones, the OSM-derived feature data, and stored plans.

Zones need a spatial index for corridor resolution. The OSM data needs the attributes routing engines do not expose — barriers, `layer`/`bridge`/`tunnel` relationships, parking areas, access restrictions, `maxspeed`. Plans need ordinary transactional storage keyed by a short code. These are one problem, and splitting them across engines would mean joining across process boundaries for queries that are naturally a single statement.

This also replaces the MongoDB of the removed prototype. That store was never reachable from the design: it indexed zone coordinates as `[latitude, longitude]` under a `2dsphere` index, which GeoJSON specifies as `[longitude, latitude]`, so every spatial query it could have served would have been wrong.

### D5 — Primary-markets data extract for the first release

**Decided.** The first release builds its data plane from an OSM extract covering **Sweden, the United Kingdom, Germany, Norway, Denmark, and Finland** — the markets named in *Geographic scope*.

The point that makes this safe: **this is a data decision, not an architectural one.** The same stack runs on a six-country extract or on the planet. Widening is a longer import against unchanged code, not a rewrite.

It does **not** answer the open question in `HANDOFF.md` §7 about self-hosting versus metered APIs at global scope. It defers the cost commitment while keeping the architecture that makes either answer reachable. The question stays open and stays owned by this document.

### D6 — Elevation sampled from Copernicus GLO-30 — *Proposed*

Global 30-metre elevation as the baseline everywhere, with national high-resolution models added later as adapters under *Global data first, local data as enhancement*.

Two distinct consumers must not be conflated:

* **Walking speed by gradient** is handled inside the routing engine, from elevation baked into its tiles.
* **Barrier and feasibility detection** under *Elevation and feasibility rules* — cliffs, embankments, retaining walls — requires sampling the raster directly along the candidate access path. A routing engine that has already declined to route somewhere cannot tell you why.

**Proposed mechanism**, and the weakest decision here: sample Cloud-Optimized GeoTIFFs via `godal`, accepting a cgo dependency. The no-cgo alternative is PostGIS raster with `ST_Value`, which keeps the toolchain pure at some cost in bulk-sampling throughput. This should be settled by measurement against real access paths, not by preference.

**Confidence note.** A 30-metre global model is coarse relative to the barriers being detected — a retaining wall is narrower than one cell. This is a known limitation, it is exactly why *Terrain confidence* exists, and stops analysed at this resolution must carry lower confidence than stops analysed against a national model. It must not be presented as though it resolved the question.

### D7 — Self-hosted vector tiles and geocoding — *Proposed*

Map tiles are rendered from the same OSM extract and served as static files; the client uses MapLibre GL JS. Geocoding for address and place search runs against the same data, per *Locations*, which requires it not be a metered external dependency. Zone-name search needs no external lookup at all — the synced zone table already holds every name.

---

## Runtime topology

```
┌─────────────────────────────┐
│  Browser (React SPA)        │
│  MapLibre GL JS             │
└──────────────┬──────────────┘
               │ HTTP / JSON
┌──────────────┴──────────────┐
│  TurfGPS service (Go)       │
│  ─ journey API              │
│  ─ solve sessions (state)   │
│  ─ optimizer + scoring      │
│  ─ access classification    │
│  ─ explanation layer        │
└──┬────────┬────────┬────────┘
   │        │        │
   │        │        └──────────────┐
┌──┴─────┐ ┌┴──────────┐ ┌──────────┴──┐
│Valhalla│ │ PostGIS   │ │ DEM rasters │
│(routing│ │ zones,    │ │ (elevation) │
│ car +  │ │ OSM feat, │ └─────────────┘
│ foot)  │ │ plans     │
└────────┘ └───────────┘
                 ▲
      ┌──────────┴──────────┐
      │ Zone sync worker    │  ── scheduled ──▶ Turf API v5
      └─────────────────────┘
```

Three properties of this shape are requirements rather than preferences:

**The service is stateful and long-lived.** Solve sessions live in it. This is not a scale-out-behind-a-load-balancer design without deliberate session affinity, and `DEPLOYMENT.md` must account for that.

**The zone sync is never on a request path.** *Retrieving zones* records `GET /v5/zones/all` as slow — a partial download reached ~82,000 zones and 23 MB before timing out at five minutes — and rate-limited to one request per 30 minutes. It is a scheduled worker writing to PostGIS, and the pipeline must tolerate the sync being mid-refresh or up to an hour stale.

**Nothing in a stored plan depends on the Turf API.** This is what makes *Never gate stored plans on the wizard* implementable: an outage degrades the volatile overlay and nothing else.

---

## Ports and adapters

*Provider adapters* requires data sources to sit behind a common interface, with the pipeline consuming the interface. The ports:

| Port | Responsibility | First implementation |
|------|----------------|---------------------|
| `RoutingProvider` | Car and pedestrian routes, matrices, edge attributes at a point | Valhalla; openrouteservice registered alongside |
| `ElevationProvider` | Point and profile sampling | Copernicus GLO-30 |
| `ZoneRepository` | Zone lookup, corridor queries, activity baseline | PostGIS |
| `TurfClient` | Player rank, held zones, region lords, zone sync | Turf API v5 |
| `PlanStore` | Short-code plan persistence and expiry | PostGIS |
| `Geocoder` | Address, place, and zone-name resolution | Self-hosted; zone names local |

The requirement this places on everything downstream is that **every consumer tolerates varying quality**, because coverage differs between one part of a route and another. The confidence model under *Terrain confidence* is the mechanism, and it extends naturally to provenance: a stop analysed against a two-metre national model is more confident than one analysed at thirty metres, and the recommendation must say so.

---

## The call budget

*External call budget* requires per-journey external call volume to be bounded and known, and asks this document to establish it with figures.

Two mechanisms bound it, and they compose:

**Self-hosting changes the unit.** Against a metered provider, a routing request is a cost. Against a local Valhalla, it is a function call over a warm tile cache. The constraint becomes latency and CPU, not money.

**Matrix batching bounds the count.** The naive pipeline routes once per candidate stop per alternative, which the concept correctly identifies as reaching thousands. `sources_to_targets` collapses each alternative's candidate set into a small number of matrix calls. The cap in *Bounding the candidate set* — 300 candidates per alternative — is what makes the matrix a fixed, known size.

**Turf API calls per journey are already bounded and small:** one `POST /v5/users` for rank, held zones, and medals; one `GET /v5/regions` for the Region Lord flag. Zone data comes from the local sync. Neither scales with candidate count.

**The concrete figures are not yet computed** and this section is incomplete until they are. They depend on the corridor arithmetic in *Bounding the candidate set*, and stating a number here before doing that arithmetic would produce exactly the invented constant the repository's conventions forbid.

---

## Awaiting the split

The following belong in this document and currently sit in `Concept.md`. They must be **moved**, not copied, per `docs/HANDOFF.md` §6.

* **Data sources and constraints** in full, **except** *The coordinate is the target* and *Rounds*, which are product facts and stay.
* The **ownership fetch strategy** — *The user's held zones are already known*.
* From *No accounts, and what that costs*: the cross-device analysis, the expiry policy, and the personal-data obligation.
* From the non-functional requirements: *Response time and progressive results*, *Global data first, local data as enhancement*, *Provider adapters*, *The cost consequence*, and *External call budget*.

Sections this document still owes independently of the split: persistence model, failure handling, observability, security architecture, and the deployment model handed to `DEPLOYMENT.md`.

---

## Open questions owned by this document

* **Self-hosting versus metered APIs at global scope.** The largest cost risk in the project. D5 defers it; the adapter pattern makes it replaceable. It is not answered.
* **The DEM sampling mechanism** — `godal` with cgo, or PostGIS raster. Settle by measurement, per D6.
* **Solve-session lifetime and residency.** Tied to the open question in `Concept.md` about *the lifetime of an unconfirmed route*: whether sessions are in-memory only, and therefore lost on restart, or persisted. The product question and the architectural one must be answered together.
* **Concrete per-journey call and latency figures**, per *The call budget*.
