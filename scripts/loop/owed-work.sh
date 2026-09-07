#!/usr/bin/env bash
# owed-work.sh — what work is owed on an open PR, and has nothing after it (#172, #178).
#
# One question, answered without an LLM: has an agent persisted an artifact that
# obliges a next hop, and did that hop never happen? Two symptoms, one detector:
#
#   A (#172)  a revision packet newer than the PR's newest commit — a remand a
#             judge ruled, persisted, and ended on, that no worker ever received.
#             Measured 2026-09-06: three of six open PRs had carried one for
#             EIGHT DAYS, and nothing anywhere recorded that the hop was owed.
#   B (#178)  a convened panel with no ruling OF ITS OWN IDENTITY — the panel
#             and its ruling both declare `sha:` and `cycle:`, and the ruling is
#             matched to the panel by those and never by posting order, which
#             the judge's own contract fixes the other way round (see B1 in the
#             body) — and a panel naming @validation-agent with no validation
#             verdict after it. Measured on
#             #135 cycle 4, #135 cycle 5 and #154 cycle 6: the pass ended before
#             the mandatory last lane, and each was caught only because a human
#             noticed a silence.
#
# It DETECTS AND REPORTS. Dispatching the worker, resuming the judge and acting
# on any line below stays with @engineering-lead (#172 § Out of scope). Nothing
# here writes to GitHub, to the board, or to the tree.
#
# ---------------------------------------------------------------------------
# MARKER DISCIPLINE — the hard criterion (#172 criterion 2)
# ---------------------------------------------------------------------------
# An artifact is identified ONLY by the `artifact:` key it DECLARED, read as the
# FIRST such line in the comment body. Never by heading text. Never by prose.
#
# This is not a preference. The 2026-09-06 survey that found the gap used two
# ad-hoc `grep` patterns and PRODUCED FALSE NEGATIVES THAT WERE REPORTED AS
# FACT — #147 was called "zero judge ledgers" when it had four, and #154 "no
# packet found" when a packet existed and had been worked. A detector written to
# fix that failure must not be able to commit it.
#
# The first-line rule is `scripts/gates/output-caps.sh`'s and is deliberately the
# same one: the first declaration is the artifact's own, and a later one belongs
# to something it relays verbatim and never reclassifies its courier.
#
# Four declared ids are read, and what each one settles:
#   revision_packet   the remand           — condition A
#   review_ledger     the convened panel   — conditions B1 and B2
#   judgment          the ruling           — discharges B1 when the `sha:` and
#                     `cycle:` it declares are the panel's own
#   reviewer_verdict  a lane's verdict     — discharges B2 when its own
#                     `reviewer:` key names validation-agent
#
# THE FOURTH ID IS THIS SCRIPT'S OWN CHOICE AND IS FLAGGED AS ONE. The contract
# this was built to names three ids and then asks for "a validation verdict",
# which is a class no id was named for. `reviewer_verdict` carrying
# `reviewer: validation-agent` is the only shape that answers it WITHOUT prose
# matching — `review-verdicts § Reviewer verdict` makes both keys mandatory — so
# it was chosen rather than varied silently. Every other candidate (a heading, a
# `validation:` line, an author name) is the thing criterion 2 forbids.
#
# READING A FIELD OF AN IDENTIFIED ARTIFACT IS NOT PROSE MATCHING, and the line
# between the two is worth stating because this script sits on both sides of it.
# Identification is by declared key, always. Once a comment is known to BE a
# `review_ledger`, asking whether its body names `validation-agent` is reading a
# field of a known artifact — the ledger's reviewer column, whose shape is fixed
# in `review-board-dispatch § Incremental review validity`. That match is
# deliberately broad (the whole-word token anywhere in the ledger), because a
# MANDATORY lane must fail toward being reported: an over-report is loud and a
# human dismisses it in a second, and the failure this file exists to prevent is
# the silent one.
#
# ---------------------------------------------------------------------------
# THREE STATES PER PR, NOT TWO. This is the load-bearing part.
# ---------------------------------------------------------------------------
#   owed        a declared artifact, and nothing after it
#   clear       a declared artifact, and something after it
#   undeclared  NO declared artifact of that class on this PR — or, where the
#               class is matched by declared identity, none carrying one this
#               script can match. The script CANNOT ANSWER. Reported distinctly,
#               NEVER printed as `clear`, and never folded into "nothing owed"
#               (#172 criterion 3).
#
# A FOURTH TOKEN, `n/a`, IS PRINTED IN THE `validation=` COLUMN AND IS NOT ONE OF
# THE THREE ABOVE: it marks a class that does not apply to this PR rather than a
# state of one that does. What it means, and what it must not be read as, is under
# WHAT IT CANNOT SEE below.
#
# WHY `undeclared` IS MANDATORY, MEASURED 2026-09-07 AND RETRIEVABLE NOWHERE
# ELSE. Across all four open PRs — #135, #140, #141, #142 — the only declared
# markers in existence were 9 x `artifact: orchestrator_comment` on #135. There
# were ZERO declared `revision_packet` and ZERO declared `judgment` comments on
# the whole board: PR #163 merged the requirement that judges emit the key, and
# every judgment, packet and ledger then alive predates it. PR #135 carried, at
# that moment, an undeclared cycle-4 revision packet (`#issuecomment-5562165264`)
# and an undeclared cycle-5 escalation.
#
# A TWO-STATE DETECTOR REPORTS #135 AS "NOTHING OWED" — reproducing, through its
# own fix, the exact false-negative-reported-as-fact that #172 exists to prevent.
# Re-derive the measurement with:
#   gh api repos/Caisesiume/TurfGPS/issues/<N>/comments --paginate \
#     --jq '.[].body' | grep -oE '^artifact: *[a-z_]+'
#
# The corollary is the exit code: `undeclared` gets `3`, not `0`. The migration
# is what makes this temporary — as declared artifacts accumulate, PRs move from
# `undeclared` to a real answer, and `3` decays toward `0` on its own.
#
# ---------------------------------------------------------------------------
# Usage: scripts/loop/owed-work.sh [pr-number ...]     (default: every open PR)
# Exit — branch on these, do not parse the prose:
#   0  every PR read, nothing owed, and every PR had declared artifacts to judge
#   1  owed work found
#   2  a source could not be read (API, auth, argument)
#   3  all read, but >= 1 PR is `undeclared` and cannot be judged — distinct from 0
#
# PRECEDENCE, which the contract does not fix and a caller needs: 2 > 1 > 3 > 0.
# An unreadable source outranks a finding because a run that cannot see the queue
# cannot vouch for what it did see, which is `fingerprint.sh`'s rule for a
# degraded component and `output-caps.sh`'s for cannot-run. Precedence governs the
# STATUS only, never what is printed and never the counts: the classes are
# counted independently, so a PR owed on one class and unanswerable on another is
# in `owed` AND in `undeclared` and the counts need not sum to the PRs read —
# every owed and every undeclared line found before the failure is still printed.
# Silence and "nothing owed" must not look
# identical (#172 criterion 3), so the summary line is printed on EVERY PATH THAT
# READ SOMETHING — including every failure on one — and `owed 0` is only ever
# reached by a run that read something.
#
# THE USAGE PATH IS THE ONE PATH WITHOUT ONE, and the claim is narrowed to the
# code rather than the code widened to the claim. `owed-work.sh badnumber` prints
# `usage: ...` to stderr and exits 2 with NO `summary:` line: the argument is
# rejected before the first API call, so no queue was ever opened and there is no
# count that would be honest to print — a `summary: owed 0 · clear 0 · ...` here
# would report a queue this run never read, which is the very confusion criterion
# 3 exists to forbid. What that criterion demands of this path is that it not be
# SILENT, and it is not: it names its own failure and takes the same `2` every
# unreadable source takes. Pinned at `X11'` in the recall corpus.
#
# WHAT IT CANNOT SEE, so no reader takes a clean line for more than it is:
#   - An UNDECLARED artifact is invisible, by construction and by criterion 2.
#     `undeclared` is how that is said out loud rather than hidden.
#   - A ledger whose own table already records validation's verdict is still
#     reported owed under B2, because this script reads the ledger for the lane
#     NAME and not for a verdict; parsing verdicts out of a markdown cell is the
#     semantic guessing criterion 2 refuses. The over-report is the chosen
#     direction — see the mandatory-lane paragraph above.
#   - A judgment does NOT discharge B2. That is the whole of #178's second
#     symptom: on #135 cycle 4 and cycle 5 a panel was persisted and a ruling
#     followed, and validation had never run. A detector that let a ruling close
#     the validation question would report exactly those cycles as complete.
#   - `validation=n/a` IS NOT `clear` AND IS NOT `undeclared`, and a reader who
#     takes it for either has been told something false. It is printed when a
#     panel WAS declared and does not name @validation-agent: the B2 obligation
#     does not apply to this PR, so there is nothing here to owe. `clear` would
#     claim a validation verdict discharged the obligation, and none did;
#     `undeclared` would claim the script could not answer, and it answered. So
#     `n/a` counts as nothing-owed AND as answered — standing alone it exits `0`,
#     never `3` — which is pinned at fixture 927 rather than left to the reader,
#     because a fourth token whose exit meaning is unpinned is exactly how
#     `undeclared` comes back as `clear` wearing a different name.
#   - `repos/O/R/pulls/N/commits` is capped by the API at 250 commits and returns
#     the OLDEST first, so a PR past that cap yields a too-old "newest commit"
#     and over-reports condition A. Loud, not silent; no open PR is near it.

set -u

GH="${GH:-/c/Program Files/GitHub CLI/gh.exe}"

# `gh --jq` uses the jq engine compiled into gh. Standalone jq is NOT installed on
# this machine, which is why `dependents.sh` and `fingerprint.sh` both do the same,
# and why the jq side below carries record SELECTION ONLY. Every owed/clear/
# undeclared decision is taken in the shell, where the hermetic suites can stub
# `gh` and exercise it with no network — the convention
# `scripts/loop/tests/dependents-declared-edges.sh` states: a stub returns what
# `gh --jq` would have returned, because a stub cannot evaluate a jq program.
#
# Three call shapes, and they are a contract with those stubs:
#   gh pr list --state open --json number --jq ...        -> one PR number per line
#   gh api repos/{owner}/{repo}/issues/<n>/comments ...   -> one comment record per line
#   gh api repos/{owner}/{repo}/pulls/<n>/commits ...     -> one commit date per line

usage() { echo "usage: owed-work.sh [pr-number ...]" >&2; }

for a in "$@"; do
  case "$a" in
    ''|*[!0-9]*) usage; exit 2 ;;
  esac
done

# --- the PR set -------------------------------------------------------------
# An empty list is ambiguous — no open PRs, or GitHub unread — so the rc of the
# call decides, and `gh auth status` is the second opinion on an empty success.
# An unreadable queue must never be able to read as a quiet one.
if [ "$#" -gt 0 ]; then
  prs="$*"
else
  prs="$("$GH" pr list --state open --limit 200 --json number --jq '.[].number' 2>/dev/null)"
  if [ "$?" -ne 0 ]; then
    echo "error: could not list open pull requests" >&2
    printf 'owed_work: 0 PRs read\nsummary: owed 0 · clear 0 · undeclared 0 · unreadable 0 — the PR list could not be read\n'
    exit 2
  fi
  if [ -z "$prs" ]; then
    "$GH" auth status >/dev/null 2>&1 || {
      echo "error: could not read GitHub" >&2
      printf 'owed_work: 0 PRs read\nsummary: owed 0 · clear 0 · undeclared 0 · unreadable 0 — GitHub could not be read\n'
      exit 2
    }
  fi
fi

# --- THE COMMENT RECORD, DECLARED IN ONE PLACE ------------------------------
# One record per comment: positional, space-separated, one field per name in
# `RECORD_FIELDS` and IN THAT ORDER. The jq program below emits them; the `read`
# further down destructures them; and this line is the only place either of them
# takes the field list from.
#
# WHY ONE PLACE. `read` folds every field past its last variable INTO that last
# variable. It is not an error, `set -u` never fires — nothing is unset — and the
# absorbed field simply becomes part of a value something downstream compares.
# So a record widened on the jq side and not here would not fail; it would answer
# wrongly and quietly, which is the failure this whole file exists to make
# impossible. `overflow` below is that guard made explicit: it is the variable an
# unannounced field lands in, it must always be empty, and a record where it is
# not is reported as unreadable rather than answered from.
#
# Every field is space-free by construction — an ISO-8601 stamp, two captured
# identifier tokens, a 0/1 flag, hex, digits — which is what makes a positional
# record safe here at all, and what makes `overflow` an assertion rather than a
# formality.
RECORD_FIELDS='ts id reviewer namesval sha cycle'

# `\r` is stripped first: GitHub serves CRLF bodies, and a trailing carriage
# return leaves `^artifact:` matching a line whose id then carries an invisible
# character into every comparison below.
#
# `sha:` and `cycle:` are read under the SAME first-declaration rule as
# `artifact:`, for the same reason: the first declaration is the comment's own
# and a later one belongs to something it relays verbatim. A `sha:` that is not
# 7-40 hex characters and a `cycle:` that is not digits are `-` — unusable for
# identity rather than guessed at. Trailing prose after either value is ignored,
# because judges write it (`cycle: 5 — last in budget`).
COMMENT_JQ='
.[]
| ((.body // "") | gsub("\r"; "")) as $b
| ($b | split("\n")) as $L
| ([$L[] | select(test("^artifact:[ \t]*[A-Za-z0-9_]"))][0] // "") as $a
| ([$L[] | select(test("^reviewer:[ \t]*[@A-Za-z0-9_-]"))][0] // "") as $r
| ([$L[] | select(test("^sha:[ \t]*[0-9a-fA-F]{7,40}([^0-9a-fA-F]|$)"))][0] // "") as $s
| ([$L[] | select(test("^cycle:[ \t]*[0-9]"))][0] // "") as $c
| [ .created_at,
    (if $a == "" then "-" else ($a | capture("^artifact:[ \t]*(?<i>[A-Za-z0-9_]+)") | .i) end),
    (if $r == "" then "-" else ($r | capture("^reviewer:[ \t]*(?<i>[@A-Za-z0-9_-]+)") | .i) end),
    (if ($b | test("(^|[^0-9A-Za-z_-])@?validation-agent([^0-9A-Za-z_-]|$)")) then "1" else "0" end),
    (if $s == "" then "-" else ($s | capture("^sha:[ \t]*(?<i>[0-9a-fA-F]+)") | .i) end),
    (if $c == "" then "-" else ($c | capture("^cycle:[ \t]*(?<i>[0-9]+)") | .i) end)
  ] | join(" ")'

# ISO-8601 UTC sorts lexicographically, so `newest` is `sort | tail -1` and
# "strictly after" is a string comparison. No date arithmetic is needed to decide
# anything; it is needed only to PRINT the gaps #172 criterion 1 asks for.
newer() { [ "$1" \> "$2" ]; }

# ONE COMMIT IS ONE COMMIT HOWEVER ITS SHA IS SPELLED, which is
# `review-board-dispatch § The claim table`'s rule for a panel key and must hold
# here too: judges declare `sha: d5f3a58` on one PR and the full 40-hex
# `297632d4e4d7…` on another, and an identity test that called those different
# would report a ruled panel unruled. Either may be the prefix of the other.
# Both sides are validated as hex of at least 7 characters first — which also
# keeps the `case` patterns literal — and anything else is refused rather than
# matched loosely, so a malformed stamp leaves the question unanswered instead
# of discharging it.
sha_same() {
  _a="$(printf '%s' "${1:-}" | tr 'ABCDEF' 'abcdef')"
  _b="$(printf '%s' "${2:-}" | tr 'ABCDEF' 'abcdef')"
  case "$_a" in ''|*[!0-9a-f]*) return 1 ;; esac
  case "$_b" in ''|*[!0-9a-f]*) return 1 ;; esac
  [ "${#_a}" -ge 7 ] && [ "${#_b}" -ge 7 ] || return 1
  case "$_a" in "$_b"*) return 0 ;; esac
  case "$_b" in "$_a"*) return 0 ;; esac
  return 1
}

# WHAT COUNTS AS AN IDENTITY AT ALL, decided in one place. `-` is a key the
# artifact did not declare; EMPTY is a record narrower than `RECORD_FIELDS`
# describes, which is `read` leaving trailing variables unset-but-set and the
# under-wide mirror of the `overflow` guard; and anything that is not hex and
# digits is not a stamp this script may match on. All three are the same answer
# — no identity — and none of them may be compared as though it were one, or a
# panel and a ruling that each declared nothing would match each other.
has_identity() {
  case "${1:-}" in ''|*[!0-9a-fA-F]*) return 1 ;; esac
  case "${2:-}" in ''|*[!0-9]*) return 1 ;; esac
  [ "${#1}" -ge 7 ]
}

# Cycles are compared as numbers, so `07` and `7` are one cycle; a non-numeric
# cycle is no cycle and is refused for the same reason a malformed sha is.
cycle_same() {
  case "${1:-}" in ''|*[!0-9]*) return 1 ;; esac
  case "${2:-}" in ''|*[!0-9]*) return 1 ;; esac
  [ "$1" -eq "$2" ]
}

# The gap in whole days, from two `YYYY-MM-DDTHH:MM:SSZ` stamps. Pure awk
# arithmetic over the civil calendar (the Julian-day formula), because `mktime`
# is a gawk extension and `date -d` is a GNU one: a control-plane script must not
# report a different number on a host that ships a different `date`.
gap_days() {
  awk -v a="$1" -v b="$2" '
    function jdn(y, m, d,   k) { k = int((14 - m) / 12)
      y = y + 4800 - k; m = m + 12 * k - 3
      return d + int((153 * m + 2) / 5) + 365 * y + int(y / 4) - int(y / 100) + int(y / 400) - 32045 }
    function secs(s) { return jdn(substr(s,1,4) + 0, substr(s,6,2) + 0, substr(s,9,2) + 0) * 86400 \
      + substr(s,12,2) * 3600 + substr(s,15,2) * 60 + substr(s,18,2) }
    BEGIN { d = int((secs(b) - secs(a)) / 86400); print (d < 0 ? 0 : d) }'
}

# The run's own clock, read ONCE so that every age printed below is measured
# from one instant rather than from wherever the loop had got to. `claim.sh`'s
# shape, fallback included: a host whose `date` gives no UTC stamp says so in
# the line instead of printing a number derived from nothing.
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)"
case "$NOW" in ????-??-??T??:??:??Z) ;; *) NOW="" ;; esac

n_owed=0; n_clear=0; n_undeclared=0; n_unreadable=0; n_read=0
lines=""; owed_detail=""; undeclared_detail=""
add() { lines="$lines
$1"; }

for pr in $prs; do
  comments="$("$GH" api "repos/{owner}/{repo}/issues/$pr/comments" --paginate --jq "$COMMENT_JQ" 2>/dev/null)"
  crc=$?
  commits="$("$GH" api "repos/{owner}/{repo}/pulls/$pr/commits" --paginate --jq '.[].commit.committer.date' 2>/dev/null)"
  krc=$?

  # `committer.date`, not `author.date`: the question is when the work LANDED on
  # this branch. A rebase or a cherry-pick preserves the author date, so a packet
  # posted before a rebase would read as newer than a commit that in fact came
  # after it — a false positive on exactly the busiest PRs.
  newest_commit="$(printf '%s\n' "$commits" | sed '/^$/d' | sort | tail -1)"

  if [ "$crc" -ne 0 ] || [ "$krc" -ne 0 ] || [ -z "$newest_commit" ]; then
    # An unreadable PR is named and fails the run. It is not a quiet queue
    # (#172 criterion 4, and `fingerprint.sh`'s rule for a degraded component).
    what=""
    [ "$crc" -ne 0 ] && what="comments"
    { [ "$krc" -ne 0 ] || [ -z "$newest_commit" ]; } && what="${what:+$what and }commits"
    add "#$pr unreadable   $what could not be read"
    n_unreadable=$((n_unreadable + 1))
    continue
  fi

  newest_packet=""; newest_panel=""; newest_val=""; panel_names_val=0
  panel_sha="-"; panel_cycle="-"; judgment_seen=0; judgment_ids=""; overflow=""
  # `$RECORD_FIELDS` is deliberately unquoted: splitting the ONE declared field
  # list is what makes it the one place the record's shape is written down. The
  # list holds nothing but names and spaces, so the split is exact.
  # shellcheck disable=SC2086
  while IFS=' ' read -r $RECORD_FIELDS overflow; do
    [ -n "${ts:-}" ] || continue
    [ -z "$overflow" ] || break
    case "$id" in
      revision_packet) newer "$ts" "$newest_packet" && newest_packet="$ts" ;;
      judgment)
        # Every ruling's identity is COLLECTED here and matched below, never
        # matched here: the panel a ruling belongs to may still be later in the
        # stream, and on a correctly ruled PR it always is (Phase 9 then 10).
        # `judgment_seen` is kept apart from the identities because the two
        # absences answer differently below: no declared ruling at all is
        # `owed`, a declared ruling carrying no identity is `undeclared`.
        judgment_seen=1
        has_identity "$sha" "$cycle" && judgment_ids="$judgment_ids
$sha $cycle" ;;
      review_ledger)
        if newer "$ts" "$newest_panel"; then
          newest_panel="$ts"; panel_names_val="$namesval"
          panel_sha="$sha"; panel_cycle="$cycle"
        fi ;;
      reviewer_verdict)
        # Only validation-agent's own verdict discharges B2, and it is its
        # declared `reviewer:` key that says so — never the author, never the prose.
        case "$reviewer" in
          @validation-agent|validation-agent) newer "$ts" "$newest_val" && newest_val="$ts" ;;
        esac ;;
    esac
  done <<EOF
$comments
EOF

  if [ -n "$overflow" ]; then
    # The only thing this can mean: the jq program emits a field `RECORD_FIELDS`
    # does not name. Answering from a record this script cannot parse is exactly
    # the silent absorption the single declaration exists to prevent, so the PR
    # is named unreadable and takes the same `2` every unread source takes.
    add "#$pr unreadable   a comment record carries a field RECORD_FIELDS does not name: '$overflow'"
    n_unreadable=$((n_unreadable + 1))
    continue
  fi
  n_read=$((n_read + 1))

  # Each class says WHY it could not be answered where it decides that it could
  # not, rather than having the reason inferred back out of the token later:
  # `undeclared` now has more than one cause, and a line that names the wrong
  # one is exactly as misleading as no line.
  packet_why=""; ruling_why=""

  # --- A (#172): a remand newer than the newest commit ----------------------
  if [ -z "$newest_packet" ]; then
    packet=undeclared; packet_why="no declared revision_packet"
  elif newer "$newest_packet" "$newest_commit"; then
    packet=owed
    # BOTH gaps, each labelled with the two stamps it spans. #172 criterion 1
    # asks for "the gap in days" and does not settle WHICH two stamps that is:
    # its sentence describes packet-minus-commit, and the table printed under it
    # reports now-minus-packet. #135's packet was ONE day past its newest commit
    # and had been carried EIGHT when the survey found it, so on the issue's own
    # headline instance the two readings differ eightfold — and this line used
    # to print the smaller one under the larger one's name. Settling that text
    # is @engineering-lead's (CORE-07, root cause `requirement`); until it is
    # settled the line privileges neither reading, and neither number can be
    # taken for the other.
    if [ -n "$NOW" ]; then
      carried="and has been carried $(gap_days "$newest_packet" "$NOW") day(s) since, to now $NOW (now-minus-packet)"
    else
      carried="and has been carried for a time this host cannot print: its \`date -u\` gave no UTC stamp"
    fi
    owed_detail="$owed_detail
  #$pr revision_packet posted $newest_packet is $(gap_days "$newest_commit" "$newest_packet") day(s) past the newest commit $newest_commit (packet-minus-commit), $carried"
  else
    packet=clear
  fi

  # --- B1 (#178): a convened panel with no ruling FOR THAT PANEL -------------
  # --- B2 (#178): that panel names the mandatory last lane, which never ran --
  #
  # B1 IS NOT A TIME TEST, and the previous "a judgment strictly after the
  # panel" was not a convention this could lean on — it was the CONTRACT read
  # backwards. `pr-judge § Phase 9` posts the judgment and `§ Phase 10` posts
  # the ledger, in that order, so on a correctly ruled PR the ledger is ALWAYS
  # the newer of the two and "judgment after panel" is false on every one of
  # them. Measured across 9 of 9 ruled cycles (#163 x6, #154 x2, #135 x1):
  # ledger strictly newer, zero counterexamples. #181 cycle 1 is judgment
  # 18:40:55Z, ledger 18:41:01Z, six seconds apart and in that order.
  # #178 criterion 2 asks for a ruling's EXISTENCE — "a convened panel with no
  # ruling" — so dropping the temporal strengthening restores the criterion
  # rather than relaxing it.
  #
  # A ruling is matched to its panel by the IDENTITY BOTH ARTIFACTS DECLARE:
  # `sha:` and `cycle:`. #135's cycle-7 judgment and ledger both declare
  # `sha: d5f3a58, cycle: 7`; #181's both declare `297632d4…` and `cycle: 1`.
  # Two artifacts of one panel say so themselves, and no clock is consulted.
  #
  # THE COMPARISON IS BETWEEN THE TWO ARTIFACTS' OWN DECLARED SHAs — never
  # against the PR's current head. A force-push moves the head, and a test of
  # the form "the panel's sha is the head" would then read a genuinely unruled
  # panel as stale history and discharge it: a SILENT FALSE NEGATIVE, which is
  # the dangerous direction for a detector whose entire purpose is to make an
  # owed hop visible. What the head is doing now is not evidence about whether
  # a panel was ever ruled.
  #
  # AND THE DISCHARGE CLASS GETS THE THIRD STATE FOR ONE OF THE TWO ABSENCES,
  # NOT BOTH. `undeclared` says: an artifact WAS declared and this script cannot
  # read an identity off it. That holds where the panel declares none, and where
  # a `judgment` was declared carrying none — a ruling this script cannot place
  # is not a ruling it may report as missing, and `owed` there would assert an
  # absence the marker discipline cannot see (#172 criterion 2).
  #
  # It does NOT hold where NO judgment was declared at all. Nothing is unreadable
  # in that case: the ruling ITSELF is what is missing, which is #178 criterion 2
  # in its own words — "a convened panel with no ruling" — and the whole symptom
  # this detector exists to catch. `undeclared` there would retire the headline
  # case into the unanswerable bucket, and would contradict B2 below on the
  # identical shape: one ledger with zero declared discharge artifacts already
  # reads `validation=owed`. `judgment_seen` is the discriminator, collected
  # above apart from the identities for exactly this.
  #
  # Where `undeclared` does hold it is loud in its own right: printed under
  # `undeclared:`, counted, and exit 3, never `clear` and never folded into
  # nothing-owed.
  if [ -z "$newest_panel" ]; then
    ruling=undeclared; ruling_why="no declared review_ledger"
    validation=undeclared
  else
    if ! has_identity "$panel_sha" "$panel_cycle"; then
      ruling=undeclared
      ruling_why="the newest declared review_ledger declares no usable sha:/cycle: to match a ruling to"
    elif [ "$judgment_seen" -eq 1 ] && [ -z "$judgment_ids" ]; then
      ruling=undeclared
      ruling_why="no declared judgment carries a sha:/cycle: to match against the panel"
    else
      ruling=owed
      while IFS=' ' read -r jsha jcycle; do
        [ -n "${jsha:-}" ] || continue
        if sha_same "$jsha" "$panel_sha" && cycle_same "$jcycle" "$panel_cycle"; then
          ruling=clear; break
        fi
      done <<EOJ
$judgment_ids
EOJ
      [ "$ruling" = owed ] && owed_detail="$owed_detail
  #$pr review_ledger $newest_panel (sha $panel_sha, cycle $panel_cycle) has no declared judgment of that identity — a panel convened and never ruled"
    fi
    if [ "$panel_names_val" != "1" ]; then
      validation=n/a
    elif newer "$newest_val" "$newest_panel"; then
      validation=clear
    else
      validation=owed
      owed_detail="$owed_detail
  #$pr review_ledger $newest_panel names @validation-agent and no validation verdict follows it — the mandatory last lane"
    fi
  fi

  # --- the three class states, then this PR's state and the counts -----------
  # THE SUBSTRING INVARIANT, stated where it is relied on: each class token is
  # fenced with `|` on both sides, so `*"|owed|"*` matches a WHOLE token and can
  # never match across a join. The previous shape tested `$packet$ruling$valida-
  # tion` — three tokens concatenated bare — where nothing but the accident of
  # today's vocabulary kept a boundary from spelling one of them.
  classes="|$packet|$ruling|$validation|"

  # PRECEDENCE DECIDES THE PRINTED STATE. IT DOES NOT DECIDE THE COUNTS, and the
  # old single `case` let it: one bucket per PR meant a PR owed on one class and
  # unanswerable on another was counted `owed` and ONLY owed, so `n_undeclared`
  # read 0 and the `undeclared:` line was never printed at all — while the
  # precedence paragraph above promises precedence governs the status and never
  # what is printed, and the consumer (`engineering-lead § Phase 1`) branches on
  # those counts. Live on 2026-09-07, #181 and #135 both read `packet=undeclared`
  # under `summary: ... undeclared 0`.
  #
  # So the classes are counted independently: a PR can appear in `owed` AND in
  # `undeclared`, and the four counts therefore need not sum to the PRs read.
  # The EXIT precedence 2 > 1 > 3 > 0 is untouched, and is where owed still
  # outranks undeclared — one status per run, all the counts it was reached by.
  state=clear
  case "$classes" in
    *"|owed|"*) state=owed; n_owed=$((n_owed + 1)) ;;
  esac
  case "$classes" in
    *"|undeclared|"*)
      [ "$state" = owed ] || state=undeclared
      n_undeclared=$((n_undeclared + 1))
      # Named, because a state the script CANNOT ANSWER is the one a reader is
      # most likely to mistake for a clean one.
      miss="$packet_why"
      [ -z "$ruling_why" ] || miss="${miss:+$miss · }$ruling_why"
      undeclared_detail="$undeclared_detail
  #$pr $miss — cannot be judged; this is NOT \"nothing owed\"" ;;
  esac
  [ "$state" = clear ] && n_clear=$((n_clear + 1))

  add "$(printf '#%-5s %-11s packet=%-11s ruling=%-11s validation=%s' \
         "$pr" "$state" "$packet" "$ruling" "$validation")"
done

printf 'owed_work: %s PR(s) read\n' "$n_read"
printf '%s\n' "$lines" | sed '/^$/d'
[ -n "$owed_detail" ]       && { printf 'owed:\n';       printf '%s\n' "$owed_detail" | sed '/^$/d'; }
[ -n "$undeclared_detail" ] && { printf 'undeclared:\n'; printf '%s\n' "$undeclared_detail" | sed '/^$/d'; }
printf 'summary: owed %s · clear %s · undeclared %s · unreadable %s\n' \
  "$n_owed" "$n_clear" "$n_undeclared" "$n_unreadable"

# 2 > 1 > 3 > 0, argued in the header. Nothing above was skipped to reach here.
[ "$n_unreadable" -eq 0 ]  || exit 2
[ "$n_owed" -eq 0 ]        || exit 1
[ "$n_undeclared" -eq 0 ]  || exit 3
exit 0
