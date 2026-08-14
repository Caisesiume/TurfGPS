# Access classification

Deciding whether and how a zone can be reached and stopped at. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-067 — Assign exactly one access class to every classified candidate

```
Statement:    The system shall assign exactly one access class — directly
              road-accessible, park-and-walk, uncertain, or excluded — to
              every candidate zone whose access classification completes.
Category:     Access classification
Source:       SPECIFICATION.md § Candidate zone identification
Priority:     MUST
Verification: test — every promoted candidate carries exactly one of the four
              classes once classification completes, and a candidate for which
              the data available establishes no class carries uncertain or
              excluded rather than none
Acceptance:   given a candidate promoted to full evaluation, when its access
              classification completes, then it carries exactly one of
              directly road-accessible, park-and-walk, uncertain and excluded
              given a candidate for which the data available establishes no
              class, when its access classification completes, then it carries
              uncertain or excluded rather than no class at all
Status:       to-build
Depends-on:   FR-064
Volatility:   settled — the source says at least two broad access types and
              the Owner closed the set at four on 7 August 2026, so what this
              record rests on is a ruling rather than a figure or an open
              question
Risk:         Every gate downstream keys on the class: uncertain never enters
              the cost model, excluded never reaches the user, and the two
              confident classes are priced by different models. A candidate
              carrying no class is read by whichever stage meets it first, and
              the class it falls into is the confident one, because that is
              the path everything else is written for. A candidate carrying
              two is priced twice and validated once.
Rationale:    The four classes are the terminals of the flowchart under
              `CalculationSpecification.md § Stop time`, and the source's at
              least two broad access types is the looser statement of the same
              thing; the Owner's ruling closes the set and this record is
              where the closure is stated. The second criterion is what the
              measure of success under
              `SPECIFICATION.md § Accessibility scope for the first release`
              turns on — not that every zone is classified, but that none is
              classified confidently and wrongly — so an unclassifiable
              candidate must land somewhere honest rather than nowhere. Which
              class a candidate reaches is decided by FR-068 and the records
              after it; this one forbids none and two.
Resolved-by:  —
```

## FR-068 — Choose the validation regime from the direct-access tolerance

```
Statement:    The system shall decide which access validation a candidate must
              pass from whether its coordinate lies within the direct-access
              tolerance defined under
              `CalculationSpecification.md § Direct-access tolerance` of an
              established stopping position on a drivable way.
Category:     Access classification
Source:       SPECIFICATION.md § Directly road-accessible zones
Priority:     MUST
Verification: test — a candidate inside the tolerance is checked for level
              compatibility and intervening barriers and is not required to
              produce a walkable path or an elevation profile, and a candidate
              outside it is required to produce both
Acceptance:   given a candidate whose coordinate lies within the direct-access
              tolerance of an established stopping position on a drivable way,
              when its access is validated, then it is checked for level
              compatibility and intervening barriers and is not required to
              produce a connected walkable path or an elevation profile
              given a candidate whose coordinate lies beyond that tolerance
              from every established stopping position, when its access is
              validated, then it is required to produce a connected walkable
              path and an obtainable elevation profile
Status:       to-build
Depends-on:   FR-067; FR-092
Volatility:   proposed-constant — the direct-access tolerance under
              `CalculationSpecification.md § Direct-access tolerance` is a
              proposed default, argued from a nominal figure
              `Architecture.md § Zone geometry` declines to rely on, and which
              candidates meet which regime moves with it
Risk:         What this tolerance admits is the set of zones that bypass the
              walk-safety gates entirely. Read as a choice between two cost
              models rather than two validations, a candidate just inside it
              is priced from the seat with no path, no profile and no gradient
              ever examined — and where that judgement is wrong the driver
              stops, finds they are not inside the zone, gets out, and becomes
              a pedestrian on ground no part of the system validated.
Rationale:    The cited calculation section states outright that the constant
              separates two validation regimes and not two cost models, and
              the flowchart under `CalculationSpecification.md § Stop time` is
              the precise statement of both branches. This record binds the
              split alone: what the direct branch must find is FR-070's and
              FR-089's, what happens when it does not is FR-071's and
              FR-072's, and what the walking branch must find is FR-076's. The
              position the tolerance is measured from is an established one,
              as `SPECIFICATION.md § What establishes a stopping position`
              defines establishment; valid was a third spelling of that
              predicate and is dropped rather than defined, because a
              tolerance measured from a position nothing constrains is
              measured from whatever drivable way lies nearest the coordinate,
              and the identification of the position itself is FR-092's. The
              tolerance is cited and never quoted, so the record stands at
              whatever value a deployment enforces — and that the value may be
              tightened and never loosened is FR-069's.
Resolved-by:  —
```

## FR-071 — Send an established level or barrier failure to the park-and-walk branch

```
Statement:    Where the data establishes that a candidate within the
              direct-access tolerance is at a level the stopping position does
              not meet, or that a barrier lies between them, the system shall
              evaluate the candidate on the park-and-walk branch rather than
              exclude it on that finding.
Category:     Access classification
Source:       SPECIFICATION.md § Directly road-accessible zones
Priority:     SHOULD
Verification: test — a candidate inside the tolerance separated from the road
              by a recorded barrier is evaluated for a walkable route and an
              elevation profile, and reaches park-and-walk, uncertain or
              excluded on what that evaluation finds rather than on the failed
              check, while one whose level relationship cannot be resolved is
              not routed to that branch by this rule
Acceptance:   given a candidate within the direct-access tolerance whose
              bridge, tunnel or layer data establishes that the stopping
              position and the zone do not meet, or for which the data records
              a barrier between them, when its access is classified, then it
              is evaluated for a connected walkable path and an obtainable
              elevation profile
              given the same candidate, when that evaluation completes, then
              its class follows from what the evaluation found and not from
              the failed direct-access check
              given a candidate within the direct-access tolerance whose level
              relationship to the stopping position cannot be resolved from
              the data, when its access is classified, then this rule does not
              route it to the park-and-walk branch
Status:       to-build
Depends-on:   FR-068; FR-070; FR-072; FR-092
Volatility:   proposed-constant — which candidates can reach this branch at
              all is decided by the direct-access tolerance under
              `CalculationSpecification.md § Direct-access tolerance`, a
              proposed default
Risk:         Excluding on this failure loses exactly the zones a conservative
              tolerance is set to catch: a zone beside a road but behind a low
              wall is an ordinary park-and-walk stop, and a classifier that
              discards it reports a clean, confident result while the coverage
              the product is judged on quietly falls. An excluded zone leaves
              no trace in the output, so nothing surfaces the loss. Widened
              past the established case the same rule runs the other way and
              is worse: it walks a candidate whose relationship to the road
              was never resolved, and a validated path then yields a confident
              class with the level question still open.
Rationale:    The cited section states that a zone is not lost by falling
              outside the direct test — it is examined more strictly — and the
              flowchart under `CalculationSpecification.md § Stop time` routes
              a failed level check into the same branch a candidate beyond the
              tolerance enters. The Owner ruled on 7 August 2026 that a failed
              direct-access test downgrades and never excludes on that failure
              alone. The antecedent is the case the data establishes, which is
              narrower than the complement of FR-070's admission test: that
              record admits the direct class only where the levels are
              established compatible, so an unresolved relationship fails it
              too, and an antecedent reading "fails the check" would send the
              unresolved case here — to a walking evaluation that can end in a
              confident class with the level relationship never resolved.
              FR-072 owns that case alone and the third criterion is the seam
              between the two. The second criterion is what makes the first
              more than a formality: an implementation can run the walking
              evaluation and still carry the failed check forward as a veto.
              The position in the antecedent is the candidate's identified
              stopping position, which is why `Depends-on` names FR-092;
              FR-070 carries the argument for evaluating the level check there
              and it is not repeated here. `Risk` is present on a should
              because this is an access-classification record and therefore a
              safety path, per `safety-path-checklist`.
Resolved-by:  —
```

## FR-072 — Classify uncertain where the direct-access evidence is ambiguous

```
Statement:    Where elevation or structural data leaves the relationship
              between a stopping position and a zone ambiguous, the system
              shall classify the zone as uncertain rather than as directly
              road-accessible.
Category:     Access classification
Source:       SPECIFICATION.md § Direct road-access validation
Priority:     MUST
Verification: test — a candidate whose bridge or tunnel relationship to the
              road cannot be resolved from the data is classified uncertain
              and not directly road-accessible, while one whose relationship
              is resolved and compatible is not made uncertain by this rule
Acceptance:   given a candidate within the direct-access tolerance whose
              bridge, tunnel or layer relationship to the stopping position
              cannot be resolved from the data available, when its access is
              classified, then it is classified uncertain and not directly
              road-accessible
              given a candidate whose level relationship to the stopping
              position is resolved and compatible, when its access is
              classified, then this rule does not make it uncertain
Status:       to-build
Depends-on:   FR-067; FR-070; FR-092
Volatility:   settled — the downgrade on ambiguous evidence is stated outright
              by the cited section and rests on no figure
Risk:         Ambiguous data resolved in the system's favour is the exact
              shape of a zone classified confidently and wrongly, which is the
              single failure the product's measure of success names. It
              arrives with no symptom: the classifier reports a directly
              road-accessible zone, the driver is told to take it from the
              car, and only the arrival shows that the road and the zone never
              met.
Rationale:    The cited section states this outright, and it acts on different
              evidence from FR-070: that record fails a candidate on data
              showing the road and the zone do not meet, this one on data
              showing neither. It is distinct from FR-071 too — a downgrade to
              the park-and-walk branch is an evaluation, while an unresolved
              level relationship is a missing fact, and walking the candidate
              would validate the very ground the ambiguity is about. Uncertain
              rather than excluded because the zone may be perfectly
              reachable; what an uncertain classification then means for the
              search is FR-084's. The position whose relationship is in
              question is the candidate's identified stopping position, which
              is why `Depends-on` names FR-092; FR-070 carries the argument
              for evaluating the level relationship there and it is not
              repeated here.
Resolved-by:  —
```

## FR-073 — Measure a candidate's access route to the zone's coordinate

```
Statement:    The system shall measure a candidate's access route to the
              zone's coordinate, and shall not treat a route ending at the
              nearest mapped position to that coordinate as having reached the
              zone.
Category:     Access classification
Source:       SPECIFICATION.md § Access-path validation
Priority:     MUST
Verification: test — a zone whose coordinate lies away from the nearest mapped
              pedestrian path yields an access estimate carrying the remaining
              approach, and an estimate stopping at that path's nearest point
              to the coordinate fails
Acceptance:   given a candidate whose coordinate lies away from the nearest
              mapped pedestrian path, when its access route is measured, then
              the estimate carries the approach from that path to the
              coordinate
              given the same candidate, when its access is classified, then a
              routed path terminating at the nearest mapped position to the
              coordinate does not on its own establish that the zone is
              reached
Status:       to-build
Depends-on:   FR-020; FR-074
Volatility:   settled — the coordinate is the only position established to lie
              inside a zone, and `Architecture.md § Zone geometry` records
              that the API exposes nothing about any zone's shape or size,
              which is a verified fact rather than a proposal
Risk:         A route measured to the nearest mapped node understates every
              walk by the last and hardest part of it — the part with no path,
              which is exactly where the fence, the embankment and the ditch
              are. It understates it in the direction that makes the zone look
              affordable, and the shortfall grows precisely as the terrain
              gets worse.
Rationale:    The cited section requires the route to connect the stopping
              point to a position inside the zone rather than merely ending at
              the geographically closest map coordinate. The statement binds
              the coordinate because that is the only position the data
              establishes to be inside the zone: no per-zone extent may be
              derived, per FR-020, and
              `SPECIFICATION.md § The coordinate is the target` is where the
              system's answer to that lives. Nothing here obliges a mapped
              path all the way to the coordinate — an unmapped final approach
              is ordinary — and what such an approach costs the estimate's
              confidence is FR-081's and what must then be said of it is
              FR-088's.
Resolved-by:  —
```

## FR-075 — Build no access path from a way the data bars to pedestrians

```
Statement:    The system shall not include in a candidate's access path a way
              the map data marks as not usable by a pedestrian or as
              access-restricted.
Category:     Access classification
Source:       SPECIFICATION.md § Access-path validation
Priority:     MUST
Verification: test — a candidate whose only mapped connection to a stopping
              position runs along a way the data bars to pedestrians is not
              classified park-and-walk on that connection, and a candidate
              with an alternative connection avoiding such a way is unaffected
Acceptance:   given a candidate whose only identified connection between a
              stopping position and the zone's coordinate runs along a way the
              map data marks as not usable by a pedestrian or as
              access-restricted, when its access is classified, then that
              connection does not establish an access path
              given a candidate for which a connection avoiding such a way
              exists, when its access is classified, then that connection is
              the access path and this rule excludes nothing
Status:       to-build
Depends-on:   FR-074
Volatility:   open-question — the access and restriction attributes this
              record reads are the OSM-derived feature data
              `Architecture.md § D4` decides the store holds, and
              `Architecture.md § Still owed by this document` records those
              tables as owed, undesigned and unreviewed rather than built
Risk:         A way that is mapped and routable is not thereby walkable: a
              hard shoulder, a service tunnel, a private drive and a fenced
              compound all route perfectly and all put a pedestrian somewhere
              they must not be. The estimate is then not merely wrong but
              confident and short, because the barred way is usually the
              shortest connection there is.
Rationale:    This is the obligation carried by the cited section's evaluation
              item on whether the path appears legally and physically usable
              by a pedestrian, and nothing else in the corpus reaches it:
              `SPECIFICATION.md § Enforceable exclusions` forbids constructing
              a route through areas the map data marks private or
              access-restricted, which binds the area and not the way, and
              `CalculationSpecification.md § Elevation-aware walking time`
              consumes surface and path type as a speed adjustment rather than
              as an admissibility test. The second criterion keeps this a rule
              about the path chosen rather than about the zone. What the data
              does not record it cannot establish, per
              `SPECIFICATION.md § Requirements the data cannot verify`, and
              this record claims nothing beyond what the data marks.
Resolved-by:  —
```

## FR-078 — Validate every leg of a multi-zone walking route

```
Statement:    The system shall establish that one stop may serve two zones
              only by validating the walking leg between them under the same
              rules it applies to the approach from the car.
Category:     Access classification
Source:       SPECIFICATION.md § Proximity between zones does not imply cheap chaining
Priority:     MUST
Verification: test — two zones lying close together but separated by a fenced
              railway are not served from one stop and are evaluated as two
              stops instead, while two zones with a validated leg between them
              may be served from one
Acceptance:   given two candidate zones between which no connected, usable
              walking leg can be validated, when stops are formed, then the
              two are not served from one stop and each is evaluated from its
              own stop
              given two candidate zones between which a walking leg is
              validated under the rules applied to the approach from the car,
              when stops are formed, then this record does not prevent one
              stop serving both
              given two candidate zones whose coordinates lie close together,
              when stops are formed, then their separation alone does not
              establish that one stop can serve both
Status:       to-build
Depends-on:   FR-021; FR-074
Volatility:   settled — the rule is stated outright by the cited section and
              rests on no figure and on nothing any document lists as open
Risk:         A straight-line reading makes barrier-separated pairs look like
              the cheapest combinations in the whole candidate set, so the
              optimizer does not merely mis-price them — it seeks them out and
              puts them at the top of the result. The driver parks once for
              two zones, finds a fenced railway between them, and has been
              given a time estimate for a walk that cannot be made.
Rationale:    The cited section states that the barrier reasoning governing
              access from a road applies equally between zones, and that a
              pair failing validation between them is simply not a chain —
              each is then evaluated from its own stop, possibly from opposite
              sides of the barrier on different parts of the drive. The third
              criterion is the failure that section warns about by name, and
              it is the one a proximity-based implementation commits while
              passing the first two. This is the second of the two levels at
              which sharing occurs: a shared driving deviation is established
              by routing under FR-086 and charged once by FR-008, and a shared
              stop is established here.
Resolved-by:  —
```

## FR-081 — Assign a confidence level to every access estimate

```
Statement:    The system shall assign a confidence level to every access
              estimate it produces for a candidate zone.
Category:     Access classification
Source:       SPECIFICATION.md § Terrain confidence
Priority:     MUST
Verification: test — every access estimate produced for a promoted candidate
              carries a confidence level, no estimate completes without one,
              and a candidate classified directly road-accessible carries one
              although its branch produced neither a path nor a profile
Acceptance:   given a candidate for which an access estimate is produced, when
              the estimate completes, then it carries a confidence level that
              everything downstream of the estimate can read
              given a candidate classified directly road-accessible, for which
              the direct branch produced no walkable path and no elevation
              profile, when its access classification completes, then an
              access estimate carrying a confidence level has been produced
              for it
Status:       to-build
Depends-on:   FR-067
Volatility:   open-question — no scale for access confidence exists:
              `CalculationSpecification.md` carries none, and what a level
              means rests on the enumeration under
              `SPECIFICATION.md § Terrain confidence` standing as a working
              default until one is authored
Risk:         Confidence is the gate the whole access model leans on — it
              decides the uncertain class, it decides what may be recommended
              at all, and it is the quantity the product's measure of success
              is stated in. An estimate carrying no confidence is read
              downstream as a confident one, because every consumer is written
              for the case where the data was there. Read as the walking
              branch's output alone it is worse than absent: the direct class
              then carries no confidence gate at any point, and that is the
              class admitted on a margin of a few metres.
Rationale:    The cited section obliges a confidence level per access estimate
              and enumerates the inputs that raise and lower one; the Owner
              ruled on 7 August 2026 that confidence is assigned per access
              estimate and that low-confidence access is what routes a zone
              into the uncertain class. This record carries the assignment and
              FR-082 carries the gate, because a system can assign a level
              correctly and gate on nothing, and can gate on the enumerated
              inputs while recording no level for anything downstream to read.
              The second criterion exists because the direct branch of the
              flowchart under `CalculationSpecification.md § Stop time`
              produces no path and no profile, so access estimate reads
              naturally as the walking estimate — and read that way the direct
              class would carry one gate and no confidence gate at all, while
              `SPECIFICATION.md § The coordinate is the target` states that
              where a classification turns on a margin of a few metres, that
              is itself grounds for treating the access as uncertain rather
              than confirmed, and the direct-access tolerance is a margin of a
              few metres by construction. The estimate a direct classification
              produces is its stop cost and the evidence behind it, which is
              an access estimate on the statement's own terms; the criterion
              states that rather than widening the statement, and one
              consequence of it is that FR-082 reaches the direct branch,
              which it otherwise would not. That the level must fall as the
              evidence thins was authored in both lanes and is NFR-006's, not
              this record's: the register puts confidence falling with thin
              data under `Coverage and data quality` by name, and a monotonic
              relation over a quality measure is that lane's. What survives
              here is the assignment alone, which nothing else obliges and
              which a system satisfying NFR-006's ordering can still fail by
              computing a level it never records. No scale for access
              confidence exists anywhere and the absence is returned with this
              batch. Confidence is a gate and never a term in the score, per
              `CalculationSpecification.md § Proposed form: value per minute`,
              and nothing here makes it one.
Resolved-by:  —
```

## FR-082 — Classify a low-confidence access estimate as uncertain

```
Statement:    The system shall classify a candidate whose access estimate
              carries low confidence as uncertain rather than as directly
              road-accessible, as park-and-walk, or as excluded on that
              confidence.
Category:     Access classification
Source:       SPECIFICATION.md § Terrain confidence;
              SPECIFICATION.md § What establishes a stopping position
Priority:     MUST
Verification: test — a candidate whose access estimate rests on any one of the
              lower-confidence inputs enumerated under
              `SPECIFICATION.md § Terrain confidence` is classified uncertain
              rather than park-and-walk and is not excluded on that
              confidence, including one resting on missing barrier information
              alone, and one resting on that section's high-confidence
              evidence throughout is not made uncertain by this rule
Acceptance:   given a candidate whose access estimate rests on any of the
              lower-confidence inputs enumerated under
              `SPECIFICATION.md § Terrain confidence`, when its access is
              classified, then it is classified uncertain and not directly
              road-accessible, park-and-walk, or excluded on that confidence
              given a candidate whose access estimate rests on missing barrier
              information alone, every other input behind it being of the
              high-confidence kind that section enumerates, when its access is
              classified, then it is classified uncertain and not directly
              road-accessible, park-and-walk, or excluded on that confidence
              given a candidate whose access estimate rests on a mapped
              stopping position, a connected pedestrian path, a complete
              elevation profile, known road and path levels and no identified
              barrier, when its access is classified, then this rule does not
              make it uncertain
Status:       to-build
Depends-on:   FR-067; FR-081
Volatility:   open-question — what counts as low confidence rests on the
              enumeration under `SPECIFICATION.md § Terrain confidence`
              standing as a working default: no scale and no threshold for
              access confidence exists in `CalculationSpecification.md`, and
              this record's boundary moves when one is authored
Risk:         This is the gate the measure of success is written about — not
              that every zone is classified, but that none is classified
              confidently and wrongly. Open it and the product's worst failure
              becomes its ordinary output: a rural candidate with no mapped
              path is offered as a priced park-and-walk stop, carrying a time
              the system does not believe, and the driver has no way to know
              which kind of estimate they are reading.
Rationale:    The cited section sends zones with low-confidence access out of
              time-sensitive recommendations or into separate presentation as
              uncertain opportunities, and
              `SPECIFICATION.md § Requirements the data cannot verify` states
              the same rule from the data's side: where what is available is
              insufficient to establish safe access, the zone is excluded or
              clearly classified uncertain rather than treated as a normal
              recommendation. That section left the choice between the two
              open and
              `SPECIFICATION.md § What establishes a stopping position`
              decides it: a position which is not established is uncertain,
              and uncertain is not excluded, so this record classifies and
              never excludes. Exclusion stays with the enforceable exclusions,
              which bind the candidate independently of this record and are
              FR-075's, FR-076's and FR-089's, and that independence is why
              the criteria refuse exclusion on the confidence rather than in
              general. The criteria are written against the inputs
              `SPECIFICATION.md § Terrain confidence` enumerates rather than
              against a confidence figure, because no scale exists to write
              one against and inventing one here would create a constant with
              no home. That absence is why the antecedent is disjunctive:
              while no scale exists these criteria are the operative
              definition of low confidence, so an antecedent conjoining two of
              the six enumerated inputs would leave the other four lowering
              confidence in nothing that is ever tested — missing barrier
              information most damagingly, being the same evidence class
              FR-070 tests on the direct branch, which is why the second
              criterion pins the single-item case there. The third criterion
              is the other side and keeps the widening from making the record
              unsatisfiable. Distinct from FR-083, whose antecedent is
              mechanical — which inputs the pricing actually consumed — where
              this one is a reading of the evidence behind the estimate: a
              system can route every fallback-priced candidate correctly and
              still call an incomplete elevation sample confident. What an
              uncertain classification then means for the search is FR-084's.
Resolved-by:  —
```

## FR-083 — Admit park-and-walk only on a routed path and an obtained elevation profile

```
Statement:    The system shall classify a candidate as park-and-walk only
              where its access estimate was computed from a routed walkable
              path and an obtained elevation profile, per
              `CalculationSpecification.md § Elevation-aware walking time`.
Category:     Access classification
Source:       CalculationSpecification.md § Stop time;
              CalculationSpecification.md § Flat-distance fallback
Priority:     MUST
Verification: test — a candidate whose walking cost was computed from a routed
              path with no elevation profile obtained is not classified
              park-and-walk, a candidate priced by the straight-line fallback
              is not either, and no park-and-walk stop in an alternative
              presented as a reliable recommendation was priced either way
Acceptance:   given a candidate for which a routed walkable path was
              identified but no elevation profile was obtained, and whose
              walking cost was computed from that path by the flat-ground
              calculation
              `CalculationSpecification.md § Elevation-aware walking time`
              states as its per-segment building block, when its access is
              classified, then it is not classified park-and-walk
              given a candidate priced by the straight-line estimate under
              `CalculationSpecification.md § Flat-distance fallback`, when its
              access is classified, then it is not classified park-and-walk
              given a journey alternative presented to the user as a reliable
              recommendation, when its stops are examined, then no
              park-and-walk stop among them was priced without an obtained
              elevation profile
Status:       to-build
Depends-on:   FR-074; FR-082
Volatility:   settled — the flowchart under
              `CalculationSpecification.md § Stop time` makes a connected path
              and an obtainable profile a conjunction on the branch that
              yields the park-and-walk class, and this record carries none of
              the figures either model is parameterized by
Risk:         Missing elevation tiles are commonest exactly where the terrain
              is worst, and a routed path is no evidence about the ground
              beside it: a hundred and eighty metres of mapped path up the
              side of a cutting, priced at its length divided by a walking
              speed, is a six-minute park-and-walk stop that enters the cost
              model and the ceiling as though it had been profiled. It is not
              a straight line, so nothing keyed on the fallback fires; the
              connectivity is certain, so nothing keyed on connectivity fires;
              and the estimate is systematically short in the direction that
              makes the zone win.
Rationale:    The obligation is positive by design. Stated as a prohibition on
              named estimates it reaches only the estimates someone thought
              of, and the corpus already holds a third —
              `CalculationSpecification.md § Elevation-aware walking time`
              gives the flat-ground calculation as its per-segment building
              block, and a routed length divided by a walking speed is neither
              the elevation-aware model nor the straight-line fallback. The
              enumeration-of-forbidden-estimates form always has one more
              estimate in it. Requiring the profile positively closes the
              class instead of chasing the estimates, and the flowchart under
              `CalculationSpecification.md § Stop time` is the authority for
              it: the branch yielding park-and-walk asks for a connected
              walkable path and an obtainable elevation profile together, and
              a candidate failing either leaves that branch. `Source` names
              the calculation sections because that is where the obligation is
              created, on FR-046's precedent. The fallback prohibition is kept
              as the second criterion rather than dropped, since a
              fallback-priced candidate is the case the cited fallback section
              marks low-confidence by name; that such a candidate may not hold
              the direct class either is FR-082's, whose enumeration begins
              with straight-line distance, and is not restated here. This
              record does not reach the direct class at all: that branch
              produces neither a path nor a profile by construction, and its
              own gates are FR-081, FR-089 and FR-070. Distinct from FR-082
              because the antecedents differ in kind — which inputs the
              pricing consumed against a reading of the evidence behind the
              estimate — so a system can satisfy either while failing the
              other, and this is the machine-checkable floor beneath the
              honesty of the access model, standing to it as FR-014 and FR-015
              stand to FR-016. Authored in this lane on `@requirements-nfr`'s
              referral and agreed: a prohibition on a behaviour is functional.
Resolved-by:  —
```

## FR-084 — Keep uncertain candidates out of the optimization

```
Statement:    The system shall not admit a candidate classified uncertain into
              a journey alternative's cost model, its Turf value, the ranking
              of alternatives, or the additional time it states.
Category:     Access classification
Source:       SPECIFICATION.md § Terrain confidence
Priority:     MUST
Verification: test — two corridors differing only in that one holds an
              uncertain candidate produce the same alternatives, the same
              values, the same ranking and the same stated additional times
Acceptance:   given two identical journeys whose corridors differ only in that
              one holds an uncertain candidate, when alternatives are produced
              for each, then the alternatives, their Turf values, their
              ranking and their stated additional times are the same for both
              given an uncertain candidate, when alternatives are scored and
              ranked, then it carries no score and contributes to no
              alternative's value
Status:       to-build
Depends-on:   FR-067; FR-082
Volatility:   settled — the exclusion from the cost model is stated absolutely
              by the sections defining the class, and it does not move with
              the reserve-pool feature, which those sections separately
              earmark for measurement
Risk:         A time estimate the system does not trust, balanced against ones
              it does, corrupts every comparison it enters and not only its
              own: the uncertain zone wins on an optimistic number, displaces
              a zone that was priced honestly, and the alternative's stated
              additional time is then a figure nobody can stand behind. The
              failure is invisible because the output is a perfectly ordinary
              route.
Rationale:    The cited section sends low-confidence access out of
              time-sensitive recommendations, and
              `SPECIFICATION.md § Handling the uncertain bucket` states the
              behaviour in full — never in the cost model, never contributing
              to value, no score, never influencing ranking, never inside a
              route's stated additional time — which is one prohibition on one
              admission and therefore stays one record.
              `CalculationSpecification.md § Proposed form: value per minute`
              gives the structural reason: confidence is a gate and never a
              term, because blending it into a score makes a well-understood
              mediocre zone and a poorly-understood excellent one
              indistinguishable. What uncertain zones are offered for instead
              — the reserve pool drawn on when the user rejects a zone, and
              the conservative upper bound that lets an accepted one be tested
              against the ceiling — is owed to the batch scoped to
              `SPECIFICATION.md § Route review and zone confirmation` and is
              deliberately not authored here.
Resolved-by:  —
```

## FR-085 — Classify a zone accessible only on an identified connection

```
Statement:    The system shall classify a zone as accessible only where it has
              identified a connection between a stopping position established
              under `SPECIFICATION.md § What establishes a stopping position`
              and the zone's coordinate.
Category:     Access classification
Source:       SPECIFICATION.md § Accessibility principle
Priority:     MUST
Verification: test — a candidate for which no connection between an
              established stopping position and its coordinate has been
              identified is not classified accessible, and neither is one
              lying within a configured radius of a road or a parking location
              with no such connection identified
Acceptance:   given a candidate for which no connection between an established
              stopping position and the zone's coordinate has been identified,
              when its access is classified, then it is not classified
              directly road-accessible or park-and-walk
              given a candidate whose coordinate lies within a configured
              radius of a road or a parking location and for which no such
              connection has been identified, when its access is classified,
              then it is not classified directly road-accessible or
              park-and-walk
Status:       to-build
Depends-on:   FR-067; FR-092
Volatility:   settled — the principle is stated as governing its whole section
              and as taking precedence over any narrower statement in it, and
              it rests on no figure
Risk:         Proximity is the measure that is always available and always
              cheap, which is why an implementation reaches for it when the
              path data is thin — and thin path data is exactly the condition
              under which proximity is worthless. A radius test returns a
              confident classification for the rural candidate above a cutting
              and for the one across a fence, and both read as ordinary
              results.
Rationale:    The cited section states that a zone is accessible only where
              the optimizer can identify a plausible, safe and sufficiently
              efficient connection between a legal stopping position and the
              zone's coordinate, and that this principle governs the whole of
              its section and takes precedence over any narrower statement in
              it. Legal is dropped rather than defined, and establishment is
              the qualifier the statement carries instead:
              `SPECIFICATION.md § What establishes a stopping position` names
              the legal stopping location of the cited principle among the
              spellings it governs, and states outright that establishment
              does not establish that stopping is lawful at the moment of
              arrival, so a criterion testing lawfulness would oblige a
              finding the data cannot support, while legal read as not marked
              restricted is unknown upgraded to permitted by defaulting — the
              failure that definition refuses by name. The obligation does not
              move with the qualifier: what changes is that the antecedent
              becomes decidable, and which positions may play the role the
              connection starts from is FR-092's. What this record carries is
              the identification: safe is carried by FR-070, FR-075, FR-076
              and FR-077, and sufficiently efficient by FR-080 together with
              the comparison FR-011 owns, so nothing here restates either and
              no part of the expression that section sets out is transcribed
              into a criterion — a record reproducing it would give that model
              a second home. The second criterion is the section's own
              prohibition, that a zone is not accessible merely because its
              centre lies within a configured radius of a road or parking
              location, and it is the branch this record fires on.
Resolved-by:  —
```

## FR-092 — Identify a stopping position rather than manufacture one

```
Statement:    The system shall identify a stopping position established under
              `SPECIFICATION.md § What establishes a stopping position` for
              every candidate it classifies as directly road-accessible or
              park-and-walk.
Category:     Access classification
Source:       SPECIFICATION.md § Direct road-access validation;
              SPECIFICATION.md § Access-path validation;
              SPECIFICATION.md § What establishes a stopping position
Priority:     MUST
Verification: test — a confidently classified candidate carries a stopping
              position identified before any check about its stop was
              evaluated, a candidate offered only the nearest drivable
              position to its coordinate with no stopping feature identified
              there is not confidently classified, and one for which no
              stopping position can be identified is classified uncertain
              rather than excluded
Acceptance:   given a candidate classified directly road-accessible or
              park-and-walk, when its access classification completes, then
              the classification carries a stopping position that was
              identified before any check about the stop was evaluated
              given a candidate whose only offered position is the position on
              a drivable way nearest its coordinate and at which the data
              carries no stopping feature, when its access is classified, then
              that position does not establish a stopping position and the
              candidate is not classified directly road-accessible or
              park-and-walk
              given a candidate for which no stopping position can be
              identified and which no enforceable exclusion removes, when its
              access is classified, then it is classified uncertain rather
              than excluded or priced from a position the system constructed
              for it
Status:       to-build
Depends-on:   FR-067
Volatility:   open-question — the cited definition names the kinds of feature
              that establish a position and leaves their identifying
              attributes to the OSM-derived feature tables
              `Architecture.md § D4` decides the store holds, which
              `Architecture.md § Still owed by this document` records as owed,
              undesigned and unreviewed rather than built
Risk:         A snapped coordinate is a stopping position the system invented,
              and one is available for every zone in the extract, so a
              classifier built on snapping never runs short of them. It
              defeats the exclusions through the definition of their own
              input: FR-076 excludes a candidate for which no connected
              walking route between a stopping position and the coordinate can
              be identified, and a position manufactured a few metres from
              that coordinate makes the route short and connected for a quarry
              rim as readily as for a car park. Nothing downstream can see it,
              because the position is a real place on a real way and every
              figure computed from it is plausible.
Rationale:    The obligation is positive by design. Stated as a prohibition on
              manufactured positions it reaches only the manufactures someone
              thought of, and the cited definition refuses four of them; the
              enumeration form always has one more member in it, which is the
              form FR-083 was re-keyed away from. The nearest point on a
              drivable way is kept as the second criterion because it is the
              manufacture an implementation reaches for first, not because it
              is the rule. The direct-access section requires evidence that
              the vehicle can stop at a valid location from which the zone can
              actually be captured, and the access-path section measures the
              walkable route from the proposed stopping position: both name
              the object and neither obliges its identification, which is why
              it appears in the corpus only inside FR-006's exception clause
              while eight records in this batch pivot on it. The first
              criterion binds the ordering, because identification obliged
              only at completion is satisfied by an implementation that
              evaluates every check against a nearest-way proxy and identifies
              the real position afterwards; FR-070's rationale carries why
              that proxy fails and it is not repeated here. The third
              criterion sends the candidate to uncertain and not to exclusion,
              on the cited definition's own rule that a position which is not
              established is uncertain and that uncertain is not excluded; the
              exclusions bind the position independently of any feature and
              are FR-075's, FR-076's and FR-089's. Its guard is the cited
              section's own, whose uncertain arm resolves the candidate where
              no exclusion applies, and it is kept rather than dropped so that
              an exclusion binding on a ground independent of the position
              still removes the candidate. None of the three reaches the
              absence of a position itself: each presupposes an identified
              position, or a connection identified from one, and FR-076's
              first criterion carries that presupposition rather than leaving
              it to be read in — left to be read in, it is met by any
              candidate for which no position was identified at all, since
              there is certainly no connected route from a position that does
              not exist, and the guard then returns this criterion's own case
              to the exclusion it was written to survive. Not FR-085's: that
              record obliges the identification of a connection between a
              stopping position and the coordinate, this one the
              identification of the position that connection starts from, and
              a system can identify a connection from a position it
              manufactured. What may be concluded about a position's legality
              is bounded by
              `SPECIFICATION.md § Requirements the data cannot verify` and is
              not claimed here — the cited definition states outright that
              establishment does not establish that stopping is lawful at the
              moment of arrival, this record obliges identification and
              refuses manufacture, and whether an identified position is one a
              driver can really use is the judged standard NFR-007 carries.
              That the road the position sits on must itself be admissible is
              FR-089's. That both halves of a stop must be computed against
              the one position is FR-093's, and the two fail independently: a
              system can identify a proper position and still resolve it
              twice.
Resolved-by:  —
```
