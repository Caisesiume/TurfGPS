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
#   B (#178)  a convened panel with no ruling after it, and a panel naming
#             @validation-agent with no validation verdict after it. Measured on
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
#   judgment          the ruling           — discharges B1
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
#   undeclared  NO declared artifact of that class on this PR. The script CANNOT
#               ANSWER. Reported distinctly, NEVER printed as `clear`, and never
#               folded into "nothing owed" (#172 criterion 3).
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
# STATUS only, never what is printed: every owed and every undeclared line found
# before the failure is still printed. Silence and "nothing owed" must not look
# identical (#172 criterion 3), so the summary line is printed on EVERY path,
# including every failure, and `owed 0` is only ever reached by a run that read
# something.
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

# One record per comment: `<created_at> <declared-id|-> <reviewer|-> <names-validation-agent:1|0>`.
# `\r` is stripped first: GitHub serves CRLF bodies, and a trailing carriage
# return leaves `^artifact:` matching a line whose id then carries an invisible
# character into every comparison below.
COMMENT_JQ='
.[]
| ((.body // "") | gsub("\r"; "")) as $b
| ($b | split("\n")) as $L
| ([$L[] | select(test("^artifact:[ \t]*[A-Za-z0-9_]"))][0] // "") as $a
| ([$L[] | select(test("^reviewer:[ \t]*[@A-Za-z0-9_-]"))][0] // "") as $r
| [ .created_at,
    (if $a == "" then "-" else ($a | capture("^artifact:[ \t]*(?<i>[A-Za-z0-9_]+)") | .i) end),
    (if $r == "" then "-" else ($r | capture("^reviewer:[ \t]*(?<i>[@A-Za-z0-9_-]+)") | .i) end),
    (if ($b | test("(^|[^0-9A-Za-z_-])@?validation-agent([^0-9A-Za-z_-]|$)")) then "1" else "0" end)
  ] | join(" ")'

# ISO-8601 UTC sorts lexicographically, so `newest` is `sort | tail -1` and
# "strictly after" is a string comparison. No date arithmetic is needed to decide
# anything; it is needed only to PRINT the gap #172 criterion 1 asks for.
newer() { [ "$1" \> "$2" ]; }

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
  n_read=$((n_read + 1))

  newest_packet=""; newest_panel=""; newest_judgment=""; newest_val=""; panel_names_val=0
  while IFS=' ' read -r ts id reviewer namesval; do
    [ -n "${ts:-}" ] || continue
    case "$id" in
      revision_packet) newer "$ts" "$newest_packet"   && newest_packet="$ts" ;;
      judgment)        newer "$ts" "$newest_judgment" && newest_judgment="$ts" ;;
      review_ledger)   if newer "$ts" "$newest_panel"; then newest_panel="$ts"; panel_names_val="$namesval"; fi ;;
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

  # --- A (#172): a remand newer than the newest commit ----------------------
  if [ -z "$newest_packet" ]; then
    packet=undeclared
  elif newer "$newest_packet" "$newest_commit"; then
    packet=owed
    owed_detail="$owed_detail
  #$pr revision_packet $newest_packet is $(gap_days "$newest_commit" "$newest_packet") day(s) newer than the newest commit $newest_commit"
  else
    packet=clear
  fi

  # --- B1 (#178): a convened panel with no ruling after it -------------------
  # --- B2 (#178): that panel names the mandatory last lane, which never ran --
  if [ -z "$newest_panel" ]; then
    ruling=undeclared; validation=undeclared
  else
    if newer "$newest_judgment" "$newest_panel"; then
      ruling=clear
    else
      ruling=owed
      owed_detail="$owed_detail
  #$pr review_ledger $newest_panel has no declared judgment after it — a panel convened and never ruled"
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

  case "$packet$ruling$validation" in
    *owed*)       state=owed;       n_owed=$((n_owed + 1)) ;;
    *undeclared*) state=undeclared; n_undeclared=$((n_undeclared + 1))
      # Named, because a state the script CANNOT ANSWER is the one a reader is
      # most likely to mistake for a clean one.
      miss=""
      [ "$packet" = undeclared ] && miss="no declared revision_packet"
      [ "$ruling" = undeclared ] && miss="${miss:+$miss · }no declared review_ledger"
      undeclared_detail="$undeclared_detail
  #$pr $miss — cannot be judged; this is NOT \"nothing owed\"" ;;
    *)            state=clear;      n_clear=$((n_clear + 1)) ;;
  esac

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
