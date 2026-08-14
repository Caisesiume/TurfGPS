# Candidate identification

Selecting which zones enter evaluation, and bounding that set. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-063 — Bound a corridor set by the corridor half-width

```
Statement:    The system shall admit a zone to a route alternative's corridor
              set only where the zone lies within the corridor half-width
              defined under
              `CalculationSpecification.md § Bounding the candidate set` of
              that route.
Category:     Candidate identification
Source:       SPECIFICATION.md § Candidate zone identification
Priority:     MUST
Verification: test — a zone lying further from a route than the corridor
              half-width derived for it is absent from that route's corridor
              set, and a journey planned with a larger additional-time
              allowance derives a corridor no narrower than one planned with a
              smaller allowance
Acceptance:   given a zone lying further from a route alternative than the
              corridor half-width derived for that route, when the corridor
              set is resolved, then the zone is not in it
              given two journeys over one route differing only in the
              additional-time allowance, when each corridor set is resolved,
              then the corridor derived for the larger allowance is no
              narrower than the one derived for the smaller, and both lie
              between the floor and the cap stated under
              `CalculationSpecification.md § Bounding the candidate set`
Status:       draft
Depends-on:   FR-023; FR-038
Volatility:   proposed-constant — the corridor half-width, its floor and its
              cap under
              `CalculationSpecification.md § Bounding the candidate set` are
              proposed defaults with nothing measured behind them
Risk:         The corridor is where per-journey call volume is decided before
              anything else bounds it: unbounded, it puts every zone in the
              extract into routing and access analysis, and the failure is not
              a wrong answer but a solve that never returns. Bounded too
              tightly it silently drops the rare high-value zone the wide
              corridor exists for, and nothing in the result says a zone was
              never looked at.
Rationale:    The obligation is created by the cited section requiring the
              spatial limits to be numbers; the numbers have one home and are
              cited, so this record survives every one of them moving. The
              second criterion states the relation that calculation section
              defines — the corridor scales with the allowance, between a
              floor and a cap — while carrying none of the three figures,
              which is what keeps the record from hardening a proposal. That
              the corridor set is reduced much further before full evaluation
              is FR-064's, and this record deliberately admits far more than
              routing can serve.
Resolved-by:  —
```

## FR-064 — Promote no more candidates to full evaluation than the cap

```
Statement:    The system shall promote no more zones from a route
              alternative's corridor set to full evaluation than the cap
              defined under
              `CalculationSpecification.md § Bounding the candidate set`.
Category:     Candidate identification
Source:       SPECIFICATION.md § Candidate zone identification
Priority:     MUST
Verification: test — a corridor set holding more zones than the cap promotes
              exactly the cap and no more, and a journey carrying several
              route alternatives bounds each one's promoted set rather than
              their total
Acceptance:   given a corridor set holding more zones than the cap, when
              candidates are promoted to full evaluation, then the number
              promoted is the cap
              given a journey carrying more than one route alternative, each
              with a corridor set holding more zones than the cap, when
              candidates are promoted, then the cap bounds each route
              alternative's promoted set rather than their total
Status:       draft
Depends-on:   FR-063
Volatility:   proposed-constant — the promotion cap under
              `CalculationSpecification.md § Bounding the candidate set` is a
              proposed default with nothing measured behind it
Risk:         This is the point at which the pipeline's cost becomes bounded
              at all: routing and access analysis are the expensive stages,
              and the corridor deliberately admits far more than they can
              serve. An unbounded promotion is a per-journey call volume
              nobody can state, spent out of the same budget every later stage
              draws on, so the failure arrives as a solve that does not finish
              rather than as a wrong answer.
Rationale:    The cap binds per route alternative because that is the unit the
              cited constant is stated over, and the second criterion exists
              because a cap applied to the journey's total is the natural
              implementation and satisfies the first. The figure has one home
              and is cited, so the record is true at whatever value a
              deployment holds. The section's other mechanism, the corridor
              itself, is FR-063's, and that the cap's binding must not be
              silent is FR-065's and FR-066's. This obligation is the one the
              reserved-ID worked example in `requirements-authoring`
              illustrates; that example is fictional in its identity alone,
              and the duty it draws on is created here and is owed a real
              record.
Resolved-by:  —
```

## FR-065 — Record where the promotion cap bound

```
Statement:    Where a route alternative's corridor set holds more qualifying
              zones than the promotion cap admits, the system shall record
              that the cap bound for that route alternative.
Category:     Candidate identification
Source:       CalculationSpecification.md § Bounding the candidate set;
              SPECIFICATION.md § Candidate zone identification
Priority:     MUST
Verification: test — a corridor set holding more qualifying zones than the cap
              admits yields a recorded binding for that route alternative, and
              one holding no more than the cap admits yields none
Acceptance:   given a route alternative whose corridor set holds more
              qualifying zones than the promotion cap admits, when candidates
              are promoted, then the system records that the cap bound for
              that route alternative
              given a route alternative whose corridor set holds no more
              qualifying zones than the cap admits, when candidates are
              promoted, then no such binding is recorded for it
Status:       draft
Depends-on:   FR-064
Volatility:   proposed-constant — the record fires against the promotion cap
              under
              `CalculationSpecification.md § Bounding the candidate set`, a
              proposed default, so how often it fires moves with that figure
              even though the obligation does not
Risk:         A result shaped by the cap and a result shaped by the user's
              preferences are indistinguishable in the output, so without this
              fact nobody can tell a dense-area answer from a thin-corridor
              one — not the user, not a reviewer asking why a zone was never
              offered, and not whoever is one day asked to move the figure. It
              is also the only evidence that would ever justify moving it.
Rationale:    The cited calculation section states that where the cap binds it
              must not do so silently, which is the obligation, and `Source`
              names that section first because it is where the duty is created
              rather than where the cap is applied — FR-046's precedent. The
              second criterion is what keeps the record from being satisfied
              by a flag that is always set. Telling the user is FR-066's: a
              system can record this perfectly and say nothing, which is the
              state the cited section's own reasoning rules out, and the Owner
              ruled on 7 August 2026 that both halves are owed.
Resolved-by:  —
```

## FR-080 — Apply the configured maximum walking distance as a bound

```
Statement:    The system shall apply the maximum walking distance the user
              configured as a bound on which park-and-walk candidates may be
              recommended, and not as the means of choosing between candidates
              within it.
Category:     Candidate identification
Source:       SPECIFICATION.md § Walking distance is priced, not filtered
Priority:     MUST
Verification: test — a candidate whose walking distance exceeds the configured
              maximum is in no produced alternative, and of two candidates
              within it the longer walk is taken where the comparison of value
              against cost favours it
Acceptance:   given a park-and-walk candidate whose walking distance exceeds
              the maximum the user configured, when journey alternatives are
              produced, then no alternative includes it
              given two park-and-walk candidates whose walking distances are
              both within that maximum, when candidates are selected, then
              neither is withheld from selection for its walking distance
              alone
              given two such candidates of which the one with the longer walk
              is the better under the comparison of value against cost, when
              candidates are selected, then it is not passed over in favour of
              the shorter walk
Status:       draft
Depends-on:   FR-011; FR-074
Volatility:   settled — the maximum is a value the user supplies rather than a
              constant, so no proposal moves this record; that a limit may
              also be set per zone attribute is permitted by the source and
              obliged by nothing
Risk:         A walking-distance filter applied before the comparison removes
              the one zone that justified setting the bound generously — the
              rare, highly ranked zone a longer walk buys — and it removes it
              silently, because a candidate filtered early leaves no trace in
              the result. Ignored in the other direction, an unbounded walk
              promises a user a stop they told the system they would not make.
Rationale:    The cited section states both halves: the configured maximum
              remains an outer bound, a statement of what the user is willing
              to do at all, and it is a constraint on the search space rather
              than the mechanism that chooses between candidates, because
              walking time is a cost that competes like any other. One record
              and not two, because its subject is the role one setting plays
              and the two criteria are the two faces of that role; split, the
              bound would read as licence to filter and the prohibition would
              read as licence to ignore the bound. The comparison that does
              choose is FR-011's, over the form stated under
              `CalculationSpecification.md § Proposed form: value per minute`,
              and neither is restated here. Per-attribute limits are permitted
              by the source and are not obliged here: the source's word is
              may, and a record obliging them would fail Necessary.
Resolved-by:  —
```
