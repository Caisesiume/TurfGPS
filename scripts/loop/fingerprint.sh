#!/usr/bin/env bash
# fingerprint.sh — the deterministic loop-state fingerprint (Owner Directive 2 §8, §9).
#
# One question, answered without an LLM: has anything the loop reacts to changed
# since the last check? @engineering-lead runs this FIRST on every cron or wake,
# because no LLM agent may run merely to discover that nothing changed.
#
# Exit codes — branch on these, do not parse the prose:
#   0   UNCHANGED   no dispatch, no digest; a one-line acknowledgement at most
#   10  CHANGED     dispatch only the agent the changed component implicates
#   2   DEGRADED    a component could not be read; treat as CHANGED, and say which
#
# An unreadable source records itself as `unavailable` rather than as "no change",
# because a broken API must never be able to read as a quiet loop.
#
# Components, one labelled line each: pr · board · main · corpus
# State lives in .claude/state/loop-fingerprint-<consumer> — one file per caller (arg 1,
# default "session"). Per-consumer state exists because a shared file would let the
# first caller consume a CHANGED signal a second caller still needs.
#   usage: fingerprint.sh [consumer]

set -u

GH="${GH:-/c/Program Files/GitHub CLI/gh.exe}"
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
CONSUMER="${1:-session}"
STATE_FILE="$ROOT/.claude/state/loop-fingerprint-$CONSUMER"
mkdir -p "$ROOT/.claude/state" 2>/dev/null || true

degraded=0
degraded_components=""

# `gh --jq` uses the jq engine compiled into gh. Standalone jq is NOT installed on
# this machine, and piping to it made both reads fail open as `unavailable` — a
# fingerprint that cannot see the board is worse than no fingerprint, because it
# reports change on every run and trains its reader to ignore it.

# 1. Open PRs — number, head SHA, draft state. These three ARE the events the loop
#    reacts to: a head SHA moving, a PR entering or leaving the open list, and
#    draft→ready. `updatedAt` was here and is deliberately gone — a comment, a
#    label, or any bot touching the PR bumps it, waking @pr-judge to re-read a diff
#    that did not move. Nothing else about a PR is consumed by the workflow, so
#    nothing else belongs in the fingerprint.
pr="$("$GH" pr list --state open --json number,headRefOid,isDraft \
      --jq 'sort_by(.number)' 2>/dev/null)"
[ -n "$pr" ] || { pr="unavailable"; degraded=1; degraded_components="$degraded_components pr"; }

# 2. Board items — id and status only. Scoped per §45: the loop reacts to status,
#    not to card bodies, so the fingerprint must not carry them. `--limit` is
#    explicit and generous: the default page is 30 and this board already carries
#    37+ items, so an implicit limit truncates the board and silently fingerprints
#    a prefix of it — items past the cut could change status forever unnoticed.
board="$("$GH" project item-list 3 --owner Caisesiume --limit 200 --format json \
         --jq '[.items[] | {id, status}] | sort_by(.id)' 2>/dev/null)"
[ -n "$board" ] || { board="unavailable"; degraded=1; degraded_components="$degraded_components board"; }

# 3+4. Trunk and corpus, both from the SAME fetched remote ref. The local checkout
#      is not a source of truth here: main came from the remote while corpus came
#      from local history, so a stale checkout reported a moved `main` with an
#      unmoved `corpus` and the requirements/ADR change that actually landed woke
#      nobody. One fetch, then both reads against origin/main. `-C "$ROOT"` because
#      the corpus pathspec is relative and this may be invoked from anywhere.
#      A fetch that fails leaves BOTH reads answerable from a stale tracking ref,
#      which would read as a quiet loop — so a failed fetch degrades explicitly.
git -C "$ROOT" fetch origin main --quiet 2>/dev/null \
  || { degraded=1; degraded_components="$degraded_components remote-fetch"; }

main="$(git -C "$ROOT" rev-parse --verify --quiet origin/main 2>/dev/null)"
[ -n "$main" ] || { main="unavailable"; degraded=1; degraded_components="$degraded_components main"; }

corpus="$(git -C "$ROOT" log -1 --format=%H origin/main -- docs/Requirements docs/adr 2>/dev/null)"
[ -n "$corpus" ] || { corpus="unavailable"; degraded=1; degraded_components="$degraded_components corpus"; }

new="$(printf 'pr:%s\nboard:%s\nmain:%s\ncorpus:%s\n' "$pr" "$board" "$main" "$corpus")"
digest="$(printf '%s' "$new" | sha256sum 2>/dev/null | cut -c1-16)"
[ -n "$digest" ] || digest="nodigest"

changed=""
if [ -f "$STATE_FILE" ]; then
  for c in pr board main corpus; do
    old_line="$(grep "^$c:" "$STATE_FILE" 2>/dev/null | head -1)"
    new_line="$(printf '%s\n' "$new" | grep "^$c:" | head -1)"
    [ "$old_line" = "$new_line" ] || changed="$changed $c"
  done
else
  changed=" first-run"
fi

if [ -z "$changed" ]; then
  echo "UNCHANGED"
else
  echo "CHANGED"
  for c in $changed; do
    if [ "$c" = "first-run" ]; then
      echo "  first-run — no previous fingerprint; treat every component as new"
    else
      printf '%s\n' "$new" | grep "^$c:" | cut -c1-110 | sed 's/^/  /'
    fi
  done
fi
[ "$degraded" -eq 0 ] || echo "  degraded:${degraded_components} — unreadable or stale; do not read this as quiet"

# `printf '%s\n'`, not '%s': command substitution stripped $new's trailing newline,
# and without it the last component and the digest share a line — which reads as a
# changed component on every subsequent run.
{ printf '%s\n' "$new"; printf 'sha256:%s\n' "$digest"; } > "$STATE_FILE" 2>/dev/null || true

[ "$degraded" -eq 0 ] || exit 2
[ -z "$changed" ] || exit 10
exit 0
