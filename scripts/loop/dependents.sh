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
# DECLARED BLOCKERS ONLY. A `Blocked by:` line carries its reason on the same line,
# and that prose names other issues and pull requests — `as of PR #67`, `which #89
# computes`. Only the leading `#N` list, before the reason begins, is an edge; the
# rest of the line is provenance. Scanning the whole line manufactured blockers
# nobody declared, and because a satisfied edge's line stays in the body forever
# (`turfgps-board-ops § Satisfied is not removed`) such a phantom blocked its story
# permanently — #136 was held on a pull request its own reason merely mentioned.
#
# SATISFIED MEANS SUCCESSFULLY COMPLETE, not merely closed (directive-4 §13). A
# blocker is satisfied iff it is CLOSED with stateReason COMPLETED — or with no
# stateReason at all, which is how a legacy plain close reads. Closed as
# NOT_PLANNED or DUPLICATE is work that never happened: it STILL BLOCKS and is
# printed with its reason, so @scrum-master files a `dependency_finding` instead
# of promoting a story onto dead work. Any other closed reason blocks too — the
# ambiguous case fails toward blocked, always.
#
# A REFERENCE MAY BE A PULL REQUEST. Issue and PR numbers share one sequence and
# `gh issue view` resolves both, so a declared blocker may be either. A MERGED pull
# request is satisfied — the work landed. One still open blocks; one closed unmerged
# never landed and blocks as `not_completed`. Which kind a reference is comes off the
# `url`, because `state` alone cannot tell a closed-unmerged pull request from the
# legacy plain-closed issue above: both are CLOSED with no stateReason.
#
# A blocker whose state cannot be read counts as STILL BLOCKING and is reported
# on its own line. An unreadable prerequisite must never be able to read as a
# satisfied one — the same rule fingerprint.sh applies to an unreadable component.
# That is unchanged; it now reaches only references that were actually declared.
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
# NOT installed on this machine. Emits one `<issue> <the whole Blocked by line>`
# record per hard-blocker line — the raw line, because which of the references on
# it are DECLARED is decided below, in the shell, where it can be tested against
# fixtures without a network.
lines="$("$GH" issue list --state open --limit 200 --json number,body --jq '
  .[]
  | select(.body != null)
  | {n: .number,
     dep: (.body | split("## Dependencies")
                 | if length > 1 then (.[1] | split("\n## ")[0]) else "" end)}
  | .n as $n
  | .dep | split("\n")[]
  | select(test("^ *\\**Blocked by"))
  | "\($n) \(.)"
' 2>/dev/null)"

# The declared blockers, then one `<issue> <blocker,blocker,...>` line per story —
# joined across a story's several `Blocked by:` lines, in the order the body gives.
# The declared list is the run of `#N` references and their separators that opens
# the line; the first token that is neither ends it, which is the reason delimiter
# in the grammar and the end of the line in a grandfathered one.
pairs="$(printf '%s\n' "$lines" | awk '
  { n = $1; rest = substr($0, index($0, " ") + 1); list = ""
    sub(/^ *[*]*Blocked by[^#]*/, "", rest)   # the label; [^#] stops at the first ref
    while (match(rest, /^#[0-9]+/)) {
      list = (list == "" ? "" : list ",") substr(rest, 2, RLENGTH - 1)
      rest = substr(rest, RLENGTH + 1)
      if (!sub(/^([ ,;&+]|and)+/, "", rest)) break
    }
    if (list == "") next
    if (!(n in b)) order[++k] = n
    b[n] = (n in b) ? b[n] "," list : list }
  END { for (i = 1; i <= k; i++) print order[i], b[order[i]] }')"

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

seen=""            # memo: |<issue>=<VERDICT>| — one API call per distinct blocker
# DONE is the only satisfying verdict. Everything else blocks.
state_of() {
  case "$seen" in
    *"|$1=OPEN|"*)          echo OPEN ;          return ;;
    *"|$1=DONE|"*)          echo DONE ;          return ;;
    *"|$1=NOT_PLANNED|"*)   echo NOT_PLANNED ;   return ;;
    *"|$1=DUPLICATE|"*)     echo DUPLICATE ;     return ;;
    *"|$1=NOT_COMPLETED|"*) echo NOT_COMPLETED ; return ;;
    *"|$1=UNKNOWN|"*)       echo UNKNOWN ;       return ;;
  esac
  raw="$("$GH" issue view "$1" --json state,stateReason,url \
         --jq '.url + " " + .state + " " + (.stateReason // "")' 2>/dev/null)"
  case "$raw"  in *' '*) u="${raw%% *}";   rest="${raw#* }" ;; *) u="";        rest="" ;; esac
  case "$rest" in *' '*) st="${rest%% *}"; rs="${rest#* }"  ;; *) st="$rest";  rs="" ;; esac
  case "$u" in */pull/*) k=pr ;; */issues/*) k=issue ;; *) k="" ;; esac
  case "$k|$st|$rs" in
    issue\|OPEN\|*)                           s=OPEN ;;
    issue\|CLOSED\|COMPLETED|issue\|CLOSED\|) s=DONE ;;          # completed, or a legacy plain close
    issue\|CLOSED\|NOT_PLANNED)               s=NOT_PLANNED ;;
    issue\|CLOSED\|DUPLICATE)                 s=DUPLICATE ;;
    issue\|CLOSED\|*)                         s=NOT_COMPLETED ;; # an unrecognized reason blocks too
    pr\|MERGED\|*)                            s=DONE ;;          # the work landed; not a blocker
    pr\|OPEN\|*)                              s=OPEN ;;
    pr\|CLOSED\|*)                            s=NOT_COMPLETED ;; # closed unmerged: it never landed
    *)                                        s=UNKNOWN ;;
  esac
  seen="$seen|$1=$s|"
  echo "$s"
}

eligible=""; blocked_lines=""; unreadable=""; not_completed=""
while IFS=' ' read -r issue blockers; do
  [ -n "$issue" ] || continue
  remaining=""
  for b in $(printf '%s' "$blockers" | tr ',' ' '); do
    st="$(state_of "$b")"
    [ "$st" = "DONE" ] && continue
    [ -n "$remaining" ] && remaining="$remaining, "
    remaining="$remaining#$b $(printf '%s' "$st" | tr 'A-Z' 'a-z')"
    case "$st" in
      UNKNOWN) unreadable="$unreadable #$b" ;;
      OPEN)    : ;;
      *)       not_completed="$not_completed #$b($st)" ;;
    esac
  done
  if [ -z "$remaining" ]; then
    eligible="$eligible #$issue"
  else
    blocked_lines="$blocked_lines
  #$issue (blockers: $remaining)"
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
[ -n "$not_completed" ] && printf 'not_completed:%s — closed without completing; still blocking, file a dependency_finding\n' "$not_completed"
[ -z "$unreadable" ] && exit 0
printf 'unreadable:%s — counted as still blocking\n' "$unreadable"
exit 0
