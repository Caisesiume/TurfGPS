#!/usr/bin/env bash
# output-caps-recall.sh — the corpus scripts/gates/output-caps.sh is measured against. (#158)
#
# #158 asks for a cap checker and says what makes one trustworthy: "A failing
# case exists before the checker is trusted." This file is that case, written
# before the checker and by a different hand, because the instruments this
# repository has shipped that could only ever pass were the ones whose author
# wrote their own proof.
#
# THE SIZES ARE THE HISTORY. Every number this corpus builds a fixture to was
# measured on an artifact that was actually posted, and each is verifiable in
# one command by someone who was not here:
#
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5458173748 --jq '.body|length'   -> 5992
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5458097789 --jq '.body|length'   -> 6048
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5458175907 --jq '.body|length'   -> 6059
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5465790998 --jq '.body|length'   -> 13442
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5458230029 --jq '.body|length'   -> 27757
#
# The first three are the three judgments on PR #141 — the discrimination test
# this corpus exists for. The fourth is the #154 lane report that had no row to
# declare until `worker_report` was added. The fifth is the #140 supersession
# comment: a 933-char notice standing over a retained judgment and its revision
# packet.
#
# THE TEXT IS NOT THE HISTORY AND DOES NOT PRETEND TO BE. A cap checker counts
# characters and cannot read words, so quoting 6,000 characters three times over
# would put 18k of prose in this file to test an arithmetic property of its
# length — and would make every fixture hostage to an editor that trimmed one
# trailing space. Each fixture is therefore GENERATED to an exact measured size
# and verifies its own construction before anything is asserted: a fixture that
# is not the size it claims is exactly the vacuous instrument this file exists
# to prevent, so a construction miss is a hard stop and never a soft failure.
#
# CHARACTERS, NOT BYTES — AND THE DISCRIMINATION TEST DIES IF THAT IS GOT WRONG.
# `agent-handoffs § Output caps` states the unit: "Characters rather than
# tokens, because a writer can count characters." The five numbers above are
# codepoint counts. The same five comments measured in BYTES are 6105, 6156,
# 6186, 13590 and 28214 — so under a byte-counting checker the 5,992 judgment,
# which is 8 characters UNDER its cap, is 105 bytes OVER it. All three of PR
# #141's judgments would flag, the finding's "under is under" would be false,
# and the instrument would be measuring UTF-8 punctuation rather than output.
# FIXTURE 1 pins this with a pair: the same 5,992-character judgment in ASCII
# and in the real comment's mixed encoding, 5992 bytes and 6105 bytes, both
# required to report `under`.
#
# WHAT THIS FILE FIXES, so the two hands meet. The checker does not exist yet;
# these are its interface, asserted here rather than described:
#
#   INVOCATION   scripts/gates/output-caps.sh <path>...   one or more artifacts.
#                The caller's working directory is not an input: every run below
#                is made from a throwaway directory outside this repository.
#   CLASSIFY     the first line matching `^artifact: <id>`, fence or no fence —
#                a YAML block posted as a comment is commonly fenced, and the
#                first match is the artifact's own declaration. A LATER match is
#                a relay of somebody else's and never wins; FIXTURE 4 is that
#                case.
#   TABLE        `agent-handoffs § The cap table`, resolved RELATIVE TO THE
#                CHECKER'S OWN PATH and not to the caller's. The checker holds no
#                knowledge of any agent: a row is looked up, never inferred.
#                FIXTURE 10 stages a table this repository does not have, which
#                is only possible if the script is relocatable, and that is the
#                reason for the rule.
#   MEASURE      characters — every character in the file, the final newline
#                included, per the row's `counts` token:
#                  whole — every character of the file.
#                  body  — every character except lines beginning with `|`, and
#                          lines from a `findings:` key up to the next key at the
#                          same indentation. THE EXCLUSION ENDS AT THAT KEY; it
#                          does not run to EOF, and FIXTURE 3 fails a checker
#                          that lets it.
#                  own   — every character outside fenced blocks. The fence lines
#                          are part of the block they delimit and are excluded
#                          with it: a rule that counts the delimiter of the thing
#                          it excludes is measuring punctuation.
#   VERDICT      `over` when measured > cap. measured == cap is `under`; the
#                caps are written `<=` where they are prose, and FIXTURE 2 pins
#                the boundary at cap and cap+1.
#   REPORT       one line per classified artifact, carrying this quadruple —
#                    <id> · <measured> · <cap> · <verdict>
#                — and the path. One line per unclassifiable artifact —
#                    unclassified · <reason> · <path>
#                — where reason contains `no artifact: key` or `no cap-table row`.
#                One summary line, always —
#                    caps · <N> measured · <M> over · <K> unclassified
#                — and, only when N > 0 and M and K are both 0, the sentinel
#                `clean · `. Every cannot-run path below asserts that sentinel is
#                ABSENT, which is the one assertion that stops a broken
#                instrument from printing what a clean run prints.
#   EXIT         0 clean · 1 an artifact over its cap · 2 the instrument cannot
#                vouch for the set: an unclassifiable artifact, a path it could
#                not read, zero artifacts measured, or a cap table it could not
#                parse. #158 requires only NON-ZERO on all three; the split
#                follows `d8-root-run-claims.sh`, because "I could not measure
#                you" and "you wrote too much" are different sentences and only
#                one of them is the author's fault. Cannot-run outranks over-cap
#                when both occur, and FIXTURE 8 asserts the outranked line is
#                still printed — precedence must never mean stopped looking.
#
# NO CAP NUMBER APPEARS IN THIS FILE. The cap table says its numbers "appear
# nowhere else in the repository", and a corpus that hardcoded 6000 would be the
# second home that rule forbids, inside the instrument built to enforce it. Every
# cap is READ from the table at run time, every expected verdict is DERIVED from
# what was read, and where a historical size only discriminates against a
# particular cap the corpus asserts the STRADDLE — 5,992 must be at or under the
# judgment cap and 6,048 must be over it — and dies saying so if the table moves
# out from under the fixture. Re-ground the fixture; never edit a number here to
# match.
#
# Hermetic: fixtures generated under mktemp, no network, no repository state
# except the cap table, which is the checker's declared input. Nothing is
# written outside TMP.
# Usage: scripts/gates/tests/output-caps-recall.sh [checker]
#        The optional argument exists for the red demonstration — pointing this
#        corpus at a stub is how it is shown to discriminate rather than merely
#        to be red because nothing is implemented yet. The checker it ran is
#        printed on every run so no output is ambiguous about what it measured.
# Exit:  0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="${1:-$DIR/../output-caps.sh}"
# Absolutised before anything else, because this corpus runs every check from a
# throwaway directory and a relative checker path — which is what the red
# demonstration hands it — would resolve against the wrong place after the `cd`.
SCRIPT="$(cd "$(dirname "$SCRIPT")" 2>/dev/null && pwd)/$(basename "$SCRIPT")"
TABLE="$DIR/../../../.claude/skills/agent-handoffs/SKILL.md"

# Guarded for the reason `d8-root-run-claims-recall.sh` guards it: only `set -u`
# is in force, so an unguarded failure turns every `$TMP/x` below into a path at
# the filesystem root and the EXIT trap into `rm -rf ""`.
TMP="$(mktemp -d)" || TMP=''
[ -n "$TMP" ] && [ -d "$TMP" ] || {
  printf 'FAIL  mktemp -d gave no usable directory; refusing to run rather than write outside one\n'; exit 1; }
trap 'rm -rf "$TMP"' EXIT

fails=0
OUT=''; RC=0

pass() { printf 'PASS  %s\n' "$1"; }
bad()  { printf 'FAIL  %s — %s\n' "$1" "$2"; fails=$((fails + 1)); }
# A construction premise that fails makes every assertion after it meaningless,
# so it stops the run rather than adding one more line to a tally.
die()  { printf 'FAIL  %s\n' "$1"; exit 1; }

[ -f "$SCRIPT" ] || die "the checker is not at $SCRIPT"
[ -f "$TABLE" ]  || die "the cap table is not at $TABLE"
printf 'checker: %s\n\n' "$SCRIPT"

# THE CALLER'S DIRECTORY IS NOT AN INPUT, and this is where that is enforced:
# every run below is made from outside the repository, so a checker that swept
# its cwd or resolved its table from it fails the whole corpus rather than
# passing by accident of where the gate was typed.
cd "$TMP" || die "could not enter $TMP"

# ---------------------------------------------------------------------------
# THE CAP TABLE IS READ, NEVER RESTATED.
# ---------------------------------------------------------------------------
ROWS="$(awk '/^\| `[a-z_]+` \| `[0-9]+` \| `[a-z]+` \|/ {
                split($0, f, "|"); a=f[2]; b=f[3]; c=f[4]
                gsub(/[` ]/, "", a); gsub(/[` ]/, "", b); gsub(/[` ]/, "", c)
                print a, b, c }' "$TABLE")"
[ -n "$ROWS" ] || die "no row in $TABLE parses under the table's own binding contract; this corpus has no ground to stand on"

cap_of()    { printf '%s\n' "$ROWS" | awk -v i="$1" '$1 == i { print $2 }'; }
counts_of() { printf '%s\n' "$ROWS" | awk -v i="$1" '$1 == i { print $3 }'; }

# Every id this corpus exercises registers itself, so the coverage printed at
# the bottom is measured rather than kept by hand — which is the failure mode
# `d8-root-run-claims-recall.sh` records for its own hand-kept rows.
USED=''
need_row() { # need_row <id> <the counts token this fixture is built for>
  local c t
  c="$(cap_of "$1")"; t="$(counts_of "$1")"
  [ -n "$c" ] || die "the cap table has no row for '$1', and this corpus's fixture for it is grounded in that row"
  [ "$t" = "$2" ] || die "'$1' now counts '$t', not '$2' — the fixture built for the '$2' rule would measure the wrong thing. Re-ground the fixture; do not edit this line."
  case " $USED " in *" $1 "*) ;; *) USED="$USED $1" ;; esac
}

# ---------------------------------------------------------------------------
# FIXTURE CONSTRUCTION. The corpus assembles from parts it labelled itself; it
# never parses a fixture back. `counted` is bytes-written minus the bytes it
# put in an excluded part minus the extra bytes its one multibyte part costs —
# construction knowledge, not a second implementation of the checker's rule.
# ---------------------------------------------------------------------------
L63='padding to the size measured on the artifact this fixture holds'
[ "${#L63}" -eq 63 ] || die "the filler line is ${#L63} characters, not 63; every generated size below is computed from that width"
MIDDOT="$(printf '\302\267')"   # U+00B7, written as bytes so this file's own encoding is not load-bearing

CUR=''; EXCL=0; MB=0
bytes_of() { wc -c < "$1" | tr -d ' \t\r\n'; }
counted()  { echo $(( $(bytes_of "$CUR") - EXCL - MB )); }

art_new() { CUR="$TMP/$1"; : > "$CUR"; EXCL=0; MB=0; }
put()     { printf '%s\n' "$1" >> "$CUR"; }
put_x()   { printf '%s\n' "$1" >> "$CUR"; EXCL=$(( EXCL + ${#1} + 1 )); }
put_block()   { local t; t="$(cat)"; printf '%s\n' "$t" >> "$CUR"; }
put_block_x() { local t n; t="$(cat)"; printf '%s\n' "$t" >> "$CUR"
                n="$(printf '%s\n' "$t" | wc -c | tr -d ' \t\r\n')"; EXCL=$(( EXCL + n )); }

# EVERY FIXTURE IS NEWLINE-TERMINATED AND THE FINAL NEWLINE IS PART OF THE
# COUNT. It is a character in the file and there is no reason for a character
# count to hold an opinion about which one it is — and the alternative costs
# more than it looks: `body` and `own` are line-oriented rules, so an
# implementation that filters lines and reassembles them re-adds a terminator
# the last line did not have, and every one of those fixtures would measure one
# over. Fixing that in the checker means bookkeeping about a POSIX convention in
# an instrument whose whole claim is that it counts characters. Fixing it here
# costs a newline. The three PR #141 judgments are 5,992 / 6,048 / 6,059
# characters either way; what moves is only which character is last.
fill_nl() { local n="$1"; while [ "$n" -ge 64 ]; do printf '%s\n' "$L63"; n=$((n - 64)); done
            [ "$n" -gt 0 ] && printf '%*s\n' "$((n - 1))" ''; return 0; }
fill_x()  { fill_nl "$1" >> "$CUR"; EXCL=$(( EXCL + $1 )); }

pad_to() { _pad "$1"; }
_pad() {
  local need; need=$(( $1 - $(counted) ))
  [ "$need" -ge 0 ] || die "construction: $CUR already counts $(counted), past its target of $1 by $(( -need ))"
  fill_nl "$need" >> "$CUR"
  [ "$(counted)" = "$1" ] || die "construction: $CUR counts $(counted) and claims $1 — a fixture that is not the size it claims proves nothing"
}

run() { OUT="$(bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }

check_rc()    { [ "$RC" = "$2" ] && { pass "$1"; return 0; }
                bad "$1" "expected rc $2, got rc $RC: $(printf '%s' "$OUT" | tr '\n' '|' | cut -c1-200)"; }
check_has()   { case "$OUT" in *"$2"*) pass "$1" ;; *) bad "$1" "output does not contain: $2" ;; esac; }
check_lacks() { case "$OUT" in *"$2"*) bad "$1" "output must not contain: $2" ;; *) pass "$1" ;; esac; }
check_count() { local n; n="$(printf '%s\n' "$OUT" | grep -cF -- "$2")"
                [ "$n" = "$3" ] && { pass "$1"; return 0; }
                bad "$1" "expected $3 line(s) carrying '$2', found $n"; }

# The report quadruple, built from what the table said rather than from a
# literal. `verdict_for` is the whole of the over/under rule, derived once.
verdict_for() { [ "$1" -gt "$2" ] && echo over || echo under; }
report_of()   { printf '%s · %s · %s · %s' "$1" "$2" "$(cap_of "$1")" "$(verdict_for "$2" "$(cap_of "$1")")"; }
summary_of()  { printf 'caps · %s measured · %s over · %s unclassified' "$1" "$2" "$3"; }

# ---------------------------------------------------------------------------
# FIXTURE 1 — THE DISCRIMINATION TEST. PR #141's three judgments, at the sizes
# they were actually posted at. An instrument that cannot separate these three
# is not measuring: 5,992 is under, 6,048 and 6,059 are over by 48 and 59.
#
# The straddle is asserted rather than the cap restated. If the judgment cap
# ever moves outside it these three numbers stop discriminating, and this corpus
# stops rather than quietly asserting whatever the new cap implies.
# ---------------------------------------------------------------------------
need_row judgment body
JCAP="$(cap_of judgment)"
[ "$JCAP" -ge 5992 ] && [ "$JCAP" -lt 6048 ] || die "PR #141's judgments measured 5992 / 6048 / 6059, and they only discriminate against a judgment cap between 5992 and 6047. The table now reads $JCAP. Re-ground the fixture triple against artifacts that straddle the new cap; do not edit these numbers."

# The under case, in ASCII, carrying the resolution table its row excludes. The
# table is ~400 characters, so a checker that counts it reports this judgment
# over — the `body` rule is asserted at the level of the VERDICT here, not only
# of the number.
mk_judgment() { # mk_judgment <file> <measured-target> <with-table:yes|no>
  art_new "$1"
  put 'artifact: judgment'
  put 'prose_licence: none'
  put 'pr: 141'
  put 'cycle: 1'
  put 'ruling: remanded'
  put 'lanes: [docs, correctness, testing]'
  put ''
  if [ "$3" = yes ]; then
    put_block_x <<'EX'
| finding | lane | severity | resolution |
|---|---|---|---|
| F-01 | docs | HIGH | upheld — the citation names no document and orphans as this file |
| F-02 | correctness | MED | upheld — the boundary is exclusive where the record says inclusive |
| F-03 | testing | MED | overturned — the case it names is covered two levels down |
| F-04 | docs | LOW | upheld — the heading cannot be cited whole without nesting |
EX
  fi
  pad_to "$2"
}

mk_judgment j-5992-ascii.md 5992 yes
mk_judgment j-6048.md       6048 yes
mk_judgment j-6059.md       6059 yes

# The same 5,992-character judgment in the encoding the real comment was posted
# in: 5992 characters occupying 6105 bytes, the pair measured on comment
# 5458173748. It carries no excluded table, so the whole file is the artifact and
# its two numbers are the real one's two numbers — the last of the 5,992 is a
# newline here and is not one there, which is the only difference and is not one
# a character count can see. A byte-counting checker reports this 105 over its
# cap and the discrimination test collapses.
art_new j-5992-utf8.md
put 'artifact: judgment'
put 'prose_licence: none'
put 'pr: 141'
put 'cycle: 1'
put 'ruling: remanded'
put ''
mid=''; i=0
while [ "$i" -lt 113 ]; do mid="$mid$MIDDOT"; i=$((i + 1)); done
printf '%s\n' "$mid" >> "$CUR"; MB=$((MB + 113))
pad_to 5992
[ "$(bytes_of "$CUR")" = 6105 ] || die "construction: the mixed-encoding judgment holds $(bytes_of "$CUR") bytes, and comment 5458173748 holds 6105 — the byte/character pin is the whole of this fixture"

run "$TMP/j-5992-ascii.md" "$TMP/j-6048.md" "$TMP/j-6059.md" "$TMP/j-5992-utf8.md"
check_rc  "triple · two of PR #141's three judgments are over cap"          1
check_has "triple · 5,992 is UNDER, and its excluded table is not counted"  "$(report_of judgment 5992)"
check_has "triple · 6,048 is over by 48 and is caught"                      "$(report_of judgment 6048)"
check_has "triple · 6,059 is over by 59 and is caught"                      "$(report_of judgment 6059)"
check_has "triple · all four were read, none skipped"                       "$(summary_of 4 2 0)"
check_has "triple · each artifact is named"                                 "j-6048.md"

# UNDER MUST BE ABLE TO GO GREEN ALONE. In the mixed run the exit status is 1
# whatever the under case did, so "under is under" is not proven there.
run "$TMP/j-5992-ascii.md"
check_rc    "under alone · a judgment 8 characters under its cap passes"   0
check_has   "under alone · ... and says so"                                "clean · "

run "$TMP/j-5992-utf8.md"
check_rc    "encoding · 5992 characters in 6105 bytes still passes"        0
check_has   "encoding · ... measured in characters, not bytes"             "$(report_of judgment 5992)"
check_lacks "encoding · ... and 6105 is not what was measured"             "6105"

# ---------------------------------------------------------------------------
# FIXTURE 2 — `whole`, and the boundary. Nothing is excluded from a `whole`
# artifact, including the fenced block its own YAML is posted in, and cap+1 is
# the tightest over-cap case there is.
# ---------------------------------------------------------------------------
need_row worker_envelope whole
WCAP="$(cap_of worker_envelope)"

# At exactly the cap, and fenced the way a YAML envelope is usually posted: the
# declaration is the first `artifact:` line whether or not a fence precedes it,
# and `whole` counts the fence too.
art_new we-at-cap.md
put '```yaml'
put 'artifact: worker_envelope'
put 'prose_licence: none'
put 'status: completed'
put 'issue: 158'
put 'files_changed: [scripts/gates/tests/output-caps-recall.sh]'
put 'confidence: 0.9'
put '```'
pad_to "$WCAP"

art_new we-over-1.md
put 'artifact: worker_envelope'
put 'prose_licence: none'
put 'status: completed'
put 'issue: 158'
pad_to $((WCAP + 1))

run "$TMP/we-at-cap.md"
check_rc  "whole · measured == cap is under, not over"          0
check_has "whole · ... and the fenced declaration classified"   "$(report_of worker_envelope "$WCAP")"

run "$TMP/we-over-1.md"
check_rc  "whole · one character over the cap is caught"        1
check_has "whole · ... and reported as over"                    "$(report_of worker_envelope $((WCAP + 1)))"

# ---------------------------------------------------------------------------
# FIXTURE 3 — `body`, and the half of the rule a checker is most likely to get
# wrong. The exclusion runs from `findings:` to the NEXT KEY AT THE SAME
# INDENTATION, not to EOF. Everything after `confidence:` here is counted, and
# it is nearly the whole artifact — so a checker that excludes to EOF measures
# far under and reports this one CLEAN. The verdict, not just the number, is
# what fails.
# ---------------------------------------------------------------------------
need_row reviewer_verdict body
VCAP="$(cap_of reviewer_verdict)"

art_new rv-terminates.md
put 'artifact: reviewer_verdict'
put 'prose_licence: none'
put 'lane: testing'
put 'verdict: request_changes'
put_block_x <<'EX'
findings:
  - id: T-01
    severity: high
    text: the corpus asserts a verdict the fixture cannot produce
  - id: T-02
    severity: medium
    text: the adversarial branch is named in the criteria and absent from the suite
EX
put 'confidence: 0.81'
put_block_x <<'EX'
| id | severity | required |
|---|---|---|
| T-01 | high | rebuild the fixture at the size it claims |
EX
pad_to $((VCAP + 1))

run "$TMP/rv-terminates.md"
check_rc  "body · the findings exclusion ends at the next key, not at EOF"  1
check_has "body · ... so one character over the cap is still over"          "$(report_of reviewer_verdict $((VCAP + 1)))"

# ---------------------------------------------------------------------------
# FIXTURE 4 — `own`, the fenced relay, and the classifier's adversarial case.
#
# A courier comment relaying a worker envelope verbatim holds TWO `artifact:`
# declarations. The first is its own and decides the row; the second belongs to
# the artifact it is carrying, is counted against that artifact's own row where
# that artifact was posted, and must not reclassify its courier. A checker that
# greps for the last match, or for any match, measures a 1,200-char `own` comment
# against a 1,500-char `whole` row and reports a number belonging to neither.
# ---------------------------------------------------------------------------
need_row orchestrator_comment own
CCAP="$(cap_of orchestrator_comment)"

art_new oc-relay.md
put 'artifact: orchestrator_comment'
put 'prose_licence: none'
put 'to: test-engineer'
put 'issue: 158'
put 'relaying: worker_envelope'
put ''
put_block_x <<'EX'
```yaml
artifact: worker_envelope
prose_licence: none
status: completed
issue: 158
EX
fill_x 3900
put_x '```'
pad_to "$CCAP"   # the counted region closes here; the block after it is excluded
put_x '```'
fill_x 400
put_x '```'
# The trailing fence above is the second half of the same rule: a fenced block
# anywhere in the file is excluded, not merely the first one. `pad_to` closed
# the counted region before it, so the courier's own words measure exactly the
# cap and the file is several times that.
[ "$(counted)" = "$CCAP" ] || die "construction: the relay's own words count $(counted) and the fixture claims $CCAP"
OC_TOTAL="$(bytes_of "$CUR")"

run "$TMP/oc-relay.md"
check_rc    "own · the courier's own words are at cap; the relay is not counted"  0
check_has   "own · ... measured against ITS row, not the relayed envelope's"      "$(report_of orchestrator_comment "$CCAP")"
check_lacks "own · ... and the relayed declaration never reclassifies its courier" "worker_envelope ·"

# ---------------------------------------------------------------------------
# FIXTURE 5 — the large historical artifacts, now that each has a row.
#
# The 13,442-char lane report on #154 is the one that had no row to declare at
# all: it is not the worker ENVELOPE, which is the YAML the manager reads. A
# checker built against the table as it stood would have measured the small
# violations and stayed silent on this.
# ---------------------------------------------------------------------------
need_row worker_report body
RCAP="$(cap_of worker_report)"
[ 13442 -gt "$RCAP" ] || die "the #154 lane report measured 13442 and the worker_report cap now reads $RCAP; the fixture no longer demonstrates anything"

art_new wr-13442.md
put 'artifact: worker_report'
put 'prose_licence: none'
put 'issue: 154'
put 'pr: 154'
put_block_x <<'EX'
| criterion | verified by | evidence |
|---|---|---|
| AC-1 | test | TestClassifyRestArea |
| AC-2 | test | TestClassifyFencedPath |
EX
pad_to 13442

run "$TMP/wr-13442.md"
check_rc  "historical · the 13,442-char #154 lane report is caught"   1
check_has "historical · ... and its size is reported, not rounded"    "$(report_of worker_report 13442)"

# ---------------------------------------------------------------------------
# FIXTURE 6 — THE KNOWN FLOOR, asserted rather than hoped.
#
# `own` sees only unfenced text, and the cap table says so in its own words: the
# measurement is a floor, not the whole answer. The #140 supersession comment is
# the case — a 933-character notice standing over a retained judgment and its
# revision packet, 27,757 characters in total. Fenced, as the table prescribes,
# it measures 933 and PASSES, and this corpus asserts that it passes. An
# instrument whose limit is written down is not the same as one that lies, and
# the way to keep it that way is to state the limit in an assertion a reader
# trips over rather than in a sentence they can skip.
# ---------------------------------------------------------------------------
need_row supersession_notice own
SCAP="$(cap_of supersession_notice)"
[ 933 -le "$SCAP" ] && [ 27757 -gt "$SCAP" ] || die "the #140 notice measured 933 own of 27757 total, and it only demonstrates the floor when 933 is within the supersession_notice cap and 27757 is not. The table now reads $SCAP."

art_new sn-floor.md
put 'artifact: supersession_notice'
put 'prose_licence: predecessor_corrected'
put 'supersedes: 5458230029'
put 'record_of_record: [5458199191, 5458242133]'
put 'retained: fenced, unedited, for audit'
pad_to 933   # the counted region closes here; the fenced retention follows
put_x '```'
fill_x $(( 27757 - 933 - 8 ))
put_x '```'
[ "$(bytes_of "$CUR")" = 27757 ] || die "construction: the floor fixture holds $(bytes_of "$CUR") characters and the #140 comment holds 27757"

run "$TMP/sn-floor.md"
check_rc  "floor · a 27,757-character artifact measuring 933 PASSES"   0
check_has "floor · ... and the floor is what was reported"             "$(report_of supersession_notice 933)"

# ---------------------------------------------------------------------------
# FIXTURE 7 — THE COMPOSITE, and the boundary of what this instrument claims.
#
# The #140 shape: a notice standing over a judgment standing over a revision
# packet, posted as one comment under one declaration. The checker does NOT
# split it. It measures the whole under the single id the artifact declared and
# flags it over cap — ONE line, not three. Detection holds; attribution is out
# of scope, and it is out of scope on purpose: apportioning 27,757 characters
# between three artifacts requires knowing where each began, which is the
# semantic judgement #139 rejected and which a character count is trusted
# precisely for not attempting.
#
# The retained copy is unfenced here, which is how it was actually posted; the
# fenced form is FIXTURE 6 and the two together are the whole of the rule.
# ---------------------------------------------------------------------------
art_new sn-composite.md
put 'artifact: supersession_notice'
put 'prose_licence: predecessor_corrected'
put 'supersedes: 5458199191'
put ''
put '## SUPERSEDED - read the reconciliation instead'
put ''
put 'Retained unedited below for audit.'
put ''
put '---'
put ''
put 'JUDGMENT - PR #140, cycle 1 of 5. RULING: REMANDED.'
put '| finding | lane | resolution |'
put '|---|---|---|'
put '| LQ-01 | safety | upheld |'
put ''
put 'REVISION PACKET - 12 required changes follow.'
pad_to 27757

run "$TMP/sn-composite.md"
check_rc    "composite · an unfenced composite is measured whole and caught"  1
check_has   "composite · ... at the size of the whole comment"                "$(report_of supersession_notice 27757)"
check_count "composite · ONE report line, not three"                          "supersession_notice · " 1
check_lacks "composite · ... the buried judgment is not attributed"           "judgment · "
check_lacks "composite · ... nor the buried revision packet"                  "revision_packet · "

# ---------------------------------------------------------------------------
# FIXTURE 8 — UNCLASSIFIABLE, the second of the three non-zero exits. Two
# shapes, and neither may be silently skipped: an artifact the instrument cannot
# classify is an artifact it did not measure, and a checker that quietly drops it
# reports a clean set it never read.
# ---------------------------------------------------------------------------
art_new no-key.md
put_block <<'EX'
## Cycle 2 report

Everything below is prose. There is no structured block, which is the defect
`agent-handoffs § The structured block comes first` exists to catch, and there
is nothing here for the checker to look up.
EX

art_new no-row.md
put 'artifact: courier_note'
put 'prose_licence: none'
put 'note: an id that no row in the cap table defines'

run "$TMP/no-key.md"
check_rc    "unclassified · an artifact with no declaration cannot be measured"  2
check_has   "unclassified · ... and the reason is named"                         "no artifact: key"
check_has   "unclassified · ... and so is the file"                              "no-key.md"
check_lacks "unclassified · ... and it never reports a clean set"                "clean · "

run "$TMP/no-row.md"
check_rc    "unclassified · an id with no row cannot be measured"                2
check_has   "unclassified · ... and the reason is named"                         "no cap-table row"
check_has   "unclassified · ... and so is the id it could not find"              "courier_note"
check_lacks "unclassified · ... and it never reports a clean set"                "clean · "

# PRECEDENCE, AND THAT IT IS NOT AN EXCUSE TO STOP LOOKING. Cannot-run outranks
# over-cap in the exit status; both lines are still printed, and the count of
# what was measured excludes only what could not be.
run "$TMP/no-key.md" "$TMP/j-6048.md" "$TMP/no-row.md"
check_rc  "mixed · cannot-run outranks over-cap in the status"        2
check_has "mixed · ... and the over-cap artifact is still reported"   "$(report_of judgment 6048)"
check_has "mixed · ... and the tally separates measured from not"     "$(summary_of 1 1 2)"

# ---------------------------------------------------------------------------
# FIXTURE 9 — ZERO MEASURED, the third non-zero exit, and the failure this whole
# instrument would otherwise have. A gate wired with a stale glob, a renamed
# directory or a typo measures nothing, and nothing is what a clean set looks
# like from the outside.
# ---------------------------------------------------------------------------
run
check_rc    "zero · no arguments is 'cannot run', not 'clean'"        2
check_has   "zero · ... and it says what it measured"                 "$(summary_of 0 0 0)"
check_lacks "zero · ... and never says clean"                         "clean · "

run "$TMP/there-is-no-such-artifact.md"
check_rc    "zero · a path it cannot read is not silently skipped"    2
check_has   "zero · ... and the path is named"                        "there-is-no-such-artifact.md"
check_lacks "zero · ... and never says clean"                         "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 10 — THE INSTRUMENT'S OWN SOURCE OF TRUTH.
#
# The cap table says the checker "holds no knowledge of any agent" and that
# adding a capped artifact "is a row, never a change to the instrument". Both
# halves are tested here, and both are only testable because the checker resolves
# its table from its own directory: the script is copied into a staged tree
# beside a table this repository does not have.
#
# A checker with the ids compiled into it passes every fixture above and fails
# the first of these two.
# ---------------------------------------------------------------------------
stage() { # stage <name> — a relocated checker and a staged table beside it
  mkdir -p "$TMP/$1/scripts/gates" "$TMP/$1/.claude/skills/agent-handoffs"
  cp "$SCRIPT" "$TMP/$1/scripts/gates/output-caps.sh"
}

# A row the instrument has never seen, appended after the table's last row.
stage ext
PROBE_ROW='| `probe_artifact` | `137` | `whole` | fixture-only row, appended by the recall corpus |'
last="$(grep -n '^| `[a-z_]*` | `[0-9]*` | `[a-z]*` |' "$TABLE" | tail -1 | cut -d: -f1)"
[ -n "$last" ] || die "could not locate the last cap-table row in $TABLE"
awk -v n="$last" -v row="$PROBE_ROW" 'NR == n { print; print row; next } { print }' \
  "$TABLE" > "$TMP/ext/.claude/skills/agent-handoffs/SKILL.md"

art_new probe-at-cap.md
put 'artifact: probe_artifact'
put 'prose_licence: none'
pad_to 137
art_new probe-over.md
put 'artifact: probe_artifact'
put 'prose_licence: none'
pad_to 138

OUT="$(bash "$TMP/ext/scripts/gates/output-caps.sh" "$TMP/probe-at-cap.md" 2>&1)"; RC=$?
check_rc  "table-driven · a row the instrument has never seen is honoured"  0
check_has "table-driven · ... and measured against that row's cap"          "probe_artifact · 137 · 137 · under"

OUT="$(bash "$TMP/ext/scripts/gates/output-caps.sh" "$TMP/probe-over.md" 2>&1)"; RC=$?
check_rc  "table-driven · and one character over that row's cap is caught"  1
check_has "table-driven · ... reported against the staged row"              "probe_artifact · 138 · 137 · over"

# A table it cannot parse. The rows are mangled so none matches the table's own
# binding contract; the instrument has no source of truth and must say so.
stage mal
sed 's/^| `/| /' "$TABLE" > "$TMP/mal/.claude/skills/agent-handoffs/SKILL.md"
OUT="$(bash "$TMP/mal/scripts/gates/output-caps.sh" "$TMP/j-5992-ascii.md" 2>&1)"; RC=$?
check_rc    "table · an unparseable cap table is 'cannot run', not 'clean'"  2
check_has   "table · ... and the table is named"                             "SKILL.md"
check_lacks "table · ... and it never reports a clean set"                   "clean · "

# ---------------------------------------------------------------------------
# COVERAGE — PRINTED, NEVER ASSERTED, and computed from the table rather than
# kept by hand. `d8-root-run-claims-recall.sh` records what a hand-kept second
# statement of a measured number does: it drifted for a whole cycle while every
# assertion of the day passed. A row added to the table with no fixture here
# appears in this list on the next run, which is the whole point of computing it.
# Do not assert it at a value: that turns coverage into a target and the first
# thing anyone reaches for is a fixture that exercises a row without testing it.
# ---------------------------------------------------------------------------
all_ids="$(printf '%s\n' "$ROWS" | awk '{print $1}' | sort)"
n_all="$(printf '%s\n' "$all_ids" | grep -c .)"
n_used="$(printf '%s' "$USED" | tr ' ' '\n' | grep -c .)"
unexercised=''
for id in $all_ids; do
  case " $USED " in *" $id "*) ;; *) unexercised="$unexercised $id" ;; esac
done
printf '\ncap-table rows exercised: %s of %s\n' "$n_used" "$n_all"
printf 'rows with no fixture here:%s\n' "${unexercised:- none}"

# The floor, printed where it cannot be missed, and with the numbers this run
# actually built rather than the numbers a sentence claims.
printf '\nthe `own` floor is real, and it is asserted rather than hoped:\n'
printf '  FIXTURE 4  a %s-character courier comment measures %s and PASSES\n' "$OC_TOTAL" "$CCAP"
printf '  FIXTURE 6  a 27757-character supersession comment measures 933 and PASSES\n'
printf 'Both are correct under the rule and both under-report. The cap table records\nthat limit; this corpus pins it so that no reader has to take it on faith.\n'

[ "$fails" -eq 0 ] || { printf '\n%s check(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
