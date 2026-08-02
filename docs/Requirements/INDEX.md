# Requirements index

One table per category, carrying **a row for every ID ever issued** — retired IDs included, their rows staying after the record is tombstoned. This is the file to skim to see the whole corpus at once, and to `grep` for a single code.

**This index is derived; the record is authoritative.** Where a row here and the record in a category file disagree, the record is right and the row is stale — a librarian finding, fixed here, never fixed by editing the record to match. That is what makes it safe to regenerate this file wholesale without ever reading it as a second source of truth.

Category vocabulary and the ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. What a record must contain: `.claude/skills/requirements-authoring/SKILL.md`.

`Verification` below is the keyword alone; the sentence saying what it proves stays on the record. `Category` is redundant under a category heading and carried anyway, so a `grep FR-001` hit answers the whole question without its surrounding lines.

## Contents

| Category | File | Records |
|---|---|---|
| [Journey definition](#journey-definition) | `journey-definition.md` | `FR-001` – `FR-006`, `FR-012` (7) |
| [Cost and time composition](#cost-and-time-composition) | `cost-and-time-composition.md` | `FR-008`, `FR-009` (2) |
| [Value model](#value-model) | `value-model.md` | `FR-007` (1) |
| [Objective selection and ranking](#objective-selection-and-ranking) | `objective-selection-and-ranking.md` | `FR-010`, `FR-011` (2) |
| [Recommendation set composition](#recommendation-set-composition) | `recommendation-set-composition.md` | `FR-013` – `FR-018` (6) |

**18 requirements, in 5 categories.** Sections below follow the register's order in `README.md`, not the ID order — a category's records are contiguous in its own file, never across the corpus. A category on the register with no records filed yet has no file and no section here; the register in `README.md` is the full list of legal names, and a name's absence from this page means unfiled, never illegal.

**All eighteen rows read `to-build`.** `FR-013` … `FR-018` — batch 2, filed 2 August 2026 — were signed off by the Owner on 2 August 2026, and that sign-off moved all six from `draft` straight to `to-build`; `approved` stays unreachable while `@requirements-reconciler` is dormant, and no record in this corpus has ever held it. Their `Resolved-by` reads `—` because story creation follows sign-off rather than sharing its pass, not because a link is missing. See `README.md` § Corpus state.

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

## Cost and time composition

`cost-and-time-composition.md`

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-008` | Measure journey cost as time added to the journey without Turf stops | Cost and time composition | MUST | to-build | test | #7 |
| `FR-009` | Charge stop time to the journey even where no detour is driven | Cost and time composition | MUST | to-build | test | #8 |

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
| `FR-013` | Offer more than one journey alternative | Recommendation set composition | MUST | to-build | test | — |
| `FR-014` | Do not offer two indistinguishable journey alternatives | Recommendation set composition | MUST | to-build | test | — |
| `FR-015` | Do not offer an alternative beaten outright by another | Recommendation set composition | SHOULD | to-build | test | — |
| `FR-016` | Offer alternatives that differ in ways a user would act on | Recommendation set composition | MUST | to-build | human-judgement | — |
| `FR-017` | Do not withhold alternatives for a missing kind of alternative | Recommendation set composition | MUST | to-build | test | — |
| `FR-018` | Do not withhold a compliant alternative for its lower Turf value | Recommendation set composition | MUST | to-build | test | — |

`human-judgement` on `FR-016` is the first appearance of that keyword in this index, and it is not a placeholder for a method not yet chosen: the record names the judge and the standard applied, and the story that resolves it will carry the `human-verified` label. What the keyword means, and why a fabricated metric would be worse, is in `.claude/skills/requirements-authoring/SKILL.md` § Verification methods.

`—` in `Resolved-by` means no story has been allocated yet. It never means the link is unknown: the single home for requirement → story is the first table in `TRACEABILITY.md`, and this column is a view of it, regenerated in the same pass. A retired record's row is the one exception, and what this column then reads is fixed by the freeze rule in `README.md` § How this folder works.
