# Cost and time composition

Composing stop cost and journey cost from routing, walking and manoeuvre components. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `.claude/skills/requirements-authoring/SKILL.md`.

## FR-008 — Measure journey cost as time added to the journey without Turf stops

```
Statement:    The system shall measure a journey alternative's cost as the
              additional time that alternative takes compared with the
              corresponding journey without Turf stops, computed as defined
              under
              *Additional journey time* in `CalculationSpecification.md`.
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
Risk:         Measured against anything other than the journey the user would
              otherwise have driven, every figure the product shows is taken
              from the wrong origin — the additional-time target, the ceiling,
              and the value-per-minute ratio alike — and the one number the
              user actually cares about is wrong in a way no later check can
              detect.
Rationale:    The section states what cost is measured against; the arithmetic
              has one home and is cited, not restated. That citation also
              fixes which journey is "corresponding" — it is stated per route
              alternative, against the general route the alternative is built
              on, per *General route alternatives* in `SPECIFICATION.md`.
Resolved-by:  #7
```

## FR-009 — Charge stop time to the journey even where no detour is driven

```
Statement:    The system shall include a stop's time, as defined under
              *Stop time* in `CalculationSpecification.md`, in a journey
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
              *Stop time* in `CalculationSpecification.md`.
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
Status:       draft
Depends-on:   FR-009
Risk:         `blocktime` is the only duration the player endpoint returns, it
              is in seconds, it varies by rank, and takeover duration — the
              thing an implementer is looking for — is not exposed by the API
              at all. Every surface property of the field invites the
              substitution and nothing about it flags the mistake.
              Substituted, it replaces a per-stop duration with one whose
              documented range, per
              *Zone lock time* in `CalculationSpecification.md`, is far above
              any takeover time, so every stop cost, every ceiling check and
              every ranking moves in the same direction at once, with no
              symptom beyond the tool advising against stops it should be
              recommending.
Rationale:    The second criterion exists because the first can be satisfied
              by an implementation that reintroduces the same quantity without
              naming the field. `Source` is the integration section that
              verified what the field means; `Category` is the subsystem the
              prohibition binds, which is how stop time is composed. That the
              field is out of scope for the first release altogether, and the
              invariant that exclusion rests on, are stated under
              *Zone lock time* in `CalculationSpecification.md` and are not
              authored here.
Resolved-by:  —
```
