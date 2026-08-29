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

## Scoped retrieval

*§45.* **A full board dump is never context.** Ask for the items a pass will act on and project only the fields it reads: `item-list` takes `--limit` and `--jq`, and the filtering happens in `gh`'s built-in jq engine before anything reaches a context window. Standalone `jq` is not installed on this machine — piping to it fails.

**Run `scripts/loop/fingerprint.sh <your-own-agent-name>` before any board read you initiate.** It reports the board component `UNCHANGED` or `CHANGED` deterministically (exit `0`/`10`, `2` degraded), and an unchanged board is a reason not to read the board at all. **Never call it bare** — the argument names your consumer, and every caller that omits it shares one state file, so the first reader spends the change for all of them. A read you were *dispatched* to perform is already gated by the `trigger:` block you carry (`agent-handoffs § The trigger block`): process that, do not re-poll.

Four statuses matter to a routine pass: **`Ready`** (promotable), **`In progress`** (WIP), **`Ordered Revision`** (preempts everything), and anything labelled **`awaiting-human`** (paused). The rest is history, and history is retrieved when a question needs it.

```bash
# status-filtered, projected to what the pass actually reads
"$GH" project item-list 3 --owner Caisesiume --limit 100 --format json \
      --jq '[.items[] | select(.status == "Ready") | {id, title, status, number: .content.number}]'

# one known item, fetched by number — never a list scanned to find it
"$GH" issue view <N> --json number,title,labels,milestone,body
```

Locality follows the role: the coordinators read the statuses they sequence, @worker-manager one assigned item, @pr-judge one PR.

## Fields

### Status

Single-select, and the item's lifecycle:

`Backlog` → `Ready` → `In progress` → `In review` → `Ordered Revision` → `Done`

Note the casing exactly: **`In progress`** and **`In review`** are lower-case after the first word. `Ready` is this board's name for what other boards call Todo.

Authority map: Backlog→Ready is the **scrum-master's** alone; workers set `In progress`/`In review` on their own item; the **pr-judge** sets `Ordered Revision`; `Done` requires merge evidence.

**Creating a work item is not finished until that item is on the board with a `Status`** — auto-add lands it with `Status` empty, not in the entry column, so **the filing agent sets `Backlog`**, whichever channel it filed through, MCP or CLI or raw API. An item nobody sets is not sitting at the head of the chain above — it is outside the chain entirely: invisible to a Backlog-filtered view, and unsequenceable by the scrum-master who would otherwise be the one to promote it. The party that filed the issue is the party that knows it exists, so it is that agent's to set, in the same pass as the labels and the milestone. **An agent that cannot set it does not simply leave it: it names the issue in its envelope and routes the board step to `@scrum-master`.** Reporting an item as created while it holds no board state, with nothing owed to anyone, files an issue and creates no work.

**Setting `Backlog` is not a promotion, and the authority map above is untouched by it.** It records the column the item is already in rather than moving it out of one; Backlog→Ready remains the scrum-master's alone, and a filing agent that writes `Backlog` has moved nothing and decided nothing about sequence. Same shape as the blocker rule under `§ Priority` below: a field left empty does not read as *not yet decided*, it reads as *sorts nowhere*, and the item is passed over rather than queued.

Applied by `@requirements-story-organizer`, which has been setting it on every filing and reporting that it did so rather than doing it silently. Recorded on 7 August 2026. **Widened on 16 August 2026 to every filing agent**, after `@engineering-lead` created issue `#108` through the API and set no Status: the issue existed, and every board query returned a board without it.

`Ordered Revision` exists so a remand is visible as its own column rather than hidden behind a label. An item sitting there **counts against its worker's WIP** — do not promote a replacement item for that worker — and revision preempts any new work.

#### The six columns carry a ten-state lifecycle

`docs/adr/ADR-0001-artifact-driven-agent-org.md § D7` ratifies the richer lifecycle of the Owner directive **as a mapping onto these six columns**, not as new columns:

| Lifecycle state | Column |
|---|---|
| Backlog | `Backlog` |
| Requirements Ready · Ready for Implementation | `Ready` |
| In Progress | `In progress` |
| Implementation Complete · Review | `In review` |
| Revision Required | `Ordered Revision` |
| Review Passed · Ready to Merge · Merged | `Done` — set at merge, not at approval |

**No schema churn, deliberately.** Finer states would have to be branched on by some agent to be worth their cost, and no agent branches on the difference between *Requirements Ready* and *Ready for Implementation* — the promotion rule is the same. The Status options were regenerated once already, on 31 July 2026, and every option ID changed with them; a second regeneration to gain names nothing reads would be cost without benefit.

Two consequences worth stating, because they are where the mapping is lossy: an item in `In review` may be either awaiting a panel or mid-panel — the PR's review ledger says which, not the board. And `Done` means **merged**; a judge approval with no merge is still `In review`.

### Priority

Single-select: `P0` · `P1` · `P2`.

Set from the requirement's MoSCoW priority, which the RE assigns and the librarian records:

| MoSCoW | Priority |
|---|---|
| MUST | `P0` |
| SHOULD | `P1` |
| COULD | `P2` |
| WON'T-now | not filed as a story at all |

A story resolving several requirements takes the **highest** priority among them. The scrum-master promotes by Priority first and dependency order second — never the reverse, because promoting a P2 whose blockers happen to be clear ahead of a ready P0 is a sequencing bug, not efficiency. That dependency order is **read off the persisted edges**, never re-derived at promotion time — see `§ The dependency representation`.

**A blocker takes the priority of the highest-priority item it blocks.** The table above derives Priority from MoSCoW, and an item with no requirement behind it — a `Task`, typically — has nothing to derive from. Left empty it sorts below every item it gates, so under *Priority first, dependency order second* it would be reached only after every `P0` had been passed over as blocked, each of them blocked on it. The invariant this protects is that **no item ever sorts below something it gates**, which is also why the priority is read off the set of items the blocker currently blocks rather than stamped once and left.

Applied by `@requirements-engineer` filing Task `#37` at `P0` inherited from `#25`, and stated in that issue as a rule not currently written down rather than applied silently. Adopted by the Owner on 4 August 2026.

### Size

Single-select: `XS` · `S` · `M` · `L` · `XL`.

Set by the story-organizer as a sizing check, not an estimate. The rule is that **a story is one reviewable PR's worth of work**. An `L` is a warning and an `XL` is a defect: re-cut it into several stories under the same Epic before filing. The bench is expensive and the judge remands sprawl.

`Estimate`, `Start date`, and `Target date` exist on the board and are **not used by the loop**. Leave them empty rather than inventing values.

## Labels

| Label | Meaning |
|---|---|
| `User Story` | A user story. Tied to an Epic (**Milestone**), body carries `Resolves: FR-*/NFR-*` |
| `Task` | Work item that is not a user story (process/infra/documentation) |
| `needs-re` | Worker-discovered problem awaiting RE tracing — must also carry `Task` and link relating stories + requirement codes |
| `human-verified` | The resolving requirement's verification method is human judgement; agent consensus cannot close it |
| `judge:approved` / `judge:remanded` | PRJudge ruling record (PR labels) |
| `awaiting-human` | Loop paused on a human decision |
| `risk:low` · `risk:medium` · `risk:high` | **PR labels.** The tier `@change-risk-assessor` returned for the diff, applied by `@pr-judge` at PR open |

The first seven exist on the repo; **the three `risk:*` labels must be created before the first PR is judged** — `"$GH" label create "risk:low" --description "..."` and so on.

**The risk tier is a label rather than a comment because it decides things other agents read.** It sets the revision budget (3 normally, 5 on `risk:high`), it drives the mandatory reviewer set, and it lets `@engineering-lead` see review load across the board without opening every PR. Applied by the judge from the assessor's PR-open output, which is authoritative over any intake prediction; re-applied if a force-push changes the diff's character.

**A `Task` sequences and promotes on the same rules as a story.** It is read for Priority and dependency order exactly as a story is, and `§ Priority` above is written for it specifically — the blocker rule exists because a `Task` has no MoSCoW to derive a priority from. Nothing about the lifecycle differs.

**Its three differences from a story are exemptions, not omissions.** A `Task` carries **no `Resolves:` line**, because no requirement stands behind it; it **joins no Milestone**, because a Milestone is an Epic and an Epic is a cluster of requirements; and it is **exempt from the coverage audit** that @requirements-story-organizer runs in both directions. That last one is the load-bearing statement: a `Task` carrying no requirement codes is **not a gap in coverage** — it is an item the audit does not range over at all. The audit is already scoped to stories by its own wording, and this records that the scoping is deliberate, so that the next reader tightening the audit does not "fix" it into reporting every `Task` as an uncovered story.

**So a `Task` missing a label, milestone, or `Resolves:` block is not a traceability defect.** That test applies to `User Story` items alone. Applying it to a `Task` manufactures a defect out of a design decision, and the traceability law below is unaffected either way — a `Task` never enters the chain it describes. **Its commits still reference its own `#N`**: the three exemptions above are exhaustive, and what that link buys is attribution — which item this commit was done for — not the requirements-tracing a `Task` stays out of.

Ratified by the Owner on 4 August 2026.

### The review ledger is a PR comment, not a board field

`@pr-judge` keeps one **review ledger** comment per PR — reviewer, domain, verdict, confidence, the diff SHA each verdict was issued against, and the cycle — updated every revision cycle, with carried-forward verdicts marked `carried (SHA)`. Its format and the incremental-validity rules that govern it live in `review-board-dispatch § Incremental review validity`; do not restate them here or on the board.

It is a comment because it is **per-cycle history**, and the board holds current state. Trying to carry it in board fields would need a field per reviewer and would still lose the SHA — and the SHA is the whole point, since a carried verdict is a claim that someone who was not there should be able to check.

Like every judgment artifact it is posted under `GH_JUDGE_TOKEN` and signed `/ The Review Ninja`.

**Auto-add is enabled.** The project's *Auto-add to project* workflow is on, so issues land on the board without a manual `project item-add`. Verify its filter in the UI before relying on label-based filtering — if it adds *every* issue rather than only labelled ones, the board will accumulate items the loop does not manage, and the scrum-master should report that rather than silently reconciling them.

**Epics are Milestones** — none exist yet. Create with:
```bash
"$GH" api repos/Caisesiume/TurfGPS/milestones -f title="<epic>" -f description="<requirement cluster>"
```

## The dependency representation

*Ratified in `docs/adr/ADR-0003-backlog-dependency-planner.md § P2`. This is the **one authoritative home** for the format: every agent that writes or reads a dependency cites this section rather than restating the grammar, so there is one place to change when it changes.*

A story's dependencies live in a **`## Dependencies` section in the issue body**, and nowhere else. One store, because a second one is the one that disagrees.

```markdown
## Dependencies

Blocked by: #41 — the entered limit must exist before this story can bind what happens to it
Soft dependency: #49 — shares the review-card surface; cheaper to land after it
Basis: FR-038 · `Architecture.md § Ports and adapters`
```

- **`Blocked by: #N — <one-line reason>`** is a **hard** edge. Several blockers are several lines, or `Blocked by: #7, #41` with one reason line per blocker below it.
- **`Soft dependency: #N — <one-line reason>`** is a **soft** edge: preferable order that **never blocks readiness**. No agent may read one as a gate.
- **`Basis:`** is optional, at most one line per edge — requirement codes and `Document.md § Section` citations. Enough that a later agent can check the edge without recomputing the plan; not the reasoning that produced it.

**Where the declared list ends.** On a `Blocked by:` line the **declared** blockers are the run of `#N` references that **opens** the line. The run continues across **glue** — anything between two references that is punctuation and space, with or without the conjunction `and` — and it **ends at the first word**, because a word is where the reason begins. So `Blocked by: #7 · #41 — the limit must exist before this binds it` declares `#7` and `#41`, and every `#N` after that first word is prose: a reason may name any issue or pull request without gating on it. Separators are deliberately **not** enumerated — the closed side of the split is the word, and enumerating five of them is what once dropped `#41` from that very line, on this repo's own house separator (#138, corrected in `ADR-0003 § A6`). Where the boundary is ambiguous it resolves toward **more** edges, never fewer: an over-read blocker holds a story where someone can see it, an under-read one promotes it with nothing to look at. `scripts/loop/dependents.sh` implements this rule and cites this section rather than restating it.

**Where it begins, and what an empty list declares.** The run opens the line's **content**, and the content starts after the label. A `Blocked by` label may carry **ornamentation** — punctuation, a bracketed aside, or a word naming what kind of reference follows — and the list behind it is still declared: `Blocked by issues #7 and #41` declares `#7` and `#41`, `Blocked by: PR #67` declares `#67`, and `Blocked by (see below): #7` declares `#7`. What ornamentation may **not** contain is the reason, and the reason begins at the **dash** — the `—` the form above already uses, and its two other spellings `–` and `--`. So **a line carrying no reference before that dash declares nothing at all**, however many its reason names: `Blocked by: none — superseded by PR #67` and `Blocked by: none.` are both empty lists. This is the one place the resolve-toward-more-edges tie-break does **not** reach, because there is no ambiguity to resolve — the list is empty and an edge read out of `none` is invented rather than over-read, and it holds the story **permanently**, a line reading `none` being one nobody ever comes back to revise. The plain ASCII hyphen is deliberately **not** a reason marker: it is ambiguous with hyphenation and with a separator, and ambiguity resolves toward more edges. What that leaves open is an empty declaration whose reason carries no dash at all (`Blocked by: superseded by PR #67`, which still reads as declaring `#67`); no line on the board is of that shape, and closing it would mean enumerating the ornament vocabulary — the enumeration this boundary was re-cut to be rid of.

**Every edge carries a concrete reason**, because an edge nobody can verify is the edge nobody dares delete — which is how a backlog quietly becomes a serial chain.

**A planned story with no edges says so: `No dependencies.`** New stories are filed with `_Pending @backlog-dependency-planner._` in this section, and **the placeholder is an explicit blocking state, not an empty graph** — unplanned is not unblocked. A story promotes only once the planner has replaced the placeholder, with edges or with `No dependencies.`; a placeholder that outlives its batch's planning pass is itself a `dependency_finding`.

**Hard versus soft is a throughput decision, not a nicety.** Hard means the downstream story must not reach `Ready` until the upstream one is `Done`; soft means preference only. With one edge type every preference is stated as a gate, and the board serializes work that could have run in parallel.

### Satisfied is not removed

*Ratified by directive 4 §8–§11 into `ADR-0003`.* A hard edge whose prerequisite is **successfully complete** is **satisfied**: it stops gating, and **its line stays in the body as provenance**. It still answers why B followed A, what capability B consumed, and what historically depended on A — which is what makes a later graph validation, an impact analysis, or a debugging pass possible at all.

**Satisfaction is derived, never written.** No `satisfied:` flag, no strike-through, no rewritten line: an edge is satisfied exactly when its prerequisite has landed — an **issue closed as completed**, or a **pull request merged** — and that fact already lives in GitHub. A declared blocker may be either kind: issue and pull-request numbers share one sequence, so a reference resolves to whichever it names (`ADR-0003 § A6`). A written mirror of GitHub state is the copy that goes stale silently and then gets believed — the same reason no agent here keeps a task list beside the board. `scripts/loop/dependents.sh` derives it deterministically, and **closed as `not planned` or as a duplicate is not completion, and neither is a pull request closed without merging** — such a prerequisite still blocks, and the mismatch is a `dependency_finding` rather than a promotion.

**Removal is a different operation with a different owner.** An edge is removed **only by @backlog-dependency-planner**, and only when the *relationship itself* no longer holds — scope changed, architecture moved the boundary, the decomposition changed, the edge was wrong, the prerequisite is no longer required. That removal is recorded in the pass's `graph_update`. Deleting an edge because its prerequisite merged destroys provenance to record something already true; keeping an edge whose basis is gone gates work on a relationship nobody believes.

> **Completion changes whether a dependency blocks; structural change changes whether it exists.**

### Who writes, who reads

**@backlog-dependency-planner writes this section and is the only agent that may** — through the GitHub MCP, per `§ Two channels, two identities — do not collapse them` above; an edge is not a judgment and never goes out under the judge's token. It writes the edges only: the narrative, the acceptance criteria, the `Resolves:` line, the labels and the fields belong to other owners.

Readers: **@scrum-master** reads the hard edges for readiness · **@project-coordinator** consumes the Ready queue's order and never reads edges to rebuild it · **`scripts/loop/dependents.sh`** reads them deterministically after a merge. Any agent that spots a wrong, missing, or stale edge files a `dependency_finding` (`handoff-payloads § Dependency findings and graph updates`) and **edits nothing**.

### The grandfather clause

The `## Dependencies` sections already on the board — **59 sections, 53 carrying `Blocked by:` lines**, measured 2026-08-10 — predate this grammar and are **valid hard edges exactly as written**, reasons in prose below the line rather than on it. No mass migration is owed: the planner brings a subgraph up to the grammar the first time an event touches it. Rewriting 59 issue bodies to gain a punctuation mark would touch every story on the board at once to change nothing any agent reads differently.

### Why not native GitHub dependencies

The native issue-dependency API is live and readable here — `gh api repos/Caisesiume/TurfGPS/issues/43/dependencies/blocked_by` returned `200` and `[]` on 2026-08-10. It is **deferred, not rejected**: the body convention already exists at scale, carries provenance and the soft type — for which the native relation has no field — and is greppable without an API call. Migration stays available if the tooling gains those two things, and `ADR-0003 § P2` owns that record.

## Traceability law

**Source document § section → requirement code (`FR-*`/`NFR-*`) → story `#N` → commit message references `#N` → PR links the work item.** The judge remands broken traceability before convening the bench.

**The last two links bind every work item; the first three are the story chain.** A `Task` stays out of the requirements half — that is `§ Labels` — but its commits still reference its own `#N` and its PR still links it. So the test a judge applies to a Task-driven PR is the tail of this chain, not its absence.

Note the first link differs from a single-specification project: TurfGPS's requirements draw on **four** upstream documents, and a citation names the document as well as the section — `SPECIFICATION.md § Enforceable exclusions`, `CalculationSpecification.md § Proposed placeholder timings`, `Architecture.md § Retrieving zones`, `DESIGN.md § Replacement and escalating scope`. A citation naming only a section is ambiguous and is a librarian finding.
