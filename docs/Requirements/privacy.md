# Privacy

What a stored plan holds, what the system declines to hold at all, what may leave the system with a hand-off, and the retention and access controls standing in place of an omission that cannot be made — `Architecture.md § Persistence and cross-device transfer` records that the origin coordinate cannot be designed out, so *the controls that matter are retention and access, not omission*, and a category holding only what is stored could carry neither. **Egress is inside that scope**: a hand-off is composed at dispatch time from live session state rather than from the stored payload, so a bound on what it may carry is a control no record about what is stored can reach. Distinct from `Observability`, which owns what must be measurable about a running system, and from `Hand-off and dispatch`, which owns the mechanics of delivering a portion: a prohibition on logging a credential is a privacy control and not an observability one, and a bound on what a hand-off may carry out is a privacy control and not a dispatch one. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## NFR-008 — Keep the plan code out of every log the service writes

```
Statement:    The TurfGPS service process shall write no plan code into any
              log output it emits, its own statements and its dependencies'
              alike, at any verbosity the deployment can select, whether as a
              request target, a structured attribute, an error message or any
              other form.
Category:     Privacy
Source:       Architecture.md § Personal data;
              DEPLOYMENT.md § Where the deployment configuration lives
Priority:     MUST
Verification: test — the whole of the log output a service process writes over
              a run that retrieves a stored plan by its plan code, and then
              attempts a retrieval with a code the store does not hold, is
              searched for the codes the run used
Acceptance:   metric — occurrences of a plan code used in the run, matched as
              a substring across the whole of the log output the service
              process writes;
              threshold — zero, in every record the run produces, including
              those written on the failure path;
              condition — the run retrieves a stored plan by its plan code and
              then attempts a retrieval with a code the store does not hold,
              with the service's logging set to the most verbose level the
              deployment can select
Status:       to-build
Depends-on:   none
Volatility:   settled — the constraint is stated outright in
              `DEPLOYMENT.md § Where the deployment configuration lives`,
              rests on no figure and appears on no open-questions list, and
              the credential argument it derives from is stated as flatly in
              `Architecture.md § Personal data`
Risk:         The plan code is the only credential for an object holding an
              origin coordinate that is frequently the user's home, per
              `Architecture.md § Personal data`. A code in a log makes every
              reader of that log a holder of that credential, and every copy
              of the log — shipped, backed up or pasted into an issue —
              carries it with them; the store cannot tell such a reader from
              the user, because there is nothing else to tell them apart by.
              The failure is silent and it is not reversible: nothing
              observable changes when a code is written, and a code cannot be
              un-disclosed afterwards.
Rationale:    `DEPLOYMENT.md § Where the deployment configuration lives`
              splits this control in two and says it is where the constraint
              is derived rather than where it is discharged. The proxy half is
              a deployment artefact and is bound to the proxy configuration
              there by NFR-015, which carries that half; this record takes
              only the service half, which that section says binds whoever
              writes the log statements in `service/`, and which no file under
              `deploy/` can constrain. The method is test rather than
              inspection deliberately, and the statement is scoped to the
              process rather than to those statements for the same reason:
              that section records that the service installs no log handler,
              so its own output and a dependency's reach the same host log,
              and both an inspection of the statements written in `service/`
              and a statement binding only them would pass over the second.
Resolved-by:  —
```

## NFR-009 — Hold no Turf username in a stored plan

```
Statement:    A stored plan shall hold no Turf username anywhere in its stored
              payload, at any nesting depth.
Category:     Privacy
Source:       Architecture.md § Persistence and cross-device transfer;
              Architecture.md § Personal data
Priority:     MUST
Verification: test — a stored plan written from a session in which the user's
              Turf username is known to the system is read back from the store
              and its whole payload searched at every nesting depth for that
              username, under a key name nested below the top level as well as
              at it
Acceptance:   metric — occurrences of the session's Turf username in a stored
              plan's payload, counted over every key and every value at every
              nesting depth rather than over top-level keys alone;
              threshold — zero;
              condition — the plan is stored from a session in which the
              system holds the user's Turf username, and the search is run
              against the payload read back from the store rather than against
              the presence of the schema constraint that refuses one
Status:       to-build
Depends-on:   none
Volatility:   open-question —
              `Architecture.md § Persistence and cross-device transfer` leaves
              two treatments of the username live, keeping it out of the
              stored object or stating its retention explicitly, and records
              the first as its recommendation rather than as a ruling; a
              ruling for the second falsifies this record
Risk:         A stored plan already holds an origin the user typed, which
              `Architecture.md § Personal data` records as frequently their
              home and as impossible to design out. The username is the one
              separable part, and it is the part that turns a coordinate under
              an opaque key into a named person's dwelling: without it the
              store holds a place, with it it holds who lives there. Once
              joined the two are one object, and removing the field later does
              not unmake a payload already written.
Rationale:    The criterion is written against the payload and not against the
              schema constraint, because `Architecture.md § Personal data`
              states that constraint's limit itself: `jsonb_exists` tests
              top-level keys only, so a nested username passes it, and that
              section says outright it is a tripwire against the obvious
              mistake rather than a proof of absence and should be described
              that way rather than relied upon. A criterion satisfied by the
              constraint's existence would report green on exactly the case
              the constraint cannot see. This record therefore obliges the
              absence and not the tripwire, and a deployment satisfies it by
              what it stores rather than by what its schema declares.
Resolved-by:  —
```

## NFR-010 — Make a plan code unpredictable

```
Statement:    A plan code shall be unpredictable from any information
              observable outside the service, drawn from a cryptographically
              secure random source and derived from no sequence, counter,
              clock reading or property of the plan it names.
Category:     Privacy
Source:       Architecture.md § Personal data;
              Architecture.md § What is unproven
Priority:     MUST
Verification: inspection — the code path that generates a plan code is
              examined from its entropy source to the value written to the
              store, for the source it draws from and for every input that
              feeds the derivation
Acceptance:   artefact — the code path in `service/` that generates a plan
              code, from its entropy source to the value written to the store;
              property — the code is drawn from a cryptographically secure
              random source, and no sequence, counter, clock reading, host or
              process identifier, or property of the plan it names feeds its
              derivation;
              location — the generator's own source, together with the
              definition site of whatever random source it names
Status:       to-build
Depends-on:   none
Volatility:   open-question — `Architecture.md § What is unproven` records at
              its tenth item that the plan code's length, alphabet and entropy
              are a security decision outstanding and belonging to review, and
              `Architecture.md § Personal data` says the same; what that
              review rules may change what this record can say
Risk:         The plan code is the sole credential for an object holding a
              dwelling coordinate, per `Architecture.md § Personal data`, and
              there is nothing behind it — no account, no password and no
              identity. A code predictable from another code, or from when a
              plan was made, is not a credential at all, and the store has no
              way to tell a guessed retrieval from the user's own.
              `Architecture.md § Personal data` names a short code that is
              both memorable and the sole access control an enumeration target
              for this reason.
Rationale:    The record obliges the property and never a figure:
              `Architecture.md § What is unproven` records the code's length,
              alphabet and entropy as outstanding and belonging to review, and
              naming any of them here would settle a security decision from
              the wrong side. Unpredictability is separable from those
              parameters and is not itself outstanding — it follows from the
              code being the only credential — so it is stated now and the
              parameters are left to the ruling, which NFR-011 obliges a home
              for. The method is inspection because unpredictability is a
              property of the derivation rather than of any run: a generator
              seeded from a clock produces output that passes every
              statistical check a test could run over a sample this system
              will ever produce.
Resolved-by:  —
```

## NFR-011 — Read the plan code's length and alphabet from configuration

```
Statement:    The system shall read a plan code's length and alphabet from
              configuration carrying a documented origin, rather than from a
              literal in the code that generates one, per
              `CalculationSpecification.md § Conventions`.
Category:     Privacy
Source:       Architecture.md § What is unproven;
              Architecture.md § Personal data
Priority:     MUST
Verification: inspection — the configuration entries defining a plan code's
              length and alphabet, and the code that generates one, are
              examined for values read at run time and for the origin recorded
              beside each entry
Acceptance:   artefact — the configuration entries defining a plan code's
              length and its alphabet, and the generator that consumes them;
              property — both values are read from configuration at run time
              and neither appears as a literal in the generator, and each
              entry records a documented origin in the form
              `CalculationSpecification.md § Conventions` requires;
              location — the configuration holding the two entries, and the
              generator's own source in `service/`
Status:       to-build
Depends-on:   NFR-010
Volatility:   open-question — the two values do not exist yet.
              `Architecture.md § What is unproven` records at its tenth item
              that the plan code's length and alphabet are a security decision
              outstanding and belonging to review, and this record obliges
              where that ruling lands rather than what it says
Risk:         `Architecture.md § Personal data` puts the code's length and
              alphabet with review and records the schema as fixing only that
              the column is text. A value chosen inline while that decision is
              outstanding closes it silently — at the moment the generator is
              first written, and by whoever writes it — and the review's
              answer then costs a code change rather than a configuration
              change. Those parameters bound how much guessing the sole
              credential for a dwelling coordinate is worth, so the gap
              between those two costs is the gap between a ruling that can be
              applied and one that is argued about.
Rationale:    The obligation is where the outstanding decision lands, not what
              it decides. `Architecture.md § What is unproven` records the
              ruling as owed to review, so a record naming a length or an
              alphabet would take that decision, and a record obliging nothing
              would leave the generator to take it. It is separate from
              NFR-010 because the two fail independently: a generator drawing
              from a secure source with a length hard-coded beside it
              satisfies unpredictability today and leaves the review's answer
              with nowhere to arrive.
Resolved-by:  —
```

## NFR-012 — Rate-limit plan retrieval per caller

```
Statement:    The system shall limit the rate at which one caller may attempt
              retrieval of a stored plan to the configured per-caller limit,
              counting every attempt whether or not the plan code it names
              exists.
Category:     Privacy
Source:       Architecture.md § Personal data
Priority:     MUST
Verification: test — one caller issuing retrieval attempts faster than the
              configured per-caller limit, naming plan codes the store does
              not hold, is refused once the limit is reached, and a second run
              mixing held and unheld codes is counted against the same limit
Acceptance:   metric — retrieval attempts naming a plan code admitted from one
              caller within the configured window;
              threshold — no more than the configured per-caller limit, with
              every attempt beyond it refused rather than served;
              condition — the attempts come from one caller and name plan
              codes the store does not hold, a second run mixes codes the
              store holds with codes it does not and is counted against the
              same limit, and both the limit and the window are read from the
              deployment's configured values rather than from figures this
              record fixes
Status:       to-build
Depends-on:   none
Volatility:   open-question — `Architecture.md § Personal data` places rate
              limiting on the plan lookup inside the same outstanding security
              decision as the code's length, alphabet and entropy, and no
              document in the repository states a rate; the criterion stands
              on a configured value that does not exist yet
Risk:         `Architecture.md § Personal data` names a short code that is
              both memorable and the sole access control an enumeration
              target, and the object behind it holds a dwelling coordinate.
              With no per-caller limit the only cost of guessing is bandwidth,
              so the code's entropy alone decides how long the store resists —
              and that entropy is itself outstanding, so neither control may
              be assumed to be carrying the other. A successful guess leaves
              no trace distinguishable from use: it is an ordinary retrieval,
              served exactly as the user's own would be.
Rationale:    The record obliges that a limit is enforced and is deliberately
              silent about where.
              `DEPLOYMENT.md § Where the deployment configuration lives` makes
              the reverse proxy a candidate enforcement point, because it is
              the one component that still knows who is asking, and says in
              the same breath that where the throttle is enforced is not
              settled; a record naming the proxy would settle it from the
              wrong side. It is equally silent about the rate, which no
              document states — the criterion is written against the
              deployment's configured value, so it neither fixes a figure the
              review owes nor can be satisfied by one chosen to pass it.
              Counting attempts that name codes the store does not hold is the
              load-bearing half: a limiter counting only retrievals that
              succeed leaves the enumeration it exists to stop entirely
              unmetered.
Resolved-by:  —
```

## NFR-013 — Bound a stored plan's retention by both clocks

```
Statement:    A stored plan shall be retained no longer than the rolling limit
              measured from its most recent retrieval and no longer than the
              absolute limit measured from its creation, both under
              `Architecture.md § Persistence and cross-device transfer`, with
              no retrieval extending the absolute limit.
Category:     Privacy
Source:       Architecture.md § Persistence and cross-device transfer;
              Architecture.md § The plan table
Priority:     MUST
Verification: test — a stored plan retrieved repeatedly at intervals shorter
              than the rolling limit carries a scheduled expiry that never
              passes the absolute limit from its creation, and an attempt to
              move that absolute limit through the path the service writes a
              stored plan by is refused
Acceptance:   metric — a stored plan's scheduled expiry, sampled after each of
              a series of retrievals spaced closer together than the rolling
              limit under
              `Architecture.md § Persistence and cross-device transfer`;
              threshold — never later than the absolute limit that section
              states, measured from the plan's creation, and never later than
              the rolling limit measured from the most recent retrieval;
              condition — the series runs past the point at which the rolling
              limit alone would carry the plan beyond the absolute limit, and
              both limits are read from the deployment's configured values
              rather than from figures this record fixes
              metric — attempts that succeed in moving a stored plan's
              creation time or its absolute expiry;
              threshold — zero;
              condition — attempted through the path the service writes a
              stored plan by, rather than through a privileged path
Status:       to-build
Depends-on:   none
Volatility:   proposed-constant — both intervals under
              `Architecture.md § Persistence and cross-device transfer` are
              proposed defaults rather than measured results, and that section
              says so of the absolute one outright; the ceiling's existence is
              ratified by `Architecture.md § The plan table` and its value is
              not
Risk:         A stored plan holds an origin the user typed, frequently their
              home, under a stable identifier and beside the date it was
              planned — personal data in substance, per
              `Architecture.md § Persistence and cross-device transfer`, that
              cannot be designed out. The rolling limit alone does not bound
              it: that section records that a plan reopened inside the rolling
              window is retained forever, so a habitual user's dwelling
              coordinate is held indefinitely on the strength of habit. The
              absolute limit is also the only thing making the store's worst
              case computable, so losing it costs the sizing as well as the
              guarantee.
Rationale:    The two limbs are one bound and not two records, because neither
              bounds retention alone: the rolling limit is reset by use, and
              the absolute limit is what makes the pair finite — which is why
              `Architecture.md § The plan table` ties them with a constraint
              instead of leaving them independent. This record states the
              upper bound only. What a retrieval does to the rolling clock is
              the reset relation stated in
              `DESIGN.md § Returning to a stored plan`, and it is not restated
              here; the bound holds whatever that relation is. Enforcement is
              cited rather than described: `Architecture.md § The plan table`
              places it in a constraint and a column grant rather than in a
              code path someone can forget, and this record obliges the
              property that arrangement exists to give.
Resolved-by:  —
```

## NFR-014 — Delete a stored plan once its expiry has passed

```
Statement:    The system shall delete a stored plan from the store once its
              expiry has passed, within one period of the sweep schedule under
              `DEPLOYMENT.md § Operational duties`.
Category:     Privacy
Source:       Architecture.md § Persistence and cross-device transfer;
              DEPLOYMENT.md § Operational duties
Priority:     MUST
Verification: test — a store seeded with stored plans whose expiry has passed
              and stored plans whose expiry has not is left to the scheduled
              sweep, and what remains present in it afterwards is counted
Acceptance:   metric — stored plans present in the store whose expiry passed
              more than one sweep period ago, the period being the configured
              cadence under `DEPLOYMENT.md § Operational duties`;
              threshold — zero, with every stored plan whose expiry has not
              passed still present;
              condition — measured after the scheduled sweep has run at least
              once over a store seeded with both, counting rows present rather
              than rows a lookup would return, and with the cadence read from
              the deployment's configured value rather than from a figure this
              record fixes
Status:       to-build
Depends-on:   NFR-013
Volatility:   proposed-constant — the sweep cadence under
              `DEPLOYMENT.md § Operational duties` is a proposed default
              rather than a measured result, and that section states it is the
              period that decides how long an expired plan outlives its own
              expiry
Risk:         `DEPLOYMENT.md § Operational duties` records that the schema
              makes the retention bound unbypassable and deletes nothing by
              itself, so without the scheduled task the bound is a stored
              timestamp and the coordinate stays. It also records why the
              failure is not housekeeping: a sweep that silently stops running
              is a privacy failure, and every signal a reader would check
              reports a healthy store while it holds data it has already
              promised to have deleted.
Rationale:    Deletion is a separate obligation from the bound and not the
              same record's second half. NFR-013 obliges the bound and is
              satisfied by correct timestamps; this record obliges that the
              row goes, and the two fail independently — a store can carry
              impeccable expiry columns and delete nothing, which is precisely
              the limit `DEPLOYMENT.md § Operational duties` records of the
              schema. The criterion counts stored plans present and not stored
              plans readable, which is the distinction a later reader is
              likeliest to collapse:
              `Architecture.md § The queries the schema exists to serve` puts
              expiry in the lookup predicate so that an expired plan is
              unreadable even before the sweep runs, and unreadable is not
              deleted. The overhang is expressed as a relation to the
              configured cadence rather than as a figure, because
              `DEPLOYMENT.md § Operational duties` states that cadence as a
              proposed default and states that the period is what decides the
              overhang. Detecting a sweep that has stopped is deliberately not
              obliged here: that same section names it the first thing
              observability should cover, and observability is owed by
              `Architecture.md § Still owed by this document` and unwritten.
Resolved-by:  —
```

## NFR-015 — Keep the plan code out of every log the reverse proxy writes

```
Statement:    The reverse proxy shall write no plan code into any log it
              produces, whether in the request target it records by default,
              in a structured field, in an error entry or in any other form.
Category:     Privacy
Source:       Architecture.md § Personal data;
              DEPLOYMENT.md § Where the deployment configuration lives
Priority:     MUST
Verification: inspection — the reverse proxy configuration the deployment
              model names is examined for every directive deciding what the
              proxy's logs record, and for the plan code among what they
              record
Acceptance:   artefact — the reverse proxy configuration named by
              `DEPLOYMENT.md § Where the deployment configuration lives`;
              property — nothing the proxy logs carries a plan code, and the
              request-target logging the proxy performs where it is left
              unconfigured is constrained by a directive rather than
              inherited;
              location — the logging directives in that configuration,
              together with the log format each of them names
Status:       to-build
Depends-on:   none
Volatility:   settled — the constraint is stated outright in
              `DEPLOYMENT.md § Where the deployment configuration lives`,
              rests on no figure and appears on no open-questions list, and
              the artefact that discharges it is a recorded debt rather than
              an open question
Risk:         The plan code is the only credential for an object holding an
              origin coordinate that is frequently the user's home, per
              `Architecture.md § Personal data`, and the code travels in the
              request line. A proxy left unconfigured logs the full request
              target, so its access log accumulates every code ever retrieved
              without anyone having decided that it should: this is the half
              of the control that fails by omission rather than by a mistaken
              statement. NFR-008 does not reach it, a proxy's access log being
              no part of what the service process emits. Every reader of that
              log holds the credential, and every copy of the log carries it
              further; nothing observable changes when it happens, and a code
              cannot be un-disclosed afterwards.
Rationale:    `DEPLOYMENT.md § Where the deployment configuration lives`
              splits this control in two and says that a deployment satisfying
              one enforcement point has satisfied half of it. NFR-008 takes
              the service half; this record takes the proxy's, and neither
              reaches the other, because a proxy directive reaches the proxy's
              access log and nothing else while no file under the deployment
              directory can constrain what the service process emits. The
              artefact is named here and not created, on the precedent
              NFR-004's second criterion sets: that section records that the
              directory does not exist yet and that no board item authors it,
              and the criterion is written against the artefact the deployment
              model names rather than against a claim. The method is
              inspection rather than the test NFR-008 uses because the whole
              of what a proxy logs is decided by its configuration, and the
              defect here is an absent directive — which an inspection of the
              configuration sees and a run against a proxy configured for the
              run does not.
Resolved-by:  —
```

## NFR-016 — Record a retrieval attempt without recording the plan code

```
Statement:    The system shall identify a retrieval attempt, in every record
              it keeps of one, by a value from which the plan code that
              attempt named cannot be recovered, rather than by the plan code
              itself.
Category:     Privacy
Source:       Architecture.md § Personal data
Priority:     MUST
Verification: inspection — the path that records a retrieval attempt is
              examined from the plan code the attempt names to every value
              that path writes, for whether any of them is the code or can be
              turned back into it
Acceptance:   artefact — the path that records a retrieval attempt, wherever
              the throttle obliged by NFR-012 is enforced, from the plan code
              the attempt names to every value that path writes;
              property — no value written is the plan code, and none can be
              turned back into it by computation or by any key, table or
              mapping the deployment holds, so an attempt is identified
              without the code being held anywhere the record of an attempt
              reaches;
              location — the derivation producing the identifying value, and
              the definition site of every field the record of an attempt
              carries
Status:       to-build
Depends-on:   NFR-012
Volatility:   open-question — `Architecture.md § Personal data` places rate
              limiting on the plan lookup inside the same outstanding security
              decision as the code's length, alphabet and entropy, so what
              that review rules about how an attempt is throttled and recorded
              may change what this record can say
Risk:         The plan code is the sole credential for an object holding a
              dwelling coordinate, per `Architecture.md § Personal data`, and
              NFR-012 obliges that every attempt is counted whether or not the
              code it names exists — so something holds a record of every
              attempt made. Keyed on the code, that store accumulates every
              code ever presented, the guessed ones beside the real ones, and
              it is the disclosure NFR-008 prevents in the log arriving in a
              store nobody calls a log and nobody thinks to search. Every
              holder of that store holds the credential, and a code cannot be
              un-disclosed afterwards.
Rationale:    The obligation is the property and never a mechanism: no hash,
              algorithm, length or scheme is named, because
              `Architecture.md § What is unproven` records at its tenth item
              that the plan code's parameters are a security decision
              belonging to review, and a mechanism chosen here would pre-empt
              that ruling from the wrong side. The record is equally silent
              about where an attempt is recorded, as NFR-012 is about where
              the throttle is enforced —
              `DEPLOYMENT.md § Where the deployment configuration lives`
              leaves that unsettled, and naming the enforcement point here
              would settle it. The method is inspection rather than test
              because the property is of the derivation and not of any run: a
              search for the code across a limiter's state passes on any
              reversible encoding of it, and leaves the code recoverable by
              whoever holds the store.
Resolved-by:  —
```

## NFR-017 — Read the per-caller retrieval limit and window from configuration

```
Statement:    The system shall read the per-caller retrieval limit and the
              window it is measured over from configuration carrying a
              documented origin, rather than from literals in the code that
              enforces the throttle, per
              `CalculationSpecification.md § Conventions`.
Category:     Privacy
Source:       Architecture.md § Personal data
Priority:     MUST
Verification: inspection — the configuration entries defining the per-caller
              retrieval limit and the window it is measured over, and the
              throttle that enforces them, are examined for values read at run
              time and for the origin recorded beside each entry
Acceptance:   artefact — the configuration entries defining the per-caller
              retrieval limit and the window it is measured over, and the
              throttle that consumes them;
              property — both values are read from configuration at run time
              and neither appears as a literal in the throttle, and each entry
              records a documented origin in the form
              `CalculationSpecification.md § Conventions` requires;
              location — the configuration holding the two entries, and the
              throttle's own definition site, wherever the enforcement point
              turns out to be
Status:       to-build
Depends-on:   NFR-012
Volatility:   open-question — the two values do not exist yet.
              `Architecture.md § Personal data` places rate limiting on the
              plan lookup inside the outstanding security decision it records,
              and no document in this repository states a rate; this record
              obliges where that ruling lands rather than what it says
Risk:         NFR-012's criterion stands on a configured limit and window that
              no document states. Chosen inline while that decision is
              outstanding, the limit is settled silently — at the moment the
              throttle is first written and by whoever writes it — and the
              review's answer then costs a code change rather than a
              configuration change. The limit and the window are what decide
              how long the sole credential for a dwelling coordinate resists
              guessing, per `Architecture.md § Personal data`, and the code's
              own entropy is outstanding beside them, so neither control may
              be assumed to be carrying the other.
Rationale:    The obligation is where the outstanding decision lands, not what
              it decides, and the record names no rate and no window. This is
              NFR-011's shape applied to the second parameter the same review
              owes: that record obliges a home for the code's length and
              alphabet, and the rate is in the same position. It is separate
              from NFR-012 because the two fail independently — a throttle
              enforcing a limit hard-coded beside it satisfies that record's
              criterion today and leaves the review's answer with nowhere to
              arrive. It is silent about the enforcement point for the reason
              NFR-012 is, and it deliberately does not classify the limit as a
              modelling or an enforcement constant under
              `CalculationSpecification.md § Conventions`: that classification
              travels with the ruling that sets the value, and taking it here
              would decide in which direction a deployment may move a figure
              nobody has set.
Resolved-by:  —
```

## NFR-018 — Bound what a hand-off carries to the target application

```
Statement:    A hand-off shall carry to the target application nothing beyond
              the stops of the dispatched portion and that portion's terminal
              points.
Category:     Privacy
Source:       SPECIFICATION.md § No accounts;
              Architecture.md § Persistence and cross-device transfer
Priority:     MUST
Verification: test — a hand-off composed in a session in which the system
              holds the user's Turf username is captured as it is sent, and
              everything it carries is compared against the dispatched
              portion's stops and that portion's terminal points
Acceptance:   metric — values a hand-off carries to the target application
              that are neither a stop of the dispatched portion nor one of
              that portion's terminal points, counted over the whole of what
              is sent rather than over a named list of fields;
              threshold — zero, with the session's Turf username among what is
              counted;
              condition — the hand-off is composed at dispatch time from a
              live session in which the system holds the user's Turf username,
              per `Architecture.md § The user's held zones are already known`,
              and the comparison is run against what leaves the system rather
              than against the stored plan's payload
Status:       to-build
Depends-on:   FR-110;
              FR-118
Volatility:   open-question —
              `Architecture.md § Persistence and cross-device transfer` leaves
              two treatments of the Turf username live, keeping it out of the
              stored object or stating its retention explicitly, and records
              the first as its recommendation rather than as a ruling; a
              ruling for the second reopens whether the username may travel
              with a hand-off
Risk:         A hand-off is composed from live session state rather than from
              the stored plan, and that state holds the user's Turf username —
              NFR-009 keeps it out of the stored payload and reaches nothing
              that leaves.
              `Architecture.md § Persistence and cross-device transfer`
              records the origin coordinate as personal data that cannot be
              designed out and the username as the separable part, and a
              hand-off carrying both is the one path on which the separable
              part is not separated: it delivers a named person and a
              coordinate that is frequently their home in a single request. It
              leaves for a third party, so no retention or access control in
              this system reaches it afterwards, and nothing observable here
              changes when it happens.
Rationale:    The bound is derived rather than invented, and it enumerates
              nothing the documents do not carry:
              `DESIGN.md § Dispatching stop by stop` dispatches as many
              consecutive stops as the target accepts, and
              `SPECIFICATION.md § The waypoint limit problem` carries the
              terminal points as parameters of their own — anything beyond the
              two serves no stated need. It is written as a closed bound
              rather than as a list of things to omit, because a denylist is
              satisfied by leaving out the one field it names and says nothing
              about the next field a hand-off gains. NFR-009 is silent here by
              its own scope and not by oversight, binding the stored payload
              alone. What a Turf stop resolves to on the wire is decided
              elsewhere and is not restated: this record bounds which things
              are sent, and not the form any of them takes.
Resolved-by:  —
```
