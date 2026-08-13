#!/usr/bin/env bash
# fingerprint-isolation.sh — proves fingerprint.sh's per-consumer state actually isolates.
# The consume bug this guards: one consumer's CHANGED must never be spent on another's
# behalf — the rule and its two halves are in `agent-handoffs § The trigger block`.
# Asserted mechanically because "it looks per-consumer" is what a shared-state version
# also looks like right up until two agents share a wake.
#
# Hermetic: stub `gh` and `git` on PATH, fixture files as the only source of change, a
# throwaway root so no real state file is touched. No jq, no network, no repo state.
# Usage: scripts/loop/tests/fingerprint-isolation.sh    Exit: 0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$DIR/../fingerprint.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

export FIX="$TMP/fix" FAKE_ROOT="$TMP/root"
mkdir -p "$FIX" "$FAKE_ROOT" "$TMP/bin"
printf '[{"number":1,"headRefOid":"aaa","isDraft":false}]\n' > "$FIX/pr"
printf '[{"id":"I_1","status":"Ready"}]\n'                    > "$FIX/board"
printf '1111111111111111111111111111111111111111\n'          > "$FIX/main"
printf '2222222222222222222222222222222222222222\n'          > "$FIX/corpus"

cat > "$TMP/bin/gh" <<'STUB'
#!/bin/sh
case "$1" in pr) cat "$FIX/pr" ;; project) cat "$FIX/board" ;; esac
exit 0
STUB
cat > "$TMP/bin/git" <<'STUB'
#!/bin/sh
# real form 1: `git rev-parse --show-toplevel`; forms 2-4 all carry `-C <root>` first
[ "$1" = "rev-parse" ] && { echo "$FAKE_ROOT"; exit 0; }
case "$3" in
  fetch)     exit 0 ;;
  rev-parse) cat "$FIX/main" ;;
  log)       cat "$FIX/corpus" ;;
esac
exit 0
STUB
chmod +x "$TMP/bin/gh" "$TMP/bin/git"
export PATH="$TMP/bin:$PATH" GH="$TMP/bin/gh"

fails=0
run() { OUT="$(bash "$SCRIPT" "$1" 2>&1)"; RC=$?; }
check() { # check <label> <expected-rc> <expected-substring>
  if [ "$RC" = "$2" ] && case "$OUT" in *"$3"*) true ;; *) false ;; esac; then
    printf 'PASS  %s\n' "$1"
  else
    printf 'FAIL  %s — expected rc %s + %s, got rc %s: %s\n' "$1" "$2" "$3" "$RC" "$(printf '%s' "$OUT" | tr '\n' '|')"
    fails=$((fails + 1))
  fi
}

run A; check "consumer A, first run -> CHANGED"            10 "CHANGED"
run A; check "consumer A, rerun -> UNCHANGED"               0 "UNCHANGED"
run B; check "consumer B unaffected by A -> CHANGED"       10 "CHANGED"
run B; check "consumer B, rerun -> UNCHANGED"               0 "UNCHANGED"

# A component moving must wake every consumer that has not yet seen it.
printf '[{"id":"I_1","status":"In progress"}]\n' > "$FIX/board"
run A; check "board mutated -> A CHANGED, names board"     10 "board:"
run B; check "board mutated -> B CHANGED independently"    10 "board:"

# An unreadable source records itself as unavailable, never as a quiet loop.
: > "$FIX/pr"
run A; check "unreadable pr -> DEGRADED"                    2 "degraded: pr"

[ "$fails" -eq 0 ] || { printf '\n%s test(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
