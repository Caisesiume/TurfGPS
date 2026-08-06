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
