---
name: requirements-authoring
description: The one definition of a TurfGPS requirement record — IEEE 29148 statement style, the separate MoSCoW priority field, the 29148 quality checklist, verification-method vocabulary including human-judgement, ID and citation rules, acceptance-criteria form, and the layout of docs/Requirements/. Load before authoring, reviewing, or filing any FR or NFR.
---

# Requirements Authoring

The single definition of what a TurfGPS requirement looks like. `@requirements-engineer`, `@requirements-fr`, `@requirements-nfr` and `@requirements-librarian` all defer to this file; their own definitions carry their lane and their judgement, not a format. This file is **mechanics, plus the three project overrides that change what a correct record may say** — and each override points at its home rather than restating it.

## The canonical record

Below is a record **exactly as it lands in a category file**: the `##` heading, then the field block inside a fence. Both are mandatory. The fence is what preserves the column alignment, and what keeps a `grep FR-021` hit and a diff readable; without it the block reflows and the corpus stops being skimmable.

````markdown
## FR-021 — Cap candidates promoted to full evaluation

```
Statement:    The system shall promote no more candidate zones to full
              evaluation than the cap defined under *Bounding the candidate
              set* in CalculationSpecification.md.
Category:     Candidate identification
Source:       SPECIFICATION.md § Candidate zone identification
Priority:     MUST
Verification: test — a corridor holding more qualifying zones than the cap
              promotes exactly the cap, and records that the cap bound
Acceptance:   given a corridor containing more qualifying zones than the
              configured cap, when candidates are promoted, then exactly the
              cap proceeds and the binding is recorded
Status:       to-build
Depends-on:   FR-018
Risk:         An unbounded promotion breaks the per-journey call budget,
              which is the only thing making pipeline cost predictable.
Rationale:    The cap binds even where the extra candidates look individually
              cheap: the call budget is a per-journey total, so an exception
              granted here is spent out of every later stage's allowance.
Resolved-by:  #14, #15
```
````

| Field | Rule | Filled by |
|---|---|---|
| `FR-###` / `NFR-###` + title | Zero-padded to three digits; title is a short imperative phrase | author |
| `Statement` | One sentence, **shall / should / may**, no RFC-2119 capitals | author |
| `Category` | One name from the **category register**, copied verbatim — it decides the file | RE assigns |
| `Source` | `Document.md § Section`; several separated by `;` | author |
| `Priority` | MoSCoW: `MUST` · `SHOULD` · `COULD` · `WON'T-now` | RE assigns |
| `Verification` | One keyword from the vocabulary below, then what it proves | author |
| `Acceptance` | given/when/then (FR) or metric/threshold/condition (NFR) | author |
| `Status` | One value from the chain below | RE / reconciler |
| `Depends-on` | Requirement codes this one presupposes, or `none`; several separated by `;` | author |
| `Risk` | What goes wrong if this is built wrong or not at all. **Required on `MUST` and safety-path records**, omit elsewhere | author |
| `Rationale` | Optional. Required where the statement would otherwise read as arbitrary or over-strict, so a later agent cannot "simplify" it away | author |
| `Resolved-by` | Story numbers, or `—` where none is allocated yet — a view of the matrix, never hand-edited here | **librarian only** |

**`Risk` and `Rationale` are as long as their argument, and no longer.** Both once read *one line* here while the corpus's own contested records ran to a dozen, which left every author choosing between the table and the practice. The practice is right: these two fields exist to stop a later agent undoing a decision it does not understand, and an argument compressed to one line stops nothing. Brevity is the default, not the ceiling — a `Rationale` earns each line by naming something a reader would otherwise get wrong. A record whose `Rationale` is long because the *statement* is doing too much is not a long-rationale problem; it is a singularity failure, and the fix is two records.

**The field block's layout.** Values begin at **column 15** — the width of the longest field name, `Verification:`, plus one space — and continuation lines are indented **14 spaces** to sit under them. Lines wrap at **78 columns**. This is not taste: the alignment is what makes a `grep FR-021` hit readable on its own and what keeps a one-field diff to the lines that field occupies. A re-wrap is formatting and belongs to `@requirements-librarian`; it is never bundled into a content edit, because in a diff the two are indistinguishable.

**Multi-value fields, and their separators.** `Source` and `Depends-on` separate entries with **`;`**. `Resolved-by` separates story numbers with **`, `** — `#14, #15`. The split is not arbitrary and is not a licence to choose: a `Source` entry carries a section name that may itself contain a comma, so a comma cannot delimit them, and `Depends-on` follows `Source` because the two reference fields sit together and reading alike is worth more than either choice on its own. A story number cannot be ambiguous, and `#14, #15` is also how `INDEX.md` renders it, so changing it would rewrite every index row for nothing. Where a field's value spans lines, the separator ends the line and the next entry starts a new one at the continuation indent, so a `grep` for one code returns one line.

**A citation is an atomic token, and a line break never falls inside one.** A split citation reads perfectly as prose and cannot be found by `grep` at all. The rule holds in **both forms** — `Document.md § Section` and the italic *Section* in `Document.md` — and at every seam the two of them offer: before the `§`, after it, between the words of a section name, and between a section name and the filename qualifying it. Wrap before the whole citation; a line that ends short is the rule working. **Where a citation does not fit even alone at the continuation indent, the 78-column wrap yields and the citation stays whole** — an atomic token that does not fit is not made to fit, exactly as a code formatter leaves a long string literal overhanging its margin. **Trailing punctuation is part of the token**: the comma, semicolon or full stop closing a citation wraps with it and never alone onto the next line, so a citation ending at column 78 carries its mark to 79 rather than stranding it at the head of the next line — where, in `Source` and `Depends-on`, a leading `;` reads as the separator above rather than the end of what precedes it.

`Depends-on` and `Risk` are not decoration: `docs/README.md § Requirements/` names risk and dependencies as corpus content, and the scrum-master sequences promotions by **Priority first and dependency order second** — never the reverse, per the Priority field in the `turfgps-board-ops` skill.

**The category register.** `Category` is a controlled vocabulary with exactly one home: the register in `docs/Requirements/README.md`. A name is legal once it appears there, and it is `@requirements-engineer` alone that seeds and extends it — a new category is an RE act, not a side effect of authoring. `@requirements-librarian` files a record under its `Category` verbatim and **flags any record whose `Category` is not on the register**, rather than inventing a near-match. The functional-area and quality-attribute lists in the RE agents' own definitions are **coverage prompts** — they exist to make an author ask "does this section impose anything here?" — and are never a source of category names. Two spellings of one area produce two files for one subsystem, and coverage per subsystem stops being checkable, which is the whole reason the field exists.

**Status chain:** `draft` → `approved` → `to-build` → `implemented-unverified` → `implemented-verified`, plus `retired`.

- **`draft`** — authored, not yet signed off. Everything `@requirements-fr` and `@requirements-nfr` return is `draft`.
- **Sign-off is an event, not a resting state.** While `@requirements-reconciler` is dormant, the Owner's sign-off moves a record from `draft` **straight to `to-build`** — there is no classification step for it to wait in `approved` for. `approved` becomes reachable only once the reconciler is live: sign-off writes `approved`, and the reconciler then writes one of its verdicts. Downstream, "**approved requirements**" — the phrase `@requirements-story-organizer` files by — means **`to-build` or later**, never the literal value `approved`.
- **`implemented-unverified`** and **`implemented-verified`** are `@requirements-reconciler`'s verdicts; its third, **`to-build`**, is the entry state.
- **`retired`** — see *Corpus layout*.

**A status states how far a requirement has got; it never instructs.** `to-build` does not mean build it now — `Priority` and the board decide that. The value bounds what may be claimed about the requirement, not what anyone should do next.

**The chain says nothing about how far the *build* has got, deliberately.** That question belongs to the board, and a record reaches it through `Resolved-by`: the story numbers are a live link, answered correctly whenever it is followed, where a status copied onto the record is an assertion the corpus has no way to keep true. A chain value whose only event lives on the board does not belong in this chain — see *What the corpus may hold about a story or an epic*.

## Obligation lives in the verb; priority lives in the field

Two independent questions, and collapsing them is the defect this skill exists to prevent.

| | Question | Vocabulary |
|---|---|---|
| Statement verb | Once built, is the system obliged? | `shall` (obliged) · `should` (expected, justified deviation allowed) · `may` (permitted, optional) |
| `Priority` field | In what order do we build it? | `MUST` · `SHOULD` · `COULD` · `WON'T-now` |

**No RFC-2119 capitals inside a statement.** In this corpus the word `MUST` has exactly one meaning: priority. A `COULD` requirement written with `shall` is entirely correct — it binds once built, it is just not built first. `MUST`/`SHOULD`/`COULD` map to the board's `P0`/`P1`/`P2`; **`WON'T-now` maps to nothing** — it is not filed as a story at all, and stays in the corpus as the record of a decided exclusion. See the Priority field in the `turfgps-board-ops` skill.

## The 29148 accept/reject checklist

Apply per record. Any reject is a rewrite, not a note.

| Characteristic | The check | Reject when |
|---|---|---|
| Necessary | Delete it — does something the documents ask for become unachievable? | No. It is a feature idea, not a requirement |
| Appropriate | Does it state the need, at the need's level of abstraction? | It names an implementation choice the need does not carry — a table, endpoint, class, or a library chosen freely. A technology **fixed by a decision in `Architecture.md § Technology decisions`** may be named where the requirement is *about* that decision, and is then cited rather than described |
| Unambiguous | Could two engineers build different things from it? | It contains *appropriate*, *reasonable*, *efficient*, *user-friendly*, *as needed*, *etc.*, or a pronoun you re-read |
| Complete | Does it stand alone — actor, condition, outcome? | The reader must open another requirement to know when it applies. (A **cited** constant is not incompleteness; nor is a `Depends-on` link, which states build order — the statement must still be comprehensible without opening what it depends on) |
| Singular | One behaviour or one quality? | An `and`/`or` joins two separately testable outcomes |
| Feasible | Achievable within `Architecture.md`'s decisions and the call budget? | It needs data that the sources named in `Architecture.md § Data sources and constraints` and its technology decisions do not expose — Turf, the map extract, **and the elevation source**, which is not OSM |
| Verifiable | Would the named method actually produce a pass/fail? | No method fits — including `human-judgement`; or `human-judgement` is named where `CalculationSpecification.md` already holds a threshold this criterion could have been written against |
| Correct | Does the cited section actually say this? | The citation is close but the section does not state it — then it is a question, not a requirement |
| Conforming | Does the record carry the shape this file defines? | RFC-2119 capitals in the statement · a citation without its document · a restated constant · a `Category` not on the register · a field the field table makes mandatory that is absent |

## Verification methods

| Method | What it is | The evidence |
|---|---|---|
| `test` | Run the system on defined inputs, compare against an expected result | the automated test |
| `analysis` | Reason, model, or calculate over the design or data without executing | the recorded calculation |
| `inspection` | Examine an artefact — code, config, schema, document | the file and section examined |
| `demonstration` | Operate the system and observe; no instrumentation | the scenario walked |
| `human-judgement` | A named person judges quality against a named standard | their recorded verdict |

**`human-judgement` must name who judges and against what**, or it is not a verification method — `human-judgement — the Owner, against the measure of success in SPECIFICATION.md § Accessibility scope for the first release`. Per `docs/DELIVERY.md § Escalation and human judgement` these never close on agent consensus, and their stories carry the `human-verified` label.

**`human-judgement` is available only where no threshold exists and none can be created without inventing one.** `docs/README.md § Conventions` puts every formula, constant and threshold in `CalculationSpecification.md` and nowhere else, so the check is a look at one document rather than an argument: scan its sections for a number this criterion could have been written against, then ask whether one could be derived from a number that is there. A threshold that exists is used; one that follows from an existing threshold is derived and cited; one that would have to be picked is not picked, and the quality is judged instead. Do the look **before** choosing the method — chosen first and justified after, the method decides what counts as a threshold.

**The look is the whole burden, and it argues for the judgement more often than against it.** Where nothing is found, `human-judgement` is correct and needs no defence beyond its judge and its standard: the reject in the checklist above is a *found* threshold that was written around, never a missing paragraph. Record what was looked for and not found in the record's `Rationale` — one clause, and an instance of the rule already governing that field, since a judged criterion reads as under-specified without it. It is there to spare the next reviewer the same search, not to buy permission for the method. An author who picks a figure to avoid writing that clause has invented a constant to save a sentence, and the sentence was by far the cheaper of the two.

`FR-016` is the worked example, and its neighbours are the other half of it. Nothing in `CalculationSpecification.md` sets a distinctness bar between journey alternatives, and any figure standing in for *meaningfully different* would have been chosen on the spot; the record says exactly that and names the Owner as judge. `FR-014` and `FR-015` carry the part of the same source obligation that could be stated without a number — identity and outright dominance are orderings, not tolerances — and both are `test`. One sentence in `SPECIFICATION.md § Recommended journey alternatives`, split by what the look found: the judgement takes the residue, never the whole.

**A fabricated metric is worse than an honest `human-judgement`.** An invented number gets measured, passes, and the quality it stood for is never examined. Where a quality is real and unmeasurable, say so.

## IDs and citations

- `FR-###` / `NFR-###`, allocated in one sequence per prefix. **Immutable: never reused, never renumbered, never recycled after deletion.** Retirement sets `Status: retired` and leaves the record in place — see *Corpus layout*.
- `Source` is always `Document.md § Section`. Four upstream documents make a bare section name ambiguous, and a bare citation is a librarian finding. Inside a statement, use the documents' own convention — an italic section name qualified by its filename, per `docs/README.md § Conventions`.
- **Outside a statement, both forms are correct**, and `Document.md § Section` is the one to reach for. `Rationale`, `Risk`, `Verification` and `Acceptance` are the corpus's own prose rather than the documents', so the `§` form reads cleanly there and matches the `Source` field sitting a few lines above it; the italic form stays available and is the natural choice where the citation is the grammatical object of a sentence — "the weight comes from *Proposed rank-to-weight curve* in `CalculationSpecification.md`". The statement is the one place the documents' own convention governs, because the statement is what a reader carries back into the source. **Nowhere is a bare section name correct.** A back-reference — "*User time constraints* in the same document" — names its document where the antecedent is the citation immediately before it, and nowhere else; if a reader has to scan upward past a second filename to resolve it, it is bare.
- **`Source` names the section that creates the obligation** — the place a reader goes to challenge whether this requirement should exist. A constant cited *inside* the statement is not a `Source` entry: `CalculationSpecification.md` supplies the value, it does not impose the duty. A record citing a calculation section as its `Source` is usually one whose real source was never found.

## Three project overrides a generic IEEE habit gets wrong

Each has a home elsewhere; only the consequence for a requirement record is stated here.

1. **Cite constants, never restate them** — `docs/README.md § Conventions`, `safety-path-checklist` § Non-negotiables. For a record: a threshold appears as a reference to its section, never as a number, in the statement *and* in the acceptance criteria.
2. **A proposed constant must never harden into a `MUST`** — `safety-path-checklist` § The proposal boundary, which also forbids an unexplained literal and forbids quoting a proposal to a user as established. For a record: the obligation is that the system reads the configured value and carries its documented origin, never that the value is a particular figure.
3. **Never infer a Turf mechanic** — `safety-path-checklist` § Domain facts are verified, not inferred. For a record: a domain assertion with no traceable source is a question for the Owner, not a requirement — do not write the record.

**Where override 1 stops: a constant is copied, a relation is expressed.** Override 1 forbids the model's *values*; it is silent on its *shape*, and that silence has to be closed here because it recurs on every cost and value record. A record **may** state the relation its cited section defines — that a cost is a difference against a baseline, that a stop's time is included in it, that value rises with rarity — because on those records the relation **is** the testable content. An acceptance criterion reading *computed correctly per the cited section* asserts nothing a test could fail; it is a citation wearing a criterion's clothes, and it passes the day the arithmetic is inverted. A record **may not** carry the numbers the relation is parameterized by, and may not walk the section's derivation step by step: a reproduced derivation is a second home for the model, which is the defect the one-home rule exists to prevent, and it is a worse one than a copied number because it looks like requirements work rather than duplication.

**The test to apply:** *if the cited section changed the values in its formula, would this record still be true?* If yes, a relation has been expressed and the record stands. If no, a constant has been copied and the record is rewritten. A record that would be falsified by a change to the section's *structure* — a difference becoming a ratio — is behaving correctly: that is a specification change, and requirements resting on it are meant to be revisited rather than insulated from it. `FR-008` and `FR-009` are the worked examples: each states what cost is measured against and what is included in it, and cites `CalculationSpecification.md` for how either is computed.

## Acceptance-criteria form

**Where the record's kind and its verification method point at different forms, the method wins.** An FR verified by `inspection` or by `human-judgement` takes that method's form below, not given/when/then. The reason is the same in both cases: the wrapper describes an execution that does not happen, and a criterion narrating a run nobody performs is the one kind of criterion that can never fail. `FR-016` is the worked example — a functional requirement whose criterion is a judgement, in the judgement form.

**Functional — given/when/then.** One criterion per branch the statement names, including the unhappy branch. Needing more than about three is a signal the requirement is not singular.

**Non-functional — metric / threshold / condition.** All three, or it is not measurable: *p95 replacement latency ≤ the threshold under `CalculationSpecification.md § Review-interaction thresholds`, from retained solve state.* Where the threshold is a proposed constant, the criterion cites it rather than quoting it — override 2 applies to acceptance criteria exactly as it does to statements.

**Under `inspection` — artefact / property / location.** Name the artefact examined, the property that must hold, and where a reader confirms it: *the configuration file naming the candidate cap carries a comment citing `CalculationSpecification.md § Bounding the candidate set` as its origin.* Do not wrap an inspection in given/when/then — nothing is executed, and the wrapper reads as a test that does not exist.

**Under `human-judgement`**, the criterion states what the judge examines and the standard applied — never a number standing in for the judgement.

## Corpus layout

Owned by `@requirements-librarian`. The corpus root is **`docs/Requirements/`** and every path below is relative to it; `docs/README.md § Requirements/` states why it is a folder rather than a file, and that every bare `Requirements/` in the repository means this path.

| Path | Holds |
|---|---|
| `README.md` | Front door: what the corpus is, the **category register**, the ID allocation ledger, and the folder's invariants |
| `INDEX.md` | One index table per category, carrying a row for every ID ever issued |
| `<category>.md` | One file per category; every record lives in exactly one |
| `TRACEABILITY.md` | The matrix, both directions |

**Why README and INDEX are separate files.** The allocation ledger is the only guard against ID reuse, and it must not sit below fifteen screens of table where an author stops scrolling. The register and the ledger change rarely; the index changes on every filing. Keeping them apart means a routine filing does not rewrite the one file the whole RE family reads first.

**The record is authoritative; every index is derived.** Where a row in `INDEX.md` and the record in a category file disagree, **the record is right and the row is stale** — a librarian finding, fixed in the index, never fixed by editing the record to match. This is the rule that makes it safe to regenerate an index without reading it as a second source of truth.

**Category files.** A file is named as the kebab-cased `Category` field, verbatim — `Candidate identification` → `candidate-identification.md` — so a record names its file and a file names its category with nothing to look up. No numeric prefixes: a file name never encodes order, so adding a category never renames a file. FRs and NFRs share the folder, never a file. Inside: the category name, one line of scope, then one record per requirement in the canonical form above, **ID ascending, an order that never changes** — not on retirement, not on re-prioritization.

**`INDEX.md`** carries a TOC, then one section per category holding this table:

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-021` | Cap candidates promoted to full evaluation | Candidate identification | MUST | to-build | test | #14, #15 |

`Verification` is the keyword alone; the evidence sentence stays on the record. `Category` is redundant under a category heading and kept anyway, so a `grep FR-021` hit answers the whole question without its surrounding lines.

**`Resolved-by` reads `—` where no story is allocated**, on the record and in the index alike, and it means *not yet allocated* — never *unknown*. One spelling in both places, because the moment the record and its view word the same emptiness differently, a reader has to decide whether the difference is meaningful.

**Tombstones.** A retired requirement does not move and is not deleted: the record stays in ID order reading `Status: retired`, and its index row stays. **The index carries a row for every ID ever issued** — that ledger duty outranks its skim duty. `README.md` carries one line, `Next free: FR-078 · NFR-023`, and **allocation is highest-ever-issued plus one, never lowest-unused**, so a retirement cannot free a number even if a tombstone is lost. A retired record leaves `TRACEABILITY.md`, so its `Resolved-by` — on the record and in the index — **freezes at its last value** rather than reverting to `—`: it is the historical record of what was built for it, and blanking it would erase the only remaining trace.

**What the corpus may hold about a story or an epic.** Requirements are documented in this repository, under `docs/Requirements/`. **Stories exist only as GitHub Issues, and epics only as GitHub Milestones** — neither is ever *documented* here. The corpus stores their **identifiers as links**, never their content and never their state: a story's **number** (`#14`) and its epic's **Milestone name** are the whole of what may appear, anywhere in the folder. A story's **title, body, acceptance criteria, assignee, labels or status** may not — not in `TRACEABILITY.md`, not in a record's `Resolved-by`, not in `INDEX.md`, and not in a column added later because it was convenient at the time.

The reason is one home per fact. A story's content and its state are owned by the board and the issue, which answer them live and answer them correctly; a copy here is stale from the moment it is written and leaves a reader holding two answers with no way to tell which is current. This is why *Story → requirement* carries **no board-status column**: the board answers status, the issue answers content, and this file answers one question only — which requirements a story resolves. **The rule has no exception.** Anything proposing one — a status copied from the board, a label mirrored here, a column added because it was convenient at the time — is a violation of it, however carefully the copy is scoped.

**`TRACEABILITY.md`** holds one table per direction: *Requirement → story* (requirement · epic Milestone · stories), one row per non-retired requirement; and *Story → requirement* (story # · epic · resolves), transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit. The first table is the single home for the requirement→story link — the record's `Resolved-by` and the index column are views of it, regenerated in the same pass. **A disagreement between the two directions is a finding for the RE, never a silent reconciliation.**
