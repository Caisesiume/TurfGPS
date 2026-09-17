#!/usr/bin/env bash
# prose-licence-recall.sh — the corpus scripts/gates/prose-licence.sh is measured against. (#171)
#
# #171 asks for the mechanical half of the licensing rule and #184 fixes what
# makes an instrument trustworthy: it must never report 0 without having
# measured something. This file is the demonstration of both. EVERY FIXTURE
# BELOW IS RED-DEMONSTRATED against a mutant of the checker with one rule
# removed, because a corpus that only ever runs against the finished script
# proves that the script is self-consistent and nothing else.
#
# THE FAILURES ARE THE HISTORY, and each is verifiable in one command by
# someone who was not here:
#
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5559078509   # PR #163's own
#       validation run, which records `prose_licence: made_up_value_not_one_of_four`
#       reported `clean`, exit 0 — beside an artifact with NO key at all,
#       reported identically. That table is the gap this checker closes.
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5571592246   # PR #154, a
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5572663377   # judgment each,
#       both declaring `prose_licence: ruling` — a value in no table, live, twice.
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5555418634   # PR #163, two
#   gh api repos/Caisesiume/TurfGPS/issues/comments/5556183187   # judgments
#       declaring `finding_overturned, predecessor_corrected` — the value this
#       gate refuses to judge rather than decide from a gate. FIXTURE 6.
#
# NO LICENCE VALUE IS RESTATED IN THIS FILE, for the reason no cap number is
# restated in `output-caps-recall.sh`: the table is the one home. Every value
# exercised below is READ from the table at run time, and the one fixture that
# needs a token the table does NOT define asserts its absence before using it —
# a fixture built on a token the table later adopts would be vacuous and green.
#
# WHAT THIS FILE FIXES, stated as assertions rather than description:
#
#   INVOCATION   scripts/gates/prose-licence.sh <path>...   one or more artifacts.
#                The caller's working directory is not an input: every run below
#                is made from a throwaway directory outside this repository.
#   CLASSIFY     the FIRST line matching `^artifact: <id>`, fence or no fence —
#                the same anchor `output-caps.sh` classifies by, deliberately, so
#                that the two gates never disagree about which artifact they are
#                looking at. A LATER declaration is a relay and never wins;
#                FIXTURE 7 is that case and its mutant takes the last.
#   RULE 1       `prose_licence:` is the SECOND key, standing directly after
#                `artifact:`. A blank line and a comment line are not keys and do
#                not break directness; anything else that is not a key ends the
#                block. FIXTURES 1 and 2.
#   RULE 2       its value is one the licence table defines. FIXTURES 3 and 4.
#   RULE 3       NOT CHECKED — the ≤ 5 sentence limit. The checker's header
#                argues why the licensed passage has no mechanically decidable
#                boundary and declines it out loud rather than faking it. There
#                is no fixture here for a rule the instrument does not hold, and
#                a green run below is not evidence about sentence counts.
#   REPORT       one line per artifact judged, carrying —
#                    <id> · <licence> · ok · <path>
#                or  <id> · <what failed> · fail · <path>
#                — and one line per artifact refused —
#                    unclassified · <reason> · <path>
#                One summary line, always —
#                    licence · <N> checked · <M> failed · <K> unclassified
#                — and, only when N > 0 and M and K are both 0, the sentinel
#                `clean · `. EVERY cannot-vouch path below asserts that sentinel
#                is ABSENT. That is #184 in assertion form and it is the single
#                thing that stops a broken instrument printing what a clean run
#                prints.
#   EXIT         0 clean · 1 a declaration that failed a rule · 2 the instrument
#                cannot vouch for the set. Cannot-vouch outranks a failure when
#                both occur, and FIXTURE 10 asserts the outranked line is still
#                printed — precedence must never mean stopped looking.
#
# Hermetic: fixtures under mktemp, no network, no repository state except the
# licence table, which is the checker's declared input. Nothing is written
# outside TMP.
#
# Usage: scripts/gates/tests/prose-licence-recall.sh [checker]
#        The optional argument exists for an external red demonstration —
#        pointing this corpus at a stub is how it is shown to discriminate
#        rather than to be red because nothing is implemented. The checker it
#        ran is printed on every run so no output is ambiguous about what it
#        exercised.
# Exit:  0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="${1:-$DIR/../prose-licence.sh}"
# Absolutised before anything else: every check below runs from a throwaway
# directory, and a relative checker path — which is what an external red
# demonstration hands it — would resolve against the wrong place after the `cd`.
SCRIPT="$(cd "$(dirname "$SCRIPT")" 2>/dev/null && pwd)/$(basename "$SCRIPT")"
TABLE="$DIR/../../../.claude/skills/agent-handoffs/SKILL.md"

# Guarded for the reason `output-caps-recall.sh` guards it: only `set -u` is in
# force, so an unguarded failure turns every `$TMP/x` below into a path at the
# filesystem root and the EXIT trap into `rm -rf ""`.
TMP="$(mktemp -d)" || TMP=''
[ -n "$TMP" ] && [ -d "$TMP" ] || {
  printf 'FAIL  mktemp -d gave no usable directory; refusing to run rather than write outside one\n'; exit 1; }
trap 'rm -rf "$TMP"' EXIT

fails=0
OUT=''; RC=0

pass() { printf 'PASS  %s\n' "$1"; }
bad()  { printf 'FAIL  %s — %s\n' "$1" "$2"; fails=$((fails + 1)); }
# A CHECK THAT COULD NOT RUN IS PRINTED AND TALLIED, NEVER OMITTED, for the
# reason `local-gates § The law` law 3 gives: silence reads as a pass. The tally
# is reported at the bottom whether or not it is zero.
skips=0
skip() { printf 'SKIP  %s — %s\n' "$1" "$2"; skips=$((skips + 1)); }
# A construction premise that fails makes every assertion after it meaningless,
# so it stops the run rather than adding one more line to a tally.
die()  { printf 'FAIL  %s\n' "$1"; exit 1; }

[ -f "$SCRIPT" ] || die "the checker is not at $SCRIPT"
[ -f "$TABLE" ]  || die "the licence table is not at $TABLE"
printf 'checker: %s\n\n' "$SCRIPT"

REAL_TABLE="$TMP/.table-copy.md"
cp "$TABLE" "$REAL_TABLE" || die "could not copy the licence table"

cd "$TMP" || die "could not enter $TMP"

check_rc()    { [ "$RC" = "$2" ] && pass "$1" || bad "$1" "exit $RC, wanted $2 · output: $OUT"; }
check_has()   { case "$OUT" in *"$2"*) pass "$1" ;; *) bad "$1" "output does not contain '$2' · output: $OUT" ;; esac; }
check_lacks() { case "$OUT" in *"$2"*) bad "$1" "output contains '$2', which it must not · output: $OUT" ;; *) pass "$1" ;; esac; }

run() { OUT="$(bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }

# ---------------------------------------------------------------------------
# THE LICENCE TABLE IS READ, NEVER RESTATED.
#
# THE GROUND-TRUTH CHECK IS NOT `output-caps-recall.sh`'s, AND COPYING IT WOULD
# HAVE BEEN VACUOUS. That corpus counts the rows its region holds against the
# lines matching the row form ANYWHERE in the file, and a disagreement is the
# signature of a row indented out of the region. The form it counts is four
# columns and belongs to one table. THE LICENCE TABLE'S FORM IS TWO COLUMNS AND
# IS SHARED — the per-field caps table earlier in the same file satisfies it
# exactly, measured here at three lines — so a form match anywhere is not
# evidence about this table at all, and the imported check failed on the
# unmodified repository the first time it was run.
#
# The drop is therefore detected DIRECTLY: a row indented out of the region is
# an indented row-shaped line inside it, which is a thing to look for rather
# than a discrepancy to infer. That is strictly stronger than the count it
# replaces — `agent-handoffs § The cap table` records that a dropped LAST row
# grounded in no fixture is caught by neither of that corpus's checks, and the
# last row here is `none`, the most-declared value in the corpus.
# ---------------------------------------------------------------------------
VALUES="$(awk '
  /^\| `prose_licence` \|/          { in_table = 1; next }
  in_table && /^\|[-: |]+\|[ \t]*$/ { next }
  in_table && /^\|/ {
    if ($0 ~ /^\| `[a-z_]+` \| [^|]*\|[ \t]*$/) {
      split($0, f, "|"); a=f[2]; gsub(/[` ]/, "", a); print a }
    next
  }
  in_table { in_table = 0 }' "$REAL_TABLE")"
[ -n "$VALUES" ] || die "no row in $TABLE parses under the licence table's binding contract; this corpus has no ground to stand on"

ROWS_HELD="$(awk '
  /^\| `prose_licence` \|/          { in_table = 1; next }
  in_table && /^\|[-: |]+\|[ \t]*$/ { next }
  in_table && /^\|/                 { n++; next }
  in_table                          { in_table = 0 }
  END { print n + 0 }' "$REAL_TABLE")"
[ "$ROWS_HELD" -gt 0 ] || die "the licence table's row region holds no rows at all"
DROPPED="$(awk '
  /^\| `prose_licence` \|/         { in_region = 1; next }
  in_region && /^#/                { in_region = 0 }
  in_region && /^[ \t]+\| `[a-z_]+` \|/ { print NR ": " $0 }' "$REAL_TABLE")"
[ -z "$DROPPED" ] || die "a row of the licence table stands indented, which takes it and every row beneath it out of the region the checker reads, silently un-defining those values while every report stays clean. Repair the row: $DROPPED"

USED=''
need_value() { # need_value <token> — this fixture is grounded in that row
  printf '%s\n' "$VALUES" | grep -qx -- "$1" ||
    die "the licence table has no value '$1', and a fixture below is grounded in it. Re-ground the fixture; never edit a value here to match."
  case " $USED " in *" $1 "*) ;; *) USED="$USED $1" ;; esac
}
# The one token that must NOT be in the table. A fixture for "a value the table
# does not define" is vacuous the day the table defines it, so it is asserted
# absent rather than assumed absent. The token is the historical one: PR #163's
# validation run used it to record that a typo passed everything.
ABSENT='made_up_value_not_one_of_four'
printf '%s\n' "$VALUES" | grep -qx -- "$ABSENT" &&
  die "the licence table now defines '$ABSENT', so every fixture below built on it asserts nothing. Choose a token the table does not define."
FIRST_VALUE="$(printf '%s\n' "$VALUES" | head -1)"
need_value "$FIRST_VALUE"

# ---------------------------------------------------------------------------
# STAGING. A staged tree is a checker plus a table at the path the checker
# resolves from ITS OWN location — which is the property FIXTURE 11 exists to
# prove, and the mechanism every mutant below rides on.
# ---------------------------------------------------------------------------
stage() { # stage <name> — a copy of the checker and of the real table
  mkdir -p "$TMP/$1/scripts/gates" "$TMP/$1/.claude/skills/agent-handoffs" ||
    die "construction: could not stage $1"
  cp "$SCRIPT" "$TMP/$1/scripts/gates/prose-licence.sh" || die "construction: could not stage the checker for $1"
  cp "$REAL_TABLE" "$TMP/$1/.claude/skills/agent-handoffs/SKILL.md" || die "construction: could not stage the table for $1"
}
staged() { printf '%s' "$TMP/$1/scripts/gates/prose-licence.sh"; }

# mutant <name> <sed expression> — the checker with ONE rule removed.
#
# A MUTANT THAT MUTATES NOTHING PROVES NOTHING, so a copy that comes back
# byte-identical is a hard stop and never a soft failure: it means the line the
# expression targets has moved, and every red demonstration riding on it would
# be asserting the behaviour of the unmodified checker.
mutant() {
  local n="$1" expr="$2" out
  stage "$n"
  out="$TMP/$n/scripts/gates/prose-licence.sh"
  sed "$expr" "$SCRIPT" > "$out" || die "construction: sed failed building mutant '$n'"
  cmp -s "$SCRIPT" "$out" &&
    die "construction: mutant '$n' is byte-identical to the checker — the line its expression targets has moved. Re-aim the expression; do not delete the demonstration."
  printf '%s' "$out"
}
runm() { local m="$1"; shift; OUT="$(bash "$m" "$@" 2>&1)"; RC=$?; }

art() { printf '%s\n' "$2" > "$TMP/$1"; }

# ---------------------------------------------------------------------------
# FIXTURE 1 — RULE 1, PRESENCE. No `prose_licence:` key at all.
#
# The case PR #163's validation run reported `clean`, exit 0.
# ---------------------------------------------------------------------------
art nokey.md "artifact: judgment
pr: 154
sha: a9862f5"

run "$TMP/nokey.md"
check_rc    "1 presence · an artifact with no prose_licence: key fails"        1
check_has   "1 presence · ... and the key that stands second is named"         "pr: stands second"
check_lacks "1 presence · ... and never says clean"                            "clean · "

M="$(mutant pres 's|checked=$((checked + 1)); failed=$((failed + 1)); continue ;;|checked=$((checked + 1)); continue ;;|')"
runm "$M" "$TMP/nokey.md"
check_rc  "1 presence · RED: with the position and presence arms neutralised it passes"  0
check_has "1 presence · RED: ... and prints exactly what a clean run prints"              "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 2 — RULE 1, POSITION. The key is present, but not second.
#
# `agent-handoffs § The structured block comes first` makes the two keys "the
# first two", so a key present in the right file at the wrong place is a
# different failure from an absent one and says so.
# ---------------------------------------------------------------------------
art late.md "artifact: judgment
pr: 154
prose_licence: $FIRST_VALUE"

run "$TMP/late.md"
check_rc    "2 position · a prose_licence: standing third fails"                    1
check_has   "2 position · ... and is distinguished from an absent key"              "is declared but pr: stands second"
check_lacks "2 position · ... and never says clean"                                 "clean · "

M="$(mutant pos 's|checked=$((checked + 1)); failed=$((failed + 1)); continue ;;|checked=$((checked + 1)); continue ;;|')"
runm "$M" "$TMP/late.md"
check_rc  "2 position · RED: with the position arms neutralised it passes"  0
check_has "2 position · RED: ... and prints what a clean run prints"        "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 3 — RULE 2, ENUM. A value the table does not define.
#
# Both the invented token from PR #163's validation table and the live one from
# PR #154's two judgments, because a checker that rejected an obviously silly
# string and accepted a plausible one would have closed nothing.
# ---------------------------------------------------------------------------
art invented.md "artifact: judgment
prose_licence: $ABSENT
pr: 163"
art ruling.md "artifact: judgment
prose_licence: ruling
pr: 154
sha: a9862f5"

run "$TMP/invented.md"
check_rc    "3 enum · an invented value fails"                  1
check_has   "3 enum · ... and the value is named"               "$ABSENT is not a value the licence table defines"
check_lacks "3 enum · ... and never says clean"                 "clean · "

run "$TMP/ruling.md"
check_rc  "3 enum · the live PR #154 value 'ruling' fails too"  1
check_has "3 enum · ... and it is named, not merely counted"    "ruling is not a value the licence table defines"

M="$(mutant enum 's|if defined "$val"; then|if true; then|')"
runm "$M" "$TMP/invented.md"
check_rc  "3 enum · RED: with the table lookup neutralised the invented value passes"  0
check_has "3 enum · RED: ... and prints what a clean run prints"                       "clean · "
runm "$M" "$TMP/ruling.md"
check_rc  "3 enum · RED: ... and so does 'ruling'"  0

# ---------------------------------------------------------------------------
# FIXTURE 4 — RULE 2, THE CONTROL. EVERY value the table defines passes, and
# the set is read from the table rather than written here.
# ---------------------------------------------------------------------------
set -- ; i=0
for v in $VALUES; do
  need_value "$v"
  art "good-$i.md" "artifact: judgment
prose_licence: $v
pr: 163"
  set -- "$@" "$TMP/good-$i.md"
  i=$((i + 1))
done
[ "$i" -eq "$ROWS_HELD" ] || die "construction: built $i control fixtures for $ROWS_HELD rows; a value the table holds would go unexercised"

run "$@"
check_rc  "4 control · every value the table defines passes"  0
check_has "4 control · ... and the run says clean"            "clean · "
check_has "4 control · ... and reports the count it measured" "licence · $i checked · 0 failed · 0 unclassified"

# ---------------------------------------------------------------------------
# FIXTURE 5 — A KEY WITH NO VALUE. `none` is a value and not an omission, which
# `agent-handoffs § Prose is licensed, and the artifact names its licence`
# states outright, so a key declaring nothing is the case that sentence names.
#
# THE MUTANT HERE FLIPS THE REASON AND NOT THE VERDICT, and that is said rather
# than hidden: with this rule removed the empty value falls through to the table
# lookup and fails there too. What the rule contributes is a sentence an author
# can act on, so that is what the demonstration asserts disappears.
# ---------------------------------------------------------------------------
art bare.md "artifact: judgment
prose_licence:
pr: 163"

run "$TMP/bare.md"
check_rc    "5 empty · a key declaring nothing fails"                        1
check_has   "5 empty · ... and is told which sentence it broke"              "none is a value rather than an omission"
check_lacks "5 empty · ... and never says clean"                             "clean · "

M="$(mutant bare 's|if . -z "$val" .; then|if false; then|')"
runm "$M" "$TMP/bare.md"
check_lacks "5 empty · RED: with the rule removed the actionable reason is gone"  "none is a value rather than an omission"

# ---------------------------------------------------------------------------
# FIXTURE 6 — TWO LICENCES AT ONCE, REFUSED RATHER THAN JUDGED.
#
# Live on PR #163 twice. The table's value column is one token, which makes the
# pair invalid; the sentence beneath it budgets the prose "however many of the
# four apply at once", which reads as though more than one may. A gate that
# failed it would be settling that from an instrument, and one that passed it
# would be the silence this file exists to end. Cannot-vouch is the third
# answer and the honest one.
# ---------------------------------------------------------------------------
V2="$(printf '%s\n' "$VALUES" | sed -n 2p)"
need_value "$V2"
art pair.md "artifact: judgment
prose_licence: $FIRST_VALUE, $V2
pr: 163"

run "$TMP/pair.md"
check_rc    "6 two licences · a value naming two is cannot-vouch, not a verdict"  2
check_has   "6 two licences · ... and the gate says it does not decide it"        "does not decide"
check_lacks "6 two licences · ... and never says clean"                           "clean · "

M="$(mutant pair 's|    \*,\*)|    __no_such_tag)|')"
runm "$M" "$TMP/pair.md"
check_rc    "6 two licences · RED: with the refusal removed it is silently judged"  1
check_lacks "6 two licences · RED: ... and is no longer refused"                    "does not decide"

# ---------------------------------------------------------------------------
# FIXTURE 7 — THE FIRST DECLARATION WINS.
#
# A relay carries somebody else's `artifact:` and `prose_licence:` inside it,
# counted against the row where THAT was posted. The artifact's own declaration
# is therefore the first one, and it is the one judged — the same rule
# `output-caps.sh` classifies by. The fixture declares a bad licence of its own
# and relays a good one, so a checker reading the last would report clean.
# ---------------------------------------------------------------------------
art relay.md "artifact: supersession_notice
prose_licence: ruling
supersedes: \"#issuecomment-5555210525\"

\`\`\`yaml
artifact: judgment
prose_licence: $FIRST_VALUE
pr: 163
\`\`\`"

run "$TMP/relay.md"
check_rc    "7 relay · the artifact's own declaration is judged, not the one it relays"  1
check_has   "7 relay · ... and the id reported is the courier's"                          "supersession_notice · ruling"
check_lacks "7 relay · ... and never says clean"                                          "clean · "

M="$(mutant relay 's|{ state = 3; next }|{ state = 1; has = 0; next }|')"
runm "$M" "$TMP/relay.md"
check_rc  "7 relay · RED: a checker reading the LAST declaration reports clean"  0
check_has "7 relay · RED: ... and prints what a clean run prints"                "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 8 — THE LINE-ANCHOR'S TWO DEBTS, which a YAML reader would have paid
# for free. The checker's header names both; this asserts them.
#
# 8a  AN UNQUOTED `#` IS A COMMENT, and a `#` with no space before it is not.
# 8b  CRLF IS STRIPPED. #186 defect 3 is that fact biting the cap checker:
#     GitHub returns CRLF bodies, and a `\r` carried into a value turns a valid
#     token into one no table defines. The pair is asserted together — the same
#     artifact in both line endings must get the same verdict — because a
#     checker that passes both by accident of encoding proves nothing.
# ---------------------------------------------------------------------------
art commented.md "artifact: judgment
prose_licence: $FIRST_VALUE   # the one licence being invoked
pr: 163"

run "$TMP/commented.md"
check_rc  "8a comment · a trailing YAML comment is not part of the value"  0
check_has "8a comment · ... and the value read is the token alone"         "judgment · $FIRST_VALUE · ok"

M="$(mutant hash 's|.*RSTART - 1)|  v = v|')"
runm "$M" "$TMP/commented.md"
check_rc    "8a comment · RED: with the comment strip removed the value fails"  1
check_lacks "8a comment · RED: ... and never says clean"                        "clean · "

printf 'artifact: judgment\r\nprose_licence: %s\r\npr: 163\r\n' "$FIRST_VALUE" > "$TMP/crlf.md"
run "$TMP/crlf.md"
check_rc  "8b CRLF · a CRLF artifact gets the same verdict as its LF twin"  0
check_has "8b CRLF · ... and the value read carries no carriage return"     "judgment · $FIRST_VALUE · ok"

# THE RED DEMONSTRATION FOR THIS RULE IS HOST-DEPENDENT, and a claim about it
# that names no host says nothing — the shape `local-gates § Backend (Go)`
# records for the race detector. Some awks hand `\r` to the program as part of
# the record and some strip it before the program runs; where it is stripped,
# the checker's own strip is a no-op that cannot be made to matter, so removing
# it changes no verdict and a mutant proves nothing. THE RULE IS STILL
# LOAD-BEARING where `\r` survives, which is what #186 defect 3 records from the
# cap checker's side: GitHub returns CRLF bodies and the platforms disagree. So
# the host is probed, the demonstration runs where it can, and where it cannot
# it is reported as not run with the reason rather than passed by accident.
CR_SURVIVES=no
case "$(printf 'x\r\n' | awk '$0 ~ /\r$/ { print "yes" }')" in *yes*) CR_SURVIVES=yes ;; esac
printf 'host probe: awk %s carry \\r into $0\n' \
  "$([ "$CR_SURVIVES" = yes ] && echo does || echo 'does NOT')"

if [ "$CR_SURVIVES" = yes ]; then
  M="$(mutant crlf 's|{ line = $0; sub(/.r$/, "", line) }|{ line = $0 }|')"
  runm "$M" "$TMP/crlf.md"
  check_rc    "8b CRLF · RED: with the strip removed the same artifact fails on one platform"  1
  check_lacks "8b CRLF · RED: ... and never says clean"                                        "clean · "
  runm "$M" "$TMP/good-0.md"
  check_rc  "8b CRLF · RED: ... while its LF twin still passes, which is the disagreement"  0
else
  skip "8b CRLF · RED: the strip cannot be shown to matter on this host" \
       "this awk strips \\r from every record before the checker's own rule runs, so the mutant and the checker agree and the demonstration would assert nothing. The rule is load-bearing where \\r survives; #186 defect 3 is the same platform split measured from the cap checker. Run this corpus on such a host to demonstrate it."
fi

# ---------------------------------------------------------------------------
# FIXTURE 9 — THE CANNOT-VOUCH SET. Every path here asserts the `clean · `
# sentinel is ABSENT. That is #184 in assertion form: an instrument that cannot
# measure must not print what a measuring one prints.
# ---------------------------------------------------------------------------
run
check_rc    "9 zero artifacts · a set of zero is not a clean set"  2
check_has   "9 zero artifacts · ... and the reason is named"       "no artifact path was given"
check_lacks "9 zero artifacts · ... and never says clean"          "clean · "

run "$TMP/there-is-no-such-file.md"
check_rc    "9 unreadable · a path it cannot read is cannot-vouch"  2
check_has   "9 unreadable · ... and the path is named"              "unreadable path"
check_lacks "9 unreadable · ... and never says clean"               "clean · "

art noart.md "RULING: REMANDED

Three findings stand, and none of them declares anything."
run "$TMP/noart.md"
check_rc    "9 no artifact: key · an undeclared artifact is cannot-vouch"  2
check_has   "9 no artifact: key · ... and the reason is named"             "no artifact: key"
check_lacks "9 no artifact: key · ... and never says clean"                "clean · "

: > "$TMP/emptyfile.md"
run "$TMP/emptyfile.md"
check_rc    "9 empty file · nothing to read is cannot-vouch"  2
check_lacks "9 empty file · ... and never says clean"         "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 10 — PRECEDENCE. Cannot-vouch outranks a failed declaration, and the
# outranked line is still printed: precedence is about the status, never about
# stopping looking.
# ---------------------------------------------------------------------------
run "$TMP/ruling.md" "$TMP/pair.md"
check_rc    "10 precedence · cannot-vouch outranks a failed declaration"  2
check_has   "10 precedence · ... and the failed declaration is still printed"  "ruling is not a value the licence table defines"
check_has   "10 precedence · ... and the refusal is still printed"            "does not decide"
check_has   "10 precedence · ... and both are counted"                        "1 checked · 1 failed · 1 unclassified"
check_lacks "10 precedence · ... and never says clean"                        "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 11 — THE TABLE IS THE SOURCE, AND IT IS RESOLVED FROM THE CHECKER'S
# OWN PATH. A staged tree whose table defines a value this repository's does not
# must make that value pass under the staged checker and fail under the real
# one. Only a relocatable, table-driven instrument can do both, which is why one
# fixture proves both.
# ---------------------------------------------------------------------------
stage tbl
EXTRA='corpus_only_probe_value'
printf '%s\n' "$VALUES" | grep -qx -- "$EXTRA" &&
  die "the licence table now defines '$EXTRA', so this fixture asserts nothing. Choose a token it does not define."
LASTROW="$(grep -n '^| `' "$REAL_TABLE" | awk -F: -v h="$(grep -n '^| \`prose_licence\` |' "$REAL_TABLE" | head -1 | cut -d: -f1)" '$1 > h { last = $1 } END { print last }')"
[ -n "$LASTROW" ] || die "construction: could not locate the last row of the licence table"
awk -v n="$LASTROW" -v row="| \`$EXTRA\` | fixture-only value, appended by the recall corpus |" \
  'NR == n { print; print row; next } { print }' \
  "$REAL_TABLE" > "$TMP/tbl/.claude/skills/agent-handoffs/SKILL.md"
grep -q "$EXTRA" "$TMP/tbl/.claude/skills/agent-handoffs/SKILL.md" ||
  die "construction: the staged table does not hold the probe value, so this fixture asserts nothing"

art probe.md "artifact: judgment
prose_licence: $EXTRA
pr: 163"

runm "$(staged tbl)" "$TMP/probe.md"
check_rc  "11 table-driven · a value the STAGED table defines passes under the staged checker"  0
check_has "11 table-driven · ... and the staged run says clean"                                 "clean · "

run "$TMP/probe.md"
check_rc  "11 table-driven · ... and the same artifact fails against the real table"  1
check_has "11 table-driven · ... which is the enum being read rather than compiled in"  "$EXTRA is not a value the licence table defines"

# A row-shaped line in the region that does not parse is NAMED, never skipped.
# Skipping it un-defines a value while leaving every report clean, which is the
# failure `agent-handoffs § The cap table` records for its own region.
stage badrow
awk -v n="$LASTROW" -v row='| `not a code span` | broken on purpose |' \
  'NR == n { print; print row; next } { print }' \
  "$REAL_TABLE" > "$TMP/badrow/.claude/skills/agent-handoffs/SKILL.md"
runm "$(staged badrow)" "$TMP/good-0.md"
check_rc    "11 broken row · a row-shaped line that does not parse is cannot-vouch"  2
check_has   "11 broken row · ... and the offending line is printed"                  "not a code span"
check_lacks "11 broken row · ... and never says clean"                               "clean · "

# A table whose header is gone leaves the checker no region to read at all.
stage noheader
awk '/^\| `prose_licence` \|/ { print "| `PROSE_LICENCE` | header renamed by the recall corpus |"; next } { print }' \
  "$REAL_TABLE" > "$TMP/noheader/.claude/skills/agent-handoffs/SKILL.md"
cmp -s "$REAL_TABLE" "$TMP/noheader/.claude/skills/agent-handoffs/SKILL.md" &&
  die "construction: the staged table is identical to the real one — the header this fixture renames has moved"
runm "$(staged noheader)" "$TMP/good-0.md"
check_rc    "11 no header · a table whose region cannot be entered is cannot-vouch"  2
check_has   "11 no header · ... and the reason is named"                             "no row in"
check_lacks "11 no header · ... and never says clean"                                "clean · "

# A checker that cannot find its table at all.
mkdir -p "$TMP/orphan/scripts/gates" || die "construction: could not stage the orphan"
cp "$SCRIPT" "$TMP/orphan/scripts/gates/prose-licence.sh" || die "construction: could not stage the orphan checker"
runm "$TMP/orphan/scripts/gates/prose-licence.sh" "$TMP/good-0.md"
check_rc    "11 no table · a checker whose table is missing refuses rather than passes"  2
check_has   "11 no table · ... and says where it looked"                                 "is not readable at"
check_lacks "11 no table · ... and never says clean"                                     "clean · "

# ---------------------------------------------------------------------------
# FIXTURE 12 — RULE 3 IS NOT CHECKED, AND THE CORPUS SAYS SO IN AN ASSERTION.
#
# A licensed artifact carrying far more than five sentences PASSES, because the
# licensed passage has no mechanically decidable boundary and the checker
# declines that rule rather than faking it. This fixture exists so that no
# reader mistakes a green run for evidence about sentence counts, and so that
# the day a boundary marker is introduced, the assertion below fails loudly and
# names the work instead of leaving a silent gap.
# ---------------------------------------------------------------------------
{
  printf 'artifact: judgment\nprose_licence: %s\npr: 163\n\n' "$FIRST_VALUE"
  i=0
  while [ "$i" -lt 40 ]; do printf 'This is a licensed sentence well past any budget of five. '; i=$((i + 1)); done
  printf '\n'
} > "$TMP/verbose.md"

run "$TMP/verbose.md"
check_rc  "12 rule 3 declined · forty sentences under one licence still passes"  0
check_has "12 rule 3 declined · ... because this gate measures the declaration, never the passage"  "clean · "

# ---------------------------------------------------------------------------
# COVERAGE — PRINTED, NEVER ASSERTED, and computed from the table rather than
# kept by hand. A value added to the table with no fixture here appears in this
# list on the next run, which is the whole point of computing it. Do not assert
# it at a number: that turns coverage into a target, and the first thing anyone
# reaches for is a fixture that exercises a value without testing it.
# ---------------------------------------------------------------------------
n_used="$(printf '%s' "$USED" | tr ' ' '\n' | grep -c .)"
unexercised=''
for v in $VALUES; do
  case " $USED " in *" $v "*) ;; *) unexercised="$unexercised $v" ;; esac
done
printf '\nlicence-table values exercised: %s of %s\n' "$n_used" "$ROWS_HELD"
printf 'values with no fixture here:%s\n' "${unexercised:- none}"

printf '\nrules this corpus holds the checker to: 2 of the 3 in the licensing rule.\n'
printf '  1 position  · prose_licence: is the second key        · FIXTURES 1, 2\n'
printf '  2 enum      · its value is one the table defines      · FIXTURES 3, 4, 11\n'
printf '  3 sentences · NOT CHECKED, and FIXTURE 12 asserts it is not.\n'
printf 'The checker header argues why rule 3 has no mechanically decidable passage.\n'
printf 'A green run below is evidence about declarations and about nothing else.\n'

printf '\nchecks not run: %s\n' "$skips"

[ "$fails" -eq 0 ] || { printf '\n%s check(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
