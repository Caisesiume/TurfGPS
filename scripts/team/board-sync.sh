#!/usr/bin/env bash
# board-sync.sh — keeps the board true to GitHub, and promotes what is free (#210).
#
# It replaces the scrum-master: one deterministic pass, nothing inferred.
#   Done              an issue closed as completed (or by a plain close), a merged PR
#   In review         an open, non-draft PR closes the issue
#   Ordered Revision  ...and that PR carries judge:remanded
#   Backlog           an open item with no Status at all — auto-add leaves it empty
#   Ready             a Backlog item `scripts/loop/dependents.sh --all` lists as eligible:
#                     it declares blockers and every one is satisfied
# Flagged, never moved: an issue closed NOT_PLANNED or DUPLICATE outside Done, an open
# issue sitting in Done, and an item In review or Ordered Revision with no open PR.
# It never sets `In progress` — the lead does when it dispatches — and never promotes
# a story that declares no blocker: that is the lead's call, because nothing here can
# tell an unplanned story from a free one.
#
# DRY RUN BY DEFAULT: it prints the moves it would make and changes nothing. --apply
# makes them, one `gh project item-edit` each, under the default account: a board
# edit is bookkeeping, not a judgment. Field and option IDs are resolved fresh on
# every run, never cached — they were regenerated once already (2026-07-31).
#
# Usage: scripts/team/board-sync.sh [--apply]
# Exit:  0  in sync, or --apply made every move
#        10 dry run, and moves are pending
#        1  --apply, and a move failed
#        2  cannot run — GitHub or the board could not be read, or the read was truncated

set -uo pipefail

GH="${GH:-/c/Program Files/GitHub CLI/gh.exe}"
REPO="${TEAM_REPO:-Caisesiume/TurfGPS}"
OWNER="${REPO%%/*}"; PROJECT=3; LIMIT=1000
DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || DIR=''
APPLY=0
case "${1:-}" in
  '') ;; --apply) APPLY=1 ;;
  *) echo "board-sync: cannot run — the only argument is --apply" >&2; exit 2 ;;
esac
die() { printf 'board-sync: cannot run — %s\n' "$*" >&2; exit 2; }
g() { env -u GH_TOKEN -u GITHUB_TOKEN -u GH_JUDGE_TOKEN "$GH" "$@"; }

pid="$(g project view "$PROJECT" --owner "$OWNER" --format json --jq .id 2>/dev/null | tr -d '\r')"
[ -n "$pid" ] || die "project $PROJECT could not be read"
fields="$(g project field-list "$PROJECT" --owner "$OWNER" --format json \
  --jq '.fields[] | select(.name == "Status") | .id as $f | .options[] | [$f, .id, .name] | @tsv' 2>/dev/null | tr -d '\r')"
[ -n "$fields" ] || die "the Status field could not be resolved"
items="$(g project item-list "$PROJECT" --owner "$OWNER" --limit "$LIMIT" --format json \
  --jq '.items[] | [.id, (.status // ""), (.content.type // ""), (.content.number // ""), (.content.repository // "")] | @tsv' 2>/dev/null | tr -d '\r')" \
  || die "the board items could not be read"
[ -n "$items" ] || die "the board read returned no items"
[ "$(printf '%s\n' "$items" | wc -l)" -lt "$LIMIT" ] || die "the board read hit its limit of $LIMIT items; raise LIMIT"
closed="$(g issue list -R "$REPO" --state closed --limit "$LIMIT" --json number,stateReason \
  --jq '.[] | [.number, (.stateReason // "")] | @tsv' 2>/dev/null | tr -d '\r')" || die "closed issues could not be read"
merged="$(g pr list -R "$REPO" --state merged --limit "$LIMIT" --json number --jq '.[].number' 2>/dev/null | tr -d '\r')" \
  || die "merged PRs could not be read"
open_prs="$(g pr list -R "$REPO" --state open --limit 200 --json number,isDraft,labels,closingIssuesReferences \
  --jq '.[] | select(.isDraft | not) | [.number, ([.labels[].name] | join(",")), ([.closingIssuesReferences[].number | tostring] | join(","))] | @tsv' \
  2>/dev/null | tr -d '\r')" || die "open PRs could not be read"
[ -n "$DIR" ] && [ -f "$DIR/../loop/dependents.sh" ] || die "scripts/loop/dependents.sh is not where this script expects it"
deps="$(GH="$GH" bash "$DIR/../loop/dependents.sh" --all 2>/dev/null)" || die "dependents.sh --all could not read the board"
eligible="$(printf '%s\n' "$deps" | sed -n 's/^eligible: //p' | tr -d '#,' | tr ' ' '\n' | grep -E '^[0-9]+$')"

# One awk pass decides every item. Inputs are tagged so each set is read once.
plan="$( {
  printf '%s\n' "$closed"   | sed '/^$/d; s/^/C\t/'
  printf '%s\n' "$merged"   | sed '/^$/d; s/^/M\t/'
  printf '%s\n' "$open_prs" | sed '/^$/d; s/^/P\t/'
  printf '%s\n' "$eligible" | sed '/^$/d; s/^/E\t/'
  printf '%s\n' "$items"    | sed '/^$/d; s/^/I\t/'
} | awk -F'\t' -v repo="$REPO" '
  $1 == "C" { reason[$2] = $3; isclosed[$2] = 1; next }
  $1 == "M" { ismerged[$2] = 1; next }
  $1 == "P" { n = split($4, cl, ","); for (i = 1; i <= n; i++) if (cl[i] != "") {
                pr[cl[i]] = $2; remanded[cl[i]] = ($3 ~ /(^|,)judge:remanded(,|$)/) } ; next }
  $1 == "E" { free[$2] = 1; next }
  $1 == "I" {
    id = $2; st = $3; type = $4; num = $5; rp = $6
    if (num == "" || rp != repo) next
    to = ""; why = ""
    if (type == "PullRequest") {
      if ((num in ismerged) && st != "Done") { to = "Done"; why = "PR merged" }
    } else if (type == "Issue") {
      if (num in isclosed) {
        r = reason[num]
        if (r == "COMPLETED" || r == "") { if (st != "Done") { to = "Done"; why = "closed as completed" } }
        else if (st != "Done") { print "F\t#" num " is closed " r " but sits in " (st == "" ? "no status" : st) }
      } else if (st == "Done") {
        print "F\t#" num " is open but sits in Done"
      } else if (num in pr) {
        want = remanded[num] ? "Ordered Revision" : "In review"
        if (st != want) { to = want; why = "PR #" pr[num] (remanded[num] ? " judge:remanded" : " open") }
      } else if (st == "In review" || st == "Ordered Revision") {
        print "F\t#" num " sits in " st " with no open PR closing it"
      } else if (st == "") { to = "Backlog"; why = "had no status" }
      else if (st == "Backlog" && (num in free)) { to = "Ready"; why = "every declared blocker is done" }
    }
    if (to != "") print "V\t" id "\t" num "\t" (st == "" ? "-" : st) "\t" to "\t" why
  }')"

moves="$(printf '%s\n' "$plan" | sed -n 's/^V\t//p')"
flags="$(printf '%s\n' "$plan" | sed -n 's/^F\t//p')"
nmoves="$(printf '%s\n' "$moves" | sed '/^$/d' | wc -l | tr -d ' ')"
nflags="$(printf '%s\n' "$flags" | sed '/^$/d' | wc -l | tr -d ' ')"

applied=0; failedm=0
if [ "$APPLY" -eq 1 ] && [ "$nmoves" -gt 0 ]; then
  while IFS=$'\t' read -r id num from to why; do
    [ -n "$id" ] || continue
    line="$(printf '%s\n' "$fields" | awk -F'\t' -v n="$to" '$3 == n { print $1 "\t" $2; exit }')"
    fid="${line%%$'\t'*}"; oid="${line#*$'\t'}"
    if [ -n "$line" ] && g project item-edit --id "$id" --project-id "$pid" --field-id "$fid" \
         --single-select-option-id "$oid" >/dev/null 2>&1; then applied=$((applied + 1))
    else failedm=$((failedm + 1)); printf 'board-sync: could not move #%s to %s\n' "$num" "$to" >&2; fi
  done <<< "$moves"
fi

printf 'board_sync:\n  mode: %s\n  items: %s\n' "$([ "$APPLY" -eq 1 ] && echo apply || echo dry-run)" "$(printf '%s\n' "$items" | wc -l | tr -d ' ')"
if [ "$nmoves" -eq 0 ]; then printf '  moves: []\n'
else
  printf '  moves:\n'
  printf '%s\n' "$moves" | sed '/^$/d' | head -n 14 | awk -F'\t' '{ printf "    - {item: \"#%s\", from: %s, to: %s, why: \"%s\"}\n", $2, $3, $4, $5 }'
  [ "$nmoves" -le 14 ] || printf '    - … %s more\n' "$((nmoves - 14))"
fi
if [ "$nflags" -eq 0 ]; then printf '  flagged: []\n'
else
  printf '  flagged:\n'
  printf '%s\n' "$flags" | sed '/^$/d' | head -n 5 | sed 's/^/    - "/; s/$/"/'
  [ "$nflags" -le 5 ] || printf '    - … %s more\n' "$((nflags - 5))"
fi
if [ "$APPLY" -eq 1 ]; then
  printf '  applied: %s · failed: %s\n' "$applied" "$failedm"
  [ "$failedm" -eq 0 ] || exit 1
  exit 0
fi
[ "$nmoves" -eq 0 ] || exit 10
exit 0
