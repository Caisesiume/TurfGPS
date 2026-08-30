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
#   12  NO_ROW    nothing was claimed here — no such panel, lane or manifest; or
#                 the claim does not cover the verdict that arrived (recorded
#                 anyway, flagged): no claim at all, or a filer who is not the
#                 one the claim recorded as expected.
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
# THE ATOMIC PRIMITIVES
#   A claim is `mkdir` of one directory. The kernel decides the winner, not a
#   read followed by a write, so two concurrent claimants on one lane at one SHA
#   produce exactly one 0 and one 10. There is no window between the test and the
#   set because there is no test. This holds on Git Bash for Windows, where the
#   call lands on CreateDirectory, which fails when the name exists.
#
#   A RECORD — a verdict, a manifest — is committed by renaming the directory
#   that already holds it onto the name readers test for, so the name cannot
#   exist without its payload. Same kernel decision, same absence of a window,
#   and additionally no half-written state for an interruption to leave behind.
#   `commit_staged` below is the one place that primitive lives.
#
# STATE
#   .claude/state/review-claims/ — inside the `/.claude/state/` entry .gitignore
#   already carries, the established home for durable local control-plane state,
#   alongside fingerprint.sh's per-consumer files. It is MACHINE-local and not
#   checkout-local: every worktree of one repository resolves the same table, by
#   the derivation argued at `table_root` below. CLAIM_TABLE_DIR overrides the
#   location, which is how a test runs against a throwaway root rather than
#   against the machine's real table. Every verb echoes the table it resolved.
#
#     <table>/PAUSED                                 flag: no new claims while it exists
#     <table>/pr-<n>/<sha>/.manifest.d/row           the selected lane set, written once
#     <table>/pr-<n>/<sha>/<lane>/holder/row         the claim
#     <table>/pr-<n>/<sha>/<lane>/verdict.d/row      the verdict, written once, immutable
#     <table>/pr-<n>/<sha>/<lane>/released-<stamp>/row   a claim given back, kept as audit
#     <table>/pr-<n>/<sha>/<lane>/released-verdict-<stamp>/row  an interrupted ruling, cleared
#
#   This script never reads GH_JUDGE_TOKEN, never invokes gh, and never records
#   the environment. That is the guarantee, it is structural, and it is the only
#   one made here.
#
#   The scrub over free text is a BACKSTOP AND NOT A SECOND GUARANTEE, and this
#   header used to claim otherwise: it said a caller who pastes a credential
#   into --artifact "cannot put it into the table either". Measured on
#   2026-08-29, fourteen credential shapes passed through it unredacted — AWS
#   access keys, `sk-ant-api03-…`, `xoxb-…`, `glpat-…`, PEM headers,
#   `user:password@` URLs, `Bearer eyJ…`, and `GHP_` in upper case, `sed -E`
#   being case-sensitive. What it actually removes is lower-case GitHub token
#   prefixes — `ghp_ gho_ ghu_ ghs_ ghr_` and `github_pat_` — and nothing else.
#
#   The defect was the sentence rather than the net, and the sentence is the
#   dangerous half: a header promising a guarantee the regex cannot deliver is
#   how a caller comes to believe the table is safe to paste into. It is not.
#   Do not paste a credential into a field; the scrub catches one shape of one
#   mistake.
#
# Usage: scripts/loop/claim.sh <subcommand> [args…]   —  `claim.sh help` prints the surface

set -u
LC_ALL=C
export LC_ALL

# THE TABLE IS MACHINE-LOCAL, NOT CHECKOUT-LOCAL
#   A mutual-exclusion table two checkouts of one repository cannot both see is
#   not a mutual-exclusion table. `--show-toplevel` answers with the caller's
#   own worktree and a linked worktree is its own toplevel, so a judge in the
#   main checkout and a reviewer dispatched into `../TurfGPS-wt/<slug>` built
#   two panels for one PR at one SHA that could not see each other: the judge
#   granted, the reviewer recorded `unclaimed` into the other table at exit 12,
#   and the judge's panel read outstanding at 10 forever. That is CLAIM-01's
#   exact shape reopened through the PATH instead of the key, and #144's failure
#   classes 1 and 2 with it.
#
#   `--git-common-dir` is the one thing git answers identically from every
#   worktree of a repository. Measured on 2026-08-29 on this host from four
#   cwds — main toplevel, a main subdir, the linked worktree, a linked-worktree
#   subdir — it returned `.git`, `../../.git`, and twice an absolute path: four
#   spellings of ONE directory. It is therefore resolved to absolute here rather
#   than used as given, because two of those four are relative to the caller's
#   cwd, which is the very thing being eliminated.
#
#   The table sits BESIDE that directory rather than inside it, which puts it at
#   the same path the old line produced from the main checkout — so no existing
#   table moves and nothing is migrated; only the worktrees that used to diverge
#   now agree. Where a repository keeps its git directory elsewhere the location
#   moves with it and all of its worktrees still agree, which is the property
#   being bought.
#
#   fingerprint.sh next door keeps `--show-toplevel` and is right to: its state
#   is a fact ABOUT one tree, so tree-local is its correct scope. A claim table
#   is not a fact about a tree. Same directory, different question — which is
#   why this file resolves its own root instead of copying the neighbour's line.
table_root() {
  _gcd="$(git rev-parse --git-common-dir 2>/dev/null)" || return 1
  [ -n "$_gcd" ] || return 1
  _gcd="$(cd "$_gcd" 2>/dev/null && pwd)" || return 1
  [ -n "$_gcd" ] || return 1
  _par="${_gcd%/*}"
  [ -n "$_par" ] || _par="/"
  printf '%s' "$_par"
}
ROOT="$(table_root || pwd)"
TABLE="${CLAIM_TABLE_DIR:-$ROOT/.claude/state/review-claims}"

# Every verb opens by naming the table it acted on. A split table is the failure
# the resolution above exists to prevent, and the cheapest moment to see one is
# the FIRST LINE of the FIRST reply — not at synthesis, where two panels are
# already built and the only evidence of the split is that neither is complete.
say_table() { printf 'table: %s\n' "$TABLE"; }
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)"
[ -n "$NOW" ] || NOW="unknown-time"
# A second stamp, for the one place a timestamp becomes a FILENAME. The ISO form
# above carries `:`, which Win32 reads as the alternate-data-stream separator and
# will not accept in a name. What that costs is NOT an `mv` that fails: measured
# on 2026-08-29 on this host, `mv` SUCCEEDS and this shell reads the row back —
# Cygwin stores the colon as U+F03A, a private-use codepoint, so the name that
# lands on disk is not the name that was asked for. PowerShell renders that
# directory `12?32?59Z`; a native open of the real `…12:32:59Z…` path raises
# ItemNotFoundException, and a native mkdir of it raises NotSupportedException.
# So the guard stays, for the reason it was always right: an audit row
# addressable only from the shell that wrote it is not an audit row. The
# mechanism is spelled out because this comment previously blamed a failing
# `mv`, and a false mechanism in a comment is how a live guard gets deleted.
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

# ONE AGENT IS ONE NAME, HOWEVER IT IS SPELLED
#   `docs-reviewer` and `@Docs-Reviewer` are the two ordinary ways an agent
#   name is written down, and dispatches use the second: `norm_lane` above
#   strips the `@` and folds the case for exactly that reason. A name arriving
#   through `--by`, `--for` or `--owner` gets no such fold — those are free
#   text and may honestly read `unnamed` — so a filer spelled `@docs-reviewer`
#   against a lane-derived `expects: docs-reviewer` would differ as strings and
#   agree as identities. That is the false anomaly `cmd_claim` below exists to
#   remove, walking back in through the spelling instead of through the field,
#   and it is the same shape as the two SHA spellings that once split a panel.
#
#   The fold is applied to the COMPARISON and never to what is stored. Each row
#   keeps the name its caller passed, because an audit row that quietly rewrites
#   what it was told is not evidence of what it was told.
agent_key() {
  printf '%s' "${1:-}" | tr -d '@' | tr 'A-Z' 'a-z'
}

# One line, no control characters, capped, and with lower-case GitHub token
# prefixes removed. What the token pattern does and does not cover is argued in
# the header — do not restate it here as a guarantee. The pattern is matched, not
# compared against a value: this script does not read GH_JUDGE_TOKEN, so it
# cannot leak it by accident of comparison either.
#
# The control strip is `[:cntrl:]` and not `\r\n`. CR and LF are what break the
# one-line record format, so they were all that got removed — and ANSI escapes
# were therefore stored raw, `od`-confirmed. A stored escape survives the round
# trip and repaints the terminal of the human reading the panel, so the row is
# intact and the reader's view of it is not, which is the one attack a ledger
# whose whole purpose is being believed cannot afford. Under the `LC_ALL=C` set
# at the top of this file the class is the ASCII control range and nothing else;
# ordinary text, including UTF-8, is untouched.
scrub() {
  printf '%s' "${1:-}" \
    | tr -d '[:cntrl:]' \
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

# THE SECOND ATOMIC PRIMITIVE: a record is committed by renaming the DIRECTORY
# that already holds it onto the name readers look for. The name therefore never
# exists without its payload, so no reader can observe a half-written record and
# no interruption can leave one — there is nothing between the two states.
#
# `-T` is load-bearing and its absence is silent. Measured on this host on
# 2026-08-29, plain `mv stage dest` with `dest` an existing directory does not
# fail: it moves the source INSIDE, producing `dest/stage/row`, which reads as a
# ruled lane holding no verdict. `mv -T` renames, and refuses with "Directory
# not empty" when the destination already holds a record — which is exactly the
# mutual exclusion wanted, decided by the kernel rather than by a test-then-set.
#
# It succeeds onto an EMPTY destination, and that is deliberate rather than
# tolerated: an empty `verdict.d` is the wreckage this primitive exists to stop
# creating, so a table carrying one from before this change is repaired by the
# next verdict instead of staying wedged.
commit_staged() { # commit_staged <stage-dir> <destination>
  mv -T "$1" "$2" 2>/dev/null
}

field() { # field <key> <file>
  [ -r "$2" ] || return 0
  grep "^$1: " "$2" 2>/dev/null | head -1 | cut -d' ' -f2-
}

# claimed · ruled · ruling-incomplete · free. `ruling-incomplete` is a verdict
# directory holding no READABLE row, and it counts as outstanding — visible
# debt, which is the whole point.
#
# No path in this file can now produce one. A ruling is committed by renaming a
# directory that already holds its row, so `verdict.d` never exists empty; the
# state remains reachable only from a table written before that change, from an
# unreadable row, or from a hand-edit. It is kept because those are real, and
# because a reader that cannot name the state it is looking at will call it
# something else — but it is no longer a state an interruption can create, and
# `release` now clears it rather than calling it finished.
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
#
# TWO IDENTITIES, BECAUSE THE HOLDER IS NEVER THE FILER ON THE ORDINARY PATH
#   A claim used to record one name, and the verdict then checked its filer
#   against it. That check fired on every honest verdict this table will ever
#   see: `pr-judge` Phase 4 claims each selected lane ON THE REVIEWER'S BEHALF
#   before dispatching it, so the judge is always the holder and the reviewer
#   is always the filer. Measured on 2026-08-30: a lane claimed by pr-judge
#   and ruled by its own reviewer recorded `attribution_mismatch: true` at
#   exit 12 — the normal case reported as the anomaly, which is how a signal
#   becomes noise and then gets switched off.
#
#   The holder cannot become the reviewer to fix it. `engineering-lead`'s
#   courier check reads `holder:` and couriers only when it is the judge that
#   asked, so a claim naming the reviewer would break the one check standing
#   between a panel and a lane carried out from under its owner. And the flag
#   cannot be dropped: a verdict written by anyone at all, attributed of
#   record to the claimant, is #140's ledger-corruption class.
#
#   So the row carries BOTH: `owner:` is who holds the claim, `expects:` is
#   who the lane's verdict is expected to come from. `verdict` checks its
#   `--by` against the SECOND. The two questions were always different and
#   one field could only ever answer them both wrong.
# ---------------------------------------------------------------------------
cmd_claim() {
  [ $# -ge 3 ] || usage_die 'claim <pr> <sha> <lane> [--owner <who>] [--for <who>]'
  PR="$(norm_pr "$1")"     || usage_die 'claim <pr> <sha> <lane> — <pr> must be digits'
  SHA="$(norm_sha "$2")"   || usage_die 'claim <pr> <sha> <lane> — <sha> must be 7-40 hex'
  LANE="$(norm_lane "$3")" || usage_die 'claim <pr> <sha> <lane> — <lane> must be [a-z0-9._-]'
  shift 3
  owner="unnamed"; expects=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --owner) [ $# -ge 2 ] || usage_die 'claim … --owner <who>'; owner="$(scrub "$2")"; shift 2 ;;
      --for)   [ $# -ge 2 ] || usage_die 'claim … --for <who>';   expects="$(scrub "$2")"; shift 2 ;;
      *)       usage_die 'claim <pr> <sha> <lane> [--owner <who>] [--for <who>]' ;;
    esac
  done
  [ -n "$owner" ] || owner="unnamed"
  # The default is the LANE NAME, and it is a default rather than a required
  # flag on purpose: a lane is named for the reviewer that rules it —
  # `norm_lane` strips a leading `@` precisely because dispatches name lanes
  # as `@agent-name`, and the case file hands a reviewer the lane name its
  # verdict is recorded under. So every call site already in the corpus
  # records the right expectation without changing a character, and `--for`
  # exists only for a lane dispatched to a name other than its own.
  [ -n "$expects" ] || expects="$LANE"
  say_table

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
      printf 'expects: %s\n'    "$expects"
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
    printf 'claim: granted\nlane: %s\npanel: pr-%s @ %s\nowner: %s\nexpects: %s\nclaimed_at: %s\n' \
      "$LANE" "$PR" "$SHA" "$owner" "$expects" "$NOW"
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
  [ $# -ge 4 ] || usage_die 'verdict <pr> <sha> <lane> <verdict> [--by <who>] [--conf <x>] [--findings <n>] [--artifact <ref>]'
  PR="$(norm_pr "$1")"     || usage_die 'verdict <pr> … — <pr> must be digits'
  SHA="$(norm_sha "$2")"   || usage_die 'verdict <pr> <sha> … — <sha> must be 7-40 hex'
  LANE="$(norm_lane "$3")" || usage_die 'verdict <pr> <sha> <lane> … — <lane> must be [a-z0-9._-]'
  V="$(scrub "$4")"
  [ -n "$V" ] || usage_die 'verdict <pr> <sha> <lane> <verdict> — <verdict> may not be empty'
  shift 4
  conf="-"; findings="-"; artifact="-"; by=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --by)                [ $# -ge 2 ] || usage_die 'verdict … --by <who>';     by="$(scrub "$2")";       shift 2 ;;
      --conf|--confidence) [ $# -ge 2 ] || usage_die 'verdict … --conf <x>';     conf="$(scrub "$2")";     shift 2 ;;
      --findings)          [ $# -ge 2 ] || usage_die 'verdict … --findings <n>'; findings="$(scrub "$2")"; shift 2 ;;
      --artifact)          [ $# -ge 2 ] || usage_die 'verdict … --artifact <ref>'; artifact="$(scrub "$2")"; shift 2 ;;
      *) usage_die 'verdict <pr> <sha> <lane> <verdict> [--by <who>] [--conf <x>] [--findings <n>] [--artifact <ref>]' ;;
    esac
  done
  say_table

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

  # THE GATE, AND IT IS THE RENAME RATHER THAN A MKDIR.
  #   The row is written inside a private staging directory and the whole
  #   directory is committed onto `verdict.d`, so the name every reader tests
  #   for comes into existence WITH its payload already inside it.
  #
  #   A `mkdir` gate followed by a write is two steps, and a process that died
  #   between them — #144's failure class 1, verbatim — left `verdict.d` present
  #   and empty. That state absorbed the lane: `claim` refused it as ruled,
  #   `verdict` refused it as already ruled, `status` counted it outstanding
  #   forever, and `release` refused it saying a lane that has ruled "is
  #   finished, not stranded", which asserted the opposite of the truth. Four
  #   verbs refusing forever, and the only recovery was `rm -rf` of the ledger:
  #   an unaudited hand-edit of the artefact whose integrity is the point. One
  #   rename has no between, so the state has no way to arise.
  #
  #   The mutual exclusion is unchanged and still the kernel's: one ruling per
  #   lane per SHA. Two judges ruled on #141 sixteen seconds apart with
  #   contradictory packets; under this gate the second is refused and the first
  #   stays of record. Six concurrent verdicts were measured against the mkdir
  #   gate as exactly one `recorded` and five refusals, and `commit_staged`
  #   above records why the rename holds the same line.
  stage="$row/.verdict-staging.$$"
  rm -rf "$stage" 2>/dev/null
  mkdir "$stage" 2>/dev/null || {
    printf 'verdict: NOT RECORDED\nreason: degraded — could not stage the ruling\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    printf 'direction: carry this verdict in your handoff; the table does not hold it\n'
    exit 2
  }
  unclaimed=false
  [ -d "$row/holder" ] || unclaimed=true
  holder_owner="$(field owner "$row/holder/row")"
  holder_expects="$(field expects "$row/holder/row")"
  # A claim row written before `expects:` existed carries none, and so does the
  # degraded claim whose directory was taken and whose row was never written.
  # The fallback is the SAME default `claim` records, so an older table judges
  # attribution by one rule rather than flagging every honest verdict in it.
  [ -n "$holder_expects" ] || holder_expects="$LANE"

  # WHO FILED IT, AGAINST WHO THE LANE EXPECTED IT FROM.
  #   `owner:` is copied out of the claim, so before `--by` existed a verdict
  #   written by anyone at all was attributed of record to the claimant, and the
  #   writer's identity was not merely unrecorded but unrecordable — there was
  #   no flag to carry it. A courier filing into a lane held by someone else
  #   produced a row reading `verdict / owner: <the reviewer that never ran>`,
  #   exit 0, panel complete, synthesise. `unclaimed` never fired, because it
  #   only fires when NO holder exists — so the likely case was the silent one.
  #
  #   THE COMPARISON IS AGAINST `expects:` AND NOT AGAINST `owner:`. Checking
  #   the filer against the holder flagged the ordinary path — the judge holds
  #   every lane it claims on a reviewer's behalf, so holder and filer differ
  #   on every honest verdict, and a flag that fires on all of them detects
  #   nothing. `cmd_claim` above argues why both names are recorded and why
  #   neither may collapse into the other.
  #
  #   Four values, each meaning exactly one thing, because conflating them is
  #   how the field stops being evidence:
  #     false        checked, and the filer is the one the claim expected,
  #                  compared by `agent_key` so a spelling cannot fake either
  #                  answer
  #     true         checked, and they differ — the anomaly, reported at 12
  #     unrecorded   no `--by` was passed, so nothing could be checked
  #     no-holder    there is no claim row to check against; see `unclaimed`
  #   `unrecorded` is not a pass. It is the row saying its own writer is unknown,
  #   which is strictly more than the old row said and is the audit signal when
  #   a caller has not been taught to pass `--by` yet; it exits 0, because
  #   refusing a verdict over a missing flag loses the work this file exists
  #   to keep.
  mismatch=unrecorded
  if [ "$unclaimed" = true ]; then
    mismatch=no-holder
  elif [ -n "$by" ]; then
    if [ "$(agent_key "$by")" = "$(agent_key "$holder_expects")" ]; then
      mismatch=false
    else
      mismatch=true
    fi
  fi
  {
    printf 'lane: %s\n'     "$LANE"
    printf 'pr: %s\n'       "$PR"
    printf 'sha: %s\n'      "$SHA"
    printf 'verdict: %s\n'  "$V"
    printf 'confidence: %s\n' "$conf"
    printf 'findings: %s\n' "$findings"
    printf 'artifact: %s\n' "$artifact"
    printf 'owner: %s\n'    "${holder_owner:--}"
    printf 'expects: %s\n'  "$holder_expects"
    printf 'filed_by: %s\n' "${by:-unrecorded}"
    printf 'attribution_mismatch: %s\n' "$mismatch"
    printf 'unclaimed: %s\n' "$unclaimed"
    printf 'ruled_at: %s\n' "$NOW"
  } > "$stage/row" 2>/dev/null || {
    rm -rf "$stage" 2>/dev/null
    printf 'verdict: NOT RECORDED\nreason: degraded — row not writable\n'
    printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
    printf 'direction: carry this verdict in your handoff; the table does not hold it\n'
    exit 2
  }

  if commit_staged "$stage" "$row/verdict.d"; then
    printf 'verdict: recorded\nlane: %s\npanel: pr-%s @ %s\nruling: %s\nruled_at: %s\n' \
      "$LANE" "$PR" "$SHA" "$V" "$NOW"
    printf 'filed_by: %s\n' "${by:-unrecorded}"
    # Durable first, loud second. Refusing either of these would lose a verdict
    # to enforce a rule about dispatch, which is the wrong trade; both anomalies
    # are recorded in the row and reported nonzero so the caller cannot miss
    # them. Both are 12 and neither is new: 12 has always meant "the claim does
    # not cover this verdict — recorded anyway, flagged". They cannot co-occur,
    # a mismatch requiring the holder whose absence is the other.
    if [ "$unclaimed" = true ]; then
      printf 'anomaly: no claim row covered this lane — the verdict is stored and flagged `unclaimed`\n'
      exit 12
    fi
    if [ "$mismatch" = true ]; then
      printf 'anomaly: filed by %s, but this lane expects its verdict from %s (held by %s) — the verdict is stored and flagged `attribution_mismatch`\n' \
        "$by" "$holder_expects" "${holder_owner:--}"
      exit 12
    fi
    exit 0
  fi

  # The commit refused, so the staged ruling is not of record and must not be
  # left behind looking like one.
  rm -rf "$stage" 2>/dev/null
  printf 'verdict: refused\nreason: already ruled\n'
  printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
  if [ -r "$row/verdict.d/row" ]; then
    printf 'of_record: %s (conf %s, findings %s) at %s\n' \
      "$(field verdict "$row/verdict.d/row")" "$(field confidence "$row/verdict.d/row")" \
      "$(field findings "$row/verdict.d/row")" "$(field ruled_at "$row/verdict.d/row")"
  else
    printf 'of_record: unreadable — a ruling is in flight or was interrupted\n'
    printf 'direction: carry this verdict in your handoff; `claim.sh release` clears the lane\n'
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
  say_table

  ensure_writable || { printf 'release: refused\nreason: degraded — table not writable\n'; exit 2; }
  row="$TABLE/pr-$PR/$SHA/$LANE"

  # A lane that has RULED is finished, not stranded, and release refuses it. A
  # verdict directory holding no readable row has NOT ruled — it is the wreckage
  # of a ruling interrupted mid-write, and it is precisely the lane `pr-judge`
  # Phase 10 sends here. This test used to be on the DIRECTORY, so release
  # answered "a lane that has ruled is finished" about a lane that had not ruled
  # and could not be recovered by any verb. The distinction is now the ROW, so
  # the instruction that names release as the remedy is true of the state it
  # names. The wreckage is moved aside as an audit row and never deleted: it is
  # evidence that a reviewer reached a ruling, and it is the one trace of it.
  cleared_ruling=false
  if [ -d "$row/verdict.d" ]; then
    if [ -r "$row/verdict.d/row" ]; then
      printf 'release: refused\nreason: ruled — a lane that has ruled is finished, not stranded\n'
      printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
      printf 'of_record: %s\n' "$(field verdict "$row/verdict.d/row")"
      exit 10
    fi
    vdst="$row/released-verdict-$STAMP-$$"
    mv "$row/verdict.d" "$vdst" 2>/dev/null || {
      printf 'release: refused\nreason: degraded — could not move the incomplete ruling aside\n'
      printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
      exit 2
    }
    # The readability test above is a read before a write, and the write it
    # guards is the only one in this file that can move a RULING OF RECORD. A
    # verdict committing in the same instant would land its payload into the
    # directory between the two, and this call would carry a real ruling off to
    # an audit row that `status` does not read. So the question is asked again
    # after the move, when the rename has already serialised the two: whatever
    # is in hand now is what was actually taken. A ruling goes back and the
    # release refuses — the direction that never discards a verdict — and if it
    # cannot go back the call says loudly where it left it rather than reporting
    # a recovery that lost one.
    if [ -r "$vdst/row" ]; then
      if commit_staged "$vdst" "$row/verdict.d"; then
        printf 'release: refused\nreason: ruled — a ruling landed while this call was clearing the lane\n'
        printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
        printf 'of_record: %s\n' "$(field verdict "$row/verdict.d/row")"
        exit 10
      fi
      printf 'release: refused\nreason: degraded — a ruling landed mid-release and could not be put back\n'
      printf 'lane: %s\npanel: pr-%s @ %s\n' "$LANE" "$PR" "$SHA"
      printf 'ruling_is_at: %s\n' "$vdst"
      printf 'direction: the verdict is on disk and is NOT of record; do not re-dispatch this lane\n'
      exit 2
    fi
    { printf 'lane: %s\npr: %s\nsha: %s\n' "$LANE" "$PR" "$SHA"
      printf 'state: released-ruling-incomplete\n'
      printf 'note: a verdict directory with no readable row — a ruling interrupted mid-write\n'
      printf 'released_at: %s\n' "$NOW"
      printf 'release_reason: %s\n' "$reason"
    } > "$vdst/row" 2>/dev/null || true
    cleared_ruling=true
  fi
  if [ ! -d "$row/holder" ]; then
    if [ "$cleared_ruling" = true ]; then
      # The incomplete ruling was the whole of the strand: there is no claim
      # left to hand back, and the recovery still happened. Reporting 12 here
      # would say nothing was done about a lane this call just unwedged.
      printf 'release: done\nlane: %s\npanel: pr-%s @ %s\nreason: %s\n' "$LANE" "$PR" "$SHA" "$reason"
      printf 'cleared: ruling-incomplete — the interrupted verdict was moved aside as an audit row\n'
      printf 'note: the lane is claimable again\n'
      exit 0
    fi
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
  [ "$cleared_ruling" = false ] || \
    printf 'cleared: ruling-incomplete — the interrupted verdict was moved aside as an audit row\n'
  printf 'note: the lane is claimable again\n'
  exit 0
}

# ---------------------------------------------------------------------------
# manifest — the set the panel must be complete AGAINST.
#
# Without one, `complete` is a claim about the rows that happen to exist, which
# is not a claim about coverage at all. A judge that claimed 2 of 7 selected
# lanes and died mid-selection left a panel reading `complete: true` at rc 0 the
# moment those two ruled; the next judge, following "0 outstanding, synthesise",
# publishes a two-lane ledger for a seven-lane board — and Phase 10 makes that
# table of record. That is #144's failure class 4, a ledger under-reporting
# lanes while asserting coverage, reproduced by the mechanism built to close it.
#
# So the selection is recorded as a row like any other, once, by the same atomic
# directory commit the verdict uses, and `status` counts a selected lane with no
# row as outstanding. A panel cannot then read complete while a lane it selected
# was never claimed. `complete` means complete against something.
#
# A panel with NO manifest behaves exactly as it did before: the table does not
# invent a set nobody recorded, and it will not start refusing panels written by
# a caller that has not been taught to select yet.
#
# It lives at `.manifest.d/`, a leading dot, which no lane can ever be called:
# `norm_lane` refuses a name beginning with `.`, and the lane globs do not match
# one either. The name is unreachable rather than merely unused — the same
# reasoning that keeps `PAUSED` upper-case, one level up.
# ---------------------------------------------------------------------------
cmd_manifest() {
  [ $# -ge 2 ] || usage_die 'manifest <pr> <sha> [--lanes "<lane> …"] [--by <who>]'
  PR="$(norm_pr "$1")"   || usage_die 'manifest <pr> <sha> — <pr> must be digits'
  SHA="$(norm_sha "$2")" || usage_die 'manifest <pr> <sha> — <sha> must be 7-40 hex'
  shift 2
  lanes=""; by=""; writing=false
  while [ $# -gt 0 ]; do
    case "$1" in
      --lanes) [ $# -ge 2 ] || usage_die 'manifest … --lanes "<lane> …"'; lanes="$2"; writing=true; shift 2 ;;
      --by)    [ $# -ge 2 ] || usage_die 'manifest … --by <who>'; by="$(scrub "$2")"; shift 2 ;;
      *)       usage_die 'manifest <pr> <sha> [--lanes "<lane> …"] [--by <who>]' ;;
    esac
  done
  say_table
  panel="$TABLE/pr-$PR/$SHA"

  if [ "$writing" = false ]; then
    ensure_readable || { printf 'manifest: unknown\nreason: degraded — table unreadable\n'; exit 2; }
    printf 'panel: pr-%s @ %s\n' "$PR" "$SHA"
    if [ -r "$panel/.manifest.d/row" ]; then
      printf 'manifest: %s\n'    "$(field lanes "$panel/.manifest.d/row")"
      printf 'count: %s\n'       "$(field count "$panel/.manifest.d/row")"
      printf 'selected_by: %s\n' "$(field selected_by "$panel/.manifest.d/row")"
      printf 'selected_at: %s\n' "$(field selected_at "$panel/.manifest.d/row")"
      exit 0
    fi
    printf 'manifest: none\n'
    printf 'note: this panel records no selected set; `complete` counts only the rows that exist\n'
    exit 12
  fi

  # Every name is canonicalised through the same gate a claim uses, so the
  # manifest and the rows agree on what a lane is called. One bad name refuses
  # the whole set at 64 and writes nothing: a manifest that silently dropped a
  # lane would be a set asserting coverage it does not have, which is the exact
  # defect this verb exists to close. The split here is deliberate word
  # splitting — the one place in this file that wants it.
  sel=""; nsel=0
  for l in $lanes; do
    nl="$(norm_lane "$l")" || usage_die 'manifest … --lanes — each lane must be [a-z0-9._-]'
    sel="$sel $nl"; nsel=$((nsel + 1))
  done
  [ "$nsel" -gt 0 ] || usage_die 'manifest … --lanes "<lane> …" — the set may not be empty'
  sel="${sel# }"

  ensure_writable || {
    printf 'manifest: NOT RECORDED\nreason: degraded — table not writable\n'
    printf 'direction: the panel has no expected set; do not treat `complete` as coverage\n'
    exit 2
  }
  mkdir -p "$panel" 2>/dev/null || {
    printf 'manifest: NOT RECORDED\nreason: degraded — could not create panel %s\n' "$panel"
    exit 2
  }
  stage="$panel/.manifest-staging.$$"
  rm -rf "$stage" 2>/dev/null
  mkdir "$stage" 2>/dev/null || {
    printf 'manifest: NOT RECORDED\nreason: degraded — could not stage the selection\n'
    exit 2
  }
  { printf 'pr: %s\nsha: %s\n' "$PR" "$SHA"
    printf 'lanes: %s\n' "$sel"
    printf 'count: %s\n' "$nsel"
    printf 'selected_by: %s\n' "${by:--}"
    printf 'selected_at: %s\n' "$NOW"
  } > "$stage/row" 2>/dev/null || {
    rm -rf "$stage" 2>/dev/null
    printf 'manifest: NOT RECORDED\nreason: degraded — row not writable\n'
    exit 2
  }

  if commit_staged "$stage" "$panel/.manifest.d"; then
    printf 'manifest: recorded\npanel: pr-%s @ %s\ncount: %s\nlanes: %s\n' "$PR" "$SHA" "$nsel" "$sel"
    exit 0
  fi
  # Written once, like a verdict, and for the same reason: a selection that can
  # be rewritten after the fact is a selection that can be shrunk to fit what
  # actually ruled.
  rm -rf "$stage" 2>/dev/null
  printf 'manifest: refused\nreason: already recorded — the first selection stays of record\n'
  printf 'panel: pr-%s @ %s\n' "$PR" "$SHA"
  printf 'of_record: %s\n' "$(field lanes "$panel/.manifest.d/row")"
  exit 10
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

  say_table
  ensure_readable || { printf 'panel: pr-%s @ %s\nstatus: degraded — table unreadable\n' "$PR" "$SHA"; exit 2; }
  panel="$TABLE/pr-$PR/$SHA"

  printf 'panel: pr-%s @ %s\n' "$PR" "$SHA"
  if is_paused; then printf 'paused: true\n'; else printf 'paused: false\n'; fi
  mlanes=""
  [ -r "$panel/.manifest.d/row" ] && mlanes="$(field lanes "$panel/.manifest.d/row")"
  if [ -n "$mlanes" ]; then printf 'manifest: %s\n' "$mlanes"; else printf 'manifest: none\n'; fi

  total=0; ruled=0; outstanding=""; one_state=""
  for d in "$panel"/*/; do
    [ -d "$d" ] || continue
    lane="$(basename "$d")"
    # A dot-name is not a lane — norm_lane refuses one, so anything the glob
    # turns up beginning with `.` is this file's own bookkeeping and never a row.
    case "$lane" in .*) continue ;; esac
    [ -z "$ONE" ] || [ "$lane" = "$ONE" ] || continue
    total=$((total + 1))
    st="$(row_state "$d")"
    [ -z "$ONE" ] || one_state="$st"
    case "$st" in
      ruled)
        ruled=$((ruled + 1))
        printf '  %-30s %-18s %s (conf %s, findings %s) at %s\n' "$lane" "$st" \
          "$(field verdict "$d/verdict.d/row")" "$(field confidence "$d/verdict.d/row")" \
          "$(field findings "$d/verdict.d/row")" "$(field ruled_at "$d/verdict.d/row")"
        # The verdict's own route to its findings. Without this the judge reads
        # "revise, conf 0.9, 7 findings" and the mechanism offers it no way to
        # reach the 7 — on the dead-parent path this table exists for, that is
        # the whole of what it was supposed to carry.
        art="$(field artifact "$d/verdict.d/row")"
        [ -z "$art" ] || [ "$art" = "-" ] || printf '      artifact: %s\n' "$art"
        # A durable flag no read verb surfaces is a flag nobody acts on.
        [ "$(field attribution_mismatch "$d/verdict.d/row")" != true ] || \
          printf '      attribution_mismatch: true — filed_by %s, expected %s (holder %s)\n' \
            "$(field filed_by "$d/verdict.d/row")" "$(field expects "$d/verdict.d/row")" \
            "$(field owner "$d/verdict.d/row")"
        [ "$(field unclaimed "$d/verdict.d/row")" != true ] || \
          printf '      unclaimed: true — no claim row covered this lane\n'
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

  # Selected and never claimed. This is the half `complete` was missing: without
  # it the panel is complete against whatever happened to get claimed, so a
  # judge that died part-way through selecting left a two-lane panel reading
  # complete for a seven-lane board. See `cmd_manifest` above.
  for l in $mlanes; do
    [ -z "$ONE" ] || [ "$l" = "$ONE" ] || continue
    [ -d "$panel/$l" ] && continue
    total=$((total + 1))
    outstanding="$outstanding $l"
    [ -z "$ONE" ] || one_state="never-claimed"
    printf '  %-30s %-18s selected by the manifest, no row exists\n' "$l" "never-claimed"
  done

  # The machine-readable answer for a single lane. `status <lane>` returns 10
  # for held, for free and for ruling-incomplete alike — three states with three
  # different remedies — and callers are told to branch on the exit status and
  # never on the prose, which left them nothing to branch on. This field is a
  # closed vocabulary of six tokens: ruled · claimed · ruling-incomplete · free
  # · never-claimed · no-row — the last being a lane this panel has never heard
  # of, which the line below returns at 12. It is a record field in the format
  # every row in this table already uses, not a sentence, and `field lane_state`
  # reads it.
  if [ -n "$ONE" ]; then
    printf 'lane: %s\n' "$ONE"
    printf 'lane_state: %s\n' "${one_state:-no-row}"
    if [ "$one_state" = claimed ]; then
      printf 'holder: %s\n' "$(field owner "$panel/$ONE/holder/row")"
      # The second identity, surfaced beside the first. A courier reads
      # `holder:` to check the judge that asked it; who the lane expects its
      # verdict from is the other half of the same question, and a field no
      # read verb surfaces is a field nobody can act on.
      printf 'expects: %s\n' "$(field expects "$panel/$ONE/holder/row")"
    elif [ "$one_state" = ruled ]; then
      printf 'filed_by: %s\n' "$(field filed_by "$panel/$ONE/verdict.d/row")"
      printf 'expects: %s\n'  "$(field expects "$panel/$ONE/verdict.d/row")"
      printf 'artifact: %s\n' "$(field artifact "$panel/$ONE/verdict.d/row")"
    fi
  fi

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
  say_table
  ensure_readable || { printf 'list: degraded — table unreadable at %s\n' "$TABLE"; exit 2; }
  if is_paused; then printf 'paused: true\n'; else printf 'paused: false\n'; fi

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
        case "$(basename "$d")" in .*) continue ;; esac   # bookkeeping, not a lane
        t=$((t + 1))
        [ "$(row_state "$d")" = ruled ] && r=$((r + 1))
      done
      # A manifest lane with no row is missing from the count above, and a panel
      # that reports complete while one is missing is the defect `manifest`
      # exists to close. `status` names them; here they are counted.
      ml=""
      [ -r "$s.manifest.d/row" ] && ml="$(field lanes "$s.manifest.d/row")"
      for l in $ml; do
        [ -d "$s$l" ] && continue
        t=$((t + 1))
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
  say_table
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
  say_table
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
  say_table
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

  claim   <pr> <sha> <lane> [--owner <who>] [--for <who>]
          Take the lane atomically. 0 granted — dispatch it. 10 held or already
          ruled — do NOT dispatch. 11 paused. 2 degraded, refused.
          TWO identities, and they are not the same one. --owner is who HOLDS
          the claim: the judge or courier making it, and what a courier checks
          against the judge that asked it. --for is who the lane's verdict is
          expected FROM, recorded as `expects:` and checked by `verdict --by`.
          It defaults to the lane name, which is the reviewer's own name, so
          pass it only for a lane dispatched to a name other than its own.

  verdict <pr> <sha> <lane> <ruling> [--by <who>] [--conf <x>] [--findings <n>]
                                     [--artifact <ref>]
          Record the ruling into the lane's own row, once. 0 recorded. 10 already
          ruled, the first stays of record. 12 recorded but the claim does not
          cover it — no claim row at all, or a --by that is not the name the
          claim recorded as `expects:`. 2 NOT recorded — carry it in your
          handoff. A pause does not block this.
          Pass --by, and pass YOUR OWN LANE NAME. It is checked against the
          claim's `expects:` and never against its holder, so the ordinary case
          — a judge claiming the lane on your behalf and you filing into it —
          is clean at 0. Without --by the row records `filed_by: unrecorded`,
          which is the row saying its own writer is unknown; that is not a pass,
          and it still exits 0.

  release <pr> <sha> <lane> --reason <text>
          Hand a stranded claim back so the lane can be re-dispatched, and clear
          a ruling that was interrupted mid-write. 0 done. 10 the lane has ruled
          — a readable verdict is finished, not stranded. 12 nothing was held
          and nothing needed clearing. 2 degraded.

  manifest <pr> <sha> [--lanes "<lane> …"] [--by <who>]
          With --lanes, record the selected lane set, once: `complete` then
          means complete against it. 0 recorded. 10 already recorded, the first
          stays of record. 64 a malformed lane — nothing is written. 2 degraded.
          Without --lanes, read it back. 0 present · 12 this panel has none.

  status  <pr> <sha> [lane]
          The panel. 0 complete — synthesise. 10 lanes outstanding. 12 no rows.
          2 degraded. With a <lane> it also prints `lane_state:`, one of SIX:
          ruled (0) · claimed, ruling-incomplete, free, never-claimed (all 10)
          · no-row (12). Four states share 10 and have four different remedies,
          so branch on that field rather than on the sentence beside it. The
          sixth, no-row, is a lane this panel has never heard of; a lane the
          manifest selected but nothing claimed reads never-claimed instead.

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
Every ambiguous state refuses. State lives under .claude/state/review-claims/
beside the repository's common git directory, so every worktree of one checkout
resolves ONE table; CLAIM_TABLE_DIR overrides the location. Every verb echoes
the table it resolved as its first line — if two agents disagree about a panel,
compare those before anything else.
HELP
  exit 0
}

sub="${1:-help}"
[ $# -eq 0 ] || shift
case "$sub" in
  claim)    cmd_claim    "$@" ;;
  verdict)  cmd_verdict  "$@" ;;
  release)  cmd_release  "$@" ;;
  manifest) cmd_manifest "$@" ;;
  status)   cmd_status   "$@" ;;
  list)    cmd_list    "$@" ;;
  pause)   cmd_pause   "$@" ;;
  resume)  cmd_resume  "$@" ;;
  paused)  cmd_paused  "$@" ;;
  help|-h|--help) cmd_help ;;
  *) printf 'unknown subcommand: %s\n' "$sub" >&2; usage_die '<claim|verdict|release|manifest|status|list|pause|resume|paused|help>' ;;
esac
