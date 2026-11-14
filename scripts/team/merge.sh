#!/usr/bin/env bash
# merge.sh — the one way an agent merges a PR (#229; ADR-0007).
#
# The lead runs it. It merges only what TheReviewNinja approved and the runner
# passed, at one head, and never a PR the Owner reserved. It runs under the
# default account and never with GH_JUDGE_TOKEN: every gh it runs has GH_TOKEN,
# GITHUB_TOKEN and GH_JUDGE_TOKEN removed from its environment, and it refuses
# to run unless that account is Caisesiume. A merge is the author's act, never
# the judge's.
#
# REFUSED, exit 1, before anything is written — #229 § Owner decision:
#   - a PR that is not open
#   - an author other than Caisesiume: an untrusted author's artifact is the
#     Owner's (ADR-0004)
#   - a moved head: the PR's head is not --sha
#   - the label awaiting-human (the Owner is deciding) or owner-merge (the
#     Owner merges it: a safety rule, accessibility classification, a
#     human-judgement requirement, an untrusted author)
#   - no approval at --sha: the latest review by TheReviewNinja whose commit is
#     --sha is not APPROVED. A review by another login, or at another SHA, is none.
#   - a required check not green at --sha: for each of `gates` and `migrations`,
#     the latest GitHub Actions check run at --sha is not completed/success.
#     A missing run is a refusal, never a pass.
#   - a safety path without a passing safety verdict at --sha, read exactly as
#     post-review.sh reads it: diff-domains.sh over GitHub's file list, with the
#     floor the union of main's and the PR head's (#252), and the latest
#     TheReviewNinja comment carrying the safety marker for --sha.
#
# DRY RUN BY DEFAULT: it runs every check and prints the merge it would make.
# --apply makes it — `gh pr merge --merge --match-head-commit <sha>`, so GitHub
# itself refuses if the head moved after the checks — then reads the PR back.
#
# WHAT IT CANNOT SEE: a review, label or push landing between its reads and the
# merge, beyond the head pin; and whether a PR that should carry owner-merge
# does. Applying that label is the lead's judgement.
#
# Usage: merge.sh --pr N --sha SHA [--apply]
# Exit:  0 merged (dry run: would merge) · 1 refused · 2 cannot run
#
# Tested by scripts/team/tests/merge-refusals.sh, with gh stubbed.

set -uo pipefail
set +x

REPO="${TEAM_REPO:-Caisesiume/TurfGPS}"
GH="${GH:-/c/Program Files/GitHub CLI/gh.exe}"
DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || DIR=''
DOMAINS="$DIR/../loop/diff-domains.sh"
JUDGE='TheReviewNinja'
AUTHOR='Caisesiume'
CHECKS='gates migrations'
T=$'\t'

die()    { printf 'merge: cannot run — %s\n' "$*" >&2; exit 2; }
refuse() { printf 'merge: refused — %s\n' "$*" >&2; exit 1; }

# Every gh runs under the default account, with no token in its environment.
rd() { env -u GH_TOKEN -u GITHUB_TOKEN -u GH_JUDGE_TOKEN "$GH" "$@"; }

# The safety floor is the union of main's and the PR head's (#252): a PR may add
# to its own gate, never remove from it. main's diff-domains.sh, beside this
# script, is run. The head's copy is fetched at --sha and only READ: the one
# regex on its `candidates=` line is matched against the files as data, and
# nothing else in it runs. It is fetched only when the PR changes that file;
# otherwise the head's floor is the merge base's, and any entry main has since
# dropped was dropped by a change that already passed this gate. A changed copy
# whose floor cannot be read stops the script (exit 2), never falls back to main's.
# post-review.sh carries the same two functions; change them together.
FLOOR_FILE='scripts/loop/diff-domains.sh'
head_floor() {
  local src pat rc
  src="$(rd api -H 'Accept: application/vnd.github.raw+json' \
    "repos/$REPO/contents/$FLOOR_FILE?ref=$SHA" 2>/dev/null)" || return 1
  pat="$(printf '%s\n' "$src" | tr -d '\r' \
    | sed -n "s/^candidates=.*| grep -E '\([^']*\)' .*\$/\1/p")"
  [ -n "$pat" ] && [ "$(printf '%s\n' "$pat" | wc -l)" -eq 1 ] || return 1
  grep -qE -- "$pat" </dev/null 2>/dev/null; rc=$?
  [ "$rc" -le 1 ] || return 1
  printf '%s\n' "$pat"
}
union_floor() {   # union_floor <main's candidates, ", "-joined> <the files, one per line>
  local pat extra
  printf '%s\n' "$2" | grep -qxF "$FLOOR_FILE" || { printf '%s' "$1"; return 0; }
  pat="$(head_floor)" || return 1
  extra="$(printf '%s\n' "$2" | grep -E -- "$pat")"
  { printf '%s\n' "$1" | sed 's/, /\n/g'; printf '%s\n' "$extra"; } \
    | awk 'NF && !seen[$0]++' | paste -sd, - | sed 's/,/, /g'
}

# ── arguments ────────────────────────────────────────────────────────────────
PR=''; SHA=''; APPLY=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --apply) APPLY=1; shift; continue ;;
    --pr|--sha) [ "$#" -ge 2 ] || die "$1 needs a value" ;;
    *) die "unknown argument: $1" ;;
  esac
  case "$1" in --pr) PR="$2" ;; --sha) SHA="$2" ;; esac
  shift 2
done
[[ "$PR" =~ ^[0-9]+$ ]] || die "--pr must be a PR number"
[[ "$SHA" =~ ^[0-9a-f]{40}$ ]] || die "--sha must be the full 40-character head SHA that was approved"

# ── the account the merge runs as ───────────────────────────────────────────
who="$(rd api user --jq .login 2>/dev/null | tr -d '\r')" || who=''
[ -n "$who" ] || die "the default gh account could not be read"
[ "$who" = "$AUTHOR" ] || die "the default gh account is '$who', not $AUTHOR; a merge never runs as anyone else"

# ── the PR as it is now ─────────────────────────────────────────────────────
pr_line="$(rd pr view "$PR" -R "$REPO" --json state,headRefOid,author,labels \
  --jq '[.state, .headRefOid, .author.login, ([.labels[].name] | join(","))] | @tsv' 2>/dev/null)" \
  || die "could not read PR #$PR"
IFS="$T" read -r state head author labels <<< "$(printf '%s' "$pr_line" | tr -d '\r')"
[ -n "${head:-}" ] || die "could not read PR #$PR's head"
[ "$state" = OPEN ] || refuse "PR #$PR is $state, not open"
[ "$author" = "$AUTHOR" ] || refuse "PR #$PR is by '$author', not $AUTHOR; an untrusted author's PR is the Owner's"
[ "$head" = "$SHA" ] || refuse "the head moved: PR #$PR is at $head, the approval was checked at $SHA"
case ",$labels," in *,awaiting-human,*) refuse "PR #$PR is labelled awaiting-human; the Owner is deciding" ;; esac
case ",$labels," in *,owner-merge,*) refuse "PR #$PR is labelled owner-merge; the Owner merges it" ;; esac

# ── the approval at this head ───────────────────────────────────────────────
reviews="$(rd api "repos/$REPO/pulls/$PR/reviews" --paginate \
  --jq '.[] | [.user.login, .commit_id, (.submitted_at // ""), .id, .state] | @tsv' 2>/dev/null)" \
  || die "could not read PR #$PR's reviews"
review="$(printf '%s\n' "$reviews" | tr -d '\r' | awk -F'\t' -v judge="$JUDGE" -v sha="$SHA" '
  $1 == judge && $2 == sha && $3 != "" { print $3 "\t" $4 "\t" $5 }' \
  | sort -t "$T" -k1,1 -k2,2n | tail -n 1 | cut -f3)"
review="${review:-none}"
[ "$review" = APPROVED ] || refuse "PR #$PR has no $JUDGE approval at $SHA (latest review there: $review)"

# ── the required checks at this head ────────────────────────────────────────
runs="$(rd api "repos/$REPO/commits/$SHA/check-runs" --paginate \
  --jq '.check_runs[] | [.name, (.app.slug // ""), .id, .status, (.conclusion // "")] | @tsv' 2>/dev/null)" \
  || die "could not read the check runs at $SHA"
for check in $CHECKS; do
  run="$(printf '%s\n' "$runs" | tr -d '\r' | awk -F'\t' -v c="$check" '
    $1 == c && $2 == "github-actions" { print $3 "\t" $4 "/" ($5 == "" ? "none" : $5) }' \
    | sort -t "$T" -k1,1n | tail -n 1 | cut -f2)"
  run="${run:-missing}"
  [ "$run" = completed/success ] || refuse "check '$check' at $SHA is $run, not completed/success"
done

# ── a safety path needs a passing safety verdict at this head ───────────────
files="$(rd api "repos/$REPO/pulls/$PR/files" --paginate \
  --jq '.[] | .filename, (.previous_filename // empty)' 2>/dev/null)" \
  || die "could not read PR #$PR's files, so whether it touches a safety path is unknown"
[ -n "$files" ] || die "PR #$PR lists no files, so whether it touches a safety path is unknown"
[ -n "$DIR" ] && [ -f "$DOMAINS" ] || die "scripts/loop/diff-domains.sh is not where this script expects it"
domains="$(printf '%s\n' "$files" | tr -d '\r' | bash "$DOMAINS" -)" \
  || die "diff-domains.sh could not classify PR #$PR"
grep -q '^safety_path_candidates: ' <<< "$domains" || die "diff-domains.sh printed no safety_path_candidates line"
safety_paths="$(printf '%s\n' "$domains" | sed -n 's/^safety_path_candidates: \[\(.*\)\]$/\1/p')"
safety_paths="$(union_floor "$safety_paths" "$(printf '%s\n' "$files" | tr -d '\r')")" \
  || die "PR #$PR changes $FLOOR_FILE and its floor at $SHA could not be read, so whether it adds a safety path is unknown"
[ -n "$safety_paths" ] || safety_paths='none'
safety_verdict='not needed'
if [ "$safety_paths" != none ]; then
  comments="$(rd api "repos/$REPO/issues/$PR/comments" --paginate \
    --jq '.[] | [.user.login, .created_at, (.body | split("\n") | .[0] | rtrimstr("\r"))] | @tsv' 2>/dev/null)" \
    || die "could not read PR #$PR's comments to find a safety verdict"
  want="<!-- team:safety v1 pr=$PR sha=$SHA result="
  latest="$(printf '%s\n' "$comments" | tr -d '\r' | awk -F'\t' -v judge="$JUDGE" -v want="$want" '
    $1 == judge && ($3 == (want "pass -->") || $3 == (want "block -->")) {
      r = substr($3, length(want) + 1); sub(/ -->$/, "", r); print $2 "\t" r }' \
    | sort -s -t "$T" -k1,1 | tail -n 1 | cut -f2)"
  safety_verdict="${latest:-none}"
  [ "$safety_verdict" = pass ] \
    || refuse "PR #$PR touches a safety path ($safety_paths) and has no passing safety verdict at $SHA (latest: $safety_verdict)"
fi

# ── the merge ───────────────────────────────────────────────────────────────
summary="checks: open · author $author · head unmoved · approved by $JUDGE · $CHECKS green · safety paths: $safety_paths · safety verdict: $safety_verdict"
if [ "$APPLY" -eq 0 ]; then
  printf 'merge · dry run · nothing merged\npr: #%s · sha: %s · as: %s\n%s\n' "$PR" "$SHA" "$who" "$summary"
  printf 'would run: gh pr merge %s -R %s --merge --match-head-commit %s\n' "$PR" "$REPO" "$SHA"
  exit 0
fi

out="$(rd pr merge "$PR" -R "$REPO" --merge --match-head-commit "$SHA" 2>&1)" \
  || die "gh pr merge did not merge #$PR at $SHA: $(printf '%s' "$out" | tr -d '\r' | tr '\n' ' ' | cut -c1-300)"
after="$(rd pr view "$PR" -R "$REPO" --json state,mergeCommit --jq '[.state, (.mergeCommit.oid // "")] | @tsv' 2>/dev/null)" \
  || die "gh pr merge returned, but PR #$PR could not be read back; read it before anything else"
IFS="$T" read -r now commit <<< "$(printf '%s' "$after" | tr -d '\r')"
[ "$now" = MERGED ] || die "gh pr merge returned, but PR #$PR is ${now:-unknown}, not merged; read it before anything else"
printf 'merged: #%s at %s as %s · merge commit %s\n%s\n' "$PR" "$SHA" "$who" "${commit:-unknown}" "$summary"
exit 0
