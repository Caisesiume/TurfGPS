---
name: review-board-dispatch
description: Mechanics for safely convening TurfGPS's review boards (craft board, Linus board, Go pipeline, SafetySentinel, ValidationAgent) — the read-only clause, tree-integrity verification, sequencing, review identity, and the scoring law from docs/DELIVERY.md. Use whenever dispatching any reviewer.
---

# Review Board Dispatch — Safe Convening Mechanics

`docs/DELIVERY.md` is the authority on what a verdict means. This skill is the authority on how a board is convened without corrupting the tree. Where the two appear to disagree, `DELIVERY.md` wins and the disagreement is a defect in this file.

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

1. Critics **parallel within a board**; boards may run in parallel with each other.
2. **ValidationAgent runs LAST and ALONE** — it executes builds/tests that must not race a critic's probing.
3. Review the **PR diff against `main`**, checked out in a dedicated review worktree (never the trunk tree, never a worker's worktree).

## The case file (same for every reviewer)

Task/story name and link · acceptance criteria + requirement codes · files modified · full diff · **safety paths touched** (see the `safety-path-checklist` skill) · gate results (see `local-gates`) · implementation summary · the read-only clause.

## Scoring law — from `docs/DELIVERY.md`

- **Each reviewer scores 0–10.** The item is shippable only when the average across all participating agents is **10.00**.
- **That is a unanimity gate, not an average.** At an average of exactly 10.00 a single sub-10 score blocks. It cannot be diluted by uninvolved agents handing out easy 10s — which is the usual way averaged review scores decay.
- **Enumerate or certify.** Every agent scoring below 10 must state precisely what would earn a 10. A score without an actionable reason is not a review: return it to the reviewer to either enumerate the gap or certify 10. A "9, nothing blocks" is a remand, not a pass.
- **N/A is not a courtesy 10.** An agent whose quality attribute the diff does not touch returns **N/A** and is excluded from the average. A documentation change has no meaningful scalability dimension, and an agent that scores 10 because it found nothing to examine has recorded a pass it never performed. N/A also means not every agent runs on every change.
- **The bench re-convenes whole** after any revision — partial re-review is not a thing.
- **Contradictory demands between reviewers** = CONFLICT → escalate to the human; never average or silently pick a side.
- **Round cap: 8.** After 8 revision rounds without reaching 10.00, the item **escalates to the repository owner** with full cycle history. Unanimity plus deliberately exacting critics can deadlock — a fix for one reviewer's objection creating another's — and this has already happened once on this repository.

## Always escalate to a human

Two categories are never settled by agent consensus, per `DELIVERY.md`:

- **Requirements marked as human-verified.** Agents can confirm a thing was built, not that it was built *well*. Whether a route recommendation is genuinely good is this product's real quality bar and is not machine-checkable.
- **Changes touching safety rules or accessibility classification.** `SPECIFICATION.md` separates safety requirements the data can enforce from those it cannot, and is explicit that a rule the system cannot verify is not a safeguard.

## Review identity

Review comments are posted under a **separate GitHub identity** from the repository owner's — authorship and approval must not share a signature. Authentication uses the personal access token held in the machine environment as **`GH_JUDGE_TOKEN`**.

> **The token is referenced by name only and must never be read, printed, logged, or echoed.** Pass it to `gh` through the environment. Nothing may cause its value to appear in a command, a comment, a log, or a transcript.

```bash
GH="/c/Program Files/GitHub CLI/gh.exe"
GH_TOKEN="$GH_JUDGE_TOKEN" "$GH" pr comment <N> --body-file <file>
```

Every review comment is signed on its own final line:

```
/ The Review Ninja
```
