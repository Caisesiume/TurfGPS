# Owner Directive — Artifact-Driven Agent Organization

**Issued:** 2026-08-10 by the repository Owner, verbatim in substance. **Ratified as:** ADR-0001 (sibling file), which records how each numbered section maps onto this repository's existing fleet. Where this directive and ADR-0001 differ on a repository-specific adaptation (names, numbers, board columns), ADR-0001 records the ratified adaptation; this file is the unaltered source order.

---

Refactor the existing Claude Code agent hierarchy into an adaptive, artifact-driven, autonomous software engineering workflow. The goal is to preserve or improve engineering quality while significantly reducing token usage, redundant agent execution, repeated context transfer, unnecessary review cycles, and human escalations.

The system must continue to support the existing workflow built around: `specification.md`, `architecture.md`, `design.md`, Requirements Engineering, GitHub Projects or Jira via MCP, Issues / work items, Milestones / epics, Sprint planning, Autonomous implementation, Git branches, Commits, Pull requests, Multi-agent peer review, Revision cycles, Merge decisions.

The primary architectural change is: **Agents must no longer behave like a continuously communicating organization. They must behave like independently executing specialists connected through small, structured handoffs and persistent engineering artifacts.**

The system should optimize for: 1. Autonomy 2. Correctness 3. Traceability 4. Minimal context transfer 5. Minimal redundant agent execution 6. Risk-proportional review 7. Explicit decision authority 8. Fast convergence 9. Low human intervention

## 1. Core hierarchy

Layer 0 — Human Owner. The human provides the original product intent through `specification.md`, `architecture.md`, `design.md`. These documents are the highest-authority product artifacts unless explicitly superseded by an approved decision or updated requirement.

The human should NOT normally participate in: implementation decisions, ordinary ambiguity resolution, code review disagreements, low-level architecture choices, prioritization between equivalent technical solutions, non-critical review findings. Human escalation is an exceptional path.

## 2. Engineering Lead

The Engineering Lead is the root orchestration agent. It must remain lightweight. It should NOT: perform detailed implementation itself, perform detailed code review itself, duplicate analysis performed by specialists, automatically forward complete subagent responses, wake every agent for every task.

Its responsibilities are: inspect the project board, identify executable work, understand dependencies between issues, construct sensible execution order, decide which specialist teams are required, delegate work, monitor workflow state, resolve ordinary cross-team decisions, enforce iteration and token budgets, escalate only genuinely product-defining questions to the human.

The Engineering Lead operates on structured summaries rather than complete agent transcripts.

## 3. Persistent source-of-truth artifacts

Product intent: `specification.md`, `architecture.md`, `design.md`. Requirements: produced by the Requirements Engineering system. Work tracking: GitHub Projects or Jira via MCP; work items should contain enough information for implementation agents to execute independently. Engineering decisions: maintain Architecture Decision Records or equivalent decision artifacts when consequential decisions are made. Source code: the repository itself is authoritative for current implementation state. Git history: commits and pull requests provide historical implementation context.

Do NOT attempt to reproduce all persistent information inside agent handoffs. Agents should retrieve authoritative artifacts when needed.

## 4. Requirements Engineering system

Create a Requirements Librarian responsible for maintaining requirements quality and consistency. It coordinates a pool of requirements specialists. Recommended specialists include: Specification Analyst, Ambiguity Analyst, Requirements Completeness Analyst, Requirements Consistency Analyst, Prioritization Analyst, Acceptance Criteria Analyst, Dependency Analyst.

These agents form a panel, but the entire panel does NOT need to execute for every change. The Requirements Librarian selects appropriate specialists according to the task. Their purpose is to transform `specification.md + architecture.md + design.md` into actionable requirements and project work items. Requirements work should result in persistent artifacts rather than large conversational handoffs.

A requirement should, where appropriate, define: purpose, expected behavior, constraints, acceptance criteria, dependencies, architectural implications, known risks, relevant source documents. Once requirements are stable, create or update GitHub Projects/Jira issues through MCP.

## 5. Decision authority for Requirements agents

Requirements agents must be explicitly authorized to resolve ordinary ambiguities themselves. They should infer intent using the following precedence:

1. Explicit specification
2. Architecture constraints
3. Design intent
4. Existing requirements
5. Existing system behavior
6. Established repository conventions
7. Most conservative reasonable interpretation

They must NOT ask the human merely because multiple technically valid interpretations exist. Choose the interpretation that best preserves product intent and document the decision.

Escalate only when: two authoritative product documents directly contradict each other; the decision materially changes product scope; the decision introduces substantial cost or irreversible architecture; legal/compliance/security intent cannot be determined; required business behavior fundamentally cannot be inferred.

## 6. Planning and risk classification

Before implementation starts, classify every work item. Create a lightweight Change Risk Assessor. It evaluates: files/components affected, architectural surface area, security relevance, authentication/authorization impact, data integrity impact, database/schema changes, external integration impact, public API changes, performance-sensitive paths, infrastructure/deployment impact, concurrency implications, backwards compatibility, test coverage, size of diff, novelty of implementation.

Return only structured decision data. Example conceptual output:

```yaml
risk:
  level: medium
  score: 0.46
domains:
  - backend
  - data
review_required:
  - correctness
  - maintainability
  - testing
review_optional:
  - architecture
review_not_required:
  - accessibility
  - frontend
  - ux
```

Do not include verbose prose unless the caller explicitly requests further explanation.

## 7. Implementation Coordinator

Create an Implementation Lead. The Engineering Lead delegates implementation work to this agent. The Implementation Lead determines which implementation specialists are required. Possible implementation specialists: Frontend Engineer, Backend Engineer, API / Integration Engineer, Data / Persistence Engineer, Infrastructure / DevOps Engineer, Test Automation Engineer.

Do NOT invoke all implementation agents automatically. Example: a backend validation change may require Backend Engineer + Test Automation Engineer. It should not automatically invoke Frontend Engineer, Data Engineer, Infrastructure Engineer unless the changed scope requires them.

## 8. Implementation contracts

Every implementation agent receives: issue identifier, objective, acceptance criteria, relevant architecture/design references, repository location, constraints, dependencies, exact scope of responsibility.

Agents should obtain source code directly from the repository. Do NOT copy large amounts of repository context into handoffs.

Implementation agents must: 1. inspect relevant code 2. implement the assigned scope 3. add/update tests where appropriate 4. verify their work 5. commit changes 6. return a compact structured handoff.

Example:

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

Do NOT return internal reasoning or a chronological description of the work.

## 9. Pull Request creation

Completed work should create a branch and pull request according to existing repository conventions. The PR must reference the relevant issue/work item. The PR becomes the primary review artifact. Reviewers must obtain information directly from: the PR diff, changed files, referenced issue, relevant requirements, relevant architecture/design artifacts. Do not duplicate these resources inside the reviewer handoff.

## 10. Peer Judge

Create a Peer Judge as the coordinator of the review process. The Peer Judge is NOT itself the primary reviewer. Its responsibilities are: 1. inspect PR metadata 2. inspect the change-risk assessment 3. inspect which areas changed 4. select relevant reviewers 5. collect structured verdicts 6. identify conflicts 7. request targeted follow-up where necessary 8. decide whether another revision cycle is justified 9. determine whether the PR can merge.

Most importantly: **Never automatically invoke the entire review panel.**

## 11. Reviewer registry

Maintain a registry of available review specialists. Recommended reviewers: Correctness, Security, Architecture, Maintainability, Readability, Testing, Performance, Data Integrity, API / Integration, Accessibility, UX, Infrastructure / DevOps, Documentation.

Each reviewer must declare activation criteria. Example — Security Reviewer activates for changes involving: authentication, authorization, credentials, encryption, user-controlled input, external requests, permissions, sensitive data, security boundaries. Do NOT activate merely because a PR exists. Accessibility Reviewer: activate primarily when user-facing interface behavior changes; do not invoke it for database migrations, backend refactors, CI changes, etc. Every reviewer must have similarly explicit triggers.

## 12. Mandatory evidence-based review

A reviewer is forbidden from issuing a valid verdict based solely on: PR description, another agent's handoff, issue description, summaries from other reviewers. Every reviewer MUST inspect the actual code or diff. Every review response must include evidence that this occurred.

Example:

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
verdict: revise
confidence: 0.96
```

If `inspected: diff: false`, the review is automatically INVALID. The Peer Judge must ignore the verdict and request a proper review.

## 13. Reviewer output contract

Reviewers should NOT return their entire reasoning process. Return decision-relevant data only. Each reviewer returns: reviewer, score, verdict (pass | revise | blocker), confidence, findings (id, severity, file, location, description, required_change), residual_risk, needs_followup.

The full review should normally remain compact. A reviewer may perform deep internal analysis, but only the conclusions necessary for downstream decision making enter the parent context.

## 14. Confidence Agent

Introduce a Confidence Assessor. This is a meta-review agent. It does NOT independently perform another complete code review. It examines: reviewer verdicts, reviewer confidence, evidence quality, disagreement between reviewers, unexplained findings, suspiciously shallow reviews, unresolved uncertainty.

Its job is to answer: **Do we have enough reliable evidence to make a decision?**

Follow-up must be targeted. Never rerun the complete review suite merely because confidence is insufficient.

## 15. Risk + Confidence decision model

The Peer Judge should use BOTH risk and confidence. Conceptually: Low risk + high confidence: merge. Medium risk + high confidence: merge if no blocker exists. High risk + high confidence: require all mandatory high-risk reviewers to pass. Low risk + low confidence: request only missing evidence. High risk + low confidence: perform targeted deeper review.

The solution to uncertainty is NOT automatically "run everybody again".

## 16. Reviewer findings and mandatory action

Avoid the concept `approved with suggestions` unless suggestions are explicitly recorded as non-actionable information. Every actionable finding must resolve into one of: `required_change`, `accepted_risk`, `invalid_finding`. There must never be an actionable suggestion that nobody owns. If a reviewer identifies something that should actually be changed, it becomes `required_change`. If it does not justify another implementation cycle, mark it as an accepted residual risk or informational observation.

## 17. Revision routing

When revision is required, do NOT restart the full workflow. The Peer Judge produces a minimal revision packet:

```yaml
revision:
  issue: PR-381
required_changes:
  - finding: SEC-01
    owner: backend
    scope: src/auth/session.ts
    change: invalidate old refresh token after successful rotation
review_after_revision:
  required:
    - security
    - correctness
  not_required:
    - architecture
    - accessibility
    - documentation
```

The Implementation Lead activates only the necessary implementation specialist. After the patch: ONLY reviewers affected by the change are rerun. Previous valid reviews remain valid unless the new diff intersects their review domain. This is critical. A one-line backend fix must not automatically cause every reviewer to execute again.

## 18. Incremental review validity

Track reviewer verdicts against reviewed diff state. A previously completed review remains valid when: subsequent changes do not touch the reviewer's relevant domain; dependencies affecting its conclusions have not changed; the Peer Judge determines the previous evidence remains applicable.

Invalidate reviews selectively. A documentation-only revision should not invalidate Security, Data Integrity, Performance. A database schema revision may invalidate Data Integrity, Backend Correctness, Performance — but probably not Accessibility, UX.

## 19. Revision convergence

Track improvement across review cycles. For each cycle record: unresolved findings, newly introduced findings, findings resolved, diff size, risk change, confidence change. The Peer Judge should determine whether the workflow is converging.

```yaml
cycle: 3
previous_findings: 5
resolved: 4
new: 0
remaining: 1
risk:
  previous: 0.61
  current: 0.31
confidence:
  previous: 0.77
  current: 0.94
converging: true
```

Do not blindly repeat review cycles.

## 20. Iteration budgets

Maintain explicit iteration budgets. Normal PR: maximum 3 autonomous revision cycles. High-risk PR: maximum 5 autonomous revision cycles. The current hard upper bound may remain as an emergency ceiling, but reaching it should be rare.

Before exceeding the normal budget, the Peer Judge must determine WHY convergence failed. Potential causes: conflicting requirement, unstable architecture, faulty reviewer, overly broad implementation, reviewer disagreement, ambiguous acceptance criteria, implementation repeatedly introduces regressions. Solve the cause rather than simply repeating the loop.

## 21. Human escalation policy

The Engineering Lead should NOT escalate ordinary engineering uncertainty. Human escalation is permitted only when at least one of the following applies:

- **Product intent is undefined** — multiple materially different product behaviors and source documents cannot distinguish between them.
- **Product documents conflict** — e.g., specification and design require mutually exclusive behavior.
- **Business tradeoff** — the choice requires knowledge unavailable to the repository, requirements or architecture.
- **Irreversible/high-impact decision** — destructive migration, major architectural replacement, substantial scope increase, externally visible breaking API change.
- **Risk exceeds autonomous authority** — a high-impact change cannot be made safe using established project constraints.

When escalating, return:

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

Never ask "What should I do?" The agent must provide a recommendation.

## 22. Default autonomous decision policy

If no escalation condition exists: **MAKE THE DECISION.** Agents have explicit authority to choose reasonable implementation details. When several solutions are valid, prefer: 1. compliance with specification 2. compliance with architecture 3. compliance with design 4. existing codebase patterns 5. lower complexity 6. smaller blast radius 7. easier reversibility 8. stronger testability 9. maintainability 10. least surprising behavior.

Record meaningful decisions instead of asking the human.

## 23. Context minimization

Never forward an entire subagent conversation to another agent unless absolutely necessary. A handoff should usually contain fewer than a few hundred tokens. Prefer references (issue, pr, finding, files) rather than copying complete requirements, complete PR descriptions, complete diffs, previous agent transcripts, chain-of-thought, repository files. The receiving agent should retrieve authoritative information itself.

## 24. Handoff principle

Every agent interaction follows: **INPUT REFERENCES, then EXECUTION, then STRUCTURED VERDICT.** Not: conversation after conversation after conversation. Agents should operate like functions with explicit contracts.

## 25. Handoff schema

Where practical, all handoffs use a shared envelope:

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

Individual agents can extend the schema. Do not populate irrelevant fields.

## 26. GitHub Projects / Jira lifecycle

The project board remains the workflow state machine. Recommended lifecycle: Backlog, Requirements Ready, Ready for Implementation, In Progress, Implementation Complete, Review, (Revision Required with revision implementation looping back to Review), Review Passed, Ready to Merge, Merged.

Agents should update project state through MCP. The Engineering Lead uses the board as the source of work rather than maintaining a duplicate task list in conversation history.

## 27. Greenfield workflow

Human (specification.md, architecture.md, design.md) feeds the Requirements Librarian (ambiguity analysis, requirement extraction, consistency analysis, acceptance criteria, dependencies, prioritization), which feeds the Project Board (Epics/Milestones, Issues, Dependencies), which feeds the Engineering Lead (work planning, risk classification), which feeds the Implementation Lead (Frontend*, Backend*, Data*, Integration*, Infrastructure*, Testing*), which produces a Pull Request, which is judged by the Peer Judge (Risk Assessor, Selected Reviewers*, Confidence Assessor): PASS merges; REVISE routes to a minimal implementation team and targeted re-review.

`*` means dynamically selected, never automatically all activated.

## 28. Feedback to earlier artifacts

Classify each finding by root cause: implementation | requirement | architecture | design | test | infrastructure.

If an implementation reviewer discovers a requirement defect, do NOT repeatedly patch the code — route the finding back to the Requirements Librarian. If a reviewer discovers an architectural contradiction, route it to the architectural decision process. Correct the highest-level faulty artifact, then propagate the resulting change downward. This prevents repeated implementation churn.

## 29. Learning without context inflation

Do not keep lessons learned indefinitely inside conversation context. Persistent project lessons should become: ADRs, repository conventions, updated requirements, updated design documentation, reviewer rules, agent activation rules. If Security Review repeatedly flags the same authentication pattern, update the engineering standard or relevant agent instruction. Future agents then retrieve the rule directly.

## 30. Token-budget awareness

Every orchestrating agent must optimize for the smallest sufficient execution graph. Before invoking an agent: does this agent have a reasonable chance of changing the outcome? If no, do not invoke it. Before repeating an agent: has the evidence relevant to this agent materially changed? If no, preserve the previous verdict. Before transferring context: can the receiving agent retrieve this information from an authoritative artifact? If yes, send the reference rather than the content.

## 31. Anti-patterns to remove

Refactor away any current behavior that causes: every reviewer running for every PR; every reviewer rerunning after every patch; complete subagent transcripts flowing upward; reviewers trusting handoff text instead of reading code; agents escalating harmless ambiguity; `10/10` being required regardless of finding significance; suggestions with no owner; repeated max-cycle loops without root-cause analysis; Engineering Lead accumulating the reasoning of all descendants; parent agents re-performing child analysis; requirements agents waking for implementation-only work; implementation specialists waking outside their domain; conversation history being treated as project memory.

## 32. Desired system behavior

Small change: Engineering Lead, Implementation Lead, 1 implementation agent, Risk Assessor, 2 relevant reviewers, Peer Judge, Merge.

Medium change: Engineering Lead, Implementation Lead, 2-3 specialists, Risk Assessor, 3-5 reviewers, Confidence Assessor, Peer Judge, targeted revision if needed, affected reviewers only, Merge.

High-risk architectural change: Engineering Lead, Requirements/Architecture validation, Implementation Lead, relevant specialists, Risk Assessor, larger reviewer panel, Confidence Assessor, Peer Judge, revision/architecture feedback as required, targeted re-review, Merge.

The complexity of the agent graph must scale with the risk and scope of the change.

## 33. Success criteria

The refactor is successful when: ordinary tasks run without human intervention; small PRs use only a handful of agents; reviewers always inspect actual code; irrelevant reviewers are not invoked; minor revisions trigger targeted review rather than complete review; subagent reasoning does not flood parent context; GitHub Projects/Jira acts as persistent workflow state; requirements and architectural artifacts act as persistent knowledge; meaningful decisions are recorded; agents possess enough authority to make routine decisions; human questions become rare and high-value; token consumption decreases substantially; review quality remains equal or improves; autonomous execution can continue for long periods without context-window exhaustion.

## Final implementation instruction

Inspect the existing agent definitions, hierarchy, prompts, MCP workflows, GitHub Projects/Jira interactions, review workflow, and handoff mechanisms. Refactor them toward the architecture described above. Preserve useful existing specialist agents rather than deleting them merely to reduce agent count.

The primary optimization is NOT fewer available agents. The primary optimization is: fewer unnecessary agent executions, smaller handoffs, selective invalidation, evidence-based review, and stronger autonomous decision authority.

Where existing agents overlap substantially, consolidate them. Where existing agents provide genuinely distinct expertise, retain them but introduce explicit activation conditions.

For every resulting agent definition, explicitly define: 1. Role 2. Responsibilities 3. Authority 4. Activation conditions 5. Required inputs 6. Required artifact retrieval 7. Required verification actions 8. Output schema 9. Allowed downstream agents 10. Escalation conditions 11. Context/handoff limits 12. Conditions under which the agent must NOT run.

Then update the orchestration rules so that this architecture is enforced during actual execution rather than merely documented.
