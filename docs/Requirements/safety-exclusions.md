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
Status:       draft
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
Status:       draft
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
              only where the road position and the zone are established to be
              at compatible levels with no intervening barrier between them.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Direct road-access validation
Priority:     MUST
Verification: test — a zone above a road that passes beneath it, a zone
              beneath a bridge carrying the road, and a zone separated from
              the road by a retaining wall are each not classified directly
              road-accessible, while a zone established to be at a compatible
              level with nothing between is not refused by this check
Acceptance:   given a candidate whose bridge, tunnel or layer data indicates
              that the road position and the zone do not meet, when its access
              is classified, then it is not classified directly
              road-accessible
              given a candidate for which the data records a barrier between
              the road position and the zone, when its access is classified,
              then it is not classified directly road-accessible
              given a candidate established to be at a level compatible with
              the road position with no barrier recorded between them, when
              its access is classified, then this check does not refuse it the
              direct class
Status:       draft
Depends-on:   FR-068
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
              sourced where that content lives. It is not FR-071's: that
              record governs where a candidate goes when this check fails, and
              a system that never runs the check fails this record while never
              firing that one. The third criterion states this record's own
              bound — the remaining conditions on the direct class are the
              road-class and speed-limit exclusions under
              `SPECIFICATION.md § Enforceable exclusions`, which FR-089 gates
              and this record does not — so it cannot be read as the whole
              test for the class.
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
Verification: test — a candidate for which no connected walking route can be
              identified is excluded rather than priced by any estimate, one
              whose path carries a section steeper than the configured
              threshold is excluded whatever the zone carries, and one so
              excluded stays excluded when the user's maximum acceptable path
              gradient is set less strict than that threshold
Acceptance:   given a candidate for which no connected walking route between a
              stopping position and the zone's coordinate can be identified,
              when its access is classified, then it is excluded and is not
              priced by a straight-line estimate
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
Status:       draft
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
              path disconnected. The fourth criterion is the branch this
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
Status:       draft
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

## FR-089 — Gate the direct class on the road-class and speed-limit exclusions

```
Statement:    The system shall classify a zone as directly road-accessible
              only where the road carrying its stopping position is one the
              road-class and speed-limit exclusions under
              `SPECIFICATION.md § Enforceable exclusions` admit as a stopping
              place.
Category:     Safety exclusions
Source:       SPECIFICATION.md § Directly road-accessible zones
Priority:     MUST
Verification: test — a candidate whose stopping position is on a motorway or a
              motorway link is not classified directly road-accessible, nor is
              one on a road the speed-limit exclusion does not admit, while
              one on an admitted road is not refused the class by this check
Acceptance:   given a candidate whose stopping position is on a motorway or a
              motorway link, when its access is classified, then it is not
              classified directly road-accessible
              given a candidate whose stopping position is on a road the
              speed-limit exclusion under
              `SPECIFICATION.md § Enforceable exclusions` does not admit as a
              stopping place, when its access is classified, then it is not
              classified directly road-accessible
              given a candidate whose stopping position is on a road those
              exclusions admit, when its access is classified, then this check
              does not refuse it the direct class
Status:       draft
Depends-on:   FR-068; FR-092
Volatility:   proposed-constant — the maximum speed limit for a stopping road
              under
              `CalculationSpecification.md § The maximum speed limit for a stopping road`
              is a proposed default with no origin, so which roads this gate
              refuses moves with it even though the gate itself does not
Risk:         This class tells a driver to take the zone without leaving the
              car, and this gate is the only thing on the direct path that
              asks what road they are being told to stop on. Without it a zone
              six metres from a drivable edge, at a compatible level with
              nothing between, reaches the confident class on a derestricted
              autobahn — the outcome
              `SPECIFICATION.md § An unknown speed limit fails the check`
              names as the single worst one the exclusion exists to prevent —
              and every check the classifier ran reports a pass, because the
              check that would have failed was never written.
Rationale:    The cited section states in one sentence that two further
              conditions apply to the direct class, each defined in full
              elsewhere and not restated there: the level-compatibility
              requirement, which FR-070 carries, and the road-class and
              speed-limit exclusions, which this record carries. Both are
              gates and neither is the machinery. What establishes a speed
              limit, what a road carrying more than one does, and that no
              lookup table may resolve a jurisdictional value are stated under
              `SPECIFICATION.md § Enforceable exclusions` and are owed to the
              batch scoped there, so none of it is restated in a criterion
              here and the criteria refer the decision to that section rather
              than making it. This record exists because FR-067 obliges every
              classified candidate into exactly one of four classes and
              FR-070's third criterion affirmatively declines to refuse the
              direct class on any other ground, which completed a path to the
              confident class around the gap this fills. The exclusion in the
              cited section is broader than this gate — no stop may be
              proposed on such a road on either branch — and the park-and-walk
              half is not authored here. `Architecture.md § D3` records that
              the routing engine returns road class, speed limit and
              drivability as edge attributes, which is what this gate reads;
              the constant those attributes are compared against has one home
              and is never quoted here, and that it may not be configured
              upward is FR-090's.
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
Status:       draft
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

## FR-091 — Plan no journey while the implausible-gradient threshold is unconfigured

```
Statement:    The system shall refuse to plan a journey while no value is
              configured for the implausible-gradient threshold recorded as
              owed under
              `CalculationSpecification.md § Enforcement constants that do not yet exist`.
Category:     Safety exclusions
Source:       CalculationSpecification.md § Enforcement constants that do not yet exist
Priority:     MUST
Verification: test — a configuration carrying no value for the
              implausible-gradient threshold is refused and no journey is
              planned under it, while one carrying a value is accepted and no
              journey is refused on this ground
Acceptance:   given a configuration carrying no value for the
              implausible-gradient threshold, when the system starts, then the
              configuration is refused and no journey is planned under it
              given a configuration carrying a value for that threshold, when
              the system starts, then the configuration is accepted and no
              journey is refused on this ground
Status:       draft
Depends-on:   FR-076
Volatility:   open-question — the threshold has no value in any document and
              is recorded as owed by the cited section, which also records
              that it may need to be per elevation provider, so the shape of
              the constant and not only its value is unsettled
Risk:         Unset, the comparison this threshold feeds is false for every
              candidate: the steep limb of FR-076 never fires, and a
              quarry-rim candidate carrying a fragment of mapped path reaches
              park-and-walk while every test written for FR-076 still passes
              on its absent and disconnected limbs. That is the shape a
              missing safety constant takes — not a check that fails but a
              check that silently never runs, on the one exclusion whose
              failure is a fall rather than a bad estimate. Every other
              enforcement constant in the set has a record refusing a loosened
              value; this is the one with no value at all, and
              `CalculationSpecification.md § Conventions` forbids the
              alternative an implementer reaches for, a literal authored on
              the spot with no origin.
Rationale:    The cited section records this constant as required for the
              system to behave as specified, states that it has no value
              anywhere, and gives the reason for recording it: the first
              implementer to need one cannot author it silently, as the
              unexplained literal `CalculationSpecification.md § Conventions`
              forbids. This record is that prohibition made enforceable, and
              it is the only shape available — the threshold cannot be
              defaulted here, because a default is the literal that section
              forbids and this lane may not pick one, and it cannot be left
              out, because omission is the dead limb FR-076's own `Risk`
              names. Refusal at start rather than a quiet downgrade of every
              park-and-walk candidate follows the reasoning FR-046 records for
              a clamp: the quiet path runs a deployment its operator did not
              configure and hides the misconfiguration behind plausible
              behaviour, and here it would hide it behind a coverage loss
              indistinguishable from thin map data. That this record blocks
              the system until the constant is authored is its intent rather
              than a side effect; authoring the value is owed to the batch
              that owns it, and this record states what may not happen
              meanwhile. Neither the value nor its shape is quoted, since
              neither exists.
Resolved-by:  —
```
