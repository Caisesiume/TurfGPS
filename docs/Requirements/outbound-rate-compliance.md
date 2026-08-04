# Outbound rate compliance

Holding the system's outbound request rate within the limits an external service publishes, counted across every caller inside the system at once — solve sessions and background jobs together, since the limit is the other party's and applies to the system as a whole. Distinct from `Call budget`, which bounds the per-journey call **volume** the system chooses for itself: this category answers to a ceiling set elsewhere, measured per unit time, and owns the aggregation point at which it is honoured. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `.claude/skills/requirements-authoring/SKILL.md`.

## NFR-001 — Hold outbound Turf calls within the API's published limits

```
Statement:    The system shall issue no more requests to a Turf API resource
              than the rate limits under
              *Data sources and constraints* in `Architecture.md` and
              *Retrieving zones* in the same document permit, counting every
              solve session and the scheduled zone sync together.
Category:     Outbound rate compliance
Source:       Architecture.md § Data sources and constraints;
              Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — several solve sessions running alongside a scheduled sync
              produce no window in which the recorded outbound rate to any
              Turf resource exceeds the limits those two sections state
Acceptance:   outbound requests to a Turf resource per unit time ≤ the
              per-resource limit under
              `Architecture.md § Data sources and constraints`, sustained
              while several solve sessions and a scheduled sync run together
              interval between successive requests to the all-zones endpoint
              recorded under `Architecture.md § Retrieving zones` ≥ the
              minimum that section states, under any sync schedule the worker
              is configured with
Status:       to-build
Depends-on:   none
Risk:         The Turf API is the only source of zone, player and region data,
              and the architecture names no alternative anywhere, so a
              throttle or a block does not degrade a feature — it stops the
              product. Nor can a breach be retried away: at the all-zones
              interval under `Architecture.md § Retrieving zones`, a wasted
              request costs the next window rather than a moment.
Rationale:    The limit belongs to the API, not to any one caller, so it binds
              the system as a whole. A rate limiter held per client instance,
              per session or per worker satisfies each caller and breaches the
              limit the moment two of them run at once — which is the ordinary
              case here, since a solve session and the sync worker reach the
              same service.
Resolved-by:  —
```
