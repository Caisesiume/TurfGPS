---
name: turfgps-board-ops
description: How to operate the TurfGPS Project Board (GitHub Project 3) and repo Caisesiume/TurfGPS from agents — gh CLI path, fresh field-ID resolution, the Status/Priority/Size fields, label and milestone conventions, the traceability law, and the PowerShell quoting pitfall. Use for any board read or mutation.
---

# TurfGPS Board Operations

## The CLI

`gh` is on PATH, but prefer the explicit path so behaviour does not depend on shell configuration:
```bash
GH="/c/Program Files/GitHub CLI/gh.exe"
"$GH" auth status    # verify first; abort with a clear report if unauthenticated
```

⚠️ **PowerShell 5.1 mangles embedded quotes in `gh api graphql` arguments. Use the Bash tool for all GraphQL and any command with nested quoting.**

## The board

- Project: **"TurfGPS Project Board" = project number 3**, owner `Caisesiume`. Project ID `PVT_kwHOBERr7s4Be_3r`.
- Repo for issues, PRs, and milestones: **`Caisesiume/TurfGPS`**.

**Resolve field and option IDs FRESH every run — never trust a cached ID.** This is not boilerplate caution: the Status options were regenerated on 31 July 2026 when `Ordered Revision` was added, and every option ID changed. Any agent holding an ID from before that moment would have written to a field that no longer existed.

```bash
"$GH" project field-list 3 --owner Caisesiume --format json
"$GH" project item-list 3 --owner Caisesiume --limit 100 --format json
"$GH" project item-edit --id <item-id> --project-id <project-id> \
      --field-id <status-field-id> --single-select-option-id <option-id>
```

## Fields

### Status (single-select) — the lifecycle

`Backlog` → `Ready` → `In progress` → `In review` → `Ordered Revision` → `Done`

Note the casing exactly: **`In progress`** and **`In review`** are lower-case after the first word. `Ready` is this board's name for what other boards call Todo.

Authority map: Backlog→Ready is the **scrum-master's** alone; workers set `In progress`/`In review` on their own item; the **pr-judge** sets `Ordered Revision`; `Done` requires merge evidence.

`Ordered Revision` exists so a remand is visible as its own column rather than hidden behind a label. An item sitting there **counts against its worker's WIP** — do not promote a replacement item for that worker — and revision preempts any new work.

### Priority (single-select) — `P0` · `P1` · `P2`

Set from the requirement's MoSCoW priority, which the RE assigns and the librarian records:

| MoSCoW | Priority |
|---|---|
| MUST | `P0` |
| SHOULD | `P1` |
| COULD | `P2` |
| WON'T-now | not filed as a story at all |

A story resolving several requirements takes the **highest** priority among them. The scrum-master promotes by Priority first and dependency order second — never the reverse, because promoting a P2 whose blockers happen to be clear ahead of a ready P0 is a sequencing bug, not efficiency.

### Size (single-select) — `XS` · `S` · `M` · `L` · `XL`

Set by the story-organizer as a sizing check, not an estimate. The rule is that **a story is one reviewable PR's worth of work**. An `L` is a warning and an `XL` is a defect: re-cut it into several stories under the same Epic before filing. The bench is expensive and the judge remands sprawl.

`Estimate`, `Start date`, and `Target date` exist on the board and are **not used by the loop**. Leave them empty rather than inventing values.

## Labels

All seven exist on the repo:

| Label | Meaning |
|---|---|
| `User Story` | A user story. Tied to an Epic (**Milestone**), body carries `Resolves: FR-*/NFR-*` |
| `Task` | Work item that is not a user story (process/infra/documentation) |
| `needs-re` | Worker-discovered problem awaiting RE tracing — must also carry `Task` and link relating stories + requirement codes |
| `human-verified` | The resolving requirement's verification method is human judgement; agent consensus cannot close it |
| `judge:approved` / `judge:remanded` | PRJudge ruling record (PR labels) |
| `awaiting-human` | Loop paused on a human decision |

**Auto-add is enabled.** The project's *Auto-add to project* workflow is on, so issues land on the board without a manual `project item-add`. Verify its filter in the UI before relying on label-based filtering — if it adds *every* issue rather than only labelled ones, the board will accumulate items the loop does not manage, and the scrum-master should report that rather than silently reconciling them.

**Epics are Milestones** — none exist yet. Create with:
```bash
"$GH" api repos/Caisesiume/TurfGPS/milestones -f title="<epic>" -f description="<requirement cluster>"
```

## Traceability law

**Source document § section → requirement code (`FR-*`/`NFR-*`) → story `#N` → commit message references `#N` → PR links the story.** The judge remands broken traceability before convening the bench.

Note the first link differs from a single-specification project: TurfGPS's requirements draw on **four** upstream documents, and a citation names the document as well as the section — `SPECIFICATION.md § Enforceable exclusions`, `CalculationSpecification.md § Proposed placeholder timings`, `Architecture.md § Retrieving zones`, `DESIGN.md § Replacement and escalating scope`. A citation naming only a section is ambiguous and is a librarian finding.
