#!/usr/bin/env bash
# dependents.sh — who was waiting on issue #N, and who is free now (ADR-0003 § P6, directive-3 §22).
#
# Runs when a story merges or closes. It answers deterministically the question
# that would otherwise wake an LLM to re-read the whole backlog: which open
# stories name #N as a HARD blocker, and which of those have no open blocker
# left. @scrum-master evaluates the `eligible:` list for Ready;
# @backlog-dependency-planner runs this before any subgraph pass. Neither
# reasons about the graph to obtain a fact grep already holds.
#
# Hard blockers only. `Blocked by:` lines inside the issue body's `## Dependencies`
# section are the graph — the grammar is in `turfgps-board-ops § The dependency
# representation`. `Soft dependency:` lines are deliberately NOT read here: a soft
# edge never blocks readiness, so counting one would manufacture a blocker and
# hold work the board never intended to hold.
#
# A blocker whose state cannot be read counts as STILL BLOCKING and is reported
# on its own line. An unreadable prerequisite must never be able to read as a
# satisfied one — the same rule fingerprint.sh applies to an unreadable component.
#
# Usage: scripts/loop/dependents.sh <issue-number>
# Exit:  0  read the board (prints `none` when nothing depended on #N)
#        2  no argument, or GitHub could not be read at all

set -u

N="${1:-}"
case "$N" in
  ''|*[!0-9]*) echo "usage: dependents.sh <issue-number>" >&2; exit 2 ;;
esac

GH="${GH:-/c/Program Files/GitHub CLI/gh.exe}"

# One board read. `gh --jq` uses the jq engine compiled into gh; standalone jq is
# NOT installed on this machine. Emits one `<issue> <blocker,blocker,...>` line
# per open story carrying at least one hard blocker.
pairs="$("$GH" issue list --state open --limit 200 --json number,body --jq '
  .[]
  | select(.body != null)
  | {n: .number,
     dep: (.body | split("## Dependencies")
                 | if length > 1 then (.[1] | split("\n## ")[0]) else "" end)}
  | {n: .n,
     b: [ .dep | split("\n")[]
        | select(test("^ *\\**Blocked by"))
        | scan("#[0-9]+") | ltrimstr("#") ]}
  | select(.b | length > 0)
  | "\(.n) \(.b | join(","))"
' 2>/dev/null)"

if [ -z "$pairs" ]; then
  "$GH" auth status >/dev/null 2>&1 || { echo "error: could not read GitHub" >&2; exit 2; }
  printf 'dependents_of: #%s\nnone\n' "$N"
  exit 0
fi

# Exact membership, not a substring match: #4 must never match #41, and comparing
# parsed numbers is stronger than any word-boundary pattern over the raw line.
deps="$(printf '%s\n' "$pairs" | awk -v n="$N" '
  { c = split($2, a, ","); for (i = 1; i <= c; i++) if (a[i] == n) { print; next } }')"

if [ -z "$deps" ]; then
  printf 'dependents_of: #%s\nnone\n' "$N"
  exit 0
fi

seen=""            # memo: |<issue>=<STATE>| — one API call per distinct blocker
state_of() {
  case "$seen" in
    *"|$1=OPEN|"*)    echo OPEN ;    return ;;
    *"|$1=CLOSED|"*)  echo CLOSED ;  return ;;
    *"|$1=UNKNOWN|"*) echo UNKNOWN ; return ;;
  esac
  s="$("$GH" issue view "$1" --json state --jq .state 2>/dev/null)"
  case "$s" in OPEN|CLOSED) : ;; *) s=UNKNOWN ;; esac
  seen="$seen|$1=$s|"
  echo "$s"
}

eligible=""; blocked_lines=""; unreadable=""
while IFS=' ' read -r issue blockers; do
  [ -n "$issue" ] || continue
  remaining=""
  for b in $(printf '%s' "$blockers" | tr ',' ' '); do
    st="$(state_of "$b")"
    [ "$st" = "CLOSED" ] && continue
    remaining="$remaining #$b"
    [ "$st" = "UNKNOWN" ] && unreadable="$unreadable #$b"
  done
  if [ -z "$remaining" ]; then
    eligible="$eligible #$issue"
  else
    blocked_lines="$blocked_lines
  #$issue (open: $(printf '%s' "$remaining" | sed 's/^ //; s/ /, /g'))"
  fi
done <<EOF
$deps
EOF

printf 'dependents_of: #%s\n' "$N"
if [ -z "$eligible" ]; then
  printf 'eligible: none\n'
else
  printf 'eligible: %s\n' "$(printf '%s' "$eligible" | sed 's/^ //; s/ /, /g')"
fi

# Capped so a wide fan-out cannot flood a context window; the full picture is the board.
if [ -z "$blocked_lines" ]; then
  printf 'still_blocked: none\n'
else
  total="$(printf '%s' "$blocked_lines" | grep -c '#' 2>/dev/null || echo 0)"
  printf 'still_blocked:\n'
  printf '%s\n' "$blocked_lines" | sed '/^$/d' | head -12
  [ "$total" -gt 12 ] && printf '  … %s more — read the board\n' "$((total - 12))"
fi
[ -z "$unreadable" ] && exit 0
printf 'unreadable:%s — counted as still blocking\n' "$unreadable"
exit 0
