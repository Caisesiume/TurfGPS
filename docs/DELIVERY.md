# Delivery and review model

How work on TurfGPS is tracked, reviewed, and judged shippable.

This describes the target working model. It is not yet set up.

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

Every item reaching test or verification is reviewed by a range of specialist agents before it is judged shippable or sent back for revision.

Most agents own a **single software quality attribute** — performance, modularity, scalability, security, maintainability, evolvability, over-engineering, user experience, documentation, test coverage, and so on. Reviewing one dimension well beats reviewing everything shallowly, and it makes each verdict attributable.

Alongside them sit the **"Linus" critics**, modelled on Linus Torvalds' direct review style: blunt, and willing to raise a small defect loudly rather than let it pass for being small.

### Scoring

Each review agent scores the diff from **0 to 10**.

The item is shippable only when the average across all participating agents is **10.00**.

Anything less sends it back for revision, and **every agent that scored below 10 must state precisely what would earn a 10**. A score without an actionable reason is not a review.

Note what this rule actually is: at an average of exactly 10.00, a single sub-10 score blocks. It is a **unanimity gate**, not an average. That is deliberate, and it has a useful property — it cannot be diluted by uninvolved agents handing out easy 10s, which is the usual way averaged review scores decay.

### Not applicable

An agent whose quality attribute a diff does not touch returns **N/A** and is excluded from the average.

It does not award a courtesy 10. A documentation change has no meaningful scalability dimension, and an agent that scores 10 because it found nothing to examine has recorded a pass it never performed. That distinction matters later, when the question is who actually approved something.

N/A also means not every agent needs to run on every change.

### Round cap

After **8 revision rounds** without reaching 10.00, the item **escalates to the repository owner** rather than cycling further.

Unanimity plus deliberately exacting critics can deadlock, with a fix for one reviewer's objection creating another's. This is not hypothetical: during the review of the product concept — since split into `SPECIFICATION.md` and its companions — a first pass produced 13 findings, and the round of fixes addressing them introduced **three of the four blockers** found by the second pass. The cap keeps work moving when convergence stalls.

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

Two categories should always reach a human rather than being settled by agent consensus:

**Requirements marked as human-verified.** Where the verification method says judgement is required, agents can confirm that a thing was built but not that it was built *well*. Whether a route recommendation is genuinely good is the product's real quality bar and is not machine-checkable.

**Changes touching safety rules or accessibility classification.** `SPECIFICATION.md` separates safety requirements the data can enforce from those it cannot, and is explicit that a rule the system cannot verify is not a safeguard. Changes in that area carry consequences beyond code quality and warrant a human decision.
