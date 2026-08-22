---
name: local-gates
description: The single source of truth for running TurfGPS's local quality gates before any PR. Documentation gates are live now; code gates activate with the Go and frontend stacks. Use whenever an agent must prove work holds up mechanically.
---

# Local Gates — One Source of Truth

**All gates green BEFORE a PR exists.** The bench reviews quality; the gates prove the work merely holds together. Do not spend the bench on what a gate could have caught.

**Which gates are live is derived from the tree, not asserted here.** The documentation gates are always live. **A stack's gates activate the moment its manifest exists** — `service/go.mod` turns the backend gates on, the frontend manifest (`web/package.json`) turns the frontend gates on — so the code gates below are written now to be found rather than invented under pressure, and no edit to this file is needed to enable them. Check before you report: a gate you skipped because a document said the stack was dormant is an unrun gate reported as inapplicable.

---

## Documentation gates

**Live now**, and required on every PR touching `docs/`.

The documentation set depends on three mechanical properties. Each has been broken at least once. Two are cheap to check; the first stopped being cheap when the citation convention landed, and how cheaply it can be run is decided file by file rather than once for the whole run.

1. **Every citation resolves, and every citation is one token.** A citation is `Document.md § Section` — self-contained, with the `§` **inside** its delimiters — or `§ Section` with no target, which means a heading in the citing file. Where the cited heading carries a stable identifier, the identifier alone is the citation: `Architecture.md § D1`. The rule and its reasoning live in `docs/README.md § Conventions`, and this gate checks that rule rather than restating it. Three shapes are defects on sight: a **bare section name**, a **filename sitting outside the delimiters** — the superseded form — and a **skill cited by its path**. A skill is cited by its **name**, standing where a filename would, and the name resolves by convention to `.claude/skills/<name>/SKILL.md`; following one therefore costs nothing, which is what makes the path form a defect rather than a longer-winded equivalent. A rename that silently orphans a reference is the common failure; a citation naming no document is the one that hides, because under the current rule it no longer reads as ambiguous, it reads as *this file*, and it orphans there.

   **A heading is checked as a citation target, not only as a heading.** One holding a `§`, a code span, or an italic span, and carrying no stable leading identifier, cannot be cited by anything: it may not be cited whole without nesting one token inside another, and it may not be shortened to fit. The rule and its reasoning live in `docs/README.md § Conventions`, and this gate checks that rule rather than restating it. This file's own headings were shortened on 5 August 2026 to satisfy it — the qualifier belongs in the line beneath, which is also where it reads better. The check is a grep over the file's own headings and it costs nothing; what makes it worth stating as part of gate 1 is that the defect is invisible from the heading's own file, and the agent who finds it is the one who needed the citation and no longer has time to fix the heading.

   **The delimiter is decided by the file, and all three classes are now settled.** The four narrative documents — `SPECIFICATION.md`, `CalculationSpecification.md`, `Architecture.md`, `DESIGN.md` — are read rendered, so their delimiter is *italics*. The requirements corpus's record fields sit inside fences where nothing renders, so theirs is the code span, with `Source` and `Depends-on` carrying no delimiter at all, per `requirements-authoring § IDs and citations`. **Every other file is a working document — consulted rather than read through — and its delimiter is the code span**, per `docs/README.md § Conventions`, which reaches the rest of `docs/`, the skill files under `.claude/skills/`, and the agent definitions under `.claude/agents/`. This gate checks delimiter conformance in all three classes; resolution and one-token form it checks regardless, neither being changed by any delimiter.

   **The pattern check is licensed per file, and which files hold that licence is not recorded here.** Deciding citation-from-emphasis by matching `§` inside a span is licensed solely for files on the **converted-file list** in `docs/README.md § Conventions`. **Read the list there, at the moment you run the gate.** Files join it one at a time, in the commit that converts or declares each one, so eligibility is a property of a named file rather than a state the set is in — and a count, a date, or a claim about what the list currently holds, restated here, would be a second home for a fact that moves without this file moving. This paragraph carried exactly that and went stale, telling agents to skip a check they were licensed to run. For every file not on the list, each citation is resolved against the cited document's heading list, one at a time. Running the pattern over a file that does not hold the licence reports its ordinary emphasis as citations and its superseded citations as nothing, which is why the licence is looked up per file and never assumed for the run.

   **The fourth part looks the other way, at the citations pointing *in*.** The three above check this PR's citations against the world; this one checks the world's citations against this PR. A commit that **renames a heading, re-splits a section, or retracts what one of them asserted** falsifies every citation naming it — in files it does not touch, with nothing in those files moving, so there is no diff to read and no reader of a citing file who could tell. The check is therefore: for each heading this diff renames, re-splits, or empties of the claim it held, grep the repository for citations naming it and confirm the repairs are in **this same diff**. The obligation, the two questions the check asks, and why this gate is its enforcement point all live in `docs/README.md § Conventions`, and this gate enforces that rule rather than restating it. **The part is conditional, and its absence is reported rather than omitted:** a PR renaming no heading and retracting no claim reports it `n/a`, which says the diff was looked at — an omitted line says only that it was omitted.
2. **No fact with one designated home is stated twice.** A fact stated in the one place designated as its home is **cited** everywhere else and restated nowhere. The formula-constant-threshold case is the original and the strictest — `docs/README.md § Conventions` puts every model in `CalculationSpecification.md` and nowhere else, and this gate checks that rule rather than restating it — but nothing in the failure is peculiar to numbers. A fact about a convention, a location, or a procedure has a home in exactly the same sense, and a paraphrase of one rots in exactly the same way: silently, with no diff in the file carrying it.

   **The test, and it is answerable without adjudicating anything:** *does this sentence state something whose home is elsewhere, in a form that would go false if that home changed?* Both halves must hold. The question is about what the sentence does under a change, not about whether the fact is important enough to deserve a home — which is what keeps two authors applying it to the same sentence at the same answer. A check that produces argument instead of verdicts is worse than a narrow one that produces neither.

   **Two things it does not reach, said plainly so the widening does not become the argument.** **A citation is not a restatement** — naming the home, and saying in a clause what a reader will find there, is the remedy this gate prescribes and can never be the thing it condemns. **A claim with no designated home anywhere is out of scope** — the gate condemns a *second* statement, so where there is no first one it has nothing to weigh against, and deciding that a homeless claim ought to acquire a home is specification work rather than a gate finding.

   **Write time, because edit time is already covered and is already too late.** The rot obligation in `docs/README.md § Conventions` catches a paraphrase when the cited file is next edited — after an interval in which the paraphrase was read as true. This gate catches it as it is written, the one moment at which deleting it costs a sentence. **Never writing it is the primary defence and the rot obligation is the backstop**, in that order: a paraphrase that must exist gets rot-checked, and one that need not exist does not get written.

   **The narrow form could not condemn the defect that widened it, and that is the whole argument for widening it.** The stale sentence repaired in gate 1 above stated which files hold the pattern check's licence — a fact about a convention, not a formula, a constant, or a threshold — so this gate as written passed it on the day it was written and on every day it was wrong. Checking for one sentence appearing in two documents still finds the coarsest violations; it is no longer the whole of the check.
3. **Every mermaid diagram parses.** A diagram that fails to render is invisible on GitHub and nobody notices until a reader reports a blank block.

> **One runner exists and it covers one duplication class; every other part of every gate is still run by hand.** `make d8-claims` fails on a restatement of the root-run model that gate 2 requires to be cited instead. That is **one class inside gate 2** — not gate 2, and nothing at all of gates 1 and 3 — and the script's own header records what it cannot see, which is to be read before its result is reported. The by-hand gates were last run in full on 31 July 2026 and returned: 0 unresolved references, 1 intentional cross-document repeat (the reference-convention line each preamble carries), 7/7 diagrams parsing. **Building the rest is owed work** — until it exists, state in the PR body how each gate was checked and what it returned. An unstated gate is an unrun gate.
>
> **The 31 July 2026 run recorded above is not a gate 1 pass, and this is the gate saying so rather than a reader discovering it.** It predates the citation convention, which was recorded on 4 August 2026 in `2ea7395`; its "0 unresolved" answers whether references resolved under the rule of the day and says nothing about token form, so it cannot be carried forward. **No set-wide gate 1 run has been recorded since**, so a clean gate line on a PR is evidence about that PR's files and about nothing else.
>
> Gate 1 has four parts — resolution, delimiter conformance, the pattern check, and the inbound check — and **a PR reporting gate 1 states which of the four it ran, and over which files.** Resolution and delimiter conformance can be judged on any file. The pattern check reaches only the files holding its licence, so a PR whose files fall on both sides of the converted-file list has used both methods and reports both; naming only the cheaper one overstates what was opened. The inbound check ranges over the diff rather than over files, and **`n/a` is one of its results, not one of its silences** — it is the answer a PR gives when it renamed no heading and retracted no claim, and it is only worth anything because a PR that did rename one cannot give it.

---

## Code gates

**Each block is live exactly when its manifest exists** — `service/go.mod` for the backend, `web/package.json` for the frontend — and dormant only until then. Neither waits on an edit to this file.

### Backend (Go)

Live once `service/go.mod` exists. Required on every PR with Go changes, per `Architecture.md § D1`.

**Run the backend gates through the Makefile, from the repository root**, where the Makefile is.

```bash
make gates
```

`gates` runs the five backend gates — `fmt`, `vet`, `lint`, `test`, `build` — and each is a target of its own when only one of them is wanted. **What those targets execute is in the Makefile and is deliberately not repeated here**, for the reason `§ When these activate` below gives.

**The working directory is now the Makefile's job, and that is the single thing it exists to get right.** Run from the repository root, `Architecture.md § D8` leaves no module to resolve — and what that costs is not one uniform silence. **Four of the five are loud from the root, but loud is not the same as discriminating**, and the two come apart in a different place for each of two gates. Measured on 22 August 2026, all five from both directories, on the reference host — Windows, go1.26.2, `make` absent, so each gate was invoked exactly as its recipe invokes it:

- **vet** and **build** each exit **1** from the root against **0** from `service/`, naming the module they could not resolve. Loud, and discriminating.
- **lint** exits **7** from the root against **0** from `service/`, and prints **`0 issues.`** on stdout in both cases — byte-identical. Its entire discrimination is the exit status plus one `does not contain main module` line on **stderr**. A recipe that captures its stdout and drops its status reads the root run as a clean one.
- **test** is host-dependent, and a claim about it that names no host says nothing. On this host, which has no C compiler, it discriminates nothing: both directories refuse the race detector with `-race requires cgo` and exit **2**, that precondition being tested before any module is resolved. Drop the detector and the same gate fails at module resolution from the root and not from `service/`, so the resolution failure is present and merely masked. **On a host carrying a C compiler the detector reaches that step and this gate discriminates like the other three** — reasoned from the masked failure above, not executed here.
- **fmt** is the quiet one — exit 0, empty output, indistinguishable from a clean run — **and not because it read nothing**: it walks the filesystem instead of resolving a module, so it reads this tree by accident of layout, every Go file sitting inside the module directory, and it reads more the moment a `.go` file lands outside.

The old defence was an agent remembering `cd service` on every command line, and a shell that resets between commands is back at the root by the second one. `§ When these activate` below requires every recipe to carry its own directory, which replaces that memory with a property of the file — so which directory a gate measured is fixed by the recipe rather than by where the command was typed.

**`make gates` prints law 1's code line itself, derived from the run that produced it** — the Makefile's header is where that derivation is argued, and where the one field it cannot derive says so instead of inventing a number. **Paste that line into the PR body rather than composing one:** a composed line is a claim about a run, and this one is the run's own output.

The race detector is not optional on this codebase. `Architecture.md § D1` chose Go specifically for a long-lived stateful service holding many concurrent solve sessions with bounded worker pools over the candidate fan-out — concurrency is the reason the language was picked, so it is the thing most likely to break.

### Frontend (Vite + React)

Live once the client's manifest (`web/package.json`) exists. Required on every PR touching the client, per `Architecture.md § D2`.

**Every command below runs from `web/`**, where `Architecture.md § D8` puts the client.

```bash
cd web          # not optional, for the same reason as the Go block above
npm run build   # tsc + vite build, no errors
npm run lint    # 0 issues
npm run test    # all pass
```

npm resolves a script against the nearest `package.json`, and there is none at the repository root, so from there these three do not run the client's build, lint, or tests at all. That failure is at least loud: npm names a missing script. The Go block is mostly loud from the root too, but not uniformly, and its one quiet gate is quiet for a reason that says nothing about whether the tree was checked — `§ Backend (Go)` above records what each of the five does. This block is wrong from the root and says so, and `cd web` is the whole of its defence until a target covers this stack.

### When these activate

**The `Makefile` at the repository root is the canonical gate runner**, introduced by the first Go PR as this section required, and this skill now points at its targets rather than listing commands. Agent prompts that duplicate command lists drift; a Makefile does not. `§ Frontend (Vite + React)` above still carries its commands, because nothing yet runs them — that block's manifest does not exist, and the Makefile's own header records where its targets will come from when it does.

**The Makefile is the commands' one home, and every other file names the target instead.** Ten agent files carried these commands inline, every one of them without the working directory, because they were copied before `Architecture.md § D8` existed and nothing pulled them forward when it did — the ten had gone stale in the same commit that made this file correct. They now **name the gate they must pass and cite this skill for how to run it**. A file that reproduces a gate command is a defect on sight, however correct the copy looks on the day it is written: a second home for the model outlives whoever checked it, and the commands have moved once already — out of this file, which was their home only for as long as there was no Makefile. Reproducing them is licensed in exactly one case — a **reviewer quoting the one check it performs itself**, as an instrument of its own review rather than as the gate — and such a file says inline why the copy is there and cites this skill for the rest.

The root is the right home for it — `make` has no notion of a module, so the one file can drive both stacks — and **every recipe sets its own working directory**, which is a standing requirement on that file rather than a task that was completed once. A recipe that invokes the Go toolchain without one puts that gate back at the root, where `§ Backend (Go)` above records what each of the five actually reports, with the `cd` no longer visible to notice missing. The Makefile's value here is that it encodes each directory once, in the only place that cannot silently be run from somewhere else.

---

## The law

1. **Report results verbatim in the PR body.** For documentation: `refs: N checked / 0 unresolved, method: headings | inbound: n/a | duplication: none | mermaid: N/N`. For code: `dir: service | fmt: clean | vet: PASS | lint: 0 | test: PASS | build: SUCCESS`. **The reference line names its method**, because gate 1 has two and which of them is licensed is decided per file by the converted-file list in `docs/README.md § Conventions` — a bare `0 unresolved` no longer says whether anything was opened. A PR whose files fall on both sides of that list names both methods rather than the more flattering one. **The reference line also carries `inbound:`**, which is the one part of gate 1 that cannot be inferred from the files the PR lists: it ranges over what the diff did to headings rather than over what the diff's files cite, so a PR omitting it is indistinguishable from one whose author never looked at that question. It reads `n/a`, or the count of headings touched and citations repaired. **The code line names its directory** for exactly the same reason: the directory a Go command ran in is what decides which tree it measured, per `Architecture.md § D8`, and `§ Backend (Go)` above records what each of the five reports from the wrong one — so a report that omits where it ran is not evidence that anything was compiled.
2. A gate that fails in a way you do not understand is a **stop-and-report**, not a retry-until-green.
3. A gate you skipped is reported as skipped, with the reason. Silence reads as a pass, and that is how an unrun gate becomes a merged defect.
4. **A test bound to a `test`-verified acceptance criterion reports its red demonstration.** The rule is `docs/DELIVERY.md § Proof that a test can fail`, which is where it is ratified and where its scope and both its awkward cases are settled; **this is only the line it is reported on**, and it is here rather than there because law 1 above is already the one place that says what a PR body must carry as evidence. It is **not a gate** — a gate is one command over the whole diff, and this is one entry per criterion — so it is reported alongside the gate lines and never folded into them. One entry per `test`-verified criterion the PR claims to satisfy:

   ```
   <REQ-ID> · <TestName> · neutralised: <what was put back> · red: <the assertion's message>
   <REQ-ID> · <TestName> · demonstration owed → <story #>
   ```

   **The entry carries the assertion's own message and never the bare word `FAIL`** — that constraint is the one part of the form that is not free-text, and `docs/DELIVERY.md § Red for the wrong reason` is where it is argued rather than here. Writing the message is what makes the entry checkable by a reader; writing the verdict is what makes it unfalsifiable.

   A PR landing no such test carries no such entries, and says so once rather than omitting the line — laws 1 and 3 apply here unchanged.
