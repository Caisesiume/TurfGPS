# ADR-0004 — Artifact authorship is a trust boundary

**Status:** accepted — 2026-09-07
**Source:** the Owner's decision of 2026-09-07, quoted verbatim in `§ The decision as given`. Unlike ADR-0001 through ADR-0003 there is no separate directive file to adapt, so the order itself is kept in this record and the sections below are its ratified form on this repository.
**Relation to ADR-0001 / ADR-0002 / ADR-0003:** a boundary added **around** the artifact mechanism ADR-0001 established, not a replacement for any part of it. No agent is added, no seat collapsed, no safety floor weakened, and no decision in those records reopened. Artifacts remain the medium; what changes is that reading one now depends on who wrote it.

**Why this record is numbered 0004.** Four sentences across three files state that no ADR-0004 exists — the directive-4 row in `docs/adr/README.md § Contents`, two in ADR-0001 (its stability note and its runtime-findings amendment), and `docs/adr/agent-org-directive-4.md` itself. Every one of them is a statement that a *particular directive* produced no record, because it decided nothing new; none reserves the number. The directive that made them says what admits one: "a new ADR only for a genuinely new decision that belongs in none of the existing records" (`docs/adr/agent-org-directive-4.md` §26). A trust boundary on artifact authorship belongs in none of them, so this is that ADR, taking the next unassigned number as `docs/adr/README.md § Form` requires. Those four sentences read as inexact once this record exists, and they are **not** edited here. Each remains true of the thing it describes — the directive it is about — one of the four is the Owner's verbatim directive text and is unalterable on any branch, and the row this record adds to `docs/adr/README.md § Contents` sits directly beneath the one that could mislead. Rewriting the rest from this branch would touch two more files to change nothing a reader acts on.

## Context

`Caisesiume/TurfGPS` is a **public** repository. That was verified rather than assumed, along with everything else in the table below.

| Checked | Found, 2026-09-07 |
|---|---|
| Repository visibility | `gh repo view --json isPrivate,visibility` → `isPrivate: false`, `visibility: PUBLIC`. Any GitHub account can comment on any open pull request. |
| What authenticates an artifact | Nothing but the artifact's own first key. `agent-handoffs § The structured block comes first` states the design in its own words: "the artifact names itself, rather than the instrument guessing from the author or the wording." |
| What a consumer retrieves | The selector proposed on PR #187 pages the PR's comments, selects on `startswith("artifact: validation_result")`, and projects each match down to `.body` alone. `.user.login`, `.created_at` and `.id` do not survive that projection. |
| Whether selection on the author was considered | It was ruled out on the same line: "A consumer selects on the declared class — not on your name, not on your wording" (PR #187, in `@validation-agent`). |
| Whether hostile input has been seen here | Yes. `#152` records a prompt-injection block observed from outside this project, reported by an agent and still open. |
| What trust boundary already exists | Exactly one, covering the judge lane only: `turfgps-board-ops § The one thing the fallback must never do` requires a ruling's author to read back as `TheReviewNinja` and makes any other value a stop-and-report. No class of artifact other than a judgment is covered. |

Put together, these are not six observations but one mechanism. A validation result is a comment whose body opens with `artifact: validation_result`; a consumer finds it by that string and reads what follows. **Any account on GitHub can write that string.** `@linus-security-critic` demonstrated it against PR #187: a comment declaring `status: pass` with the live head SHA is read as a pass for gates that nobody ran, by a consumer that has no field left in its hands with which to doubt it. The projection is what closes off the doubt — a consumer that wanted to check provenance could not, because the published command threw the author away before the consumer saw it.

The judge lane shows the shape of the answer, having already needed it. It does not extend to this one: it is a rule about a *verdict's* signature, and a validation result is deliberately not a verdict.

**On the finding IDs cited below.** `@linus-security-critic` filed these as `SEC-02` and `SEC-04` in its review of PR #187. Those identifiers are scoped to that review and are not repository-wide: `ADR-0002 § O17` already discusses a different `SEC-02`, from PR #67, about reverse-proxy exposure. Every reference here therefore names the pull request alongside the code. The findings had not been posted to PR #187 when this record was written — checked 2026-09-07, the PR carried no reviews and no comments — so they are cited by the review that produced them, while the mechanism they describe is verified above against the repository itself.

## The decision as given

The Owner's decision of 2026-09-07, verbatim:

> "Harden the readers, as that option suggest. But it should remain the same way as today, but only add @Caisesiume and @TheReviewNinja as trusted users. All other cases, with other users than them, the agents MUST escalate the input from such user to me in question form to deem it's readiness and correctness. .user.login and created at are still needed and a great approach in addition."

## Decision

### D1 — The allowlist is two names, and for them nothing changes

Artifact authorship is trusted for exactly **`Caisesiume`** and **`TheReviewNinja`**. An artifact authored by either is read exactly as it is read today: the declared class selects it, the body is consumed, and no new step is added on the path that carries all of this repository's real traffic.

This is the Owner's "it should remain the same way as today," and it is a deliberate choice of the cheapest possible hardening. The reader gains a comparison against a two-element list. It gains no policy engine, no signature scheme, and no new artifact class — every one of which would be a larger change to the mechanism than the vector justifies.

The list is closed. Adding a third name is an Owner decision and an amendment to this record, not a judgement any agent makes at read time.

### D2 — Any other author is escalated to the Owner as a question, and is never discarded

An artifact whose author is not on the list is **neither trusted nor dropped**. The agent stops consuming it and puts it to the Owner **in question form**, for the Owner to judge — in the Owner's own terms — "it's readiness and correctness."

**Discarding it would be wrong, and this record says so plainly rather than leaving it to be inferred.** A comment from an account outside the list is not by construction an attack. It may be a legitimate contribution: a reader who reproduced a failure, a passer-by who spotted a real defect, a future collaborator. Silently dropping it makes the repository unable to receive that, and does so invisibly — nobody is told, so nobody can notice the policy is costing something. The one party entitled to decide whether such input is sound is the Owner, and the escalation exists to put it there rather than to have an agent guess in either direction.

Both failure directions are therefore refused. Consuming an unknown author's artifact treats a stranger's claim as a measurement; discarding it silently treats a contribution as noise. The escalation is the only path that resolves neither way on its own.

The escalation travels the route every human-bound question already travels — `handoff-payloads § Escalation packet` for the shape that reaches the Owner, and `handoff-payloads § Structured uncertainty (blocked)` where the agent's own domain cannot settle it. This record adds a trigger to that machinery; it does not add a second channel beside it.

### D3 — The selector retains `.user.login` and `.created_at`

The Owner called this out specifically: "`.user.login` and created at are still needed and a great approach in addition." Any retrieval that a consumer is instructed to run **carries the author and the creation timestamp through to the consumer**. A projection that discards them is a defect, because it removes from the reader the only fields on which D1 and D2 can be decided at all — an allowlist is unenforceable against a body that arrived alone.

`.created_at` earns its place twice. The second use is the ordering gap `@linus-security-critic` filed as `SEC-04` on PR #187: when several results exist at one head SHA, a consumer needs a rule for which one is the result, and a timestamp makes that tie-break computable instead of leaving it to whichever comment the page happened to return first.

**What is normative here is that the three fields survive projection**, not any particular command. The command's one home is the agent definition that publishes it, and this record deliberately does not become a second one — `local-gates § Documentation gates` gate 2 is the reason, and the selector fragments described in `§ Context` above are dated evidence of a shape observed on a specific day, not a copy of the instruction.

### D4 — No artifact is a clearance

An artifact reporting `pass` is **evidence that a measurement was taken**. It is an input to `@pr-judge`'s ruling and never a substitute for it. This holds for artifacts from allowlisted authors too — D1 admits an artifact to being *read*, and admits it to nothing else.

This is the same principle `docs/DELIVERY.md § Escalation and human judgement` already states for the board: "a clean board is a **recommendation to the human**, not an approval." The reason it is restated as a decision rather than left as a citation is that the allowlist creates the temptation it guards against: a trusted author makes a `pass` feel dispositive precisely when the mechanism has just been described as hardened. It has not been. It has been narrowed.

## Consequences

**What this obliges:**

- **Every consumer of a declared artifact acquires a branch it did not have.** Read, escalate, or stop — decided per artifact, on the author. The cost is one comparison against a two-element list on the path that carries essentially all traffic, and this is the cheapest form the decision could take.
- **Two files must change, and neither changes here.** The reader hardening lands in `@validation-agent` and in `agent-handoffs`, which is exactly the diff PR #187 has open and under judgment. This record was written without touching either: a second writer in a file already being judged is the hazard the one-writer rule exists to prevent, and a decision record that quietly implements itself is not a decision record. The implementation cites this ADR when it lands.
- **A new stall is possible.** An artifact from an unlisted author now blocks its consumer until the Owner answers. That is a stall by design, and preferable to both alternatives, but it is a real cost and it is the Owner who pays it.
- **The list is a maintenance surface.** Two names are written into the repository's trust model. An account rename, a new collaborator, or a second review identity is now a change to this record.

**What this gives up:**

- **Authorship on GitHub is not authentication.** An allowlist tests a login, and a login is not a signature. It is what is available without a key-management scheme this project does not have and does not need; it stops an anonymous stranger, which is the vector actually demonstrated.
- **It does not close the collapse `SEC-02` recorded on PR #187, and pretending otherwise would be the worse outcome.** The default CLI is authenticated as `Caisesiume` — the same account that authors the branches under review. `turfgps-board-ops § The one thing the fallback must never do` names that in its own words for the judgment case: a judgment posted through the plain CLI "would appear as `Caisesiume` — the author approving their own work ... and it would look completely normal in the history." An allowlist admitting `Caisesiume` therefore cannot tell a worker's artifact from the branch author's own tooling, because at the transport layer they are the same account. **The stranger vector is closed. The self-authorship vector is not**, and no widening of this list can close it — a list that admits an identity admits everything wearing it. What answers that one is D4: the artifact is not the clearance, so an artifact that authenticates as the author still buys nothing but a reading.
- **Nothing is gained against a compromised trusted account**, and the same D4 argument is the only thing standing behind that case too.

**Reversibility: high.** The decision is a comparison and two retained fields. Removing the comparison restores today's behaviour exactly; the retained fields are additive and harm nothing if never read. No data migrates and no artifact already posted becomes invalid.
