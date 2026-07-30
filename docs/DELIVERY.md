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

Unanimity plus deliberately exacting critics can deadlock, with a fix for one reviewer's objection creating another's. This is not hypothetical: during the review of `Concept.md`, a first pass produced 13 findings, and the round of fixes addressing them introduced **three of the four blockers** found by the second pass. The cap keeps work moving when convergence stalls.

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

**Changes touching safety rules or accessibility classification.** `Concept.md` separates safety requirements the data can enforce from those it cannot, and is explicit that a rule the system cannot verify is not a safeguard. Changes in that area carry consequences beyond code quality and warrant a human decision.
