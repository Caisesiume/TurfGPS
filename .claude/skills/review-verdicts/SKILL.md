---
name: review-verdicts
description: What a convened TurfGPS reviewer returns, and the standard that verdict is measured against — the reviewer verdict schema with its findings, severity, confidence and residual risk, the unsatisfiable status that keeps insufficient evidence distinct from low confidence, and the evidence law: a reviewer does not accept a claim it could check, the VERIFIED INDEPENDENTLY / ACCEPTED ON TRUST block, how far the obligation reaches, and the two incidents that made it a rule rather than a habit; and the obligation to record that verdict into its own claim-table row before the pass ends. Load alongside `agent-handoffs` before returning any review verdict.
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
verdict: revise            # pass | revise | blocker | N/A
confidence: 0.96           # a number; `unassessed` belongs to `unsatisfiable` and never here
residual_risk:
needs_followup: false
evidence: |
  VERIFIED INDEPENDENTLY:
    · …
  ACCEPTED ON TRUST:
    · …
```

**Mandatory keys:** `artifact` · `prose_licence` · `reviewer` · `status` · `inspected` · `files_inspected` · `findings` · `confidence` · `evidence`. An empty `findings: []` is an answer; an absent `findings` is not, and the judge cannot tell it from a lane that never looked.

**`verdict` is mandatory under `status: valid_review` and absent under `status: unsatisfiable`**, which is the one thing `status` decides and the reason it is a field rather than a formality. A lane that could not be satisfied has no judgement to record, so it records none — it does not record a fourth kind of judgement. The shape is fixed in `§ Insufficient evidence is not low confidence` below.

**`inspected: diff: false` makes the verdict automatically invalid** and the judge ignores it — unless the return carries `status: unsatisfiable` in the form `§ Insufficient evidence is not low confidence` below fixes, which is the one shape that reports `false` honestly. That flag is the floor; the standard is the `evidence` block, defined in `§ The report block`.

Return decision-relevant data only. Deep internal analysis is welcome; only its conclusions enter the parent's context.

**Every finding a reviewer files will be resolved by the judge into exactly one of five outcomes** — `required_change` · `accepted_risk` · `invalid_finding` · `future_work` · `informational`. A reviewer does not resolve its own findings, but knowing the vocabulary changes how it writes them: a finding filed as though everything must block is a finding the judge has to reclassify, and one filed as a passing remark is one that disappears. The five are defined in `docs/DELIVERY.md § Findings and their owners`.

### Insufficient evidence is not low confidence

**A lane that could not gather evidence and a lane that gathered weak evidence are different results, and one field cannot carry both.** `status: unsatisfiable`, with no `verdict` at all, says the review could not be performed. A low `confidence` says it was performed and the reviewer does not trust the answer. Collapsing the first into the second hands the judge a number where there was no measurement — and the judge then weighs an unrun lane against a run one, which is #144's ledger-corruption class arriving through vocabulary instead of through a missing row.

**It is `status` that carries this and never `verdict`, and the two are different kinds of statement.** `pass` · `revise` · `blocker` are judgements about the code; *I could not measure* is a statement about the lane. **A fourth verdict value was tried and withdrawn** — the Owner ruled on 6 September 2026, on #158, that `insufficient_evidence` is not a verdict, that an unassessable lane is expressed through `status:` with `verdict:` absent, and that the capability motivating the fourth value is preserved by the shape below rather than lost. The enum kept in `docs/DELIVERY.md § Verdicts` was left untouched by that ruling, and the rules that invalidate a verdict reach this shape no differently for it: **a lane that produced no verdict is not a weak verdict, and there is nothing here to invalidate.**

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
# no verdict key — the absence is the statement, per the paragraph above
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

**`confidence: unassessed` is mandatory with this status and a number is forbidden.** Any number offered here is manufactured, and manufacturing one is the failure a reviewer exists to catch in others.

**`evidence_gap` is mandatory too, and `closable_by` is the load-bearing field.** It is what separates a gap a dispatch can close — send the artifact inline — from one it cannot: **a follow-up question cannot close a tooling gap**, and a judge that does not know which it is facing will spend a cycle asking.

**The judge records an unsatisfiable lane as unsatisfiable and converts it into neither a pass nor a fail.** Ignoring it, as `§ Reviewer verdict` above has the judge ignore an ordinary `diff: false`, is how an absent measurement becomes a silently passed lane; reading `unassessed` as a low number is how it becomes a weak one. Both are the same error in opposite directions, and the ledger row carries the word rather than a value.

**The class was first recorded on PR #135, 29 August 2026.** `@confidence-assessor` holds `Read, Grep, Glob` and no Bash or GitHub access, so it could not read the verdicts it had been convened to weigh; it checked for a local mirror before reporting the gap, then returned `evidence_quality: unknown` — *"not weak — unassessed"* — with `followup: none`, *"a reviewer follow-up doesn't fix a tooling gap."* The cycle-3 ruling recorded the lane **unsatisfiable, not low**. That vocabulary is this section, and the payload half of it is `handoff-payloads § Confidence assessment`.

## Record your verdict into its row before your pass ends

**Return the verdict above to `@pr-judge`, and record it into the claim table first.** The return value is the convenience; the row is the record. A pass that ends holding a verdict only in its own context has produced nothing a dead parent can collect, which is the stranding class issue #144 records — and it is why the row comes first rather than after.

```bash
scripts/loop/claim.sh verdict <pr> <sha> <lane> <ruling> \
  --conf <x> --findings <n> --artifact <where the full verdict is>
```

`<ruling>` is the verdict word from `§ Reviewer verdict` above; the table enforces no vocabulary of its own, so a word this skill does not define is a word nothing will refuse. **The dispatch carries the other three arguments** — PR number, head SHA, and your lane name — per `review-board-dispatch § The case file (same for every reviewer) — references, not content`. A dispatch that does not carry them convened you outside the table: record what you can, and say so in your verdict rather than guessing a panel key.

**This obligation lives here and in no agent file.** Every convened reviewer already loads this skill, and stating it in each reviewer definition instead would create two dozen copies to keep true; a reader looks in one place. What a reviewer supplies of its own is its lane name and its ruling.

**Branch on the exit status; never parse the prose.** The full set and what each code means are in the header of `scripts/loop/claim.sh`. These are the ones a reviewer meets, and what each one obliges:

| Status | What you do |
|---|---|
| **0** recorded | Nothing further. The ruling is of record and survives your process. |
| **10** already ruled | **Stop. Do not retry and do not overwrite.** A ruling for your lane at this SHA already stands and stays of record; report the collision in your envelope. |
| **12** recorded, unclaimed | The ruling is durable and no claim row covered your lane — you were dispatched without a claim, which is a defect in the dispatch. Name it in your envelope. |
| **2** NOT recorded | The table did not take it. **Carry the whole verdict in your handoff** and say plainly that the table does not hold it. |

**Record before you report, not after.** The order is the entire mechanism: a row written before the pass ends survives a parent that never reads the return value, and `agent-handoffs § An outstanding continuation is not left behind` is the general form of the same obligation.

**Recording is not ruling a second time.** The row carries the ruling, the confidence, the finding count, and a reference to where the full verdict lives. The findings themselves stay in the verdict you return — the table holds no analysis and adjudicates nothing.

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

