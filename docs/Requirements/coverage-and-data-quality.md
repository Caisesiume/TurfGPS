# Coverage and data quality

Graceful degradation where map or elevation data is thin, and confidence falling with it. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## NFR-006 — Fall access confidence with the evidence behind the estimate

```
Statement:    The system shall record no higher access confidence for an
              estimate resting on the lower-confidence evidence enumerated
              under `SPECIFICATION.md § Terrain confidence` than for an
              estimate of the same zone resting on the high-confidence
              evidence that section enumerates.
Category:     Coverage and data quality
Source:       SPECIFICATION.md § Terrain confidence
Priority:     MUST
Verification: test — access estimates computed for one zone and one stopping
              position under a series of evidence sets drawn from the two
              enumerations under `SPECIFICATION.md § Terrain confidence` carry
              recorded confidence that never rises as the evidence weakens and
              is strictly lower at the weak end of the series
Acceptance:   metric — the access confidence recorded for a zone's access
              estimate, sampled over a series of evidence sets drawn from the
              two enumerations under `SPECIFICATION.md § Terrain confidence`;
              threshold — strictly lower for the sample resting on that
              section's lower-confidence evidence than for the sample resting
              on its high-confidence evidence, and never higher for a sample
              whose evidence is a subset of another sample's;
              condition — the samples differ only in which of that section's
              evidence items are available, with the zone, the stopping
              position, the access path and every other input held fixed, and
              the series spans both ends of the envelope — one sample carrying
              the full high-confidence evidence set and one resting on the
              straight-line estimate under
              `CalculationSpecification.md § Flat-distance fallback` — with
              the confidence levels compared being the deployment's own rather
              than a scale this record fixes
Status:       to-build
Depends-on:   FR-081
Volatility:   settled — the two evidence enumerations under
              `SPECIFICATION.md § Terrain confidence` are stated flatly, carry
              no proposal marking and appear on no open-questions list, and
              this record fixes neither a confidence scale nor a boundary on
              one
Risk:         Access confidence is a gate and not a term, per
              `CalculationSpecification.md § Proposed form: value per minute`,
              and it is the gate that routes a zone into the uncertain class.
              A confidence that does not fall as the evidence thins holds that
              gate open on exactly the estimates it exists to catch: a zone
              priced from a straight-line fallback enters the cost model
              beside one priced from a mapped path and a sampled profile. The
              failure is invisible downstream, because a confident
              classification and a wrongly confident one are the same object
              to everything that consumes one.
Rationale:    The obligation is the relation, not the assignment. That an
              access estimate carries a confidence level at all is behaviour
              and is FR-081's; what this record adds is that the level tracks
              the evidence, which nothing functional obliges — a system
              recording one fixed level for every estimate satisfies the
              assignment perfectly and defeats the gate entirely. Both lanes
              authored the relation and it is filed here on the register's own
              wording, which puts confidence falling with thin data under this
              category by name. The look found no threshold to write against
              and none to derive: `CalculationSpecification.md` carries no
              confidence scale at all,
              `SPECIFICATION.md § Currency confidence` records that access
              confidence is a third dimension whose scale is unsettled, and
              `CalculationSpecification.md § Proposed form: value per minute`
              carries confidence as a gate rather than a quantity. An ordering
              needs the levels to be comparable and not to sit on a scale, so
              the method is test rather than human-judgement even though the
              look found nothing — NFR-002's precedent, on the other
              confidence dimension. Two residues are deliberately not carried.
              Whether the fall is large enough to be honest to the user is
              NFR-002's residue and is owed to the batch scoped to
              `SPECIFICATION.md § Confidence and uncertainty`. That confidence
              must also fall with the resolution of the elevation provider is
              created by `Architecture.md § D6` and
              `Architecture.md § Ports and adapters` rather than by the cited
              section, and naming provider resolution in the condition above
              would have claimed an obligation owed to the batch scoped there.
Resolved-by:  —
```
