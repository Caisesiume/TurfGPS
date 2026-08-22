# Cost and time composition

Composing stop cost and journey cost from routing, walking and manoeuvre components. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-008 — Measure journey cost as time added to the journey without Turf stops

```
Statement:    The system shall measure a journey alternative's cost as the
              additional time that alternative takes compared with the
              corresponding journey without Turf stops, computed as defined
              under `CalculationSpecification.md § Additional journey time`.
Category:     Cost and time composition
Source:       SPECIFICATION.md § The journey as an optimization problem
Priority:     MUST
Verification: test — the cost reported for an alternative equals its estimated
              duration less the estimated duration of the corresponding
              journey without Turf stops, and an alternative holding no Turf
              stop is charged no additional time
Acceptance:   given a journey alternative holding at least one Turf stop, when
              its cost is computed, then the cost equals that alternative's
              estimated duration less the estimated duration of the
              corresponding journey without Turf stops
              given a journey alternative holding no Turf stop, when its cost
              is computed, then no additional time is attributed to it
Status:       to-build
Depends-on:   FR-001
Volatility:   settled — the record names what cost is measured against and
              cites the arithmetic rather than restating it, so no proposed
              constant reaches it; gained on the opportunistic transition,
              this record's first edit since the field existed
Risk:         Measured against anything other than the general route the
              alternative is built on, this figure stops comparing Turf
              strategies and starts comparing roads: two strategies on one
              road are no longer separable, and the value-per-minute ratio is
              taken from the wrong origin. The quantity the stated limit and
              the absolute ceiling are tested against is built on this one,
              per FR-061, so an error here reaches both of them while every
              check still reports compliance.
Rationale:    The section states what cost is measured against; the arithmetic
              has one home and is cited, not restated. That citation also
              fixes which journey is "corresponding" — it is stated per route
              alternative, against the general route the alternative is built
              on, per `SPECIFICATION.md § General route alternatives`. What
              this figure is not is the quantity the user's stated limit
              tests. The Owner ruled on 7 August 2026 that the limit bounds an
              alternative's total additional time against the fastest
              conventional route — the general-route deviation and this figure
              together — and FR-061 carries that quantity. The obligation here
              is untouched by the ruling: each general route stays its own
              baseline, which is what keeps the two figures separately
              computable and two Turf strategies on one road comparable. Only
              `Risk` moved, which had claimed the additional-time target and
              the ceiling rest on this record's quantity; they rest on
              FR-061's, and this one is what FR-061's is built from.
Resolved-by:  #7
```

## FR-009 — Charge stop time to the journey even where no detour is driven

```
Statement:    The system shall include a stop's time, as defined under
              `CalculationSpecification.md § Stop time`, in a journey
              alternative's additional time even where capturing that zone
              changes no driving time.
Category:     Cost and time composition
Source:       SPECIFICATION.md § The journey as an optimization problem
Priority:     MUST
Verification: test — an alternative whose only stop is a zone on the road
              already driven, requiring no deviation, is charged more
              additional time than the same journey planned without that stop
Acceptance:   given a journey alternative capturing a zone that lies on the
              road it already drives, so that its driving time is unchanged,
              when its additional time is computed, then that time exceeds the
              additional time of the same journey planned without the stop, by
              the stop's time
              given a journey alternative capturing a zone that requires a
              driving deviation, when its additional time is computed, then
              that time carries both the changed driving time and the stop's
              time
Status:       to-build
Depends-on:   FR-008
Risk:         A zone priced at its detour alone is free whenever it sits on
              the route, so the optimizer fills the journey with roadside
              stops the user never agreed to spend time on, and the
              additional-time target and the absolute ceiling are then
              enforced against a figure that omits every minute actually spent
              standing still.
Rationale:    Not redundant with FR-008: a routing provider returns driving
              duration only, so an implementation computing FR-008's
              difference from routing output alone satisfies its arithmetic
              while omitting the whole stop.
              `CalculationSpecification.md § Additional journey time` names
              that precise omission as what the system must estimate beyond
              the routing service. What a stop costs, which differs by access
              class, is owned elsewhere and is cited rather than restated
              here.
Resolved-by:  #8
```

## FR-027 — Exclude blocktime from stop time

```
Statement:    The system shall not use the `blocktime` value returned for a
              player as a component of stop time, as composed under
              `CalculationSpecification.md § Stop time`.
Category:     Cost and time composition
Source:       Architecture.md § Player data
Priority:     MUST
Verification: test — two players whose `blocktime` values differ and whose
              rank and takeover-bonus state are equal receive the same stop
              time for the same zone
Acceptance:   given two players whose rank and takeover-bonus state are equal
              and whose `blocktime` values differ, when stop time is computed
              for the same zone, then the two results are equal
              given a journey that captures each of its zones once, when stop
              time is composed for each stop, then no stop's time carries a
              period during which the zone is locked against being taken again
Status:       to-build
Depends-on:   FR-009
Risk:         `blocktime` is the only duration the player endpoint returns, it
              is in seconds, it varies by rank, and takeover duration — the
              thing an implementer is looking for — is not exposed by the API
              at all. Every surface property of the field invites the
              substitution and nothing about it flags the mistake.
              Substituted, it replaces a per-stop duration with one whose
              documented range, per
              `CalculationSpecification.md § Zone lock time`, is far above any
              takeover time, so every stop cost, every ceiling check and every
              ranking moves in the same direction at once, with no symptom
              beyond the tool advising against stops it should be
              recommending.
Rationale:    The second criterion exists because the first can be satisfied
              by an implementation that reintroduces the same quantity without
              naming the field. `Source` is the integration section that
              verified what the field means; `Category` is the subsystem the
              prohibition binds, which is how stop time is composed. That the
              field is out of scope for the first release altogether, and the
              invariant that exclusion rests on, are stated under
              `CalculationSpecification.md § Zone lock time` and are not
              authored here.
Resolved-by:  #32
```

## FR-062 — Keep the general-route deviation separable from the Turf stop time

```
Statement:    The system should compute a journey alternative's deviation from
              the fastest conventional route separately from the additional
              time its Turf stops add, and retain both quantities for that
              alternative.
Category:     Cost and time composition
Source:       SPECIFICATION.md § General route alternatives
Priority:     SHOULD
Verification: test — both quantities are separately available for an
              alternative built on a general road route slower than the
              fastest conventional one, and they sum to the total FR-061 tests
Acceptance:   given a journey alternative built on a general road route slower
              than the fastest conventional one and capturing at least one
              zone, when its cost is composed, then its general-route
              deviation and its Turf stop time are each available as
              quantities in their own right and their sum is the total FR-061
              tests
              given a journey alternative built on the fastest conventional
              route, when its cost is composed, then its general-route
              deviation is zero and its total additional time is its Turf stop
              time
Status:       to-build
Depends-on:   FR-008; FR-061
Volatility:   settled — the record obliges two quantities to stay distinct and
              states neither of them, so nothing proposed or open moves it
Rationale:    The section obliges the system to distinguish general route
              selection from zone-level detours, and that distinction is
              observable only as two quantities: an implementation composing
              the correct total in one step satisfies FR-061 while being
              unable to say which half of it the user is buying. That is the
              independence witness, and it is why this is a record rather than
              a criterion on FR-061; it is also the weaker of the two, which
              is why the verbs differ. Nothing here obliges either quantity to
              be shown — what an explanation must carry is
              `Recommendation disclosure`'s, and nothing in the cited section
              creates it — so the record binds availability and stops there.
              Risk is omitted and its argument folded here, on FR-025's
              precedent: the ceiling and the stated limit are tested against
              the total by FR-061 whether or not the halves are separable.
Resolved-by:  —
```

## FR-074 — Compute access cost from the routed path and its elevation profile

```
Statement:    The system shall compute a park-and-walk candidate's walking
              distance and walking time from the routed walkable path and that
              path's elevation profile, per
              `CalculationSpecification.md § Elevation-aware walking time`,
              wherever pedestrian-path and elevation data are obtainable.
Category:     Cost and time composition
Source:       SPECIFICATION.md § Access-path validation
Priority:     MUST
Verification: test — a candidate whose routed path is longer than the straight
              line to its coordinate is charged the routed length, and of two
              candidates whose routed paths are of equal length the one
              carrying substantial ascent is charged the greater walking time
Acceptance:   given a park-and-walk candidate for which a routed walkable path
              and its elevation profile are obtainable, when its access cost
              is computed, then the walking distance used is the length of
              that path and not the straight-line distance to the zone's
              coordinate
              given two park-and-walk candidates whose routed paths are of
              equal length, one flat and one carrying substantial ascent, when
              their access costs are computed, then the climbing one is
              charged the greater walking time
              given a park-and-walk candidate for which a connected walkable
              path is identified but its elevation profile cannot be
              obtained, when its access cost is computed, then the
              straight-line estimate under
              `CalculationSpecification.md § Flat-distance fallback` is what
              prices it
Status:       to-build
Depends-on:   FR-009; FR-068; FR-092
Volatility:   settled — the record names which inputs the computation takes
              and asserts an ordering over them, so it survives every figure
              in the cited model moving
Risk:         A straight-line walk beside a fenced railway or across a cutting
              is wrong by the whole detour to the crossing, and it is wrong in
              the direction that makes the zone look cheap. Priced that way
              the zone beats everything honest around it, so the error does
              not merely mis-state one stop — it selects for exactly the stops
              it mis-states.
Rationale:    The cited section permits a straight-line measurement as an
              early candidate filter and forbids it as the final walking
              distance where topographical or path data is available;
              `CalculationSpecification.md § Elevation-aware walking time`
              states the computation and names the fallback as its only
              permitted departure. This record binds which of the two applies
              and cites both rather than restating either. The second
              criterion states the relation the cited model defines — vertical
              effort is priced, not only horizontal movement — while carrying
              no figure the model is parameterized by. The third states the
              fallback's own condition, so the record cannot be read as
              forbidding the fallback where it is the honest estimate; that a
              candidate priced by it may never hold a confident class is
              FR-083's.
Resolved-by:  —
```

## FR-079 — Price a park-and-walk stop over the set of zones it serves

```
Statement:    The system shall price a park-and-walk stop over the set of
              zones taken from it, per
              `CalculationSpecification.md § Elevation-aware walking time`,
              rather than as a separate out-and-back walk for each zone.
Category:     Cost and time composition
Source:       SPECIFICATION.md § Stops serving several zones
Priority:     MUST
Verification: test — a stop from which two zones are taken on one walk is
              charged less than two stops each serving one of them, and a zone
              whose coordinate lies on a walk already being made, reached over
              a leg validated under FR-078, is charged its takeover time and
              no further walking
Acceptance:   given a stop from which two zones are taken on one walking
              route, when the stop is priced, then its time is less than the
              total of two stops each serving one of those zones
              given a zone whose coordinate lies on the walking route already
              taken to another zone from the same stop, the leg reaching it
              having been validated under the rules FR-078 applies to every
              leg of a multi-zone walk, when the stop is priced, then
              including it adds its takeover time and no further walking
              given a stop serving a single zone, when it is priced, then it
              is priced as a walking route to that zone's coordinate and back
Status:       to-build
Depends-on:   FR-009; FR-078
Volatility:   settled — the record states the shape of the cited model, over a
              set rather than per zone, and carries none of the timings that
              model is parameterized by
Risk:         Priced per zone, the cheapest additions the system can ever find
              — a zone lying on a walk already being made — are invisible, and
              they are invisible in the direction that removes them from the
              result. Priced by nearness to the walk instead, the error runs
              the other way and lands inside the ceiling: this is the only
              record pricing a multi-zone walk, so a zone counted as on the
              route across a fence omits the detour to a legal crossing from
              every figure built on the stop, including the total FR-061
              composes and the ceiling FR-045 is enforced against.
Rationale:    The cited section makes a stop serving several zones the common
              case and states that this is why the park-and-walk stop-time
              model is expressed over a set of zones; the model has one home
              and is cited rather than restated. The second criterion is the
              marginal-cost consequence the section names, written against a
              zone actually on the route so that it needs no threshold for
              almost on the way — and it names FR-078 because the obvious
              implementation of on the route is a distance to the walking
              polyline, which a coordinate three metres away across a fence
              satisfies. That is the failure
              `SPECIFICATION.md § Proximity between zones does not imply cheap chaining`
              warns of by name, and naming the validation in the criterion
              connects the corpus's line rather than leaving it merely
              present. The third criterion states the single-zone case so the
              record cannot be read as forbidding an out-and-back where that
              is what the walk is. The direct road-access model is
              deliberately single-zone and this record does not reach it — two
              zones taken from two stopping positions are two stops, per
              `CalculationSpecification.md § Direct road-access calculation` —
              and that a direct stop is never modelled as a park-and-walk stop
              with no walking is returned with this batch rather than authored
              here. The order in which the zones are walked is method rather
              than outcome and is not authored.
Resolved-by:  —
```

## FR-093 — Compute both halves of a stop against one resolved stopping position

```
Statement:    The system shall compute a stop's driving cost and its walking
              cost against one stopping position as the driving costing model
              resolves it onto the network, rather than resolving that
              position independently for each cost model.
Category:     Cost and time composition
Source:       SPECIFICATION.md § Access-path validation
Priority:     MUST
Verification: test — a stop beside a carriageway carrying a parallel mapped
              footway computes its walking cost from the point the driving
              costing model resolved, and the same stop priced with its
              walking cost measured from the pedestrian costing model's own
              resolution of the same input coordinate fails
Acceptance:   given a stop whose driving cost and whose walking cost are both
              computed, when the stop is priced, then both are computed
              against the stopping position as the driving costing model
              resolved it onto the network
              given a stopping position beside a carriageway carrying a
              parallel mapped footway, which a driving costing model and a
              pedestrian costing model resolve to different points of the
              network, when the stop is priced, then the walking cost is
              measured from the point the driving model resolved and not from
              the point the pedestrian model would
Status:       to-build
Depends-on:   FR-074; FR-086; FR-092
Volatility:   settled — `Architecture.md § D3` is decided and states the
              failure this record prevents as its reason for refusing split
              engines, the chaining that makes the car's resolution the
              authoritative one is the canonical model under
              `CalculationSpecification.md § Elevation-aware walking time`,
              and the record rests on no figure
Risk:         The two halves of one stop are priced by two costing models
              against what is nominally one place, and one input coordinate
              does not make it one place. Where a carriageway and its parallel
              footway are separately mapped, a car costing and a pedestrian
              costing have different edges available to them, so the same
              coordinate resolves to two points and the walk is measured from
              a point the car never occupies. The error is not random. It is
              short by the offset, on every stop of that shape, in the
              direction that makes the zone win. Nothing else in the corpus
              binds the two: FR-086 routes the driving cost through the
              stopping position and FR-074 measures the walking cost from it,
              and two resolutions of one coordinate satisfy both. Every stop
              still yields a plausible number, which `Architecture.md § D3`
              calls the worst available failure mode against a product whose
              measure of success is that no zone is classified confidently and
              wrongly.
Rationale:    The cited section chains the walk from the proposed stopping
              position to the zone's coordinate, and
              `CalculationSpecification.md § Elevation-aware walking time`
              prices that walk as a route visiting each zone's coordinate and
              returning to the car — car, walk, car through one point.
              `Architecture.md § D3` refuses split engines on exactly that
              ground, and the reason it gives is about resolution rather than
              about engines: the point is snapped to two graphs, so the two
              halves of one stop are computed against different geometry. That
              reason outlives the decision it was written for and is not spent
              by it. Each costing model correlates the input to the network
              under its own access rules, and nothing in the decision to run
              one instance establishes that a car costing and a pedestrian
              costing put one coordinate at the same point — so an
              implementation handing one coordinate to both satisfies a
              criterion reading resolved once and reproduces the failure
              whole. That is why the object named here is the resolved
              position and not the coordinate the models are given. Whether
              the two in fact diverge on a given engine is an external
              behaviour nobody in this repository has verified, and this
              record does not rest on the answer: obliging one resolution
              costs nothing if they agree and is the whole of the safeguard if
              they do not. The car's resolution is the authoritative one
              because it is where the vehicle physically stops — the driving
              cost is the cost of bringing the car to that point, and the walk
              chained to it begins and ends where the car actually is. Naming
              the costing models is the one case the Appropriate check in
              `requirements-authoring` permits, the requirement being about
              that decision, and it is cited rather than described. Separate
              from FR-092 because the two fail independently and neither test
              finds the other: a system can refuse every manufactured position
              and still resolve the honest one twice, and one that resolves
              once can resolve a manufactured point. Separate from FR-086,
              which fixes how the driving cost is obtained and says nothing
              about what the walking half is measured from. Nothing here
              restates either cost model — what each half computes is FR-074's
              and FR-086's, and this record binds only that they agree about
              where the car stopped.
Resolved-by:  —
```
