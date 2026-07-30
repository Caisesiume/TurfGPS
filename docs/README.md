# TurfGPS documentation

This folder is the front door to the TurfGPS documentation set. Each document answers one question, and the question is what decides where a piece of writing belongs.

## The documents

| Document | The question it answers |
|----------|------------------------|
| `SPECIFICATION.md` | What is the system, and how should it behave conceptually? |
| `Requirements/` | What precisely must the system satisfy? |
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

### Architecture.md

Component boundaries, runtime topology, scaling model, ports and adapters, extensibility, persistence model, data flows, security architecture, failure handling, and architectural decisions.

### DESIGN.md

User interaction design, user experience, graphic profile, the feelings the design should express, simplicity, user-flow diagrams, design visions, use-case detail, and page layouts.

### DEPLOYMENT.md

Operational detail: how to deploy, target OS, hosting options with price and complexity comparisons, pipeline and CI detail, full deployment architecture with diagrams, and operational direction.

## Current state

The documentation set is **not yet in this shape**. As of 31 July 2026:

- **`Concept.md`** holds the authoritative product definition at roughly 1,670 lines. It is the source of truth for intent — the repository contains almost none of the system as code, so the document leads the implementation, not the other way round.
- **`Architecture.md`** and **`CalculationSpecification.md`** are outline stubs.

`Concept.md` grew past its purpose because formulas, thresholds, caching policy, and interaction flow had nowhere else to live. Splitting it into the structure above is the next planned step.

Two notes for whoever performs that split:

**It is a move operation, not an authoring one.** `Concept.md` explicitly forbids restating its own formulas — several sections say a model is stated *only* there. Content must therefore be relocated rather than copied, or the anti-duplication rules it relies on are broken.

**The calculation specification is already written**, inside `Concept.md`, in finished form with numeric defaults. The stub understates the work remaining by a wide margin in one direction and overstates it in another: there is a lot of content, but it needs moving rather than writing.

## Conventions

**Numeric constants are proposals unless stated otherwise.** `Concept.md` marks them explicitly. A proposed default exists so that specification work begins with a concrete number to argue against rather than a blank to fill; disagreeing with one is useful, leaving it undecided is not.

**Facts about the Turf API were verified against the live API**, not assumed. Several contradict the obvious guess. Where a document states an API behaviour, it was checked. Where something remains inferred, it says so.

**Open questions are catalogued, not silently decided.** `Concept.md` carries an *Open questions* section separating what is proposed, what needs measurement, what needs evidence from real use, and what is deferred. When the split happens, each document should carry the open questions belonging to its own content.
