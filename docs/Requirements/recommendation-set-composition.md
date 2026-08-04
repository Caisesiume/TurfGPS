# Recommendation set composition

Which journey alternatives reach the user and what the offered set must hold: how many, how varied, what is dropped as redundant or dominated, and what may never be withheld. Distinct from `Route alternative generation`, which produces the corridors, and from `Objective selection and ranking`, which decides the objective and the order — this category decides membership of the offered set, not its ranking. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `.claude/skills/requirements-authoring/SKILL.md`.

## FR-013 — Offer more than one journey alternative

```
Statement:    The system should offer more than one journey alternative for a
              planned journey rather than a single recommendation.
Category:     Recommendation set composition
Source:       SPECIFICATION.md § Recommended journey alternatives
Priority:     MUST
Verification: test — a journey for which more than one produced alternative
              survives FR-015 and FR-016 yields more than one alternative
              offered to the user; a journey whose produced alternatives
              those records reduce to one has nothing added to the offered
              set to reach two
Acceptance:   given a journey for which more than one alternative differing
              in its general route or in its captured zone set has been
              produced, and more than one of them survives the removals
              FR-015 and FR-016 require, when recommendations are offered,
              then more than one alternative is offered
              given a journey whose produced alternatives those removals
              reduce to a single alternative, when recommendations are
              offered, then no removed alternative is returned to the set to
              reach two
Status:       to-build
Depends-on:   FR-001
Risk:         A single answer is the product the user already has. The choice
              between a faster route and a richer one is the decision this
              tool exists to support, and a set of one conceals that the
              choice was made at all — by the optimizer, on the user's behalf
              and unexplained.
Rationale:    The source's word is *several*; the statement says *more than
              one* because *several* is not a count two engineers read alike,
              and because the source's own contrast is with a single
              supposedly perfect answer. No size is fixed beyond that floor: a
              result-set size would be a constant, and it has no home in
              `CalculationSpecification.md`. The verb follows the source and
              carries the case where the search yields only one alternative
              worth offering. That some route is always offered at all is a
              claim this record does not make: it belongs to
              `SPECIFICATION.md § User time constraints`, which declares
              itself the single definition of the allowance, and is not
              restated here. This record does not bound curation from below
              and must not be read as doing so. FR-014 cannot reduce the set
              below the floor: what it removes are alternatives agreeing on
              both properties the first criterion above requires them to
              differ in, so where it fires the floor was never engaged. FR-015
              and FR-016 can, and where either cannot hold together with this
              record, that record governs and the set is offered short rather
              than filled out. Padding a set to reach a count is fabricated
              variety — the same dishonesty as a fabricated metric, and the
              user is the one who pays for it. This is the justified deviation
              the statement's *should* carries, and it is the only one.
Resolved-by:  #12
```

## FR-014 — Do not offer two indistinguishable journey alternatives

```
Statement:    The system should not offer two journey alternatives that use
              the same general route and whose captured zone sets compare
              equal.
Category:     Recommendation set composition
Source:       SPECIFICATION.md § Recommended journey alternatives
Priority:     MUST
Verification: test — two produced alternatives using the same general route
              and capturing the same zone set yield one offered alternative
              rather than two
Acceptance:   given two produced journey alternatives that use the same
              general route and whose captured zone sets compare equal, when
              recommendations are offered, then at most one of the two is
              offered
Status:       to-build
Depends-on:   FR-013
Risk:         Two alternatives a user cannot tell apart spend the offered
              set's scarcest resource — the user's attention — on a choice
              that is not one, and whatever they displaced is never seen.
Rationale:    This is the identity floor of the source's *effectively
              duplicates*, not a definition of it. The source's word is
              *effectively*, and the similarity threshold that would decide
              when two different alternatives are effectively the same has no
              home in `CalculationSpecification.md`; introducing one here
              would create it, which is the reasoning FR-012 records for
              proximity to an origin. What such a threshold would have covered
              is FR-016's, judged rather than measured, and this record does
              not discharge it. The two properties compared are the ones the
              documents already use to tell alternatives apart:
              `SPECIFICATION.md § General route alternatives` separates
              general route selection from zone-level detours, so two
              alternatives agreeing on both are one alternative found twice.
Resolved-by:  #13
```

## FR-015 — Do not offer an alternative beaten outright by another

```
Statement:    The system should not offer a journey alternative for which
              another alternative it offers carries at least the same Turf
              value at no greater additional journey cost, and is better on
              one of the two measures.
Category:     Recommendation set composition
Source:       SPECIFICATION.md § Recommended journey alternatives
Priority:     SHOULD
Verification: test — of a constructed pair in which one alternative costs more
              additional journey time and carries no more Turf value than the
              other, only the other is offered; of a pair in which the more
              costly alternative also carries the greater value, both are
              offered
Acceptance:   given two produced journey alternatives, one carrying no more
              Turf value than the other at a greater additional journey cost,
              when recommendations are offered, then the more costly
              alternative is not among those offered
              given two produced journey alternatives, the more costly of
              which carries the greater Turf value, when recommendations are
              offered, then the more costly alternative is not withheld for
              being the more costly
Status:       to-build
Depends-on:   FR-007; FR-008; FR-013
Rationale:    The source's words are *do not provide a meaningful trade-off*.
              An alternative beaten on both measures asks the user to give up
              something for nothing, and it is the one case needing no
              threshold: the relation is an ordering on each measure, not a
              tolerance between them, so nothing here creates the constant
              FR-014's Rationale records as missing. Turf value is compared
              under one user's preferences, per FR-011; the comparison is not
              user-independent and this record does not make it so. It is not
              the strict ordering by value against cost that FR-010's
              Rationale declined: a stretch alternative costs more and carries
              more value, so it is never beaten outright, and the deliberately
              varied set survives this record intact. The second criterion is
              what states that, and it is there so the record cannot be built
              as the ordering FR-010 refused.
Resolved-by:  #14
```

## FR-016 — Offer alternatives that differ in ways a user would act on

```
Statement:    The system should offer a set of journey alternatives in which
              each alternative gives the user a distinct reason to choose it
              over the others.
Category:     Recommendation set composition
Source:       SPECIFICATION.md § Recommended journey alternatives
Priority:     MUST
Verification: human-judgement — the Owner, against
              `SPECIFICATION.md § Recommended journey alternatives`, over the
              alternative sets returned for real journeys: whether the set is
              several meaningfully different alternatives rather than
              variations of one
Acceptance:   the Owner reviews the offered alternative set for each of a
              sample of real journeys and judges, against
              `SPECIFICATION.md § Recommended journey alternatives`, whether a
              user choosing between them would have a reason to prefer each
              one; a set whose alternatives the Owner cannot separate on
              grounds a user would act on fails
Status:       to-build
Depends-on:   FR-013
Risk:         A set whose alternatives all say the same thing returns the user
              to the manual comparison this product exists to replace, and it
              does so invisibly: every countable signal — how many
              alternatives were offered, how many of the named kinds they
              covered — reports success while the choice on the screen is not
              a choice.
Rationale:    FR-014 and FR-015 are the machine-checkable floor of this
              obligation, not the obligation itself: two alternatives
              differing by one ordinary zone and a minute are neither
              identical nor beaten outright, and they are exactly what the
              source's *effectively duplicates* describes. Closing that gap
              with a number would need a similarity threshold that has no
              home, and a fabricated one would be measured, would pass, and
              would leave the quality it stood for unexamined. So the residue
              is judged: `DELIVERY.md § Escalation and human judgement` names
              whether a route recommendation is genuinely good as the
              product's real bar and as not machine-checkable, and which
              alternatives reach a user is that bar's clearest instance. The
              four kinds named in the cited section are the judge's reference
              for what *meaningfully different* means — they are offered
              permissively there and create no obligation to produce any of
              them, which is FR-017's branch. This record is functional rather
              than a quality attribute because it constrains which produced
              alternatives reach the user, an observable output, not how well
              any one of them performs.
Resolved-by:  #17
```

## FR-017 — Do not withhold alternatives for a missing kind of alternative

```
Statement:    The system shall not withhold a journey alternative it has
              produced for the reason that the alternatives produced do not
              include every kind of alternative named under
              *Recommended journey alternatives* in `SPECIFICATION.md`.
Category:     Recommendation set composition
Source:       SPECIFICATION.md § Recommended journey alternatives
Priority:     MUST
Verification: test — a journey for which alternatives of only one of the named
              kinds are produced still yields recommendations, with no
              produced alternative dropped and the offered set not emptied for
              the other kinds being absent
Acceptance:   given a journey for which alternatives of only some of the kinds
              named in `SPECIFICATION.md § Recommended journey alternatives`
              have been produced, when recommendations are offered, then no
              produced alternative is withheld for the absence of a kind that
              was not produced
              given a journey for which every produced alternative is of one
              and the same kind named in
              `SPECIFICATION.md § Recommended journey alternatives`, when
              recommendations are offered, then the offered set is not emptied
              for the other kinds being absent
Status:       to-build
Depends-on:   FR-013
Risk:         Read as a required set, the four named kinds turn an ordinary
              result into a failed search: a corridor with no attributed zone,
              or a budget with no room for a stretch, is the common case
              outside a city, and the user is told nothing was found on a
              journey that has several usable routes.
Rationale:    The cited section offers the four kinds permissively — a typical
              result set *may* contain them — and then states outright that
              not every search must return all four. The obligation is
              therefore the negative one: the set is not conditioned on the
              kinds. The statement forbids one ground for withholding and
              obliges nothing else about the offered set, deliberately: an
              obligation to offer what was produced would contradict FR-014,
              FR-015 and FR-016, each of which removes alternatives on grounds
              this record does not touch. What each kind is remains unauthored
              here, also deliberately. Three of the four are defined in terms
              the cited section does not own — a budget used comfortably, more
              of, or most of — whose home is
              `SPECIFICATION.md § User time constraints`, and the fourth in
              terms of highly ranked attributes, whose home is
              `SPECIFICATION.md § Attribute preference`.
Resolved-by:  #15
```

## FR-018 — Do not withhold a compliant alternative for its lower Turf value

```
Statement:    Where no zone carrying an attribute is reachable within the
              user's stated additional-time limit, per
              *User time constraints* in `SPECIFICATION.md`, the system should
              not withhold a journey alternative that lies within that limit
              for carrying less Turf value than an alternative that exceeds
              it.
Category:     Recommendation set composition
Source:       SPECIFICATION.md § Recommended journey alternatives
Priority:     MUST
Verification: test — a corridor whose zones reachable within the stated limit
              all carry no attribute yields the produced within-limit
              alternative among those offered, and a corridor whose only
              attributed zone lies beyond that limit does not drop its
              within-limit alternative in favour of the one reaching that
              zone
Acceptance:   given a journey for which no zone carrying an attribute is
              reachable within the user's stated additional-time limit and a
              within-limit alternative has been produced, when recommendations
              are offered, then that alternative is not withheld for carrying
              less Turf value than an alternative exceeding the limit
              given a journey whose only zone carrying an attribute is
              reachable only by exceeding that limit, when recommendations are
              offered, then the produced within-limit alternative is not
              excluded in favour of the alternative that exceeds it
Status:       to-build
Depends-on:   FR-007; FR-008; FR-013
Risk:         The zone-poor corridor is the ordinary case away from cities. A
              system reading no attributed zone as no result gives the user an
              empty screen on the journey where one ordinary zone was still
              worth the stop, and leaves an over-limit alternative as the only
              thing it has to show — which the absolute ceiling may then have
              to refuse as well.
Rationale:    The source's *best compliant route available* is the ordinary
              ranking applied unchanged, not a second ordering rule: what
              makes one alternative better than another is owned by
              *Optimization objectives* in `SPECIFICATION.md` and by
              *The objective function* in `CalculationSpecification.md`, so
              this record names neither and cites the limit rather than
              restating it. `SPECIFICATION.md § User time constraints`
              declares itself the single definition of the allowance, so the
              limit is named and never quantified — and that at least one
              alternative within it always exists is that section's
              obligation, deferred to the batch scoped to it rather than
              restated here. What this section uniquely creates is the branch
              stated above: an alternative already produced within the limit
              is not dropped for being the poorer in Turf terms, which is the
              one ground the source's *fewer or only ordinary zones* describes
              twice. The prohibition is on that ground alone, so it neither
              forbids ranking the within-limit alternative below a stretch nor
              collides with FR-015 — an alternative exceeding the limit costs
              more, so it never beats a within-limit one outright. The verb
              follows the source's own *should present*; reading `shall` off
              the *must* in the sentence before it would restate the existence
              obligation this record defers.
Resolved-by:  #16
```
