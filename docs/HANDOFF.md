# Session handoff

Written 31 July 2026 to carry context between machines. Read this first if you are starting fresh on this project.

---

## 1. What this project is

**TurfGPS** is a route-planning and decision-support system for players of **Turf**, the Swedish GPS location game. A user is already driving somewhere — Örebro to Jönköping, say — and wants to know which roads and which individual zones make the best Turf journey for the extra time they can spare.

It is not a zone map and not a navigation app. It is an optimizer with an explanation layer: it compares whole journeys, prices the real cost of each stop, and tells the user *why* one route beats another.

**The documentation leads the implementation.** The repository is a near-empty Next.js skeleton; almost nothing described in the docs exists as code. `docs/` is the source of truth for intent, not the source tree.

## 2. Read these, in this order

1. **`docs/README.md`** — the front door. Explains the six-document structure, what question each answers, and the current state of the folder.
2. **`docs/Concept.md`** — roughly 1,670 lines, the authoritative product definition. Purpose, optimization framing, accessibility classification, value model, user flow, safety rules, boundaries, open questions.
3. **`docs/DELIVERY.md`** — how work is tracked and reviewed.

`docs/Architecture.md` and `docs/CalculationSpecification.md` are outline stubs.

Do not skim `Concept.md`. It is long because it settles a great many decisions, and re-deciding one that is already settled wastes the work that produced it.

## 3. Facts that must not be guessed

Domain and API facts in `Concept.md` were **verified against the live Turf API**, not assumed. Several contradict the obvious guess, and each one was originally gotten wrong by reasoning rather than checking:

- **`blocktime` is not takeover time.** It is how long a zone stays locked after being taken. Takeover time follows a published formula: `30 − (0.2 × rank)` seconds.
- **The Region Lord bonus is global.** Holding any region grants it everywhere, so takeover time is one per-journey constant, not per-zone.
- **Zones expose only a coordinate.** No radius, no shape. The system models no capture area and treats the coordinate as the target.
- **`currentOwner` and `dateLastTaken` are round-scoped.** Their absence means "not taken this round", never "never taken".
- **`POST /v5/users` returns the user's currently-held zone ids**, so the ownership indicator needs no per-zone lookup and no cache.
- **The bbox limit is an area product**, `Δlat × Δlng > 0.05`, not a per-axis limit.

The API self-documents at `GET https://api.turfgame.com/v5`. **Verify against it rather than inferring.** Every assumption made without checking during the original work turned out wrong.

**Numeric constants in `Concept.md` are proposals unless stated otherwise.** They exist so specification work starts with a concrete number to argue against rather than a blank to fill. Do not cite one as measured.

## 4. Save these to memory

Memory files do not travel between machines, so they must be recreated. **Save what the repository does not already record** — do not duplicate the docs into memory.

Worth saving:

- **A pointer**, not a copy: `docs/README.md` is the front door and `docs/Concept.md` is the source of truth for intent; the repo's code does not reflect the design.
- **The verified-facts warning** from section 3, specifically that inference about Turf mechanics has a poor track record and the live API should be consulted.
- **The user's working preferences**, in section 5 below. These are not in the repository and are the most easily lost.
- **The current step**, section 6.

Not worth saving: the docs structure and the delivery model, both of which are now committed as `docs/README.md` and `docs/DELIVERY.md`.

## 5. How this user works

Learned over a long design session, and worth respecting:

- **Raise concerns explicitly, as interview questions.** When something is ambiguous or contradictory, ask rather than deciding quietly. This was requested directly: *"If you still are unsure or have a concern in any of the content of the document, raise it as a concern in forms of interview questions to me. That way we solve ambiguity."*
- **Propose an answer alongside every question.** They asked for proposed defaults, not open-ended questions. A question with a recommendation is useful; a question without one is work handed back.
- **Prefer concrete numbers to blanks.** Mark them as proposals, but supply them.
- **Write decisions into the documents**, not just into replies — *"so that we have a repo-wide agreement, regardless who views it."*
- **They correct domain facts directly and often.** Take corrections at face value, update the docs, and do not over-apologise. State confidence honestly so they know what to check.
- **They use multi-agent review heavily** and value adversarial verification over a single confident pass. Two review rounds during the concept work found real defects that a single pass missed.

## 6. Where the work stands

All documentation is merged to `main` at commit `5677603`. The working tree is clean.

**The next step is splitting `docs/Concept.md`** into the structure described in `docs/README.md`. It has grown past its purpose — it now contains a calculation specification, an architecture directive, and a UX specification alongside the product concept.

Two facts govern the split:

**It is a move operation, not authoring.** `Concept.md` explicitly forbids restating its own formulas — several sections say a model is stated *only* there. Copying breaks the anti-duplication rules the document depends on.

**The calculation specification is already written**, inside `Concept.md`, in finished form with numeric defaults. Roughly 20 of its 21 formula blocks are calculation content. Whoever fills that stub is relocating text, not composing it.

### What moves where

Determined by a full audit of the document.

**To `CalculationSpecification.md`** — elevation-aware walking time and the canonical stop-time model; the flat-distance fallback; direct road-access calculation; speed and manoeuvre calculations including the placeholder timing table; the rank-to-takeover-time rule and the default when rank is unknown; the proposed rank-to-weight curve and its table; the activity formulas, difficulty multiplier and three guards; *Proposed form: value per minute* as the canonical objective; the corridor figures (1 km floor, 15 km cap, 300 candidates per alternative); the 10-metre direct-access tolerance; the conservative upper bound for uncertain stops; and the two interaction thresholds — three rejections within 2 km, and 20% of remaining journey length. Those last two are constants that happen to surface in an interface; they belong here, not in the design document.

**To `Architecture.md`** — all of *Data sources and constraints* **except** two subsections that are product facts and must stay: *The coordinate is the target* and *Rounds*. Also the ownership fetch strategy; the cross-device analysis, expiry policy and personal-data obligation from *No accounts*; and from the non-functional requirements: response time and progressive results, global-data-first, provider adapters, the cost consequence, and the external call budget.

**To a new `DESIGN.md`** — first-run initialization; entering a journey; reviewing zones one at a time; replacement and escalating scope; when replacement runs out; when nothing fits at all; dispatching stop by stop; communicating a round rollover.

**Stays in `SPECIFICATION.md`** — purpose, primary use case, the planning player, the optimization framing as prose, general route alternatives, the individual-zones argument, why attributes matter, the weighting is extreme, optimization objectives and their priority order, optimizer and advisor, the four journey alternatives, the rationale for route review, the waypoint-limit problem, the persistence requirement, stored routes going stale, safety and legality, product boundaries, and product vision.

### Two traps in the split

**The attribute rarity table stays** in the specification despite being a table of numbers. Four separate sections cite it as *evidence* for product decisions rather than using it as formula input. Have the calculation spec reference it rather than take it.

**The zone-activity section splits rather than moves.** The observations stay as concept evidence — median around 26 takes per month, water zones at 1.8, islands and bathing places lowest — because they are the argument for the whole signal. The formulas and guards move.

After the split the specification should be around 700–750 lines: readable front to back in one sitting, which is the point.

### Then: requirements

`Requirements/` is the bottleneck for everything else. Epics and User Stories derive from it, so the GitHub Project board described in `docs/DELIVERY.md` cannot sensibly be built first. `DESIGN.md` and `DEPLOYMENT.md` can lag.

## 7. Still genuinely open

Do not close these silently. `Concept.md` carries a fuller list in its *Open questions* section.

- **Manoeuvre timings** — slowdown and rejoin values. The only constant needing a stopwatch. Placeholders are in place and clearly marked.
- **Direct-access yield has never been counted.** Nobody knows how many zones qualify as directly road-accessible per 100 km of rural corridor. This determines whether the product's core promise holds outside towns, and it is measurable now from the zone data plus road data.
- **The uncertain-bucket share is unknown.** If it is 5% of rural candidates the design works as written; at 60% the review flow becomes mostly reserve-pool negotiation and the product feels different.
- **Self-hosting versus metered APIs at global scope** — the largest cost risk in the project. The provider-adapter pattern makes the choice replaceable, not answered.
- **The lifetime of an unconfirmed route.** Persistence covers confirmed plans; a solved, partly-reviewed route carrying accumulated exclusions has no defined lifetime.
- **The 10-metre direct-access tolerance** is a proposal sitting exactly on the boundary between two cost models, and is worth validating early.

## 8. Do not

- Do not treat a proposed constant as measured.
- Do not decide an open question without asking.
- Do not infer Turf mechanics — check the API.
- Do not read, print, or echo `GH_JUDGE_TOKEN`. It is referenced by name only, and passed to `gh` through the environment.
