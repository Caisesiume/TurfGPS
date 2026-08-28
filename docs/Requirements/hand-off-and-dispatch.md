# Hand-off and dispatch

Delivering a chosen plan to the device and the navigation app that will drive it. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-109 — Hold each hand-off target's waypoint limits as configuration

```
Statement:    The system shall read each hand-off target's waypoint limits
              from configuration recording, for every limit, the date it was
              checked and the published source it was read from, rather than
              from constants in code.
Category:     Hand-off and dispatch
Source:       DESIGN.md § Dispatching stop by stop;
              SPECIFICATION.md § The waypoint limit problem
Priority:     MUST
Verification: inspection — the configuration the dispatch path reads each
              target's waypoint limits from, and the dispatch path itself,
              which holds none of them as a literal
Acceptance:   the configuration the dispatch path reads each hand-off target's
              waypoint limits from carries, with every limit, the date that
              limit was checked and the published source it was read from; a
              reader confirms this in that configuration and in the dispatch
              path that reads it, and confirms in
              `SPECIFICATION.md § The waypoint limit problem` that these are
              third-party figures carrying the date they were checked
              no waypoint limit for any hand-off target appears as a literal
              in code; a reader confirms this in the dispatch path
Status:       to-build
Depends-on:   none
Volatility:   settled — `DESIGN.md § Dispatching stop by stop` requires the
              limits be held as configuration, and
              `DESIGN.md § Open questions owned by this document` records
              their movement as a standing obligation rather than a question
Risk:         These figures are owned by a third party and change without
              reference to this product. Compiled into code, a change on the
              target's side becomes a release rather than a configuration
              update; and an out-of-date allowance is not a visible failure,
              because the surplus stops are dropped by the target without
              warning, so the product ships a hand-off that appears to work
              and delivers a plain drive.
Rationale:    The obligation is that the value is read from configuration and
              carries its documented origin, never that it is a particular
              figure, and no figure is named in the statement or in the
              criteria. The checked date is part of that origin rather than
              decoration: `SPECIFICATION.md § The waypoint limit problem`
              requires these limits be re-checked rather than assumed, and a
              configured figure with no date cannot be re-checked, only
              re-guessed.
Resolved-by:  —
```

## FR-110 — Size the first dispatch from the intermediate-waypoint allowance

```
Statement:    The system shall size each portion of a confirmed route, on the
              first attempt to dispatch it, at the target's
              intermediate-waypoint allowance in intermediate stops plus the
              one further stop the portion's destination slot carries, or,
              where no more than that many stops remain, at every stop that
              remains.
Category:     Hand-off and dispatch
Source:       DESIGN.md § Dispatching stop by stop
Priority:     MUST
Verification: test — a confirmed route with more stops remaining than one
              portion's whole length yields a first dispatch whose
              intermediate stops number the target's intermediate-waypoint
              allowance and whose stop count exceeds that allowance by exactly
              the destination slot, and routes with exactly that many stops
              remaining and with fewer each yield every stop that remains
Acceptance:   given a confirmed route with more stops remaining than the
              target's intermediate-waypoint allowance plus the one stop a
              portion's destination slot carries, when the next portion is
              first dispatched, then the portion's intermediate stops number
              that allowance and one further stop occupies the destination
              slot
              given a confirmed route with exactly that many stops remaining,
              when the next portion is first dispatched, then the portion
              carries every stop that remains and withholds none
              given a confirmed route with fewer than that many stops
              remaining, when the next portion is first dispatched, then the
              portion carries every stop that remains and withholds none
Status:       to-build
Depends-on:   FR-094;
              FR-109
Volatility:   unverified-fact — the arithmetic rests on the allowance counting
              intermediate stops only, which
              `SPECIFICATION.md § The waypoint limit problem` records as an
              inference from where the target states the cap rather than as a
              published guarantee
Risk:         Sized against a total rather than against the intermediates,
              every portion is one stop short and the driver returns to the
              app once more per portion for the whole length of the journey.
              Sized one stop longer than the allowance admits, the portion
              rests on a slot the published limit does not cover, and the
              product cannot detect what the target does with the surplus:
              `SPECIFICATION.md § Waypoints may be dropped without warning`
              records that a hand-off can appear to succeed while delivering a
              plain drive with the Turf stops removed, discovered by arriving.
Rationale:    The cited section sizes the portion at as many consecutive stops
              as the target accepts and states the arithmetic in the same
              breath — a portion whose destination slot holds a zone carries
              one Turf stop more than the allowance admits. That arithmetic is
              placed here, on the record that sizes the portion, so that a
              reader computing a portion's length reads one record rather than
              assembling it from two; FR-111 states only what the destination
              slot may carry, and the two together yield one length. The
              allowance itself is configuration under FR-109 and is cited
              rather than restated, the cited section requiring these
              third-party limits be held as configuration rather than as
              constants in code. The two branches are partitioned by the
              portion's whole length and not by the allowance alone, so more
              than that many remaining and that many or fewer divide the cases
              with no gap and nothing decided twice; the second criterion is
              the boundary itself, which the two branches previously left
              between them. The record binds the first attempt only. FR-112
              obliges a refused hand-off be attempted again carrying fewer
              stops, and read as binding every attempt this record would
              forbid that on every one of them, so FR-112 could never be
              satisfied at all — which is the test that decides it.
              `DECISIONS.md § RD-034` carries both halves.
Resolved-by:  —
```

## FR-111 — Carry the portion's last Turf stop in the destination slot

```
Statement:    Where a dispatched portion ends short of the journey's final
              destination, the system should place the portion's last Turf
              stop in the target's destination slot rather than in an
              intermediate slot.
Category:     Hand-off and dispatch
Source:       DESIGN.md § Dispatching stop by stop;
              SPECIFICATION.md § The waypoint limit problem
Priority:     SHOULD
Verification: test — a portion ending at a zone hands off with that zone as
              the target's destination and the stops before it as
              intermediates, so the intermediates it consumes number one fewer
              than the Turf stops the portion carries
Acceptance:   given a dispatched portion whose last stop is a Turf zone rather
              than the journey's final destination, when the portion is handed
              off, then that zone occupies the target's destination slot and
              only the stops before it consume the intermediate-waypoint
              allowance
              given a dispatched portion whose last stop is the journey's
              final destination, when the portion is handed off, then that
              destination occupies the destination slot and this rule adds no
              stop to the portion
Status:       to-build
Depends-on:   FR-109; FR-110
Volatility:   unverified-fact — the intermediate-waypoint allowance counting
              intermediate stops only is an inference
              `SPECIFICATION.md § The waypoint limit problem` records as one,
              resting on where the target states the cap rather than on any
              sentence stating its scope
Rationale:    The cited design section makes the destination slot the reason a
              portion carries one Turf stop more than the
              intermediate-waypoint allowance admits; that arithmetic is
              FR-110's, and this record states only what the slot may hold, so
              the two are read together and neither restates the other. The
              verb is should rather than shall because the allowance's
              intermediates-only scope is an inference the source records as
              one: a target counting its own destination inside the cap makes
              this the wrong placement, and
              `SPECIFICATION.md § The waypoint limit problem` names such a
              conflicting convention already published upstream.
Resolved-by:  —
```

## FR-112 — Reduce a refused hand-off rather than abandon it

```
Statement:    Where a hand-off to the target is refused, the system shall
              attempt it again carrying fewer stops rather than abandon the
              hand-off.
Category:     Hand-off and dispatch
Source:       DESIGN.md § Dispatching stop by stop
Priority:     MUST
Verification: test — a refused hand-off is attempted again carrying fewer
              stops, and a hand-off still refused carrying a single stop is
              abandoned with the stored plan intact and the user told
Acceptance:   given a hand-off the target refuses, when the system responds,
              then it attempts the hand-off again carrying fewer stops than
              the refused attempt
              given a hand-off the target still refuses when it carries a
              single stop, when the system stops attempting, then the stored
              plan is unchanged and the user is told that the hand-off did not
              complete
Status:       to-build
Depends-on:   FR-110
Volatility:   settled — `DESIGN.md § Dispatching stop by stop` requires the
              dispatch path to degrade by sending fewer stops rather than
              failing outright
Risk:         The hand-off is made at the roadside, at the moment the user is
              ready to drive. A path that fails outright leaves them with a
              plan they cannot act on and nothing smaller to try, for a
              target-side limit a shorter portion would have satisfied.
Rationale:    The trigger is a refusal the system observes.
              `SPECIFICATION.md § Waypoints may be dropped without warning`
              records that where a target does not support waypoints the
              parameter is ignored rather than rejected and that the system
              cannot detect it, so the truncation
              `DESIGN.md § Dispatching stop by stop` names beside a rejection
              is read here as an observable one; the undetectable case is
              carried by FR-113 as disclosure instead, because a requirement
              to react to it would oblige a detection the sources say does not
              exist. The second criterion is the branch the source does not
              reach: the reduction terminates, and what it terminates at
              leaves the plan intact, which FR-114 obliges independently.
Resolved-by:  —
```

## FR-113 — Tell the user what a hand-off may drop, before it is made

```
Statement:    The system shall tell the user, before a hand-off is made, that
              the target may drop the Turf stops the hand-off carries without
              reporting it.
Category:     Hand-off and dispatch
Source:       SPECIFICATION.md § Waypoints may be dropped without warning
Priority:     MUST
Verification: test — the telling is presented before the hand-off is made, and
              is presented for a portion within the target's
              intermediate-waypoint allowance as well as for one that fills it
Acceptance:   given a portion of a stored plan ready to be dispatched, when
              the hand-off is offered, then the user is told, before the
              hand-off is made, that the target may drop the Turf stops it
              carries without reporting it
              given a portion within the target's intermediate-waypoint
              allowance, when the hand-off is offered, then that telling is
              presented for it too
Status:       to-build
Depends-on:   FR-110
Volatility:   settled —
              `SPECIFICATION.md § Waypoints may be dropped without warning`
              states the obligation and accepts the underlying constraint for
              the first release
Risk:         A hand-off can appear to succeed while delivering a plain drive
              to the destination with every Turf stop removed, discovered by
              arriving. For a product whose entire value is the stops, an
              undisclosed hand-off is the most expensive failure it can ship:
              the user has lost the journey and has no reason to suspect the
              product of it.
Rationale:    The obligation is the telling and its timing; how it is worded
              and where it appears is an interface question
              `SPECIFICATION.md § Waypoints may be dropped without warning`
              explicitly leaves unsettled, and is not authored here. The
              second criterion is what keeps the telling honest: the drop is a
              property of which Maps product the user's tap reaches and not of
              how full the portion is, so a statement shown only for a full
              portion would say least in exactly the case the user has most
              reason to trust. This is the stance
              `SPECIFICATION.md § Confidence and uncertainty` sets for an
              estimate, applied to a hand-off. The quantity that criterion
              names is the registered noun, `intermediate-waypoint allowance`,
              and not a second spelling of it: `DECISIONS.md § RD-012` settled
              the register row this record now writes verbatim.
Resolved-by:  —
```

## FR-114 — Leave the stored plan unchanged by a hand-off

```
Statement:    The system shall leave a stored plan's zones, their order and
              the state retained with it unchanged by any hand-off of a
              portion of that plan.
Category:     Hand-off and dispatch
Source:       DESIGN.md § Dispatching stop by stop
Priority:     MUST
Verification: test — a stored plan is unchanged after a hand-off that
              succeeds, after one the target refuses, and after one that is
              abandoned
Acceptance:   given a portion of a stored plan dispatched successfully, when
              the stored plan is next opened, then its zones, their order and
              the state retained with it are as they were before the hand-off
              given a hand-off that was refused or abandoned, when the stored
              plan is next opened, then its zones, their order and the state
              retained with it are as they were before the hand-off
Status:       to-build
Depends-on:   FR-094;
              FR-096
Volatility:   settled — `DESIGN.md § Dispatching stop by stop` states that the
              plan itself is never at risk in a hand-off because it lives in
              this product
Risk:         FR-112's fallback has nothing to fall back to if a hand-off can
              consume or mutate the plan, and the loss lands at the roadside,
              mid-journey, on the one copy of the most expensive computation
              the product performs.
Rationale:    The record is what makes this product the holder of the plan
              rather than an exporter of it: dispatching a portion is a read.
              It says nothing about how far along the journey the user has
              got, which is not plan state.
              `SPECIFICATION.md § Handing off the confirmed route` rules out
              live position tracking, and FR-110 sizes the portion offered for
              dispatch from the stops that remain without obliging where the
              count of those already dispatched is held — so this record
              bounds what a hand-off may change in the stored plan and leaves
              that question to whatever record answers it, rather than
              deciding it here by implication.
Resolved-by:  —
```

## FR-115 — Offer hand-off only to the target named for the first release

```
Statement:    The system shall offer, as the means of sending a stored plan to
              a navigation application, only the target named for the first
              release under
              `SPECIFICATION.md § Handing off the confirmed route`.
Category:     Hand-off and dispatch
Source:       SPECIFICATION.md § Handing off the confirmed route
Priority:     MUST
Verification: test — the hand-off options offered for a stored plan name that
              one target, and offer no other navigation application and no
              file-based export
Acceptance:   given a stored plan ready to be driven, when the hand-off
              options are presented, then the only navigation application
              offered is the one named for the first release under
              `SPECIFICATION.md § Handing off the confirmed route`
              given the same stored plan, when the hand-off options are
              presented, then no other navigation application and no
              file-based export is offered
Status:       to-build
Depends-on:   none
Volatility:   settled — `SPECIFICATION.md § Handing off the confirmed route`
              names one target for the first release and defers the rest
Risk:         A second target is a second dispatch model rather than a second
              link: the one documented alternative accepts a single
              destination and no waypoints at all, so supporting it means
              building a per-stop hand-off beside the portion model, and the
              release spends its budget on a path the specification defers.
Rationale:    The target is cited rather than named, on the rule that this
              corpus does not become a second home for a fact one document
              owns; whoever builds the dispatch reads the name there. The
              documented alternative is carried upstream for its constraint
              and is explicitly out of scope, as are other applications and
              file-based export. That the product provides no turn-by-turn
              guidance, no live position tracking and no zone-arrival prompts
              is a further boundary of the same section and is not authored
              here.
Resolved-by:  —
```

## FR-118 — Dispatch each Turf stop at its stopping position

```
Statement:    The system shall use, as the waypoint dispatched for each Turf
              stop of a portion, that stop's stopping position rather than the
              zone's coordinate.
Category:     Hand-off and dispatch
Source:       DESIGN.md § Dispatching stop by stop;
              SPECIFICATION.md § What establishes a stopping position
Priority:     MUST
Verification: test — a portion whose Turf stops carry stopping positions
              distinct from their zones' coordinates hands off carrying those
              positions in the intermediate slots and in the destination slot,
              and the same portion handed off carrying any of those zones'
              coordinates fails
Acceptance:   given a dispatched portion whose Turf stops each carry a
              stopping position distinct from that zone's coordinate, when the
              portion is handed off, then the waypoint dispatched for each of
              those stops is its stopping position and no zone's coordinate is
              dispatched
              given a dispatched portion whose last stop is a Turf zone
              occupying the target's destination slot, when the portion is
              handed off, then the point dispatched in that slot is that
              stop's stopping position
Status:       to-build
Depends-on:   FR-092;
              FR-110
Volatility:   settled — the stopping position is already the point every
              access and exclusion decision is taken at, FR-086, FR-089 and
              FR-092 fixing it there, and `DECISIONS.md § RD-028` records the
              ruling that dispatch carries the same point
Risk:         The zone's coordinate is a point nothing in this corpus
              established as stoppable. FR-089 refuses a stopping position on
              a road the exclusions refuse and FR-092 refuses a manufactured
              one, and both checks were run against the other point — so a
              hand-off dispatching the coordinate sends the driver to a place
              the exclusions may refuse and the terrain may not admit, and
              does it after the product has priced the journey as though it
              had not. The estimate and the destination disagree, with nothing
              on screen inconsistent with anything else, and the disagreement
              is invisible until arrival, at the roadside, on a product whose
              whole value is the stops.
Rationale:    `DESIGN.md § Dispatching stop by stop` sizes a portion in stops
              and never says what a stop resolves to on the wire; this record
              closes that gap rather than reading it out of that section. What
              decides it is already filed: FR-086 routes a detour's cost
              through the proposed stopping position, FR-089 refuses one on an
              excluded road, and FR-092 obliges one be identified before any
              check about the stop is evaluated — so the stopping position is
              where every question about whether the stop can be made has been
              answered, and dispatch is the one place that chain reaches the
              driver. The second criterion is not the first restated: the
              destination slot is a different field on the wire from the
              intermediate slots, FR-111 puts a Turf zone in it, and it is the
              slot an implementation reaches for the zone's coordinate in. The
              journey's own origin and final destination are not Turf stops
              and are untouched here.
Resolved-by:  —
```
