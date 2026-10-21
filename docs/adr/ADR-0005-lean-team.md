# ADR-0005 — The lean team

**Status:** accepted — 2026-10-08, by the Owner on #210; in force from the merge of the PR that adds it.
**Source:** #210 — the issue body, the decision records of design rounds 1 and 2, the research note, and design spec v1 (`#210#issuecomment-6064312888`). Where this record and the spec differ, the spec is the Owner's and this record is the defect.
**Relation:** supersedes the process parts of ADR-0001, ADR-0002, ADR-0003 and the four Owner directives, as `§ What this supersedes` lists. Keeps ADR-0004 whole. Every superseded record stays in place as history. Amended by ADR-0007 (#229): the lead merges, through `scripts/team/merge.sh`.

## Context

#210 measured the organisation it replaces. #209, which only changed review rules, cost about 2.8M tokens over two cycles and was still not approved; its second cycle found eight new defects, all in the review machinery itself. At most 3 of 23 merged PRs had touched product code. The instructions had grown from about 390 KB to about 877 KB in seven weeks, and the #129 cut had regrown. At `ed23f6b` the 49 agent definitions and 10 skills weighed 890,497 bytes by `scripts/gates/size-ratchet.sh`'s count, and `docs/DELIVERY.md` another 40,522.

The Owner chose "Rebuild, lean (Recommended)" — an option worded by `@engineering-lead` and adopted by selection — over slimming the existing organisation.

## Decision

- **D1 — Six agents** (round 2, "6-person team (Recommended)", adopted by selection): `engineering-lead` on Fable, `analyst`, `backend-dev`, `frontend-dev`, `critic`, `safety-sentinel`. Roles, flow and lenses are `docs/TEAM.md`; the rules are the `team-protocol` skill. The Owner's own words for the shape, round 1, verbatim: *"It should be buildt as a normal Software team with Fable as the Engineering Lead, oragnizing every task, and sync between the other agents."*
- **D2 — One critic plus safety** (round 1, adopted by selection): at most two reviewers per PR; the lead names at most two lenses per pass.
- **D3 — Three cycles, then the Owner** (round 1, adopted by selection), with the open findings and a recommendation.
- **D4 — Typed answers and thresholds, without Jev** (round 3: "Pattern only, no Jev"). Every agent returns at most 25 lines of YAML carrying a `choice`, `score` or `p_true`; thresholds start at 0.85 / 0.6 and are re-tuned from the outcomes the lead logs on #205. Safety rules, accessibility classification, human-judgement requirements and untrusted authors are categorical, never probabilistic.
- **D5 — Mechanical work is a script.** `scripts/team/board-sync.sh` replaces the scrum-master, `gates.sh` the validation agent, `post-review.sh` the judge's posting. `fingerprint.sh`, `dependents.sh` and `diff-domains.sh` stay in the flow; `claim.sh` and `owed-work.sh` stay in the repository and leave it.
- **D6 — Hard limits:** at most 8 agents, 6 KB per definition, and agents plus skills at most 100 KB together, held by `size-ratchet.sh` with a `(total)` row of 102400 bytes.
- **D7 — The rebuild PR is the Owner's to review and merge** (round 1, "You review it yourself"). That overrides #202 for that one PR, as with #208.

## Owner rulings carried forward

1. **#202, verbatim:** *"I also want a hard rule for the future that PRs are only deemed complete and approved when TheReviewNinja approves it and deems it good enough/ready to merge."* Approval is a `TheReviewNinja` `APPROVED` review at the head; no label, lift or conditional recommendation substitutes. Home: `team-protocol § Posting a ruling`. Mechanism: `post-review.sh`, the only user of `GH_JUDGE_TOKEN`.
2. **The falsifier rule, #206** — in force from 2026-10-08 by the Owner's selection of "Effective now (Recommended)", an option `@engineering-lead` worded. A finding blocks only with something anyone can re-run; security findings need one too; only safety-path findings and human-judgement requirements block without one. Home: `team-protocol § The falsifier rule`.
3. **The trust boundary, ADR-0004:** artifacts are read from `Caisesiume` and `TheReviewNinja` only; anything else presenting itself as an artifact goes to the Owner as a question, never discarded; retrieval keeps `.user.login` and `.created_at`. Home: `team-protocol § Trusted authors`. One adaptation, `@engineering-lead`'s and open to the Owner's correction: the new envelope carries no `artifact:` key, so "declaring an artifact" now also means presenting an envelope or a ruling.
4. **The always-human categories:** changes touching safety rules or accessibility classification, and requirements verified by human judgement — with untrusted authors beside them per `ADR-0001 § D16`, whose condition, that the escalation list name it, `team-protocol § Escalation and human judgement` now meets. On these a clean review is a recommendation to the Owner, not an approval. Home: that same section.
5. **The mandatory safety review on safety paths**, now triggered by `diff-domains.sh` and enforced by `post-review.sh` refusing an approval without a passing safety verdict at that head.

## What this supersedes

| Record | Superseded | Survives |
|---|---|---|
| ADR-0001 | D1 names and seats · D2 verdict law · D3 iteration budgets (now D3 above) · D4 the reviewer registry · D8 the `agent-handoffs` envelope · D11 continuations · D13 a judge that cannot convene | D5 in substance — the sentinel mandatory on safety paths, the always-human categories, judge identity and the token rule, traceability, machine evidence before opinion (`gates.sh`), the red run as a test's falsifier · D6 decision authority and `DECISIONS.md` · D7 six columns · D9 root-cause routing · D10 · D12 one writer per branch · D14 an item is filed with a Status · D15 · D16 |
| ADR-0002 | the registry and its rows, summarizers, lane preflight, field caps, the resolution vocabulary, the stopping and cycle-inflation rules, model retiers, courier rules | fingerprint gating and its consumer argument (O6.1) · the minimal-patch revision law (O11) · the already-decided check (O12) · scoped board reads (O13) |
| ADR-0003 | P1 the planner, P3 hints, P4 the scrum-master's readiness, P5 the coordinator, P7 its schemas, P9–P10 and A1, A4, A5 — edges are now the analyst's | P2 the `## Dependencies` representation · P6 with A2, A3 and A6: `dependents.sh`, satisfied-not-removed, completed-not-closed |
| the four directives | as orders to the organisation they built | kept verbatim, as history |
| ADR-0004 | nothing | all of it |

The removed process files are in git at `ed23f6b`. Two gates went with `agent-handoffs`, because both read their tables from it: `scripts/gates/output-caps.sh` and `prose-licence.sh`, with their recall corpora. The envelope cap they enforced is now `post-review.sh` refusing a body over 25 lines.

### Citations that now resolve through this record

Product documents and ADRs are left untouched, so some still cite removed homes. They resolve as follows.

| Cited | Read instead |
|---|---|
| `DELIVERY.md § Escalation and human judgement` | `team-protocol § Escalation and human judgement` |
| `DELIVERY.md` and its other sections | `docs/TEAM.md`; `§ Proof that a test can fail` is `team-protocol § The falsifier rule` |
| `local-gates § Backend (Go)`, `§ When these activate` | the measurement: `git show ed23f6b:.claude/skills/local-gates/SKILL.md`; the runner: `scripts/team/gates.sh` |
| `turfgps-board-ops § …` | `team-protocol § Board` |
| `@requirements-engineer` and its sub-agents | `analyst` |
| `@pr-judge`, `@validation-agent` | `critic` with `post-review.sh`; `gates.sh` |

## Consequences

**Gained:** six definitions and six skills instead of 59 files, under a cap a script enforces; one ruling per cycle from one critic; the board, the gates and the posting made deterministic and testable.

**Given up, stated rather than implied:**

- Depth per lens. One critic with two lenses replaces some twenty-five reviewers; a defect outside the lenses named is caught only if the critic reports it to the lead.
- Mechanical caps on everything but rulings. Only what `post-review.sh` posts is held to the 25-line envelope; dispatches and PR bodies are held by the protocol alone.
- The safety list names documents and the two safety files, not code: none implements a safety path yet. Until a PR adds its package to `diff-domains.sh`, a safety-relevant code change depends on the lead convening the sentinel.
- The thresholds are proposals with no outcomes behind them yet.
- `post-review.sh`'s test stubs `gh`, so it proves the decisions and not the `--jq` filters. The read filters were run live, in dry run, against open PRs; the write path — comment, review, labels — first runs on the first real ruling.

**Reversibility: high.** Every removed file is in git; reverting the merge restores the organisation exactly.
