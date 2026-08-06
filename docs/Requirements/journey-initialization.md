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
Status:       draft
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
              assumed, and whether such a default is admissible at all is an
              open question returned with this batch.
Resolved-by:  —
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
Status:       draft
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
Resolved-by:  —
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
Status:       draft
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
Resolved-by:  —
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
Status:       draft
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
Resolved-by:  —
```
