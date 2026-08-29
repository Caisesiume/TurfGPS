#!/usr/bin/env bash
# claim-table-known-defects.sh — assertions claim.sh does NOT currently satisfy.
#
# THIS FILE IS RED ON PURPOSE. It is not a broken test; it is a defect found by
# falsifying scripts/loop/claim.sh and kept in the failing state rather than bent
# until it passed, per the standing rule that a test bent until it passes is the
# instrument class this repository has already shipped too many of.
#
# It asserts the behaviour `#144 § Acceptance` requires. It turns green when the
# defect is fixed in the author lane, at which point it should be folded into
# claim-table.sh and this file deleted.
#
# THE DEFECT — one commit is not one epoch
#
# #144 puts the head SHA on each row as the epoch, so that "a `claimed` row is
# visible; a second judge sees the panel convened and stops". The row key is used
# verbatim after a case fold and a 7-40 hex length check, so the SAME commit
# spelled two legitimate ways produces TWO panels that cannot see each other:
#
#   git log --oneline   ->  ec9ee33                                    (7 hex)
#   git rev-parse HEAD  ->  ec9ee330b2cfe0f9164eaa7f3dee22c23c4afdc3   (40 hex)
#
# Both pass validation. Neither is exotic — they are the two ordinary ways an
# agent obtains the head SHA. The consequence, measured on 2026-08-29:
#
#   * two judges on one commit and one lane are BOTH granted a claim, and each
#     one's `status` shows a single-lane panel with no sign of the other
#     — failure class 2, duplicate dispatch, which this table exists to close;
#   * a verdict recorded at one spelling is filed `unclaimed` against the other
#     spelling's row and exits 12, while the claiming judge's panel stays
#     `outstanding` at rc 10 forever, and a second panel for the same commit
#     reads `complete` at rc 0 — failure classes 1 and 6, stranding and silent
#     debt, reopened through the mechanism meant to close them.
#
# The same root — an uncanonicalised row key — also splits `pr-0144` from
# `pr-144`. That spelling is far less likely from an agent, and it is included
# here because one fix closes both.
#
# Not fixed here: this suite's author does not write claim.sh. Whether the fix is
# to canonicalise the epoch, to require a 40-hex SHA, or to resolve short to long
# through git is a design decision for the author lane.
#
# Usage: scripts/loop/tests/claim-table-known-defects.sh
# Exit: 1 while the defect stands · 0 once it is fixed

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="${CLAIM_SH:-$DIR/../claim.sh}"
[ -r "$SCRIPT" ] || { printf 'cannot read the script under test: %s\n' "$SCRIPT" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

FULL=ec9ee330b2cfe0f9164eaa7f3dee22c23c4afdc3
SHORT=ec9ee33
PR=144

fails=0; cases=0; CASE=0
fresh() { CASE=$((CASE + 1)); CLAIM_TABLE_DIR="$TMP/tbl$CASE"; export CLAIM_TABLE_DIR; }
run() { OUT="$(bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }
flat() { printf '%s' "${1:-}" | tr '\n' '|'; }
has() { case "$1" in *"$2"*) return 0 ;; *) return 1 ;; esac; }
pass() { cases=$((cases + 1)); printf 'PASS  %s\n' "$1"; }
fail() { cases=$((cases + 1)); fails=$((fails + 1)); printf 'FAIL  %s\n' "$1"; }
check() {
  if [ "$RC" = "$2" ] && has "$OUT" "$3"; then pass "$1"
  else fail "$1 — want rc $2 with [$3], got rc $RC: $(flat "$OUT")"; fi
}
is() { if [ "$2" = "$3" ]; then pass "$1"; else fail "$1 — got [$2], want [$3]"; fi; }

printf -- '-- one commit is one epoch, however its SHA is spelled\n'

# The headline: two judges, one commit, one lane, both granted.
fresh
run claim $PR $FULL go-architecture-critic --owner judge-A
check 'the first judge claims the lane at the full SHA'      0  'claim: granted'
run claim $PR $SHORT go-architecture-critic --owner judge-B
check 'a second judge spelling the SAME commit short is REFUSED' 10 'reason: held'
is '  and the commit holds ONE panel, not two' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR" 2>/dev/null | wc -l | tr -d ' ')" 1

# Claim short, spell long: the same hole in the other direction.
fresh
run claim $PR $SHORT correctness --owner judge-A
check 'the first judge claims the lane at the short SHA'     0  'claim: granted'
run claim $PR $FULL correctness --owner judge-B
check 'a second judge spelling the SAME commit long is REFUSED'  10 'reason: held'

# The stranding consequence, which is what makes this more than untidiness.
fresh
run claim $PR $SHORT correctness --owner judge
run verdict $PR $FULL correctness approved --conf 0.9
check 'a verdict at the other spelling covers the claim that exists' 0 'verdict: recorded'
VOUT=$OUT
if has "$VOUT" 'unclaimed'; then
  fail 'the verdict is not filed as unclaimed — a claim for this commit exists'
else
  pass 'the verdict is not filed as unclaimed'
fi
run status $PR $SHORT
check 'the claiming judge sees its lane ruled and the panel complete' 0 'complete: true'

# One fix closes the PR spelling too.
fresh
run claim 0144 abc1234 lane-a --owner judge-A
run claim 144 abc1234 lane-a --owner judge-B
check 'a leading-zero PR number is the same panel'           10 'reason: held'

printf '\n%s checks · %s failed\n' "$cases" "$fails"
if [ "$fails" -eq 0 ]; then
  printf 'the defect is fixed — fold these into claim-table.sh and delete this file\n'
  exit 0
fi
printf 'the defect stands: one commit still yields more than one panel\n'
exit 1
