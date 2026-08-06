# Multi-leg budget allocation

Dividing a journey's additional-time budget across its legs before any leg is optimized, and redistributing what a leg does not use. The mirror of `Cost and time composition`: that category composes a cost from its components, this one decomposes an allowance into shares. Distinct from `Safety exclusions`, which owns the journey-level target and ceiling the shares must sum within, and from `Journey initialization`, which owns the capture of the budget being divided. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-050 — Allocate the budget across legs before optimizing any leg

```
Statement:    The system shall allocate a journey's additional-time budget
              across its legs before optimizing any leg of it.
Category:     Multi-leg budget allocation
Source:       SPECIFICATION.md § Journeys with several legs
Priority:     MUST
Verification: test — on a journey of more than one leg, each leg is optimized
              against a share determined before the first leg was optimized,
              and no leg is optimized against the journey's whole
              additional-time budget
Acceptance:   given a journey with more than one leg, when a leg is optimized,
              then the budget that leg is optimized against is the share
              allocated to it and not the journey's whole additional-time
              budget
              given a journey with more than one leg, when its first leg is
              optimized, then a share has already been allocated to every leg
Status:       draft
Depends-on:   FR-002; FR-038
Risk:         Without the allocation each leg consumes the whole allowance and
              the finished journey exceeds it once per leg — the promise made
              to the user broken several times over by a system every one of
              whose per-leg checks passed.
              `SPECIFICATION.md § Journeys with several legs` names this as
              what happens in the absence of an explicit rule, which is to say
              it is the default behaviour rather than a mistake somebody has
              to make.
Rationale:    The obligation is the allocation's existence and its timing, not
              its rule: how a share is sized is FR-051's, what an unused share
              does is FR-052's, and what the sum must satisfy is FR-053's. The
              second criterion states the timing the section makes explicit,
              and it is not redundant — an allocation computed lazily as each
              leg is reached lets an early leg spend against a budget the
              later legs have not yet claimed from, which is the same failure
              arriving one leg at a time.
Resolved-by:  —
```

## FR-051 — Size each leg's share in proportion to its baseline driving time

```
Statement:    The system should size each leg's share of a journey's
              additional-time budget in proportion to that leg's baseline
              driving time.
Category:     Multi-leg budget allocation
Source:       SPECIFICATION.md § Journeys with several legs
Priority:     SHOULD
Verification: test — on a journey of two legs whose baseline driving times
              stand in a given ratio, the shares allocated to them stand in
              the same ratio
Acceptance:   given a journey of two legs whose baseline driving times stand
              in a given ratio, when the budget is allocated, then the shares
              allocated to them stand in that same ratio
              given a journey of two legs of equal baseline driving time, when
              the budget is allocated, then the two shares are equal
Status:       draft
Depends-on:   FR-050
Rationale:    A relation and not a formula: the record states the proportion
              the section defines and fixes no figure, so a change in how
              baseline driving time is obtained leaves it true. Baseline
              driving time is the quantity
              `CalculationSpecification.md § Additional journey time` already
              assigns to the route provider, so nothing new is introduced by
              naming it. The verb follows the section's should: an equal split
              still respects the journey-level target and ceiling, which
              FR-053 binds, so a deviation here costs result quality rather
              than the promise made to the user. Both criteria are stated at
              allocation time, before FR-052 redistributes anything — after
              pooling the shares no longer stand in the baseline ratio, and a
              criterion measured then would fail against a correct
              implementation.
Resolved-by:  —
```

## FR-052 — Return an unused share to a pool for the remaining legs

```
Statement:    The system should return the unused remainder of a leg's share
              of the additional-time budget to a pool available to the legs
              not yet optimized.
Category:     Multi-leg budget allocation
Source:       SPECIFICATION.md § Journeys with several legs
Priority:     SHOULD
Verification: test — a journey whose earlier leg cannot use its share yields a
              later leg able to draw beyond its own allocated share, bounded
              by the remainder that leg left
Acceptance:   given a journey of several legs in which one leg's optimization
              leaves part of its share unused, when a later leg is optimized,
              then that leg may draw additional time beyond its own allocated
              share, up to the remainder returned
              given a journey in which no leg leaves any part of its share
              unused, when a later leg is optimized, then it may draw no more
              than its own allocated share
Status:       draft
Depends-on:   FR-050; FR-051
Rationale:    The section's reason for the pool is a leg along which too few
              accessible zones exist, and it states the behaviour for the
              remaining legs — a leg already optimized is not revisited, so
              this record obliges no re-solve behind the current one. The verb
              follows the section's should: the pool improves how much of the
              budget a journey can use and takes nothing away from FR-053,
              which bounds the sum however it is distributed. Where the leg
              leaving the remainder is the last, there is no remaining leg to
              offer it to and the journey simply uses less than the user
              allowed; the section obliges nothing there and neither does this
              record.
Resolved-by:  —
```
