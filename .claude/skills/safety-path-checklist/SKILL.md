---
name: safety-path-checklist
description: What every agent must know before touching a TurfGPS safety path — accessibility classification, the enforceable exclusions, the absolute time ceiling, the uncertain bucket, and the proposal/measurement boundary. Use before implementing or reviewing anything in access classification, stop selection, routing exclusions, time budgeting, or the constants that feed them.
---

# Safety-Path Checklist

TurfGPS moves no money. Its equivalent stakes are physical: a defect here proposes that a driver stop somewhere they must not, or promises a stop they cannot make. A **safety path** = access classification, stop-position selection, routing exclusions, the time ceiling, and the constants feeding any of them. Diffs here always convene @safety-sentinel at the bench, and per `docs/DELIVERY.md` they always reach a human.

## The measure of success

> **Not that every zone is classified, but that no zone is classified confidently and wrongly.**

That is from *Accessibility scope for the first release* in `SPECIFICATION.md`, and it is the sentence to test a change against. A change that classifies more zones while loosening confidence is a regression, however good the coverage number looks.

## Enforceable exclusions (hard rules — a zone failing any is excluded regardless of Turf value)

- No stop on a motorway, motorway link, or any road whose speed limit is not **established** to be at or below the maximum for a stopping road. It is a test the road must pass, so an unknown limit excludes it. What establishes a limit, what does not, and the constant itself are stated under *Enforceable exclusions* in `SPECIFICATION.md` — enforce it from there and hold no copy here. A nearby rest area, service road, parking area, or exit may still make the zone accessible; the high-speed carriageway itself never is.
- No stop on a road not marked drivable by the map data.
- No direct road-access classification where road and zone are at incompatible levels, or where bridge/tunnel/`layer` data indicates they do not meet.
- No accessibility across an access path that is absent, disconnected, or implausibly steep.
- No recommendation may assume capture while the vehicle is moving. **Every stop is a stop.**
- No route through areas the map data marks private or access-restricted.

**Time is never grounds for relaxing any of these.** Time is the quantity being optimized; safety is a constraint on the search space, not a term in the objective function. A reviewer that sees safety traded for minutes returns a blocking finding.

## What the data cannot verify — and must not be claimed

Whether stopping is *legally permitted* at a roadside position is generally absent from map data at usable coverage. So is whether a manoeuvre would be locally unsafe. The system honours restrictions where recorded, treats them as **unknown** where not, labels any stop whose legality could not be established, and never presents an unverified roadside stop with the confidence of a mapped parking area.

## The absolute ceiling

The stated additional-time limit is a **soft target**; **115% of it is a hard ceiling**, re-checked after every change during review. No recommendation exceeds it for any reason, however valuable the zone. An uncertain stop accepted during review is tested against its conservative upper bound, and where that bound would breach 115% **the acceptance is refused**. A stop that might breach the ceiling is treated as one that does.

## The uncertain bucket

Uncertain zones **never enter the optimizer's cost model** and never contribute to a journey's value — a time estimate the system does not trust cannot be balanced against one it does. They carry no score, never influence ranking, and never appear inside a route's stated additional time. They are offered only as a reserve when the user rejects a zone, and accepting one visibly widens the route's estimate.

Confidence is a **gate, not a term**. Blending it into the score would make a well-understood mediocre zone and a poorly-understood excellent one indistinguishable.

## The proposal boundary

Nearly every constant in `CalculationSpecification.md` is a **proposed default**, not a measurement. Two consequences bind every agent:

- **Never quote a proposed constant to a user as established.** The manoeuvre timings in particular are uncalibrated guesses and the largest single source of error in the time model.
- **Never harden a proposal into a fixed requirement or an unexplained literal.** Requirements cite the constant by name; code reads it from configuration and carries its documented origin.

## Domain facts are verified, not inferred

Every assumption about Turf mechanics made by reasoning rather than checking has turned out wrong. `blocktime` is not takeover time; the Region Lord bonus is global; zones expose only a coordinate; `currentOwner` is round-scoped. The verified facts are in `Architecture.md` under *Data sources and constraints*, and the API self-documents at `GET https://api.turfgame.com/v5`. **Check it rather than inferring.** A domain assertion with no traceable source is a blocking finding.

## Non-negotiables

- Detour cost is obtained by **routing**, never inferred from geometry. A straight-line estimate beside a dual carriageway can be twenty minutes wrong in the direction that matters.
- Any geometry that must agree with itself comes from **one** routing engine — the car and walk halves of a stop snapped to different graphs fail silently, with every stop still yielding a plausible number.
- A model is stated in exactly **one** document. A formula restated in code comments, in requirements, or in another document is a second thing to keep correct, and the two will diverge.
