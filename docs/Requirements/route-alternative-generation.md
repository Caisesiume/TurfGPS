# Route alternative generation

Producing the alternatives a search yields: the candidate corridors a journey may take, and whether alternatives are produced at all and in what number. Distinct from `Recommendation set composition`, which decides which of them reach the user — this category owns the search's output, that one owns the offered set. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-042 — Search for alternatives whatever the size of the stated limit

```
Statement:    The system shall not decline to produce journey alternatives by
              reason of the size of the additional-time limit the user stated.
Category:     Route alternative generation
Source:       SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — a journey planned with a limit at the low end of the
              realistic range, along a corridor holding an admissible
              candidate zone whose capture fits within that limit, produces an
              alternative capturing such a zone rather than nothing
Acceptance:   given a limit at the low end of the realistic range described
              under `SPECIFICATION.md § User time constraints` and a corridor
              holding at least one admissible candidate zone whose capture
              would keep an alternative within that limit, when planning runs,
              then at least one produced alternative captures such a zone
              given a limit at the low end of that range, when the user starts
              planning, then planning runs and is not refused for the limit's
              size
Status:       to-build
Depends-on:   FR-038; FR-040
Risk:         The user with the smallest budget is the one this product
              converts from a curiosity into a habit, and an empty screen is
              indistinguishable to them from a corridor with no zones in it.
              The failure needs no minimum-budget gate to occur: a pipeline
              tuned and tested at generous budgets can return nothing at a
              short one and look correct doing it.
Rationale:    The section requires the optimizer to behave sensibly across the
              whole range, and the first criterion is that requirement's
              testable floor at the short end — a capture that fits is found —
              which needs no threshold because it is stated against what the
              corridor actually holds. The second forecloses the gate. FR-040
              is the interface half of the same sentence and this is the
              optimizer half, which is the split the section itself draws.
              Nothing is authored here for the long end of the range: what a
              large budget must not degenerate into is owned by
              `SPECIFICATION.md § Individual zones rather than local collection routes`
              and restating it here would give that prohibition a second home.
Resolved-by:  #49
```

## FR-058 — Generate more than one general road route

```
Statement:    The system should generate more than one general road route
              between the journey's required locations where more than one
              viable road route exists.
Category:     Route alternative generation
Source:       SPECIFICATION.md § General route alternatives
Priority:     MUST
Verification: test — a journey whose required locations are connected by more
              than one viable road route yields more than one general road
              route, and no two of the routes produced traverse the same roads
Acceptance:   given a journey whose required locations are connected by more
              than one viable road route, when general road routes are
              generated, then more than one general road route is produced
              given the general road routes produced for such a journey, when
              they are compared, then no two of them traverse the same roads
              between those locations
Status:       to-build
Depends-on:   FR-001
Volatility:   settled — the obligation is a count and a non-identity, and
              rests on no figure and on nothing any document lists as open
Risk:         Every Turf-enhanced alternative is built on a general road
              route, so a search that finds one can only ever offer variations
              of a single journey. The user then chooses between detours off a
              road nobody asked them to confirm, and nothing in the result
              shows that the road itself was decided for them.
Rationale:    The section's word is multiple; the statement says more than one
              because a count two engineers read alike is the whole of what
              this record can state without a threshold. The second criterion
              is the identity floor of the section's meaningfully different,
              on FR-014's precedent — two routes traversing the same roads are
              one route found twice — and the judged residue above that floor
              is FR-059's. The verb follows the source. This is not the
              obligation `README.md` records as owed on producing more than
              one journey alternative, whose home is this same category: that
              debt is about the search's output once zones are combined, and
              one general road route can still yield several journey
              alternatives. Whether this record narrows it is returned with
              this batch rather than decided here.
Resolved-by:  —
```

## FR-059 — Produce general road routes that are different drives

```
Statement:    The general road routes the system produces for a journey should
              each represent a different way of making the drive rather than a
              variation of another.
Category:     Route alternative generation
Source:       SPECIFICATION.md § General route alternatives
Priority:     MUST
Verification: human-judgement — the Owner, against
              `SPECIFICATION.md § General route alternatives`, over the
              general road routes produced for real journeys: whether each is
              a way of making the drive that a driver would recognise as
              different from the others
Acceptance:   the Owner reviews the general road routes produced for each of a
              sample of real journeys and judges, against
              `SPECIFICATION.md § General route alternatives`, whether each
              route is a different way of making the drive rather than a
              variation of another; a set whose routes the Owner can separate
              only by a short local difference fails
Status:       to-build
Depends-on:   FR-058
Volatility:   settled — nothing this record rests on is proposed or open, and
              the quantity it lacks is a distinctness bar, whose absence is
              why the method is judged rather than something that will move
Risk:         A search fed near-identical general road routes spends its
              candidate budget twice on one corridor, and the waste is
              invisible downstream: FR-016 removes the effective duplicate
              from the offered set, so the user meets a short set rather than
              a wasted search, and nothing distinguishes that from a corridor
              with nothing in it.
Rationale:    FR-058's second criterion is the machine-checkable floor of this
              obligation and not the obligation itself: two routes differing
              at one junction traverse different roads and are still the same
              drive, which is exactly what the section's not merely a
              variation describes. The look for a bar found none —
              `CalculationSpecification.md` sets no distinctness measure
              between routes anywhere, and `README.md` already records that
              neither `Architecture.md` nor `DESIGN.md` holds one — and a
              similarity figure chosen here would be measured, would pass, and
              would leave the quality it stood for unexamined. So the residue
              is judged, on FR-016's precedent. This record reaches a layer
              FR-016 does not: that record judges the alternatives offered to
              the user, and by the time it applies the search has already
              spent its budget on whatever general road routes were produced.
Resolved-by:  —
```

## FR-060 — Build on more than the fastest conventional route

```
Statement:    The system should not treat the fastest conventional route as
              the only general road route on which Turf-enhanced journey
              alternatives are built.
Category:     Route alternative generation
Source:       SPECIFICATION.md § General route alternatives
Priority:     SHOULD
Verification: test — a journey for which more than one general road route was
              produced yields at least one journey alternative built on a
              general road route other than the fastest conventional one, and
              the fastest conventional route is still among the routes
              alternatives are built on
Acceptance:   given a journey for which more than one general road route has
              been produced, when journey alternatives are produced, then at
              least one of them is built on a general road route other than
              the fastest conventional one
              given the same journey, when journey alternatives are produced,
              then the fastest conventional route is among the general road
              routes they are built on
Status:       to-build
Depends-on:   FR-058
Volatility:   settled — the record states a residue that survives whatever
              scoring model is one day authored for general road routes, and
              rests on no figure
Rationale:    The section asks the optimizer to score the usability of each
              general road route on ordinary travel characteristics and Turf
              potential, and no such model exists:
              `CalculationSpecification.md § The objective function` scores
              candidate zones rather than roads, and inventing a
              route-usability score here would create a model with no home.
              What survives without one is the section's own conclusion — the
              fastest route stays an important reference and must not
              automatically be treated as the only valid foundation — and this
              record states that and no more. The second criterion carries the
              important reference half, so the record cannot be built as a
              rule that skips the fastest route. Risk is omitted and its
              argument folded here, on FR-025's precedent: the failure costs
              result variety and puts no driver anywhere.
Resolved-by:  —
```
