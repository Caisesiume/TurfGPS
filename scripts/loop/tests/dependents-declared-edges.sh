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
# genuinely cannot be read still counts as blocking, every closed-but-not-completed
# reason still blocks, and a legacy plain close still satisfies (`§ P6`, `§ A3`).
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
208 Blocked by: #41 — the first of this story's two lines, an open issue
208 Blocked by: #142 — the second line, an open pull request
209 Blocked by: #45 — a blocker closed as not planned: work that never happened
210 Blocked by: #46 — a blocker closed as a duplicate
211 Blocked by: #47 — a blocker closed for a reason the verdict table does not enumerate
212 Blocked by: #48 — a legacy plain close: CLOSED carrying no stateReason at all
213 Blocked by: none — superseded by PR #67
214 Blocked by: none — the limit moved into #41 and this no longer waits on it
215 Blocked by issues #7 and #41
216 Blocked by: PR #67
217 Blocked by (see below): #7
218 Blocked by: #7 AND #41
219 Blocked by: #7#41
220 **Blocked by:** #7, #41
221 Blocked by: #7 — #41 both must land
222 Blocked by: #41 — see #142
223 Blocked by: none.
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
# The four CLOSED issue verdicts. `REOPENED` is a real member of GitHub's stateReason
# enum that the table does not enumerate, so it stands in for any member it never sees.
printf '%s/issues/45 CLOSED NOT_PLANNED\n' "$R" > "$FIX/view.45"
printf '%s/issues/46 CLOSED DUPLICATE\n'   "$R" > "$FIX/view.46"
printf '%s/issues/47 CLOSED REOPENED\n'    "$R" > "$FIX/view.47"
printf '%s/issues/48 CLOSED \n'            "$R" > "$FIX/view.48"
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
absent() { # absent <label> <substring-that-must-not-appear> <positive-anchor>
  # An absence is evidence only from a run that could have carried the thing absent.
  # Inspecting $OUT alone, this passed on empty output and on rc 2 — an assertion that
  # cannot fail, the anti-pattern #138 exists to close. So rc 0 is required, and with it
  # an anchor: a substring THIS run must print, chosen to prove the declared-edge
  # machinery answered for this very query rather than falling through to `none`.
  if [ "$RC" = 0 ] \
     && case "$OUT" in *"$3"*) true  ;; *) false ;; esac \
     && case "$OUT" in *"$2"*) false ;; *) true  ;; esac; then
    printf 'PASS  %s\n' "$1"
  else
    printf 'FAIL  %s — expected rc 0 + %s without %s, got rc %s: %s\n' \
      "$1" "$3" "$2" "$RC" "$(printf '%s' "$OUT" | tr '\n' '|')"
    fails=$((fails + 1))
  fi
}

# Defect 1 — only the declared list is an edge. #136 and #206 each declare #37 alone;
# the prose behind the reason names a merged PR and an open issue, and neither gates.
run 37;  check  "prose refs are not edges -> both stories free"  0 "eligible: #136, #206"
run 41;  absent "an open issue named in prose is not an edge"      "#206" "#202 (blockers: #41 open)"

# Defect 2 — a declared blocker that is a merged pull request is satisfied, not UNKNOWN.
run 67;  check  "declared blocker is a merged PR -> satisfied"   0 "eligible: #201"
run 67;  absent "the PR named in #136's prose gained no dependent" "#136" "eligible: #201"

# The declared list itself is kept whole, both members of it.
run 7;   check  "multi-blocker list: #7 done, #41 still blocks"  0 "#202 (blockers: #41 open)"

# The boundary is a rule about glue, not a list of five separators: `·` is in no
# whitelist and is this repo's house separator, so enumerating separators dropped
# #41 here and printed #207 as eligible with a declared blocker still open.
run 7;   check  "an unenumerated separator loses no declared blocker" 0 "#207 (blockers: #41 open)"

# A story's several `Blocked by:` lines are ONE declared list, joined in the order the
# body gives them — 12 of 96 live stories carried more than one when #147 was judged.
# #208 is the only fixture with two lines, so it is the only one that can separate an
# accumulating join from a last-line-wins or a first-line-wins one. It is asked from BOTH
# ends because the join and the membership scan are separate mechanisms: the first query
# proves the list was assembled whole and in body order, the second that a member
# contributed by the LATER line is still found when the scan reaches it — a scan reading
# only the head of the joined list passes the first query and fails the second.
run 41;  check  "two Blocked by lines join into one list"        0 "#208 (blockers: #41 open, #142 open)"
run 142; check  "and the later line does not replace the earlier" 0 "#208 (blockers: #41 open, #142 open)"

# The other two pull-request states block, and the closed one is named as not completed.
run 142; check  "an open pull request blocks"                    0 "#203 (blockers: #142 open)"
run 143; check  "a pull request closed unmerged blocks"          0 "#143(NOT_COMPLETED)"

# The CLOSED verdict table (`ADR-0003 § A3`). Each branch is asserted in the direction a
# wrong answer is SILENT in: closed-but-not-completed is work that never happened, and
# reading it as satisfied promotes a story onto dead work and prints nothing anyone would
# look at. The verdict NAME is asserted with the block, because the reason is what
# @scrum-master files a `dependency_finding` on.
run 45;  check  "closed NOT_PLANNED still blocks, by that name"  0 "#209 (blockers: #45 not_planned)"
run 46;  check  "closed DUPLICATE still blocks, by that name"    0 "#210 (blockers: #46 duplicate)"
run 47;  check  "an unenumerated closed reason blocks too"       0 "#211 (blockers: #47 not_completed)"
# The other half of the discriminator #143 tests. `CLOSED` with no stateReason is the ONE
# shape `state` cannot resolve — a closed-unmerged pull request and a legacy plain close
# are identical in it — and only the url separates them. #143 must block; #48, the same
# shape on an issue, must satisfy. Either assertion alone leaves the discriminator untested.
run 48;  check  "a legacy plain-closed issue is satisfied"       0 "eligible: #212"

# Where the declared list BEGINS — the other end of the boundary above. The rule is
# in `turfgps-board-ops § The dependency representation`; these only pin it.
# (i) The label strip used to skip prose to the first `#`, so #213 and #214 each
# invented an edge out of `none`: one satisfied, one that would have held its story
# forever — the harm exactly, because a line reading `none` is never revised. Live
# story #41 reads `Blocked by: none.` and is one word from the trigger.
run 67;  absent "an empty declared list yields no edge"     "#213" "eligible: #201"
run 41;  absent "and no phantom that would block for good"  "#214" "#208 (blockers: #41 open, #142 open)"

# (ii) The bound that rejected the candidate strip `[^#0-9A-Za-z]*`: it buys (i) by
# refusing every label ornament, reading all three of these to empty and so losing a
# declared blocker in silence. #215 is asked from #7, which is satisfied — so it is
# visible at all only if the WHOLE list was read, not merely its tail.
run 7;   check  "a kind-qualified label keeps both members" 0 "#215 (blockers: #41 open)"
run 67;  check  "a PR-qualified label declares its PR"      0 "eligible: #201, #216"
run 7;   check  "a parenthetical aside is not a reason"     0 "eligible: #217"

# (iii) The three branches of that boundary no fixture reached. Each was measured
# REMOVABLE with every other assertion still green — the green-but-broken trap — and
# each fails by UNDER-reading: the list ends early, the story's one open blocker is
# never counted, and it is promoted with nothing anyone would look at. So all three are
# asked from #7, which is satisfied: a story that lost its tail does not merely print a
# shorter list, it moves to `eligible`, which is the harm itself and is what reds here.
#   Uppercase `AND` — the conjunction is spelled per character, so narrowing the class
#   to `and` keeps #215 above green and only this fixture can see it.
#   Adjacent refs — the glue run is starred, not plussed, and no other fixture puts two
#   references together, so nothing else distinguishes zero glue from one separator.
#   Bold label — `[*]*` in the strip. Without it the line matches no label at all, the
#   declared list is empty and the story leaves the report entirely rather than short.
run 7;   check  "an uppercase conjunction is still glue"    0 "#218 (blockers: #41 open)"
run 7;   check  "adjacent references need no glue at all"   0 "#219 (blockers: #41 open)"
run 7;   check  "a bold label strips like a plain one"      0 "#220 (blockers: #41 open)"

# (iv) The dash, in both directions. `turfgps-board-ops § The dependency
# representation` gives one row per shape; this pins a fixture to each of the two the
# dash owns, and does not restate the rule. Three cycles moved this boundary and each
# swing was found only afterwards, from whichever shapes that cycle happened to
# consider — a row carrying no fixture of its own is how that kept happening. Both
# assertions name the DECLARED LIST rather than a symptom, and the closing `)` is
# load-bearing: it ends the printed list, so one extra member fails the substring.
#   #221 is asked from #7, which is satisfied and therefore never printed — the story
#   is in this report at all only because #7 is in its declared list, and `#41 open`
#   proves the member on the far side of the dash survived with it. Read as a reason,
#   the list is {7} and #221 moves to `eligible:` carrying an open blocker: the
#   under-read direction, and what reds here.
#   #222 is the mirror, the reference behind the dash NOT declared, and it is asked
#   from #41 so the whole list prints. #142 is open, so reading that reference as a
#   member prints it inside the parentheses and reds. #136 and #206 carry this shape
#   already, but every reference on both of those lines is satisfied, so neither can
#   print a list that separates the two readings — they can assert only an absence.
run 7;   check  "a dash between two references loses neither"  0 "#221 (blockers: #41 open)"
run 41;  check  "and a reference behind one is still a reason" 0 "#222 (blockers: #41 open)"

# (v) The empty list without a dash — the shape live story #41 carries today, and the
# one (i) does not reach: there the dash-cut empties the line, here the label strip
# alone must. Said here rather than left to be discovered: no neutralisation of
# today's parser reds this, because the line carries no `#` for any reading of it to
# find. It is present because the shape table has a row for it and a row without a
# fixture is the defect above, and it pins that row against a parser declaring an edge
# from the LABEL rather than from a reference — the over-read direction, which is the
# only direction this shape has ever failed in.
run 41;  absent "a bare none declares nothing at all"       "#223" "#208 (blockers: #41 open, #142 open)"

# Unchanged: a DECLARED blocker that cannot be read fails toward blocked.
run 999; check  "unreadable declared blocker still blocks"       0 "#205 (blockers: #999 unknown)"
run 999; check  "and is named on its own line"                   0 "unreadable: #999"

[ "$fails" -eq 0 ] || { printf '\n%s test(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
