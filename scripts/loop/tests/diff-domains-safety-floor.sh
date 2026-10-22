#!/usr/bin/env bash
# diff-domains-safety-floor.sh — pins the safety-path floor of diff-domains.sh (#212).
# A path on the floor makes safety-sentinel mandatory and post-review.sh refuses an
# approval without a passing safety verdict at that head, so an entry that silently
# stops matching is a sentinel nobody convenes. service/internal/config/ holds the
# registry of the constants FR-091 refuses to run without (SAFE-10, PR #140).
#
# Hermetic: stdin mode (`diff-domains.sh -`), so no repository state, no git, no network.
# Usage: scripts/loop/tests/diff-domains-safety-floor.sh    Exit: 0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$DIR/../diff-domains.sh"

fails=0
run() { OUT="$(printf '%s\n' "$@" | bash "$SCRIPT" - 2>&1)"; RC=$?; }
check() { # check <label> <expected safety_path_candidates line, whole>
  if [ "$RC" = 0 ] && printf '%s\n' "$OUT" | grep -qxF "$2"; then
    printf 'PASS  %s\n' "$1"
  else
    printf 'FAIL  %s — expected rc 0 and the line [%s], got rc %s: %s\n' "$1" "$2" "$RC" \
      "$(printf '%s\n' "$OUT" | grep '^safety_path_candidates:' || printf 'no safety_path_candidates line')"
    fails=$((fails + 1))
  fi
}

# Given the changed paths include owed.go, the floor names it and nothing else.
run service/internal/config/owed.go service/cmd/turfgps/main.go web/src/App.tsx
check "config/owed.go is a safety path" \
  'safety_path_candidates: [service/internal/config/owed.go]'

# Any file under the directory, its tests included.
run service/internal/config/owed_test.go docs/Architecture.md
check "config/owed_test.go is a safety path" \
  'safety_path_candidates: [service/internal/config/owed_test.go]'

# Given paths touching none of the listed candidates, the floor is still empty:
# the entry is an addition, not a widening. Each near miss is on one side of an anchor.
run service/internal/configx/owed.go service/internal/configuration/x.go \
    service/internal/config service/internal/sync/config.go \
    xservice/internal/config/owed.go web/src/config/owed.ts docs/Architecture.md
check "near misses are not safety paths" \
  'safety_path_candidates: []'

# service/internal/ownership/ decides held zones and region lordship, which feed
# takeover time and so the ceiling (#234, from PR #230's safety verdict, S4).
run service/internal/ownership/lordship.go service/internal/ownership/held_test.go \
    service/internal/turf/session.go
check "internal/ownership/ is a safety path" \
  'safety_path_candidates: [service/internal/ownership/lordship.go, service/internal/ownership/held_test.go]'

# Near misses on either side of the ownership anchors.
run service/internal/ownershipx/held.go service/internal/ownership \
    xservice/internal/ownership/held.go service/internal/turf/ownership/held.go
check "ownership near misses are not safety paths" \
  'safety_path_candidates: []'

# service/internal/valhalla/ is the routing adapter: its request decides whether
# a car is kept off ways the map data closes to it (#6).
run service/internal/valhalla/valhalla.go service/internal/valhalla/testdata/README.md \
    service/internal/planner/planner.go
check "internal/valhalla/ is a safety path" \
  'safety_path_candidates: [service/internal/valhalla/valhalla.go, service/internal/valhalla/testdata/README.md]'

# Near misses on either side of the valhalla anchors.
run service/internal/valhallax/valhalla.go service/internal/valhalla \
    xservice/internal/valhalla/valhalla.go service/internal/routing/valhalla/x.go
check "valhalla near misses are not safety paths" \
  'safety_path_candidates: []'

# The six entries the floor already held still match, in input order.
run docs/CalculationSpecification.md docs/SPECIFICATION.md \
    docs/Requirements/safety-exclusions.md docs/Requirements/access-classification.md \
    .claude/skills/safety-path-checklist/SKILL.md .claude/agents/safety-sentinel.md
check "the six existing entries are untouched" \
  'safety_path_candidates: [docs/CalculationSpecification.md, docs/SPECIFICATION.md, docs/Requirements/safety-exclusions.md, docs/Requirements/access-classification.md, .claude/skills/safety-path-checklist/SKILL.md, .claude/agents/safety-sentinel.md]'

[ "$fails" -eq 0 ] || { printf '\n%s test(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
