# Requirements index

One table per category, carrying **a row for every ID ever issued** — retired IDs included, their rows staying after the record is tombstoned. This is the file to skim to see the whole corpus at once, and to `grep` for a single code.

**This index is derived; the record is authoritative.** Where a row here and the record in a category file disagree, the record is right and the row is stale — a librarian finding, fixed here, never fixed by editing the record to match. That is what makes it safe to regenerate this file wholesale without ever reading it as a second source of truth.

Category vocabulary and the ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. What a record must contain: `requirements-authoring`.

`Verification` below is the keyword alone; the sentence saying what it proves stays on the record. `Category` is redundant under a category heading and carried anyway, so a `grep FR-001` hit answers the whole question without its surrounding lines.

## Contents

| Category | File | Records |
|---|---|---|
| [Journey definition](#journey-definition) | `journey-definition.md` | `FR-001` – `FR-006`, `FR-012` (7) |
| [Journey initialization](#journey-initialization) | `journey-initialization.md` | `FR-038` – `FR-041`, `FR-055` – `FR-057` (7) |
| [Route alternative generation](#route-alternative-generation) | `route-alternative-generation.md` | `FR-042`, `FR-058` – `FR-060` (4) |
| [Turf data integration](#turf-data-integration) | `turf-data-integration.md` | `FR-019`, `FR-022` – `FR-026`, `FR-028` – `FR-031`, `FR-033` – `FR-037` (15) |
| [Zone data fidelity](#zone-data-fidelity) | `zone-data-fidelity.md` | `FR-020`, `FR-021`, `FR-032` (3) |
| [Candidate identification](#candidate-identification) | `candidate-identification.md` | `FR-063` – `FR-065`, `FR-080` (4) |
| [Access classification](#access-classification) | `access-classification.md` | `FR-067`, `FR-068`, `FR-071` – `FR-073`, `FR-075`, `FR-078`, `FR-081` – `FR-085`, `FR-092` (13) |
| [Route construction](#route-construction) | `route-construction.md` | `FR-086`, `FR-087` (2) |
| [Cost and time composition](#cost-and-time-composition) | `cost-and-time-composition.md` | `FR-008`, `FR-009`, `FR-027`, `FR-062`, `FR-074`, `FR-079`, `FR-093` (7) |
| [Multi-leg budget allocation](#multi-leg-budget-allocation) | `multi-leg-budget-allocation.md` | `FR-050` – `FR-052` (3) |
| [Value model](#value-model) | `value-model.md` | `FR-007` (1) |
| [Objective selection and ranking](#objective-selection-and-ranking) | `objective-selection-and-ranking.md` | `FR-010`, `FR-011` (2) |
| [Recommendation set composition](#recommendation-set-composition) | `recommendation-set-composition.md` | `FR-013` – `FR-018`, `FR-043`, `FR-049` (8) |
| [Recommendation disclosure](#recommendation-disclosure) | `recommendation-disclosure.md` | `FR-047`, `FR-048`, `FR-054`, `FR-066`, `FR-088` (5) |
| [Safety exclusions](#safety-exclusions) | `safety-exclusions.md` | `FR-044` – `FR-046`, `FR-053`, `FR-061`, `FR-069`, `FR-070`, `FR-076`, `FR-077`, `FR-089` – `FR-091` (12) |
| [Coverage and data quality](#coverage-and-data-quality) | `coverage-and-data-quality.md` | `NFR-006` (1) |
| [Data currency and confidence](#data-currency-and-confidence) | `data-currency-and-confidence.md` | `NFR-002` (1) |
| [Classification correctness](#classification-correctness) | `classification-correctness.md` | `NFR-007` (1) |
| [Outbound rate compliance](#outbound-rate-compliance) | `outbound-rate-compliance.md` | `NFR-001` (1) |
| [Runtime and deployment shape](#runtime-and-deployment-shape) | `runtime-and-deployment-shape.md` | `NFR-003` – `NFR-005` (3) |

**100 requirements, in 20 categories.** Sections below follow the register's order in `README.md`, not the ID order — a category's records are contiguous in its own file, never across the corpus. A category on the register with no records filed yet has no file and no section here; the register in `README.md` is the full list of legal names, and a name's absence from this page means unfiled, never illegal.

**62 rows read `to-build` and 38 read `draft`.** Batch 5 — `FR-058` … `FR-093`, `NFR-006` and `NFR-007` — was filed on 7 August 2026 and is unsigned, so the `draft` column is occupied again; it had been empty since `FR-055` … `FR-057` were signed off earlier the same day. Every row filed before this batch reads `to-build`: batch 4 — `FR-038` … `FR-054` — was signed off on 6 August 2026 and all seventeen moved from `draft` **straight to `to-build`**, as batch 3b did earlier the same day, as batch 3 did on 4 August 2026 and as batches 1 and 2 each did on the day they were filed, and `FR-055` … `FR-057` moved the same way on 7 August 2026. `approved` stays unreachable while `@requirements-reconciler` is dormant, and no record in this corpus has ever held it. **Nothing in this column says how far the build has got** — that question is answered by following `Resolved-by` to the stories, never by a value written here. See `README.md § Corpus state`.

**62 rows name at least one story in `Resolved-by`, and the 38 batch-5 rows read `—`.** The sixty-two were allocated in six passes — eighteen on 2 August 2026, the twenty-one batch-3 rows on 4 August 2026, the three batch-3b rows and seventeen batch-4 rows on 6 August 2026, and three more on 7 August 2026, each allocation following the Owner's sign-off as a pass of its own — and thirteen epics and fifty-nine stories carry them. **The rows naming a story and the rows reading `to-build` are the same sixty-two**, and the rows reading `—` and the rows reading `draft` are the same thirty-eight. That is the ordinary interval between a filing pass and the story pass that follows it: **sign-off gates story creation, not the reverse**, so no story exists for a batch-5 record and none should. `—` says *not yet allocated* here and never *unknown*; for these thirty-eight it carries the earlier of its two readings, *filed and not yet signed*, which is the same one batch 3b's and `FR-055` … `FR-057`'s carried. Which stories, and which epic each requirement sits under, is `TRACEABILITY.md`; this column is a view of its first table.

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

## Journey initialization

`journey-initialization.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-038` | Plan no journey without an additional-time limit the user entered | Journey initialization | MUST | to-build | test | #41 |
| `FR-039` | Derive no additional-time limit from the journey's length | Journey initialization | MUST | to-build | test | #42 |
| `FR-040` | Admit an additional-time limit anywhere in the realistic range | Journey initialization | MUST | to-build | test | #43 |
| `FR-041` | Present neither end of the realistic range as unusual | Journey initialization | SHOULD | to-build | test | #48 |
| `FR-055` | Open the planner with no additional-time limit entered | Journey initialization | MUST | to-build | test | #60 |
| `FR-056` | Refuse an additional-time limit that is not positive | Journey initialization | MUST | to-build | test | #58 |
| `FR-057` | Impose no maximum admissible additional-time limit | Journey initialization | SHOULD | to-build | test | #59 |

## Route alternative generation

`route-alternative-generation.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-042` | Search for alternatives whatever the size of the stated limit | Route alternative generation | MUST | to-build | test | #49 |
| `FR-058` | Generate more than one general road route | Route alternative generation | MUST | draft | test | — |
| `FR-059` | Produce general road routes that are different drives | Route alternative generation | MUST | draft | human-judgement | — |
| `FR-060` | Build on more than the fastest conventional route | Route alternative generation | SHOULD | draft | test | — |

The second-hole debt recorded in `README.md § ID allocation ledger` — that nothing in the corpus obliges the system to produce more than one alternative — names this category as its home, and `FR-058`'s `Rationale` addresses it by name without settling it. Whether the debt is discharged or narrowed is `@requirements-engineer`'s and is not answered on this page.

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
| `FR-035` | Update dateCreated on zones the local copy already holds | Turf data integration | MUST | to-build | test | #38 |
| `FR-036` | Treat an absent zone type as ordinary data | Turf data integration | MUST | to-build | test | #39 |
| `FR-037` | Determine a zone's country where the region subkey is absent | Turf data integration | COULD | to-build | test | #40 |

`NFR-001` and `NFR-002` were filed under this category on 3 August 2026 and moved out the same day, to `Outbound rate compliance` and `Data currency and confidence`, on the Owner's ruling that a `Category` never spans both kinds — see `README.md § Category register`. Their rows are under those two headings below, and nothing about either record changed in the move but its `Category` field.

## Zone data fidelity

`zone-data-fidelity.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-020` | Derive no per-zone extent from the nominal zone size | Zone data fidelity | MUST | to-build | inspection | #29 |
| `FR-021` | Measure the distance between two zones for that pair | Zone data fidelity | MUST | to-build | test | #34 |
| `FR-032` | Assume no minimum distance between two zones | Zone data fidelity | MUST | to-build | test | #30 |

## Candidate identification

`candidate-identification.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-063` | Bound a corridor set by the corridor half-width | Candidate identification | MUST | draft | test | — |
| `FR-064` | Promote no more candidates to full evaluation than the cap | Candidate identification | MUST | draft | test | — |
| `FR-065` | Record where the promotion cap bound | Candidate identification | MUST | draft | test | — |
| `FR-080` | Apply the configured maximum walking distance as a bound | Candidate identification | MUST | draft | test | — |

`FR-064` carries the obligation the reserved-ID worked example in `requirements-authoring` illustrates. The example is `FR-000` and is fictional in its identity alone; the duty it draws on is created here, and the two are not the same record. See `README.md § ID allocation ledger` for why the example's number is reserved.

## Access classification

`access-classification.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-067` | Assign exactly one access class to every classified candidate | Access classification | MUST | draft | test | — |
| `FR-068` | Choose the validation regime from the direct-access tolerance | Access classification | MUST | draft | test | — |
| `FR-071` | Send an established level or barrier failure to the park-and-walk branch | Access classification | SHOULD | draft | test | — |
| `FR-072` | Classify uncertain where the direct-access evidence is ambiguous | Access classification | MUST | draft | test | — |
| `FR-073` | Measure a candidate's access route to the zone's coordinate | Access classification | MUST | draft | test | — |
| `FR-075` | Build no access path from a way the data bars to pedestrians | Access classification | MUST | draft | test | — |
| `FR-078` | Validate every leg of a multi-zone walking route | Access classification | MUST | draft | test | — |
| `FR-081` | Assign a confidence level to every access estimate | Access classification | MUST | draft | test | — |
| `FR-082` | Classify a low-confidence access estimate as uncertain | Access classification | MUST | draft | test | — |
| `FR-083` | Admit park-and-walk only on a routed path and an obtained elevation profile | Access classification | MUST | draft | test | — |
| `FR-084` | Keep uncertain candidates out of the optimization | Access classification | MUST | draft | test | — |
| `FR-085` | Classify a zone accessible only on an identified connection | Access classification | MUST | draft | test | — |
| `FR-092` | Identify a stopping position rather than manufacture one | Access classification | MUST | draft | test | — |

Thirteen records at first filing, which makes this the corpus's second-largest category behind `Turf data integration`. **It is not the whole of the access model**, and the neighbours are worth naming because a reviewer checking access coverage will otherwise read one file and stop: the exclusions that refuse a class sit under `Safety exclusions`, the pricing of a validated stop under `Cost and time composition`, what the user is told about an uncertain estimate under `Recommendation disclosure`, and the judged correctness bar over all of them under `Classification correctness`. Each line is drawn on the register in `README.md`.

## Route construction

`route-construction.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-086` | Route a detour's cost through the proposed stopping position | Route construction | MUST | draft | test | — |
| `FR-087` | Reuse no candidate's detour cost across journeys | Route construction | MUST | draft | test | — |

## Cost and time composition

`cost-and-time-composition.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-008` | Measure journey cost as time added to the journey without Turf stops | Cost and time composition | MUST | to-build | test | #7 |
| `FR-009` | Charge stop time to the journey even where no detour is driven | Cost and time composition | MUST | to-build | test | #8 |
| `FR-027` | Exclude blocktime from stop time | Cost and time composition | MUST | to-build | test | #32 |
| `FR-062` | Keep the general-route deviation separable from the Turf stop time | Cost and time composition | SHOULD | draft | test | — |
| `FR-074` | Compute access cost from the routed path and its elevation profile | Cost and time composition | MUST | draft | test | — |
| `FR-079` | Price a park-and-walk stop over the set of zones it serves | Cost and time composition | MUST | draft | test | — |
| `FR-093` | Compute both halves of a stop against one resolved stopping position | Cost and time composition | MUST | draft | test | — |

`FR-027` is the first record in the corpus whose `Source` document differs from the document its category's other records cite: it is sourced from `Architecture.md § Player data` and filed under `Cost and time composition`. That is not a filing error and is not to be normalized — `Source` is where the obligation comes from, `Category` is the subsystem it binds.

## Multi-leg budget allocation

`multi-leg-budget-allocation.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-050` | Allocate the budget across legs before optimizing any leg | Multi-leg budget allocation | MUST | to-build | test | #46 |
| `FR-051` | Size each leg's share in proportion to its baseline driving time | Multi-leg budget allocation | SHOULD | to-build | test | #53 |
| `FR-052` | Return an unused share to a pool for the remaining legs | Multi-leg budget allocation | SHOULD | to-build | test | #55 |

`FR-050` … `FR-052` were authored proposing `Cost and time composition` and filed here on `@requirements-engineer`'s assignment of 6 August 2026, the register entry being added in the same pass. The two categories are mirrors — that one composes a cost from its components, this one decomposes an allowance into shares — and the reasoning is on the register in `README.md`, not here.

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
| `FR-043` | Offer a zone-capturing within-limit alternative where one was produced | Recommendation set composition | MUST | to-build | test | #47 |
| `FR-049` | Offer an above-limit alternative only where value justifies it | Recommendation set composition | MUST | to-build | human-judgement | #57 |

## Recommendation disclosure

`recommendation-disclosure.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-047` | Present an above-limit alternative as a stretch alternative | Recommendation disclosure | MUST | to-build | test | #45 |
| `FR-048` | Name what the stated limit was exceeded for | Recommendation disclosure | MUST | to-build | test | #52 |
| `FR-054` | Present the additional time for the journey as a whole | Recommendation disclosure | SHOULD | to-build | test | #54 |
| `FR-066` | Tell the user where the promotion cap shaped the result | Recommendation disclosure | SHOULD | draft | test | — |
| `FR-088` | State the material uncertainty in an access estimate | Recommendation disclosure | SHOULD | draft | test | — |

`FR-047`, `FR-048` and `FR-054` were authored proposing `Recommendation set composition` and filed here on `@requirements-engineer`'s assignment of 6 August 2026, the register entry being added in the same pass. That category disclaims presentation in its own scope line, so the three could not have filed there; the reasoning is on the register in `README.md`, not here.

## Safety exclusions

`safety-exclusions.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-044` | Derive the absolute ceiling from the stated additional time | Safety exclusions | MUST | to-build | test | #44 |
| `FR-045` | Offer no journey alternative above the absolute ceiling | Safety exclusions | MUST | to-build | test | #50 |
| `FR-046` | Reject a configured ceiling multiplier above the permitted maximum | Safety exclusions | MUST | to-build | test | #51 |
| `FR-053` | Test the limit and the ceiling against the sum across all legs | Safety exclusions | MUST | to-build | test | #56 |
| `FR-061` | Test the stated limit and the ceiling against the whole additional time | Safety exclusions | MUST | draft | test | — |
| `FR-069` | Reject a direct-access tolerance above the permitted value | Safety exclusions | MUST | draft | test | — |
| `FR-070` | Require compatible levels and no barrier for direct road access | Safety exclusions | MUST | draft | test | — |
| `FR-076` | Exclude an absent, disconnected, or implausibly steep access path | Safety exclusions | MUST | draft | test | — |
| `FR-077` | Let no zone's value admit an unsafe or infeasible access route | Safety exclusions | MUST | draft | test | — |
| `FR-089` | Propose no stop on a road the exclusions refuse | Safety exclusions | MUST | draft | test | — |
| `FR-090` | Reject a maximum stopping speed above the permitted value | Safety exclusions | MUST | draft | test | — |
| `FR-091` | Plan no journey while an owed enforcement constant is unconfigured | Safety exclusions | MUST | draft | test | — |

The category triples in one pass and its subject widens with it: `FR-044` … `FR-046` and `FR-053` are the time ceiling, and the eight added here are the enforceable exclusions — the other half of the scope line the register has carried since consolidation. `FR-069`, `FR-090` and `FR-091` are configuration records rather than behaviour at solve time, and they sit here because the constants they guard are the exclusions' own; each states a direction or an absence and none states a figure.

## Coverage and data quality

`coverage-and-data-quality.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `NFR-006` | Fall access confidence with the evidence behind the estimate | Coverage and data quality | MUST | draft | test | — |

## Data currency and confidence

`data-currency-and-confidence.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `NFR-002` | Lower confidence as the Turf data behind a recommendation ages | Data currency and confidence | SHOULD | to-build | test | #36 |

`NFR-002` and `NFR-006` are the two halves the register separates by trigger: age against thinness. Both are filed now, adjacent on this page and in separate files, which is the arrangement the two scope lines were written to produce.

## Classification correctness

`classification-correctness.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `NFR-007` | Classify no zone confidently and wrongly | Classification correctness | MUST | draft | human-judgement | — |

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

**The split across the corpus is 92 `test` · 4 `inspection` · 4 `human-judgement`**, which is 100. Batch 5 added thirty-six `test` and two `human-judgement` on 7 August 2026 — `FR-059` and `NFR-007` — and is the first batch to file more than one judged record, doubling the corpus's count of them. It added no `inspection`. Batch 4 added sixteen `test` and one `human-judgement` on 6 August 2026, and `FR-055` … `FR-057` three further `test` the same day. `FR-055`'s `Verification` was amended twice on 7 August 2026 — its second half before the record's signature and its first half after — and **its keyword moved on neither**, so the split is untouched by both: this column carries the keyword alone, and a change to what the method proves leaves no trace here by design. The record is where that sentence is read, and `README.md § ID allocation ledger` is where the two amendments are told apart.

`inspection` appears in this index for the first time with batch 3 — `FR-019`, `FR-020`, `FR-026` and `NFR-004` — and it is a chosen method rather than a test not yet written. Each of the four names the artefact examined, the property that must hold and where a reader confirms it, in the form `requirements-authoring § Acceptance-criteria form` requires; none is wrapped in given/when/then, because nothing is executed and a criterion narrating a run nobody performs can never fail. `NFR-004` is the first record in the corpus whose kind and whose method point at different criterion forms, and it takes the method's — which is the rule that section states, not an exception made for it.

`human-judgement` on `FR-016`, `FR-049`, `FR-059` and `NFR-007` is likewise not a placeholder for a method not yet chosen: each of the four names the judge and the standard applied. What the keyword means, how a story resolving such a record is labelled, and why a fabricated metric would be worse, are in `requirements-authoring § Verification methods`.

An earlier wording of that sentence ended *and the story that resolves it carries the `human-verified` label*, which was a **story's label mirrored in the corpus** and is forbidden without exception by `requirements-authoring § Corpus layout`. Raised by `@requirements-librarian` on 4 August 2026 and repaired the same day. The reading that saved it — that the skill states the same convention, so this was a cross-reference rather than a second home — is exactly why it survived three passes: it asserted the label of **one live issue**, which the board answers correctly and this page cannot. The repair points at the rule instead of copying the instance, which is what a cross-reference should have done from the start.

`—` in `Resolved-by` means no story has been allocated yet. It never means the link is unknown: the single home for requirement → story is the first table in `TRACEABILITY.md`, and this column is a view of it, regenerated in the same pass. A retired record's row is the one exception, and what this column then reads is fixed by the freeze rule in `README.md § How this folder works`.
