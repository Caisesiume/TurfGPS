---
name: turfgps-board-ops
description: How to operate the TurfGPS Project Board (GitHub Project 3) and repo Caisesiume/TurfGPS from agents — the two-channel rule (MCP for board and issues, gh CLI for judgments), fresh field-ID resolution, the Status/Priority/Size fields, label and milestone conventions, the traceability law, and the PowerShell quoting pitfall. Use for any board read or mutation.
---

# TurfGPS Board Operations

## ⚠️ Two channels, two identities — do not collapse them

This is a deliberate split, not an accident of tooling. **Anyone "simplifying" it breaks a requirement in `docs/DELIVERY.md`.**

| Channel | Used for | Identity |
|---|---|---|
| **GitHub MCP** (server `github`) | Board reads and item edits, issues, milestones, labels, PR reads | The repository owner, via `GITHUB_MCP_TOKEN` |
| **`gh` CLI** | **Judgments only** — PR review verdicts and review comments | The judge — login **`TheReviewNinja`** — via `GH_JUDGE_TOKEN` |

`docs/DELIVERY.md` requires review comments under a **separate GitHub identity** from the repository owner's, because authorship and approval must not share a signature — self-approval is not review, and a distinct identity makes the boundary visible in the history rather than merely intended.

**An MCP server carries one identity.** If the judge posted through the same MCP connection as the agent that wrote the code, that boundary disappears and nothing in the history would show it. So:

- **Never set `GITHUB_MCP_TOKEN` to the judge's token**, and never the reverse. Two tokens, two accounts, always.
- **@pr-judge issues its rulings through the CLI**, prefixed with `GH_TOKEN="$GH_JUDGE_TOKEN"`, even when the MCP is connected and would be more convenient.
- **The token is referenced by name only and must never be read, printed, logged, or echoed.** Pass it through the environment.

Everything that is *not* a judgment — the scrum-master's promotions, the story-organizer's issue creation, the coordinator's reads — should prefer the MCP. It avoids the shell-quoting hazards below entirely.

## The fallback rule — MCP is a convenience, never a dependency

**When the MCP is unavailable, errors, or cannot perform an action, fall back to the `gh` CLI and complete the task.** Do not stop, do not report a blocker, and do not wait for the connector.

The CLI is authenticated as **`Caisesiume`** and holds exactly the permissions the owner has through the web UI. Anything doable in the UI is doable through the CLI, so a failed MCP call is a transport problem, never an authorization one. Treat "the MCP did not work" as a reason to change tool, not a reason to halt the loop.

The ladder, in order:

1. **GitHub MCP** — preferred for board and issue work.
2. **`gh` CLI** — full fallback for all of it. Use the Bash tool.
3. **`gh api graphql`** — for the handful of things the porcelain cannot do at all, such as adding an option to a single-select field. Bash tool only; see the PowerShell warning below.

Note the ladder is not strictly a capability ordering: some operations only step 3 can perform. Reach down as far as the task needs.

Mention in your report which channel you used, so a persistently failing MCP shows up as a pattern rather than staying invisible behind successful fallbacks.

### The one thing the fallback must never do

**The fallback is a change of transport, not a change of identity.**

Falling back for a *judgment* means using the CLI **with `GH_TOKEN="$GH_JUDGE_TOKEN"`**, exactly as @pr-judge already does. A judgment posted through the plain CLI would appear as `Caisesiume` — the author approving their own work, which is the precise failure `docs/DELIVERY.md` exists to prevent, and it would look completely normal in the history.

If the judge token is missing or its call fails, that **is** a stop-and-report. There is no acceptable fallback for a ruling's identity.

**The identity is checkable — so check it.** The token is the input; the login is the evidence. Confirm a ruling landed under the judge by reading the review author back off the PR:

```bash
"$GH" pr view <N> --json reviews --jq '.reviews[].author.login'
```

Expect **`TheReviewNinja`**. A verdict showing `Caisesiume` is the exact failure this rule exists to catch — **stop and report it**; do not delete it, re-post over it, or treat it as a formality. The same check against `--json comments` covers the judgment comment, which `pr comment` also posts under whichever token was in the environment.

## The CLI

Configured project-scoped in `.mcp.json`, so the MCP travels with the repository; the CLI is the fallback and the judgment channel.

`gh` is on PATH, but prefer the explicit path so behaviour does not depend on shell configuration:
```bash
GH="/c/Program Files/GitHub CLI/gh.exe"
"$GH" auth status    # verify first; abort with a clear report if unauthenticated
```

⚠️ **PowerShell 5.1 mangles embedded quotes in `gh api graphql` arguments. Use the Bash tool for all GraphQL and any command with nested quoting.** This is the main reason to prefer the MCP for issue bodies: a story body carries a multi-line narrative, given/when/then criteria, and a `Resolves:` block, and heredoc quoting is the fragile part of filing it — not the content.

**Some operations have no CLI equivalent at all.** `gh project` cannot add an option to a single-select field; that needs raw GraphQL (`updateProjectV2Field`), which replaces the whole option list and regenerates every option ID. Reach for the MCP first.

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
