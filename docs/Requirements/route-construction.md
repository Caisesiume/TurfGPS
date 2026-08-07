# Route construction

Assembling one journey alternative from a general route and a set of classified candidate zones: which zones are combined, the order they are visited in, how the marginal cost of adding a zone to an existing sequence is obtained, and the rule that a detour's cost is routed rather than inferred from geometry. Distinct from `Route alternative generation`, which owns whether alternatives are produced at all and in what number — that category owns the set the search yields, this one owns how one member of it is built — and from `Cost and time composition`, which composes a stop's cost out of its parts once the sequence exists. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-086 — Route a detour's cost through the proposed stopping location

```
Statement:    The system shall obtain the cost of a detour to a candidate zone
              by routing the journey through the proposed stopping location
              and comparing the result against the baseline, and shall derive
              it from no geometric measure.
Category:     Route construction
Source:       SPECIFICATION.md § Detour cost must always be routed, never inferred
Priority:     MUST
Verification: test — a candidate lying a short distance from the route line
              but reachable only by continuing to a distant exit, turning and
              returning is charged the routed cost of that detour rather than
              one derived from its distance to the line
Acceptance:   given a candidate lying a short distance from the route line but
              reachable only by continuing to a distant exit, turning and
              returning, when its detour cost is obtained, then the cost is
              the difference between the journey routed through its stopping
              location and the baseline
              given the same candidate, when its detour cost is obtained, then
              the cost is not derived from its distance to the route line,
              from a radius, or from any other straight-line measure
Status:       draft
Depends-on:   FR-008; FR-064; FR-092
Volatility:   settled — the rule is stated absolutely by the cited section and
              by `safety-path-checklist § Non-negotiables`, and rests on no
              figure
Risk:         A straight-line estimate beside a dual carriageway can be twenty
              minutes wrong in the direction that matters, and it is wrong in
              the direction that makes the zone look cheap — so the error does
              not distribute, it selects. Every alternative built on it is
              inside the stated limit on paper and outside it on the road, and
              the absolute ceiling is then enforced against the same wrong
              number.
Rationale:    The cited section states the rule and the failures it prevents:
              one-way systems, central reservations, restricted turns and
              asymmetric exits all produce the same effect, and routing
              through the stop point captures every one of them with no
              special handling. The one legitimate use of a straight-line
              measure is cheaply reducing the corridor to a set worth routing,
              which FR-063 and FR-064 carry; a straight-line walking distance
              is refused by FR-074 and a proximity-based access classification
              by FR-085, so no separate record is authored for the permission
              itself. That a driving deviation shared between two zones is
              charged once follows from this record together with FR-008 and
              is not authored separately. That the resulting cost is specific
              to one journey and one direction is FR-087's.
Resolved-by:  —
```

## FR-087 — Reuse no candidate's detour cost across journeys

```
Statement:    The system shall obtain a candidate zone's detour cost for the
              journey and the direction of travel it is being evaluated for,
              and shall not reuse a cost obtained for another journey or
              another direction.
Category:     Route construction
Source:       SPECIFICATION.md § Detour cost must always be routed, never inferred
Priority:     MUST
Verification: test — one candidate evaluated on two journeys passing it in
              opposite directions along a road whose exits are not symmetric
              is charged two different detour costs, each routed for its own
              journey
Acceptance:   given a candidate evaluated on two journeys that pass it in
              opposite directions along a road whose exits are not symmetric,
              when its detour cost is obtained for each, then the two costs
              differ and each is the one routed for that journey
              given a candidate already evaluated on an earlier journey, when
              it is evaluated on a later one, then its detour cost is obtained
              for the later journey rather than taken from the earlier
Status:       draft
Depends-on:   FR-086
Volatility:   settled — the direction-dependence is stated as a consequence of
              routing by the cited section and again by
              `CalculationSpecification.md § Additional journey time`, and
              rests on no figure
Risk:         A cached detour cost is the cheapest optimization available in
              this pipeline and is indistinguishable from a correct one in any
              single-journey test: the figure is real, it was routed, and it
              simply belongs to another journey. The error appears only across
              journeys, in whichever direction the exits happen to favour, and
              it arrives on the stop a driver is being told to make.
Rationale:    The cited section states that a candidate's cost is meaningful
              only for a specific journey travelled in a specific direction
              and must not be cached or reused across journeys as though it
              were a property of the zone;
              `CalculationSpecification.md § Additional journey time` names
              the same consequence for every formula built on it. Separate
              from FR-086 because the two fail independently and neither test
              finds the other: a system that routes every detour honestly and
              memoizes the result by zone passes FR-086 on every fresh
              computation while serving the northbound figure to a southbound
              journey, and a system that never caches anything can still
              estimate from geometry. The failure modes differ in kind as well
              — FR-086's is visible in one journey's numbers, this one's only
              in the second journey, which is exactly why it survives the
              tests written for the first.
Resolved-by:  —
```
