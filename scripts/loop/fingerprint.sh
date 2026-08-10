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

# 1. Open PRs — number, head SHA, updated_at. A head SHA moving is the event.
pr="$("$GH" pr list --state open --json number,headRefOid,updatedAt \
      --jq 'sort_by(.number)' 2>/dev/null)"
[ -n "$pr" ] || { pr="unavailable"; degraded=1; degraded_components="$degraded_components pr"; }

# 2. Board items — id and status only. Scoped per §45: the loop reacts to status,
#    not to card bodies, so the fingerprint must not carry them.
board="$("$GH" project item-list 3 --owner Caisesiume --format json \
         --jq '[.items[] | {id, status}] | sort_by(.id)' 2>/dev/null)"
[ -n "$board" ] || { board="unavailable"; degraded=1; degraded_components="$degraded_components board"; }

# 3. Trunk — the remote ref, not the local one: a stale local main is not an event.
main="$(git ls-remote origin refs/heads/main 2>/dev/null | cut -f1)"
[ -n "$main" ] || { main="unavailable"; degraded=1; degraded_components="$degraded_components main"; }

# 4. Corpus and ADR head — what is true about requirements and ratified decisions.
corpus="$(git log -1 --format=%H -- docs/Requirements docs/adr 2>/dev/null)"
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
[ "$degraded" -eq 0 ] || echo "  degraded:${degraded_components} — read as unavailable; do not read this as quiet"

# `printf '%s\n'`, not '%s': command substitution stripped $new's trailing newline,
# and without it the last component and the digest share a line — which reads as a
# changed component on every subsequent run.
{ printf '%s\n' "$new"; printf 'sha256:%s\n' "$digest"; } > "$STATE_FILE" 2>/dev/null || true

[ "$degraded" -eq 0 ] || exit 2
[ -z "$changed" ] || exit 10
exit 0
