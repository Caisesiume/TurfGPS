# ADR-0003 — A dedicated backlog dependency planner

**Status:** accepted — 2026-08-10
**Source:** `docs/adr/agent-org-directive-3.md`, the Owner's third directive, kept verbatim. Where that file and this one differ on a repository-specific adaptation, **this record is the ratified form**. Section references (`§N`) point at that directive.
**Relation to ADR-0001 / ADR-0002:** a targeted change **on** that architecture, not a replacement for it. No specialist is deleted, no safety floor weakened, no settled decision reopened. Every ADR-0002 optimization listed in §30 survives — fingerprint gating, scoped retrieval, references over context, compact envelopes, selective review invalidation, deterministic checks before LLM calls.

## Context

`@scrum-master` owned two different jobs: **board truth and readiness**, and **semantic dependency reasoning**. The second one ran on every sync. A board change woke it, it read the backlog, re-read the architecture, and re-derived an implementation ordering — data plane before consumers, ports before adapters, schema before code, backend before frontend — that had not changed since the previous run and would not change again until the work graph did. Dependency reasoning is *durable*: it belongs to the moment the graph changes, not to the moment somebody looks at the board.

### Pre-edit inspection — what already existed

Per the directive's final instruction, inspected before any edit; recorded so a later reader does not mistake an absence below for an oversight.

| Inspected | Found, 2026-08-10 |
|---|---|
| Who wrote dependencies | `@requirements-story-organizer`, at story creation — `Blocked by: #41` plus per-edge provenance prose in a `## Dependencies` body section |
| Who re-derived them | `@scrum-master`, every sync, in its Phase 3 architectural-ordering block |
| How many live sections | **59** open issues carry `## Dependencies`; **53** carry at least one `Blocked by:` line; **0** carry a soft edge, because no syntax for one existed |
| Native GitHub dependency API | Live and readable — `gh api repos/Caisesiume/TurfGPS/issues/43/dependencies/blocked_by` → `200`, `[]` |
| Who consumed the ordering | `@project-coordinator`, described as consuming "@scrum-master's dependency analysis" |

So the graph already existed and was already persisted in GitHub. What was missing was an **owner** for it, a **type** on its edges, and the rule that it is *read* rather than *recomputed*.

## Decision

### P1 — `@backlog-dependency-planner` exists

*Context: §1–§5, §15–§19. The durable reasoning needed an owner, and it was not going to be the agent that runs every sync.*

New agent `.claude/agents/backlog-dependency-planner.md` (opus; Read, Grep, Glob, Bash, Skill, `mcp__github`). It answers **what must be true before this work can safely begin?** and owns edge existence, edge type, edge provenance, conflict classification, and selective re-evaluation. It does **not** own requirements truth, story content, board Status, worker assignment, implementation, PR review, priority itself, or runtime scheduling. Its duties: minimum necessary ordering, maximum safe parallelism, and a concrete one-line reason on every edge — **relatedness is not a dependency**.

### P2 — The `## Dependencies` body section is the authoritative representation

*Context: §6, §16. The directive's preference order puts native relationships first; the inspection found the third option already deployed at scale, with two properties the first does not have.*

The format's one home is **`turfgps-board-ops § The dependency representation`**: the grammar (`Blocked by: #N — reason` hard, `Soft dependency: #N — reason` soft, optional one-line `Basis:`), the hard/soft semantics, who writes and who reads, and the grandfather clause. Everything else cites it.

**The grandfather clause is deliberate.** The 59 existing sections are valid hard edges exactly as written; the planner brings a subgraph up to the grammar the first time an event touches it. A migration pass over 59 issue bodies would touch every story on the board at once to change nothing any agent reads differently.

**Native-API migration is deferred, not rejected** — and this record owns that decision. The API is readable today. Against it: the body convention already exists at scale, it carries **provenance** and the **soft type** natively where the relation has no field for either, and it is greppable without an API call, which is what makes `dependents.sh` possible. If the native relationship gains a type and a reason, migrating is a mechanical pass and this section is where the option was recorded.

### P3 — The organizer emits hints, not edges

*Context: §7, §8. A hint written as an edge is indistinguishable from a verified one the moment it is on the board.*

`@requirements-story-organizer` stays Story Architect and **stops writing `Blocked by:` lines**. New stories carry the `## Dependencies` heading with `_Pending @backlog-dependency-planner._`. What it noticed while cutting goes into `dependency_hints` in its handoff. Its completion is the planner's trigger; the dispatch itself is routed by whoever commissioned it — normally `@requirements-engineer` through `@engineering-lead` — because the organizer holds no Agent tool and agents do not dispatch sideways.

### P4 — The scrum-master evaluates readiness against the persisted graph

*Context: §9–§11, §27. This is where the saving is: the second path must be substantially cheaper than the first.*

Phase 3's architectural-ordering block is **deleted**. In its place: verify traceability · read the persisted hard edges · confirm every hard blocker is Done and merged · confirm no explicit blocking state remains · apply the Priority/WIP policy · promote. **It never silently repairs the graph**: a story whose ACs plainly consume another with no persisted edge, or an edge naming a nonexistent issue, returns a `dependency_finding` to the planner. Everything ADR-0002 gave it survives — fingerprint gate, scoped retrieval, priority-first promotion, WIP, traceability flagging, statelessness — and the file is shorter than before.

### P5 — The coordinator stays runtime-only

*Context: §12. Smallest edit in the wave, and the one that keeps the rebuilt-graph problem from reappearing one lane down.*

It consumes the Ready queue in the order `@scrum-master` gave it from the persisted graph, and **never inspects the backlog to reconstruct dependencies**.

### P6 — `scripts/loop/dependents.sh`

*Context: §22, §41. When an upstream story completes, no LLM reasons from scratch about every blocked story.*

Argument: the merged or closed issue number. It reads open issues once, parses hard blockers out of each `## Dependencies` section, checks each blocker's state, and prints `eligible:` and `still_blocked:`. Soft edges are deliberately not read — counting one would manufacture a blocker. A blocker whose state cannot be read **counts as still blocking**, the same rule `fingerprint.sh` applies to an unreadable component. Output is capped; `none` and exit 0 when nothing depended on the issue.

### P7 — Two schemas in `agent-handoffs`

**Both now live in `handoff-payloads`**, split out of `agent-handoffs` on 28 August 2026; the heading records where P7 put them and the citation below is the current home.

*Context: §11, §18, §20. `handoff-payloads § Dependency findings and graph updates`.*

**`dependency_finding`** — reporter, story, suspected or missing prerequisite, evidence references, recommendation. Returned by `@scrum-master`, specialists, `@worker-manager`, `@pr-judge`; **always** routed to the planner; reporters never edit edges. **`graph_update`** — stories examined, edges added/removed/preserved with type and reason, newly unblocked, newly blocked, parallelizable sets, affected Epics. The organizer's envelope may extend with `dependency_hints`.

### P8 — `dependency` and `planning` join the root-cause vocabulary

*Context: §21. Do not repeatedly patch around an invalid backlog graph.*

In `pr-judge.md § Phase 8` and `DELIVERY.md § Root cause`. Work that failed because it ran in the wrong structural order — a consumer implemented before its contract — routes to the planner, not to another revision cycle.

### P9 — Implementation feedback, and who may dispatch the planner

*Context: §13, §14, §20, §30.*

`@worker-manager` returns a `dependency_finding` for a prerequisite discovered mid-implementation and edits no edge. `@engineering-lead` consumes graph health rather than deriving order, and dispatches the planner **only on a graph event** — story batches created or changed, scope changes, an architecture decision moving a boundary, hints, or a finding arriving. **Never on cadence, a poll, or a fingerprint change**: a fingerprint detects that something moved, and treating it as a graph event would restore precisely the per-sync recomputation this record removes.

### P10 — Governance

*Context: §24, §29. One authoritative home per rule; do not copy the policy into every agent.*

`DELIVERY.md § Who owns what` carries the ownership table — requirements truth · nodes · edges · readiness · runtime · implementation routing · review — and `DELIVERY.md § Root cause` carries the new routing line. `docs/adr/README.md` lists the directive and this record.

## §31 acceptance criteria — where each is satisfied

| # | Criterion | Satisfied by |
|---|---|---|
| 1 | A dedicated owner for the persistent graph | P1 |
| 2 | Organizer creates work, emits hints, is not the authority | P3 |
| 3 | Scrum-master no longer reconstructs architectural order | P4 — the Phase 3 block is deleted |
| 4 | Scrum-master consumes persisted hard blockers | P4 |
| 5 | Coordinator remains runtime-only | P5 |
| 6 | Lead consumes graph state | P9 |
| 7 | Analysis runs only when graph-relevant artifacts change | P1 activation list · P9 dispatch rule |
| 8 | One story change invalidates only its subgraph | P1 — selective invalidation, on the intersection-test philosophy |
| 9 | Hard and soft are distinguished | P2 grammar |
| 10 | Safe parallelism explicitly preserved | P1 duties · `graph_update.parallelizable` |
| 11 | Provenance persisted compactly | P2 `Basis:` line |
| 12 | Hidden dependencies from implementation route to the planner | P7 · P9 |
| 13 | Planning defects from review route to the planner | P8 |
| 14 | Unrelated board changes cause no recomputation | P9 — no fingerprint row wakes the planner |
| 15 | Newly satisfied dependencies detected cheaply after merges | P6 |
| 16 | GitHub remains the source of dependency truth | P2 — the edges live in issue bodies |
| 17 | The pipeline progresses without lead or scrum-master rebuilding the graph | P3 → P1 → P4 → P5 flow |

## Consequences

**What this obliges:**

- **One more agent in the chain between stories and Ready.** A story batch is not schedulable until the planner has run over it. That gap is visible by design — the placeholder says so on the issue — but a batch filed while nobody dispatches the planner sits with no edges at all, and the scrum-master will read that as *unblocked* and promote it. **— Superseded, 2026-08-13.** That was the historical consequence before the Directive-4 amendment; the pending placeholder is now an **explicit blocking state**, so an unplanned batch does not promote (`turfgps-board-ops § The dependency representation`). The obligation the bullet describes is unchanged — the continuation is still mandatory — but the failure mode is now a *stall*, which is visible, rather than a *false promotion*, which was not.
- **A third load-bearing shell script.** `dependents.sh` parses issue bodies, so a story that writes `Blocked by` in unexpected shape is invisible to it. It fails toward *blocked* rather than *eligible* in every ambiguous case, which is the safe direction but will hold work when the grammar drifts.
- **Two grammars on the board for a while.** The grandfather clause means 59 sections in the old shape coexist with the new one until events touch them. Both are readable; only the new one carries a type.

**What this gives up:**

- **The scrum-master's second opinion.** It used to re-derive the ordering every sync, so an edge that was wrong had a chance of being noticed by an agent reading the architecture fresh. That check is gone on purpose — it cost a full architectural re-read per sync to almost always reproduce the same answer — and what replaces it is `dependency_finding`, which fires only when something looks wrong from where a consumer stands.
- **Immediacy.** Dependency reasoning now happens at graph-change time, so a graph event that nobody routes leaves stale edges in place until the next one. The failure mode moves from *expensive and current* to *cheap and possibly stale*.

**Reversibility: high.** The representation is unchanged issue-body text that predates this record, the new agent is one file, the script is one file that can simply stop being called, and restoring the scrum-master's Phase 3 block is a revert of a deletion. Nothing here migrates data.

## Amendment — 2026-08-13 (Owner Directive 4)

*Source: `docs/adr/agent-org-directive-4.md` §8–§16. Three clarifications to the decisions above; no agent added, no ownership moved, nothing reopened.*

### A1 — The story-batch continuation is dispatched by the RE, not relayed through the lead

*Amends P3 and P9. Context: directive-4 §14–§16.* The route as built woke `@engineering-lead` after every story batch **solely to relay `required_agents: [backlog-dependency-planner]`** — a hop with no decision in it, which §14 says to remove. So `@requirements-engineer` now dispatches the planner **directly and one-shot** once `@requirements-story-organizer` returns, carrying the batch's story numbers and the organizer's `dependency_hints` as references.

Three guards keep this from becoming a general cross-domain dispatch pattern, and all three are stated in both agents' files: **the RE does not own edges** — it triggers the pass and never writes a `## Dependencies` section; **the planner is not an RE sub-agent** — it is outside the pool of five, dispatched and finished, never managed or re-tasked; and **the RE may trigger it only on graph events arising from its own requirements or story changes.** Every other graph event — a `dependency_finding`, an architecture decision moving a boundary, a suspected graph defect — remains `@engineering-lead`'s dispatch under P9. **One shot, no ping-pong** (§16): the planner's findings return in its envelope and are corrected at the highest faulty layer, never answered with a second dispatch.

The planner therefore has exactly **two dispatchers**, and its Activation contract names both.

### A2 — Satisfied and removed are different operations

*Extends P2 and P4. Context: directive-4 §8–§11, ratified into `turfgps-board-ops § Satisfied is not removed`, which is the one home for the rule.*

A hard edge whose prerequisite is **successfully complete** is **satisfied**: it stops gating and **its line stays in the body as provenance** — why B followed A, what B consumed, what historically depended on A. **Satisfaction is derived, never written**: no `satisfied:` flag, because a written mirror of GitHub Status is the second source of truth that goes stale silently (§11). An edge is **removed** only by the planner, and only when the *relationship itself* no longer holds — scope change, architecture change, changed decomposition, a wrong edge — recorded in that pass's `graph_update`.

**Completion changes whether a dependency blocks; structural change changes whether it exists.**

### A3 — `dependents.sh` distinguishes completed from merely closed

*Corrects P6. Context: directive-4 §12–§13.* The script treated **any** `CLOSED` blocker as satisfying, so a prerequisite closed as **not planned** — work that never happened — read as done and could free a dependent for `Ready`. It now reads `state,stateReason`: satisfied **iff** `CLOSED` with stateReason `COMPLETED`, or with no stateReason at all (a legacy plain close). `NOT_PLANNED`, `DUPLICATE`, and any unrecognized closed reason **still block** and print on a `not_completed:` line with their reason, so `@scrum-master` files a `dependency_finding` instead of promoting a story onto dead work. The ambiguous case still fails toward *blocked*, as it did for an unreadable blocker.

## Amendment — 2026-08-13 (runtime hardening)

*Source: the Owner's runtime-hardening directive, which is deliberately not filed as a separate document — its clarifications are recorded in the existing ADRs. One clarification; no agent added, no ownership moved, nothing reopened.*

### A4 — Dispatch authority over the planner is exactly two agents, and reporting is not dispatching

*Consolidates P9 and A1, which stated the rule correctly, against three agent files that contradicted it.* The planner's dispatchers are **`@engineering-lead`** — every graph event — and **`@requirements-engineer`** — its own story batches, and only those (A1's three guards are unchanged). Nobody else wakes it.

`@pr-judge` and `@worker-manager` read as though they did: the judge routed `dependency`/`planning`-root-cause findings "to `@backlog-dependency-planner`" and listed it downstream, and the manager "carried a `dependency_finding` up to" it. Both now **classify and report upward**, and `@engineering-lead` dispatches. The distinction is the point: **a finding is evidence, a dispatch is a decision about whether the graph should be re-reasoned now**, and a planner with four dispatchers is one that runs on every reporter's estimate of its own urgency — the per-event discipline this record exists to hold. Root-cause classification is untouched and remains load-bearing: without a correctly named `dependency` or `planning` root cause, there is nothing for the lead to route.

Neither agent lost anything it owned. `handoff-payloads § Dependency findings and graph updates` states the same routing in the schema's own home.

## Amendment — 2026-08-16 (first live loop cycle)

*Source: the Owner's runtime-findings directive. One observation; no rule of this record changes.*

### A5 — The mandatory continuation is the one the loop actually dropped

*Concerns P3, P9, and A1. Recorded here because the failure landed on this record's pipeline, while the rule that closes it is general and lives in ADR-0001.*

A1 made `@requirements-engineer` the direct dispatcher of `@backlog-dependency-planner` after a story batch, on the reasoning that a mandatory pipeline step should not be relayed through an orchestrator. **Twice in the first live cycle the RE's process ended before its background children returned, so that continuation never fired** — once leaving a finished `FR-019` field block held by nobody.

**Nothing about A1 is wrong, and this is not a reason to restore the hop.** Relaying the dispatch through `@engineering-lead` would have added the decision-free hop A1 removed without making the continuation any more durable: the same process would still have ended, and the relay would have died with it. The missing rule was general — an agent must not end a pass while a continuation it owns is outstanding — and is recorded as **`ADR-0001 § D11`**, with its home in `agent-handoffs § An outstanding continuation is not left behind` and its citation at `requirements-engineer § Mode A`.

The consequence this record already anticipated holds, and the incidents sharpen it. An unplanned batch **stalls** rather than promoting falsely, because the `_Pending @backlog-dependency-planner._` placeholder is an explicit blocking state (`turfgps-board-ops § The dependency representation`), and that stall is what made both incidents visible at all. **A visible stall still ships nothing**, which is the half D11 closes.

## Amendment — 2026-08-29 (issue #138)

*Source: `#138` — reported by `@scrum-master` during the #37 reconciliation, both defects confirmed at source and the blast radius measured by `@engineering-lead`. One correction; no agent added, no ownership moved, nothing reopened.*

### A6 — `dependents.sh` counts the declared edges, and resolves a reference that is a pull request

*Corrects `§ P6`. Context: #138.* The script scanned the **whole** `Blocked by:` line for `#N`, so every issue and pull request named in an edge's *reason* became an edge of its own. Nothing prunes such a phantom: a satisfied edge's line stays in the body permanently as provenance (`§ A2`, `turfgps-board-ops § Satisfied is not removed`), so the story is held forever on a relationship nobody declared, and the exposure grows as prose accumulates cross-references. The script now parses only the **declared** list — the run of `#N` references that opens the line, ending where the reason begins. Measured over the 112 live `Blocked by:` lines on 2026-08-29: 49 over-extracted, 48 of them a harmless repeat of the blocker's own number that the membership test already absorbed, and one real phantom — `#136`, held on `PR #67`, merged three weeks earlier.

Second, a declared blocker may **be a pull request**: issue and pull-request numbers share one sequence and `gh issue view` resolves both. A `MERGED` pull request had no case and read `UNKNOWN`, which blocks — so the two defects compounded, the phantom being unresolvable as well as undeclared. A merged pull request is now **satisfied**, the work having landed; one still open blocks; one **closed unmerged** never landed and blocks on the `not_completed:` line. Which kind a reference is comes off the `url`, because `state` alone cannot separate a closed-unmerged pull request from the legacy plain close `§ A3` admits — both are `CLOSED` with no stateReason.

**The ambiguous case still fails toward blocked, exactly as `§ P6` and `§ A3` require, and that is unchanged.** A *declared* blocker whose state genuinely cannot be read is still counted as blocking and still named on the `unreadable:` line. What ends is blocking on a reference that was never an edge. Asserted by `scripts/loop/tests/dependents-declared-edges.sh`, which is hermetic and which fails against the pre-fix script.
