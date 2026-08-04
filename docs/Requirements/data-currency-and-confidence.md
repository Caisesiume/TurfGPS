# Data currency and confidence

How the age of the data a recommendation was built from bounds the confidence the system may record for it. Its trigger is the **age** of data that is otherwise complete, which is what separates it from `Coverage and data quality`, whose trigger is data that is **thin**; and its subject is the confidence carried by the recommendation rather than the calibration of the time estimate inside it, which is what separates it from `Estimate accuracy`. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `.claude/skills/requirements-authoring/SKILL.md`.

## NFR-002 — Lower confidence as the Turf data behind a recommendation ages

```
Statement:    The system should record lower confidence for a recommendation
              as the Turf data it was built from ages.
Category:     Data currency and confidence
Source:       Architecture.md § Volatile and optional fields
Priority:     SHOULD
Verification: test — recommendations computed from Turf snapshots that differ
              only in age carry confidence that never rises with age, and that
              is strictly lower for the older snapshot at least once across
              the ages sampled
Acceptance:   metric — the confidence the system records for a recommendation,
              sampled over a series of Turf snapshots of increasing age;
              threshold — that confidence never rises as snapshot age rises,
              and is strictly lower for at least one pair in the series;
              condition — the samples differ only in the age of the snapshot
              behind them, with the same zones, the same route and every other
              uncertainty driver held fixed, and the ages sampled are read
              from the deployment's own confidence configuration rather than
              fixed by this record
Status:       to-build
Depends-on:   none
Rationale:    The source states a relation and not its shape, so this record
              tests the relation and fixes no age at which confidence changes.
              An age boundary would be a constant, and it has no home:
              `CalculationSpecification.md` carries no confidence scale, no
              age bucket and no data-age term, and
              `SPECIFICATION.md § Confidence and uncertainty` enumerates the
              drivers of confidence without naming data age among them. The
              access confidence carried under
              `CalculationSpecification.md § Proposed form: value per minute`
              is a gate on eligibility rather than a scale this record could
              have moved. A criterion written against a boundary would have
              created the missing constant on the first implementer's desk,
              and would have chosen the implementation besides: a confidence
              decaying continuously with age satisfies the statement and
              straddles no boundary at all. An ordering is what makes this
              testable without a number — the source compares a more confident
              recommendation with a less confident one, which needs the two to
              be comparable, not to sit on a scale — so the method is test
              rather than human-judgement even though the look found no
              threshold; the judged method is for a residue no ordering
              catches. That residue is real and is deliberately not carried
              here: whether the fall this record obliges is large enough to be
              honest to the user reading the recommendation, and whether that
              section's drivers are the right standard for data age at all,
              are owed to the batch scoped to
              `SPECIFICATION.md § Confidence and uncertainty`.
Resolved-by:  #36
```
