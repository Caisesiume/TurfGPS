#!/usr/bin/env bash
# owed-work-recall.sh — the corpus scripts/loop/owed-work.sh is measured against. (#172, #178)
#
# #172 criterion 5 asks for "a fixture that goes red without it, and written by a
# hand other than the script's". This file is that fixture. It was written from
# #172's and #178's acceptance criteria and from the dispatched contract, by an
# author who did not write the detector.
#
# INDEPENDENCE IS SUBSTANTIVE OR IT IS NOTHING, AND CYCLE 1 PROVED THAT ON THIS
# FILE. The first version of this corpus was written by a different hand and still
# REPRODUCED THE RULE IT WAS TESTING: its B1 fixtures pinned "a judgment strictly
# after the panel", which is the detector's own line read back to it, and fixture
# B1c CERTIFIED the inversion by asserting judgment-then-ledger as `owed` — the
# shape every correctly ruled PR in this repository has. A corpus that copies the
# rule cannot catch the rule being wrong. Three things follow, and they are the
# method this file now holds itself to:
#
#   1. EVERY EXPECTED ANSWER IS DERIVED FROM A NAMED SOURCE, cited at the fixture.
#      B1's source is `pr-judge § Phase 9` -> `§ Phase 10`, quoted below. Where the
#      detector's answer differs from the source, this corpus goes red and the
#      source wins.
#   2. THE TRANSPORT HOLDS NO ALGORITHM. The stubbed `gh` used to re-implement the
#      detector's marker extraction in awk — a twin that would have been wrong in
#      the same way at the same time. Each fixture comment now DECLARES the record
#      it must yield, hand-derived from criterion 2, and the stub is a lookup. See
#      FIXTURE AUTHORING.
#   3. THE `none` STUB IN `owed-work-stubs.sh` IS A SECOND IMPLEMENTATION, not a
#      transcription of the first. Its divergences are stated there.
#
# WHAT THIS FILE TOOK FROM THE SCRIPT, AND WHAT IT REFUSED TO. Every expected
# STATE and every expected EXIT STATUS below is derived from the criteria and the
# contract, never from the detector's body. Two things were taken by OBSERVATION,
# because a corpus cannot assert against an interface it may not look at:
#   · the `gh` invocations and the shape of the comment record, discovered by
#     running the detector against a stub that logged its argv. That is the
#     convention `dependents-declared-edges.sh` uses — a stub returns what
#     `gh --jq` would have returned, because standalone `jq` is NOT installed on
#     this host.
#   · the shape of the report lines, discovered by running the detector once.
# Neither observation decides whether an answer is RIGHT.
#
# WHAT REMAINS UNEXECUTED, SAID OUT LOUD RATHER THAN LEFT TO BE FOUND AGAIN.
# `COMMENT_JQ` — the detector's marker extractor, and the whole of #172
# criterion 2 — is NOT executed by this corpus, because there is no jq engine on
# this host that a hermetic test may reach: standalone `jq` is absent, and the one
# compiled into `gh` is reachable only through an HTTP request. The awk twin that
# used to stand in for it has been DELETED rather than replaced, so the corpus no
# longer holds a second copy of the rule that could be wrong in step with the
# first; what it holds instead is hand-derived data. The residual gap is a detector
# whose `COMMENT_JQ` is broken while its shell is right, and this corpus cannot see
# it. Filed rather than papered over.
#
# ---------------------------------------------------------------------------
# THE CONTRACT ASSERTED HERE
# ---------------------------------------------------------------------------
# MARKERS. An artifact is identified ONLY by the `artifact:` key declared in a
# comment's leading YAML block. Never by heading text. Never by prose. The
# 2026-09-06 survey used two ad-hoc `grep` patterns and produced FALSE NEGATIVES
# REPORTED AS FACT — #147 called "zero judge ledgers" when it had four, #154 "no
# packet found" when a packet existed and had been worked (#172 criterion 2).
# FIXTURES A7, A8 and A9 are that failure in both directions.
#
# THE THREE CONDITIONS.
#   A  (#172)  the newest declared `revision_packet` is newer than the PR's newest
#              commit — a remand a judge persisted and no worker ever received.
#   B1 (#178)  a convened panel (a declared `review_ledger`) with no ruling OF ITS
#              OWN IDENTITY. See the derivation below; this is NOT a time test.
#   B2 (#178)  a panel naming `@validation-agent` with no validation verdict
#              after it.
#
# B1 IS DERIVED FROM THE POSTING CONTRACT AND NOT FROM A CLOCK, AND THAT IS THE
# CORRECTION THIS CYCLE MAKES. `pr-judge § Phase 9 — Rule` posts the judgment
# comment; `§ Phase 10 — Ledger, convergence, budget` posts the ledger. In that
# order, on every path — approve and remand alike. So ON EVERY CORRECTLY RULED PR
# THE LEDGER IS THE NEWER OF THE TWO, and "a judgment after the panel" is FALSE on
# exactly the PRs that were handled properly. Measured across 9 of 9 ruled cycles
# (#163 x6, #154 x2, #135 x1): ledger strictly newer, zero counterexamples. #181
# cycle 1 is judgment 18:40:55Z, ledger 18:41:01Z — six seconds apart, in the
# contract's order, and the first version of this corpus asserted that shape was
# owed work.
#
# What #178 criterion 2 actually asks for is EXISTENCE — "a convened panel with no
# ruling" — and both artifacts declare the identity that pairs them: `sha:` and
# `cycle:`. #135's cycle-7 judgment and ledger both declare `sha: d5f3a58,
# cycle: 7`; #181's both declare `297632d4…` and `cycle: 1`. So the corpus asserts
# on IDENTITY, and FIXTURES B1b and B1e pin that posting order decides nothing:
# the same identity discharges the panel whichever comment came first.
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
# WHAT B2'S DISCHARGE IS IS NOT SETTLED, AND THIS CORPUS DOES NOT SETTLE IT.
# `artifact: reviewer_verdict` carrying `reviewer: validation-agent` is a shape
# `@validation-agent` is contractually forbidden to emit (`validation-agent.md:83`),
# so the discharge branch is unreachable in life. #182 is the item that declares an
# artifact class for a validation result; #178 is blocked on it. FIXTURES B2c and
# U2 carry the provisional shape and are marked where they sit — they are the only
# assertions here that a contract decision may invalidate, and no new one was added
# this cycle.
#
# THREE STATES PER CLASS, NOT TWO. This is the load-bearing part.
#   owed        a declared artifact of that class, and nothing after it
#   clear       a declared artifact of that class, and something after it
#   undeclared  the detector CANNOT ANSWER — no declared artifact of that class,
#               or (on a discharge class) none carrying an identity to match.
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
# EVERY path including every failure — FIXTURES X1..X11 assert both halves.
#
# AN OWED CLASS IS ASSERTED ON FOUR SURFACES, NOT ON ONE TABLE CELL: the class
# pair in the PR's own line, the PR's STATE token, the run's EXIT STATUS, and the
# SUMMARY COUNTS the consumer branches on (`engineering-lead § Phase 1`). A cell is
# the cheapest of the four to get right by accident, and under the previous corpus
# a detector that lost a whole condition moved exactly one of them. See the
# S-block.
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
# A fixture comment is a RAW BODY, exactly as an agent would post it, TOGETHER
# WITH THE RECORD IT MUST YIELD, written out by hand from #172 criterion 2. The
# stubbed `gh` then serves that record verbatim and computes nothing.
#
# WHY THE RECORD IS DATA AND NOT CODE. The stub used to derive the record from the
# body with an awk program that re-implemented the detector's marker extraction —
# the same first-declaration rule, the same word-boundary regex, character for
# character. Two copies of one rule fail together: a corpus built that way is green
# against a detector that is wrong in the way the corpus is also wrong, which is
# the failure @pr-judge found on cycle 1 of this PR. Hand-written records have no
# such coupling. What they cost is transcription error, and the TRANSCRIPTION
# GUARDS below buy that back by re-reading the bodies with a plain `grep` — a check
# on the author, deliberately weaker than the rule and never consumed by any
# detector.
#
# Two records are declared per comment because the red demonstration needs both:
#   declared  what the leading `artifact:`/`reviewer:`/`sha:`/`cycle:` keys yield —
#             what `gh --jq "$COMMENT_JQ"` returns to a contract-faithful detector
#   heading   what a detector keyed off HEADING TEXT would see instead. It differs
#             only in the identified artifact, which is the whole of the defect,
#             and defaults to "identified nothing" where the body has no heading.
# Holding the second as data is what lets `owed-work-stubs.sh` build the 2026-09-06
# survey's failure as one query change rather than as a second extractor — and it
# is why the `validation-agent` word-boundary regex, which used to stand
# character-identical in three files, now stands in the detector alone.
#
# Field order is the record's, observed once from the detector's own destructuring:
#   <created_at> <artifact-id> <reviewer> <names-validation-agent> <sha> <cycle>
# `-` is "the comment declared no such key".
# ---------------------------------------------------------------------------
CUR=''; CN=0
pr()      { CUR="$FIX/pr$1"; mkdir -p "$CUR"; CN=0
            : > "$CUR/records.declared"; : > "$CUR/records.heading"; }
commits() { printf '%s\n' "$@" > "$CUR/commits"; }        # oldest first, as the API returns
# cmt <created_at> <id> <reviewer> <namesval> <sha> <cycle> [<heading-id> <heading-reviewer>] <<'BODY'
cmt() {
  CN=$((CN + 1)); F="$(printf '%s/c%02d.txt' "$CUR" "$CN")"
  cat > "$F"
  printf '%s %s %s %s %s %s\n' "$1" "$2"      "$3"      "$4" "$5" "$6" >> "$CUR/records.declared"
  printf '%s %s %s %s %s %s\n' "$1" "${7:--}" "${8:--}" "$4" "$5" "$6" >> "$CUR/records.heading"
}

# The two timestamps the historical cases are grounded in. #135's newest commit was
# 2026-08-28 22:19 and its revision packet 2026-08-29 22:23; this corpus keeps the
# clock-time and stretches the date to the EIGHT DAYS #172's table measured, so the
# gap the detector prints is the gap that was reported.
LASTCOMMIT='2026-08-28T22:19:00Z'
PACKET_8D='2026-09-05T22:23:00Z'

# The identities the B-fixtures are built on are the real ones the judge named:
# #181 cycle 1 declared the 40-hex form, #135 cycle 7 the seven-character prefix.
SHA181='297632d4e4d714249dbf3cc90d5c411815e2f848'
SHA135='d5f3a58'
SHA135_LONG='d5f3a58c9b1e2f3a4b5c6d7e8f90123456789abc'
SHA_OLD='aaaa111bbbb222cccc333dddd444eeee5555ffff'

# THE CONTRACT'S OWN POSTING TIMES. `pr-judge § Phase 9` posts the judgment and
# `§ Phase 10` the ledger; on #181 cycle 1 that was 18:40:55Z then 18:41:01Z. Every
# correctly-ruled fixture below uses that pair, so the corpus is asserting against
# the shape the contract actually produces rather than against a convenient one.
T_JUDGMENT='2026-09-05T18:40:55Z'
T_LEDGER='2026-09-05T18:41:01Z'

# --- A1  owed, the #172 headline: a packet eight days past the newest commit ----
pr 900; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" revision_packet - 0 - 4 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 4
required_changes:
  - id: SEC-01
    owner: implementation-engineer
BODY

# --- A2  clear: a packet, and a commit after it ---------------------------------
pr 901; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T22:23:00Z" revision_packet - 0 - 2 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY

# --- A3  boundary: the packet is exactly as old as the newest commit ------------
# "newer than" is strict. Equal is not newer, so nothing is owed.
pr 902; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$LASTCOMMIT" revision_packet - 0 - 1 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY

# --- A4  boundary: the packet is ONE SECOND newer than the newest commit --------
# The gap in days is 0 and the packet is still owed. A detector that reported only
# gaps of a whole day would lose every remand of the same working day.
pr 903; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-08-28T22:19:01Z" revision_packet - 0 - 1 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY

# --- A5  the NEWEST packet decides, and here it is discharged -------------------
# An older packet that was never worked, then a newer one that was. Reading "any
# packet newer than the newest commit" reports this PR owed forever.
pr 904; commits "2026-08-20T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-08-19T09:00:00Z" revision_packet - 0 - 1 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY
cmt "2026-09-05T22:23:00Z" revision_packet - 0 - 2 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY

# --- A6  the NEWEST packet decides, and here it is owed -------------------------
# A packet, a commit that discharged it, and a SECOND packet after that commit.
# Reading "the oldest packet" or "a packet with any commit after it" reports clear.
pr 905; commits "2026-08-20T09:00:00Z" "2026-09-01T10:00:00Z"
cmt "2026-08-19T09:00:00Z" revision_packet - 0 - 1 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY
cmt "$PACKET_8D" revision_packet - 0 - 2 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY

# --- A7  MARKER DISCIPLINE, direction one: prose is not a marker ----------------
# A comment whose heading says "Revision packet" and whose prose says everything a
# packet says, declaring NO key. It is not a packet. The class is `undeclared` —
# the detector cannot answer — and it is emphatically not `clear`. The HEADING
# record is what the 2026-09-06 survey's greps saw instead.
pr 906; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" - - 0 - - revision_packet - <<'BODY'
## Revision packet — cycle 4

This remand orders a revision packet. The required changes are listed below and
the packet's owner is `@implementation-engineer`.

| id | required change | owner |
|---|---|---|
| SEC-01 | invalidate the old refresh token | implementation-engineer |
BODY

# --- A8  MARKER DISCIPLINE, direction two: a heading is not a disqualifier ------
# A DECLARED packet whose body carries no heading at all. A detector keyed off
# heading text calls this PR clear and loses the remand. This is #154's "no packet
# found" reported as fact.
pr 907; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" revision_packet - 0 - 4 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 4
BODY

# --- A9  a relay never reclassifies its courier ---------------------------------
# The LEADING declaration is the artifact's own; a later one belongs to something
# the comment quotes verbatim. An orchestrator comment that reposts a packet is an
# orchestrator comment, so this PR has no declared packet of its own. The relayed
# `cycle: 4` IS this comment's first `cycle:` line and is transcribed as such —
# the first-declaration rule is per key, and the courier's own id is what decides
# that the cycle is never read.
pr 908; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$PACKET_8D" orchestrator_comment - 0 - 4 <<'BODY'
artifact: orchestrator_comment
prose_licence: none
note: relaying the judge's packet for the record

```yaml
artifact: revision_packet
prose_licence: none
cycle: 4
```
BODY

# --- B1a  owed: a convened panel with no ruling at all --------------------------
# #178 criterion 2 in its own words — "a convened panel with no ruling". No
# judgment of any identity exists on this PR.
pr 910; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 3 <<BODY
artifact: review_ledger
prose_licence: none
pr: 910
sha: $SHA181
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY

# --- B1b  clear: THE CONTRACT'S OWN ORDER, and the fixture cycle 1 got backwards -
# `pr-judge § Phase 9` posts the judgment, `§ Phase 10` posts the ledger, six
# seconds later on #181 cycle 1 and in that order on 9 of 9 ruled cycles. So the
# ledger is the NEWER artifact on every correctly ruled PR, and a detector asking
# for "a judgment after the panel" answers `owed` here — on the one shape that
# proves the panel WAS ruled. The previous version of this corpus asserted exactly
# that, which is why this fixture is the centre of the B1 set.
pr 911; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_JUDGMENT" judgment - 0 "$SHA181" 3 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA181
cycle: 3
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 3 <<BODY
artifact: review_ledger
prose_licence: none
pr: 911
sha: $SHA181
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY

# --- B1c  owed: the ruling that exists belongs to ANOTHER panel ------------------
# The same posting order as B1b, and the opposite answer — because the judgment
# declares the PREVIOUS cycle's identity. This is what B1b's clear rests on: not
# that a judgment exists, and not when it arrived, but that it declares THIS
# panel's `sha:` and `cycle:`. A detector asking only "does a judgment exist"
# reports clear and loses the unruled panel.
pr 912; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-04T10:00:00Z" judgment - 0 "$SHA_OLD" 2 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA_OLD
cycle: 2
decision: remand
BODY
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 3 <<BODY
artifact: review_ledger
prose_licence: none
pr: 912
sha: $SHA181
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY

# --- B1d  undeclared: no panel was ever declared on this PR ---------------------
pr 913; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-05T20:00:00Z" revision_packet - 0 - 1 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY

# --- B1e  clear: the ruling posted AFTER the panel, same identity ---------------
# The mirror of B1b, and the reason B1 must not simply INVERT the old rule. Order
# decides nothing in either direction: a judge that posts its ledger first has
# still ruled, and the identity both artifacts declare is what says so.
pr 914; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 3 <<BODY
artifact: review_ledger
prose_licence: none
pr: 914
sha: $SHA181
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY
cmt "2026-09-05T23:00:00Z" judgment - 0 "$SHA181" 3 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA181
cycle: 3
decision: approve
BODY

# --- B1f  clear: ONE COMMIT HOWEVER ITS SHA IS SPELLED --------------------------
# #135 cycle 7 declared `sha: d5f3a58` where another artifact carries the 40-hex
# form, and judges write `cycle: 07` as readily as `cycle: 7`. An identity test
# that called those different reports a ruled panel unruled. Both spellings in one
# fixture, because both are the same question: what makes two declarations one
# identity.
pr 915; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_JUDGMENT" judgment - 0 "$SHA135" 07 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA135
cycle: 07
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 0 "$SHA135_LONG" 7 <<BODY
artifact: review_ledger
prose_licence: none
pr: 915
sha: $SHA135_LONG
cycle: 7

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY

# --- B1g  undeclared: the panel declares no identity to match a ruling to -------
# The discharge class gets the third state the source classes have. A ruling this
# detector CANNOT MATCH is not a ruling it may report as missing: an undeclared
# artifact is invisible by construction (#172 criterion 2), so `owed` would assert
# an absence the marker discipline cannot see, and `clear` would claim a discharge
# nothing performed.
pr 916; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_LEDGER" review_ledger - 0 - - <<'BODY'
artifact: review_ledger
prose_licence: none
pr: 916

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY

# --- B1h  undeclared: a ruling exists and declares no identity ------------------
# The other half of B1g, and a different absence: the panel is matchable and the
# judgment is not. Same answer, and it must not be `owed` for the same reason.
pr 917; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_JUDGMENT" judgment - 0 - - <<'BODY'
artifact: judgment
prose_licence: none
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 3 <<BODY
artifact: review_ledger
prose_licence: none
pr: 917
sha: $SHA181
cycle: 3

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
BODY

# WHAT THE B1 SET DELIBERATELY DOES NOT ASSERT, said here so its absence is not
# read as an oversight. The detector compares the two artifacts' OWN declared
# SHAs and never the PR's current head — a force-push moves the head, and a test
# of the form "the panel's sha is the head" would discharge a genuinely unruled
# panel as stale history. That guard cannot be exercised at this interface: the
# commits call the detector was observed to make returns DATES and no shas at all,
# so a head-comparing detector cannot be built against this transport. Any detector
# that reached for a commit sha would ask the stubbed `gh` a question it does not
# serve, and the UNKNOWN guard at the foot of this file stops the run.

# --- B2a  owed: the #135 cycle-4 shape -----------------------------------------
# A panel naming the mandatory last lane, and no validation verdict after it.
# Caught in life only because a human noticed a 90-minute silence.
pr 920; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 4 <<BODY
artifact: review_ledger
prose_licence: none
pr: 920
sha: $SHA181
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
| validation | @validation-agent | pending |
BODY

# --- B2b  A JUDGMENT DOES NOT DISCHARGE B2, and it is asserted on four surfaces --
# PR #135 cycles 4 and 5 exactly: a panel was persisted, three verdicts returned,
# a ruling followed, and @validation-agent was never dispatched at all. A detector
# that let a ruling close the validation question reports both cycles complete.
#
# THE PACKET AND THE RULING ARE BOTH DISCHARGED HERE ON PURPOSE, so validation is
# the ONLY owed class on the PR. That isolation is what lets the S-block assert the
# state token, the exit status and the summary counts on this one fixture: with a
# second owed class present, a detector that lost B2 entirely would still print
# `owed`, still exit 1 and still count 1, and only the table cell would move — one
# red, which is exactly what this fixture produced on cycle 1. The judgment and the
# ledger are posted in the contract's order and declare one identity, so the ruling
# is `clear` on the same grounds as B1b.
pr 921; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T17:00:00Z" revision_packet - 0 - 4 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 4
BODY
cmt "$T_JUDGMENT" judgment - 0 "$SHA181" 4 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA181
cycle: 4
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 4 <<BODY
artifact: review_ledger
prose_licence: none
pr: 921
sha: $SHA181
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
| validation | @validation-agent | pending |
BODY

# --- B2c  clear: a validation verdict after the panel ---------------------------
# PROVISIONAL, AND THE ONLY KIND OF ASSERTION HERE THAT IS. Both keys are
# mandatory per `review-verdicts § Reviewer verdict` — but `@validation-agent` is
# forbidden to emit a `reviewer_verdict` of its own (`validation-agent.md:83`), so
# this discharge shape is unreachable in life and #182 is the item that declares
# what a validation result's artifact class actually is. Kept, marked, and NOT
# built on: no fixture added this cycle asserts a B2 discharge, and when #182 lands
# this fixture and U2 are the two that change.
pr 922; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 4 <<BODY
artifact: review_ledger
prose_licence: none
pr: 922
sha: $SHA181
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T22:00:00Z" reviewer_verdict validation-agent 1 - - <<'BODY'
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
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 4 <<BODY
artifact: review_ledger
prose_licence: none
pr: 923
sha: $SHA181
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T22:00:00Z" reviewer_verdict correctness 0 - - <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: correctness
status: valid_review
verdict: pass
BODY

# --- B2e  a validation verdict BEFORE the panel does not discharge it -----------
# The previous cycle's validation says nothing about this cycle's panel. Unlike B1,
# B2 IS a time test and stays one: #178 criterion 3 asks for a verdict AFTER the
# panel, and the two conditions differ because a verdict declares no panel identity
# to be matched by.
pr 924; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "2026-09-04T10:00:00Z" reviewer_verdict validation-agent 1 - - <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: validation-agent
status: valid_review
verdict: pass
BODY
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 5 <<BODY
artifact: review_ledger
prose_licence: none
pr: 924
sha: $SHA181
cycle: 5

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY

# --- B2f  a panel that does not name the mandatory lane owes no validation ------
# B2 is "a panel NAMING @validation-agent". This one does not, so there is nothing
# owed — and the panel itself is declared and ruled, so this is `clear` and not
# `undeclared`.
pr 925; commits "2026-08-27T09:00:00Z" "$LASTCOMMIT"
cmt "$T_JUDGMENT" judgment - 0 "$SHA181" 1 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA181
cycle: 1
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 1 <<BODY
artifact: review_ledger
prose_licence: none
pr: 925
sha: $SHA181
cycle: 1

| lane | reviewer | verdict |
|---|---|---|
| documentation | @docs-reviewer | pass |
BODY

# --- B2f'  the same panel, with the other two classes discharged too ------------
# B2f alone cannot show what `n/a` MEANS, because its packet class is undeclared
# and the run exits 3 for that reason instead. This PR discharges A and B1, so the
# only class left is the inapplicable one and the exit status is `n/a`'s alone.
pr 927; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T17:00:00Z" revision_packet - 0 - 1 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 1
BODY
cmt "$T_JUDGMENT" judgment - 0 "$SHA181" 1 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA181
cycle: 1
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 0 "$SHA181" 1 <<BODY
artifact: review_ledger
prose_licence: none
pr: 927
sha: $SHA181
cycle: 1

| lane | reviewer | verdict |
|---|---|---|
| documentation | @docs-reviewer | pass |
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
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 4 <<BODY
artifact: review_ledger
prose_licence: none
pr: 926
sha: $SHA181
cycle: 4

| lane | reviewer | verdict |
|---|---|---|
| correctness | @linus-quality-critic | pass |
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
  cmt "$(printf '2026-09-0%dT08:00:00Z' "$i")" orchestrator_comment - 0 - - <<'BODY'
artifact: orchestrator_comment
prose_licence: none
note: relaying the panel state for the record
BODY
  i=$((i + 1))
done
cmt "2026-09-10T22:23:00Z" - - 1 - - revision_packet - <<'BODY'
## Cycle 4 — required changes

Remanded. `@implementation-engineer` owns SEC-01; `@validation-agent` re-runs the
suite once the change lands. This comment predates the declared-key requirement
and declares nothing.
BODY

# --- U2  every class discharged -------------------------------------------------
# Carries the same provisional B2 discharge B2c does, and is marked for the same
# reason: #182 settles what artifact class a validation result declares.
pr 940; commits "2026-08-27T09:00:00Z" "2026-09-06T10:00:00Z"
cmt "2026-09-05T17:00:00Z" revision_packet - 0 - 2 <<'BODY'
artifact: revision_packet
prose_licence: none
cycle: 2
BODY
cmt "$T_JUDGMENT" judgment - 0 "$SHA181" 2 <<BODY
artifact: judgment
prose_licence: none
sha: $SHA181
cycle: 2
decision: approve
BODY
cmt "$T_LEDGER" review_ledger - 1 "$SHA181" 2 <<BODY
artifact: review_ledger
prose_licence: none
pr: 940
sha: $SHA181
cycle: 2

| lane | reviewer | verdict |
|---|---|---|
| validation | @validation-agent | pending |
BODY
cmt "2026-09-05T21:00:00Z" reviewer_verdict validation-agent 1 - - <<'BODY'
artifact: reviewer_verdict
prose_licence: none
reviewer: validation-agent
status: valid_review
verdict: pass
BODY

# --- X  unreadable sources ------------------------------------------------------
# 930: the commits endpoint fails.  931: the comments endpoint fails.
# Both are marked by the ABSENCE of the fixture file, which is what the stubbed
# `gh` turns into a non-zero exit — an API that answered with an error, not an API
# that answered with nothing.
pr 930
cmt "$T_LEDGER" review_ledger - 0 - - <<'BODY'
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
# It answers the three calls the detector was observed to make, and it computes
# NOTHING: the comment records are the ones the fixtures declared. Two comment
# queries are served — the declared-marker record (what `gh --jq "$COMMENT_JQ"`
# returns) and the heading-derived record (what a prose-matching detector's own
# program would return, so that one can be built and failed against this corpus).
# A query it does not recognise is recorded and stops the run: a stub that silently
# answered an unknown question with nothing would hand every detector an empty
# queue and call it clean.
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
      *'^artifact:'*|DECLARED) cat "$FIX/pr$n/records.declared"; exit 0 ;;
      HEADING)                 cat "$FIX/pr$n/records.heading";  exit 0 ;;
      *) echo "gh stub: unrecognised --jq program for $path" >> "$FIX/UNKNOWN"; exit 1 ;;
    esac ;;
esac
echo "gh stub: unrecognised api path: $path" >> "$FIX/UNKNOWN"
exit 1
STUB
chmod +x "$TMP/bin/gh"
export PATH="$TMP/bin:$PATH" GH="$TMP/bin/gh"

# ---------------------------------------------------------------------------
# CONSTRUCTION AND TRANSCRIPTION GUARDS
#
# A fixture that is not what it claims is exactly the vacuous instrument this file
# exists to prevent, so a miss is a hard stop.
#
# THE TRANSCRIPTION GUARD IS A CHECK ON THE AUTHOR AND NOT A SECOND COPY OF THE
# RULE. It re-reads each body with a plain `grep` and asks whether the record the
# fixture declared is the one the body's leading keys say it is. It is deliberately
# WEAKER than the detector's marker discipline — it does not validate a sha's
# length or a cycle's digits, and it matches `validation-agent` as a fixed string
# rather than on word boundaries — and NO DETECTOR EVER CONSUMES IT. Its failure
# mode is a typo in the table above; the rule's failure mode is not its business.
# ---------------------------------------------------------------------------
first_key() { # first_key <file> <key>  -> the value of the body's first `<key>:` line
  grep -m1 "^$2:" "$1" 2>/dev/null | sed "s/^$2:[[:space:]]*//" | awk '{print $1}'; }

for d in "$FIX"/pr*; do
  n="${d##*/pr}"
  [ -f "$d/NO_COMMENTS" ] && continue
  # Comments must be chronological, because every condition here is "X with nothing
  # after it" and a fixture whose records arrived out of order tests a different
  # question than the one it names.
  awk -v pr="$n" '
    $1 < prev { printf "pr %s: comment %d at %s precedes %s\n", pr, NR, $1, prev; bad = 1 }
    { prev = $1 }
    END { exit bad + 0 }' "$d/records.declared" \
    || die "fixture comments are not in chronological order (above)"

  k=0
  while IFS=' ' read -r ts id rev nval sha cyc; do
    k=$((k + 1)); f="$(printf '%s/c%02d.txt' "$d" "$k")"
    for pair in "artifact $id" "reviewer $rev" "sha $sha" "cycle $cyc"; do
      key="${pair%% *}"; want="${pair##* }"
      got="$(first_key "$f" "$key")"
      if [ "$want" = '-' ]; then
        [ -z "$got" ] || die "pr $n comment $k declares '$key: $got' and its record says it declares none"
      else
        [ "$got" = "$want" ] \
          || die "pr $n comment $k: record says $key=$want, the body's first '$key:' line says '${got:-<none>}'"
      fi
    done
    if grep -qF 'validation-agent' "$f"; then
      [ "$nval" = 1 ] || die "pr $n comment $k names validation-agent and its record says 0"
    else
      [ "$nval" = 0 ] || die "pr $n comment $k does not name validation-agent and its record says 1"
    fi
    # The heading record may differ ONLY in the artifact it identifies and that
    # artifact's reviewer; anything else drifting would make the prose stub's reds
    # attributable to something other than heading-matching.
    h="$(sed -n "${k}p" "$d/records.heading")"
    [ "$(printf '%s' "$h" | awk '{print $1, $4, $5, $6}')" = "$ts $nval $sha $cyc" ] \
      || die "pr $n comment $k: the heading record differs from the declared record outside the artifact id"
  done < "$d/records.declared"
done

# The fixtures whose whole point is WHICH id was extracted say so once more here,
# in the direction the 2026-09-06 survey got wrong in each case.
dec() { awk -v k="$2" 'NR == k {print $2}' "$FIX/pr$1/records.declared"; }
hdg() { awk -v k="$2" 'NR == k {print $2}' "$FIX/pr$1/records.heading"; }
[ "$(dec 906 1)" = '-' ] && [ "$(hdg 906 1)" = 'revision_packet' ] \
  || die "A7 must declare NOTHING and look like a packet to a heading-matcher"
[ "$(dec 907 1)" = 'revision_packet' ] && [ "$(hdg 907 1)" = '-' ] \
  || die "A8 must declare revision_packet and look like nothing to a heading-matcher"
[ "$(dec 908 1)" = 'orchestrator_comment' ] \
  || die "A9's courier must extract its OWN leading key, not the packet it relays"
[ "$(awk '$2 != "orchestrator_comment" {print $2}' "$FIX/pr135/records.declared" | tr -d '\n')" = '-' ] \
  || die "U1 must hold nine couriers and one comment declaring nothing — the 2026-09-07 board"
[ "$(awk '$2 == "review_ledger" {print $4}' "$FIX/pr920/records.declared")" = '1' ] \
  || die "B2a's ledger must NAME @validation-agent or the mandatory-lane condition is never reached"
[ "$(awk '$2 == "review_ledger" {print $4}' "$FIX/pr925/records.declared")" = '0' ] \
  || die "B2f's ledger must NOT name @validation-agent or it is the same fixture as B2a"
[ "$(awk '$2 == "reviewer_verdict" {print $3}' "$FIX/pr926/records.declared")" = '' ] \
  || die "B2g must declare NO reviewer_verdict — the over-report it pins is the ledger's table standing alone"
# B1's centre: the two artifacts of one panel declare ONE identity, and the
# judgment is the OLDER of the two — `pr-judge § Phase 9` before `§ Phase 10`.
[ "$(awk 'NR == 1 {print $2, $5, $6}' "$FIX/pr911/records.declared")" = "judgment $SHA181 3" ] \
  && [ "$(awk 'NR == 2 {print $2, $5, $6}' "$FIX/pr911/records.declared")" = "review_ledger $SHA181 3" ] \
  || die "B1b must be judgment-then-ledger of ONE identity or it is not the contract's own order"
[ "$(awk 'NR == 1 {print $5}' "$FIX/pr912/records.declared")" != "$SHA181" ] \
  || die "B1c's ruling must declare ANOTHER panel's identity or it is the same fixture as B1b"

cd "$TMP" || die "could not enter $TMP"

# ---------------------------------------------------------------------------
# ASSERTIONS
# ---------------------------------------------------------------------------
run() { OUT="$(bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }
flat() { printf '%s' "$OUT" | tr '\n' '|'; }
line_for() { printf '%s\n' "$OUT" | grep -E "^#$1[[:space:]]" | head -1; }

check_rc() { # check_rc <label> <expected>
  [ "$RC" = "$2" ] && pass "$1" || bad "$1" "expected rc $2, got rc $RC: $(flat)"; }
check_has() { # check_has <label> <substring>
  case "$OUT" in *"$2"*) pass "$1" ;; *) bad "$1" "expected to find '$2': $(flat)" ;; esac; }

# The per-PR line, matched tolerantly: the corpus asserts that a CLASS and a STATE
# co-occur in the PR's own line, which is the contract's "three states per class".
# Spacing and ordering are the detector's to choose; the pairing is not.
cls() { # cls <label> <pr> <packet-state> <ruling-state> <validation-state>
  L="$(line_for "$2")"
  if [ -z "$L" ]; then bad "$1" "no per-PR line for #$2: $(flat)"; return; fi
  for pair in "packet $3" "ruling $4" "validation $5"; do
    k="${pair%% *}"; v="${pair##* }"
    printf '%s' "$L" | grep -Eq "$k[[:space:]]*[=:][[:space:]]*$v([^a-z]|\$)" || {
      bad "$1" "expected $k=$v, got: $L"; return; }
  done
  pass "$1"
}
# THE PR'S OWN STATE TOKEN, which is a different claim from any class pair: it is
# what precedence resolved the three classes DOWN to, and a detector can get every
# cell right and still roll them up wrongly.
state_is() { # state_is <label> <pr> <state>
  L="$(line_for "$2")"; S="$(printf '%s' "$L" | awk '{print $2}')"
  [ "$S" = "$3" ] && pass "$1" || bad "$1" "expected state '$3', got '${S:-<none>}' in: ${L:-<no line for #$2>}"; }

# AN ABSENCE IS EVIDENCE ONLY FROM A RUN THAT COULD HAVE CARRIED THE THING ABSENT,
# and this corpus asserts absences in exactly one place — X15, on #135's own line.
# It is written inline there rather than as a helper because the scoping is the
# whole of the assertion: an absence read over the WHOLE report fails on the
# summary's `clear 0` tally, which is a count and not a verdict on any PR. That is
# the shape `dependents-declared-edges.sh` records as the #138 anti-pattern —
# an assertion that cannot fail, or one that fails for a reason it does not name.

READABLE='900 901 902 903 904 905 906 907 908 910 911 912 913 914 915 916 917 920 921 922 923 924 925 926 927 135 940'

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

# THE GAP, ASSERTED CLOCK-TOLERANTLY. #172 criterion 1 asks for "the gap in days"
# and does not settle which two stamps it spans, so the line now prints BOTH and
# labels each (CORE-07 is the item that settles the text). Only the fixed one is
# asserted by value: now-minus-packet moves every day this suite runs, and a corpus
# that pinned it would go red on the calendar rather than on a defect. Its LABEL is
# asserted, because the whole point of printing two numbers is that neither can be
# taken for the other.
check_has "A1  ... and the packet-minus-commit gap is the measured 8"       "8 day(s) past the newest commit"
check_has "A4  ... and a zero-day gap is reported rather than swallowed"    "0 day(s) past the newest commit"
check_has "A1  ... and the second gap is labelled, not left to be confused" "(now-minus-packet)"

# CONDITION B1 (#178 criterion 2) — a convened panel with no ruling of its own
# identity. Derived from `pr-judge § Phase 9` -> `§ Phase 10`; see the header.
cls "B1a a panel with no ruling at all is OWED"                             910 undeclared owed       n/a
cls "B1b THE CONTRACT'S ORDER: judgment then ledger, one identity, CLEAR"   911 undeclared clear      n/a
cls "B1c a ruling of ANOTHER panel's identity does not rule this one"       912 undeclared owed       n/a
cls "B1d no panel declared: the detector cannot answer"                     913 owed       undeclared undeclared
cls "B1e the ruling posted AFTER the panel rules it just the same"          914 undeclared clear      n/a
cls "B1f one commit however its sha is spelled, one cycle however written"  915 undeclared clear      n/a
cls "B1g a panel declaring no identity cannot be judged, and is not OWED"   916 undeclared undeclared n/a
cls "B1h a ruling declaring no identity cannot be matched, and is not OWED" 917 undeclared undeclared n/a

# CONDITION B2 (#178 criterion 3) — the mandatory last lane, silently skipped.
cls "B2a a panel naming @validation-agent with no verdict is OWED"          920 undeclared owed       owed
cls "B2b A JUDGMENT DOES NOT DISCHARGE B2 — #135 cycles 4 and 5"            921 clear      clear      owed
cls "B2c a reviewer_verdict from validation-agent DISCHARGES it (#182)"     922 undeclared owed       clear
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
cls "U2  every class discharged (#182 governs the third)"                   940 clear      clear      clear

# ---- S: an owed class is four claims, not one cell -----------------------------
# CELL, STATE TOKEN, EXIT STATUS, SUMMARY COUNTS. The old corpus asserted the first
# alone on B2, and a detector that lost the mandatory-lane condition entirely moved
# exactly one check. Each fixture below is run ALONE so the status and the counts
# are about it and nothing else.
run 921
cls       "S1  B2b isolated: validation is the only owed class"             921 clear clear owed
state_is  "S1  ... and the PR's own state token is owed"                    921 owed
check_rc  "S1  ... and the run exits 1, which is what a caller branches on"  1
check_has "S1  ... and the summary counts one owed"                         "owed 1"
check_has "S1  ... and counts nothing clear, so the roll-up moved with it"  "clear 0"
check_has "S1  ... and the owed block names the mandatory lane"             "@validation-agent"

run 900
state_is  "S2  an owed packet rolls up to an owed PR"                       900 owed
check_rc  "S2  ... exit 1"                                                   1
check_has "S2  ... summary owed 1"                                          "owed 1"

run 135
state_is  "S3  an unanswerable PR rolls up to undeclared"                   135 undeclared
check_rc  "S3  ... exit 3"                                                   3
check_has "S3  ... summary undeclared 1"                                    "undeclared 1"
check_has "S3  ... and nothing owed was invented to reach it"               "owed 0"

run 927
state_is  "S4  a PR whose only open class is inapplicable is clear"         927 clear
check_rc  "S4  ... exit 0"                                                   0
check_has "S4  ... summary clear 1"                                         "clear 1"

run 930
state_is  "S5  an unreadable source is named on its own line"               930 unreadable
check_rc  "S5  ... exit 2"                                                   2
check_has "S5  ... summary unreadable 1"                                    "unreadable 1"

# THE COUNTS ARE PER CLASS AND NEED NOT SUM TO THE PRs READ, which the consumer
# branches on (`engineering-lead § Phase 1`) and the detector's header promises.
# #913 is owed on A and unanswerable on B in one PR: a run that bucketed one state
# per PR reports `undeclared 0` here and never prints the undeclared block at all.
run 913
cls       "S6  one PR, owed on one class and unanswerable on another"       913 owed undeclared undeclared
check_has "S6  ... counted owed"                                            "owed 1"
check_has "S6  ... AND counted undeclared, from the same single PR"         "undeclared 1"
check_has "S6  ... over one PR read"                                        "1 PR"
check_has "S6  ... and the undeclared block is printed, not swallowed"      "undeclared:"

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
# The usage path reads nothing at all, so it has no queue to summarise, and the
# settled contract asks it for no summary: the line prints on every path THAT
# READ SOMETHING, and this run opened no queue whose counts would be honest to
# print. What it is held to instead is LOUDNESS — it names its own failure and
# takes the same 2 every unreadable source takes. That is the property the
# summary rule exists to protect, and it is what X11' asserts.
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
L135="$(line_for 135)"
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
