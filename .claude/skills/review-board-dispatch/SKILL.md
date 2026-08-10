---
name: review-board-dispatch
description: Mechanics for safely convening TurfGPS's reviewers — the read-only clause, tree-integrity verification, sequencing, review identity, the evidence obligation, and the reviewer registry that decides who is convened at all. Use whenever dispatching any reviewer.
---

# Review Board Dispatch — Safe Convening Mechanics

`docs/DELIVERY.md` is the authority on what a verdict means and who must be convened. This skill is the authority on how a reviewer is convened without corrupting the tree, and it holds the registry that `DELIVERY.md` delegates to. Where the two appear to disagree, `DELIVERY.md` wins and the disagreement is a defect in this file.

## The read-only clause (learned the hard way)

Critics have been observed **mutation-testing in place and corrupting the shared tree**. Every reviewer dispatch prompt MUST contain, verbatim:

> *"You must not modify, create, or delete any file. Report only."*

Belt-and-braces around the whole board:
```bash
# before convening: fingerprint the tree
git -C <tree> status --porcelain > /tmp/pre_board_status.txt
git -C <tree> stash list > /tmp/pre_board_stash.txt
# after all reviewers return: verify nothing moved
git -C <tree> status --porcelain | diff /tmp/pre_board_status.txt - && echo TREE-CLEAN
```
If the tree changed, the board run is **invalid**: restore, identify the mutating reviewer, re-run.

## Sequencing

1. Convened critics run **parallel within a board**; boards may run in parallel with each other.
2. **ValidationAgent runs LAST and ALONE** — it executes builds/tests that must not race a critic's probing.
3. Review the **PR diff against `main`**, checked out in a dedicated review worktree (never the trunk tree, never a worker's worktree).

## The case file (same for every reviewer) — references, not content

PR number and the story it links · acceptance criteria and requirement codes · the head SHA under review · files modified · **safety paths touched** (see `safety-path-checklist`) · where the gate results are (see `local-gates`) · the review worktree path · the read-only clause.

**Send the references; the reviewer opens the artifacts itself.** Pasting the diff, the requirement text, or the PR body into the dispatch pays for the same bytes twice and hands the reviewer a copy that can already be stale. It also quietly invites the failure the next section exists to prevent.

## A reviewer does not accept a claim it could check

Everything in the case file above is a **claim**. None of it is evidence, and a reviewer accepts none of it where the means to check it is at hand: not the PR body's account of what changed, not the author's stated gate results, not a count in a commit message, not "the cited section says X." Where the diff, the tree, the section, or the command is available, **the reviewer looks**.

Checking is read-only. `§ The read-only clause (learned the hard way)` still binds — read the diff, read the tree, open the cited heading, run a command that only reads. A check that would write anything is not available to you; that claim goes under `ACCEPTED ON TRUST` naming `@validation-agent` as its owner.

### The report block

Every verdict carries this, in two halves:

```
VERIFIED INDEPENDENTLY:
  · …
ACCEPTED ON TRUST:
  · …
```

**The second half is the load-bearing one.** Listing what you checked is easy and flattering, and a reviewer will fill that half without effort. Naming what you took on faith is the only part of this that makes a reviewer *notice* they took something on faith — which is the entire point, because nothing else in a review surfaces an inherited premise. Write that half first if it helps.

**An empty `VERIFIED INDEPENDENTLY` block is itself a finding.** A reviewer that checked nothing has reviewed the PR body, not the work, and has returned an opinion where a verdict was asked for. Say so plainly rather than letting it pass as brevity.

**This block is the standard; `inspected: diff: true` is only the floor.** The verdict schema in `agent-handoffs` carries that flag, and a verdict reporting `false` is automatically invalid. But a flag is a self-assessment and the block is an enumeration, so a verdict may satisfy the flag and still fail here. The judge checks the block.

### What the obligation reaches

**It reaches what the verdict rests on** — any claim your own verdict depends on. It is not a re-run of the suite: that is `@validation-agent`'s job, it runs last and alone for exactly that reason, and duplicating it across the bench would double the cost of every PR to learn nothing new. Where a claim's truth would not move your verdict, take it on trust and list it.

**`ACCEPTED ON TRUST` is not a dumping ground.** A claim the verdict rests on, written in that half, **is the finding** — the reviewer has just recorded that its own verdict is unsupported. Check it, or file the gap as a finding, but do not list it and rule as though you had.

### Why this is a rule and not a habit

Both of these were found by an agent that checked a premise it had been handed, and neither was found by the pass meant to find it.

- **The board agent that could not see the board.** Its own definition told it an empty board was "a complete and correct run" and to stop; the board held **37 items**. That instruction was reachable on every run, and the run that reached it would have reported an empty backlog and recorded itself as complete. Found 4 August 2026 while sweeping citation delimiters — `c091046`.
- **The gates that could pass having read nothing.** `Architecture.md § D8` puts the Go module in `service/`. The gate block carried no working directory, so from the repository root all five commands resolve against nothing, exit zero, and print exactly what a clean module prints — and the prescribed report line was character-for-character what that vacuous run produces. Eleven agent files and the PR-body template carried the same directory-less copy, so the path ran unbroken from command to report line. Closed before any PR in this repository existed to carry it; the instrument, not a reviewer, was the thing that would have lied. Found 5 August 2026 because a layout decision recorded its own cost honestly — `d6a7e3e`, `1928a28`.

Neither is something a reviewer catches by reading attentively. Both were **instruments reporting success**, and the only thing that separated the report from the truth was an agent running the thing itself.

## Selection law

From `docs/DELIVERY.md § Selection`, which is where it is ratified; this section is how it is applied at dispatch. It replaces the withdrawn scoring law — see `docs/adr/ADR-0001-artifact-driven-agent-org.md § D2, D4`.

- **Convene from the registry, never by default.** A reviewer with no matching row does not run. *A PR exists* is not an activation condition.
- **Smallest sufficient panel.** Before dispatching, ask whether this reviewer has a reasonable chance of changing the outcome. If not, not dispatching it is the decision, not an omission.
- **Two floors are exempt from selection.** `@validation-agent` on every PR, last and alone. `@safety-sentinel` on every safety-path diff, at every tier, never softened by a budget. The sentinel is also the one reviewer any agent may convene directly on a safety concern, outside selection entirely — that door stays open.
- **Verdicts are `pass` / `revise` / `blocker`** with confidence and severity-tagged findings — schema in `agent-handoffs`. A `revise` or `blocker` naming no concrete finding is invalid and goes back; so is a `pass` that names an actionable problem without filing it.
- **N/A is not a courtesy pass.** A *convened* reviewer that finds its lane genuinely untouched returns `N/A`. Selection means this should now be rare — a common `N/A` is evidence the registry row is wrong, and is worth reporting as such.
- **Contradictory demands between reviewers** = CONFLICT. The judge resolves it by ruling one finding `invalid_finding` with a reason, or escalates it; never averages, never silently picks a side.

## The reviewer registry

One row per reviewer. `Invalidated by` is what re-runs it after a revision; anything not listed leaves a prior verdict standing.

| Agent | Domain | Activate when | Never when | Invalidated by |
|---|---|---|---|---|
| `@change-risk-assessor` | risk classification | Item intake (prediction) and PR open (authoritative, from the diff) | Docs-only typo/formatting fix — auto-low, no run | Any force-push or scope change to the PR |
| `@validation-agent` | machine evidence: build, lint, test, gates | **Every PR**, last and alone | Never skipped; never in parallel with anything | Any new commit on the head branch |
| `@safety-sentinel` | safety paths: access classification, stop selection, routing exclusions, time ceiling, feeding constants | **Any diff touching a safety path — mandatory at every tier** | The diff touches no safety path | Any revision touching a safety path, however small |
| `@docs-reviewer` | documentation accuracy and honesty | `docs/`, `Requirements/`, `README`, or code that changes documented behaviour | Pure-code refactor with no documented-behaviour surface | Changes to docs, or to behaviour a document describes |
| `@linus-quality-critic` | backend correctness, bluntly | Behavioural backend/Go change at medium+ tier, or correctness flagged | Docs-only; pure formatting | Any further behavioural change to the same code |
| `@linus-structure-critic` | code structure and layout | Go diff at high tier, or structure flagged | Docs-only; config-only | File moves, splits, or signature changes |
| `@linus-architecture-critic` | module boundaries, ports/adapters, concurrency design | Cross-boundary change | Change confined inside one package's internals | Any boundary, port, adapter, or concurrency-design change |
| `@linus-security-critic` | security surface | Auth, input validation, spatial queries, stored personal data, plan retrieval, secrets, external requests, data-touching migrations | Pure styling; docs-only | Any change to a listed surface |
| `@go-quality-critic` | idiomatic Go | Any Go diff | No Go in the diff | Any Go change |
| `@go-structure-critic` | package and file organization | Packages or files added or moved | Edits confined to the existing file set | Further adds, moves, or renames |
| `@go-architecture-critic` | Go interfaces and boundaries | Interface or boundary change | Leaf implementation-only change behind a stable interface | Further interface or boundary change |
| `@ux-reviewer` | user-facing behaviour | Frontend diff | Backend, migrations, CI, infrastructure | Further frontend change |
| `@design-reviewer` | visual and interaction design | Frontend diff | Backend, migrations, CI, infrastructure | Further frontend change |
| `@ui-engineer` | component structure, state strategy | Frontend component-structure or state-strategy change | Styling-only or copy-only change | Further component or state-strategy change |
| `@maintainability-reviewer` | future readers' cost | Medium+ tier, or a new module, or a diff over ~150 lines | Trivial low-tier diff | Any further substantive code change |
| `@modularity-reviewer` | seams between units | New packages or types, or boundary moves | Edits inside an existing unit | New packages/types, or moves across a boundary |
| `@evolvability-reviewer` | known extension seams: routing provider, elevation adapters, country widening, points/medals | The diff touches a seam | No seam in the diff | Changes at a seam |
| `@scalability-reviewer` | growth under load | Concurrency, pools, caps, fan-out, back-pressure | Single-request synchronous path with no shared resource | Changes to concurrency, limits, or fan-out |
| `@performance-reviewer` | hot paths: solve loop, spatial queries, candidate fan-out | The diff touches a hot path | Cold paths; startup-only code | Changes to a hot path |
| `@code-smell-reviewer` | local code health | Any code diff at medium+ tier | Docs-only; low tier | Any further code change |
| `@over-engineering-reviewer` | unearned complexity | New abstractions, layers, or configuration surface | Diffs that only delete or inline | New abstraction, layer, or config knob |
| `@confidence-assessor` | evidence sufficiency (meta) | Medium tier with ≥3 verdicts or any disagreement; **always at high tier** | Fewer than 3 verdicts and no disagreement; never as a code reviewer | New or changed verdicts |
| `@craft-review-summarizer` · `@linus-review-summarizer` · `@go-review-summarizer` | one voice per board | **≥3 members of that board ran this cycle** | Fewer than 3 ran — the judge reads the verdicts directly | Any member of that board re-running |

**A summarizer below three verdicts is re-narration, not synthesis** — it adds a hop and a paraphrase between the judge and evidence the judge can read in full.

**@ui-engineer also has an architect half** — commissioned by @worker-manager during implementation, which is not board convening; its registry row covers only the reviewer half. The omission is deliberate.

### Mandatory sets by risk tier

| Tier | Panel |
|---|---|
| **low** | `@validation-agent` + 1–2 domain-matched reviewers from the registry |
| **medium** | the low set + the correctness lane for the language touched + `@maintainability-reviewer` + `@confidence-assessor` if ≥3 verdicts or any disagreement |
| **high** | the medium set + the architecture lane + the security lane where flagged + `@confidence-assessor` **always**. **Every mandatory reviewer must return `pass`** |

Plus, at every tier without exception: `@safety-sentinel` on any safety-path diff.

The tiers are floors, not ceilings. Adding a reviewer the registry activates is correct at any tier; removing one the tier mandates is not.

## Incremental review validity

**A verdict is issued against a specific diff state, so it must be recorded against one.** The judge keeps a **review ledger** as a structured comment on the PR, updated every cycle:

```markdown
## Review ledger — PR #<n>

| reviewer | domain | verdict | conf | diff SHA | cycle |
|---|---|---|---|---|---|
| validation-agent | machine evidence | pass | 1.00 | a1b2c3d | 2 |
| linus-security-critic | security | pass | 0.91 | 9f8e7d6 | 1 — carried (9f8e7d6) |
| docs-reviewer | documentation | revise | 0.88 | a1b2c3d | 2 |

Convergence — cycle 2: previous 5 · resolved 4 · new 0 · remaining 1 ·
risk 0.61 → 0.31 · confidence 0.77 → 0.94 · converging: true
```

A row marked **`carried (SHA)`** states plainly that nobody looked at this cycle's diff for that lane, and on whose earlier evidence the merge will rest. That is the point: carried validity is a claim, and a claim someone who was not there can check is worth more than one they must trust.

### The intersection test

After a revision, a prior verdict **survives unless the new diff meets one of that reviewer's `Invalidated by` conditions**. Test on both axes:

- **Files** — does the new diff touch files inside the reviewer's domain?
- **Domain** — is the *nature* of the change in that reviewer's lane?

Both must hit. A typo fixed in a comment inside `auth/session.go` touches the security reviewer's files but not its domain, and does not invalidate it. A concurrency change in a file the performance reviewer never listed still invalidates performance if it sits on a hot path.

**Where the answer is genuinely unclear, re-run.** The cost of one extra reviewer is one execution; the cost of a wrongly carried verdict is a defect merged under a signature that never saw it. **On safety paths there is no unclear case** — any touch re-runs `@safety-sentinel`.

A documentation-only revision does not invalidate security, data integrity, or performance. A schema revision may invalidate data integrity, backend correctness, and performance, and almost certainly does not invalidate accessibility or UX.

## Always escalate to a human

Two categories are never settled by agent consensus, per `DELIVERY.md`:

- **Requirements marked as human-verified.** Agents can confirm a thing was built, not that it was built *well*. Whether a route recommendation is genuinely good is this product's real quality bar and is not machine-checkable.
- **Changes touching safety rules or accessibility classification.** `SPECIFICATION.md` separates safety requirements the data can enforce from those it cannot, and is explicit that a rule the system cannot verify is not a safeguard.

On both, a clean panel is a recommendation to the human, not an approval. Everything else follows the single escalation policy in `DELIVERY.md § Escalation and human judgement` — with a recommendation attached, always.

## Review identity

Review comments are posted under a **separate GitHub identity** from the repository owner's — authorship and approval must not share a signature. Authentication uses the personal access token held in the machine environment as **`GH_JUDGE_TOKEN`**.

**This is why judgments go through the `gh` CLI and never through the GitHub MCP**, even though the MCP is connected and would be more convenient. An MCP server carries one identity; routing a ruling through the same connection the implementing agent used collapses the boundary and leaves nothing in the history to show it. The full rule is in the `turfgps-board-ops` skill under *Two channels, two identities* — board and issue work through the MCP, judgments through the CLI, two tokens, two accounts, always.

That skill also sets a general **fallback rule**: when the MCP fails, agents drop to the CLI and finish the job rather than halting. **That rule does not extend to rulings.** A judgment falls back to the CLI *with the judge token*, never to the plain CLI — the plain CLI is authenticated as the repository owner, so a ruling posted through it is the author approving their own work, and it would look entirely normal in the history. A missing or failing `GH_JUDGE_TOKEN` is a stop-and-report, not something to work around.

> **The token is referenced by name only and must never be read, printed, logged, or echoed.** Pass it to `gh` through the environment. Nothing may cause its value to appear in a command, a comment, a log, or a transcript.

```bash
GH="/c/Program Files/GitHub CLI/gh.exe"
GH_TOKEN="$GH_JUDGE_TOKEN" "$GH" pr comment <N> --body-file <file>
```

Every review comment — the ledger included — is signed on its own final line:

```
/ The Review Ninja
```
