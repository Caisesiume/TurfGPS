# Safety exclusions

The enforceable exclusions and the absolute ceiling, as behaviour the system must exhibit. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-044 — Derive the absolute ceiling from the stated additional time

```
Statement:    The system shall derive the absolute additional-time ceiling
              from the additional-time limit the user stated, and never from
              the journey's total duration, per
              `CalculationSpecification.md § The absolute additional-time ceiling`.
Category:     Safety exclusions
Source:       SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — two journeys whose stated limits are equal and whose
              baseline durations differ several-fold yield equal ceilings, and
              one value of additional time passes or fails the ceiling check
              identically on both
Acceptance:   given two journeys whose stated additional-time limits are equal
              and whose baseline durations differ several-fold, when the
              ceiling is determined for each, then the two ceilings are equal
              given one and the same amount of additional time on each of
              those two journeys, when each is tested against its own ceiling,
              then either both pass or both fail
Status:       to-build
Depends-on:   FR-008; FR-038
Risk:         Read against the journey's duration, the ceiling stops being the
              user's and becomes the road's: a long drive silently admits
              several times the detour the driver agreed to, and every ceiling
              check reports compliance while it happens. Nothing downstream
              can detect it, because the arithmetic is right and only the base
              is wrong.
Rationale:    The cited calculation section names this as the misreading the
              constant exists to foreclose, and the section this record is
              sourced from states the same rule in product terms; both are
              cited and neither is quoted, since the multiplier is a proposed
              default whose value has one home. Separate from FR-045, which
              forbids exceeding the ceiling: an implementation can enforce a
              ceiling perfectly and still have computed it over the wrong
              quantity, and that is the failure with no symptom. Separate from
              FR-039 in the same way — that record binds the limit the user
              stated, this one binds the allowance built on top of it.
Resolved-by:  #44
```

## FR-045 — Offer no journey alternative above the absolute ceiling

```
Statement:    The system shall not offer a journey alternative whose
              additional time exceeds the absolute ceiling defined under
              `CalculationSpecification.md § The absolute additional-time ceiling`.
Category:     Safety exclusions
Source:       SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — an alternative whose additional time exceeds the ceiling
              is absent from the offered set, and one exceeding it while
              capturing the zone that carries the user's highest-ranked
              attribute is equally absent
Acceptance:   given a produced journey alternative whose additional time
              exceeds the ceiling, when recommendations are offered, then that
              alternative is not among them
              given a produced journey alternative whose additional time
              exceeds the ceiling and which captures the zone carrying the
              attribute the user ranked highest, when recommendations are
              offered, then that alternative is not among them
              given a journey for which every produced alternative exceeds the
              ceiling, when recommendations are offered, then no alternative
              above the ceiling is offered in place of an empty set
Status:       to-build
Depends-on:   FR-008; FR-044
Risk:         This is the product's one absolute promise about the user's
              time, and the pressure to break it arrives where it is least
              visible — a single highly valued zone just past the line, on a
              journey the optimizer has otherwise solved well.
              `SPECIFICATION.md § The weighting is extreme` makes that
              pressure structural: attribute weights are steep enough that the
              optimizer will spend the whole allowance to reach one zone, so
              this ceiling is the only thing standing between the value model
              and a route the driver never agreed to drive.
Rationale:    The obligation is enforcement, not arithmetic: whatever value
              the configured constant holds, no recommendation exceeds it,
              which is why the statement cites the constant and states no
              figure. The section's "for any reason, however valuable the
              zone" is the second criterion rather than a flourish — it is the
              only branch on which this record ever fires, since an
              alternative nobody wants to offer needs no rule to exclude it.
              The third forecloses the exception that arrives disguised as
              helpfulness: where everything produced is above the ceiling, the
              honest outcome is owned by
              `DESIGN.md § When nothing fits at all` and it is not an
              above-ceiling recommendation. The re-check after every change
              during review is a separate obligation, created by
              `SPECIFICATION.md § Consequences for the optimizer`, and is not
              authored here. Time is never grounds for relaxing a safety rule
              and value is never grounds for relaxing this one, per
              `safety-path-checklist § The absolute ceiling`.
Resolved-by:  #50
```

## FR-046 — Reject a configured ceiling multiplier above the permitted maximum

```
Statement:    The system shall reject a configuration that sets the
              additional-time ceiling multiplier above the maximum stated
              under
              `CalculationSpecification.md § The absolute additional-time ceiling`.
Category:     Safety exclusions
Source:       CalculationSpecification.md § The absolute additional-time ceiling
Priority:     MUST
Verification: test — a configuration setting the multiplier above the stated
              maximum is refused and no journey is planned under it, while one
              setting it below is accepted and the lower ceiling is the one
              enforced
Acceptance:   given a configuration setting the multiplier above the maximum
              stated under
              `CalculationSpecification.md § The absolute additional-time ceiling`,
              when the system starts, then the configuration is refused and no
              journey is planned under it
              given a configuration setting the multiplier below that maximum,
              when the system starts, then the configuration is accepted and
              the ceiling enforced is the one it sets
Status:       to-build
Depends-on:   FR-044
Risk:         A ceiling that can be raised is not a ceiling. Raising it is a
              one-line configuration edit, it leaves every ceiling check
              passing, and nothing in the running system distinguishes a
              deployment that widened the user's allowance from one that never
              touched it.
Rationale:    The cited section states both properties this record rests on:
              that the multiplier's strict direction is downward, and that a
              configuration setting it higher is invalid rather than merely
              unusual. Refusal rather than silent clamping follows from
              invalid — a clamp runs the deployment its operator did not
              configure and hides the misconfiguration behind correct
              behaviour. The second criterion is the other half of the same
              property, and it is what keeps this record from hardening a
              proposed default into a fixed figure: lowering the multiplier
              stays permitted and no value is quoted here. `Source` is the
              calculation section because that is where this obligation is
              created; the model the constant serves is stated under
              `SPECIFICATION.md § User time constraints` and is cited by
              FR-044 and FR-045 for it.
Resolved-by:  #51
```

## FR-053 — Test the limit and the ceiling against the sum across all legs

```
Statement:    The system shall test a journey's additional time against the
              limit the user stated and against the absolute ceiling as the
              sum across all of the journey's legs.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Journeys with several legs
Priority:     MUST
Verification: test — a multi-leg alternative each of whose legs is within its
              own allocated share, and whose legs sum to more than the
              ceiling, is not offered; and one whose legs sum to more than the
              stated limit is not presented as being within that limit
Acceptance:   given a multi-leg journey alternative each of whose legs is
              within the share allocated to it and whose legs' additional
              times sum to more than the absolute ceiling, when
              recommendations are offered, then that alternative is not among
              them
              given a multi-leg journey alternative whose legs' additional
              times sum to more than the limit the user stated, when it is
              presented, then it is not presented as being within that limit
Status:       to-build
Depends-on:   FR-045; FR-047; FR-050
Risk:         A ceiling applied per leg compounds: four legs each inside the
              allowance produce a journey far outside it, and every check the
              system ran passed. The per-leg loop is the natural way to build
              a multi-leg optimizer, so this is the shape the failure arrives
              in by default rather than by mistake.
Rationale:    This restates neither FR-045's prohibition nor FR-047's
              identification; it fixes the quantity each of them is measured
              over, which neither states and which a per-leg optimizer answers
              wrongly while satisfying both inside every leg it looks at. The
              section's "what must hold in every case" is one obligation over
              two limits — the same sum, tested twice — so it stays one
              record.
              `CalculationSpecification.md § The absolute additional-time ceiling`
              states the ceiling half in the same terms, and FR-045 is where
              that constant is cited. That the sum may exceed the stated limit
              at all is the stretch band, which is why the second criterion
              binds how the alternative is presented rather than refusing it.
Resolved-by:  #56
```

## FR-061 — Test the stated limit and the ceiling against the whole additional time

```
Statement:    The system shall test a journey alternative against the
              additional-time limit the user stated and against the absolute
              ceiling derived from it using the alternative's total additional
              time relative to the fastest conventional route, comprising both
              its general-route deviation and the time its Turf stops add,
              composed per
              `CalculationSpecification.md § Additional journey time`.
Category:     Safety exclusions
Source:       SPECIFICATION.md § General route alternatives;
              SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — an alternative whose Turf stops alone fall within the
              stated limit, but which is built on a general road route slower
              than the fastest conventional one by enough to carry the pair
              past it, is not treated as within that limit, and the same pair
              carried past the ceiling is not treated as within the ceiling
Acceptance:   given a journey alternative whose Turf stop time alone is within
              the limit the user stated and whose general-route deviation and
              Turf stop time together exceed it, when it is tested against
              that limit, then it is not treated as being within it
              given a journey alternative whose Turf stop time alone is within
              the absolute ceiling and whose general-route deviation and Turf
              stop time together exceed it, when it is tested against that
              ceiling, then it is not treated as being within it
              given a journey alternative built on the fastest conventional
              route, when it is tested against either of them, then the
              quantity tested is its Turf stop time alone
              given two journey alternatives carrying equal Turf stop time,
              one built on the fastest conventional route and one on a slower
              general road route, when each is tested against that limit, then
              the one on the slower route is measured as the more costly
Status:       to-build
Depends-on:   FR-008; FR-038; FR-044
Volatility:   settled — the quantity and the baseline it is measured against
              were ruled by the Owner on 7 August 2026; the ceiling it also
              tests is derived from a proposed multiplier under
              `CalculationSpecification.md § The absolute additional-time ceiling`,
              and this record binds the quantity rather than that figure
Risk:         The general-route deviation is the larger of the two terms on
              any journey where the roads genuinely differ, so a limit tested
              against the Turf half alone understates the promise by the
              biggest number in it: a route half an hour longer by road is
              offered as costing the minutes its stops took. The ceiling is
              derived from the same stated limit under FR-044 and is enforced
              by FR-045 against whatever quantity its implementer reads, so
              the understatement propagates into the one check that may never
              be wrong, and every check still reports compliance.
Rationale:    Three readings of the quantity are available and two of them
              survive a single criterion: the limit measured against the Turf
              half alone, against the general-route deviation alone, and
              against the sum. The first, third and fourth criteria close
              them. The second criterion carries the ceiling, and it is here
              because FR-045 forbids offering an alternative whose additional
              time exceeds the ceiling while nothing in the corpus composes
              that quantity — leaving its implementer two figures to choose
              from, of which the Turf half is the one FR-062 obliges the
              system to retain separately. This stays one record rather than
              two on FR-053's precedent: one composed sum tested against two
              limits is one obligation, not a second. FR-045 is signed and is
              not amended here; this record fixes the quantity it tests,
              exactly as FR-044 fixes the base its ceiling is derived from.
              This record does not touch what FR-008 measures — the Turf
              figure stays per alternative against the general road route it
              is built on, which is the baseline that keeps two Turf
              strategies on one road comparable — and the amendment recording
              that on FR-008 is `@requirements-engineer`'s. FR-053 fixes the
              same quantity across a journey's legs and this record fixes what
              enters it on each leg; neither states the other. The arithmetic
              has one home and is cited, per `docs/README.md § Conventions`.
Resolved-by:  —
```

## FR-069 — Reject a direct-access tolerance above the permitted value

```
Statement:    The system shall reject a configuration that sets the
              direct-access tolerance above the value stated under
              `CalculationSpecification.md § Direct-access tolerance`.
Category:     Safety exclusions
Source:       CalculationSpecification.md § Direct-access tolerance
Priority:     MUST
Verification: test — a configuration setting the tolerance above the stated
              value is refused and no journey is planned under it, while one
              setting it below is accepted and the tighter tolerance is the
              one enforced
Acceptance:   given a configuration setting the direct-access tolerance above
              the value stated under
              `CalculationSpecification.md § Direct-access tolerance`, when
              the system starts, then the configuration is refused and no
              journey is planned under it
              given a configuration setting that tolerance below the stated
              value, when the system starts, then the configuration is
              accepted and the tolerance enforced is the one it sets
Status:       to-build
Depends-on:   FR-068
Volatility:   proposed-constant — the stated value is a proposed default the
              cited section marks as worth checking early against real
              captures, so the point this record enforces at is expected to
              move; the direction it may be configured in is not
Risk:         Raising this constant mis-prices nothing, which is why it looks
              harmless in a configuration diff: it enlarges the set of zones
              handed to a driver as capturable from the seat with no
              walk-safety gate ever entered. Every access check still passes,
              because the checks that would have failed are precisely the ones
              a raised tolerance skips.
Rationale:    The cited section names this an enforcement constant whose
              strict direction is downward, and
              `CalculationSpecification.md § Conventions` states what that
              means: a deployment may tighten it and may never loosen it, and
              the documented value is the limit of what is permitted rather
              than a midpoint to tune around. `Source` is the calculation
              section because that is where this obligation is created,
              exactly as FR-046 records for the ceiling multiplier; the model
              the constant serves is
              `SPECIFICATION.md § Directly road-accessible zones`'s and is
              cited by FR-068 for it. Refusal rather than silent clamping
              follows the reasoning FR-046 records: a clamp runs the
              deployment its operator did not configure and hides the
              misconfiguration behind correct behaviour. The second criterion
              is what keeps this record from hardening a proposed default into
              a fixed figure.
Resolved-by:  —
```

## FR-070 — Require compatible levels and no barrier for direct road access

```
Statement:    The system shall classify a zone as directly road-accessible
              only where the stopping position and the zone are established to
              be at compatible levels with no intervening barrier between
              them.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Direct road-access validation
Priority:     MUST
Verification: test — a zone above a road that passes beneath it, a zone
              beneath a bridge carrying the road, and a zone separated from
              the road by a retaining wall are each not classified directly
              road-accessible, while a zone established to be at a compatible
              level with nothing between is not refused by this check
Acceptance:   given a candidate whose bridge, tunnel or layer data indicates
              that the stopping position and the zone do not meet, when its
              access is classified, then it is not classified directly
              road-accessible
              given a candidate for which the data records a barrier between
              the stopping position and the zone, when its access is
              classified, then it is not classified directly road-accessible
              given a candidate established to be at a level compatible with
              the stopping position with no barrier recorded between them,
              when its access is classified, then this check does not refuse
              it the direct class
Status:       to-build
Depends-on:   FR-068; FR-092
Volatility:   open-question — the bridge, tunnel, layer and barrier attributes
              this record tests are the OSM-derived feature data
              `Architecture.md § D4` decides the store holds, and
              `Architecture.md § Still owed by this document` records those
              tables as owed, undesigned and unreviewed rather than built
Risk:         This class tells a driver they can take the zone without leaving
              the car. Wrong by a level, it directs them to stop on a road
              that passes under or over the zone entirely, at a place where
              the capture cannot happen and where stopping was justified only
              by the capture. The two-dimensional test that produces this
              failure is the obvious implementation and looks correct on a
              map.
Rationale:    This is one of the exclusions under
              `SPECIFICATION.md § Enforceable exclusions`, which states the
              prohibition and points here for its content, so the record is
              sourced where that content lives. The position tested is the
              candidate's stopping position — the one FR-092 obliges be
              identified — and not whatever way lies nearest its coordinate,
              which is why `Depends-on` names that record. A bridge and its
              approach are two ways carrying different `layer` values, and
              both can lie within the direct-access tolerance of one
              coordinate, so a check evaluated against a position other than
              the one the car will stop at can pass on the approach while the
              car stops on the bridge. It is not FR-071's: that record governs
              where a candidate goes when this check fails, and a system that
              never runs the check fails this record while never firing that
              one. The third criterion states this record's own bound and not
              the class's: what it declines to refuse on is the level and
              barrier evidence this record tests, and every other condition on
              the direct class stands where that condition is written —
              among them the road-class and speed-limit exclusions under
              `SPECIFICATION.md § Enforceable exclusions`, which FR-089
              carries for both branches and this record does not — so it
              cannot be read as the whole test for the class.
Resolved-by:  —
```

## FR-076 — Exclude an absent, disconnected, or implausibly steep access path

```
Statement:    The system shall exclude a candidate zone whose access path is
              absent, disconnected, or steeper than the configured
              implausible-gradient threshold recorded as owed under
              `CalculationSpecification.md § Enforcement constants that do not yet exist`,
              which is a distinct value from the maximum acceptable path
              gradient the user configures under
              `SPECIFICATION.md § User-configurable terrain tolerance` and is
              never relaxed by it.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Elevation and feasibility rules
Priority:     MUST
Verification: test — a candidate with an identified stopping position from
              which no connected walking route to the zone's coordinate can be
              identified is excluded rather than priced by any estimate, one
              whose path carries a section steeper than the configured
              threshold is excluded whatever the zone carries, and one so
              excluded stays excluded when the user's maximum acceptable path
              gradient is set less strict than that threshold
Acceptance:   given a candidate for which a stopping position has been
              identified and for which no connected walking route between that
              position and the zone's coordinate can be identified, when its
              access is classified, then it is excluded and is not priced by a
              straight-line estimate
              given a candidate whose identified access path carries a section
              steeper than the configured implausible-gradient threshold, when
              its access is classified, then it is excluded
              given the same candidate and a maximum acceptable path gradient
              the user configured less strict than that threshold, when its
              access is classified, then it is still excluded and the user's
              setting does not admit it
              given a candidate excluded under this rule which carries the
              attribute the user ranked highest, when journey alternatives are
              produced, then no alternative includes it
Status:       to-build
Depends-on:   FR-067; FR-074; FR-092
Volatility:   open-question — the implausible-gradient threshold has no value
              in any document and is recorded as an owed enforcement constant
              under
              `CalculationSpecification.md § Enforcement constants that do not yet exist`,
              which also records that it may need to be per elevation
              provider, so the shape of the constant and not only its value is
              unsettled
Risk:         This is the exclusion standing between a recommendation and a
              driver climbing an embankment or a quarry edge at the roadside,
              and it is the one the data supports least:
              `Architecture.md § D6` records that a global thirty-metre model
              cannot resolve a retaining wall, which is narrower than one
              cell. Enforced against an optimistic threshold the system
              recommends the stop with full confidence and the failure is not
              a bad estimate but a fall. Enforced against none while the value
              is owed, the steep limb of this record is dead for every
              candidate while every test written for it still passes on the
              other two — the state FR-091 refuses.
Rationale:    The cited section lists the conditions on which a candidate is
              excluded or strongly penalized, and
              `SPECIFICATION.md § Enforceable exclusions` makes three of them
              hard rules — absent, disconnected, implausibly steep — which is
              why this record obliges exclusion and authors no penalty: no
              penalty scale exists, and a hard rule stated as a penalty is one
              that can be outbid. The threshold is cited to the section that
              owns it rather than to
              `SPECIFICATION.md § Enforceable exclusions`, which names the
              exclusion qualitatively and states no threshold at all; an
              implementer following a citation to a phrase writes the
              unexplained literal
              `CalculationSpecification.md § Conventions` forbids. The
              statement separates that threshold from the user's maximum
              acceptable path gradient under
              `SPECIFICATION.md § User-configurable terrain tolerance` because
              nothing else in the corpus does, and wiring the two together is
              the obvious implementation: it reads as respecting the user
              while relaxing the exclusion exactly where the user asked for
              more tolerance. That the user's setting may tighten the
              effective limit is permitted by that tolerance section and is
              obliged nowhere, on the ground FR-080 records for the
              walking-distance bound. The statement binds the system to read
              the configured threshold and enforce it rather than to any
              figure, which is the only correct shape while the value is owed,
              per `safety-path-checklist § The proposal boundary`; that it may
              not be enforced against no value at all is FR-091's. The
              remaining conditions in that list are covered rather than
              dropped: an approach across unmapped terrain reaches the
              fallback and FR-083, an insufficiently confident estimate is
              FR-082's, and a cliff or wall the data shows is what makes the
              path disconnected. The first criterion presupposes an identified
              stopping position and cannot fire without one. Read without that
              presupposition it is satisfied vacuously wherever no position
              was identified at all — there is certainly no connected route
              from a position that does not exist — so this record would
              exclude the case
              `SPECIFICATION.md § What establishes a stopping position` sends
              to uncertain, and would do it through the very guard FR-092's
              third criterion rests on. Narrowing the antecedent gives up no
              exclusion and costs no safety: absent, disconnected and
              implausibly steep each bind exactly as before once a position
              exists, and the no-position case lands in uncertain, which
              FR-084 keeps out of the cost model and the ranking so that no
              walk is priced for it. The fourth criterion is the branch this
              record ever fires on, on FR-045's precedent — an excluded
              candidate nobody wanted needs no rule.
Resolved-by:  —
```

## FR-077 — Let no zone's value admit an unsafe or infeasible access route

```
Statement:    The system shall not admit a candidate whose access route is
              classified as unsafe, illegal, or physically infeasible on the
              grounds of the Turf value the zone carries or of the terrain
              tolerance the user configured.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Elevation and feasibility rules;
              SPECIFICATION.md § User-configurable terrain tolerance
Priority:     MUST
Verification: test — a candidate excluded on access grounds stays excluded
              when it carries the user's highest-ranked attribute, stays
              excluded when the user's terrain tolerance is configured
              generously enough to admit its walk, and stays excluded when the
              user's maximum acceptable path gradient is configured generously
              enough to admit its climb
Acceptance:   given a candidate whose access route is classified unsafe,
              illegal, or physically infeasible and which carries the
              attribute the user ranked highest, when journey alternatives are
              produced, then no alternative includes it
              given the same candidate and a terrain tolerance the user
              configured generously enough to admit its walking distance and
              its elevation gain, when journey alternatives are produced, then
              no alternative includes it
              given a candidate whose access path carries a section steeper
              than the configured implausible-gradient threshold and a maximum
              acceptable path gradient the user configured generously enough
              to admit that path, when journey alternatives are produced, then
              no alternative includes it
Status:       to-build
Depends-on:   FR-076
Volatility:   settled — the record rests on the precedence of an access
              classification over value and over a user setting, which both
              cited sections state and neither leaves open; the threshold its
              third criterion fires against is owed, per FR-076, which moves
              how often that criterion fires rather than whether the record
              holds
Risk:         The pressure is structural rather than occasional:
              `SPECIFICATION.md § The weighting is extreme` makes attribute
              weights steep enough that the optimizer will spend everything it
              has to reach one zone, so the case where value argues against an
              access classification is the ordinary case at the top of the
              ranking rather than an edge. A tolerance setting is the second
              route to the same place, and it arrives looking like the user's
              own choice.
Rationale:    Both cited sections state the same precedence from opposite
              sides — a highly valued attribute may justify additional walking
              time or elevation gain but must not override an access route
              classified unsafe, illegal or physically infeasible, and the
              terrain settings influence which feasible zones are recommended
              without permitting unsafe or inaccessible ones — so this is one
              obligation with two grounds rather than two records. It is
              FR-045's shape one level down: that record refuses value as a
              reason to exceed the ceiling, this one refuses value as a reason
              to cross an access classification, and
              `SPECIFICATION.md § Requirements the data cannot verify` states
              the general form, that safety is a constraint on the search
              space and never a term in the objective. The third criterion is
              in the gradient dimension because the second is not, and the gap
              matters: nothing outside FR-076's statement says the enforcement
              threshold and the user's maximum acceptable path gradient are
              different objects, so an implementation computing physically
              infeasible from the user's setting would leave this record an
              antecedent that never fires — a criterion that can never fail,
              which is the one kind the corpus rejects. Its antecedent is
              FR-076's threshold precisely because that quantity does not move
              with the user's setting.
Resolved-by:  —
```

## FR-089 — Propose no stop on a road the exclusions refuse

```
Statement:    The system shall propose no stopping position on a road the
              road-class and speed-limit exclusions under
              `SPECIFICATION.md § Enforceable exclusions` do not admit as a
              stopping place.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions;
              SPECIFICATION.md § Directly road-accessible zones
Priority:     MUST
Verification: test — a candidate whose stopping position is on a motorway or a
              motorway link reaches neither confident class, a zone off a
              rural road posted above the maximum and reachable by a mapped
              track is not offered as a park-and-walk stop from a position on
              that carriageway, a candidate within the direct-access tolerance
              refused on this ground is evaluated on the park-and-walk branch
              rather than excluded, and one whose stopping position is on an
              admitted road is refused neither class by this check
Acceptance:   given a park-and-walk candidate whose identified stopping
              position is on a road the road-class and speed-limit exclusions
              under `SPECIFICATION.md § Enforceable exclusions` do not admit
              as a stopping place, when its access is classified, then that
              position is not proposed for it and it is not classified
              park-and-walk on that position
              given a candidate within the direct-access tolerance whose
              identified stopping position is on such a road, when its access
              is classified, then it is not classified directly
              road-accessible and is evaluated on the park-and-walk branch
              rather than excluded on that finding
              given a candidate whose identified stopping position is on a
              road those exclusions admit as a stopping place, when its access
              is classified, then this check refuses it neither confident
              class
Status:       to-build
Depends-on:   FR-068; FR-092
Volatility:   proposed-constant — the maximum speed limit for a stopping road
              under
              `CalculationSpecification.md § The maximum speed limit for a stopping road`
              is a proposed default with no origin, so which roads this record
              refuses moves with it even though the refusal itself does not
Risk:         A confident class tells a driver to stop, and this record is the
              only thing on either access path that asks what road they are
              being told to stop on. On the direct path a zone six metres from
              a drivable edge, at a compatible level with nothing between,
              reaches the class on a derestricted autobahn — the outcome
              `SPECIFICATION.md § An unknown speed limit fails the check`
              names as the single worst one the exclusion exists to prevent.
              On the walking path the same hole is quieter and no less
              reachable: a rural trunk road posted above the maximum, a zone
              forty metres off it, a mapped track between them, and every
              remaining gate passes — the output is a confident park-and-walk
              stop requiring the driver to stop on that carriageway and stand
              on its verge. In both cases every check the classifier ran
              reports a pass, because the check that would have failed was
              never written.
Rationale:    The obligation is on the stop and not on a class, because that
              is where the cited section puts it: no stop may be proposed on
              such a road, on either branch. `Architecture.md § D3` reads the
              same way from the other side, naming the edge attributes the
              routing engine returns as what that section must test against
              when validating a stopping position. An earlier form of this
              record gated the direct class alone, and what that left is
              FR-068's other branch — a candidate beyond the direct-access
              tolerance goes to the walking evaluation with nothing there
              asking what road its stopping position sits on. Binding the
              position covers both branches in one record, because FR-092
              obliges an identified stopping position for both confident
              classes and this record refuses a class of road for whichever
              branch produced it. Where a refusal sends the candidate is
              stated rather than left for an implementer to pick, and the
              cited section supplies the answer: a nearby rest area, service
              road, parking area or exit may still make the zone accessible
              while the high-speed carriageway itself never is, so what is
              refused is the position and not the zone. On the direct branch
              that is a downgrade to the park-and-walk branch, FR-071's shape
              on an antecedent that record does not reach, its own being a
              level or barrier finding. The composition is not a way round
              this record: a candidate downgraded here meets it again on the
              walking branch, where a stopping position on the same
              carriageway is refused by the first criterion, and where no
              admissible position is identified at all FR-092's third
              criterion carries the destination. This record exists because
              FR-067 obliges every classified candidate into exactly one of
              four classes while FR-070's third criterion affirmatively
              declines to refuse the direct class on any other ground, which
              completed a path to a confident class around the gap it fills.
              What establishes a speed limit, what a road carrying more than
              one does, and that no lookup table may resolve a jurisdictional
              value are stated under
              `SPECIFICATION.md § Enforceable exclusions` and are owed to the
              batch scoped there, so none of it is restated in a criterion
              here and the criteria refer the decision to that section rather
              than making it. The constant those attributes are compared
              against has one home and is never quoted here, and that it may
              not be configured upward is FR-090's.
Resolved-by:  —
```

## FR-090 — Reject a maximum stopping speed above the permitted value

```
Statement:    The system shall reject a configuration that sets the maximum
              speed limit for a stopping road above the value stated under
              `CalculationSpecification.md § The maximum speed limit for a stopping road`.
Category:     Safety exclusions
Source:       CalculationSpecification.md § The maximum speed limit for a stopping road
Priority:     MUST
Verification: test — a configuration setting the maximum above the stated
              value is refused and no journey is planned under it, while one
              setting it below is accepted and the lower maximum is the one
              enforced
Acceptance:   given a configuration setting the maximum speed limit for a
              stopping road above the value stated under
              `CalculationSpecification.md § The maximum speed limit for a stopping road`,
              when the system starts, then the configuration is refused and no
              journey is planned under it
              given a configuration setting that maximum below the stated
              value, when the system starts, then the configuration is
              accepted and the maximum enforced is the one it sets
Status:       to-build
Depends-on:   FR-089
Volatility:   proposed-constant — the stated value is a proposed default by
              default, carrying no origin, no measurement and no cited source,
              so the point this record enforces at is expected to move; the
              direction it may be configured in is not
Risk:         Raising this constant admits roads the cost model prices nothing
              for. The roadside rows under
              `CalculationSpecification.md § Proposed placeholder timings` run
              only as far as the exclusion does, so a raised maximum puts
              stopping places into the model with nothing calibrated or even
              guessed for stopping on them, and the driver is sent to the
              verge of a carriageway faster than any road the timings were
              written for. The pressure to raise it is structural rather than
              occasional — the market holding the second-largest national
              share of the zone corpus is one whose rural network the
              exclusion removes — so it arrives as a coverage argument with a
              one-line configuration edit behind it, and every exclusion check
              still passes afterwards.
Rationale:    `CalculationSpecification.md § Conventions` names the maximum
              speed limit for a stopping road an enforcement constant and
              states what that means: a deployment may tighten it and may
              never loosen it, and the documented value is the limit of what
              is permitted rather than a midpoint to tune around. The section
              owning the constant states that its strict direction is downward
              and gives the reason. `Source` is that calculation section
              because that is where this obligation is created, on FR-046's
              and FR-069's precedent; the exclusion the constant serves is
              `SPECIFICATION.md § Enforceable exclusions`'s and is gated by
              FR-089 for it. Refusal rather than silent clamping follows the
              reasoning FR-046 records: a clamp runs the deployment its
              operator did not configure and hides the misconfiguration behind
              correct behaviour. The second criterion is what keeps this
              record from hardening a proposed default into a fixed figure. No
              figure is quoted anywhere here, and
              `SPECIFICATION.md § The United Kingdom is a park-and-walk market`
              is deliberately absent from `Source`: it measures the
              consequence of the constant against the zone corpus and states
              outright that it is not a source for the constant itself.
Resolved-by:  —
```

## FR-091 — Plan no journey while an owed enforcement constant is unconfigured

```
Statement:    The system shall refuse to plan a journey while any enforcement
              constant recorded as owed under
              `CalculationSpecification.md § Enforcement constants that do not yet exist`
              has no configured value.
Category:     Safety exclusions
Source:       CalculationSpecification.md § Enforcement constants that do not yet exist
Priority:     MUST
Verification: test — a configuration carrying no value for a constant that
              section records as owed is refused and no journey is planned
              under it, one carrying a value for some of them and none for the
              rest is refused likewise, and one carrying a value for every
              constant it records as owed is accepted and no journey is
              refused on this ground
Acceptance:   given a configuration carrying no value for a constant recorded
              as owed under
              `CalculationSpecification.md § Enforcement constants that do not yet exist`,
              when the system starts, then the configuration is refused and no
              journey is planned under it
              given a configuration carrying a value for some of the constants
              that section records as owed and no value for the rest, when the
              system starts, then the configuration is refused and no journey
              is planned under it
              given a configuration carrying a value for every constant that
              section records as owed, when the system starts, then the
              configuration is accepted and no journey is refused on this
              ground
Status:       to-build
Depends-on:   FR-076
Volatility:   open-question — neither constant the cited section records as
              owed has a value in any document, and that section records that
              the implausible-gradient threshold may need to be per elevation
              provider, so for that one the shape of the constant and not only
              its value is unsettled
Risk:         Unset, a constant of this class is not a check that fails but a
              check that silently never runs. The implausible-gradient
              threshold is the sharper case: the comparison it feeds is false
              for every candidate, so the steep limb of FR-076 never fires and
              a quarry-rim candidate carrying a fragment of mapped path
              reaches park-and-walk while every test written for FR-076 still
              passes on its absent and disconnected limbs — on the one
              exclusion whose failure is a fall rather than a bad estimate.
              The conservative upper bound for an uncertain stop fails the
              other way and just as quietly: the cited section makes it the
              only sensor the absolute ceiling has for an uncertain stop, so
              computed optimistically the ceiling check passes, FR-045's
              refusal never fires, and a user accepts a zone that commits them
              past the limit they stated. Every other enforcement constant in
              the set has a record refusing a loosened value; these are the
              ones with no value at all, and
              `CalculationSpecification.md § Conventions` forbids the
              alternative an implementer reaches for, a literal authored on
              the spot with no origin.
Rationale:    The cited section records these constants as required for the
              system to behave as specified, states that they have no value
              anywhere, and gives the reason for recording them: the first
              implementer to need one cannot author it silently, as the
              unexplained literal `CalculationSpecification.md § Conventions`
              forbids. This record is that prohibition made enforceable. The
              antecedent is the class that section records rather than one
              member of it — an earlier form of this record covered the
              gradient threshold alone, which left the second constant free
              and would have gone stale the day a third was added. It is the
              only shape available for any member: a value cannot be defaulted
              here, because a default is the literal that section forbids and
              this lane may not pick one, and it cannot be left out, because
              omission is the dead limb FR-076's own `Risk` names. Refusal at
              start rather than a quiet downgrade follows the reasoning FR-046
              records for a clamp — the quiet path runs a deployment its
              operator did not configure and hides the misconfiguration behind
              plausible behaviour, which for the gradient threshold means a
              coverage loss indistinguishable from thin map data. `Depends-on`
              names FR-076 alone because it is the only consuming record that
              exists: what consumes the conservative upper bound is owed to
              the batch scoped to
              `SPECIFICATION.md § Reconciling this with the absolute ceiling`,
              and binding the class rather than the instance is exactly what
              lets this refusal stand before that record is written. That the
              system is blocked until the constants are authored is the intent
              and not a side effect; authoring either value is owed to the
              batch that owns it, and this record states what may not happen
              meanwhile. Neither value nor shape is quoted, since neither
              exists.
Resolved-by:  —
```

## FR-125 — Route no driving leg through a private or access-restricted way or barrier

```
Statement:    The system shall produce no driving leg that runs along a way,
              or passes a barrier, that the map data marks as private or
              access-restricted, other than in the run FR-126 admits at a
              location the user entered or as a warned leg FR-129 admits.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract holding
              such ways and barriers, a leg between ways not so marked whose
              quickest drivable connection runs along a marked way or through
              a marked barrier is produced on a connection avoiding it, a leg
              to a stopping position the system chose beside a marked way runs
              along none other than as a warned leg FR-129 admits, where
              every drivable connection between two
              points on ways not so marked is marked no leg is produced along
              one other than a warned leg FR-129 admits, and a leg with one
              end at a location the user entered on a marked way runs beyond
              FR-126's run on a connection avoiding every marking, or is not
              produced where none does other than as a warned leg FR-129
              admits; and a leg
              whose quickest connection runs along a way or past a barrier
              that `DECISIONS.md § RD-063` leaves unmarked, by its precedence
              or by a value it adds to the permissive set, runs along it,
              while one carrying a value outside that set on the least
              specific key alone is avoided, and one carrying a value outside
              that set on the most specific of the four keys it carries is
              avoided though it also carries a value within that set on a key
              that is none of the four; and one so marked, as
              `DECISIONS.md § RD-064`, `DECISIONS.md § RD-065` and
              `DECISIONS.md § RD-066` define, by a value outside that set only
              on a `:forward` or `:backward` form is avoided in the direction
              the form does not name, as is one so marked by such a value on a
              directional form while the most specific plain key it carries
              permits, and a node carrying such a value with no `barrier` key,
              while one carrying such a value only on a `:conditional` form is
              driven; and one carrying a value outside
              that set on the most specific of the four keys it carries is
              avoided though it carries a value within that set on both the
              `:forward` and the `:backward` form of one of them, or on a
              `:conditional` form of one
Acceptance:   given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way, or passes a
              barrier, that the map data marks as private or
              access-restricted, and between whose ends a drivable connection
              avoiding every such way and barrier exists, when the leg is
              produced, then it runs along no such way and passes no such
              barrier
              given a driving leg one end of which is a stopping position the
              system chose, lying nearer a way so marked than any way not so
              marked, and the other end of which lies on a way not so marked,
              when the leg is produced, then it runs along no such way and
              passes no such barrier, other than as a warned leg FR-129
              admits
              given two points on ways not so marked, every drivable
              connection between which runs along such a way or passes such a
              barrier, when the journey is planned, then no driving leg is
              produced that runs along such a way or passes such a barrier,
              other than a warned leg FR-129 admits
              given a driving leg one end of which is a location the user
              entered lying on a way so marked, whose quickest drivable
              connection beyond the run FR-126 admits at that location runs
              along such a way or passes such a barrier, when the journey is
              planned, then beyond that run the leg runs along no such way and
              passes no such barrier, being produced on a drivable connection
              from that run's end that avoids every such way and barrier
              where one exists, and not produced where none does other than
              as a warned leg FR-129 admits
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along no way and passes no
              barrier so marked, one way or barrier on it carrying a value
              outside the permissive set on a key less specific than another
              it carries with a value within that set, when the leg is
              produced, then it runs along that way or passes that barrier
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along no way and passes no
              barrier so marked, one way or barrier on it carrying on its most
              specific key a value that `DECISIONS.md § RD-063` adds to the
              permissive set, when the leg is produced, then it runs along
              that way or passes that barrier
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way carrying a value
              outside the permissive set on the least specific of the four
              keys and none of the other three, and between whose ends a
              drivable connection avoiding every such way and barrier exists,
              when the leg is produced, then it runs along no such way
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way, or passes a
              barrier, carrying a value outside the permissive set on the most
              specific of the four keys it carries and a value within that set
              on a key that is none of the four (`bicycle`, `foot`, `psv` or
              `motorcycle`), and between whose ends a drivable connection
              avoiding every such way and barrier exists, when the leg is
              produced, then it runs along no such way and passes no such
              barrier
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way, or passes a
              barrier, carrying none of the four keys and a value outside the
              permissive set on a `:forward` or `:backward` form of one of
              them, and so marked as `DECISIONS.md § RD-064`,
              `DECISIONS.md § RD-065` and `DECISIONS.md § RD-066` define, that
              connection running along the way or passing the barrier in the
              direction the form does not name, and between whose ends a
              drivable connection avoiding every such way and barrier exists,
              when the leg is produced, then it runs along no such way and
              passes no such barrier
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection passes a node carrying no `barrier`
              key and a value outside the permissive set on the most specific
              of the four keys it carries, and between whose ends a drivable
              connection avoiding every such node and every way and barrier so
              marked exists, when the leg is produced, then it passes no such
              node
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way, or passes a
              barrier, carrying a value within the permissive set on the most
              specific of the four keys it carries and a value outside that
              set on a `:forward` or `:backward` form of a less specific one
              (`motorcar=yes` with `motor_vehicle:backward=no`), and so marked
              as `DECISIONS.md § RD-064`, `DECISIONS.md § RD-065` and
              `DECISIONS.md § RD-066` define, and between whose ends a
              drivable connection avoiding every such way and barrier exists,
              when the leg is produced, then it runs along no such way and
              passes no such barrier
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along no way and passes no
              barrier so marked, one way on it carrying none of the four keys
              or their `:forward` and `:backward` forms and a value outside
              the permissive set on a `:conditional` form of one of them, when
              the leg is produced, then it runs along that way
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way, or passes a
              barrier, carrying a value outside the permissive set on the most
              specific of the four keys it carries and a value within that set
              on both the `:forward` and the `:backward` form of one of them
              (`motor_vehicle=no` with `motor_vehicle:forward=yes` and
              `motor_vehicle:backward=yes`), and between whose ends a drivable
              connection avoiding every such way and barrier exists, when the
              leg is produced, then it runs along no such way and passes no
              such barrier
              given a driving leg whose ends lie on ways not so marked, whose
              quickest drivable connection runs along a way, or passes a
              barrier, carrying a value outside the permissive set on the most
              specific of the four keys it carries and a value within that set
              on a `:conditional` form of one of them (`access=private` with
              `motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00)`), and
              between whose ends a drivable connection avoiding every such way
              and barrier exists, when the leg is produced, then it runs along
              no such way and passes no such barrier
Status:       to-build
Depends-on:   FR-006; FR-126; FR-129
Volatility:   open-question — how the tags are read is designed in
              `Architecture.md § The access-marking tables`, which reads
              neither a leg's direction nor a conditional key yet; a way or
              barrier is so marked where the most specific of
              `DECISIONS.md § RD-059`'s keys it carries has a value outside
              the permissive set, the precedence and the set being
              `DECISIONS.md § RD-063`'s, which supersedes
              `DECISIONS.md § RD-062`'s second and third edges; the keys'
              directional and conditional forms, and a node carrying no
              `barrier` key, are read as `DECISIONS.md § RD-064`,
              `DECISIONS.md § RD-065` and `DECISIONS.md § RD-066` decide; and
              a spelling found wrong is corrected without narrowing it; and
              where nothing avoids a marking between two locations the user
              entered, FR-129's warned leg replaces the refusal on the
              Owner's ruling in `DECISIONS.md § RD-067`, whose open edges
              await the Owner
Risk:         A routing engine that prices a private road as a cost rather
              than a prohibition takes the shortcut through a gated estate or
              along a residents' road exactly where it saves the most time,
              and the leg comes back as one a car can drive, so FR-006's check
              passes. The driver is sent through a gate they have no right to
              open, and nothing in the result says so.
Rationale:    The cited section makes this a hard rule: no route may be
              constructed through an area the map data marks private or
              access-restricted. FR-075 keeps the walking access path off a
              way so marked and reads the member as binding the area rather
              than the way; this record keeps that reading and takes the
              area's marking from what a driving leg is made of, the ways it
              runs along and the barriers across them, and FR-128 adds the
              ways an area so marked encloses. Way and barrier stay one record
              because a gate and the road behind it mark one area, and a
              record for either alone leaves the other as the way round it.
              The member's antecedent is a route through an area and never a
              stop at a position inside one
              (`SPECIFICATION.md § What establishes a stopping position`), so
              this record excludes no stop and no location a journey names; it
              binds what a leg travels along. Its two exceptions are the
              Owner's. The first is a run at a location the user entered, and
              FR-126 states it with the bound the Owner set
              (`DECISIONS.md § RD-056`, `DECISIONS.md § RD-057`). The second
              is a warned leg where nothing avoids a marking between two
              locations the user entered, and FR-129 states it
              (`DECISIONS.md § RD-067`). A stopping position the user chose is
              such a location (`DECISIONS.md § RD-061`); one the system chose
              is outside the exception (`DECISIONS.md § RD-052`), which the
              second criterion tests, unless a warned leg FR-129 admits
              reaches it, as the Owner allows along the warned leg a location
              the user entered requires (`DECISIONS.md § RD-079`), and FR-127
              decides when a leg may end
              short of a stopping position. The fourth criterion holds the
              exclusion on the rest of a leg that has one end at such a
              location, so the exception admits the run and no more of the
              leg: a leg leaving a home on a private road takes the least run
              out and is then held to the rule in transit. The third
              criterion keeps the
              exclusion hard: where nothing avoids the area the answer is no
              leg through it, never a least-bad crossing presented as a road.
              The warned leg both criteria now except is not that crossing:
              FR-129 admits it only between two locations the user entered,
              and to a stop the system chose only within the bound it sets
              on the legs joining them, FR-130 keeps every avoidable marking
              off it, FR-131 to FR-133
              identify, show and warn of every marked stretch on it, and
              FR-134 lets no journey holding it be confirmed until the user
              confirms the warning, so the crossing is never silent and never
              chosen for the user. The cited section does not yet carry that
              exception; `DECISIONS.md § RD-067` records the ruling and
              raises the section. The fifth and
              sixth criteria hold the Owner's loosening of the floor
              (`DECISIONS.md § RD-063`): a way a more specific key permits,
              or whose value the Owner added to the permissive set, is driven,
              so a build that still reads either as marked fails; the seventh
              keeps a value outside the set on the least specific key alone
              marked, so the loosening lifts nothing else. The eighth keeps a
              permission on a key that is none of the four from lifting a
              marking, so a build that takes the most specific key of any
              mode, not of a car, fails. The ninth to the twelfth test the
              Owner's tag decisions (`DECISIONS.md § RD-064`,
              `DECISIONS.md § RD-065`), the ninth's and the eleventh's
              premises reading "so marked" as those entries and
              `DECISIONS.md § RD-066` define: a way marked in one direction
              alone is avoided in the other, so a build that ignores the
              directional forms, or reads a leg's direction before the check
              can, fails; a node so marked with no `barrier` key is a barrier;
              a directional form outside the set marks though a plain, more
              specific key permits, but not where a more specific directional
              form of the same direction permits, which
              `DECISIONS.md § RD-065`'s second edge reads as unmarked
              (`motorcar:backward=yes` with `motor_vehicle:backward=no`, with
              or without `motorcar=yes`);
              and a value outside the set on a `:conditional` form
              alone is driven, so a build that reads a condition it cannot
              evaluate as a marking fails. The thirteenth holds the Owner's
              ruling that a directional form only adds markings
              (`DECISIONS.md § RD-066`): a plain marking stands though both
              directional forms of a key permit, so a build that lets a
              directional permission take its key's place fails. The
              fourteenth holds the other half of the conditional decision: a
              `:conditional` permission lifts no marking, so a build that
              reads a condition it cannot evaluate as a permission fails, and
              the half of that decision that errs strict is held as firmly as
              the twelfth holds the half that does not. No mechanism is named,
              and a penalty is not an exclusion.
Resolved-by:  #248, #255, #256, #271
```

## FR-126 — Bound the private run that reaches a location the user entered

```
Statement:    The system shall reach or leave a location the user entered — a
              required location of the journey or a stopping position the user
              chose — that lies on a way the map data marks as private or
              access-restricted by a driving leg that runs along such ways,
              and passes such barriers, at that location only in one run
              joining the location to the first way not so marked, ending at
              no stopping position the system chose that no warned leg FR-129
              admits reaches, and carrying no more such travel than the least
              that joins the two.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract holding a
              location the user entered on a marked way, a leg leaving or
              reaching it starts or ends at it and leaves the marked ways at
              the first way not so marked, a stopping position the system
              chose on that run is reached from the location along none of it
              unless a warned leg FR-129 admits reaches it, and where two runs
              join the location to ways not so marked the
              leg takes no more marked length than the shorter
Acceptance:   given a location the user entered — the journey's origin, its
              destination, an intermediate destination or a stopping position
              the user chose — lying on a way so marked that a car reaches
              only along such ways, when a leg leaving or reaching that
              location is produced, then the leg starts or ends at that
              location, and its travel along such ways and past such barriers
              there is one run that ends at the first way not so marked it
              meets
              given a stopping position the system chose, lying on such ways
              between a location the user entered on a way so marked and the
              ways not so marked nearest that location, that no warned leg
              FR-129 admits reaches, when the journey is planned, then no
              driving leg runs along such ways between that location and that
              stopping position
              given a location the user entered on a way so marked, joined to
              ways not so marked by two runs of which the quickest leg would
              take the one carrying more such travel, when that leg is
              produced, then the length of such ways it runs along at that
              location is no more than the least that joins the location to a
              way not so marked
Status:       to-build
Depends-on:   FR-006; FR-129
Volatility:   open-question — the markings this record reads are FR-125's,
              their spelling designed in
              `Architecture.md § The access-marking tables` and their floor a
              value outside the permissive set on the most specific key
              carried, as `DECISIONS.md § RD-063` reads
              `DECISIONS.md § RD-062`'s allow-list, with the keys' directional
              and conditional forms read as `DECISIONS.md § RD-064`,
              `DECISIONS.md § RD-065` and `DECISIONS.md § RD-066` decide; that
              a stop the system chose on the run of a warned leg is reached is
              the Owner's in `DECISIONS.md § RD-079`
Risk:         Read wider than the least run to a way not so marked, the
              exception becomes transit: a home on a private road whose marked
              network runs on through an estate to a second public road is
              left by the far side when that is quicker, and a stop the system
              chose on the run draws one leg into the area and the next out of
              it, so the journey crosses the area. Read narrower, a home on a
              private road is not planned to the door, which the Owner ruled
              it is.
Rationale:    The member admits no route through a marked area, and the
              Owner's decisions on PR #251 let a leg enter one to reach a
              location the user entered, because the user has a reason to be
              there, and bound the run that does it (`DECISIONS.md § RD-056`,
              `DECISIONS.md § RD-057`). A stopping position the user chose is
              such a location, and one the system chose is not
              (`DECISIONS.md § RD-061`), one the system offered and the user
              accepted included (FR-136). The three bounds are one record
              because together they are the run, and each is a criterion
              because each closes one way the run becomes transit: the first
              keeps it to the stretch between the location and the first way
              not so marked; the second keeps a stop the system chose off it,
              a leg ending at such a stop on the run being a leg along marked
              ways to no location the user entered; the third refuses the far
              exit. The first way not so marked is the one the least marked
              travel reaches, and that travel is measured as length, so a fast
              road through the area is never the least because it is quick
              (`DECISIONS.md § RD-057`). A home behind a gate marked private,
              on a driveway the map data does not mark, lies on a way not so
              marked: its run is empty and the gate is passed in transit,
              which FR-125 refuses, as the Owner accepted, unless FR-129
              admits the leg as a warned leg because its other end is a
              location the user entered too (`DECISIONS.md § RD-067`). A stop
              the user chose on a way not so marked behind such a gate is the
              same case, and FR-127 refuses a leg ending short of it. On a
              warned leg the run is part of the private stretch the user
              confirms, and the Owner allows a stop the system chose on it
              (`DECISIONS.md § RD-079`): the statement's exception and the
              second criterion's reach only a stop a warned leg FR-129 admits
              reaches, whose legs FR-129 holds to the warned leg's own ways,
              so the run is still never transit.
Resolved-by:  #248, #256
```

## FR-127 — Serve no stopping position past a private or access-restricted way or barrier

```
Statement:    The system shall serve a stopping position by a driving leg
              ending short of it only where no way or barrier that the map
              data marks as private or access-restricted lies between the
              leg's end and the position.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract, a
              stopping position a car reaches only along a marked way or past
              a marked barrier is served by no leg where the system chose it
              and no warned leg FR-129 admits reaches it, and where the user
              chose it on a way not so marked by no leg ending short of it
              and by a leg ending at it only as a warned
              leg FR-129 admits, and one with
              nothing so marked between it and the end of the leg serving it
              is not refused by this record, and a leg ending short of a
              position that a connection avoiding every marking also reaches
              is refused where a marked way or barrier lies on a drivable
              connection between the leg's end and the position
Acceptance:   given a stopping position the system chose, that a car reaches
              only by running along a way, or passing a barrier, that the map
              data marks as private or access-restricted — a parking area
              behind a gate so marked — and that no warned leg FR-129 admits
              reaches, when the journey is planned, then no driving leg ends
              at it or short of it, and the journey includes no stop there
              given a driving leg ending short of a stopping position, no way
              or barrier so marked lying between the leg's end and the
              position, when the journey is planned, then this record refuses
              neither the leg nor the stop
              given a stopping position the user chose, lying on a way not so
              marked, that a car reaches only by running along a way, or
              passing a barrier, so marked — a parking area not itself so
              marked behind a gate so marked — when the journey is planned,
              then no driving leg ends short of it, and a driving leg ends at
              it only as a warned leg FR-129 admits
              given a stopping position that a car also reaches along a
              drivable connection avoiding every way and barrier so marked — a
              parking area with a public entrance on its far side — and a
              driving leg ending short of it at a point from which a drivable
              connection to the position runs along such a way or passes such
              a barrier — the leg ending before a gate so marked on the
              parking area's near side — when the journey is planned, then no
              driving leg ends short of it at that point, and a leg serving
              it ends at it, or short of it with nothing so marked between
Status:       to-build
Depends-on:   FR-125; FR-126; FR-129
Volatility:   open-question — the markings this record reads are FR-125's,
              their spelling designed in
              `Architecture.md § The access-marking tables` and their floor a
              value outside the permissive set on the most specific key
              carried, as `DECISIONS.md § RD-063` reads
              `DECISIONS.md § RD-062`'s allow-list, with the keys' directional
              and conditional forms read as `DECISIONS.md § RD-064`,
              `DECISIONS.md § RD-065` and `DECISIONS.md § RD-066` decide; how
              far short of its position a leg may end is a tolerance no
              document gives
Risk:         A leg that ends at a gate marked private, within the tolerance
              for how far short of its position a leg may end, serves the
              parking area behind it: the driver drives through a gate they
              have no right to open or finds it locked, and the car half of
              the stop ends where the walk half does not start.
Rationale:    FR-125 keeps a leg to a stopping position off such ways and
              barriers, and a leg that cannot reach the position along any
              other connection may still end short of it, which FR-125 alone
              does not refuse. The Owner's decision makes the gap part of what
              is checked: ending short honours the exclusion only where
              nothing so marked lies between the leg's end and the position,
              and otherwise there is no leg and the stop is not served
              (`DECISIONS.md § RD-058`). A way or barrier lies between them
              where it lies on a drivable connection joining the two, since
              the driver closes the gap by driving it. The rule binds every
              stopping position a leg ends short of. One the system chose is
              outside the Owner's exception (`DECISIONS.md § RD-052`), which
              the first criterion tests, unless a warned leg FR-129 admits
              reaches it, as the Owner allows along the warned leg a location
              the user entered requires (`DECISIONS.md § RD-079`); and one
              the system offered and the user accepted is one the system
              chose (FR-136, `DECISIONS.md § RD-069`). One the user chose is a
              location the
              user entered (`DECISIONS.md § RD-061`): on a way so marked,
              FR-126's run reaches it and ends at it, never short of it; on a
              way not so marked, its run is empty, so a marking between it and
              the leg's end leaves it unserved by a leg ending short of it,
              and a leg reaches it across the marking only as a warned leg
              FR-129 admits, whose other end is a location the user entered
              too (`DECISIONS.md § RD-067`), which the third criterion tests.
              A marking on any connection joining the leg's end to the
              position lies between them, so a public entrance elsewhere does
              not admit a leg ending at the gate, which the fourth criterion
              tests. The second criterion keeps the rule to the gap, so a
              position a few metres off the drivable network with nothing so
              marked between is not refused by it; how far that gap may run is
              a tolerance this record does not set. That a stopping position
              the system chose behind such a marking is not offered at all, as
              a recommended stop, a replacement or a reserve candidate, is
              FR-138's (`DECISIONS.md § RD-074`).
Resolved-by:  #256
```

## FR-128 — Treat the ways inside an area marked private or access-restricted as marked

```
Statement:    The system shall treat a way, or the part of one, lying within
              an area that the map data marks as private or access-restricted
              on an area feature enclosing it as a way so marked for FR-125,
              FR-126, FR-127, FR-129, FR-130, FR-131, FR-137 and FR-138.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract holding an
              area feature so marked whose own ways carry no marking, a leg
              whose quickest connection crosses the area is produced on a
              connection avoiding it, a location the user entered inside it is
              reached along the run FR-126 admits, and a stopping position the
              system chose reached only through it, and by no warned leg
              FR-129 admits, is served by no leg, and a leg with one end at a
              location the user entered runs within no such area beyond
              FR-126's run where a connection avoids it
Acceptance:   given two points on ways outside an area that the map data marks
              as private or access-restricted on an area feature enclosing it,
              whose quickest drivable connection runs along a way within that
              area carrying no such marking of its own, and between which a
              drivable connection avoiding every such area exists, when the
              leg is produced, then it runs along no way within that area
              given a location the user entered — a required location of the
              journey or a stopping position the user chose — lying within
              such an area on a way carrying no such marking of its own, when
              a leg leaving or reaching that location is produced, then the
              leg starts or ends at that location and runs within the area
              only in the run FR-126 admits
              given a stopping position the system chose that a car reaches
              only along ways within such an area, and that no warned leg
              FR-129 admits reaches, when the journey is planned, then no
              driving leg ends at it or short of it, and the journey includes
              no stop there
              given a driving leg one end of which is a location the user
              entered lying within such an area or on a way so marked, whose
              quickest drivable connection beyond the run FR-126 admits at
              that location runs along a way within such an area, when the
              journey is planned, then beyond that run the leg runs along no
              way within such an area, being produced on a drivable
              connection from that run's end that avoids every such area
              where one exists, and not produced where none does other than
              as a warned leg FR-129 admits
Status:       to-build
Depends-on:   FR-125; FR-126; FR-127; FR-129; FR-130; FR-131; FR-137; FR-138
Volatility:   open-question — which area features, and which tags on them,
              spell the marking is proposed in
              `Architecture.md § What the tables hold` and not yet confirmed;
              an area feature is so marked where the most specific of
              `DECISIONS.md § RD-059`'s keys it carries has a value outside
              the permissive set, the precedence and the set being
              `DECISIONS.md § RD-063`'s reading of
              `DECISIONS.md § RD-062`'s allow-list, with the keys'
              directional and conditional forms read as
              `DECISIONS.md § RD-064`, `DECISIONS.md § RD-065` and
              `DECISIONS.md § RD-066` decide; and a spelling found wrong is
              corrected without narrowing it
Risk:         A car park, an estate or a compound drawn as one area and marked
              customers-only or private on that area alone routes as public
              road, because its own ways carry no marking, and the leg cuts
              through it exactly where that saves time. Every check reading
              the ways' own markings passes.
Rationale:    The cited section binds the area, and FR-125 reads the area's
              marking from what a leg is made of. Where the map data marks an
              area only on the feature enclosing it, its ways are unmarked and
              FR-125 alone admits the cut-through the member forbids. Treating
              those ways as so marked for FR-125, FR-126 and FR-127 states the
              rule, the Owner's exception and its bound once, so an area so
              marked is crossed, entered and served exactly as a marked way is
              (`DECISIONS.md § RD-060`). Treating them so for FR-129 to FR-131
              as well makes a warned leg across such an area a warned leg as
              one across a marked way is, its avoidable crossings kept off it
              and its stretch within the area identified
              (`DECISIONS.md § RD-067`). Treating them so for FR-137 and
              FR-138 keeps the area from reopening what those records close: a
              stopping position the system chose beyond a gated estate drawn
              as one area so marked, whose own ways carry no marking, lies
              behind the area's marking, and without this record no check
              reading the ways' own markings sees it, so the position and the
              zones a walk from it are offered and the driver is sent through
              the estate for zones the planner picked, where no warned leg
              FR-129 admits reaches it (`DECISIONS.md § RD-079`). Those two
              records' own criteria test the area, so this record's do not. A
              stop the user chose in a customers-only car park drawn as such
              an area is therefore reached along FR-126's run, as
              `SPECIFICATION.md § What establishes a stopping position` keeps
              it reachable (`DECISIONS.md § RD-061`). A way's own markings do
              not lift the marking of the area enclosing it. The record
              reaches a driving leg; the walking access path's half stays owed
              with FR-075 until an access path is built. No mechanism is
              named, and a penalty is not an exclusion.
Resolved-by:  #253
```

## FR-129 — Produce a warned leg where nothing avoids a marking between two locations the user entered

```
Statement:    Where every drivable connection between two locations the user
              entered, with no other location the user entered between them in
              the journey, runs along a way, or passes a barrier, that the map
              data marks as private or access-restricted, other than in the
              run FR-126 admits at either, the system shall produce as a
              warned leg, rather than refuse the journey, each driving leg
              between them that runs along or passes such a way or barrier,
              where the driving legs between them together pass a barrier so
              marked, or drive along a part of a way that is so marked or that
              a car reaches only across such a way or barrier, only where the
              warned leg the system would produce joining the two were no stop
              between them, a leg the two locations alone fix, passes that
              barrier or drives along that part, and no more times than that
              leg does, or where that part is one not itself so marked within
              a parking area establishing a stop between them, driven on a leg
              to or from that stop, and a car turns directly, past no such
              barrier, onto a part not itself so marked within that parking
              area from a part that warned leg drives along and that is so
              marked or that a car reaches only across such a way or barrier.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract, a leg
              between two locations the user entered every drivable connection
              between which crosses a marked way or barrier beyond FR-126's
              run is produced and identified as warned; with a parking area
              the system chose beside the part of the private road that leg
              drives, and separately a stop the system chose on a public road
              beyond the marked gate it passes, between them, each leg between
              the two locations that crosses a marking is produced and
              identified as warned; with one stop the system chose behind that
              gate and one outside it, the legs between the two locations pass
              the gate no more times than the warned leg joining them with no
              stop between them does; with one stop the system chose in a
              parking area beside the part of the private road that leg drives
              and one in a second parking area that a car reaches only through
              the ways within the first, the journey includes no stop in the
              second, though it may hold one in the first; with stops the
              system chose in the first on both sides of one in the second,
              where an unmarked part within the second meets the public road
              that leg drives before a marked gate of its own, and separately
              where the second meets that private road only along a marked
              part or past a marked barrier of its own, the journey includes
              no stop in the second; with one stop
              the system chose in a parking area drawn within a larger one
              that meets that private road where the smaller does not, the
              stop is reached by warned legs along the unmarked ways within
              the larger one; with one stop the system chose where two parking
              areas overlap, only one of which meets that road, no leg to it
              drives along a part within the other outside the first; with one
              stop the system chose in a parking area beside that private road
              that a car reaches from the road directly and, more briefly,
              through the ways within another parking area beside it, the leg
              to it drives along no part within the other; a leg to a parking
              area the system chose behind a second marked gate off that
              private road, to one beside a single unsplit way past the
              location the user entered that the warned leg reaches along that
              way, where a gate node stands on the way with an unmarked part
              of it before the gate, to one beside the other of two
              connections that cross markings alike where the warned leg with
              no stop between them takes the first, to one that a car reaches
              only through the ways within another parking area beside the
              part of the private road that leg drives, or to one behind a
              marked gate between two locations the user entered that a
              connection avoiding every marking joins, is produced neither
              warned nor otherwise; and a leg between two locations the user
              entered that a connection avoiding every marking joins is
              produced on such a connection and not identified as warned
Acceptance:   given two locations the user entered — each the journey's
              origin, its destination, an intermediate destination or a
              stopping position the user chose — that one driving leg joins,
              every drivable connection between which runs along such a way
              or passes such a barrier beyond the run FR-126 admits at
              either, when the journey is planned, then that leg is produced
              and identified as a warned leg, and the journey is not refused
              for it
              given a driving leg one end of which is a stopping position the
              system chose, that runs along such a way or passes such a
              barrier beyond any run FR-126 admits, where a drivable
              connection avoiding every such way and barrier joins the
              locations the user entered on either side of the stop, or where
              every pair of connections joining the stop to those two
              locations together passes such a barrier, or drives along a part
              of a way so marked or that a car reaches only across such a way
              or barrier, that the warned leg the system would produce joining
              the two were no stop between them does not pass or drive along,
              or more times than it does, other than a part not itself so
              marked within a parking area establishing the stop where a car
              turns directly, past no such barrier, onto a part not itself so
              marked within that parking area from a part that warned leg
              drives along and that is so marked or that a car reaches only
              across such a way or barrier — a parking area behind a second
              gate off that leg's private road; one beside a single unsplit
              way past the location the user entered that the warned leg
              reaches along it, beyond a gate standing on that way; of two
              connections that cross markings alike, one beside the connection
              that warned leg does not take; or one that a car reaches from
              the part of the private road that warned leg drives only through
              the ways within another parking area beside it, though it meets,
              past a marked barrier or along a marked part of its own, that
              private road or a public road that warned leg drives — when the
              journey is planned, then that leg is not produced, warned or
              not, and the journey includes no stop there
              given stopping positions the system chose between two locations
              the user entered, every drivable connection between which runs
              along such a way or passes such a barrier beyond the run FR-126
              admits at either, that a car reaches from the two locations by
              legs that together pass such a barrier, or drive along a part of
              a way so marked or that a car reaches only across such a way or
              barrier, only where the warned leg the system would produce
              joining the two were no stop between them passes or drives along
              it, and no more times than it does, or along a part not itself
              so marked within a parking area establishing a stop, on the legs
              to and from that stop, where a car turns directly, past no such
              barrier, onto a part not itself so marked within that parking
              area from a part that warned leg drives along and that is so
              marked or that a car reaches only across such a way or barrier —
              a parking area beside the part of the private road that leg
              drives, or one on a public road beyond the gate it passes — when
              the journey is planned, then each driving leg between the two
              locations that runs along such a way or passes such a barrier is
              produced and identified as a warned leg, the journey is not
              refused for it, and the driving legs between the two locations
              together pass or drive along no such barrier or part but those,
              and none more times than that warned leg does but a part so
              allowed within such a parking area, so that a journey holding
              one such stop behind the gate and one outside it passes the gate
              no more times than that warned leg passes it, and one holding
              stops in a parking area beside the part of the private road that
              leg drives holds none in a parking area that a car reaches from
              that road only through the ways within the first
              given two locations the user entered that one driving leg
              joins, between which a drivable connection avoiding every such
              way and barrier beyond FR-126's run exists, when the journey is
              planned, then the leg is produced on such a connection and is
              not identified as a warned leg
Status:       to-build
Depends-on:   FR-006; FR-126
Volatility:   open-question — the markings are FR-125's, read as
              `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066` decide; the
              ruling is `DECISIONS.md § RD-067`'s, and its first edge, that a
              leg is warned only where both its ends are locations the user
              entered, the Owner's in `DECISIONS.md § RD-068`, superseded
              along the warned leg by the Owner's ruling in
              `DECISIONS.md § RD-079`, whose "no detours onto other private
              roads" the statement's bound reads, measured as
              `DECISIONS.md § RD-080` decides, its parking-area exception
              narrowed by `DECISIONS.md § RD-081` and written as what the legs
              may use by `DECISIONS.md § RD-082`; the exception in
              `SPECIFICATION.md § Enforceable exclusions` is the amendment PR
              #274 drafts for the Owner's approval at merge
Risk:         Read wider than the warned leg, to every stop behind the marking
              it crosses, the warning becomes the price of a zone: the user
              confirms the gate their own location lies behind, and the
              planner sends the driver deeper into the gated site, or through
              a second gate, for a stop it picked, which the Owner ruled out.
              Measured over the ways the warned leg runs along rather than
              the part of each it drives, the bound turns on how the map data
              splits its ways, and an unsplit road past the user's stop
              carries the driver on into the site; measured over which
              markings are crossed and not how often, the planner leaves and
              re-enters through the gate the user confirmed once.
              Excepting the unmarked parts within every parking area holding a
              stop, on every leg, the planner sends the driver through the
              aisles of a parking area beside the stretch into a second one
              deeper in the site, for a zone it picked.
              Counting a parking area as meeting the leg wherever a part
              within it touches any part the leg drives, a public road or the
              stretch past a marked entrance of its own included, the planner
              sends the driver through a car park beside the stretch into a
              gated staff car park beyond it, for a zone it picked.
              Read as both ends the user's, a parking area beside the private
              stretch the user confirmed cannot serve the zones along it,
              which the Owner ruled it may. Built as the refusal it replaces,
              a user whose home lies behind a private gate gets a routing
              error and no journey, which the Owner ruled against. Built as an
              ordinary leg without the identification, a client draws a
              private crossing as a road the driver may take, and nothing in
              the result says so.
Rationale:    FR-125 refuses a leg nothing joins without a marking, and the
              Owner ruled that between locations the user entered the route is
              shown instead, warned (`DECISIONS.md § RD-067`): the user has a
              reason to be at each end (`DECISIONS.md § RD-056`) and decides
              whether to drive it. The identification is what every later
              record reads: FR-130 keeps the avoidable markings off the leg,
              FR-131 names its marked stretches, FR-132 and FR-133 show and
              warn of them, and FR-134 holds the journey unconfirmed until the
              user confirms. The Owner first ruled that only a leg between two
              locations the user entered may be warned
              (`DECISIONS.md § RD-068`), and then that the route a location
              the user entered or placed requires may have zones along it,
              "even if the stop for them is on the private stretch you already
              accepted", with "no detours onto other private roads" and the
              warning kept (`DECISIONS.md § RD-079`). A stopping position the
              system chose between the two locations divides the warned leg
              into legs, so the statement warns each that crosses a marking
              and bounds them together by the warned leg the system would
              produce joining the two with no stop between them, as a list of
              what they may use behind a marking and not as exceptions to a
              count (`DECISIONS.md § RD-082`): the marked barriers that leg
              passes and the parts of ways it drives that are marked or that a
              car reaches only across a marking, each no more times than it
              does, and the unmarked parts within a parking area establishing
              a stop that meets the stretch the user accepted, on the legs to
              and from that stop (`DECISIONS.md § RD-080`,
              `DECISIONS.md § RD-081`); nothing else behind a marking.
              Reaching a part only across a marking is what a gate mapped on
              an unmarked road needs: the ways past it carry no marking, and
              without it every stop past the same gate rides on the user's
              confirmation. The bound is over parts of ways and not ways, so
              that it holds however the map data splits them; over how often
              and not only which, so that the legs never leave and re-enter
              through the marking the user confirmed once; and against the leg
              produced with no stop, so that where markings tie the planner
              never picks the private road carrying its own zones, the route
              being the one the user's locations cause and not one the system
              chose for a zone. A parking area meets the stretch where a car
              turns directly, past no marked barrier, onto an unmarked part
              within it from a part that leg drives that is marked or that a
              car reaches only across a marking. Meeting a public road the leg
              drives, or the stretch only along a marked part or past a marked
              barrier of the parking area's own, is not meeting it, so the
              list never carries the driver through one parking area into
              another deeper in the site, whatever the second touches
              elsewhere, nor across a parking area on the way to some other
              stop, which is driving for that other stop and not for its own.
              Where a stop lies within more than one parking area, nested or
              overlapping, each establishes it and each is tested on its own,
              so the driver stays within one that meets the stretch
              (`DECISIONS.md § RD-082`). The second criterion holds the bound
              at its edges, the journey that needs no warned leg among them,
              where a stop the system chose behind a marking is still never
              reached; the third holds the ruling, a stop beside the private
              stretch, or on a public road beyond the gate, being reached by
              warned legs and the journey not refused. The fourth keeps the
              warning a last resort: it replaces a refusal and never a leg
              FR-125 admits, so a build that warns where it could avoid fails.
              The cited section carries the exception on the amendment the
              Owner approves at PR #274's merge (`DECISIONS.md § RD-068`), and
              a stop along the route adds no crossing to it. A stopping
              position the system offered stays one the system chose after the
              user accepts it (FR-136, `DECISIONS.md § RD-069`), so an
              acceptance never makes it a location the user entered with a
              warned leg of its own; and zones are taken from every stop these
              legs reach and from no other behind a marking (FR-137). No
              mechanism and no wire shape is named.
Resolved-by:  #271
```

## FR-130 — Keep every avoidable marking off a warned leg

```
Statement:    The system shall produce no warned leg that runs along a way,
              or passes a barrier, that the map data marks as private or
              access-restricted where another drivable connection between the
              leg's ends avoids that way or barrier and runs along and passes
              no such way or barrier that the warned leg does not.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract, a warned
              leg whose ends every drivable connection joins only past one
              marked barrier, and whose quickest connection also runs along a
              marked way further on that a connection past that barrier alone
              avoids, runs along no such way; and a warned leg whose ends
              every connection joins past one of two marked barriers passes
              one and not both
Acceptance:   given two locations the user entered that every drivable
              connection joins only past a barrier so marked, whose quickest
              drivable connection also runs along a way so marked further on
              that a connection passing that barrier alone avoids, when the
              journey is planned, then the warned leg passes that barrier and
              runs along no such way
              given two locations the user entered that every drivable
              connection joins only past one of two barriers so marked, each
              avoidable by a connection passing the other, when the journey
              is planned, then the warned leg passes one of them and not both
Status:       to-build
Depends-on:   FR-129
Volatility:   open-question — the markings are FR-125's, read as
              `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066` decide; which
              connection a warned leg takes, the second edge
              `DECISIONS.md § RD-067` carried, is the Owner's "least private"
              in `DECISIONS.md § RD-068`
Risk:         A warned leg taken as the engine's quickest answer carries every
              marking on it, the unavoidable one and any other the quickest
              connection happens to take: the user confirms a warning about
              the gate their home lies behind and is sent along a residents'
              road further on that a short way round avoids, which no leg not
              warned would ever run along.
Rationale:    The ruling turns a refusal into a warned route because nothing
              joins the user's locations without a marking, and nothing in it
              admits a marking that something avoids
              (`DECISIONS.md § RD-067`); the Owner ruled it so, that the
              warned leg crosses no marking it could have avoided
              (`DECISIONS.md § RD-068`). FR-125's first criterion avoids such
              a marking on a leg not warned, and this record holds the same on
              a warned leg: the warning covers only the markings the user's
              locations cannot be joined without, and the leg is held to the
              rule everywhere else. It is stated over the set of markings
              crossed, never over length or time, because a set is what
              avoiding a marking changes and what the warning lists: a
              shorter or quicker connection that trades one marking for
              another is no less a crossing. Which connection the leg then
              takes among those the record leaves is not obliged; the ruling's
              "show the route" names none.
Resolved-by:  #271
```

## FR-134 — Confirm no journey holding a warned leg until the user confirms its warning

```
Statement:    The system shall let no journey alternative holding a warned
              leg become the confirmed route until the user, after its marked
              stretches and its warning were presented, has explicitly
              confirmed taking it.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions;
              DESIGN.md § Reviewing the recommended route
Priority:     MUST
Verification: test — a presented journey alternative holding a warned leg
              becomes the confirmed route, and has a portion dispatched or is
              stored, only after the user's explicit confirmation of its
              warning, an alternative holding no warned leg asks for none,
              and a change after confirmation that adds a marked stretch asks
              again
Acceptance:   given a presented journey alternative holding a warned leg
              whose warning the user has not confirmed, when the user accepts
              the alternative, then it does not become the confirmed route, no
              portion of it is dispatched and it is not stored
              given the same alternative, when the user explicitly confirms
              its warning, then it may become the confirmed route as the
              review otherwise allows
              given a presented journey alternative holding no warned leg,
              when the user accepts it, then no confirmation of a warning is
              asked for
              given a route whose warning the user confirmed, when a change to
              it, in the review or in a later review of its stored plan, makes
              it hold a warned leg or a marked stretch that the confirmed
              warning did not present, then the changed route does not become
              the confirmed route, no portion of it is dispatched and it is
              not stored until the user confirms its warning afresh
Status:       to-build
Depends-on:   FR-131; FR-132; FR-133
Volatility:   open-question — where the step sits in the review is owed by
              `DESIGN.md § Reviewing the recommended route`; what "used"
              means, and that one confirmation covers every warned leg of an
              alternative, the third edge `DECISIONS.md § RD-067` carried, is
              the Owner's in `DECISIONS.md § RD-068`; that a change adding a
              crossing asks again is `DECISIONS.md § RD-071`'s reading of it
Risk:         A warning the user can pass without acting on it is a warning in
              name: the route goes to the navigation app, which shows no
              marking, and the driver meets the gate with the plan's
              endorsement and nothing on the screen in front of them to
              reconsider it by.
Rationale:    The Owner's scope has the user confirm before the route is used
              (`DECISIONS.md § RD-067`), and the Owner ruled that "used" is
              the journey becoming the confirmed route
              (`DECISIONS.md § RD-068`): a dispatch is sized
              from a confirmed route (FR-110) and a stored plan is the
              persisted confirmed plan, so holding the confirmation there
              gates both with one rule, and the first criterion tests both.
              The confirmation is an act of its own, never implied by
              accepting the alternative or its zones, and it follows the
              presentation, so the user confirms what they were shown. The
              Owner ruled one confirmation per alternative, the warning being
              about the route, which the Owner's words address whole, and
              that the alternative may be viewed unconfirmed: the record holds
              back the confirmed route and nothing before it. The
              third criterion keeps the step to the warned journey, so a
              build that asks on every route, and so trains the user to
              confirm without reading, fails. The fourth keeps the
              confirmation to what was shown: a change in the review, or in
              a later review re-solved from what FR-096 retains, that brings
              a crossing the confirmed warning never presented makes a route
              the user did not confirm, and "once per alternative" counts the
              alternative shown (`DECISIONS.md § RD-071`).
Resolved-by:  #273
```

## FR-136 — Hold a stopping position the system offered as the system's once the user accepts it

```
Statement:    The system shall hold a stopping position it offered — the
              position of a recommended stop, of a replacement or of a reserve
              candidate — as a stopping position the system chose, and not as
              one the user chose, after the user accepts it.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions;
              SPECIFICATION.md § What establishes a stopping position
Priority:     MUST
Verification: test — against routing data built from a test extract, a reserve
              candidate whose stopping position a car reaches only past a
              marked gate, or only along a marked way, that no warned leg
              FR-129 admits reaches, offered after a rejection and accepted,
              is served by no driving leg, warned or not, while a stopping
              position the user placed at the same point is not refused by
              this record
Acceptance:   given a stopping position the system offered — a recommended
              stop, a replacement or a reserve candidate — that a car reaches
              only by running along a way, or passing a barrier, that the map
              data marks as private or access-restricted — a parking area
              behind a gate so marked, or one on a service road so marked —
              and that no warned leg FR-129 admits reaches, when the user
              accepts it and the journey is planned, then no driving leg ends
              at it or short of it, warned or not, and the journey includes
              no stop there
              given a stopping position the user placed at the same point,
              when the journey is planned, then this record refuses neither
              the leg reaching it nor the stop
Status:       to-build
Depends-on:   FR-127; FR-129
Volatility:   open-question — the markings are FR-125's, read as
              `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066` decide; who
              chose a stop rests on the Owner's scoping answer as
              `DECISIONS.md § RD-069` reads it; how a user places a stopping
              position other than as a location of the journey is FR-139's,
              its interaction owed to `DESIGN.md` as the Owner directed in
              `DECISIONS.md § RD-073`; that one a warned leg FR-129 admits
              reaches is served is the Owner's in `DECISIONS.md § RD-079`
Risk:         Read as the user's once accepted, the exception becomes the
              price of a zone: the reserve offers a car park behind a private
              gate, the user accepts it, and one confirmation sends the driver
              through the gate for a stop the system picked, every check
              reading a location the user entered.
Rationale:    The Owner's scope gives "your own stops" a warned route and
              keeps "Stops the app picks itself (zones) behind private roads"
              never offered, since "the app can just pick another zone"
              (`DECISIONS.md § RD-069`). Nothing said who chose a stop, and
              `SPECIFICATION.md § What establishes a stopping position` has
              the user "place the stop there and have it recorded as
              user-chosen", which an acceptance could be read to satisfy. The
              offer settles it: the system picked the position, and accepting
              it judges that pick and places nothing, as
              `DESIGN.md § Reviewing zones one at a time` gives a zone two
              actions, accept it or ask for a different zone. A position so
              stays the system's however it reached the plan — recommended,
              replaced, drawn from the reserve or reinstated by an undo — and
              is the user's only where the user placed it; a required
              location named as a zone (`DESIGN.md § Locations`) is a location
              the user entered, as before. Stated once here, the rule binds
              every record turning on who chose a stop: FR-126's run, FR-127's
              gap, FR-128's area and FR-129's warned leg, and FR-126, FR-127
              and FR-129 point to it. It removes nothing the cited section
              keeps, which reaches a restricted position by the user placing
              the stop there; the second criterion keeps the record to the
              offer, so a build that refuses every stop at such a point
              fails. FR-138 withholds from the offer a stopping position the
              system chose behind a marking, and no way a car runs along to
              the position is the position's own (`DECISIONS.md § RD-075`), so
              the first criterion is the layer beneath it: it holds where such
              a position reaches the plan regardless, and a test of it tells
              the planner directly which position was offered and accepted.
              One the system offered on a parking area whose own feature
              carries the marking is served on acceptance in no geometry,
              which is FR-140's (`DECISIONS.md § RD-078`). One that a warned
              leg FR-129 admits reaches is served along that leg because a
              location the user entered brought the marking into the journey,
              never because the user accepted it (`DECISIONS.md § RD-079`),
              and the first criterion's last clause keeps the two apart: the
              acceptance places nothing, so it never causes a warned leg of
              its own. A stop the user places at an offered candidate's point
              is a stop the user placed, and the candidate stays the system's
              (FR-139).
Resolved-by:  #275
```

## FR-137 — Offer no zone taken from a stop behind a marking that no warned leg FR-129 admits reaches

```
Statement:    The system shall offer no zone it chose to be taken from a stop
              that a car reaches only by running along a way, or passing a
              barrier, that the map data marks as private or
              access-restricted, other than a stop at a location the user
              entered that a leg reaches along the run FR-126 admits there, or
              a stop that a warned leg FR-129 admits reaches.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract, the
              journey's origin, an intermediate destination, its destination
              and a stopping position the user chose, each in turn behind a
              marked gate standing on another way and joined to the next
              location the user entered by a warned leg, a stopping position
              the user chose past a marked gate on the unmarked road it lies
              on, a destination on a driveway itself marked reached along
              FR-126's run, and a parking area the system chose beside the
              part of the private road a warned leg drives, each with a zone a
              short walk away, are refused no zone by this record; while a
              parking area the system chose behind a second marked gate off
              that road, one beside a single unsplit way past the location the
              user entered that the warned leg reaches along it, beyond a gate
              node on that way, and one behind a marked gate in a journey
              whose locations the user entered a connection avoiding every
              marking joins, each with a zone a short walk away, have no zone
              the system chose offered to be taken from a stop there
Acceptance:   given a location the user entered — the journey's origin, its
              destination, an intermediate destination or a stopping position
              the user chose — that a car reaches only by running along a way,
              or passing a barrier, so marked, and that a warned leg FR-129
              admits reaches or a leg reaches along the run FR-126 admits
              there — a home behind a gate on its own driveway, a street
              within an estate entered through a gated road, a stop the user
              placed past a gate on the road it lies on, a home on a driveway
              itself so marked — with a zone a walk from it could take, when
              the journey is planned, then this record refuses no zone taken
              from a stop there, and a leg reaching it across a marking beyond
              that run is still a warned leg
              given a stopping position the system chose that a warned leg
              FR-129 admits reaches — a parking area beside the part of the
              private road that the warned leg between two locations the user
              entered drives — with a zone a walk from it could take, when the
              journey is planned, then this record refuses no zone taken from
              a stop there
              given a stop that a car reaches only by running along a way, or
              passing a barrier, so marked, that no warned leg FR-129 admits
              reaches and that is at no location the user entered reached
              along FR-126's run — a parking area behind a second gate off the
              private road a warned leg runs along, one beside a single
              unsplit way past the location the user entered that the warned
              leg reaches along it, beyond a gate standing on that way, or one
              behind a gate so marked in a journey whose locations the user
              entered a connection avoiding every such way and barrier joins
              — with a zone a walk from it could take, when the journey is
              planned, then no zone the system chose is offered to be taken
              from a stop there, in the journey, as a replacement or as a
              reserve candidate
Status:       to-build
Depends-on:   FR-126; FR-129; FR-136
Volatility:   open-question — the markings are FR-125's, read as
              `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066` decide; that
              zones are taken along a warned leg a location the user entered
              or placed requires, and from no stop off it behind a marking, is
              the Owner's in `DECISIONS.md § RD-079`, bounded as FR-129
              states and measured as `DECISIONS.md § RD-080`,
              `DECISIONS.md § RD-081` and `DECISIONS.md § RD-082` decide; that
              a location reached along FR-126's run keeps its zones is the
              Owner's in `DECISIONS.md § RD-072`
Risk:         Built on the exclusion this record replaced, a user whose home
              or placed stop lies behind a private gate is stripped of the
              zones around it after confirming the warned route there, which
              the Owner ruled against. Built on the gate alone, every stop
              behind the gate the user confirmed rides on that confirmation:
              the planner sends the driver deeper into the gated site, or
              through a second gate, for zones it picked, which is the detour
              the Owner ruled out. Built on the warned leg's two ends alone, a
              parking area beside the private stretch the user confirmed
              cannot serve the zones along it, which the Owner ruled it may.
              Built on the ways the warned leg runs along rather than the part
              of each it drives, a stop past the user's own on an unsplit
              private road keeps the zones the planner picked there.
Rationale:    The Owner ruled that the app never on its own suggests zones and
              routes along private roads, and that where a stop the user
              places needs one, "the route is okay to have zones along it",
              the warning still shown, "even if the stop for them is on the
              private stretch you already accepted", with "no detours onto
              other private roads" (`DECISIONS.md § RD-079`). A zone is taken
              on foot from a stop
              (`SPECIFICATION.md § Stops serving several zones`), so the
              record reads the stop and how a car reaches it. A location the
              user entered behind a marking is reached by a warned leg FR-129
              admits, or along FR-126's run, whoever entered it and however,
              so every such location keeps its zones, which the first
              criterion tests: the bounds `DECISIONS.md § RD-070`,
              `DECISIONS.md § RD-072` and `DECISIONS.md § RD-076` drew
              location by location, which turned on the way a gate stands on
              and so on how the map data splits its ways, are superseded where
              they took zones away. A stop the system chose is held to the
              warned leg's bound, stated once in FR-129 as what the legs
              between the two locations the user entered may use behind a
              marking: the marked barriers, and the parts of ways that are
              marked or lie behind a marking, that the warned leg the system
              would produce joining them with no stop between them passes or
              drives, no more times than it does; and on the legs to and from
              a stop, the unmarked parts within a parking area establishing it
              onto which a car turns directly, past no marked barrier, from a
              part that leg drives that is marked or lies behind a marking
              (`DECISIONS.md § RD-080`, `DECISIONS.md § RD-081`,
              `DECISIONS.md § RD-082`).
              The second criterion tests a stop beside the private stretch,
              and the third the detour, a second gate or a stop deeper behind
              the first, and the journey that needs no warned leg at all,
              where nothing behind a marking is ever proposed. FR-138 keeps
              such a stop out of the offer, and this record keeps zones off
              one that reaches the plan regardless; FR-128 has the ways of an
              area marked only on its enclosing feature read as marked here. A
              stop the system offered stays the system's when accepted
              (FR-136), so an acceptance never makes a position a location the
              user entered with a warned leg of its own. A zone the user named
              as a location of the journey (`DESIGN.md § Locations`) is
              entered and not chosen by the system. The walking access path's
              own crossing stays owed with FR-075.
Resolved-by:  #275
```

## FR-138 — Offer no stopping position the system chose behind a private or access-restricted way or barrier

```
Statement:    The system shall offer no stopping position it chose — the
              position of a recommended stop, of a replacement or of a reserve
              candidate — that no warned leg FR-129 admits reaches and that a
              car reaches only by passing a barrier, or by running along a
              way, that the map data marks as private or access-restricted,
              other than a way that lies within a parking area establishing
              the position and carries the marking only as that parking
              area's.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Enforceable exclusions;
              SPECIFICATION.md § What establishes a stopping position
Priority:     MUST
Verification: test — against routing data built from a test extract, a parking
              area not itself marked that a car reaches only past a marked
              gate, one reached only along a marked private road, a position
              on a service road itself marked and reached only along it, and a
              parking area reached only along the unmarked ways of an area
              marked on an area feature enclosing it, each the only stopping
              position of a zone in a journey whose locations the user entered
              a connection avoiding every marking joins, is offered as no
              recommended stop, no replacement and no reserve candidate,
              before a rejection or after one, nor is a parking area behind a
              second marked gate off the private road a warned leg runs along,
              nor one beside a single unsplit way past the location the user
              entered that the warned leg reaches along it, beyond a gate node
              on that way; while a parking area whose own feature is marked
              customers-only, drawn as a point and separately as an area whose
              ways carry no marking of their own, reached along no other
              marked way or barrier, a position a connection avoiding every
              marking reaches, and a parking area beside the part of the
              private road a warned leg FR-129 admits drives, are withheld
              from no offer by this record
Acceptance:   given a stopping position the system chose that a car reaches
              only by passing a barrier so marked, or by running along a way
              so marked that does not lie within a parking area establishing
              the position, and that no warned leg FR-129 admits reaches — a
              parking area not itself so marked behind a gate so marked, a
              position on a service road itself so marked, a parking area
              beyond an area so marked on an area feature enclosing it, one
              behind a second gate off the private road a warned leg runs
              along, or one beside a single unsplit way past the location the
              user entered that the warned leg reaches along it, beyond a gate
              standing on that way — when the journey is planned and when the
              user asks for a different zone, then the position is offered as
              no recommended stop, no replacement and no reserve candidate
              given a stopping position the system chose in a parking area
              whose own feature is so marked — a parking area marked
              customers-only, drawn as a point, or as an area whose ways carry
              no such marking of their own — that a car reaches past no
              barrier so marked and along no way so marked other than the ways
              within that parking area, when the journey is planned and when
              the user asks for a different zone, then this record withholds
              the position from no offer
              given a stopping position the system chose that a drivable
              connection avoiding every such way and barrier reaches, when the
              journey is planned and when the user asks for a different zone,
              then this record withholds the position from no offer
              given a stopping position the system chose that a warned leg
              FR-129 admits reaches — a parking area beside the part of the
              private road that the warned leg between two locations the user
              entered drives — when the journey is planned and when the user
              asks for a different zone, then this record withholds the
              position from no offer
Status:       to-build
Depends-on:   FR-127; FR-129; FR-136
Volatility:   open-question — the markings are FR-125's, read as
              `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066` decide; which
              features establish a position is owed with the OSM-derived
              feature tables
              `SPECIFICATION.md § What establishes a stopping position` names;
              the reading is `DECISIONS.md § RD-074`'s, and that no way a car
              runs along is the position's own feature is the Owner's in
              `DECISIONS.md § RD-075`; that a position a warned leg FR-129
              admits reaches is offered is the Owner's in
              `DECISIONS.md § RD-079`, bounded as FR-129 states and measured
              as `DECISIONS.md § RD-080`, `DECISIONS.md § RD-081` and
              `DECISIONS.md § RD-082` decide
Risk:         Built as a refusal of service alone, a reserve candidate
              behind a private gate is shown to a user who has just rejected a
              zone, accepted, and silently dropped from the plan: the Owner's
              "never offered" is broken in the one place the user sees it, and
              the user loses the zone they took in its place. Built over every
              feature the cited section names as establishing a position, a
              service road among them, a position along a private service road
              is the road's own and offered, and a stop the user places on its
              marker is driven the road's length along FR-126's run, warned by
              nothing. Built over a parking area's own feature and not the
              ways within it, a customers-only car park drawn as an area is
              withheld through its own aisles; built over the parking area as
              a whole, the gated staff car park the cited section cannot tell
              from it is offered through its gate. Built without the warned
              leg's exception, a parking area beside the private stretch the
              user confirmed, which the Owner ruled may serve the zones along
              it, is never shown; built over every position behind the marking
              that leg crosses, the planner offers stops deeper into the gated
              site for the user's confirmation to carry.
Rationale:    The Owner keeps "Stops the app picks itself (zones) behind
              private roads" never offered, since "the app can just pick
              another zone" (`DECISIONS.md § RD-069`). FR-127's first
              criterion and FR-136's refuse to serve such a position, which
              is not refusing to offer it, and
              `SPECIFICATION.md § What establishes a stopping position` gives
              an excluded candidate "no presentation, no reserve, no second
              look" (`DECISIONS.md § RD-074`). The two are layers: this record
              keeps the position out of the review, and FR-127 and FR-136
              keep it out of the plan should it reach the review regardless.
              The same section sends a restriction recorded on the stopping
              feature itself to uncertain and not to exclusion, so that "the
              user may still place the stop there" and a customers-only car
              park stays reachable; a marking the position's own feature
              carries is therefore not one it lies behind, which the second
              criterion tests. The Owner ruled that a road the car drives
              along is never that feature, and only a dead-end feature such as
              a parking area can be (`DECISIONS.md § RD-075`): a way is how a
              car reaches the position, so its marking is one the position
              lies behind even where the cited section names the way as the
              establishing feature, and a barrier is never the position's own.
              The ways within a parking area drawn as an area carry its
              marking as FR-128 reads it, and are exempt only while they carry
              it as the parking area's alone, so the usual mapping of a
              customers-only car park is offered and an aisle marked on its
              own account, or a car park inside a marked estate, is not. Such
              a position stays the system's when accepted (FR-136), and
              FR-140 serves it in no geometry: it is had only where the user
              places the stop there (FR-139), its uncertainty shown (FR-141);
              that its stated reason names the restriction is owed to the
              review batch with the uncertain bucket's other reasons. FR-128
              has the ways of an area marked only on its enclosing feature
              read as marked here. The Owner ruled that the route a location
              the user entered or placed requires may have zones along it, the
              stop for them on the private stretch the user accepted included,
              with no detour onto another private road
              (`DECISIONS.md § RD-079`): a position a warned leg FR-129 admits
              reaches, within the bound that record sets, is offered, which
              the fourth criterion tests, and one behind a marking that leg
              does not cross, or along a part of a way behind one that it does
              not drive, though the map data draws that part and the leg's as
              one way (`DECISIONS.md § RD-080`), is not, which the first
              tests. A position that a connection avoiding every
              marking reaches lies behind nothing and is outside this record,
              and the legs joining it to a location the user entered behind a
              marking are warned legs FR-129 admits. Stated over a stopping
              position and not over a zone, because a zone with a second
              position that nothing so marked separates from the road may be
              offered from that one.
Resolved-by:  #276
```

## FR-140 — Serve no stopping position the system chose on a parking area marked private or access-restricted

```
Statement:    The system shall serve no stopping position it chose in a
              parking area whose own feature the map data marks as private or
              access-restricted, however that parking area is drawn.
Category:     Safety exclusions
Source:       SPECIFICATION.md § What establishes a stopping position
Priority:     MUST
Verification: test — against routing data built from a test extract, a
              stopping position the system offered in a parking area marked
              customers-only on its own feature, drawn as a point beside a way
              not so marked and separately as an area whose ways carry no
              marking of their own, accepted by the user, is served by no
              driving leg, warned or not, in either case; while a stop the
              user placed at the same point is not refused by this record
Acceptance:   given a stopping position the system offered — a recommended
              stop, a replacement or a reserve candidate — in a parking area
              whose own feature is so marked, a parking area marked
              customers-only drawn as a point beside a way not so marked or as
              an area, when the user accepts it and the journey is planned,
              then no driving leg ends at it or short of it, warned or not,
              and the journey includes no stop there
              given a stopping position the user placed at the same point,
              when the journey is planned, then this record refuses neither
              the leg reaching it nor the stop
Status:       to-build
Depends-on:   FR-136; FR-138
Volatility:   open-question — the markings are FR-125's, read as
              `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066` decide; which
              features establish a position, a parking area among them, is
              owed with the OSM-derived feature tables
              `SPECIFICATION.md § What establishes a stopping position` names;
              the one outcome is `DECISIONS.md § RD-078`'s
Risk:         Left to the ways between the position and the road, a
              customers-only car park drawn as a point is served on acceptance
              as the system's while the same car park drawn as an area is
              dropped through its own aisles: the user cannot see which, so
              the same act on the same kind of place keeps the zone one time
              and silently loses it the next. Built to serve both instead, a
              private staff car park with no gate mapped is driven into for a
              stop the system picked, on an acceptance, which is the
              restricted stop the cited section refuses.
Rationale:    `SPECIFICATION.md § What establishes a stopping position` holds
              that where the data records a restriction against the feature
              itself "the feature does not establish a position", sends the
              position to uncertain "so the user may still place the stop
              there and have it recorded as user-chosen", and says this
              paragraph "is what refuses the restricted stop". FR-138 keeps
              such a position in review as uncertain, and FR-136 holds it the
              system's once accepted, but neither refuses it on acceptance:
              FR-127 and FR-128 refuse only a position reached across a
              marking, which the aisles of one drawn as an area are and the
              road beside one drawn as a point is not. This record makes the
              outcome one, and the restrictive one, because the section
              refuses the system's restricted stop and keeps it reachable
              only by the user's placement, and because serving both would
              lift a safety exclusion for a stop the system chose
              (`DECISIONS.md § RD-078`). The stop is had where the user places
              one there (FR-139), which the Owner allows on its marker before
              its restriction reason is stated, its uncertainty shown (FR-141,
              `DECISIONS.md § RD-077`); the second criterion keeps the record
              to the offer, so a build that refuses every stop at such a point
              fails. That the review makes placing, and not accepting, the act
              on such a candidate, so that accepting it is never a silent
              loss, is owed with the interaction #278 drafts for `DESIGN.md`
              and the uncertain bucket's stated reasons. A position the
              system chose behind a marking is never offered (FR-138), and a
              way a car runs along is never a parking area's own feature
              (`DECISIONS.md § RD-075`), so this record reaches only the
              position on its own marking. No mechanism is named.
Resolved-by:  #276
```
