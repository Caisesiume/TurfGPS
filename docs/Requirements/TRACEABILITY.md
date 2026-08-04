# Traceability matrix

Requirement ↔ epic ↔ story, in both directions. The first table below is the **single home** for the requirement → story link: each record's `Resolved-by` field and the `Resolved-by` column in `INDEX.md` are views of it, regenerated in the same librarian pass. The second table is transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit this file.

A disagreement between the two directions is a finding for `@requirements-engineer`, never a silent reconciliation.

**Story allocation — 2 August 2026, in two passes.** `@requirements-story-organizer` cut three epics as GitHub Milestones and eleven stories as GitHub Issues for batch 1, following the Owner's sign-off of 1 August 2026, and then a fourth epic and six further stories for batch 2 — `FR-013` … `FR-018` — following its sign-off of 2 August 2026. Story creation follows sign-off as a pass of its own rather than sharing it, which is why the two allocations are separate events on the same day. Every one of `FR-001` … `FR-018` carries an epic and a story; both directions below were filled in the same librarian pass as each allocation, and they agree. Neither table holds anything about a story or an epic beyond its identifier — a story's number and its epic's Milestone name — per `.claude/skills/requirements-authoring/SKILL.md § Corpus layout`.

**Story allocation — 4 August 2026.** `@requirements-story-organizer` cut five further epics as GitHub Milestones and nineteen stories as GitHub Issues, `#18` … `#36`, for batch 3 — `FR-019` … `FR-034` and `NFR-001` … `NFR-005` — following the Owner's sign-off of the same day. Story creation follows sign-off as a pass of its own, which is why this is a separate event from the filing and the sign-off that preceded it. Both directions below were filled in the same librarian pass and they agree. **Every requirement in the corpus now carries an epic and at least one story**, and `—` appears nowhere under `Resolved-by` — on a record, in `INDEX.md`, or in either table here.

## Requirement → story

One row per non-retired requirement. Every requirement in the corpus appears here, whether or not a story exists for it — a requirement missing from this table is a filing error. A requirement that retires leaves it, taking the link's single home with it; what its `Resolved-by` then reads is fixed by the freeze rule in `README.md § How this folder works`.

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
| `FR-019` — Address every Turf API request to the current API version | A Turf client that stays inside the API's rules | #18 |
| `FR-020` — Derive no per-zone extent from the nominal zone size | Zone facts the API actually provides | #29 |
| `FR-021` — Measure the distance between two zones for that pair | Zone facts the API actually provides | #34 |
| `FR-022` — Refresh the local zone copy from a scheduled background job | The synced zone set, and how current it is | #20 |
| `FR-023` — Resolve a route corridor's zones against the local copy | The synced zone set, and how current it is | #27 |
| `FR-024` — Plan against a mid-refresh or stale local copy | The synced zone set, and how current it is | #33 |
| `FR-025` — Build bounding-box requests against the permitted area product | A Turf client that stays inside the API's rules | #21 |
| `FR-026` — Answer no ownership question from the local zone copy | Zone facts the API actually provides | #29 |
| `FR-027` — Exclude blocktime from stop time | The player's own Turf state | #32 |
| `FR-028` — Decide the user's own holdings by membership in the held-zone list | The player's own Turf state | #23 |
| `FR-029` — Determine region lordship once, from a single region response | The player's own Turf state | #24 |
| `FR-030` — Carry an absent ownership field as absent | Zone facts the API actually provides | #22 |
| `FR-031` — Do not read an absent ownership field as a zone never taken | Zone facts the API actually provides | #31 |
| `FR-032` — Assume no minimum distance between two zones | Zone facts the API actually provides | #30 |
| `FR-033` — Plan no journey for lack of zone data only where no copy is held | The synced zone set, and how current it is | #33 |
| `FR-034` — Record a result as built from data stale beyond the bound | The synced zone set, and how current it is | #35 |
| `NFR-001` — Hold outbound Turf calls within the API's published limits | A Turf client that stays inside the API's rules | #28 |
| `NFR-002` — Lower confidence as the Turf data behind a recommendation ages | The synced zone set, and how current it is | #36 |
| `NFR-003` — Build the service as one self-contained executable | The service's runtime and deployment shape | #19 |
| `NFR-004` — Run the service as one long-running process | The service's runtime and deployment shape | #25 |
| `NFR-005` — Serve the client as static files | The service's runtime and deployment shape | #26 |

39 of 39 requirements listed, and **every one carries an epic and at least one story** — nine epics and thirty-six stories, allocated in three passes. `—` appears under neither column. The rule stays one row per non-retired requirement rather than one row per requirement that has a story: a requirement missing from this table is a filing error whatever its status, and a row added only once a story exists would make the table's completeness unverifiable exactly when it matters.

**The orphan rule is scoped to an epic in flight**: a requirement reading `—` under *Stories* while its epic is in flight is an orphan and a finding, and a requirement with no epic at all has not reached that test yet. **No requirement reads `—` under either column today**, so the rule has nothing to bite on in this table; it stays stated because it bites again on the first record filed after its epic is cut. Every requirement code here appears in *Story → requirement* below, and every code there appears here — a code with a story in one direction and no counterpart in the other is a finding.

## Story → requirement

One row per story, transcribed from the issue's `Resolves:` line.

**This table carries no board-status column, and none is to be added.** A story's state is owned by the board, which answers it live; a column here would be a second home for it, stale from the moment it is written. The rule this follows — what the corpus may hold about a story, and what it may never hold — is stated in `.claude/skills/requirements-authoring/SKILL.md § Corpus layout`.

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
| #18 | A Turf client that stays inside the API's rules | `FR-019` |
| #19 | The service's runtime and deployment shape | `NFR-003` |
| #20 | The synced zone set, and how current it is | `FR-022` |
| #21 | A Turf client that stays inside the API's rules | `FR-025` |
| #22 | Zone facts the API actually provides | `FR-030` |
| #23 | The player's own Turf state | `FR-028` |
| #24 | The player's own Turf state | `FR-029` |
| #25 | The service's runtime and deployment shape | `NFR-004` |
| #26 | The service's runtime and deployment shape | `NFR-005` |
| #27 | The synced zone set, and how current it is | `FR-023` |
| #28 | A Turf client that stays inside the API's rules | `NFR-001` |
| #29 | Zone facts the API actually provides | `FR-020`, `FR-026` |
| #30 | Zone facts the API actually provides | `FR-032` |
| #31 | Zone facts the API actually provides | `FR-031` |
| #32 | The player's own Turf state | `FR-027` |
| #33 | The synced zone set, and how current it is | `FR-024`, `FR-033` |
| #34 | Zone facts the API actually provides | `FR-021` |
| #35 | The synced zone set, and how current it is | `FR-034` |
| #36 | The synced zone set, and how current it is | `NFR-002` |

36 of 36 stories listed, transcribed as filed. Neither the batch-2 nor the batch-3 rows run in requirement order, and that is how they were filed rather than a transcription slip: `#17` resolves `FR-016` and `#15` resolves `FR-017`; `#22` resolves `FR-030` and `#34` resolves `FR-021`. A requirement code appearing in more than one row here is not a disagreement between the two directions and is never normalized away: `FR-001` is resolved by two stories, and both tables record it. A row naming two codes is the same case seen from the other side — `#29` and `#33` each resolve two records — and is likewise transcribed as it stands rather than split into a row apiece. A story here naming a requirement code that has no row in `INDEX.md` — the list of every ID ever issued — is a finding, not a row to keep.
