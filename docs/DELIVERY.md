# Delivery and review model

How work on TurfGPS is tracked, reviewed, and judged shippable.

Documentation work runs through this model today; the code half of it activates with the Go and frontend stacks. The agent organization that applies it is ratified in `docs/adr/ADR-0001-artifact-driven-agent-org.md`.

## Work tracking

Epics and User Stories are extracted from the requirements **before implementation begins**, and stored as GitHub issues on a GitHub Project board.

**Milestones serve as Epics**, tying related issues together.

Agents work the board using the `gh` CLI and the GitHub MCP connector.

### Requirements come first

Issues are derived from the requirements in `Requirements/`, and cite the requirement IDs they satisfy. This is the dependency that governs sequencing: a board built before the requirements exist would contain guesses, and review agents would have nothing objective to check work against.

Consequently `SPECIFICATION.md` and `Requirements/` block the board. `DESIGN.md` and `DEPLOYMENT.md` do not, and can follow.

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

Confirming a demonstration instead of reading the claim of one is an instance of `review-board-dispatch § A reviewer does not accept a claim it could check`, and is governed there rather than here. One consequence is worth naming because it is not obvious: **re-running a demonstration is a write**, since it neutralises an implementation in a tree, so it is not available to a critic under that section's read-only bound and belongs in the `ACCEPTED ON TRUST` half. What is available read-only is the shape of the recorded evidence — that an entry exists for every `test`-verified criterion the item claims, and that each carries an assertion's message rather than a build failure.

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

### Verdicts

Each convened reviewer returns **`pass`**, **`revise`**, or **`blocker`**, with a confidence and severity-tagged findings. The schema is in the `agent-handoffs` skill; the evidence obligations are in `review-board-dispatch`.

- **Enumerate or certify.** A `revise` or `blocker` with no concrete finding is invalid and goes back — it is an impression where a verdict was asked for. So is a `pass` that mentions an actionable problem without filing it: the problem is a finding with an owner, or it is informational and says so.
- **Evidence or the verdict does not count.** A review that did not inspect the actual code or diff is invalid on its face. The full obligation — the `VERIFIED INDEPENDENTLY` / `ACCEPTED ON TRUST` block — is in `review-board-dispatch § A reviewer does not accept a claim it could check`, and it is the standard, not a formality.
- **N/A is not a courtesy pass.** A convened reviewer that finds its lane genuinely untouched returns `N/A`. It does not pass. An agent that passes because it found nothing to examine has recorded an approval it never performed, and that matters later, when the question is who actually approved something.

### Findings and their owners

**Every actionable finding resolves into exactly one of `required_change`, `accepted_risk`, or `invalid_finding`, and each carries a named owner.**

There is no *approved with suggestions*. A suggestion nobody owns is how a real defect leaves the room dressed as politeness. If something should actually change, it is a `required_change`. If it should not change now, it is an `accepted_risk` recorded with its owner, or an informational observation explicitly marked non-actionable.

`invalid_finding` is a real outcome and must state why — a reviewer out of lane, a misread of the diff, a claim the artifact contradicts. Ruling it invalid without a reason is the judge substituting its own opinion for a reviewer's.

### The merge decision

The judge decides on **risk and confidence together**:

| | High confidence | Low confidence |
|---|---|---|
| **Low risk** | Merge | Request only the missing evidence |
| **Medium risk** | Merge if no blocker | Targeted follow-up on the weak point |
| **High risk** | Merge only if every mandatory high-tier reviewer passes | Targeted deeper review |

In all cases: **zero unresolved `required_change`** among the convened reviewers, machine evidence green, and the traceability chain intact.

**The answer to uncertainty is never "run everybody again."** `@confidence-assessor` names the specific weak point — a shallow review, a conflict, an unexplained finding — and follow-up is one reviewer and one question.

### Revision, and what stays valid

A remand produces a **revision packet**, not a restart: the required changes with their owners and scope, and the list of reviewers to re-run afterwards. `@worker-manager` activates only the specialist that owns each finding.

**After a patch, only reviewers whose domain intersects the new diff re-run.** Previous verdicts remain valid otherwise, and are carried forward marked with the SHA they were issued against. A documentation-only revision does not invalidate security, data integrity, or performance; a schema revision may invalidate data integrity, backend correctness, and performance, and almost certainly does not invalidate accessibility.

**The judge keeps a review ledger** as a structured comment on the PR — reviewer, verdict, confidence, diff SHA reviewed, domain — updated every cycle. The ledger is what makes carried-forward validity checkable by someone who was not there, which is the same reason the red-demonstration rule prefers evidence to sequence.

### Convergence and budgets

**3 autonomous revision cycles normally; 5 for a high-risk PR.**

Each cycle records unresolved findings, newly introduced findings, findings resolved, diff size, and the movement in risk and confidence. That is what *converging* means here, and it is a determination the judge makes rather than an assumption it holds.

**Before exceeding the budget, the judge must determine why convergence failed.** Conflicting requirement · unstable architecture · faulty reviewer · overly broad implementation · reviewer disagreement · ambiguous acceptance criteria · an implementation that keeps reintroducing regressions. Solve the cause; repeating the loop is not a plan.

**8 rounds remains the absolute ceiling**, and reaching it escalates to the repository owner with the full cycle history. It is kept because deliberately exacting critics can deadlock, with a fix for one reviewer's objection creating another's, and that is not hypothetical on this repository: during the review of the product concept — since split into `SPECIFICATION.md` and its companions — a first pass produced 13 findings, and the round of fixes addressing them introduced **three of the four blockers** found by the second pass. Under the budgets above, reaching the ceiling is now itself a reportable failure rather than a normal ending.

### Root cause

**Every finding is classified by root cause:** implementation · requirement · architecture · design · test · infrastructure.

A requirement defect routes back to `@requirements-engineer`. An architectural contradiction routes to the ADR process. **Repeatedly patching code around an upstream defect is forbidden** — it is how a broken requirement becomes permanent, expensive, and invisible. Correct the highest faulty artifact and let the change propagate down.

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

A question reaches the human only when one of these holds: **product intent is undefined** and the source documents cannot distinguish between materially different behaviours; **two product documents conflict**; the choice is a **business tradeoff** requiring knowledge not in the repository, requirements, or architecture; the decision is **irreversible or high-impact** — a destructive migration, a major architectural replacement, a substantial scope increase, an externally visible breaking change; or **risk exceeds autonomous authority** and the change cannot be made safe within established project constraints.

Every escalation carries a precise question, why the artifacts cannot answer it, the options, a **recommended option**, and the impact. A question without a recommendation is work handed back.

Two categories always reach a human rather than being settled by agent consensus, regardless of verdicts:

**Requirements marked as human-verified.** Where the verification method says judgement is required, agents can confirm that a thing was built but not that it was built *well*. Whether a route recommendation is genuinely good is the product's real quality bar and is not machine-checkable.

**Changes touching safety rules or accessibility classification.** `SPECIFICATION.md` separates safety requirements the data can enforce from those it cannot, and is explicit that a rule the system cannot verify is not a safeguard. Changes in that area carry consequences beyond code quality and warrant a human decision.

On both, a clean board is a **recommendation to the human**, not an approval.
