---
name: agent-handoffs
description: The shared envelope and schemas for every handoff between TurfGPS agents — the input-references/execution/verdict principle, the size limit, and the five payload schemas (worker completion, reviewer verdict, revision packet, escalation packet, risk assessment). Every agent that dispatches another agent or reports to one loads this before it writes a handoff.
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

### Worker completion (§8)

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

### Reviewer verdict (§13)

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

**`inspected: diff: false` makes the verdict automatically invalid** and the judge ignores it. That flag is the floor. The standard is the `evidence` block, whose two halves are defined in `review-board-dispatch § A reviewer does not accept a claim it could check` — and the load-bearing half is `ACCEPTED ON TRUST`, because naming what you took on faith is the only part of a review that makes you notice you took something on faith. An empty `VERIFIED INDEPENDENTLY` is itself a finding: the reviewer read the PR body, not the work.

Return decision-relevant data only. Deep internal analysis is welcome; only its conclusions enter the parent's context.

### Revision packet (§17)

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

### Escalation packet (§21)

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

### Risk assessment (§6)

Produced by `@change-risk-assessor`, consumed by `@worker-manager` at intake and `@pr-judge` at PR open. Structured data only, no prose. Its shape is in that agent's definition.

## Before you invoke anything (§30)

Three questions, every time:

1. **Does this agent have a reasonable chance of changing the outcome?** If no, do not invoke it. "It might as well look" is the habit this whole model was written to remove.
2. **Has the evidence relevant to this agent materially changed?** If no, preserve its previous verdict rather than re-running it.
3. **Can the receiving agent retrieve this from an authoritative artifact?** If yes, send the reference instead of the content.
