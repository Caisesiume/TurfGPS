---
name: local-gates
description: The single source of truth for running TurfGPS's local quality gates before any PR. Documentation gates are live now; code gates activate with the Go and frontend stacks. Use whenever an agent must prove work holds up mechanically.
---

# Local Gates — One Source of Truth

**All gates green BEFORE a PR exists.** The bench reviews quality; the gates prove the work merely holds together. Do not spend the bench on what a gate could have caught.

The repository is currently documentation-only, so the documentation gates below are the live ones. The code gates are written now so they are not invented under pressure later, and they activate the moment their stack exists.

---

## Documentation gates — **live now**, required on every PR touching `docs/`

The documentation set depends on three mechanical properties. Each has been broken at least once. Two are cheap to check; the first stopped being cheap when the citation convention landed, and the note below the list says exactly how much of it can be run today.

1. **Every citation resolves, and every citation is one token.** A citation is `Document.md § Section` — self-contained, with the `§` **inside** its delimiters — or `§ Section` with no target, which means a heading in the citing file. Where the cited heading carries a stable identifier, the identifier alone is the citation: `Architecture.md § D1`. The rule and its reasoning live in `docs/README.md § Conventions`, and this gate checks that rule rather than restating it. Two shapes are defects on sight: a **bare section name**, and a **filename sitting outside the delimiters** — the superseded form. A rename that silently orphans a reference is the common failure; a citation naming no document is the one that hides, because under the current rule it no longer reads as ambiguous, it reads as *this file*, and it orphans there.

   **The delimiter is decided by the file, and only two classes are settled.** The four narrative documents — `SPECIFICATION.md`, `CalculationSpecification.md`, `Architecture.md`, `DESIGN.md` — are read rendered, so their delimiter is *italics*. The requirements corpus's record fields sit inside fences where nothing renders, so theirs is the code span, with `Source` and `Depends-on` carrying no delimiter at all, per `.claude/skills/requirements-authoring/SKILL.md § IDs and citations`. **Every other file under `docs/` is unsettled, and this gate does not check its delimiter** — it checks that the citation resolves and that it is one token, neither of which any delimiter changes.

   **The pattern check runs against converted files only, and today that is none of them.** Deciding citation-from-emphasis by matching `§` inside a span is licensed solely for files on the **converted-file list** in `docs/README.md § Conventions`. That list is empty as of 4 August 2026, so **no file is eligible for it and the pattern check may not be run at all**: every citation must be resolved against the cited document's heading list, one at a time. A file joins the list in the same commit that converts it, so the cheap check becomes available file by file and never before. Running the pattern over an unconverted file reports its ordinary emphasis as citations and its superseded citations as nothing.
2. **No model is stated twice.** A formula, constant, or threshold lives in `CalculationSpecification.md` and nowhere else. Check for a sentence appearing in two documents — that is how the anti-duplication rule dies.
3. **Every mermaid diagram parses.** A diagram that fails to render is invisible on GitHub and nobody notices until a reader reports a blank block.

> **No runner exists yet.** These were last run by hand on 31 July 2026 and returned: 0 unresolved references, 1 intentional cross-document repeat (the reference-convention line each preamble carries), 7/7 diagrams parsing. **Building a runner is owed work** — until it exists, state in the PR body how each gate was checked and what it returned. An unstated gate is an unrun gate.
>
> **Gate 1 has never been run in the form described above, and this is the gate saying so rather than a reader discovering it.** That 31 July 2026 run predates the citation convention, which was recorded on 4 August 2026 in `2ea7395`; its "0 unresolved" answers whether references resolved under the rule of the day and says nothing about token form, so it cannot be carried forward as a pass. Of gate 1's three parts, **one is runnable today and two are not**: resolution can be checked by hand against heading lists, at the cost of opening every cited file; the **pattern check cannot be run against anything**, because the converted-file list that licenses it is empty; and **delimiter conformance cannot be judged outside the four documents and the corpus's record fields**, because no rule has been ratified for the rest. A PR reporting gate 1 states which of the three it ran.

---

## Code gates — **dormant** until the stacks exist

### Backend (Go) — required on every PR with Go changes, per D1

```bash
gofmt -l .              # clean (empty output = pass)
go vet ./...
golangci-lint run       # 0 issues
go test -race -count=1 ./...
go build ./...
```

The race detector is not optional on this codebase. D1 chose Go specifically for a long-lived stateful service holding many concurrent solve sessions with bounded worker pools over the candidate fan-out — concurrency is the reason the language was picked, so it is the thing most likely to break.

### Frontend (Vite + React) — required on every PR touching the client, per D2

```bash
npm run build   # tsc + vite build, no errors
npm run lint    # 0 issues
npm run test    # all pass
```

### When these activate

The first Go or frontend PR should also introduce a **`Makefile` at the repository root** as the canonical gate runner, and this skill then points at the Makefile targets rather than listing commands. Agent prompts that duplicate command lists drift; a Makefile does not. Until that PR, the commands above are the list.

---

## The law

1. **Report results verbatim in the PR body.** For documentation: `refs: N checked / 0 unresolved, method: headings | duplication: none | mermaid: N/N`. For code: `fmt: clean | vet: PASS | lint: 0 | test: PASS | build: SUCCESS`. **The reference line names its method**, because gate 1 now has two and only the expensive one is licensed — a bare `0 unresolved` no longer says whether anything was opened.
2. A gate that fails in a way you do not understand is a **stop-and-report**, not a retry-until-green.
3. A gate you skipped is reported as skipped, with the reason. Silence reads as a pass, and that is how an unrun gate becomes a merged defect.
