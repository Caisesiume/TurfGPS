# Runtime and deployment shape

The form the system is built and run in: what the build produces, how the client is served, the statefulness the solve session imposes, and the deployment topologies the architecture rules out. Distinct from `Platform support`, which is about the browsers the client must work in. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## NFR-003 — Build the service as one self-contained executable

```
Statement:    The build shall produce the TurfGPS service as a single
              executable that runs on a clean host of its target platform with
              no installed language runtime and no application server, per
              *D1* in `Architecture.md`.
Category:     Runtime and deployment shape
Source:       Architecture.md § D1
Priority:     MUST
Verification: test — the built artefact is copied alone into an image carrying
              no language runtime and no project dependencies, and starts and
              serves a request there
Acceptance:   files required to start the service on a clean target-platform
              host: exactly one executable, in an image carrying no language
              runtime, no application server and no project-installed
              dependencies
              service start-up and first served request in that image: both
              succeed, with no installation step performed between the copy
              and the start
Status:       to-build
Depends-on:   none
Risk:         A single binary is one of the three properties
              `Architecture.md § D1` chose Go for, and it is the one that lets
              the deployment model — written on 14 August 2026, and leaving the
              host open under
              `DEPLOYMENT.md § What this document does not decide` — avoid
              specifying a runtime environment at all. A build that needs an
              installed runtime or a tree of side files spends that property
              the model now rests on, and reclaiming it later means changing
              how the service is built rather than how it is deployed.
Rationale:    The obligation is the property the decision buys, not the
              decision itself: a record obliging the backend to be written in
              Go would restate a decision `Architecture.md` already states as
              binding, and could not fail any check a reviewer could run
              against an artefact — while the artefact count and the absence
              of an installed runtime can both be checked on the first build.
              It is deliberately silent about linkage: *D6* in
              `Architecture.md` proposes a raster sampler carrying a cgo
              dependency, which would leave the binary linked against system
              libraries, and that proposal is open — a record demanding a
              statically linked, cgo-free binary would settle it from the
              wrong side.
Resolved-by:  #19
```

## NFR-004 — Run the service as one long-running process

```
Statement:    The TurfGPS service shall run as a single long-running server
              process, admitting no deployment target that ends that process
              between requests, per *D1* in `Architecture.md`.
Category:     Runtime and deployment shape
Source:       Architecture.md § D1;
              Architecture.md § D2
Priority:     MUST
Verification: inspection — the service's entry point and its session registry,
              and the deployment configuration once one exists, are examined
              for a process whose lifetime spans many requests
Acceptance:   the service's entry point holds a server that listens for the
              lifetime of the process and a session registry created once at
              start-up rather than per request, and holds no per-invocation
              handler entry point, in the service's main package
              the deployment configuration names a target that keeps the
              service process running between requests, as the deployment
              model defines it in
              `DEPLOYMENT.md § Where the deployment configuration lives`
Status:       to-build
Depends-on:   none
Risk:         A per-invocation target discards the candidate set, the access
              classifications and the computed costs between requests, and
              rebuilding them is the expensive half of the pipeline. The loss
              is invisible to a functional test, which passes on rebuilt state
              exactly as it does on retained state, so it surfaces only as
              cost and latency once the shape is load-bearing.
              `Architecture.md § D2` records the same hazard arriving through
              the frontend, as steady pressure toward the topology
              `Architecture.md § D1` rules out.
Rationale:    The requirement here is the exclusion, not the retention. What
              must survive a solve is stated in `SPECIFICATION.md`, in a
              section this corpus has not swept, and restating it from an
              architecture section that only observes it would give one
              obligation two homes. This record takes only the consequence
              `Architecture.md § D1` draws in its own right — that the process
              outlives the request — so that the first skeleton is not written
              against a shape that has to be undone. The second criterion is
              deliberately unsatisfiable today and names why: the deployment
              model in
              `DEPLOYMENT.md § Where the deployment configuration lives` names
              the artefact it examines and does not create it.
Resolved-by:  #25
```

## NFR-005 — Serve the client as static files

```
Statement:    The client shall be built to files a static file host can serve
              with no application code executing on the server, per
              *D2* in `Architecture.md`.
Category:     Runtime and deployment shape
Source:       Architecture.md § D2
Priority:     MUST
Verification: test — the built client is served by a plain static file server
              with no application runtime present, and loads and completes a
              call to the service over HTTP
Acceptance:   server-side application code executed to serve the client: none,
              with the build output served by a static file host carrying no
              application runtime
              client load and first successful call to the service in that
              arrangement: both succeed, with the service reachable only over
              HTTP
Status:       to-build
Depends-on:   none
Risk:         Server-side rendering adds a build and a deployment layer
              serving no stated requirement, and the frameworks offering it
              default to serverless targets — the pressure
              `Architecture.md § D2` names toward the topology
              `Architecture.md § D1` rules out and NFR-004 forbids. Both are
              near-free to adopt while scaffolding and expensive to remove
              afterwards, because by then the client's data loading is written
              against them.
Rationale:    The obligation is the property, not the toolchain: a record
              obliging Vite and React would restate a binding decision and
              could not fail a check run against the artefact, while what the
              build emits and what serving it requires are both checkable on
              the first build. This is also what keeps the client deployable
              independently of the service, which the deployment model in
              `DEPLOYMENT.md § What is deployed` would otherwise have had to
              reconstruct.
Resolved-by:  #26
```
