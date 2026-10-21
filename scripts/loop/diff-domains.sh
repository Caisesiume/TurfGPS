#!/usr/bin/env bash
# diff-domains.sh — deterministic file-domain classification of a diff (Owner Directive 2 §6, §7; #210).
#
# Runs before any LLM in the review path. It names the domains a diff touches,
# the critic lenses each domain offers (`docs/TEAM.md § The flow`; the lead
# picks at most two), and the safety paths among the files. It assigns no risk
# and judges nothing.
#
# `safety_path_candidates` is a FLOOR, not a ceiling (#210). A listed path makes
# @safety-sentinel mandatory, and scripts/team/post-review.sh refuses an
# approval without a passing safety verdict at that head. An unlisted path
# proves nothing: a safety rule can move through a constant in a file this list
# has never heard of, so the lead still sends the sentinel on judgement. A PR
# that creates code implementing a safety path adds that path here.
#
# Usage: scripts/loop/diff-domains.sh [BASE] [HEAD]     (default: origin/main HEAD)
#        scripts/loop/diff-domains.sh -                 (the changed paths on stdin, one per line)
# Exit:  0 classified (an empty diff included) · 2 the diff could not be read

set -u
BASE="${1:-origin/main}"
HEAD="${2:-HEAD}"

if [ "$BASE" = "-" ]; then
  files="$(tr -d '\r')"
else
  files="$(git diff --name-only "$BASE...$HEAD" 2>/dev/null)"
fi
if [ -z "$files" ]; then
  if [ "$BASE" = "-" ] || { git rev-parse --verify --quiet "$BASE" >/dev/null 2>&1 \
     && git rev-parse --verify --quiet "$HEAD" >/dev/null 2>&1; }; then
    printf 'total_files: 0\ndocs_only: false\nlenses: {}\nsafety_path_candidates: []\nnote: empty diff — classification not applicable\n'
    exit 0
  fi
  printf 'total_files: 0\nerror: could not diff %s...%s\n' "$BASE" "$HEAD"
  exit 2
fi

go=0; frontend=0; docs=0; agent_system=0; scripts=0; schema=0; ci=0; data=0; other=0; total=0

# Order is load-bearing: .claude/ is the agent system and scripts/ the tooling,
# not documentation, and .github/ is CI rather than either, however suffixed.
while IFS= read -r f; do
  [ -n "$f" ] || continue
  total=$((total + 1))
  case "$f" in
    .github/*)                             ci=$((ci + 1)) ;;
    .claude/*)                             agent_system=$((agent_system + 1)) ;;
    scripts/*|*.sh|Makefile|*Dockerfile|*.dockerignore) scripts=$((scripts + 1)) ;;
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

# The lens table of `docs/TEAM.md § The flow`, one row per domain touched.
# Instructions under .claude/ are read as documents.
lenses=""
add() { [ "$1" -gt 0 ] && lenses="$lenses, $2: [$3]"; }
add $((docs + agent_system)) docs 'accuracy, references'
add "$go"                    go   'correctness, security'
add "$schema"                sql  'data integrity, security'
add "$frontend"              web  'states, accessibility, design'
add $((scripts + ci))        scripts 'correctness, supply chain'

# The homes where safety rules are written, the two files that decide how they
# are reviewed, and the code that implements one: service/internal/config/ is
# the registry of the constants FR-091 refuses to run without, where a default
# for one of them would be authored (#212); and service/internal/ownership/
# decides held zones and region lordship, which feed takeover time and so the
# ceiling (#234). Pinned by scripts/loop/tests/diff-domains-safety-floor.sh.
candidates="$(printf '%s\n' "$files" | grep -E '^(docs/CalculationSpecification\.md|docs/SPECIFICATION\.md|docs/Requirements/(safety-exclusions|access-classification)\.md|\.claude/skills/safety-path-checklist/SKILL\.md|\.claude/agents/safety-sentinel\.md|service/internal/config/.+|service/internal/ownership/.+)$' 2>/dev/null)"

printf 'total_files: %s\n' "$total"
printf 'domains: {go: %s, frontend: %s, docs: %s, agent_system: %s, scripts: %s, schema: %s, ci: %s, data: %s, other: %s}\n' \
  "$go" "$frontend" "$docs" "$agent_system" "$scripts" "$schema" "$ci" "$data" "$other"
printf 'docs_only: %s\n' "$docs_only"
printf 'lenses: {%s}\n' "${lenses#, }"
if [ -z "$candidates" ]; then
  printf 'safety_path_candidates: []\n'
else
  printf 'safety_path_candidates: [%s]\n' "$(printf '%s' "$candidates" | tr '\n' ',' | sed 's/,$//; s/,/, /g')"
fi
printf 'note: a listed safety path makes safety-sentinel mandatory; an unlisted one proves nothing\n'
exit 0
