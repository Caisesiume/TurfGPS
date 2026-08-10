# Owner Directive 3 — Dedicated Backlog Dependency Planner

**Issued:** 2026-08-10 by the repository Owner, third in the series. **Ratified as:** ADR-0003 (sibling file), which records the pre-edit inspection and each repository-specific adaptation — including which dependency representation is authoritative and why. A targeted change on the ADR-0001/ADR-0002 architecture; where this file and ADR-0003 differ on an adaptation, ADR-0003 governs.

---

Refine the current organization by extracting durable backlog dependency reasoning from `@scrum-master` into a dedicated `@backlog-dependency-planner`. Targeted architectural change; do not redesign the wider hierarchy. Preserve the current responsibilities of the requirements family, `@engineering-lead`, `@project-coordinator`, `@worker-manager`, `@pr-judge`, and the review/risk/confidence architecture.

The problem: `@scrum-master` owns both board-state reconciliation/readiness AND semantic dependency/order analysis. Those are different responsibilities. Durable dependency reasoning should happen when the work graph changes, be persisted to GitHub, and be consumed cheaply by runtime orchestration. Desired flow: Requirements → Story/Epic decomposition → Backlog Dependency Planner → persisted dependency graph → Scrum Master readiness evaluation → Project Coordinator runtime scheduling → Implementation.

## 1. Create @backlog-dependency-planner

`.claude/agents/backlog-dependency-planner.md`. Role: owner of the persistent dependency graph between Epics and User Stories. Core question: **what must be true before this work can safely begin?** It owns dependency reasoning. It does NOT own: requirements truth, story content, board Status, worker assignment, implementation, PR review, priority itself, runtime scheduling.

## 2. Responsibilities

Analyze dependencies from: validated requirements, story acceptance criteria, story scope, architecture, design, existing dependency metadata, implementation boundaries, external prerequisites. Identify: hard story dependencies, soft story dependencies, Epic relationships, safe parallel work, critical-path hints where useful, dependency conflicts, stale or invalid edges.

## 3. Dependency reasoning examples

Schema before code that reads/writes it; migration before code requiring migrated state; domain model before adapters; interface/port before adapter implementation; backend API contract before frontend consumer; authentication foundation before protected features; ingestion before algorithms requiring ingested data; persistence before services querying it; core capability before its presentation; infrastructure before deployment-dependent work. Examples only — do not create dependencies merely because two stories are related.

## 4. Minimum necessary ordering

Prefer minimum necessary ordering and maximum safe parallelism. Do NOT turn the backlog into an unnecessarily serial chain. A dependency must have a concrete reason.

## 5. Hard versus soft

`type: hard` — downstream must not become Ready until upstream is complete. `type: soft` — this order is preferable but does not block execution. Each edge carries blocked story, prerequisite story, type, reason.

## 6. Persist the graph in GitHub

Output must not live only in conversation. Prefer, in order: (1) native GitHub issue dependency/blocking relationships if supported by current MCP/API/tooling; (2) structured project fields if already available; (3) machine-readable issue-body metadata (`Blocked by: #142`, `Soft dependency: #149`); (4) another existing repository-approved representation. Do not introduce a second competing dependency store. One authoritative representation, documented.

## 7. Story Organizer emits hints, not final authority

Preserve `@requirements-story-organizer` as Story Architect. It may infer local dependency hints while decomposing (`dependency_hints: story 143 likely_blocked_by 142, reason`). Hints only — it does not become authority over the global graph. After a story batch is created or changed: organizer → planner. The planner verifies and persists the actual graph.

## 8. Story Organizer handoff

Extend the compact handoff where useful: `stories_created` (issue, resolves, epic) + `dependency_hints` (downstream, upstream, reason). References only; no story bodies, no requirements.

## 9. Move semantic dependency analysis out of Scrum Master

Remove/substantially reduce its repeated architectural ordering reasoning (data plane before consumers, ports before adapters, schema before users, backend before frontend). That reasoning belongs to the planner. Scrum Master consumes the graph, never reconstructs it.

## 10. New Scrum Master responsibility

For each Backlog candidate: verify board/item traceability; read persisted hard dependencies; confirm every hard dependency complete; confirm no explicit blocking state remains; apply priority/WIP readiness policy; promote eligible work to Ready. It may validate that dependency metadata is internally consistent. It does not rediscover architectural dependencies each sync. Evidence of a stale/inconsistent graph routes as a finding to the planner.

## 11. Scrum Master must not silently repair graph semantics

Seeing a story with no blocker whose acceptance criteria clearly consume another story's API, it does NOT invent the dependency. It returns a `dependency_finding` (story, suspected prerequisite, evidence, recommended_action: planner verify). The planner owns the graph.

## 12. Project Coordinator remains runtime-only

It answers: of the work already Ready, what runs next? It consumes Ready state, priority, WIP, risk hints, merge collision information. It does not own persistent dependency reasoning and does not inspect the entire backlog to reconstruct dependencies.

## 13. Engineering Lead consumes the graph

It does not rebuild dependency ordering. It consumes board status, graph health, blocked items, Ready items, graph-change findings. It may dispatch the planner when the graph needs recomputation; it never performs the analysis itself.

## 14. Activation conditions

Run when: new stories created; stories split/merged; story scope materially changes; an Epic is added/reorganized; a requirement change affects implementation prerequisites; an architecture decision changes boundaries; the organizer emits hints; scrum-master reports a suspected graph defect; worker-manager discovers a missing dependency; the judge finds a planning/dependency root cause. Do NOT run merely because: time passed; the board was polled; a worker picked up a Ready item; a PR changed status; a review completed; an unrelated story moved columns.

## 15. Selective graph invalidation

Do not recompute the whole graph for one story change. Determine affected scope: the changed node, its prerequisites, its direct dependents, downstream nodes whose validity actually depends on it. Preserve unrelated edges — the same selective-invalidation philosophy PR review already uses.

## 16. Persist dependency provenance

Where practical, persist why an edge exists (basis: architecture section, requirement code). Not verbose reasoning — enough evidence that a future agent can verify the edge without recomputing the plan.

## 17. Already-decided rule

Before deriving a dependency: inspect existing dependency metadata, relevant ADRs, relevant requirement records, relevant architecture/design sections. Already decided with unchanged basis = reuse, do not re-litigate.

## 18. Output contract

Compact structured output: stories examined; edges added/removed/preserved (upstream, downstream, type, reason); newly unblocked; newly blocked; parallelizable sets; affected epics; confidence. No chronological narrative, no story text, no requirements.

## 19. Dependency conflict handling

Incompatible ordering implied by two artifacts: do not guess silently. Ordinary technical ambiguity: decide autonomously under the normal ladder. Escalate only on §21 conditions. First classify the conflict — story decomposition defect, requirement defect, architecture contradiction, or dependency metadata defect — and route to the highest faulty layer.

## 20. Feedback from implementation

Worker Manager and specialists may discover hidden prerequisites. They must not edit global dependency semantics. They return a `dependency_finding` (current story, missing prerequisite, reason, evidence, recommendation) routed to the planner.

## 21. Feedback from review

Extend PR Judge root-cause vocabulary with `dependency` and `planning`. Where a failure exists because work executed in the wrong structural order (consumer implemented before its contract), route to the planner. Do not repeatedly patch around an invalid backlog graph.

## 22. Updating the board after merge

When an upstream story completes, no LLM reasons from scratch about every blocked story. A cheap mechanism identifies stories whose persisted hard dependencies may now all be satisfied; scrum-master evaluates those candidates for Ready. Use deterministic tooling where possible.

## 23. Graph maintenance vs board status

Planner owns "A blocks B". Scrum Master owns "B is now eligible for Ready". Coordinator owns "B is the next Ready item to execute". Worker Manager owns "these specialists implement B". Do not blur.

## 24. Ownership model

Requirements truth @requirements-engineer · requirement authorship @requirements-fr / @requirements-nfr · corpus @requirements-librarian · story/Epic decomposition @requirements-story-organizer · persistent dependency graph @backlog-dependency-planner · board truth and readiness @scrum-master · runtime assignment/merge sequencing @project-coordinator · implementation routing @worker-manager · review @pr-judge. Reflect this table in contracts and governing documentation.

## 25. Greenfield flow

Human documents → Requirements Engineer → FR/NFR authors → validated corpus → Story Organizer → Epics + stories → Dependency Planner → persisted DAG → Scrum Master → Ready → Coordinator → Worker Manager → Implementation → PR Judge.

## 26. Example

Organizer creates #100 schema, #101 ingest, #102 spatial query, #103 solve endpoint, #104 planner UI, #105 CI. Planner persists 100→101→102→103→104 with 105 independent. Board exposes Ready: 100 and 105; the rest blocked. When 100 merges, no agent rethinks the architecture — the graph already says 101 depended on it; scrum-master checks blockers and promotes.

## 27. Token-efficiency requirement

Before: board changes → scrum-master reads backlog → rereads architecture → re-derives ordering. After: graph-changing event → planner reasons once → persisted; normal board change → scrum-master reads persisted state. The second path must be substantially cheaper.

## 28. Do not over-create agents

Only this new responsibility — unless inspection shows an existing agent cleanly owns it and can be refactored without breaking its runtime role. Do NOT leave durable dependency reasoning in scrum-master. The separation is the goal.

## 29. Update governing skills and documentation

Inspect and update as necessary: the story-organizer, scrum-master, project-coordinator, engineering-lead, worker-manager, pr-judge agent files; turfgps-board-ops and agent-handoffs skills; DELIVERY.md; ADR-0001; ADR-0002; any other artifact assigning dependency reasoning to scrum-master. One authoritative home for the dependency semantics; do not copy the policy into every agent.

## 30. Do not regress existing token optimizations

Preserve: fingerprint gating; scoped Project retrieval; references over pasted context; compact envelopes; smallest-sufficient execution graph; selective PR-review invalidation; risk-based review; conditional Confidence Assessor; deterministic checks before LLM calls; no LLM run merely to discover unchanged state.

## 31. Acceptance criteria

Complete when: (1) a dedicated owner exists for the persistent dependency graph; (2) the organizer creates work and emits hints but is not final authority; (3) scrum-master no longer reconstructs architectural order per run; (4) scrum-master consumes persisted hard blockers for readiness; (5) coordinator remains runtime-only; (6) engineering-lead consumes graph state; (7) analysis runs only when graph-relevant artifacts change; (8) one story change invalidates only its affected subgraph; (9) hard and soft are distinguished; (10) safe parallelism is explicitly preserved; (11) provenance is persisted compactly; (12) hidden dependencies found in implementation route to the planner; (13) planning defects found in review route to the planner; (14) unrelated board changes cause no recomputation; (15) newly satisfied dependencies are detected cheaply after merges; (16) GitHub remains the source of dependency truth; (17) the greenfield pipeline progresses without the lead or scrum-master rebuilding the graph.

## Final instruction

Inspect the current implementation first; do not blindly implement what already exists. Then perform the smallest coherent refactor establishing: **story decomposition creates the nodes; the planner owns the edges; scrum-master decides which nodes are ready; the coordinator decides which ready node runs next. Persist the edges. Do not repeatedly rediscover them.**
