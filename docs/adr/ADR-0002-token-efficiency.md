# ADR-0002 — Token efficiency of the agent organization

**Status:** accepted — 2026-08-10
**Source:** `docs/adr/agent-org-directive-2.md`, the Owner's second directive, kept verbatim. Where that file and this one differ on a repository-specific adaptation, **this record is the ratified form**. Section references (`§N`) point at that directive.
**Relation to ADR-0001:** this is wave 2 **on** the ADR-0001 architecture, not a replacement for it. Every property listed in §7 as *must remain* remains. Nothing here deletes a specialist, weakens a safety floor, or reopens a decision ADR-0001 settled.

## Context

ADR-0001 made review *selective*. It did not make execution *cheap*, and the audit below found that the remaining cost was concentrated in places the first wave never looked at: what every agent loads before it starts, what runs when nothing has happened, and what runs because a rule said "any diff of this kind" rather than "this decision is unresolved".

The governing sentence of this wave, and the one every decision below is an application of:

> **No agent runs because its role is relevant in the abstract. It runs because there is a concrete unresolved decision in its domain.**

### Already solved — do not re-implement

The directive proposes several things this repository had already built. They are recorded here so a future reader does not mistake their absence from the Decision section for an oversight:

| Directive | Where it already lives |
|---|---|
| §25/§26 — reviewers do not rerun the machine suite | `review-verdicts § What the obligation reaches`; `@validation-agent` runs last and alone |
| §17/§18 — files **and** domain intersection test | `review-board-dispatch § The intersection test` |
| §35 — confidence follow-ups are one-hop | `DELIVERY.md § The merge decision`; `@confidence-assessor` |
| §30 — pre-invocation questions | `agent-handoffs § Before you invoke anything` |
| §21 — escalation packet carries a recommendation | `DELIVERY.md § Escalation and human judgement` |
| §19 — review ledger with carried verdicts and SHAs | `review-board-dispatch § Incremental review validity` |

### §52 audit appendix

Measured on this repository, on 2026-08-10, before any edit in this wave.

**`token_leaks`**

| ID | Leak | Measured |
|---|---|---|
| L1 | The evidence law lived in the judge-side dispatch skill, so **26 agent files** loaded reviewer-registry and dispatch mechanics to reach the standard their own verdict is measured against | 26 of 48 agent files reference `review-board-dispatch`; **all 23** that cite the evidence law already reference `agent-handoffs` |
| L2 | Unconditional ~25-minute monitoring dispatches — a full `@scrum-master` analysis every cadence tick whether or not anything moved | `engineering-lead.md § Session Cadence`, before this wave |
| L3 | Activation rows broad enough to be effectively mandatory: *any Go diff*, *any code diff at medium+*, *frontend diff*, *medium+ tier* | 5 registry rows |
| L4 | Invalidation rows broad enough to re-run on unrelated revisions: *any Go change*, *any further code change*, *any further substantive code change* | same 5 rows |
| L5 | Summarizer threshold of ≥3 on boards with 3 and 4 members — met by an ordinary panel | Go board 3 members, Linus board 4 |
| L7 | `review_not_required` was a soft hint with no recorded cost to ignoring it | `change-risk-assessor.md`, before this wave |
| L9 | Handoff fields unbounded below the ~300-token envelope total | `agent-handoffs § The limit`, before this wave |
| L13 | **43 of 48** agent files on the top model tier, including summarizers and structured classifiers | `grep '^model:' .claude/agents/*.md` |
| L14 | **6,873** total agent prompt lines, with **seven critics over 200 lines each** | go-quality 297 · go-architecture 253 · linus-security 236 · linus-quality 221 · linus-architecture 207 · linus-structure 206 · go-structure 202 |
| L16 | Unscoped board reads — full card-and-field dumps where only IDs and statuses are consumed | `turfgps-board-ops` |

**`redundant_agent_pairs`** — seven families, each kept and none merged: go-quality ↔ linus-quality · go-architecture ↔ linus-architecture · maintainability ↔ code-smell · modularity ↔ structure critics · evolvability ↔ architecture lanes · performance ↔ scalability · ux ↔ design ↔ ui-engineer.

**`deterministic_replacements`** — loop-state fingerprint · diff-domain classification · docs-only classification · draft-PR skip · unchanged-SHA carry · summarizer threshold. All six were LLM judgements about facts Git, GitHub, a glob, or a SHA answers exactly.

**`prompt_duplication`** — the evidence law, one home instead of 26 loads · the reviewer registry, loaded by non-judges that never convene anyone · philosophy restated in the seven >200-line critics.

## Decision

### O1 — The evidence law moved to `agent-handoffs`

*Context: L1, §42's "one authoritative home per shared rule", §43's "load the smallest sufficient instruction set".*

`review-verdicts § A reviewer does not accept a claim it could check` — the rule, the two-halves block, what the obligation reaches, and both recorded incidents — moved to **`agent-handoffs`**, beside the verdict schema that carries it. `review-board-dispatch` keeps the heading with a one-line pointer, so the 23 existing citations still land somewhere true until a fleet pass retargets them. `pr-judge.md` and `DELIVERY.md` are retargeted in this wave.

The law itself is **unchanged in substance** — this is a move, not a rewrite. Its purpose is that reviewers stop loading the judge-side skill entirely; every one of the 23 already loads `agent-handoffs`, so the new home costs them nothing.

**The home moved once more on 28 August 2026, and this decision is unchanged by it.** `agent-handoffs` was split, and the evidence law travelled with the verdict schema it measures into `review-verdicts` — still one skill for a reviewer, now without the payload schemas of every other role attached. What O1 decided is where this law sits *relative to the verdict*, and that has held through both moves; the citations were retargeted in the same diff that made them false.

### O2 — Five tightened registry rows, and one mandatory set

*Context: L3, L4, §3, §10.*

`@go-quality-critic`, `@code-smell-reviewer`, `@maintainability-reviewer`, `@ux-reviewer`, `@design-reviewer` receive evidence-based activation and narrowed invalidation — the rows are in `review-board-dispatch § The reviewer registry` and are the authority. **Medium tier's mandatory set drops `@maintainability-reviewer`**; the floor is validation, the correctness lane, and conditional confidence. The reviewer is not weakened, only un-mandated by *tier alone*: its row still fires on a new module, ~150+ changed lines, or a risk assessment requesting the lane.

**What this gives up:** a rename-only Go diff no longer gets an idiomatic-Go pass, and a medium diff introducing no new concept no longer gets a maintainability read. The wager is that neither was changing an outcome.

### O3 — The marginal contribution rule

*Context: the seven redundant families, §2.*

A reviewer overlapping one already convened runs **only with a recorded `marginal_question`** naming what it uniquely answers; if that cannot be stated, do not invoke. Within a family, default to the one whose trigger matched most specifically. The rule and the family table are in `review-board-dispatch § The marginal contribution rule`. **The agents stay registered and distinct — this is a convening condition, not a merge.**

### O4 — `review_not_required` is a hard negative

*Context: L7, §4.*

The judge convenes a `not_required` lane only with a recorded `reviewer_override` (reviewer, assessment, reason). `review_optional` runs only on a concrete signal from the §4 list. In `review-board-dispatch § Negative routing`, `pr-judge.md § Phase 3`, and `change-risk-assessor.md`. An assessment that can be ignored at no cost is an assessment nobody has to write honestly.

### O5 — Summarizer demotion

*Context: L5, §5, §14.*

A board summarizer runs when **5+ members of that board ran this cycle**, or the judge records multiple substantive cross-reviewer conflicts requiring synthesis, or the combined payload is genuinely too large to weigh directly. For 1–4 compact verdicts the judge reads them directly. This replaces the ADR-0001 §D4 ≥3 rule everywhere it appeared; the re-narration reasoning stands.

**Stated plainly because it is the intent, not a side effect:** the three-member Go board and four-member Linus board **can no longer meet the count condition** — their summarizers are conflict-triggered only.

### O6 — Deterministic preflight

*Context: all six deterministic replacements, §6–§9.*

Two scripts, both tolerant of failure and both tiny in output:

- **`scripts/loop/fingerprint.sh`** — open PRs, board item IDs and statuses, remote `main` SHA, corpus-and-ADR head. `UNCHANGED` exit 0 · `CHANGED` exit 10 · degraded exit 2.
- **`scripts/loop/diff-domains.sh`** — per-domain file counts, `docs_only`, `lanes_closed`, `safety_path_candidates`.

`.claude/state/` is gitignored. The preflight is normative in `review-board-dispatch § Deterministic preflight` and `pr-judge § Phase 0`: exact negatives close lanes with no LLM; a draft PR stops; an unchanged head SHA is a full carry. **`docs_only` is not itself a skip** — whether the risk assessor runs is decided by its row's `Never when` in `review-board-dispatch § The reviewer registry`, and O17 below is why this record states no flat rule of its own.

**The §50 guard is stated verbatim in both homes:** *deterministic checks close lanes only where the file-domain mapping is exact; anything semantic — safety paths above all — stays with the assessor and the judge.* This is why `diff-domains.sh` only ever **closes** a lane, never opens one, and why `safety_path_candidates` prints `hint_only: sentinel activation is semantic` beside itself. A file list can prove a Go critic has nothing to read; it cannot prove a safety rule was untouched, and those two errors do not cost the same.

### O7 — Field caps, the context ladder, tool-output discipline

*Context: L9, §17, §19, §20, §27, §28. All three in `agent-handoffs`.*

Per-field caps: `summary` 2 sentences · each finding 1 description + 1 required action · `decisions` IDs and outcome unless the reasoning is unretrievable · one `recommended_next_action`. **Verbosity is a contract violation, not a style issue** — the receiver pays for every word and cannot decline.

The context ladder: **Level 0** IDs and metadata → **1** the named issue, requirement, or diff section → **2** related architecture and design sections → **3** the wider corpus. Each level is earned by evidence the one below was insufficient; **never start at 3**.

Tool output: success is a one-line confirmation; failure returns the failed command, exit status, failing assertion, and relevant excerpt, expanded only if diagnosis requires. Filenames before patches, failed tests before logs, one requirement before a directory.

### O8 — Five finding resolutions

*Context: §37, §38.*

`required_change | accepted_risk | invalid_finding | future_work | informational`. **`future_work`** is valid work outside scope: the judge records it as a traceable issue reference, or hands it to `@engineering-lead` to route — **never a revision trigger, never lost.** **`informational`** is recorded and actioned by nobody. In `DELIVERY.md § Findings and their owners`, `review-verdicts § Reviewer verdict`, `pr-judge § Phase 8`.

### O9 — Stopping rule, cycle justification, accounting

*Context: §29, §30, §33, §38, §39, §49.*

**The stopping rule**, in `DELIVERY.md § The stopping rule` and applied at `pr-judge.md § Phase 7`: `required_changes: 0 · machine_gates: green · required_review_lanes: satisfied · confidence: sufficient_for_risk · human_gate: false` → MERGE. No final-thoughts round, no one-last-review, no polish cycle. **Stopping is part of correctness**; there is always another refactor.

**Cycle justification:** a remand records `next_cycle_justification`, and an empty `expected_new_information` means **do not start the cycle**. Targeted validation of a fix is not semantic re-review.

**Accounting:** the ledger gains a compact metrics footer — agents invoked, reviewers invoked and carried, lanes closed by preflight, summarizers, cycles, escalations — **assembled by Bash and bookkeeping, never by reasoning** (§41). Three bloat signals (small change >7 specialist executions · low-risk >3 domain reviewers · a revision invoking more than the original without increased risk) prompt inspection; the decision stays semantic.

**Duplicate normalization (§33):** same file + location + root cause + required change = **one** finding, with `supported_by` listing each reviewer's ID. Multiple votes strengthen evidence; they never multiply fixes.

### O10 — Model retiers

*Context: L13, §40. Owned by a fleet pass over the agent files; recorded here as the ratified target.* **Amended by O19 (28 August 2026)**, which narrows the semantic-reviewer exclusion below to §40's own classification. Read the two together; where they differ, O19 is the live rule.

To `sonnet`: `@craft-review-summarizer`, `@go-review-summarizer`, `@linus-review-summarizer`, `@scrum-master`, `@state-reporter`, `@validation-agent`. **Every semantic reviewer is unchanged** — a narrow lane is not a cheap judgement, and downgrading a critic to save tokens is exactly the trade §50 forbids.

**§32 outcome-based tuning is explicitly deferred** until real PR data exists. Tuning activation policy from zero merged PRs would encode a guess as evidence, and the metric when the data arrives is not how often a reviewer passes but **whether its findings were unique and outcome-changing**.

### O11 — The minimal-patch revision law

*Context: §12, §13. In `DELIVERY.md § The minimal-patch revision law`; `@worker-manager` states it in every revision dispatch.*

A revision changes the smallest semantic blast radius that resolves the named finding, and the specialist asks the one-file question before touching an additional file. Unrelated cleanup becomes `future_work`. The reason is mechanical: extra changed surface meets `Invalidated by` conditions and re-convenes reviewers whose verdicts would otherwise have carried.

### O12 — The already-decided check

*Context: §22, §23.*

A fourth pre-invocation question in `agent-handoffs § Before you invoke anything`, and one line in `DELIVERY.md § Escalation and human judgement`: before reasoning about an ambiguity, search `DECISIONS.md`, the ADRs, the requirement, and the board or PR record — **reuse, do not re-litigate.**

### O13 — Scoped board operations

*Context: L16, §45. Owned by a fleet pass over `turfgps-board-ops`; recorded here as ratified.*

Board reads are scoped and filtered rather than dumped: the coordinator receives items relevant to scheduling, `@worker-manager` one assigned item, `@pr-judge` one PR. `fingerprint.sh` already applies this to its own board component, carrying IDs and statuses only.

### O14 — Delivery vehicle

This wave lands on **PR #61's branch, `refactor/agent-org-v2`** — one constitution, one merge decision, rather than a constitutional change split across two reviews.

## Consequences

**What this obliges:**

- **Two shell scripts are now load-bearing.** `pr-judge` Phase 0, `engineering-lead`'s cadence, and `change-risk-assessor`'s intake all begin with one. A script that fails silently would degrade to the old behaviour at best, so both report their own degradation rather than returning a confident wrong answer — `fingerprint.sh` exits 2 and says which component it could not read.
- **The registry is stricter and therefore more fragile.** Narrower rows mean a genuinely relevant reviewer can now be missed by a row that does not quite match. `N/A` verdicts and missed findings are the evidence to watch, and the ledger is where they will show.
- **A fleet pass is owed** to strip `review-board-dispatch` from the reviewers that no longer need it, apply O10's retiers, and scope `turfgps-board-ops`. Until it runs, those files are unoptimized but not wrong.

**What this gives up:**

- **Ceremonial coverage that occasionally caught something.** A reviewer convened on a broad rule sometimes found a defect outside the reason it was called. Narrower activation loses some of those, exactly as ADR-0001's selective convening did, and for the same recorded wager.
- **The summarizer's smoothing.** Four verdicts read directly by the judge is four voices rather than one, and the judge now does that reconciliation itself.

**Reversibility: high.** The registry is a table, the thresholds are numbers, the scripts are two files that can simply stop being called, and the finding vocabulary is five words. If defect escape rises, the ledger records which lanes were closed and by what — which is the reason the accounting footer counts lanes closed by preflight separately from reviewers carried.

## Amendment — 2026-08-13 (runtime hardening)

*Source: the Owner's runtime-hardening directive, deliberately not filed as a separate document — its clarifications are recorded in the existing ADRs. Corrections to O6 as built; the decision is unchanged.*

### O6.1 — The consumer argument is mandatory, and a dispatched agent does not re-gate

Per-consumer state existed from the start; **every call site omitted the argument**, so all four gated agents shared the default `session` file and the first to read it consumed the signal for the rest — precisely the bug the argument was added to prevent. The rule now has two halves and both are stated wherever the gate appears: an agent waking **autonomously** self-gates on `fingerprint.sh <its-own-agent-name>`; an agent dispatched with an explicit **`trigger:` block** (`agent-handoffs § The trigger block`) processes that trigger and does **not** re-run the gate, because an event detector whose dispatch is rejected by its own already-consumed signal detects nothing. Isolation is now asserted mechanically by `scripts/loop/tests/fingerprint-isolation.sh`.

### O6.2 — Three component corrections

O6's component list stands; what each component *reads* was wrong in ways that either woke agents for nothing or hid change from them.

- **`pr` fingerprints `number,headRefOid,isDraft`** — `updatedAt` is gone. A comment, a label, or any bot bumped it and woke @pr-judge on an unmoved diff. What remains are the events the loop reacts to: a head SHA moving, list membership changing, and draft→ready.
- **`board` passes an explicit `--limit 200`.** `gh project item-list` defaults to 30 against a 37+-item board, so the fingerprint covered a *prefix* of it — items past the cut could change status forever without registering.
- **`main` and `corpus` both derive from `origin/main` after one `git fetch`.** `main` came from the remote while `corpus` came from local history, so on a stale checkout a requirements or ADR change that had landed remotely read as no change at all. A failed fetch degrades explicitly rather than answering from a stale tracking ref, because an unreachable remote must never read as a quiet loop.

## Amendment — 2026-08-16 (first live loop cycle)

*Source: the Owner's runtime-findings directive, recorded in the existing ADRs rather than filed separately. Three defects **observed** in the loop's first live cycle. O1–O14 stand unchanged: two of these are costs the wave did not anticipate, and the third is the first piece of the counter-evidence this wave asked for by name.*

### O15 — Confirm a panel is not already running before couriering a reviewer into it

*Observed once, on **PR #67**, and it was the convener's own defect: `@pr-judge` dispatched a second `@docs-reviewer` into the panel it had already convened, producing duplicate verdicts. It recorded the duplicate against itself three times — in the judgment ("that was **my** duplicate dispatch"), in the bloat signals ("a judge-process defect"), and in the ledger row.*

**The rule the incident earns reaches the courier, not only the convener.** `@engineering-lead` holds the Agent tool a judge may lack, so it will sometimes courier a reviewer on the judge's behalf. That does not make it the convener — **the panel is `@pr-judge`'s and the ledger is its record** — so before couriering it reads the ledger and the judge's envelope to establish whether the lane is already convened, already carried, or closed by the preflight. Two rows for one lane at one SHA is not coverage: it is a duplicate the judge must reconcile, and it inflates the accounting footer in the direction that looks like thoroughness. The observed duplicate came from inside the panel; the courier route reaches the same duplicate from outside it, and is the half nothing yet checked.

**A duplicate dispatch is a bloat signal to record, not to absorb** — whichever seat produced it. Stated in `engineering-lead § Before you invoke anything`, and named in that agent's graph-bloat list at `engineering-lead § Phase 2`. A bloat register the orchestrator can only ever point at the judge with is one it has quietly exempted itself from.

### O16 — A gate command is `@validation-agent`'s, and a reviewer's lapse is noted rather than fatal

*Observed once. A `@docs-reviewer` ran gate commands, exceeding its dispatch.* What a reviewer needing a measurement does instead is `review-verdicts § What the obligation reaches`, applied at dispatch by `review-board-dispatch § Read-only is not the whole of the boundary`; this record adds no third statement of it.

**The judge's disposition is recorded alongside it, because the risk here runs both ways.** In the observed case the tree verified clean and the measurements proved load-bearing, and the judge **noted the lapse rather than invalidating the verdict** — which is correct, and is now written down so the next judge does not over-correct into discarding one. The validity table asks whether a verdict is *evidenced*; discarding one whose evidence held buys nothing and costs a full re-review. **A tree that moved remains the separate and fatal failure.** In `review-board-dispatch § Read-only is not the whole of the boundary`.

### O17 — A document that decides a boundary is not docs-only

*Observed once, and it is precisely the counter-evidence this wave's Consequences named as the thing to watch: "a genuinely relevant reviewer can now be missed by a row that does not quite match."*

`@linus-security-critic`'s row read `Never when: docs-only.` A judge crossed it deliberately on **PR #67**, recorded why on the PR, and the reviewer returned **two high findings on paths to stored personal data**. The row is amended: **a docs change that decides one of the security surfaces the row's `Activate when` column already names is not docs-only.** That column is the whole of the bound and this record does not copy it — the amendment reaches what a document *decides* about a surface already listed there, and #67's evidence (stored personal data) sits squarely inside it. The exclusion survives for what it was written for — a typo, a heading, a rewording that settles nothing — and the test is what the document *decides*, not what it is named. In `review-board-dispatch § The reviewer registry`.

**Extended 18 August 2026, on this record's own evidence gap.** O17 makes the `Activate when` column the whole of the bound, and PR #110's security review found that column naming neither **inbound network exposure / deployment reachability** nor **supply chain / build provenance** — so a document deciding either fell outside the lane this record had just widened. PR #67 is O17's entire evidentiary basis and its own `SEC-02` sits in that gap: the reverse-proxy exposure invariant, Valhalla's HTTP daemon not binding loopback, is ingress, where `external requests` is egress. **`supply chain / build provenance` is earned by ownership rather than by an incident, and is recorded that way rather than given a #67 story it does not have:** `linus-security-critic § Core Identity` names the third-party supply chain among the surfaces the lane guards, `linus-security-critic § Attribute Ownership` puts it in the Security row's own list of what the lane checks, and `linus-security-critic § Contract` both sweeps it as a responsibility and carries the `govulncheck` verification action — none of the three touched by this diff, so the column now names a surface the lane already owned. Both surfaces are now named in the column. **No disjunct was restored** — the column remains a closed enumeration, and this record still does not copy it.

**This widens a row O2 narrowed, on the evidence O2 asked for, and it is not a reversal of the wave.** `docs_only` remains an exact deterministic fact about file extensions and O6's §50 guard is untouched. What it stops being is an automatic answer about the **security lane** — which was never a file-extension question, and which §50 already reserved for the assessor and the judge on exactly this reasoning.

## Amendment — 2026-08-28 (measured session)

*Source: issue #128, which measured one session of live operation end to end — ~12.7M subagent tokens, `@pr-judge` ~28% of them and the convened reviewers ~23%. The measurement's finding is that the cost sits in what agents **read** and how often they **re-do**, not in what they write: 35 judgment comments totalled ~120k tokens, about 2% of the runs that produced them. O1–O9 and O11–O17 stand; **O10 is amended by O19 below**, on the revisit condition O10 itself named. This amendment adds the one rule the distribution asks for by name, and settles the model tiers the same session made decidable.*

### O18 — Cycle inflation has a floor

At cycle 3 or later, a new `low`-severity finding in text the previous cycle created resolves to **`future_work` by default**, and an override states what makes that instance different. `medium`+ severity, safety paths, security findings, and false statements of fact are outside it entirely. The law is `DELIVERY.md § The cycle-inflation rule`, which is where it is ratified and where its bounds are settled; `pr-judge § Phase 8` is where it is applied.

**Ratified because three records converged without conferring, and they do not all come from the same seat** — PR #67 cycle 2, which is `@linus-security-critic`'s **verdict** persisted under *"no ruling issued"* rather than a judge's ledger (eight new findings, every one attached to prose cycle 1 did not contain), and PR #120 cycle 2's **judge ledger** (four of five new findings created by discharging cycle 1); then PR #110's final ruling, which established mechanism rather than correlation: a **compliant** execution of the revision packet mechanically produced the next finding. That is a loop whose input is its own output, and it stops on the budget rather than on the work.

**The cost is why it carries a number.** A revision cycle measured at roughly **570,000 tokens**; the session #128 measured would have avoided three to four of them. A rule with a number attached is applied, and one without it is weighed against whatever the last reviewer noticed.

**What this gives up, stated plainly:** some real `low` findings will merge unfixed, carried as `future_work` issues instead. That is the wager — the same one `future_work` was created to make in O8 — and the counter-evidence to watch for is a `future_work` backlog that grows without ever being worked, which would mean the class is being used to close cycles rather than to defer them.

### O19 — O10's semantic-reviewer exclusion is narrowed to the directive it implements

*Amends O10. Filed as a blocker by `@linus-architecture-critic` on PR #129 cycle 1: the diff retiered nine semantic reviewers while O10 read* **"Every semantic reviewer is unchanged."** *Both could not stand. Resolved by amending O10 rather than withdrawing the retiers, on two grounds — and the amendment costs two of them.*

**O10 went beyond the directive it implements.** §40 sorts roles by the shape of their work and names **"narrow reviewers"** explicitly among those that *"may use a cheaper one."* O10 read that as a blanket exclusion of every reviewer. That is stricter than the directive, so this narrows an over-broad implementation rather than reversing an Owner position — §40's carve-outs for *"difficult architecture"* and the high-reasoning roles are untouched, and it is those, not the word *reviewer*, that O10 was protecting.

**O10 named its own revisit condition, and the data now exists.** It deferred §32 outcome-based tuning *"until real PR data exists"* and fixed the metric in advance: **not how often a reviewer passes, but whether its findings were unique and outcome-changing.** Merged PRs with recorded ledgers now exist. The condition O10 set is met rather than circumvented, and applying the metric O10 chose is the amendment it asked for.

**The metric is applied in both directions, because it cut both ways. Two retiers are withdrawn.**

- **`@maintainability-reviewer` returns to `opus`.** On PR #106 its `MAINT-06` reached the `Makefile` false pass — a report line derived from nothing the recipes produced — independently of `@validation-agent`'s `VAL-01` blocker; together they carried `CORE-01`, the `required_change` that PR turned on. `MAINT-01` and `MAINT-02` carried two more at `high`. **Recorded honestly: `MAINT-06` was not unique, and the metric asks for both.** It counts anyway, because a semantic lane arriving at a machine lane's blocker by its own route is judgement rather than duplication, and the class it caught is a **false pass** — the class a cheaper model is likeliest to wave through.
- **`@worker-manager` returns to `opus`.** The retier held that its Phases 2–4 are checklist-shaped — a roster, a written order, two named commands, mechanical verification — and that the residual risk, **negative routing**, is contained by three mandatory floors. `@linus-architecture-critic` dismantled that and is right: **all three floors are predicates over the diff, and a lane that never runs produces an absence from the diff rather than a signal in it.** Worse, the high-tier floor **inverts** — under-routing yields a smaller diff, which yields a lower tier, which yields a smaller panel. The containment is weakened by the very failure it exists to hold. The uncovered risk sits in Phase 1 and the dormancy call, which the justification never reached.

**Eleven retiers stand, and their basis is the classification, not outcome data.** Eight are narrow reviewer lanes with an explicit checklist — `@code-smell-reviewer`, `@design-reviewer`, `@evolvability-reviewer`, `@go-structure-critic`, `@modularity-reviewer`, `@performance-reviewer`, `@scalability-reviewer`, `@ux-reviewer` — and three are structured filing roles: `@project-coordinator`, `@requirements-librarian`, `@requirements-reconciler`. **That is §40's classification and nothing more. Most have never been convened, so no outcome data exists for them and none is claimed** — asserting otherwise would be exactly the guess O10 refused to encode. **O10's metric governs them in both directions once data arrives:** a lane whose findings prove unique and outcome-changing returns to `opus` on the same evidence that just returned two, and the ledger is where that will show.

**Counter-evidence to watch:** a `sonnet` lane returning `pass` on a diff a later cycle or a merged defect proves it should not have. Reversing one is a row in the ledger and a line of frontmatter.
