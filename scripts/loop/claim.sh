#!/usr/bin/env bash
# claim.sh — the review claim table: a lane is claimed BEFORE it is dispatched,
# and its verdict is durable the instant it exists (issue #144).
#
# It replaces the post-hoc ledger of
# `review-board-dispatch § Incremental review validity`, which was composed after
# the panel ran and could therefore omit a lane it never heard back from — on
# #140 it recorded 4 lanes and 11 findings where 5 and 13 had landed, while
# asserting that every lane above had run at this SHA. Here
# the row exists before the dispatch does, so the table IS the panel rather than
# a retelling of it, and it cannot omit a row it created. The consequence that
# matters most: the absence of a verdict stops being indistinguishable from the
# absence of a dispatch, which is the inference that put three judges on #141
# and four on #140.
#
# Deterministic and LLM-free. No network, no GitHub read, no judgement. The table
# answers three questions and no others — does a row exist · is it held · does it
# carry a verdict. It knows nothing about what any agent reviews or what counts
# as done for it: a checker that did would be a second, unreviewed opinion about
# completion, which is the design considered and rejected for #139. That is also
# why no verdict vocabulary is enforced below; `review-verdicts` owns what a
# verdict word means, and a vocabulary check here could only ever refuse a
# verdict it did not recognise, losing the work this script exists to keep.
#
# Exit codes — branch on these, do not parse the prose:
#   0   OK        claim granted · verdict recorded · panel complete · flag set
#   10  REFUSED   the lane is already held, or already ruled; or the panel is not
#                 complete yet. An expected, ordinary "no".
#   11  PAUSED    new claims are paused. Rows already held may still record.
#   12  NO_ROW    nothing was claimed here — no such panel or lane; or a verdict
#                 arrived for a lane no claim covers (recorded anyway, flagged).
#   2   DEGRADED  state could not be read, written, or locked.
#   64  USAGE     malformed arguments. Nothing was read and nothing was written.
#
# WHICH WAY THIS FAILS, AND WHY THAT DIRECTION
#   Unreadable state, an unwritable table, a missing directory that cannot be
#   made, a claim directory that cannot be taken: every one of them REFUSES.
#   For `claim` the closed direction is refusing to dispatch — never dispatching
#   twice. For `verdict` the closed direction is the opposite one, and saying so
#   is the point: a verdict that cannot be written exits NONZERO and says it was
#   not recorded, so the reviewer carries it in its handoff instead of believing
#   the table holds it. Silent success is the direction that loses work; it is
#   never taken by either.
#
# THE ATOMIC PRIMITIVE
#   A claim is `mkdir` of one directory. The kernel decides the winner, not a
#   read followed by a write, so two concurrent claimants on one lane at one SHA
#   produce exactly one 0 and one 10. There is no window between the test and the
#   set because there is no test. This holds on Git Bash for Windows, where the
#   call lands on CreateDirectory, which fails when the name exists.
#
# STATE
#   .claude/state/review-claims/ — inside the `/.claude/state/` entry .gitignore
#   already carries, the established home for durable local control-plane state,
#   alongside fingerprint.sh's per-consumer files. CLAIM_TABLE_DIR overrides the
#   location, which is how a test runs against a throwaway root rather than
#   against the machine's real table.
#
#     <table>/PAUSED                                 flag: no new claims while it exists
#     <table>/pr-<n>/<sha>/<lane>/holder/row         the claim
#     <table>/pr-<n>/<sha>/<lane>/verdict.d/row      the verdict, written once, immutable
#     <table>/pr-<n>/<sha>/<lane>/released-<stamp>/row   a claim given back, kept as audit
#
#   This script never reads GH_JUDGE_TOKEN, never invokes gh, and never records
#   the environment. Free-text fields are additionally scrubbed of token-shaped
#   substrings before they are written, so a caller who pastes a credential into
#   --note or --artifact cannot put it into the table either.
#
# Usage: scripts/loop/claim.sh <subcommand> [args…]   —  `claim.sh help` prints the surface

set -u
LC_ALL=C
export LC_ALL

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
TABLE="${CLAIM_TABLE_DIR:-$ROOT/.claude/state/review-claims}"
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)"
[ -n "$NOW" ] || NOW="unknown-time"
# A second stamp, for the one place a timestamp becomes a FILENAME. NTFS rejects
# `:`, so the ISO form above cannot name a directory on this machine — `mv` would
# fail and the release would report degraded on a table that was perfectly fine.
STAMP="$(date -u +%Y%m%dT%H%M%SZ 2>/dev/null)"
[ -n "$STAMP" ] || STAMP="unknown"

# ---------------------------------------------------------------------------
# Argument canonicalisation. These values become path components, so each is
# validated rather than quoted-and-hoped: a lane carrying `/` or `..` would
# write outside the table.
#
# Lane and SHA are lower-cased. NTFS is case-insensitive, so `Foo` and `foo`
# are one directory on this machine and two on Linux; canonicalising makes the
# row the same one everywhere, and in the safe direction — a lane must not
# become claimable a second time by changing the case of its name.
#
# ONE COMMIT IS ONE EPOCH, HOWEVER ITS SHA IS SPELLED
#   The epoch used to be the caller's string verbatim, and `ec9ee33` and
#   `ec9ee330b2cfe0f9164eaa7f3dee22c23c4afdc3` are the two ordinary ways an agent
#   obtains one head SHA — `git log --oneline` and `git rev-parse HEAD`. Both
#   validate, so one commit built TWO panels that could not see each other: both
#   judges granted, the verdict filed `unclaimed` against the spelling its
#   claimant did not use, one panel outstanding at rc 10 forever while a second
#   read complete at rc 0. That is #144's failure classes 1, 2 and 6 reopened
#   through the mechanism that exists to close them.
#
#   The row key is therefore the first SEVEN characters of the folded SHA, and
#   seven is not a taste: it is the shortest spelling this script accepts, so it
#   is the only prefix every accepted spelling of one commit is guaranteed to
#   share. A longer key cannot unify a 7-hex caller with a 40-hex one. The whole
#   string is still validated as 7-40 hex first — the truncation is the KEY and
#   never the check — and the key is echoed back, so a caller who passes 40 sees
#   which panel it actually joined rather than the one it typed.
#
#   Resolving short to long through `git rev-parse` was the other candidate and
#   is rejected on purpose: it would put a subprocess, a repository and a new
#   failure mode into a script whose whole value is that it answers without one,
#   and it would answer differently — or not at all — on the degraded path this
#   file is built around. Scanning the panel for a prefix match was rejected for
#   a harder reason: it is a read before the write, and the claim gate is one
#   `mkdir` precisely so that no such window exists.
#
#   The trade, stated here rather than discovered later: two DIFFERENT commits
#   sharing a seven-hex prefix under ONE pr number would share a panel. That is
#   28 bits against the handful of head SHAs a single PR ever carries, and the
#   alternative costs the git dependency above. If it ever bites, the answer is
#   a longer key, not a lookup.
#
#   `pr-0144` and `pr-144` split a panel the same way and for the same reason,
#   so the PR number sheds its leading zeros here too. One fix, both keys.
# ---------------------------------------------------------------------------

norm_pr() {
  s="${1:-}"
  case "$s" in
    ''|*[!0-9]*) return 1 ;;
  esac
  # `0144` and `144` are one PR. Shed leading zeros, never the last digit:
  # `0` is a number and the empty string is not a path component.
  while :; do
    case "$s" in
      0?*) s="${s#0}" ;;
      *)   break ;;
    esac
  done
  printf '%s' "$s"
}

norm_sha() {
  s="$(printf '%s' "${1:-}" | tr 'A-Z' 'a-z')"
  case "$s" in
    ''|*[!0-9a-f]*) return 1 ;;
  esac
  n=${#s}
  [ "$n" -ge 7 ] && [ "$n" -le 40 ] || return 1
  # The epoch — the first seven characters, not the spelling the caller used.
  # See ONE COMMIT IS ONE EPOCH above. The guard is not redundant with the bound
  # on the line above it: it keeps the truncation honest if that bound ever
  # moves, because a key that came out empty would put every panel in one row.
  case "$s" in
    ???????*) rest="${s#???????}"; s="${s%"$rest"}" ;;
  esac
  printf '%s' "$s"
}

norm_lane() {
  s="${1:-}"
  s="${s#@}"                                   # dispatches name lanes as @agent-name
  s="$(printf '%s' "$s" | tr 'A-Z' 'a-z')"
  case "$s" in
    ''|*[!a-z0-9._-]*) return 1 ;;
    .*|*..*)           return 1 ;;             # no traversal, no hidden rows
  esac
  printf '%s' "$s"
}

# One line, no control characters, capped, and with anything token-shaped removed.
# The pattern is matched, not compared against a value: this script does not read
# GH_JUDGE_TOKEN, so it cannot leak it by accident of comparison either.
scrub() {
  printf '%s' "${1:-}" \
    | tr -d '\r\n' \
    | sed -E 's/(gh[pousr]_[A-Za-z0-9]{16,}|github_pat_[A-Za-z0-9_]{16,})/[redacted]/g' \
    | cut -c1-200
}

usage_die() {
  printf 'usage: claim.sh %s\n' "$1" >&2
  printf 'run `claim.sh help` for the full command surface\n' >&2
  exit 64
}

# ---------------------------------------------------------------------------
# State access. Writability is established by writing, not by `[ -w ]`: under
# MSYS the test operator answers from a translated view of the Windows ACL and
# has been wrong in both directions, and creating a file is the exact capability
# every write path here needs.
# ---------------------------------------------------------------------------

ensure_writable() {
  mkdir -p "$TABLE" 2>/dev/null || return 1
  [ -d "$TABLE" ] || return 1
  p="$TABLE/.probe.$$"
  : > "$p" 2>/dev/null || return 1
  rm -f "$p" 2>/dev/null
  return 0
}

# A table that does not exist yet is not a fault — it is a table with no rows.
# A table that exists and cannot be listed is a fault, and refuses.
ensure_readable() {
  [ -e "$TABLE" ] || return 0
  [ -d "$TABLE" ] || return 1
  ls "$TABLE" >/dev/null 2>&1 || return 1
  return 0
}

is_paused() { [ -e "$TABLE/PAUSED" ]; }

field() { # field <key> <file>
  [ -r "$2" ] || return 0
  grep "^$1: " "$2" 2>/dev/null | head -1 | cut -d' ' -f2-
}

# claimed · ruled · ruling-incomplete · free. `ruling-incomplete` is a verdict
# directory holding no row: a reviewer died mid-write. It is reported rather than
# repaired, and it counts as outstanding — visible debt, which is the whole point.
row_state() { # row_state <row-dir>
  if [ -d "$1/verdict.d" ]; then
    if [ -r "$1/verdict.d/row" ]; then echo ruled; else echo ruling-incomplete; fi
  elif [ -d "$1/holder" ]; then
    echo claimed
  else
    echo free
  fi
}

# ---------------------------------------------------------------------------
# claim — the atomic gate. Everything else in this file is bookkeeping.
# ---------------------------------------------------------------------------
cmd_claim() {
  [ $# -ge 3 ] || usage_die 'claim <pr> <sha> <lane> [--owner <who>]'
  PR="$(norm_pr "$1")"     || usage_die 'claim <pr> <sha> <lane> — <pr> must be digits'
  SHA="$(norm_sha "$2")"   || usage_die 'claim <pr> <sha> <lane> — <sha> must be 7-40 hex'
  LANE="$(norm_lane "$3")" || usage_die 'claim <pr> <sha> <lane> — <lane> must be [a-z0-9._-]'
  shift 3
  owner="unnamed"
  while [ $# -gt 0 ]; do
    case "$1" in
      --owner) [ $# -ge 2 ] || usage_die 'claim … --owner <who>'; owner="$(scrub "$2")"; shift 2 ;;
      *)       usage_die 'claim <pr> <sha> <lane> [--owner <who>]' ;;
    esac
  done
  [ -n "$owner" ] || owner="unnamed"

  ensure_writable || {
    printf 'claim: refused\nreason: degraded — table not writable at %s\n' "$TABLE"
    printf 'direction: refusing to dispatch; an unwritable table must never read as an unclaimed lane\n'
    exit 2
  }

  # The pause gate is here and nowhere else. It stops the cascade at the next
  # hop — no new claim, therefore no new dispatch — while every row already held
  # can still record its verdict below. That is what a pause should mean.
  if is_paused; then
    printf 'claim: refused\nreason: paused\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    [ -r "$TABLE/PAUSED" ] && sed 's/^/  /' "$TABLE/PAUSED" 2>/dev/null
    exit 11
  fi

  row="$TABLE/pr-$PR/$SHA/$LANE"
  mkdir -p "$row" 2>/dev/null || {
    printf 'claim: refused\nreason: degraded — could not create row %s\n' "$row"
    exit 2
  }

  if [ -d "$row/verdict.d" ]; then
    printf 'claim: refused\nreason: ruled\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    if [ -r "$row/verdict.d/row" ]; then
      printf 'of_record: %s\n' "$(field verdict "$row/verdict.d/row")"
    else
      # A ruling is in flight, or one was interrupted mid-write. Both refuse the
      # claim: re-dispatching a lane that is being ruled is how #141 acquired two
      # contradictory packets sixteen seconds apart. `status` shows it as
      # ruling-incomplete rather than as a lane anyone may pick up.
      printf 'of_record: unreadable — a ruling is in flight or was interrupted\n'
    fi
    exit 10
  fi

  # THE gate. `mkdir` without -p fails when the name exists, so exactly one of
  # two concurrent callers reaches the next line.
  if mkdir "$row/holder" 2>/dev/null; then
    if [ -d "$row/verdict.d" ]; then
      # A verdict landed between the check above and the gate. Give the claim
      # straight back; a ruled lane is never re-dispatched.
      mv "$row/holder" "$row/released-$STAMP-$$" 2>/dev/null
      printf 'claim: refused\nreason: ruled\nlane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
      exit 10
    fi
    {
      printf 'lane: %s\n'       "$LANE"
      printf 'pr: %s\n'         "$PR"
      printf 'sha: %s\n'        "$SHA"
      printf 'state: claimed\n'
      printf 'owner: %s\n'      "$owner"
      printf 'claimed_at: %s\n' "$NOW"
    } > "$row/holder/row" 2>/dev/null || {
      # The directory is held and its row could not be written. The directory
      # STAYS: removing it would readmit the second dispatch this call just shut
      # out. The caller does not dispatch, and `release` is how it is given back.
      printf 'claim: refused\nreason: degraded — holder taken, row not writable\n'
      printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
      printf 'note: the lane is held and undispatched; release it with `claim.sh release`\n'
      exit 2
    }
    printf 'claim: granted\nlane: %s\npanel: pr-%s @ %s\nowner: %s\nclaimed_at: %s\n' \
      "$LANE" "$PR" "$SHA" "$owner" "$NOW"
    exit 0
  fi

  # The gate refused. Held is the ordinary reason; anything else is degraded,
  # and both refuse — only the label differs.
  if [ -d "$row/holder" ]; then
    printf 'claim: refused\nreason: held\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    printf 'holder: %s since %s\n' \
      "$(field owner "$row/holder/row")" "$(field claimed_at "$row/holder/row")"
    exit 10
  fi
  printf 'claim: refused\nreason: degraded — could not take the claim directory\n'
  printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
  exit 2
}

# ---------------------------------------------------------------------------
# verdict — durability. Written once, by its own reviewer, at the moment it
# exists. A pause does not block it: in-flight work completes.
# ---------------------------------------------------------------------------
cmd_verdict() {
  [ $# -ge 4 ] || usage_die 'verdict <pr> <sha> <lane> <verdict> [--conf <x>] [--findings <n>] [--artifact <ref>] [--note <text>]'
  PR="$(norm_pr "$1")"     || usage_die 'verdict <pr> … — <pr> must be digits'
  SHA="$(norm_sha "$2")"   || usage_die 'verdict <pr> <sha> … — <sha> must be 7-40 hex'
  LANE="$(norm_lane "$3")" || usage_die 'verdict <pr> <sha> <lane> … — <lane> must be [a-z0-9._-]'
  V="$(scrub "$4")"
  [ -n "$V" ] || usage_die 'verdict <pr> <sha> <lane> <verdict> — <verdict> may not be empty'
  shift 4
  conf="-"; findings="-"; artifact="-"; note="-"
  while [ $# -gt 0 ]; do
    case "$1" in
      --conf|--confidence) [ $# -ge 2 ] || usage_die 'verdict … --conf <x>';     conf="$(scrub "$2")";     shift 2 ;;
      --findings)          [ $# -ge 2 ] || usage_die 'verdict … --findings <n>'; findings="$(scrub "$2")"; shift 2 ;;
      --artifact)          [ $# -ge 2 ] || usage_die 'verdict … --artifact <ref>'; artifact="$(scrub "$2")"; shift 2 ;;
      --note)              [ $# -ge 2 ] || usage_die 'verdict … --note <text>';  note="$(scrub "$2")";     shift 2 ;;
      *) usage_die 'verdict <pr> <sha> <lane> <verdict> [--conf <x>] [--findings <n>] [--artifact <ref>] [--note <text>]' ;;
    esac
  done

  ensure_writable || {
    printf 'verdict: NOT RECORDED\nreason: degraded — table not writable at %s\n' "$TABLE"
    printf 'direction: carry this verdict in your handoff; the table does not hold it\n'
    exit 2
  }

  row="$TABLE/pr-$PR/$SHA/$LANE"
  mkdir -p "$row" 2>/dev/null || {
    printf 'verdict: NOT RECORDED\nreason: degraded — could not create row %s\n' "$row"
    printf 'direction: carry this verdict in your handoff; the table does not hold it\n'
    exit 2
  }

  # Same primitive as the claim, for the same reason: one ruling per lane per
  # SHA, decided by the kernel. Two judges ruled on #141 sixteen seconds apart
  # with contradictory packets; under this gate the second one is refused and
  # the first stays of record.
  if mkdir "$row/verdict.d" 2>/dev/null; then
    unclaimed=false
    [ -d "$row/holder" ] || unclaimed=true
    tmp="$row/verdict.d/.row.$$"
    {
      printf 'lane: %s\n'     "$LANE"
      printf 'pr: %s\n'       "$PR"
      printf 'sha: %s\n'      "$SHA"
      printf 'verdict: %s\n'  "$V"
      printf 'confidence: %s\n' "$conf"
      printf 'findings: %s\n' "$findings"
      printf 'artifact: %s\n' "$artifact"
      printf 'note: %s\n'     "$note"
      printf 'owner: %s\n'    "$(field owner "$row/holder/row")"
      printf 'unclaimed: %s\n' "$unclaimed"
      printf 'ruled_at: %s\n' "$NOW"
    } > "$tmp" 2>/dev/null && mv -f "$tmp" "$row/verdict.d/row" 2>/dev/null || {
      rm -f "$tmp" 2>/dev/null
      printf 'verdict: NOT RECORDED\nreason: degraded — row not writable\n'
      printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
      printf 'direction: carry this verdict in your handoff; the table does not hold it\n'
      exit 2
    }
    printf 'verdict: recorded\nlane: %s\npanel: pr-%s @ %s\nruling: %s\nruled_at: %s\n' \
      "$LANE" "$PR" "$SHA" "$V" "$NOW"
    if [ "$unclaimed" = true ]; then
      # Durable first, loud second. Refusing here would lose a verdict to enforce
      # a rule about dispatch, which is the wrong trade; the anomaly is recorded
      # in the row and reported nonzero so the caller cannot miss it.
      printf 'anomaly: no claim row covered this lane — the verdict is stored and flagged `unclaimed`\n'
      exit 12
    fi
    exit 0
  fi

  printf 'verdict: refused\nreason: already ruled\n'
  printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
  if [ -r "$row/verdict.d/row" ]; then
    printf 'of_record: %s (conf %s, findings %s) at %s\n' \
      "$(field verdict "$row/verdict.d/row")" "$(field confidence "$row/verdict.d/row")" \
      "$(field findings "$row/verdict.d/row")" "$(field ruled_at "$row/verdict.d/row")"
  else
    printf 'of_record: unreadable — a ruling is in flight or was interrupted\n'
  fi
  exit 10
}

# ---------------------------------------------------------------------------
# release — the resume tool whose absence caused two of the duplicate dispatches
# ("I caused two of those dispatches trying to recover a verdict I had no tool
# to resume"). It hands a stranded claim back so the lane can be re-dispatched.
# It is deliberately explicit: a reason is required, a ruled lane is refused, and
# the released claim is kept as an audit row rather than deleted.
# ---------------------------------------------------------------------------
cmd_release() {
  [ $# -ge 3 ] || usage_die 'release <pr> <sha> <lane> --reason <text>'
  PR="$(norm_pr "$1")"     || usage_die 'release <pr> … — <pr> must be digits'
  SHA="$(norm_sha "$2")"   || usage_die 'release <pr> <sha> … — <sha> must be 7-40 hex'
  LANE="$(norm_lane "$3")" || usage_die 'release <pr> <sha> <lane> … — <lane> must be [a-z0-9._-]'
  shift 3
  reason=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --reason) [ $# -ge 2 ] || usage_die 'release … --reason <text>'; reason="$(scrub "$2")"; shift 2 ;;
      *)        usage_die 'release <pr> <sha> <lane> --reason <text>' ;;
    esac
  done
  [ -n "$reason" ] || usage_die 'release <pr> <sha> <lane> --reason <text> — a reason is required'

  ensure_writable || { printf 'release: refused\nreason: degraded — table not writable\n'; exit 2; }
  row="$TABLE/pr-$PR/$SHA/$LANE"

  if [ -d "$row/verdict.d" ]; then
    printf 'release: refused\nreason: ruled — a lane that has ruled is finished, not stranded\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    exit 10
  fi
  if [ ! -d "$row/holder" ]; then
    printf 'release: refused\nreason: no claim to release\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    exit 12
  fi

  dst="$row/released-$STAMP-$$"
  mv "$row/holder" "$dst" 2>/dev/null || {
    printf 'release: refused\nreason: degraded — could not move the claim aside\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    exit 2
  }
  # Rewritten rather than appended. A second `state:` line would leave the audit
  # row still reading `claimed` to anything that takes the first match, which is
  # what every reader in this file does.
  tmp="$dst/.row.$$"
  if [ -r "$dst/row" ]; then
    { sed 's/^state: .*/state: released/' "$dst/row" 2>/dev/null
      printf 'released_at: %s\n' "$NOW"
      printf 'release_reason: %s\n' "$reason"
    } > "$tmp" 2>/dev/null && mv -f "$tmp" "$dst/row" 2>/dev/null || rm -f "$tmp" 2>/dev/null
  else
    # The claim directory was taken but its row never written — the degraded
    # branch of `claim`. The audit row is composed from the arguments instead.
    { printf 'lane: %s\npr: %s\nsha: %s\n' "$LANE" "$PR" "$SHA"
      printf 'state: released\n'
      printf 'owner: unrecorded — the claim row was never written\n'
      printf 'released_at: %s\n' "$NOW"
      printf 'release_reason: %s\n' "$reason"
    } > "$dst/row" 2>/dev/null || true
  fi
  printf 'release: done\nlane: %s\npanel: pr-%s @ %s\nreason: %s\n' "$LANE" "$PR" "$SHA" "$reason"
  printf 'note: the lane is claimable again\n'
  exit 0
}

# ---------------------------------------------------------------------------
# status — the trigger. The judge synthesises when the panel is complete, and
# this is how it learns that without an LLM and without receiving anything back.
# ---------------------------------------------------------------------------
cmd_status() {
  [ $# -ge 2 ] || usage_die 'status <pr> <sha> [lane]'
  PR="$(norm_pr "$1")"   || usage_die 'status <pr> <sha> — <pr> must be digits'
  SHA="$(norm_sha "$2")" || usage_die 'status <pr> <sha> — <sha> must be 7-40 hex'
  ONE=""
  if [ $# -ge 3 ]; then
    ONE="$(norm_lane "$3")" || usage_die 'status <pr> <sha> [lane] — <lane> must be [a-z0-9._-]'
  fi

  ensure_readable || { printf 'panel: pr-%s @ %s\nstatus: degraded — table unreadable\n' "$PR" "$SHA"; exit 2; }
  panel="$TABLE/pr-$PR/$SHA"

  printf 'panel: pr-%s @ %s\n' "$PR" "$SHA"
  if is_paused; then printf 'paused: true\n'; else printf 'paused: false\n'; fi

  total=0; ruled=0; outstanding=""
  for d in "$panel"/*/; do
    [ -d "$d" ] || continue
    lane="$(basename "$d")"
    [ -z "$ONE" ] || [ "$lane" = "$ONE" ] || continue
    total=$((total + 1))
    st="$(row_state "$d")"
    case "$st" in
      ruled)
        ruled=$((ruled + 1))
        printf '  %-30s %-18s %s (conf %s, findings %s) at %s\n' "$lane" "$st" \
          "$(field verdict "$d/verdict.d/row")" "$(field confidence "$d/verdict.d/row")" \
          "$(field findings "$d/verdict.d/row")" "$(field ruled_at "$d/verdict.d/row")"
        ;;
      claimed)
        outstanding="$outstanding $lane"
        printf '  %-30s %-18s owner %s since %s\n' "$lane" "$st" \
          "$(field owner "$d/holder/row")" "$(field claimed_at "$d/holder/row")"
        ;;
      *)
        # ruling-incomplete or free: neither ruled nor running. Counted as
        # outstanding, because the panel is not complete and a row that reads as
        # complete while holding no verdict is the corruption this replaces.
        outstanding="$outstanding $lane"
        printf '  %-30s %-18s no verdict on record\n' "$lane" "$st"
        ;;
    esac
  done

  if [ "$total" -eq 0 ]; then
    printf 'lanes: 0\ncomplete: false\nnote: no claim row here — nothing was dispatched under this panel\n'
    exit 12
  fi
  printf 'lanes: %s · ruled: %s · outstanding: %s\n' "$total" "$ruled" "$((total - ruled))"
  if [ -z "$outstanding" ]; then
    printf 'complete: true\n'
    exit 0
  fi
  printf 'outstanding:%s\n' "$outstanding"
  printf 'complete: false\n'
  exit 10
}

# ---------------------------------------------------------------------------
# list — every panel the table holds, or every panel of one PR.
# ---------------------------------------------------------------------------
cmd_list() {
  PR=""
  if [ $# -ge 1 ]; then
    PR="$(norm_pr "$1")" || usage_die 'list [pr]'
  fi
  ensure_readable || { printf 'list: degraded — table unreadable at %s\n' "$TABLE"; exit 2; }
  if is_paused; then printf 'paused: true\n'; else printf 'paused: false\n'; fi
  printf 'table: %s\n' "$TABLE"

  found=0
  for p in "$TABLE"/pr-*/; do
    [ -d "$p" ] || continue
    n="$(basename "$p")"; n="${n#pr-}"
    [ -z "$PR" ] || [ "$n" = "$PR" ] || continue
    for s in "$p"*/; do
      [ -d "$s" ] || continue
      sha="$(basename "$s")"
      t=0; r=0
      for d in "$s"*/; do
        [ -d "$d" ] || continue
        t=$((t + 1))
        [ "$(row_state "$d")" = ruled ] && r=$((r + 1))
      done
      [ "$t" -gt 0 ] || continue
      found=$((found + 1))
      if [ "$t" -eq "$r" ]; then c=complete; else c=incomplete; fi
      printf '  pr-%-6s %-42s lanes %-3s ruled %-3s %s\n' "$n" "$sha" "$t" "$r" "$c"
    done
  done
  [ "$found" -gt 0 ] || printf '  none\n'
  exit 0
}

# ---------------------------------------------------------------------------
# pause / resume / paused — the flag that stops the cascade at the next hop.
# ---------------------------------------------------------------------------
cmd_pause() {
  reason=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --reason) [ $# -ge 2 ] || usage_die 'pause [--reason <text>]'; reason="$(scrub "$2")"; shift 2 ;;
      *)        usage_die 'pause [--reason <text>]' ;;
    esac
  done
  ensure_writable || { printf 'pause: failed\nreason: degraded — table not writable\n'; exit 2; }
  if is_paused; then
    printf 'paused: true\nnote: already paused; the first pause stands\n'
    sed 's/^/  /' "$TABLE/PAUSED" 2>/dev/null
    exit 0
  fi
  { printf 'paused_at: %s\n' "$NOW"; printf 'reason: %s\n' "${reason:--}"; } > "$TABLE/PAUSED" 2>/dev/null || {
    printf 'pause: failed\nreason: degraded — flag not writable\n'
    exit 2
  }
  printf 'paused: true\npaused_at: %s\nreason: %s\n' "$NOW" "${reason:--}"
  printf 'note: no new claims; lanes already held may still record verdicts\n'
  exit 0
}

cmd_resume() {
  [ $# -eq 0 ] || usage_die 'resume'
  ensure_writable || { printf 'resume: failed\nreason: degraded — table not writable\n'; exit 2; }
  if ! is_paused; then printf 'paused: false\nnote: was not paused\n'; exit 0; fi
  rm -f "$TABLE/PAUSED" 2>/dev/null
  if is_paused; then
    printf 'resume: failed\nreason: degraded — flag could not be removed\n'
    exit 2
  fi
  printf 'paused: false\nresumed_at: %s\n' "$NOW"
  exit 0
}

cmd_paused() {
  [ $# -eq 0 ] || usage_die 'paused'
  ensure_readable || { printf 'paused: unknown\nreason: degraded — table unreadable\n'; exit 2; }
  if is_paused; then
    printf 'paused: true\n'
    sed 's/^/  /' "$TABLE/PAUSED" 2>/dev/null
    exit 11
  fi
  printf 'paused: false\n'
  exit 0
}

cmd_help() {
  cat <<'HELP'
claim.sh — the review claim table. A lane is claimed before dispatch; its verdict
is durable the instant it exists. No LLM, no network, no judgement.

  claim   <pr> <sha> <lane> [--owner <who>]
          Take the lane atomically. 0 granted — dispatch it. 10 held or already
          ruled — do NOT dispatch. 11 paused. 2 degraded, refused.

  verdict <pr> <sha> <lane> <ruling> [--conf <x>] [--findings <n>]
                                     [--artifact <ref>] [--note <text>]
          Record the ruling into the lane's own row, once. 0 recorded. 10 already
          ruled, the first stays of record. 12 recorded but no claim covered it.
          2 NOT recorded — carry it in your handoff. A pause does not block this.

  release <pr> <sha> <lane> --reason <text>
          Hand a stranded claim back so the lane can be re-dispatched. 0 done.
          10 the lane already ruled. 12 nothing was held. 2 degraded.

  status  <pr> <sha> [lane]
          The panel. 0 complete — synthesise. 10 lanes outstanding. 12 no rows.
          2 degraded.

  list    [pr]                Every panel, or one PR's. 0 always, 2 degraded.
  pause   [--reason <text>]   No new claims. Held lanes still record. 0.
  resume                      Lift the pause. 0.
  paused                      0 not paused · 11 paused · 2 degraded.
  help                        This.

A <sha> is 7-40 hex and the panel key is its first SEVEN characters, so one
commit is one panel however it is spelled — `ec9ee33` and `ec9ee330b2…` are the
same row and not two. A <pr> sheds leading zeros for the same reason. Both are
echoed back canonical, so the reply names the panel you actually joined.

Exit codes: 0 OK · 10 REFUSED · 11 PAUSED · 12 NO_ROW · 2 DEGRADED · 64 USAGE.
Every ambiguous state refuses. State lives under .claude/state/review-claims/;
CLAIM_TABLE_DIR overrides the location.
HELP
  exit 0
}

sub="${1:-help}"
[ $# -eq 0 ] || shift
case "$sub" in
  claim)   cmd_claim   "$@" ;;
  verdict) cmd_verdict "$@" ;;
  release) cmd_release "$@" ;;
  status)  cmd_status  "$@" ;;
  list)    cmd_list    "$@" ;;
  pause)   cmd_pause   "$@" ;;
  resume)  cmd_resume  "$@" ;;
  paused)  cmd_paused  "$@" ;;
  help|-h|--help) cmd_help ;;
  *) printf 'unknown subcommand: %s\n' "$sub" >&2; usage_die '<claim|verdict|release|status|list|pause|resume|paused|help>' ;;
esac
