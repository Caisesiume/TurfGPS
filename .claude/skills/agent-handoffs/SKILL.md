---
name: agent-handoffs
description: The shared envelope for every handoff between TurfGPS agents — the input-references/execution/verdict principle, the size limit with its per-field caps, the output caps every capped artifact is held to, the trigger block, the obligation not to end a pass with a continuation outstanding, the context-escalation ladder, tool-output discipline, and the four questions asked before any invocation. The role-specific payload schemas are in `handoff-payloads`; the reviewer verdict and the evidence law are in `review-verdicts`. Every agent that dispatches another agent or reports to one loads this before it writes a handoff.
---

# Agent handoffs — the envelope

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

### Output caps

**Every capped artifact is capped here, and this section is its one home.** A contract that produces one of these cites this table rather than restating a number. Characters rather than tokens, because a writer can count characters.

| Artifact | Cap |
|---|---|
| reviewer verdict | **≤ 1,200 chars** plus its findings table; the evidence block **≤ 10 bullets, one line each** |
| judgment | **≤ 6,000 chars** |
| review ledger | **≤ 2,000 chars**, and it **supersedes** — one table per PR, rewritten each cycle, never appended |
| worker envelope | **≤ 1,500 chars** |
| `graph_update` | **≤ 1,200 chars** |
| an `@engineering-lead` dispatch | **≤ 1,200 chars** |
| an Owner report | **≤ 2,000 chars** |

**Prose inside a capped artifact is licensed for exactly four things:** a finding **overturned** · a conflict **dissolved** · a rule **renegotiated** · a predecessor **corrected**. In each of those the reasoning *is* the decision and cannot be recovered from the outcome. **A finding that simply holds gets a row, not a paragraph** — the row states it, the artifact the row names carries the rest, and a reader wanting the argument opens what it cites.

**The caps are what give the rule above teeth.** *Verbosity is a contract violation, not a style preference* named no number a writer could fail, and a rule nobody can fail is a preference. The cost being cut is not the artifact itself — issue #128 measured 35 judgment comments at about 120k tokens total, roughly 2% of the runs that produced them. It is the **re-read multiplier**: each cycle ingests its predecessors, so on PR #67 the fifth cycle could take in ~25k tokens of earlier prose before reading a line of the diff. A cap is paid once where it is written and refunded on every later pass.

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

### The trigger block

A dispatch made because something *moved* carries why, so the receiver never has to re-detect it:

```yaml
trigger:
  type: merge_completed          # board_changed | pr_changed | corpus_changed | merge_completed | explicit_request
  source: engineering-lead
  fingerprint_component: main    # or `explicit_sync: true` when a human or an agent asked directly
```

**An explicitly dispatched agent knows why it was invoked and does not re-poll to second-guess its parent.** The fingerprint's state is per-consumer and already consumed by the dispatcher; re-running the generic gate would let the very signal that caused the dispatch reject it. Self-gating on `scripts/loop/fingerprint.sh <your-own-agent-name>` is for waking *autonomously* — cron, `/loop`, a scheduled run — and the argument is never omitted, because the bare default puts every caller on one shared state file.

## Where the payloads live

This skill is the envelope. The schemas that fill it were split out on 28 August 2026, because it is loaded by nearly every agent in the organization and each of them was carrying every other role's payload to reach its own:

- **`review-verdicts`** — the reviewer verdict schema, and the evidence law that verdict is measured against. Every convened reviewer loads it.
- **`handoff-payloads`** — worker completion, the revision packet, the escalation packet, the risk assessment, structured uncertainty, and the dependency finding and graph update.

Load the one your role returns, and cite it rather than restating it; neither is summarized here.

## An outstanding continuation is not left behind

*Ratified in `docs/adr/ADR-0001-artifact-driven-agent-org.md § D11`, on two observations in the first live loop cycle. Binds every agent that dispatches; the dispatchers cite this section rather than restating it.*

**An agent must not end a pass while a continuation it owns is outstanding.** Dispatching a child creates an obligation that outlives the dispatch, and a parent whose process ends before its children return takes the obligation with it: the mandatory next step never fires, and finished work is left with nobody holding it.

A dispatching agent therefore has exactly two honest endings:

- **Await the children** and continue the pass, or
- **Persist their output to a durable artifact** — the issue, the PR, the corpus file — and name in its envelope **what remains owed and to whom**.

```yaml
continuation_owed:
  to: backlog-dependency-planner
  for: "the story batch just filed"
  persisted: "issue #18 comment — the finished FR-019 field block"
```

Ending with the work held only in the pass's own context is neither. The second ending is not the lesser one: an agent that cannot await its children still discharges the obligation by making the state retrievable and the debt explicit — the same artifact-over-conversation principle this skill applies to every other handoff. **A pass that ends silently with a continuation outstanding is a defect in that pass**, not an accident of scheduling, and it is invisible by construction: nothing anywhere records that a step was owed.

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
