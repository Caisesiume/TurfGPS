# Decision log

Every ambiguity `@requirements-engineer` resolved on its own authority rather than asking the Owner, one entry each. The authority is granted by `docs/adr/ADR-0001-artifact-driven-agent-org.md § D6`: ordinary ambiguity is settled under the seven-rung precedence ladder and recorded here, and only a genuinely escalation-qualifying question goes to the Owner. **This file is how that delegation stays visible** — the Owner reads it through `@state-reporter`'s digest, which reports every new entry, non-blocking, as a standing invitation to overturn any of it. A decision made autonomously and never written down is indistinguishable, from the Owner's side, from a decision nobody made.

Records: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## Entry format

One entry per resolved question, under a `## RD-NNN — <the question, in a line>` heading, with a field block laid out as `requirements-authoring § The canonical record` prescribes — values aligned one space past the longest field name below, continuation lines under them, citations kept whole and never split across a line break.

```
Date:           YYYY-MM-DD
Question:       the question as it was actually asked, not as it reads once
                answered
Interpretation: the reading chosen, stated so a later agent can apply it
                without reconstructing the argument
Rung:           N — the ladder rung it rested on, named
Affects:        the records and documents the interpretation binds
```

**`Rung` is the load-bearing field.** It is what lets a later reader check whether the ladder was followed or whether something merely landed somewhere reasonable, and an entry without it records that a decision happened while withholding the only thing the record exists to prove. The seven rungs, in precedence: explicit specification · architecture constraints · design intent · existing requirements · existing system behavior · established repository conventions · most conservative reasonable interpretation.

**`RD-NNN` IDs are immutable and never reused**, on the same rule as `FR-*` and `NFR-*`. Entries are not rewritten to read better after the fact: a decision that no longer holds is superseded by a **new** entry saying so, never by editing the old one, because a log that can be quietly revised is not evidence of anything. `@requirements-librarian` owns this file's structure, as it owns the rest of the corpus's shape; `@requirements-engineer` writes the entries, and neither judges whether a decision was right — that is the Owner's.

## Entries

## RD-001 — Does FR-092 widen by enumerating manufactures, or by obliging establishment?

```
Date:           2026-08-14
Question:       FR-092 refuses only the position on a drivable way nearest the
                zone's coordinate, and was held pending a definition of what
                establishes that a vehicle can be stopped at a position. With
                that definition landed, does the record widen by enumerating
                the manufactures it refuses, or by obliging establishment
                positively?
Interpretation: Positively. The record obliges a stopping position established
                as `SPECIFICATION.md § What establishes a stopping position`
                defines it, and the nearest-drivable-point case becomes one
                instance of non-establishment rather than the rule. An
                enumeration of forbidden manufactures always has one more
                manufacture in it — the form FR-083 was re-keyed away from on
                the safety board's second finding, and the same reasoning
                governs here.
Rung:           1 — explicit specification. The cited section supplies the
                positive predicate: a position is established only where the
                map data carries a feature that is itself a place to stop, and
                no lookup, no default and no road attribute short of the
                feature establishes one.
Affects:        FR-092. By reference FR-068, FR-070, FR-071, FR-072, FR-076,
                FR-085, FR-086, FR-089 and FR-093, each of which names the
                object this record obliges be identified.
```

## RD-002 — Where does a candidate with no established stopping position go?

```
Date:           2026-08-14
Question:       FR-092's third criterion and both of FR-082's read "uncertain
                or excluded".
                `SPECIFICATION.md § Requirements the data cannot verify` left
                that choice open. Does the new section decide it, and does the
                decision reach FR-082, whose antecedent is low access
                confidence generally rather than establishment?
Interpretation: Both narrow to uncertain. Exclusion is left to the records
                carrying the enforceable exclusions — FR-075, FR-076 and
                FR-089 — and neither of these two may reach it. The decision
                does reach FR-082: the new section opens by naming the rule it
                decides, and that rule is the one FR-082's own Rationale
                already cites from the data's side. FR-076 is untouched, its
                exclusion being an enforceable one and the third arm of the
                section's own three-way resolution.
Rung:           1 — explicit specification. The section states that a position
                which is not established is uncertain and that uncertain is
                not excluded, and names the inverted reading as the failure
                that would discard exactly the candidates the reserve pool
                exists to carry.
Affects:        FR-082, FR-092.
```

## RD-003 — What predicate qualifies the stopping position in FR-085 and FR-068?

```
Date:           2026-08-14
Question:       FR-085 obliges a connection from "a legal stopping position";
                FR-068 keys the validation regime to "a valid stopping
                position". Neither qualifier is defined anywhere in the
                corpus, and read as "not marked restricted", legal is unknown
                upgraded to permitted by defaulting.
Interpretation: Both read "established", on the settled predicate. Legal is
                dropped rather than defined: the cited section states outright
                that establishment does not establish that stopping is lawful
                at the moment of arrival, so a record testing lawfulness would
                oblige a finding the data cannot support. Valid is dropped as
                the same predicate under a third spelling. The obligation
                moves in neither record — what changes is that its antecedent
                becomes decidable.
Rung:           1 — explicit specification. The section names "legal stopping
                location" among the five spellings of the one object, governs
                it under every one of them, and separates establishment from
                lawfulness in terms.
Affects:        FR-068, FR-085.
```

## RD-004 — When must a stopping position be identified?

```
Date:           2026-08-14
Question:       FR-092's first criterion obliges a stopping position when
                access classification completes, while FR-070, FR-071 and
                FR-072 test at that position during classification. An
                implementation could therefore evaluate the level check
                against a nearest-way proxy and identify the real position
                afterwards, satisfying every criterion as written and
                reopening the bridge-and-approach hole.
Interpretation: Identification is obliged before any check about the stop is
                evaluated. This is the repair the batch-5 ledger names and
                defers for want of the predicate. The term register's entry
                for `stopping position` already denotes it as the point at
                which every check about the stop is evaluated, so the
                criterion catches up to the register rather than extending it,
                and FR-070's Rationale already names the proxy failure.
Rung:           1 — explicit specification, the settled predicate making the
                obligation writable; rungs 4 and 6 corroborate through the
                term register and FR-070's stated argument, and neither was
                sufficient alone.
Affects:        FR-092. FR-070, FR-071 and FR-072 inherit the ordering through
                Depends-on and are not themselves amended for it.
```

## RD-005 — Is the restricted-feature obligation authored here, or deferred?

```
Date:           2026-08-14
Question:       The new section makes a restriction recorded against a
                stopping feature decisive against establishment and sends the
                position to uncertain, and states that no existing exclusion
                covers it. No record carries this. Author it into batch 5, or
                defer it with a home named?
Interpretation: Deferred, its home being the batch scoped to
                `SPECIFICATION.md § Enforceable exclusions`, alongside the
                four exclusion rules already recorded as owed there. The
                section distinguishes its own restricted-stop paragraph from
                that section's private-and-access-restricted member by
                antecedent, so the two are decided together or not at all.
                Coverage meanwhile is by reference and not by silence: FR-092
                obliges a position established as the cited section defines
                it, and that section refuses a restricted feature, so no path
                to a confident class runs through a restricted stop while the
                debt stands.
Rung:           6 — established repository conventions. This corpus records an
                obligation arising in an unswept section as a debt with its
                home named rather than authoring one paragraph of a section a
                later batch owns; FR-036's companion and the Portability
                country record are the standing instances.
Affects:        The batch-5 debt list in `README.md`. FR-092 carries the
                coverage by citation.
```

## RD-006 — Is the labelling debt still owed in the form batch 5 recorded it?

```
Date:           2026-08-14
Question:       Batch 5 recorded a debt that a stop whose legality could not
                be established must be labelled as such, owed to
                `SPECIFICATION.md § Requirements the data cannot verify`. The
                new section states that this duty is discharged by the
                uncertain classification together with its stated reason, and
                never by annotating a confidently priced recommendation. Does
                the debt survive in that form?
Interpretation: It survives only as the stated reason, and collapses into the
                debt already owed for what an uncertain zone's stated reason
                must name. FR-082 carries the classification and FR-081 the
                recording of it; what remains unowned is the reason's content,
                already owed to the batch scoped to
                `SPECIFICATION.md § Route review and zone confirmation`. No
                separate labelling record is owed, so the debt is re-pointed
                rather than closed. FR-088 is untouched: its duty is material
                uncertainty in an estimate that has already passed the gate,
                which is a different obligation from labelling one that has
                not.
Rung:           1 — explicit specification. The section states which artefact
                discharges the duty and which does not.
Affects:        The batch-5 debt list in `README.md`. FR-088 confirmed
                unchanged.
```

## RD-007 — Does the no-position case get repaired on FR-092's guard, or on FR-076's antecedent?

```
Date:           2026-08-14
Question:       RD-002 routes a candidate with no established stopping
                position to uncertain, guarded "and which no enforceable
                exclusion removes". FR-076 excludes a candidate for which no
                connected walking route between a stopping position and the
                coordinate can be identified — an antecedent an implementer
                reads as satisfied when there is no position at all, which
                hands the case back to exclusion and leaves RD-002 unfailable
                in the only direction that matters. Repair the guard, or the
                antecedent?
Interpretation: The antecedent. FR-076's first criterion now presupposes an
                identified stopping position and cannot fire without one;
                FR-092's guard is kept verbatim. The guard is the source's own
                wording — the section resolves the candidate to uncertain
                where no exclusion applies — so removing it contradicts the
                section, and narrowing it to "except an exclusion keyed on the
                absence of a position" would give FR-076's presupposition a
                second home. After the repair the guard is correct rather than
                vacuous: it still binds where an exclusion fires on a ground
                independent of the position, which the restricted-feature
                record deferred under RD-005 will be.
Rung:           1 — explicit specification; rung 4 corroborates and was not
                sufficient alone. FR-076 already named FR-092 in `Depends-on`
                while its first criterion contradicted that dependency, so the
                record was internally inconsistent and the repair brings the
                criterion up to an edge the record had already declared.
Affects:        FR-076, FR-092. FR-085 and FR-075 carry the same
                vacuous-antecedent shape and neither is hazardous — FR-085's
                consequent is a refusal of the confident class, which is the
                correct outcome for this case, and FR-075's antecedent
                presupposes an identified connection and so fails rather than
                fires.
```

## RD-008 — Must a record oblige the Turf adapter FR-019 inspects, or must FR-019 stop naming it?

```
Date:           2026-08-15
Question:       FR-019 verifies by inspection of "the Turf API adapter" and
                the configuration supplying its base path.
                `Architecture.md § Ports and adapters` declares a `TurfClient`
                port, and no record in this corpus obliges it — or any other
                of that section's six ports — into existence. Raised as DEP-01
                by `@backlog-dependency-planner` against #18. Does a new
                record oblige the port, or does FR-019 re-key onto the first
                story that constructs a Turf request?
Interpretation: Neither. The defect is in FR-019's own Verification and
                Acceptance, and it is repaired there: both re-key from the
                named artefact onto the class of paths that construct a Turf
                API request, identified by what they do rather than by what
                they are called. The Statement does not move, and Depends-on
                stays none — the obligation is over every request the system
                makes, presupposes no other record, and binds whatever the
                code is organized into. Obliging the port is deferred as a
                debt whose home is the batch scoped to
                `Architecture.md § Ports and adapters`, which decides all six
                together; whether `TurfClient` needs a record at all is that
                batch's to weigh, since the section argues from adding a
                country's dataset and Turf has no second provider. Coverage
                meanwhile is by reference and not by silence: the repaired
                record binds every path constructing a Turf request, so none
                escapes the version constraint whether or not the code sits
                behind a port.
Rung:           1 — explicit specification, for the repair.
                `Architecture.md § API version` keys its obligation to the
                requests the system issues and names no artefact at all, so a
                verification keyed to an artefact verifies something the
                source does not oblige. 6 — established repository
                conventions, for the deferral: this corpus records an
                obligation arising in an unswept section as a debt with its
                home named rather than authoring one paragraph of a section a
                later batch owns, which is RD-005's ground. Rung 2 was checked
                and does not decide it — that section ratifies the port, which
                is what makes this a deferral rather than a refusal, but a
                section ratifying a shape is not a record obliging one, per
                the Owner's ruling of 3 August 2026 that scaffolding carries
                no traceability exception. Rung 4 corroborates and was not
                sufficient alone: NFR-001's ruled repair is this exact shape,
                a criterion keyed to an artefact that need not exist reporting
                green while measuring nothing.
Affects:        FR-019. The `FR-019` repair's debt list in `README.md`. By
                reference #18, #20, #21 and #28.
```

## RD-009 — Does #25 author the deployment configuration, or only inspect one?

```
Date:           2026-08-22
Question:       `DEPLOYMENT.md`, in the section stating where the deployment
                configuration lives, assigns #25 four artefacts — a systemd
                unit, a proxy configuration carrying TLS termination, a
                provisioning directory and a README — and calls the first of
                them the file NFR-004's second criterion is satisfied against.
                #25 scopes itself to inspecting a configuration and states
                that no work inside it can produce one. Raised as PLAN-01-A by
                `@backlog-dependency-planner` against #25 and corroborated by
                two documentation reviewers on PR #67, where the attribution
                was additionally traced to #37's note rather than to #25's
                own. Which side is right?
Interpretation: #25 only inspects; `DEPLOYMENT.md` overstates it, and #37's
                note is where the overstatement originates. NFR-004 is the
                whole of that story's traceability, and it obliges an
                exclusion — no deployment target that ends the process between
                requests — rather than an artefact. Three of its fields place
                the artefact's arrival elsewhere: `Verification` reads *once
                one exists*, the second criterion reads *wherever the
                deployment model ... comes to define it*, and `Rationale`
                states outright that the artefact it examines arrives with the
                owed deployment model. An inspection criterion in artefact /
                property / location form names what a reader confirms; it does
                not oblige the artefact into being. The authoring node
                therefore does not exist, and it is deliberately not filed
                here: its Done-when would be defined by that same section of
                `DEPLOYMENT.md`, unmerged on PR #67, and it would be
                constrained by #109's security architecture section,
                unwritten. Recorded instead as a debt with its home named. #25
                and #37 are repaired to state this. Two claims sit in the
                section of `DEPLOYMENT.md` stating where the deployment
                configuration lives, and this ruling condemns one of them. The
                authoring attribution falls: that the `deploy/` directory is
                created by the first deployment work and that work is #25, and
                that the proxy configuration is one #25 writes. The proxy
                attribution shares its sentence with the only clause binding
                anything to satisfy *no host log may record a plan code*, so
                it is edited out of that sentence and the sentence is not
                deleted; the obligation stands, bound to the proxy
                configuration as before. The referent identification stands,
                and is endorsed here rather than merely left alone: naming
                `deploy/turfgps.service` as the file NFR-004's second
                criterion is satisfied against is that criterion operating as
                written, because the criterion defers to wherever the owed
                deployment model comes to define the deployment configuration.
                Revising the authoring attribution is PR #67's, on its own
                PLAN-01-A; the referent identification is not #67's to delete.
                No upstream document is edited here.
Rung:           1 — explicit specification, and it cannot decide alone.
                `Architecture.md § Still owed by this document` owes the
                deployment model handed to `DEPLOYMENT.md`, so nothing
                upstream obliges the four artefacts into existence or assigns
                authoring of them to a story. That forecloses the enlarged
                scope without settling the narrower one: an upstream silence
                about who authors an artefact does not by itself state what a
                story does, and what #25 does is settled by the record #25
                resolves. 4 — existing requirements, decisive on that.
                NFR-004's own `Verification`, second criterion and `Rationale`
                each place the artefact with the owed deployment model, and a
                record signed to-build is the corpus's settled statement of
                its own scope. 6 — established repository conventions, twice:
                `docs/README.md § DEPLOYMENT.md` defines that document as
                operational detail and direction rather than as one of the
                four approved specification documents, so the document map
                corroborates at this rung and never at rung 1; and the
                deferral of the authoring node rests here, on RD-008's ground.
                Rung 7 was checked and agrees rather than deciding, the
                conservative reading being the narrower scope. Corroborated
                structurally, and not sufficient alone: three of the four
                artefacts carry security obligations NFR-004 is silent on.
                `deploy/turfgps.service` falls under the boundary that no
                secret material lives under `deploy/`, stated in the section
                of `DEPLOYMENT.md` naming where the deployment configuration
                lives, and is the file an `Environment=` line would commit a
                database URL or the proxy's TLS private key into.
                `deploy/proxy/` carries TLS termination, the constraint that
                no host log record a plan code, and the caller-address
                preservation without which a per-caller throttle cannot be
                written. `deploy/provisioning/` carries the exposure
                invariant. The fourth, `deploy/README.md`, carries apply order
                and no such obligation. The unit file is the strongest
                instance and not merely the gravest. NFR-004's second
                criterion names no file: it defers to the deployment
                configuration, which `DEPLOYMENT.md`'s standing referent
                identification resolves to this artefact. Through that
                identification the criterion asks only that the configuration
                name a target keeping the process running, so it reports green
                on a unit whose `Environment=` line carries the credential. A
                story authoring the three would close against a requirement
                that cannot fail on any of them — the defect shape RD-008
                ruled, a criterion keyed to an artefact reporting green while
                measuring nothing.
Affects:        NFR-004, which is unrepaired and stands as written, on three
                grounds and none of them namedness. That axis does not
                separate it from the FR-019 defect, whose old criterion was
                keyed to *the Turf API adapter* — the same definite singular,
                naming no file — and RD-008's predicate turns on the
                artefact's possible absence, which rung 1 above establishes
                for this artefact too. What separates them is this. Its second
                criterion fails closed where FR-019's failed open: it asks
                that a configuration be read and a target be named in it, so
                an absent artefact yields no evidence and no green, where
                FR-019's single criterion had nothing to examine and passed
                vacuously. Its first criterion gives the exclusion independent
                coverage, binding an entry point and a session registry in a
                main package that exists and is obliged to, where FR-019
                carried no second criterion at all. And there is no coverage
                escape: absent a deployment configuration there is no
                deployment, so the obligation has no instance to miss, where a
                path constructing a Turf request existed whether or not
                anything was called an adapter. The security obligations
                `Rung` enumerates are a corpus coverage gap, owned by
                PLAN-01-B on #109, whose Done-when covers all three artefacts
                and gates #25. The debt list in `README.md`. By reference #25,
                #37, #109 and PR #67.
```

## RD-010 — What in RD-009 did PR #67 falsify, now that it has merged?

```
Date:           2026-08-28
Question:       RD-009 was written on 22 August 2026 against `NFR-004` and
                `DEPLOYMENT.md` as they then stood. PR #67 merged at `030d9e3`
                and rewrote both. `@pr-judge` ruled DOC-06 on cycle 4 — two of
                RD-009's verbatim quotations go stale on merge, the staleness
                does not gate merge, and a tracked issue is owed before it —
                and folded DOC-14 and `@validation-agent`'s `:327-328` hint
                into the same repair. #126 tracks it.
                `DECISIONS.md § Entry format` forbids repairing an entry in
                place, so the compliant fix is this entry. Which of RD-009's
                statements are now false, and does its ruling survive them?
Interpretation: **RD-009's ruling stands entirely and is untouched.** #25
                inspects the deployment configuration and authors none of it;
                the authoring node is still a debt with its home named;
                `NFR-004` is still unrepaired and still stands as written. The
                new text supports that ruling more directly than the wording
                it was argued from, so nothing decided is disturbed and no
                record moves. **Four statements inside RD-009 are superseded,
                and a reader meeting any of them reads this entry instead.**
                First, RD-009's Interpretation says `NFR-004`'s second
                criterion reads *wherever the deployment model ... comes to
                define it*. As merged it reads *as the deployment model
                defines it in*
                `DEPLOYMENT.md § Where the deployment configuration lives`.
                The deferral RD-009's endorsement rests on survives the
                rewording and is if anything plainer, the criterion now naming
                the section it defers to. Second, RD-009 says that criterion's
                `Rationale` states the artefact it examines *arrives with the
                owed deployment model*. As merged it states that the
                deployment model in that same section *names the artefact it
                examines and does not create it* — which is RD-009's own
                conclusion, now asserted by the record rather than inferred
                from it. Third, RD-009's `Rung` says `deploy/proxy/` carries
                the constraint that no host log record a plan code. After the
                SEC-15 split it carries **the proxy half only**:
                `DEPLOYMENT.md § Where the deployment configuration lives`
                gives that artefact *the caller-address constraint whole, and
                the proxy half of the log constraint*, and assigns the other
                half to whoever writes the log statements in `service/`. That
                half is what NFR-008 now obliges, filed this batch. Fourth,
                RD-009's `Rung` cites
                `Architecture.md § Still owed by this document` as owing the
                deployment model handed to `DEPLOYMENT.md`. That section no
                longer owes it — the model *left this list on 14 August 2026*,
                and the heading survives while the claim does not, which is
                the rot shape `docs/README.md § Conventions` names. RD-009's
                conclusion at that rung is unaffected: nothing upstream
                obliged the four artefacts into existence when it was written,
                and the section that has since arrived does not oblige them
                either. **RD-009's other two quotations still resolve and are
                not superseded** — `NFR-004`'s `Verification` still reads
                *once one exists*, and *no host log may record a plan code*
                resolves verbatim, and for the first time resolvably, to
                `DEPLOYMENT.md § Where the deployment configuration lives`.
                Every quotation above was checked against `main` at `5ccd75d`
                rather than against the branch that wrote it.
Rung:           6 — established repository conventions, decisive on the form.
                `DECISIONS.md § Entry format` makes an entry immutable and
                supersedes by a new entry rather than an edit, so the
                `Architecture.md § D8` precedent of retracting in place is
                unavailable here by rule and not by preference; and
                `docs/README.md § Conventions` puts the repair of a rotted
                citation on whoever reads it once the editing commit has
                passed. 4 — existing requirements, decisive on the substance:
                `NFR-004` as merged is what the entry must now be read
                against, and its second criterion and `Rationale` each carry
                RD-009's conclusion more plainly than the text RD-009 quoted.
                Rung 1 was checked and reaches nothing: no upstream document
                decides what a decision log says about itself.
Affects:        RD-009, superseded in the four statements above and in no
                other. `NFR-004`, unchanged and unrepaired. NFR-008, which
                takes up the service half of the log constraint the third
                statement misdescribed. By reference #126, #25, #37, #109 and
                PR #67.
```

## RD-011 — What is the persisted object called, and what is the key that retrieves it?

```
Date:           2026-08-28
Question:       Batch 6 obliges thirty things about one persisted object and
                its retrieval key. The upstream documents carry several nouns
                for each, sometimes in one paragraph:
                `SPECIFICATION.md § Route persistence` opens *Confirmed routes
                persist*, says *A stored route keeps everything*, and then
                says *The stored plan does not change*.
                `Architecture.md § Persistence and cross-device transfer`
                writes *short code* and *the code*;
                `DESIGN.md § Returning to a stored plan` writes *The retrieval
                code*. Which nouns does the corpus write?
Interpretation: The persisted object is a **stored plan**. Its retrieval key
                is a **plan code**. Both are registered in `README.md` and are
                written verbatim in every field of every record. This is the
                term register operating on its second case, and the case is
                stronger than the first: *route* and *plan* are not merely two
                spellings but two readings, one naming the geography and one
                naming the whole stored object with its candidate set and
                costs — and FR-096 and FR-106 oblige different things about
                those two readings. `plan` is chosen over `route` because
                `Architecture.md § The plan table` is the settled design and
                uses it throughout, down to the column names; `plan code` over
                `short code` because the code's length is an open security
                decision under `Architecture.md § Personal data` and a noun
                asserting shortness would pre-empt it. The upstream documents
                are not bound and keep their own words, and a record quoting
                one of them in reported speech keeps the source's word.
Rung:           6 — established repository conventions, decisive. The term
                register in `README.md` exists for exactly this, and
                `requirements-authoring § The term register` states that a
                second noun for one object is a decision rather than a
                synonym. 2 corroborates on which noun:
                `Architecture.md § The plan table` is the settled artefact and
                it says *plan*. Rung 1 was checked and cannot decide —
                `SPECIFICATION.md § Route persistence` uses both nouns for the
                same object in adjacent sentences, which is what raised the
                question.
Affects:        The term register in `README.md`. Every record in batch 6,
                FR-094 through FR-116 and NFR-008 through NFR-014. No filed
                record is renamed: nothing before this batch obliges anything
                about the stored plan.
```

## RD-012 — May the term register hold a quantity, or only a physical object?

```
Date:           2026-08-28
Question:       The set-level consistency pass found FR-110 sizing a
                dispatched portion against *the target's configured waypoint
                allowance* while FR-111 obliged one stop more than *the
                target's intermediate-waypoint allowance*. Two noun phrases,
                and nothing in either record said whether they name one
                quantity or two — so on every portion FR-111 reached, one
                record obliged N stops and the other N+1. The register is
                scoped to *physical objects several records oblige things
                about*, and a waypoint allowance is not one. Does it reach
                this?
Interpretation: It does, and the register's scope is widened to **an object or
                a quantity**. `intermediate-waypoint allowance` is registered
                and denotes the target's published cap on intermediate stops,
                counting neither the origin nor the destination. FR-110 and
                FR-111 are rewritten to write it verbatim, with FR-110
                carrying the arithmetic and FR-111 stating only what the
                destination slot holds. The widening is minimal and
                deliberate: the register's own argument — that a duplicate
                noun costs an *obligation* where a duplicate category costs
                only coverage checking — is about what two records can be made
                to oblige, and nothing in it turns on the referent being
                physical. This case is the proof: both records read correctly
                alone, cited their sources faithfully, and could not both be
                satisfied. The register stays short and the bar is unchanged —
                a noun earns a row only once several records oblige something
                about the same referent and a second spelling has appeared.
                Here it had appeared, in the same batch, between two adjacent
                records.
Rung:           6 — established repository conventions, decisive. The
                register's stated reason in `README.md` and in
                `requirements-authoring § The term register` is that two nouns
                for one referent produce two obligations from one intent, and
                that reason is indifferent to whether the referent is
                physical. 7 corroborates: widening the scope by one word is
                the narrower repair against the alternative, which is to leave
                the register alone and fix the wording in the two records —
                the same choice that was available for `stopping position` and
                that left three nouns running for a year.
Affects:        The term register in `README.md`, both its scope sentence and
                its new row. FR-110 and FR-111, rewritten. The upstream
                documents are unbound:
                `SPECIFICATION.md § The waypoint limit problem` keeps
                *waypoints* and *intermediate stops* as it writes them.
```

## RD-013 — Does a stored plan keeping everything include the ownership it was planned with?

```
Date:           2026-08-28
Question:       `SPECIFICATION.md § Route persistence` says a stored route
                keeps **everything** and enumerates the candidate set, the
                access classifications, the computed costs and the
                additional-time allowance. `Architecture.md § Round rollover`
                says the plan payload stores **no ownership** and that the
                indicator is recomputed on every open. Read together, does
                *everything* include zone ownership, and if not, is this a
                contradiction between two documents?
Interpretation: It is a refinement and not a contradiction, and no escalation
                is owed. `SPECIFICATION.md § Stored routes go stale` draws the
                line itself, in the same document and about the same object:
                what a stored route keeps is the plan and the state behind it,
                and what *decays* is the volatile overlay — ownership, current
                points, and any activity-derived estimate — which refreshes
                rather than persisting. *Everything* is therefore everything a
                later re-solve needs, and ownership is not state a re-solve
                needs. FR-096 obliges the enumeration and FR-098 obliges the
                one exclusion, each naming the other so that a reader meeting
                either does not read it as the whole rule.
Rung:           1 — explicit specification, decisive. The distinction is drawn
                inside `SPECIFICATION.md § Stored routes go stale` and needs
                no inference: the roads and zones remain the plan, the
                volatile data attached to it decays. 2 corroborates on the
                mechanism: `Architecture.md § Round rollover` gives the reason
                a plan spans on the order of a dozen rounds and anything
                round-scoped inside it is wrong for most of its life.
Affects:        FR-096, FR-098, FR-106, FR-107. NFR-009 and NFR-013, which
                bound what else a stored plan may hold and for how long.
```

## RD-014 — Is omitting the Turf username an obligation, or one of two treatments?

```
Date:           2026-08-28
Question:       `Architecture.md § Persistence and cross-device transfer`
                leaves two treatments of the Turf username live — keep it out
                of the stored object, or state its retention explicitly — and
                records the first as *simpler* and *the recommendation*. A
                recommendation is not a ruling. May the corpus oblige the
                omission?
Interpretation: It may, and it does. `Architecture.md § Personal data` adopts
                the recommendation rather than restating the choice: it says
                there is no column for the username and gives the payload a
                constraint refusing one. A design that has built the
                enforcement has taken the decision. NFR-009 therefore obliges
                the absence outright, and the residual openness is carried in
                its `Volatility` rather than by weakening the statement to
                *should*. **The obligation is on the payload and not on the
                constraint**, which is the load-bearing half:
                `Architecture.md § Personal data` states that constraint's own
                limit — it tests top-level keys only, so a nested username
                passes it, and it is *a tripwire against the obvious mistake,
                not a proof of absence*. A criterion satisfied by the
                constraint's existence would report green on exactly the case
                the constraint cannot see.
Rung:           2 — architecture constraints, decisive. The schema is where
                the recommendation became a decision, and it did so by
                building the refusal rather than by arguing for it. Rung 1
                reaches nothing: `SPECIFICATION.md § No accounts` excludes
                identity without naming the username. Rung 7 agrees rather
                than deciding — not storing a separable identifier is the
                conservative reading.
Affects:        NFR-009. FR-096, whose enumeration is bounded by it.
```

## RD-015 — May an expired plan code and an unrecognised one be told apart from outside?

```
Date:           2026-08-28
Question:       `DESIGN.md § Returning to a stored plan` says *A code that has
                expired or is not recognised must produce a plain, unalarming
                explanation and an obvious way to start a new plan, rather
                than an error*. Does that oblige one response for both cases,
                or two responses each of which is plain?
Interpretation: One response, indistinguishable from outside. The sentence
                gives both causes one treatment and names no difference
                between them, and there is nothing a user could do with the
                distinction: in both cases the plan is not there and the way
                on is the same. The security reading points the same way and
                is why the ruling is worth recording rather than left to an
                implementer's taste — two distinguishable responses make the
                store an oracle for which plan codes exist, and
                `Architecture.md § Personal data` names a short code that is
                the sole access control an enumeration target for an object
                holding a dwelling coordinate. FR-103 obliges the
                indistinguishability; what the explanation says and where it
                appears stays `DESIGN.md`'s.
Rung:           1 — explicit specification, decisive: one sentence, one
                treatment, two causes, no distinction drawn. 7 corroborates
                and does not decide — the conservative reading of an ambiguous
                disclosure rule is the one that discloses less. Rung 2 was
                checked: `Architecture.md § Personal data` supplies the motive
                but obliges no response shape.
Affects:        FR-103. NFR-012, which bounds the rate at which the oracle
                could be questioned even were it answering.
```

## RD-016 — May the corpus oblige anything about the plan code while its length is open?

```
Date:           2026-08-28
Question:       `Architecture.md § Personal data` says the code's *length,
                alphabet and entropy are a security decision rather than a
                schema one*, outstanding and belonging to review, and
                `Architecture.md § What is unproven` records the same at its
                tenth item. Is the whole subject closed to the corpus until
                that review runs?
Interpretation: No. Two things are separable and only one is open. **That the
                code must be unpredictable is not outstanding** — it follows
                from the code being the sole credential for an object holding
                a dwelling coordinate, which the same section states as
                settled fact, and a code derived from a sequence, a counter or
                a clock is not a credential at any length. NFR-010 obliges it
                now, by inspection of the derivation rather than by any test
                over a sample. **What the review owes is the parameters**, and
                NFR-011 obliges only that they have somewhere to land — read
                from configuration with a documented origin per
                `CalculationSpecification.md § Conventions` rather than
                written inline. No figure, bit count or alphabet appears in
                either record. This is deliberately not an escalation: the
                security *intent* is determinable and stated, and only the
                parameter is open, so the §21 condition about undeterminable
                security intent is not met.
Rung:           2 — architecture constraints, decisive. The sole-credential
                premise and the enumeration-target reading are both stated in
                `Architecture.md § Personal data` as facts about the design,
                not as questions. 6 corroborates on the second record:
                `CalculationSpecification.md § Conventions` already requires
                every constant to be configurable and to carry a documented
                origin, so obliging a home for the ruling restates no
                decision. 7 decides the boundary — obliging the property and
                refusing the figure is the narrowest reading that leaves the
                review its decision.
Affects:        NFR-010, NFR-011. FR-100, whose `Rationale` parks the code's
                strength here rather than asserting it.
```

## RD-017 — Whose obligation is *no host log may record a plan code*, and how far does it reach?

```
Date:           2026-08-28
Question:       PR #67 left SEC-07 open:
                `DEPLOYMENT.md § Where the deployment configuration lives`
                states *No host log may record a plan code — not the proxy's,
                not the service's*, and no requirement stood behind it.
                `@backlog-dependency-planner` classified it as the second
                instance of the class where a document asserts a property of
                code that nothing obliges. What does the corpus owe, and what
                does it cite, given `DEPLOYMENT.md` is not one of the four
                approved specification documents?
Interpretation: The corpus owes **the service half only**, as NFR-008. That
                section splits the control itself and says it is *where the
                constraint is derived rather than where it is discharged*: the
                proxy half is a deployment artefact bound to `deploy/proxy/`,
                and the other half binds whoever writes the log statements in
                `service/`, which no file under `deploy/` can constrain.
                **`Source` names `Architecture.md § Personal data` first and
                `DEPLOYMENT.md § Where the deployment configuration lives`
                second.** Architecture creates the obligation — the code is a
                plan's only credential and the plan holds a dwelling
                coordinate — and `DEPLOYMENT.md` states it; RD-009 already
                established that the document map puts `DEPLOYMENT.md` at rung
                6 and never at rung 1, so it may be cited but may not carry a
                requirement alone. **The reach is everything the service
                process emits, not only the statements written in
                `service/`.** That section records that the service installs
                no log handler, so its own output and a dependency's land in
                the same journal, and an inspection of the statements in
                `service/` would pass over the second — which is why NFR-008
                is verified by `test` against the whole of a run's output.
Rung:           2 — architecture constraints, decisive on both the source and
                the reach. 6 corroborates on the citation order, on RD-009's
                own ground: `docs/README.md § DEPLOYMENT.md` defines that
                document as operational detail and direction rather than as
                one of the four approved specification documents. Rung 1 was
                checked and reaches nothing — no approved document states the
                log constraint.
Affects:        NFR-008, which closes SEC-07. RD-009, whose `Rung`
                misdescribed the proxy's share of this constraint and is
                superseded on that statement by RD-010.
```

## RD-018 — What resets a stored plan's rolling clock, and what does its expiry do?

```
Date:           2026-08-28
Question:       `Architecture.md § The plan table` says the rolling period is
                *reset on every open*, and
                `DESIGN.md § Returning to a stored plan` says *Reopening a
                plan restarts its expiry clock*. Neither says what counts as
                an open over an interface whose only verb is retrieval by
                code. And once a plan's expiry has passed, is the plan gone or
                merely unreadable?
Interpretation: **The clock resets on a successful retrieval of the stored
                plan by its plan code**, and on nothing else. A background
                refresh of volatile data is not an open, and opening a stored
                plan from the device's own list is one —
                `Architecture.md § Persistence and cross-device transfer`
                holds the plan code in local storage precisely so the usual
                case never requires the user to see it, which makes opening
                from the list a retrieval by code. The enumeration worry this
                raises — that a caller hitting a valid code extends its life —
                is answered by NFR-012's throttle and by the ceiling, not by
                refusing the reset, which would reintroduce the loss the reset
                exists to prevent. **Expiry means the stored plan is deleted,
                not made unreadable.** `DEPLOYMENT.md § Operational duties`
                says the schema makes the bound unbypassable and *does not
                delete anything by itself*, and
                `Architecture.md § The queries the schema exists to serve`
                already makes an expired plan unreadable through the lookup
                predicate before any sweep runs — so unreadable cannot be what
                the retention obligation buys. NFR-014 therefore counts rows
                present, not rows a lookup would return.
Rung:           2 — architecture constraints, decisive on both halves: the
                reset relation and the two columns are stated in
                `Architecture.md § The plan table`, and the local-storage
                argument that makes a list open a retrieval is stated in
                `Architecture.md § Persistence and cross-device transfer`. 1
                corroborates on deletion — that section says a plan *should be
                deleted* at the ceiling, which unreadability does not satisfy.
Affects:        FR-104, NFR-013, NFR-014. FR-103, which is what a user meets
                once either clock has run out.
```

## RD-019 — Is opening a stored plan gated on the initialization wizard?

```
Date:           2026-08-28
Question:       `DESIGN.md § First-run initialization` says *Only when both
                steps are complete does the planner open*.
                `DESIGN.md § Never gate stored plans on the wizard` says *A
                returning user with stored values must be able to open
                existing plans while the Turf API is unavailable* — which is
                scoped to a user who has already completed it. Is a user
                opening a plan by its plan code on a device that has never run
                the wizard inside the gate or outside it?
Interpretation: Outside it. The same section states the intent in terms that
                decide the case it did not name: *The gate exists to stop
                someone planning without the data that makes a plan good. It
                must not stop someone reading a plan they already made.* A
                plan retrieved by its plan code on a second device is a plan
                they already made, and reading it is not planning. FR-105
                therefore obliges the open without the wizard and without the
                Turf API, presenting the plan with the volatile overlay absent
                rather than withholding it — which is the degradation that
                section already licenses, applied to the case where the
                username is unknown rather than to the case where the API is
                down. The gate on *planning* is untouched.
Rung:           3 — design intent, decisive. The section states its own
                purpose and the statement of purpose reaches the unnamed case
                where the scoped sentence does not. 1 corroborates:
                `SPECIFICATION.md § Route persistence` makes the failure being
                prevented a user arriving on the morning of departure to find
                the work gone, and a wizard lockout on a borrowed device
                produces it exactly.
Affects:        FR-105, FR-107. NFR-009, whose omission of the username is
                what makes a wizard-less open coherent.
```

## RD-020 — Does the corpus oblige the stored-plan list's ordering?

```
Date:           2026-08-28
Question:       `DESIGN.md § Returning to a stored plan` says stored plans are
                *listed, most recent first, each identified by its origin and
                destination and the date it was planned*. Both halves are
                testable. Does the corpus oblige both?
Interpretation: The identifying content binds; the ordering does not. What a
                user must be told to tell one stored plan from another is a
                requirement — FR-102 obliges the origin, the destination and
                the date planned, which the source calls *the two things a
                user actually remembers*. The sort order is presentation, and
                the corpus has already drawn this line once: the Owner's
                ruling of 1 August 2026 on FR-010 put the obligation on the
                offered *set* and never on presentation order, and
                `Recommendation set composition` disclaims presentation in its
                own scope line. Obliging *most recent first* here would
                contradict that line in the act of using it, and would freeze
                into the corpus a choice `DESIGN.md` is free to revise.
Rung:           4 — existing requirements, decisive. FR-010 and the ruling
                behind it already settled that ordering is not the corpus's,
                and applying the settled line is cheaper and safer than
                re-drawing it. 6 corroborates: the category register states
                the same boundary on two rows. Rung 1 was checked and cannot
                decide — `DESIGN.md` states both halves in one sentence and
                marks neither.
Affects:        FR-102.
```

## RD-021 — How much of the round-rollover treatment may harden into a requirement?

```
Date:           2026-08-28
Question:       `DESIGN.md § Communicating a round rollover` specifies a
                non-blocking banner, an optional *Re-check zones* action, and
                the wording of the message.
                `DESIGN.md § Open questions owned by this document` lists
                round-rollover messaging as *Proposed, and open to revision*.
                What may the corpus oblige?
Interpretation: The obligation, not the treatment. FR-108 obliges that a
                rollover the system has determined is told to the user before
                they act on the plan, that the message states the route is
                unchanged, and that any refresh offered is optional rather
                than an automatic edit. The banner, its placement and its
                wording are not obliged and the record carries
                `Volatility: open-question` naming the entry that keeps them
                open. This is override 2's shape applied to an interaction
                rather than to a constant: hardening a proposal makes the
                record read better and falsifies it the day the proposal
                moves. The two clauses that are **not** proposals and do bind
                are the ones the section argues from rather than proposes —
                that a rollover is not presented as a problem, every zone in
                the plan now being unclaimed, and that the route's survival is
                stated plainly.
Rung:           1 — explicit specification, decisive on what binds:
                `SPECIFICATION.md § Stored routes go stale` obliges that
                anything material is surfaced as information and never as an
                automatic edit, independently of how `DESIGN.md` proposes to
                surface it. 6 decides the boundary —
                `docs/README.md § Conventions` makes a proposal a position to
                argue against, and
                `safety-path-checklist § The proposal boundary` forbids
                quoting one as established.
Affects:        FR-108. FR-106 and FR-107, which carry the
                never-an-automatic-edit and refresh-after-presentation halves
                of the same source rule.
```

## RD-022 — Is code-keyed server-side plan storage decided, or still proposed?

```
Date:           2026-08-28
Question:       `Architecture.md § D4` reads **Decided** and says plans need
                *ordinary transactional storage keyed by a short code*.
                `Architecture.md § Persistence and cross-device transfer`
                calls the same mechanism *the proposed resolution*, and
                `Architecture.md § Open questions owned by this document`
                still lists it as *Proposed, and open to revision*. One
                document says both. May the corpus oblige retrieval by code at
                all?
Interpretation: It may, and the records rest on it as `open-question` rather
                than as `settled`. **The obligation stands**: `D4` is a
                ratified decision, and `Architecture.md § The plan table`
                builds concrete DDL on it,
                `DESIGN.md § Returning to a stored plan` specifies its
                interaction as settled, and
                `DEPLOYMENT.md § Operational duties` schedules its sweep. Four
                places build on the mechanism, which is more than a proposal
                carries. **The volatility does not follow the obligation**:
                the open-questions entry says *open to revision*, not
                *undecided*, and a record resting on something a document
                lists as open is by definition `open-question` under
                `requirements-authoring`. FR-100, FR-101 and FR-103 were
                corrected from `settled` on this ruling. This is **not** an
                escalation under §21: that condition names two authoritative
                documents contradicting each other, and this is one document
                inconsistent with itself, on a point that changes no scope
                either way. It is raised upstream as a finding instead.
Rung:           2 — architecture constraints, decisive. Between two statements
                in one document, the one carrying a ratified decision marker
                and three downstream artefacts built on it governs the one
                carrying neither. 6 corroborates on the volatility:
                `requirements-authoring` defines `open-question` as resting on
                something a document lists as open, which this is, whatever
                the obligation's standing.
Affects:        FR-100, FR-101, FR-103, and FR-104 by subject though its
                `Volatility` names the sooner-falsifying proposals instead.
                Raised upstream against `Architecture.md`.
```

## RD-023 — Is the time a plan was computed the same instant as the plan's creation?

```
Date:           2026-08-28
Question:       Found by the set-level consistency pass. FR-099 obliges
                recording *the time at which the plan was computed*, from
                `SPECIFICATION.md § Stored routes go stale`. NFR-013 bounds
                retention by an absolute limit *measured from its creation*,
                from
                `Architecture.md § Persistence and cross-device transfer`.
                Both read correctly alone. Are they one instant?
Interpretation: They are two, and neither record may be satisfied by the
                other's. A plan is computed, then reviewed — possibly at
                length — then confirmed and stored, so the computation instant
                precedes the storage instant by an unbounded interval.
                FR-099's instant answers *how stale is what this plan was
                built from*, which is why
                `SPECIFICATION.md § Stored routes go stale` asks for it;
                NFR-013's answers *how long has this object been held*, which
                is what a retention ceiling bounds. Collapsing them is the
                plausible mistake because one column looks like it would serve
                both, and it fails in the privacy direction: a ceiling run
                from the computation instant expires the object early and
                harmlessly, while one run from a stored computation time that
                is later than creation would hold a dwelling coordinate past
                the ceiling. Recorded rather than repaired, because neither
                record is wrong — what was missing was anything saying they
                are different, and this entry is it.
Rung:           2 — architecture constraints, decisive.
                `Architecture.md § The plan table` keeps `created_at` and
                `last_opened_at` as separate columns and runs the ceiling from
                creation, so the schema already distinguishes the instants
                this corpus was about to merge. 7 corroborates on which to
                prefer if a build ever must choose — the earlier instant,
                which expires sooner.
Affects:        FR-099, NFR-013. FR-108 and NFR-002, both of which measure an
                age from the computation instant rather than from creation.
```

## RD-024 — At what aggregation point is plan retrieval throttled?

```
Date:           2026-08-28
Question:       `Architecture.md § Personal data` names *rate limiting on Q3*
                as part of the answer to a short code being an enumeration
                target, and states no rate and no aggregation point. Per
                caller, or across the system? NFR-001 is the precedent for the
                question mattering.
Interpretation: Per caller.
                `DEPLOYMENT.md § Where the deployment configuration lives`
                obliges the proxy to preserve the caller's address precisely
                because *no throttle can be written against a single
                indistinguishable caller*, which is only coherent if the limit
                is per caller. NFR-012 obliges it there and is deliberately
                silent about **where** it is enforced, that same section
                making the proxy a *candidate* enforcement point and saying
                outright that the question is not settled. The rate itself is
                named nowhere in the repository, so the criterion is written
                against the deployment's configured value and no figure is
                invented — which is raised as a gap rather than closed.
                **Every attempt counts, whether or not the plan code exists**:
                a limiter counting only successful retrievals leaves the
                enumeration it exists to stop entirely unmetered. This is the
                mirror of NFR-001, which names its aggregation point for the
                opposite reason — there the limit is the other party's and
                applies to the system as a whole, here it is ours and applies
                to each caller.
Rung:           2 — architecture constraints, decisive. The caller-address
                obligation exists for this throttle and states its own reason,
                which fixes the aggregation point without fixing the rate. 4
                corroborates by contrast rather than by analogy: NFR-001's
                system-wide point was chosen because the ceiling is external,
                and neither reason applies here.
Affects:        NFR-012. FR-100, whose first criterion is scoped inside this
                limit.
```

## RD-025 — What triggers a hand-off fallback, and where does the reduction stop?

```
Date:           2026-08-28
Question:       `DESIGN.md § Dispatching stop by stop` says the dispatch path
                must degrade gracefully — *if a hand-off is rejected or
                truncated, the system falls back to sending fewer stops rather
                than failing outright*. But
                `SPECIFICATION.md § Waypoints may be dropped without warning`
                says a truncation is undetectable: the parameter is ignored
                rather than rejected, and the system cannot tell. So what
                fires the fallback, and what happens when a single-stop
                hand-off is still refused?
Interpretation: **The trigger is an observable refusal only.** Reading
                *truncated* as a trigger would oblige a detection the sources
                say does not exist, and a requirement that cannot fire is
                worse than none — it reports green while the user drives away
                without their stops. The undetectable case is carried as
                **disclosure** instead: FR-113 obliges that the user is told
                before the hand-off what it may drop, which is the stance
                `SPECIFICATION.md § The waypoint limit problem` itself takes,
                applying the confidence-and-uncertainty rule to a hand-off
                rather than to an estimate. **The reduction terminates by
                abandoning**, with the stored plan unchanged and the user told
                the hand-off did not complete. The source does not reach this
                branch; silently retrying forever, or leaving the plan
                half-dispatched, are both worse than saying so, and FR-114
                obliges the plan's integrity independently.
Rung:           1 — explicit specification, decisive on the trigger:
                `SPECIFICATION.md § Waypoints may be dropped without warning`
                states that the system cannot detect the silent case and
                accepts it as a first-release constraint, which forecloses
                reading *truncated* as detectable. 7 decides the termination
                branch, which no document reaches — abandoning visibly with
                the plan intact is the most conservative of the available
                endings.
Affects:        FR-112, FR-113, FR-114.
```

## RD-026 — Does *never overwrites* reach eviction under storage pressure?

```
Date:           2026-08-28
Question:       `SPECIFICATION.md § Route persistence` says *starting a new
                journey never overwrites an existing one*. That names the
                moment of confirmation. Does it also forbid discarding an
                older stored plan later, to make room?
Interpretation: It does. The failure the sentence names is a user losing
                planning work without warning, and eviction under pressure is
                that same failure arriving later — identical from the user's
                side, and worse for being untied to any action of theirs.
                FR-095 therefore carries two criteria, and the second is the
                record's point rather than a restatement of the first: a store
                that satisfies only the first passes every test written on a
                device holding two plans. Bounding the store is NFR-013's job
                and it does it by **time**, which is the bound
                `Architecture.md § Persistence and cross-device transfer`
                chose and argued for; a count-based eviction would be a
                second, unargued bound on the same object.
Rung:           7 — most conservative reasonable interpretation, decisive. The
                sentence is silent on eviction and either reading is
                available; the one that keeps the user's work is the
                conservative one. 1 corroborates rather than deciding: the
                failure `SPECIFICATION.md § Route persistence` exists to
                prevent is stated as an outcome for the user, not as a moment
                in the code. 4 corroborates — NFR-013 already bounds the
                store, so no unbounded-growth argument forces the other
                reading.
Affects:        FR-095. NFR-013, which is the bound that makes this
                affordable.
```
