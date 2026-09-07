#!/usr/bin/env bash
# owed-work-stubs.sh — the RED DEMONSTRATION for owed-work-recall.sh. (#172, #178)
#
# `docs/DELIVERY.md § Proof that a test can fail`: a test only earns its place if
# it would fail on the bug it names. `owed-work-recall.sh` being green against
# `owed-work.sh` PROVES NOTHING — a corpus that asserted nothing would be just as
# green. This script is what makes the green mean something.
#
# HOW THE DEMONSTRATION IS MADE HONEST. Five detectors are generated from ONE
# template, differing in a single `DEFECT` flag. The first has no defect and must
# go GREEN; the other four each reproduce a documented failure and must go RED.
# Generating them from one template is the whole method.
# `docs/DELIVERY.md § Red for the wrong reason` excludes a demonstration that is
# red because something failed to build or never reached its assertion, and four
# independently-written stubs would each be red for reasons nobody could
# attribute to a defect. Here the four differ from a detector that PASSES THE WHOLE
# CORPUS by one behaviour each, so every red is that behaviour and nothing else.
#
# ---------------------------------------------------------------------------
# THE `none` STUB IS A SECOND IMPLEMENTATION, AND CYCLE 1 SHOWED WHAT HAPPENS
# WHEN IT IS ONLY NOMINALLY ONE
# ---------------------------------------------------------------------------
# @pr-judge's finding on cycle 1 was that this stub's B1 was "logically identical
# to the line under test" — it asked "is there a judgment newer than the ledger",
# which was the detector's own expression transcribed. A control that computes the
# same thing the same way cannot tell you the thing is wrong. So the B1 here is
# written from the criteria and DELIBERATELY BY ANOTHER METHOD:
#
#   the detector      collects each ruling's `sha`/`cycle` into a list and walks it,
#                     comparing shas by testing whether either is a PREFIX of the
#                     other after lower-casing and hex-validating both.
#   this stub         CANONICALISES each identity to a single key — the first seven
#                     hex characters, lower-cased, joined to the cycle read as a
#                     number — and asks one set-membership question. No list, no
#                     walk, no prefix test.
#
# THE TWO METHODS ARE NOT EQUIVALENT, AND SAYING SO IS THE POINT. Two shas that
# share seven characters and differ at the eighth are ONE identity to this stub and
# TWO to the detector. No fixture in the corpus contains that pair, so the two agree
# everywhere they are measured — but they are genuinely different programs, and the
# divergence is recorded rather than hidden. Which reading is correct is a contract
# question (`review-board-dispatch § The claim table` on how a panel key is
# spelled) and is not settled here.
#
# THE OTHER SHARED PREDICATE IS GONE ENTIRELY. The marker extraction used to live
# in an awk twin inside the corpus and a second one here, with the `validation-agent`
# word-boundary regex standing character-identical in three files. The corpus now
# declares each comment's record as DATA; this template asks for the declared
# records or the heading-derived ones and computes neither. The regex now exists in
# the detector alone, and the `prose` defect below is one query name.
#
# WHAT THE `none` STUB IS FOR, beyond being a control: it is what distinguishes
# "the detector is wrong" from "the corpus over-specifies" when the two disagree.
# A corpus no contract-faithful detector can satisfy is a broken corpus. A corpus
# this stub satisfies and the deliverable does not is a finding against the
# deliverable.
#
# THE FOUR DEFECTS, and the failure each one reproduces:
#   two_state       folds `undeclared` into `clear` and never exits 3. Measured
#                   2026-09-07: all four open PRs carried ZERO declared
#                   revision_packet/judgment comments, so this detector prints
#                   #135 — which held an undischarged cycle-4 packet — as nothing
#                   owed. #172's own bug, reproduced by its fix.
#   prose           identifies artifacts from HEADING TEXT instead of the declared
#                   `artifact:` key (#172 criterion 2), by asking the transport for
#                   the heading-derived record. This is the 2026-09-06 survey's two
#                   ad-hoc greps: #147 called "zero judge ledgers" when it had four,
#                   #154 "no packet found" when a packet existed and had been worked.
#   quiet_fail      exits 0 when a source is unreadable (#172 criterion 4, the
#                   `fingerprint.sh` precedent): an unreadable source is not a
#                   quiet queue.
#   b2_discharged   lets a ruling close the validation question (#178 criterion 3):
#                   the mandatory last lane silently skipped, which is PR #135
#                   cycles 4 and 5 and PR #154 cycle 6. Its red used to rest on ONE
#                   table cell, and on an ordering the posting contract makes
#                   impossible; the corpus now isolates that defect on a fixture
#                   whose only owed class is validation and asserts the cell, the
#                   state token, the exit status and the summary counts.
#
# Hermetic: everything generated under mktemp, nothing written outside it, no
# network, no repository state. The corpus it drives is hermetic in its own right.
# Usage: scripts/loop/tests/owed-work-stubs.sh
# Exit:  0 the `none` stub passed AND all four defective stubs were caught
#        1 any stub behaved otherwise — which is a corpus that does not discriminate

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
CORPUS="$DIR/owed-work-recall.sh"

TMP="$(mktemp -d)" || TMP=''
[ -n "$TMP" ] && [ -d "$TMP" ] || {
  printf 'FAIL  mktemp -d gave no usable directory; refusing to run\n'; exit 1; }
trap 'rm -rf "$TMP"' EXIT

[ -f "$CORPUS" ] || { printf 'FAIL  construction · the corpus is not at %s\n' "$CORPUS"; exit 1; }

# ---------------------------------------------------------------------------
# THE TEMPLATE. One contract-faithful detector; `__DEFECT__` is substituted per
# stub and is the ONLY difference between them.
# ---------------------------------------------------------------------------
cat > "$TMP/template" <<'TEMPLATE'
#!/usr/bin/env bash
# Generated by owed-work-stubs.sh. Not a deliverable — a control.
set -u
DEFECT="__DEFECT__"
GH="${GH:-gh}"
R='repos/{owner}/{repo}'

usage() { echo "usage: owed-work.sh [pr-number ...]" >&2; }

if [ "$#" -gt 0 ]; then
  for a in "$@"; do
    case "$a" in ''|*[!0-9]*) usage; exit 2 ;; esac
  done
  PRS="$*"
else
  PRS="$("$GH" pr list --state open --limit 200 --json number --jq '.[].number' 2>/dev/null)"; lrc=$?
  if [ "$lrc" -ne 0 ]; then
    echo "owed_work: 0 PR(s) read"
    echo "summary: owed 0 · clear 0 · undeclared 0 · unreadable 1"
    exit 2
  fi
fi

# DEFECT: identify artifacts by HEADING TEXT rather than by the declared key. The
# transport serves both readings, so the defect is which one is asked for.
QUERY=DECLARED
[ "$DEFECT" = prose ] && QUERY=HEADING

n_owed=0; n_clear=0; n_undecl=0; n_unread=0; n_read=0
TABLE=''; OWED=''; UNDECL=''; UNREAD=''

# The gap in whole days. `date -u -d` is a GNU extension and the deliverable
# deliberately avoids it in favour of civil-calendar arithmetic; a control is free
# to use it, and it being a different computation is the point.
secs() { date -u -d "$1" +%s 2>/dev/null || echo 0; }
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)"

for pr in $PRS; do
  recs="$("$GH" api "$R/issues/$pr/comments" --paginate --jq "$QUERY" 2>/dev/null)"; crc=$?
  dates="$("$GH" api "$R/pulls/$pr/commits" --paginate --jq '.[].commit.committer.date' 2>/dev/null)"; drc=$?

  if [ "$crc" -ne 0 ] || [ "$drc" -ne 0 ] || [ -z "$dates" ]; then
    n_unread=$((n_unread + 1))
    TABLE="$TABLE
$(printf '#%-5s %-11s %s' "$pr" "unreadable" "a source could not be read")"
    UNREAD="$UNREAD
  #$pr a source could not be read"
    continue
  fi
  n_read=$((n_read + 1))

  # The API returns commits oldest first, so the newest is the last line. ISO-8601
  # Z timestamps compare correctly as strings, which is why `date` is reached for
  # only when a gap must be printed.
  newest_commit="${dates##*$'\n'}"

  # ONE pass for every derived value. B1 is decided HERE, by canonical-key set
  # membership — see the header for why it is not the detector's method:
  #   ident(sha, cycle) = first seven hex characters, lower-cased, "@", cycle as a
  #                       number. A sha shorter than seven characters, or a cycle
  #                       that is not digits, HAS NO IDENTITY and is never keyed.
  # The five B1 answers, derived from #178 criterion 2 and from the third state the
  # source classes carry:
  #   no panel                          -> undeclared (nothing convened to owe)
  #   panel with no identity            -> undeclared (no key to look up)
  #   panel, and NO ruling of any kind  -> OWED — "a convened panel with no ruling",
  #                                        which is #178 criterion 2 word for word
  #   panel, rulings, none with identity-> undeclared (a ruling that cannot be
  #                                        placed is not a ruling reported missing)
  #   panel, and its key is among them  -> clear; otherwise OWED
  vals="$(printf '%s\n' "$recs" | awk '
    function ident(s, c,   k) { k = tolower(substr(s, 1, 7))
      if (k !~ /^[0-9a-f]{7}$/ || c !~ /^[0-9]+$/) return ""
      return k "@" (c + 0) }
    $2 == "revision_packet" { pkt = $1 }
    $2 == "review_ledger"   { led = $1; ledval = $4; lsha = $5; lcyc = $6 }
    $2 == "judgment" { nj++; k = ident($5, $6); if (k != "") { ruled[k] = 1; nid++ } }
    $2 == "reviewer_verdict" && $3 ~ /^@?validation-agent$/ { v[++nv] = $1 }
    END {
      rule = "undeclared"
      if (led != "") {
        lk = ident(lsha, lcyc)
        if      (lk == "")     rule = "undeclared"
        else if (nj + 0 == 0)  rule = "owed"
        else if (nid + 0 == 0) rule = "undeclared"
        else if (lk in ruled)  rule = "clear"
        else                   rule = "owed"
        for (i = 1; i <= nv; i++) if (v[i] > led && v[i] > ver) ver = v[i]
      }
      printf "%s|%s|%d|%s|%s\n", pkt, led, ledval + 0, rule, ver
    }')"
  IFS='|' read -r pkt led led_names_val ruling ver <<EOF
$vals
EOF

  # --- condition A: the newest packet, newer than the newest commit -------------
  if [ -z "$pkt" ]; then packet=undeclared
  elif [[ "$pkt" > "$newest_commit" ]]; then
    packet=owed
    g1=$(( ( $(secs "$pkt") - $(secs "$newest_commit") ) / 86400 ))
    g2=$(( ( $(secs "$NOW") - $(secs "$pkt") ) / 86400 ))
    [ "$g2" -lt 0 ] && g2=0
    OWED="$OWED
  #$pr revision_packet posted $pkt is $g1 day(s) past the newest commit $newest_commit (packet-minus-commit), and has been carried $g2 day(s) since, to now $NOW (now-minus-packet)"
  else packet=clear
  fi

  [ "$ruling" = owed ] && OWED="$OWED
  #$pr review_ledger $led has no declared judgment of its own identity — a panel convened and never ruled"

  # --- condition B2: the mandatory last lane -----------------------------------
  # DEFECT: a ruling closes the validation question. It does not — #135 c4/c5.
  [ "$DEFECT" = b2_discharged ] && [ "$ruling" = clear ] && ver="closed-by-the-ruling"
  if   [ -z "$led" ];                 then validation=undeclared
  elif [ "$led_names_val" -eq 0 ];    then validation=n/a
  elif [ -n "$ver" ];                 then validation=clear
  else validation=owed
    OWED="$OWED
  #$pr review_ledger $led names @validation-agent and no validation verdict follows it — the mandatory last lane"
  fi

  # DEFECT: two states. `undeclared` becomes `clear` and the distinction is gone.
  if [ "$DEFECT" = two_state ]; then
    [ "$packet"     = undeclared ] && packet=clear
    [ "$ruling"     = undeclared ] && ruling=clear
    [ "$validation" = undeclared ] && validation=clear
  fi

  # THE CLASSES ARE COUNTED INDEPENDENTLY. A PR owed on one class and unanswerable
  # on another belongs in BOTH counts, so the four counts need not sum to the PRs
  # read; precedence decides only the single state token printed beside the PR.
  # Tested class by class rather than by matching a substring of the three joined
  # together — a join has boundaries, and nothing but today's vocabulary keeps one
  # from spelling a token that is not there.
  has_owed=0; has_undecl=0
  for c in "$packet" "$ruling" "$validation"; do
    [ "$c" = owed ]       && has_owed=1
    [ "$c" = undeclared ] && has_undecl=1
  done
  state=clear
  [ "$has_owed" -eq 1 ]   && { state=owed; n_owed=$((n_owed + 1)); }
  if [ "$has_undecl" -eq 1 ]; then
    [ "$state" = owed ] || state=undeclared
    n_undecl=$((n_undecl + 1))
    UNDECL="$UNDECL
  #$pr no declared artifact of at least one class — cannot be judged; this is NOT \"nothing owed\""
  fi
  [ "$state" = clear ] && n_clear=$((n_clear + 1))

  TABLE="$TABLE
$(printf '#%-5s %-11s packet=%-11s ruling=%-11s validation=%s' \
    "$pr" "$state" "$packet" "$ruling" "$validation")"
done

printf 'owed_work: %d PR(s) read\n' "$n_read"
[ -n "$TABLE" ]  && printf '%s\n' "${TABLE#
}"
[ -n "$OWED" ]   && { printf 'owed:\n'; printf '%s\n' "${OWED#
}"; }
[ -n "$UNDECL" ] && { printf 'undeclared:\n'; printf '%s\n' "${UNDECL#
}"; }
[ -n "$UNREAD" ] && { printf 'unreadable:\n'; printf '%s\n' "${UNREAD#
}"; }
printf 'summary: owed %d · clear %d · undeclared %d · unreadable %d\n' \
  "$n_owed" "$n_clear" "$n_undecl" "$n_unread"

# Precedence 2 > 1 > 3 > 0.
# DEFECT: an unreadable source exits 0 — a quiet queue, which is what
# `fingerprint.sh` was fixed to stop being.
if [ "$n_unread" -gt 0 ] && [ "$DEFECT" != quiet_fail ]; then exit 2; fi
[ "$n_owed"   -gt 0 ] && exit 1
[ "$n_undecl" -gt 0 ] && exit 3
exit 0
TEMPLATE

mk() { # mk <defect>
  sed "s/__DEFECT__/$1/" "$TMP/template" > "$TMP/$1.sh"
  chmod +x "$TMP/$1.sh"
  cmp -s "$TMP/template" "$TMP/$1.sh" && {
    printf 'FAIL  construction · the %s stub is byte-identical to the template — a stub that mutates nothing proves nothing\n' "$1"
    exit 1; }
}
for d in none two_state prose quiet_fail b2_discharged; do mk "$d"; done

fails=0
report() { # report <defect> <must-be: green|red>
  out="$(bash "$CORPUS" "$TMP/$1.sh" 2>&1)"; rc=$?
  n_pass="$(printf '%s\n' "$out" | grep -c '^PASS' || true)"
  n_fail="$(printf '%s\n' "$out" | grep -c '^FAIL' || true)"
  if [ "$2" = green ]; then
    if [ "$rc" -eq 0 ]; then
      printf 'PASS  %-14s contract-faithful · %s checks passed, corpus rc 0\n' "$1" "$n_pass"
    else
      printf 'FAIL  %-14s contract-faithful stub did NOT satisfy the corpus (rc %s, %s red)\n' "$1" "$rc" "$n_fail"
      printf '%s\n' "$out" | grep '^FAIL' | sed 's/^/        /'
      fails=$((fails + 1))
    fi
  else
    if [ "$rc" -ne 0 ] && [ "$n_fail" -gt 0 ]; then
      printf 'PASS  %-14s CAUGHT · %s checks red, %s still green, corpus rc %s\n' "$1" "$n_fail" "$n_pass" "$rc"
      printf '%s\n' "$out" | grep '^FAIL' | sed 's/^\(FAIL  \)/        red: /' \
        | cut -c1-118 | head -8
      [ "$n_fail" -gt 8 ] && printf '        red: … and %s more\n' "$((n_fail - 8))"
    else
      printf 'FAIL  %-14s NOT CAUGHT — the corpus is green against a detector with this defect\n' "$1"
      fails=$((fails + 1))
    fi
  fi
  printf '\n'
}

printf 'corpus: %s\n\n' "$CORPUS"
report none           green
report two_state      red
report prose          red
report quiet_fail     red
report b2_discharged  red

[ "$fails" -eq 0 ] || { printf '%s stub(s) behaved wrongly\n' "$fails"; exit 1; }
printf 'the corpus discriminates: one contract-faithful detector passes it and four defective ones do not\n'
exit 0
