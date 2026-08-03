# Traceability matrix

Requirement ↔ epic ↔ story, in both directions. The first table below is the **single home** for the requirement → story link: each record's `Resolved-by` field and the `Resolved-by` column in `INDEX.md` are views of it, regenerated in the same librarian pass. The second table is transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit this file.

A disagreement between the two directions is a finding for `@requirements-engineer`, never a silent reconciliation.

**Story allocation — 2 August 2026, in two passes.** `@requirements-story-organizer` cut three epics as GitHub Milestones and eleven stories as GitHub Issues for batch 1, following the Owner's sign-off of 1 August 2026, and then a fourth epic and six further stories for batch 2 — `FR-013` … `FR-018` — following its sign-off of 2 August 2026. Story creation follows sign-off as a pass of its own rather than sharing it, which is why the two allocations are separate events on the same day. Every one of `FR-001` … `FR-018` carries an epic and a story; both directions below were filled in the same librarian pass as each allocation, and they agree. Neither table holds anything about a story or an epic beyond its identifier — a story's number and its epic's Milestone name — per `.claude/skills/requirements-authoring/SKILL.md` § Corpus layout.

**No story allocation has run for batch 3.** `FR-019` … `FR-032` and `NFR-001` … `NFR-005` were filed on 3 August 2026 and are `draft`, awaiting the Owner's sign-off; each takes a row in the first table below with no epic and no story, which is what the one-row-per-non-retired-requirement rule requires and is not an orphan. The second table is unchanged by that filing — no story was cut, so no row could be transcribed.

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
| `FR-013` — Offer more than one journey alternative | A real choice of journey alternatives | #12 |
| `FR-014` — Do not offer two indistinguishable journey alternatives | A real choice of journey alternatives | #13 |
| `FR-015` — Do not offer an alternative beaten outright by another | A real choice of journey alternatives | #14 |
| `FR-016` — Offer alternatives that differ in ways a user would act on | A real choice of journey alternatives | #17 |
| `FR-017` — Do not withhold alternatives for a missing kind of alternative | A real choice of journey alternatives | #15 |
| `FR-018` — Do not withhold a compliant alternative for its lower Turf value | A real choice of journey alternatives | #16 |
| `FR-019` — Address every Turf API request to the current API version | — | — |
| `FR-020` — Derive no per-zone extent from the nominal zone size | — | — |
| `FR-021` — Measure the distance between two zones for that pair | — | — |
| `FR-022` — Refresh the local zone copy from a scheduled background job | — | — |
| `FR-023` — Resolve a route corridor's zones against the local copy | — | — |
| `FR-024` — Plan against a mid-refresh or stale local copy | — | — |
| `FR-025` — Build bounding-box requests against the permitted area product | — | — |
| `FR-026` — Answer no ownership question from the local zone copy | — | — |
| `FR-027` — Exclude blocktime from stop time | — | — |
| `FR-028` — Decide the user's own holdings by membership in the held-zone list | — | — |
| `FR-029` — Determine region lordship once, from a single region response | — | — |
| `FR-030` — Carry an absent ownership field as absent | — | — |
| `FR-031` — Do not read an absent ownership field as a zone never taken | — | — |
| `FR-032` — Assume no minimum distance between two zones | — | — |
| `NFR-001` — Hold outbound Turf calls within the API's published limits | — | — |
| `NFR-002` — Lower confidence as the Turf data behind a recommendation ages | — | — |
| `NFR-003` — Build the service as one self-contained executable | — | — |
| `NFR-004` — Run the service as one long-running process | — | — |
| `NFR-005` — Serve the client as static files | — | — |

37 of 37 requirements listed. **Eighteen carry an epic and at least one story; nineteen — batch 3, `FR-019` … `FR-032` and `NFR-001` … `NFR-005` — carry neither, and read `—` under both columns.** They are here because the rule is one row per non-retired requirement and a `draft` is non-retired: a requirement missing from this table is a filing error whatever its status, and a row added only once a story exists would make the table's completeness unverifiable exactly when it matters.

**The orphan rule is scoped to an epic in flight**: a requirement reading `—` under *Stories* while its epic is in flight is an orphan and a finding, and a requirement with no epic at all has not reached that test yet. **None of the nineteen is an orphan** — they have no epic, and story creation follows the Owner's sign-off as a pass of its own. The rule bites on them only once an epic is cut. Every requirement code carrying a story here also appears in *Story → requirement* below; a code with a story in one direction and no counterpart in the other is a finding, and a code reading `—` in this table is expected to be absent from the second.

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
| #12 | A real choice of journey alternatives | `FR-013` |
| #13 | A real choice of journey alternatives | `FR-014` |
| #14 | A real choice of journey alternatives | `FR-015` |
| #15 | A real choice of journey alternatives | `FR-017` |
| #16 | A real choice of journey alternatives | `FR-018` |
| #17 | A real choice of journey alternatives | `FR-016` |

17 of 17 stories listed, transcribed as filed. The batch-2 rows are not in requirement order, and that is how they were filed rather than a transcription slip: `#17` resolves `FR-016` and `#15` resolves `FR-017`. A requirement code appearing in more than one row here is not a disagreement between the two directions and is never normalized away: `FR-001` is resolved by two stories, and both tables record it. A story here naming a requirement code that has no row in `INDEX.md` — the list of every ID ever issued — is a finding, not a row to keep.
