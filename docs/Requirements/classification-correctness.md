# Classification correctness

The product's stated measure of success: no zone classified confidently and wrongly. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## NFR-007 — Classify no zone confidently and wrongly

```
Statement:    The system shall place no zone in a confident access class —
              directly road-accessible or park-and-walk — whose access does
              not hold on the ground.
Category:     Classification correctness
Source:       SPECIFICATION.md § Accessibility scope for the first release;
              SPECIFICATION.md § Accessibility principle
Priority:     MUST
Verification: human-judgement — the Owner, against the measure of success
              under
              `SPECIFICATION.md § Accessibility scope for the first release`,
              over a sample of confidently classified zones checked against
              what is actually there
Acceptance:   the Owner examines a sample of zones the system classified as
              directly road-accessible or as park-and-walk, drawn from real
              journeys across urban and rural corridors alike, against what is
              actually there — the stopping position, the way between it and
              the zone coordinate, and any barrier between the two — and
              judges, against the measure of success under
              `SPECIFICATION.md § Accessibility scope for the first release`,
              whether any of them was classified confidently and wrongly. One
              confident classification the Owner finds wrong fails the
              criterion; a zone the system declined to classify, or classified
              uncertain, does not, however many of those the sample holds.
Status:       draft
Depends-on:   FR-067; FR-085
Volatility:   settled — the measure of success under
              `SPECIFICATION.md § Accessibility scope for the first release`
              is stated flatly, is marked no proposal, appears on no
              open-questions list, and this record names no constant
Risk:         This is the product's stated measure of success, and the failure
              it names is the one the data cannot catch. A zone classified
              confidently and wrongly is handed to a driver as a stop they can
              make: they leave the road for a position that is not one, or
              step out beside a carriageway onto ground no part of the system
              validated, because the branch that would have validated it was
              never entered — the hazard
              `CalculationSpecification.md § Direct-access tolerance` states
              for its own constant. Every countable signal reports success
              while this fails. Coverage rises, the uncertain bucket shrinks,
              and each classification carries evidence that looked sufficient
              at the moment it was made.
Rationale:    The look was run before the method was chosen and found nothing
              to write against: `CalculationSpecification.md` holds no
              misclassification rate, no accuracy bar and no sampling rule for
              access classification, and nothing there yields one — the
              candidate cap and the direct-access tolerance bound a search and
              a validation regime, not a rate of being right. What licenses
              the judged method here rather than forbidding it is that the
              standard exists and is stated: the cited section's own sentence,
              carried also by
              `safety-path-checklist § The measure of success`, is the bar the
              judge applies, so nothing is invented in the act of applying it.
              That is the distinction the sixth considered zero turns on, and
              this record is the case on its other side. A rate would have
              been fabricated, would then have been measured, would have
              passed, and would have left the quality it stood for unexamined.
              The asymmetry in the criterion is the source's own and is the
              whole content of the standard: coverage yields to correctness,
              so an unclassified zone is not a failure and a confidently wrong
              one is. A machine-checkable floor beneath this record exists and
              is deliberately not here — that a candidate priced by the
              degraded estimate under
              `CalculationSpecification.md § Flat-distance fallback` never
              reaches a confident class is a prohibition on a behaviour and is
              FR-083's, standing to this record as FR-014 and FR-015 stand to
              FR-016.
Resolved-by:  —
```
