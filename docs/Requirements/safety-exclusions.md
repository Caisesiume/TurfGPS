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
Resolved-by:  —
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
Resolved-by:  —
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
Resolved-by:  —
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
Resolved-by:  —
```
