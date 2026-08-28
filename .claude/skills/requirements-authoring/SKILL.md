---
name: requirements-authoring
description: The one definition of a TurfGPS requirement record — IEEE 29148 statement style, the separate MoSCoW priority field, the 29148 quality checklist, verification-method vocabulary including human-judgement, ID and citation rules, acceptance-criteria form, and the layout of docs/Requirements/. Load before authoring, reviewing, or filing any FR or NFR.
---

# Requirements Authoring

The single definition of what a TurfGPS requirement looks like. `@requirements-engineer`, `@requirements-fr`, `@requirements-nfr` and `@requirements-librarian` all defer to this file; their own definitions carry their lane and their judgement, not a format. This file is **mechanics, plus the three project overrides that change what a correct record may say** — and each override points at its home rather than restating it.

## The canonical record

Below is a record **exactly as it lands in a category file**: the `##` heading, then the field block inside a fence. Both are mandatory. The fence is what preserves the column alignment, and what keeps a `grep FR-000` hit and a diff readable; without it the block reflows and the corpus stops being skimmable.

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
              promotes exactly the cap, and one holding fewer promotes all
              of them
Acceptance:   given a corridor containing more qualifying zones than the
              configured cap, when candidates are promoted, then exactly the
              cap proceeds
              given a corridor containing fewer qualifying zones than the
              cap, when candidates are promoted, then all of them proceed
Status:       to-build
Depends-on:   FR-018
Volatility:   proposed-constant — the cap under
              `CalculationSpecification.md § Bounding the candidate set` is a
              proposed default with nothing measured behind it
Risk:         An unbounded promotion breaks the per-journey call budget,
              which is the only thing making pipeline cost predictable.
Rationale:    The cap binds even where the extra candidates look individually
              cheap: the call budget is a per-journey total, so an exception
              granted here is spent out of every later stage's allowance.
Resolved-by:  #14, #15
```
````

**The example's ID is reserved, not chosen.** Allocation runs upward from `FR-001`, so `FR-000` and `NFR-000` can never be issued to a real record, and an illustration carrying one can never answer to a live requirement's identity. That collision costs more than it sounds: this record is fictional, so a `grep` for its code returning it *alongside* a real one hands a reader two answers, one of which was never a requirement — and a `Rationale` citing that code by number then resolves to either. **Illustrate with the reserved code and never with a live one**; the live records this file cites by name — `FR-008`, `FR-009`, `FR-014`, `FR-015`, `FR-016`, `FR-022`, `NFR-001` — are genuine cross-references and mean the records they name. The reservation's home is the allocation ledger in `docs/Requirements/README.md`; this file only spends it. Do not renumber the example to look more realistic.

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
| `Volatility` | What the record rests on and how likely that is to move: one value from the vocabulary below, then what it rests on. **Mandatory** — see the dated transition | author |
| `Risk` | What goes wrong if this is built wrong or not at all. **Required on `MUST` and safety-path records**, omit elsewhere | author |
| `Rationale` | Optional. Required where the statement would otherwise read as arbitrary or over-strict, so a later agent cannot "simplify" it away | author |
| `Resolved-by` | Story numbers, or `—` where none is allocated yet — a view of the matrix, never hand-edited here | **librarian only** |

**`Risk` and `Rationale` are as long as their argument, and no longer.** Both once read *one line* here while the corpus's own contested records ran to a dozen, which left every author choosing between the table and the practice. The practice is right: these two fields exist to stop a later agent undoing a decision it does not understand, and an argument compressed to one line stops nothing. Brevity is the default, not the ceiling — a `Rationale` earns each line by naming something a reader would otherwise get wrong. A record whose `Rationale` is long because the *statement* is doing too much is not a long-rationale problem; it is a singularity failure, and the fix is two records.

**`Volatility` says what the record stands on.** One value from `proposed-constant` · `unverified-fact` · `open-question` · `settled`, then — for the first three — what it rests on, cited. A record is `proposed-constant` where a figure it depends on is marked a proposal; `unverified-fact` where it depends on an external fact nobody in this repository has checked against a published source; `open-question` where it depends on something a document lists as open, or on a working default standing until a ruling; `settled` where none of the three applies. Where two apply, name the one that would falsify the record soonest and let `Rationale` carry the rest.

**The field exists because a record's precision is sometimes deliberately limited, and nothing else on the record says so.** `FR-011`'s criteria assert non-identity rather than naming which alternative wins, because the rank-to-weight mapping is a proposal and a criterion naming the winner would harden it. That record does carry the argument — in the last sentence of its longest field. A reader deciding whether to sharpen those criteria must reach the bottom of that field to learn they must not, and a reviewer asking which records move when a given proposal moves has to read every `Rationale` in the corpus. `Volatility` puts both answers on one line a pattern can find.

**Mandatory, not conditional, and `settled` is a legal value.** `Risk`'s conditionality works because another field decides it: a reader meeting no `Risk` checks `Priority` and knows whether the absence is legal. Nothing on a record decides whether it rests on a proposal, so an absent `Volatility` would be undecidable between *settled* and *never considered* — the ambiguity `Resolved-by: —` needed a stated meaning to escape. What the field buys is the **look**, exactly as the threshold look before `human-judgement` does: its worth is that it was run, and a field that may be omitted is a look that leaves no trace when it is skipped. The cost is one short line on a settled record, and that is the price of the line meaning something when it is not `settled`.

**The transition is dated, and the records filed before 7 August 2026 do not carry it.** Retrofitting means judging what each already-signed record rests on — content work across five signed batches, in one commit, on records nobody is otherwise editing. The same judgement made once per record as it is next touched is the better trade. So: **mandatory on every record authored from 7 August 2026**, and gained by an earlier record opportunistically, when that record is next edited for a reason of its own. An absent `Volatility` on a record filed before that date is not a finding; an absent one on a record filed after it is a `Conforming` reject.

**The field block's layout.** Values begin at **column 15** — the width of the longest field name, `Verification:`, plus one space — and continuation lines are indented **14 spaces** to sit under them. Lines wrap at **78 columns**, and **fill is greedy** — a line takes every word that fits under the ceiling. This is not taste: the alignment is what makes a `grep FR-000` hit readable on its own and what keeps a one-field diff to the lines that field occupies, and a stated fill is what makes two agents wrapping the same field produce the same bytes. A ceiling without a fill leaves a minimal fix and a full re-flow both legal on the same field, and the difference surfaces later as churn that reads like a content edit. **Greedy fill binds on new writes; existing fields are normalized opportunistically, when the field is next edited for another reason** — never as a dedicated re-wrap pass across the corpus. Several fields carry a word's slack from earlier batches and are not defects. **Three kinds of short line are not slack and are never filled:** an `Acceptance` criterion begins a new line, each entry of a `;`-separated field begins a new line — both below — and a line ending short before a citation is the atomicity rule working, not room going spare. A re-wrap is formatting and belongs to `@requirements-librarian`; it is never bundled into a content edit, because in a diff the two are indistinguishable.

**Multi-value fields, and their separators.** `Source` and `Depends-on` separate entries with **`;`**. `Resolved-by` separates story numbers with **`, `** — `#14, #15`. The split is not arbitrary and is not a licence to choose: a `Source` entry carries a section name that may itself contain a comma, so a comma cannot delimit them, and `Depends-on` follows `Source` because the two reference fields sit together and reading alike is worth more than either choice on its own. A story number cannot be ambiguous, and `#14, #15` is also how `INDEX.md` renders it, so changing it would rewrite every index row for nothing. Where a field's value spans lines, the separator ends the line and the next entry starts a new one at the continuation indent, so a `grep` for one code returns one line.

**A citation is an atomic token, and a line break never falls inside one.** A split citation reads perfectly as prose and cannot be found by `grep` at all. The rule holds in the one form there is — `Document.md § Section` — at every seam it offers: before the `§`, after it, and between the words of a section name. Wrap before the whole citation; a line that ends short is the rule working. **Where a citation does not fit even alone at the continuation indent, the 78-column wrap yields and the citation stays whole** — an atomic token that does not fit is not made to fit, exactly as a code formatter leaves a long string literal overhanging its margin. **Trailing punctuation is part of the token**: the comma, semicolon or full stop closing a citation wraps with it and never alone onto the next line, so a citation ending at column 78 carries its mark to 79 rather than stranding it at the head of the next line — where, in `Source` and `Depends-on`, a leading `;` reads as the separator above rather than the end of what precedes it. **Checking this rule means consulting the cited document's heading list, not matching a pattern — until the file converts.** A regex over this corpus's delimiter — the code span — and `§` reports its bare filenames and requirement codes, `INDEX.md` and `FR-018` among them, as split citations, and reads "a constant with no home in `CalculationSpecification.md`" as one, though it names no section and asserts an absence. Against the headings all three resolve to nothing and the check returns zero. **On a converted file the pattern is enough**, which is precisely what the token form was chosen to buy — `docs/README.md § Conventions` states the property that makes it so. The consequence here is the one to remember: the cheap check may be run only against the converted-file list kept there, and a file absent from that list still needs the headings.

`Depends-on` and `Risk` are not decoration: `docs/README.md § Requirements/` names risk and dependencies as corpus content, and the scrum-master sequences promotions by **Priority first and dependency order second** — never the reverse, per `turfgps-board-ops § Priority`.

Two fields take their values from a **register** rather than from the author's judgement — `Category`, and any noun naming a registered object. Both registers live in `docs/Requirements/README.md` and both have a section of their own below.

**Status chain:** `draft` → `approved` → `to-build` → `implemented-unverified` → `implemented-verified`, plus `retired`.

- **`draft`** — authored, not yet signed off. Everything `@requirements-fr` and `@requirements-nfr` return is `draft`.
- **The `to-build` transition is an event, not a resting state.** While `@requirements-reconciler` is dormant — a condition derived from the tree per `codebase-map § Which map is authoritative — check the tree, do not assume`, never asserted here or inherited downstream — `@requirements-engineer` records the move from `draft` **straight to `to-build`** on its own authority, once every question in the batch is resolved under the `§5` precedence ladder and logged in `docs/Requirements/DECISIONS.md` — there is no classification step for it to wait in `approved` for, and no Owner sign-off gating it (`ADR-0001 § D6`). A record carrying a `§21`-qualifying open question stays `draft` and blocks **by itself**, never with its batch behind it. `approved` becomes reachable only once the reconciler is live: the transition writes `approved`, and the reconciler then writes one of its verdicts. Downstream, "**approved requirements**" — the phrase `@requirements-story-organizer` files by — means **`to-build` or later**, never the literal value `approved`.
- **`implemented-unverified`** and **`implemented-verified`** are `@requirements-reconciler`'s verdicts; its third, **`to-build`**, is the entry state.
- **`retired`** — see `§ Corpus layout`.

**A status states how far a requirement has got; it never instructs.** `to-build` does not mean build it now — `Priority` and the board decide that. The value bounds what may be claimed about the requirement, not what anyone should do next.

**The chain says nothing about how far the *build* has got, deliberately.** That question belongs to the board, and a record reaches it through `Resolved-by`: the story numbers are a live link, answered correctly whenever it is followed, where a status copied onto the record is an assertion the corpus has no way to keep true. A chain value whose only event lives on the board does not belong in this chain — see `§ What the corpus may hold about a story or an epic`.

## The category register

`Category` is a controlled vocabulary with exactly one home: the register in `docs/Requirements/README.md`. A name is legal once it appears there, and it is `@requirements-engineer` alone that seeds and extends it — a new category is an RE act, not a side effect of authoring. `@requirements-librarian` files a record under its `Category` verbatim and **flags any record whose `Category` is not on the register**, rather than inventing a near-match. The functional-area and quality-attribute lists in the RE agents' own definitions are **coverage prompts** — they exist to make an author ask "does this section impose anything here?" — and are never a source of category names. Two spellings of one area produce two files for one subsystem, and coverage per subsystem stops being checkable, which is the whole reason the field exists.

## The term register

**`Category` is not the only controlled vocabulary this corpus needs.** Where several records oblige things about **one physical object**, the noun for that object is registered in `docs/Requirements/README.md` alongside the category register, and an author writes the registered noun verbatim in every field. It is `@requirements-engineer` alone that settles or extends it, on the category register's reasoning exactly: a second noun is a decision, never a synonym an author reaches for mid-sentence. `@requirements-librarian` **flags a record using an unregistered noun for a registered object** rather than translating it.

**The reason is sharper here than for categories, and it is worth stating rather than inferring.** A duplicate category costs coverage checking; a duplicate noun costs an *obligation*. Two records can each read perfectly, cite their sources faithfully, and pass every per-record check while obliging things about what a reader takes to be one object and an implementer builds as two — and no per-record check can see it, because neither record is wrong. That is the set-level shape, and the corpus has paid for it: *road position* and *stopping position* ran side by side with nothing binding the point a level check passed at to the point the car stopped at, and the count and the repair are recorded at the settlement in `docs/Requirements/README.md § Term register`.

**Three bounds, because each is a repair someone will otherwise make.** The **upstream documents are not bound** by this register and carry several spellings of some objects; where a document owes a glossary, the settled noun is an upstream finding for the Owner rather than a licence to edit the document. **Reported speech is not translated** — a record describing what a cited section says quotes that section's noun, because translating it changes what the record attributes to its source. And a **different object is not a violation**: *stopping place* applied to a road, or *parking location* naming a place-class, are other objects that happen to read alike, and the register binds the object rather than the word.

**A settled noun is applied by reading, not by pattern.** The corpus wraps its fields, so an occurrence broken across a line break survives any per-line replacement — which is exactly how four of them survived the settlement of 7 August 2026, one of them inside the acceptance criterion of a record whose statement had already been converted. A settlement pass is checked by reading the records that carry the object, and the check is not a `grep`.

## Obligation lives in the verb; priority lives in the field

Two independent questions, and collapsing them is the defect this skill exists to prevent.

| | Question | Vocabulary |
|---|---|---|
| Statement verb | Once built, is the system obliged? | `shall` (obliged) · `should` (expected, justified deviation allowed) · `may` (permitted, optional) |
| `Priority` field | In what order do we build it? | `MUST` · `SHOULD` · `COULD` · `WON'T-now` |

**No RFC-2119 capitals inside a statement.** In this corpus the word `MUST` has exactly one meaning: priority. A `COULD` requirement written with `shall` is entirely correct — it binds once built, it is just not built first. `MUST`/`SHOULD`/`COULD` map to the board's `P0`/`P1`/`P2`; **`WON'T-now` maps to nothing** — it is not filed as a story at all, and stays in the corpus as the record of a decided exclusion. See `turfgps-board-ops § Priority`.

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

- `FR-###` / `NFR-###`, allocated in one sequence per prefix. **Immutable: never reused, never renumbered, never recycled after deletion.** Retirement sets `Status: retired` and leaves the record in place — see `§ Corpus layout`.
- `Source` is always `Document.md § Section`. Four upstream documents make a bare section name ambiguous, and a bare citation is a librarian finding. **That value is the citation token itself** — the form is defined in `docs/README.md § Conventions` and is not restated here — and the corpus writes that one token in every field, statement included.
- **The statement's exception is gone, and the second form with it.** A statement once took the documents' own convention, an italic section name qualified by its filename, on the ground that the statement is what a reader carries back into the source. That reason held only while the two halves differed in *content* — section-before-filename against filename-before-section. The token form makes them the same token and leaves only the delimiter differing, so the exception has nothing left to protect: a reader carries back `Architecture.md § Retrieving zones` whichever field it was read in. **Nowhere is a bare section name correct**, and a back-reference no longer escapes that. "*Retrieving zones* in the same document" was legal only while a citation could lean on the antecedent before it, and a **self-contained** token by definition cannot lean on anything. A second citation to a document already named repeats the document.
- **Which token to write is settled above; how to delimit it is a separate question.** In the corpus the delimiter is a **code span** wherever the citation sits inside a sentence — `Architecture.md § D1` — which is `Statement`, `Rationale`, `Risk`, `Verification` and `Acceptance`. There the span is punctuation rather than formatting: it marks where the citation starts and where it stops, the same service the atomicity rule performs across a line break. A filename standing alone in prose, outside any citation, carries the span too, and always has.
- **`Source` and `Depends-on` take no delimiter, and the migration never reaches them.** The entire value *is* the citation, so there is nothing to separate it from — a span there marks a boundary the field name already marks, and spends two of the 78 columns doing it. Across the 39 records on 4 August 2026, all **42** `Source` citations and all **32** `Depends-on` codes are bare, and that is this rule holding rather than drift waiting to be swept. **Converting a file does not touch its `Source` or `Depends-on` lines.** A commit that backticks them has not finished the migration; it has broken this rule, and it has done so across the field a `grep FR-000` reader depends on most. **The fence settles nothing here.** Inside a fence no markdown renders, so the span is never doing cosmetic work — which is an argument that the delimiter must earn its place field by field, not one that grants or refuses it everywhere at once.
- **The identifier citation the convention permits applies here in full** — `Architecture.md § D3`, whose heading on 3 August 2026 ran *D3 — Valhalla as the default routing engine; openrouteservice as a registered adapter*. This is not a shortcut but the more stable of the two citations: in an architectural decision the identifier is the durable part of the heading and the tail is the volatile part, so an identifier citation survives a rewording that would falsify a verbatim one — as the date on that quotation concedes. Citing such a heading whole would also run every ADR reference past a hundred characters, permanently overhanging the wrap, and would couple the corpus to titles that are meant to be rephrasable. The identifier is then **the whole atomic token**, and the wrap treats it as one. **The bound is narrow, or this becomes licence to abbreviate anything:** it applies only where the heading itself carries a stable identifier as its leading token — `Architecture.md`'s `D1`–`D7` — and never to shortening an ordinary section name to its opening words. **That refusal is one half of a rule**, and the half binding the heading's author lives in `docs/README.md § Conventions`: a heading that cannot be cited whole and carries no identifier is a defect in the heading, not a licence to trim it. Read here alone the bound looks like a dead end, and the two moves it invites at the moment someone meets one — trimming the heading, or dropping the citation into prose — are both refused there.
- **`Source` names the section that creates the obligation** — the place a reader goes to challenge whether this requirement should exist. A constant cited *inside* the statement is not a `Source` entry: `CalculationSpecification.md` supplies the value, it does not impose the duty. A record citing a calculation section as its `Source` is usually one whose real source was never found.
- **A citation rots when the *cited* file changes, and nothing in the citing file moves.** The obligation lives in `docs/README.md § Conventions` — what the check asks, why heading-existence alone is not it, which party owes the repair, and where it is enforced — and **it is not restated here**, having moved there on 7 August 2026 once it was clear that a duty about citations cannot be scoped to the corpus. Two of its clauses land on this corpus by name, so read it rather than assuming this file has covered you: one falls on whoever edits **this file**, since the corpus cites these sections by heading and a rename or a re-split here breaks those citations; and one is `@requirements-librarian`'s, whose scope that section states and this one does not.

## Three project overrides a generic IEEE habit gets wrong

Each has a home elsewhere; only the consequence for a requirement record is stated here.

1. **Cite constants, never restate them** — `docs/README.md § Conventions`, `safety-path-checklist § Non-negotiables`. For a record: a threshold appears as a reference to its section, never as a number, in the statement *and* in the acceptance criteria.
2. **A proposed constant must never harden into a `MUST`** — `safety-path-checklist § The proposal boundary`, which also forbids an unexplained literal and forbids quoting a proposal to a user as established. For a record: the obligation is that the system reads the configured value and carries its documented origin, never that the value is a particular figure.
3. **Never infer a Turf mechanic** — `safety-path-checklist § Domain facts are verified, not inferred`. For a record: a domain assertion with no traceable source is a question for the Owner, not a requirement — do not write the record.

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
| `README.md` | Front door: what the corpus is, the **category register**, the **term register**, the ID allocation ledger, and the folder's invariants |
| `INDEX.md` | One index table per category, carrying a row for every ID ever issued |
| `<category>.md` | One file per category; every record lives in exactly one |
| `TRACEABILITY.md` | The matrix, both directions |

**Why README and INDEX are separate files.** The allocation ledger is the only guard against ID reuse, and it must not sit below fifteen screens of table where an author stops scrolling. The register and the ledger change rarely; the index changes on every filing. Keeping them apart means a routine filing does not rewrite the one file the whole RE family reads first.

**The record is authoritative; every index is derived.** Where a row in `INDEX.md` and the record in a category file disagree, **the record is right and the row is stale** — a librarian finding, fixed in the index, never fixed by editing the record to match. This is the rule that makes it safe to regenerate an index without reading it as a second source of truth.

**Category files.** A file is named as the kebab-cased `Category` field, verbatim — `Candidate identification` → `candidate-identification.md` — so a record names its file and a file names its category with nothing to look up. No numeric prefixes: a file name never encodes order, so adding a category never renames a file. FRs and NFRs share the folder, never a file. Inside: the category name, one line of scope, then one record per requirement in the canonical form above, **ID ascending, an order that never changes** — not on retirement, not on re-prioritization.

**`INDEX.md`** carries a TOC, then one section per category holding this table:

| ID | Title | Category | Priority | Status | Verification | Resolved-by |
|---|---|---|---|---|---|---|
| `FR-000` | Cap candidates promoted to full evaluation | Candidate identification | MUST | to-build | test | #14, #15 |

`Verification` is the keyword alone; the evidence sentence stays on the record. `Category` is redundant under a category heading and kept anyway, so a `grep FR-000` hit answers the whole question without its surrounding lines.

**`Resolved-by` reads `—` where no story is allocated**, on the record and in the index alike, and it means *not yet allocated* — never *unknown*. One spelling in both places, because the moment the record and its view word the same emptiness differently, a reader has to decide whether the difference is meaningful.

**Tombstones.** A retired requirement does not move and is not deleted: the record stays in ID order reading `Status: retired`, and its index row stays. **The index carries a row for every ID ever issued** — that ledger duty outranks its skim duty. `README.md` carries one line, `Next free: FR-078 · NFR-023`, and **allocation is highest-ever-issued plus one, never lowest-unused**, so a retirement cannot free a number even if a tombstone is lost. A retired record leaves `TRACEABILITY.md`, so its `Resolved-by` — on the record and in the index — **freezes at its last value** rather than reverting to `—`: it is the historical record of what was built for it, and blanking it would erase the only remaining trace.

### What the corpus may hold about a story or an epic

**What the corpus may hold about a story or an epic.** Requirements are documented in this repository, under `docs/Requirements/`. **Stories exist only as GitHub Issues, and epics only as GitHub Milestones** — neither is ever *documented* here. The corpus stores their **identifiers as links**, never their content and never their state: a story's **number** (`#14`) and its epic's **Milestone name** are the whole of what may appear, anywhere in the folder. A story's **title, body, acceptance criteria, assignee, labels or status** may not — not in `TRACEABILITY.md`, not in a record's `Resolved-by`, not in `INDEX.md`, and not in a column added later because it was convenient at the time.

The reason is one home per fact. A story's content and its state are owned by the board and the issue, which answer them live and answer them correctly; a copy here is stale from the moment it is written and leaves a reader holding two answers with no way to tell which is current. This is why *Story → requirement* carries **no board-status column**: the board answers status, the issue answers content, and this file answers one question only — which requirements a story resolves. **The rule has no exception.** Anything proposing one — a status copied from the board, a label mirrored here, a column added because it was convenient at the time — is a violation of it, however carefully the copy is scoped.

**`TRACEABILITY.md`** holds one table per direction: *Requirement → story* (requirement · epic Milestone · stories), one row per non-retired requirement; and *Story → requirement* (story # · epic · resolves), transcribed from each issue's `Resolves:` line, which stays the source of truth and is never edited to fit. The first table is the single home for the requirement→story link — the record's `Resolved-by` and the index column are views of it, regenerated in the same pass. **A disagreement between the two directions is a finding for the RE, never a silent reconciliation.**
