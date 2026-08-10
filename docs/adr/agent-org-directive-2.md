# Owner Directive 2 — Token-Efficiency Optimization of the Agent Organization

**Issued:** 2026-08-10 by the repository Owner, second directive in the series. **Ratified as:** ADR-0002 (sibling file), which records the audited leaks, what was already solved, and each repository-specific adaptation. Preserves the ADR-0001 architecture; this order optimizes its execution cost. Where this file and ADR-0002 differ on an adaptation, ADR-0002 governs.

---

Review and refine the existing agentic engineering workflow. The system has already been substantially refactored toward an artifact-driven, selective, autonomous agent organization. **Preserve that architecture. Do not redesign from scratch.** Properties that must remain: lightweight Engineering Lead; autonomous requirements decision authority; skill-footprint worker routing; dynamic reviewer selection; risk-sized review; evidence-sufficiency (not re-review) confidence assessment; artifact retrieval over copied context; compact envelopes; diff-inspection evidence; cross-revision verdict validity; targeted revision packets; board as workflow state; requirements/ADRs/docs/PRs/issues/commits as persistent memory.

The task: **find and remove remaining sources of unnecessary LLM execution and token amplification**, preserving the quality bar. Optimize for the smallest sufficient execution graph and the smallest sufficient context for every unit of work.

Many registered specialists is not a problem. The problem is: agents running without a meaningful chance of changing the outcome; multiple agents answering substantially the same question; re-reading or re-summarizing what is already usable; re-evaluating unchanged information; orchestrators waking to discover nothing changed; parents retaining what artifacts already hold; revisions needlessly invalidating prior work; summarization as another expensive layer; routine monitoring consuming LLM calls. The desired system has 40–50 registered specialists while invoking a handful for a normal task.

## 1. Audit every agent invocation

For every invocation path answer: (1) What specific decision can this invocation change? (2) What evidence causes it to run? (3) What evidence causes it NOT to run? (4) Does another active agent already cover substantially the same decision? (5) Could deterministic logic replace this LLM call? (6) Could an existing artifact answer the question? (7) Could a prior verdict remain valid? (8) Is it invoked merely because it exists in a hierarchy? Any invocation without a strong answer to (1) is suspect.

**General law: no agent runs merely because its role is relevant in the abstract. It runs because there is a concrete unresolved decision in its domain.**

## 2. The Marginal Contribution Rule

Before invoking any additional reviewer whose domain substantially overlaps an already-convened reviewer, the caller must state the additional question this reviewer is expected to answer:

```yaml
candidate_reviewer: linus-quality-critic
overlaps_with: [go-quality-critic]
marginal_question:
  - "Could the implementation be logically incorrect despite being idiomatic Go?"
expected_unique_value: true
```

If `marginal_question` cannot be stated clearly: DO NOT INVOKE. Applies particularly to: Go Quality vs Linus Quality; Go Architecture vs Linus Architecture; Maintainability vs Code Smell; Modularity vs Structure; Evolvability vs Architecture; Performance vs Scalability; Design vs UX vs UI Engineering. Keep these agents available; do not merge them; require an explicit reason for simultaneous activation.

## 3. Audit reviewer activation rules for excessive breadth

Identify rules like "any Go diff", "medium+", "any code diff", "frontend diff" that make reviewers effectively mandatory at low contribution. Replace broad rules with evidence-based triggers — e.g. non-trivial behavioural change, new concurrency, new public functions/interfaces, error-handling changes, context propagation changes, substantial implementation logic, risk assessment requesting the lane. Do not weaken correctness coverage: eliminate ceremonial review, not critical review.

## 4. Make risk assessment drive the execution graph more strongly

`review_not_required` is a strong negative routing signal. The judge does not activate an agent in a not_required domain without recording a specific override reason:

```yaml
reviewer_override:
  reviewer: performance-reviewer
  risk_assessment: not_required
  reason: revision introduced an O(n^2) candidate loop not present in original assessment
```

Without an override reason: DO NOT RUN IT. `review_optional` does not mean "run if budget permits" — run only on a concrete signal: unusual diff structure; implementer flags uncertainty; reviewer conflict; acceptance criterion depends on the domain; prior defect history in the component; confidence gap.

## 5. Remove unnecessary review summarization

Structured verdicts are already compact; a summarizer consuming three 150–300-token verdicts to produce another summary may cost more than the judge reading the originals. **Structured data should not be summarized by another LLM merely to make it structured again.** A summarizer runs only when at least one applies: total verdict payload exceeds a defined threshold; at least five same-board reviewers ran; multiple substantive conflicts require synthesis; the judge explicitly cannot resolve overlapping findings directly; a genuinely large corpus is being reduced. For ordinary panels of 1–4: the judge reads verdicts directly. Do not preserve an agent solely because the historical architecture used it.

## 6. Deterministic routing before LLM routing

Examples: docs-only typo = risk low; no frontend files = no UX/design reviewers; no Go files = no Go reviewers; no schema files = no migration review; unchanged SHA = no re-review; no new board state = no state-analysis agents; CI red = no subjective panel yet; draft PR = no full review; no safety path touched = no sentinel; below summarizer threshold = no summarizer. Enforce with deterministic checks, scripts, skills, shell commands, repository metadata, or orchestration rules BEFORE an LLM is called. **Do not pay an LLM to answer a question Git, GitHub, a glob, a label, a SHA, or a script can answer exactly.**

## 7. Separate cheap detection from expensive interpretation

Two-stage pattern: cheap deterministic detection → only if signal exists → LLM specialist analysis. "Did auth-related code change? yes → security review", not "security reviewer reads every PR to determine nothing changed". Likewise "did board state change? yes → analysis", not periodic LLM board inspection discovering no change.

## 8. Fix heartbeat and monitoring token leakage

Periodic activity must not cause full LLM runs merely because time passed. **No LLM agent runs merely to discover that nothing changed.** Before invoking Scrum Master, Project Coordinator, State Reporter, or Engineering Lead monitoring passes, compute a cheap state fingerprint (board item ids+statuses+assignments+updated_at; open PR numbers+head SHAs+review/CI states; corpus revision; main SHA). If unchanged since the previous check: NO ANALYSIS AGENT RUNS. Monitoring cadence may remain; unchanged state must be nearly token-free.

## 9. Event-driven thinking over polling

Prefer execution caused by meaningful transitions — item became Ready, assignment changed, PR opened, head SHA changed, CI completed, verdict arrived, revision packet created, requirement changed — over "every N minutes, inspect everything". Where true events are unavailable, emulate with lightweight state fingerprints. Polling discovers events cheaply; LLMs interpret events, not perform the polling.

## 10. Strengthen incremental review validity

Ensure every reviewer's `Invalidated by` is as narrow as safely possible. Avoid "any code change invalidates" unless genuinely necessary. Prefer semantic invalidation (security: auth/authorization behaviour change, trust-boundary change, validation change, secret handling change, data exposure change — not an unrelated comment, formatting, or isolated other-domain change). For each revision compute a delta (changed files, changed domains, changed behaviors, findings addressed) and compare against each prior lane. Only invalidated verdicts rerun; carried verdicts remain explicit in the ledger.

## 11. Never rerun an agent because "the PR changed"

A changed SHA by itself does not invalidate semantic review. Machine verification (build, tests, lint, validation) may rerun every commit. Semantic specialist review reruns only when the revision intersects its domain. Enforce globally.

## 12. Minimize revision diff surface

Revision agents operate under a strict minimal-patch rule: no unrelated cleanup, refactoring, cross-file formatting, opportunistic abstraction, or while-I-am-here improvements — every additional changed surface risks invalidating reviews and triggering specialists. **During a revision cycle, minimizing semantic blast radius is itself a token-efficiency requirement.** Desirable-but-unrelated cleanup becomes a separate future work item.

## 13. Make revision cost visible to implementation agents

Add to their contract: before changing an additional file during a revision, ask — does this file need to change to resolve the named finding? If no: do not touch it. Initial implementation may refactor coherently; review remediation patches narrowly.

## 14. Eliminate recursive confirmation agents

Audit for chains where layers merely confirm another layer did its job. Valid: reviewer → judge; or reviewers → confidence assessor → judge when evidence is complex. **An agent whose primary job is restating another agent's structured output is probably unnecessary.**

## 15. Distinguish persistent memory from working context

Persistent information belongs in issues, the board, PR comments, the ledger, requirement records, DECISIONS.md, ADRs, documentation, commits, repository state. Conversation context holds only immediate execution state, active identifiers, current decisions, unresolved dependencies. No duplicate lists of tasks, reviewer history, requirements, decisions, board state, or implementation summaries inside the Engineering Lead conversation — reference the artifact.

## 16. Use IDs as memory pointers

Prefer `issue: 142 / pr: 381 / requirement: FR-24 / decision: DEC-031 / adr: ADR-0004 / finding: SEC-01 / review_sha: a1b2c3` over prose describing contents. Retrieve on demand. Applies at every layer.

## 17. Tighten the shared handoff envelope

The ~300-token limit is correct; audit field abuse. Per-field guidance: `summary` maximum 2 sentences; findings maximum 1 concise description + 1 required action each; `decisions` IDs and outcome only unless the reasoning cannot be retrieved elsewhere; `recommended_next_action` one action. Avoid arbitrary hard rejection where clarity suffers, but verbosity is an explicit contract violation.

## 18. Keep reasoning local

Agents may reason deeply internally; the complete analysis does not enter the parent context. Return conclusion, evidence pointers, uncertainty, required action — not the path taken. If more is needed later: one targeted question or re-open the artifact. Never preload downstream agents with speculative reasoning.

## 19. Avoid repeated retrieval of unchanged large artifacts

Artifact-driven architecture still wastes tokens if every subagent rereads the complete specification, architecture, design document, or requirement corpus for a narrow task. Introduce scoped retrieval: dispatch references point to specific sections (`Architecture.md § Solver boundaries`; `FR-142`; `DESIGN.md § Route result card`). Read the relevant section first; broaden only if the initial evidence demands it. **Retrieve progressively, not comprehensively.**

## 20. Introduce context escalation

Progressive context levels: **Level 0** — IDs and metadata (routing). **Level 1** — relevant issue/requirement/diff section (normal specialist work). **Level 2** — related architecture/design sections (cross-cutting reasoning). **Level 3** — larger project corpus (genuine systemic ambiguity only). Do not start at Level 3; broader context is earned through evidence that local context is insufficient.

## 21. Reduce repeated requirement reading during implementation

The work item carries enough traceability for the specialist to know what to inspect. Worker Manager routes `issue / requirements / architecture_sections / design_sections / scope`; the specialist retrieves only these first. On contradiction it escalates upward rather than loading the entire requirements universe.

## 22. Cache decisions through artifacts, not prompts

Recurring resolutions are recorded: requirement interpretation in DECISIONS.md; architecture choice as an ADR; recurring review rule in the registry/skill; coding convention in repository guidance; safety invariant in the specification/checklist; activation lesson in the registry. Future agents retrieve the rule rather than reason from scratch.

## 23. Add an "already decided?" check

Before reasoning about an ambiguity: search DECISIONS.md; search ADRs; inspect the relevant requirement; inspect the relevant board/PR record. If already decided: reuse it, do not re-litigate. Particularly for Requirements Engineering and Architecture.

## 24. Prevent review findings from regenerating repeatedly

A finding repeatedly discovered by an LLM is a candidate for automation: static analysis, lint rule, test helper, repository convention, agent instruction, architecture invariant, CI check. **Every recurring review finding is an opportunity to replace future LLM tokens with deterministic enforcement.** Track repeated finding categories where practical.

## 25. Prefer machines over reviewers for machine-checkable qualities

Move deterministic checks out of subjective agents: formatting, linting, type checking, compilation, schema validation, coverage thresholds, dead code, static analysis, dependency scans, generated-file consistency, API schema compatibility. Reviewers receive the resulting evidence; they do not rerun the same machine checks unless verifying a disputed claim.

## 26. Avoid validation duplication

ValidationAgent owns machine verification. Other reviewers do not independently rerun the full build/lint/test suite unless their verdict depends on reproducing a suspicious result. They inspect the machine evidence. Specialists rerun a targeted instrument only where their domain uniquely requires it.

## 27. Audit tool output size

Large tool output consumes context even when the handoff is small. Avoid: full diffs where changed-file filtering suffices; unlimited git log; full board dumps; complete requirement directories; broad recursive grep output; full logs after successful tests; entire CI logs when only failed steps matter. Progressive retrieval applies to tool output: changed filenames first → relevant patch; failed tests only → specific failure; specific requirement ID → surrounding section.

## 28. Treat successful output differently from failed output

Successful commands return minimal confirmation (`tests: status passed, suite: service/auth`). Failures capture failed command, exit status, failing test, relevant error excerpt — expand logs only if diagnosis requires it.

## 29. Introduce execution accounting

Per work item/PR, record per invocation: agent, reason, decision expected, triggered by, reused previous result. At completion, aggregate: agents invoked, reviewers invoked, reviewers carried, agents avoided by routing, revision cycles, summarizers invoked, human escalations. Not another LLM-generated report — structured metadata or a compact artifact. Purpose: answer "why did this PR cost as much as it did?"

## 30. Add a graph-bloat detector

Anomaly signals, not automatic failures: small change with more than ~5–7 meaningful specialist executions; low-risk PR with more than 3 domain reviewers; revision cycle invoking more agents than the original without increased risk; one-line revision rerunning several semantic reviewers. Inspect when the signal fires; the decision stays semantic.

## 31. Track marginal value of review cycles

Per cycle: new findings, resolved findings, new blockers, reviewers run, reviewers carried, diff size, risk change, confidence change. A cycle adding no meaningful finding, changing no risk or confidence, yet rerunning several agents means the routing policy wasted work — use it to improve activation rules.

## 32. Learn activation policy from actual outcomes

Once enough PRs exist, analyze: reviewers frequently returning N/A; frequently passing with no findings; findings usually duplicates; often overturning another verdict; discovering high-severity unique findings. Improve registry triggers from that evidence. Do NOT remove a reviewer merely because it passes often — a security reviewer passing 99% of the time may be essential. The metric: **when it finds something, was the finding unique and outcome-changing?**

## 33. Detect duplicate findings before another cycle

The judge normalizes findings by file, location, root cause, required change. Two reviewers identifying the same defect = one underlying issue with multiple supporting reviewers (`finding: CORE-07, supported_by: [go-quality: GOQ-03, maintainability: MAINT-02]`). Multiple votes strengthen evidence; they do not imply multiple fixes, tasks, or cycles.

## 34. Confidence Assessor remains conditional

Do not invoke it merely because review happened. Conditions: high risk; disagreement; sufficiently large panel; suspicious evidence; coverage gap. For low-risk PRs with two strong matching verdicts the judge decides directly. **Never a mandatory review of the reviews.**

## 35. Confidence follow-ups remain one-hop

One reviewer, one question. Never confidence → judge → three reviewers → summarizers. A confidence gap results in one targeted evidence request wherever possible.

## 36. Human escalation does not restart context

Persist the escalation packet on the issue/PR. When the human responds: resume from that artifact; do not reconstruct the prior conversation. The answer updates the owning persistent document if it changes durable truth.

## 37. Distinguish blocking from interesting

Every discovery classifies as `required_change | accepted_risk | future_work | informational`. `future_work` covers valid improvements outside current scope — create a traceable issue where warranted; do not start a revision because something is interesting, and do not lose it either. This prevents autonomous perfection loops.

## 38. Prevent 10/10 syndrome from returning

No lane requires subjective perfection. A PR passes when required behaviour is satisfied, gates pass, no unresolved required change exists, risk is acceptable, evidence is sufficient — not when every reviewer can imagine no possible improvement. There is always another refactor. The loop must have a stopping rule.

## 39. Formalize the stopping rule

The judge stops when: `required_changes: 0 · machine_gates: green · required_review_lanes: satisfied · confidence: sufficient_for_risk · human_gate: false`. At that point: MERGE/APPROVE. Do not ask agents for final thoughts. Do not perform one last review. Do not run a polish cycle. **Stopping is part of correctness in an autonomous loop.**

## 40. Review the model assigned to each agent

Do not blindly downgrade. High-reasoning roles (Engineering Lead, Requirements Engineer, PR Judge, difficult architecture) may need the stronger model. Mostly structured/classification roles (risk assessor, confidence assessor, state classification, straightforward filing, narrow reviewers) may use a cheaper one. Deterministic work uses no LLM: board-state diff, SHA comparison, file categorization by path, CI extraction, unchanged-state detection, ledger bookkeeping. Spend the highest-quality model where difficult synthesis and judgment occur.

## 41. Avoid expensive orchestrators doing cheap work

Engineering Lead and PR Judge do not spend expensive context on formatting YAML, counting cards, comparing SHAs, extracting filenames, copying labels, or compiling ledger rows — delegate to tools/scripts. Expensive cognition goes to conflict resolution, scope interpretation, sequencing tradeoffs, risk judgment, merge decisions.

## 42. Audit agent prompt size itself

Repeated invocation makes static prompt content an ongoing cost. Audit for repeated philosophy, handoff schemas, reviewer laws, escalation policies, tool commands, repository background, duplicated examples. Move shared rules into skills where loading behaviour makes it beneficial. Do NOT blindly shorten: remove duplication, not expertise. **One authoritative home per shared rule.** Agent files contain primarily unique role identity, activation conditions, domain expertise, authority, output expectations.

## 43. Avoid loading every shared skill automatically

Do not replace prompt duplication with every agent loading ten skills at startup. A docs reviewer does not need implementation routing rules; a backend worker does not need the reviewer registry; a risk assessor does not need full DELIVERY doctrine. **Load the smallest sufficient instruction set.**

## 44. Progressive skill loading

Start with the agent's local contract; load one relevant shared skill; load additional policy only if the task crosses that boundary. Avoid loading full organizational doctrine into every child agent.

## 45. Audit board operations

Board retrieval is scoped. Do not repeatedly retrieve every card and field when only Ready / In Progress / Ordered Revision / Awaiting Human matter. Prefer query/filter operations over full dumps. Coordinator receives only items relevant to scheduling; Worker Manager one assigned item; PR Judge one PR. Hierarchical information locality.

## 46. One item creates one context island

Each board item/PR is a mostly isolated execution context. Do not carry another item's implementation/review state in unless a declared dependency exists. Engineering Lead may know IDs and status globally; detailed context stays local. Long-running sessions must not accumulate the entire project.

## 47. Dependency context is explicit

If B depends on A: reference A (`depends_on: issue 142, contract_used: API /route/solve`) — never copy A's implementation summary. Retrieve detail only if required.

## 48. Prevent cross-agent conversational ping-pong

When a specialist needs another domain's clarification it returns a structured uncertainty (`status: blocked, needs_domain_decision: {domain, question, evidence, recommendation}`). The orchestrator routes one targeted request; the answer becomes an artifact or compact decision. No open-ended chat between agents.

## 49. Every additional cycle justifies itself

Before another revision/review cycle the judge records `next_cycle_justification: {unresolved_required_changes, expected_reviewers, expected_new_information}`. If `expected_new_information` is empty: do not start another review cycle. Fixing an implementation issue may justify targeted validation; it does not necessarily justify semantic re-review from every previous lane.

## 50. Preserve quality while reducing tokens

Optimization must NOT: skip required security review; skip safety review; hide uncertainty; weaken traceability; ignore architecture changes; accept unsupported verdicts; replace semantic judgment with naive path matching; merge because a budget ran out. **Remove duplicated cognition, not necessary cognition.**

## 51. Desired execution examples

Tiny documentation fix: deterministic classifier → docs reviewer if semantic content changed → validation if required → judge (1–2 LLM agents). Not a ten-agent chain through Lead, Scrum Master, Coordinator, Worker Manager, Risk Assessor, Writer, Reviewer, Summarizer, Confidence, Judge.

Small Go bug fix: Worker Manager → Change Risk Assessor → Go Worker → Validation Agent → 1–2 relevant reviewers → PR Judge; Confidence Assessor only if evidence/disagreement warrants.

Small revision after a security finding: revision packet → owning specialist → Validation Agent → security reviewer → possibly correctness if semantic behaviour changed → PR Judge. All unrelated prior reviews carried forward.

High-risk architecture change: Engineering Lead → relevant Requirements/Architecture decision → Worker Manager → several specialists → Risk Assessor → larger selected panel → Confidence Assessor → PR Judge. The large graph is justified because the risk is large.

## 52. Produce an optimization report before editing

Inspect the repository first; produce a plan of `token_leaks` (source, current behavior, why expensive, proposed change, quality risk, expected saving), `redundant_agent_pairs` (agents, overlapping domain, proposed routing rule), `deterministic_replacements` (current LLM task, deterministic signal, proposed implementation), `prompt_duplication` (shared rule, repeated in, authoritative home). Drive the edits from the analysis. Do not mechanically implement suggestions the repository shows are already solved.

## 53. Then update the actual orchestration

Modify `.claude/agents/*.md`, `.claude/skills/**`, `docs/DELIVERY.md`, ADRs, routing registries, handoff schemas, supporting scripts/configuration as necessary. One authoritative source per shared rule. Do not merely write another optimization document — **the execution behaviour must change.**

## 54. Preserve existing good decisions

Do not regress: evidence-based review; VERIFIED INDEPENDENTLY / ACCEPTED ON TRUST; structured handoffs; the review ledger; selective invalidation; risk×confidence; autonomous requirements decisions; narrow revision packets; explicit escalation policy; smallest-sufficient-panel; board-as-memory; artifact references; safety and human-verification gates.

## 55. Final acceptance criteria

Complete only when: (1) no reviewer runs simply because a PR exists, beyond machine/safety floors; (2) overlapping reviewers require a unique marginal question; (3) structured verdicts are not automatically summarized; (4) confidence review is conditional; (5) unchanged state causes no expensive monitoring executions; (6) deterministic checks precede LLM routing wherever possible; (7) prior semantic reviews survive unrelated revisions; (8) revisions minimize blast radius; (9) context is retrieved progressively; (10) shared policies are not duplicated across prompts; (11) shared skills are not universally loaded; (12) tool output is scoped, success output compact; (13) repeated findings become deterministic enforcement candidates; (14) every invocation has a concrete decision it may change; (15) every additional cycle states its expected new information; (16) the loop has an explicit stopping rule; (17) routine ambiguity stays autonomous; (18) human escalation stays exceptional; (19) high-risk changes still get deep scrutiny; (20) token optimization never overrides safety or correctness.

## Final principle

Do not ask "how do we make each agent use fewer tokens?" Ask first: **"Why was this agent invoked at all?"** Then: "Why did it need this much context?" Then: "Why must this work be repeated?" The largest savings come from work that never needs to happen. The registry may stay large; the runtime execution graph should be small. The ideal system is not fifty agents communicating efficiently — it is fifty agents available, and only the four whose judgment can change this particular outcome ever waking up.
