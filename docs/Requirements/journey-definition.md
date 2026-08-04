# Journey definition

What a journey is: the destinations a request must and may name, the order they are visited in, their preservation through optimization, and the mode of travel between them. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-001 — Plan a journey from an origin to a destination

```
Statement:    The system shall plan a journey specified by a single origin and
              a single destination.
Category:     Journey definition
Source:       SPECIFICATION.md § Primary use case
Priority:     MUST
Verification: test — a request naming only an origin and a destination, and a
              request also naming intermediate destinations, each yields a
              journey whose first location is the entered origin and whose
              last is the entered destination
Acceptance:   given a request naming an origin and a destination, when
              planning runs, then every journey the system produces begins at
              the origin and ends at the destination
Status:       to-build
Depends-on:   none
Risk:         A → B is the design centre of the whole product; every later
              requirement presupposes a journey exists to attach zones to.
Rationale:    Kept separate from FR-002 on singularity, not on verb strength:
              planning A → B and planning a journey with intermediate
              destinations are two separately testable capabilities, and
              `SPECIFICATION.md § Initial product boundaries` lists them as
              two scope items — one origin and one required destination, and
              optional ordered intermediate destinations. Both statements have
              read `shall` since the Owner's ratification of 1 August 2026, so
              the verbs no longer carry the distinction; merging the records
              would join two separately testable outcomes in one statement,
              which the singular check rejects.
Resolved-by:  #1, #2
```

## FR-002 — Plan a journey with intermediate destinations

```
Statement:    The system shall plan a journey specified by an origin, a
              destination, and one or more intermediate destinations entered
              by the user.
Category:     Journey definition
Source:       SPECIFICATION.md § Primary use case
Priority:     MUST
Verification: test — a request naming at least one intermediate destination is
              planned rather than refused, and the produced journey passes
              through that destination
Acceptance:   given a request naming an origin, a destination and at least one
              intermediate destination, when planning runs, then the system
              plans the journey and the produced journey passes through the
              intermediate destination
Status:       to-build
Depends-on:   FR-001
Risk:         Built wrong or omitted, a capability the first release committed
              to (`SPECIFICATION.md § Initial product boundaries`) is silently
              dropped, and `SPECIFICATION.md § Journeys with several legs` is
              left with no subject — per-leg budget allocation, remainder
              pooling and the journey-level ceiling become unimplementable and
              untestable — so a user with a stop en route falls back to
              chaining separate single-leg plans, each granted the full
              additional-time allowance, defeating the journey-level ceiling
              outside any place the system can enforce it.
Rationale:    The source writes "should also support", but
              `SPECIFICATION.md § Initial product boundaries` commits optional
              ordered intermediate destinations to the first release, and
              FR-003 and FR-004 place strictly-obliged duties on the waypoints
              this record admits. An expectation-obliged precondition carrying
              strictly-obliged consequences is a strength inversion, so the
              verb is `shall` on the Owner's ratification of 1 August 2026.
              `DESIGN.md § Required and optional inputs` lists intermediate
              waypoints as optional *input* — optional to supply, never
              optional to support, which is the distinction the earlier
              `should` blurred.
Resolved-by:  #2
```

## FR-003 — Visit mandatory waypoints in the order entered

```
Statement:    The system shall visit a journey's mandatory waypoints in the
              order in which the user entered them.
Category:     Journey definition
Source:       SPECIFICATION.md § Primary use case
Priority:     MUST
Verification: test — a journey whose waypoints would be reached sooner in a
              different order is still produced in the entered order
Acceptance:   given a journey whose mandatory waypoints were entered in a
              given order, when planning runs, then every journey the system
              produces reaches them in that order
              given a journey whose mandatory waypoints would be reached in
              less time in a different order, when planning runs, then the
              produced journey still reaches them in the entered order
Status:       to-build
Depends-on:   FR-002
Risk:         Reordering silently substitutes a different journey for the one
              the user needs to make, and the substitution is invisible until
              they are driving it.
Resolved-by:  #2
```

## FR-004 — Preserve the entered mandatory-waypoint set

```
Statement:    The system shall produce each journey such that the journey's
              mandatory-waypoint set is identical to the set of mandatory
              waypoints the user entered.
Category:     Journey definition
Source:       SPECIFICATION.md § Primary use case
Priority:     MUST
Verification: test — a journey holding a mandatory waypoint that raises
              journey cost, and one holding a mandatory waypoint with a
              better-scoring nearby alternative, both yield plans whose
              mandatory-waypoint set compares equal to the entered set
Acceptance:   given an entered mandatory waypoint that raises the journey's
              cost relative to routing without it, when the system produces
              journey alternatives, then each alternative's mandatory-waypoint
              set compares equal to the entered set, the costly waypoint
              having not been dropped
              given an entered mandatory waypoint that has a nearby location
              the optimizer scores more favourably, when the system produces
              journey alternatives, then each alternative's mandatory-waypoint
              set compares equal to the entered set, the nearby location
              having not been substituted for it
Status:       to-build
Depends-on:   FR-002
Risk:         A plan that quietly drops or relocates a required destination is
              not the journey the user has to make, and the product's whole
              premise is optimizing inside a trip they already need to take.
Rationale:    Set identity, not inclusion — "identical" is what makes swapping
              in a nearby location a failure, which is the source's "replace".
Resolved-by:  #3
```

## FR-005 — Require a destination

```
Statement:    The system shall not plan a journey for a request that does not
              name a destination.
Category:     Journey definition
Source:       SPECIFICATION.md § Primary use case;
              SPECIFICATION.md § Genuinely out of reach or out of scope
Priority:     MUST
Verification: test — a request carrying an origin and a time budget but no
              destination produces no journey
Acceptance:   given a request naming an origin and a time budget but no
              destination, when the user attempts to start planning, then the
              system plans no journey
Status:       to-build
Depends-on:   none
Risk:         Open-ended trip generation is a materially larger optimization
              problem than the bounded one this release solves; admitting it
              by omission changes the product's scope without a decision.
Rationale:    The source qualifies the exclusion as "at least initially", and
              `SPECIFICATION.md § Genuinely out of reach or out of scope`
              repeats it as a deliberate product boundary. The corresponding
              positive branch — a request that does name a destination is
              planned — is FR-001's, and is not restated here.
Resolved-by:  #4
```

## FR-006 — Plan journey travel as travel by car

```
Statement:    The system shall plan all travel within a journey as travel by
              car, except the walk between a stopping location and a zone, per
              *Park-and-walk zones* in `SPECIFICATION.md`.
Category:     Journey definition
Source:       SPECIFICATION.md § Primary use case
Priority:     MUST
Verification: test — every leg of a produced journey other than a
              park-and-walk leg is routed for a car, and a leg routed over a
              path a car cannot use fails
Acceptance:   given any journey, when its legs are produced, then every leg
              other than the walk between a stopping location and a zone is
              routed for a car
              given a candidate leg traversable only on foot or by bicycle,
              when legs are produced, then it is not offered as a driving leg
Status:       to-build
Depends-on:   none
Risk:         A leg routed on a path a car cannot use yields a journey the
              user cannot drive and a cost estimate that means nothing, and
              the error is invisible until the user is at the roadside.
Rationale:    `SPECIFICATION.md § The planning player` owns the premise — the
              primary mode is a car, and the product does not generate cycling
              routes. The consequence it draws for urban zones is owned by
              `SPECIFICATION.md § Individual zones rather than local collection routes`
              and is not recorded here. The stated exception keeps this record
              from contradicting the park-and-walk model.
              `Architecture.md § D3` confirms feasibility: one routing engine
              serves both car and pedestrian costing.
Resolved-by:  #6
```

## FR-012 — Require an intermediate destination on a return to the origin

```
Statement:    The system shall not plan a journey for a request that names a
              single location as both its origin and its destination and
              names no intermediate destination distinct from that location.
Category:     Journey definition
Source:       SPECIFICATION.md § Genuinely out of reach or out of scope
Priority:     MUST
Verification: test — a request naming one location as both its origin and its
              destination produces no journey when it names no intermediate
              destination, and none when every intermediate destination it
              names is that location; the same request naming an intermediate
              destination distinct from that location is not refused for
              having named its origin as its destination
Acceptance:   given a request naming a single location as both its origin and
              its destination and naming no intermediate destination, when
              the user attempts to start planning, then the system plans no
              journey
              given a request naming a single location as both its origin and
              its destination and naming at least one intermediate
              destination, none of them distinct from that location, when the
              user attempts to start planning, then the system plans no
              journey
              given a request naming a single location as both its origin and
              its destination and naming at least one intermediate
              destination distinct from that location, when the user attempts
              to start planning, then the request is not refused for having
              named its origin as its destination
Status:       to-build
Depends-on:   FR-002
Risk:         Admitted, such a request has a zero-length journey to price
              against, so every minute driven counts as additional time, and
              the additional-time target and the absolute ceiling then bound
              the trip's whole length rather than what Turf adds to it — the
              open-ended generation this release excludes, arrived at by
              omission rather than by decision.
Rationale:    Refusing a request whose named destination is its origin reads
              as arbitrary without the boundary it belongs to. The named
              intermediate destination is what separates driving out and
              coming back — a journey with a *there* to drive out to, and so a
              real baseline to price the additional time against — from
              driving around and coming back, which the cited section keeps
              excluded. That baseline is also why the intermediate must name
              a location distinct from the origin, on the Owner's ruling of
              1 August 2026: an intermediate equal to the origin is no *there*
              to drive out to, so the base route is still zero-length and the
              whole journey still prices as additional time — a wording
              requiring only an intermediate destination admits A → A → A
              unchanged. The cited section now carries that qualifier itself,
              amended upstream on 2 August 2026: the drift closed by the
              source moving to the record's position, never by this record
              being edited to match its source. The condition is on what the
              request *names*: an intermediate is distinct when it names a
              different location, and the test is exact identity rather than
              proximity — a tolerance for an intermediate lying *near* the
              origin would be a constant with no home in
              `CalculationSpecification.md`, and introducing one here would
              create it. The third criterion states only that such a request
              is not refused, never that it is planned: planning a journey
              with intermediate destinations is FR-002's obligation, and
              FR-005's Rationale sets the precedent that a positive branch
              stays on the record that owns it.
Resolved-by:  #5
```
