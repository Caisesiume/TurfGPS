# Delivery and review model

How work on TurfGPS is tracked, reviewed, and judged shippable.

Documentation work runs through this model today; the code half of it activates with the Go and frontend stacks. The agent organization that applies it is ratified in `docs/adr/ADR-0001-artifact-driven-agent-org.md`, and the cost of executing it is governed by `docs/adr/ADR-0002-token-efficiency.md`.

## The general law of invocation

**No agent runs because its role is relevant in the abstract. It runs because there is a concrete unresolved decision in its domain.**

Everything else in this document about selection, carrying, budgets, and stopping is that sentence applied to a particular moment. It is worth stating on its own because the failure it prevents never looks like a failure from inside: convening a reviewer whose lane the diff plausibly touches feels like diligence, and produces a `pass` that changed nothing, cost a full execution, and adds a signature to the record from an agent that had nothing to weigh.

The test is not *could this agent have an opinion here* — it is **what decision is currently unresolved that this agent's judgement would settle.** If that question has no answer, the invocation has no purpose, and not making it is the decision rather than the omission.

## Work tracking

Epics and User Stories are extracted from the requirements **before implementation begins**, and stored as GitHub issues on a GitHub Project board.

**Milestones serve as Epics**, tying related issues together.

Agents work the board using the `gh` CLI and the GitHub MCP connector.

### Requirements come first

Issues are derived from the requirements in `Requirements/`, and cite the requirement IDs they satisfy. This is the dependency that governs sequencing: a board built before the requirements exist would contain guesses, and review agents would have nothing objective to check work against.

Consequently `SPECIFICATION.md` and `Requirements/` block the board. `DESIGN.md` and `DEPLOYMENT.md` do not, and can follow.

### Who owns what

**One home for the split**, ratified in `docs/adr/ADR-0003-backlog-dependency-planner.md`. Contracts elsewhere cite this table rather than restating it.

| The question | The owner |
|---|---|
| What is true about the requirements | `@requirements-engineer`, authored by `@requirements-fr` / `@requirements-nfr`, corpus curated by `@requirements-librarian` |
| What the work items **are** — Epics and stories, the *nodes* | `@requirements-story-organizer` |
| What must precede what — the persistent dependency *edges* | `@backlog-dependency-planner` |
| Which Backlog items are **eligible to enter Ready** now, and board truth | `@scrum-master` |
| Which **Ready item runs next**, and merge order | `@project-coordinator` |
| Which specialists **implement** it | `@worker-manager` |
| Which reviewers are convened, and whether it is **shippable** | `@pr-judge` |
| Which owner must act next, and the human channel | `@engineering-lead` |

The three middle rows are the ones that merge if nobody watches, and they are why the table exists: decomposition creates the nodes, the planner owns the edges, readiness is a third question asked of both. An agent answering two of them has answered one of them without review — which is how a backlog acquires ordering nobody verified and cannot later distinguish from ordering that was.

**Rows 4 and 5 are the second pair that merges**, and the split is exact: *eligible for Ready?* is `@scrum-master`'s; *which Ready item runs next?* is `@project-coordinator`'s. Promotion order within the WIP limit belongs to the first; runtime execution order belongs to the second. `@engineering-lead` orchestrates all of them and reconstructs none of them.

### The architecture is stable

**Ratified by ADR-0001 and hardened by `agent-org-directive-4.md § 33`: this organization is now stable, and further organizational change is driven by observed operational evidence rather than by hypothetical optimization.** The evidence that justifies a change is the kind you can point at from a real run — repeated unnecessary invocations, routing failures, recurrent missing dependencies, review blind spots, excessive escalation, context pressure, measurable duplicate-reviewer cost, or a defect attributable to missing ownership. No metrics platform is owed for this rule; the judge's existing review accounting is the first place to look. Absent such evidence, the correct action on the organization is to run it.

## Proof that a test can fail

**Every test written for a `test`-verified acceptance criterion is demonstrated to fail without the change under test, and the pull request states the demonstration.** A criterion whose test has never been red is asserted, not verified.

The corpus has already caught this defect one level up. `NFR-001` carried an acceptance criterion that would have reported green on a version bump while measuring nothing, and it was rewritten rather than noted — `docs/Requirements/README.md § ID allocation ledger` records the reject and its reasoning. A test that passes whether or not the implementation is present is that same criterion in executable form, and it is the worse of the two: the requirement was caught by a reviewer reading it, while a vacuous test reports success in the one place nobody re-reads.

**The rule is the evidence, not the sequence.** Test-first was considered and refused, and the reasoning is recorded because it is the obvious answer and will be proposed again: nothing in a diff shows the order its files were written in, so *the test came first* is a claim a worker can make and no reviewer can check. *Revert the change and the test goes red* is checkable, and checkable by someone who was not there.

`@test-engineer` already holds the standard — a test earns its place only if it would fail on the bug it names — so this gives an existing belief teeth rather than importing a new one.

### Red for the wrong reason

**The test must fail as an assertion.** A test that is red because the package no longer compiles, because the function it calls was deleted, or because the code panicked before reaching the assertion demonstrates nothing. Without this clause the rule is satisfiable by breaking the build, which is easier than writing a real test.

**So the demonstration neutralises the change rather than deleting it**: keep the signature, return the zero value, or invert the single behaviour under test. Deleting a newly added function takes its test out of the build along with it, so a new implementation's demonstration fails by construction unless this is stated. `panic("not implemented")` is the reflexive stub and is excluded for the same reason — the test never reaches its assertion.

**The evidence is therefore the assertion's own failure message**, the observed value against the expected one. That is what makes the clause self-enforcing rather than an appeal to good faith: a package that does not build prints `[build failed]` and no `--- FAIL: TestName` line at all, and a panic prints a stack trace rather than a comparison. Only a test that compiled, ran, and evaluated the thing it asserts can produce that message.

### Where there is nothing to revert

**The rule reaches `test` and no other verification method.** The five methods are defined under `requirements-authoring § Verification methods`, and `test` is the only one whose evidence is an automated test — an assertion being the only thing there is to show red. The distinction is not simply run-versus-no-run: `demonstration` is a run, an operator working the system and observing it, and it still has no red state to demonstrate because nothing in it asserts. `requirements-authoring § Acceptance-criteria form` already refuses to wrap an `inspection` criterion in given/when/then on the neighbouring ground that nothing executes at all. **The scope is by method rather than by census**, so it does not rot when a record is filed under a method nothing currently uses. The exclusion is not theoretical either: on 6 August 2026 five of the corpus's records verified by `inspection` or `human-judgement`, with the live count derivable from `docs/Requirements/INDEX.md`.

**Where the change does not exist yet, the demonstration is owed and not waived.** Tests may legitimately land before or apart from the code they cover. The pull request then states per criterion that the demonstration is owed and names the story that will discharge it, and the implementing pull request carries it — reverting its own change being precisely the demonstration. What may not happen is a criterion passing out of both pull requests with the demonstration in neither, which is what waiving it would amount to and is the whole of what this clause prevents.

**Where the implementation lands in the same pull request as its tests, the stub stage already is the demonstration.** Write the signature returning the zero value, run the tests, record the red; then implement and run them green. Nothing has to be neutralised afterwards, because the state a later author would have had to construct existed on the way through — which is why complying with this rule costs almost nothing in the ordinary case, and why an author who reports no demonstration is usually reporting that they never watched their own test fail.

### What a reviewer does with it

Confirming a demonstration instead of reading the claim of one is an instance of `review-verdicts § A reviewer does not accept a claim it could check`, and is governed there rather than here. One consequence is worth naming because it is not obvious: **re-running a demonstration is a write**, since it neutralises an implementation in a tree, so it is not available to a critic under that section's read-only bound and belongs in the `ACCEPTED ON TRUST` half. What is available read-only is the shape of the recorded evidence — that an entry exists for every `test`-verified criterion the item claims, and that each carries an assertion's message rather than a build failure.

The form the pull request reports this in sits with the gate report law, under `local-gates § The law`.

## Review

Every item reaching test or verification is reviewed by specialist agents before it is judged shippable or sent back for revision. **Which** specialists is a decision, not a default.

Most agents own a **single software quality attribute** — performance, modularity, scalability, security, maintainability, evolvability, over-engineering, user experience, documentation, test coverage, and so on. Reviewing one dimension well beats reviewing everything shallowly, and it makes each verdict attributable.

Alongside them sit the **"Linus" critics**, modelled on Linus Torvalds' direct review style: blunt, and willing to raise a small defect loudly rather than let it pass for being small.

### What this supersedes

**This document supersedes the 10.00 unanimity gate and the whole-bench re-convening rule.** Both are withdrawn, and an agent file still describing them is out of date rather than authoritative.

They are withdrawn because of what they cost against what they bought. Convening every reviewer on every diff mostly produced `N/A` from agents whose attribute the change never touched, and re-convening all of them after a one-line fix reproduced the previous verdicts at full price. That is a large, repeated token cost for no marginal quality.

The old gate did have one property worth keeping: **it could not be diluted.** A single sub-10 blocked nine 10s, so uninvolved agents handing out easy passes could not average a real objection away. That property is retained by a different mechanism — **findings are owned and resolved, never counted**. Nothing is averaged now, so there is nothing to dilute. A finding leaves the process only by being fixed, by being accepted as a recorded risk with a named owner, or by being ruled invalid with a reason.

### Risk classification

**Every change is classified before review, and the classification decides the panel.** `@change-risk-assessor` produces it — predicted at item intake, and authoritatively from the actual diff when the PR opens. The tier is recorded on the PR as `risk:low`, `risk:medium`, or `risk:high`.

**High is mandatory, not a judgement call, when the diff touches any of:** a safety path (access classification, stop-position selection, routing exclusions, the time ceiling, or the constants feeding them) · database schema or DDL · migrations · the authentication or plan-retrieval surface · a breaking change to a public API · an architecture with no precedent in this repository.

Everything else is medium or low on the assessor's factors — surface area, data-integrity impact, external integrations, performance-sensitive paths, concurrency, backwards compatibility, test coverage, diff size, novelty. A documentation typo fix is low without being assessed.

### Selection

**The registry governs who is convened.** `review-board-dispatch § The reviewer registry` holds one row per reviewer — domain, activate-when, never-when, invalidated-by — and a reviewer with no matching row does not run. Activation criteria are explicit for every reviewer; *a PR exists* is not one of them.

**Convene the smallest sufficient panel.** Before invoking a reviewer, the question is whether it has a reasonable chance of changing the outcome. If not, not invoking it is the correct decision rather than a corner cut.

Two floors are not subject to selection:

- **`@validation-agent` runs on every PR**, last and alone. It is machine evidence, and machine evidence is worthless if it depends on someone first deciding it was relevant.
- **`@safety-sentinel` runs on every diff touching a safety path**, at every risk tier, never softened by a budget. A safety path does not become less dangerous because the diff is small.

### Lanes are claimed before they are dispatched

**The review ledger is a claim table written before the work starts, not a summary written after it.** Three clauses, each closing a failure observed across PRs #135, #140, #141 and #142 and recorded in issue #144:

- **A lane is claimed before it is dispatched.** The judge writes every selected lane into the table first, and a lane with no claim row is not convened. A second dispatch against a claimed lane at the same head SHA is **refused**, not discouraged.
- **A verdict is durable before its reviewer's pass ends.** The reviewer records the ruling into its own row, itself. A parent that then dies costs synthesis, not the review.
- **The judge synthesises by reading the table.** Completeness is a property of the table and never of what happened to come back. That is what makes the absence of a verdict distinguishable from the absence of a dispatch, which is the confusion that put several judges on one PR.

**A refusal from the table is an answer, and stopping is the correct response to it.** An agent that retries a refusal, or reads a nonzero status as a transient failure, reinstates the duplicate dispatch this exists to close and does it while reporting success. That is the one way the mechanism fails, and it fails that way because **the table cannot force a dispatch to ask first**: it is durable state, not a gate on the Agent tool. What enforces it is this rule and the caller-side obligations it delegates to.

**A pause stops new claims and lets claimed lanes finish.** That is what a pause should mean, and it is enforceable here in the way a declared intention was not.

The mechanics are `review-board-dispatch § The claim table`; the reviewer's own obligation is `review-verdicts § Record your verdict into its row before your pass ends`; the judge's two are `@pr-judge § Phase 4` and `@pr-judge § Phase 10`.

### Verdicts

Each convened reviewer returns **`pass`**, **`revise`**, or **`blocker`**, with a confidence and severity-tagged findings. The schema and the evidence obligations are both in `review-verdicts`.

- **Enumerate or certify.** A `revise` or `blocker` with no concrete finding is invalid and goes back — it is an impression where a verdict was asked for. So is a `pass` that mentions an actionable problem without filing it: the problem is a finding with an owner, or it is informational and says so.
- **Evidence or the verdict does not count.** A review that did not inspect the actual code or diff is invalid on its face. The full obligation — the `VERIFIED INDEPENDENTLY` / `ACCEPTED ON TRUST` block — is in `review-verdicts § A reviewer does not accept a claim it could check`, and it is the standard, not a formality.
- **N/A is not a courtesy pass.** A convened reviewer that finds its lane genuinely untouched returns `N/A`. It does not pass. An agent that passes because it found nothing to examine has recorded an approval it never performed, and that matters later, when the question is who actually approved something.

### Findings and their owners

**Every finding resolves into exactly one of five outcomes, and each carries a named owner:**

| Resolution | Meaning | What it triggers |
|---|---|---|
| `required_change` | it should actually change, now | a revision cycle |
| `accepted_risk` | real, and not worth another cycle | nothing — recorded with its owner and reason |
| `invalid_finding` | out of lane, a misread, or contradicted by a named artifact | nothing — recorded with the reason |
| `future_work` | **valid, and outside this item's scope** | a traceable issue reference — never a revision |
| `informational` | no action is called for | nothing — recorded |

There is no *approved with suggestions*. A suggestion nobody owns is how a real defect leaves the room dressed as politeness.

**`future_work` is the outcome that keeps the loop from perfecting itself to a standstill.** Before it existed, a valid improvement outside scope had two homes and both were wrong: `required_change`, which starts a cycle the item never asked for, or a passing remark, which loses it. So the judge records it as a **traceable issue reference** — or hands it to `@engineering-lead` to route where the scope decision is not the judge's to make. It is **never a revision trigger and never lost**, and those two clauses are equally load-bearing: dropping the first is how an autonomous loop refactors forever, dropping the second is how it learns to call real findings out-of-scope.

`invalid_finding` must state *why* — a reviewer out of lane, a misread of the diff, a claim the artifact contradicts. Ruling it invalid without a reason is the judge substituting its own opinion for a reviewer's.

### Duplicate findings are one finding

**Two reviewers naming the same file, the same location, the same root cause, and the same required change have found one defect, not two.** The judge normalizes them into a single finding whose `supported_by` lists each reviewer's own ID:

```yaml
finding: CORE-07
supported_by: [go-quality: GOQ-03, maintainability: MAINT-02]
```

**Multiple votes strengthen the evidence; they never multiply the fix.** Left unnormalized they become two tasks, two revisions, or a count of "8 findings" that overstates the state of the PR and drives a cycle the work does not need.

### The merge decision

The judge decides on **risk and confidence together**:

| | High confidence | Low confidence |
|---|---|---|
| **Low risk** | Merge | Request only the missing evidence |
| **Medium risk** | Merge if no blocker | Targeted follow-up on the weak point |
| **High risk** | Merge only if every mandatory high-tier reviewer passes | Targeted deeper review |

In all cases: **zero unresolved `required_change`** among the convened reviewers, machine evidence green, and the traceability chain intact.

**The answer to uncertainty is never "run everybody again."** `@confidence-assessor` names the specific weak point — a shallow review, a conflict, an unexplained finding — and follow-up is one reviewer and one question.

### The stopping rule

**The loop stops, and merges, when all five of these hold:**

```
required_changes: 0 · machine_gates: green · required_review_lanes: satisfied
confidence: sufficient_for_risk · human_gate: false
```

At that point the ruling is MERGE. **Do not ask agents for final thoughts. Do not perform one last review. Do not run a polish cycle.**

No lane requires subjective perfection, and none is entitled to it. A PR passes because the required behaviour is satisfied, the gates are green, no required change is unresolved, the risk is acceptable, and the evidence is sufficient — not because every reviewer has run out of things it could imagine improving. **There is always another refactor.** An autonomous loop with no stopping rule does not converge on quality; it converges on whatever the last reviewer happened to notice, at full price per lap.

**Stopping is part of correctness here, not a concession to budget.** A judge that keeps going when the five conditions are met is not being careful — it is failing to make the decision it exists to make.

### Merge and readiness

**A completed merge that can satisfy a downstream dependency must reliably cause readiness reconciliation** — the route is deterministic and has exactly one shape: the fingerprint's `main` component wakes `@engineering-lead`, which dispatches `@scrum-master` on `trigger: {type: merge_completed}`, which reconciles the merged items to Done, runs `scripts/loop/dependents.sh` per completed story, and evaluates the `eligible:` list for Ready. It never wakes `@backlog-dependency-planner`: satisfaction is not a graph event.

### Revision, and what stays valid

A remand produces a **revision packet**, not a restart: the required changes with their owners and scope, and the list of reviewers to re-run afterwards. `@worker-manager` activates only the specialist that owns each finding.

**After a patch, only reviewers whose domain intersects the new diff re-run.** Previous verdicts remain valid otherwise, and are carried forward marked with the SHA they were issued against. A documentation-only revision does not invalidate security, data integrity, or performance; a schema revision may invalidate data integrity, backend correctness, and performance, and almost certainly does not invalidate accessibility.

**A changed SHA is not, by itself, an invalidation.** The two kinds of rerun are governed differently and conflating them is what re-convened the bench on one-line fixes:

- **Machine verification** — build, lint, tests, gates, `@validation-agent` — **reruns on every commit.** It is cheap, deterministic, and worthless if stale.
- **Semantic specialist review** reruns **only when the revision intersects that reviewer's domain**, on the files-and-domain test in `review-board-dispatch § The intersection test`. "The PR changed" is not a reason; *the PR changed in this lane* is.

### The minimal-patch revision law

**A revision changes the smallest surface that resolves the named finding.** No unrelated cleanup, no opportunistic refactor, no cross-file formatting, no while-I-am-here improvement. Before touching an additional file, the implementing specialist asks one question — **does this file have to change to resolve the named finding?** If the answer is no, it is not touched.

This is a token-efficiency requirement as much as a review-hygiene one, and the mechanism is worth being explicit about: every extra changed surface is a surface that can meet some reviewer's `Invalidated by` condition, so an unrelated tidy-up in a revision cycle does not merely add lines — it re-convenes reviewers whose verdicts would otherwise have carried. Desirable-but-unrelated cleanup becomes `future_work` with a traceable issue.

Initial implementation may refactor coherently; **review remediation patches narrowly.** `@worker-manager` states this law in every revision dispatch, because the specialist receiving a remand is precisely the agent most tempted to improve one more thing while it is in there.

**The judge keeps a review ledger** as a structured comment on the PR — reviewer, verdict, confidence, diff SHA reviewed, domain — updated every cycle. The ledger is what makes carried-forward validity checkable by someone who was not there, which is the same reason the red-demonstration rule prefers evidence to sequence.

### The cycle-inflation rule

**At cycle 3 or later, a new `low`-severity finding located in text the previous cycle created resolves to `future_work` by default, not `required_change`.** The default is a starting point and not a bar: **overriding it requires a stated reason naming what makes that instance different**, recorded with the resolution like any other, and a judge that can state one has satisfied the rule rather than escaped it.

**Four things it never reaches**, and the list is closed: a finding at `medium` severity or above · anything on a safety path · a security finding · a **false statement of fact**. Those resolve on their merits at every cycle. The rule bounds the *precision* work a late cycle generates about itself, and precision is the only thing it bounds.

**Two of those four are terms of art, and they are defined here because nothing else defines them.** An exclusion a judge cannot apply the same way twice is not an exclusion.

- **The severity scale is `blocker` > `high` > `medium` > `low` > `info`** — the vocabulary every reviewer's finding schema already carries, stated here as the ordinal it is. ***`medium` or above*** therefore means `blocker`, `high`, or `medium`. A finding that arrives without a severity is not `low` by default: the judge assigns one, and records it, before this rule is applied to it.
- **A false statement of fact is a sentence a reader could act on that the artifact it names does not support** — a citation to a section that does not say it, a count that does not match what was counted, a claim about a command's output the command does not produce, an attribution to the wrong author or artifact. **The test is the check, not the wording.** It is a false statement of fact if a reader can open the named artifact and find otherwise, however the sentence is phrased; hedging it into an opinion does not move it out of the class, and only correcting it or removing the claim does. This matters on a prose corpus, where the alternative to a definition is a rule escapable by rewording.

**The evidence is three independent records, and only two of them are a judge's.** PR #67's cycle 2 is **`@linus-security-critic`'s verdict**, persisted on the PR by `@engineering-lead` under *"no ruling issued"* because the judge process ended before it could rule: eight new findings, *"all second-order — every one attaches to prose that did not exist in cycle 1."* A reviewer named the shape from inside the panel, which is where it is first visible. PR #120's **cycle-2 judgment comment** recorded that *"four of the five new findings are precision defects in text that the discharge of cycle 1's findings created"* — and scoped its next cycle to the named edits for that reason. PR #110's final ruling put the mechanism beyond inference: the `DOC-16` fix **mechanically produced** `DOC-17`, because *"a faithful, correct execution of the packet generated the next finding."* A loop whose input is its own output terminates on the budget rather than on the work.

**Two of those three records are cycle 2, and the rule still begins at cycle 3 — they establish the mechanism, not the threshold.** Cycle 2 is the earliest a finding *can* sit in text a previous cycle created, so it is where the shape first becomes visible, and it is also the panel's **first** reading of that correction — ordinary review, and where a regression the fix introduced actually lands. Cycle 3 is the first reading of a correction to a correction the panel has already read once. The floor sits one cycle above the first occurrence deliberately: the record that shows whether this is working is written at the cycle it declines to bind.

**The number is why this is a rule rather than a preference.** A revision cycle costs roughly **570,000 tokens** — the judge, the worker, and the re-run lanes — measured across one session of live operation and recorded in issue #128. That is what the default is worth each time it holds, weighed against one `low` finding which `future_work` records as a traceable issue and therefore does not lose. `§ Findings and their owners` governs that half unchanged: never a revision trigger, and never lost.

### Convergence and budgets

**3 autonomous revision cycles normally; 5 for a high-risk PR.**

Each cycle records unresolved findings, newly introduced findings, findings resolved, diff size, and the movement in risk and confidence. That is what *converging* means here, and it is a determination the judge makes rather than an assumption it holds.

**Before exceeding the budget, the judge must determine why convergence failed.** Conflicting requirement · unstable architecture · faulty reviewer · overly broad implementation · reviewer disagreement · ambiguous acceptance criteria · an implementation that keeps reintroducing regressions. Solve the cause; repeating the loop is not a plan.

**Every additional cycle justifies itself before it starts.** The judge records what the next one is expected to produce:

```yaml
next_cycle_justification:
  unresolved_required_changes: [SEC-01]
  expected_reviewers: [linus-security-critic]
  expected_new_information: "whether the rotation fix closes the reuse window"
```

**If `expected_new_information` is empty, do not start the cycle.** A cycle that can name no new information it expects to obtain is a cycle that will produce the verdicts it already has. Note the narrower point inside this: fixing an implementation issue justifies **targeted validation** of the fix; it does not, by itself, justify semantic re-review from every lane that ran before.

**8 rounds remains the absolute ceiling**, and reaching it escalates to the repository owner with the full cycle history. It is kept because deliberately exacting critics can deadlock, with a fix for one reviewer's objection creating another's, and that is not hypothetical on this repository: during the review of the product concept — since split into `SPECIFICATION.md` and its companions — a first pass produced 13 findings, and the round of fixes addressing them introduced **three of the four blockers** found by the second pass. Under the budgets above, reaching the ceiling is now itself a reportable failure rather than a normal ending.

### Root cause

**Every finding is classified by root cause:** implementation · requirement · architecture · design · test · infrastructure · dependency · planning.

A requirement defect routes back to `@requirements-engineer`. An architectural contradiction routes to the ADR process. A **`dependency` or `planning`** defect — work that failed because it ran in the wrong structural order, a consumer built before the contract it consumes — routes to `@backlog-dependency-planner`, which owns the backlog graph that order came from. Where the faulty artifact is instead **the story cut itself** — a story that cannot be built as written, partial or overlapping coverage of the requirements it claims, an acceptance criterion that decomposition should have split — it routes to `@requirements-story-organizer`, which owns the nodes; correcting an edge around a badly cut node leaves the node wrong. **Repeatedly patching code around an upstream defect is forbidden** — it is how a broken requirement becomes permanent, expensive, and invisible. Correct the highest faulty artifact and let the change propagate down.

## Execution shapes

**These are normative, not illustrative.** They state what the execution graph is *supposed* to look like at each size of work, so that a graph much larger than its shape is visible as a defect rather than admired as thoroughness. The graph scales with risk and scope — and only with those.

| The work | The shape |
|---|---|
| **A tiny documentation fix** | deterministic classifier → `@docs-reviewer` *if semantic content changed* → validation if required → judge. **1–2 LLM agents.** Not a ten-agent chain through lead, scrum master, coordinator, worker manager, risk assessor, writer, reviewer, summarizer, confidence, judge. |
| **A small Go bug fix** | `@worker-manager` → `@change-risk-assessor` → `@go-worker` → `@validation-agent` → 1–2 relevant reviewers → `@pr-judge`. `@confidence-assessor` only if the evidence or a disagreement warrants it. |
| **A small revision after a security finding** | revision packet → the owning specialist → `@validation-agent` → the security reviewer → correctness *only if semantic behaviour changed* → `@pr-judge`. **All unrelated prior verdicts carry forward.** |
| **A high-risk architecture change** | `@engineering-lead` → the requirements or architecture decision → `@worker-manager` → several specialists → `@change-risk-assessor` → a larger selected panel → `@confidence-assessor` → `@pr-judge`. |

The last row is the one that keeps the others honest: **the large graph is not forbidden, it is earned.** A ten-agent chain on a high-risk architecture change is correct, and the same chain on a typo is the same cost buying nothing.

**Bloat is a signal to inspect, not an automatic failure.** Three signals worth stopping on: a small change with more than ~7 meaningful specialist executions · a low-risk PR with more than 3 domain reviewers · a revision cycle invoking more agents than the original without an increase in risk. Each means look at the routing; the decision about whether it was justified stays semantic.

## Review identity

Review agents comment on the pull request or issue under a **separate GitHub identity** from the repository owner's.

Authorship and approval must not share a signature. Self-approval is not review, and a distinct identity makes the boundary visible in the history rather than merely intended.

Authentication uses a personal access token held in the machine environment as **`GH_JUDGE_TOKEN`**.

> **The token is referenced by name only and must never be read, printed, logged, or echoed.** Pass it to `gh` through the environment. Nothing should ever cause its value to appear in a command, a comment, a log, or a transcript.

Every review comment is signed:

```
/ The Review Ninja
```

## Escalation and human judgement

**Human escalation is an exceptional path, and it is the same path everywhere.** Ordinary engineering uncertainty is decided, not escalated: where several solutions are valid, prefer compliance with the specification, then architecture, then design, then existing codebase patterns, then lower complexity, smaller blast radius, easier reversibility, stronger testability, maintainability, and least surprising behaviour. Record the decision instead of asking.

**First, check whether it has already been decided.** Before reasoning about any ambiguity, search `docs/Requirements/DECISIONS.md`, the ADRs in `docs/adr/`, the requirement record itself, and the board or PR record — if the question is settled, **reuse the answer rather than re-litigating it.** Re-deciding a settled question costs a full execution and, at best, reproduces the answer that already existed.

A question reaches the human only when one of these holds: **product intent is undefined** and the source documents cannot distinguish between materially different behaviours; **two product documents conflict**; the choice is a **business tradeoff** requiring knowledge not in the repository, requirements, or architecture; the decision is **irreversible or high-impact** — a destructive migration, a major architectural replacement, a substantial scope increase, an externally visible breaking change; or **risk exceeds autonomous authority** and the change cannot be made safe within established project constraints.

Every escalation carries a precise question, why the artifacts cannot answer it, the options, a **recommended option**, and the impact. A question without a recommendation is work handed back.

Two categories always reach a human rather than being settled by agent consensus, regardless of verdicts:

**Requirements marked as human-verified.** Where the verification method says judgement is required, agents can confirm that a thing was built but not that it was built *well*. Whether a route recommendation is genuinely good is the product's real quality bar and is not machine-checkable.

**Changes touching safety rules or accessibility classification.** `SPECIFICATION.md` separates safety requirements the data can enforce from those it cannot, and is explicit that a rule the system cannot verify is not a safeguard. Changes in that area carry consequences beyond code quality and warrant a human decision.

On both, a clean board is a **recommendation to the human**, not an approval.
