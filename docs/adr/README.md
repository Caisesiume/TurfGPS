# Architecture Decision Records

This folder holds the decisions that would otherwise be re-argued.

A decision recorded in a conversation is lost the moment the conversation ends; a decision recorded in an agent file is invisible to everyone not reading that file. An ADR is where a decision goes when it outlives both.

## When one is owed

**A consequential decision gets an ADR** — the directive's word, and the test is consequence, not size. A decision is consequential when reversing it later would cost significantly more than making it now, when more than one agent or document has to obey it, or when the reasoning is the part that matters and the outcome alone would read as arbitrary.

**A recurring lesson gets an ADR, a registry rule, or a skill update — never resident context.** If the same finding keeps being raised, the fix is to write the rule where future agents retrieve it, not to carry the lesson forward in a conversation that will end. That is the whole of the directive's §29, and this folder is one of its three destinations.

Not everything is owed one. A choice with a single obeying agent belongs in that agent's definition; a rule about how a record is shaped belongs in the `requirements-authoring` skill; a rule about how a board is convened belongs in `review-board-dispatch`. **Put the rule where the agent that must obey it will look**, and use an ADR when that place is "everywhere" or "the reason, not the rule".

## Form

- **Filename** `ADR-NNNN-short-slug.md`, numbered sequentially from `0001` and never reused. A number is permanent even if the record is later superseded.
- **Status** one of `proposed` · `accepted` · `superseded by ADR-NNNN` · `rejected`. A superseded record stays in place with its status updated; deleting it destroys the reasoning that explains why the successor exists.
- **One decision per record.** Two decisions in one file cannot be superseded independently, which is exactly what happens to them.
- **Sections:** Status · Context (what the situation was and what it cost) · Decision (what is now true) · Consequences (what this obliges, and what it gives up).

State the cost as well as the benefit. A record that lists only advantages is advocacy, and the next reader will assume the tradeoff was never examined.

## Contents

| Record | Subject |
|---|---|
| `agent-org-directive.md` | Not an ADR — the Owner's directive of 2026-08-10, kept verbatim as the source order ADR-0001 adapts |
| `ADR-0001-artifact-driven-agent-org.md` | The ratified adaptation of that directive onto this repository's fleet |
| `agent-org-directive-2.md` | Not an ADR — the Owner's second directive of 2026-08-10, kept verbatim as the source order ADR-0002 adapts |
| `ADR-0002-token-efficiency.md` | Token efficiency of the agent organization — the audited leaks and the fourteen ratified optimizations |
| `agent-org-directive-3.md` | Not an ADR — the Owner's third directive of 2026-08-10, kept verbatim as the source order ADR-0003 adapts |
| `ADR-0003-backlog-dependency-planner.md` | A dedicated owner for the persistent Epic/story dependency graph — the edges leave `@scrum-master`, which now consumes them |
