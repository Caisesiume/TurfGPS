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
