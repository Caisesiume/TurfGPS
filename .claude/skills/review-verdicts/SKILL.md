---
name: review-verdicts
description: What a convened TurfGPS reviewer returns, and the standard that verdict is measured against — the reviewer verdict schema with its findings, severity, confidence and residual risk, and the evidence law: a reviewer does not accept a claim it could check, the VERIFIED INDEPENDENTLY / ACCEPTED ON TRUST block, how far the obligation reaches, and the two incidents that made it a rule rather than a habit; and the obligation to record that verdict into its own claim-table row before the pass ends. Load alongside `agent-handoffs` before returning any review verdict.
---

# Review verdicts — the schema and the evidence law

*Split out of `agent-handoffs` on 28 August 2026, which keeps the envelope, the size limit, and the output caps and is not restated here. This file is what a convened reviewer returns and the standard it is held to; a reviewer loads both, and nothing else.*

## Reviewer verdict

Returned by every convened reviewer to `@pr-judge`.

```yaml
reviewer: security
status: valid_review
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
confidence: 0.96
residual_risk:
needs_followup: false
evidence: |
  VERIFIED INDEPENDENTLY:
    · …
  ACCEPTED ON TRUST:
    · …
```

**`inspected: diff: false` makes the verdict automatically invalid** and the judge ignores it. That flag is the floor; the standard is the `evidence` block, defined in the next section.

Return decision-relevant data only. Deep internal analysis is welcome; only its conclusions enter the parent's context.

**Every finding a reviewer files will be resolved by the judge into exactly one of five outcomes** — `required_change` · `accepted_risk` · `invalid_finding` · `future_work` · `informational`. A reviewer does not resolve its own findings, but knowing the vocabulary changes how it writes them: a finding filed as though everything must block is a finding the judge has to reclassify, and one filed as a passing remark is one that disappears. The five are defined in `docs/DELIVERY.md § Findings and their owners`.

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

