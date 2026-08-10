#!/usr/bin/env bash
# diff-domains.sh — deterministic file-domain classification of a diff (Owner Directive 2 §6, §7).
#
# Runs before any LLM in the review path. It CLOSES review lanes only where the
# file→domain mapping is exact — a diff containing no Go files cannot need a Go
# critic, and no reasoning is required to establish that. It never OPENS a lane,
# never assigns a risk tier, and never decides that a safety path was touched:
# everything semantic stays with @change-risk-assessor and @pr-judge (§50).
#
# `safety_path_candidates` is a hint and is labelled as one. Sentinel activation
# is semantic — a safety rule can be changed by a constant in a file this list
# has never heard of, which is exactly why the list may not be a gate.
#
# Usage: scripts/loop/diff-domains.sh [BASE] [HEAD]     (default: origin/main HEAD)

set -u
BASE="${1:-origin/main}"
HEAD="${2:-HEAD}"

files="$(git diff --name-only "$BASE...$HEAD" 2>/dev/null)"
if [ -z "$files" ]; then
  if git rev-parse --verify --quiet "$BASE" >/dev/null 2>&1 \
     && git rev-parse --verify --quiet "$HEAD" >/dev/null 2>&1; then
    printf 'total_files: 0\ndocs_only: false\nlanes_closed: []\nnote: empty diff — classification not applicable\n'
    exit 0
  fi
  printf 'total_files: 0\nerror: could not diff %s...%s\n' "$BASE" "$HEAD"
  exit 2
fi

go=0; frontend=0; docs=0; agent_system=0; schema=0; ci=0; data=0; other=0; total=0

# Order is load-bearing: .claude/ and scripts/ are the agent system rather than
# documentation, and .github/ is CI rather than either, however they are suffixed.
while IFS= read -r f; do
  [ -n "$f" ] || continue
  total=$((total + 1))
  case "$f" in
    .github/*)                             ci=$((ci + 1)) ;;
    .claude/*|scripts/*)                   agent_system=$((agent_system + 1)) ;;
    *.sql|migrations/*|*/migrations/*)     schema=$((schema + 1)) ;;
    *.go|go.mod|go.sum|*/go.mod|*/go.sum)  go=$((go + 1)) ;;
    *.tsx|*.ts|*.jsx|*.css|frontend/*|web/*) frontend=$((frontend + 1)) ;;
    docs/*)                                docs=$((docs + 1)) ;;
    data/*)                                data=$((data + 1)) ;;
    *.md)  case "$f" in */*) other=$((other + 1)) ;; *) docs=$((docs + 1)) ;; esac ;;
    *)                                     other=$((other + 1)) ;;
  esac
done <<EOF
$files
EOF

docs_only=false
[ "$docs" -eq "$total" ] && docs_only=true

closed=""
[ "$go" -eq 0 ]       && closed="$closed go-quality-critic go-structure-critic go-architecture-critic"
[ "$frontend" -eq 0 ] && closed="$closed ux-reviewer design-reviewer ui-engineer"
[ "$schema" -eq 0 ]   && closed="$closed schema-migration-lane"

# The homes where safety rules are currently written. Membership is a prompt to
# look, not a finding, and non-membership proves nothing.
candidates="$(printf '%s\n' "$files" | grep -E '^(docs/CalculationSpecification\.md|docs/Requirements/safety-exclusions\.md|docs/SPECIFICATION\.md)$' 2>/dev/null)"

printf 'total_files: %s\n' "$total"
printf 'domains: {go: %s, frontend: %s, docs: %s, agent_system: %s, schema: %s, ci: %s, data: %s, other: %s}\n' \
  "$go" "$frontend" "$docs" "$agent_system" "$schema" "$ci" "$data" "$other"
printf 'docs_only: %s\n' "$docs_only"
if [ -z "$closed" ]; then
  printf 'lanes_closed: []\n'
else
  printf 'lanes_closed: [%s]\n' "$(printf '%s' "${closed# }" | tr ' ' ',' | sed 's/,/, /g')"
fi
if [ -z "$candidates" ]; then
  printf 'safety_path_candidates: []\n'
else
  printf 'safety_path_candidates: [%s]\n' "$(printf '%s' "$candidates" | tr '\n' ',' | sed 's/,$//; s/,/, /g')"
fi
printf 'hint_only: sentinel activation is semantic\n'
exit 0
