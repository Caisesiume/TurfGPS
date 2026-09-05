#!/usr/bin/env bash
# output-caps.sh — is a capped artifact within the cap its row sets? (#158)
#
# The enforcement point for `agent-handoffs § Output caps`. Every capped
# artifact is capped there, and the rule for counting each one is there; this
# file holds neither. It reads the table, looks up the id the artifact
# declared, counts characters, and compares. Adding a capped artifact is a row
# there, never a change here.
#
# That claim is narrower than "the one home of every cap", which this line used
# to make and which is false: the ~300-token handoff limit and the per-field
# caps both sit in `agent-handoffs` OUTSIDE `§ Output caps`. Only the artifacts
# this instrument measures are capped there.
#
# WHAT IT KNOWS ABOUT AGENTS: NOTHING, and that is a property to keep rather
# than a stage it is at. An artifact self-identifies through the mandatory
# `artifact:` key that `agent-handoffs § The structured block comes first`
# requires, and this script reads the FIRST such line in the file. The first is
# the artifact's own declaration; a later one belongs to something it relays
# verbatim, is counted against its own row where that was posted, and never
# reclassifies its courier. Nothing here infers an author from a wording, a
# filename or a lane. An instrument that guessed would be a semantic classifier,
# which is the direction #139 rejected, and a character count is worth trusting
# precisely because it attempts nothing of the kind.
#
# THE TABLE IS RESOLVED FROM THIS SCRIPT'S OWN PATH and never from the caller's
# working directory. The gate must measure against the same table wherever it is
# typed from, and `scripts/gates/tests/output-caps-recall.sh` stages a table this
# repository does not have to prove that it does.
#
# CHARACTERS, NOT BYTES. `agent-handoffs § Output caps` sets the unit and gives
# the reason: a writer can count characters. The distinction is not pedantic —
# the three judgments on PR #141 sit a few dozen characters either side of their
# cap and all three sit above it in bytes, so a byte count flags the one that
# was under and the discrimination this instrument exists for is gone. That
# corpus records both numbers for each of the five historical artifacts and the
# one command that re-derives them; they are not restated here.
#
# Every count below is therefore BYTES MINUS UTF-8 CONTINUATION BYTES, taken
# under LC_ALL=C. `wc -m` and awk's `length()` both answer in whatever the
# ambient locale says — bytes under C, codepoints under a UTF-8 locale — and a
# gate whose verdict depends on the caller's LANG is not a gate. Subtracting
# continuation bytes is locale-independent and exact for UTF-8, which is what
# this repository writes.
#
# WHAT `whole`, `body` AND `own` EXCLUDE is defined in
# `agent-handoffs § The cap table`; the filter below implements that definition
# and does not restate it. What is recorded here is the implementation's own
# edges, which the table does not decide and a reader cannot infer from it:
#
#   - a findings row is a line whose FIRST character is `|`. An indented one is
#     ordinary text, because a table being quoted inside prose is prose.
#   - a blank line inside a `findings:` block decides nothing and stays inside
#     it. The block's boundaries are the table's rule and are not restated here;
#     what the table does not decide is what an empty line between two findings
#     means, and treating one as a boundary would end the block there and resume
#     counting the rest of it.
#   - a fence is a line whose first non-blank characters are three backticks,
#     and it is excluded with the block it delimits. A rule that counted the
#     delimiter of the thing it excludes would be measuring punctuation.
#   - the final newline is a character in the file and is counted. A character
#     count has no reason to hold an opinion about which character is last, and
#     a file that does not end in one is measured without inventing it.
#   - A MEASUREMENT OF ZERO IS NOT A MEASUREMENT, and is cannot-run. This is the
#     one refusal below that the table does not reach, so it is argued here.
#     Zero is under every cap in the table, which makes it the single value that
#     turns a failure to measure into a clean line — and it is reachable with no
#     failure at all: an `own` artifact posted entirely inside one fence, the
#     canonical shape `agent-handoffs § The structured block comes first`
#     prescribes, measures exactly zero. This instrument refused a set of zero
#     ARTIFACTS from the start; it refuses zero CHARACTERS for the same reason.
#
# TWO FURTHER REFUSALS BELOW ARE THE TABLE'S RULES AND NOT THIS FILE'S, and are
# implemented rather than argued here: an `own` artifact whose fences do not
# balance is unmeasurable, and a row-shaped line in the cap table that does not
# parse is named rather than skipped. `agent-handoffs § The cap table` carries
# both, with the reasoning; what is written here would go false the day either
# one moved, which is the whole reason it is not written here.
#
# WHAT IT CANNOT SEE, so that no reader takes a clean line for more than it is:
#
#   - `own` measures a floor rather than the whole answer, for the reason
#     `agent-handoffs § The cap table` gives where it says so. An artifact that
#     fences ordinary prose measures small here and is under-reported, and it is
#     a reader who catches that, not this script.
#   - A COMPOSITE IS NOT SPLIT. A notice standing over a judgment standing over
#     a revision packet, posted under one declaration, is measured whole under
#     that one id — one line, not three. Detection holds; attribution does not,
#     and it is out of scope deliberately: apportioning one comment between
#     three artifacts requires knowing where each began, which is the semantic
#     judgement above that this instrument makes none of. Attribution is
#     separate work, not a widening of this file.
#   - A cap is a length. Whether the prose inside an artifact was licensed is a
#     question about meaning: `prose_licence:` is declared for a reader and is
#     not measured here.
#
# Run it with `make output-caps ARTIFACTS="<path>..."`, which runs the recall
# corpus FIRST and only then this check. A verdict from an instrument whose
# recall was not just demonstrated is the thing that target exists to refuse.
#
# Usage: scripts/gates/output-caps.sh <path>...
# Exit:  0  every artifact measured is within the cap its row sets
#        1  at least one is over it; each is printed with what it measured
#        2  cannot run — an artifact declaring no id, an id no row defines, a
#           path it could not read, zero artifacts measured, an artifact
#           measuring zero characters, an `own` artifact whose fences do not
#           balance, or a cap table it could not find, could not parse, or
#           holds a row-shaped line that does not parse. "I could not measure
#           you" and "you wrote too much" are different sentences and only one
#           of them is the author's fault, so they are different statuses.
#           Cannot-run outranks
#           over-cap when both occur, and the outranked line is still printed:
#           precedence is about the status, never about stopping looking.
#           Never reported as clean.

set -uo pipefail

DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || DIR=''
TABLE="$DIR/../../.claude/skills/agent-handoffs/SKILL.md"

measured=0; over=0; unclassified=0; cannot=0

# The summary is printed on EVERY path, including every cannot-run one. "The
# check ran and found nothing" and "the check did not run" must not print the
# same thing, and the sentinel below is what separates them: it is emitted only
# when something was measured and nothing was over or unclassified, so a broken
# instrument cannot print what a clean run prints.
finish() {
  printf 'caps · %s measured · %s over · %s unclassified\n' "$measured" "$over" "$unclassified"
  [ "$cannot" -eq 0 ] || exit 2
  [ "$over" -eq 0 ]   || exit 1
  printf 'clean · every artifact measured is within the cap its row sets\n'
  exit 0
}

if [ -z "$DIR" ]; then
  echo "output-caps: cannot run — this script could not resolve its own directory, so the cap table cannot be located" >&2
  cannot=1; finish
fi
if [ ! -f "$TABLE" ] || [ ! -r "$TABLE" ]; then
  printf 'output-caps: cannot run — the cap table is not readable at %s\n' "$TABLE" >&2
  cannot=1; finish
fi

# The table's own binding contract, as `agent-handoffs § The cap table` fixes
# it: three code spans, one snake_case id, a bare integer, one lowercase token.
#
# WHICH LINES ARE ROWS IS DECIDED BY POSITION, not by whether they parse. The
# table is entered at its header and left at the first line that is not a table
# line, and every line between them other than the alignment separator is a row
# this script must be able to read. That is what lets a row-shaped line that
# does NOT parse be named — a fail-open where it is dropped instead is the whole
# of MAINT-01, and it is invisible from the table's own file. Deciding by the
# contract alone cannot do it: the two-column `prose_licence` table below fails
# the same contract and is not a broken row, and the header row fails it too.
#
# Each emitted line is tagged `+` for a row that parsed and `-` for one that did
# not, because command substitution returns one stream and the two answers must
# stay separable within it.
ROWTEXT="$(awk '
  /^\| `artifact` \| `cap_chars` \| `counts` \|/ { in_table = 1; next }
  in_table && /^\|[-: |]+\|[ \t]*$/             { next }
  in_table && /^\|/ {
    if ($0 ~ /^\| `[a-z_]+` \| `[0-9]+` \| `[a-z]+` \|/) {
      split($0, f, "|"); a=f[2]; b=f[3]; c=f[4]
      gsub(/[` ]/, "", a); gsub(/[` ]/, "", b); gsub(/[` ]/, "", c)
      printf "+ %s %s %s\n", a, b, c
    } else printf "- %s\n", $0
    next
  }
  in_table { in_table = 0 }' < "$TABLE")"

BADROWS="$(printf '%s\n' "$ROWTEXT" | sed -n 's/^- //p')"
ROWS="$(printf '%s\n' "$ROWTEXT" | sed -n 's/^+ //p')"

if [ -n "$BADROWS" ]; then
  printf 'output-caps: cannot run — a row of the cap table in %s does not parse under its own binding contract:\n' "$TABLE" >&2
  printf '%s\n' "$BADROWS" | while IFS= read -r bad; do
    printf 'output-caps:   %s\n' "$bad" >&2
  done
  cannot=1; finish
fi
if [ -z "$ROWS" ]; then
  printf 'output-caps: cannot run — no row in %s parses under the binding contract of the cap table\n' "$TABLE" >&2
  cannot=1; finish
fi

cap_of()    { printf '%s\n' "$ROWS" | awk -v i="$1" '$1 == i { print $2 }'; }
counts_of() { printf '%s\n' "$ROWS" | awk -v i="$1" '$1 == i { print $3 }'; }

# The line filter. It emits the counted region of an artifact and nothing else,
# so the counting below is one arithmetic rule over whatever it is handed.
#
# Printing is deferred by one record so that END knows which line was last:
# where a file does not end in a newline, re-adding one to its final line would
# measure a character the file does not hold.
FILTER='
function indent_of(s) { if (match(s, /^[ \t]*/)) return RLENGTH; return 0 }
{
  keep = 1
  if (mode == "body") {
    if (in_findings) {
      # Membership, not termination. A line deeper than the key is inside the
      # block; a `-` item at that same indent is inside it; a blank line is
      # held inside, so that one empty line between two findings is not a
      # boundary. Anything else is outside, and falls through to be counted.
      if ($0 ~ /^[ \t]*$/) keep = 0
      else if (indent_of($0) > find_indent) keep = 0
      else if (indent_of($0) == find_indent && $0 ~ /^[ \t]*-([ \t]|$)/) keep = 0
      else in_findings = 0
    }
    if (keep) {
      if ($0 ~ /^\|/) keep = 0
      else if ($0 ~ /^[ \t]*findings[ \t]*:/) { in_findings = 1; find_indent = indent_of($0); keep = 0 }
    }
  } else if (mode == "own") {
    if ($0 ~ /^[ \t]*```/) { keep = 0; in_fence = !in_fence }
    else if (in_fence) keep = 0
  }
  if (seen && prev_keep) printf "%s\n", prev
  prev = $0; prev_keep = keep; seen = 1
}
END {
  if (seen && prev_keep) { if (nonl == 1) printf "%s", prev; else printf "%s\n", prev }
}'

digits_of() { local s="$1"; printf '%s' "${s//[!0-9]/}"; }

# EVERY FILE IS FED ON STDIN AND NEVER AS AN OPERAND. awk consumes a bare
# operand matching `name=value` as a variable assignment and reads stdin
# instead, so a path such as `v=1.md` — which `make output-caps ARTIFACTS=...`
# can deliver — measured 0 characters and reported clean with stdin closed, and
# hung the gate forever with stdin open. `pipefail` cannot see either: awk
# exited 0, having done exactly what it was told. sed and tail take the same
# operand and, path-leading `-` aside, want the same treatment for the same
# reason — the redirection is the shell resolving the path, which is the one
# place in this pipeline that cannot reinterpret it.
fences_of() { # fences_of <path> — how many fence lines the file holds
  LC_ALL=C awk '/^[ \t]*```/ { n++ } END { print n + 0 }' < "$1"
}

measure() { # measure <path> <counts token> — prints characters, or nothing
  local f="$1" mode="$2" nonl=0 b c
  # EVERY PIPELINE'S STATUS IS TESTED, and `set -o pipefail` above is what makes
  # testing the last one enough. A filter that dies — an awk that will not parse
  # this program, a file it cannot read — writes nothing, and nothing counts as
  # zero characters, which is under every cap in the table. That is the shape
  # this whole instrument exists to refuse: a measurement that was never made,
  # reported as a clean one. It was live here until a mutant of this script's own
  # line filter produced exactly it.
  if [ "$mode" = whole ]; then
    b="$(LC_ALL=C wc -c < "$f")" || return 1
    c="$(LC_ALL=C tr -dc '\200-\277' < "$f" | LC_ALL=C wc -c)" || return 1
  else
    if [ -s "$f" ]; then
      case "$(LC_ALL=C tail -c 1 < "$f" | LC_ALL=C od -An -to1 | tr -d '[:space:]')" in
        012) ;;
        *) nonl=1 ;;
      esac
    fi
    b="$(LC_ALL=C awk -v mode="$mode" -v nonl="$nonl" "$FILTER" < "$f" | LC_ALL=C wc -c)" || return 1
    c="$(LC_ALL=C awk -v mode="$mode" -v nonl="$nonl" "$FILTER" < "$f" | LC_ALL=C tr -dc '\200-\277' | LC_ALL=C wc -c)" || return 1
  fi
  b="$(digits_of "$b")"; c="$(digits_of "$c")"
  [ -n "$b" ] && [ -n "$c" ] || return 1
  printf '%s' $(( b - c ))
}

if [ "$#" -eq 0 ]; then
  echo "output-caps: cannot run — no artifact path was given, and a set of zero artifacts is not a clean set" >&2
  cannot=1; finish
fi

for f in "$@"; do
  if [ ! -f "$f" ] || [ ! -r "$f" ]; then
    printf 'unclassified · unreadable path · %s\n' "$f"
    unclassified=$((unclassified + 1)); cannot=1; continue
  fi

  # The FIRST declaration, and only it. `head -1` is the whole of that rule.
  id="$(sed -n 's/^artifact:[[:space:]]*\([A-Za-z0-9_][A-Za-z0-9_]*\).*$/\1/p' < "$f" | head -1)"
  if [ -z "$id" ]; then
    printf 'unclassified · no artifact: key · %s\n' "$f"
    unclassified=$((unclassified + 1)); cannot=1; continue
  fi

  cap="$(cap_of "$id")"
  mode="$(counts_of "$id")"
  if [ -z "$cap" ] || [ -z "$mode" ]; then
    printf 'unclassified · no cap-table row for %s · %s\n' "$id" "$f"
    unclassified=$((unclassified + 1)); cannot=1; continue
  fi
  case "$mode" in
    whole|body|own) ;;
    *) printf 'unclassified · the row for %s counts %s, which this instrument does not implement · %s\n' "$id" "$mode" "$f"
       unclassified=$((unclassified + 1)); cannot=1; continue ;;
  esac

  # An unbalanced fence is refused BEFORE anything is measured, because the
  # number a toggle produces from one is not a floor, a ceiling or an estimate —
  # it is the measurement of whichever region the inversion happened to select.
  if [ "$mode" = own ]; then
    nf="$(fences_of "$f")"; nf="$(digits_of "$nf")"
    if [ -z "$nf" ] || [ $(( nf % 2 )) -ne 0 ]; then
      printf 'unclassified · %s fence lines do not balance, so the counted region is not decidable · %s\n' "${nf:-unreadable}" "$f"
      unclassified=$((unclassified + 1)); cannot=1; continue
    fi
  fi

  n="$(measure "$f" "$mode")"
  case "$n" in
    ''|*[!0-9]*)
      printf 'unclassified · could not count the characters of this artifact · %s\n' "$f"
      unclassified=$((unclassified + 1)); cannot=1; continue ;;
  esac

  # Zero is under every cap in the table, so it is the one value that turns a
  # failure to measure into a clean line. This instrument refused a set of zero
  # artifacts from the start; it refuses zero characters for the same reason.
  if [ "$n" -eq 0 ]; then
    printf 'unclassified · measured zero characters, which is not a measurement this gate will vouch for · %s\n' "$f"
    unclassified=$((unclassified + 1)); cannot=1; continue
  fi

  if [ "$n" -gt "$cap" ]; then verdict=over; over=$((over + 1)); else verdict=under; fi
  printf '%s · %s · %s · %s · %s\n' "$id" "$n" "$cap" "$verdict" "$f"
  measured=$((measured + 1))
done

finish
