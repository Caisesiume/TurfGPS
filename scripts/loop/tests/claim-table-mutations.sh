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
# A MUTATION THAT MATCHES NOTHING IS A DEFECT HERE, AND THREE OF THEM WERE.
# M02, M13 and M20 addressed lines the review board's fixes removed — the mkdir
# verdict gate, the `\r\n` strip, the `mv -f` of a bare row file. Each still
# named a real behaviour, so each was re-aimed at the line that now carries it
# rather than dropped; a mutation retired quietly is an assertion that stops
# being demonstrated with nothing saying so. The `cmp -s` guard below is what
# caught all three, which is the argument for keeping it.
#
# Usage: scripts/loop/tests/claim-table-mutations.sh [id …]
#        Exit: 0 every mutation killed · 1 any mutation survived or misapplied
# The whole suite runs once per mutation. Measured on 2026-08-30 on the
# reference host: one suite run is ~90s and the full matrix of 43 is a little
# under an hour, so it is a background gate rather than an inner-loop one. Named
# ids run a subset — `claim-table-mutations.sh M28 M33` — and that is how a
# single behaviour is re-demonstrated after a change without paying for all 43.

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
M38 M39 M40 M41 M42 M43'
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
    M33) sed -i '/stage="\$row\/\.verdict-staging/i mkdir -p "$row/verdict.d" 2>/dev/null' "$f" ;;
    M34) sed -i 's@\[ -d "$panel/$l" \] && continue@[ -e "$panel" ] \&\& continue@' "$f" ;;
    M35) sed -i 's@if commit_staged "$stage" "$panel/.manifest.d"; then@if rm -rf "$panel/.manifest.d" 2>/dev/null \&\& commit_staged "$stage" "$panel/.manifest.d"; then@' "$f" ;;
    M36) sed -i 's@printf .lane_state: %s.n. "${one_state:-no-row}"@printf "lane_state_omitted: %s\\n" "${one_state:-no-row}"@' "$f" ;;
    M37) sed -i 's@field artifact @field artifact-never @g' "$f" ;;
    M38) sed -i 's@else mismatch=true; fi@else mismatch=false; fi@' "$f" ;;
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
if [ -s "$TMP/all.labels" ]; then
  sort -u "$TMP/red.ord" > "$TMP/red.u"
  cut -f1 "$TMP/all.labels" | sort -u > "$TMP/all.u"
  comm -23 "$TMP/all.u" "$TMP/red.u" > "$TMP/never.u"
  printf '\n-- assertion coverage: %s of %s assertions were shown red\n' \
    "$(wc -l < "$TMP/red.u" | tr -d ' ')" "$(wc -l < "$TMP/all.u" | tr -d ' ')"
  nnever="$(wc -l < "$TMP/never.u" | tr -d ' ')"
  if [ "$nnever" -gt 0 ]; then
    printf '   %s never failed under any mutation, and are undemonstrated:\n' "$nnever"
    while read -r o; do
      printf '     %s\n' "$(awk -F'\t' -v k="$o" '$1==k{print $2}' "$TMP/all.labels")"
    done < "$TMP/never.u"
  fi
fi

printf '\n%s killed · %s not demonstrated\n' "$killed" "$bad"
[ "$bad" -eq 0 ] || { printf 'a mutation that is not killed is an assertion that has never been red\n'; exit 1; }
printf 'every mutation was caught by a named assertion\n'
exit 0
