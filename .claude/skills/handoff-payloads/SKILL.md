---
name: handoff-payloads
description: The role-specific payloads that fill the TurfGPS handoff envelope — worker completion with its red demonstration, the revision packet a remand produces, the escalation packet that is the only shape reaching the human, the risk assessment, the confidence assessment and its unassessed case, structured uncertainty when another domain must decide, and the dependency finding and graph update. Load alongside `agent-handoffs`, which holds the envelope itself, its limit and the output caps; the reviewer verdict and the evidence law are in `review-verdicts`.
---

# Handoff payloads — the schemas that fill the envelope

*Split out of `agent-handoffs` on 28 August 2026. That skill holds the envelope every payload extends, the size limit, and the output caps. **The reviewer verdict is not one of these payloads** — it, and the evidence law it is measured against, are in `review-verdicts`. None of either is restated here. Load the payload your role actually returns; an agent that reports work does not need the graph schemas, and an agent that repairs the graph does not need the escalation packet.*

**Every payload here opens with its structured block, and licensed prose — where there is any — comes after it**, per `agent-handoffs § The structured block comes first`. A payload whose artifact holds a row in `agent-handoffs § The cap table` declares `artifact:` and `prose_licence:` as its first two keys; the shape rule and both key definitions live there and are not restated here.

## Worker completion

Returned by any implementation specialist to `@worker-manager`.

```yaml
artifact: worker_envelope
prose_licence: none
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

**Mandatory keys:** `artifact` · `prose_licence` · `status` · `issue` · `changes` · `files_changed` · `tests` · `risks` · `requires_review` · `confidence`. `risks: [none_known]` is an answer; an absent `risks` is not.

`requires_review` is a **hint from the person who wrote the code** about where it is weakest. It informs selection; it does not decide it — the registry and the risk assessment do, because an author's sense of where their own work is weak is exactly the thing under review.

Where an acceptance criterion is `test`-verified, the completion also carries the red demonstration required by `docs/DELIVERY.md § Proof that a test can fail` — the assertion's own failure message, per criterion, or the story that owes it.

## Revision packet

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

## Escalation packet

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

## Risk assessment

Produced by `@change-risk-assessor`, consumed by `@worker-manager` at intake and `@pr-judge` at PR open. Structured data only, no prose. Its shape is in that agent's definition.

`review_not_required` in that assessment is a **hard negative**, not a hint — see `review-board-dispatch § Negative routing`.

## Confidence assessment

Produced by `@confidence-assessor`, consumed by `@pr-judge`. It weighs the verdicts and never the change; the lane's own limits are in that agent's definition and are not restated here.

```yaml
artifact: worker_envelope
prose_licence: none
agent: confidence-assessor
aggregate_confidence: 0.71        # a number, or `unassessed`
evidence_quality: adequate        # strong | adequate | weak | unknown
conflicts:
  - between: [linus-security-critic, go-quality-critic]
    about: SEC-02 — whether the input is already validated upstream
coverage_gaps:
  - testing lane required by risk assessment, no verdict present
followup:
  reviewer: linus-security-critic
  question: does the upstream validation the pass relies on exist in this diff, or was it assumed?
recommendation: targeted_followup # decide_now | targeted_followup | insufficient_evidence
```

**Mandatory keys:** `artifact` · `prose_licence` · `agent` · `aggregate_confidence` · `evidence_quality` · `conflicts` · `coverage_gaps` · `followup` · `recommendation`. Where the evidence is sufficient, `followup: none` and `recommendation: decide_now` — **said plainly, as a row.** A meta-reviewer that always finds something to check is as useless as one that never does, and it is more expensive.

### Unknown is not weak

**`evidence_quality: unknown` means unassessed** — the lane could not reach the verdicts at all. `weak` means it read them and they are thin. The law separating the two, and the incident that made it a rule, are in `review-verdicts § Insufficient evidence is not low confidence` and are not restated here. What this payload adds is the shape `unknown` obliges:

```yaml
evidence_quality: unknown
aggregate_confidence: unassessed
recommendation: insufficient_evidence
evidence_gap:
  what: the collected verdicts
  why: tooling                    # tooling | access | artifact_absent | out_of_scope
  closable_by: a dispatch carrying the verdict bodies inline
followup: none
```

**`unknown`, `unassessed`, `insufficient_evidence` and `evidence_gap` move together, and no three of them are a valid payload.** `unknown` beside a numeric `aggregate_confidence` is the conflation this vocabulary exists to prevent, wearing the form that licenses it; `unknown` without `evidence_gap` tells the judge a lane failed without telling it what would make the lane run.

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

## Dependency findings and graph updates

*`ADR-0003 § P7`. The persisted grammar both schemas refer to is `turfgps-board-ops § The dependency representation`, and is not restated here.*

### Dependency finding

Returned by `@scrum-master`, an implementation specialist, `@worker-manager`, or `@pr-judge` when the graph looks wrong from where they stand. It **always ends at `@backlog-dependency-planner` — and always reaches it via the orchestration path**: the reporter puts the finding in its envelope, `@engineering-lead` dispatches the planner. Dispatch authority over the planner is exactly two agents (`ADR-0003 § P9`, amended): `@engineering-lead` for every graph event, and `@requirements-engineer` for its own story batches, which it continues directly because relaying a mandatory pipeline step through an orchestrator adds a hop with no decision in it. Reporting a finding is not dispatching, and **the reporter edits no edge**: a graph repaired in passing by four agents is a graph with no owner, and the repair nobody reviewed is the one that survives.

```yaml
dependency_finding:
  reporter: scrum-master
  story: 46
  suspected_prerequisite: 43          # or missing_prerequisite / invalid_edge: <n>
  evidence: ["#46 AC-2 reads the persisted classification", "#43 is what creates it"]
  recommendation: verify and persist a hard edge — 46 blocked by 43
```

### Graph update

Returned by `@backlog-dependency-planner`. Edges, not prose: no story text, no requirement text, no account of the pass.

```yaml
artifact: graph_update
prose_licence: none
graph_update:
  stories_examined: [41, 43, 46, 47]
  added:   [{blocked: 46, prerequisite: 43, type: hard, reason: "consumes the persisted classification"}]
  removed: [{blocked: 47, prerequisite: 52, reason: "scope moved out of #52; the relationship no longer holds"}]
  preserved: 9                        # a count — edges outside the pass are untouched, not re-listed
  newly_unblocked: [43]               # derived from prerequisite state — a satisfied edge is never `removed`
  newly_blocked: []
  parallelizable: [[43, 52]]
  affected_epics: ["Access classification"]
```

`@requirements-story-organizer` may extend its envelope with **`dependency_hints`** — `{downstream, upstream, reason}`, references only: hints for the planner, never authority over the graph.

