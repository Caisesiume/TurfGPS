#!/usr/bin/env bash
# dependents-declared-edges.sh — proves dependents.sh counts only DECLARED blockers,
# and resolves one that is a pull request (#138). The two defects it guards:
#   1. the reason prose on a `Blocked by:` line names other issues and pull requests,
#      and every one of them used to become an edge. A satisfied edge's line stays in
#      the body forever (`turfgps-board-ops § Satisfied is not removed`), so a phantom
#      blocked its story permanently — #136 was held on a pull request its own reason
#      merely mentioned, and prose accumulates cross-references as the corpus matures;
#   2. a declared blocker may BE a pull request — numbers share one sequence — and a
#      MERGED one had no case, so it read UNKNOWN and blocked forever.
# What must NOT change, and is asserted here too: a DECLARED blocker whose state
# genuinely cannot be read still counts as blocking (`ADR-0003 § P6`, `ADR-0003 § A3`).
#
# Hermetic: stub `gh` on PATH, fixture files as the only input, no repo state, no
# network. The stub returns what `gh --jq` would have returned, which is the same
# convention `fingerprint-isolation.sh` uses — standalone jq is not installed, so a
# stub cannot evaluate a jq program, and the logic under test is in the shell.
# Usage: scripts/loop/tests/dependents-declared-edges.sh    Exit: 0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$DIR/../dependents.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

export FIX="$TMP/fix"
mkdir -p "$FIX" "$TMP/bin"

# `gh issue list --json number,body --jq …` — one `<issue> <the whole Blocked by line>`
# record per hard-blocker line, in body order.
cat > "$FIX/list" <<'FIXTURE'
136 Blocked by: #37 — `DEPLOYMENT.md` is on `main` as of PR #67, and the residency half of this ruling is stated against the runtime model that document fixes.
201 Blocked by: #67 — the merged pull request itself, declared as the blocker
202 Blocked by: #7, #41
203 Blocked by: #142 — an open pull request, declared as the blocker
204 Blocked by: #143 — a pull request closed without merging, declared as the blocker
205 Blocked by: #999 — a declared blocker whose state cannot be read at all
206 Blocked by: #37 — the limit arrives with it; #41 is where that limit is stored and is still open
207 Blocked by: #7 · #41 — two declared blockers on this repo's own house separator
FIXTURE

# `gh issue view <n> --json state,stateReason,url --jq '.url + " " + .state + …'`.
# The trailing space is real: a pull request and an open issue carry no stateReason.
R=https://github.com/Caisesiume/TurfGPS
printf '%s/issues/37 CLOSED COMPLETED\n' "$R" > "$FIX/view.37"
printf '%s/issues/7 CLOSED COMPLETED\n'  "$R" > "$FIX/view.7"
printf '%s/issues/41 OPEN \n'            "$R" > "$FIX/view.41"
printf '%s/pull/67 MERGED \n'            "$R" > "$FIX/view.67"
printf '%s/pull/142 OPEN \n'             "$R" > "$FIX/view.142"
printf '%s/pull/143 CLOSED \n'           "$R" > "$FIX/view.143"
# no view.999 — the reference cannot be read at all

cat > "$TMP/bin/gh" <<'STUB'
#!/bin/sh
# real forms: `gh issue list … --jq …` · `gh issue view <n> … --jq …` · `gh auth status`
case "$1 $2" in
  "issue list") cat "$FIX/list" ;;
  "issue view") cat "$FIX/view.$3" 2>/dev/null || exit 1 ;;
esac
exit 0
STUB
chmod +x "$TMP/bin/gh"
export PATH="$TMP/bin:$PATH" GH="$TMP/bin/gh"

fails=0
OUT=""; RC=0
run() { OUT="$(bash "$SCRIPT" "$1" 2>&1)"; RC=$?; }
check() { # check <label> <expected-rc> <expected-substring>
  if [ "$RC" = "$2" ] && case "$OUT" in *"$3"*) true ;; *) false ;; esac; then
    printf 'PASS  %s\n' "$1"
  else
    printf 'FAIL  %s — expected rc %s + %s, got rc %s: %s\n' "$1" "$2" "$3" "$RC" "$(printf '%s' "$OUT" | tr '\n' '|')"
    fails=$((fails + 1))
  fi
}
absent() { # absent <label> <substring-that-must-not-appear>
  if case "$OUT" in *"$2"*) false ;; *) true ;; esac; then
    printf 'PASS  %s\n' "$1"
  else
    printf 'FAIL  %s — %s must not appear, got: %s\n' "$1" "$2" "$(printf '%s' "$OUT" | tr '\n' '|')"
    fails=$((fails + 1))
  fi
}

# Defect 1 — only the declared list is an edge. #136 and #206 each declare #37 alone;
# the prose behind the reason names a merged PR and an open issue, and neither gates.
run 37;  check  "prose refs are not edges -> both stories free"  0 "eligible: #136, #206"
run 41;  absent "an open issue named in prose is not an edge"      "#206"

# Defect 2 — a declared blocker that is a merged pull request is satisfied, not UNKNOWN.
run 67;  check  "declared blocker is a merged PR -> satisfied"   0 "eligible: #201"
run 67;  absent "the PR named in #136's prose gained no dependent" "#136"

# The declared list itself is kept whole, both members of it.
run 7;   check  "multi-blocker list: #7 done, #41 still blocks"  0 "#202 (blockers: #41 open)"

# The boundary is a rule about glue, not a list of five separators: `·` is in no
# whitelist and is this repo's house separator, so enumerating separators dropped
# #41 here and printed #207 as eligible with a declared blocker still open.
run 7;   check  "an unenumerated separator loses no declared blocker" 0 "#207 (blockers: #41 open)"

# The other two pull-request states block, and the closed one is named as not completed.
run 142; check  "an open pull request blocks"                    0 "#203 (blockers: #142 open)"
run 143; check  "a pull request closed unmerged blocks"          0 "#143(NOT_COMPLETED)"

# Unchanged: a DECLARED blocker that cannot be read fails toward blocked.
run 999; check  "unreadable declared blocker still blocks"       0 "#205 (blockers: #999 unknown)"
run 999; check  "and is named on its own line"                   0 "unreadable: #999"

[ "$fails" -eq 0 ] || { printf '\n%s test(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
