# Persistence and staleness

Storing a plan, retrieving it, and what makes a stored plan no longer trustworthy. This scope is the category's entry on the register in `README.md`, which is its home.

Index: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## FR-094 — Keep a confirmed plan reopenable after an arbitrary interval

```
Statement:    The system shall keep each plan the user confirms as a stored
              plan that can be reopened on the same device after an arbitrary
              interval.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Route persistence
Priority:     MUST
Verification: test — a plan confirmed in one session reopens on the same
              device in a later session, after the session that made it has
              ended and after an arbitrary interval
Acceptance:   given a plan the user has confirmed, when the browser is closed
              and opened again on the same device after an arbitrary interval,
              then that stored plan is available to reopen
              given a plan the user has confirmed and a session that has since
              ended, when a new session begins on the same device, then that
              stored plan is available to reopen with no state carried over
              from the ended session
Status:       to-build
Depends-on:   none
Volatility:   settled — `SPECIFICATION.md § Route persistence` states the
              obligation directly, and `Architecture.md § D4` decides the
              store that holds it
Risk:         A user who plans a road trip weeks ahead, closes the browser and
              finds the work gone does not redo the planning — they abandon
              the product. The pipeline behind a plan is the most expensive
              computation this product performs, and losing it costs the user
              the one thing they cannot cheaply repeat.
Rationale:    The record obliges availability and not a mechanism, because
              `Architecture.md § D4` decides where a stored plan lives and
              this corpus does not become a second home for that decision.
              Nothing filed obliges the confirmation step that produces the
              object stored here —
              `SPECIFICATION.md § Route review and zone confirmation` belongs
              to `Review and replacement`, which is reserved — so the
              confirmed plan is named as the input it is and `Depends-on`
              reads `none` until that category opens. What a stored plan
              retains is FR-096's; how long it is kept is the non-functional
              lane's and is deliberately absent here.
Resolved-by:  —
```

## FR-095 — Add a new stored plan without displacing an existing one

```
Statement:    The system shall add each newly confirmed plan to the stored
              plans a device already holds, displacing none of them.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Route persistence
Priority:     MUST
Verification: test — a device already holding stored plans takes a further
              confirmed plan and afterwards holds that plan and every earlier
              one
Acceptance:   given a device holding one or more stored plans, when a further
              plan is confirmed, then that plan and every stored plan held
              before it are all available to reopen
              given a device holding stored plans, when a further plan is
              confirmed, then no stored plan is removed to make room for it
Status:       to-build
Depends-on:   FR-094
Volatility:   settled — `SPECIFICATION.md § Route persistence` states that
              several plans are held at once and that starting a new journey
              never overwrites an existing one
Risk:         Overwriting is silent and irreversible: the user plans next
              weekend's route, plans a summer trip, and finds on the Friday
              that the first is gone, with nothing having warned them.
              Eviction under pressure is the same loss deferred until the
              device holds enough plans for it.
Rationale:    The second criterion is the record's point rather than a
              restatement of the first. Overwriting on confirmation is the
              failure the source names; discarding an older stored plan to
              make room is the same failure arriving later, and a store that
              satisfies only the first criterion passes every test written on
              a device holding two plans.
Resolved-by:  —
```

## FR-096 — Retain with a stored plan the state a later re-solve needs

```
Statement:    The system shall retain, with each stored plan, the candidate
              set, the access classifications, the computed costs, and the
              additional-time allowance the journey was planned against,
              alongside the confirmed zones and their order.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Route persistence
Priority:     MUST
Verification: test — a reopened stored plan yields a replacement zone from the
              state retained with it, without the pipeline being rerun from
              the beginning
Acceptance:   given a stored plan reopened after an arbitrary interval, when
              the state retained with it is examined, then it holds the
              candidate set, the access classifications, the computed costs,
              and the additional-time allowance the journey was planned
              against
              given a stored plan reopened after an arbitrary interval, when a
              zone in it is rejected during a later review, then the
              replacement is drawn from the retained candidate set without the
              pipeline being rerun from the beginning
Status:       to-build
Depends-on:   FR-094; FR-038
Volatility:   settled — `SPECIFICATION.md § Route persistence` enumerates what
              a stored plan keeps and records that the larger option was
              chosen deliberately
Risk:         A stored plan holding only the confirmed zones cannot be
              re-solved, so every zone swapped during a later review costs a
              full pipeline run — the cost the source says persistence exists
              to spare. The user meets it as a wait long enough that the
              review loop stops being usable, on the morning of departure.
Rationale:    The enumeration is bounded from two directions and neither bound
              is authored here. Round-scoped ownership is excluded by FR-098,
              which refines the source's *everything* rather than
              contradicting it. The Turf username is kept out of the stored
              object under `Architecture.md § Personal data`, and what a
              stored plan may hold and for how long is the non-functional
              lane's. The allowance is on the list for a reason this record
              does not restate: it is what a later review must judge a
              replacement against, which is FR-097's obligation and not this
              one's.
Resolved-by:  —
```

## FR-097 — Judge a later review against the allowance retained with the plan

```
Statement:    The system shall judge a zone swapped into a reopened stored
              plan during a later review against the additional-time allowance
              retained with that stored plan, and against no other limit.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Route persistence
Priority:     MUST
Verification: test — a replacement in a reopened stored plan is admitted or
              refused on the retained allowance, including where a different
              additional-time limit has been entered since for another journey
Acceptance:   given a reopened stored plan whose retained additional-time
              allowance differs from any limit entered since, when a zone in
              it is replaced, then the replacement is judged against the
              retained allowance
              given a further journey planned since against a different
              additional-time limit, when a zone in the reopened stored plan
              is replaced, then that later limit has no effect on the
              judgement
Status:       to-build
Depends-on:   FR-094; FR-096; FR-038
Volatility:   settled — `SPECIFICATION.md § Route persistence` states that the
              allowance is a property of the journey and travels with the
              stored plan for the same reason the costs do
Risk:         A stored plan re-solved against any other limit is re-solved
              against a constraint the user never agreed to, and nothing about
              the result looks wrong: the route comes back longer or shorter
              than the one they approved, with no figure on screen that is
              inconsistent with any other. It is the staleness failure that
              leaves no trace.
Rationale:    Separated from FR-096 on singularity: retaining the allowance
              and judging against it are outcomes a system can satisfy
              independently, and a store that keeps the allowance and then
              re-solves against a current default satisfies FR-096 exactly.
              The obligation arises from persistence, which is why it is
              drafted in this category; whether it files here or with the
              review loop once `Review and replacement` opens is
              `@requirements-engineer`'s.
Resolved-by:  —
```

## FR-098 — Hold no zone ownership in a stored plan

```
Statement:    The system shall hold no zone-ownership state in a stored plan,
              so that no ownership presented for a reopened stored plan is
              read from the plan.
Category:     Persistence and staleness
Source:       Architecture.md § Round rollover
Priority:     MUST
Verification: test — a stored plan written while the user held zones in it
              carries no ownership state, and the ownership presented when it
              is reopened in a later round is the one that session's held-zone
              list gives
Acceptance:   given a plan confirmed while the user holds zones in it, when
              the plan is stored, then the stored plan holds no ownership
              state for any zone
              given a stored plan stored in one round and reopened in a later
              one, when an ownership indicator is presented for a zone in it,
              then the value presented is the one that session's held-zone
              list gives
Status:       to-build
Depends-on:   FR-028
Volatility:   settled — `Architecture.md § Round rollover` states that the
              plan payload stores no ownership and that the indicator is
              recomputed on every open
Risk:         Ownership is round-scoped and a stored plan lives on the order
              of a dozen rounds, so ownership recorded in one is wrong for
              most of the plan's life. A stale *you already own this* is worse
              than showing nothing: it makes the user skip a zone they could
              have taken, confidently, on the strength of the product's own
              display.
Rationale:    This is the one exclusion from what
              `SPECIFICATION.md § Route persistence` calls everything, and it
              is a refinement rather than a conflict — that section lists the
              state a re-solve needs, and ownership is not state a re-solve
              needs. The recomputation is FR-028's, which decides held state
              from the player-data call made at the start of a review session;
              what this record adds is that the stored plan is not a competing
              source for the same fact, which is the rule FR-026 applies to
              the local synced zone copy applied to the other store. No stale
              marker survives a rollover to be cleared, because none is
              stored.
Resolved-by:  —
```

## FR-099 — Record with each stored plan when it was computed

```
Statement:    The system shall record, with each stored plan, the time at
              which the plan was computed.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Stored routes go stale
Priority:     MUST
Verification: test — a stored plan carries its computation time when reopened,
              and reopening it does not advance that time
Acceptance:   given a plan that has been computed and confirmed, when it is
              stored and later reopened, then the time at which the plan was
              computed is available from the stored plan
              given a stored plan that has been reopened, when the recorded
              computation time is read again, then it is unchanged by the
              reopening
Status:       to-build
Depends-on:   FR-094
Volatility:   settled — `SPECIFICATION.md § Stored routes go stale` requires a
              stored route to record when it was computed
Risk:         Without it the gap between planning and departure is
              unmeasurable, and every judgement resting on that gap fails
              silently: nothing can decide whether the volatile data is stale,
              whether a round boundary has passed since, or what confidence a
              reopened plan may carry.
Rationale:    The recorded time is the plan's computation and not its last
              reopening, which is what the second criterion fixes;
              `Architecture.md § The plan table` keeps the two as separate
              columns and this record obliges the distinction without naming
              either. Whether the staleness provenance FR-034 records for a
              result survives storage and retrieval is answered by FR-116 and
              no longer outstanding; this record carries the computation time
              and not that marking.
Resolved-by:  —
```

## FR-100 — Require no identity to retrieve a plan by its plan code

```
Statement:    The system shall require no account, no login, no stored
              identity and no prior use of the presenting device as a
              condition of retrieving the plan a plan code identifies.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § No accounts;
              Architecture.md § Persistence and cross-device transfer
Priority:     MUST
Verification: test — a plan code issued on one device returns its plan on a
              second device holding no session and no stored state for that
              plan, against a store holding no user record, and the retrieval
              creates none
Acceptance:   given a plan stored from one device and its plan code presented
              from a second device holding no session for that plan, when the
              caller is within the per-caller retrieval limit NFR-012 sets,
              then the stored plan is returned
              given a caller for whom the system holds no account and no
              stored identity, when that caller presents a plan code, then
              retrieval proceeds on the code alone and no account, login or
              identity is required of them
Status:       to-build
Depends-on:   FR-094
Volatility:   open-question — retrieval by a short code is the proposed
              resolution under
              `Architecture.md § Persistence and cross-device transfer` rather
              than a ruling, and `Architecture.md § What is unproven` records
              the code's own entropy as an open security decision with the
              store as the only thing behind it
Risk:         Read as a device or an account condition, this obligation
              inverts into the thing `SPECIFICATION.md § No accounts`
              excludes: a retrieval path that asks who is calling is an
              account in all but name, and the desk-to-phone case the code
              exists to close fails on the one journey it was built for. Read
              as an obligation to serve, it fails the other way and forbids
              the refusal NFR-012 requires, which would leave a dwelling
              coordinate behind a short code with nothing bounding how often
              it may be guessed.
Rationale:    The cited sections oblige that no identity stands between a plan
              code and its plan — no login, no stored identity, no server-side
              user record — and neither obliges that every presentation of a
              code is served. This record therefore binds what retrieval may
              be conditioned on, not whether a given attempt succeeds. The
              rate is NFR-012's, which refuses attempts beyond the configured
              per-caller limit and counts every attempt whether or not the
              code exists; it is named here, and its bound written into the
              first criterion, so that a reader meets the limit on this record
              rather than discovering it in another category's file.
Resolved-by:  —
```

## FR-101 — Make a stored plan's plan code available from its list entry

```
Statement:    The system shall make a stored plan's plan code available from
              that stored plan's entry in the list of stored plans.
Category:     Persistence and staleness
Source:       DESIGN.md § Returning to a stored plan
Priority:     SHOULD
Verification: test — each entry in a list of several stored plans makes a plan
              code available, and the code taken from one entry retrieves that
              entry's plan and no other
Acceptance:   given a device holding several stored plans, when the list of
              stored plans is presented, then each entry makes that stored
              plan's plan code available
              given the plan code taken from one entry, when it is presented,
              then the stored plan returned is that entry's and no other
Status:       to-build
Depends-on:   FR-094; FR-100
Volatility:   open-question — the cross-device transfer entry under
              `Architecture.md § Open questions owned by this document` lists
              code-keyed server-side storage as proposed and open to revision
Rationale:    `SHOULD` because retrieval by plan code is FR-100's and works
              without this record; what this adds is moving one stored plan of
              several deliberately, which
              `Architecture.md § Persistence and cross-device transfer` says
              the usual case never requires — the code is held locally, and
              the user sees it only when they want the plan elsewhere. The
              statement binds once built, which is why it reads `shall`.
Resolved-by:  —
```

## FR-102 — Identify each listed stored plan by origin, destination and date

```
Statement:    The system shall identify each entry in the list of stored plans
              by that stored plan's origin, its destination, and the date the
              plan was made.
Category:     Persistence and staleness
Source:       DESIGN.md § Returning to a stored plan
Priority:     MUST
Verification: test — every entry in a list of several stored plans carries all
              three, and two plans between the same origin and destination
              made on different dates are told apart from the listed content
              alone
Acceptance:   given a device holding several stored plans, when the list of
              stored plans is presented, then each entry carries that stored
              plan's origin, its destination and the date the plan was made
              given two stored plans between the same origin and destination
              made on different dates, when the list is presented, then their
              entries are distinguishable from the content each carries
Status:       to-build
Depends-on:   FR-094
Volatility:   settled — `DESIGN.md § Returning to a stored plan` names the
              three as the things a user actually remembers about a route
Risk:         A user who cannot tell which entry is the trip they are driving
              today has lost the plan as surely as if it had been deleted: the
              failure `SPECIFICATION.md § Route persistence` exists to prevent
              is reproduced one step later, by a list nobody can read.
Rationale:    The obligation is on what each entry carries and never on the
              order the entries appear in.
              `DESIGN.md § Returning to a stored plan` also states that the
              list runs most recent first; that is presentation, and the
              ruling recorded in `README.md § Category register` on FR-010
              puts an obligation on the set and never on presentation order.
              The second criterion is what makes the first an identification
              rather than a decoration: three fields that do not separate two
              plans between the same two towns identify neither of them.
Resolved-by:  —
```

## FR-103 — Answer an expired plan code and an unrecognised one alike

```
Statement:    The system shall answer a plan code that has expired and a plan
              code it does not recognise with the same plain explanation,
              offering from it a way to start a new plan.
Category:     Persistence and staleness
Source:       DESIGN.md § Returning to a stored plan
Priority:     MUST
Verification: test — an expired plan code and an unrecognised one produce
              responses carrying the same explanation and the same way on,
              neither presented as an error
Acceptance:   given a plan code whose stored plan is no longer held, when it
              is presented, then the response is a plain explanation offering
              a way to start a new plan, and is not presented as an error
              given a plan code that was never issued, when it is presented,
              then the response is indistinguishable from the response in the
              criterion above
Status:       to-build
Depends-on:   FR-100
Volatility:   open-question — the cross-device transfer entry under
              `Architecture.md § Open questions owned by this document` lists
              code-keyed server-side storage as proposed and open to revision
Risk:         Two distinguishable responses make the store an oracle for which
              plan codes exist, which is a working enumeration signal against
              objects holding a dwelling coordinate, per
              `Architecture.md § Personal data`. The user-facing half fails at
              the worst moment: a plan gone at its ceiling is met on the
              morning of departure, and an error there reads as the product
              having lost the work.
Rationale:    One record and not two, because it is one response: an
              explanation leaving the user nowhere to go is not the plain,
              unalarming answer the source requires, and an
              indistinguishability rule with nothing to be indistinguishable
              from asserts nothing. What the explanation says and where it
              appears is `DESIGN.md`'s and is not authored here. Why a plan
              code stopped resolving is deliberately outside the record — the
              retention bounds behind it are the non-functional lane's, and
              this record is what makes the causes look alike from outside.
Resolved-by:  —
```

## FR-104 — Extend a stored plan's rolling retention within its ceiling

```
Statement:    The system shall extend a stored plan's rolling retention period
              to run from each successful retrieval of that stored plan by its
              plan code, and shall not extend it past the ceiling measured
              from the plan's creation, both periods being defined under
              `Architecture.md § Persistence and cross-device transfer`.
Category:     Persistence and staleness
Source:       Architecture.md § Persistence and cross-device transfer;
              DESIGN.md § Returning to a stored plan
Priority:     MUST
Verification: test — a retrieval well inside the ceiling moves the rolling
              period to run from that retrieval, and a retrieval whose
              extension would pass the ceiling moves it only as far as the
              ceiling
Acceptance:   given a stored plan whose ceiling is further off than the
              rolling period would reach, when it is retrieved by its plan
              code, then its rolling retention runs from that retrieval
              given a stored plan whose ceiling is nearer than the rolling
              period would reach, when it is retrieved by its plan code, then
              its retention is extended only as far as the ceiling, which the
              retrieval does not move
Status:       to-build
Depends-on:   FR-100
Volatility:   proposed-constant — both periods under
              `Architecture.md § Persistence and cross-device transfer` are
              proposed defaults rather than measured results
Risk:         Without the reset a plan in ordinary use is lost to the idle
              timer between one journey and the next, deleted for an
              inactivity it never had. Without the bound a plan reopened just
              inside the rolling period is retained forever, and what is
              retained indefinitely is a dwelling coordinate under a stable
              identifier.
Rationale:    The relation is expressed and neither period is restated, so the
              record survives either figure moving. The two clocks are
              independent: a retrieval moves the rolling one and nothing moves
              the ceiling.
              `Architecture.md § Persistence and cross-device transfer` holds
              the plan code in local storage, so opening a stored plan from
              the list is a retrieval by its plan code and resets the clock
              exactly as presenting the code on another device does. What
              happens when either period runs out is not this record's —
              deletion and the retention bounds are the non-functional lane's,
              and what the user meets afterwards is FR-103's.
Resolved-by:  —
```

## FR-105 — Open a stored plan without the wizard and without the Turf API

```
Statement:    The system shall open a stored plan without requiring the
              initialization wizard to have been completed and without
              requiring the Turf API to be available, presenting the plan
              without the volatile data it cannot obtain rather than
              withholding the plan.
Category:     Persistence and staleness
Source:       DESIGN.md § Never gate stored plans on the wizard
Priority:     MUST
Verification: test — a stored plan opens with the initialization values
              incomplete, and opens with the Turf API unavailable, presenting
              the plan with the volatile data absent
Acceptance:   given a device whose initialization values are missing or
              incomplete, when a stored plan is opened, then the plan is
              presented and the initialization wizard is not required first
              given a Turf API that is unavailable, when a stored plan is
              opened, then the plan is presented with the volatile data absent
              rather than being withheld
Status:       to-build
Depends-on:   FR-094
Volatility:   settled — `DESIGN.md § Never gate stored plans on the wizard`
              states the obligation directly and names the failure it prevents
Risk:         An outage-triggered lockout produces exactly the failure
              persistence exists to prevent, on exactly the morning it
              matters, for a plan that is entirely intact — and it does so
              because a third party is down, which the user cannot act on and
              the stored plan does not need.
Rationale:    The gate is not weakened: `DESIGN.md § First-run initialization`
              still holds the planner closed until both steps are complete,
              because the gate exists to stop someone planning without the
              data that makes a plan good. Reading a stored plan is not
              planning. Nothing in a stored plan depends on the Turf API — the
              roads and the zones are geography — so the absent overlay is a
              degraded display and never a reason to withhold the plan.
Resolved-by:  —
```

## FR-106 — Present a reopened stored plan exactly as it was confirmed

```
Statement:    The system shall present a reopened stored plan with the zones
              the user confirmed, in the order they confirmed them, and shall
              not recompute it into a different plan.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Stored routes go stale
Priority:     MUST
Verification: test — a stored plan reopened long after it was made presents
              the confirmed zones in the confirmed order, and still does so
              once volatile data that has changed materially since has been
              refreshed
Acceptance:   given a stored plan reopened after an arbitrary interval, when
              it is presented, then its zones and their order are the ones the
              user confirmed
              given a stored plan whose volatile data has changed materially
              since it was computed, when the plan is presented and that data
              refreshed, then its zones and their order are still the ones the
              user confirmed
Status:       to-build
Depends-on:   FR-094
Volatility:   settled — `SPECIFICATION.md § Stored routes go stale` states
              that the stored plan does not change and is shown exactly as the
              user approved it
Risk:         Recomputing a stored plan discards exactly the judgement the
              confirmation step exists to capture — the user's own knowledge
              of which zones they have already taken, which the Turf API does
              not expose — and it discards it invisibly, because a silently
              substituted plan looks like a plan.
Rationale:    The route does not decay: the roads and the chosen zones are
              geography, and geography does not change. What decays is the
              volatile data attached to it, whose refresh is FR-107's. This
              record is the never-an-automatic-edit half of the source's rule;
              the telling half is FR-108's, for the one material change the
              system can determine.
Resolved-by:  —
```

## FR-107 — Refresh a reopened plan's volatile data after it is on screen

```
Statement:    The system shall refresh a reopened stored plan's volatile data
              after the plan is presented, and shall not withhold the plan
              while that refresh is outstanding.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Stored routes go stale;
              DESIGN.md § Returning to a stored plan
Priority:     MUST
Verification: test — a reopened stored plan is on screen before its volatile
              refresh completes, carries the refreshed values once it has, and
              stays on screen with that data absent when it cannot
Acceptance:   given a stored plan being reopened, when the plan is presented,
              then it is presented before the refresh of its volatile data has
              completed
              given a reopened stored plan on screen, when the refresh
              completes, then the ownership and points presented are the
              refreshed values
              given a reopened stored plan on screen, when the refresh cannot
              complete, then the plan stays on screen with the volatile data
              absent
Status:       to-build
Depends-on:   FR-094; FR-105; FR-028
Volatility:   settled — `SPECIFICATION.md § Stored routes go stale` states
              that the refresh happens in the background after the plan is
              already on screen
Risk:         A plan gated on its refresh reintroduces through the back door
              the lockout FR-105 removes, on the same morning. A plan
              presented with weeks-old ownership as though it were current
              makes the user skip zones they could have taken, which
              `DESIGN.md § Zones the user already owns` calls worse than
              showing nothing.
Rationale:    The ordering is the obligation: the plan first, the refresh
              after. Which data is volatile is
              `SPECIFICATION.md § Stored routes go stale`'s — ownership,
              current points, and anything derived from them — and is not
              enumerated again here. That a refresh may change what is
              displayed but never the plan is FR-106's. The third criterion is
              the branch a stored plan reaches whenever FR-105's outage holds,
              and without it that record and this one would each assume the
              other had covered it.
Resolved-by:  —
```

## FR-108 — Tell the user a round has rolled over before they act on the plan

```
Statement:    Where the system determines that a round boundary has passed
              since a stored plan was computed, it shall tell the user so when
              that stored plan is reopened, and what it tells them shall state
              that the route is unchanged.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Stored routes go stale;
              DESIGN.md § Communicating a round rollover
Priority:     MUST
Verification: test — a stored plan reopened after a determined round boundary
              carries the telling with the plan and states the route
              unchanged, the plan stays usable where the user does nothing,
              and a plan reopened within the same round carries no telling
Acceptance:   given a stored plan reopened after the system has determined
              that a round boundary passed since the plan was computed, when
              the plan is presented, then the user is told with it, rather
              than after they have acted on it, that a new round has started
              and that the route is unchanged
              given that telling, when the user takes no action on it, then
              the stored plan stays usable and its zones and their order are
              unchanged
              given a stored plan reopened with no round boundary determined
              to have passed since it was computed, when the plan is
              presented, then no such telling is presented
Status:       to-build
Depends-on:   FR-099; FR-107
Volatility:   open-question — round-rollover messaging is proposed and open to
              revision under
              `DESIGN.md § Open questions owned by this document`
Risk:         A rollover is the largest change that can occur between planning
              and departure, and it is good news: every zone in the plan is
              unclaimed again. Unstated, the user meets it as ownership
              markers that have silently vanished from weeks-old work, with no
              way to tell whether the plan itself survived — and that
              reassurance is the one thing the system can give with certainty.
Rationale:    What is obliged is the telling, its timing, and one thing it
              must say. The banner proposed under
              `DESIGN.md § Communicating a round rollover`, its form and its
              wording, is open to revision and is deliberately not hardened
              here; the second criterion carries only that the offered
              re-check is the user's to take, and the rule that nothing edits
              the plan is FR-106's. The detection is conditioned and not
              obliged: the round's start is derivable from the floor of
              `dateLastTaken` per
              `Architecture.md § Volatile and optional fields`, which records
              that as indicated by sampling rather than confirmed, and
              `Architecture.md § Round rollover` states the sync does not
              supply it. The third criterion is what makes the first
              informative — a telling presented on every reopening tells the
              user nothing.
Resolved-by:  —
```

## FR-116 — Carry a stale-data marking through storage and retrieval

```
Statement:    The system shall carry the marking that a result was built from
              a local synced copy of the zone set stale beyond the bound
              recorded under `Architecture.md § Retrieving zones` through to
              every reopening of the stored plan made from that result.
Category:     Persistence and staleness
Source:       SPECIFICATION.md § Route persistence
Priority:     MUST
Verification: test — a stored plan made from a marked result carries the
              marking when it is reopened, one made from an unmarked result
              carries none, and a refresh of the local synced copy since the
              plan was computed does not clear it
Acceptance:   given a result marked as built from a local synced copy of the
              zone set stale beyond the bound recorded under
              `Architecture.md § Retrieving zones`, when the plan made from it
              is stored and later reopened, then the reopened stored plan
              carries that marking
              given a result carrying no such marking, when the plan made from
              it is stored and later reopened, then the reopened stored plan
              carries no such marking
              given a stored plan made from a marked result, when it is
              reopened after the local synced copy of the zone set has been
              refreshed within that bound, then the marking is still carried
Status:       to-build
Depends-on:   FR-034
Volatility:   settled — `SPECIFICATION.md § Route persistence` states that a
              stored route keeps everything, and the bound under
              `Architecture.md § Retrieving zones` is recorded rather than
              proposed
Risk:         A marking lost on storage defeats FR-034 outright for exactly
              the plans that live longest: a plan built from a copy stale
              beyond the bound reopens indistinguishable from one built from a
              current copy, and the degradation the ruling accepted on
              condition of disclosure becomes invisible on the morning of
              departure, when the user acts on it.
Rationale:    `SPECIFICATION.md § Route persistence` says a stored route keeps
              everything, and the marking is a property of the result rather
              than of the session that produced it, so it is on that list;
              FR-034 is corroboration rather than source, being defeated
              outright by a marking that does not survive storage. The record
              is retention and survival only: what a reopened stored plan
              tells the user about the marking belongs to
              `Recommendation disclosure` and is not authored here. The
              marking is fixed at computation and is not the round-scoped
              ownership marker FR-098 keeps out of a stored plan, which is why
              a later refresh of the local synced copy does not clear it.
Resolved-by:  —
```
