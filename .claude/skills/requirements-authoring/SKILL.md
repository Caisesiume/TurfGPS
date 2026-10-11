---
name: requirements-authoring
description: The one definition of a TurfGPS requirement record — statement style, the MoSCoW priority field, the 29148 checklist, verification methods including human-judgement, IDs and citations, acceptance-criteria form, and the layout of docs/Requirements/. Read before authoring, reviewing or filing any FR or NFR.
---

# Requirements authoring

The shape of a record and of the corpus. The `analyst` owns this file and the corpus; anyone reviewing a record checks it against this file.

## The record

A `##` heading, then the fields inside a fence, exactly as it lands in a category file:

````markdown
## FR-000 — Cap candidates promoted to full evaluation

```
Statement:    The system shall promote no more candidate zones to full
              evaluation than the cap defined under
              `CalculationSpecification.md § Bounding the candidate set`.
Category:     Candidate identification
Source:       SPECIFICATION.md § Candidate zone identification
Priority:     MUST
Verification: test — a corridor holding more qualifying zones than the cap
              promotes exactly the cap
Acceptance:   given a corridor containing more qualifying zones than the
              configured cap, when candidates are promoted, then exactly the
              cap proceeds
Status:       to-build
Depends-on:   FR-018
Volatility:   proposed-constant — the cap is a proposed default with nothing
              measured behind it
Risk:         An unbounded promotion breaks the per-journey call budget.
Resolved-by:  #14, #15
```
````

`FR-000` and `NFR-000` are reserved for illustration and are never issued; allocation runs from `001`.

| Field | Rule |
|---|---|
| `FR-###` / `NFR-###` + title | three digits; the title a short imperative phrase |
| `Statement` | one sentence, `shall` / `should` / `may`, no RFC-2119 capitals |
| `Category` | one name from the category register, verbatim — it decides the file |
| `Source` | `Document.md § Section`, several separated by `;` |
| `Priority` | `MUST` · `SHOULD` · `COULD` · `WON'T-now` |
| `Verification` | one method keyword, then what it proves |
| `Acceptance` | given/when/then, or the method's own form below |
| `Status` | one value of the chain below |
| `Depends-on` | requirement codes it presupposes, or `none`; `;`-separated |
| `Volatility` | `proposed-constant` · `unverified-fact` · `open-question` · `settled`, then — for the first three — what it rests on, cited. Mandatory on records authored from 7 August 2026 |
| `Risk` | what goes wrong if it is built wrong; required on `MUST` and safety-path records |
| `Rationale` | optional; required where the statement would otherwise read as arbitrary, so nobody "simplifies" it away |
| `Resolved-by` | story numbers `#14, #15`, or `—` for none yet; a view of `TRACEABILITY.md`, never hand-edited |

`Risk` and `Rationale` are as long as their argument and no longer; a long one because the statement does two things means two records.

**Layout.** Values start at column 15; continuation lines indent 14 spaces; wrap at 78 columns with greedy fill on new writes. An `Acceptance` criterion and each `;`-separated entry start a new line. A citation is one atomic token: never break a line inside it, keep its trailing punctuation with it, and let it overhang column 78 rather than split. Re-wrapping existing fields is a separate change, never bundled into a content edit.

**Statuses:** `draft` → `to-build` → `implemented-unverified` → `implemented-verified`, plus `retired`. The analyst records `draft` → `to-build` once every question in the batch is decided and logged (no Owner sign-off, `ADR-0001 § D6`); a record with an open Owner question stays `draft` alone. The two `implemented-` values are classified from the tree. `approved` is a legacy value no transition writes. A status says how far the requirement has got; it never instructs, and build progress lives on the board.

## Registers

`Category` and any noun naming a registered physical object take their values from the registers in `docs/Requirements/README.md`, which only the analyst extends. A record with an unregistered category or noun is flagged, never translated to a near-match. Upstream documents are not bound by the term register, and reported speech quotes its source's noun.

## Obligation and priority are separate

The verb says whether the built system is obliged — `shall` (obliged), `should` (expected, justified deviation allowed), `may` (permitted). `Priority` says build order: `MUST` `P0`, `SHOULD` `P1`, `COULD` `P2`; `WON'T-now` is never filed as a story and stays as the record of a decided exclusion. A `COULD` record written with `shall` is correct.

## The 29148 checklist

Any reject is a rewrite, not a note.

| Check | Reject when |
|---|---|
| Necessary | deleting it leaves nothing the documents ask for unachievable |
| Appropriate | it names an implementation choice the need does not carry, unless a decision in `Architecture.md § Technology decisions` fixed it and is cited |
| Unambiguous | *appropriate, reasonable, efficient, user-friendly, as needed, etc.*, or a pronoun you re-read |
| Complete | a reader must open another record to know when it applies (a cited constant or a `Depends-on` is fine) |
| Singular | an `and`/`or` joins two separately testable outcomes |
| Feasible | it needs data the sources in `Architecture.md § Data sources and constraints` do not expose |
| Verifiable | no method yields a pass/fail, or `human-judgement` is named where a threshold exists |
| Correct | the cited section does not actually say it — then it is a question |
| Conforming | RFC-2119 capitals, a citation without its document, a restated constant, an unregistered category, or a mandatory field missing |

## Verification methods

| Method | Evidence |
|---|---|
| `test` | the automated test |
| `analysis` | the recorded calculation |
| `inspection` | the file and section examined |
| `demonstration` | the scenario walked |
| `human-judgement` | a named judge's recorded verdict against a named standard |

`human-judgement` names who judges and against what, and is allowed only where `CalculationSpecification.md` holds no threshold and none can be derived from one there — look first, then choose the method, and record in `Rationale` what was looked for. Its stories carry `human-verified` and always reach the Owner (`team-protocol § Escalation and human judgement`). An invented number is worse than an honest judgement.

## IDs and citations

- IDs are immutable: never reused, renumbered or recycled. Allocation is highest-ever-issued plus one, from the `Next free:` line in `docs/Requirements/README.md`.
- A citation is the token `Document.md § Section`, or `Document.md § D3` where the heading leads with a stable identifier — never a bare section name, never a trimmed heading. Inside prose fields it sits in a code span; `Source` and `Depends-on` take it bare.
- `Source` names the section that creates the obligation, not the one supplying a constant.
- Renaming or re-splitting a heading obliges repairing every citation to it in the same change (`docs/README.md § Conventions`).

## Three overrides

1. **Cite constants, never restate them** — no number from `CalculationSpecification.md` in a statement or criterion. A record may express the relation its section defines (a cost is a difference against a baseline); the test is: *if the section changed its values, would this record still be true?*
2. **A proposed constant never hardens into a `MUST`** (`safety-path-checklist § The proposal boundary`): the obligation is to read the configured value and carry its origin.
3. **Never infer a Turf mechanic** (`safety-path-checklist § Domain facts are verified, not inferred`): an untraceable domain claim is a question for the Owner, not a record.

## Acceptance criteria

The verification method decides the form. **Functional:** given/when/then, one per branch including the unhappy one; more than about three means the record is not singular. **Non-functional:** metric / threshold / condition, the threshold cited, never quoted. **Inspection:** artefact / property / location. **Human-judgement:** what the judge examines and the standard applied.

## The corpus — `docs/Requirements/`

| Path | Holds |
|---|---|
| `README.md` | the category and term registers, the ID allocation ledger, the folder's invariants |
| `INDEX.md` | one table per category, a row for every ID ever issued |
| `<category>.md` | one file per category, kebab-cased from its name; records in ascending ID order, forever |
| `TRACEABILITY.md` | requirement → story and story → requirement |
| `DECISIONS.md` | every ambiguity decided on the ladder, with its rung |

The record is authoritative and every index is derived: where a row disagrees with the record, the row is stale. A retired record stays in place as `Status: retired` with its index row, and its `Resolved-by` freezes. Stories exist only as GitHub issues and Epics only as Milestones: the corpus holds their numbers and Milestone names, never their content or status. `TRACEABILITY.md`'s story → requirement table is transcribed from each issue's `Resolves:` line; a disagreement between the two directions is a finding, never a silent fix.
