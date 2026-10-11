# Team

How work on TurfGPS is planned, built, reviewed and merged. It replaced `DELIVERY.md` on #210, as `adr/ADR-0005-lean-team.md` records. The rules every agent follows are the `team-protocol` skill; this page is the shape. Skills are read as files — `.claude/skills/<name>/SKILL.md` — because no agent holds the Skill tool.

## The team

| Agent | Model | Job |
|---|---|---|
| `engineering-lead` | fable | Plans and assigns every ticket, syncs the team, and is the only agent that talks to the Owner. Writes no product code. The session's default agent. |
| `analyst` | inherit | Requirements → stories with acceptance criteria and `Blocked by:` edges. Keeps `Requirements/` and its `DECISIONS.md`. Never invents scope. |
| `backend-dev` | inherit | The Go service, PostGIS and operations — one ticket at a time, in its own worktree from `origin/main`. |
| `frontend-dev` | inherit | The React client, built to `DESIGN.md`. |
| `critic` | inherit | Reviews one PR through the lenses the lead names and posts the one ruling as TheReviewNinja. |
| `safety-sentinel` | inherit | Mandatory when `diff-domains.sh` flags a safety path. Rule-based; its blocking findings need no falsifier. |

At most two reviewers per PR: the critic, and the sentinel on a safety path. Mechanical work is a script, never an agent.

## The flow

1. The analyst puts stories on the board, with acceptance criteria and dependency lines.
2. `board-sync.sh` reconciles statuses and promotes Ready by `dependents.sh`.
3. The lead takes a Ready ticket, picks the developer by stack, sets `In progress` and dispatches by reference only.
4. The developer implements, tests and documents, runs `gates.sh`, then tries to break the change: at least one falsifier per acceptance criterion, recorded in the PR body. Then it opens the PR.
5. The lead dispatches `safety-sentinel` if a safety path is touched, and the critic with at most two lenses from the table below.
6. The sentinel posts its verdict, and the critic posts the ruling and the GitHub review, both through `post-review.sh`. Blocking findings each carry a falsifier.
7. A remand fixes only the blocking findings. A PR still blocked after cycle 3 goes to the Owner, with the open findings and a recommendation.
8. The Owner merges; `board-sync.sh` runs.

| Change | Lenses |
|---|---|
| docs | accuracy, references |
| Go | correctness, security |
| SQL / migration | data integrity, security |
| web | states, accessibility, design |
| scripts / CI | correctness, supply chain |

Every agent answers in the typed envelope — at most 25 lines of YAML with a `choice`, `score` or `p_true` — and the thresholds in `team-protocol § Thresholds — act, look again, or ask` decide whether to act, look again or ask. That is the Jev pattern the Owner chose on #210: typed answers and thresholds, built with our own agents, no new vendor.

## Scripts

| Script | What it answers |
|---|---|
| `scripts/loop/fingerprint.sh engineering-lead` | has anything moved since the lead last looked — the first thing every wake runs |
| `scripts/team/board-sync.sh` | the board made true to GitHub: Done, In review, Ordered Revision, Backlog → Ready. Dry run unless `--apply` |
| `scripts/loop/dependents.sh` | who waits on `#N`; with `--all`, which stories are free |
| `scripts/loop/diff-domains.sh` | a diff's domains, its candidate lenses, and the safety paths among its files |
| `scripts/team/gates.sh` | every gate the change owes, as one typed result: pass, fail or incomplete |
| `.github/workflows/gates.yml` | the same `gates.sh --all` on every pull request and every push to `main`, on a GitHub-hosted runner carrying what the reference host lacks: a C compiler for the race detector and a container runtime for the image test. Where it and a local run disagree, the runner's result is authoritative — a local `unrun`, or a red naming a missing host capability, is the host's; the runner's red is the change's. Making `gates` a required status check is branch protection the Owner sets |
| `scripts/team/post-review.sh` | the one way a verdict, ruling or review is posted. The only user of `GH_JUDGE_TOKEN`; refuses a moved head, an approval under `awaiting-human`, and an approval on a safety path without a passing safety verdict at that head; never merges |
| `scripts/gates/size-ratchet.sh` | no agent or skill past its baseline, and all of them under 100 KB together |

`scripts/loop/claim.sh` and `scripts/loop/owed-work.sh` stay in the repository but are not part of the flow.

## Escalation and human judgement

What reaches the Owner, and the four categories that always do — safety rules, accessibility classification, requirements verified by `human-judgement`, and an artifact from an untrusted author — is `team-protocol § Escalation and human judgement`. On those, a clean review is a recommendation to the Owner, never an approval. Only the lead asks: one question, its options, a recommendation for each set, and the impact. The question and the answer are recorded on the issue or PR, and a decision of consequence becomes an ADR.

## Setting up a machine

- `GH_JUDGE_TOKEN` in the environment holds TheReviewNinja's token. Nothing but `post-review.sh` reads it, and that script checks the login before it posts.
- Allow `bash scripts/team/post-review.sh *` in the session's permissions. Machine-local settings are not in the repository.
- `gh` is authenticated as Caisesiume for everything that is not a judgment.
- Remote control is per user: `/config`, enable it for all sessions, so the lead's questions reach the Owner's phone.

## Measured

After each merge the lead posts one line on #205: the PR, its cycles, each ruling's `p_ready`, and the tokens it cost. That log is what the thresholds are re-tuned from, and what decides whether this shape stays.
