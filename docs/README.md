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

As of 4 August 2026 the set is in this shape, less one document that does not yet exist.

`Concept.md` held the authoritative product definition at roughly 1,670 lines, having grown past its purpose because formulas, thresholds, integration facts, and interaction flow had nowhere else to live. It was **split** on 31 July 2026, as a move operation rather than an authoring one, into `SPECIFICATION.md`, `CalculationSpecification.md`, `Architecture.md`, and `DESIGN.md`. It no longer exists; it survives in git history.

- **`SPECIFICATION.md`** is the source of truth for intent. The repository contains none of the system as code, so the documents lead the implementation, not the other way round.
- **`CalculationSpecification.md`** is complete for the first release, less a domain glossary it still owes.
- **`Architecture.md`** carries binding technology decisions and the Turf API facts. It still owes failure handling, observability, security, and schema.
- **`DESIGN.md`** carries the full interaction flow. It still owes the visual layer.
- **`Requirements/`** exists and is being authored — on 4 August 2026 its records were filed across ten categories, with the live count in its own `INDEX.md`. It is **no longer the bottleneck** for the project board described in `DELIVERY.md`: its signed-off records have been cut into Epics and stories, and the board is stocked. Authoring continues alongside the work the board now carries, still far short of the volume anticipated in `§ Requirements/` above. **`DEPLOYMENT.md`** does not exist yet.

Each document ends with what it still owes and the open questions it owns. Those are the shortest route to what is unfinished.

## Conventions

**Every model has exactly one home.** A formula, constant, or threshold is stated in `CalculationSpecification.md` and nowhere else; other documents reference it by section name. This is not tidiness — a second statement of a model is a second thing to keep correct, and the two will diverge.

**A citation is one self-contained token, and the `§` sits inside its delimiters.** A cross-document citation names its target within the delimiters — as the four documents write it, *SPECIFICATION.md § Park-and-walk zones* — and a same-file citation omits the target and opens on the marker — *§ Current state*. The marker being inside is the whole of the design: **a delimited span containing `§` is a citation, and one without it is emphasis.** The two are then decidable by looking at the span — no list of headings to consult, and no filter to separate citations from the ordinary emphasis sharing their delimiter. It is also what makes a bare cross-document citation *unwritable* rather than merely discouraged: the target-less form already means *this file*, so no spelling is left over for "some other document, unnamed".

The form kept here until now put the filename outside the delimiters — *Park-and-walk zones* in `SPECIFICATION.md`. It carries the same information and fails the same test: the italic half cannot be told from emphasis without opening the other file and reading its headings. On 4 August 2026 the four documents held **152 italic spans carrying no filename** — each either a same-document citation or ordinary emphasis, with nothing in the syntax to say which. That is the cost the token form removes, and it is why the token form was chosen over it.

**A filename inside a citation carries the citation's delimiter and no other.** Write *Architecture.md § D1*, never a code-spanned filename nested inside the italics. A filename standing alone in prose, outside any citation, keeps the code span it has always had — `Architecture.md`.

Five consequences of the token being atomic:

- **`above` and `below` sit outside it** — *§ Conventions* above — because they say where the reader will find the thing, which is prose about the citation rather than part of it.
- **A status suffix normalises away.** The `D6` heading in `Architecture.md` ends ` — *Proposed*`; a citation drops that tail. A suffix records the decision's standing on the day it was read, so a citation carrying one falsifies itself the moment the decision is ratified — and it also nests a second pair of italics inside the first. **A heading carrying a status suffix is better cited by its identifier.**
- **Where a heading carries a stable identifier, the identifier alone is the citation** — *Architecture.md § D6* — and it is then the whole atomic token.
- **A line break never falls inside a token**, at any of its seams.
- **A heading is citable only as a plain-text label.** A heading holding a `§` or a delimited span cannot be cited whole without nesting one token inside another, and it may not be shortened to fit — `requirements-authoring § IDs and citations` refuses that outright. So **every heading is either a plain-text label or carries a stable leading identifier**, and one that is neither cannot be cited by anything.

**The fifth is a duty on the heading's author, and it is the half that was missing.** `@docs-writer` found the gap on 5 August 2026 by walking into it: repairing bare identifiers, it wrote a citation *into* a heading, discovered it could not then cite that heading, and undid the change. Nothing was broken at the time — all thirty citations then in the repository resolved — because the twelve headings carrying a qualifier happened to be ones nobody had yet needed to name. That is the whole of what made it latent, and it is not a property anyone can rely on twice.

**It is not a new mechanism.** The identifier consequence above already is the general resolution: an identifier is what buys a heading the right to carry a volatile tail and stay citable. The status suffix normalising away is a *case* of that rather than a precedent of its own — both headings in the repository carrying a suffix, `D6` and `D7`, also carry identifiers, so the normalisation has never once been the operative rule. What was absent was the duty pointing the other way, at whoever writes the heading rather than whoever cites it.

**A rule normalising any trailing qualifier away was considered and refused**, and the reasoning is recorded because it is the attractive answer and will be proposed again. It would hand the citation's author the decision of where substance ends and qualifier begins — precisely the decision the identifier bound refuses to delegate, and it would delegate it to the party under time pressure. The em-dash does four different jobs here: it separates an identifier from a title, a title from a subtitle, a title from a status, and — in `safety-path-checklist`'s heading `What the data cannot verify — and must not be claimed` — a title from the operative half of its own meaning, where dropping the tail drops the injunction. It would also reach nothing, since a `§` or a span may sit anywhere in a heading and not only after a dash. And it would turn heading-existence, the check `requirements-authoring § IDs and citations` insists on running by lookup, from an equality test into a prefix match.

**The check is a grep over a file's own headings**, run when a heading is written rather than when someone fails to cite it. On 5 August 2026 it condemned three headings, all in skill files, and they were repaired in the commit recording this rule. Two of the three had already done damage: `requirements-authoring` needed to cite the board's Priority field twice, could not, and wrote *the Priority field in the `turfgps-board-ops` skill* instead — a bare section reference, which `local-gates § Documentation gates` calls a defect on sight. **That is the failure mode to expect.** The predicted one was that a citation's author would shorten the heading; the observed one is that they dissolve the citation into prose, which breaks the same rule while reading perfectly and leaving nothing for a checker to catch.

**A skill is cited by its name, and the name stands where a filename would** — `safety-path-checklist § The proposal boundary`. **The name resolves by convention to `.claude/skills/<name>/SKILL.md`**, and that substitution is the whole of the resolution: a checker following a skill citation to the file it names needs nothing beyond it. A citation carrying the path in place of the name is a defect rather than a longer-winded equivalent, and a skill named in prose outside any citation keeps the code span every bare name has — `safety-path-checklist`.

This is **the identifier rule above reaching the token's other half**, which is why it is recorded here and not as a case of its own. There, a heading carrying a stable identifier is cited by that identifier because the identifier is the durable part of the heading and the tail is the volatile part. The same split runs through the target: a skill's **name** is how every agent invokes it, it is the `name` field of the skill's own frontmatter, and it survives the file being moved — while the path around it is a fixed directory and a fixed filename that carry nothing the name does not, and that change without the skill changing. One stable key standing for one location, in both halves of the token.

**A bare filename cited from inside `docs/Requirements/` resolves sibling-first.** `INDEX.md`, `TRACEABILITY.md`, and `README.md` written in a corpus file mean the corpus's own three files, never `docs/`'s. This is not a special case granted to the corpus — it is what the citations there already mean, consistently, and it is the only reading under which they are true: `README.md § Category register` and `README.md § How this folder works` name headings that exist in `docs/Requirements/README.md` and in no other README. Where a corpus file means *this* file it writes the path in full — `docs/README.md § Conventions` — and every occurrence in the corpus already does, so the two are distinguishable today without any file being edited.

**The rule is a property of the resolver, not of the citation.** Nothing is missing from `INDEX.md` and nothing needs adding to it; a checker that resolves it from the repository root and reports a dangle has found a defect in itself and filed it against the corpus. @requirements-engineer built exactly that resolver and it reported three false dangles, one of them to a file sitting beside the document citing it. It is recorded here rather than left to be re-derived because **the tempting response to a screaming checker is to quieten it**, and loosening a resolver until the false dangles stop is precisely how the noise filter that once hid a real orphan came to be written. The correct repair is the resolution order. Widening the tolerance would trade three false negatives for an unknown number of true ones, which is the worse trade every time.

**Two renderings, one token.** The four documents are read rendered, so their delimiters are italic. The requirements corpus's record fields sit inside fences where no markdown renders at all, so grep-ability is the only criterion that applies there and the delimiters are code spans. That half is owned by `requirements-authoring § IDs and citations` — including the two fields the delimiters are deliberately kept off — and is not restated here. Nothing about the token itself differs between them.

**Every other file takes the code span, and a property decides it rather than a tally.** This file, `DELIVERY.md`, `HANDOFF.md`, the corpus's own `README.md`, `INDEX.md` and `TRACEABILITY.md`, the skill files under `.claude/skills/`, and the agent definitions under `.claude/agents/` are all read rendered, and none is one of the four documents, so neither rendering's stated reason reaches them. What does reach them is how they are used: **these are working documents, consulted rather than read through.** A reader arrives at one by searching for a term, or by loading it whole into an agent's context — never front to back — so the criterion that settles the corpus's record fields, that a citation be findable by `grep`, applies here in its own right rather than because a fence suppresses rendering. That is the third class, and its delimiter is the code span.

**The agent definitions were in no class at all, and this is where they land.** The 46 files under `.claude/agents/` on 4 August 2026 were named by neither rendering's reason and by no list, so nothing said what a citation in one should look like, and nothing said what to check one against. The property above absorbs them without amendment — they are consulted rather than read through, exactly as the skill files are — and that it needs no amendment to reach them is a point in its favour. Recorded here so the next reader does not re-derive it.

**No file in this class is ever eligible for the converted-file list, and the reason is that its citations are repaired rather than migrated.** The token form has bound every file in the repository since it was recorded on 4 August 2026 in `2ea7395`; for this class only the delimiter was open, and the paragraph above closes it. A citation here carrying its filename outside the delimiters therefore breaks a rule that already applied to it. It is a **defect** — not the residue of a superseded convention this class once followed, because that convention was the four documents' and this class never held it. Repairing a defect is not a migration, and only a migration puts a file on the list below. That keeps the list meaning exactly one thing — a file whose citations were **migrated** from a form it legitimately held to the new token — and a list meaning one thing is the only kind that can license a check.

**The migration runs per file, and only as files are touched.** A file converts **whole**, in a commit already editing it for another reason — never as a sweep across the set. The unit is the file rather than the corpus because a **half-converted file cannot be checked by looking at it**: a reader finding one span with a marker and one without cannot tell an unconverted citation from a deliberate emphasis. A half-converted *set* has no such defect, because each file answers for itself — provided it is known which files have converted.

**A file born in the token form joins by declaration, and that is the second and last route onto the list.** The paragraph above admits a file by **migration**, and a file that never held the old form was never migrated — so the corpus's newest and cleanest files were permanently ineligible, while the list stayed empty and licensed nothing. That is the list meaning exactly one thing at the cost of the check reaching nothing, and the cost runs the wrong way. **A file whose citations are token-form from its first commit joins the list in that same commit, declared in the commit message.** Raised by `@requirements-engineer` on 6 August 2026, when batch 4 created five such files, and ratified by the Owner the same day.

**The two routes admit on the same evidence and differ only in when it is gathered.** A migration establishes token-form completeness by converting; a declaration establishes it by inspection at creation. Neither admits a file whose state is merely assumed — **a file joins on a check that was actually run**, and a declaration made without running it is precisely the drift the list exists to prevent, wearing the form that licenses it.

**The list keeps its name, which now records how it began rather than how it is joined.** A born-token-form file was never converted, so the name is inexact for one of the two routes. Renaming it would touch `requirements-authoring § IDs and citations`, which cites it by that name, and the ruling admitting the second route used the name too — so the mismatch is stated here instead of chased through three files. What the list means is unchanged and is the only thing a reader needs: **every citation in a file on it is token-form.**

**Converted files:**

- `docs/Requirements/journey-initialization.md` — born token-form
- `docs/Requirements/multi-leg-budget-allocation.md` — born token-form
- `docs/Requirements/recommendation-disclosure.md` — born token-form
- `docs/Requirements/route-alternative-generation.md` — born token-form
- `docs/Requirements/safety-exclusions.md` — born token-form

**Those five were admitted one commit late, and it is recorded rather than smoothed over.** They were created in `1bb42e5`, before this route existed, and admitted in the commit signing batch 4 off — on a check run at admission rather than at creation. That is the whole of the exception and it is not a precedent: a file born after 6 August 2026 declares in its creating commit or does not join.

That list is what licenses the syntactic check: citation-or-emphasis may be decided by pattern **only for files on it**, and every file not on it must still be checked against the cited document's heading list. **A file joins the list in the same commit that converts it, and in no other commit.** A list that can drift from the files it describes is worse than no list, because the check it licenses would then run against files that cannot bear it.

`70e035f` converted four corpus files' prose citations to code spans on 4 August 2026 and is **not** a conversion under this convention: it settled which delimiter a citation carries, not what the citation contains, and all four files still hold citations with the filename outside the delimiters.

**Numeric constants are proposals unless stated otherwise.** A proposed default exists so that specification work begins with a concrete number to argue against rather than a blank to fill; disagreeing with one is useful, leaving it undecided is not.

**Facts about the Turf API were verified against the live API**, not assumed. Several contradict the obvious guess. Where a document states an API behaviour, it was checked. Where something remains inferred, it says so.

**Open questions are catalogued, not silently decided.** Each document carries the open questions belonging to its own content, rather than one shared list.
