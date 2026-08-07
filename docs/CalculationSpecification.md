# Calculation specification

Every formula, constant, and threshold behind the numbers TurfGPS puts in front of a user.

This document owns *how the system computes*. It does not restate what the system does or argue why it should — `SPECIFICATION.md` owns that, and the reasoning that justifies these models lives there. Conversely, a model stated here is stated **only** here. Nothing else in the documentation set may restate one, because a second statement of a model is a second thing to keep correct, and the two will diverge.

Where a section of another document is referenced, the document is named. An unqualified italic name refers to a section of this one.

**Status: moved out of `Concept.md` on 31 July 2026, complete for the first release.** The models below are finished and carry proposed defaults throughout. The domain glossary this document also owes is listed under *Still owed by this document*.

---

## Conventions

**Numeric constants are proposals unless stated otherwise**, per `docs/README.md`. A proposed default exists so that work begins with a concrete number to argue against rather than a blank to fill. None of the proposals here is a measured value, and none may be quoted to a user as though it were.

**Every constant is configurable and carries a documented origin**, never embedded as an unexplained literal, per *Estimate accuracy and calibration* in `SPECIFICATION.md`. For an uncalibrated constant that origin is the assumption made and why; for a settled one it is the source and the date it was checked.

**Modelling constants and enforcement constants are not configurable in the same way.** The rule above was written for constants where being wrong degrades the quality of an answer. It does not fit constants where being wrong puts a driver somewhere they did not agree to be, and applying it unchanged to one of those turns a safety limit into a deployment knob.

* A **modelling constant** — the candidate cap, the corridor width, the placeholder timings, the rank-to-weight anchors — fails in the direction of a worse recommendation. It remains configurable in both directions under the rule above.
* An **enforcement constant** — the absolute additional-time ceiling, the direct-access tolerance, the maximum speed limit for a stopping road — fails in the direction of an outcome the user did not agree to. It is **configurable in the strict direction only**: a deployment may tighten it and may never loosen it, and its documented value is the limit of what is permitted rather than a midpoint to tune around. It must be **named as an enforcement constant at its own definition site, together with which direction is the strict one**, so both properties travel with the constant instead of depending on a reader recalling this section.

**The strict direction is not always downward.** Stating this rule arithmetically rather than as strictness inverts it on at least one constant. For the ceiling, strict is lower. For *A conservative upper bound for an uncertain stop*, strict is **higher**: a larger bound refuses more acceptances, so a rule reading "downward only", applied verbatim by a later author, would turn the ceiling's only sensor for an uncertain stop into a rubber stamp. Each definition site names its own strict direction, and no reader may infer one from this section.

The test is not how important a constant looks but **which way its failure runs**: if moving a constant in some direction can produce an outcome the user did not agree to, it is an enforcement constant. That direction is the loose one; the opposite is the strict direction it may be configured in. Anything feeding the enforceable exclusions or the absolute ceiling in `SPECIFICATION.md` is presumed enforcement unless argued otherwise.

The reason for stating this as a general distinction rather than a carve-out is that an enforcement constant filed here silently inherits "every constant is configurable". Every ceiling check would still pass against a loosened value, and every test asserting compliance would still return green, while the user was routed past what they agreed to.

---

## Bounding the candidate set

"Reasonable spatial limits", in *Candidate zone identification* in `SPECIFICATION.md`, must be a number, or the requirement that per-journey call volume be bounded has nothing bounding it. Two mechanisms apply together, and both are **proposed defaults**.

**A corridor scaled to the time budget.** How far a zone can lie from the route and remain worth considering depends on how much extra time the user has allowed. The corridor half-width is derived from the per-leg time allowance — roughly the distance drivable in half the remaining allowance — then **floored at 1 km and capped at 15 km**.

The floor keeps a corridor usable even for a very small budget. The cap is generous on purpose: a highly ranked attribute is rare enough, per *Attribute rarity* in `SPECIFICATION.md`, that a genuinely valuable zone may sit well off the main road, and a tight corridor would never see it.

**A hard cap on candidates promoted to routing.** Corridor membership is cheap; routing and access analysis are not. Zones inside the corridor are scored on straight-line proximity and estimated value, and only the best **300 per route alternative** proceed to full evaluation.

This is the point at which the pipeline's cost becomes bounded and predictable, and it is what makes the wide corridor affordable. Both figures are configuration, and the arithmetic connecting them to the per-journey call volume belongs under *The call budget* in `Architecture.md`.

Where the cap binds, it must not do so silently. If a corridor contains more qualifying candidates than the cap admits, that fact should be recorded, because a user in a dense area is receiving a result shaped by a limit rather than by their preferences.

---

## Direct-access tolerance

**This is an enforcement constant, and its strict direction is downward.**

*Directly road-accessible zones* in `SPECIFICATION.md` covers two situations: a zone sitting directly **on** a drivable road, and a zone sitting **beside** one, close enough that a stopped car is already within it.

The second case needs a stated tolerance, because *The coordinate is the target* deliberately refuses to model the capture area. That rule governs how far a player must **travel**, and it is right to be conservative there. Direct access asks a different question — whether a stopped vehicle is already inside the zone — and answering it requires admitting that the zone has some extent.

The proposed rule is that a zone qualifies as directly road-accessible when its coordinate lies within **10 metres** of a valid stopping position on a drivable way. The value is defined here. The flowchart under *Stop time* and every rule that consumes the tolerance name the constant rather than repeating the figure, because a diagram or an implementation holding its own copy will silently disagree with a deployment that has tightened it.

**What this constant separates is two validation regimes, not two cost models.** The flowchart under *Stop time* is the precise statement **of that split and of nothing wider** — it decides which model prices a stop, it opens on a candidate already at a valid stopping position, and it carries no road-class, speed-limit, or establishment gate on any path, because those sit upstream of it. A candidate inside the tolerance is checked for level compatibility and intervening barriers, and enters no part of the walk-safety branch: it is never required to produce a connected walkable path, an obtainable elevation profile, or a gradient that is not implausibly steep.

**The constant separates that pair and exempts a candidate from nothing outside it.** This paragraph read *and nothing else* until 7 August 2026, which — in a paragraph whose first sentence says the constant decides which validation a candidate must pass — said that a candidate inside the tolerance faces two checks in the whole system. The road-class and speed-limit exclusions, and the requirement that a stopping position be established at all, bind on both sides of this constant; *SPECIFICATION.md § Directly road-accessible zones* carries the correction and names all three of the requirements the old wording excluded. A candidate outside the tolerance must pass that branch or be excluded or downgraded to uncertain. Raising this constant therefore does not mis-price a zone — it **enlarges the set of zones that bypass the walk-safety gates entirely**, and the zone is then handed to a driver who is told to take it from the car. Where that judgement is wrong, the driver stops, finds they are not inside the zone, gets out, and becomes a pedestrian on ground no part of the system ever validated, because the branch that would have validated it was never entered.

**Why downward is the strict direction.** Set below its true value, a zone that could have been taken from the car becomes a park-and-walk stop, and a park-and-walk stop is strictly more validated: it must produce a connected path and an elevation profile, or fall to the uncertain bucket, or be excluded. The worst outcome of a tolerance set too low is therefore lost coverage on zones that were in fact reachable from the seat — which is the loss the measure of success under *Accessibility scope for the first release* in `SPECIFICATION.md` accepts, since it asks not that every zone be classified but that none be classified confidently and wrongly. The worst outcome of one set too high is a pedestrian beside a road.

**The derivation is an argument, not arithmetic.** Ten metres was reasoned against a nominal 25-metre zone, whose half-extent would be about 12.5 metres — but *Zone geometry* in `Architecture.md` records that figure as a guideline rather than a guarantee, that real sizes vary considerably because each zone is fitted to its place, and that the API exposes nothing about the shape or size of any individual zone. No per-zone extent is knowable, so nothing here is a calculation from a measured quantity, and the arithmetic must not be read as one. It is a conservative choice argued from a nominal figure the corpus itself declines to rely on.

Beyond the tolerance the zone is not lost: it enters the park-and-walk branch, where it may be priced with a short walk, downgraded to uncertain, or excluded on the evidence found there.

This 10-metre figure is a proposal and is exactly the kind of threshold worth checking against real captures early, since it decides which candidates are exempted from walk-safety validation.

---

## The maximum speed limit for a stopping road

**This is an enforcement constant, and its strict direction is downward.**

The constant is **90 km/h**: the highest speed limit a road may be **established** to carry and still be considered as a place to stop. The rule that consumes it is stated under *Enforceable exclusions* in `SPECIFICATION.md` and is not restated here. Only the figure lives here.

**The value is a proposal; the exclusion is not.** Nothing in this documentation set establishes 90 km/h. It carries no origin, no measurement, and no cited source, and it is not enumerated in *Estimate accuracy and calibration* in `SPECIFICATION.md`, which enumerates which constants are settled, which need measurement, and which are irreducible. Under *Conventions* in `docs/README.md`, where numeric constants are proposals unless stated otherwise, it is therefore a **proposed default by default** — held to that status because nothing supports a stronger one, not because anyone argued for the number. It must not be presented to a user, a requirement, or a reviewer as a legal threshold or as a figure taken from any road-traffic source; no such attribution exists to make. Whatever value the constant holds, **the exclusion built on it is absolute**: a zone whose only stopping position is on a road above the limit is excluded regardless of its Turf value, and time is never grounds for relaxing it.

**Why downward is the strict direction.** Lowering the constant excludes more roads, which is the direction that costs coverage rather than safety, and it degrades gracefully in the cost model: the roadside rows in *Proposed placeholder timings* run from 30 km/h to this limit, so every stopping context a tightening leaves admissible still has a row. Raising it does not. Those rows stop where the exclusion stops, so raising the constant admits roads faster than the table prices — putting stopping places into the model with nothing calibrated, or even guessed, for stopping on them.

**What this constant costs, measured.** In the `zones-dump-2025-01-04.json` corpus of **138,830 zones**, Great Britain holds **44,830 — 32.3%**, the second-largest national share after Sweden. The UK national speed limit is 60 mph (96.6 km/h) on a rural single carriageway and 70 mph (112.7 km/h) on a dual carriageway, both above 90 km/h. Essentially the entire UK rural network at national speed limit is therefore excluded as a stopping place, so a market holding roughly a third of all zones is one whose rural network is reached by parking and walking rather than by stopping at the roadside. **This is a recorded product decision, not a defect.** The exclusion was held at 90 km/h with this figure in front of the decision-maker, and it is stated in product terms under *The United Kingdom is a park-and-walk market* in `SPECIFICATION.md`. Nobody should raise or inherit this constant without knowing it.

What is measured here is the **consequence** of the constant, not the constant. This paragraph is not a source for 90 km/h and must not be cited as one; the value remains a proposed default with no origin, exactly as stated above.

**The two halves of that paragraph do not stand on the same evidence.** The **44,830 — 32.3%** share is measured, against the dated dump named above. The 60 and 70 mph limits are **recalled UK road law, not checked against a published source by anyone in this repository** — citable in principle, uncited in fact, and registered as unverified under *Estimate accuracy and calibration* in `SPECIFICATION.md`. The conclusion needs both halves: without the limits, the corpus share says nothing about which roads the exclusion removes. A later pass that verifies them puts the citation here and retires this marking.

The measurement **strengthens the case for downward being the strict direction rather than weakening it.** The pressure a third of the corpus creates is upward — one number moved, and a large market becomes far more directly accessible. That is precisely the direction with no priced rows above the limit to fall back on, and the direction whose failure mode is a driver stopping on a 96.6 km/h carriageway. A coverage argument of this size is exactly the kind this constant must be able to refuse.

---

## Additional journey time

The cost metric is additional journey time. The system must estimate more than the additional driving duration returned by a routing service. It must also account for slowing down, parking, leaving the car, walking, taking the zone, returning to the car, and rejoining traffic.

The route provider calculates the baseline driving time and the driving time of any rerouted journey. The Turf optimizer then adds stop-specific service time that an ordinary routing provider does not include.

For every route alternative:

```text
additional journey time =
    Turf-enhanced journey duration
    − baseline journey duration
```

The Turf-enhanced journey duration consists of:

```text
rerouted driving time
+ all zone stop times
```

This distinction prevents the system from double-counting the time spent driving to a zone. The routing provider accounts for the changed road route, while the Turf stop model accounts for actions performed after or during the stop.

Detour cost is obtained by routing and never inferred from geometry, per *Detour cost must always be routed, never inferred* in `SPECIFICATION.md`. A consequence for every formula below is that a stop's cost is meaningful only for one journey travelled in one direction, and must not be cached or reused across journeys as though it were a property of the zone.

---

## The absolute additional-time ceiling

**This is an enforcement constant, and its strict direction is downward.**

The ceiling is **115% of the user's stated additional-time limit** — an allowance of 15% above the figure the user gave. A deployment may lower this multiplier; it may never raise it. 115% is the maximum permitted value, not a default to be tuned in both directions, and a configuration that sets it higher is invalid rather than merely unusual.

The model this constant serves — soft target, hard ceiling, and the clearly-labelled stretch band between them — is stated under *User time constraints* in `SPECIFICATION.md` and is not restated here.

**The base is the additional time, never the journey duration.** The multiplier applies to the additional time the user permitted, not to the total duration of the journey. This is the misreading the constant exists to foreclose, because it fails in the direction that puts a driver somewhere they did not agree to be: on a six-hour drive with a twenty-minute budget, the ceiling is **23 minutes of additional time**. Read against journey duration it would be 54 minutes — 2.7 times what the driver agreed to, while every ceiling check still reported compliance.

**The ceiling is journey-level, never per-leg.** Where a journey has several legs, the quantity the ceiling tests is the sum of additional time across all of them, per *Journeys with several legs* in `SPECIFICATION.md`. Applied per leg instead, four legs at 115% each would compound far past the promise made to the user.

**Two properties, both binding.** The 15% figure is a **proposed default**: nothing measured establishes it, and it may not be quoted to a user as though something did. What it fixes is the **maximum permitted value** stated above, and revising that is a change to this specification, ratified here — the maximum is not a deployment setting. The **value a deployment enforces** is a separate object, governed by the strict direction stated above: a deployment running stricter than the maximum revises nothing, and needs no ratification to do it. The ceiling's **enforcement is not a proposal**: whatever value the constant holds, no recommendation may exceed it, for any reason, however valuable the zone. A statement carrying only the first property leaves the constraint tunable in both directions; one carrying only the second falsely claims a measured figure. Both are stated because either alone is wrong.

This ceiling is **unrelated to the 15 km corridor cap** under *Bounding the candidate set*, which shares no more than a numeral with it. That cap bounds where candidates may be looked for, and being wrong costs result quality; this ceiling bounds what may be recommended, and being wrong routes a driver off their journey.

---

## Stop time

Which model prices a stop follows from its access classification:

```mermaid
flowchart TD
    A[Candidate zone] --> B{Coordinate within the direct-access<br/>tolerance of a valid stopping<br/>position on a drivable way?}
    B -- yes --> C{Levels compatible and<br/>no intervening barrier?}
    C -- yes --> D["Direct road-access calculation<br/>no walking, no car exit"]
    C -- no --> E
    B -- no --> E{Connected walkable path and<br/>elevation profile obtainable?}
    E -- yes --> F["Elevation-aware walking time<br/>park-and-walk stop model"]
    E -- no --> G{Path absent, disconnected,<br/>or implausibly steep?}
    G -- yes --> H[Excluded]
    G -- no --> I["Flat-distance fallback<br/>low confidence, uncertain bucket"]
```

The tolerance tested at the first decision is defined under *Direct-access tolerance*. This diagram references it and deliberately does not repeat its value, so a deployment that tightens the constant tightens the diagram with it.

Takeover time is a separate component of every one of these models and is defined under *Takeover time* below. It applies to park-and-walk zones and directly road-accessible zones alike.

### Elevation-aware walking time

Walking time is calculated from the actual access path and its elevation profile rather than from horizontal distance alone. Walking distance is measured along a routed walkable path to the zone's coordinate, and the path-based, elevation-aware calculation is the default whenever pedestrian-path and elevation data are available.

The basic flat-ground calculation is:

```text
flat walking time =
    walking distance ÷ configured walking speed
```

The walking speed is a configurable average adult walking speed unless the user provides a personal value in advanced settings.

For elevation-aware routing, the path is divided into segments. The effective walking speed of each segment is adjusted according to its gradient and, when available, its surface or path type.

Conceptually:

```text
walking time =
    sum of time required for each path segment
```

```text
segment time =
    segment distance ÷ grade-adjusted walking speed
```

Uphill and downhill travel must not be assumed to require the same amount of time. A steep ascent normally reduces walking speed, while a steep descent may also be slower because of footing and safety. The return journey is therefore calculated from the reversed elevation profile rather than by doubling the outbound time.

A stop may serve one zone or several — the common case rather than the exceptional one, per *Stops serving several zones* in `SPECIFICATION.md` — so the model is expressed over the set of zones taken from it:

```text
park-and-walk stop time =
    slowdown and parking time
    + exit and lock time
    + terrain-adjusted walking time along the route that visits
      each selected zone's coordinate and returns to the car
    + takeover time × number of zones taken
    + unlock and enter time
    + time to rejoin the road and accelerate
```

Where a stop serves a single zone, that walking route is simply an out-and-back to one zone coordinate, computed from the outbound profile and its reverse.

The access cost therefore reflects both horizontal movement and vertical effort.

This is the canonical park-and-walk stop-time model. It is stated only here, and the distances feeding it are measured along the routed path to each zone coordinate. The flat calculation above is the per-segment building block, not an alternative to the model. The only permitted departure is the explicitly degraded estimate under *Flat-distance fallback*.

### Flat-distance fallback

Where a connected path is identified but its elevation profile cannot be obtained, the system may fall back to a symmetric straight-line estimate:

```text
total walking distance =
    straight-line distance from stopping point to the zone coordinate × 2
```

```text
walking time =
    total walking distance ÷ walking speed
```

This fallback is explicitly a degraded estimate. It ignores terrain, gradient, and barriers, and it assumes the outbound and return journeys take equal time — an assumption *Elevation-aware walking time* rejects wherever real data exists. Any candidate priced with this fallback must be marked low-confidence and is subject to the exclusion rules under *Elevation and feasibility rules* in `SPECIFICATION.md`. It must never be used to justify a recommendation presented as reliable.

**The fallback prices an uncertainty about a path that exists; it never prices an absence.** An access path that is absent or disconnected is excluded outright under *Enforceable exclusions* in `SPECIFICATION.md`, which is a hard rule rather than a preference — and a hard rule is not satisfied by pricing the candidate anyway, however degraded the estimate and however plainly it is labelled. What reaches this fallback is a path the system found and could not profile; what never reaches it is a path the system could not find. The flowchart under *Stop time* is the decision procedure that sorts the two, and where its reading and this wording ever appear to differ, the flowchart governs.

The opening condition above was written as a disjunction until 7 August 2026 — *no connected path **or** elevation profile* — which granted the fallback over an absent path and contradicted both that exclusion and the flowchart. The Owner ruled the flowchart's reading. That is recorded rather than quietly corrected, because the disjunction is the reading a reader arrives with: doubling a straight line to a path nobody could find produces a plausible number for a candidate that should never have been priced at all, and it produces it in the direction that makes the candidate look cheap.

### Direct road-access calculation

For a zone that can be captured directly from a legally and safely stoppable road position, the model becomes:

```text
direct road stop time =
    slowdown and safe stopping time
    + takeover time
    + time to rejoin traffic and accelerate
```

No walking, parking, car-exit, or car-entry time is required.

The system must not assume that the player can capture a zone while driving. The vehicle is expected to stop.

Unlike the park-and-walk model, this one is deliberately single-zone. A stopping position from which two zones are simultaneously capturable is possible where zones sit unusually close — see *Distance between zones* in `Architecture.md` — but it is rare enough that the general case is one zone per stopping position. Two zones taken from two positions are two stops, each priced separately.

Where a stop does turn out to serve two zones from the car, the correct treatment is the same as above with takeover time counted once per zone. It must not be modelled as a park-and-walk stop with zero walking distance, because the parking and car-exit components do not apply.

### Speed and manoeuvre calculations

Road speed affects the time required to slow down, stop, and return to normal traffic. However, speed multiplied by a deceleration time produces a distance, not a time. The stop model therefore uses time-based functions or lookup tables.

Conceptually:

```text
slowdown and parking time =
    function of speed limit, road type, stopping location, and parking type
```

```text
rejoin and acceleration time =
    function of speed limit, road type, traffic context, and stopping location
```

This will likely produce more reliable estimates than attempting to derive every manoeuvre from idealized vehicle acceleration and deceleration physics.

### Proposed placeholder timings

These values are **uncalibrated estimates**. They exist so the system can run and produce plausible output before anyone has driven the measurements, and they are the single largest source of error in the time model. Every one must be replaced by measurement, and none may be quoted to a user as though it were established.

| Stopping context | Slow down and stop | Rejoin and accelerate |
|------------------|--------------------|-----------------------|
| Parking area or rest area | 45 s | 30 s |
| Urban roadside, 30–50 km/h | 25 s | 20 s |
| Regional roadside, 70 km/h | 40 s | 35 s |
| Regional roadside, 80 km/h | 50 s | 45 s |
| Regional roadside, 90 km/h | 55 s | 50 s |
| Exit and re-entry via intersection | 60 s | 60 s |

A further **15-second buffer** applies per stop, covering the small unpredictable delays described below.

Roads above *The maximum speed limit for a stopping road* do not appear in this table because stopping on them is excluded outright under *Enforceable exclusions* in `SPECIFICATION.md`. Where such a zone is reachable from an adjacent rest area or service road, the relevant row is the one describing that stopping place, not the road the zone sits beside.

The roadside rows run to the same limit the exclusion does, so every road permitted as a stopping place resolves to a row. Where a recorded limit falls between two roadside rows, it takes the **faster** row rather than the nearer one. The top of this table tracks that constant rather than standing on its own: tightening the constant strands the rows above the new limit, and they should be removed with it so that this table never prices a road the system will not stop on.

**The 80 and 90 km/h rows are the weakest numbers in this table, and are biased high on purpose.** They were not estimated independently; they were extrapolated from the two slower roadside rows, which are themselves guesses, so they inherit that error and add their own. The trend those rows describe is +15 s in both columns between the top of the urban band (50 km/h) and the 70 km/h row — 7.5 s per 10 km/h — continued to 80 and 90 and rounded **up** to the 5-second granularity the rest of the table uses. Upward is the deliberate direction: understating a manoeuvre understates the cost of the stop it belongs to, which makes the stop look cheaper than it is and lets a plan built on it run closer to the absolute ceiling than its true cost allows. A reader who needs these to be wrong in a known direction should read them as too generous, never as too tight. What replaces them is measurement — the same stopwatch runs the other rows are waiting on, driven on roads posted at these limits rather than interpolated from slower ones.

The timing model also includes small configurable buffers because real-world actions are not perfectly deterministic. Parking availability, traffic, road crossings, GPS delay, and the exact zone geometry can all affect the actual duration.

---

## Takeover time

The time required to capture a Turf zone varies by player rank. The optimizer must include takeover duration because it directly affects the true cost of every recommended stop.

Rank is available from `POST /v5/users` by username, so this path is supported by real data rather than aspirational. The API does **not** expose takeover duration itself, so that value must be held by this system.

### Rank-to-takeover-time table

Takeover time is not an unknown requiring measurement. It follows a documented linear rule over the rank range 0 to 60:

```text
base takeover time (seconds) =
    30 − (0.2 × rank)
```

This gives 30 seconds at rank 0 and 18 seconds at rank 60.

The system carries this as a maintained, hard-coded table of rank to seconds, generated from the rule above rather than transcribed by hand. A table is preferred over evaluating the formula at runtime because it is directly auditable against the source, and because the rule is a published game constant that may be revised without notice.

Because the table is a copy of external information, it must carry its provenance in the source file: where the values came from, the date they were checked, and the formula they were generated from. A silent drift between the game's real takeover time and this table would corrupt every stop estimate in the system while leaving no visible symptom.

The authoritative sources are the Turf wiki's *Takeover time* page and the game's own information pages. Neither publishes a machine-readable form, so this remains a manual synchronization point and should be re-checked periodically.

### Bonuses that reduce takeover time

Rank alone does not determine takeover time. Two bonuses each reduce it by five seconds, and they stack, with the documented floor being eight seconds for a rank-60 player holding both.

**Region Lord** applies globally, not regionally. A player who holds the title in *any* region receives the bonus on *every* takeover, anywhere. Being Region Lord of Zuid-Holland shortens takeovers in Nordnorge by the same five seconds.

The bonus is therefore a single property of the player, not of the zone. It resolves to one boolean per journey: does this user hold a region lordship anywhere? If so, five seconds comes off the base time for every stop on the route.

This is computable, and cheaply — the single call that answers it is recorded under *Region lords* in `Architecture.md`. There is no need to join zones to regions for this purpose, and no need to vary takeover time from stop to stop.

Region lordship changes hands during play, so the flag is volatile and is resolved per journey rather than cached indefinitely.

**GPS bonus** also reduces takeover time by five seconds, but it is **permanently excluded from this model** rather than deferred.

It is earned by keeping Turf's GPS on and remaining visible to other players while travelling between zones. That is a choice each player makes, it cannot be controlled or predicted by this system, and it must not be assumed of every player. There is no version of this system that could reliably know whether the bonus applies.

Excluding it biases estimates conservative, which is the right direction for an error of this kind: a stop may prove faster than promised, never slower.

### Default when rank is unknown

A username is always present, because initialization requires one before the planner opens — see *First-run initialization* in `DESIGN.md`. The default below therefore covers only the case where a known username fails to resolve to a rank at request time: an API outage, a renamed account, or a transient failure. It is a fallback, not the normal path.

The rule above makes this default derivable rather than arbitrary. The mean over ranks 0 to 60 is the value at the midpoint rank of 30:

```text
default takeover time =
    30 − (0.2 × 30) = 24 seconds
```

That derivation must be recorded alongside the constant. The requirement that the default be transparent is met by showing this calculation, not by asserting a number.

No bonus is assumed for an unknown player.

### Zone lock time

Because `blocktime` governs when a zone becomes takeable again, it only affects a journey that would take the same zone twice. The model excludes it, and the exclusion rests on one invariant: that a planned journey **takes each zone at most once**. While that holds, zone lock time has no bearing on any cost in the model and does not appear in it.

That invariant is an assumption, not a guaranteed property — nothing in the requirements corpus currently obliges a plan to take each zone at most once. It is named here so that the exclusion has a stated condition to fail: if a plan can ever take one zone twice, this exclusion must be revisited rather than inherited.

The journey's shape does not settle it. A round trip is in scope, per *Genuinely out of reach or out of scope* in `SPECIFICATION.md`, and a journey of any shape may pass the same zone twice — an out-and-back over one road does so however its endpoints are named. What keeps lock time out of the model is the plan taking each zone at most once, not the journey running from an origin to a different destination.

Should the invariant not hold, the documented behaviour is that lock time rises with rank, from ten minutes at rank 0 to twenty-five minutes at rank 60, and that a revisit is locked for a static five minutes regardless of rank.

The units are confirmed: `blocktime` is **in seconds**. A live query for a rank-60 player returned 1500, which is exactly the twenty-five minutes the published maximum describes. The API's own example response showing 30 is stale documentation and should be disregarded.

`blocktime` is otherwise **out of scope**. It plays no part in the first release.

---

## Zone value

Value is carried by the zone's attribute, ranked 1 to 11 by the user during initialization, per *Attribute preference* in `SPECIFICATION.md`. That ranking is ordinal and the optimizer needs a number. The curve below is the translation.

### Proposed rank-to-weight curve

The following is the **proposed default**. The numbers are a starting position to argue with, not measured values, and the anchors are configurable.

An ordinary zone without an attribute has a base value of **1.0**. Weights decay geometrically from rank 1, each rank worth roughly 57% of the one above:

| Rank | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|------|---|---|---|---|---|---|---|---|---|----|----|
| Weight | 500 | 288 | 166 | 95 | 55 | 32 | 18 | 10 | 6 | 3.5 | 2 |

The curve is anchored at rank 1 = 500 and rank 11 = 2, with the intermediate values following from the constant ratio.

Three properties make this shape a reasonable default:

* **Rank 1 clears the stated bar with headroom.** At 500 to 1, one top-ranked attribute zone outweighs any plausible number of ordinary zones in a single journey — which is the design target set out under *The weighting is extreme* in `SPECIFICATION.md`.
* **Adjacent ranks stay comparable.** Rank 1 beats rank 3 by roughly three to one, so a nearby rank-3 zone can still beat a distant rank-1 zone. The ranking expresses preference without becoming an absolute lexicographic order between adjacent attributes.
* **The lowest-ranked attribute still matters.** At 2.0 it remains worth two ordinary zones, which reflects that ranking an attribute eleventh is not the same as having no interest in it.

Attributed zones add value according to this ranking, and a highly ranked attribute justifies a far larger walk or deviation than any quantity of ordinary zones.

---

## Access difficulty

Difficulty is derived from how often a zone is actually taken, relative to its neighbours. The reasoning behind that signal, and the sampling supporting it, are in *Zone activity as a difficulty and hold-time signal* in `SPECIFICATION.md`. The arithmetic is here.

### Takeover rate

Every zone reports `totalTakeovers` and `dateCreated` — both on 100% of records, per *Retrieving zones* in `Architecture.md`. Together these give a **takeover rate**:

```text
takes per month =
    totalTakeovers ÷ months since dateCreated
```

**The divisor is never exactly zero, and that is why a guard is required.** `dateCreated` carries **second** precision, not day granularity: across the complete corpus its values hold **47,831 distinct times of day**, **83,879** records carry a non-zero seconds field, and every value is `+0000`. A zone created minutes ago therefore yields a divisor that is very small and **not zero**. The arithmetic never raises — **it returns**, and what it returns looks plausible. Zone `810622` was created at `2026-08-02T07:44:52+0000` and carries **one** takeover; at the reference instant below it is **1.10 days** old, giving a divisor of **0.0362 months** and a rate of **27.7 takes per month**, which only **8.7%** of the corpus reaches or exceeds, against a median of **1.3**. A single takeover on a day-old zone is indistinguishable from one of the busiest zones in the game, and nothing in the calculation says so.

**A guard written as `if months == 0` will never fire.** An exact zero is unreachable at second precision, so a zero-divisor test protects nothing while reading as though it protects everything. The **minimum-age** guard specified immediately below is what actually handles this case, and it is the only thing that does — the failure it prevents is a confidently wrong rate returned silently, not an exception raised loudly.

**Reference instant: `2026-08-03T10:10Z`**, when the corpus dump was retrieved. Every age-dependent count in this section is stated against it: **11 zones** were then under 24 hours old and **888** under thirty days, and the youngest record's `dateCreated` is `2026-08-03T00:00:00+0000`, roughly ten hours earlier. These counts move with the instant chosen — measured at midnight the same day they are 12 and 890 — so a later reader finding a small discrepancy has found a different clock rather than an error.

*Zone activity as a difficulty and hold-time signal* in `SPECIFICATION.md` already calls the measure weak for young zones and asks that they be treated as unknown rather than as low-activity. The corpus shows the arithmetic returns a confident wrong answer before the interpretation is ever reached, so the guard is required for correctness and not only for meaningfulness.

**The guard's value is not set here.** A minimum age below which the rate is not computed — and the zone is routed to the unknown treatment `SPECIFICATION.md` already specifies — is required, but no figure in this documentation set establishes one, and none is invented here. It must be chosen deliberately and recorded in this document when it is.

**A zero numerator is a different case and must not be conflated with it.** **4,555 zones** carry `totalTakeovers` of zero. Where the zone is old enough to clear the age guard, that is a well-defined rate of zero and a genuine signal — a zone nobody takes — not a missing value. Only the divisor is undefined; the numerator is merely small.

### The activity baseline

Activity is judged relative to its neighbourhood, never against a global figure. The system computes a local activity baseline — the typical takeover rate across the surrounding zones — and scores each zone against that baseline rather than against an absolute number.

What matters is the ratio:

```text
relative activity =
    zone takes per month ÷ local baseline takes per month
```

The neighbourhood is proposed as the **nearest 100 zones bounded to a 25 km radius**, whichever limit binds first.

One caution in construction: a fixed count of nearest zones spans a very large area in sparse terrain, and may reach into a distant town whose activity has no bearing on the local one. The neighbourhood must be bounded by distance as well as by count, so the baseline stays genuinely local.

**This division has its own zero, and it is the baseline rather than the divisor above.** Where every zone in the neighbourhood has never been taken, the baseline is zero and the ratio is undefined. The 25 km bound makes this reachable rather than theoretical: where the radius binds before the count does, the neighbourhood may hold a handful of zones instead of a hundred. In the corpus of **3 August 2026**, **45 regions consist entirely of zones with zero takeovers** — all small, the largest being Pakistan at 7 zones — and six regions of 100 zones or more are **between 50% and 81% zero** (Gyeongsangnam 125/155, Chungcheongnam 88/122, Gangwon 75/118). A guard is required here too, and as above no value for it is established.

These region-level concentrations are measured. **That a specific 25 km neighbourhood resolves to an all-zero set is inferred from them, not computed** — the spatial neighbourhood query does not exist yet to test it against. The inference is strong enough to require the guard and not strong enough to quantify how often it fires.

### Proposed adjustment

The following is a **proposed default**, not a measured relationship. It feeds the difficulty multiplier in *Proposed form: value per minute*:

```text
difficulty multiplier =
    1 + 0.6 × max(0, 1 − relative activity)      capped at 1.6
```

The behaviour this produces:

| Relative activity | Multiplier | Reading |
|-------------------|-----------|---------|
| 1.0 or above | 1.00 | At or above local normal — no penalty |
| 0.5 | 1.30 | Half the local rate |
| 0.25 | 1.45 | A quarter of the local rate |
| approaching 0 | 1.60 | Capped |

The penalty is deliberately mild and strictly bounded. A quiet zone becomes less attractive but never disqualified, which matches the strength of the evidence: low activity is a hint about difficulty, not a measurement of it.

Three guards apply, and in each case the adjustment is simply not applied:

* Fewer than **20 zones** within the neighbourhood radius, so no meaningful baseline exists.
* A zone younger than **12 months**, where the takeover rate is too unstable to trust.
* A neighbourhood spanning more than **25 km**, beyond which "local" stops meaning anything.

---

## The objective function

### Proposed form: value per minute

The proposed structure keeps **value and cost separate** and scores candidates by their ratio:

```text
value = attribute weight, or 1.0 for an unattributed zone

cost  = marginal additional minutes × difficulty multiplier

score = value ÷ cost
```

```mermaid
flowchart LR
    A[Attribute rank 1–11] --> B["value<br/>weight, or 1.0 unattributed"]
    C[Marginal additional minutes] --> D["cost<br/>minutes × difficulty"]
    E[Relative activity] --> F[difficulty multiplier]
    F --> D
    B --> G["score = value ÷ cost"]
    D --> G
    H[Access confidence] -. gate, never a term .-> G
```

A ratio is preferred over a subtractive form such as `value − λ × cost`. The subtractive form requires λ — an exchange rate between value and minutes — which is one more invented constant with no natural anchor. The ratio needs no such constant and states something directly explainable to the user: *this zone gives you the most value per minute spent*.

Two structural rules follow from this shape:

* **Difficulty inflates cost; it never reduces value.** A `Monument` zone up a steep path is still a `Monument` zone — it simply costs more to reach. Keeping difficulty on the cost side preserves the meaning of both quantities and keeps the units clean.
* **Confidence is a gate, not a term.** Access confidence decides whether a zone is eligible at all, per *Handling the uncertain bucket* in `SPECIFICATION.md`. It does not adjust the score of a zone that qualifies. Blending confidence into the score would let a well-understood mediocre zone and a poorly-understood excellent one become indistinguishable.

The minutes on the cost side are **marginal**, not an isolated round-trip detour: the cost of adding this zone to a sequence that already exists. Where a walk or a driving deviation is shared with another zone, only the increment counts, per *Individual zones rather than local collection routes* in `SPECIFICATION.md`.

At route level the problem is then to maximize total value subject to the time budget. Selecting greedily by this ratio and improving the result with local search is sufficient at the candidate counts involved; exact methods are not warranted.

Where the **Points** objective is selected, points move from being one factor among several to being the criterion that ranks candidates, and expected hold time — the inverse of activity — gains substantial weight. Points is not in the first release, per *Points is deferred* in `SPECIFICATION.md`.

---

## Bounds the interaction layer depends on

The figures below are constants that happen to surface in an interface. They are recorded here rather than in `DESIGN.md` because they are numbers the optimizer tests, not layout.

### A conservative upper bound for an uncertain stop

**This is an enforcement constant, and its strict direction is upward.** It is the one constant in this document where the strict direction is not downward, and a rule applied from memory as "downward only" inverts here. A larger bound refuses more acceptances; a smaller one admits more. **No value is specified below** — the bound is stated qualitatively and is recorded as a named gap under *Open questions owned by this document*.

An uncertain zone accepted during review carries no time estimate, which would leave the check against *The absolute additional-time ceiling* with nothing to evaluate, per *Reconciling this with the absolute ceiling* in `SPECIFICATION.md`.

The resolution is a **conservative upper bound** on the uncertain stop's cost — the worst plausible case given what little is known about its access. It is used for two purposes only: the ceiling check, and the upper end of the widened range shown to the user.

It is used for nothing else. It never enters scoring, never affects ranking, and never makes an uncertain zone comparable to a priced one. The rule that uncertain zones stay out of the cost model is unchanged; this bound exists solely so that an absolute constraint has a number to test.

A useful property falls out of this, and it holds independently of where the ceiling's own multiplier is defined: what the user sees as the widened upper total and what the ceiling tests are **the same number**, so the constraint and the display can never disagree.

That property cuts both ways, which is why the constant is an enforcement one. The number that protects the user is also the number that alarms them, so anyone who finds the widened range too wide to show has a direct route to narrowing it — and narrowing it loosens the ceiling by the same stroke, silently, with every ceiling check still passing. Whoever authors the value must treat a display objection as an argument about presentation, never as a reason to move this bound.

**Until it has a value, the feature it guards does not operate.** The reserve pool is not offered while this bound is unset, per *SPECIFICATION.md § Reconciling this with the absolute ceiling*, which is where that rule is stated and argued. It is recorded here because this is the definition site an implementer reaches when they discover the constant is missing, and the tempting repair at that moment is to supply a plausible figure rather than to refuse — which is the failure the entry under *Enforcement constants that do not yet exist* describes, arriving through the door marked *unblock the feature*.

### Review-interaction thresholds

Rejections cluster, so per-zone replacement stops working after a few attempts in the same place, per *Replacement and escalating scope* in `DESIGN.md`. Two proposed constants govern the escalation:

* Coarser replacement controls appear after **three rejections within 2 km** of one another, which is the point at which per-zone replacement has demonstrably stopped working.
* The coarsest of those controls offers a zone at least **20% of the remaining journey length** further along.

---

## Still owed by this document

A **domain glossary** formalizing the terms these formulas operate on, so that each has one definition rather than an implied one: route, route alternative, route leg, mandatory waypoint, zone, zone attribute, candidate zone, stop location, direct road-access zone, park-and-walk zone, access path, attribute preference, time budget, stretch alternative, accessibility confidence, takeover duration, terrain cost, and route score.

---

## Open questions owned by this document

### Enforcement constants that do not yet exist

Two constants are named by this documentation set, are required for the system to behave as specified, and **have no value anywhere**. They are recorded here rather than authored now: each is written when the batch of work that owns it runs. The point of recording them is that the first implementer to need one cannot author it silently, as an unexplained literal, which *Conventions* forbids.

* **The conservative upper bound for an uncertain stop**, per *A conservative upper bound for an uncertain stop*. Specified entirely qualitatively — "the worst plausible case" — with no value and no origin. It is the only sensor the absolute ceiling has for an uncertain stop: computed optimistically, the ceiling check passes, the refusal never fires, and a user accepts a reserve zone that commits them past the limit they stated. Its **strict direction is upward**. Its specific hazard is that it is the same number the user sees as the top of the widened range, so an implementer who finds that range too alarming to display has a direct incentive to shrink it — and shrinking it loosens the ceiling while every check still reports compliance. **While it is unset the reserve pool does not operate**, ruled on 7 August 2026 and stated under *SPECIFICATION.md § Reconciling this with the absolute ceiling*, so this constant now blocks a feature outright rather than quietly degrading one.
* **The "implausibly steep" gradient threshold**, named as a member of *Enforceable exclusions* in `SPECIFICATION.md` and reached by the last decision of the flowchart under *Stop time*, with no value stated in any document. An exclusion phrased qualitatively cannot be evaluated, so it will otherwise arrive as exactly the unexplained literal *Conventions* forbids. Its **strict direction is downward**. Whoever authors it inherits a coupling no other constant here has: *D6 — Elevation sampled from Copernicus GLO-30* in `Architecture.md` records that a 30-metre global model cannot resolve a retaining wall, which is narrower than one cell, so a threshold calibrated against a 1-metre national model is wrong when the same check runs on the GLO-30 fallback. It may need to be **per elevation provider**, and a single global figure may be the wrong shape for this constant.

### Proposed constants awaiting evidence

* **The rank-to-weight curve** — geometric decay anchored at 500 and 2, per *Proposed rank-to-weight curve*. The anchors and the ratio are all configurable.
* **The core scoring function** — value divided by cost, with difficulty inflating cost and confidence acting as a gate, per *Proposed form: value per minute*.
* **The activity adjustment** — a mild capped multiplier over a neighbourhood-relative ratio, per *Proposed adjustment*, including the neighbourhood size, its distance bound, and the three guards.
* **The base value of an ordinary zone**, proposed as 1.0.
* **The 10-metre direct-access tolerance**, which decides which candidates skip walk-safety validation entirely, per *Direct-access tolerance*, and is worth validating early against real captures.
* **The 90 km/h maximum speed limit for a stopping road**, per *The maximum speed limit for a stopping road*. A proposal by default rather than by argument: no source, measurement, or reasoning for the figure exists anywhere in this documentation set. The exclusion it feeds is absolute regardless; it is the number that is unevidenced, not the rule.
* **Manoeuvre timings.** The placeholders under *Proposed placeholder timings* let the system run, but they are guesses, and they are the largest single source of error in the time model. This is the one thing in the system that requires someone to drive the manoeuvres with a stopwatch — **every row of that table, including the 80 and 90 km/h rows that exist to match the range the exclusion admits.** Those two are the weakest entries in the document: extrapolated from the slower roadside rows rather than observed even once, and rounded upward on purpose, so they are guesses drawn from guesses. They are owed the same measurement as the rest, on roads posted at those limits, and closing this question means retiring all of them and not merely the ones somebody happened to drive.

### Standing obligations

* The **rank-to-takeover-time table** is a standing obligation rather than a question: it is synchronized manually against the Turf wiki, and is subject to change without notice.
