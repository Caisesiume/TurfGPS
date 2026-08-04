# Requirements index

One table per category, carrying **a row for every ID ever issued** — retired IDs included, their rows staying after the record is tombstoned. This is the file to skim to see the whole corpus at once, and to `grep` for a single code.

**This index is derived; the record is authoritative.** Where a row here and the record in a category file disagree, the record is right and the row is stale — a librarian finding, fixed here, never fixed by editing the record to match. That is what makes it safe to regenerate this file wholesale without ever reading it as a second source of truth.

Category vocabulary and the ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. What a record must contain: `.claude/skills/requirements-authoring/SKILL.md`.

`Verification` below is the keyword alone; the sentence saying what it proves stays on the record. `Category` is redundant under a category heading and carried anyway, so a `grep FR-001` hit answers the whole question without its surrounding lines.

## Contents

| Category | File | Records |
|---|---|---|
| [Journey definition](#journey-definition) | `journey-definition.md` | `FR-001` – `FR-006`, `FR-012` (7) |
| [Turf data integration](#turf-data-integration) | `turf-data-integration.md` | `FR-019`, `FR-022` – `FR-026`, `FR-028` – `FR-031`, `FR-033` – `FR-034` (12) |
| [Zone data fidelity](#zone-data-fidelity) | `zone-data-fidelity.md` | `FR-020`, `FR-021`, `FR-032` (3) |
| [Cost and time composition](#cost-and-time-composition) | `cost-and-time-composition.md` | `FR-008`, `FR-009`, `FR-027` (3) |
| [Value model](#value-model) | `value-model.md` | `FR-007` (1) |
| [Objective selection and ranking](#objective-selection-and-ranking) | `objective-selection-and-ranking.md` | `FR-010`, `FR-011` (2) |
| [Recommendation set composition](#recommendation-set-composition) | `recommendation-set-composition.md` | `FR-013` – `FR-018` (6) |
| [Data currency and confidence](#data-currency-and-confidence) | `data-currency-and-confidence.md` | `NFR-002` (1) |
| [Outbound rate compliance](#outbound-rate-compliance) | `outbound-rate-compliance.md` | `NFR-001` (1) |
| [Runtime and deployment shape](#runtime-and-deployment-shape) | `runtime-and-deployment-shape.md` | `NFR-003` – `NFR-005` (3) |

**39 requirements, in 10 categories.** Sections below follow the register's order in `README.md`, not the ID order — a category's records are contiguous in its own file, never across the corpus. A category on the register with no records filed yet has no file and no section here; the register in `README.md` is the full list of legal names, and a name's absence from this page means unfiled, never illegal.

**All 39 rows read `to-build`, and `draft` is empty.** The last twenty-one — batch 3, `FR-019` … `FR-034` and `NFR-001` … `NFR-005` — were signed off by the Owner on 4 August 2026 and moved from `draft` **straight to `to-build`**, as batches 1 and 2 each did on the day they were filed. `approved` stays unreachable while `@requirements-reconciler` is dormant, and no record in this corpus has ever held it. See `README.md` § Corpus state.

**No row reads `—` in `Resolved-by`**: all 39 name at least one story, allocated in three passes — eighteen on 2 August 2026 and the twenty-one batch-3 rows on 4 August 2026, each allocation following the Owner's sign-off as a pass of its own. Nine epics and thirty-six stories carry them. Which stories, and which epic each requirement sits under, is `TRACEABILITY.md`; this column is a view of its first table.

## Journey definition

`journey-definition.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-001` | Plan a journey from an origin to a destination | Journey definition | MUST | to-build | test | #1, #2 |
| `FR-002` | Plan a journey with intermediate destinations | Journey definition | MUST | to-build | test | #2 |
| `FR-003` | Visit mandatory waypoints in the order entered | Journey definition | MUST | to-build | test | #2 |
| `FR-004` | Preserve the entered mandatory-waypoint set | Journey definition | MUST | to-build | test | #3 |
| `FR-005` | Require a destination | Journey definition | MUST | to-build | test | #4 |
| `FR-006` | Plan journey travel as travel by car | Journey definition | MUST | to-build | test | #6 |
| `FR-012` | Require an intermediate destination on a return to the origin | Journey definition | MUST | to-build | test | #5 |

## Turf data integration

`turf-data-integration.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-019` | Address every Turf API request to the current API version | Turf data integration | MUST | to-build | inspection | #18 |
| `FR-022` | Refresh the local zone copy from a scheduled background job | Turf data integration | MUST | to-build | test | #20 |
| `FR-023` | Resolve a route corridor's zones against the local copy | Turf data integration | MUST | to-build | test | #27 |
| `FR-024` | Plan against a mid-refresh or stale local copy | Turf data integration | MUST | to-build | test | #33 |
| `FR-025` | Build bounding-box requests against the permitted area product | Turf data integration | COULD | to-build | test | #21 |
| `FR-026` | Answer no ownership question from the local zone copy | Turf data integration | MUST | to-build | inspection | #29 |
| `FR-028` | Decide the user's own holdings by membership in the held-zone list | Turf data integration | MUST | to-build | test | #23 |
| `FR-029` | Determine region lordship once, from a single region response | Turf data integration | MUST | to-build | test | #24 |
| `FR-030` | Carry an absent ownership field as absent | Turf data integration | MUST | to-build | test | #22 |
| `FR-031` | Do not read an absent ownership field as a zone never taken | Turf data integration | MUST | to-build | test | #31 |
| `FR-033` | Plan no journey for lack of zone data only where no copy is held | Turf data integration | MUST | to-build | test | #33 |
| `FR-034` | Record a result as built from data stale beyond the bound | Turf data integration | MUST | to-build | test | #35 |

`NFR-001` and `NFR-002` were filed under this category on 3 August 2026 and moved out the same day, to `Outbound rate compliance` and `Data currency and confidence`, on the Owner's ruling that a `Category` never spans both kinds — see `README.md` § Category register. Their rows are under those two headings below, and nothing about either record changed in the move but its `Category` field.

## Zone data fidelity

`zone-data-fidelity.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-020` | Derive no per-zone extent from the nominal zone size | Zone data fidelity | MUST | to-build | inspection | #29 |
| `FR-021` | Measure the distance between two zones for that pair | Zone data fidelity | MUST | to-build | test | #34 |
| `FR-032` | Assume no minimum distance between two zones | Zone data fidelity | MUST | to-build | test | #30 |

## Cost and time composition

`cost-and-time-composition.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-008` | Measure journey cost as time added to the journey without Turf stops | Cost and time composition | MUST | to-build | test | #7 |
| `FR-009` | Charge stop time to the journey even where no detour is driven | Cost and time composition | MUST | to-build | test | #8 |
| `FR-027` | Exclude blocktime from stop time | Cost and time composition | MUST | to-build | test | #32 |

`FR-027` is the first record in the corpus whose `Source` document differs from the document its category's other records cite: it is sourced from `Architecture.md § Player data` and filed under `Cost and time composition`. That is not a filing error and is not to be normalized — `Source` is where the obligation comes from, `Category` is the subsystem it binds.

## Value model

`value-model.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-007` | Derive a journey's Turf value from the zones it captures | Value model | MUST | to-build | test | #9 |

## Objective selection and ranking

`objective-selection-and-ranking.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-010` | Do not prefer a journey alternative for its zone count | Objective selection and ranking | MUST | to-build | test | #10 |
| `FR-011` | Balance value against cost from the individual user's preferences | Objective selection and ranking | MUST | to-build | test | #11 |

## Recommendation set composition

`recommendation-set-composition.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-013` | Offer more than one journey alternative | Recommendation set composition | MUST | to-build | test | #12 |
| `FR-014` | Do not offer two indistinguishable journey alternatives | Recommendation set composition | MUST | to-build | test | #13 |
| `FR-015` | Do not offer an alternative beaten outright by another | Recommendation set composition | SHOULD | to-build | test | #14 |
| `FR-016` | Offer alternatives that differ in ways a user would act on | Recommendation set composition | MUST | to-build | human-judgement | #17 |
| `FR-017` | Do not withhold alternatives for a missing kind of alternative | Recommendation set composition | MUST | to-build | test | #15 |
| `FR-018` | Do not withhold a compliant alternative for its lower Turf value | Recommendation set composition | MUST | to-build | test | #16 |

## Data currency and confidence

`data-currency-and-confidence.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `NFR-002` | Lower confidence as the Turf data behind a recommendation ages | Data currency and confidence | SHOULD | to-build | test | #36 |

## Outbound rate compliance

`outbound-rate-compliance.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `NFR-001` | Hold outbound Turf calls within the API's published limits | Outbound rate compliance | MUST | to-build | test | #28 |

## Runtime and deployment shape

`runtime-and-deployment-shape.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `NFR-003` | Build the service as one self-contained executable | Runtime and deployment shape | MUST | to-build | test | #19 |
| `NFR-004` | Run the service as one long-running process | Runtime and deployment shape | MUST | to-build | inspection | #25 |
| `NFR-005` | Serve the client as static files | Runtime and deployment shape | MUST | to-build | test | #26 |

## Reading the Verification column

**The split across the corpus is 34 `test` · 4 `inspection` · 1 `human-judgement`.**

`inspection` appears in this index for the first time with batch 3 — `FR-019`, `FR-020`, `FR-026` and `NFR-004` — and it is a chosen method rather than a test not yet written. Each of the four names the artefact examined, the property that must hold and where a reader confirms it, in the form `.claude/skills/requirements-authoring/SKILL.md` § Acceptance-criteria form requires; none is wrapped in given/when/then, because nothing is executed and a criterion narrating a run nobody performs can never fail. `NFR-004` is the first record in the corpus whose kind and whose method point at different criterion forms, and it takes the method's — which is the rule that section states, not an exception made for it.

`human-judgement` on `FR-016` is likewise not a placeholder for a method not yet chosen: the record names the judge and the standard applied. What the keyword means, how a story resolving such a record is labelled, and why a fabricated metric would be worse, are in `.claude/skills/requirements-authoring/SKILL.md` § Verification methods.

An earlier wording of that sentence ended *and the story that resolves it carries the `human-verified` label*, which was a **story's label mirrored in the corpus** and is forbidden without exception by § Corpus layout. Raised by `@requirements-librarian` on 4 August 2026 and repaired the same day. The reading that saved it — that the skill states the same convention, so this was a cross-reference rather than a second home — is exactly why it survived three passes: it asserted the label of **one live issue**, which the board answers correctly and this page cannot. The repair points at the rule instead of copying the instance, which is what a cross-reference should have done from the start.

`—` in `Resolved-by` means no story has been allocated yet. It never means the link is unknown: the single home for requirement → story is the first table in `TRACEABILITY.md`, and this column is a view of it, regenerated in the same pass. A retired record's row is the one exception, and what this column then reads is fixed by the freeze rule in `README.md` § How this folder works.
