#!/usr/bin/env bash
# post-review.sh — the one way a judgment reaches GitHub (#210; rulings #202, #206).
#
#   safety  posts safety-sentinel's verdict at one head, as a PR comment
#   ruling  posts the critic's ruling comment, then the review pinned to the
#           same head, then the judge:* label
#
# Both go out as TheReviewNinja. THIS IS THE ONLY FILE THAT USES GH_JUDGE_TOKEN,
# and it uses it by name only: the value reaches gh through the environment of
# the one command that writes, and is never read into a variable, printed,
# logged or written to a file. Every heredoc is quoted, and tracing is switched
# off below, so nothing here can expand it into output.
#
# REFUSED, exit 1, before anything is written:
#   - a moved head: the PR's head is not --sha, the SHA the verdict was formed on
#   - a PR that is not open
#   - approve while the PR carries `awaiting-human`: the Owner is deciding
#   - approve on a safety path without a passing safety verdict at --sha. Safety
#     path = scripts/loop/diff-domains.sh lists one of the PR's files, read from
#     GitHub's own file list. Passing verdict = the latest comment by
#     TheReviewNinja whose first line is this script's safety marker for --sha
#     reads result=pass. A verdict at another SHA, or by another login, is none.
#   - a live run whose token is not TheReviewNinja's
# IT NEVER MERGES. There is no merge verb or event, and nothing here calls one.
#
# --dry-run uses no token at all: every gh it runs has GH_TOKEN, GITHUB_TOKEN
# and GH_JUDGE_TOKEN removed from its environment, so reads go out under the
# default account. It runs every check and prints what it would post.
# Live, a missing or failing token is a stop-and-report (exit 2). There is no
# fallback to the default account: that account is the author's.
#
# Usage:
#   post-review.sh safety --pr N --sha SHA --result pass|block --body FILE [--dry-run]
#   post-review.sh ruling --pr N --sha SHA --event approve|request-changes|comment --body FILE [--dry-run]
# The body is the poster's envelope, at most 25 lines (team-protocol § The envelope).
# Exit: 0 posted (dry run: would post) · 1 refused · 2 cannot run
#
# Tested in dry run by scripts/team/tests/post-review-refusals.sh, with gh stubbed.

set -uo pipefail
set +x

REPO="${TEAM_REPO:-Caisesiume/TurfGPS}"
GH="${GH:-/c/Program Files/GitHub CLI/gh.exe}"
DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || DIR=''
DOMAINS="$DIR/../loop/diff-domains.sh"
JUDGE='TheReviewNinja'
MAX_LINES=25
SIGNATURE='/ The Review Ninja'

die()    { printf 'post-review: cannot run — %s\n' "$*" >&2; exit 2; }
refuse() { printf 'post-review: refused — %s\n' "$*" >&2; exit 1; }

# Every read runs under the default account, with no token in its environment.
rd() { env -u GH_TOKEN -u GITHUB_TOKEN -u GH_JUDGE_TOKEN "$GH" "$@"; }

# ── arguments ────────────────────────────────────────────────────────────────
VERB="${1:-}"
case "$VERB" in
  safety|ruling) shift ;;
  *) die "the first argument is 'safety' or 'ruling' — there is no other verb" ;;
esac
PR=''; SHA=''; RESULT=''; EVENT=''; BODY=''; DRY=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run) DRY=1; shift; continue ;;
    --pr|--sha|--result|--event|--body) [ "$#" -ge 2 ] || die "$1 needs a value" ;;
    *) die "unknown argument: $1" ;;
  esac
  case "$1" in
    --pr) PR="$2" ;; --sha) SHA="$2" ;; --result) RESULT="$2" ;;
    --event) EVENT="$2" ;; --body) BODY="$2" ;;
  esac
  shift 2
done

[[ "$PR" =~ ^[0-9]+$ ]] || die "--pr must be a PR number"
[[ "$SHA" =~ ^[0-9a-f]{40}$ ]] || die "--sha must be the full 40-character head SHA the verdict was formed on"
if [ "$VERB" = safety ]; then
  [ -z "$EVENT" ] || die "safety takes --result, not --event"
  case "$RESULT" in pass|block) ;; *) die "--result is pass or block" ;; esac
else
  [ -z "$RESULT" ] || die "ruling takes --event, not --result"
  case "$EVENT" in
    approve) REVIEW_EVENT=APPROVE ;;
    request-changes) REVIEW_EVENT=REQUEST_CHANGES ;;
    comment) REVIEW_EVENT=COMMENT ;;
    *) die "--event is approve, request-changes or comment — nothing else, and never a merge" ;;
  esac
fi

# ── the body ────────────────────────────────────────────────────────────────
[ -n "$BODY" ] && [ -f "$BODY" ] && [ -r "$BODY" ] || die "--body must name a readable file"
lines="$(tr -d '\r' < "$BODY" | awk 'END { print NR }')"
[ "${lines:-0}" -gt 0 ] || die "the body is empty"
[ "$lines" -le "$MAX_LINES" ] || refuse "the body is $lines lines; the envelope cap is $MAX_LINES (team-protocol § The envelope)"
if grep -qE '^```|<!-- team:' "$BODY"; then
  refuse "the body may not carry a code fence or a team: marker; this script writes both"
fi

# ── the PR as it is now ─────────────────────────────────────────────────────
pr_line="$(rd pr view "$PR" -R "$REPO" --json state,headRefOid,labels \
  --jq '[.state, .headRefOid, ([.labels[].name] | join(","))] | @tsv' 2>/dev/null)" \
  || die "could not read PR #$PR"
IFS=$'\t' read -r state head labels <<< "$(printf '%s' "$pr_line" | tr -d '\r')"
[ -n "${head:-}" ] || die "could not read PR #$PR's head"
[ "$state" = OPEN ] || refuse "PR #$PR is $state, not open"
[ "$head" = "$SHA" ] || refuse "the head moved: PR #$PR is at $head, the verdict was formed on $SHA"

safety_paths='none'; safety_verdict='not needed'
if [ "$VERB" = ruling ] && [ "$EVENT" = approve ]; then
  case ",$labels," in
    *,awaiting-human,*) refuse "PR #$PR is labelled awaiting-human; the Owner decides before any approval" ;;
  esac

  files="$(rd api "repos/$REPO/pulls/$PR/files" --paginate \
    --jq '.[] | .filename, (.previous_filename // empty)' 2>/dev/null)" \
    || die "could not read PR #$PR's files, so whether it touches a safety path is unknown"
  [ -n "$files" ] || die "PR #$PR lists no files, so whether it touches a safety path is unknown"
  [ -n "$DIR" ] && [ -f "$DOMAINS" ] || die "scripts/loop/diff-domains.sh is not where this script expects it"
  domains="$(printf '%s\n' "$files" | tr -d '\r' | bash "$DOMAINS" -)" \
    || die "diff-domains.sh could not classify PR #$PR"
  safety_paths="$(printf '%s\n' "$domains" | sed -n 's/^safety_path_candidates: \[\(.*\)\]$/\1/p')"
  grep -q '^safety_path_candidates: ' <<< "$domains" \
    || die "diff-domains.sh printed no safety_path_candidates line"
  [ -n "$safety_paths" ] || safety_paths='none'

  if [ "$safety_paths" != none ]; then
    comments="$(rd api "repos/$REPO/issues/$PR/comments" --paginate \
      --jq '.[] | [.user.login, .created_at, (.body | split("\n") | .[0] | rtrimstr("\r"))] | @tsv' 2>/dev/null)" \
      || die "could not read PR #$PR's comments to find a safety verdict"
    want="<!-- team:safety v1 pr=$PR sha=$SHA result="
    latest="$(printf '%s\n' "$comments" | tr -d '\r' | awk -F'\t' -v judge="$JUDGE" -v want="$want" '
      $1 == judge && ($3 == (want "pass -->") || $3 == (want "block -->")) {
        r = substr($3, length(want) + 1); sub(/ -->$/, "", r); print $2 "\t" r }' \
      | sort -s -t "$(printf '\t')" -k1,1 | tail -n 1 | cut -f2)"
    safety_verdict="${latest:-none}"
    [ "$safety_verdict" = pass ] \
      || refuse "PR #$PR touches a safety path ($safety_paths) and has no passing safety verdict at $SHA (latest: $safety_verdict)"
  fi
fi

# ── what would be posted ────────────────────────────────────────────────────
TMP="$(mktemp -d 2>/dev/null)" && [ -d "$TMP" ] || die "mktemp -d gave no directory"
trap 'rm -rf "$TMP"' EXIT
if [ "$VERB" = safety ]; then
  marker="<!-- team:safety v1 pr=$PR sha=$SHA result=$RESULT -->"
else
  marker="<!-- team:ruling v1 pr=$PR sha=$SHA event=$EVENT -->"
fi
{
  printf '%s\n```yaml\n' "$marker"
  tr -d '\r' < "$BODY" | sed -e '$a\'
  printf '```\n%s\n' "$SIGNATURE"
} > "$TMP/comment.md"

add=''; remove=''
if [ "$VERB" = ruling ]; then
  case "$EVENT" in
    approve)         add='judge:approved'; remove='judge:remanded' ;;
    request-changes) add='judge:remanded'; remove='judge:approved' ;;
    comment)         remove='judge:approved' ;;
  esac
fi

if [ "$DRY" -eq 1 ]; then
  printf 'post-review · dry run · nothing posted · no token used\n'
  printf 'pr: #%s · sha: %s · verb: %s · %s\n' "$PR" "$SHA" "$VERB" "${EVENT:-result $RESULT}"
  printf 'checks: open · head unmoved · safety paths: %s · safety verdict: %s\n' "$safety_paths" "$safety_verdict"
  printf -- '--- comment, as %s ---\n' "$JUDGE"
  cat "$TMP/comment.md"
  if [ "$VERB" = ruling ]; then
    printf -- '--- review, as %s: %s, pinned to %s ---\n' "$JUDGE" "$REVIEW_EVENT" "$SHA"
    printf 'Ruling at %s: %s. The findings are in the ruling comment.\n%s\n' "${SHA:0:7}" "$EVENT" "$SIGNATURE"
    printf -- '--- labels, default account ---\nadd: %s · remove: %s\n' "${add:-none}" "${remove:-none}"
  fi
  exit 0
fi

# ── live: the token, then the writes ────────────────────────────────────────
[ -n "${GH_JUDGE_TOKEN:-}" ] || die "GH_JUDGE_TOKEN is not set; stop and report — never post a judgment as the default account"
who="$(GH_TOKEN="$GH_JUDGE_TOKEN" "$GH" api user --jq .login 2>/dev/null)" \
  || die "GH_JUDGE_TOKEN did not authenticate; stop and report"
who="$(printf '%s' "$who" | tr -d '\r')"
[ "$who" = "$JUDGE" ] || refuse "GH_JUDGE_TOKEN belongs to '$who', not $JUDGE"

posted="$(GH_TOKEN="$GH_JUDGE_TOKEN" "$GH" api -X POST "repos/$REPO/issues/$PR/comments" \
  -F "body=@$TMP/comment.md" --jq '[.user.login, .html_url] | @tsv' 2>/dev/null)" \
  || die "the comment was not posted"
IFS=$'\t' read -r by url <<< "$(printf '%s' "$posted" | tr -d '\r')"
[ "$by" = "$JUDGE" ] || die "the comment posted as '$by', not $JUDGE: stop and report it; do not delete or re-post it ($url)"
printf 'posted: %s comment %s\n' "$VERB" "$url"
[ "$VERB" = ruling ] || exit 0

now="$(rd pr view "$PR" -R "$REPO" --json headRefOid --jq .headRefOid 2>/dev/null | tr -d '\r')"
[ "$now" = "$SHA" ] || refuse "the head moved to ${now:-unknown} after the ruling comment; no review or label was posted"

printf 'Ruling at %s: %s. The findings are in %s\n\n%s\n' "${SHA:0:7}" "$EVENT" "$url" "$SIGNATURE" > "$TMP/review.md"
review="$(GH_TOKEN="$GH_JUDGE_TOKEN" "$GH" api -X POST "repos/$REPO/pulls/$PR/reviews" \
  -f "event=$REVIEW_EVENT" -f "commit_id=$SHA" -F "body=@$TMP/review.md" \
  --jq '[.user.login, .state, .html_url] | @tsv' 2>/dev/null)" \
  || die "the ruling comment is posted but the review is not: stop and report ($url)"
IFS=$'\t' read -r rby rstate rurl <<< "$(printf '%s' "$review" | tr -d '\r')"
[ "$rby" = "$JUDGE" ] || die "the review posted as '$rby', not $JUDGE: stop and report it ($rurl)"
printf 'posted: review %s %s\n' "$rstate" "$rurl"

label_args=()
[ -n "$add" ] && label_args+=(--add-label "$add")
[ -n "$remove" ] && label_args+=(--remove-label "$remove")
rd pr edit "$PR" -R "$REPO" "${label_args[@]}" >/dev/null 2>&1 \
  || die "the review is posted but the labels were not set: set them by hand (${add:+add $add} ${remove:+remove $remove})"
printf 'labels: %s\n' "${add:++$add }${remove:+-$remove}"
exit 0
