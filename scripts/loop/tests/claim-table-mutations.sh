#!/usr/bin/env bash
# claim-table-mutations.sh — the red demonstration for claim-table.sh, mechanised.
#
# `docs/DELIVERY.md § Proof that a test can fail` requires every test bound to a
# criterion to be shown failing without the change under test. A claim that it was
# shown red is not checkable; this is, by anyone, at any later date.
#
# For each mutation it: copies scripts/loop/claim.sh to a scratch file, NEUTRALISES
# one behaviour in the copy, runs the suite against it via CLAIM_SH, and requires
# the named assertion to be among the failures. The tracked script is never touched
# and every copy is removed on exit.
#
# `docs/DELIVERY.md § Red for the wrong reason` excludes a test that is red
# because nothing ran, so each mutant is required to PARSE and to answer `help`
# before the suite is run against it. A mutation that does not apply, or that
# breaks the script outright, is reported as a defect in this harness rather than
# counted as a demonstration.
#
# A SURVIVING MUTATION IS A FAILURE HERE. It means the suite cannot see the
# behaviour that mutation removed — an assertion that has never been red and
# therefore proves nothing, which is the exact instrument class #144 was filed over.
#
# It also reports, at the end, every assertion in the suite that NO mutation made
# fail. Those are undemonstrated: they may still be sound, but this harness has not
# shown it, and saying so is worth more than a count that implies otherwise. They
# are reported rather than failed, because some of them — the hermeticity guards —
# assert a property of the suite rather than a behaviour of claim.sh, and no
# mutation of claim.sh should be able to move them.
#
# A MUTATION THAT MATCHES NOTHING IS A DEFECT HERE, AND FIVE OF THEM HAVE BEEN.
# M02, M13 and M20 addressed lines the review board's fixes removed — the mkdir
# verdict gate, the `\r\n` strip, the `mv -f` of a bare row file — and M38 the
# one-line `else mismatch=true; fi` that SEC-01's rebuild in this same branch
# replaced with a multi-line `if`. M45 is the fifth: SEC-05 rewrote `agent_key`
# into two statements and DELETED the `| tr -d '@' | tr 'A-Z' 'a-z'` pipeline the
# edit aimed at, so the mutation stopped applying and the identity fold stopped
# being demonstrated at the same head that hardened it. Each still
# named a real behaviour, so each was re-aimed at the line that now carries it
# rather than dropped; a mutation retired quietly is an assertion that stops
# being demonstrated with nothing saying so. The `cmp -s` guard below is what
# caught all five, which is the argument for keeping it.
#
# Usage: scripts/loop/tests/claim-table-mutations.sh [id …]
#        Exit: 0 every mutation killed · 1 any mutation survived or misapplied
# The whole suite runs once per mutation, plus once for the baseline. Re-measured
# on 2026-09-06 on the reference host, at 366 assertions: one suite run is 131 s,
# and the full matrix of 51 is 52 of them — about two hours, so it is a background
# gate rather than an inner-loop one. Named ids run a subset —
# `claim-table-mutations.sh M28 M33` — and that is how a single behaviour is
# re-demonstrated after a change without paying for all 51. A subset run does not
# print the undemonstrated list; the block at the foot of this file says why.

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
ORIG="$DIR/../claim.sh"
SUITE="$DIR/claim-table.sh"
[ -r "$ORIG" ]  || { printf 'cannot read %s\n' "$ORIG" >&2; exit 1; }
[ -r "$SUITE" ] || { printf 'cannot read %s\n' "$SUITE" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
Q="'"

ALL='M01 M02 M03 M04 M05 M06 M07 M08 M09 M10 M11 M12 M13 M14 M15 M16 M17 M18
M19 M20 M21 M22 M23 M24 M25 M26 M27 M28 M29 M30 M31 M32 M33 M34 M35 M36 M37
M38 M39 M40 M41 M42 M43 M44 M45 M46 M47 M48 M49 M50 M51'
WANT="${*:-$ALL}"

# what each mutation neutralises
desc() {
  case "$1" in
    M01) printf 'claim gate loses atomicity: mkdir -> mkdir -p on the holder' ;;
    M02) printf 'the ruling commit clobbers a ruling of record instead of refusing' ;;
    M03) printf 'claim direction inverted: a degraded table reads as granted' ;;
    M04) printf 'verdict direction inverted: a lost verdict reads as recorded' ;;
    M05) printf 'release no longer requires a --reason' ;;
    M06) printf 'release admits a ruled lane' ;;
    M07) printf 'the pause flag is never seen' ;;
    M08) printf 'the pause blocks a verdict from an already-claimed lane' ;;
    M09) printf 'the lane is no longer case-folded' ;;
    M10) printf 'the sha is no longer case-folded' ;;
    M11) printf 'the @ prefix is no longer stripped from a lane' ;;
    M12) printf 'the token scrub no longer matches a token' ;;
    M13) printf 'free text is no longer stripped of newlines' ;;
    M14) printf 'the release stamp carries a colon in its directory name' ;;
    M15) printf 'an unclaimed verdict is recorded silently, exit 0' ;;
    M16) printf 'a ruling interrupted mid-write counts as ruled' ;;
    M17) printf 'a usage error writes state before refusing' ;;
    M18) printf 'claim admits a lane whose ruling is in flight' ;;
    M19) printf 'a usage error exits 0 instead of 64' ;;
    M20) printf 'the verdict row never lands on disk, but reads recorded' ;;
    M21) printf 'the release audit row keeps its stale claimed state' ;;
    M22) printf 'claim grants when it could not create the row' ;;
    M23) printf 'both verbs fail in the SAME direction' ;;
    M24) printf 'status calls a panel complete while lanes are outstanding' ;;
    M25) printf 'the sha length bounds are removed' ;;
    M26) printf 'a lane may traverse or hide again' ;;
    M27) printf 'status returns 0 for a panel with no rows' ;;
    M28) printf 'the panel key is the caller SPELLING again, not the seven-hex epoch' ;;
    M29) printf 'a leading-zero pr number is a panel of its own again' ;;
    M30) printf 'the table follows the caller worktree again, not the repository' ;;
    M31) printf 'no verb names the table it acted on' ;;
    M32) printf 'the ruling commit drops -T and moves INSIDE an existing directory' ;;
    M33) printf 'the ruling directory is created before its payload — the two-step gate' ;;
    M34) printf 'status stops counting a selected lane that has no row' ;;
    M35) printf 'the manifest can be rewritten after the fact' ;;
    M36) printf 'status prints no lane_state field' ;;
    M37) printf 'no read verb surfaces the artifact' ;;
    M38) printf 'a verdict filed by another hand never reads as a mismatch' ;;
    M39) printf 'the control strip is CR and LF again, so an ANSI escape is stored' ;;
    M40) printf 'release calls an interrupted ruling finished again' ;;
    M41) printf 'no read verb surfaces the unclaimed flag' ;;
    M42) printf 'the table follows the caller cwd, so a subdirectory is its own panel' ;;
    M43) printf 'the table enforces a verdict vocabulary and refuses one it does not know' ;;
    M44) printf 'the filer is checked against the HOLDER again, not against expects' ;;
    M45) printf 'agent_key stops folding, so one agent is two identities by spelling' ;;
    M46) printf 'a claim row predating expects no longer falls back to the lane name' ;;
    M47) printf 'the deleted --note is quietly accepted and dropped again' ;;
    M48) printf 'a ruling the table cannot commit is deleted and called already-ruled' ;;
    M49) printf 'a verdict builds a panel nobody ever opened, under a pause' ;;
    M50) printf 'an amendment DELETES the selection it replaced instead of keeping it' ;;
    M51) printf 'an expectation this table inferred reads as one a claim recorded' ;;
  esac
}

# the assertion that must go red. One per mutation, chosen as the most
# characteristic; the harness also reports how many others died with it.
kills() {
  case "$1" in
    M01) printf 'a second claim on the held lane is REFUSED' ;;
    M02) printf 'a SECOND ruling is refused' ;;
    M03) printf 'and NEVER reads as granted' ;;
    M04) printf 'unwritable table: verdict says NOT RECORDED' ;;
    M05) printf 'release with NO --reason is a usage error' ;;
    M06) printf 'release of a RULED lane is refused' ;;
    M07) printf 'a NEW claim under pause is refused with 11' ;;
    M08) printf 'an ALREADY-CLAIMED lane still records under pause' ;;
    M09) printf 'and written lower-cased on disk' ;;
    M10) printf 'and written lower-cased on disk' ;;
    M11) printf 'and written without the @ on disk' ;;
    M12) printf 'and redacted in the row on disk' ;;
    M13) printf 'but cannot forge a second verdict line' ;;
    M14) printf 'whose directory name carries no colon' ;;
    M15) printf 'and is loud about the anomaly' ;;
    M16) printf 'and status counts it outstanding' ;;
    M17) printf 'and creates no table' ;;
    M18) printf 'a lane whose ruling is IN FLIGHT is not claimable' ;;
    M19) printf '64 · pr is not digits' ;;
    M20) printf 'the verdict survives the writer being killed -9' ;;
    M21) printf 'and reads released, not claimed' ;;
    M22) printf 'unmakeable row: claim refuses' ;;
    M23) printf 'on ONE degraded table the two verbs fail in opposite directions' ;;
    M24) printf 'status: complete is 0' ;;
    M25) printf '64 · sha too short' ;;
    M26) printf '64 · lane starting with a dot' ;;
    M27) printf 'status: no rows is 12' ;;
    M28) printf 'the SAME commit spelled short is the SAME row' ;;
    M29) printf 'a leading-zero PR number is the same panel' ;;
    M30) printf 'the LINKED WORKTREE resolves the same one' ;;
    M31) printf 'names its table on line 1' ;;
    M32) printf 'and a READABLE row stands where the wreckage was' ;;
    M33) printf 'twelve killed writers leave NO empty ruling directory' ;;
    M34) printf 'two ruled of seven selected is NOT a complete panel' ;;
    M35) printf 'a SECOND selection is refused' ;;
    M36) printf 'lane_state claimed' ;;
    M37) printf 'the whole-panel view surfaces the artifact too' ;;
    M38) printf 'a verdict filed by someone OTHER than the holder is 12' ;;
    M39) printf 'and no escape byte reaches the verdict row on disk' ;;
    M40) printf 'release RECOVERS it rather than calling it finished' ;;
    M41) printf 'and a read verb surfaces THAT anomaly too' ;;
    M42) printf 'a main SUBDIRECTORY resolves the same one' ;;
    M43) printf 'a verdict of `pass` is recorded' ;;
    M44) printf 'and the ORDINARY path is never called an anomaly' ;;
    M45) printf 'a filer spelled [@docs-reviewer] is the identity expected' ;;
    M46) printf 'a legacy row falls back to the lane name, at 0' ;;
    M47) printf 'verdict --note is a usage error at 64' ;;
    M48) printf 'a ruling blocked by a FILE at verdict.d is NOT RECORDED' ;;
    M49) printf 'a verdict into a lane with NO row is refused under pause' ;;
    M50) printf 'the prior selection kept WHOLE, not deleted' ;;
    M51) printf 'an INFERRED expectation says so in the row it wrote' ;;
  esac
}

# Each edit keeps the shape of the line it changes. Nothing is deleted: a deleted
# branch takes its own assertion out of the run, which
# `docs/DELIVERY.md § Red for the wrong reason` excludes as a demonstration.
mutate() { # mutate <id> <file>
  m="$1"; f="$2"
  case "$m" in
    M01) sed -i 's@if mkdir "$row/holder" 2>/dev/null; then@if mkdir -p "$row/holder" 2>/dev/null; then@' "$f" ;;
    # M02 used to turn the verdict's `mkdir` gate into `mkdir -p`. That gate is
    # gone: the ruling is committed by renaming a staged directory, so the
    # mutual exclusion now lives in `commit_staged` and this edit follows it
    # there. Clearing the destination first is the same neutralisation — the
    # kernel stops deciding, and the second ruling of a lane wins.
    M02) sed -i 's@^  mv -T "$1" "$2" 2>/dev/null$@  rm -rf "$2" 2>/dev/null; mv -T "$1" "$2" 2>/dev/null@' "$f" ;;
    M03) sed -i -e 's@claim: refused\(.*table not writable\)@claim: granted\1@' \
                -e '/refusing to dispatch/,+1 s@exit 2@exit 0@' "$f" ;;
    M04) sed -i -e 's@verdict: NOT RECORDED@verdict: recorded@g' \
                -e '/carry this verdict in your handoff/,+1 s@exit 2@exit 0@' "$f" ;;
    M05) sed -i 's@\[ -n "$reason" \] || usage_die.*@[ -n "$reason" ] || reason=unstated@' "$f" ;;
    M06) sed -i '/^cmd_release() {$/,/^}$/ s@\[ -d "$row/verdict.d" \]@[ -d "$row/verdict.d/never" ]@' "$f" ;;
    M07) sed -i 's@is_paused() { \[ -e "$TABLE/PAUSED" \]; }@is_paused() { [ -e "$TABLE/PAUSED.never" ]; }@' "$f" ;;
    M08) sed -i '/^cmd_verdict() {$/,/^}$/ s@^  ensure_writable || {$@  if is_paused; then printf "verdict: refused\\nreason: paused\\n"; exit 11; fi\n  ensure_writable || {@' "$f" ;;
    M09) sed -i -e '/^norm_lane() {$/,/^}$/ s@ | tr .A-Z. .a-z.)"@)"@' \
                -e '/^norm_lane() {$/,/^}$/ s@\[!a-z0-9._-\]@[!a-zA-Z0-9._-]@' "$f" ;;
    M10) sed -i -e '/^norm_sha() {$/,/^}$/ s@ | tr .A-Z. .a-z.)"@)"@' \
                -e '/^norm_sha() {$/,/^}$/ s@\[!0-9a-f\]@[!0-9a-fA-F]@' "$f" ;;
    M11) sed -i 's|s="${s#@}"|s="${s#@@}"|' "$f" ;;
    M12) sed -i -e 's@(gh\[pousr\]_@(zz[pousr]_@' -e 's@|github_pat_@|zzthub_pat_@' "$f" ;;
    # M13 used to aim at `tr -d '\r\n'`, which SEC-05 replaced with the whole
    # control range. Same neutralisation against the line that now carries it:
    # CR alone, so a newline survives into the record format again.
    M13) sed -i "s@tr -d ${Q}\[:cntrl:\]${Q}@tr -d ${Q}\\\\r${Q}@" "$f" ;;
    # M14 is caught by the directory-name assertion, not by a failed `mv`.
    # claim.sh's header says NTFS "would have rejected it outright"; measured on
    # 2026-08-29 Cygwin's mv creates `released-2026-08-29T12:32:59Z-473354`
    # happily, because Cygwin maps `:` into a private-use codepoint. The guard is
    # still right — a native Win32 reader of that path is not — but the stated
    # mechanism is not what happens through this shell.
    M14) sed -i 's@+%Y%m%dT%H%M%SZ@+%Y-%m-%dT%H:%M:%SZ@' "$f" ;;
    M15) sed -i '/anomaly: no claim row covered/,+1 s@exit 12@exit 0@' "$f" ;;
    M16) sed -i 's@if \[ -r "$1/verdict.d/row" \]; then echo ruled; else echo ruling-incomplete; fi@echo ruled@' "$f" ;;
    M17) sed -i '/^usage_die() {$/a mkdir -p "$TABLE" 2>/dev/null' "$f" ;;
    M18) sed -i '/^cmd_claim() {$/,/^}$/ s@^  if \[ -d "$row/verdict.d" \]; then$@  if [ -r "$row/verdict.d/row" ]; then@' "$f" ;;
    M19) sed -i '/^usage_die() {$/,/^}$/ s@exit 64@exit 0@' "$f" ;;
    # M20 used to aim at the `mv -f` of a bare row file into an already-created
    # verdict.d. The row is now written inside the staging directory before the
    # commit, so the same neutralisation — a ruling that reports itself recorded
    # while no readable row lands — is a row written under another name. The
    # range matters: `cmd_manifest` stages its row identically.
    M20) sed -i '/^cmd_verdict() {$/,/^}$/ s@> "$stage/row" 2>/dev/null@> "$stage/row.notyet" 2>/dev/null@' "$f" ;;
    M21) sed -i 's@{ sed .*"$dst/row" 2>/dev/null@{ cat "$dst/row" 2>/dev/null@' "$f" ;;
    M22) sed -i -e '/^cmd_claim() {$/,/^}$/ s@claim: refused\(.*could not create row\)@claim: granted\1@' \
                -e '/^cmd_claim() {$/,/^}$/ { /could not create row/,+1 s@exit 2@exit 0@ }' "$f" ;;
    M23) sed -i 's@direction: refusing to dispatch; an unwritable table must never read as an unclaimed lane@direction: carry this verdict in your handoff; the table does not hold it@' "$f" ;;
    M24) sed -i 's@if \[ -z "$outstanding" \]; then@if [ -n "$outstanding" ]; then@' "$f" ;;
    M25) sed -i 's@\[ "$n" -ge 7 \] && \[ "$n" -le 40 \]@[ "$n" -ge 1 ] \&\& [ "$n" -le 400 ]@' "$f" ;;
    M26) sed -i 's@^    \.\*|\*\.\.\*)@    .zzz*|*..zzz*)@' "$f" ;;
    M27) sed -i '/nothing was dispatched under this panel/,+1 s@exit 12@exit 0@' "$f" ;;
    # The truncation is the KEY and never the check, so this empties the suffix
    # rather than removing the bound above it: the length check still passes and
    # the row key becomes the caller's spelling again. That is CLAIM-01 exactly,
    # and M01-M27 contained no mutant of it — the fix's only coverage lived in a
    # file the harness never ran, so nothing here could have gone red for it.
    M28) sed -i 's@rest="${s#???????}"@rest=""@' "$f" ;;
    M29) sed -i 's@0?\*) s="${s#0}" ;;@00?*) s="${s#0}" ;;@' "$f" ;;
    # Both edits together are the line SEC-03 filed: the caller's own toplevel,
    # which a linked worktree answers with a path of its own. One alone would
    # not neutralise it — `--show-toplevel` with the parent still taken lands
    # both checkouts on their shared parent directory and they agree by accident.
    M30) sed -i -e 's@git rev-parse --git-common-dir@git rev-parse --show-toplevel@' \
                -e 's@_par="${_gcd%/\*}"@_par="$_gcd"@' "$f" ;;
    M31) sed -i 's@say_table() { printf .table: %s.n. "$TABLE"; }@say_table() { printf "tbl: %s\\n" "$TABLE"; }@' "$f" ;;
    M32) sed -i 's@^  mv -T "$1" "$2" 2>/dev/null$@  mv "$1" "$2" 2>/dev/null@' "$f" ;;
    # The two-step gate LA-02 filed, at its widest: the name readers test for
    # exists before the payload does, so an interruption leaves it empty.
    #
    # M33 IS THE ONE MUTATION WHOSE KILL IS ITSELF A RACE, and it lost that race
    # three times in four. The assertion can only see this defect if the suite's
    # `kill -9` lands between the directory appearing and its payload committing,
    # and measured on 2026-08-30 that interval was a small enough slice of one
    # verdict that the mutant SURVIVED three runs of four — the one kill catching
    # 1 wrecked writer of 12. A demonstration that lands a quarter of the time is
    # not a demonstration; it is a gate that goes red on the schedule instead of
    # on the defect, and this harness's whole claim is to be checkable by anyone
    # at any later date.
    #
    # The `sleep` HOLDS the mutant between the two steps rather than hoping the
    # scheduler will. That is failure class 1 reproduced deliberately instead of
    # waited for, and it neutralises nothing extra: `verdict.d` still comes into
    # existence before its payload, which is the whole of what M33 asserts. The
    # suite calibrates its kill delays against the MUTANT's own verdict time, so
    # the window scales with the sleep rather than being outrun by it.
    M33) sed -i '/stage="\$row\/\.verdict-staging/i mkdir -p "$row/verdict.d" 2>/dev/null; sleep 0.5' "$f" ;;
    M34) sed -i 's@\[ -d "$panel/$l" \] && continue@[ -e "$panel" ] \&\& continue@' "$f" ;;
    M35) sed -i 's@if commit_staged "$stage" "$panel/.manifest.d"; then@if rm -rf "$panel/.manifest.d" 2>/dev/null \&\& commit_staged "$stage" "$panel/.manifest.d"; then@' "$f" ;;
    M36) sed -i 's@printf .lane_state: %s.n. "${one_state:-no-row}"@printf "lane_state_omitted: %s\\n" "${one_state:-no-row}"@' "$f" ;;
    M37) sed -i 's@field artifact @field artifact-never @g' "$f" ;;
    # M38 used to aim at `else mismatch=true; fi`, a one-line form SEC-01's
    # rebuild replaced with a multi-line `if`. Same neutralisation against the
    # line that now carries it: the anomaly branch never sets the flag, so a
    # verdict filed by another hand reads clean. Re-aimed rather than dropped,
    # for the reason M02, M13 and M20 above were.
    M38) sed -i 's@^      mismatch=true$@      mismatch=false@' "$f" ;;
    M39) sed -i "s@tr -d ${Q}\[:cntrl:\]${Q}@tr -d ${Q}\\\\r\\\\n${Q}@" "$f" ;;
    M40) sed -i '/^cmd_release() {$/,/^}$/ s@if \[ -r "$row/verdict.d/row" \]; then@if [ -d "$row/verdict.d" ]; then@' "$f" ;;
    M41) sed -i 's@field unclaimed @field unclaimed-never @g' "$f" ;;
    # SEC-03's sentence, in its purest form: the table root is the caller's cwd.
    # M30 above is the same defect through `--show-toplevel`, which is blind to a
    # subdirectory of one checkout; this one is not, and the four-cwd claim needs
    # both to be demonstrated rather than one.
    M42) sed -i 's@ROOT="$(table_root || pwd)"@ROOT="$(pwd)"@' "$f" ;;
    # The one behaviour here that is an ABSENCE — no verdict vocabulary is
    # enforced, deliberately, because a check could only ever refuse a verdict it
    # did not recognise and lose the work. Neutralising an absence means adding
    # the check, so this is the shape of the well-meant "fix" that would silently
    # start refusing `@validation-agent`'s own two words.
    M43) sed -i 's@^  \[ -n "$V" \] || usage_die@  case "$V" in pass|fail) usage_die "verdict — unrecognised ruling" ;; esac; [ -n "$V" ] || usage_die@' "$f" ;;
    # M44 IS THE DEFECT THAT SHIPPED, and it shipped because nothing here could
    # see it. SEC-01's first remedy compared `--by` against `owner:`; the judge
    # holds every lane it claims on a reviewer's behalf, so that comparison
    # flagged every honest verdict this table will ever see. It was caught by
    # hand, on a throwaway table, after the commit. This mutation puts the
    # comparand back exactly as it was, and the assertion it must kill is the
    # one whose ABSENCE let it through — the clean ordinary path.
    M44) sed -i 's@agent_key "$holder_expects"@agent_key "$holder_owner"@' "$f" ;;
    # The same false anomaly, arriving through the SPELLING instead of the
    # field: dispatches name agents `@docs-reviewer`, `norm_lane` folds a lane
    # and `--by` is free text that is not folded. Without the fold the two agree
    # as identities and differ as strings, which is the shape of the two SHA
    # spellings that once split a panel.
    #
    # RE-AIMED, because the line it removed no longer exists. SEC-05 anchored the
    # strip and added a trim, splitting `agent_key` into two statements and taking
    # `| tr -d '@' | tr 'A-Z' 'a-z'` out with it — so this edit matched nothing and
    # M45 demonstrated nothing at the very head that hardened the function. The
    # neutralisation is unchanged in SCOPE as well as in kind: both halves of the
    # old pipeline are removed where they now live, one per expression, and the
    # trim SEC-05 added is left alone because it is a different behaviour and
    # neutralising it here would make the kill ambiguous between the two.
    M45) sed -i -e "/^agent_key() {\$/,/^}\$/ s@ | tr ${Q}A-Z${Q} ${Q}a-z${Q}@@" \
                -e '/^agent_key() {$/,/^}$/ s|"${_ak#@}"|"$_ak"|' "$f" ;;
    # And a third time, by DATA AGE: a claim row written before `expects:`
    # existed carries none, so without the fallback an upgraded table flags
    # every honest verdict already in it.
    M46) sed -i 's@|| holder_expects="$LANE"@|| holder_expects="none-recorded"@' "$f" ;;
    # A deleted flag that is quietly ACCEPTED and dropped is worse than one that
    # never existed: the caller believes the table holds text it does not hold.
    # This is the shape of `--note` being put back by a caller that missed LA-10,
    # so the refusal is what must be demonstrated rather than assumed.
    M47) sed -i "/^cmd_verdict() {\$/,/^  done\$/ s@^      \*) usage_die@      --note) shift 2 ;;\n      *) usage_die@" "$f" ;;
    # CLAIM-B, the verdict this mechanism DESTROYED. `mv -T` refuses onto a
    # non-empty directory and that refusal is the write-once gate — but it also
    # refuses onto a plain file standing where `verdict.d` belongs, and reading
    # every refusal as the gate reported `already ruled` at 10 forever, after
    # `rm -rf` had taken the only copy of the ruling. Making the test false is
    # the whole neutralisation: control falls through to exactly the two lines
    # that shipped, so the mutant IS the defect rather than a caricature of it.
    M48) sed -i 's@^  if \[ ! -d "$row/verdict.d" \]; then$@  if [ -d "$row/verdict.d/never" ]; then@' "$f" ;;
    # LQ-05. `verdict` is deliberately not pause-gated, so the gate is on
    # CREATION and nothing else: making its first test false lets `mkdir -p`
    # build `pr-<n>/`, the SHA and the lane out of nothing again, under the one
    # state in which the least should be believed about coverage. `is_paused` is
    # left in place so this cannot be confused with M07.
    M49) sed -i 's@^  if \[ ! -d "$row" \] && is_paused; then$@  if [ -d "$row/never" ] \&\& is_paused; then@' "$f" ;;
    # CLAIM-C, at the property that makes an amendable manifest safe. The
    # amendment still happens, still requires its reason, and still cites the set
    # it replaced by name and count — the prior is simply not THERE any more, so
    # every word of the amendment is true and unverifiable. That is the silent
    # shrink write-once existed to prevent, wearing the audit trail as a coat.
    M50) sed -i 's@^    mv "$panel/.manifest.d" "$pdst" 2>/dev/null || {$@    rm -rf "$panel/.manifest.d" 2>/dev/null; mkdir -p "$pdst" 2>/dev/null || {@' "$f" ;;
    # SEC-03. The row still carries the field, the read verbs still surface it,
    # and it is always `recorded` — which is the state before the field existed
    # with a label on it saying otherwise. An inference reading as something a
    # claim asked for is the accept direction of the attribution check, arriving
    # one field to the left of the flag it qualifies.
    M51) sed -i 's@^  if \[ -n "$holder_expects" \]; then expects_source=recorded; else expects_source=inferred; fi$@  expects_source=recorded@' "$f" ;;
  esac
}

bad=0; killed=0
printf 'red demonstration for %s\n' "$SUITE"
printf 'script under test: %s\n\n' "$ORIG"

# The baseline names every assertion, in order, so the coverage report below can
# say which of them no mutation ever moved. Every assertion is unconditional, so
# the Nth reported line is the same assertion in every run.
BASE="$TMP/baseline.log"
if ! bash "$SUITE" > "$BASE" 2>&1; then
  printf 'the suite is not green against the unmutated script — fix that first:\n' >&2
  grep '^FAIL' "$BASE" >&2
  exit 1
fi
awk '/^(PASS|FAIL)  /{n++; lbl=$0; sub(/^(PASS|FAIL)  /,"",lbl); printf "%03d\t%s\n", n, lbl}' \
  "$BASE" > "$TMP/all.labels"
: > "$TMP/red.ord"

for m in $WANT; do
  MUT="$TMP/$m.sh"
  cp "$ORIG" "$MUT"
  mutate "$m" "$MUT"

  if cmp -s "$ORIG" "$MUT"; then
    printf '%s  MISAPPLIED  %s — the edit matched nothing; it demonstrates nothing\n' "$m" "$(desc "$m")"
    bad=$((bad + 1)); continue
  fi
  if ! bash -n "$MUT" 2>/dev/null; then
    printf '%s  BROKEN      %s — the mutant does not parse; red would be for the wrong reason\n' "$m" "$(desc "$m")"
    bad=$((bad + 1)); continue
  fi
  if ! bash "$MUT" help >/dev/null 2>&1; then
    printf '%s  BROKEN      %s — the mutant cannot answer `help`; red would be for the wrong reason\n' "$m" "$(desc "$m")"
    bad=$((bad + 1)); continue
  fi
  # `help` and `bash -n` are too weak on their own: both passed a mutant whose
  # norm_sha had been turned to garbage by an unescaped & in a sed replacement,
  # and its 107 spurious failures were counted as red. No mutation in this matrix
  # should stop an ordinary claim on an empty table, so that is the real smoke test.
  if ! CLAIM_TABLE_DIR="$TMP/$m.smoke" bash "$MUT" claim 1 abc1234 smoke >/dev/null 2>&1; then
    printf '%s  BROKEN      %s — the mutant cannot grant an ordinary claim; red would be for the wrong reason\n' "$m" "$(desc "$m")"
    bad=$((bad + 1)); continue
  fi

  LOG="$TMP/$m.log"
  CLAIM_SH="$MUT" bash "$SUITE" > "$LOG" 2>&1
  rc=$?
  nfail="$(grep -c '^FAIL' "$LOG" 2>/dev/null | tr -d ' ')"
  want="$(kills "$m")"
  line="$(grep '^FAIL' "$LOG" 2>/dev/null | grep -F "$want" | head -1)"

  if [ "$rc" -eq 0 ]; then
    printf '%s  SURVIVED    %s — the suite stayed GREEN. Nothing here asserts it.\n' "$m" "$(desc "$m")"
    bad=$((bad + 1)); continue
  fi
  if [ -z "$line" ]; then
    printf '%s  WRONG-RED   %s — suite red (%s failures) but [%s] was not among them\n' "$m" "$(desc "$m")" "$nfail" "$want"
    bad=$((bad + 1)); continue
  fi

  awk '/^(PASS|FAIL)  /{n++; if ($0 ~ /^FAIL  /) printf "%03d\n", n}' "$LOG" >> "$TMP/red.ord"

  killed=$((killed + 1))
  printf '%s  KILLED      %s\n' "$m" "$(desc "$m")"
  printf '      red: %s\n' "$(printf '%s' "$line" | sed 's/^FAIL  *//')"
  printf '      (%s assertions failed under this mutation)\n' "$nfail"
done

# Which assertions no mutation ever made fail. Reported, not failed: some of them
# guard this suite's own hermeticity rather than a behaviour of claim.sh, and no
# mutation of claim.sh should be able to move those.
#
# AND IT IS ONLY THAT REPORT AFTER A FULL RUN. `comm -23 all red` computes "every
# assertion no mutation IN THIS RUN made fail", and under a named subset that is
# very nearly every assertion in the suite. Measured on 2026-09-06 at this head:
# `M28 M33` — the subset the usage line above recommends — kills both its
# mutations and shows 11 of 366 assertions red, so the report as it stood named
# the other 355 as undemonstrated. Not one of the 355 was a finding about the
# suite; they are an artefact of what was asked for, printed in the imperative
# voice of a finding, directly above the exit code. An alarm that fires on the
# ordinary use of the tool is an alarm its reader learns to skip, and the
# assertions it names after a FULL run are the ones that most need reading. So
# the full matrix keeps the report and a subset states what it measured and
# stops there. The comparison is on the SET of ids rather than on the string,
# because naming all 51 explicitly is a full run and must not read as a partial.
if [ -s "$TMP/all.labels" ]; then
  sort -u "$TMP/red.ord" > "$TMP/red.u"
  cut -f1 "$TMP/all.labels" | sort -u > "$TMP/all.u"
  nred="$(wc -l < "$TMP/red.u" | tr -d ' ')"
  nall="$(wc -l < "$TMP/all.u" | tr -d ' ')"
  want_set="$(printf '%s\n' $WANT | sort -u | tr '\n' ' ')"
  all_set="$(printf '%s\n' $ALL  | sort -u | tr '\n' ' ')"
  if [ "$want_set" = "$all_set" ]; then
    comm -23 "$TMP/all.u" "$TMP/red.u" > "$TMP/never.u"
    printf '\n-- assertion coverage: %s of %s assertions were shown red\n' "$nred" "$nall"
    nnever="$(wc -l < "$TMP/never.u" | tr -d ' ')"
    if [ "$nnever" -gt 0 ]; then
      printf '   %s never failed under any mutation, and are undemonstrated:\n' "$nnever"
      while read -r o; do
        printf '     %s\n' "$(awk -F'\t' -v k="$o" '$1==k{print $2}' "$TMP/all.labels")"
      done < "$TMP/never.u"
    fi
  else
    printf '\n-- assertion coverage: PARTIAL. %s of %s assertions were shown red by the\n' "$nred" "$nall"
    printf '   %s mutation(s) asked for:%s\n' "$(printf '%s\n' $WANT | wc -l | tr -d ' ')" \
      "$(printf ' %s' $WANT)"
    printf '   This says nothing about the other %s. The undemonstrated list is a claim\n' \
      "$(( nall - nred ))"
    printf '   about the whole suite and needs the whole matrix: run `%s` with no arguments.\n' \
      "$(basename "$0")"
  fi
fi

printf '\n%s killed · %s not demonstrated\n' "$killed" "$bad"
[ "$bad" -eq 0 ] || { printf 'a mutation that is not killed is an assertion that has never been red\n'; exit 1; }
printf 'every mutation was caught by a named assertion\n'
exit 0
