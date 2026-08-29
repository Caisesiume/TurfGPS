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

# The table this suite must never touch. Computed exactly as claim.sh computes it,
# so the guard at the end compares like with like.
GITROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
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

# A ruling in flight — verdict.d present, its row not yet written. A reviewer died
# mid-write. Neither claimable nor releasable.
fresh
run claim $PR $SHA inflight --owner r1
mkdir -p "$(row_of $PR $SHA inflight)/verdict.d"
run release $PR $SHA inflight --reason 'looks stuck to me'
check 'release of a lane whose ruling is IN FLIGHT is refused' 10 'reason: ruled'
run claim $PR $SHA inflight --owner r2
check '  and it is not claimable either'                      10 'reason: ruled'
check '  saying the record is unreadable, not absent'         10 'unreadable'
run status $PR $SHA
check '  and status counts it outstanding'                    10 'ruling-incomplete'
refute '  and never reports the panel complete'                  'complete: true'

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
run verdict $PR $SHA orphan rejected --note 'no claim covered this'
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

run verdict $PR $SHA secret approved --note "leaked $TOK here"
is 'a token-shaped note is redacted in the row' \
   "$(grep -c "$TOK" "$(row_of $PR $SHA secret)/verdict.d/row" 2>/dev/null | tr -d ' ')" 0

fresh
run claim $PR $SHA forge --owner j1
run verdict $PR $SHA forge pending --note "$(printf 'x\nverdict: approved-by-injection')"
check 'a note carrying a newline is accepted'                 0  'verdict: recorded'
is '  but cannot forge a second verdict line' \
   "$(grep -c '^verdict: ' "$(row_of $PR $SHA forge)/verdict.d/row" 2>/dev/null | tr -d ' ')" 1
is '  and the ruling of record is the real one' \
   "$(grep '^verdict: ' "$(row_of $PR $SHA forge)/verdict.d/row" 2>/dev/null | head -1)" 'verdict: pending'

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

printf '\n%s checks · %s failed\n' "$cases" "$fails"
[ "$fails" -eq 0 ] || exit 1
printf 'all passed\n'
exit 0
