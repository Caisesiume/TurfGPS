# ADR-0001 — Artifact-driven agent organization

**Status:** accepted — 2026-08-10 · **architecture stable — 2026-08-13**
**Stability note (2026-08-13, Owner Directive 4 — `docs/adr/agent-org-directive-4.md`):** the organization ratified here is **stable as of this date**. Directive 4 added no agent, collapsed no seat, and reopened no decision below; it hardened the *wording* of the orchestration contracts so that no agent's Role, Authority, Focus, or Contract claims a decision another agent owns. Subsequent organizational change is **evidence-driven** — the maintenance rule is in `docs/DELIVERY.md § The architecture is stable`, and it cites this record. Per directive 4 §26 there is no ADR-0004: nothing genuinely new was decided.
**Source:** `docs/adr/agent-org-directive.md`, the Owner's directive, kept verbatim. Where that file and this one differ on a repository-specific adaptation, **this record is the ratified form** and the directive is the unaltered order it adapts. Section references below (`§N`) point at the directive.

## Context

The fleet grew as a *standing organization*: agents that talked to each other continuously, a review bench that convened whole, and handoffs that carried conversation rather than conclusions. It produced good work and it produced defects nothing else caught — the record of that is kept, not disowned.

What it cost:

- **The bench convened whole on every PR.** Twenty reviewers examined a diff so that most of them could return `N/A`. Selection existed only *after* the invocation, which is the one place it cannot save anything.
- **The whole bench re-convened after every revision.** A one-line fix to one file re-ran every reviewer whose domain the fix could not possibly have touched. The re-run's verdicts were, by construction, the previous verdicts.
- **A single scale carried every kind of finding.** `0–10` with a `10.00` unanimity gate meant a naming quibble and a data-loss bug both arrived as "below 10", and the gate could only be opened by resolving both to the same standard.
- **Transcripts flowed upward.** Parent agents accumulated the reasoning of their descendants and then re-performed the analysis they had just been handed.
- **The Owner was a blocking dependency for ordinary ambiguity.** Requirement batches waited on sign-off for questions the precedence of the source documents already answered.

The directive's diagnosis is that agents must stop behaving like a continuously communicating organization and start behaving like **independently executing specialists connected through small structured handoffs and persistent artifacts**. The decisions below are how that lands on *this* fleet, which already has twenty-odd working specialists and a body of law earned from real incidents.

The governing constraint on the whole adaptation: **the primary optimization is fewer unnecessary executions and smaller handoffs, not fewer agents.** Nothing here deletes a specialist.

## Decision

### D1 — Names stay; seats are filled by existing agents

Agent IDs and filenames are stable. The directive's new seats map onto the existing fleet:

| Directive seat (§) | This repository |
|---|---|
| Peer Judge (§10) | `pr-judge` |
| Implementation Lead (§7) | `worker-manager` |
| Requirements Librarian, *coordinator sense* (§4) | `requirements-engineer` |
| Change Risk Assessor (§6) | `change-risk-assessor` — new |
| Confidence Assessor (§14) | `confidence-assessor` — new |

**Two agents answer to the name "librarian" and they are not the same job.** `requirements-engineer` holds the directive's coordinating seat — it selects specialists, owns the corpus, and resolves ambiguity. The existing `requirements-librarian` keeps its name and its narrow lane: document management only — structure, stable IDs, category filing, the index, the traceability matrix. It authors nothing and coordinates nothing. Renaming either one would break every cross-reference in the fleet to save a word of ambiguity, which this paragraph removes more cheaply.

No deletions, no renames. Two additions.

### D2 — The verdict law is replaced

The `0–10` scale and the `10.00` unanimity gate are **gone**. Reviewers return `verdict ∈ {pass, revise, blocker}`, a `confidence`, and severity-tagged findings (§13).

Merge requires: **zero unresolved `required_change` among the convened reviewers**, and the risk×confidence model of §15 satisfied. Every actionable finding resolves to `required_change`, `accepted_risk`, or `invalid_finding`, each with a recorded owner (§16).

Two clauses of the old law survive translated, because they were never really about numbers:

- **Enumerate or certify.** A `revise` or `blocker` carrying no concrete finding is **invalid** — it is an impression, not a verdict. A `pass` that names an actionable problem it did not file is equally invalid: the problem is either filed as a finding with an owner, or it is not actionable and says so.
- **N/A is not a courtesy pass.** Selection now happens at convening time, so `N/A` is no longer how a reviewer excuses itself from a diff it should never have been sent. It survives for the narrow real case: a *convened* reviewer that finds its lane genuinely untouched. A reviewer that passes because it found nothing to examine has recorded an approval it never performed, and that still matters later when the question is who approved something.

**What is given up:** the old gate's dilution-resistance — one sub-10 blocking nine 10s, immune to uninvolved agents handing out easy passes. That property is retained by a different mechanism: **per-finding ownership**. A finding cannot be diluted by a majority because it is not counted, it is *resolved* — by fixing it, by accepting it as a recorded risk with a named owner, or by ruling it invalid with a reason. Nobody averages it away, because nothing is averaged.

### D3 — Iteration budgets replace the round cap

**3 autonomous revision cycles normally, 5 for high-risk** (§20). Exceeding the budget is not a matter of asking for more rope: the judge must first make a **root-cause determination** from §20's causes and route it per §28 — fix the cause or escalate.

The historical **8-round cap survives as the absolute emergency ceiling** that ends in a human, and its recorded history stays in `docs/DELIVERY.md`: during the review of the product concept, a first pass produced 13 findings, and the round of fixes addressing them introduced three of the four blockers found by the second pass. That episode is why a ceiling exists at all. Under the new budgets, reaching it should be rare enough to be a report in itself.

### D4 — Review is selective, and revalidation is incremental

Reviewers are convened from the **registry** in `review-board-dispatch`, driven by `change-risk-assessor` output. **Never the whole bench by default** (§10).

Incremental validity per §17–18 is made checkable rather than trusted: the judge maintains a **review ledger** as a structured PR comment — one row per reviewer with verdict, confidence, the **diff SHA reviewed**, and domain — updated every cycle. After a revision, only reviewers whose domain intersects the new diff re-run; the rest carry forward marked `carried (SHA)`. Convergence is tracked in the same ledger per §19.

Board summarizers (`craft-`, `linus-`, `go-review-summarizer`) convene **only when ≥3 members of their board ran this cycle**. Below that the judge reads the verdicts directly; a summarizer aggregating two verdicts is a re-narration, not a synthesis.

PRs carry a `risk:low` / `risk:medium` / `risk:high` label so the tier is visible on the board rather than only inside a comment.

### D5 — Preserved law

The following survive **verbatim in substance**. Weakening any of them is a defect, not a simplification — each was bought with an incident:

- The **red-demonstration rule** (`DELIVERY.md § Proof that a test can fail`), whole, including the wrong-reason and nothing-to-revert clauses.
- The **`VERIFIED INDEPENDENTLY` / `ACCEPTED ON TRUST`** block and *a reviewer does not accept a claim it could check*. This **is** §12's evidence law in a stronger form, and §12's `inspected: diff: true` maps onto it: the flag is the floor, the block is the standard. Both recorded incidents stay.
- The **read-only dispatch clause** and tree fingerprinting.
- **ValidationAgent runs last and alone, on every PR.** It is machine evidence, and it is exempt from selection: the point of machine evidence is that it does not depend on someone deciding it was relevant.
- The **review worktree** discipline.
- **Traceability as part of the case** — PR ↔ story ↔ requirement codes.
- **Machine evidence precedes opinion** — a red gate ends the hearing before it starts.
- The **two always-human categories**: requirements whose verification method is human judgement, and any change touching safety rules or accessibility classification.
- **`@safety-sentinel` is mandatory on any safety-path diff at every risk tier**, and is never softened by a budget. A safety path does not become less dangerous because the diff is small, and "low risk" is a statement about the change, not about where it landed.
- **Judge identity**: `GH_JUDGE_TOKEN`, the two-channel rule, the `/ The Review Ninja` signature, and the token never echoed.

### D6 — Decision authority

- §5's precedence ladder binds the requirements agents; §22's preference ladder binds everyone else.
- **§21 is the only human-escalation policy in the repository.** Its packet schema is used as written, and every escalation carries a recommendation. "What should I do?" is not an escalation.
- **Requirement batches no longer block on Owner sign-off.** The requirements engineer resolves ordinary ambiguity itself and records each resolution in `docs/Requirements/DECISIONS.md` — ID, date, question, interpretation chosen, precedence rung relied on, affected records. `@requirements-librarian` owns that file's structure, as it owns the rest of the corpus's shape. The Owner receives a **non-blocking decisions digest** through `@engineering-lead`. Only §21-qualifying questions block.

This is the largest behavioural change in the record and the one most likely to be misread. It does not make the Owner optional — it stops spending the Owner on questions the four documents already answer in precedence order, so that the questions that do reach them are worth their attention.

### D7 — The board keeps its six columns

Status stays `Backlog` → `Ready` → `In progress` → `In review` → `Ordered Revision` → `Done`. §26's richer lifecycle is a **mapping**, documented in `turfgps-board-ops`:

| §26 state | Column |
|---|---|
| Backlog | `Backlog` |
| Requirements Ready · Ready for Implementation | `Ready` |
| In Progress | `In progress` |
| Implementation Complete · Review | `In review` |
| Revision Required | `Ordered Revision` |
| Review Passed · Ready to Merge · Merged | `Done` at merge |

No schema churn. The board's Status options were regenerated once already and every option ID changed with them; a second regeneration to gain finer names that no agent branches on would be cost without benefit.

### D8 — Handoffs are envelopes

§23–25 — the shared envelope, the input-references/execution/structured-verdict principle, and the ~300-token typical limit — live in a new skill, **`agent-handoffs`**, loaded by every dispatching or reporting agent. References, not content; no transcripts, no chain-of-thought, no chronology of the work.

### D9 — Findings are routed by root cause

Every finding is classified `implementation | requirement | architecture | design | test | infrastructure` (§28). A requirement defect goes to `@requirements-engineer`; an architectural contradiction goes to the ADR process. **Repeatedly patching code to work around an upstream defect is forbidden** — it is the mechanism by which a broken requirement becomes permanent and expensive.

### D10 — Lessons become artifacts

§29. A recurring lesson becomes an ADR, a registry rule, or a skill update. It never becomes resident context, because context ends and the next agent starts without it.

## Consequences

**What this obliges:**

- Every reviewer needs an explicit activation condition, and the registry in `review-board-dispatch` is now load-bearing law rather than documentation. A reviewer with no row does not get convened.
- The judge does more work per PR — risk assessment, selection, validity checks, finding resolution, the ledger — in exchange for far fewer reviewer executions. The judge is the right place for that cost: it is one agent, and it was already the only one holding the whole case.
- Leaf agent definitions across the fleet still carry the `0–10` and `10/10` language. Until each is patched, those files contradict this record. **This record governs**; the contradiction is a defect in the leaf file.
- `docs/Requirements/DECISIONS.md` is a new corpus artifact and must exist before the requirements engineer next resolves an ambiguity, or the authority granted in D6 has nowhere to be recorded.

**What this gives up:**

- **Belt-and-braces coverage.** Under whole-bench review, a reviewer occasionally found something outside the lane it was convened for. Selective convening will lose some of those. The wager — the Owner's, recorded here as such — is that the loss is smaller than the cost, and that per-finding ownership plus the mandatory floors (validation always, safety sentinel always on safety paths, the high-tier mandatory set) catch what matters.
- **A single legible number.** `10.00` was easy to report and easy to check. `zero unresolved required_change` is the honest version of the same statement and is harder to read at a glance; the ledger exists to make it legible again.

**Reversibility:** high. The registry is a table, the budgets are numbers, and the verdict vocabulary is three words. If selective review demonstrably lets defects through, the evidence will be in the ledger — which is the reason the ledger records the SHA each verdict was issued against.

## Amendment — 2026-08-16 (first live loop cycle)

*Source: the Owner's runtime-findings directive, deliberately not filed as a separate document — its rules are recorded in the existing ADRs. Five contract defects **observed** during the loop's first live cycle, which is exactly what `docs/DELIVERY.md § The architecture is stable` requires before an organizational change: these are operational evidence, not hypotheses. No agent added, no seat collapsed, no decision above reopened, and **still no ADR-0004** — each of these is a contract the agents already implied and none of them stated.*

### D11 — An agent must not end a pass while a continuation it owns is outstanding

*Observed twice. `@requirements-engineer` dispatched children in the background and its own process ended before they returned, so the mandatory planner continuation never fired; the second time a finished `FR-019` field block was left with nobody holding it, and survived only because it was parked by hand in a comment on issue `#18`.*

Dispatching creates an obligation that outlives the dispatch. An agent that dispatches therefore ends its pass in exactly one of two ways: it **awaits its children**, or it **persists their output to a durable artifact and names in its envelope what remains owed and to whom**. Holding the work only in the pass's own context is neither.

Stated once, in `agent-handoffs § An outstanding continuation is not left behind` — it binds every dispatching agent, and none of them should have to open another agent's file to find it. `requirements-engineer § Mode A` cites it at the continuation that has actually been dropped.

**This is a durability rule, not a scheduling one.** The failure is invisible by construction — nothing anywhere records that a step was owed — which is why the obligation is discharged into an artifact rather than trusted to a process staying alive long enough.

### D12 — One writer per branch

*Observed once. Two agents worked `#37`'s branch concurrently, and one committed the other's uncommitted edits as its own.*

**Before dispatching a writer, verify that no other agent holds that branch or its worktree.** A branch with two writers has no author: the commit is attributed to the wrong lane, the diff the judge reads is not the diff either agent wrote, and nothing in the history distinguishes that from ordinary work — a traceability defect that passes every traceability check we have.

`worker-manager § One writer per branch` is the home and carries the check. `@project-coordinator` cites it, because assignment is the moment the collision is either created or prevented.

### D13 — A judge that cannot convene rules nothing, and says where it stopped

*Observed once and **handled correctly by the agent that hit it**: one `@pr-judge` process had no dispatch capability and stopped at Phase 4 with a durable resume point rather than fabricating a panel. Another instance convened normally, so the capability is not something a judge may assume about itself.*

A judge without the means to convene has selected a panel and heard none of it, and ruling anyway would enter a merge decision under a signature that reviewed the diff itself. It first asks `@engineering-lead` to courier the dispatch, and where that too is unavailable it **stops, rules nothing, and emits a resume packet** carrying what is expensive to re-derive — the head SHA, the selected panel with its reasons, and what must not be convened — so Phases 0–3, the expensive half, are not paid for twice. The preflight's closed lanes are deliberately not in it: a script re-runs, and a cache of its output can go stale in a way the script cannot. In `pr-judge § When the panel cannot be convened`.

**Recorded because it was improvised well, not because it went wrong.** A behaviour that exists only in the instance that invented it is a behaviour the next process does not have.

### D14 — Creating a work item includes putting it on the board

*Observed once. `@engineering-lead` filed issue `#108` through the API and set no board Status; the issue existed, and every board query returned a board without it.*

An item with no `Status` is not at the head of the chain — it is outside the chain: invisible to every filtered view, and unsequenceable by the agent who would otherwise promote it. The rule and its `@scrum-master` fallback are in `turfgps-board-ops § Status`, whose text was already correct but read as binding only `@requirements-story-organizer`; **it is now written for every filing agent, through every channel**, which is the whole of what this amendment changed.

### D15 — Commits reference a `Task`, and its three exemptions are exhaustive

*Ruled rather than observed. `turfgps-board-ops § Labels` enumerated three `Task` exemptions and separately said a `Task` "never enters the chain it describes", which left the commit link genuinely ambiguous to a reader applying both sentences.*

**They must.** A `Task`'s commits reference its own `#N`. That link buys **attribution** — which item this commit was done for — and not requirements-tracing, which is the chain a `Task` legitimately stays out of. The three exemptions (`Resolves:`, Milestone, the coverage audit) are the complete list, and no fourth is inferred from the chain sentence. One sentence, in `turfgps-board-ops § Labels`.
