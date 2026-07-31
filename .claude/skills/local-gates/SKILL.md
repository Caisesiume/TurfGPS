---
name: local-gates
description: The single source of truth for running TurfGPS's local quality gates before any PR. Documentation gates are live now; code gates activate with the Go and frontend stacks. Use whenever an agent must prove work holds up mechanically.
---

# Local Gates — One Source of Truth

**All gates green BEFORE a PR exists.** The bench reviews quality; the gates prove the work merely holds together. Do not spend the bench on what a gate could have caught.

The repository is currently documentation-only, so the documentation gates below are the live ones. The code gates are written now so they are not invented under pressure later, and they activate the moment their stack exists.

---

## Documentation gates — **live now**, required on every PR touching `docs/`

The documentation set depends on three mechanical properties. Each has been broken at least once and each is cheap to check.

1. **Every cross-reference resolves.** An italic section name qualified with a filename must match a real heading in that file; an unqualified one must match a heading in its own file. A rename that silently orphans a reference is the common failure.
2. **No model is stated twice.** A formula, constant, or threshold lives in `CalculationSpecification.md` and nowhere else. Check for a sentence appearing in two documents — that is how the anti-duplication rule dies.
3. **Every mermaid diagram parses.** A diagram that fails to render is invisible on GitHub and nobody notices until a reader reports a blank block.

> **No runner exists yet.** These were last run by hand on 31 July 2026 and returned: 0 unresolved references, 1 intentional cross-document repeat (the reference-convention line each preamble carries), 7/7 diagrams parsing. **Building a runner is owed work** — until it exists, state in the PR body how each gate was checked and what it returned. An unstated gate is an unrun gate.

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

1. **Report results verbatim in the PR body.** For documentation: `refs: 0 unresolved | duplication: none | mermaid: N/N`. For code: `fmt: clean | vet: PASS | lint: 0 | test: PASS | build: SUCCESS`.
2. A gate that fails in a way you do not understand is a **stop-and-report**, not a retry-until-green.
3. A gate you skipped is reported as skipped, with the reason. Silence reads as a pass, and that is how an unrun gate becomes a merged defect.
