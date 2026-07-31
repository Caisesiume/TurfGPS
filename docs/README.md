# TurfGPS documentation

This folder is the front door to the TurfGPS documentation set. Each document answers one question, and the question is what decides where a piece of writing belongs.

## The documents

| Document | The question it answers |
|----------|------------------------|
| `SPECIFICATION.md` | What is the system, and how should it behave conceptually? |
| `Requirements/` | What precisely must the system satisfy? |
| `CalculationSpecification.md` | How is every number the system produces worked out? |
| `Architecture.md` | How will the system feasibly satisfy those requirements? |
| `DESIGN.md` | What design qualities are required to satisfy those requirements? |
| `DEPLOYMENT.md` | How should this system operate with as little complexity as possible? |
| `DELIVERY.md` | How does work get tracked, reviewed, and shipped? |

### SPECIFICATION.md

Vision, potential commercial model, major user capabilities, terminology, behaviour, safety philosophy, conceptual invariants, provider strategy, operational expectations, and known product-policy questions.

This is the document someone reads to understand *what this product is*. It should be readable front to back in one sitting.

### Requirements/

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

As of 31 July 2026 the set is in this shape, less two documents that do not yet exist.

`Concept.md` held the authoritative product definition at roughly 1,670 lines, having grown past its purpose because formulas, thresholds, integration facts, and interaction flow had nowhere else to live. It was **split** on that date, as a move operation rather than an authoring one, into `SPECIFICATION.md`, `CalculationSpecification.md`, `Architecture.md`, and `DESIGN.md`. It no longer exists; it survives in git history.

- **`SPECIFICATION.md`** is the source of truth for intent. The repository contains none of the system as code, so the documents lead the implementation, not the other way round.
- **`CalculationSpecification.md`** is complete for the first release, less a domain glossary it still owes.
- **`Architecture.md`** carries binding technology decisions and the Turf API facts. It still owes failure handling, observability, security, and schema.
- **`DESIGN.md`** carries the full interaction flow. It still owes the visual layer.
- **`Requirements/`** and **`DEPLOYMENT.md`** do not exist yet. `Requirements/` is the bottleneck for the project board described in `DELIVERY.md`.

Each document ends with what it still owes and the open questions it owns. Those are the shortest route to what is unfinished.

## Conventions

**Every model has exactly one home.** A formula, constant, or threshold is stated in `CalculationSpecification.md` and nowhere else; other documents reference it by section name. This is not tidiness — a second statement of a model is a second thing to keep correct, and the two will diverge.

**Cross-document references name the document.** An italic section name qualified with a filename points elsewhere; an unqualified one points within the same document.

**Numeric constants are proposals unless stated otherwise.** A proposed default exists so that specification work begins with a concrete number to argue against rather than a blank to fill; disagreeing with one is useful, leaving it undecided is not.

**Facts about the Turf API were verified against the live API**, not assumed. Several contradict the obvious guess. Where a document states an API behaviour, it was checked. Where something remains inferred, it says so.

**Open questions are catalogued, not silently decided.** Each document carries the open questions belonging to its own content, rather than one shared list.
