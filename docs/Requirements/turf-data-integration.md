# Turf data integration

How zone, player and region data is obtained from the Turf API and what may be relied on from it: which source answers which question, the local synced copy and the staleness the pipeline tolerates in it, the shape constraints the endpoints place on a single request, and the fields whose absence or round scope constrains what can be concluded. Two neighbours abut it and neither is folded in: **rate over time against the API is `Outbound rate compliance`; per-journey call volume is `Call budget`** — this category owns what is fetched and what it means, never how often it may be asked for. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-019 — Address every Turf API request to the current API version

```
Statement:    The system shall issue every request it makes to the Turf API
              against the API version recorded as current under
              *API version* in `Architecture.md`.
Category:     Turf data integration
Source:       Architecture.md § API version
Priority:     MUST
Verification: inspection — the code paths that construct a Turf API request,
              and any configuration they draw a base path from, name between
              them exactly one API version, and it is the one recorded as
              current; every such path is examined, the fallback and error
              paths as well as the primary ones
Acceptance:   the code paths that construct a Turf API request, identified by
              what they construct rather than by what they are called,
              together with any configuration they draw a base path from, name
              exactly one API version between them, and it is the version
              recorded as current; the fallback and error paths are in that
              class and no path in it is exempt; a reader confirms this by
              enumerating the class from every point at which the system
              issues a Turf API request, reading each path's request
              construction and the configuration it draws on, and confirms in
              `Architecture.md § API version` which version is current
Status:       to-build
Depends-on:   none
Volatility:   settled — the version is cited rather than named, so a version
              bump moves the cited document and not this record, and the
              criteria are keyed to what a path constructs rather than to any
              component, so a decision about how the code is organized does
              not move them either; gained on the opportunistic transition,
              this record's first edit since the field existed
Risk:         The preceding version is deprecated, so a call path built on it
              works until the day it is withdrawn and then fails for every
              user at once, with no local change to point at. Worse than the
              outage is the silence before it: every fact the rest of these
              requirements rest on was verified against the current version,
              so a request answered by the older one voids those guarantees
              while still returning data that looks right.
Rationale:    Verification is inspection rather than test because the property
              must hold on every path that can construct a request, including
              the fallback and error paths a suite is least likely to
              exercise, and because a single version across all of them is
              what makes it checkable at all. The version is cited rather than
              named so that this record does not become a second home for it.
              Both criteria are keyed to what a path does — construct a Turf
              API request — rather than to any component it sits in, so the
              single-version property belongs to that class of paths and not
              to a named artefact's configuration, and the record presupposes
              no component, which is why `Depends-on` is none. A criterion
              keyed to an artefact is satisfied vacuously wherever that
              artefact is absent, reporting green while measuring nothing; the
              criteria this record now carries fail closed instead, each a
              cardinality claim over that class — its paths name exactly one
              version between them — which an empty class falsifies rather
              than satisfies, since a class naming no version has not named
              exactly one.
Resolved-by:  #18
```

## FR-022 — Refresh the local zone copy from a scheduled background job

```
Statement:    The system shall refresh its local synced copy of the zone set
              from the all-zones endpoint recorded under
              *Retrieving zones* in `Architecture.md` on a schedule, by a
              background job that no user request initiates or waits on.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — the refresh runs from its schedule with no request
              outstanding, and a planning request handled while the copy is
              due for refresh issues no all-zones request and does not block
              on the job
Acceptance:   given a local synced copy due for refresh and no user request in
              flight, when the schedule fires, then the refresh runs and
              updates the copy
              given a planning request handled while the local synced copy is
              due for refresh, when the request is served, then it issues no
              all-zones request and does not wait on the refresh job
Status:       to-build
Depends-on:   none
Risk:         The all-zones endpoint admits one request per interval, so a
              request path that triggers it can spend the whole interval's
              allowance on a single page load. The copy is then frozen until
              the interval elapses, so the damage outlives the request that
              caused it. Freshness is bounded by that limit however fast the
              transfer is, so a fast download is no reason to move the sync
              onto the request path.
Rationale:    This is one obligation rather than two: a job that runs on a
              schedule and can also be triggered by a request is not off the
              request path, and the source states the schedule and the
              prohibition in one breath. The record fixes neither the interval
              nor the mechanism — the interval the endpoint permits is cited,
              and how the job is scheduled is left to the design.
Resolved-by:  #20
```

## FR-023 — Resolve a route corridor's zones against the local copy

```
Statement:    The system shall resolve the zones lying within a route corridor
              against its local synced copy of the zone set, without issuing a
              Turf API request for that resolution.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — resolving a corridor against a local copy within the
              tolerated staleness bound returns the zones lying within it and
              issues no Turf API request
Acceptance:   given a local synced copy within the staleness bound recorded
              under `Architecture.md § Retrieving zones` and a route corridor,
              when the zones lying within that corridor are resolved, then
              they are returned from the local copy and no Turf API request is
              issued
Status:       to-build
Depends-on:   FR-022
Risk:         Resolving corridors through the fallback endpoint puts a tiled
              sequence of Turf requests on the planning path, so planning time
              and call volume both grow with corridor length against a
              per-second limit, and a Turf outage stops candidate discovery
              outright rather than degrading only what the local copy cannot
              answer.
Rationale:    The record fixes where corridor resolution is answered from, not
              which endpoints the system may call: the same section keeps the
              bounding-box endpoint as the fallback and as the means of
              refreshing volatile per-zone fields, and the criterion is
              conditioned on a current local copy so that it forbids neither
              use. What the local copy may be asked is bounded by FR-026.
Resolved-by:  #27
```

## FR-024 — Plan against a mid-refresh or stale local copy

```
Statement:    The system shall complete journey planning from its local synced
              copy of the zone set while a refresh of that copy is in
              progress, and while the copy is stale by up to the bound
              recorded under *Retrieving zones* in `Architecture.md`.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — planning completes and returns recommendations both while
              a refresh job is running and against a copy as old as the
              tolerated staleness bound
Acceptance:   given a refresh of the local synced copy in progress, when a
              journey is planned, then planning completes and returns
              recommendations
              given a local synced copy whose last completed refresh is as old
              as the staleness bound recorded under
              `Architecture.md § Retrieving zones`, when a journey is planned,
              then planning completes and returns recommendations
Status:       to-build
Depends-on:   FR-022; FR-023
Risk:         The refresh runs on a schedule no user sees, so a pipeline that
              refuses to plan while it is running, or against a copy within
              the tolerated age, fails at moments nobody outside can predict
              and nobody inside can reproduce on demand. It presents as an
              outage rather than as the design choice it is, and the refresh
              interval makes the window neither rare nor brief.
Rationale:    One obligation under two conditions rather than two obligations:
              the outcome asserted is identical in both criteria, and
              splitting them would leave each half readable as permission to
              fail the other. What the system may conclude from a copy of that
              age is a confidence question, and it is NFR-002's.
Resolved-by:  #33
```

## FR-025 — Build bounding-box requests against the permitted area product

```
Statement:    The system shall construct every bounding-box request to the
              Turf zones endpoint so that the product of its latitude span and
              its longitude span satisfies the area constraint recorded under
              *Retrieving zones* in `Architecture.md`.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     COULD
Verification: test — an area whose span product exceeds the constraint is
              divided into requests that each satisfy it, and an area
              satisfying the constraint is issued whole even where one of its
              spans alone exceeds the constraint's value
Acceptance:   given an area whose latitude span multiplied by its longitude
              span exceeds the constraint recorded under
              `Architecture.md § Retrieving zones`, when requests are
              constructed for it, then every request issued satisfies that
              constraint
              given an area whose span product satisfies that constraint but
              one of whose spans, taken alone, exceeds the constraint's value,
              when requests are constructed for it, then the area is issued as
              a single request
Status:       to-build
Depends-on:   none
Rationale:    The second criterion is the record's point and not a refinement
              of the first. A per-axis reading of the constraint admits no
              illegal request, so it passes any test written from the first
              criterion alone; what it does instead is refuse legal ones,
              splitting a long narrow corridor whose product sits well inside
              the limit into many requests, each then spent against the
              general per-second limit. The defect is invisible as a
              correctness fault and surfaces only as call volume the design
              never planned for, which is why the shape of the constraint — a
              product, and not a bound on either span — is stated in the
              record rather than left to the citation. The relation is
              expressed; the value stays in the cited section.
Resolved-by:  #21
```

## FR-026 — Answer no ownership question from the local zone copy

```
Statement:    The system shall derive no statement of zone ownership from the
              local synced copy of the zone set described under
              *Retrieving zones* in `Architecture.md`.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: inspection — the local synced copy's schema carries no ownership
              field, and the ingest mapping for the all-zones response
              populates none
Acceptance:   the schema of the local synced copy carries no owner and no
              last-taken column, and the ingest mapping for the all-zones
              response populates neither; a reader confirms this in the zone
              schema definition and in the ingest mapping, and confirms in
              `Architecture.md § Retrieving zones` that the endpoint omits
              both fields
Status:       to-build
Depends-on:   FR-022
Risk:         The all-zones response omits `currentOwner` and `dateLastTaken`
              from every record, so a copy carrying those columns holds a
              uniform absence that is indistinguishable, in the store, from
              the meaningful absence a bounding-box response carries — where a
              missing `currentOwner` identifies a zone nobody holds this
              round. A consumer reading ownership from the synced copy
              therefore concludes that every zone in the game is unheld, and
              the conclusion is uniform, confident and wrong.
Rationale:    The obligation binds the store and its consumers rather than any
              one query, because the defect is a column that is always empty
              being read as a fact. The same field means opposite things in
              the two responses, and the source records those two facts in
              separate subsections without joining them, so the join is made
              here. Where ownership is needed the first release takes it from
              the player-data call, per
              *The user's held zones are already known* in `Architecture.md`;
              the same section records the bounding-box endpoint as able to
              answer per-zone ownership, and records that the first release
              does not require it. The copy this record binds is the one the
              all-zones sync writes; a volatile overlay held separately, from
              a response that does carry these fields, is governed by FR-030
              instead.
Resolved-by:  #29
```

## FR-028 — Decide the user's own holdings by membership in the held-zone list

```
Statement:    The system shall determine whether the user holds a zone shown
              during review by testing that zone's id against the held-zone
              list returned by the player-data call made at the start of the
              review session, issuing no further Turf API request.
Category:     Turf data integration
Source:       Architecture.md § The user's held zones are already known
Priority:     MUST
Verification: test — a review session over a route containing both held and
              unheld zones decides every zone from the one held-zone list and
              issues no further Turf request
Acceptance:   given a review session whose player-data call has returned the
              user's held-zone list, when a route containing both held and
              unheld zones is reviewed, then each zone's held state is decided
              by membership in that list and no further Turf request is issued
              given a review session in which the user reviews further zones
              after the first, when their held state is decided, then the
              number of Turf requests the session has made is unchanged
Status:       to-build
Depends-on:   none
Risk:         A per-zone ownership lookup turns one call into one call per
              zone reviewed, against a per-second rate limit, so the review
              loop slows in proportion to route length and the journey's call
              budget is spent answering a question already answered in hand —
              spent, moreover, during the one interaction where the user is
              waiting.
Rationale:    The section records the equivalence as verified against sampled
              ids rather than assumed, which is what lets a membership test
              stand in for a lookup. The record obliges neither a cache nor a
              refresh policy: the same section states that no cache is
              required, and how the age of that single fetch is surfaced is
              owned by *Zones the user already owns* in `DESIGN.md`. What the
              list deliberately does not answer — whether any other player
              holds a zone — is a scope limit of the same section and is not
              authored here.
Resolved-by:  #23
```

## FR-029 — Determine region lordship once, from a single region response

```
Statement:    The system shall determine the user's region-lord status as a
              single value covering all regions, obtained by scanning one
              region response for the user's id.
Category:     Turf data integration
Source:       Architecture.md § Region lords
Priority:     MUST
Verification: test — a user holding a lordship in exactly one region, planning
              a journey whose zones lie entirely in other regions, resolves as
              holding one, from a single region request
Acceptance:   given a user who holds a region lordship in exactly one region
              and a journey whose zones lie entirely in other regions, when
              region-lord status is determined, then it resolves as held and
              exactly one region request is issued
              given a user who holds no region lordship, when region-lord
              status is determined, then it resolves as not held and exactly
              one region request is issued
Status:       to-build
Depends-on:   none
Risk:         The Region Lord takeover bonus applies globally rather than per
              region, so a status determined per zone or per region is wrong
              for every stop outside the region the user leads, and buys a
              join between zones and regions the data does not require. The
              error under-applies the bonus, which lengthens estimates and is
              therefore conservative — so it passes every sanity check a
              reviewer would think to run.
Rationale:    The section names exactly one use for region data, whether the
              user holds a lordship anywhere, and the first criterion is what
              separates a correct implementation from a per-region one: it
              puts the lordship and the journey in different regions. How
              often the status is re-resolved is stated under
              *Bonuses that reduce takeover time* in `CalculationSpecification.md`
              rather than in the cited section, and is deliberately not
              authored here.
Resolved-by:  #24
```

## FR-030 — Carry an absent ownership field as absent

```
Statement:    The system shall process a zone record in which `currentOwner`
              or `dateLastTaken` is absent without substituting a value for
              the absent field.
Category:     Turf data integration
Source:       Architecture.md § Volatile and optional fields
Priority:     MUST
Verification: test — a zone response omitting either field, or both, is
              processed without error and the field's absence is carried
              forward rather than defaulted
Acceptance:   given a zone response from which `currentOwner` is absent, when
              that zone is processed, then processing completes without error
              and the field is carried as absent rather than as an owner value
              given a zone response from which `dateLastTaken` is absent, when
              that zone is processed, then processing completes without error
              and the field is carried as absent rather than as a date
Status:       to-build
Depends-on:   none
Risk:         Both fields are absent by design for a zone nobody has taken
              this round, so absence is the ordinary case rather than an
              exceptional one and a consumer requiring them fails on routine
              data. A consumer that defaults them instead destroys the only
              signal the absence carries, and destroys it silently: every zone
              then reports an owner or a date the API never sent.
Rationale:    Distinct from FR-026, which forbids the synced copy from
              carrying these fields at all: this record governs a zone record
              that does carry them, from the endpoint that returns them, and
              obliges the absence to survive into whatever consumes it.
              Without that, FR-031 has nothing left to reason over.
Resolved-by:  #22
```

## FR-031 — Do not read an absent ownership field as a zone never taken

```
Statement:    The system shall not conclude from the absence of `currentOwner`
              or `dateLastTaken` on a zone that the zone has never been taken.
Category:     Turf data integration
Source:       Architecture.md § Volatile and optional fields
Priority:     MUST
Verification: test — a zone reporting a non-zero `totalTakeovers` and neither
              field present is not treated by any derived signal as a zone
              with no takeover history
Acceptance:   given a zone reporting a non-zero `totalTakeovers` and no
              `dateLastTaken`, when a signal derived from that zone's takeover
              history is computed, then the zone is not treated as one that
              has never been taken
              given a zone reporting a non-zero `totalTakeovers` and no
              `currentOwner`, when a signal derived from that zone's takeover
              history is computed, then the zone is not treated as one that
              has never been taken
Status:       to-build
Depends-on:   FR-030
Risk:         The section records zones carrying substantial lifetime takeover
              counts and no date at all, so the absence is ordinary data
              rather than a curiosity. Read as never taken, such a zone is
              scored as untouched terrain, and the misreading is systematic
              rather than random: it lands on exactly the zones nobody has
              held since the round began.
Rationale:    The obligation is defensive and holds whichever way the
              round-scope question is finally settled. The section records
              `currentOwner` resetting at a round boundary as confirmed and
              the round scope of `dateLastTaken` as strongly indicated by
              sampling rather than established, and the prohibition stands
              under either reading — which is why it is stated as what may not
              be concluded rather than as what the fields mean. Nothing here
              obliges the system to derive the round's start date from the
              data; the section calls that useful in its own right and does
              not require it.
Resolved-by:  #31
```

## FR-033 — Plan no journey for lack of zone data only where no copy is held

```
Statement:    The system shall treat the absence of a local synced copy of the
              zone set, and not the age of an existing copy beyond the bound
              recorded under *Retrieving zones* in `Architecture.md`, as the
              ground on which no journey is planned for lack of zone data.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — planning completes and returns recommendations against a
              local copy older than the tolerated staleness bound, and no
              journey is planned where no local copy exists at all
Acceptance:   given a local synced copy of the zone set whose last completed
              refresh is older than the bound recorded under
              `Architecture.md § Retrieving zones`, when a journey is planned,
              then planning completes and returns recommendations
              given no local synced copy of the zone set, when a journey is
              planned, then the system plans no journey
Status:       to-build
Depends-on:   FR-022; FR-023; FR-024
Risk:         A refusal beyond the bound puts the whole product behind a
              background job no user can see, retry or reach: one failed or
              delayed sync stops planning for everyone at once, and the outage
              lasts as long as the job stays down rather than as long as a
              request. What the refusal withholds is the stable,
              non-round-scoped zone data the sync carries — which zones exist
              and what they are — so the product is removed to protect an
              answer that had barely moved.
Rationale:    The cited section obliges the pipeline to tolerate a copy stale
              up to the bound and stops there, saying nothing about a copy
              older than that; this record closes that silence on the Owner's
              ruling of 3 August 2026 — plan against whatever copy exists, and
              refuse only where none exists — so the unwritten case does not
              settle itself as a refusal on first contact with it. FR-024 owns
              the tolerated range and the mid-refresh window; this record
              begins past the bound and re-asserts neither, which is why the
              condition is stated as age beyond the cited bound rather than as
              staleness generally. The two branches are one gate and not two
              records: the decision reads one input, whether a copy is held,
              and separated the halves read as permission to fail each other —
              a record obliging planning from any copy invites planning from
              none, and a record refusing where none is held is satisfied by
              refusing always. What a plan built this way may claim about
              itself is not settled here: the ordering of confidence against
              data age is NFR-002's and is neither restated nor extended, and
              the fact the result must carry is FR-034, separated because a
              plan produced and a plan marked are separately testable
              outcomes. Nothing here fixes how a refusal is surfaced, which
              has no source yet.
Resolved-by:  #33
```

## FR-034 — Record a result as built from data stale beyond the bound

```
Statement:    The system shall record, on each result built from a local
              synced copy of the zone set stale beyond the bound recorded
              under *Retrieving zones* in `Architecture.md`, that the result
              was built from data stale beyond that bound.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — a result planned from a local copy older than the
              tolerated staleness bound carries that fact, and a result
              planned from a copy within the bound does not
Acceptance:   given a local synced copy of the zone set whose last completed
              refresh is older than the bound recorded under
              `Architecture.md § Retrieving zones`, when a journey is planned,
              then the result returned carries the fact that it was built from
              data stale beyond that bound
              given a local synced copy of the zone set whose last completed
              refresh is within that bound, when a journey is planned, then
              the result returned carries no such fact
Status:       to-build
Depends-on:   FR-022; FR-033
Risk:         A plan built from data past the tolerated bound and returned
              indistinguishable from one built from a current copy makes the
              degradation the ruling accepted invisible: the product keeps
              answering while the basis of its answers quietly erodes, and
              nothing downstream — the confidence recorded for it, a later
              diagnosis, the user deciding whether to trust it — can tell
              which answers to discount. The loss is silent, and it is total
              exactly when it matters most: a prolonged sync failure reaches
              every result at once and none of them says so.
Rationale:    Separated from FR-033 on singularity: producing the plan and
              recording what it was built from are two outcomes on one branch,
              and a statement joining them would be half-satisfiable by the
              half that keeps the product working — which is the half nobody
              forgets to build. It carries the second half of the Owner's
              ruling of 3 August 2026, that the result is marked as built from
              stale data. The second criterion is as much the record's point
              as the first: a fact recorded on every result records nothing,
              and the marking informs only where it tracks the condition that
              produced it; that criterion asserts nothing about whether
              planning within the bound completes, which is FR-024's. The
              boundary against NFR-002 is stated because the two read as one
              record until they are separated. NFR-002 owns the relation
              between data age and the confidence a recommendation carries and
              remains its only home — no scale, no ordering and no confidence
              value is set or implied here. This record carries a fact about
              provenance, true or false of a given result, and a system could
              satisfy either without the other: record the fact and never let
              it move a confidence value, or move confidence and never record
              what moved it. The record fixes no mechanism — where the fact is
              held, what it is called, and whether or how it reaches a user
              are design and interface questions, and the interface has no
              source in the corpus yet. The obligation is that the result
              carries the fact.
Resolved-by:  #35
```

## FR-035 — Update dateCreated on zones the local copy already holds

```
Statement:    The system shall store the `dateCreated` a refresh returns for a
              zone in its local synced copy of the zone set, whether or not
              the copy already holds that zone.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — a refresh returning a revised `dateCreated` for an id the
              copy already holds leaves the stored value equal to the returned
              one, and an id new to the copy is stored with the value the
              refresh returned
Acceptance:   given a zone id the local synced copy already holds and a
              refresh returning for that id a `dateCreated` different from the
              stored one, when the refresh completes, then the stored
              `dateCreated` for that id is the value the refresh returned
              given a zone id the local synced copy does not hold, when a
              refresh returning that id completes, then the copy holds that id
              with the `dateCreated` the refresh returned
Status:       to-build
Depends-on:   FR-022
Risk:         A row inserted once and never updated keeps a `dateCreated` the
              API has since revised, and anything derived from the months
              since that date is then computed against the wrong denominator —
              `CalculationSpecification.md § Takeover rate` divides lifetime
              takeovers by exactly that interval, so a stale date inflates the
              rate rather than merely blurring it, and
              `Architecture.md § Retrieving zones` measures how far the rate
              moved for the one zone it observed being re-sited. Nothing
              surfaces the error: the zone carries a plausible date, a
              plausible count and a plausible rate, and no signal separates it
              from a zone whose date never moved. Rarity is what makes the
              defect dangerous rather than negligible — that section found the
              revision exceptional across the whole interval it compared, so
              no fixture, no sample and no reviewer meets the case by
              accident, and a write path that gets it wrong reads as correct
              in review for as long as the product runs.
Rationale:    The record binds the write path of the refresh and not its
              trigger, which is FR-022's: a job running on schedule that
              writes only ids it has not seen before satisfies FR-022 while
              leaving every held zone's date frozen at first sight. The
              obligation is stated as the value stored rather than as an
              upsert keyed on `id`, because the cited section offers that as
              one sufficient way to meet the need rather than as the need — a
              write path meeting it another way is compliant, and the same
              section states that no detection mechanism is required, so none
              is obliged here either. The two criteria are one obligation
              under two conditions: the outcome asserted is identical in both,
              that the stored value is the value the refresh returned, and
              separating them would leave the held-id half readable as an
              exception. What the section observes moving alongside the date
              on that zone — its coordinates and its takeover count — it
              observes rather than obliges, so no field beyond the one it
              names is bound here.
Resolved-by:  #38
```

## FR-036 — Treat an absent zone type as ordinary data

```
Statement:    The system shall treat a zone record carrying no `type` as
              ordinary data wherever that record is ingested, stored or
              consumed, and not as an anomaly or as a defect in the sync.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     MUST
Verification: test — a zone set in which some records carry no `type` is
              ingested with every record retained, and no error, rejection or
              data-defect signal is raised on account of the absent field
Acceptance:   given a zone set in which some records carry a `type` and others
              carry none, when the set is ingested, then every record in it is
              retained and the ingest completes as it does for a set in which
              every record carries one
              given a zone record carrying no `type`, when that record is
              ingested, stored or consumed, then no error, rejection or
              data-defect signal is raised on account of the absence
Status:       to-build
Depends-on:   none
Risk:         `Architecture.md § Retrieving zones` measures the field on a
              minority of zones, so a schema, a validator or an ingest guard
              that requires it rejects most of the corpus rather than a fringe
              of it, and candidate discovery then runs against a fraction of
              the game while every later stage reasons correctly about the
              wrong world. The escalating half is quieter and worse: a
              data-defect signal raised on the ordinary case fires on nearly
              every record, so whoever reads it stops reading it, and the sync
              failure it was built to catch then arrives in a channel already
              written off as noise.
Rationale:    One record and not two, and the pair it is measured against is
              FR-030 with FR-031. Those two split because a correctly carried
              absence can still be misread downstream — the absence survives
              ingest and a consumer concludes the zone was never taken — so
              carrying and concluding are independent, and a system can
              satisfy either while failing the other. Here the halves are not
              independent: rejecting the record and raising an alert on it are
              two forms of one conclusion, that the absence is a fault, and
              the cited section states them in one breath as one prohibition,
              leaving no second inference for a record of its own. What is
              genuinely separate is substitution — an ingest supplying a value
              for the absent field escalates nothing and is not reached by
              this record; the duty that the absence survive unsubstituted
              into consuming code is stated by
              `SPECIFICATION.md § Zone attributes and user-defined value`,
              which is unswept, and belongs to the batch scoped there in
              FR-030's shape. The record states the need rather than the
              storage choice the section's evidence illustrates it with: what
              the local copy's schema may declare is `Architecture.md § D4`'s,
              and `Architecture.md § The zone table` now declares that column
              and declares it nullable, so a record naming a column would give
              a decision already taken a second home rather than pre-empt one
              not yet taken. `Architecture.md § The schema` marks that DDL a
              proposal, which sharpens the point rather than blunting it: a
              record naming the column would harden a proposal into an
              obligation. Verification is test rather than inspection for the
              reason FR-032 gives: the assumption can hide in a schema
              constraint, a validator, an ingest guard, a metric or an alert
              rule, each of which reads as ordinary diligence in review, while
              one fixture set carrying records without the field fails every
              form of it at once.
Resolved-by:  #39
```

## FR-037 — Determine a zone's country where the region subkey is absent

```
Statement:    The system shall obtain a zone's country, wherever one is
              required, from `region.country` where that subkey is present and
              from `region.name` where it is absent.
Category:     Turf data integration
Source:       Architecture.md § Retrieving zones
Priority:     COULD
Verification: test — a zone whose `region` carries no `country` subkey yields
              a country, and a set holding such zones alongside zones that
              carry the subkey places every one of them in a group when
              grouped by country
Acceptance:   given a zone whose `region` carries no `country` subkey, when
              that zone's country is required, then it is obtained from
              `region.name` and the zone yields a country
              given a zone set holding zones that carry the `country` subkey
              and zones that do not, when the set is grouped by country, then
              every zone in the set falls into a group and none is dropped
Status:       to-build
Depends-on:   none
Rationale:    The record exists against a named failure mode, and the mode is
              silence: a grouping taken on the subkey drops the zones lacking
              it rather than failing, so the aggregate looks complete and is
              not, and nothing in the result separates a country with few
              zones from a country whose zones were dropped whole.
              `Architecture.md § Retrieving zones` records the zones lacking
              the subkey as concentrated under a handful of country names
              rather than scattered across the set, so the loss is not a thin
              uniform sampling error but the removal of whole countries from
              any per-country view — which is exactly the shape a reader takes
              for a fact about the game rather than a defect in the query. Two
              readings of the cited paragraph were considered and this is the
              one it supports. The section supplies the resolution itself:
              where the subkey is missing the country's name is carried in
              `region.name`, so a country can be obtained for every zone and
              the aggregate made genuinely complete. The alternative, that an
              aggregate short of zones must fail visibly, would oblige a
              failure behaviour and a destination for it that no document
              states, and it would leave the section's own resolving fact with
              nothing to do. That is also why it is not a second record: the
              two are not separately testable outcomes of one source, they are
              one outcome and one construction the source does not make. The
              statement names both branches of the determination rather than
              the absent one alone because they are two branches of one gate
              reading one input, whether the subkey is present, exactly as
              FR-033's two are; a record stating only the fallback would read
              as licence to take the subkey wherever it happens to be there.
              The country obtained here serves aggregates over the fetched
              zone set and nothing else, and is in particular never an input
              to speed-limit resolution:
              `SPECIFICATION.md § An unknown speed limit fails the check`
              forbids a resolved limit resting on a country attribution
              supplied by the resolver rather than by the data. The statement
              is `shall` while nothing in the first release requires a zone's
              country: the obligation binds the moment anything does, and
              `COULD` records that no such consumer is built yet rather than
              any weakness in the duty. `WON'T-now` would file no story and
              leave a later aggregate free to be silently short.
Resolved-by:  #40
```
