#!/usr/bin/env bash
# owed-work-recall.sh — the corpus scripts/loop/owed-work.sh is measured against. (#172, #178)
#
# #172 criterion 5 asks for "a fixture that goes red without it, and written by a
# hand other than the script's". This file is that fixture. It was written from
# #172's and #178's acceptance criteria and from the dispatched contract, by an
# author who did not write the detector — the instruments this repository has
# shipped that could only ever pass were the ones whose author wrote their own
# proof.
#
# WHAT THIS FILE TOOK FROM THE SCRIPT, AND WHAT IT REFUSED TO. Every expected
# STATE and every expected EXIT STATUS below is derived from the criteria and the
# contract, never from the detector's body. Two things were taken by OBSERVATION,
# because a corpus cannot assert against an interface it may not look at:
#   · the `gh` invocations, discovered by running the detector against a stub that
#     logged its argv and printed nothing else. That is the same convention
#     `dependents-declared-edges.sh` uses — the stub returns what `gh --jq` would
#     have returned, because standalone `jq` is NOT installed on this host.
#   · the shape of the report lines, discovered by running the detector once.
# Neither observation decides whether an answer is RIGHT. Where the detector's
# answer differs from the criteria, this corpus goes red and the criteria win.
#
# ---------------------------------------------------------------------------
# THE CONTRACT ASSERTED HERE
# ---------------------------------------------------------------------------
# MARKERS. An artifact is identified ONLY by the `artifact:` key declared in a
# comment's leading YAML block. Never by heading text. Never by prose. The 2026-09-06
# survey used two ad-hoc `grep` patterns and produced FALSE NEGATIVES REPORTED AS
# FACT — #147 called "zero judge ledgers" when it had four, #154 "no packet found"
# when a packet existed and had been worked (#172 criterion 2). FIXTURES A6, A7 and
# A8 are that failure in both directions: a declared packet under an unrelated
# heading MUST be found, and a heading that says "Revision packet" over a comment
# declaring nothing MUST NOT be.
#
# THE THREE CONDITIONS.
#   A  (#172)  the newest declared `revision_packet` is newer than the PR's newest
#              commit — a remand a judge persisted and no worker ever received.
#   B1 (#178)  a persisted convened panel (a declared `review_ledger`) with no
#              declared `judgment` after it.
#   B2 (#178)  a panel naming `@validation-agent` with no VALIDATION VERDICT after
#              it, where a validation verdict is a comment declaring
#              `artifact: reviewer_verdict` whose own `reviewer:` key is
#              `validation-agent` — both keys mandatory per
#              `review-verdicts § Reviewer verdict`.
#
# TWO CONSEQUENCES OF B2 THAT THE CORPUS PINS BECAUSE THEY ARE WHERE IT WOULD BE
# MISSED. A `judgment` does NOT discharge B2 (FIXTURE B2b) — that is exactly PR
# #135 cycles 4 and 5, where a panel was persisted, a ruling followed, and the
# mandatory last lane was never dispatched at all. And B2 OVER-REPORTS a ledger
# whose own table already records validation (FIXTURE B2g); that is pinned as built
# behaviour, not as an ideal, because a mandatory lane must fail toward being
# reported and reading a verdict out of a markdown cell is the semantic guessing
# criterion 2 forbids.
#
# THREE STATES PER CLASS, NOT TWO. This is the load-bearing part.
#   owed        a declared artifact of that class, and nothing after it
#   clear       a declared artifact of that class, and something after it
#   undeclared  NO declared artifact of that class. The detector CANNOT ANSWER.
#               Never `clear`, never folded into "nothing owed" (#172 criterion 3).
#
# WHY `undeclared` CARRIES THE WHOLE CORPUS. Measured 2026-09-07 across all four
# open PRs — #135, #140, #141, #142 — the only declared markers in existence were
# `artifact: orchestrator_comment` on #135. There were ZERO declared
# `revision_packet` and ZERO declared `judgment` comments on the entire board: PR
# #163 merged the requirement that judges emit the key, and every packet, ledger
# and judgment then alive predates it. PR #135 carried, at that moment, an
# UNDECLARED cycle-4 revision packet. A TWO-STATE DETECTOR THEREFORE PRINTS #135
# AS "NOTHING OWED" — reproducing, through its own fix, the exact
# false-negative-reported-as-fact that #172 exists to prevent. FIXTURE U1 is that
# PR, and `owed-work-stubs.sh` runs this corpus against a two-state detector to
# prove the fixture catches it. Re-derive the measurement with:
#   gh api repos/Caisesiume/TurfGPS/issues/<N>/comments --paginate \
#     --jq '.[].body' | grep -oE '^artifact: *[a-z_]+'
#
# EXIT CODES, precedence 2 > 1 > 3 > 0.
#   0  every PR read, nothing owed, nothing undeclared
#   1  owed work found
#   2  a source could not be read
#   3  all read, >= 1 undeclared and cannot be judged — DISTINCT FROM 0
# Precedence governs the STATUS only. Every owed and every undeclared line found
# before a failing status is still PRINTED, and a `summary:` line is printed on
# EVERY path including every failure — FIXTURES X1..X9 assert both halves, because
# a detector that stopped looking on the first unreadable PR would satisfy the exit
# code and lose the queue.
#
# GREEN AGAINST THE REAL DETECTOR PROVES NOTHING, which is why this file takes an
# optional argument and why `owed-work-stubs.sh` exists beside it.
#
# Hermetic: fixtures under mktemp, a stubbed `gh` on PATH and in $GH, no network,
# no repository state, no live API, nothing written outside TMP.
# Usage: scripts/loop/tests/owed-work-recall.sh [detector]
#        The optional argument is the red demonstration's handle — pointing this
#        corpus at a deliberately-broken detector is how it is shown to
#        DISCRIMINATE rather than merely to be green. The detector it ran is
#        printed on every run so no output is ambiguous about what it measured.
# Exit:  0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="${1:-$DIR/../owed-work.sh}"
# Absolutised before anything else: every run below is made from a throwaway
# directory, and a relative detector path — which is what the red demonstration
# hands it — would resolve against the wrong place after the `cd`.
SCRIPT="$(cd "$(dirname "$SCRIPT")" 2>/dev/null && pwd)/$(basename "$SCRIPT")"

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
# A construction premise that fails makes every assertion after it meaningless, so
# it stops the run rather than adding one more line to a tally.
die()  { printf 'FAIL  construction · %s\n' "$1"; exit 1; }

[ -f "$SCRIPT" ] || die "the detector is not at $SCRIPT"
printf 'detector: %s\n\n' "$SCRIPT"

export FIX="$TMP/fix"
mkdir -p "$FIX" "$TMP/bin"

# ---------------------------------------------------------------------------
# FIXTURE AUTHORING
#
# A fixture comment is a RAW BODY, exactly as an agent would post it, and the
# stubbed `gh` derives the declared-marker record from it. That is deliberate: if
# the fixtures held pre-extracted marker records, the marker discipline would be
# implemented by the transport and no prose-matching detector could be built to
# fail against them. The bodies are the ground truth.
# ---------------------------------------------------------------------------
CUR=''; CN=0
pr()      { CUR="$FIX/pr$1"; mkdir -p "$CUR"; CN=0; }
commits() { printf '%s\n' "$@" > "$CUR/commits"; }        # oldest first, as the API returns
cmt()     { CN=$((CN + 1)); F="$(printf '%s/c%02d.txt' "$CUR" "$CN")"
            printf '%s\n' "$1" > "$F"; cat >> "$F"; }     # cmt <created_at> <<'BODY'

# The two timestamps the historical cases are grounded in. #135's newest commit was
# 2026-08-28 22:19 and its revision packet 2026-08-29 22:23; this corpus keeps the
# clock-time and stretches the date to the EIGHT DAYS #172's table measured, so the
# gap the detector prints is the gap that was reported.
LASTCOMMIT='2026-08-28T22:19:00Z'
PACKET_8D='2026-09-05T22:23:00Z'

# --- A1  owed, the #172 headline: a packet eight days past the newest commit ----
pr 900; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 4
required_changes:
  - id: SEC-01
    owner: implementation-engineer
BODY

# --- A2  clear: a packet, and a commit after it ---------------------------------
pr 901; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T22:23:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY

# --- A3  boundary: the packet is exactly as old as the newest commit ------------
# "newer than" is strict. Equal is not newer, so nothing is owed.
pr 902; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$LASTCOMMIT" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY

# --- A4  boundary: the packet is ONE SECOND newer than the newest commit --------
# The gap in days is 0 and the packet is still owed. A detector that reported only
# gaps of a whole day would lose every remand of the same working day.
pr 903; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-08-28T22:19:01Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY

# --- A5  the NEWEST packet decides, and here it is discharged -------------------
# An older packet that was never worked, then a newer one that was. Reading "any
# packet newer than the newest commit" reports this PR owed forever.
pr 904; commits "2026-08-20T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-08-19T09:00:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY
cmt "2026-09-05T22:23:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY

# --- A6  the NEWEST packet decides, and here it is owed -------------------------
# A packet, a commit that discharged it, and a SECOND packet after that commit.
# Reading "the oldest packet" or "a packet with any commit after it" reports clear.
pr 905; commits "2026-08-20T09:00:00Z" "2026-09-01T10:00:00Z"
cmt "2026-08-19T09:00:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY
cmt "$PACKET_8D" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY

# --- A7  MARKER DISCIPLINE, direction one: prose is not a marker ----------------
# A comment whose heading says "Revision packet" and whose prose says everything a
# packet says, declaring NO key. It is not a packet. The class is `undeclared` —
# the detector cannot answer — and it is emphatically not `clear`.
pr 906; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" <<'BODY'
## Revision packet — cycle 4

This remand orders a revision packet. The required changes are listed below and
the packet's owner is `@implementation-engineer`.

| id | required change | owner |
|---|---|---|
| SEC-01 | invalidate the old refresh token | implementation-engineer |
BODY

# --- A8  MARKER DISCIPLINE, direction two: a heading is not a disqualifier ------
# A DECLARED packet whose heading text says nothing about packets. A detector
# keyed off heading text calls this PR clear and loses the remand. This is #154's
# "no packet found" reported as fact.
pr 907; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 4
BODY
# (the body carries no heading at all — the declared key is the whole marker)

# --- A9  a relay never reclassifies its courier ---------------------------------
# The LEADING declaration is the artifact's own; a later one belongs to something
# the comment quotes verbatim. An orchestrator comment that reposts a packet is an
# orchestrator comment, so this PR has no declared packet of its own.
pr 908; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" <<'BODY'
artifact: orchestrator_comment
prose_licence: none
note: relaying the judge's packet for the record

```yaml
artifact: revision_packet
prose_licence: none
cycle: 4
```
BODY

# --- B1a  owed: a convened panel with no ruling after it ------------------------
pr 910; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 910
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @correctness-reviewer | pending |
BODY

# --- B1b  clear: a panel, then a ruling ----------------------------------------
pr 911; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 911
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @correctness-reviewer | pass |
BODY
cmt "2026-09-05T23:00:00Z" <<'BODY'
artifact: judgment
prose_licence: none
decision: approve
BODY

# --- B1c  a ruling BEFORE the panel does not rule it ----------------------------
# Cycle 2's judgment cannot discharge cycle 3's panel. "No judgment AFTER it" is
# the contract; a detector asking only "does a judgment exist" reports clear.
pr 912; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-04T10:00:00Z" <<'BODY'
artifact: judgment
prose_licence: none
decision: remand
BODY
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 912
cycle: 3
BODY

# --- B1d  undeclared: no panel was ever declared on this PR ---------------------
pr 913; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY

# --- B2a  owed: the #135 cycle-4 shape -----------------------------------------
# A panel naming the mandatory last lane, and no validation verdict after it.
# Caught in life only because a human noticed a 90-minute silence.
pr 920; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 920
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| correctness | @correctness-reviewer | pass |
| validation | @validation-agent | pending |
BODY

# --- B2b  A JUDGMENT DOES NOT DISCHARGE B2 -------------------------------------
# PR #135 cycles 4 and 5 exactly: a panel was persisted, three verdicts returned,
# a ruling followed, and @validation-agent was never dispatched at all. A detector
# that let a ruling close the validation question reports both cycles complete.
# The ruling IS discharged here — `ruling=clear` — and validation is still owed.
pr 921; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 921
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| correctness | @correctness-reviewer | pass |
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T23:00:00Z" <<'BODY'
artifact: judgment
prose_licence: none
decision: approve
BODY

# --- B2c  clear: a validation verdict after the panel ---------------------------
# Both keys mandatory: `artifact: reviewer_verdict` AND `reviewer: validation-agent`.
pr 922; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 922
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T22:00:00Z" <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: validation-agent
status: valid_review
verdict: pass
BODY

# --- B2d  ANOTHER LANE'S verdict does not discharge the mandatory one -----------
# On #154 cycle 4 three lanes parked their head figures in ACCEPTED ON TRUST naming
# validation as owner, and validation never ran — so those figures were attested by
# no lane at all. A detector keyed on `artifact: reviewer_verdict` alone, ignoring
# the `reviewer:` key, reports this PR clear.
pr 923; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 923
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T22:00:00Z" <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: correctness
status: valid_review
verdict: pass
BODY

# --- B2e  a validation verdict BEFORE the panel does not discharge it -----------
# The previous cycle's validation says nothing about this cycle's panel.
pr 924; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-04T10:00:00Z" <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: validation-agent
status: valid_review
verdict: pass
BODY
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 924
cycle: 5

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY

# --- B2f  a panel that does not name the mandatory lane owes no validation ------
# B2 is "a panel NAMING @validation-agent". This one does not, so there is nothing
# owed — and the panel itself is declared, so this is `clear` and not `undeclared`.
pr 925; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 925
cycle: 1

| lane | reviewer | verdict |
|---|---|---|
| documentation | @documentation-reviewer | pass |
BODY
cmt "2026-09-05T23:00:00Z" <<'BODY'
artifact: judgment
prose_licence: none
decision: approve
BODY

# --- B2f'  the same panel, with the other two classes discharged too ------------
# B2f alone cannot show what `n/a` MEANS, because its packet class is undeclared
# and the run exits 3 for that reason instead. This PR discharges A and B1, so the
# only class left is the inapplicable one and the exit status is `n/a`'s alone.
pr 927; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T19:00:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 927
cycle: 1

| lane | reviewer | verdict |
|---|---|---|
| documentation | @documentation-reviewer | pass |
BODY
cmt "2026-09-05T23:00:00Z" <<'BODY'
artifact: judgment
prose_licence: none
decision: approve
BODY

# --- B2g  THE BUILT-IN OVER-REPORT, pinned as behaviour and not as an ideal -----
# The ledger's own table already records validation's verdict, and no separate
# reviewer_verdict was ever declared. This is reported OWED. Reading a verdict out
# of a markdown cell is the semantic guessing #172 criterion 2 forbids, and a
# MANDATORY lane must fail toward being reported: an over-report is loud and a
# human dismisses it in a second, while the failure this detector exists to prevent
# is the silent one. If this check ever goes red because the answer became `clear`,
# that is a contract change and belongs in #178 — not a loosening here.
pr 926; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 926
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| correctness | @correctness-reviewer | pass |
| validation | @validation-agent | fail |
BODY

# --- U1  THE MEASURED BOARD, 2026-09-07: PR #135 --------------------------------
# Nine declared `artifact: orchestrator_comment` couriers, and an UNDECLARED
# cycle-4 revision packet. Zero declared packets, ledgers or judgments anywhere.
# Every class is `undeclared`; the detector cannot answer, and saying so is the
# whole of #172 criterion 3. A two-state detector prints this PR as nothing owed.
pr 135; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
i=1
while [ "$i" -le 9 ]; do
  cmt "$(printf '2026-09-0%dT08:00:00Z' "$i")" <<'BODY'
artifact: orchestrator_comment
prose_licence: none
note: relaying the panel's state for the record
BODY
  i=$((i + 1))
done
cmt "2026-09-10T22:23:00Z" <<'BODY'
## Cycle 4 — required changes

Remanded. `@implementation-engineer` owns SEC-01; `@validation-agent` re-runs the
suite once the change lands. This comment predates the declared-key requirement
and declares nothing.
BODY

# --- U2  every class clear: the only shape that may exit 0 ----------------------
pr 940; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY
cmt "2026-09-05T20:30:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 940
cycle: 2

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T21:00:00Z" <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: validation-agent
status: valid_review
verdict: pass
BODY
cmt "2026-09-05T23:00:00Z" <<'BODY'
artifact: judgment
prose_licence: none
decision: approve
BODY

# --- X  unreadable sources ------------------------------------------------------
# 930: the commits endpoint fails.  931: the comments endpoint fails.
# Both are marked by the ABSENCE of the fixture file, which is what the stubbed
# `gh` turns into a non-zero exit — an API that answered with an error, not an API
# that answered with nothing.
pr 930
cmt "2026-09-05T20:00:00Z" <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 930
BODY
rm -f "$FIX/pr930/commits"
pr 931; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
: > "$FIX/pr931/NO_COMMENTS"

# ---------------------------------------------------------------------------
# THE STUBBED `gh`
#
# It answers the three calls the detector was observed to make, and it answers
# them from the RAW fixture bodies. Two comment queries are served: the declared
# marker record (what the detector's own `--jq` program computes) and the raw
# bodies (what a prose-matching detector would ask for, so that one can be built
# and failed against this corpus). A query it does not recognise is recorded and
# stops the run — a stub that silently answered an unknown question with nothing
# would hand every detector an empty queue and call it clean.
# ---------------------------------------------------------------------------
cat > "$TMP/bin/gh" <<'STUB'
#!/bin/sh
# observed forms:
#   gh pr list --state open --limit 200 --json number --jq '.[].number'
#   gh auth status
#   gh api repos/{owner}/{repo}/issues/<N>/comments --paginate --jq <program>
#   gh api repos/{owner}/{repo}/pulls/<N>/commits   --paginate --jq <program>
[ -f "$FIX/AUTH_FAIL" ] && { echo "gh: authentication required" >&2; exit 1; }
case "$1 $2" in
  "auth status") exit 0 ;;
  "pr list")     cat "$FIX/open-prs"; exit 0 ;;
esac
[ "$1" = api ] || { echo "gh stub: unrecognised call: $*" >> "$FIX/UNKNOWN"; exit 1; }
path="$2"; prog="${5:-}"
case "$path" in
  */pulls/*/commits)
    n="${path#*/pulls/}"; n="${n%/commits}"
    [ -f "$FIX/pr$n/commits" ] || { echo "API error: could not read commits for $n" >&2; exit 1; }
    cat "$FIX/pr$n/commits"; exit 0 ;;
  */issues/*/comments)
    n="${path#*/issues/}"; n="${n%/comments}"
    [ -d "$FIX/pr$n" ] || { echo "API error: no such pull request $n" >&2; exit 1; }
    [ -f "$FIX/pr$n/NO_COMMENTS" ] && { echo "API error: could not read comments for $n" >&2; exit 1; }
    case "$prog" in
      *'^artifact:'*|DECLARED)
        # what `gh --jq` returns for the marker program: one record per comment,
        #   <created_at> <declared-artifact-or-dash> <declared-reviewer-or-dash> <names-validation-agent>
        # The artifact and reviewer are the FIRST such line in the body — a later
        # one belongs to something the comment relays verbatim.
        for f in "$FIX/pr$n"/c*.txt; do
          [ -e "$f" ] || continue
          awk '
            NR == 1 { ts = $0; next }
            { body = body $0 "\n"
              if (art == "" && $0 ~ /^artifact:[ \t]*[A-Za-z0-9_]/) {
                art = $0; sub(/^artifact:[ \t]*/, "", art); sub(/[^A-Za-z0-9_].*$/, "", art) }
              if (rev == "" && $0 ~ /^reviewer:[ \t]*[@A-Za-z0-9_-]/) {
                rev = $0; sub(/^reviewer:[ \t]*/, "", rev); sub(/[^@A-Za-z0-9_-].*$/, "", rev) } }
            END { val = (body ~ /(^|[^0-9A-Za-z_-])@?validation-agent([^0-9A-Za-z_-]|$)/) ? 1 : 0
                  printf "%s %s %s %s\n", ts, (art == "" ? "-" : art), (rev == "" ? "-" : rev), val }
          ' "$f"
        done
        exit 0 ;;
      *RAWBODY*)
        for f in "$FIX/pr$n"/c*.txt; do
          [ -e "$f" ] || continue
          awk 'NR == 1 { printf "#REC %s\n", $0; next } { print }' "$f"
        done
        exit 0 ;;
      *) echo "gh stub: unrecognised --jq program for $path" >> "$FIX/UNKNOWN"; exit 1 ;;
    esac ;;
esac
echo "gh stub: unrecognised api path: $path" >> "$FIX/UNKNOWN"
exit 1
STUB
chmod +x "$TMP/bin/gh"
export PATH="$TMP/bin:$PATH" GH="$TMP/bin/gh"

# ---------------------------------------------------------------------------
# CONSTRUCTION GUARDS. A fixture that is not what it claims is exactly the vacuous
# instrument this file exists to prevent, so a miss is a hard stop.
# ---------------------------------------------------------------------------
declared_ids() { "$TMP/bin/gh" api "repos/o/r/issues/$1/comments" --paginate --jq DECLARED; }

# Comments must be chronological, because every condition here is "X with nothing
# after it" and a fixture whose records arrived out of order tests a different
# question than the one it names.
for d in "$FIX"/pr*; do
  n="${d##*/pr}"
  [ -f "$d/NO_COMMENTS" ] && continue
  declared_ids "$n" | awk -v pr="$n" '
    $1 < prev { printf "pr %s: comment %d at %s precedes %s\n", pr, NR, $1, prev; bad = 1 }
    { prev = $1 }
    END { exit bad + 0 }' || die "fixture comments are not in chronological order (above)"
done

# The three fixtures whose whole point is WHICH id was extracted verify it here.
[ "$(declared_ids 906 | awk '{print $2}')" = '-' ] \
  || die "A7 must declare NOTHING — a prose-only packet that extracts an id is not the fixture this corpus needs"
[ "$(declared_ids 907 | awk '{print $2}')" = 'revision_packet' ] \
  || die "A8 must declare revision_packet under no heading at all"
[ "$(declared_ids 908 | awk '{print $2}')" = 'orchestrator_comment' ] \
  || die "A9's courier must extract its OWN leading key, not the packet it relays"
[ "$(declared_ids 135 | awk '$2 != "orchestrator_comment"' | awk '{print $2}' | tr -d '\n')" = '-' ] \
  || die "U1 must hold nine couriers and one comment declaring nothing — the 2026-09-07 board"
[ "$(declared_ids 920 | awk '$2 == "review_ledger" { print $4 }')" = '1' ] \
  || die "B2a's ledger must NAME @validation-agent or the mandatory-lane condition is never reached"
[ "$(declared_ids 925 | awk '$2 == "review_ledger" { print $4 }')" = '0' ] \
  || die "B2f's ledger must NOT name @validation-agent or it is the same fixture as B2a"
[ "$(declared_ids 926 | awk '$2 == "reviewer_verdict" { print $3 }')" = '' ] \
  || die "B2g must declare NO reviewer_verdict — the over-report it pins is the ledger's table standing alone"

cd "$TMP" || die "could not enter $TMP"

# ---------------------------------------------------------------------------
# ASSERTIONS
# ---------------------------------------------------------------------------
run() { OUT="$(bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }
flat() { printf '%s' "$OUT" | tr '\n' '|'; }

check_rc() { # check_rc <label> <expected>
  [ "$RC" = "$2" ] && pass "$1" || bad "$1" "expected rc $2, got rc $RC: $(flat)"; }
check_has() { # check_has <label> <substring>
  case "$OUT" in *"$2"*) pass "$1" ;; *) bad "$1" "expected to find '$2': $(flat)" ;; esac; }

# The per-PR line, matched tolerantly: the corpus asserts that a CLASS and a STATE
# co-occur in the PR's own line, which is the contract's "three states per class".
# Spacing and ordering are the detector's to choose; the pairing is not.
cls() { # cls <label> <pr> <packet-state> <ruling-state> <validation-state>
  L="$(printf '%s\n' "$OUT" | grep -E "^#$2[[:space:]]" | head -1)"
  if [ -z "$L" ]; then bad "$1" "no per-PR line for #$2: $(flat)"; return; fi
  for pair in "packet $3" "ruling $4" "validation $5"; do
    k="${pair%% *}"; v="${pair##* }"
    printf '%s' "$L" | grep -Eq "$k[[:space:]]*[=:][[:space:]]*$v([^a-z]|\$)" || {
      bad "$1" "expected $k=$v, got: $L"; return; }
  done
  pass "$1"
}
# AN ABSENCE IS EVIDENCE ONLY FROM A RUN THAT COULD HAVE CARRIED THE THING ABSENT,
# and this corpus asserts absences in exactly one place — X15, on #135's own line.
# It is written inline there rather than as a helper because the scoping is the
# whole of the assertion: an absence read over the WHOLE report fails on the
# summary's `clear 0` tally, which is a count and not a verdict on any PR. That is
# the shape `dependents-declared-edges.sh` records as the #138 anti-pattern —
# an assertion that cannot fail, or one that fails for a reason it does not name.

READABLE='900 901 902 903 904 905 906 907 908 910 911 912 913 920 921 922 923 924 925 926 927 135 940'

# ---- one run over every readable PR: the classification corpus ----------------
# shellcheck disable=SC2086
run $READABLE

# CONDITION A (#172) — a remand nobody worked.
cls "A1  a packet eight days past the newest commit is OWED"                900 owed       undeclared undeclared
cls "A2  a packet with a commit after it is CLEAR"                          901 clear      undeclared undeclared
cls "A3  a packet exactly as old as the newest commit is not NEWER"         902 clear      undeclared undeclared
cls "A4  a packet one second newer is OWED — same-day remands count"        903 owed       undeclared undeclared
cls "A5  the NEWEST packet decides, and this one was worked"                904 clear      undeclared undeclared
cls "A6  the NEWEST packet decides, and this one was not"                   905 owed       undeclared undeclared
cls "A7  prose is not a marker: heading + prose declares nothing"           906 undeclared undeclared undeclared
cls "A8  a declared packet under no heading is still found (#154)"          907 owed       undeclared undeclared
cls "A9  a courier relaying a packet is not a packet"                       908 undeclared undeclared undeclared

# The gap in days is #172 criterion 1's own words, and 8 is the measured figure.
check_has "A1  ... and the gap is reported in days, as measured: 8"          "8 day"
check_has "A4  ... and a zero-day gap is reported rather than swallowed"     "0 day"

# CONDITION B1 (#178) — a convened panel with no ruling.
cls "B1a a panel with no ruling after it is OWED"                           910 undeclared owed       n/a
cls "B1b a panel with a ruling after it is CLEAR"                           911 undeclared clear      n/a
cls "B1c a ruling BEFORE the panel does not rule it"                        912 undeclared owed       n/a
cls "B1d no panel declared: the detector cannot answer"                     913 owed       undeclared undeclared

# CONDITION B2 (#178 criterion 3) — the mandatory last lane, silently skipped.
cls "B2a a panel naming @validation-agent with no verdict is OWED"          920 undeclared owed       owed
cls "B2b A JUDGMENT DOES NOT DISCHARGE B2 — #135 cycles 4 and 5"            921 undeclared clear      owed
cls "B2c a reviewer_verdict from validation-agent DISCHARGES it"            922 undeclared owed       clear
cls "B2d another lane's verdict does not discharge the mandatory one"       923 undeclared owed       owed
cls "B2e a validation verdict BEFORE the panel does not discharge it"       924 undeclared owed       owed
cls "B2f a panel not naming the lane owes no validation"                    925 undeclared clear      n/a
cls "B2f' the same, with A and B1 discharged so n/a stands alone"           927 clear      clear      n/a
cls "B2g BUILT-IN OVER-REPORT: a ledger recording its own validation"       926 undeclared owed       owed

# THE FOURTH STATE, AND WHY IT IS PINNED HERE RATHER THAN ACCOMMODATED. The
# contract names three states per class. The detector emits a FOURTH, `n/a`, for
# the one sub-case that has no third: a declared panel that does not name the
# mandatory lane owes nothing and is not unanswerable. The vocabulary is reported
# upstream rather than settled here — but its MEANING is asserted, because a
# fourth token whose exit-code meaning is unpinned is exactly how `undeclared`
# would come back as `clear` under a different name. `n/a` must count as
# nothing-owed AND as answered — asserted in the exit-code section below as X1',
# where a run of its own can be made without clobbering this one's report.

# THE THREE STATES, and the one that carries the whole item.
cls "U1  the 2026-09-07 board: #135 is UNDECLARED in every class"           135 undeclared undeclared undeclared
cls "U2  every class discharged"                                            940 clear      clear      clear

# ---- exit codes and their precedence ------------------------------------------
run 940;         check_rc  "X1  all read, nothing owed, nothing undeclared -> 0"  0
run 940;         check_has "X1  ... and a summary line is printed"                "summary:"
# X1' — what the fourth state MEANS. #927's only class not discharged is the
# inapplicable one, so if `n/a` ever came to mean "cannot answer" this exits 3.
run 927;         check_rc  "X1' n/a is nothing-owed AND answered -> 0, not 1, not 3" 0
run 927;         check_has "X1' ... and it is a run that read something"          "1 PR"
run 900;         check_rc  "X2  owed found -> 1"                                  1
run 135;         check_rc  "X3  all read, >=1 undeclared -> 3, NEVER 0"           3
run 930;         check_rc  "X4  commits unreadable -> 2"                          2
run 931;         check_rc  "X5  comments unreadable -> 2"                         2
run badnumber;   check_rc  "X6  an argument that is not a PR -> 2"                2

run 900 135;     check_rc  "X7  precedence 1 > 3: owed outranks undeclared"       1
run 900 135;     check_has "X7  ... and the undeclared PR is still printed"       "#135"
run 940 135;     check_rc  "X8  precedence 3 > 0: undeclared outranks clear"      3
run 900 930;     check_rc  "X9  precedence 2 > 1: unreadable outranks owed"       2
run 900 930;     check_has "X9  ... and the owed PR found before it still prints" "#900"
run 135 930;     check_rc  "X10 precedence 2 > 3"                                 2
run 135 930;     check_has "X10 ... and the undeclared PR still prints"           "#135"

# A summary line on EVERY path THAT READ SOMETHING, including every failure. A run
# that died quietly after printing nothing is the shape `fingerprint.sh` was fixed
# to stop, and it is the shape a caller cannot distinguish from an empty queue.
for a in 940 900 135 930 931; do
  run "$a"; check_has "X11 summary printed on the '$a' path (rc $RC)" "summary:"
done
# The usage path reads nothing at all, so it has no queue to summarise. What it
# must not be is SILENT: it names its own failure and it exits 2. The contract's
# words are "a summary line prints on every path", and this path does not print
# one — reported upstream as a deviation rather than accommodated by deleting the
# check. What is asserted here is the property the rule exists to protect.
run badnumber
check_rc  "X11'the usage path is a source that could not be read -> 2"           2
check_has "X11'... and it names its own failure rather than going quiet"         "usage"

# CRITERION 3 ITSELF, and this is the check the whole item turns on: SILENCE AND
# "NOTHING OWED" MUST NOT LOOK IDENTICAL. Asserted as the difference between two
# runs rather than as a word in one of them — a detector can print any word it
# likes, and what a caller acts on is the status and the count. #135 declares
# nothing and #940 declares everything and owes nothing; the two must be
# distinguishable on both.
run 135; RC135="$RC"; OUT135="$OUT"
run 940; RC940="$RC"; OUT940="$OUT"
[ "$RC135" != "$RC940" ] \
  && pass "X12 the undeclared board and the all-clear board do not share an exit status" \
  || bad  "X12 the undeclared board and the all-clear board do not share an exit status" \
          "both exited $RC135 — a two-state detector is indistinguishable from a three-state one here"
case "$OUT135" in *"undeclared 1"*) pass "X13 the undeclared board counts its unanswerable PR" ;;
  *) bad "X13 the undeclared board counts its unanswerable PR" "no 'undeclared 1' in: $(printf '%s' "$OUT135" | tr '\n' '|')" ;; esac
case "$OUT940" in *"undeclared 0"*) pass "X14 the all-clear board counts none" ;;
  *) bad "X14 the all-clear board counts none" "no 'undeclared 0' in: $(printf '%s' "$OUT940" | tr '\n' '|')" ;; esac

# And #135's own line never calls itself clear. Line-scoped, because `clear 0` in
# the summary is a tally and not a verdict on this PR — an absence read over the
# whole report would fail on the tally and prove nothing about the line.
OUT="$OUT135"
L135="$(printf '%s\n' "$OUT135" | grep -E '^#135[[:space:]]' | head -1)"
if [ -n "$L135" ] && ! printf '%s' "$L135" | grep -q 'clear'; then
  pass "X15 #135's own line never reads clear"
else
  bad "X15 #135's own line never reads clear" "got: ${L135:-<no line for #135>}"
fi

# ---- the default queue: no arguments means every open PR ----------------------
printf '900\n135\n' > "$FIX/open-prs"
run
check_rc  "X16 with no arguments the open-PR queue is read"                       1
check_has "X16 ... and the owed PR from that queue is reported"                   "#900"
check_has "X16 ... and the undeclared PR from that queue is too"                  "#135"

# Auth that cannot be read is a source that cannot be read. The fixture fails the
# whole CLI and not only `auth status`, because an unauthenticated `gh` cannot
# list pull requests either — a stub that failed only the check the detector runs
# LAST would let the queue arrive and prove nothing about the failure.
: > "$FIX/AUTH_FAIL"
run
check_rc  "X17 auth that cannot be read is a source that cannot be read -> 2"     2
check_has "X17 ... and it still prints a summary rather than dying quietly"       "summary:"
rm -f "$FIX/AUTH_FAIL"

# ---------------------------------------------------------------------------
[ -s "$FIX/UNKNOWN" ] && {
  printf '\nFAIL  construction · the detector asked the stubbed gh a question it does not serve:\n'
  sed 's/^/        /' "$FIX/UNKNOWN"
  printf '        Every assertion above measured a stub that answered with an error.\n'
  exit 1; }

[ "$fails" -eq 0 ] || { printf '\n%s check(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
