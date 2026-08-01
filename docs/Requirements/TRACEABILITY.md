# Traceability matrix

Requirement ↔ epic ↔ story, in both directions. The first table below is the **single home** for the requirement → story link: each record's `Resolved-by` field and the `Resolved-by` column in `INDEX.md` are views of it, regenerated in the same librarian pass. The second table is transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit this file.

A disagreement between the two directions is a finding for `@requirements-engineer`, never a silent reconciliation.

**Both tables are empty of story links on purpose — 1 August 2026.** No epic and no story exists. The Owner signed the corpus off on 1 August 2026, so story creation is now **authorized and not yet run**: it awaits `@engineering-lead`'s verification that the sign-off conditions have been applied, after which `@requirements-story-organizer` cuts the first Epics. The shape below is the finished shape; only rows are missing. A `—` is an allocation that has not happened yet, not a link that was lost.

## Requirement → story

One row per non-retired requirement. Every requirement in the corpus appears here, whether or not a story exists for it — a requirement missing from this table is a filing error. A requirement that retires leaves it, taking the link's single home with it; what its `Resolved-by` then reads is fixed by the freeze rule in `README.md` § How this folder works.

| Requirement | Epic (Milestone) | Stories |
|---|---|---|
| `FR-001` — Plan a journey from an origin to a destination | — | — |
| `FR-002` — Plan a journey with intermediate destinations | — | — |
| `FR-003` — Visit mandatory waypoints in the order entered | — | — |
| `FR-004` — Preserve the entered mandatory-waypoint set | — | — |
| `FR-005` — Require a destination | — | — |
| `FR-006` — Plan journey travel as travel by car | — | — |
| `FR-007` — Derive a journey's Turf value from the zones it captures | — | — |
| `FR-008` — Measure journey cost as time added to the journey without Turf stops | — | — |
| `FR-009` — Charge stop time to the journey even where no detour is driven | — | — |
| `FR-010` — Do not prefer a journey alternative for its zone count | — | — |
| `FR-011` — Balance value against cost from the individual user's preferences | — | — |
| `FR-012` — Require an intermediate destination on a return to the origin | — | — |

12 of 12 requirements listed. 0 carry an epic. 0 carry a story. Once an epic is in flight, a requirement in it still reading `—` under *Stories* is an orphan and a finding.

## Story → requirement

One row per story, transcribed from the issue's `Resolves:` line.

| Story | Epic | Resolves | Board status |
|---|---|---|---|

**No rows, and that is correct.** No issue exists to transcribe. The first story filed adds the first row here and, in the same pass, the matching story number in the table above. A story here naming a requirement code that has no row in `INDEX.md` — the list of every ID ever issued — is a finding, not a row to keep.
