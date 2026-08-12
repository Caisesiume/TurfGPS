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

The documentation set depends on three mechanical properties. Each has been broken at least once. Two are cheap to check; the first stopped being cheap when the citation convention landed, and the note below the list says exactly how much of it can be run today.

1. **Every citation resolves, and every citation is one token.** A citation is `Document.md § Section` — self-contained, with the `§` **inside** its delimiters — or `§ Section` with no target, which means a heading in the citing file. Where the cited heading carries a stable identifier, the identifier alone is the citation: `Architecture.md § D1`. The rule and its reasoning live in `docs/README.md § Conventions`, and this gate checks that rule rather than restating it. Three shapes are defects on sight: a **bare section name**, a **filename sitting outside the delimiters** — the superseded form — and a **skill cited by its path**. A skill is cited by its **name**, standing where a filename would, and the name resolves by convention to `.claude/skills/<name>/SKILL.md`; following one therefore costs nothing, which is what makes the path form a defect rather than a longer-winded equivalent. A rename that silently orphans a reference is the common failure; a citation naming no document is the one that hides, because under the current rule it no longer reads as ambiguous, it reads as *this file*, and it orphans there.

   **A heading is checked as a citation target, not only as a heading.** One holding a `§`, a code span, or an italic span, and carrying no stable leading identifier, cannot be cited by anything: it may not be cited whole without nesting one token inside another, and it may not be shortened to fit. The rule and its reasoning live in `docs/README.md § Conventions`, and this gate checks that rule rather than restating it. This file's own headings were shortened on 5 August 2026 to satisfy it — the qualifier belongs in the line beneath, which is also where it reads better. The check is a grep over the file's own headings and it costs nothing; what makes it worth stating as part of gate 1 is that the defect is invisible from the heading's own file, and the agent who finds it is the one who needed the citation and no longer has time to fix the heading.

   **The delimiter is decided by the file, and all three classes are now settled.** The four narrative documents — `SPECIFICATION.md`, `CalculationSpecification.md`, `Architecture.md`, `DESIGN.md` — are read rendered, so their delimiter is *italics*. The requirements corpus's record fields sit inside fences where nothing renders, so theirs is the code span, with `Source` and `Depends-on` carrying no delimiter at all, per `requirements-authoring § IDs and citations`. **Every other file is a working document — consulted rather than read through — and its delimiter is the code span**, per `docs/README.md § Conventions`, which reaches the rest of `docs/`, the skill files under `.claude/skills/`, and the agent definitions under `.claude/agents/`. This gate checks delimiter conformance in all three classes; resolution and one-token form it checks regardless, neither being changed by any delimiter.

   **The pattern check runs against converted files only, and today that is none of them.** Deciding citation-from-emphasis by matching `§` inside a span is licensed solely for files on the **converted-file list** in `docs/README.md § Conventions`. That list is empty as of 4 August 2026, so **no file is eligible for it and the pattern check may not be run at all**: every citation must be resolved against the cited document's heading list, one at a time. A file joins the list in the same commit that converts it, so the cheap check becomes available file by file and never before. Running the pattern over an unconverted file reports its ordinary emphasis as citations and its superseded citations as nothing.
2. **No model is stated twice.** A formula, constant, or threshold lives in `CalculationSpecification.md` and nowhere else. Check for a sentence appearing in two documents — that is how the anti-duplication rule dies.
3. **Every mermaid diagram parses.** A diagram that fails to render is invisible on GitHub and nobody notices until a reader reports a blank block.

> **No runner exists yet.** These were last run by hand on 31 July 2026 and returned: 0 unresolved references, 1 intentional cross-document repeat (the reference-convention line each preamble carries), 7/7 diagrams parsing. **Building a runner is owed work** — until it exists, state in the PR body how each gate was checked and what it returned. An unstated gate is an unrun gate.
>
> **Gate 1 has never been run in the form described above, and this is the gate saying so rather than a reader discovering it.** That 31 July 2026 run predates the citation convention, which was recorded on 4 August 2026 in `2ea7395`; its "0 unresolved" answers whether references resolved under the rule of the day and says nothing about token form, so it cannot be carried forward as a pass. Of gate 1's three parts, **two are runnable today and one is not**: resolution can be checked by hand against heading lists, at the cost of opening every cited file; **delimiter conformance can now be judged on any file**, because the third and last class was settled in `docs/README.md § Conventions` on 4 August 2026 and no file is left outside the three; and the **pattern check cannot be run against anything**, because the converted-file list that licenses it is empty. A PR reporting gate 1 states which of the three it ran.

---

## Code gates

**Each block is live exactly when its manifest exists** — `service/go.mod` for the backend, `web/package.json` for the frontend — and dormant only until then. Neither waits on an edit to this file.

### Backend (Go)

Live once `service/go.mod` exists. Required on every PR with Go changes, per `Architecture.md § D1`.

**Every command below runs from `service/`**, where `Architecture.md § D8` puts the Go module — never from the repository root.

```bash
cd service              # not optional, and not once per session — see the note below
gofmt -l .              # clean (empty output = pass)
go vet ./...
golangci-lint run       # 0 issues
go test -race -count=1 ./...
go build ./...
```

**Run from the repository root, all five of these pass.** Each resolves against the working directory, has nothing to inspect there, exits zero, and prints what a clean tree prints — the board comes back `fmt: clean | vet: PASS | lint: 0 | test: PASS | build: SUCCESS` having read no code at all. `Architecture.md § D8` accepts that false pass as the price of the layout; what it costs *here* is that this one `cd` is the entire defence against it, so do not simplify it away. An agent issuing these one command at a time carries the directory into every one — `cd service && go vet ./...` — because a shell that resets between commands is back at the root by the second line, which is exactly where they all succeed against nothing.

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

npm resolves a script against the nearest `package.json`, and there is none at the repository root, so from there these three do not run the client's build, lint, or tests at all. That failure is at least visible, which the Go block's is not — the directory here buys a client that is genuinely checked rather than a report that is merely honest. Both blocks are wrong from the root; only one of them is quiet about it.

### When these activate

The first Go or frontend PR should also introduce a **`Makefile` at the repository root** as the canonical gate runner, and this skill then points at the Makefile targets rather than listing commands. Agent prompts that duplicate command lists drift; a Makefile does not. Until that PR, the commands above are the list.

**"The list" means this block and no other, and that is now enforced by there being no other.** Ten agent files carried these commands inline, every one of them without the working directory, because they were copied before `Architecture.md § D8` existed and nothing pulled them forward when it did — the ten had gone stale in the same commit that made this file correct. They now **name the gate they must pass and cite this skill for how to run it**. An agent file that reproduces the commands is a defect on sight, however correct the copy looks on the day it is written: a second home for the model outlives whoever checked it, and the gates are still owed a Makefile and a `service/` that exists, so the commands will move again. Reproducing them is licensed in exactly one case — a **reviewer quoting the one check it performs itself**, as an instrument of its own review rather than as the gate — and such a file says inline why the copy is there and cites this skill for the rest.

The root is the right home for it — `make` has no notion of a module, so the one file can drive both stacks — but **every recipe sets its own working directory**, and a recipe that invokes the Go toolchain without one reintroduces the false pass above with the `cd` no longer visible to notice missing. That is the single thing to get right in that PR: the Makefile's value here is that it encodes each directory once, in the only place that cannot silently be run from somewhere else.

---

## The law

1. **Report results verbatim in the PR body.** For documentation: `refs: N checked / 0 unresolved, method: headings | duplication: none | mermaid: N/N`. For code: `dir: service | fmt: clean | vet: PASS | lint: 0 | test: PASS | build: SUCCESS`. **The reference line names its method**, because gate 1 now has two and only the expensive one is licensed — a bare `0 unresolved` no longer says whether anything was opened. **The code line names its directory** for exactly the same reason: five passes over an empty root are character-for-character the five passes over a clean module, so a report that omits where it ran is not evidence that anything was compiled.
2. A gate that fails in a way you do not understand is a **stop-and-report**, not a retry-until-green.
3. A gate you skipped is reported as skipped, with the reason. Silence reads as a pass, and that is how an unrun gate becomes a merged defect.
4. **A test bound to a `test`-verified acceptance criterion reports its red demonstration.** The rule is `docs/DELIVERY.md § Proof that a test can fail`, which is where it is ratified and where its scope and both its awkward cases are settled; **this is only the line it is reported on**, and it is here rather than there because law 1 above is already the one place that says what a PR body must carry as evidence. It is **not a gate** — a gate is one command over the whole diff, and this is one entry per criterion — so it is reported alongside the gate lines and never folded into them. One entry per `test`-verified criterion the PR claims to satisfy:

   ```
   <REQ-ID> · <TestName> · neutralised: <what was put back> · red: <the assertion's message>
   <REQ-ID> · <TestName> · demonstration owed → <story #>
   ```

   **The entry carries the assertion's own message and never the bare word `FAIL`** — that constraint is the one part of the form that is not free-text, and `docs/DELIVERY.md § Red for the wrong reason` is where it is argued rather than here. Writing the message is what makes the entry checkable by a reader; writing the verdict is what makes it unfalsifiable.

   A PR landing no such test carries no such entries, and says so once rather than omitting the line — laws 1 and 3 apply here unchanged.
