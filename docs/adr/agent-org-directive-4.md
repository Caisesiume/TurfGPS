# Owner Directive 4 — Final Contract Hardening

**Issued:** 2026-08-10 by the repository Owner, fourth and final in the series. **Ratified into:** the existing ADRs — ADR-0001 (ownership and the stabilization rule), ADR-0002 (efficiency rules touched), ADR-0003 (dependency semantics and the continuation route). Per its own §26, no ADR-0004 exists: this pass changes no genuinely new architectural decision, it hardens the existing ones. Where this file and an ADR differ, the ADR governs.

---

Perform a narrow final hardening pass over the current organization on `main`. The architecture is now fundamentally correct; this is not a redesign. Do not add agents, orchestration layers, reviewer families, planning layers, or telemetry systems. No broad prompt-compression campaign. No model retiers unless a concrete correctness defect requires it. Fix the remaining places where written contracts can cause agents to cross responsibility boundaries or re-reason about persistent state. The desired result: a stable architecture that can be run and evaluated on real operational evidence.

## 1. Inspect before editing

Inspect the current orchestration and dependency-pipeline files and the governing ADRs/skills/scripts before changing anything. Do not assume this prompt describes the repository perfectly; treat the repository as the implementation to refine.

## 2. Preserve the current ownership model

Requirements truth @requirements-engineer · FR authoring @requirements-fr · NFR authoring @requirements-nfr · corpus @requirements-librarian · story/Epic decomposition @requirements-story-organizer · persistent dependency relationships @backlog-dependency-planner · backlog readiness/promotion @scrum-master · runtime selection and assignment @project-coordinator · implementation-team selection @worker-manager · review orchestration and ruling @pr-judge · human-facing orchestration @engineering-lead. Do not collapse these roles; do not create another owner for any of them.

## 3. Fix contradictory authority language

Audit description / Role / Authority / Focus / Invocation / Contract / Guiding Philosophy in the orchestration agents. The protocol may be correct while a broader sentence near the top still implies wider authority — a model may act on the broadest instruction it sees. Wording must match actual decision ownership exactly.

## 4. Engineering Lead orchestrates, never reconstructs

It must NOT independently determine: the dependency graph, which story is structurally executable, whether a hard dependency is satisfied, which story enters Ready, persistent story ordering, implementation-specific dependency structure. It consumes dependency findings, persisted graph state, Ready/Blocked state, assignment state, review state, escalation state — and decides which owning agent needs to act. Prefer: "consumes persisted organizational state, identifies the next required organizational action, dispatches the owning agent, resolves cross-team decisions within its authority, keeps the execution graph small." Avoid anything readable as dependency or readiness analysis.

## 5. Scrum Master owns readiness, not runtime scheduling

Its decision: which Backlog items are eligible to enter Ready now? It owns reconciliation, Backlog→Ready, persisted-hard-dependency checking, traceability gates, WIP-aware stocking, suspected-defect reporting. It does NOT decide which Ready item executes next — that is @project-coordinator. Reword "what gets selected next" / "decides the next item" wherever the boundary blurs. The precise split: Scrum Master — eligible for Ready? Coordinator — which Ready item runs next?

## 6. Project Coordinator owns runtime selection only

Consumes already-Ready work; decides, given the Ready set and WIP, which item is assigned next. May consider priority, remands, capacity, merge collision, intake risk hints, in-flight work. Must not derive hard dependencies, promote Backlog items, redefine readiness, reconstruct requirements, or pick individual implementation specialists — @worker-manager owns specialist selection.

## 7. The four-stage execution boundary

Planner → persistent structural relationships. Scrum Master → readiness from those relationships. Coordinator → next Ready work to execute. Worker Manager → the implementation specialists. Engineering Lead orchestrates these owners; no other agent recreates these decisions.

## 8. Review dependency completion semantics

Inspect the persisted grammar and `scripts/loop/dependents.sh`. Required invariant: completing a prerequisite must change whether the dependency currently blocks execution, without unnecessarily destroying useful provenance. Do NOT assume literal `Blocked by:` text must remain forever; do NOT assume completed edges must be deleted. Choose the smallest representation cleanly distinguishing "dependency relationship exists" from "dependency currently blocks execution".

## 9. Preserve useful dependency provenance

After A merges, B stops being blocked — but that B depended on A often remains useful (why was B sequenced after A; which capability; was the edge invalidated later; what historically depended on A). Keep only what materially supports future graph validation, debugging, auditability, or change-impact analysis.

## 10. Distinguish these two operations

**Dependency becomes satisfied** — the relationship remains valid, its prerequisite is complete; it stops blocking execution. **Dependency is removed** — the relationship itself no longer holds (scope changed, architecture changed, the edge was wrong, the prerequisite is no longer required, decomposition changed). Never conflate "prerequisite completed" with "the relationship was wrong".

## 11. Prefer derived satisfaction over duplicated state

If consumers can already derive `hard dependency + prerequisite Done = satisfied`, do not introduce a redundant `satisfied: true` state mirroring GitHub Status. Avoid two sources of truth. Add explicit dependency state only if the current representation genuinely cannot express the distinction safely.

## 12. Update deterministic tooling if necessary

`scripts/loop/dependents.sh` should answer, without an LLM: which stories depend on #N; which dependencies still block; which are satisfied because the prerequisite is Done. If the representation already supports this cheaply, keep it; otherwise make the smallest coherent change.

## 13. Scrum Master consumes the resulting semantics

For every hard dependency: prerequisite successfully complete → satisfied; else → still blocks. It does not rebuild why an edge exists and asks no architecture questions merely to decide whether a persisted prerequisite completed.

## 14. Inspect the Story Organizer → Dependency Planner continuation

Once a new or materially changed story batch exists, dependency planning is usually mandatory pipeline continuation, not a new strategic decision. Does the continuation wake Engineering Lead solely to relay `required_agents: [backlog-dependency-planner]` without a real decision? If no, leave the route alone. If yes, remove that pure relay in the smallest architecture-consistent way.

## 15. Direct RE → Planner dispatch is allowed, not required

Acceptable if it provably removes a real semantic hop. Preserve ownership: the RE does not own edges; the planner is not conceptually an RE sub-agent; the RE may trigger it only on concrete graph events from requirements/story changes. Do not introduce a general cross-domain dispatch pattern for one path; if the existing route is already cheap and deterministic, preserve it.

## 16. No agent-to-agent conversation

The continuation is a one-shot structured dispatch. No RE → Planner → RE → Planner ping-pong. A requirement or decomposition defect found by the planner returns as a structured finding to the responsible owner; the artifact is corrected at the highest faulty layer.

## 17. Remove stale live-state snapshots from orchestration prompts

Board counts, milestone counts, temporary PR numbers, corpus counts, "all work is documentation right now". A fact that changes through normal operation belongs in an authoritative artifact, not a static agent definition — replace with "retrieve the current state from the authoritative source".

## 18. Not a whole-fleet cleanup

Primary scope: the orchestration and dependency pipeline. Search the rest of the fleet only for exact contradictions to ownership rules changed by this pass. Do not rewrite unrelated specialist agents, reviewer expertise, or registries.

## 19. No broad prompt-size optimization

Do not rewrite large specialist prompts for being large. Remove duplicated text only where it sits in files already being changed, the rule has a clear authoritative home, and removal cannot weaken the local contract.

## 20. Shared rules remain centralized

Preserve the centralized homes (agent-handoffs, turfgps-board-ops, review-board-dispatch, requirements-authoring, DELIVERY.md, the ADRs). Prefer concise references where an in-scope agent restates them — but keep local authority boundaries explicit; do not make agents cryptic to save tokens.

## 21. Preserve all current review optimizations

No regression of: deterministic preflight, lane closing, safety-path semantic guard, smallest sufficient panel, hard-negative review_not_required, marginal-contribution rule, conditional Confidence Assessor, evidence requirements, incremental validity, carried verdicts, duplicate-finding normalization, future_work, stopping rule, cycle justification, minimal-patch law, review accounting.

## 22. Preserve event-driven dependency planning

The planner never runs because time passed, the board was polled, a worker changed Status, a PR opened, an unrelated issue moved, or the generic fingerprint changed. It runs because dependency truth may have changed: new stories, changed scope, split/merge, prerequisite-affecting requirement changes, architecture boundary changes, explicit findings, decomposition changes.

## 23. Preserve selective graph invalidation

One story's change re-evaluates only the affected subgraph; unrelated relationships survive. Do not weaken.

## 24. Do not downgrade the Dependency Planner

Keep its current capable model absent strong operational evidence. Dependency decisions are infrequent with large downstream leverage. This pass is contract correctness, not model-cost tuning.

## 25. Remove stale-state wording conservatively

For phrases like "currently / right now / today / as of / there are N / the board has": classify each as stable project invariant, useful historical rationale, or live-state snapshot. Only the third is removed from runtime prompts. Historical dates in ADRs explaining decisions remain.

## 26. Update ADRs only where semantics actually change

Ownership → ADR-0001; token/execution efficiency → ADR-0002; dependency planning → ADR-0003. No ADR-0004 merely restating this prompt; a new ADR only for a genuinely new decision that belongs in none of the existing records.

## 27. Explicit non-goals

No new agents, reviewer-hierarchy redesign, orchestration layers, metrics agents, planning systems, board replacement, requirements-architecture changes, verdict-architecture changes, broad retiers, fleet-wide compression, renames for aesthetics, board-schema rebuilds, or reopening ratified decisions.

## 28. Validate the final ownership chain

Authors write requirements · RE owns requirements truth · Organizer creates Epic and story nodes · Planner owns dependency relationships · Scrum Master decides Ready eligibility · Coordinator decides which Ready item executes next · Worker Manager decides the implementing specialists · Judge chooses the panel and rules · Engineering Lead orchestrates and owns the human channel. No two agents claim the same decision.

## 29. Validate a greenfield batch

Trace: source section → RE → FR/NFR authoring → validated to-build requirements → Organizer (Epics + stories + hints) → Planner (persisted graph) → Scrum Master (Ready) → Coordinator → Worker Manager. Check: no human decision for ordinary ambiguity; no agent reconstructs another's persisted decision; dependency planning occurs exactly once per graph event; no mandatory continuation is lost.

## 30. Validate prerequisite completion

#41 prerequisite of #47; #41 merges → provenance remains available → #41 no longer blocks #47 → deterministic tooling identifies dependents → Scrum Master checks remaining blockers → #47 may become Ready. The planner does not re-plan merely because #41 finished: completion changes blocking state, not dependency truth.

## 31. Validate dependency invalidation

Architecture or scope changes → basis may no longer hold → planner runs → affected subgraph only → edges added/removed/changed → persisted graph updated. That is when dependency truth changes.

## 32. Validate defect routing

Implementation defect → Worker Manager · requirement defect → RE · story decomposition defect → Organizer · dependency/planning defect → Planner · architecture contradiction → ADR route · review evidence gap → targeted review / Confidence Assessor. Correct the highest faulty artifact; never encode a workaround at a lower layer.

## 33. Architecture stabilization rule

After this pass the architecture is stable. Record a concise maintenance rule: major changes to the agent organization are driven by observed operational evidence, not hypothetical optimization — e.g. repeated unnecessary invocations, routing failures, recurrent missing dependencies, review blind spots, excessive escalation, context pressure, measurable duplicate-reviewer cost, defects attributable to missing ownership. No metrics platform for this rule.

## 34. Use existing accounting first

The judge's execution accounting is preserved. A cheap metric may be added only where it falls naturally out of existing artifacts. No LLM-based metrics pipeline. The next optimization phase uses observed data from real runs.

## 35. Required output

Make the repository changes — not recommendations — and return the concise `final_contract_hardening` report: ownership conflicts fixed; dependency semantics (representation, completion behavior, invalidation behavior); orchestration hops (story→planner before/after, change made, reason); stale snapshots removed; files changed; agents added: 0; architecture redesigned: false; remaining risks; recommended next step: run the architecture and collect operational evidence before another organizational change. A justified no-op is preferable to unnecessary churn.

## Acceptance criteria

(1) EL wording no longer overlaps dependency/readiness ownership; (2) Scrum Master owns readiness, not runtime selection; (3) Coordinator owns selection from Ready; (4) Worker Manager keeps specialist selection; (5) Planner is sole semantic owner of relationships; (6) completion and invalidation are clearly distinguished; (7) completed prerequisites stop blocking without destroying useful provenance; (8) deterministic tooling handles satisfaction where practical; (9) story creation reliably triggers planning without an unnecessary semantic relay; (10) any dispatch-path change preserves ownership; (11) live-state snapshots are out of runtime prompts; (12) ADR history preserved; (13) no broad fleet refactor; (14) no new agents; (15) no review redesign; (16) token-efficiency rules intact; (17) greenfield works end-to-end; (18) planning stays selective and event-driven; (19) defects route to the highest faulty owner; (20) the architecture is explicitly stable after this pass.

## Final principle

Smallest changes necessary for internal consistency; optimize nothing that has not demonstrated a problem. Story Organizer creates nodes. Dependency Planner owns relationships. Scrum Master determines readiness. Project Coordinator chooses the next Ready work. Worker Manager chooses the team. PR Judge governs review. Engineering Lead orchestrates without reconstructing specialist decisions. **Completion changes whether a dependency blocks; structural change changes whether it exists.** After this pass, stop redesigning the organization and run it long enough to learn where the real costs and failures are.
