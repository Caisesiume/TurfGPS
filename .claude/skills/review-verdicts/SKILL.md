---
name: review-verdicts
description: What a convened TurfGPS reviewer returns, and the standard that verdict is measured against — the reviewer verdict schema with its findings, severity, confidence and residual risk, the unsatisfiable verdict that keeps insufficient evidence distinct from low confidence, and the evidence law: a reviewer does not accept a claim it could check, the VERIFIED INDEPENDENTLY / ACCEPTED ON TRUST block, how far the obligation reaches, and the two incidents that made it a rule rather than a habit. Load alongside `agent-handoffs` before returning any review verdict.
---

# Review verdicts — the schema and the evidence law

*Split out of `agent-handoffs` on 28 August 2026, which keeps the envelope, the size limit, and the output caps and is not restated here. This file is what a convened reviewer returns and the standard it is held to; a reviewer loads both, and nothing else.*

## Reviewer verdict

Returned by every convened reviewer to `@pr-judge`.

**The block comes first and any licensed prose comes after it**, per `agent-handoffs § The structured block comes first`, which is also where `artifact:` and `prose_licence:` are defined.

```yaml
artifact: reviewer_verdict
prose_licence: none
reviewer: security
status: valid_review       # valid_review | unsatisfiable
inspected:
  diff: true
files_inspected:
  - src/auth/session.ts
findings:
  - id: SEC-01
    severity: high
    file: src/auth/session.ts
    line: 142
    description: refresh tokens can be reused after rotation
    required_change: invalidate the old refresh token on successful rotation
    root_cause: implementation
verdict: revise            # pass | revise | blocker | insufficient_evidence | N/A
confidence: 0.96           # a number, or `unassessed` — never a number standing in for one
residual_risk:
needs_followup: false
evidence: |
  VERIFIED INDEPENDENTLY:
    · …
  ACCEPTED ON TRUST:
    · …
```

**Mandatory keys:** `artifact` · `prose_licence` · `reviewer` · `status` · `inspected` · `files_inspected` · `findings` · `verdict` · `confidence` · `evidence`. An empty `findings: []` is an answer; an absent `findings` is not, and the judge cannot tell it from a lane that never looked.

**`inspected: diff: false` makes the verdict automatically invalid** and the judge ignores it — unless the verdict is the unsatisfiable one defined in `§ Insufficient evidence is not low confidence` below, which is the one shape that reports `false` honestly. That flag is the floor; the standard is the `evidence` block, defined in `§ The report block`.

Return decision-relevant data only. Deep internal analysis is welcome; only its conclusions enter the parent's context.

**Every finding a reviewer files will be resolved by the judge into exactly one of five outcomes** — `required_change` · `accepted_risk` · `invalid_finding` · `future_work` · `informational`. A reviewer does not resolve its own findings, but knowing the vocabulary changes how it writes them: a finding filed as though everything must block is a finding the judge has to reclassify, and one filed as a passing remark is one that disappears. The five are defined in `docs/DELIVERY.md § Findings and their owners`.

### Insufficient evidence is not low confidence

**A lane that could not gather evidence and a lane that gathered weak evidence are different results, and one field cannot carry both.** `verdict: insufficient_evidence` says the review could not be performed. A low `confidence` says it was performed and the reviewer does not trust the answer. Collapsing the first into the second hands the judge a number where there was no measurement — and the judge then weighs an unrun lane against a run one, which is #144's ledger-corruption class arriving through vocabulary instead of through a missing row.

A lane that cannot be satisfied returns:

```yaml
artifact: reviewer_verdict
prose_licence: none
reviewer: security
status: unsatisfiable
inspected:
  diff: false
files_inspected: []
findings: []
verdict: insufficient_evidence
confidence: unassessed
evidence_gap:
  what: the diff — the PR body was reachable, the patch was not
  why: tooling                # tooling | access | artifact_absent | out_of_scope
  closable_by: a dispatch carrying the patch, or a lane holding repository access
evidence: |
  VERIFIED INDEPENDENTLY:
    · nothing — see evidence_gap
  ACCEPTED ON TRUST:
    · nothing was accepted; no verdict was formed
```

**`confidence: unassessed` is mandatory with this verdict and a number is forbidden.** Any number offered here is manufactured, and manufacturing one is the failure a reviewer exists to catch in others.

**`evidence_gap` is mandatory too, and `closable_by` is the load-bearing field.** It is what separates a gap a dispatch can close — send the artifact inline — from one it cannot: **a follow-up question cannot close a tooling gap**, and a judge that does not know which it is facing will spend a cycle asking.

**The judge records an unsatisfiable lane as unsatisfiable and converts it into neither a pass nor a fail.** Ignoring it, as `§ Reviewer verdict` above has the judge ignore an ordinary `diff: false`, is how an absent measurement becomes a silently passed lane; reading `unassessed` as a low number is how it becomes a weak one. Both are the same error in opposite directions, and the ledger row carries the word rather than a value.

**The class was first recorded on PR #135, 29 August 2026.** `@confidence-assessor` holds `Read, Grep, Glob` and no Bash or GitHub access, so it could not read the verdicts it had been convened to weigh; it checked for a local mirror before reporting the gap, then returned `evidence_quality: unknown` — *"not weak — unassessed"* — with `followup: none`, *"a reviewer follow-up doesn't fix a tooling gap."* The cycle-3 ruling recorded the lane **unsatisfiable, not low**. That vocabulary is this section, and the payload half of it is `handoff-payloads § Confidence assessment`.

## A reviewer does not accept a claim it could check

*Ratified in `docs/adr/ADR-0001-artifact-driven-agent-org.md § D5`; moved out of `review-board-dispatch` by `docs/adr/ADR-0002-token-efficiency.md § O1`, and into this file on 28 August 2026 when `agent-handoffs` was split. This is the home of the evidence law, and it still sits beside the verdict schema it measures — one skill holds both, and holding them here means a reviewer loads neither the judge's dispatch mechanics nor the payload schemas of roles it does not have.*

Everything in a dispatch's case file is a **claim**. None of it is evidence, and a reviewer accepts none of it where the means to check it is at hand: not the PR body's account of what changed, not the author's stated gate results, not a count in a commit message, not "the cited section says X." Where the diff, the tree, the section, or the command is available, **the reviewer looks**.

Checking is read-only. `review-board-dispatch § The read-only clause (learned the hard way)` still binds — read the diff, read the tree, open the cited heading, run a command that only reads. A check that would write anything is not available to you; that claim goes under `ACCEPTED ON TRUST` naming `@validation-agent` as its owner.

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

**This block is the standard; `inspected: diff: true` is only the floor.** The verdict schema above carries that flag, and a verdict reporting `false` is automatically invalid. But a flag is a self-assessment and the block is an enumeration, so a verdict may satisfy the flag and still fail here. The judge checks the block.

### What the obligation reaches

**It reaches what the verdict rests on** — any claim your own verdict depends on. It is not a re-run of the suite: that is `@validation-agent`'s job, it runs last and alone for exactly that reason, and duplicating it across the bench would double the cost of every PR to learn nothing new. Where a claim's truth would not move your verdict, take it on trust and list it.

**`ACCEPTED ON TRUST` is not a dumping ground.** A claim the verdict rests on, written in that half, **is the finding** — the reviewer has just recorded that its own verdict is unsupported. Check it, or file the gap as a finding, but do not list it and rule as though you had.

### Why this is a rule and not a habit

Both of these were found by an agent that checked a premise it had been handed, and neither was found by the pass meant to find it.

- **The board agent that could not see the board.** Its own definition told it an empty board was "a complete and correct run" and to stop; the board held **37 items**. That instruction was reachable on every run, and the run that reached it would have reported an empty backlog and recorded itself as complete. Found 4 August 2026 while sweeping citation delimiters — `c091046`.
- **The gates whose report could not say where they ran.** `Architecture.md § D8` puts the Go module in `service/`. The gate block carried no working directory, and neither did the report line it then prescribed — that form had no field for one, so no line written in it could name the tree it measured, whatever the commands had actually done. Whether the underlying results differed at all varies by gate, and `local-gates § Backend (Go)` is where that is measured; the form could carry none of it either way. `local-gates § The law` now requires that field. Eleven agent files and the PR-body template carried the same directory-less copy, so the path ran unbroken from command to report line. Closed before any PR in this repository existed to carry it; the instrument, not a reviewer, was the thing that would have lied. Found 5 August 2026 because a layout decision recorded its own cost honestly — `d6a7e3e`, `1928a28`. **The mechanism first recorded here — all five commands passing vacuously over an empty root — was itself measured false and retracted on 22 August 2026; `Architecture.md § D8` carries the retraction and `local-gates § Backend (Go)` the measurement. The incident stands and so does the law it justifies**: the report still could not say where it ran, and the quiet gate is quiet from the wrong directory for a reason that says nothing about what it read.

Neither is something a reviewer catches by reading attentively. Both were **instruments reporting success**, and the only thing that separated the report from the truth was an agent running the thing itself.

