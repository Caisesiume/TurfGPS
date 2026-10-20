# ADR-0007 — The lead merges approved, green PRs through a guarded script

**Status:** accepted — 2026-10-20, by the Owner on #229; in force from the merge of the PR that adds it.
**Source:** #229 § Owner decision — the Owner's request, and the scope the Owner chose through AskUserQuestion in the lead session, with the lead's recommendation. Where this record and that section differ, the issue is the Owner's and this record is the defect.
**Relation:** amends ADR-0005 in one clause, the one it homed in `team-protocol § Posting a ruling`: "No agent merges — the Owner does" becomes the rule below. It narrows ADR-0004 § E1, whose backstop against an adversary holding a trusted account is human review before merge: that review now backs only `owner-merge` PRs (§ Consequences). The rest of ADR-0005 stands, and so does #202: approval is still only a TheReviewNinja `APPROVED` review at the head.

## Context

Under ADR-0005 every PR waited for the Owner to merge it, after TheReviewNinja had approved it and the runner had passed it. The merge added no judgement the review and the checks had not already made, and it made the Owner the queue every finished change waited in. The Owner asked that the lead may merge PRs itself, and chose the scope: all PRs except the Owner-reserved ones.

## Decision

- **The lead merges** a PR, and only through `scripts/team/merge.sh`, when the script finds all of: the PR open and authored by `Caisesiume`; its head equal to the SHA given; the latest TheReviewNinja review at that head `APPROVED`; the required checks `gates` and `migrations` successful at that head, a missing check counting as a failure; no `awaiting-human` label; and, when `scripts/loop/diff-domains.sh` lists a safety path among the PR's files, a TheReviewNinja safety verdict `pass` at that head. Any other finding is a refusal, and nothing is merged.
- **Owner-reserved, never merged by the lead:** a change to a safety rule, accessibility classification, a requirement verified by human judgement, an artifact from an untrusted author, and anything labelled `awaiting-human`. The lead labels such a PR `owner-merge`, and the script refuses it.
- **The merge runs under the default account** (`Caisesiume`), never with `GH_JUDGE_TOKEN`: the judge approves and the author's account merges, so no one identity both judges and lands a change. That separation is held by rule, not by the environment (§ Consequences). The method is a merge commit, pinned to the approved head with `--match-head-commit`.
- **Dry run by default.** `merge.sh` prints the merge it would make; `--apply` makes it.
- **The Owner turns on branch protection for `main`** requiring the `gates` and `migrations` checks: a gate outside any agent's reach, so that `merge.sh` is not the only one. On 2026-10-20 it was not yet on: `GET /branches/main/protection` answered 404, and the active ruleset's `required_status_checks` list was empty.
- The PR that adds this record is the last merged by hand under the old rule.

## Consequences

**Gained:** an approved, green PR lands without waiting for the Owner; the Owner's attention goes to the reserved categories, which the label and the script keep in the Owner's hands.

**Given up, stated rather than implied:**

- A last human look before `main` moves. A defect the critic and the runner both missed now lands without one.
- **ADR-0004 § E1's backstop, for every PR the lead merges.** ADR-0004 names human review before merge as the backstop against an adversary holding a trusted account; that review now backs only `owner-merge` PRs. Any other PR a trusted account can carry to `main` through the judge's approval and the lead's merge with no human looking.
- **The judge's token and the merging account share one environment.** Every session holds `GH_JUDGE_TOKEN` (`docs/TEAM.md § Setting up a machine`) beside the default account, so the separation in § Decision is held by rule only: a session that disobeys it can both approve and merge.
- The reserved categories are guarded by a label, and applying it is the lead's judgement. The script cannot tell a safety-rule change that was never labelled from an ordinary one; the safety-path check catches only what `diff-domains.sh` lists, which is a floor (ADR-0005).
- The checks and the merge are not atomic. The head pin makes GitHub refuse a moved head, but a review, label or verdict that lands between the script's reads and the merge is not seen.
- `merge.sh`'s test stubs `gh`, so it proves the decisions, not the `--jq` filters; those were run live, read-only.

**Reversibility: high.** Reverting the PR restores "no agent merges"; nothing merged under this rule depends on it.
