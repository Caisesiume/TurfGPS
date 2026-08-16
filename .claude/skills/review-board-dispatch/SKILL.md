---
name: review-board-dispatch
description: Mechanics for safely convening TurfGPS's reviewers — the deterministic preflight that closes lanes before any LLM runs, the read-only clause, tree-integrity verification, sequencing, review identity, the selection and negative-routing law, the reviewer registry that decides who is convened at all, and the marginal-contribution rule for overlapping reviewers. Dispatch-side only: the evidence law reviewers must satisfy lives in `agent-handoffs`. Use whenever dispatching any reviewer.
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

### Read-only is not the whole of the boundary

**Running a gate is `@validation-agent`'s.** The gates in `local-gates` are machine evidence; they run **last and alone**, on every PR, so that nothing races them and one agent owns the result the PR body reports. A reviewer that runs them anyway has not broken read-only — a green `go test` leaves nothing in the tree — but it has produced a second, unattributed measurement of the thing the ledger records under another name, and paid again for the execution selective review exists to avoid (`agent-handoffs § What the obligation reaches`).

**A reviewer needing a measurement asks for it rather than taking it:** file the gap as a finding naming `@validation-agent` as its owner, or list the claim under `ACCEPTED ON TRUST` with that owner. Both are already the evidence law's prescribed moves. What stays licensed is a reviewer's **own instrument** — `govulncheck` and `gitleaks` for the security lane, a grep, a heading list, anything read-only that *is* the review rather than a re-run of the suite. `local-gates § When these activate` draws that line and this clause does not redraw it.

**The judge's disposition, stated so the next judge does not over-correct.** A reviewer that ran a gate has **exceeded its dispatch**, and that is noted on the PR and in the ledger. It has not, by that act alone, invalidated its verdict. Where the tree verified clean and the measurement proved load-bearing, **the verdict stands and the lapse is recorded** — `pr-judge.md § Phase 5` asks whether a verdict is *evidenced*, and a verdict whose evidence held is not improved by discarding it. Discard it where the **tree moved**, which is the clause above and a different failure. Observed once, on a `@docs-reviewer`, and disposed of exactly this way.

## Sequencing

1. Convened critics run **parallel within a board**; boards may run in parallel with each other.
2. **ValidationAgent runs LAST and ALONE** — it executes builds/tests that must not race a critic's probing.
3. Review the **PR diff against `main`**, checked out in a dedicated review worktree (never the trunk tree, never a worker's worktree).

## The case file (same for every reviewer) — references, not content

PR number and the story it links · acceptance criteria and requirement codes · the head SHA under review · files modified · **safety paths touched** (see `safety-path-checklist`) · where the gate results are (see `local-gates`) · the review worktree path · the read-only clause.

**Send the references; the reviewer opens the artifacts itself.** Pasting the diff, the requirement text, or the PR body into the dispatch pays for the same bytes twice and hands the reviewer a copy that can already be stale. It also quietly invites the failure the next section exists to prevent.

## A reviewer does not accept a claim it could check

**Moved.** The evidence law — the rule, the `VERIFIED INDEPENDENTLY` / `ACCEPTED ON TRUST` block, what the obligation reaches, and both recorded incidents — now lives in **`agent-handoffs § A reviewer does not accept a claim it could check`**, which every reviewer already loads. It was here, so no reviewer could reach its own standard without loading the judge's dispatch mechanics too. This heading remains only so existing citations land somewhere true; cite the new home.

## Deterministic preflight

**Run this before anything else in a review, and before an LLM is called at all.** Git, GitHub, a glob, a SHA, and a script answer several routing questions exactly, and paying a model to answer them is paying for a worse version of a certain answer.

```bash
scripts/loop/diff-domains.sh                 # defaults to origin/main...HEAD
scripts/loop/diff-domains.sh <base> <head>   # explicit
```

It classifies every changed file by domain and returns counts, `docs_only`, `lanes_closed`, and `safety_path_candidates`. Act on its **exact negatives**:

| Deterministic fact | Consequence, with no LLM involved |
|---|---|
| no `*.go` files in the diff | `@go-quality-critic`, `@go-structure-critic`, `@go-architecture-critic` are **closed** |
| no frontend files | `@ux-reviewer`, `@design-reviewer`, `@ui-engineer` are **closed** |
| no schema files | the schema and migration lane is **closed** |
| `docs_only: true` | **not a close on its own.** The assessor's auto-low exemption is narrower than `docs_only` and its test is semantic — apply the row in `§ The reviewer registry`, and assess anything that row does not exempt |
| PR is a draft | **stop.** No panel convenes on a draft |
| head SHA unchanged since the last ledger entry | **full carry.** Nothing re-reviews; update the ledger and stop |

**The §50 guard, and it is not negotiable:** *deterministic checks close lanes only where the file-domain mapping is exact; anything semantic — safety paths above all — stays with `@change-risk-assessor` and `@pr-judge`.*

That guard is why the script only ever **closes** a lane and never opens one, and why `safety_path_candidates` prints `hint_only: sentinel activation is semantic` next to itself. A file list can prove a Go critic has nothing to read. It cannot prove a safety rule was not changed, because a safety rule can be changed by a constant in a file no hard-coded list has heard of — and the cost of those two errors is not remotely symmetric.

**The docs row is the one place a deterministic fact stops short of its consequence.** `docs_only` is exact and is a statement about file extensions; whether a docs diff is a typo, a formatting change, or a link fix is a judgement about what it *does*, and that exemption belongs to the assessor — stated in its registry row below and in its own contract, which is why this table applies that row instead of carrying a third copy of it. **PR #67 is the evidence:** `docs_only: true` across three Markdown files, and the authoritative PR-open assessment came back `medium / 0.54` naming `security` as a required lane. The observed behaviour already followed the assessor's row; it was this table that was wrong, and §50 had reserved the semantic half for the assessor and the judge before the case arose. Narrowed 16 August 2026 — `docs/adr/ADR-0002-token-efficiency.md § O17`.

## Selection law

From `docs/DELIVERY.md § Selection`, which is where it is ratified; this section is how it is applied at dispatch. It replaces the withdrawn scoring law — see `docs/adr/ADR-0001-artifact-driven-agent-org.md § D2, D4`.

- **Convene from the registry, never by default.** A reviewer with no matching row does not run. *A PR exists* is not an activation condition.
- **Smallest sufficient panel.** Before dispatching, ask whether this reviewer has a reasonable chance of changing the outcome. If not, not dispatching it is the decision, not an omission.
- **Two floors are exempt from selection.** `@validation-agent` on every PR, last and alone. `@safety-sentinel` on every safety-path diff, at every tier, never softened by a budget. The sentinel is also the one reviewer any agent may convene directly on a safety concern, outside selection entirely — that door stays open.
- **Verdicts are `pass` / `revise` / `blocker`** with confidence and severity-tagged findings — schema in `agent-handoffs`. A `revise` or `blocker` naming no concrete finding is invalid and goes back; so is a `pass` that names an actionable problem without filing it. **`@validation-agent` is outside this vocabulary**: it returns a machine result, `validation: {status: pass | fail}`, because evidence and judgement must be distinguishable in the ledger without a special case.
- **N/A is not a courtesy pass.** A *convened* reviewer that finds its lane genuinely untouched returns `N/A`. Selection means this should now be rare — a common `N/A` is evidence the registry row is wrong, and is worth reporting as such.
- **Contradictory demands between reviewers** = CONFLICT. The judge resolves it by ruling one finding `invalid_finding` with a reason, or escalates it; never averages, never silently picks a side.

### Negative routing

`@change-risk-assessor` returns `review_required`, `review_optional`, and `review_not_required`. The third is **a hard negative, not a shrug.**

**A lane listed under `review_not_required` is not convened.** The judge may overrule it, but only by recording why, on the PR, in this shape:

```yaml
reviewer_override:
  reviewer: performance-reviewer
  risk_assessment: not_required
  reason: revision introduced an O(n^2) candidate loop not present in the original assessment
```

Without a recorded override: **do not run it.** An assessment that can be quietly ignored is an assessment nobody has to write honestly, and the whole selective model rests on that output meaning something.

**`review_optional` is not "run if there is budget."** It runs on a concrete signal, and the signal is named when it fires: unusual diff structure · the implementer flagged uncertainty · a conflict between reviewers · an acceptance criterion that depends on that domain · prior defect history in the component · a confidence gap. No signal, no run.

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
| `@linus-security-critic` | security surface | Auth, input validation, spatial queries, stored personal data, plan retrieval, secrets, external requests, data-touching migrations — **including a document that decides one of them** | Pure styling; docs-only **that decides no exposure boundary, no trust boundary, and nothing an infrastructure item will build against** | Any change to a listed surface |
| `@go-quality-critic` | idiomatic Go | Go diff with behavioural or interface change: new/changed exported identifiers, error-handling or context-propagation changes, concurrency primitives, or non-trivial implementation logic (roughly 40+ changed Go lines); or the risk assessment requests the correctness lane | Rename-, move-, comment-, or formatting-only Go diffs; docs-only | Further change to the functions or packages it reviewed, or a new exported API, error-path, or concurrency change elsewhere |
| `@go-structure-critic` | package and file organization | Packages or files added or moved | Edits confined to the existing file set | Further adds, moves, or renames |
| `@go-architecture-critic` | Go interfaces and boundaries | Interface or boundary change | Leaf implementation-only change behind a stable interface | Further interface or boundary change |
| `@ux-reviewer` | user-facing behaviour | Frontend diff changing user-visible behaviour: flows, states, feedback, copy, information hierarchy | Logic or state refactors with unchanged rendering; backend, migrations, CI | Further user-visible behaviour change |
| `@design-reviewer` | visual and interaction design | Frontend diff changing layout, composition, design tokens, theme, or visual states | Logic-only; copy-only (that is `@ux-reviewer`); backend, migrations, CI | Further visual change |
| `@ui-engineer` | component structure, state strategy | Frontend component-structure or state-strategy change | Styling-only or copy-only change | Further component or state-strategy change |
| `@maintainability-reviewer` | future readers' cost | A new module, roughly 150+ changed lines, or the risk assessment requests the lane | Trivial diffs; minimal-patch revisions introducing no new concept | A revision restructuring what it reviewed, introducing a new module or concept, or addressing one of its own findings |
| `@modularity-reviewer` | seams between units | New packages or types, or boundary moves | Edits inside an existing unit | New packages/types, or moves across a boundary |
| `@evolvability-reviewer` | known extension seams: routing provider, elevation adapters, country widening, points/medals | The diff touches a seam | No seam in the diff | Changes at a seam |
| `@scalability-reviewer` | growth under load | Concurrency, pools, caps, fan-out, back-pressure | Single-request synchronous path with no shared resource | Changes to concurrency, limits, or fan-out |
| `@performance-reviewer` | hot paths: solve loop, spatial queries, candidate fan-out | The diff touches a hot path | Cold paths; startup-only code | Changes to a hot path |
| `@code-smell-reviewer` | local code health | Medium+ tier diff adding roughly 100+ changed implementation lines, or a duplication/nesting/dead-code signal flagged by the risk assessment or another reviewer | Docs-only; low tier; minimal-patch revision diffs adding no new logic | A revision adding new implementation logic — not comment, config, or doc edits |
| `@over-engineering-reviewer` | unearned complexity | New abstractions, layers, or configuration surface | Diffs that only delete or inline | New abstraction, layer, or config knob |
| `@confidence-assessor` | evidence sufficiency (meta) | Medium tier with ≥3 verdicts or any disagreement; **always at high tier** | Fewer than 3 verdicts and no disagreement; never as a code reviewer | New or changed verdicts |
| `@craft-review-summarizer` · `@linus-review-summarizer` · `@go-review-summarizer` | one voice per board | **5+ members of that board ran this cycle**, OR the judge records multiple substantive cross-reviewer conflicts requiring synthesis, OR the combined verdict payload is genuinely too large to weigh directly (recorded as such) | 1–4 compact verdicts — the judge reads them directly | Any member of that board re-running |

**Structured data is not summarized by another LLM merely to make it structured again.** A verdict is already 150–300 tokens of schema; a summarizer that consumes three of them and emits a fourth has added a hop, a paraphrase, and a full execution between the judge and evidence the judge can read whole.

**State the effect plainly, because it is the point rather than a side effect:** the Go board has three members and the Linus board has four, so **neither can ever meet the count condition**. Their summarizers are now conflict-triggered only. That is not an oversight to be corrected by lowering the number — it is what happens when a threshold is set by what synthesis actually costs rather than by what the historical architecture happened to do.

**@ui-engineer also has an architect half** — commissioned by @worker-manager during implementation, which is not board convening; its registry row covers only the reviewer half. The omission is deliberate.

**`docs-only` is a statement about file extensions, not about consequence — and the security row now says so.** A document can decide an exposure boundary, a trust boundary, or the thing an infrastructure item will be built against, and a diff doing any of those has moved the security surface without changing a line of code. **PR #67 is the evidence, and is why the row was widened rather than argued:** a judge crossed the old flat `docs-only` exclusion deliberately, recorded why on the PR, and `@linus-security-critic` returned **two high findings on paths to stored personal data**. The exclusion still holds for what it was written for — a typo, a heading, a rewording that settles nothing — and the test is what the document *decides*, not what it is named. Amended 16 August 2026 on that evidence; `docs/adr/ADR-0002-token-efficiency.md § O17` records it.

### Mandatory sets by risk tier

| Tier | Panel |
|---|---|
| **low** | `@validation-agent` + 1–2 domain-matched reviewers from the registry |
| **medium** | the low set + the correctness lane for the language touched + `@confidence-assessor` if ≥3 verdicts or any disagreement |
| **high** | the medium set + the architecture lane + the security lane where flagged + `@confidence-assessor` **always**. **Every mandatory reviewer must return `pass`** |

Plus, at every tier without exception: `@safety-sentinel` on any safety-path diff.

**Medium's floor is validation, the correctness lane, and conditional confidence — `@maintainability-reviewer` is no longer mandatory there.** It is not weakened, only moved: its registry row above fires on a new module, on roughly 150+ changed lines, or when the risk assessment asks for the lane, and a medium-tier diff meeting any of those still gets it. What is removed is its activation by *tier alone*, which convened it on medium diffs that introduced no new concept for it to weigh.

The tiers are floors, not ceilings. Adding a reviewer the registry activates is correct at any tier; removing one the tier mandates is not.

## The marginal contribution rule

Several reviewers in the registry share a border. Two of them convened on the same diff will usually both find something, and will usually both find *the same* thing in two vocabularies — which reads like corroboration and costs like coverage.

**Before convening a reviewer whose domain substantially overlaps one already convened, record the question the second one uniquely answers:**

```yaml
candidate_reviewer: linus-quality-critic
overlaps_with: [go-quality-critic]
marginal_question:
  - "Could the implementation be logically incorrect despite being idiomatic Go?"
expected_unique_value: true
```

**If `marginal_question` cannot be stated clearly: do not invoke.** The inability to say what the second reviewer adds *is the answer* — it is not a prompt to think harder about the justification.

| Family | The second one is convened only if… |
|---|---|
| `@go-quality-critic` ↔ `@linus-quality-critic` | Linus adds: could this be **logically incorrect or fragile at runtime** despite being idiomatic? |
| `@go-architecture-critic` ↔ `@linus-architecture-critic` | Linus adds: is this **operationally sound system-wide** — resilience, observability — beyond Go-idiomatic boundaries? |
| `@maintainability-reviewer` ↔ `@code-smell-reviewer` | Is the **cost of the next change** genuinely distinct from the smell census? |
| `@modularity-reviewer` ↔ `@go-`/`@linus-structure-critic` | Are **coupling and dependency direction** at issue, beyond file and package shape? |
| `@evolvability-reviewer` ↔ the architecture lanes | Is a **named extension seam** concretely implicated — routing provider, elevation adapters, country widening, points/medals? |
| `@performance-reviewer` ↔ `@scalability-reviewer` | Are **both now-cost and growth-behaviour** concretely implicated? |
| `@ux-reviewer` ↔ `@design-reviewer` ↔ `@ui-engineer` | Route by what changed — behaviour → ux, visuals → design, component/state architecture → ui-engineer. A second one joins only with its own stated question. |

**Default within a family: convene the one whose trigger matched most specifically.** A broad match and a narrow match on the same diff is one reviewer's evidence, not two.

These agents stay registered and stay distinct — **the rule is a convening condition, not a merge.** The moment a diff genuinely raises both questions, both run, and the recorded `marginal_question` is what makes that defensible rather than habitual.

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
