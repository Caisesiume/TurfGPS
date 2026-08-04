# TurfGPS documentation

This folder is the front door to the TurfGPS documentation set. Each document answers one question, and the question is what decides where a piece of writing belongs.

## The documents

| Document | The question it answers |
|----------|------------------------|
| `SPECIFICATION.md` | What is the system, and how should it behave conceptually? |
| `docs/Requirements/` | What precisely must the system satisfy? |
| `CalculationSpecification.md` | How is every number the system produces worked out? |
| `Architecture.md` | How will the system feasibly satisfy those requirements? |
| `DESIGN.md` | What design qualities are required to satisfy those requirements? |
| `DEPLOYMENT.md` | How should this system operate with as little complexity as possible? |
| `DELIVERY.md` | How does work get tracked, reviewed, and shipped? |

### SPECIFICATION.md

Vision, potential commercial model, major user capabilities, terminology, behaviour, safety philosophy, conceptual invariants, provider strategy, operational expectations, and known product-policy questions.

This is the document someone reads to understand *what this product is*. It should be readable front to back in one sitting.

### Requirements/

The corpus lives at **`docs/Requirements/`**, inside this folder alongside the other documents. Every bare mention of `Requirements/` across the repository means that path.

Formal functional and non-functional requirements, constraints, acceptance criteria, verification methods, priority, source, risk, dependencies, and traceability IDs.

A **folder rather than a single file**, because the expected volume — on the order of 150–250 requirements — makes one file unnavigable exactly when it becomes load-bearing. Reference and dependency tables live as separate files, with a `README.md` inside the folder as its own front door.

Every requirement carries an explicit **verification method**. Much of this product's quality bar is human judgement rather than anything machine-checkable — whether a recommended route is a *good* Turf route cannot be asserted by a test — and the requirement must say so, or review will claim to have verified something it did not.

Traceability IDs matter beyond tidiness: issues on the project board cite the requirements they satisfy, and review agents check against them. See `DELIVERY.md`.

### CalculationSpecification.md

Every formula, constant, and threshold: candidate bounding, access classification, stop and journey time, takeover time, the value model, the difficulty model, and the objective function.

It exists because the optimization logic is central enough that it should not be buried inside the architecture document, and because a model needs exactly one home. Anything stated there is stated **only** there.

### Architecture.md

Component boundaries, runtime topology, scaling model, ports and adapters, extensibility, persistence model, data flows, security architecture, failure handling, and architectural decisions. It also owns the properties of the Turf API the system integrates against.

### DESIGN.md

User interaction design, user experience, graphic profile, the feelings the design should express, simplicity, user-flow diagrams, design visions, use-case detail, and page layouts.

### DEPLOYMENT.md

Operational detail: how to deploy, target OS, hosting options with price and complexity comparisons, pipeline and CI detail, full deployment architecture with diagrams, and operational direction.

## Current state

As of 1 August 2026 the set is in this shape, less one document that does not yet exist.

`Concept.md` held the authoritative product definition at roughly 1,670 lines, having grown past its purpose because formulas, thresholds, integration facts, and interaction flow had nowhere else to live. It was **split** on 31 July 2026, as a move operation rather than an authoring one, into `SPECIFICATION.md`, `CalculationSpecification.md`, `Architecture.md`, and `DESIGN.md`. It no longer exists; it survives in git history.

- **`SPECIFICATION.md`** is the source of truth for intent. The repository contains none of the system as code, so the documents lead the implementation, not the other way round.
- **`CalculationSpecification.md`** is complete for the first release, less a domain glossary it still owes.
- **`Architecture.md`** carries binding technology decisions and the Turf API facts. It still owes failure handling, observability, security, and schema.
- **`DESIGN.md`** carries the full interaction flow. It still owes the visual layer.
- **`Requirements/`** exists and is being authored — its first requirements are filed across four categories, with the live count in its own `INDEX.md`. It remains the bottleneck for the project board described in `DELIVERY.md`, now because it is far short of the expected volume rather than because it is absent. **`DEPLOYMENT.md`** does not exist yet.

Each document ends with what it still owes and the open questions it owns. Those are the shortest route to what is unfinished.

## Conventions

**Every model has exactly one home.** A formula, constant, or threshold is stated in `CalculationSpecification.md` and nowhere else; other documents reference it by section name. This is not tidiness — a second statement of a model is a second thing to keep correct, and the two will diverge.

**A citation is one self-contained token, and the `§` sits inside its delimiters.** A cross-document citation names its target within the delimiters — as the four documents write it, *SPECIFICATION.md § Park-and-walk zones* — and a same-file citation omits the target and opens on the marker — *§ Current state*. The marker being inside is the whole of the design: **a delimited span containing `§` is a citation, and one without it is emphasis.** The two are then decidable by looking at the span — no list of headings to consult, and no filter to separate citations from the ordinary emphasis sharing their delimiter. It is also what makes a bare cross-document citation *unwritable* rather than merely discouraged: the target-less form already means *this file*, so no spelling is left over for "some other document, unnamed".

The form kept here until now put the filename outside the delimiters — *Park-and-walk zones* in `SPECIFICATION.md`. It carries the same information and fails the same test: the italic half cannot be told from emphasis without opening the other file and reading its headings. On 4 August 2026 the four documents held **152 italic spans carrying no filename** — each either a same-document citation or ordinary emphasis, with nothing in the syntax to say which. That is the cost the token form removes, and it is why the token form was chosen over it.

**A filename inside a citation carries the citation's delimiter and no other.** Write *Architecture.md § D1*, never a code-spanned filename nested inside the italics. A filename standing alone in prose, outside any citation, keeps the code span it has always had — `Architecture.md`.

Four consequences of the token being atomic:

- **`above` and `below` sit outside it** — *§ Conventions* above — because they say where the reader will find the thing, which is prose about the citation rather than part of it.
- **A status suffix normalises away.** The `D6` heading in `Architecture.md` ends ` — *Proposed*`; a citation drops that tail. A suffix records the decision's standing on the day it was read, so a citation carrying one falsifies itself the moment the decision is ratified — and it also nests a second pair of italics inside the first. **A heading carrying a status suffix is better cited by its identifier.**
- **Where a heading carries a stable identifier, the identifier alone is the citation** — *Architecture.md § D6* — and it is then the whole atomic token.
- **A line break never falls inside a token**, at any of its seams.

**Two renderings, one token.** The four documents are read rendered, so their delimiters are italic. The requirements corpus's record fields sit inside fences where no markdown renders at all, so grep-ability is the only criterion that applies there and the delimiters are code spans. That half is owned by `.claude/skills/requirements-authoring/SKILL.md § IDs and citations` — including the two fields the delimiters are deliberately kept off — and is not restated here. Nothing about the token itself differs between them.

**Every other file takes the code span, and a property decides it rather than a tally.** This file, `DELIVERY.md`, `HANDOFF.md`, the corpus's own `README.md`, `INDEX.md` and `TRACEABILITY.md`, the skill files under `.claude/skills/`, and the agent definitions under `.claude/agents/` are all read rendered, and none is one of the four documents, so neither rendering's stated reason reaches them. What does reach them is how they are used: **these are working documents, consulted rather than read through.** A reader arrives at one by searching for a term, or by loading it whole into an agent's context — never front to back — so the criterion that settles the corpus's record fields, that a citation be findable by `grep`, applies here in its own right rather than because a fence suppresses rendering. That is the third class, and its delimiter is the code span.

**The agent definitions were in no class at all, and this is where they land.** The 46 files under `.claude/agents/` on 4 August 2026 were named by neither rendering's reason and by no list, so nothing said what a citation in one should look like, and nothing said what to check one against. The property above absorbs them without amendment — they are consulted rather than read through, exactly as the skill files are — and that it needs no amendment to reach them is a point in its favour. Recorded here so the next reader does not re-derive it.

**No file in this class is ever eligible for the converted-file list, and the reason is that its citations are repaired rather than migrated.** The token form has bound every file in the repository since it was recorded on 4 August 2026 in `2ea7395`; for this class only the delimiter was open, and the paragraph above closes it. A citation here carrying its filename outside the delimiters therefore breaks a rule that already applied to it. It is a **defect** — not the residue of a superseded convention this class once followed, because that convention was the four documents' and this class never held it. Repairing a defect is not a migration, and only a migration puts a file on the list below. That keeps the list meaning exactly one thing — a file whose citations were **migrated** from a form it legitimately held to the new token — and a list meaning one thing is the only kind that can license a check.

**The migration runs per file, and only as files are touched.** A file converts **whole**, in a commit already editing it for another reason — never as a sweep across the set. The unit is the file rather than the corpus because a **half-converted file cannot be checked by looking at it**: a reader finding one span with a marker and one without cannot tell an unconverted citation from a deliberate emphasis. A half-converted *set* has no such defect, because each file answers for itself — provided it is known which files have converted.

**Converted files:**

- *(none yet)*

That list is what licenses the syntactic check: citation-or-emphasis may be decided by pattern **only for files on it**, and every file not on it must still be checked against the cited document's heading list. **A file joins the list in the same commit that converts it, and in no other commit.** A list that can drift from the files it describes is worse than no list, because the check it licenses would then run against files that cannot bear it.

`70e035f` converted four corpus files' prose citations to code spans on 4 August 2026 and is **not** a conversion under this convention: it settled which delimiter a citation carries, not what the citation contains, and all four files still hold citations with the filename outside the delimiters.

**Numeric constants are proposals unless stated otherwise.** A proposed default exists so that specification work begins with a concrete number to argue against rather than a blank to fill; disagreeing with one is useful, leaving it undecided is not.

**Facts about the Turf API were verified against the live API**, not assumed. Several contradict the obvious guess. Where a document states an API behaviour, it was checked. Where something remains inferred, it says so.

**Open questions are catalogued, not silently decided.** Each document carries the open questions belonging to its own content, rather than one shared list.
