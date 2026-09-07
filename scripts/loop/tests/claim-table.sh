#!/usr/bin/env bash
# claim-table.sh — the falsifying suite for scripts/loop/claim.sh (issue #144).
#
# The builder shipped claim.sh saying "no test asserts any of this yet — nothing
# here is trusted until it can be falsified." This is that suite, written by a
# different agent than the script: the author/falsifier split is what stops a test
# being bent until it passes.
#
# What it pins, in the order `#144 § Acceptance` states it:
#   1. A second claim on a held lane is REFUSED, never granted.
#   2. A verdict is durable BEFORE its writer's pass ends — read by another
#      process while the writer lives, and still on record after it is killed -9.
#   3. A `paused` table refuses a NEW claim yet admits an already-claimed one.
# Then `release`, adversarially, because it is the one path that can readmit a
# second dispatch and is therefore the single hole in the design.
#
# Then the review board's own findings, each of which named a state in which the
# table asserted something untrue, and each of which is falsified here rather
# than described: one commit is one panel however its SHA is spelled (folded in
# from claim-table-known-defects.sh, which held these while they were red and is
# deleted in the same commit as the fold) · one table per repository, exercised
# across a REAL linked worktree · every verb naming the table it acted on · a
# ruling committing as one directory, so no failure of the script can leave the
# empty `verdict.d` that used to wedge four verbs · `release` recovering that
# wreckage instead of calling it finished · a manifest, so `complete` is
# complete against something · a verdict naming who FILED it as distinct from
# who held the lane · `lane_state`, the field a caller may branch on · and a
# control-character strip a stored ANSI escape cannot survive.
#
# Then cycle 3's findings, in the same form: a ruling the table CANNOT commit
# kept on disk and named rather than deleted · a pause that stops the table
# growing without stopping an already-opened lane recording · a manifest amended
# out loud, keeping the set it replaced whole · and an expectation this table
# INFERRED marked apart from one a claim recorded, in both status views.
#
# BOTH FAILURE DIRECTIONS ARE PINNED SEPARATELY, and that is the part that earns
# the suite. `claim` must fail toward REFUSING to dispatch; `verdict` must fail
# toward loudly NOT recording. An inversion of either is silent at runtime and
# restores exactly the defect #144 exists to close, so every degraded case asserts
# the DIRECTION — the word the caller branches on — and not merely a nonzero code.
#
# Hermetic: CLAIM_TABLE_DIR points at a throwaway root for every case, so the
# machine's real table under .claude/state/ is never read and never written. That
# is asserted at the end rather than assumed.
#
# NTFS: this host folds case, so a rc-only test of the lane and SHA
# canonicalisation would pass even if claim.sh did no folding at all. Those cases
# therefore assert the directory NAME on disk and the echoed field, both of which
# NTFS preserves, so they cannot pass on the filesystem's behalf.
#
# CRLF: scripts/loop/*.sh carry no eol attribute, so on a core.autocrlf=true host
# they check out CRLF — fingerprint-isolation.sh already is CRLF in this working
# tree. This file avoids multi-line quoted literals for that reason, and the CRLF
# section asserts the script under test answers identically in both forms.
#
# Usage: scripts/loop/tests/claim-table.sh    Exit: 0 all pass · 1 any failure
# CLAIM_SH overrides the script under test — claim-table-mutations.sh sets it to
# run this suite against a deliberately broken copy, which is how each assertion
# here is shown red.

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="${CLAIM_SH:-$DIR/../claim.sh}"
[ -r "$SCRIPT" ] || { printf 'cannot read the script under test: %s\n' "$SCRIPT" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# The table this suite must never touch. Computed exactly as claim.sh computes
# it — the parent of the COMMON git directory, resolved absolute — and no longer
# from `--show-toplevel`. The two answers differ in a linked worktree, which is
# where this file runs as often as not: the old line watched
# `<worktree>/.claude/state/review-claims`, a path claim.sh does not write, so
# the guard was watching an empty room while the real table sat elsewhere. That
# the two now agree is asserted below rather than assumed, because a hermeticity
# guard pointed at the wrong path is indistinguishable from a hermetic run.
_gc="$(git rev-parse --git-common-dir 2>/dev/null)"
[ -z "$_gc" ] || _gc="$(cd "$_gc" 2>/dev/null && pwd)"
if [ -n "$_gc" ]; then GITROOT="${_gc%/*}"; else GITROOT="$(pwd)"; fi
[ -n "$GITROOT" ] || GITROOT="/"
REALTABLE="$GITROOT/.claude/state/review-claims"
REAL_BEFORE=absent; [ -e "$REALTABLE" ] && REAL_BEFORE=present
REAL_N_BEFORE="$(ls -a "$REALTABLE" 2>/dev/null | wc -l | tr -d ' ')"

PR=144
SHA=deadbee
SHA40=00112233445566778899aabbccddeeff00112233
CR="$(printf '\r')"

fails=0; cases=0; CASE=0

fresh() { CASE=$((CASE + 1)); CLAIM_TABLE_DIR="$TMP/tbl$CASE"; export CLAIM_TABLE_DIR; }
row_of() { printf '%s' "$CLAIM_TABLE_DIR/pr-$1/$2/$3"; }
run() { OUT="$(bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }
# run, from a chosen working directory. Almost every case here is indifferent to
# the cwd, and `run` above leaves it wherever the suite was started for exactly
# that reason. `--lanes` is not indifferent to it: it is the one CALLER argument
# this script word-splits, and an unquoted expansion is GLOB-expanded as well as
# split, so a pattern in it is matched against the CALLER'S cwd. For that case
# the cwd is part of the input and has to be a directory this suite built rather
# than one it inherited.
run_in() { _rd="$1"; shift; OUT="$(cd "$_rd" && bash "$SCRIPT" "$@" 2>&1)"; RC=$?; }
flat() { printf '%s' "${1:-}" | tr '\n' '|'; }
has() { case "$1" in *"$2"*) return 0 ;; *) return 1 ;; esac; }
pass() { cases=$((cases + 1)); printf 'PASS  %s\n' "$1"; }
fail() { cases=$((cases + 1)); fails=$((fails + 1)); printf 'FAIL  %s\n' "$1"; }

# check — the last run must carry this exit code AND this substring.
check() {
  if [ "$RC" = "$2" ] && has "$OUT" "$3"; then pass "$1"
  else fail "$1 — want rc $2 with [$3], got rc $RC: $(flat "$OUT")"; fi
}
# refute — the last run must NOT carry this substring. This is how a direction is
# pinned: `claim: granted` must be ABSENT from every refusal, not merely
# outnumbered by other text in it.
refute() {
  if has "$OUT" "$2"; then fail "$1 — want [$2] ABSENT, got rc $RC: $(flat "$OUT")"
  else pass "$1"; fi
}
is() { if [ "$2" = "$3" ]; then pass "$1"; else fail "$1 — got [$2], want [$3]"; fi; }

section() { printf '\n-- %s\n' "$1"; }

# ---------------------------------------------------------------------------
section 'condition 1 — a second claim on a held lane is refused'
# ---------------------------------------------------------------------------
fresh
run claim $PR $SHA lane-a --owner judge-1
check 'claim on a free lane is granted'                       0  'claim: granted'
check '  and names the lane it granted'                       0  'lane: lane-a'

run claim $PR $SHA lane-a --owner judge-2
check 'a second claim on the held lane is REFUSED'            10 'reason: held'
refute '  and never reads as granted'                            'claim: granted'
check '  and names the holder it lost to'                     10 'holder: judge-1'

run claim $PR $SHA lane-b --owner judge-2
check 'a different lane in the same panel is free'            0  'claim: granted'
run claim $PR c0ffee1 lane-a --owner judge-2
check 'the same lane at another SHA is a different row'       0  'claim: granted'

# The atomic primitive, asserted rather than reasoned about. The builder measured
# twelve concurrent claimants producing one 0 and eleven 10s; nothing pinned it.
fresh
RD="$TMP/race1"; mkdir -p "$RD"
i=1
while [ "$i" -le 12 ]; do
  ( bash "$SCRIPT" claim $PR $SHA hot-lane --owner "c$i" > "$RD/out.$i" 2>&1
    printf '%s' $? > "$RD/rc.$i" ) &
  i=$((i + 1))
done
wait
g=0; r=0; other=0
for f in "$RD"/rc.*; do
  v="$(cat "$f" 2>/dev/null)"
  case "$v" in 0) g=$((g + 1)) ;; 10) r=$((r + 1)) ;; *) other=$((other + 1)) ;; esac
done
is 'twelve concurrent claimants: exactly one is granted'      "$g"     1
is '  and the other eleven are refused with 10'               "$r"     11
is '  and none returns any other code'                        "$other" 0
is '  and exactly one holder directory exists' \
   "$(ls -d "$(row_of $PR $SHA hot-lane)"/holder 2>/dev/null | wc -l | tr -d ' ')" 1

# ---------------------------------------------------------------------------
section 'condition 2 — a verdict is durable BEFORE its writer pass ends'
# ---------------------------------------------------------------------------
fresh
export MID="$TMP/mid.out" SCRIPT_X="$SCRIPT" PR_X=$PR SHA_X=$SHA
cat > "$TMP/writer.sh" <<'ENDWRITER'
#!/usr/bin/env bash
# A reviewer's pass: claim, record its verdict, then DIE without returning anything.
bash "$SCRIPT_X" claim "$PR_X" "$SHA_X" durable --owner reviewer >/dev/null 2>&1
bash "$SCRIPT_X" verdict "$PR_X" "$SHA_X" durable approved --conf 0.91 --findings 3 >/dev/null 2>&1
# Still alive on this line. A SEPARATE process must ALREADY see the ruling — that
# is what "durable before the pass ends" means, as against "returned at the end".
bash "$SCRIPT_X" status "$PR_X" "$SHA_X" durable > "$MID" 2>&1
kill -9 $$
ENDWRITER
bash "$TMP/writer.sh" >/dev/null 2>&1
WRC=$?
MIDOUT="$(cat "$MID" 2>/dev/null)"
if has "$MIDOUT" 'ruled' && has "$MIDOUT" 'approved'; then
  pass 'the ruling is on record while its writer is still alive'
else
  fail "the ruling is on record while its writer is still alive — want [ruled] and [approved] from a separate reader mid-pass, got: $(flat "$MIDOUT")"
fi
is '  and the writer then died without returning anything' \
   "$([ "$WRC" -ne 0 ] && printf died || printf returned)" died

run status $PR $SHA durable
check 'the verdict survives the writer being killed -9'       0  'approved'
check '  and the panel reads complete from the table alone'   0  'complete: true'
run status $PR $SHA
check '  and a whole-panel read sees it too'                  0  'complete: true'

VROW="$(row_of $PR $SHA durable)/verdict.d/row"
is '  and the row is on disk, not in a return value' \
   "$([ -r "$VROW" ] && printf yes || printf no)" yes
is '  carrying the ruling' \
   "$(grep '^verdict: ' "$VROW" 2>/dev/null | head -1)" 'verdict: approved'
is '  carrying the confidence' \
   "$(grep '^confidence: ' "$VROW" 2>/dev/null | head -1)" 'confidence: 0.91'

# ---------------------------------------------------------------------------
section 'condition 3 — paused refuses a NEW claim, admits an already-claimed one'
# ---------------------------------------------------------------------------
fresh
run claim $PR $SHA early --owner reviewer
check 'a lane claimed before the pause'                       0  'claim: granted'
run pause --reason 'owner halt'
check 'pause is set'                                          0  'paused: true'
check '  and records its reason'                              0  'reason: owner halt'

run claim $PR $SHA late --owner reviewer
check 'a NEW claim under pause is refused with 11'            11 'reason: paused'
refute '  and never reads as granted'                            'claim: granted'
is '  and leaves no row behind for the lane it refused' \
   "$([ -e "$(row_of $PR $SHA late)" ] && printf exists || printf absent)" absent

run verdict $PR $SHA early approved --conf 0.8
check 'an ALREADY-CLAIMED lane still records under pause'     0  'verdict: recorded'
refute '  and is not refused for being paused'                   'reason: paused'

# A PAUSE STOPS THE TABLE GROWING, AND `mkdir -p` GREW IT ANYWAY.
# `verdict` is deliberately not pause-gated, which is right — a lane already
# dispatched must be able to record. But `mkdir -p "$row"` built `pr-<n>/`, the
# SHA directory and the lane out of nothing, so a verdict naming a panel NOBODY
# HAD OPENED created one under the pause and then filled its only row; `list`
# then printed that panel `complete` off a single unclaimed verdict no judge
# ever dispatched, in the one state where the least should be believed about
# coverage. The gate is on CREATION and never on recording, so both halves are
# pinned: the row that exists still records (above), and the row this call would
# have to invent is refused.
run verdict $PR $SHA ghost approved --conf 0.5
check 'a verdict into a lane with NO row is refused under pause' 11 'verdict: NOT RECORDED'
refute '  and never reads as recorded'                              'verdict: recorded'
check '  telling the caller to carry it rather than lose it'     11 'carry this verdict in your handoff'
is '  and inventing no row for the lane it refused' \
   "$([ -e "$(row_of $PR $SHA ghost)" ] && printf exists || printf absent)" absent
run verdict 999 abc1234 ghost approved --conf 0.5
check '  nor a PANEL for a pr this table never opened'           11 'verdict: NOT RECORDED'
is '  which stays off the table entirely' \
   "$([ -e "$CLAIM_TABLE_DIR/pr-999" ] && printf exists || printf absent)" absent

run pause --reason 'second halt'
check 'a second pause is idempotent'                          0  'already paused'
is '  and the FIRST pause stands' \
   "$(grep '^reason: ' "$CLAIM_TABLE_DIR/PAUSED" 2>/dev/null | head -1)" 'reason: owner halt'

run paused
check 'the paused query reports 11 while paused'              11 'paused: true'
run status $PR $SHA
check 'status reports the pause even on a complete panel'     0  'paused: true'
check '  and the pause does not make a complete panel read incomplete' 0 'complete: true'

run resume
check 'resume lifts it'                                       0  'paused: false'
run claim $PR $SHA late --owner reviewer
check '  and the refused lane is now claimable'               0  'claim: granted'
run verdict $PR $SHA ghost approved --conf 0.5
check '  and the verdict the pause refused records once lifted' 12 'verdict: recorded'
run resume
check 'resume when not paused is a no-op, not an error'       0  'was not paused'
run paused
check '  and the paused query reports 0'                      0  'paused: false'

# ---------------------------------------------------------------------------
section 'release — the one path that can readmit a second dispatch'
# ---------------------------------------------------------------------------
fresh
run claim $PR $SHA stranded --owner dead-judge
check 'a lane is claimed and its judge strands it'            0  'claim: granted'

run release $PR $SHA stranded
check 'release with NO --reason is a usage error'             64 'usage: claim.sh release'
is '  and the claim is still held' \
   "$([ -d "$(row_of $PR $SHA stranded)/holder" ] && printf held || printf freed)" held

run release $PR $SHA stranded --reason ''
check 'release with an EMPTY --reason is refused'             64 'a reason is required'
is '  and the claim is still held' \
   "$([ -d "$(row_of $PR $SHA stranded)/holder" ] && printf held || printf freed)" held

run release $PR $SHA stranded --reason
check 'release with --reason and no value is refused'         64 'usage: claim.sh'
is '  and the claim is still held' \
   "$([ -d "$(row_of $PR $SHA stranded)/holder" ] && printf held || printf freed)" held

run release $PR $SHA stranded --reason 'judge stranded at 21:45Z'
check 'release with a stated reason succeeds'                 0  'release: done'
check '  and says the lane is claimable again'                0  'claimable again'

# The hole, opened and pinned deliberately: release is the ONLY readmission path.
run claim $PR $SHA stranded --owner fresh-judge
check 'after release the lane is claimable again'             0  'claim: granted'
run claim $PR $SHA stranded --owner third-judge
check '  and the hole closes behind itself'                   10 'reason: held'
refute '  with no second grant'                                  'claim: granted'

# The audit row is what makes the readmission accountable rather than silent.
AUD="$(ls -d "$(row_of $PR $SHA stranded)"/released-* 2>/dev/null | head -1)"
is 'the released claim is kept as an audit row' \
   "$([ -n "$AUD" ] && printf kept || printf destroyed)" kept
is '  whose directory name carries no colon (NTFS rejects it)' \
   "$(case "$(basename "${AUD:-x}")" in *:*) printf colon ;; *) printf clean ;; esac)" clean
is '  and reads released, not claimed' \
   "$(grep '^state: ' "${AUD:-/nonexistent}/row" 2>/dev/null | head -1)" 'state: released'
is '  with no stale claimed line left beneath it' \
   "$(grep -c '^state: claimed' "${AUD:-/nonexistent}/row" 2>/dev/null | tr -d ' ')" 0
is '  and records why it was released' \
   "$(grep -c '^release_reason: ' "${AUD:-/nonexistent}/row" 2>/dev/null | tr -d ' ')" 1

# A ruled lane is finished, not stranded. Releasing one would readmit a second
# ruling on a lane already of record — the #141 shape.
fresh
run claim $PR $SHA ruled-lane --owner r1
run verdict $PR $SHA ruled-lane approved --conf 0.9
check 'a lane is claimed and ruled'                           0  'verdict: recorded'
run release $PR $SHA ruled-lane --reason 'let me back in'
check 'release of a RULED lane is refused'                    10 'reason: ruled'
refute '  and does not report done'                              'release: done'
run status $PR $SHA ruled-lane
check '  and the ruling is still of record'                   0  'approved'
run claim $PR $SHA ruled-lane --owner r2
check '  and the lane cannot be re-dispatched'                10 'reason: ruled'

# A ruling interrupted mid-write — verdict.d present, its row never written. No
# path in claim.sh can produce this state any more, because a ruling is now
# committed by renaming a directory that already holds its row; the suite builds
# it by hand, which is what a table written before that change still carries and
# what a hand-edit leaves. It used to absorb the lane — four verbs refusing
# forever, with `release`, the verb `pr-judge` Phase 10 names as the remedy,
# answering that a lane that has ruled "is finished, not stranded" about a lane
# that had not ruled and could not be recovered by any verb at all.
fresh
run claim $PR $SHA inflight --owner r1
mkdir -p "$(row_of $PR $SHA inflight)/verdict.d"
run claim $PR $SHA inflight --owner r2
check 'a lane whose ruling is IN FLIGHT is not claimable'     10 'reason: ruled'
check '  saying the record is unreadable, not absent'         10 'unreadable'
refute '  and never re-dispatches a lane being ruled'            'claim: granted'
run status $PR $SHA
check '  and status counts it outstanding'                    10 'ruling-incomplete'
refute '  and never reports the panel complete'                  'complete: true'

run release $PR $SHA inflight --reason 'looks stuck to me'
check 'release RECOVERS it rather than calling it finished'   0  'release: done'
check '  and names what it cleared'                           0  'cleared: ruling-incomplete'
refute '  never calling an unruled lane ruled'                   'reason: ruled'
VAUD="$(ls -d "$(row_of $PR $SHA inflight)"/released-verdict-* 2>/dev/null | head -1)"
is '  keeping the interrupted ruling as an audit row' \
   "$([ -n "$VAUD" ] && printf kept || printf destroyed)" kept
is '  which names the state it was found in' \
   "$(grep '^state: ' "${VAUD:-/nonexistent}/row" 2>/dev/null | head -1)" 'state: released-ruling-incomplete'
is '  and why it was cleared' \
   "$(grep -c '^release_reason: ' "${VAUD:-/nonexistent}/row" 2>/dev/null | tr -d ' ')" 1
run status $PR $SHA inflight
check '  leaving the lane free rather than wedged'            10 'lane_state: free'
run claim $PR $SHA inflight --owner r3
check '  so the lane is claimable again — the wedge is gone'  0  'claim: granted'

# The same wreckage with NO holder under it — a reviewer killed before its claim
# was written, or one whose claim was already released. There is nothing to hand
# back and the recovery still happened, so a 12 here would report that nothing
# was done about a lane this call had just unwedged.
fresh
mkdir -p "$(row_of $PR $SHA orphaned)/verdict.d"
run release $PR $SHA orphaned --reason 'no holder, just wreckage'
check 'release clears wreckage with no claim under it'        0  'release: done'
check '  and says what it cleared there too'                  0  'cleared: ruling-incomplete'
run claim $PR $SHA orphaned --owner r1
check '  leaving that lane claimable as well'                 0  'claim: granted'

# The line the recovery must not cross. `release` refuses a READABLE verdict at
# 10, asserted above; what this pins is that the two are told apart by the ROW
# and not by the DIRECTORY — the distinction the old test lost, which is how one
# verb came to say "finished" about a lane that had not ruled.
fresh
run claim $PR $SHA ruled-not-stranded --owner r1
run verdict $PR $SHA ruled-not-stranded approved
run release $PR $SHA ruled-not-stranded --reason 'let me back in'
check 'a lane with a READABLE row is still refused at 10'     10 'reason: ruled'
is '  with its ruling untouched on disk' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA ruled-not-stranded)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: approved'
is '  and nothing moved aside from a lane that had ruled' \
   "$(ls -d "$(row_of $PR $SHA ruled-not-stranded)"/released-verdict-* 2>/dev/null | wc -l | tr -d ' ')" 0

fresh
run release $PR $SHA never-existed --reason 'x'
check 'release of a lane with no row at all is 12'            12 'no claim to release'
refute '  and does not report done'                              'release: done'

fresh
run claim $PR $SHA held-then --owner r1
run release $PR $SHA held-then --reason 'first'
run release $PR $SHA held-then --reason 'again'
check 'a second release of the same lane is 12'               12 'no claim to release'
run status $PR $SHA
check 'a released lane shows as outstanding debt'             10 'no verdict on record'
refute '  and never as a complete panel'                         'complete: true'

# release must not be a way around the pause.
fresh
run claim $PR $SHA under-pause --owner r1
run pause --reason 'halt'
run release $PR $SHA under-pause --reason 'freeing it'
check 'release still works while paused'                      0  'release: done'
run claim $PR $SHA under-pause --owner r2
check '  but the freed lane still cannot be re-claimed'       11 'reason: paused'
refute '  so release is no way around the pause'                 'claim: granted'

# ---------------------------------------------------------------------------
section 'failure direction 1 — claim fails toward REFUSING to dispatch'
# ---------------------------------------------------------------------------
# A table path occupied by a regular file: mkdir -p cannot make it, so the table
# is unwritable. Were this to read as an unclaimed lane, every judge dispatches.
fresh
: > "$CLAIM_TABLE_DIR"
run claim $PR $SHA lane-a --owner j1
check 'unwritable table: claim refuses'                       2  'claim: refused'
refute '  and NEVER reads as granted'                            'claim: granted'
check '  and names its direction explicitly'                  2  'refusing to dispatch'

# The table is fine; the row cannot be made.
fresh
mkdir -p "$CLAIM_TABLE_DIR"; : > "$CLAIM_TABLE_DIR/pr-$PR"
run claim $PR $SHA lane-a --owner j1
check 'unmakeable row: claim refuses'                         2  'claim: refused'
refute '  and NEVER reads as granted'                            'claim: granted'
check '  and says the row could not be created'               2  'could not create row'

# ---------------------------------------------------------------------------
section 'failure direction 2 — verdict fails toward loudly NOT recording'
# ---------------------------------------------------------------------------
fresh
: > "$CLAIM_TABLE_DIR"
run verdict $PR $SHA lane-a approved
check 'unwritable table: verdict says NOT RECORDED'           2  'verdict: NOT RECORDED'
refute '  and never claims to have recorded'                     'verdict: recorded'
check '  and tells the reviewer to carry it in the handoff'   2  'carry this verdict in your handoff'

fresh
mkdir -p "$CLAIM_TABLE_DIR"; : > "$CLAIM_TABLE_DIR/pr-$PR"
run verdict $PR $SHA lane-a approved
check 'unmakeable row: verdict says NOT RECORDED'             2  'verdict: NOT RECORDED'
refute '  and never claims to have recorded'                     'verdict: recorded'
check '  and tells the reviewer to carry it'                  2  'carry this verdict'

# The two directions are OPPOSITE, and that asymmetry is the contract. A uniform
# direction either dispatches twice or loses a verdict; this asserts neither.
fresh
: > "$CLAIM_TABLE_DIR"
run claim $PR $SHA both --owner j1; CDIR="$OUT"
run verdict $PR $SHA both approved;  VDIR="$OUT"
if has "$CDIR" 'refusing to dispatch' && has "$VDIR" 'carry this verdict in your handoff'; then
  pass 'on ONE degraded table the two verbs fail in opposite directions'
else
  fail "on ONE degraded table the two verbs fail in opposite directions — want claim [refusing to dispatch] and verdict [carry this verdict in your handoff], got claim: $(flat "$CDIR") and verdict: $(flat "$VDIR")"
fi

# A verdict for a lane no claim covers is DURABLE FIRST and loud second. Refusing
# it would lose the work in order to enforce a rule about dispatch.
fresh
run verdict $PR $SHA orphan rejected --artifact 'https://github.com/x/pull/144#orphan'
check 'an unclaimed verdict is still recorded'                12 'verdict: recorded'
check '  and is loud about the anomaly'                       12 'anomaly: no claim row covered'
is '  and the row is on disk' \
   "$([ -r "$(row_of $PR $SHA orphan)/verdict.d/row" ] && printf yes || printf no)" yes
is '  flagged unclaimed in the row itself' \
   "$(grep '^unclaimed: ' "$(row_of $PR $SHA orphan)/verdict.d/row" 2>/dev/null | head -1)" 'unclaimed: true'

# One ruling per lane per SHA. Two judges ruled on #141 sixteen seconds apart with
# contradictory packets; the second must be refused and the FIRST stay of record.
fresh
run claim $PR $SHA once --owner j1
run verdict $PR $SHA once approved --conf 0.9 --findings 2
check 'the first ruling is recorded'                          0  'verdict: recorded'
run verdict $PR $SHA once rejected --conf 0.4 --findings 20
check 'a SECOND ruling is refused'                            10 'already ruled'
refute '  and does not report itself recorded'                   'verdict: recorded'
check '  and the first is named as of record'                 10 'of_record: approved'
is '  and the table still holds the first' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA once)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: approved'
is '  with the first findings count, not the second' \
   "$(grep '^findings: ' "$(row_of $PR $SHA once)/verdict.d/row" 2>/dev/null | head -1)" 'findings: 2'

fresh
run claim $PR $SHA hotrule --owner j1
RD="$TMP/race2"; mkdir -p "$RD"
i=1
while [ "$i" -le 8 ]; do
  ( bash "$SCRIPT" verdict $PR $SHA hotrule "ruling-$i" > "$RD/out.$i" 2>&1
    printf '%s' $? > "$RD/rc.$i" ) &
  i=$((i + 1))
done
wait
g=0; r=0; other=0
for f in "$RD"/rc.*; do
  v="$(cat "$f" 2>/dev/null)"
  case "$v" in 0) g=$((g + 1)) ;; 10) r=$((r + 1)) ;; *) other=$((other + 1)) ;; esac
done
is 'eight concurrent rulings: exactly one records'            "$g"     1
is '  and the other seven are refused with 10'                "$r"     7
is '  and none returns any other code'                        "$other" 0
is '  and exactly one ruling is of record' \
   "$(grep -c '^verdict: ' "$(row_of $PR $SHA hotrule)/verdict.d/row" 2>/dev/null | tr -d ' ')" 1

# ---------------------------------------------------------------------------
section 'exit codes as contract — callers branch on these'
# ---------------------------------------------------------------------------
fresh
run claim $PR $SHA codes --owner j1;  check '0  · claim granted'              0  'granted'
run claim $PR $SHA codes --owner j2;  check '10 · claim held'                 10 'held'
run pause
run claim $PR $SHA other --owner j3;  check '11 · claim paused'               11 'paused'
run resume
run status 999 $SHA;                  check '12 · status, no such panel'      12 'no claim row here'
fresh; : > "$CLAIM_TABLE_DIR"
run status $PR $SHA;                  check '2  · degraded, unreadable table' 2  'degraded'
fresh
run frobnicate;                       check '64 · unknown subcommand'         64 'unknown subcommand'
run claim;                            check '64 · claim with no arguments'    64 'usage: claim.sh claim'
run claim $PR $SHA;                   check '64 · claim missing the lane'     64 'usage: claim.sh claim'
run claim 14x $SHA lane-a;            check '64 · pr is not digits'           64 'must be digits'
run claim -1 $SHA lane-a;             check '64 · pr is negative'             64 'must be digits'
run claim $PR abc12 lane-a;           check '64 · sha too short'              64 '7-40 hex'
run claim $PR "${SHA40}f" lane-a;     check '64 · sha too long'               64 '7-40 hex'
run claim $PR abcdefg lane-a;         check '64 · sha is not hex'             64 '7-40 hex'
run claim $PR $SHA40 lane-a;          check '0  · a 40-hex sha is accepted'   0  'granted'
run claim $PR abc1234 lane-a;         check '0  · a 7-hex sha is accepted'    0  'granted'
run claim $PR $SHA '../evil';         check '64 · lane traversal is refused'  64 '[a-z0-9._-]'
run claim $PR $SHA 'a/b';             check '64 · lane with a slash'          64 '[a-z0-9._-]'
run claim $PR $SHA '.hidden';         check '64 · lane starting with a dot'   64 '[a-z0-9._-]'
run claim $PR $SHA 'a..b';            check '64 · lane containing ..'         64 '[a-z0-9._-]'
run claim $PR $SHA lane-a --wat x;    check '64 · unknown flag'               64 'usage: claim.sh claim'
run claim $PR $SHA lane-a --owner;    check '64 · flag with no value'         64 'usage: claim.sh claim'
run verdict $PR $SHA lane-a;          check '64 · verdict with no ruling'     64 'usage: claim.sh verdict'
run verdict $PR $SHA lane-a '';       check '64 · verdict may not be empty'   64 'may not be empty'
run resume extra;                     check '64 · resume takes no arguments'  64 'usage: claim.sh resume'
run paused extra;                     check '64 · paused takes no arguments'  64 'usage: claim.sh paused'
run help;                             check '0  · help prints the surface'    0  'Exit codes:'
run list;                             check '0  · list always succeeds'       0  'table:'

# "Nothing was read and nothing was written" is a header claim; it is checkable.
fresh
run claim 14x $SHA lane-a
check '64 · a usage failure writes nothing'                   64 'must be digits'
is '  and creates no table' \
   "$([ -e "$CLAIM_TABLE_DIR" ] && printf created || printf absent)" absent

# The three states a panel can be in, and the code each returns.
fresh
run status $PR $SHA;                  check 'status: no rows is 12'           12 'nothing was dispatched'
run claim $PR $SHA s1 --owner j1
run status $PR $SHA;                  check 'status: outstanding is 10'       10 'complete: false'
check '  and names the outstanding lane'                      10 'outstanding: s1'
run verdict $PR $SHA s1 approved
run status $PR $SHA;                  check 'status: complete is 0'           0  'complete: true'

# ---------------------------------------------------------------------------
section 'NTFS — canonicalisation asserted on disk, not on the exit code'
# ---------------------------------------------------------------------------
# This host folds case, so `claim Foo` after `claim foo` is refused whether or not
# claim.sh normalises anything. These cases assert the directory NAME and the
# echoed field, which NTFS preserves, so they cannot pass on its behalf.
fresh
run claim $PR $SHA 'Go-Architecture-Critic' --owner j1
check 'a mixed-case lane is accepted'                         0  'claim: granted'
check '  and echoed lower-cased'                              0  'lane: go-architecture-critic'
is '  and written lower-cased on disk' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR/$SHA" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')" 'go-architecture-critic'
run claim $PR $SHA 'go-architecture-critic' --owner j2
check '  and the folded name is the same row'                 10 'reason: held'
is '  leaving exactly one row, not two' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR/$SHA" 2>/dev/null | wc -l | tr -d ' ')" 1

fresh
run claim $PR 'ABCDEF1' lane-a --owner j1
check 'a mixed-case sha is accepted'                          0  'claim: granted'
check '  and echoed lower-cased'                              0  '@ abcdef1'
is '  and written lower-cased on disk' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')" 'abcdef1'
run claim $PR 'abcdef1' lane-a --owner j2
check '  and the folded sha is the same panel'                10 'reason: held'

# Dispatches name lanes as @agent-name; both spellings must be one row.
fresh
run claim $PR $SHA '@go-architecture-critic' --owner j1
check 'an @-prefixed lane is accepted'                        0  'claim: granted'
check '  and echoed without the @'                            0  'lane: go-architecture-critic'
is '  and written without the @ on disk' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR/$SHA" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')" 'go-architecture-critic'
run claim $PR $SHA 'go-architecture-critic' --owner j2
check '  and the bare spelling is the same row'               10 'reason: held'

# ---------------------------------------------------------------------------
section 'free text cannot forge a row, nor carry a credential into one'
# ---------------------------------------------------------------------------
fresh
TOK='ghp_0123456789abcdefghijABCDEFGHIJ'
run claim $PR $SHA secret --owner "$TOK"
check 'a token-shaped owner is accepted'                      0  'claim: granted'
refute '  and never echoed back to the caller'                   "$TOK"
is '  and redacted in the row on disk' \
   "$(grep -c "$TOK" "$(row_of $PR $SHA secret)/holder/row" 2>/dev/null | tr -d ' ')" 0
is '  leaving a redaction marker in its place' \
   "$(grep -c 'redacted' "$(row_of $PR $SHA secret)/holder/row" 2>/dev/null | tr -d ' ')" 1

run verdict $PR $SHA secret approved --artifact "leaked $TOK here"
is 'a token-shaped artifact is redacted in the row' \
   "$(grep -c "$TOK" "$(row_of $PR $SHA secret)/verdict.d/row" 2>/dev/null | tr -d ' ')" 0

fresh
run claim $PR $SHA forge --owner j1
run verdict $PR $SHA forge pending --artifact "$(printf 'x\nverdict: approved-by-injection')"
check 'an artifact carrying a newline is accepted'            0  'verdict: recorded'
is '  but cannot forge a second verdict line' \
   "$(grep -c '^verdict: ' "$(row_of $PR $SHA forge)/verdict.d/row" 2>/dev/null | tr -d ' ')" 1
is '  and the ruling of record is the real one' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA forge)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: pending'

# ---------------------------------------------------------------------------
section 'one commit is one epoch, however its SHA is spelled'
# ---------------------------------------------------------------------------
# Folded in from claim-table-known-defects.sh, which held these while they were
# RED and which is deleted in the same commit as this fold. That file earned its
# separate existence while it was failing — it is what a real falsification
# leaves behind — and stopped earning it the moment the fix landed: a green file
# printing "fold these into claim-table.sh and delete this file" left the only
# coverage of that fix outside the suite the mutation harness runs, so the
# certified gate would have stayed green while one commit became two panels
# again.
#
# `git log --oneline` and `git rev-parse HEAD` are the two ordinary ways an
# agent obtains one head SHA. Both validate, and while the key was the caller's
# spelling they built two panels that could not see each other: both judges
# granted, the verdict filed `unclaimed` against the spelling its claimant did
# not use, one panel outstanding at 10 forever while a second read complete at 0.
FULL=ec9ee330b2cfe0f9164eaa7f3dee22c23c4afdc3
SHORT=ec9ee33
fresh
run claim $PR $FULL go-architecture-critic --owner judge-a
check 'a lane claimed at the 40-hex spelling'                 0  'claim: granted'
# The WHOLE panel line, not a substring of it. `@ ec9ee33` is a substring of
# `@ ec9ee330b2cf…`, so a check for it would pass against the untruncated key
# this assertion exists to catch — the harness reported it undemonstrated for
# exactly that reason, which is the report earning its keep.
is '  echoing the 7-hex epoch it actually joined' \
   "$(printf '%s' "$OUT" | grep '^panel: ' | head -1)" "panel: pr-$PR @ $SHORT"
run claim $PR $SHORT go-architecture-critic --owner judge-b
check 'the SAME commit spelled short is the SAME row'         10 'reason: held'
refute '  and is never granted a second dispatch'                'claim: granted'
is '  so the commit holds ONE panel, not two' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR" 2>/dev/null | wc -l | tr -d ' ')" 1
is '  keyed by the epoch and not by the spelling' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')" "$SHORT"

fresh
run claim $PR $SHORT correctness --owner judge-a
check 'a lane claimed at the 7-hex spelling'                  0  'claim: granted'
run claim $PR $FULL correctness --owner judge-b
check '  refuses the long spelling of the same commit'        10 'reason: held'

# The stranding consequence, which is what made this more than untidiness.
fresh
run claim $PR $SHORT correctness --owner judge
run verdict $PR $FULL correctness approved --conf 0.9
check 'a verdict at the other spelling covers the claim that exists' 0 'verdict: recorded'
refute '  and is not filed unclaimed against a panel of its own'    'unclaimed'
run status $PR $SHORT
check '  so the claiming judge sees its own panel complete'   0  'complete: true'

# One fix, both keys: `pr-0144` and `pr-144` split a panel the same way.
fresh
run claim 0144 abc1234 lane-a --owner judge-a
run claim 144  abc1234 lane-a --owner judge-b
check 'a leading-zero PR number is the same panel'            10 'reason: held'
is '  leaving one pr directory, not two' \
   "$(ls -d "$CLAIM_TABLE_DIR"/pr-* 2>/dev/null | wc -l | tr -d ' ')" 1

# ---------------------------------------------------------------------------
section 'one table per repository — a linked worktree is not a second panel'
# ---------------------------------------------------------------------------
# `--show-toplevel` answers with the CALLER's worktree and a linked worktree is
# its own toplevel, so a judge in the main checkout and a reviewer dispatched
# into ../TurfGPS-wt/<slug> built two panels for one PR at one SHA that could
# not see each other: the judge granted, the reviewer recorded `unclaimed` into
# the other table at 12, and the judge's panel read outstanding at 10 forever.
# That is CLAIM-01's shape through the PATH instead of the key.
#
# Four cwds are not enough on their own — two derivations can agree on a path
# and still disagree about a claim — so the mutual exclusion is exercised ACROSS
# the two checkouts as well. CLAIM_TABLE_DIR is unset throughout this section,
# because the derivation is the thing under test, and every call is made inside
# a throwaway repository built here rather than in this one.
WTREPO="$TMP/wt-repo"; WTLINK="$TMP/wt-linked"
mkdir -p "$WTREPO/sub"
git init "$WTREPO" >/dev/null 2>&1
: > "$WTREPO/sub/f"
git -C "$WTREPO" add -A >/dev/null 2>&1
git -C "$WTREPO" -c user.email=t@example.invalid -c user.name=t commit -m init >/dev/null 2>&1
git -C "$WTREPO" worktree add -b wt-branch "$WTLINK" >/dev/null 2>&1
is 'a throwaway repository with a linked worktree was built' \
   "$([ -d "$WTLINK/sub" ] && printf built || printf unbuilt)" built

# at <dir> <args…> — run the script from <dir> with no CLAIM_TABLE_DIR at all.
# `|| exit 3` is load-bearing: a cd that failed would run this against the
# machine's REAL table, which is the one thing this suite may never do.
at() { atd="$1"; shift; ( cd "$atd" 2>/dev/null || exit 3; unset CLAIM_TABLE_DIR; bash "$SCRIPT" "$@" 2>&1 ); }
WTTABLE="$WTREPO/.claude/state/review-claims"
wt1="$(at "$WTREPO"     list | head -1)"
wt2="$(at "$WTREPO/sub" list | head -1)"
wt3="$(at "$WTLINK"     list | head -1)"
wt4="$(at "$WTLINK/sub" list | head -1)"
is 'the main toplevel resolves the repository table'          "$wt1" "table: $WTTABLE"
is '  a main SUBDIRECTORY resolves the same one'              "$wt2" "$wt1"
is '  the LINKED WORKTREE resolves the same one'              "$wt3" "$wt1"
is '  and a linked-worktree subdirectory too'                 "$wt4" "$wt1"

OUT="$(at "$WTLINK" claim 7 abc1234 shared --owner reviewer-in-worktree --for reviewer-in-worktree)"; RC=$?
check 'a claim taken from the linked worktree is granted'     0  'claim: granted'
OUT="$(at "$WTREPO" claim 7 abc1234 shared --owner judge-in-main)"; RC=$?
check '  and the SAME lane from the main checkout is refused' 10 'reason: held'
refute '  never as a second panel of its own'                    'claim: granted'
check '  naming the holder that sits in the other checkout'   10 'holder: reviewer-in-worktree'
OUT="$(at "$WTREPO" verdict 7 abc1234 shared approved --by reviewer-in-worktree)"; RC=$?
check 'a verdict filed from the main checkout reaches that claim' 0 'verdict: recorded'
refute '  and is not filed unclaimed into a table of its own'    'unclaimed'
OUT="$(at "$WTLINK" status 7 abc1234)"; RC=$?
check '  so the worktree reads its own panel complete'        0  'complete: true'
is '  and exactly one table exists across both checkouts' \
   "$(find "$WTREPO" "$WTLINK" -type d -name review-claims 2>/dev/null | wc -l | tr -d ' ')" 1

# ---------------------------------------------------------------------------
section 'every verb names the table it acted on, as its FIRST line'
# ---------------------------------------------------------------------------
# A split table is what the derivation above prevents, and the cheapest moment
# to SEE one is the first line of the first reply — not at synthesis, where two
# panels are already built and the only evidence of the split is that neither is
# complete. Asserted as line 1 exactly: a `table:` anywhere in the output would
# pass while the caller still had to hunt for it. The verbs run in an order that
# puts three of them on a REFUSAL path, because that is the reply a split
# produces and therefore the one that must carry the line.
fresh
for verb in claim verdict release manifest status list pause paused resume; do
  case "$verb" in
    claim)    run claim $PR $SHA saystable --owner j1 ;;
    verdict)  run verdict $PR $SHA saystable approved ;;
    release)  run release $PR $SHA saystable --reason 'it has ruled' ;;
    manifest) run manifest $PR $SHA ;;
    status)   run status $PR $SHA ;;
    list)     run list ;;
    pause)    run pause --reason 'halt' ;;
    paused)   run paused ;;
    resume)   run resume ;;
  esac
  is "$verb names its table on line 1" \
     "$(printf '%s' "$OUT" | head -1)" "table: $CLAIM_TABLE_DIR"
done

# ---------------------------------------------------------------------------
section 'a ruling commits as ONE directory — the gate cannot succeed empty'
# ---------------------------------------------------------------------------
# #144's failure class 1, verbatim: a `mkdir` gate followed by a write is two
# steps, and a process that died between them left `verdict.d` present and
# empty. That state absorbed the lane — `claim` refused it as ruled, `verdict`
# refused it as already ruled, `status` counted it outstanding forever, and
# `release` refused it saying a lane that has ruled "is finished, not stranded",
# which asserted the opposite of the truth. Four verbs refusing forever, and the
# only recovery was `rm -rf` of the ledger.
#
# The ruling is now the rename of a directory that already holds its row. This
# section falsifies "the state has no way to arise", which is a claim about the
# FAILURE paths — so it takes the two this suite can construct, and then the
# interruption itself, which neither of them reaches.

# 1. The wreckage a table written before that change still carries is REPAIRED
#    by the next ruling instead of absorbing the lane. This is also what pins
#    `mv -T`: a plain `mv` onto an existing empty directory does not fail, it
#    moves the staging directory INSIDE, which reads as a ruled lane holding no
#    verdict — the wedge rebuilt by the verb meant to clear it.
fresh
run claim $PR $SHA repair --owner r1
mkdir -p "$(row_of $PR $SHA repair)/verdict.d"
run verdict $PR $SHA repair approved --conf 0.77
check 'a ruling onto empty wreckage is recorded'              0  'verdict: recorded'
is '  and a READABLE row stands where the wreckage was' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA repair)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: approved'
is '  the ruling directory holding the row and nothing else' \
   "$(ls "$(row_of $PR $SHA repair)/verdict.d" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')" 'row'
run status $PR $SHA repair
check '  so the lane reads ruled, not ruling-incomplete'      0  'lane_state: ruled'

# 2. Every degraded exit of `verdict` leaves NO ruling directory at all. A
#    verdict that was not written must not leave behind the name readers test.
fresh
: > "$CLAIM_TABLE_DIR"
run verdict $PR $SHA degraded-a approved
check 'unwritable table: the verdict refuses'                 2  'verdict: NOT RECORDED'
is '  leaving no ruling directory anywhere' \
   "$(find "$CLAIM_TABLE_DIR" -type d -name 'verdict.d' 2>/dev/null | wc -l | tr -d ' ')" 0
fresh
mkdir -p "$CLAIM_TABLE_DIR"; : > "$CLAIM_TABLE_DIR/pr-$PR"
run verdict $PR $SHA degraded-b approved
check 'unmakeable row: the verdict refuses'                   2  'verdict: NOT RECORDED'
is '  leaving no ruling directory either' \
   "$(find "$CLAIM_TABLE_DIR" -type d -name 'verdict.d' 2>/dev/null | wc -l | tr -d ' ')" 0

# 3. And the interruption, which is the case the two above cannot reach: a
#    writer killed -9 partway through. One full call is timed first and the
#    kills are swept across a window measured in FRACTIONS of it, from 60% to
#    170%, so they land where the two-step gate was vulnerable rather than
#    before the script has done anything. A sweep in fixed wall-clock seconds
#    drifts into "red for the wrong reason" on a faster or a slower host — and a
#    sweep in fractions still can, because the calibrating call is one sample:
#    measured on 2026-08-30, a run in which every kill landed before the commit
#    reported nothing at all about the gate. So the third assertion below asks
#    whether the sweep reached the commit window, the sweep is repeated once at
#    twice the delays when it did not, and neither is a pass on its own. The
#    invariant itself is absolute and is asserted over every trial rather than
#    on average: a ruling directory that exists holds a readable row.
sweep_kills() { # sweep_kills <base-ms> <lane-prefix>
  swept_wrecked=0; swept_landed=0
  k=1
  while [ "$k" -le 12 ]; do
    bash "$SCRIPT" verdict $PR $SHA "$2-$k" approved --conf 0.5 >/dev/null 2>&1 &
    kp=$!
    kms=$(( $1 * (50 + k * 10) / 100 ))
    sleep "$(printf '%d.%03d' $((kms / 1000)) $((kms % 1000)))"
    kill -9 "$kp" 2>/dev/null
    wait "$kp" 2>/dev/null
    kd="$(row_of $PR $SHA "$2-$k")"
    if [ -d "$kd/verdict.d" ]; then
      if [ -r "$kd/verdict.d/row" ]; then swept_landed=$((swept_landed + 1))
      else swept_wrecked=$((swept_wrecked + 1)); fi
    fi
    k=$((k + 1))
  done
}
fresh
KB="$(date +%s%N 2>/dev/null | tr -cd '0-9')"; [ -n "$KB" ] || KB=0
run verdict $PR $SHA calibrate approved
KE="$(date +%s%N 2>/dev/null | tr -cd '0-9')"; [ -n "$KE" ] || KE=0
TFULL=$(( (KE - KB) / 1000000 ))
[ "$TFULL" -ge 20 ] && [ "$TFULL" -le 5000 ] || TFULL=400
sweep_kills "$TFULL" killed
if [ $((swept_landed + swept_wrecked)) -eq 0 ]; then
  sweep_kills $((TFULL * 2)) killed-again
fi
is 'twelve killed writers leave NO empty ruling directory'    "$swept_wrecked" 0
is '  and no lane in the panel reads ruling-incomplete' \
   "$(bash "$SCRIPT" status $PR $SHA 2>&1 | grep -c 'ruling-incomplete' | tr -d ' ')" 0
is '  with the sweep reaching the commit window at all' \
   "$([ $((swept_landed + swept_wrecked)) -ge 1 ] && printf reached || printf missed)" reached

# ---------------------------------------------------------------------------
section 'a ruling that CANNOT be committed is kept, never deleted'
# ---------------------------------------------------------------------------
# `mv -T` refuses onto a non-empty directory, and that refusal IS the write-once
# gate. It also refuses for reasons that are not that gate — a plain FILE
# standing where `verdict.d` belongs is the one this suite can construct — and
# every one of them was read as `already ruled` at 10, permanently, because
# nothing about a file in the way changes on a retry. The staged row, the only
# copy of that verdict, had been `rm -rf`'d one line above the message telling
# the caller to run `release`, which clears a lane and cannot bring back what
# was deleted. A verdict destroyed by the verb whose one job is durability.
#
# THE EXIT CODE IS THE SMALL HALF OF THIS. What made it a lost verdict rather
# than a wrong answer is the deletion, so the assertions below follow the
# printed path onto disk and read the ruling out of it.
fresh
run claim $PR $SHA blocked --owner pr-judge
: > "$(row_of $PR $SHA blocked)/verdict.d"
run verdict $PR $SHA blocked approved --conf 0.91 --by blocked
check 'a ruling blocked by a FILE at verdict.d is NOT RECORDED'  2 'verdict: NOT RECORDED'
refute '  and is never called the ordinary already-ruled refusal' 'reason: already ruled'
refute '  nor read as recorded'                                   'verdict: recorded'
check '  naming what actually stands in the way'                 2 'is not a ruling directory'
check '  and naming where the ruling it could not commit is'     2 'staged_ruling_is_at:'
STAGED="$(printf '%s' "$OUT" | sed -n 's/^staged_ruling_is_at: //p' | head -1)"
is '  the path it printed being a directory that really exists' \
   "$([ -d "$STAGED" ] && printf exists || printf absent)" exists
is '  holding the ruling that was not committed' \
   "$(grep '^verdict: ' "$STAGED/row" 2>/dev/null | head -1)" 'verdict: approved'
is '  with the confidence its filer passed, re-fileable by hand' \
   "$(grep '^confidence: ' "$STAGED/row" 2>/dev/null | head -1)" 'confidence: 0.91'
# And it stays degraded on a retry rather than curdling into a permanent 10:
# the blocking file does not change, so a caller that reads 10 here believes a
# ruling of record exists and stops carrying its own.
run verdict $PR $SHA blocked approved --conf 0.91 --by blocked
check '  a retry says the same thing, and still not `already ruled`' 2 'verdict: NOT RECORDED'
refute '  which is what a caller must never mistake for a ruling'    'of_record:'

# ---------------------------------------------------------------------------
section 'manifest — complete means complete AGAINST something'
# ---------------------------------------------------------------------------
# Nothing recorded which lanes were selected, so `complete` was a claim about
# the rows that happened to exist and not a claim about coverage at all. A judge
# that claimed 2 of the 7 lanes it selected and died mid-selection left a panel
# reading `complete: true` at rc 0 the moment those two ruled; the next judge
# follows "0 outstanding, synthesise" and publishes a two-lane ledger for a
# seven-lane board, which Phase 10 then makes of record. That is #144 failure
# class 4 — a ledger under-reporting lanes while asserting coverage.
fresh
run manifest $PR $SHA
check 'a panel with no manifest says so, at 12'               12 'manifest: none'
refute '  and never invents a set nobody recorded'               'count:'
run manifest $PR $SHA --lanes 'correctness @Security-Critic docs testing safety ux design' --by judge-1
check 'the selected set is recorded'                          0  'manifest: recorded'
check '  counting every lane selected'                        0  'count: 7'
check '  canonicalised exactly as a claim would be'           0  'lanes: correctness security-critic docs'
run manifest $PR $SHA
check 'the set reads back'                                    0  'count: 7'
check '  naming who selected it'                              0  'selected_by: judge-1'

run claim   $PR $SHA correctness --owner j1
run verdict $PR $SHA correctness approved --conf 0.9
run claim   $PR $SHA security-critic --owner j2
run verdict $PR $SHA security-critic approved --conf 0.8
run status $PR $SHA
check 'two ruled of seven selected is NOT a complete panel'   10 'complete: false'
refute '  and never reads complete'                              'complete: true'
check '  naming a selected lane no row was ever made for'     10 'outstanding: docs'
check '  counting the set it was measured against'            10 'lanes: 7'
run status $PR $SHA docs
check '  a selected lane with no row has a state of its own'  10 'lane_state: never-claimed'
run list $PR
check '  and list counts it into the same seven'              0  'lanes 7'
check '  reporting the panel incomplete'                      0  'incomplete'

# Written once, like a verdict, and for the same reason: a selection that can be
# rewritten is a selection that can be shrunk to fit whatever actually ruled.
run manifest $PR $SHA --lanes 'correctness security-critic'
check 'a SECOND selection is refused'                         10 'already recorded'
refute '  and does not report itself recorded'                   'manifest: recorded'
check '  the first staying of record'                         10 'of_record: correctness security-critic docs'
run status $PR $SHA
check '  so the panel is still measured against seven'        10 'lanes: 7'

# WRITTEN ONCE IS A PANEL THAT CAN WEDGE, AND `rm -rf` WAS THE ONLY WAY OUT.
# A lane the manifest selected and nothing ever claimed counts outstanding
# forever: `status` refuses at 10, `complete` never arrives, and no verb could
# take that lane back out of the set — leaving the unaudited hand-edit of the
# artefact whose integrity is the whole point, which is the corner `verdict` was
# in before `release`. `--amend --reason` is the way out that leaves evidence.
# What replaces immutability is not weaker, it is DIFFERENT: a reason refused at
# 64 when absent, a new row citing what it replaced by name and by count, and
# the prior set kept WHOLE rather than deleted. A shrink is still possible; a
# shrink nobody can see is not, and that was the property worth keeping.
run manifest $PR $SHA --lanes 'correctness security-critic' --amend
check 'an amendment with no reason is a usage error'          64 'requires a reason'
run manifest $PR $SHA --amend --reason 'docs was never dispatched'
check 'an amendment naming no new set is a usage error'       64 'must name the new set'
run manifest $PR $SHA --lanes 'correctness security-critic' --reason 'no flag'
check 'a reason without --amend is a usage error'             64 'means nothing without --amend'
is '  and not one of the three touched the set of record' \
   "$(grep '^count: ' "$CLAIM_TABLE_DIR/pr-$PR/$SHA/.manifest.d/row" 2>/dev/null | head -1)" 'count: 7'

run manifest $PR $SHA --lanes 'correctness security-critic' --amend --reason 'docs was never dispatched' --by judge-2
check 'an amendment with a reason supersedes the selection'   0  'manifest: amended'
check '  naming the set it replaced, and its count'           0  'superseded: correctness security-critic docs'
check '  and where that set is kept'                          0  'prior_is_at:'
PRIOR="$(printf '%s' "$OUT" | sed -n 's/^prior_is_at: //p' | head -1)"
is '  the prior selection kept WHOLE, not deleted' \
   "$(grep '^lanes: ' "$PRIOR/row" 2>/dev/null | head -1)" \
   'lanes: correctness security-critic docs testing safety ux design'
is '  under an audit key of its own, saying why it was replaced' \
   "$(grep '^superseded_reason: ' "$PRIOR/row" 2>/dev/null | head -1)" \
   'superseded_reason: docs was never dispatched'
run status $PR $SHA
check '  so the wedged panel completes against the new set'   0  'complete: true'
refute '  and the dropped lane is outstanding no longer'         'outstanding: docs'
run manifest $PR $SHA
check '  the set of record being the amended one'             0  'count: 2'
check '  which says out loud that it superseded another'      0  'amended: yes'
check '  and why, where a reader of the manifest will see it' 0  'amend_reason: docs was never dispatched'

# The shrink is still not available quietly: an ordinary second `--lanes`
# refuses exactly as it did before, so nothing amends by accident.
run manifest $PR $SHA --lanes 'correctness'
check 'a plain second selection STILL refuses at 10'          10 'already recorded'
refute '  and never reports itself amended'                      'manifest: amended'
check '  naming the flag that is the only way past it'        10 '--amend --reason'
is '  the amended set staying of record' \
   "$(grep '^count: ' "$CLAIM_TABLE_DIR/pr-$PR/$SHA/.manifest.d/row" 2>/dev/null | head -1)" 'count: 2'

# An amendment with nothing to amend is asked before anything is created, so it
# leaves the table exactly as it found it rather than writing a first selection
# under a flag that claims to be replacing one.
fresh
run manifest $PR $SHA --lanes 'correctness' --amend --reason 'nothing here yet'
check 'an amendment on a panel with NO selection is refused'  12 'nothing to amend'
refute '  and never reports itself recorded'                     'manifest: recorded'
is '  having created no table at all' \
   "$([ -e "$CLAIM_TABLE_DIR" ] && printf created || printf absent)" absent

# One bad name refuses the whole set and writes nothing: a manifest that
# silently dropped a lane would be a set asserting coverage it does not have.
fresh
run manifest $PR $SHA --lanes 'ok ../evil'
check 'one malformed lane refuses the whole selection'        64 '[a-z0-9._-]'
is '  and writes nothing at all' \
   "$([ -e "$CLAIM_TABLE_DIR" ] && printf created || printf absent)" absent
run manifest $PR $SHA --lanes '   '
check 'an empty selection is refused'                         64 'may not be empty'

# THE SET IS WORD-SPLIT AND IT IS NOT GLOBBED, WHICH ARE TWO EXPANSIONS AND NOT
# ONE. claim.sh has three unquoted `for … in $x` loops and this is the only one
# reading a CALLER argument — the other two split a `lanes:` field `norm_lane`
# gated on the way in, where no pattern character can have survived. This one
# wants the split and nothing else; unquoted, the same word is also matched
# against the CALLER'S cwd, so `--lanes '*'` recorded whatever files happened to be
# sitting in the directory the caller ran from as the selected panel. That is a
# lane set derived from a directory listing, written once, that `complete` is
# then measured against for the life of the panel — coverage asserted over names
# nobody selected. `set -f` around the loop is the whole of the control, and
# nothing above can see it: every other case here passes a set with no pattern
# in it, so the glob never fires and the guard is never exercised.
#
# The cwd is therefore built and populated, because a case that ran from an
# EMPTY directory would leave `*` unmatched, bash would pass the pattern through
# unchanged, and the refusal would arrive for the right reason with the control
# removed. The decoy is named as a lane `norm_lane` would accept, so the failure
# under a missing guard is a set quietly RECORDED rather than a second refusal.
fresh
mkdir -p "$TMP/glob$CASE" 2>/dev/null
: > "$TMP/glob$CASE/decoy-lane"
run_in "$TMP/glob$CASE" manifest $PR $SHA --lanes '*'
check 'a pattern in --lanes is a malformed lane, not a listing' 64 '[a-z0-9._-]'
refute '  and nothing is recorded from the caller working directory' 'manifest: recorded'
refute '  the file sitting there never becoming a selected lane'     'decoy-lane'
is '  and no table is created for it either' \
   "$([ -e "$CLAIM_TABLE_DIR" ] && printf created || printf absent)" absent

# A panel with no manifest behaves exactly as it did before: the table does not
# start refusing panels written by a caller that has not been taught to select.
fresh
run claim $PR $SHA solo --owner j1
run verdict $PR $SHA solo approved
run status $PR $SHA
check 'an unmanifested panel still completes on its own rows' 0  'complete: true'
check '  saying it has no expected set'                       0  'manifest: none'

# ---------------------------------------------------------------------------
section 'the amendment between-state is NAMED, and is never a complete panel'
# ---------------------------------------------------------------------------
# `--amend` renames the prior set aside BEFORE committing its replacement,
# because the write-once gate refuses a name that already holds a record. So
# between the two renames the panel holds a superseded set and nothing at
# `.manifest.d`, and every read verb tested for `.manifest.d` alone: the window
# read as a panel that had never selected anything, and `complete` fell back to
# counting the rows that happen to exist. Measured on this host on 2026-09-07,
# on the panel this case builds, against read verbs that test for `.manifest.d`
# alone — which is what `claim-table-mutations.sh M54` restores, so the figures
# are reproducible rather than remembered. Outside the window: `lanes: 3 ·
# ruled: 2 · outstanding: 1` and `complete: false` at 10. Inside it: `manifest:
# none`, `lanes: 2 · ruled: 2 · outstanding: 0` and `complete: true` at 0, with
# `list` printing that same panel `lanes 2 ruled 2 complete` at 0. That is
# failure class 4 arriving through a state that lasts two renames — a two-lane
# answer for a panel whose set of record names three, and the exact reply that
# tells a judge to synthesise.
#
# WHAT IS ASSERTED IS THAT `complete: true` CANNOT APPEAR, not merely that the
# new label does. The regression was a false COMPLETENESS; a case that checked
# only for `amend-in-flight` would pass while `complete: true` stood two lines
# above it, which is how the shipped defect would survive its own test.
#
# The window is entered by taking `.manifest.d` off a panel that has REALLY been
# amended, so the superseded directory left behind is one this script named. A
# case that spelled the marker itself would stay green if the prefix
# `cmd_manifest` writes and the prefix `amend_in_flight` globs for ever stopped
# being the same string, and that divergence is silent in every other case here.
fresh
run manifest $PR $SHA --lanes 'alpha beta gamma delta' --by judge-1
run claim   $PR $SHA alpha --owner j1 --for j1
run verdict $PR $SHA alpha approved --by j1
run claim   $PR $SHA beta  --owner j1 --for j1
run verdict $PR $SHA beta  approved --by j1
run status $PR $SHA
check 'four selected and two ruled is an incomplete panel'    10 'complete: false'
run manifest $PR $SHA --lanes 'alpha beta gamma' --amend --reason 'delta was never dispatched' --by judge-2
check '  and an amendment leaves a superseded set on disk'    0  'manifest: amended'
PANEL="$CLAIM_TABLE_DIR/pr-$PR/$SHA"

# A COMPLETED amendment is not the between-state. The superseded set is still
# there and a set of record is back beside it, so a predicate that fired on the
# marker alone would degrade every panel that had ever been amended — the false
# positive that would make the new state useless within one board.
run status $PR $SHA
check '  a completed amendment is an ordinary panel'          10 'outstanding: gamma'
refute '  and is never reported as an amendment in flight'       'amend-in-flight'

# The window itself. Nothing about this panel changes but the presence of the
# set of record, which is what the window IS.
mv "$PANEL/.manifest.d" "$TMP/held$CASE"
run status $PR $SHA
check 'in the window status is degraded, at 2'                2  'manifest_state: amend-in-flight'
check '  naming the state in the manifest field too'          2  'manifest: amend-in-flight'
check '  and answering the completeness question FALSE'       2  'complete: false'
refute '  and NEVER complete: true, which is what shipped'       'complete: true'
refute '  nor a panel that merely selected nothing'              'manifest: none'
check '  telling the caller not to read these rows as coverage' 2 'do not read these rows as coverage'

run list $PR
check 'in the window list is degraded, at 2'                  2  'amend-in-flight on 1 panel(s)'
is '  the panel carrying the state where its verdict would go' \
   "$(printf '%s' "$OUT" | sed -n 's/^  pr-.* //p' | head -1)" 'amend-in-flight'

run manifest $PR $SHA
check 'in the window the manifest read verb is degraded, at 2' 2 'manifest: amend-in-flight'
refute '  and never calls it a panel that selected nothing'      'manifest: none'
check '  saying it is between two selections, not without one' 2 'not without one'

# THE WRITE VERBS IN THE WINDOW, WHICH IS THE PATH THAT ACTUALLY SHIPPED. Every
# assertion made INSIDE the window above this line is a read. A mutation matrix
# scoring every one of them would still have passed a script whose WRITER walked
# into the window, because nothing here had ever called `--lanes` or `--amend`
# with the set of record moved aside — and the verb that opens this state is the
# same verb that walks back into it.
#
# WHAT IS ASSERTED IS AGAIN THE CONSEQUENCE AND NOT THE GUARD. A case checking
# only that a refusal is printed stays green against a script that prints one
# and writes anyway, and green against one that refuses for an unrelated reason.
# Both leave the panel answering `complete: true` over a set it has not
# fulfilled, which is the failure this section exists to make unreachable. So
# each refusal below is followed by asking `status` AGAIN and refuting
# `complete: true` in the answer — the question a landed write gets wrong
# whatever the verb that landed it printed on its way through.
#
# Measured on this host on 2026-09-07 against a copy whose `amend_in_flight`
# always answers false, which is the state before the guard in `cmd_manifest`
# and is what `claim-table-mutations.sh M54` restores, so these figures are
# reproducible rather than remembered.
#
# The amendment goes first because it is the ROUTE and not merely a second door.
# `--amend` in the window also finds nothing at the set of record, so before the
# window check was ordered AHEAD of the nothing-to-amend check this verb
# answered `nothing to amend` at 12 and directed the caller to record it without
# `--amend`. Refuting that direction is what pins the order of the two checks; an
# assertion on the refusal alone cannot see it, and a caller obeying the printed
# direction is exactly how the write below was reached.
run manifest $PR $SHA --lanes 'alpha beta' --amend --reason 'gamma stood down'
check 'in the window an amendment is refused, at 2'           2  'manifest: NOT RECORDED'
refute '  and never as a panel with nothing to amend'            'nothing to amend'
run status $PR $SHA
refute 'the refused amendment leaves NO complete: true behind'   'complete: true'

# And the plain selection, which is where that direction sent the caller. It
# answered `manifest: recorded` at 0 — `commit_staged` finds the name free, so
# nothing downstream could refuse it — leaving the panel holding a two-lane set
# with both lanes ruled, its three-lane set of record orphaned under the audit
# name and cited in no `supersedes:`, and `status` turned from `complete: false`
# at 10 into `complete: true` at 0. Failure class 4, reached through a writer.
run manifest $PR $SHA --lanes 'alpha beta'
check 'in the window a plain selection is refused too, at 2'  2  'manifest: NOT RECORDED'
refute '  and never reports a set recorded'                      'manifest: recorded'
check '  saying a set written here would orphan the one on disk' 2 'would orphan the one already on disk'
run status $PR $SHA
refute 'the refused selection leaves no complete: true either'   'complete: true'
check '  the panel still measured incomplete, at 2'           2  'complete: false'

# And the state is transient rather than sticky: the second rename lands and the
# panel is an ordinary one again. A degraded state a panel cannot leave would be
# a wedge of its own, and this one is defined by disk state on both sides.
mv "$TMP/held$CASE" "$PANEL/.manifest.d"
run status $PR $SHA
check 'once the second rename lands the panel reads normally'  10 'complete: false'
refute '  with the between-state gone'                            'amend-in-flight'
check '  measured against the amended set again'               10 'lanes: 3'

# ---------------------------------------------------------------------------
section 'a verdict names who FILED it, not only who held the lane'
# ---------------------------------------------------------------------------
# `owner:` is copied out of the claim, so a verdict written by anyone at all was
# attributed of record to the claimant, and the writer's identity was not merely
# unrecorded but unrecordable — there was no flag to carry it. A courier filing
# into a lane held by someone else produced a row reading
# `verdict / owner: <the reviewer that never ran>`, exit 0, panel complete,
# synthesise. `unclaimed` never fired, because it fires only where NO holder
# exists: the likely case was the silent one. Four values, each asserted in the
# row on disk, because a flag no reader can tell apart is not a flag.
fresh
run claim $PR $SHA agree --owner alice --for alice
run verdict $PR $SHA agree approved --by alice
check 'a verdict filed by the holder is recorded'             0  'verdict: recorded'
check '  naming its filer'                                    0  'filed_by: alice'
is '  and the row records the check as having passed' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA agree)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: false'

run claim $PR $SHA differ --owner judge --for alice
run verdict $PR $SHA differ approved --by mallory
check 'a verdict filed by someone OTHER than the holder is 12' 12 'anomaly: filed by mallory'
check '  naming both identities'                              12 'expects its verdict from alice (held by judge)'
check '  and it is RECORDED, never discarded to enforce that' 12 'verdict: recorded'
is '  flagged in the row itself' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA differ)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: true'
is '  which keeps the filer distinct from the holder on disk' \
   "$(grep '^filed_by: ' "$(row_of $PR $SHA differ)/verdict.d/row" 2>/dev/null | head -1)" 'filed_by: mallory'
run status $PR $SHA differ
check '  and a read verb surfaces the mismatch'               0  'attribution_mismatch: true'

run claim $PR $SHA silent --owner alice
run verdict $PR $SHA silent approved
check 'a verdict with no --by is still recorded'              0  'verdict: recorded'
check '  and says its own writer is unknown'                  0  'filed_by: unrecorded'
is '  which is a value and not a pass' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA silent)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: unrecorded'

run verdict $PR $SHA no-holder approved --by carol
check 'a verdict on a lane no claim covers is 12'             12 'anomaly: no claim row covered'
is '  distinguishing no holder from the wrong holder' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA no-holder)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: no-holder'
run status $PR $SHA no-holder
check '  and a read verb surfaces THAT anomaly too'           0  'unclaimed: true'

# ---------------------------------------------------------------------------
section 'the holder and the expected filer are two identities, not one'
# ---------------------------------------------------------------------------
# SEC-01's first remedy checked `--by` against `owner:` and so fired on the
# ORDINARY path. `pr-judge` Phase 4 claims each selected lane ON THE REVIEWER'S
# BEHALF before dispatching it, and `engineering-lead` couriers the same way, so
# the judge is always the holder and the reviewer is always the filer: every
# honest verdict this table will ever see recorded `attribution_mismatch: true`
# at exit 12. The normal case reported as the anomaly, which is how a signal
# becomes noise and then gets switched off by the third reader.
#
# NOTHING ABOVE COULD SEE IT. The section before this one pins a holder filing
# into its own lane — a shape the courier design never produces — so the suite
# stayed green while the flag fired on every real dispatch, and the defect was
# caught by hand instead. That absent assertion is why this section exists, and
# the clean ordinary path is therefore pinned FIRST and as a DIRECTION:
# `anomaly` must be ABSENT from it, not merely outnumbered inside it.
fresh
run claim $PR $SHA docs-reviewer --owner pr-judge
check 'a judge claims a lane on its reviewer behalf'          0  'claim: granted'
check '  the row naming who HOLDS it'                         0  'owner: pr-judge'
check '  and, separately, who it EXPECTS to file'             0  'expects: docs-reviewer'
is '  the second identity defaulting to the lane, unpassed' \
   "$(grep '^expects: ' "$(row_of $PR $SHA docs-reviewer)/holder/row" 2>/dev/null | head -1)" 'expects: docs-reviewer'

run verdict $PR $SHA docs-reviewer approved --by docs-reviewer
check '(a) the lane own reviewer files: RECORDED at 0'        0  'verdict: recorded'
refute '  and the ORDINARY path is never called an anomaly'      'anomaly'
refute '  nor flagged a mismatch in the caller own output'       'mismatch'
is '  the row recording the check as PASSED, not as absent' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA docs-reviewer)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: false'
is '  while the holder stays distinct from the filer on disk' \
   "$(grep '^owner: ' "$(row_of $PR $SHA docs-reviewer)/verdict.d/row" 2>/dev/null | head -1)" 'owner: pr-judge'

# (b) The attack the flag exists for, which must still fire once the ordinary
# path is quiet — a flag made quiet by being switched off detects nothing either.
run claim $PR $SHA sec-reviewer --owner pr-judge
run verdict $PR $SHA sec-reviewer approved --by mallory
check '(b) a lane filed by ANOTHER hand is 12'                12 'anomaly: filed by mallory'
check '  naming the identity expected, and the holder too'    12 'expects its verdict from sec-reviewer (held by pr-judge)'
check '  and RECORDED, never discarded to enforce the rule'   12 'verdict: recorded'
is '  flagged true in the row itself' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA sec-reviewer)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: true'

# (c) and (d), the two values that are neither pass nor attack, kept apart
# because a flag whose four values collapse to two is evidence of nothing.
run claim $PR $SHA quiet-l --owner pr-judge
run verdict $PR $SHA quiet-l approved
check '(c) no --by at all is recorded at 0'                   0  'verdict: recorded'
refute '  and an unchecked filer is not reported as an attack'   'anomaly'
is '  the row saying its own writer is unknown' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA quiet-l)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: unrecorded'

run verdict $PR $SHA never-claimed-l approved --by docs-reviewer
check '(d) no claim row at all is 12'                         12 'anomaly: no claim row covered'
is '  distinguished from a wrong filer by its own token' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA never-claimed-l)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: no-holder'

# ONE AGENT IS ONE NAME, HOWEVER IT IS SPELLED.
# `norm_lane` folds a lane; `--by` is free text and is not folded, so without
# `agent_key` the same false anomaly walks back in through the SPELLING rather
# than through the field. Measured on this tree before `agent_key` existed:
# `--by @docs-reviewer` against a lane-derived `expects: docs-reviewer` was rc
# 12. A live direction, not a hypothetical, and all four spellings a dispatch
# actually uses are pinned rather than one standing for the rest.
for spelling in 'docs-reviewer' '@docs-reviewer' 'Docs-Reviewer' '@Docs-Reviewer'; do
  fresh
  run claim $PR $SHA "$spelling" --owner pr-judge
  run verdict $PR $SHA docs-reviewer approved --by "$spelling"
  check "a filer spelled [$spelling] is the identity expected"  0  'verdict: recorded'
  refute "  and is never an anomaly on the spelling alone"         'anomaly'
done
# The fold is for the COMPARISON only. A row that quietly rewrites the name it
# was handed is not evidence of what it was handed.
is '  yet the row stores the spelling its caller passed' \
   "$(grep '^filed_by: ' "$(row_of $PR $SHA docs-reviewer)/verdict.d/row" 2>/dev/null | head -1)" 'filed_by: @Docs-Reviewer'
# and the fold does not make every name agree — a different one still fires.
fresh
run claim $PR $SHA docs-reviewer --owner pr-judge
run verdict $PR $SHA docs-reviewer approved --by '@Mallory'
check '  while a genuinely different name still fires at 12'  12 'anomaly: filed by @Mallory'

# THE STRIP IS ANCHORED: A LEADING @ IS A SPELLING, AN EMBEDDED ONE IS A NAME.
# The fold was `tr -d '@'`, which deleted EVERY `@` in the name, so `a@b` and
# `ab` produced one key and two different identities compared EQUAL. That is the
# mismatch flag answering `false` about a filer it never saw — the ACCEPT
# direction, and the one this field cannot afford, since every other value it
# carries is only worth reading if `false` means checked.
#
# The four spellings above cannot see this and neither can the mutation aimed at
# them: they all carry the `@` in the one position both the anchored rule and
# the unanchored one strip, so the two rules agree on every case in this file
# and the anchoring is invisible. An embedded `@` is the only input that
# separates the two, which is why this is asserted rather than folded in above.
fresh
run claim $PR $SHA embedded-l --owner pr-judge --for 'a@b'
check 'a lane may expect a filer whose name CONTAINS an @'    0  'expects: a@b'
run verdict $PR $SHA embedded-l approved --by 'ab'
check '  a filer differing only by that @ is an anomaly at 12' 12 'anomaly: filed by ab'
check '  naming the identity it is not'                       12 'expects its verdict from a@b'
is '  and flagged in the row, not only argued in the prose' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA embedded-l)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: true'

# `--for`, the override, for a lane dispatched to a name other than its own.
fresh
run claim $PR $SHA lane-x --owner pr-judge --for docs-reviewer
check 'a lane may expect a filer other than its own name'     0  'expects: docs-reviewer'
run verdict $PR $SHA lane-x approved --by docs-reviewer
check '  and that filer is the one it accepts'                0  'verdict: recorded'
refute '  without an anomaly'                                    'anomaly'
run claim $PR $SHA lane-y --owner pr-judge --for docs-reviewer
run verdict $PR $SHA lane-y approved --by lane-y
check '  while the LANE NAME is no longer the one expected'   12 'anomaly: filed by lane-y'

# A claim row written before `expects:` existed carries none. It must fall back
# to the same default `claim` records, or an upgrade flags every honest verdict
# in an older table — the regression above, arriving a second time by data age.
fresh
run claim $PR $SHA legacy-l --owner pr-judge
sed -i '/^expects: /d' "$(row_of $PR $SHA legacy-l)/holder/row" 2>/dev/null
is '  a pre-expects claim row really carries no expects' \
   "$(grep -c '^expects: ' "$(row_of $PR $SHA legacy-l)/holder/row" 2>/dev/null | tr -d ' ')" 0
run verdict $PR $SHA legacy-l approved --by legacy-l
check 'a legacy row falls back to the lane name, at 0'        0  'verdict: recorded'
refute '  and is not flagged merely for predating the field'     'anomaly'
is '  recording the check as passed against the fallback' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA legacy-l)/verdict.d/row" 2>/dev/null | head -1)" 'attribution_mismatch: false'
fresh
run claim $PR $SHA legacy-m --owner pr-judge
sed -i '/^expects: /d' "$(row_of $PR $SHA legacy-m)/holder/row" 2>/dev/null
run verdict $PR $SHA legacy-m approved --by mallory
check '  yet a legacy row still catches a wrong filer'        12 'anomaly: filed by mallory'

# AN INFERENCE THIS TABLE MADE MUST NOT READ AS SOMETHING A CLAIM RECORDED.
# The fallback above is right and stays, but written into the row unmarked it is
# indistinguishable from an expectation a claim actually asked for:
# `expects: docs-reviewer` beside `attribution_mismatch: false` says "checked,
# and the filer is the one the claim expected" about a claim that expected
# nothing — and, on `cmd_claim`'s degraded branch, about a claim row that was
# never written at all. Two states collapsing into the one that reads clean is
# what an absent `--by` reading `unrecorded` already refuses one field to the
# left. Neither value refuses anything; the difference is on disk, where a
# reader can act on it, and BOTH read verbs surface it.
fresh
run claim   $PR $SHA docs-reviewer --owner pr-judge
run verdict $PR $SHA docs-reviewer approved --by docs-reviewer
is 'a RECORDED expectation says where it came from, in the row' \
   "$(grep '^expects_source: ' "$(row_of $PR $SHA docs-reviewer)/verdict.d/row" 2>/dev/null | head -1)" \
   'expects_source: recorded'
run status $PR $SHA docs-reviewer
check '  and the single-lane view surfaces it'                0  'expects_source: recorded'
run status $PR $SHA
refute '  while the panel view stays quiet about the clean case' 'expects_source:'

fresh
run claim $PR $SHA docs-reviewer --owner pr-judge
sed -i '/^expects: /d' "$(row_of $PR $SHA docs-reviewer)/holder/row" 2>/dev/null
run verdict $PR $SHA docs-reviewer approved --by docs-reviewer
check 'an INFERRED expectation still records at 0'            0  'verdict: recorded'
is 'an INFERRED expectation says so in the row it wrote' \
   "$(grep '^expects_source: ' "$(row_of $PR $SHA docs-reviewer)/verdict.d/row" 2>/dev/null | head -1)" \
   'expects_source: inferred'
is '  beside the flag it qualifies, which still reads false' \
   "$(grep '^attribution_mismatch: ' "$(row_of $PR $SHA docs-reviewer)/verdict.d/row" 2>/dev/null | head -1)" \
   'attribution_mismatch: false'
run status $PR $SHA docs-reviewer
check '  the single-lane view naming it inferred'             0  'expects_source: inferred'
run status $PR $SHA
check '  and the WHOLE-PANEL view surfacing it too'           0  'expects_source: inferred'
check '  in words, not merely a token a reader must decode'   0  'the lane name stood in'

# ---------------------------------------------------------------------------
section 'the deleted --note is a usage error, not a silent no-op'
# ---------------------------------------------------------------------------
# LA-10 asked that each free-text flag name an obliged caller or go. `--note`
# could not, so it was deleted. A deleted flag quietly ACCEPTED and dropped is
# worse than one that never existed: the caller believes the table holds text it
# does not hold. It must refuse, and refuse at 64 — nothing read, nothing
# written — which is the same direction every other malformed call takes here.
fresh
run claim $PR $SHA gone-l --owner pr-judge
run verdict $PR $SHA gone-l approved --by gone-l --note 'seven findings'
check 'verdict --note is a usage error at 64'                 64 'usage: claim.sh verdict'
refute '  and never reads as recorded'                           'verdict: recorded'
refute '  the usage line no longer offering the flag'            '--note'
is '  and nothing was written for it' \
   "$([ -e "$(row_of $PR $SHA gone-l)/verdict.d" ] && printf yes || printf no)" no
run help
refute 'nor does help name the flag any more'                    '--note'

# ---------------------------------------------------------------------------
section 'status answers in a field a caller can branch on'
# ---------------------------------------------------------------------------
# `status <lane>` returned 10 for held, for free and for ruling-incomplete alike
# — three states with three different remedies — while every caller is told to
# branch on the exit status and never on the prose, which left them nothing to
# branch on. `lane_state:` is that field, and all five of its tokens are pinned
# here; the sixth this suite can produce, `no-row`, is pinned beside them.
fresh
run manifest $PR $SHA --lanes 'ruled-l claimed-l inflight-l free-l never-l'
run claim $PR $SHA ruled-l --owner j1 --for j1
run verdict $PR $SHA ruled-l approved --by j1 --artifact 'https://github.com/x/pull/154#note-1'
run claim $PR $SHA claimed-l --owner j2
run claim $PR $SHA inflight-l --owner j3
mkdir -p "$(row_of $PR $SHA inflight-l)/verdict.d"
run claim $PR $SHA free-l --owner j4
run release $PR $SHA free-l --reason 'its judge died'

run status $PR $SHA ruled-l
check 'lane_state ruled'                                      0  'lane_state: ruled'
check '  and the ruled lane offers its route to the findings' 0  'artifact: https://github.com/x/pull/154'
check '  and who filed it'                                    0  'filed_by: j1'
run status $PR $SHA claimed-l
check 'lane_state claimed'                                    10 'lane_state: claimed'
check '  naming the holder to ask'                            10 'holder: j2'
run status $PR $SHA inflight-l
check 'lane_state ruling-incomplete'                          10 'lane_state: ruling-incomplete'
run status $PR $SHA free-l
check 'lane_state free'                                       10 'lane_state: free'
run status $PR $SHA never-l
check 'lane_state never-claimed'                              10 'lane_state: never-claimed'
run status $PR $SHA ghost-l
check 'a lane the panel never heard of is no-row at 12'       12 'lane_state: no-row'
run status $PR $SHA
check 'the whole-panel view surfaces the artifact too'        10 'artifact: https://github.com/x/pull/154'

# ---------------------------------------------------------------------------
section 'control characters cannot repaint the reader of a row'
# ---------------------------------------------------------------------------
# The strip was `tr -d` of CR and LF alone — they are what break the one-line
# record format, so they were all that was removed, and ANSI escapes were
# therefore stored raw. A stored escape survives the round trip and repaints the
# terminal of the human reading the panel: the row is intact and the reader's
# view of it is not, which is the one attack a ledger whose whole purpose is
# being believed cannot afford. The strip is now the ASCII control range under
# LC_ALL=C — and ordinary text must survive it, or the scrub becomes a second
# way to lose a verdict, so the UTF-8 half is asserted beside it and built from
# octal escapes rather than typed, so no checkout encoding can decide it.
fresh
ESC="$(printf '\033')"
U8="$(printf '\303\245\303\244\303\266')"
run claim $PR $SHA ansi --owner "$(printf 'a\033[31mj1\033[0m')"
run verdict $PR $SHA ansi approved --artifact "$(printf 'x\033[2Jy\001z\tw ')$U8"
check 'an artifact carrying ANSI escapes is accepted'         0  'verdict: recorded'
is '  and no escape byte reaches the verdict row on disk' \
   "$(grep -c "$ESC" "$(row_of $PR $SHA ansi)/verdict.d/row" 2>/dev/null | tr -d ' ')" 0
is '  nor any other control byte' \
   "$(grep -c '[[:cntrl:]]' "$(row_of $PR $SHA ansi)/verdict.d/row" 2>/dev/null | tr -d ' ')" 0
is '  nor the holder row either' \
   "$(grep -c "$ESC" "$(row_of $PR $SHA ansi)/holder/row" 2>/dev/null | tr -d ' ')" 0
is '  while ordinary UTF-8 text survives the strip' \
   "$(grep -c "$U8" "$(row_of $PR $SHA ansi)/verdict.d/row" 2>/dev/null | tr -d ' ')" 1

# ---------------------------------------------------------------------------
section 'the table enforces no verdict vocabulary'
# ---------------------------------------------------------------------------
# `@validation-agent`'s schema returns `status: pass | fail`, and neither word
# is in `review-verdicts`' vocabulary. A check here could only ever refuse a
# verdict it did not recognise, losing the work this script exists to keep — so
# both record, and both read back as themselves rather than as a judgement.
fresh
run claim $PR $SHA vpass --owner v1 --for v1
run verdict $PR $SHA vpass pass --by v1
check 'a verdict of `pass` is recorded'                       0  'ruling: pass'
is '  and reads back as itself' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA vpass)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: pass'
run claim $PR $SHA vfail --owner v1 --for v1
run verdict $PR $SHA vfail fail --by v1
check 'a verdict of `fail` is recorded, not refused'          0  'ruling: fail'
is '  and it too reads back as itself' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA vfail)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: fail'
run status $PR $SHA
check '  and a panel of pass and fail is complete, not judged' 0 'complete: true'

# ---------------------------------------------------------------------------
section 'CRLF — the checkout form must not change any answer'
# ---------------------------------------------------------------------------
# scripts/loop/*.sh carry no eol attribute, so a core.autocrlf=true host checks
# them out CRLF; fingerprint-isolation.sh already is CRLF in this working tree.
CRLFSH="$TMP/crlf-claim.sh"
sed 's/$/\r/' "$SCRIPT" > "$CRLFSH"
fresh; LFT="$CLAIM_TABLE_DIR"
run claim   $PR $SHA p --owner j1;    lf1=$RC
run claim   $PR $SHA p --owner j2;    lf2=$RC
run verdict $PR $SHA p approved;      lf3=$RC
run release $PR $SHA p --reason x;    lf4=$RC
run claim   $PR $SHA '../x';          lf5=$RC
run status  $PR $SHA;                 lf6=$RC
fresh
OUT="$(bash "$CRLFSH" claim   $PR $SHA p --owner j1 2>&1)";  cr1=$?
OUT="$(bash "$CRLFSH" claim   $PR $SHA p --owner j2 2>&1)";  cr2=$?
OUT="$(bash "$CRLFSH" verdict $PR $SHA p approved 2>&1)";    cr3=$?
OUT="$(bash "$CRLFSH" release $PR $SHA p --reason x 2>&1)";  cr4=$?
OUT="$(bash "$CRLFSH" claim   $PR $SHA '../x' 2>&1)";        cr5=$?
OUT="$(bash "$CRLFSH" status  $PR $SHA 2>&1)";               cr6=$?
is 'a CRLF checkout returns the same code for six verbs' \
   "$cr1/$cr2/$cr3/$cr4/$cr5/$cr6" "$lf1/$lf2/$lf3/$lf4/$lf5/$lf6"
is '  and writes the same lane directory name' \
   "$(ls "$CLAIM_TABLE_DIR/pr-$PR/$SHA" 2>/dev/null)" "$(ls "$LFT/pr-$PR/$SHA" 2>/dev/null)"
is '  and no CR reached any row this suite wrote' \
   "$(grep -rlU "$CR" "$CLAIM_TABLE_DIR" 2>/dev/null | wc -l | tr -d ' ')" 0

# ---------------------------------------------------------------------------
section 'hermetic — the real table under .claude/state was never touched'
# ---------------------------------------------------------------------------
REAL_AFTER=absent; [ -e "$REALTABLE" ] && REAL_AFTER=present
REAL_N_AFTER="$(ls -a "$REALTABLE" 2>/dev/null | wc -l | tr -d ' ')"
is 'the real claim table is in the state the suite found it'  "$REAL_AFTER"   "$REAL_BEFORE"
is '  with the same number of entries'                        "$REAL_N_AFTER" "$REAL_N_BEFORE"
is '  and no row for this suite PR number' \
   "$([ -e "$REALTABLE/pr-$PR" ] && printf leaked || printf clean)" clean
# The guard is only worth anything if it watched the table the script would have
# written. `list` is the one read-only verb — it never creates the table — so
# asking it, with no CLAIM_TABLE_DIR, is safe and is the only way to check that
# the derivation above matches claim.sh's own rather than a path nothing writes.
is '  and the table it guarded is the one claim.sh resolves' \
   "$( (unset CLAIM_TABLE_DIR; bash "$SCRIPT" list 2>&1 | head -1) )" "table: $REALTABLE"

printf '\n%s checks · %s failed\n' "$cases" "$fails"
[ "$fails" -eq 0 ] || exit 1
printf 'all passed\n'
exit 0
