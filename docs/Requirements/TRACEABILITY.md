# Traceability matrix

Requirement ↔ epic ↔ story, in both directions. The first table below is the **single home** for the requirement → story link: each record's `Resolved-by` field and the `Resolved-by` column in `INDEX.md` are views of it, regenerated in the same librarian pass. The second table is transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit this file.

A disagreement between the two directions is a finding for `@requirements-engineer`, never a silent reconciliation.

**First story allocation — 2 August 2026.** Story creation has run for batch 1. `@requirements-story-organizer` cut three epics as GitHub Milestones and eleven stories as GitHub Issues on 2 August 2026, following the Owner's sign-off of 1 August 2026, and every one of `FR-001` … `FR-012` carries a story. Both directions below were filled in the same pass and agree. **Batch 2 — `FR-013` … `FR-018`, filed 2 August 2026 — is deliberately not in that allocation**: those records are still `draft`, story creation follows sign-off and not filing, and they appear in the first table with no epic and no story so that the table keeps its row per requirement. Neither table holds anything about a story or an epic beyond its identifier — a story's number and its epic's Milestone name — per `.claude/skills/requirements-authoring/SKILL.md` § Corpus layout.

## Requirement → story

One row per non-retired requirement. Every requirement in the corpus appears here, whether or not a story exists for it — a requirement missing from this table is a filing error. A requirement that retires leaves it, taking the link's single home with it; what its `Resolved-by` then reads is fixed by the freeze rule in `README.md` § How this folder works.

| Requirement | Epic (Milestone) | Stories |
|---|---|---|
| `FR-001` — Plan a journey from an origin to a destination | Journey request and structure | #1, #2 |
| `FR-002` — Plan a journey with intermediate destinations | Journey request and structure | #2 |
| `FR-003` — Visit mandatory waypoints in the order entered | Journey request and structure | #2 |
| `FR-004` — Preserve the entered mandatory-waypoint set | Journey request and structure | #3 |
| `FR-005` — Require a destination | Journey request and structure | #4 |
| `FR-006` — Plan journey travel as travel by car | Drivable legs and journey cost | #6 |
| `FR-007` — Derive a journey's Turf value from the zones it captures | Turf value and per-user ranking | #9 |
| `FR-008` — Measure journey cost as time added to the journey without Turf stops | Drivable legs and journey cost | #7 |
| `FR-009` — Charge stop time to the journey even where no detour is driven | Drivable legs and journey cost | #8 |
| `FR-010` — Do not prefer a journey alternative for its zone count | Turf value and per-user ranking | #10 |
| `FR-011` — Balance value against cost from the individual user's preferences | Turf value and per-user ranking | #11 |
| `FR-012` — Require an intermediate destination on a return to the origin | Journey request and structure | #5 |
| `FR-013` — Offer more than one journey alternative | — | — |
| `FR-014` — Do not offer two indistinguishable journey alternatives | — | — |
| `FR-015` — Do not offer an alternative beaten outright by another | — | — |
| `FR-016` — Offer alternatives that differ in ways a user would act on | — | — |
| `FR-017` — Do not withhold alternatives for a missing kind of alternative | — | — |
| `FR-018` — Do not withhold a compliant alternative for its lower Turf value | — | — |

18 of 18 requirements listed. `FR-001` … `FR-012` each carry an epic and at least one story. `FR-013` … `FR-018` carry neither, and that is the expected state rather than a gap: they are batch 2, filed 2 August 2026 and not yet signed off, and `@requirements-story-organizer` is invoked only after sign-off — as it was for batch 1. **The orphan rule is scoped to an epic in flight**: a requirement reading `—` under *Stories* while its epic is in flight is an orphan and a finding, and a requirement with no epic at all has not reached that test yet. No batch-2 code appears in *Story → requirement*, and one appearing there before those records are signed off and stories are cut would be a finding.

## Story → requirement

One row per story, transcribed from the issue's `Resolves:` line.

**This table carries no board-status column, and none is to be added.** A story's state is owned by the board, which answers it live; a column here would be a second home for it, stale from the moment it is written. The rule this follows — what the corpus may hold about a story, and what it may never hold — is stated in `.claude/skills/requirements-authoring/SKILL.md` § Corpus layout.

| Story | Epic | Resolves |
|---|---|---|
| #1 | Journey request and structure | `FR-001` |
| #2 | Journey request and structure | `FR-001`, `FR-002`, `FR-003` |
| #3 | Journey request and structure | `FR-004` |
| #4 | Journey request and structure | `FR-005` |
| #5 | Journey request and structure | `FR-012` |
| #6 | Drivable legs and journey cost | `FR-006` |
| #7 | Drivable legs and journey cost | `FR-008` |
| #8 | Drivable legs and journey cost | `FR-009` |
| #9 | Turf value and per-user ranking | `FR-007` |
| #10 | Turf value and per-user ranking | `FR-010` |
| #11 | Turf value and per-user ranking | `FR-011` |

11 of 11 stories listed, transcribed as filed. A requirement code appearing in more than one row here is not a disagreement between the two directions and is never normalized away: `FR-001` is resolved by two stories, and both tables record it. A story here naming a requirement code that has no row in `INDEX.md` — the list of every ID ever issued — is a finding, not a row to keep.
