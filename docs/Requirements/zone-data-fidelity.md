# Zone data fidelity

What the system may and may not assume about an individual zone, and about the relationship between two zones, given what the Turf API exposes and what it merely *guides*. Distinct from `Turf data integration`, which owns how that data is obtained and kept current — this category owns the limits the data's nature places on everything downstream of it, including the cost model and the optimizer. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `.claude/skills/requirements-authoring/SKILL.md`.

## FR-020 — Derive no per-zone extent from the nominal zone size

```
Statement:    The system shall derive no dimension, extent or footprint for an
              individual zone from the nominal zone size recorded under
              *Zone geometry* in `Architecture.md`.
Category:     Zone data fidelity
Source:       Architecture.md § Zone geometry
Priority:     MUST
Verification: inspection — the zone schema and the ingest that populates it
              carry no per-zone extent, radius or footprint, and the nominal
              size appears in neither
Acceptance:   the persisted zone representation and the ingest mapping that
              writes it carry a zone's coordinate and no extent, radius or
              footprint attribute, and neither names the nominal size recorded
              under `Architecture.md § Zone geometry`; a reader confirms this
              in the zone schema definition and in the ingest mapping
Status:       to-build
Depends-on:   none
Risk:         The API exposes nothing about the shape or size of any
              individual zone, and the nominal figure is a guideline that real
              zones vary considerably from, so a footprint built from it is
              fabricated data wearing the appearance of a measurement. Every
              later consumer reads it as a fact about that zone, and the error
              is invisible precisely because a fabricated footprint looks
              plausible everywhere.
Rationale:    The nominal size sits in the corpus and is exactly the shape a
              spatial buffer wants, which is why the prohibition is worth
              stating rather than assuming. What the system models instead of
              an area is owned by
              *The coordinate is the target* in `SPECIFICATION.md` and is not
              restated here; this record carries only what
              *Zone geometry* in `Architecture.md` establishes about what the
              API exposes, and therefore about what may not be relied on per
              zone.
Resolved-by:  #29
```

## FR-021 — Measure the distance between two zones for that pair

```
Statement:    The system shall obtain the distance between two zones by
              measuring it from their coordinates for that pair, rather than
              from the placement guideline recorded under
              *Distance between zones* in `Architecture.md`.
Category:     Zone data fidelity
Source:       Architecture.md § Distance between zones
Priority:     MUST
Verification: test — a pair of zones placed closer together than the placement
              guideline yields a measured distance below it, and an operation
              consuming that distance completes on the measured value
Acceptance:   given two zones whose coordinates lie closer together than the
              placement guideline recorded under
              `Architecture.md § Distance between zones`, when the distance
              between them is required, then it is computed from those two
              coordinates and the result is less than that guideline
              given the same pair of zones, when an operation that consumes
              the distance between them runs, then it completes and its result
              follows from the measured separation
Status:       to-build
Depends-on:   none
Risk:         The guideline has known exceptions — water zones are commonly,
              and often considerably, closer than it — so a floor assumed
              rather than measured is wrong exactly where zones are densest,
              which is where the candidate set is richest. A minimum
              separation baked into chaining, de-duplication or spatial
              bucketing discards real zones or merges distinct ones, and
              nothing in the output shows that an assumption stood in for a
              measurement.
Rationale:    The section's own test — that any logic which breaks when two
              zones sit far closer than the guideline is incorrect — is
              carried by FR-032, which reaches logic that obtains no distance
              at all. The guideline is not repudiated here; the same section
              keeps it as a rough expectation of candidate density and
              excludes it only from the cost model.
Resolved-by:  #34
```

## FR-032 — Assume no minimum distance between two zones

```
Statement:    The system shall assume no minimum distance between two zones in
              any computation, including one that never measures the distance
              between them.
Category:     Zone data fidelity
Source:       Architecture.md § Distance between zones
Priority:     MUST
Verification: test — a pair of zones placed closer together than the placement
              guideline survives ingest and every stage that groups, indexes
              or selects zones as two distinct zones, neither merged nor
              dropped
Acceptance:   given two zones whose coordinates lie closer together than the
              placement guideline recorded under
              `Architecture.md § Distance between zones`, when the zone set
              holding them is ingested and stored, then both are stored as
              distinct zones and neither is merged with nor dropped in favour
              of the other
              given the same pair of zones, when a stage that groups, indexes
              or selects zones returns a zone set covering the area holding
              them, then both appear in that set as separate zones
Status:       to-build
Depends-on:   none
Risk:         Water zones are commonly closer together than the placement
              guideline, and often considerably closer, so the case this
              record protects is ordinary rather than exotic, and it is the
              densest terrain — where the candidate set is richest — that an
              assumed floor gets wrong. The failure is silent: a pair
              collapsed into one still yields a plausible journey at a
              plausible cost, and nothing surfaces except a zone the user was
              never offered.
Rationale:    This record and FR-021 split one section's obligation by where
              the assumption hides. FR-021 governs how a distance is obtained
              when one is needed, and catches a guideline standing in for a
              measurement; this record governs logic that obtains no distance
              at all, and catches a structure sized so that a close pair
              cannot survive it — a proximity de-duplication radius, a
              clustering bucket, a spatial index tuned to the guideline. None
              of those consumes the distance between two zones, so a test
              written from FR-021 passes against every one of them, and
              neither record presupposes the other. It is stated as a
              prohibition because the assumption is cheap to make and easy to
              hide: each of those forms reads as an ordinary optimization in
              review and drops one zone of a close pair without saying so. It
              is verified by test rather than inspection for the same reason —
              what a reviewer misses in a diff, a close-pair fixture fails
              whatever form the assumption takes. The floor forbidden is one
              of any size rather than the guideline's alone, which is why the
              statement cites no figure: a separation assumed locally at a
              smaller figure is the same defect wearing a smaller number.
Resolved-by:  #30
```
