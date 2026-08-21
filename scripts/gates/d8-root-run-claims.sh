#!/usr/bin/env bash
# d8-root-run-claims.sh — does anything state the root-run model instead of citing it? (#118)
#
# ONE model — how a Go command behaves when it is run from a repository root
# that holds no module — was stated in 13 places across 9 files, paraphrased
# differently at every site. It was measured false, and because every site
# owned its own wording, correcting the model corrected none of them and left
# no diff in any of them to notice. Two documentation passes missed the class.
#
# The homes are `Architecture.md § D8` (the layout decision and its cost),
# `local-gates § Backend (Go)` (what each gate reports from the wrong
# directory), and the Makefile (where each recipe's directory is encoded).
# Everything else CITES one of them. That is `local-gates § The law` gate 2.
#
# MEASURED RECALL, FIRST, BECAUSE IT IS THE UNFLATTERING HALF. Both parts were
# run against the pre-repair tree at 6fbf7de, where all eight dependent sites
# were still live and known by enumeration. The result:
#
#   part 1 — structural, sets the exit status ....... 0 of 8 sites
#   part 2 — phrase list, advisory only ............. 6 of 8 sites
#
# PART 1 CAUGHT NONE OF THEM, and the reason is the assumption it was built on.
# The obvious structural test is "does a file state this model without citing
# its home", and in this repository the answer at every one of the thirteen
# sites was that the file DID cite the home — and then restated it anyway, in
# the next clause. Gate-2 duplication here does not look like a missing
# citation. It looks like a citation with a gloss that drifted off it. So part 1
# guards only a shape the historical class contains no instance of: a NEW
# document restating the model having never cited D8 at all. That is worth an
# exit status because it is certain and costs nothing, and it is emphatically
# not the thing #118 was about. Do not read a green part 1 as an absence.
#
# PART 2 IS THE PART WITH RECALL, AND IT IS EXACTLY THE PHRASE-MATCHING
# @docs-reviewer CERTIFIED INSUFFICIENT. That certification is correct and is
# not softened here: "passes vacuously", "resolves against nothing", "selects
# from no packages", "inspects zero files", "finds no packages", "reports no
# vulnerabilities" and "passes against nothing" were seven wordings of one
# claim, and an eighth nobody has written yet will pass this list. It is
# advisory for that reason — it prints lines for a human to read and never
# fails a build on its own opinion. It is in the file because 6 of 8 measured
# beats 0 of 8 measured, and because a reading aid that names the six lines
# worth opening is worth more than a verdict that is wrong about all eight.
#
# THE TWO SITES PART 2 MISSED, so nobody has to re-derive them:
#   · local-gates/SKILL.md:97 — inside a home file. The home exemption is
#     per FILE, and that file is the home of § Backend (Go) while its law 1 was
#     a dependent of § D8. A home can hold a stale dependent and this check
#     cannot see it. Unclosed.
#   · go-quality-critic.md:41 — the outcome word and the directory word sat in
#     different sentences. The check is per line. Its sibling at :48 flagged the
#     file, so a reader would have reached it; that is luck, not design.
#
# WHAT NEITHER PART CAN DO is decide whether a line beside a citation is a
# faithful gloss or a restatement. Only reading it decides that. This file
# narrows what must be read from the whole corpus to a handful of lines; it
# does not remove the reading, and a PR that reports it green has evidence
# about the surface and none about the prose.
#
# Usage: scripts/gates/d8-root-run-claims.sh
# Exit:  0  no file states the model while citing no home (part 2 may still
#           have printed lines to read — check the summary count, not just $?)
#        1  at least one file does; each is printed
#        2  could not run (not a git repository, or no files to check)

set -uo pipefail

cd "$(git rev-parse --show-toplevel 2>/dev/null)" || { echo "not a git repository" >&2; exit 2; }

# Named commands, AND the generic nouns. The generic half is not optional: the
# fourteenth statement reads "every CI job invoking the Go toolchain" and names
# no command at all, so a command-only list missed the one line this check was
# written against. Measured, not assumed -- the probe is in the PR body.
TOOLCHAIN='gofmt|go vet|go test|go build|go run|go mod|golangci-lint|staticcheck|govulncheck|go tool|go gate|go command|go recipe'

# Root or working-directory language. A statement of this model cannot avoid it
# either — the model is entirely about which directory a command resolved from.
DIRECTORY='repository root|repo root|from the root|at its root|working director'

# A citation of a home, in the token form `docs/README.md § Conventions`
# requires. `§ Backend (Go)` with no document is a same-file citation and is
# correct inside local-gates itself.
CITES_HOME='Architecture\.md § D8|local-gates § Backend \(Go\)|§ Backend \(Go\)'

# The homes themselves. These STATE the model; that is their job.
is_home() {
  case "$1" in
    docs/Architecture.md|.claude/skills/local-gates/SKILL.md|Makefile) return 0 ;;
    *) return 1 ;;
  esac
}

# Success-outcome vocabulary, for part 2 only. Open-ended by nature; see above.
#
# It is deliberately narrower than the obvious wording. A bare `passes` or a
# bare `silent` matched two lines on a clean tree that state nothing about a
# root run — "a passing gate says nothing about" (an ordinary gate result) and
# "silently invalidated the first line" (a documented command going stale, not
# a gate reporting success). Both are true sentences that survive the
# retraction, so matching them made part 2 noise on a clean tree, and a check
# that cries wolf on a clean tree is one nobody reads. Precision bought at the
# cost of recall, on a part that is advisory precisely because its recall was
# never going to be complete.
#
# The mandatory space after `no` and `zero` is load-bearing and was added after this
# check flagged its own repair: `selects from no` matches `selects from
# NONE of this repository's packages`, which is a true statement of scope and
# the exact wording that replaced the retired claim. A check that fires on the
# fix for the defect it exists to find is a check that trains its reader to
# ignore it. A word-boundary escape was tried first and is NOT used: it went
# into the file as a literal backspace byte, matched nothing, and silently
# disabled this whole alternative -- the check's own failure mode, in the
# check. Requiring the space is portable, and it is visible in a diff.
OUTCOME='vacuous|silently pass|silent pass|passes silently|passes against nothing|exits? zero|exit 0|reports? success|(finds?|reports?|inspects?|scans?|selects from) (no|zero) +(go +)?(files|packages|faults|vulnerabilities|code|tests|dependencies)|finds? nothing|reads? nothing|read nothing|zero faults|indistinguishable from|prints what a clean'

# A line that explicitly retracts or records the retraction is not restating it.
RETRACTS='retract|withdraw|measured false|no longer|superseded|was itself measured'

# `-co --exclude-standard` so a NEW, not-yet-committed document is checked too.
# The fourteenth statement this check exists to stop arrived in a file that did
# not exist on main, so a tracked-files-only sweep would have missed the one
# case it was written for.
files=$(git ls-files -co --exclude-standard '*.md' 'Makefile')
[ -n "$files" ] || { echo "no files to check" >&2; exit 2; }

fail=0
checked=0
advisory=0

for f in $files; do
  is_home "$f" && continue
  grep -qiE "$TOOLCHAIN" "$f" || continue
  grep -qiE "$DIRECTORY" "$f" || continue
  checked=$((checked + 1))

  if ! grep -qE "$CITES_HOME" "$f"; then
    echo "part 1 · states the root-run model and cites no home: $f"
    grep -niE "$DIRECTORY" "$f" | head -3 | sed 's/^/    /'
    fail=1
    continue
  fi

  # Part 2 is ADVISORY and never sets the exit status. It cannot distinguish a
  # citation glossed correctly from a citation restated wrongly — only reading
  # the line does that — so it prints lines to read rather than verdicts.
  while IFS= read -r line; do
    printf '%s' "$line" | grep -qiE "$RETRACTS" && continue
    echo "review · $f cites a home; read this line for a restatement beside it:"
    printf '    %s\n' "$line"
    advisory=$((advisory + 1))
  done < <(grep -niE "$DIRECTORY" "$f" | grep -iE "$OUTCOME")
done

if [ "$fail" -eq 0 ]; then
  echo "d8-root-run-claims: clean · $checked file(s) on the surface, every one cites a home · $advisory line(s) flagged for reading"
else
  echo "d8-root-run-claims: FAIL · see above; the homes are Architecture.md § D8 and local-gates § Backend (Go)"
fi

exit "$fail"
