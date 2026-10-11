#!/usr/bin/env bash
# gates.sh — every mechanical gate a change owes, in one pass, as one typed result (#210).
#
# It replaces the validation agent: machine evidence before any opinion. The
# developer runs it in its worktree and pastes the YAML into the PR body; the
# critic re-runs it in the review worktree at the head it reviews.
#
# WHAT RUNS is decided by what the change touches — the committed diff against
# BASE plus anything uncommitted — and a gate not run is printed as `skip` with
# its reason, never left out:
#   go       service/ or any .go file     fmt, vet, lint, test (-race), build — each from service/
#   web      web/, once web/package.json   build, lint (no warnings), test — each from web/
#   docs     any .md, or the Makefile      d8-claims: its recall corpus first, then the checker
#   agents   always                        size-ratchet: per-file baselines and the total cap
#   scripts  always                        bash -n over every .sh; scripts/team/tests/*.sh;
#                                          any other suite whose script or test changed
# --all runs everything.
#
# The Go and web commands are the Makefile's recipes, run directly because the
# reference host has no make; change a recipe there and here together. The Go
# build writes into this run's log directory, not the tree. Go commands run in
# service/, the module's directory (`Architecture.md § D8`).
#
# EVERY FIELD IS DERIVED FROM ITS GATE'S EXIT STATUS, never written as a literal.
# `unrun` is a gate the host could not run — the race detector without a C
# compiler, a missing tool. It is not a pass: the result is then `incomplete`,
# and the race-free `go test` is reported beside it as `test_norace`, never in
# its place.
#
# Usage: scripts/team/gates.sh [--base REF] [--all]        (REF default origin/main)
# Exit:  0 pass · 1 a gate failed · 3 incomplete (nothing failed, something unrun)
#        2 cannot run (bad argument, unreadable diff)
# Each gate's output is kept in the `logs:` directory the result names.

set -uo pipefail

ROOT="$(cd "$(dirname "$0")/../.." 2>/dev/null && pwd)" || { echo 'gates: cannot run — cannot resolve the repository' >&2; exit 2; }
BASE='origin/main'; ALL=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --all) ALL=1; shift ;;
    --base) [ "$#" -ge 2 ] || { echo 'gates: cannot run — --base needs a ref' >&2; exit 2; }; BASE="$2"; shift 2 ;;
    *) echo "gates: cannot run — unknown argument: $1" >&2; exit 2 ;;
  esac
done

LOG="$(mktemp -d 2>/dev/null)" && [ -d "$LOG" ] || { echo 'gates: cannot run — mktemp -d failed' >&2; exit 2; }
cd "$ROOT" || exit 2

changed=''
if [ "$ALL" -eq 0 ]; then
  committed="$(git diff --name-only "$BASE...HEAD" 2>/dev/null)" \
    || { echo "gates: cannot run — cannot diff $BASE...HEAD; fetch it, pass --base, or use --all" >&2; exit 2; }
  changed="$(printf '%s\n%s\n%s\n' "$committed" "$(git diff --name-only HEAD 2>/dev/null)" \
    "$(git ls-files --others --exclude-standard 2>/dev/null)" | sed '/^$/d' | sort -u)"
fi
touched() { [ "$ALL" -eq 1 ] || printf '%s\n' "$changed" | grep -qE "$1"; }

failed=0; unrun=0
# gate <name> <command...> — runs the command with its output in $LOG/<name>.log and
# prints pass or fail. It runs inside $( ), so it cannot set a flag here: every result
# goes through note(), which is what raises `failed`.
gate() { local n="$1"; shift; if "$@" > "$LOG/$n.log" 2>&1; then echo pass; else echo fail; fi; }
note() { local r; for r in "$@"; do [ "$r" = fail ] && failed=1; done; return 0; }

# ── go ──
if [ ! -f service/go.mod ]; then go_line='skip — no service/go.mod'
elif ! touched '^service/|\.go$|(^|/)go\.(mod|sum)$'; then go_line='skip — untouched'
else
  fmt_check() { local out; out="$(cd service && gofmt -l .)" || return 1; [ -z "$out" ] || { printf 'not gofmt-clean:\n%s\n' "$out"; return 1; }; }
  r_fmt="$(gate fmt fmt_check)"
  r_vet="$(gate vet sh -c 'cd service && go vet ./...')"
  if command -v golangci-lint >/dev/null 2>&1; then r_lint="$(gate lint sh -c 'cd service && golangci-lint run')"
  else r_lint='unrun (no golangci-lint)'; unrun=1; fi
  r_test="$(gate test sh -c 'cd service && go test -race -count=1 ./...')"; r_norace=''; norace=''
  if [ "$r_test" = fail ] && grep -q -- '-race requires cgo' "$LOG/test.log"; then
    r_test='unrun (race detector needs a C compiler)'; unrun=1
    r_norace="$(gate test-norace sh -c 'cd service && go test -count=1 ./...')"
    norace=", test_norace: $r_norace"
  fi
  mkdir -p "$LOG/bin"
  r_build="$(gate build sh -c "cd service && go build -trimpath -o '$LOG/bin/' ./cmd/...")"
  note "$r_fmt" "$r_vet" "$r_lint" "$r_test" "$r_norace" "$r_build"
  [ "$r_fmt" = pass ] && r_fmt=clean
  go_line="{dir: service, fmt: $r_fmt, vet: $r_vet, lint: $r_lint, test: $r_test$norace, build: $r_build}"
fi

# ── web ──
if [ ! -f web/package.json ]; then web_line='skip — no web/package.json'
elif ! touched '^web/'; then web_line='skip — untouched'
elif ! command -v npm >/dev/null 2>&1; then web_line='unrun (no npm)'; unrun=1
else
  r_install=pass; [ -d web/node_modules ] || r_install="$(gate web-install sh -c 'cd web && npm ci')"
  r_wbuild="$(gate web-build sh -c 'cd web && npm run build')"
  r_wlint="$(gate web-lint sh -c 'cd web && npm run lint -- --max-warnings 0')"
  r_wtest="$(gate web-test sh -c 'cd web && npm run test')"
  note "$r_install" "$r_wbuild" "$r_wlint" "$r_wtest"
  web_line="{dir: web, install: $r_install, build: $r_wbuild, lint: $r_wlint, test: $r_wtest}"
fi

# ── docs ──
if touched '\.md$|^Makefile$'; then
  d8_check() { bash scripts/gates/tests/d8-root-run-claims-recall.sh && bash scripts/gates/d8-root-run-claims.sh; }
  r_d8="$(gate d8-claims d8_check)"; note "$r_d8"
  docs_line="{d8_claims: $r_d8}"
else docs_line='skip — no .md or Makefile touched'; fi

# ── agents ──
r_size="$(gate size-ratchet bash scripts/gates/size-ratchet.sh)"; note "$r_size"
agents_line="{size_ratchet: $r_size}"

# ── scripts ──
syntax_check() {
  local f bad=0
  while IFS= read -r f; do
    [ -f "$f" ] || continue
    bash -n "$f" 2>&1 || { echo "syntax: $f"; bad=1; }
  done < <({ git ls-files '*.sh'; git ls-files --others --exclude-standard '*.sh'; } | sort -u)
  return "$bad"
}
r_syntax="$(gate syntax syntax_check)"; note "$r_syntax"
suites=''; ran=0; suite_fail=0
for t in scripts/*/tests/*.sh; do
  [ -f "$t" ] || continue
  dir="${t%/tests/*}"; base="$(basename "$t" .sh)"; due=0
  case "$dir" in scripts/team) due=1 ;; esac
  [ "$ALL" -eq 1 ] && due=1
  if [ "$due" -eq 0 ]; then
    if printf '%s\n' "$changed" | grep -qxF "$t"; then due=1; fi
    for s in "$dir"/*.sh; do
      n="$(basename "$s" .sh)"
      case "$base" in "$n"|"$n"-*) printf '%s\n' "$changed" | grep -qxF "$s" && due=1 ;; esac
    done
  fi
  [ "$due" -eq 1 ] || continue
  ran=$((ran + 1))
  if bash "$t" > "$LOG/test-$base.log" 2>&1; then :; else suite_fail=1; failed=1; suites="$suites $base"; fi
done
if [ "$ran" -eq 0 ]; then r_suites='none due'
elif [ "$suite_fail" -eq 0 ]; then r_suites="pass ($ran)"
else r_suites="fail:${suites}"; fi
scripts_line="{syntax: $r_syntax, tests: $r_suites}"

result=pass; rc=0
[ "$unrun" -eq 1 ] && { result=incomplete; rc=3; }
[ "$failed" -eq 1 ] && { result=fail; rc=1; }
dirty=''; [ -n "$(git status --porcelain 2>/dev/null)" ] && dirty=' (+ uncommitted changes)'
printf 'gates:\n'
printf '  base: %s\n' "$([ "$ALL" -eq 1 ] && echo 'all — every gate run' || echo "$BASE")"
printf '  head: %s%s\n' "$(git rev-parse HEAD 2>/dev/null)" "$dirty"
printf '  go: %s\n' "$go_line"
printf '  web: %s\n' "$web_line"
printf '  docs: %s\n' "$docs_line"
printf '  agents: %s\n' "$agents_line"
printf '  scripts: %s\n' "$scripts_line"
printf '  result: %s\n' "$result"
printf '  logs: %s\n' "$LOG"
exit "$rc"
