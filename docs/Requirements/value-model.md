# Value model

What a journey's Turf value is made of: the zones counted toward it, the attributes that weight them, and the user's stated preferences over those attributes. The mirror of `Cost and time composition`; the two are combined by `Objective selection and ranking`, which owns neither. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-007 — Derive a journey's Turf value from the zones it captures

```
Statement:    The system shall derive a journey alternative's Turf value from
              the zones that alternative captures rather than from the route
              it travels.
Category:     Value model
Source:       SPECIFICATION.md § The journey as an optimization problem
Priority:     MUST
Verification: test — two alternatives capturing the same zones over different
              roads receive equal Turf value, and lengthening an alternative's
              driving time without changing its captured zones leaves its Turf
              value unchanged
Acceptance:   given two journey alternatives that capture the same set of
              zones while travelling different roads, when the Turf value of
              each is computed, then the two values compare equal
              given a journey alternative whose driving time changes while its
              captured zone set does not, when its Turf value is recomputed,
              then the value is unchanged
Status:       to-build
Depends-on:   FR-001
Risk:         Once a property of the route leaks into value, the comparison of
              value against cost stops meaning anything — the same minutes are
              counted as a benefit on one side and a penalty on the other —
              and no recommendation built on it can be explained to the user,
              which is the product's stated purpose.
Rationale:    The source determines value by the zones a journey holds and
              cost by the time it adds, so the route belongs to cost alone;
              *Proposed form: value per minute* in `CalculationSpecification.md`
              corroborates the same separation from the other side, keeping
              difficulty on the cost side so it never reduces value. The
              statement does not say "solely from the zones", because the
              weight a captured zone carries comes from the user's stated
              preferences as well, per FR-011. This record took "captures"
              while the source still read "the zones it contains" in one
              sentence and "the zones it captures" in the next, since a
              journey passing a zone it does not stop at gains nothing;
              `SPECIFICATION.md § The journey as an optimization problem` now
              reads "captures" in both, corrected to that reading on 1 August
              2026 with the Owner's approval.
Resolved-by:  #9
```
