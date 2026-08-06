# Recommendation disclosure

What the system must tell the user about a recommendation it offers: the figures it must show, the labels an alternative must carry, and the reasons it must give for what the optimizer did. Distinct from `Recommendation set composition`, which decides membership of the offered set and disclaims presentation in its own scope line — this category owns what is said about an alternative, never how it looks, which is `DESIGN.md`'s. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-047 — Present an above-limit alternative as a stretch alternative

```
Statement:    The system shall present a journey alternative whose additional
              time exceeds the additional-time limit the user stated as a
              stretch alternative, distinguished from the alternatives it
              presents as within that limit.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § User time constraints;
              SPECIFICATION.md § Recommended journey alternatives
Priority:     MUST
Verification: test — an offered alternative above the stated limit carries the
              stretch identification where the user sees it, and an offered
              alternative within the limit does not
Acceptance:   given an offered journey alternative whose additional time
              exceeds the limit the user stated, when the offered set is
              presented, then that alternative is identified to the user as a
              stretch alternative
              given an offered journey alternative whose additional time is
              within that limit, when the offered set is presented, then it is
              not identified as a stretch alternative
Status:       draft
Depends-on:   FR-008; FR-038
Risk:         An unlabelled stretch is the quiet exceedance the section
              forbids: the user picks the richest-looking route believing it
              is the one they asked for, and meets the difference as time on
              the road. It also disarms the one check the product cannot run
              for itself — the user's own judgement of whether the extra
              minutes are worth it.
Rationale:    One record and not two, though the section states the
              identification and its visibility in separate sentences: an
              identification the user never sees is not an identification, so
              the two sentences bound one behaviour, and splitting them would
              leave a record satisfiable by an internal flag. The second
              criterion is what makes the first a distinction rather than a
              decoration. What the identification looks like is `DESIGN.md`'s
              and is not authored here; that a stretch alternative must also
              carry a reason is FR-048's, and whether it should have been
              offered at all is FR-049's.
Resolved-by:  —
```

## FR-048 — Name what the stated limit was exceeded for

```
Statement:    Where the system offers a journey alternative whose additional
              time exceeds the additional-time limit the user stated, the
              explanation it presents for that alternative shall name the zone
              or the zone attribute for which the limit was exceeded.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — the explanation presented with an offered above-limit
              alternative names the zone or the zone attribute for which the
              limit was exceeded, and an explanation carrying only that
              alternative's additional time fails
Acceptance:   given an offered journey alternative whose additional time
              exceeds the limit the user stated and which captures a zone that
              the alternatives within that limit do not, when its explanation
              is presented, then the explanation names that zone or the
              attribute it carries
              given an offered alternative above that limit whose explanation
              states its additional time, when the explanation is presented,
              then it also names the zone or the attribute for which the limit
              was exceeded
Status:       draft
Depends-on:   FR-047
Risk:         A stretch alternative the user cannot account for is a route
              they have to take on trust, on precisely the decision the
              section reserves to them. It also removes the only external
              check the value model has: unable to see which zone bought the
              extra minutes, the user cannot notice that the optimizer spent
              them on something they do not care about.
Rationale:    The section names the content the explanation must carry and
              nothing about its form, so this record binds the content alone.
              What the criteria demand is a cause, which is why the second one
              exists: an above-limit alternative shown with its additional
              time and no cause satisfies every reading of explained that
              counts characters and none that helps the user decide. That
              every recommendation carries an explanation at all is owned by
              `SPECIFICATION.md § Explaining combined objectives` and is not
              authored here. Where more than one zone accounts for the
              exceedance the section is silent, and that gap is returned with
              this batch rather than closed by a rule invented here.
Resolved-by:  —
```

## FR-054 — Present the additional time for the journey as a whole

```
Statement:    The system should present a journey alternative's additional
              time to the user as the total across the whole journey.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § Journeys with several legs
Priority:     SHOULD
Verification: test — a multi-leg alternative is presented with the total
              additional time for the whole journey, and a presentation
              carrying only per-leg figures fails
Acceptance:   given a journey alternative with more than one leg, when it is
              presented, then the additional time shown for it is the total
              across all of its legs
              given the same alternative presented with a per-leg breakdown,
              when it is presented, then the journey total is shown alongside
              the per-leg figures
Status:       draft
Depends-on:   FR-002; FR-008
Rationale:    The section makes the total the figure that matters and the
              breakdown useful detail, so this record obliges the total and
              permits the breakdown rather than forbidding it; the second
              criterion states that, and it is there so the record cannot be
              built as a prohibition on per-leg figures. The quantity
              presented is the one FR-053 tests, which is what keeps the
              number the user reads and the number the ceiling checks from
              being two different figures.
Resolved-by:  —
```
