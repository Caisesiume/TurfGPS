---
name: agent-handoffs
description: The shared envelope and schemas for every handoff between TurfGPS agents — the input-references/execution/verdict principle, the size limit and its per-field caps, the five payload schemas (worker completion, reviewer verdict, revision packet, escalation packet, risk assessment), the evidence law a reviewer's verdict must satisfy, the context-escalation ladder, and tool-output discipline. Every agent that dispatches another agent, reports to one, or returns a verdict loads this before it writes a handoff.
---

# Agent handoffs — the envelope and its payloads

Ratified in `docs/adr/ADR-0001-artifact-driven-agent-org.md § D8`, from the Owner directive §23–25.

## The principle

**INPUT REFERENCES → EXECUTION → STRUCTURED VERDICT.**

Agents operate like functions with explicit contracts, not like colleagues in a conversation. A handoff is the return value, not a report of the meeting.

What follows from that, and is the whole point of this skill: **the receiving agent retrieves what it needs itself.** The board item, the diff, the requirement record, the architecture section — all of them are authoritative artifacts that the receiver can open. Copying them into a handoff produces a second, staler copy and pays for it twice, once in the sender's output and once in the receiver's input.

## The limit

**A handoff is typically under ~300 tokens.** It carries references — issue, PR, finding IDs, file paths, requirement codes, section names — not content.

Never send: a subagent transcript · chain-of-thought or internal reasoning · a chronological account of how the work went · a complete diff, requirement, or PR description · repository file contents.

**A chronology is the most common violation and the easiest to spot**: if the handoff would read differently had the work been done in a different order, it is describing the work rather than its result. The downstream agent acts on the result.

Exceeding the limit is allowed when the content genuinely cannot be referenced — an escalation's option analysis, a conflict between two reviewers stated in their own words. Exceeding it because summarizing was harder than pasting is not.

### Per-field caps

The ~300 tokens are spent field by field, and a limit with no per-field budget is spent entirely by whichever field the writer found most interesting. **Verbosity here is a contract violation, not a matter of style** — the receiver pays for every word in its own context window, so an overlong field is a cost imposed on someone who cannot decline it.

| Field | Cap |
|---|---|
| `summary` | **2 sentences maximum**, of outcome |
| each finding | **1 concise description + 1 required action** — not a paragraph of analysis |
| `decisions` | **IDs and the outcome only**, unless the reasoning is genuinely unretrievable elsewhere |
| `recommended_next_action` | **one** action |

Where the reasoning matters and is retrievable, cite where it lives — `DECISIONS.md § RD-007`, the ADR, the PR comment. Where it matters and lives nowhere yet, that is a signal to write it into an artifact, not to widen the envelope.

## The envelope

All handoffs use this shape. Extend it where an agent needs to; **do not populate irrelevant fields** — an empty field is noise that the next reader must still check.

```yaml
task_id:
agent:
status: completed | blocked | failed | decision_required
summary:
artifacts:
  issues: []
  pull_requests: []
  commits: []
  files: []
findings: []
decisions: []
risk:
confidence:
recommended_next_action:
required_agents: []
human_escalation: false
```

`summary` is one or two sentences of *outcome*. `confidence` is a number you would defend, not a courtesy 0.9.

## Payload schemas

### Worker completion

Returned by any implementation specialist to `@worker-manager`.

```yaml
status: completed
issue: ENG-142
changes:
  - added token refresh handling
  - added expired-session recovery
  - added regression tests
files_changed:
  - src/auth/session.ts
  - tests/auth/session.test.ts
tests:
  status: passed
  commands:
    - npm test -- session.test.ts
risks:
  - none_known
requires_review:
  - security
  - correctness
  - testing
confidence: 0.93
```

`requires_review` is a **hint from the person who wrote the code** about where it is weakest. It informs selection; it does not decide it — the registry and the risk assessment do, because an author's sense of where their own work is weak is exactly the thing under review.

Where an acceptance criterion is `test`-verified, the completion also carries the red demonstration required by `docs/DELIVERY.md § Proof that a test can fail` — the assertion's own failure message, per criterion, or the story that owes it.

### Reviewer verdict

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

### Revision packet

Produced by `@pr-judge` on remand, consumed by `@worker-manager`.

```yaml
revision:
  issue: PR-381
  cycle: 2
required_changes:
  - finding: SEC-01
    owner: backend
    scope: src/auth/session.ts
    change: invalidate old refresh token after successful rotation
accepted_risks:
  - finding: PERF-03
    owner: pr-judge
    reason: cost bounded by candidate cap; revisit if the cap moves
review_after_revision:
  required:
    - security
    - correctness
  not_required:
    - architecture
    - accessibility
    - documentation
```

`review_after_revision.not_required` is not decoration — it is the instruction that stops a one-line fix from re-running the bench, and it is checkable against the ledger afterwards.

### Escalation packet

The only shape in which anything reaches the human, and only `@engineering-lead` sends it.

```yaml
human_decision_required: true
question: <one precise question>
reason: <why existing artifacts cannot answer it>
options:
  - ...
recommended_option: ...
impact:
  ...
```

**Never ask "What should I do?"** An escalation without a recommendation is work handed back, and the Owner has said so directly. The qualifying conditions are in `docs/DELIVERY.md § Escalation and human judgement`; nothing else qualifies.

### Risk assessment

Produced by `@change-risk-assessor`, consumed by `@worker-manager` at intake and `@pr-judge` at PR open. Structured data only, no prose. Its shape is in that agent's definition.

`review_not_required` in that assessment is a **hard negative**, not a hint — see `review-board-dispatch § Negative routing`.

## Structured uncertainty (blocked)

*§48.* An agent that needs another domain's judgement does not open a conversation with that domain. It **stops and returns**:

```yaml
status: blocked
needs_domain_decision:
  domain: architecture
  question: <one precise question>
  evidence: [Architecture.md § Retrieving zones, FR-24, PR 381]
  recommendation: <what you would do, and why>
```

The orchestrator that dispatched you routes **one** targeted request, and the answer becomes an artifact or a recorded decision — an ADR, a `DECISIONS.md` entry, an amended requirement — so the next agent retrieves it rather than asking again. **Agents never chat**: a back-and-forth costs a full execution per turn and leaves nothing behind that anyone can retrieve.

## A reviewer does not accept a claim it could check

*Ratified in `docs/adr/ADR-0001-artifact-driven-agent-org.md § D5`; moved here by ADR-0002 § O1. This is the home of the evidence law. It lives beside the verdict schema that carries it because every reviewer already loads this skill, and none of them should have to load the judge's dispatch mechanics to find the standard their own verdict is measured against.*

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
- **The gates that could pass having read nothing.** `Architecture.md § D8` puts the Go module in `service/`. The gate block carried no working directory, so from the repository root all five commands resolve against nothing, exit zero, and print exactly what a clean module prints — and the prescribed report line was character-for-character what that vacuous run produces. Eleven agent files and the PR-body template carried the same directory-less copy, so the path ran unbroken from command to report line. Closed before any PR in this repository existed to carry it; the instrument, not a reviewer, was the thing that would have lied. Found 5 August 2026 because a layout decision recorded its own cost honestly — `d6a7e3e`, `1928a28`.

Neither is something a reviewer catches by reading attentively. Both were **instruments reporting success**, and the only thing that separated the report from the truth was an agent running the thing itself.

## The context escalation ladder

Artifact-driven retrieval still wastes a window if every agent opens the whole corpus for a narrow task. **Retrieve progressively, not comprehensively.** Four levels, and each is *earned* by evidence that the level below it was insufficient:

| Level | What you hold | Typical use |
|---|---|---|
| **0** | IDs and metadata — issue, PR, requirement code, SHA, status | routing and assignment |
| **1** | The named issue, requirement record, or diff section | ordinary specialist work |
| **2** | The related architecture and design sections | cross-cutting reasoning |
| **3** | The wider project corpus | genuine systemic ambiguity, and nothing less |

**Never start at Level 3.** Opening everything is not thoroughness — it is the decision not to work out what the task needs, paid for by the context window that has to survive the rest of the session. When you climb a level, you should be able to say what the lower level failed to answer; if you cannot, you have not earned the climb.

The dispatch tells you where to start: `@worker-manager` and `@pr-judge` send scoped references — requirement IDs, `Architecture.md § <section>`, `DESIGN.md § <section>`, the scope. **Read those first.** Broaden only when the evidence in front of you demands it, and on a contradiction between what you were sent and what you found, escalate rather than loading the universe to adjudicate it yourself.

## Tool-output discipline

A small handoff does not save anything if the work that produced it dragged a full CI log into context. **Tool output obeys the same progressive rule as artifacts.**

**A successful command returns a one-line confirmation.** `tests: passed · suite: service/auth` is the whole report. Nobody reads a green log, and pasting one costs exactly as much as pasting a red one.

**A failure returns, in this order:** the failed command · the exit status · the failing assertion or error · the relevant excerpt. Expand beyond that only when diagnosis actually requires it — and then only the part that does.

Order every retrieval cheapest-first: **filenames before patches · failed tests before logs · one requirement before a directory · a scoped query before a full board dump.** Each of those pairs has a left side that usually ends the question.

## Before you invoke anything

Four questions, every time:

1. **Does this agent have a reasonable chance of changing the outcome?** If no, do not invoke it. "It might as well look" is the habit this whole model was written to remove.
2. **Has the evidence relevant to this agent materially changed?** If no, preserve its previous verdict rather than re-running it.
3. **Can the receiving agent retrieve this from an authoritative artifact?** If yes, send the reference instead of the content.
4. **Has this already been decided?** Search `docs/Requirements/DECISIONS.md`, the ADRs in `docs/adr/`, the requirement record itself, and the board or PR record. If it has: **reuse the decision, do not re-litigate it.** Re-deciding a settled question is the most expensive kind of work, because it costs a full execution and its usual output is the answer that already existed — or, worse, a different one.
