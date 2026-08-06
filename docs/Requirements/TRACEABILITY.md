# Traceability matrix

Requirement ↔ epic ↔ story, in both directions. The first table below is the **single home** for the requirement → story link: each record's `Resolved-by` field and the `Resolved-by` column in `INDEX.md` are views of it, regenerated in the same librarian pass. The second table is transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit this file.

A disagreement between the two directions is a finding for `@requirements-engineer`, never a silent reconciliation.

**Story allocation — 2 August 2026, in two passes.** `@requirements-story-organizer` cut three epics as GitHub Milestones and eleven stories as GitHub Issues for batch 1, following the Owner's sign-off of 1 August 2026, and then a fourth epic and six further stories for batch 2 — `FR-013` … `FR-018` — following its sign-off of 2 August 2026. Story creation follows sign-off as a pass of its own rather than sharing it, which is why the two allocations are separate events on the same day. Every one of `FR-001` … `FR-018` carries an epic and a story; both directions below were filled in the same librarian pass as each allocation, and they agree. Neither table holds anything about a story or an epic beyond its identifier — a story's number and its epic's Milestone name — per `requirements-authoring § Corpus layout`.

**Story allocation — 4 August 2026.** `@requirements-story-organizer` cut five further epics as GitHub Milestones and nineteen stories as GitHub Issues, `#18` … `#36`, for batch 3 — `FR-019` … `FR-034` and `NFR-001` … `NFR-005` — following the Owner's sign-off of the same day. Story creation follows sign-off as a pass of its own, which is why this is a separate event from the filing and the sign-off that preceded it. Both directions below were filled in the same librarian pass and they agree. **Every requirement filed up to that day — `FR-001` … `FR-034` and `NFR-001` … `NFR-005` — carries an epic and at least one story**, and on 4 August 2026 `—` appeared nowhere under `Resolved-by`: not on a record, not in `INDEX.md`, and not in either table here. That second half stopped being true of the corpus on 6 August 2026 and is left standing as what it was — the state batch 3 left behind — rather than edited into a claim about today.

**Story allocation — 6 August 2026.** `@requirements-story-organizer` cut three stories as GitHub Issues, `#38` … `#40`, for batch 3b — `FR-035` … `FR-037` — following the Owner's sign-off of the same day. **No new Milestone was cut**, and this is the first allocation of which that is true: all three stories sit under epics that already existed, `#38` under *The synced zone set, and how current it is* and `#39` with `#40` under *Zone facts the API actually provides*, so the corpus stands at nine epics still. Story creation follows sign-off as a pass of its own, which is why this is a separate event from the filing and the sign-off that preceded it. Both directions below were filled in the same librarian pass and they agree. **Between its filing and this allocation, both on 6 August 2026, batch 3b carried no epic and no story and read `—` under both columns of the first table** — the first `—` in this file, and in the corpus, since story creation caught up with filing on 4 August 2026, and it carried its one meaning throughout: *not yet allocated*, never *unknown*. That is left standing as what it was, the interval batch 3b spent in `draft`, rather than edited into a claim about today.

**Batch 4 filed and signed off, 6 August 2026 — no story allocation has run for it yet.** `FR-038` … `FR-054` were filed that day and signed off the same day, and they take a row apiece in the first table below reading `—` under both columns. This is the ordinary interval between a sign-off and the story pass that follows it: sign-off gates story creation, not the reverse, and story creation is a pass of its own rather than part of the signature, so no epic and no story exists for any of the seventeen at this moment. **The interval is the ordinary one, but the shape is new** — every previous batch's story pass ran on the day of its sign-off, so this is the first time the corpus has held signed-off requirements with no story behind them. It is not a gap and does not become one until an epic is cut; what would make it one is stated in the orphan rule below.

**`FR-055` … `FR-057` filed `draft`, 6 August 2026.** Three records from the three rulings that came with batch 4's signature, taking a row apiece reading `—` under both columns for the earlier of the two reasons: they are ahead of sign-off as well as ahead of any story. Nothing was added in the *Story → requirement* direction for either group, which carries one row per story and has no story to carry. The em-dash means *not yet allocated* and never *unknown* in all twenty cases.

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
| `FR-035` — Update dateCreated on zones the local copy already holds | The synced zone set, and how current it is | #38 |
| `FR-036` — Treat an absent zone type as ordinary data | Zone facts the API actually provides | #39 |
| `FR-037` — Determine a zone's country where the region subkey is absent | Zone facts the API actually provides | #40 |
| `NFR-001` — Hold outbound Turf calls within the API's published limits | A Turf client that stays inside the API's rules | #28 |
| `NFR-002` — Lower confidence as the Turf data behind a recommendation ages | The synced zone set, and how current it is | #36 |
| `NFR-003` — Build the service as one self-contained executable | The service's runtime and deployment shape | #19 |
| `NFR-004` — Run the service as one long-running process | The service's runtime and deployment shape | #25 |
| `NFR-005` — Serve the client as static files | The service's runtime and deployment shape | #26 |
| `FR-038` — Plan no journey without an additional-time limit the user entered | — | — |
| `FR-039` — Derive no additional-time limit from the journey's length | — | — |
| `FR-040` — Admit an additional-time limit anywhere in the realistic range | — | — |
| `FR-041` — Present neither end of the realistic range as unusual | — | — |
| `FR-042` — Search for alternatives whatever the size of the stated limit | — | — |
| `FR-043` — Offer a zone-capturing within-limit alternative where one was produced | — | — |
| `FR-044` — Derive the absolute ceiling from the stated additional time | — | — |
| `FR-045` — Offer no journey alternative above the absolute ceiling | — | — |
| `FR-046` — Reject a configured ceiling multiplier above the permitted maximum | — | — |
| `FR-047` — Present an above-limit alternative as a stretch alternative | — | — |
| `FR-048` — Name what the stated limit was exceeded for | — | — |
| `FR-049` — Offer an above-limit alternative only where value justifies it | — | — |
| `FR-050` — Allocate the budget across legs before optimizing any leg | — | — |
| `FR-051` — Size each leg's share in proportion to its baseline driving time | — | — |
| `FR-052` — Return an unused share to a pool for the remaining legs | — | — |
| `FR-053` — Test the limit and the ceiling against the sum across all legs | — | — |
| `FR-054` — Present the additional time for the journey as a whole | — | — |
| `FR-055` — Open the planner with no additional-time limit entered | — | — |
| `FR-056` — Refuse an additional-time limit that is not positive | — | — |
| `FR-057` — Impose no maximum admissible additional-time limit | — | — |

62 of 62 requirements listed. **Forty-two carry an epic and at least one story** — nine epics and thirty-nine stories, allocated in four passes — and **twenty read `—` under both columns**: batch 4's seventeen, signed off on 6 August 2026 with its story pass not yet run, and `FR-055` … `FR-057`, filed `draft` the same day. The rule stays one row per non-retired requirement rather than one row per requirement that has a story: a requirement missing from this table is a filing error whatever its status, and a row added only once a story exists would make the table's completeness unverifiable exactly when it matters. **The rule was first exercised by batch 3b** — between its filing and its story allocation, both on 6 August 2026, `FR-035` … `FR-037` were the first three rows this table had ever held with no story behind them — and batch 4 is the second and much larger instance of the same shape. It is also the first instance in which the requirements are **signed off** rather than `draft`, which changes nothing about the rows and everything about what happens next: the story pass is now due for those seventeen and is not due for the three.

**The orphan rule is scoped to an epic in flight**: a requirement reading `—` under *Stories* while its epic is in flight is an orphan and a finding, and a requirement with no epic at all has not reached that test yet. **Twenty rows read `—` today and none of them is an orphan**, because each also reads `—` under *Epic (Milestone)* — no epic has been cut for batch 4 and none for `FR-055` … `FR-057`, so none of the twenty has reached the test. The rule bites on the first of them to still read `—` under *Stories* once its epic is cut. **Sign-off does not bite either, and that is worth stating now that seventeen of the twenty carry it**: the test is an epic in flight, not a signature, so a signed-off requirement with no epic is exactly as far from being an orphan as a `draft` one. `FR-035` … `FR-037` held the same shape for part of 6 August 2026 and none of the three was ever an orphan, for the same reason. Every requirement code here that names a story appears in the second table below, and every code there appears here — a code with a story in one direction and no counterpart in the other is a finding. A requirement with no story is absent from that table by construction and is not that finding.

## Story → requirement

One row per story, transcribed from the issue's `Resolves:` line.

**This table carries no board-status column, and none is to be added.** A story's state is owned by the board, which answers it live; a column here would be a second home for it, stale from the moment it is written. The rule this follows — what the corpus may hold about a story, and what it may never hold — is stated in `requirements-authoring § Corpus layout`.

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
| #38 | The synced zone set, and how current it is | `FR-035` |
| #39 | Zone facts the API actually provides | `FR-036` |
| #40 | Zone facts the API actually provides | `FR-037` |

39 of 39 stories listed, transcribed as filed. **The numbers run to `#40` and are not contiguous, which is not a missing row.** This table carries one row per story, and the board numbers stories alongside items that are not stories and resolve no requirement; a number absent from this column says only that the board issued it to something this table does not carry. What *would* be a missing row is a story whose `Resolves:` line names a requirement and has no entry here. Neither the batch-2 nor the batch-3 rows run in requirement order, and that is how they were filed rather than a transcription slip: `#17` resolves `FR-016` and `#15` resolves `FR-017`; `#22` resolves `FR-030` and `#34` resolves `FR-021`. A requirement code appearing in more than one row here is not a disagreement between the two directions and is never normalized away: `FR-001` is resolved by two stories, and both tables record it. A row naming two codes is the same case seen from the other side — `#29` and `#33` each resolve two records — and is likewise transcribed as it stands rather than split into a row apiece. A story here naming a requirement code that has no row in `INDEX.md` — the list of every ID ever issued — is a finding, not a row to keep.
