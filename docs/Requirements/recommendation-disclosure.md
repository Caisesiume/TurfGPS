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
Status:       to-build
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
Resolved-by:  #45
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
Status:       to-build
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
Resolved-by:  #52
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
Status:       to-build
Depends-on:   FR-002; FR-008
Rationale:    The section makes the total the figure that matters and the
              breakdown useful detail, so this record obliges the total and
              permits the breakdown rather than forbidding it; the second
              criterion states that, and it is there so the record cannot be
              built as a prohibition on per-leg figures. The quantity
              presented is the one FR-053 tests, which is what keeps the
              number the user reads and the number the ceiling checks from
              being two different figures.
Resolved-by:  #54
```

## FR-066 — Tell the user where the promotion cap shaped the result

```
Statement:    Where the system has recorded that the promotion cap bound for a
              route alternative, it should tell the user that the result was
              shaped by a limit on how many candidate zones were examined.
Category:     Recommendation disclosure
Source:       CalculationSpecification.md § Bounding the candidate set;
              SPECIFICATION.md § Candidate zone identification
Priority:     SHOULD
Verification: test — a journey for which the cap's binding was recorded is
              presented with that fact stated to the user, and one for which
              it was not carries no such statement
Acceptance:   given a journey for which the system recorded that the promotion
              cap bound, when its recommendations are presented, then the user
              is told that the result was shaped by a limit on how many
              candidate zones were examined
              given a journey for which no such binding was recorded, when its
              recommendations are presented, then no such statement is made
Status:       to-build
Depends-on:   FR-065
Volatility:   proposed-constant — the disclosure fires against the promotion
              cap under
              `CalculationSpecification.md § Bounding the candidate set`, a
              proposed default, so how often a user meets it moves with that
              figure
Rationale:    The cited calculation section's stated reason for recording the
              binding is the user — someone in a dense area is receiving a
              result shaped by a limit rather than by their preferences — so
              an internal flag does not discharge it, and the Owner ruled on 7
              August 2026 that the disclosure is a second record. What is said
              follows the posture under
              `SPECIFICATION.md § Confidence and uncertainty`: material
              uncertainty is communicated without overwhelming the user, so
              this record binds that the user is told and never how, which is
              `DESIGN.md`'s. The verb is should because the cited section
              states the recording as a should and this is the softer half of
              it; the second criterion exists so the obligation cannot be
              built as a standing notice on every journey. Risk is omitted and
              its argument folded here, on FR-025's precedent: the failure
              withholds a caveat rather than sending a driver anywhere, and
              FR-065 carries the diagnostic loss.
Resolved-by:  —
```

## FR-088 — State the material uncertainty in an access estimate

```
Statement:    Where a recommendation rests on an access estimate carrying
              material uncertainty, the system should state that uncertainty
              and what it rests on to the user.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § Terrain confidence
Priority:     SHOULD
Verification: test — a recommended stop whose access estimate rests on an
              unmapped final approach is presented with that fact stated, and
              one resting on a mapped stopping position, a connected path and
              a complete profile carries no such statement
Acceptance:   given a recommended stop whose access estimate rests on an
              unmapped final approach or an incomplete elevation profile, when
              the recommendation is presented, then the presentation states
              that the estimate carries that uncertainty and what it rests on
              given a recommended stop whose access estimate rests on a mapped
              stopping position, a connected pedestrian path and a complete
              elevation profile, when the recommendation is presented, then no
              such statement is made for it
Status:       to-build
Depends-on:   FR-081
Volatility:   open-question — what counts as material rests on the enumeration
              under `SPECIFICATION.md § Terrain confidence` standing as a
              working default, no scale for access confidence existing in
              `CalculationSpecification.md`
Risk:         What the user is told is the last gate before the driver acts,
              and this record is the whole of it. A stop presented without its
              caveat reads exactly like a mapped, profiled estimate, so the
              driver arrives at an unmapped final approach and an estimated
              climb having been given no reason to expect either.
              `SPECIFICATION.md § Requirements the data cannot verify`
              reserves the judgement at the roadside to the driver precisely
              because the system cannot make it, and an estimate that says
              nothing takes that judgement back by withholding the one thing
              it would be made on.
Rationale:    The cited section obliges the recommendation to communicate
              material uncertainty and gives a worked example of what that
              means for an access estimate — a mapped parking location, an
              unmapped final approach, an estimated elevation gain — and the
              criteria are written against the inputs it enumerates rather
              than against a materiality threshold, because none exists.
              Distinct from the confidence this batch assigns and gates on:
              FR-081 produces the level and FR-082 decides what may be
              recommended at all, while this record is about what is said of a
              stop that has already passed that gate. Distinct also from the
              currency dimension, which
              `SPECIFICATION.md § Currency confidence` separates by name and
              whose honesty residue `README.md` already records as owed to
              another batch. What the statement looks like is `DESIGN.md`'s;
              the second criterion exists so the record cannot be built as a
              standing disclaimer on every stop. `Risk` is present rather than
              folded into this field because what the user is told about a
              stop is a lane of `safety-path-checklist` rather than an
              adjacent concern:
              `SPECIFICATION.md § Requirements the data cannot verify` makes
              the labelling of an unverified stop an obligation in the same
              list as the exclusions, and the record is on the safety set on
              the sentinel's ruling of 7 August 2026.
Resolved-by:  —
```

## FR-131 — Identify every marked stretch of a warned leg

```
Statement:    The system shall identify, with every warned leg it produces,
              each stretch of the leg that runs along a way, and each point at
              which it passes a barrier, that the map data marks as private or
              access-restricted.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § Enforceable exclusions
Priority:     MUST
Verification: test — against routing data built from a test extract, a warned
              leg that runs along a marked way and passes a marked barrier
              carries an identification of each, locating it on the leg's
              shape, and a warned leg leaving a location the user entered
              along FR-126's run carries the run's stretch as well
Acceptance:   given a warned leg that runs along a way so marked and passes a
              barrier so marked, when the journey is returned, then the way's
              stretch and the barrier's point are each identified, each
              locating its own part of the leg's shape
              given a warned leg that leaves or reaches a location the user
              entered along the run FR-126 admits there, when the journey is
              returned, then the run's stretch is identified too
Status:       to-build
Depends-on:   FR-129
Volatility:   open-question — the markings are FR-125's, their spelling
              designed in `Architecture.md § The access-marking tables` and
              read as `DECISIONS.md § RD-063` to `DECISIONS.md § RD-066`
              decide; the ruling identifying them is
              `DECISIONS.md § RD-067`'s
Risk:         An unidentified stretch is a private crossing the user is never
              shown: the warning reads as covering the leg while one gate on
              it goes undrawn, and the confirmation the user gives is for
              less than they will drive.
Rationale:    The Owner ruled that the route "becomes clearly marked as
              private" and that the private stretch is highlighted
              (`DECISIONS.md § RD-067`), and a client can highlight only what
              the produced leg identifies, so the obligation sits with the leg
              and FR-132 draws it. Every stretch is identified, the run's
              included, because the run is private travel the user will
              drive and the ruling marks the route as private, not some of
              it; a leg that is not warned is unchanged by the ruling, and its
              run stays as FR-126 admits it. Locating each identification on
              the leg's shape is what lets the stretch be drawn where it is
              and not merely listed. FR-128 makes a way within an area so
              marked a way so marked for this record. How the identification
              is carried is not this record's.
Resolved-by:  #271
```

## FR-132 — Show each marked stretch of a warned leg on the map

```
Statement:    The system shall show, wherever it presents a journey
              alternative holding a warned leg on the map, each stretch and
              point FR-131 identifies on that leg distinguished from the rest
              of the journey.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § Enforceable exclusions;
              DESIGN.md § Reviewing the recommended route
Priority:     MUST
Verification: test — a presented alternative holding a warned leg draws each
              identified stretch and point in the form `DESIGN.md` gives a
              marked stretch, distinct from the rest of the journey, and an
              alternative holding no warned leg draws nothing in that form
Acceptance:   given a journey alternative holding a warned leg whose marked
              stretches and points FR-131 identifies, when it is presented on
              the map, then each of them is shown distinguished from the rest
              of the journey
              given a journey alternative holding no warned leg, when it is
              presented on the map, then no part of it is shown as a marked
              stretch
Status:       to-build
Depends-on:   FR-131
Volatility:   open-question — how a marked stretch is drawn is owed by
              `DESIGN.md`, to which the Owner delegated it
              (`DECISIONS.md § RD-067`)
Risk:         A private stretch drawn like the rest of the route leaves the
              user a warning with nothing to point at: they cannot see
              whether the private part is the gate at their own door or a
              shortcut in the middle they could plan around, and reconsider
              nothing.
Rationale:    The Owner ruled that the private stretch is highlighted
              (`DECISIONS.md § RD-067`). This category owns what is said about
              an alternative and never how it looks, so the record obliges the
              distinction and `DESIGN.md` owns its form. The second criterion
              keeps the form meaning one thing, as FR-047's second does for a
              stretch alternative: a form also worn by an ordinary stretch
              distinguishes nothing. Whether the distinction is clear enough
              is FR-135's, which no test answers.
Resolved-by:  #273
```

## FR-133 — Warn that a warned route includes private property and should be reconsidered

```
Statement:    The system shall tell the user, with every journey alternative
              it presents holding a warned leg, that the route includes
              private property and should be reconsidered.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § Enforceable exclusions;
              DESIGN.md § Reviewing the recommended route
Priority:     MUST
Verification: test — a presented alternative holding a warned leg carries the
              warning in `DESIGN.md`'s wording, saying both things, before the
              user is asked to confirm it, and an alternative holding no
              warned leg carries none
Acceptance:   given a journey alternative holding a warned leg, when it is
              presented, then the user is told, before being asked to confirm
              it, that the route includes private property and that it should
              be reconsidered
              given a journey alternative holding no warned leg, when it is
              presented, then no such warning is shown
Status:       to-build
Depends-on:   FR-129
Volatility:   open-question — the wording is owed by `DESIGN.md`, to which
              the Owner delegated it (`DECISIONS.md § RD-067`); what it must
              say is the Owner's and is settled
Risk:         A warning that says only that the route is private reads as a
              note on a route the system stands behind: the user drives it
              because the plan offered it. Without the warning the
              highlighted stretch is a colour with no meaning.
Rationale:    The Owner's words name both halves: the route is "displayed
              with warnings that the route should be reconsidered and that
              this route includes private property"
              (`DECISIONS.md § RD-067`). They are one record and not two,
              because they are one warning the Owner stated whole and a test
              finds both in it or fails; what each says in words is
              `DESIGN.md`'s, delegated by the Owner, and is not authored
              here. Shown before the confirmation is asked, so FR-134
              confirms a warning the user has seen.
Resolved-by:  #273
```

## FR-135 — Owner judgement — a warned route's private stretches and warning are unmistakable

```
Statement:    The system shall present a journey alternative holding a warned
              leg so that a user sees, before being asked to confirm it,
              which part of the route is private and that the route should be
              reconsidered.
Category:     Recommendation disclosure
Source:       SPECIFICATION.md § Enforceable exclusions;
              DESIGN.md § Reviewing the recommended route
Priority:     MUST
Verification: human-judgement — the Owner, on the built client at a mobile
              width, against the standard in the acceptance below, the
              verdict recorded on the story
Acceptance:   the Owner examines the client presenting a journey alternative
              holding a warned leg with one marked stretch at a location the
              user entered and one in transit, at a mobile width, and judges
              against the Owner's own words, that the route "becomes clearly
              marked as private" (`DECISIONS.md § RD-067`), whether a user
              who reads nothing but that screen can tell which parts of the
              route are private and that the route should be reconsidered,
              before the confirmation is asked
Status:       to-build
Depends-on:   FR-132; FR-133
Volatility:   settled — the standard is the Owner's ruling in
              `DECISIONS.md § RD-067`
Risk:         A stretch drawn in a form a phone screen in daylight does not
              tell from the route, or a warning below the fold, passes every
              test FR-132 and FR-133 carry and leaves the user confirming a
              crossing they did not see.
Rationale:    FR-132 and FR-133 test that the stretch is distinguished and
              the warning shown; whether either is clear is what the Owner's
              "clearly" asks, and no test answers it. Looked for first:
              `CalculationSpecification.md` holds no threshold for the
              legibility or contrast of anything presented, and none can be
              derived from one there, and `DESIGN.md § Still owed by this document`
              still owes the graphic profile and the page layouts, so no
              standard short of the Owner's own words exists to test against.
Resolved-by:  #273
```
