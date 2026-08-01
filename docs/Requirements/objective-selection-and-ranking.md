# Objective selection and ranking

The objective a plan optimizes, and the ordering of alternatives under it. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `.claude/skills/requirements-authoring/SKILL.md`.

## FR-010 — Do not prefer a journey alternative for its zone count

```
Statement:    The system should not prefer one journey alternative to another
              on the number of zones it captures, where the other
              alternative's Turf value set against its additional cost is the
              greater.
Category:     Objective selection and ranking
Source:       SPECIFICATION.md § The journey as an optimization problem
Priority:     MUST
Verification: test — a constructed pair in which the alternative capturing
              fewer but more highly valued zones has the better value against
              cost, and the alternative capturing more zones is not offered in
              preference to it
Acceptance:   given two journey alternatives, one capturing more zones than
              the other, and the other's Turf value set against its additional
              cost being the greater, when the system offers its
              recommendations, then the alternative capturing more zones is
              not offered in preference to the other
Status:       to-build
Depends-on:   FR-007; FR-008
Risk:         Count maximization is the failure mode this product exists to
              avoid: it produces exactly the urban zone-collection route the
              scope excludes and buries the single attribute zone the planning
              user made the journey for, while every internal metric reports
              success.
Rationale:    The statement forbids preferring an alternative *for its count*,
              not counting zones at all — the Zones objective under
              *Optimization objectives* in `SPECIFICATION.md` legitimately
              values zone count, and there the extra zones carry the value
              this comparison already weighs. The source's word is "blindly".
              The statement verb is `should`, following the source's "should
              not blindly maximize"; the priority is separate and high. The
              source's converse — a valuable zone that is still a poor
              recommendation because reaching it costs too much — is
              deliberately not authored here: stated generally it would oblige
              a strict ordering by value against cost, which the deliberately
              varied alternatives under *Recommended journey alternatives* in
              `SPECIFICATION.md` do not follow, and the bound on what costs
              too much is owned by *User time constraints* in the same
              document.
Resolved-by:  —
```

## FR-011 — Balance value against cost from the individual user's preferences

```
Statement:    The system shall balance Turf value against additional journey
              cost according to the preferences stated by the individual user
              rather than a balance fixed across all users.
Category:     Objective selection and ranking
Source:       SPECIFICATION.md § The journey as an optimization problem
Priority:     MUST
Verification: test — the same journey planned for two users whose stated
              preferences over an attribute available along the route are
              opposed yields recommendations that are not identical, and
              changing one user's preferences changes their recommendations
Acceptance:   given the same journey requested by two users, one ranking an
              attribute carried by a zone reachable from the route above all
              others and the other ranking it below all others, when
              recommendations are produced for each, then the two sets of
              recommendations are not identical
              given a user whose stated preferences are changed and whose
              journey is then planned again unaltered, when recommendations
              are produced, then the recommendations change
Status:       to-build
Depends-on:   FR-007; FR-008
Risk:         A single global trade-off makes the tool useless to precisely
              the user it is built for — the attribute hunter for whom one
              Monument zone justifies a long detour — because their journey is
              scored by someone else's idea of a good route, and nothing in
              the output reveals that the balance was never theirs.
Rationale:    The source's "must" is the strongest obligation in the section.
              What the user may state, and how a ranking becomes a weight, are
              owned by *Attribute preference* in `SPECIFICATION.md` and
              *Proposed rank-to-weight curve* in
              `CalculationSpecification.md`; this record obliges the optimizer
              to consume those preferences, not to collect them. The criteria
              assert non-identity rather than which alternative wins: that
              mapping is a proposal, not a settled value, and a criterion
              naming the winner would harden it into an obligation.
Resolved-by:  —
```
