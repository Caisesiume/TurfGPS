#!/usr/bin/env bash
# prose-licence.sh — is the licence DECLARED, and is it one the table defines? (#171)
#
# A COMPANION TO `scripts/gates/output-caps.sh`, NEVER A WIDENING OF IT. That
# instrument measures a length, and its own header draws the boundary this one
# is built beside: "A cap is a length. Whether the prose inside an artifact was
# licensed is a question about meaning: `prose_licence:` is declared for a
# reader and is not measured here." THAT BOUNDARY STAYS, and nothing below
# crosses it. What is mechanical is not the licence's MEANING but its
# DECLARATION, and exactly two properties of a declaration are checked here:
#
#   1. POSITION — `prose_licence:` is the SECOND key of the artifact, standing
#      directly after `artifact:`, per
#      `agent-handoffs § The structured block comes first`, which makes both
#      keys mandatory and makes them the first two.
#   2. ENUM — its value is one of the values
#      `agent-handoffs § Prose is licensed, and the artifact names its licence`
#      defines. A typo passed everything until this file existed, and that is
#      recorded rather than supposed: PR #163's own validation run reports
#      `prose_licence: made_up_value_not_one_of_four` as `clean`, exit 0,
#      beside an artifact with no key at all reported identically —
#      `gh api repos/Caisesiume/TurfGPS/issues/comments/5559078509`. Live
#      artifacts fail it too: two judgments on PR #154 declare
#      `prose_licence: ruling`, comments 5571592246 and 5572663377.
#
# THE THIRD RULE — ≤ 5 SENTENCES OF LICENSED PROSE — IS NOT CHECKED HERE, and
# this is the file saying so rather than a reader discovering it from a clean
# line. The rule is real and `agent-handoffs § Prose is licensed, and the
# artifact names its licence` states it; what is missing is a mechanically
# decidable boundary for the passage it governs. The licensing rule places
# prose "after" the structured block, and no artifact marks where that block
# ends. Measured on the real judgment at
# `gh api repos/Caisesiume/TurfGPS/issues/comments/5555418634`: its YAML block
# closes after six keys, and everything after it — roughly forty sentences —
# is a mixture of the `judgment` shape's OWN field values written as bold-lead
# lines (`**Risk** medium · 0.38 · carried`), gate lines quoted verbatim, a
# findings tally, and the licensed argument itself. Separating the third from
# the first two is reading, not parsing. A checker that took "everything after
# the block" would fail that artifact at eight times the limit while naming
# nothing an author could act on, and a checker that guessed which paragraphs
# were argument would be the semantic classifier #139 rejected. So rule 3 is
# declined, out loud, and the gate reports two rules rather than three.
# If a boundary marker is ever introduced, the third rule is a fixture and a
# branch here — not a redesign.
#
# NO LICENCE VALUE IS COMPILED INTO THE MEMBERSHIP TEST — it is read from the
# table, for the reason the cap numbers are read rather than restated: one home
# per fact. Values are NAMED in the comments below, where they are examples of
# what this gate can and cannot see; naming one there decides nothing, and the
# claim worth making is about the test rather than about the prose. The
# table is resolved from THIS SCRIPT'S OWN PATH and never from the caller's
# working directory, and `scripts/gates/tests/prose-licence-recall.sh` stages a
# table this repository does not have to prove that it is.
#
# LINE-ANCHORED, NOT A YAML READER, AND HERE IS WHICH AND WHY. A capped
# artifact is normally posted as a fenced YAML block inside a markdown comment,
# so the file handed to this gate is not a YAML document and a parser cannot be
# pointed at it without first deciding where the block begins and ends — the
# same undecidable boundary rule 3 dies on above. Worse, it would decide it
# DIFFERENTLY from `output-caps.sh`, which classifies on the first line matching
# `^artifact:`, fence or no fence; two gates disagreeing about which artifact
# they are looking at is a defect neither of them could report. The anchor here
# is therefore that same one, deliberately — NEARLY AND NOT EXACTLY, and the
# gap is named rather than claimed away. This anchor's whitespace class is
# `[ \t]` where `output-caps.sh:327` uses `[[:space:]]`, so on the three inputs
# that separate them — a vertical tab, a form feed, or a mid-record CR between
# `artifact:` and the id — that gate classifies the artifact and this one
# reports `unclassified` at rc 2. Measured over 9 fixtures: 6 agree, 3 diverge
# that way. The divergence is toward REFUSAL, which is the safe side of this
# instrument: an artifact it declines to classify is reported as unvouched and
# read by a human, where the failure that matters is passing one silently.
#
# WHAT A LINE ANCHOR THEN OWES, because a real YAML reader would have done it
# for free and part 1 of #171 was bitten by exactly this class:
#
#   - AN UNQUOTED `#` IS A COMMENT. `prose_licence: none   # nothing licensed`
#     declares `none`, and the value is cut at the first `#` that follows
#     whitespace or opens the line. A `#` with no space before it is part of
#     the value, which is YAML's rule and not a convenience.
#   - QUOTES ARE NOT PART OF THE VALUE. `"none"` and `'none'` declare `none`,
#     as a reader would have resolved them.
#   - CRLF IS STRIPPED FROM EVERY RECORD BEFORE ANYTHING IS MATCHED. GitHub
#     returns CRLF comment bodies — #186 defect 3 is that fact biting the cap
#     checker — and a `\r` carried into a value turns `none` into `none\r`,
#     which is in no enum. An artifact must not pass on one platform and fail
#     on the other. WHETHER THE STRIP DOES ANYTHING IS ITSELF HOST-DEPENDENT:
#     some awks hand `\r` to the program as part of the record and some drop it
#     before the program runs, which is that same platform split seen from this
#     side. On a host that drops it this line is a no-op that cannot be shown to
#     matter, so `prose-licence-recall.sh` probes the host and reports its red
#     demonstration for this rule as NOT RUN rather than passing it by accident
#     — measured on the reference host, Windows, 17 September 2026: this awk
#     does NOT carry `\r`, and that check did not run. The line stays because
#     the other kind of host exists and is where this gate runs in CI.
#
# WHAT IT DOES NOT REQUIRE, so that its scope is legible from its refusals: a
# CAP-TABLE ROW. Whether an id is capped is `output-caps.sh`'s question and it
# answers it; `escalation_packet` is a live id with no row (#186 § 4, AC 4 —
# "Every artifact id in live use has a cap-table row") and refusing it here
# would report #186's gap as this rule's violation. Anything declaring an
# `artifact:` id has its licence checked, rowed or not.
#
# WHAT IT CANNOT SEE:
#
#   - A DECLARATION WRAPPED IN MARKUP. `` `artifact: orchestrator_comment` ``
#     does not match `^artifact:`, so this gate reports no key exactly as
#     `output-caps.sh` does. That is #186 defect 2, it is filed, and it is
#     inherited here ON PURPOSE: fixing the anchor in one instrument and not
#     the other is how the two start classifying different artifacts.
#   - WHETHER THE LICENCE IS TRUE. `predecessor_corrected` on an artifact that
#     corrected no predecessor is a declaration this gate passes and a reader
#     catches. Declaring a licence is checkable; deserving one is not.
#   - HOW MANY SENTENCES THE LICENCE BOUGHT — rule 3, declined above.
#   - A SECOND `prose_licence:` KEY. A duplicate is first-wins and silent: the
#     `!has` guard means the second declaration is never observed, so
#     `none` followed by `ruling` reports `none · ok` at rc 0. Named here
#     because a list of what an instrument cannot see is worthless if it is
#     exhaustive about everything except the case a writer could exploit.
#   - A CR-ONLY ARTIFACT. Old-Mac line endings make the whole file one record,
#     so the scan reaches state 1 and EOF yields `no key follows artifact:` on
#     an artifact `output-caps.sh` classifies fine. The verdict is a false
#     FAILURE and not a false pass, so it is accepted here rather than fixed by
#     stripping CR globally: this gate's end-anchored strip is the CRLF rule
#     above, whose scope is the line ending GitHub actually returns.
#
# ONE VALUE IT REFUSES TO JUDGE RATHER THAN FAIL, and the refusal is the
# honest answer instead of a casting vote. A value naming two licences at once
# — `prose_licence: finding_overturned, predecessor_corrected`, live on PR #163
# at comments 5555418634 and 5556183187 — sits in a gap in the rule, not in a
# gap in this instrument. The table's value column is one of five tokens, which
# makes the pair invalid; the sentence beneath it budgets the licensed prose at
# five sentences "however many of the four apply at once", which reads as
# though more than one may. Failing it decides that question from a gate, and
# passing it is the silence this file exists to end. It is therefore named and
# reported as cannot-vouch, which is never clean, and #171's PR carries it to
# `@engineering-lead` as the one reconciliation this gate cannot make for
# itself.
#
# Run it with `make prose-licence ARTIFACTS="<path>..."`, which runs the recall
# corpus FIRST and only then this check — the ordering `make output-caps` uses
# and for the same reason.
#
# Usage: scripts/gates/prose-licence.sh <path>...
# Exit:  0  every artifact declares a value from the table as its second key
#        1  at least one does not; each is printed with what failed
#        2  cannot vouch — a path it could not read, zero artifacts, an
#           artifact declaring no `artifact:` key, a value naming more than one
#           licence, or a licence table it could not find, could not parse, or
#           which holds a row-shaped line that does not parse. "I could not
#           judge you" and "you declared the wrong thing" are different
#           sentences and only one of them is the author's fault, so they are
#           different statuses. Cannot-vouch outranks a failed declaration when
#           both occur, and the outranked line is still printed: precedence is
#           about the status, never about stopping looking.
#           NEVER REPORTED AS CLEAN, and never 0 without having checked
#           something (#184) — the sentinel below is emitted on one path only.

set -uo pipefail

DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || DIR=''
TABLE="$DIR/../../.claude/skills/agent-handoffs/SKILL.md"

checked=0; failed=0; unclassified=0; cannot=0

# The summary prints on EVERY path, including every cannot-vouch one, for the
# reason `local-gates § Artifact caps` gives: a check that did not run and a
# check that found nothing must not print the same thing. The sentinel is
# emitted only when something was checked and nothing failed or was refused, so
# a broken instrument cannot print what a clean run prints.
finish() {
  printf 'licence · %s checked · %s failed · %s unclassified\n' \
    "$checked" "$failed" "$unclassified"
  [ "$cannot" -eq 0 ] || exit 2
  [ "$failed" -eq 0 ] || exit 1
  # THE LAST GUARD BEFORE THE SENTINEL, and it is deliberately unreachable by
  # any path above: every branch that judges nothing already sets `cannot`. It
  # stands because #184 is about the ONE value that turns a failure to check
  # into a clean line, and a zero that arrives by a route nobody anticipated
  # must not be the route that prints `clean`. It says why rather than exiting
  # mutely, so that if it ever does fire, the reader is not left guessing.
  if [ "$checked" -le 0 ]; then
    echo "prose-licence: cannot vouch — nothing was checked, and a run that judged no declaration is not a clean run" >&2
    exit 2
  fi
  printf 'clean · every artifact declares a licence the table defines as its second key\n'
  exit 0
}

if [ -z "$DIR" ]; then
  echo "prose-licence: cannot vouch — this script could not resolve its own directory, so the licence table cannot be located" >&2
  cannot=1; finish
fi
if [ ! -f "$TABLE" ] || [ ! -r "$TABLE" ]; then
  printf 'prose-licence: cannot vouch — the licence table is not readable at %s\n' "$TABLE" >&2
  cannot=1; finish
fi

# THE LICENCE TABLE'S BINDING CONTRACT, and it is asserted here because the
# table does not yet assert it of itself. `agent-handoffs § The cap table`
# fixes a form for the FOUR-column cap table and says in the same breath that
# "the two-column `prose_licence` table below is outside that region and is not
# judged against a contract that was never its own" — true of that instrument
# and no longer true of the file, which this gate now machine-reads. The
# sentence that fixes the two-column form belongs in `agent-handoffs`, which
# #171 holds read-only on this branch; until it lands, the contract lives here
# and the PR carries it as owed.
#
# The form: two columns, the first a code span holding one snake_case token.
# The header anchor is the first column's own name rather than the whole header
# line, because the second column is prose and anchoring on prose is how an
# instrument breaks on a wording change that means nothing.
#
# ROWS ARE FOUND BY POSITION, not by whether they parse, for the reason
# `agent-handoffs § The cap table` gives about its own region: a row-shaped
# line that fails the form is a defect in the table and is NAMED, because
# dropping it silently un-defines a value while leaving every report clean.
LICTEXT="$(awk '
  /^\| `prose_licence` \|/          { in_table = 1; next }
  in_table && /^\|[-: |]+\|[ \t]*$/ { next }
  in_table && /^\|/ {
    if ($0 ~ /^\| `[a-z_]+` \| [^|]*\|[ \t]*$/) {
      split($0, f, "|"); a=f[2]; gsub(/[` ]/, "", a); printf "+ %s\n", a
    } else printf "- %s\n", $0
    next
  }
  in_table { in_table = 0 }' < "$TABLE")"

BADROWS="$(printf '%s\n' "$LICTEXT" | sed -n 's/^- //p')"
VALUES="$(printf '%s\n' "$LICTEXT" | sed -n 's/^+ //p')"

if [ -n "$BADROWS" ]; then
  printf 'prose-licence: cannot vouch — a row of the licence table in %s does not parse under its own binding contract:\n' "$TABLE" >&2
  printf '%s\n' "$BADROWS" | while IFS= read -r bad; do
    printf 'prose-licence:   %s\n' "$bad" >&2
  done
  cannot=1; finish
fi
if [ -z "$VALUES" ]; then
  printf 'prose-licence: cannot vouch — no row in %s parses under the binding contract of the licence table\n' "$TABLE" >&2
  cannot=1; finish
fi

# THE CANDIDATE REACHES awk THROUGH THE ENVIRONMENT AND NOT THROUGH `-v`, and
# the difference is the whole of whether this enum can fail open. A `-v`
# assignment is processed as a string literal, so awk decodes backslash escapes
# in it before the program compares anything: `prose_licence: non\145` arrived
# at the comparison already decoded to `none` and the gate reported `ok` at
# rc 0 on a value no row defines. `ENVIRON[]` is that same value uninterpreted.
# An enum that fails open is worse than no enum, because it reports clean while
# deciding nothing, and #171 AC1's own words are that it cannot fail open.
defined() { printf '%s\n' "$VALUES" | V="$1" awk '$1 == ENVIRON["V"] { found = 1 } END { exit !found }'; }

# THE PARSER. It reads an artifact and reports one OBSERVATION; it holds no
# enum, makes no verdict, and the shell below does both against the table it
# read. Splitting them that way is what keeps the enum in one place.
#
# `id` is extracted by the same rule `output-caps.sh` extracts it by, on the
# first `^artifact:` line and no later one: a later declaration belongs to
# something the artifact relays verbatim and never reclassifies its courier.
# The scan for the second key stops at a relayed `artifact:` for the same
# reason — a relay's own licence is not this artifact's.
SCAN='
function trim(v)  { sub(/^[ \t]+/, "", v); sub(/[ \t]+$/, "", v); return v }
function uncomment(v) {
  # A YAML comment opens at a `#` that begins the value or follows whitespace.
  # A `#` with no space before it is part of the value.
  if (match(v, /(^|[ \t])#/)) v = substr(v, 1, RSTART - 1)
  return v
}
function unquote(v) {
  if (length(v) > 1 && (v ~ /^".*"$/ || v ~ /^'"'"'.*'"'"'$/))
    v = substr(v, 2, length(v) - 2)
  return v
}
{ line = $0; sub(/\r$/, "", line) }

state == 0 {
  if (line ~ /^artifact:[ \t]*[A-Za-z0-9_]/) {
    id = line
    sub(/^artifact:[ \t]*/, "", id)
    sub(/[^A-Za-z0-9_].*$/, "", id)
    state = 1
  }
  next
}

# DIRECTLY AFTER means the next KEY, and a blank line and a comment line are
# not keys. Anything else that is not a key ends the block, and an artifact
# whose block held one key has no second one to check.
state == 1 {
  if (line ~ /^[ \t]*$/) next
  if (line ~ /^[ \t]*#/)  next
  state = 2
  if (line ~ /^prose_licence[ \t]*:/) {
    has = 1
    val = line
    sub(/^prose_licence[ \t]*:/, "", val)
    val = unquote(trim(uncomment(val)))
  } else if (line ~ /^[A-Za-z_][A-Za-z0-9_-]*[ \t]*:/) {
    second = line
    sub(/[ \t]*:.*$/, "", second)
  } else {
    second = ""
  }
  next
}

state == 2 {
  if (line ~ /^artifact:[ \t]*[A-Za-z0-9_]/) { state = 3; next }
  if (!has && line ~ /^prose_licence[ \t]*:/) later = 1
  next
}

END {
  if (state == 0)   { print "NOARTIFACT\n\n";          exit }
  if (state == 1)   { print "NOSECOND\n" id "\n";      exit }
  if (has)          { print "VALUE\n" id "\n" val;     exit }
  if (second == "") { print "NOTAKEY\n" id "\n";       exit }
  if (later)        { print "LATER\n" id "\n" second;  exit }
  print "OTHERKEY\n" id "\n" second
}'

if [ "$#" -eq 0 ]; then
  echo "prose-licence: cannot vouch — no artifact path was given, and a set of zero artifacts is not a clean set" >&2
  cannot=1; finish
fi

for f in "$@"; do
  if [ ! -f "$f" ] || [ ! -r "$f" ]; then
    printf 'unclassified · unreadable path · %s\n' "$f"
    unclassified=$((unclassified + 1)); cannot=1; continue
  fi

  # Fed on stdin and never as an operand, for the reason `output-caps.sh`
  # records: awk consumes a bare operand matching `name=value` as a variable
  # assignment, reads stdin instead, and exits 0 having done what it was told.
  obs="$(LC_ALL=C awk "$SCAN" < "$f")" || obs=''
  # A PARSER THAT DIES WRITES NOTHING, and nothing must not read as an artifact
  # that declared nothing. `output-caps.sh` records this shape going live in its
  # own filter — a program that would not parse produced a clean line — so the
  # empty observation is a refusal here rather than a fallthrough.
  if [ -z "$obs" ]; then
    printf 'unclassified · could not parse this artifact · %s\n' "$f"
    unclassified=$((unclassified + 1)); cannot=1; continue
  fi

  # Three lines, not three tab-separated fields. Tab is IFS WHITESPACE, so
  # `IFS=<tab> read -r a b c` collapses a run of tabs into one delimiter and an
  # empty middle field silently shifts the value into the id.
  tag=''; id=''; val=''
  { IFS= read -r tag || tag=''
    IFS= read -r id  || id=''
    IFS= read -r val || val=''; } <<EOF
$obs
EOF

  case "$tag" in
    NOARTIFACT)
      printf 'unclassified · no artifact: key · %s\n' "$f"
      unclassified=$((unclassified + 1)); cannot=1; continue ;;
    NOSECOND)
      printf '%s · no key follows artifact: · fail · %s\n' "$id" "$f"
      checked=$((checked + 1)); failed=$((failed + 1)); continue ;;
    NOTAKEY)
      printf '%s · the line after artifact: is not a key, so prose_licence: is not the second one · fail · %s\n' "$id" "$f"
      checked=$((checked + 1)); failed=$((failed + 1)); continue ;;
    LATER)
      printf '%s · prose_licence: is declared but %s: stands second · fail · %s\n' "$id" "$val" "$f"
      checked=$((checked + 1)); failed=$((failed + 1)); continue ;;
    OTHERKEY)
      printf '%s · no prose_licence: key; %s: stands second · fail · %s\n' "$id" "$val" "$f"
      checked=$((checked + 1)); failed=$((failed + 1)); continue ;;
  esac

  # A key with nothing after it. `agent-handoffs § Prose is licensed, and the
  # artifact names its licence` settles this one outright: "`none` is a value
  # and not an omission", so an empty value is an omission wearing a key and
  # is the case that sentence was written against.
  if [ -z "$val" ]; then
    printf '%s · prose_licence: declares an empty value, and none is a value rather than an omission · fail · %s\n' "$id" "$f"
    checked=$((checked + 1)); failed=$((failed + 1)); continue
  fi

  # Two licences at once — refused rather than judged, per the header.
  case "$val" in
    *,*)
      printf 'unclassified · prose_licence: names more than one licence (%s), which the table does not define and this gate does not decide · %s\n' "$val" "$f"
      unclassified=$((unclassified + 1)); cannot=1; continue ;;
  esac

  if defined "$val"; then
    printf '%s · %s · ok · %s\n' "$id" "$val" "$f"
    checked=$((checked + 1))
  else
    printf '%s · %s is not a value the licence table defines · fail · %s\n' "$id" "$val" "$f"
    checked=$((checked + 1)); failed=$((failed + 1))
  fi
done

finish
