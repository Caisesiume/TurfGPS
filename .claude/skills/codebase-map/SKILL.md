---
name: codebase-map
description: Where things live in TurfGPS — load before recon so you start from the map instead of re-deriving the layout. There is no code yet, so the map is currently the architecture documents; this skill says which section answers which question, and what changes once code exists.
---

# Codebase Map — Start Here for Recon

## There is no code

The repository holds documentation and one gitignored data file. The Next.js prototype was deleted on 31 July 2026 because it was a different application, not a partial implementation — it survives in git history only, and **nothing is being ported from it**. Do not read it for orientation; it will mislead you. Its MongoDB store, in particular, indexed coordinates in the wrong order and was never reachable from this design.

**Until code exists, `docs/Architecture.md` is the map.** Three sections answer most recon questions:

- ***Runtime topology*** — the components and what talks to what.
- ***Ports and adapters*** — the six ports (`RoutingProvider`, `ElevationProvider`, `ZoneRepository`, `TurfClient`, `PlanStore`, `Geocoder`) and their first implementations. This is the seam every provider sits behind.
- ***Technology decisions*** D1–D7 — what was chosen, why, and what it costs. Binding until revised.

For behaviour rather than structure, the routing is: what the product does → `SPECIFICATION.md`; how a number is computed → `CalculationSpecification.md`; what the user sees → `DESIGN.md`. Each document ends with what it still owes and the open questions it owns.

## Orientation rules

- **Ports before adapters.** The pipeline consumes the interface, never a concrete provider. Adding a country's dataset is implementing an adapter and registering it, not modifying the optimizer.
- **Geometry lives in PostGIS.** D1 chose Go with the thinnest geospatial ecosystem of the candidates, and the mitigation is load-bearing: corridor buffers, proximity filtering, and the nearest-neighbour query behind the activity baseline are SQL, not in-process geometry.
- **The service is stateful.** Solve sessions live in it, which rules out serverless regardless of language. A design that recomputes rather than retains fails `Architecture.md § Response time and progressive results`.
- **One engine owns geometry that must agree with itself.** Car and pedestrian routing come from the same Valhalla tiles deliberately; splitting them makes a stop's two halves disagree silently.
- **Recon before code, always** — the map orients you; it does not replace verifying the specific behaviour your item names.

## When code arrives

The first structural work should create **`docs/CODEBASE_MAP.md`** — a maintained directory-level map — and from that point this skill points there first and at `Architecture.md` second. After any structural change (new package, moved responsibility), the same PR updates the map. A map drifted from disk is a `needs-re`-worthy docs defect.
