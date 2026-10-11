#!/usr/bin/env bash
# size-ratchet-cap.sh — proves size-ratchet.sh's total cap (#210) fires, in both directions.
#
# The per-file rows were already proved by hand on #207 (one file grown by a byte fails,
# one shrunk passes). What #210 added is the `(total)` row, and the case it exists for is
# the one per-file rows cannot see: every file within its own row while a NEW file, with
# a row of its own, pushes the set past the cap. That case, its control, a missing cap row
# and a doubled one are asserted here.
#
# Hermetic: a copy of the script in a throwaway tree, fixture files only.
# Usage: scripts/gates/tests/size-ratchet-cap.sh    Exit: 0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
TMP="$(mktemp -d)" || TMP=''
[ -n "$TMP" ] && [ -d "$TMP" ] || { echo 'FAIL  mktemp -d gave no directory'; exit 1; }
trap 'rm -rf "$TMP"' EXIT
mkdir -p "$TMP/scripts/gates" "$TMP/.claude/agents" "$TMP/.claude/skills/s"
cp "$DIR/../size-ratchet.sh" "$TMP/scripts/gates/size-ratchet.sh"
head -c 100 /dev/zero | tr '\0' 'a' > "$TMP/.claude/agents/a.md"
head -c 50  /dev/zero | tr '\0' 'b' > "$TMP/.claude/skills/s/SKILL.md"

fails=0
case_() { # case_ <label> <expected-rc> <expected-substring> <baseline rows...>
  local label="$1" want="$2" sub="$3" out rc; shift 3
  printf '%s\n' "$@" > "$TMP/scripts/gates/size-ratchet.baseline"
  out="$(bash "$TMP/scripts/gates/size-ratchet.sh" 2>&1)"; rc=$?
  if [ "$rc" = "$want" ] && case "$out" in *"$sub"*) true ;; *) false ;; esac; then
    printf 'PASS  %s\n' "$label"
  else
    printf 'FAIL  %s — expected rc %s + %s, got rc %s: %s\n' "$label" "$want" "$sub" "$rc" "$(printf '%s' "$out" | tr '\n' '|')"
    fails=$((fails + 1))
  fi
}
A='100 .claude/agents/a.md fixture'; S='50 .claude/skills/s/SKILL.md fixture'

case_ "within every row and the cap: clean"       0 "within its cap"   '150 (total) cap' "$A" "$S"
case_ "the cap exactly met: clean"                0 "cap 150"          '150 (total) cap' "$A" "$S"
case_ "set over the cap: fails, with its numbers" 1 "over cap · the set · 150 > 149 (+1)" '149 (total) cap' "$A" "$S"
case_ "no (total) row: cannot run"                2 "no (total) row"   "$A" "$S"
case_ "two (total) rows: cannot run"              2 "a second time"    '150 (total) cap' '150 (total) again' "$A" "$S"

# The case the cap exists for: a new file arrives WITH its own row, so no per-file row
# fails, and only the total can see the set grow.
head -c 10 /dev/zero | tr '\0' 'c' > "$TMP/.claude/agents/new.md"
case_ "a new file with its own row, set over the cap: fails" 1 "over cap · the set · 160 > 150 (+10)" \
      '150 (total) cap' "$A" "$S" '10 .claude/agents/new.md fixture'
case_ "...and with the cap raised in the baseline: clean"    0 "within its cap" \
      '160 (total) cap raised, visibly' "$A" "$S" '10 .claude/agents/new.md fixture'

[ "$fails" -eq 0 ] || { printf '\n%s test(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
