# Session handoff

Written 31 July 2026 to carry context between machines. Read this first if you are starting fresh on this project.

**Amended 10 August 2026.** Section 6 is history: the corpus is stocked and the board carries its stories — `docs/README.md § Current state` is the live picture. The agent organization now runs under `ADR-0001` and `DELIVERY.md`: reviewers convene by registry, ordinary ambiguity is decided under a precedence ladder and logged, and only escalation-qualifying questions reach the Owner.

---

## 1. What this project is

**TurfGPS** is a route-planning and decision-support system for players of **Turf**, the Swedish GPS location game. A user is already driving somewhere — Örebro to Jönköping, say — and wants to know which roads and which individual zones make the best Turf journey for the extra time they can spare.

It is not a zone map and not a navigation app. It is an optimizer with an explanation layer: it compares whole journeys, prices the real cost of each stop, and tells the user *why* one route beats another.

**The documentation leads the implementation.** `docs/` is the source of truth for intent, and it runs ahead of the tree rather than describing it. **What the repository actually holds is derived and never asserted here**: `codebase-map § Which map is authoritative — check the tree, do not assume` is the one place that answers it, and it carries both the one-line check and the reason the deleted Next.js prototype is not a partial implementation of this design. This paragraph asserted the answer instead, and was wrong from the merge of PR #106 — where the Go service became reachable on `main` — until 28 August 2026. The span opens on a merge rather than a date because a commit's own date is the day it was written on a branch, not the day `main` acquired it.

## 2. Read these, in this order

1. **`docs/README.md`** — the front door. Explains what each document answers, the current state of the folder, and the conventions the whole set depends on.
2. **`docs/SPECIFICATION.md`** — roughly 1,065 lines, the authoritative product definition. Purpose, optimization framing, accessibility classification, value model, safety rules, boundaries.
3. Then whichever of **`docs/CalculationSpecification.md`**, **`docs/Architecture.md`**, and **`docs/DESIGN.md`** the work touches.
4. **`docs/DELIVERY.md`** — how work is tracked and reviewed.
5. **`docs/adr/`** — the decision records. `ADR-0001` defines how the agent organization runs.

Do not skim `SPECIFICATION.md`. It is long because it settles a great many decisions, and re-deciding one that is already settled wastes the work that produced it.

Every document ends with **what it still owes** and **the open questions it owns**. Those two sections are the fastest route to what is unfinished.

## 3. Facts that must not be guessed

Domain and API facts were **verified against the live Turf API**, not assumed. They are recorded under `Architecture.md § Data sources and constraints`. Several contradict the obvious guess, and each one was originally gotten wrong by reasoning rather than checking:

- **`blocktime` is not takeover time.** It is how long a zone stays locked after being taken. Takeover time follows a published formula: `30 − (0.2 × rank)` seconds.
- **The Region Lord bonus is global.** Holding any region grants it everywhere, so takeover time is one per-journey constant, not per-zone.
- **Zones expose only a coordinate.** No radius, no shape. The system models no capture area and treats the coordinate as the target.
- **`currentOwner` and `dateLastTaken` are round-scoped.** Their absence means "not taken this round", never "never taken".
- **`POST /v5/users` returns the user's currently-held zone ids**, so the ownership indicator needs no per-zone lookup and no cache.
- **The bbox limit is an area product**, `Δlat × Δlng > 0.05`, not a per-axis limit.

The API self-documents at `GET https://api.turfgame.com/v5`. **Verify against it rather than inferring.** Every assumption made without checking during the original work turned out wrong.

**Numeric constants are proposals unless stated otherwise.** They live in `CalculationSpecification.md` and are marked there. They exist so specification work starts with a concrete number to argue against rather than a blank to fill. Do not cite one as measured.

## 4. Save these to memory

Memory files do not travel between machines, so they must be recreated. **Save what the repository does not already record** — do not duplicate the docs into memory.

Worth saving:

- **A pointer**, not a copy: `docs/README.md` is the front door and `docs/SPECIFICATION.md` is the source of truth for intent. What the tree holds is worth saving to no memory at all — it is derived on the spot per `codebase-map § Which map is authoritative — check the tree, do not assume`, and a remembered answer is stale by construction.
- **The verified-facts warning** from section 3, specifically that inference about Turf mechanics has a poor track record and the live API should be consulted.
- **The user's working preferences**, in section 5 below. These are not in the repository and are the most easily lost.
- **The current step**, section 6.
- **`data/zones-dump-2025-01-04.json`** exists on the original machine — a 56 MB full zone dump, gitignored, and the fastest way to answer the measurement questions in section 7. It does not travel with the repository.

Not worth saving: the docs structure, the delivery model, and the split itself, all of which are now committed.

## 5. How this user works

Learned over a long design session, and worth respecting:

- **Raise concerns explicitly, as interview questions.** When something is ambiguous or contradictory, ask rather than deciding quietly. This was requested directly: *"If you still are unsure or have a concern in any of the content of the document, raise it as a concern in forms of interview questions to me. That way we solve ambiguity."*
- **Propose an answer alongside every question.** They asked for proposed defaults, not open-ended questions. A question with a recommendation is useful; a question without one is work handed back.
- **Prefer concrete numbers to blanks.** Mark them as proposals, but supply them.
- **Write decisions into the documents**, not just into replies — *"so that we have a repo-wide agreement, regardless who views it."*
- **They correct domain facts directly and often.** Take corrections at face value, update the docs, and do not over-apologise. State confidence honestly so they know what to check.
- **They use multi-agent review heavily** and value adversarial verification over a single confident pass. Two review rounds during the concept work found real defects that a single pass missed.

## 6. Where the work stands

`Concept.md` has been **split**. It no longer exists; it survives in git history. The move ran on 31 July 2026 and distributed roughly 1,670 lines into four documents:

| Document | Lines | Holds |
|----------|-------|-------|
| `SPECIFICATION.md` | ~1,065 | The product: purpose, behaviour, arguments, safety, boundaries |
| `CalculationSpecification.md` | ~444 | Every formula, constant, and threshold |
| `Architecture.md` | ~346 | Technology decisions, topology, Turf API facts, call budget, persistence |
| `DESIGN.md` | ~251 | The full interaction flow, from wizard to dispatch |

It was a move, not authoring. The only content added was **mermaid diagrams** — seven across the four documents — and the connective sentences needed to keep each document readable where a block had been lifted out of it.

The rule that made the split necessary now governs the set: **a model has exactly one home.** Formulas live in `CalculationSpecification.md` and are referenced elsewhere by section name, never restated. Cross-document references name the document they point at.

### Then: requirements

`Requirements/` is the bottleneck for everything else, and is the next piece of work. Epics and User Stories derive from it, so the GitHub Project board described in `docs/DELIVERY.md` cannot sensibly be built first.

Expect on the order of 150–250 requirements, in a folder rather than a single file, each carrying a **verification method**. Much of this product's quality bar is human judgement rather than anything machine-checkable — whether a recommended route is a *good* Turf route cannot be asserted by a test — and the requirement must say so, or review will claim to have verified something it did not.

`DEPLOYMENT.md` can lag. So can the visual layer of `DESIGN.md`.

## 7. Still genuinely open

Do not close these silently. Each document carries its own *Open questions* section; this is the cross-cutting summary.

- **Manoeuvre timings** — slowdown and rejoin values, owned by `CalculationSpecification.md`. The only constant needing a stopwatch. Placeholders are in place and clearly marked.
- **Direct-access yield has never been counted.** Nobody knows how many zones qualify as directly road-accessible per 100 km of rural corridor. This determines whether the product's core promise holds outside towns, and it is measurable now from the zone data plus road data.
- **The uncertain-bucket share is unknown.** If it is 5% of rural candidates the design works as written; at 60% the review flow becomes mostly reserve-pool negotiation and the product feels different.
- **Self-hosting versus metered APIs at global scope** — the largest cost risk in the project, owned by `Architecture.md`. The provider-adapter pattern makes the choice replaceable, not answered.
- **The lifetime of an unconfirmed route**, owned by `SPECIFICATION.md`, with its architectural half — solve-session lifetime and residency — owned by `Architecture.md`. They must be answered together.
- **The 10-metre direct-access tolerance** is a proposal, and an enforcement constant whose strict direction is downward. It separates two validation regimes, not two cost models: inside it a candidate skips walk-safety validation entirely. It exempts a candidate from nothing outside that pair of branches — the road-class and speed-limit exclusions and the requirement that a stopping position be established still bind. Worth validating early, per `CalculationSpecification.md § Direct-access tolerance`.

## 8. Do not

- Do not treat a proposed constant as measured.
- Do not close a catalogued open question without the Owner. Ordinary ambiguity is decided under the precedence ladder and recorded — `ADR-0001 § D6`.
- Do not infer Turf mechanics — check the API.
- Do not restate a formula outside `CalculationSpecification.md`.
- Do not read, print, or echo `GH_JUDGE_TOKEN`. It is referenced by name only, and passed to `gh` through the environment.
