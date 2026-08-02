# TurfGPS requirements

The front door to the requirements corpus. Every functional and non-functional requirement lives in exactly one category file in this folder, in the canonical record form defined by `.claude/skills/requirements-authoring/SKILL.md`. This page holds the three things that change rarely and that everything else in the folder depends on: the **category register**, the **ID allocation ledger**, and the folder's **invariants**.

Why a folder rather than a file: `docs/README.md` § Requirements/.

**Corpus state — 2 August 2026.** 18 requirements, 5 categories in use. **All eighteen are `to-build`.** `FR-001` … `FR-012` were signed off by the Owner on 1 August 2026 and `FR-013` … `FR-018` on 2 August 2026; each sign-off moved every record it covered from `draft` **straight to `to-build`**, because `@requirements-reconciler` is dormant and there is no classification step for a record to wait in `approved` for — no record in this corpus has ever held `approved`. Story creation has since run for all eighteen: `@requirements-story-organizer` filed four epics and seventeen stories on 2 August 2026, in two passes, one per sign-off, and every requirement in the corpus is resolved by at least one story. The links live in `TRACEABILITY.md`; each record's `Resolved-by` and the column in `INDEX.md` are views of them.

**`FR-013` … `FR-018` are batch 2, filed 2 August 2026 and signed off the same day.** All six read `Status: to-build`, and all six now carry a story: story creation ran for them on 2 August 2026 as a pass of its own following sign-off, exactly as it did for batch 1, under a fourth epic — `A real choice of journey alternatives`. No record in the corpus reads `Resolved-by: —` today. The em-dash keeps its meaning for the next batch filed: *not yet allocated*, never *unknown*, and never an orphan — the orphan test applies to a requirement whose epic is in flight.

## The files in this folder

| File | Holds |
|---|---|
| `README.md` | This page: front door, category register, ID allocation ledger, folder invariants |
| `INDEX.md` | One index table per category, carrying a row for every ID ever issued |
| `journey-definition.md` | The `Journey definition` records — one file per category, named from the register below |
| `cost-and-time-composition.md` | The `Cost and time composition` records |
| `value-model.md` | The `Value model` records |
| `objective-selection-and-ranking.md` | The `Objective selection and ranking` records |
| `recommendation-set-composition.md` | The `Recommendation set composition` records |
| `TRACEABILITY.md` | The requirement ↔ epic ↔ story matrix, both directions |

**Why the index is a separate file.** The allocation ledger is the only guard against ID reuse, and it must not sit below fifteen screens of index table where an author stops scrolling. The register and the ledger change rarely; the index changes on every filing. Keeping them apart means a routine filing does not rewrite the one file the whole RE family reads first, and two agents filing different categories do not serialize against a single file.

## Category register

The corpus's controlled `Category` vocabulary, and its only home. A `Category` value is legal if and only if it appears in the **Category** column below, copied verbatim. `@requirements-engineer` alone adds a name here — a new category is a decision, never a side effect of authoring. `@requirements-librarian` files each record under its `Category` and raises a record carrying a name not on this list as a finding, rather than filing it under a near-match.

**A name on this list is permission to file, not a promise that requirements exist.** Most rows read `reserved`: the source sections behind them have not been broken down yet. `reserved` means no file and no records, and nothing may be concluded from the absence — a category's absence from the *filed* set means unfiled, never illegal.

These names were consolidated on 1 August 2026 from coverage prompts that previously stood in four agent definitions, which named the same areas in four different phrasings and would have produced several files for one subsystem. Those lists remain in place as coverage prompts — questions an author asks per source section — and are explicitly no longer a source of category names.

**Extensions since.** `Value model`, added 1 August 2026 by `@requirements-engineer` while integrating `SPECIFICATION.md § The journey as an optimization problem`. The register as consolidated carried `Cost and time composition` with no counterpart on the value side, while the objective the whole product optimizes is value set against cost. Without it, everything the value half of that objective is built from — zone attributes, attribute rarity and preference, the rank-to-weight curve — would have filed under `Objective selection and ranking`, whose scope is how value and cost are *combined*, not what either is made of. Recorded here because a register extension is a decision, and a decision with no stated reason is indistinguishable later from a name someone coined in passing.

`Recommendation set composition`, added 2 August 2026 by `@requirements-engineer` while integrating `SPECIFICATION.md § Recommended journey alternatives`. The register carried `Route alternative generation` for producing corridors and `Objective selection and ranking` for the objective and the ordering under it, and nothing for the question that section actually answers: which of the alternatives produced are put in front of the user. Filing those records under `Objective selection and ranking` would have put set membership and set ordering in one file under a scope line that names only the second, and coverage per subsystem would have stopped being checkable for either. Recorded here because a register extension is a decision, and a decision with no stated reason is indistinguishable later from a name someone coined in passing.

| Category | Kind | File | Scope | State |
|---|---|---|---|---|
| Journey definition | FR | `journey-definition.md` | The destinations a request must and may name, the order they are visited in, their preservation through optimization, and the mode of travel between them | **in use** |
| Journey initialization | FR | — | How a journey is entered and a planning session begins: input capture, the time budget, validation, session start. **What a journey must and may name is `Journey definition`, not this** — the split is the act of entering against the thing entered | reserved |
| Route alternative generation | FR | — | Producing the candidate corridors a journey may take | reserved |
| Candidate identification | FR | — | Selecting which zones enter evaluation, and bounding that set | reserved |
| Access classification | FR | — | Deciding whether and how a zone can be reached and stopped at | reserved |
| Cost and time composition | FR | `cost-and-time-composition.md` | Composing stop cost and journey cost from routing, walking and manoeuvre components | **in use** |
| Value model | FR | `value-model.md` | What a journey's Turf value is made of: the zones counted toward it, the attributes that weight them, and the user's stated preferences over those attributes. The mirror of `Cost and time composition`; the two are combined by `Objective selection and ranking`, which owns neither | **in use** |
| Objective selection and ranking | FR | `objective-selection-and-ranking.md` | The objective a plan optimizes, and the ordering of alternatives under it | **in use** |
| Recommendation set composition | FR | `recommendation-set-composition.md` | Which journey alternatives reach the user and what the offered set must hold: how many, how varied, what is dropped as redundant or dominated, and what may never be withheld. Distinct from `Route alternative generation`, which produces the corridors, and from `Objective selection and ranking`, which decides the objective and the order — this category decides membership of the offered set, not its ranking | **in use** |
| Review and replacement | FR | — | The user's review loop: rejecting a zone and receiving a replacement | reserved |
| Hand-off and dispatch | FR | — | Delivering a chosen plan to the device and the navigation app that will drive it | reserved |
| Persistence and staleness | FR | — | Storing a plan, retrieving it, and what makes a stored plan no longer trustworthy | reserved |
| Safety exclusions | FR | — | The enforceable exclusions and the absolute ceiling, as behaviour the system must exhibit | reserved |
| Response time | NFR | — | How long the initial solve and the review-time replacement may take | reserved |
| Estimate accuracy | NFR | — | Calibration of presented estimates, and the width of the ranges carrying them | reserved |
| Coverage and data quality | NFR | — | Graceful degradation where map or elevation data is thin, and confidence falling with it | reserved |
| Classification correctness | NFR | — | The product's stated measure of success: no zone classified confidently and wrongly | reserved |
| Call budget | NFR | — | Bounded and known per-journey external call volume | reserved |
| Privacy | NFR | — | What a stored plan holds, and what the system declines to hold at all | reserved |
| Portability | NFR | — | Adding a country as an adapter rather than a rewrite | reserved |
| Platform support | NFR | — | Mobile-first web: desktop and mobile, iOS and Android browsers | reserved |
| Observability | NFR | — | What must be measurable about a running system, including the decisions the design earmarks for measurement | reserved |
| Maintainability | NFR | — | The cost of the next safe change, including the one-home-per-model invariant | reserved |

A category file is created the first time a record is filed under its name, and the **File** column changes from `—` to the filename in the same pass.

## ID allocation ledger

**Next free: FR-019 · NFR-001**

**Allocation is highest-ever-issued plus one, never lowest-unused**, so a retirement cannot free a number even if a tombstone is lost. One ascending sequence per prefix, the two counters running independently.

**No NFR has ever been issued.** `@requirements-nfr` returned a considered zero for batch 1, scoped to `SPECIFICATION.md § Primary use case` and its subsection *The planning player* — neither states a quality attribute with an obligation attached — a second considered zero, argued and accepted, for `SPECIFICATION.md § The journey as an optimization problem`, the section behind `FR-007` … `FR-011`, and a third for `SPECIFICATION.md § Recommended journey alternatives`, the section behind `FR-013` … `FR-018`, argued at the foot of this ledger. `NFR-001` is therefore free, not missing, and the first NFR filed will take it. An empty NFR sequence here is a recorded outcome, not a gap.

**What no zero covered, recorded so the emptiness above is not over-read.** `SPECIFICATION.md § Genuinely out of reach or out of scope` is cited by `FR-005` and is `FR-012`'s sole source, but it entered batch 1 only as a corroborating second citation and was never swept as a section: five of its six bullets carry no requirement at all, and `FR-012` arrived from an Owner ruling rather than from a batch scoped to it. No NFR question has been put to it, and the absence of one there is **unexamined rather than a considered zero of its own**. That section is owed a batch of its own — an earlier wording of this paragraph described batch 1 as covering "the two source sections behind `FR-001` … `FR-006`", which counted a citation as a sweep and is corrected here.

Owed for the same reason, and noticed while `FR-012` was authored: *The planning player* states that the tool is used in a planning session rather than while driving, "which relaxes how quickly the first answer must appear". A relaxed target is still a target — but the obligation's home is `Architecture.md § Response time and progressive results`, not the section observing it, which is the same reasoning that produced the second zero below. **A `Response time` NFR is owed when a batch scoped to that Architecture section runs**, and is deliberately not allocated here.

**`SPECIFICATION.md § User time constraints` is owed a batch of its own, and this debt is not merely an unswept section.** Batch 2 actively deferred obligations to it: `FR-013`, `FR-017` and `FR-018` each disclaim in their `Rationale` an obligation that section owns — that there must always be at least one journey alternative within the user's stated additional-time limit — naming it as the single definition of the allowance rather than restating it. The consequence has to be written down plainly, because it cannot be seen from the records themselves: **until that batch runs, no record in the corpus obliges the offered set to be non-empty**, and every batch-2 record is vacuously satisfiable by a system that offers nothing at all. Owed to the same batch and deferred for the same reason: the allowance itself, the absolute ceiling over it, the value-based justification a stretch alternative requires, and a stretch alternative's identification as one.

**A second hole of the same shape, and it is not the one above.** Found by `@requirements-story-organizer` while cutting batch 2's stories, and recorded here because it cannot be seen from any record on its own. The debt above is that the offered set may be **empty**; this one is that it may be **singular**. Every batch-2 criterion is conditioned on alternatives *having been produced* — `FR-013`'s first criterion opens *given a journey for which more than one alternative … has been produced* — and **nothing anywhere in the corpus obliges the system to produce more than one alternative in the first place**. A search that only ever yields one satisfies `FR-013` without ever putting a choice in front of a user, which is the whole of what that record exists to secure.

The obligation's home is `Route alternative generation`, which is `reserved` on the register above: no file, no records, and the section behind it unswept. **It is owed to the batch scoped there**, and it is deliberately not authored now — writing it here would be admitting scope through the side door that story creation opened, and the register entry exists precisely so it is not. Recorded rather than storied, and carried on `#12` as an openly-stated hole so no worker reads that story as guaranteeing a plural result.

**Why the second zero, recorded so it is not re-litigated.** That section names the one-home-per-model invariant while complying with it — the scoring model is defined once, in `CalculationSpecification.md`, and not restated. The obligation is created by `docs/README.md` § Conventions, not by the section observing it, so an NFR sourced from the observation would cite the wrong home; and narrowing it to the scoring model alone would shard a corpus-wide invariant per occurrence, giving the one-home rule more than one home — the exact defect it exists to prevent. **One corpus-wide `Maintainability` NFR, verified by `inspection`, is owed when a batch scoped to `docs/README.md § Conventions` runs.** It is deliberately not allocated here.

**Third considered zero — `SPECIFICATION.md § Recommended journey alternatives`, 2 August 2026.** The section states one obligation of its own — curate the offered set, avoiding effective duplicates and alternatives with no meaningful trade-off — and everything quality-*sounding* in it is the standard by which that one duty is judged, not a second duty. "Meaningfully different", "effectively duplicates" and "meaningful trade-off" are the adverbs of the curation rule; the corpus's home for a standard of judgement is the `Verification` field of the record carrying the duty, never a companion record. An NFR restating them would be the same sentence filed twice, split only by verification method, and would collide head-on with the FR sweep of this same section. One obligation, one home.

No threshold exists to hang a metric on. "Several" is not a count, and nothing in `CalculationSpecification.md`, `Architecture.md` or `DESIGN.md` sets a number of alternatives to present or a distinctness bar between them. The nearest thing in the corpus — "a route alternative is not merely a variation caused by visiting one additional zone", `SPECIFICATION.md § General route alternatives` — is a *definition* rather than a threshold, and it governs the general-road-route layer, not the Turf-strategy layer this section governs. Any number filed here would have been invented.

The fallback is not `Coverage and data quality`, and this is the trap the section sets. That category degrades on **thin map or elevation data**, with confidence falling as it thins. This section degrades on **zone availability** — there genuinely is no attributed zone inside the user's limit — which is a fact about the world, not a deficiency in our data; the system knows it perfectly well and its confidence is unaffected. Filing it under that category would give "graceful degradation" two triggers and leave the later data-thinness batch working around a record that had already claimed the name. What the fallback obliges is an action in a branch, with an unhappy path and no quality target attached. That is functional, and it is `FR-018`.

The allowance, the absolute ceiling over it and the at-least-one-compliant guarantee are all sourced from `SPECIFICATION.md § User time constraints`, which states it is their single definition; they are owed to a batch scoped there. No `Response time` question was put to this section — that NFR is already owed to `Architecture.md § Response time and progressive results` and is not pre-empted here. An `Observability` record on how often the fallback fires was considered and rejected: no document asks for it, and authoring it would be inventing scope rather than reading it.

## How this folder works

- **IDs are immutable.** Never reused, never renumbered, never recycled after deletion. Whoever issues the next one takes it from the ledger above and nowhere else.
- **`INDEX.md` carries a row for every ID ever issued**, including retired ones. A retired requirement keeps its record, in place and in ID order, reading `Status: retired`; it is never deleted and never moved. There are no tombstones today because no ID has ever been retired — `FR-001` … `FR-018` is contiguous with no gaps.
- **A retired record's `Resolved-by` freezes at its last value.** Retirement removes the requirement from the *Requirement → story* table in `TRACEABILITY.md`, so the link's single home disappears with it — but the field on the record and the column in `INDEX.md` keep whatever they last read rather than reverting to `—`. They are then the only surviving record of what was built for that requirement, and blanking them would erase it. A frozen value is not stale and is not a finding.
- **Records sit in ID ascending order inside their file, and that order never changes** — not on retirement, not on re-prioritization. A file's order is therefore a fact about IDs alone, and a diff never moves a record it did not touch.
- **The record is authoritative; every index is derived.** Where a row in `INDEX.md` and the record in a category file disagree, the record is right and the row is stale — a librarian finding, fixed in the index, never fixed by editing the record to match.
- **Records are filed verbatim inside fenced blocks**, preserving the field alignment of the canonical record in `requirements-authoring`. The librarian moves, groups, indexes and cross-links records; it never re-words one. Content questions go to `@requirements-engineer`.
- **Category files are named as the kebab-cased `Category` field**, verbatim and unnumbered, so a record names its file and a file names its category with nothing to look up, and adding a category never renames one.
