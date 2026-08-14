# Journey initialization

How a journey is entered and a planning session begins: input capture, the time budget, validation, session start. **What a journey must and may name is `Journey definition`, not this** — the split is the act of entering against the thing entered. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-038 — Plan no journey without an additional-time limit the user entered

```
Statement:    The system shall not plan a journey for a request in which the
              user has entered no additional-time limit.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints;
              DESIGN.md § Required and optional inputs
Priority:     MUST
Verification: test — a request naming an origin and a destination but carrying
              no additional-time limit entered by the user produces no
              journey, and no limit is supplied in place of the one not
              entered
Acceptance:   given a request naming an origin and a destination in which the
              user has entered no additional-time limit, when the user
              attempts to start planning, then the system plans no journey
              given that same request, when the user attempts to start
              planning, then the system supplies no additional-time limit of
              its own in place of the one not entered
Status:       to-build
Depends-on:   none
Risk:         A limit the user never chose is a promise made on their behalf.
              The target, the stretch band and the absolute ceiling are all
              derived from it, so a substituted value commits a driver to time
              they never agreed to spend, while every check the system runs
              reports compliance.
Rationale:    The section's words are that this is always a user-supplied
              parameter and that the system must not assume a typical value.
              Assuming and inferring are the only two ways to reach a limit
              without the user, so what the two prohibitions leave is the gate
              stated here, on FR-005's precedent: the positive branch — a
              request that does carry a limit is planned — belongs to the
              records that own planning and is not restated.
              `DESIGN.md § Required and optional inputs` is the second source
              because it is what places the limit before the search rather
              than inside it, which is when this gate fires. The criteria turn
              on what the user entered and not on what the request carries: a
              pre-filled default the user never touches is a value the system
              assumed. Whether such a default is admissible at all was the
              open question this record returned, and the Owner ruled on 6
              August 2026 that it is not and that the field starts empty. That
              obligation is FR-055's rather than this record's, being
              separately testable and testable at a different moment: this
              gate fires when planning is attempted, and FR-055 binds what the
              input holds before anything is attempted.
Resolved-by:  #41
```

## FR-039 — Derive no additional-time limit from the journey's length

```
Statement:    The system shall not derive a journey's additional-time limit
              from the journey's length or duration.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — two requests stating the same additional-time limit, on
              journeys whose baseline durations differ several-fold, are each
              planned against the limit stated rather than against one scaled
              to the journey
Acceptance:   given two requests stating the same additional-time limit, one
              for a journey several times the baseline duration of the other,
              when each is planned, then the limit in force for each equals
              the limit stated
              given a request whose journey is long enough that an allowance
              proportional to it would exceed the limit stated, when it is
              planned, then the limit in force is still the limit stated
Status:       to-build
Depends-on:   FR-038
Risk:         An inferred limit is FR-038's substituted value arriving through
              the one door that record leaves open — a limit the user did
              state, silently enlarged because the drive is long. It fails in
              the direction that puts a driver further from their journey than
              they agreed to go, and the enlargement shows up in no figure the
              system displays.
Rationale:    Distinct from FR-044, which fixes what the ceiling is derived
              from. This record binds the limit the user stated; that one
              binds the allowance built on top of it, and the two failures are
              independent — a correctly captured limit can still carry a
              ceiling read off the journey's duration, and an inflated limit
              yields a ceiling that is correct arithmetic over the wrong base.
              The second criterion names the specific inference the section
              forbids rather than leaving it to be read out of the first.
Resolved-by:  #42
```

## FR-040 — Admit an additional-time limit anywhere in the realistic range

```
Statement:    The system shall admit as a journey's additional-time limit,
              unaltered, any value within the realistic range described under
              `SPECIFICATION.md § User time constraints`.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints
Priority:     MUST
Verification: test — a limit at the low end of that range and a limit at its
              high end are each planned against as entered, neither refused
              nor adjusted to a nearer value
Acceptance:   given a limit at the low end of the realistic range described
              under `SPECIFICATION.md § User time constraints`, when the user
              starts planning, then planning runs and the limit in force
              equals the value entered
              given a limit at the high end of that range, when the user
              starts planning, then planning runs and the limit in force
              equals the value entered
Status:       to-build
Depends-on:   FR-038
Risk:         Both ends of this range are ordinary users of the product, and a
              bound placed on the input excludes one of them outright — the
              player squeezing zones into a trip they have to make, or the one
              whose day is built around the journey. An adjustment does it
              silently: the plan is built against a limit the user did not
              choose, and nothing on the screen says so.
Rationale:    The section gives the range's ends as examples of ordinary
              values, not as validation bounds, and states no minimum or
              maximum admissible limit anywhere; the criteria therefore cite
              the section for its ends rather than quoting them, and this
              record creates no bound of its own. Refusal and adjustment are
              one behaviour in two guises — under both, the value the plan is
              built against is not the value entered — which is why they are
              one record. That neither end may be presented as unusual is the
              other half of the section's sentence and is FR-041's, being
              separately testable: a value can be admitted unaltered and still
              be argued with.
Resolved-by:  #43
```

## FR-041 — Present neither end of the realistic range as unusual

```
Statement:    The system shall not present an additional-time limit at either
              end of the realistic range described under
              `SPECIFICATION.md § User time constraints` as unusual.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints
Priority:     SHOULD
Verification: test — a limit entered at either end of that range raises no
              warning, caution or extra confirmation that a value in the
              middle of the range does not raise
Acceptance:   given a limit at the low end of the realistic range described
              under `SPECIFICATION.md § User time constraints`, when it is
              entered, then the interface presents no warning, caution or
              confirmation step that a value in the middle of that range does
              not carry
              given a limit at the high end of that range, when it is entered,
              then the interface presents no warning, caution or confirmation
              step that a value in the middle of that range does not carry
Status:       to-build
Depends-on:   FR-040
Rationale:    The obligation is comparative rather than absolute — the section
              forbids treating either end as unusual, not every affordance
              around the field — so each criterion measures an end against the
              middle of the same range rather than against nothing. A value
              the section calls ordinary, questioned by the product that asked
              for it, tells the user their trip is the wrong shape for the
              tool, and the range is in the section precisely to say that it
              is not. Priority is SHOULD because nothing becomes unbuildable
              without this record: it forbids an addition rather than obliging
              a capability, and it binds whenever the input is built.
Resolved-by:  #48
```

## FR-055 — Open the planner with no additional-time limit entered

```
Statement:    The system shall present the journey's additional-time limit
              with no value entered when the planner opens.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints;
              DESIGN.md § Required and optional inputs
Priority:     MUST
Verification: test — the additional-time limit carries no value when the
              planner opens, on a device where no limit from an earlier
              session is available and on one where such a limit is available,
              whatever its provenance
Acceptance:   given a device on which no additional-time limit from an earlier
              session is available, whatever its provenance, when the planner
              opens, then the additional-time limit carries no value
              given a device on which an additional-time limit from an earlier
              session is available, whatever its provenance, when the planner
              opens again, then the additional-time limit carries no value
Status:       to-build
Depends-on:   FR-038
Risk:         A pre-filled field is the assumption the section forbids,
              wearing the user's own consent. The value is on the screen, the
              user presses start, and from that point nothing downstream can
              tell an entered limit from an assumed one — the target, the
              stretch band and the absolute ceiling are all derived from that
              figure, and every check reports a user-supplied limit. The one
              input the specification insists must never be assumed becomes
              the one input nobody can audit.
Rationale:    The Owner ruled on 6 August 2026 that a pre-filled default is
              not user-supplied and that the field starts empty, closing the
              open question FR-038 returned. The reason is that a default
              converts the question the product is asking — how much time will
              you spare — into an answer nobody chose, and
              `SPECIFICATION.md § User time constraints` declares itself the
              single definition of an allowance model built entirely on that
              figure. This is a separate record and not an extension of FR-038
              for two reasons. It is testable at a different moment: FR-038
              fires when planning is attempted, and this fires when the
              planner opens, before the user has done anything. And it is what
              makes FR-038's criteria decidable at all — those criteria turn
              on whether the user entered a limit, and while a default may sit
              in the field two engineers can read an untouched one as entered
              and as not entered. Absorbing it would leave FR-038 carrying a
              gate and an initial state, which is two behaviours. The default
              is the pattern the design uses elsewhere, which is why the
              prohibition needs stating:
              `DESIGN.md § First-run initialization` pre-orders the attribute
              list by rarity so an untouched list is already valid, and
              `DESIGN.md § Required and optional inputs` defaults the
              objective selection. What separates this input from those is
              that an untouched value here is a typical value assumed, which
              the cited section forbids by name. The second criterion reads
              the ruling literally: a limit restored from an earlier session
              was entered for a different journey, and the allowance is a
              property of the journey being planned rather than of the device.
              That reading names one source, and the Owner widened it on 7
              August 2026: no value is present when the journey begins,
              whatever its provenance. The criterion carries that universal
              rather than an enumeration, because a named provenance reads as
              a scope — an engineer holding a limit the system stored rather
              than the user finds the obligation silent on it, and the
              population default returns with a more personal-looking source.
              Both criteria sit on retention rather than planning history, so
              the pair covers every device state; a never-used device has
              nothing available. What the record does not reach is a stored
              plan being reopened. A stored route keeps the classifications
              and costs behind it, per `SPECIFICATION.md § Route persistence`,
              so the limit it was approved against belongs to that journey and
              travels with it; reopening one is
              `DESIGN.md § Returning to a stored plan` rather than the planner
              opening, and `DESIGN.md § Never gate stored plans on the wizard`
              draws that line itself. That is scope and not an obligation:
              nothing here says what a reopened plan shows.
Resolved-by:  #60
```

## FR-056 — Refuse an additional-time limit that is not positive

```
Statement:    The system shall not admit as a journey's additional-time limit
              a value that is not positive.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints;
              DESIGN.md § Required and optional inputs
Priority:     MUST
Verification: test — a request carrying an additional-time limit that is not
              positive produces no journey, and no positive limit is planned
              against in its place
Acceptance:   given a request in which the user has entered an additional-time
              limit that is not positive, when the user attempts to start
              planning, then the system plans no journey
              given that same request, when the user attempts to start
              planning, then the system supplies no additional-time limit of
              its own in place of the value refused
Status:       to-build
Depends-on:   FR-038
Risk:         Every quantity in the allowance model is derived from the stated
              limit, per `SPECIFICATION.md § User time constraints`, so a
              limit that is not positive propagates into all of them and none
              of them notices. A limit of nothing spends a full solve to reach
              an allowance that can hold no stop, and what the user is then
              shown is the honest no-answer message of
              `DESIGN.md § When nothing fits at all` — which names the
              constraints that bound hardest and invites the user to relax
              them, on a journey where the zones were never the problem. A
              negative limit is worse: it asks for a Turf journey shorter than
              the drive without one, and every ceiling check reports
              compliance with it.
Rationale:    The Owner ruled on 6 August 2026 that a limit of zero or less is
              refused, on the ground that a non-positive allowance is a
              request the product cannot answer. The boundary creates no
              constant and cites none: positivity is a property of a quantity
              of time rather than a value that could have been chosen
              otherwise, which is the reasoning FR-015 records for an ordering
              needing no tolerance, and the look for a figure found nothing to
              cite.
              `CalculationSpecification.md § The absolute additional-time ceiling`
              derives the allowance from the stated limit and sets no floor
              under it, and
              `CalculationSpecification.md § Additional journey time` defines
              the cost metric without bounding it. This is not FR-040's: that
              record admits unaltered any value within the realistic range and
              creates no bound of its own, and the non-positive case sits
              below the range, where FR-040 neither admits nor refuses. Nor is
              it FR-038's, where the user entered nothing at all; here they
              entered something the model cannot use. The two criteria are one
              behaviour in FR-038's shape rather than two: refusing the value
              and substituting a workable one for it are the two ways the same
              outcome fails, and a record stating only the first would be met
              by a system that plans against a limit of its own choosing. What
              the interface says when it refuses is not authored here and no
              document states it.
Resolved-by:  #58
```

## FR-057 — Impose no maximum admissible additional-time limit

```
Statement:    The system shall impose no maximum on the additional-time limit
              a journey may be planned against.
Category:     Journey initialization
Source:       SPECIFICATION.md § User time constraints
Priority:     SHOULD
Verification: test — two limits above the high end of the realistic range
              described under `SPECIFICATION.md § User time constraints`, one
              several times the other, are each planned against as entered,
              neither refused nor adjusted
Acceptance:   given an additional-time limit above the high end of the
              realistic range described under
              `SPECIFICATION.md § User time constraints`, when the user starts
              planning, then planning runs and the limit in force equals the
              value entered
              given a further limit several times that one, when the user
              starts planning, then planning runs and the limit in force
              equals the value entered
Status:       to-build
Depends-on:   FR-040
Rationale:    The Owner ruled on 6 August 2026 that no maximum is imposed.
              FR-040 does not already carry this: it obliges admission within
              the realistic range and creates no bound of its own, so a
              maximum placed just above that range satisfies FR-040 while
              refusing a user the section calls ordinary — and since the
              range's high end is *several hours* rather than a figure, a
              bound placed near it is arguable against FR-040 and indefensible
              against the section. Stated as a prohibition on adding a bound
              rather than as an obligation to admit any positive value,
              deliberately: the obliging form would subsume FR-040 and give
              one duty two homes, and the prohibiting form reaches only the
              addition FR-040 leaves open. Separate from FR-056 because the
              outcomes are opposite and separately testable — one obliges a
              refusal and this forbids one — where FR-040's refusal and
              adjustment were one behaviour in two guises because both end
              with the plan built against a value the user did not enter. No
              maximum is quoted anywhere because there is none to quote: any
              figure would be a constant with no home in
              `CalculationSpecification.md`, and the criteria therefore cite
              the section for the range's high end as the point a test starts
              above, not as a bound this record creates. Priority is SHOULD on
              FR-041's precedent: nothing becomes unbuildable without this
              record, which forbids an addition rather than obliging a
              capability, and it binds whenever the input is built.
Resolved-by:  #59
```
