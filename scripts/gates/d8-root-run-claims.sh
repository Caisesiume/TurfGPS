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
# Everything else CITES one of them. That is the duplication gate,
# `local-gates § Documentation gates` gate 2. It is NOT gate 2 of
# `local-gates § The law`, which is the stop-and-report rule and was the
# address this file cited until cycle 2: that heading resolved, and did not
# hold the claim attributed to it.
#
# Run it with `make d8-claims`, which runs the recall corpus FIRST and only
# then this check. A verdict from an instrument whose recall was not just
# demonstrated is the thing this file exists to refuse.
#
# ---------------------------------------------------------------------------
# WHAT SETS THE EXIT STATUS, AND WHAT EACH PART CAN ACTUALLY SEE
#
# PART 1 — a non-home file that states the model and cites no home at all.
# It caught 0 of the 8 historical sites, and that is the shape rather than a
# tuning defect: at every one of the thirteen sites the file DID cite the home
# and then restated it anyway in the next clause. Gate-2 duplication here does
# not look like a missing citation. It looks like a citation with a gloss that
# drifted off it.
#
# `CITES_HOME` is tested PER FILE, so one routine `§ Backend (Go)` citation
# anywhere in a document exempts that whole document from part 1 permanently.
# Part 1 therefore fires only on a document citing nothing at all — narrower
# than the "certain" this header claimed until cycle 2 measured it. It is kept
# because it costs nothing and that document is a real shape; it is not the
# check for this class, and a green part 1 is not an absence. Part 2 is.
#
# PART 2 — a non-home file that restates the RETRACTED OUTCOME beside its
# citation. It now SETS THE EXIT STATUS. It is allowed to, because of what it
# matches and what it deliberately does not:
#
#   flagged      a sentence claiming a wrong-directory run SUCCEEDED, found
#                nothing, or is indistinguishable from a clean one — the
#                vacuous-pass claim, which is what was measured false and
#                retracted.
#   not flagged  a sentence stating SCOPE ("scans none of this service's
#                dependencies") or the surviving operational instruction
#                ("report the directory you ran it in"). That is the remedy
#                `local-gates § Documentation gates` gate 2 prescribes — a
#                citation with a clause saying what a reader will find there —
#                and a check that condemned the remedy would be worthless.
#
# All eight repairs on this branch replaced the first with the second, which is
# why the repaired tree is green and the pre-repair tree is not. Both halves
# are asserted by the corpus, so neither is a claim taken on trust.
#
# MATCHING IS PER SENTENCE, NOT PER LINE, and that is load-bearing. A markdown
# paragraph is one line. Per line, the outcome word and the location word may
# sit in different sentences about different subjects: measured, that missed
# `go-quality-critic.md:41`, where the two sat two sentences apart, and it
# flagged `validation-agent.md:69`, where "indistinguishable from" is about a
# semantic guess versus a test result and concerns no directory at all.
# Requiring both in ONE sentence fixed both, in opposite directions.
#
# THE HOME EXEMPTION IS PER SECTION, not per file, for the two markdown homes.
# A home file can hold a stale DEPENDENT: one of the eight sites was
# `local-gates` law 1, a dependent of `Architecture.md § D8` sitting inside
# the file that is the home of `local-gates § Backend (Go)`. A per-file
# exemption cannot see that, and did not.
# The Makefile is a home of ONE claim — where each recipe's directory is
# encoded — and of nothing part 2 matches: `Architecture.md § D8` gives what each
# gate REPORTS from the wrong directory to `local-gates § Backend (Go)` alone.
# So it is exempt from part 1, a home having nothing to cite itself for, and
# from no line of part 2. It was `ALL` until cycle 2, which skipped the whole
# file ahead of BOTH parts — and the thirteenth restatement of this very model
# was sitting in it, at `Makefile:16`, while this check called the tree clean.
# The one file the class had most recently occupied was the one file the sweep
# would not open. A home file whose home section cannot be located is exit 2,
# never a verdict.
#
# MEASURED RECALL, FIRST, BECAUSE IT IS THE UNFLATTERING HALF — and it is now
# a command rather than a claim:
#
#     scripts/gates/tests/d8-root-run-claims-recall.sh
#
#   part 1, exit status ......................... 0 of 9 sites
#   part 2, exit status ......................... 8 OF 9 SITES
#   the 9 repaired sites, which must NOT flag ... 0 of 9 flagged
#   held-out wordings, invented for the corpus .. 0 OF 4 CAUGHT
#
# THOSE FOUR ROWS ARE ASSERTED BY THAT CORPUS, FOR AGREEMENT WITH WHAT IT JUST
# MEASURED — never for a value. They were a hand-kept second statement of it
# until cycle 3, which is the duplication shape #118 exists to collapse,
# reproduced inside the instrument built to enforce it: adding one OUTCOME
# alternative moved the corpus's held-out count to `1 of 4` while the last row
# still read `0 OF 4 CAUGHT`, every assertion passed and nothing failed. The
# corpus now reads these four rows and goes red when a number here is not the
# number it just measured. The held-out row stays PRINTED rather than asserted
# at a value — what is asserted is only that this row repeats it — so it still
# cannot be tuned to a target, only kept honest.
#
# THE NINE IN THE FIRST THREE ROWS IS: the 8 dependent sites at 6fbf7de, plus
# `Makefile:16-23` at c8088fa^, the thirteenth restatement and the one part 2
# misses. IT IS NOT the nine the recall corpus asserts a flag count over in its
# FIXTURE 1 — that one is the same 8 dependents plus a synthetic
# `Architecture.md` dependent-section case, and it EXCLUDES the Makefile. Two
# disjoint nines, and each is now named where it is printed or asserted.
#
# The 9 are the historical class, so that recall is IN-SAMPLE: it says this
# check catches the shape it was built from, and not all of that. The site it
# misses is the ninth, the Makefile's own — see below. The last line is the
# out-of-sample one and it is the number to believe. Four restatements were
# written for the corpus in wordings drawn from none of the thirteen — "compiles
# nothing and still exits 0", "a green build over an empty tree", "quietly
# succeeds having examined nothing", "sees a directory rather than a module" —
# and this check catches none of them. The corpus prints that count on every
# run and does not assert it, so it stays in view and cannot be tuned away by
# adding those four phrases here; adding them would only make the number lie.
#
# So: this fails on most restatements of the shape the thirteen took, and it
# will not fail on one phrased in a way nobody here has phrased it yet.
#
# THE ONE IN-SAMPLE SITE IT MISSES is the Makefile's, `Makefile:16-23` as it
# stood at c8088fa^ — the thirteenth, and the most recent. Bringing the file
# into reach in cycle 2 was necessary and was not sufficient: the sweep opens it
# now and still returns clean over that paragraph. The cause was measured and it
# is NOT the hard wrapping — unwrapped into single lines, the verdict does not
# change.
#
# TWO BLOCKERS, NOT ONE, AND EACH ALONE IS ENOUGH. Until cycle 3 this paragraph
# named a single cause — that the location cue sat two sentences from the
# outcome and part 2 requires both in ONE sentence — and that was false in the
# costly direction: it pointed the next maintainer at a lever that on its own
# does nothing. Re-measured against `git show c8088fa^:Makefile`:
#
#   1. THE RANGE CARRIES NO RECOGNISED LOCATION TOKEN AT ALL. `LOCATION` matches
#      nothing in lines 16-23 — zero hits for the assembled regex, and zero for
#      every one of its alternatives probed singly. "Measured from this root"
#      reads like the cue and is not one: the list carries `from the root`, and
#      `from this root` appears in no alternative. The nearest recognised
#      location word is `at the repository root` at `Makefile:12`, four lines
#      ABOVE the range — that is what puts the file on the surface, which is why
#      this miss happens after the file is opened rather than before.
#   2. The outcome and that non-token also sit in different sentences.
#
# SO NEITHER LEVER WORKS ALONE, and this was probed rather than reasoned:
# collapsing the whole paragraph into one sentence flags 0; adding
# `from this root` to `LOCATION_ALTS` flags 0; doing BOTH flags 1.
#
# THE ABSENCE OF `from this root` IS DELIBERATE AS OF CYCLE 3, AND WAS NOT
# BEFORE. Nothing here recorded it as an exclusion, and the sentence removed
# above treated the phrase as already recognised, which is what an oversight
# looks like rather than a design. Its cost was measured before this was
# written: with it added, output byte-identical to the run without it — 0 new
# flags, 0 false positives — because `this root` occurs 0 times in the swept
# corpus at head, this PR having repaired the one wording that used it. So it is
# free, and it buys nothing: it does not close the miss, and it would put an
# alternative in a phrase list with no fixture behind it. It stays out, and this
# paragraph is the record of that choice rather than of the omission it replaces.
#
# OF THE CHECKER-SIDE LEVERS, dropping the pairing conjunct is still the only
# one that catches it — measured on a site-9-only tree, where it flags
# `Makefile:19` and `:20`. Because the surface filter already requires a
# file-level location word, dropping that conjunct pairs at FILE scope rather
# than at sentence scope, which is how it reaches a paragraph carrying no
# location token of its own. Its cost was measured on the repaired tree at
# b651862: 5 sentences flagged, not one of them a restatement — `Makefile:118`,
# `Makefile:137`, `validation-agent.md:69`, `local-gates/SKILL.md:104`,
# `docs/Architecture.md:558`. Two are this file's certified repair;
# `validation-agent.md:69` is the exact false positive the per-sentence rule was
# introduced to remove. The ceiling here is the same one the vocabulary has: a
# check that condemns the fix is worse than a narrow one. Both halves are
# asserted by the corpus, the miss included.
#
# THE VOCABULARY CANNOT BE BROADENED TO CLOSE THAT GAP, and this was measured
# rather than assumed. Widening the outcome list from the vacuous-pass family
# to the general success/negation family (`pass`, `clean`, `none`, `nothing`,
# `green`, `empty`) flags 3 of the 7 glosses certified as correct repairs —
# `go-quality-critic.md:48`, `linus-quality-critic.md:98` and
# `linus-security-critic.md:104` — plus a YAML example line. A check that fails
# the build on the accepted remedy is worse than a narrow one. The ceiling on
# this check's recall is set by the requirement that it not condemn the fix,
# and that is a property of the class rather than of this list.
#
# WHAT IT STILL CANNOT DO is decide whether a line beside a citation is a
# faithful gloss or a restatement in wording no alternative below covers. Only
# reading it decides that. This file narrows what must be read from the whole
# corpus to a handful of lines; a PR that reports it green has evidence about
# the surface and none about the prose.
# ---------------------------------------------------------------------------
#
# Usage: scripts/gates/d8-root-run-claims.sh
# Exit:  0  no file states the retracted model outside a home
#        1  at least one does; each is printed with the sentence
#        2  could not run — not a git repository, an empty corpus, nothing on
#           the surface, no file reaching part 2, an unlocatable home section,
#           a corpus file that is not text (a NUL byte, which is also every
#           UTF-16 file), or a phrase list that does not compile. Never
#           reported as clean.

set -uo pipefail

cd "$(git rev-parse --show-toplevel 2>/dev/null)" || { echo "not a git repository" >&2; exit 2; }

# Every list below is ONE ALTERNATIVE PER LINE, so a ninth phrase is a one-line
# addition that a diff shows as an addition. The previous form put 26
# alternatives on one 346-character line, and that review surface is what let a
# literal backspace byte into this file unnoticed.

# Named commands, AND the generic nouns. The generic half is not optional: the
# fourteenth statement reads "every CI job invoking the Go toolchain" and names
# no command at all, so a command-only list missed the one line this check was
# written against. Measured, not assumed -- the probe is in the PR body.
TOOLCHAIN_ALTS=(
  'gofmt'
  'go vet'
  'go test'
  'go build'
  'go run'
  'go mod'
  'golangci-lint'
  'staticcheck'
  'govulncheck'
  'go tool'
  'go gate'
  'go command'
  'go recipe'
)

# Root or working-directory language, used as the FILE-level filter: it decides
# which files reach either part. A statement of this model cannot avoid it, the
# model being entirely about which directory a command resolved from.
DIRECTORY_ALTS=(
  'repository root'
  'repo root'
  'from the root'
  'at its root'
  'working director'
)

# The SENTENCE-level filter, a superset of the file-level one. It is wider
# because a sentence can point at the wrong directory without naming the root:
# `go-quality-critic.md:41` said "from the wrong place", and the file-level list
# above never had to reach it because the same file names `repo root` two lines
# up. Widening the sentence list costs nothing — a sentence must ALSO carry an
# outcome below to be flagged — while widening the file list would move files
# into part 1. That is why the two are separate lists.
LOCATION_ALTS=(
  "${DIRECTORY_ALTS[@]}"
  'from there'
  'from here'
  'the wrong (place|directory)'
  'wrong directory'
  'empty root'
  'over a root'
  'outside the module'
)

# A citation of a home, in the token form `docs/README.md § Conventions`
# requires. `§ Backend (Go)` with no document is a same-file citation and is
# correct inside local-gates itself. Part 1 only.
CITES_HOME_ALTS=(
  'Architecture\.md § D8'
  'local-gates § Backend \(Go\)'
  '§ Backend \(Go\)'
)

# The retracted outcome, and only it. See the header for why this list may not
# be widened into the general success vocabulary: three of the certified
# repairs sit inside that wider family.
OUTCOME_ALTS=(
  'vacuous'
  'silently pass'
  'silent pass'
  'passes silently'
  'passes against nothing'
  'resolves? against nothing'
  'exits? zero'
  'exit 0'
  'reports? success'
  '(finds?|reports?|inspects?|scans?|selects from) (no|zero) +(go +)?(files|packages|faults|vulnerabilities|code|tests|dependencies)'
  'finds? nothing'
  'reads? nothing'
  'read nothing'
  'zero faults'
  'indistinguishable from'
  'prints what a clean'
  'what a clean (tree|module|run) prints'
  'character-for-character'
  'identical to (what )?a clean'
)
# The mandatory space after `no` and `zero` above is load-bearing and was added
# after this check flagged its own repair: `selects from no` matches `selects
# from NONE of this repository's packages`, which is a true statement of scope
# and the exact wording that replaced the retired claim at three sites. A check
# that fires on the fix for the defect it exists to find is a check that trains
# its reader to ignore it. A word-boundary escape was tried first and is NOT
# used: it went into the file as a literal backspace byte, matched nothing, and
# silently disabled that whole alternative -- the check's own failure mode, in
# the check. Requiring the space is portable, it is visible in a diff, and the
# non-printing-byte guard below now refuses that byte outright.

# A sentence that explicitly RECORDS the retraction is not restating it. Tested
# per sentence, not per line: a retraction word anywhere in a 400-word markdown
# paragraph used to exempt every sentence in it. `no longer` was dropped from
# this list in cycle 2 — it marks a change rather than a retraction, so it
# would exempt "run from the root, go vet no longer resolves the module and
# exits zero", which is the defect wearing the exemption. Dropping it changes
# no verdict on either the pre-repair tree or the repaired one; measured.
#
# An exempted sentence is PRINTED and COUNTED, never skipped in silence. This
# exemption is the one thing standing between the repaired tree and a red gate
# today, and a suppression that decides an exit status has to be visible.
RETRACTS_ALTS=(
  'retract'
  'withdraw'
  'measured false'
  'superseded'
  'was itself measured'
)

join_alts() { local IFS='|'; printf '%s' "$*"; }

# A broken phrase list must not be indistinguishable from a clean tree. Two
# failures are refused here, both of which this file has already suffered or
# was one edit away from: an alternative that does not compile, and an
# alternative carrying a control byte, which compiles and matches nothing.
# Each is named individually, which is why the lists are arrays.
#
# The test is `[[:cntrl:]]` under LC_ALL=C, and not `[^[:print:]]`. The wider
# form was tried first and rejected THIS FILE: `§` is two bytes of UTF-8, which
# the C locale reads as non-printing, so the guard against a stray control byte
# would have condemned every citation in the file it guards. Control bytes are
# the class that hides — the backspace this list already shipped is 0x08 — and
# UTF-8 text bytes are not in it.
validate_alts() { # validate_alts <list-name> <alternative>...
  local name="$1" a st; shift
  for a in "$@"; do
    if printf '%s' "$a" | LC_ALL=C grep -q '[[:cntrl:]]'; then
      printf 'd8-root-run-claims: cannot run — %s carries a control byte in the alternative: %s\n' "$name" "$a" >&2
      exit 2
    fi
    printf '' | grep -qE -- "$a" >/dev/null 2>&1
    st=$?
    if [ "$st" -gt 1 ]; then
      printf 'd8-root-run-claims: cannot run — %s does not compile at the alternative: %s\n' "$name" "$a" >&2
      exit 2
    fi
  done
}

validate_alts TOOLCHAIN  "${TOOLCHAIN_ALTS[@]}"
validate_alts DIRECTORY  "${DIRECTORY_ALTS[@]}"
validate_alts LOCATION   "${LOCATION_ALTS[@]}"
validate_alts CITES_HOME "${CITES_HOME_ALTS[@]}"
validate_alts OUTCOME    "${OUTCOME_ALTS[@]}"
validate_alts RETRACTS   "${RETRACTS_ALTS[@]}"

TOOLCHAIN="$(join_alts "${TOOLCHAIN_ALTS[@]}")"
DIRECTORY="$(join_alts "${DIRECTORY_ALTS[@]}")"
LOCATION="$(join_alts "${LOCATION_ALTS[@]}")"
CITES_HOME="$(join_alts "${CITES_HOME_ALTS[@]}")"
OUTCOME="$(join_alts "${OUTCOME_ALTS[@]}")"
RETRACTS="$(join_alts "${RETRACTS_ALTS[@]}")"

# Compiling is not matching. This sentence is `linus-quality-critic.md:98` as it
# stood at 6fbf7de, and both live lists must still match it: an assembled regex
# that matches nothing is the backspace byte's exact signature.
CANARY='run from the repository root, -run selects from no packages and passes'
printf '%s' "$CANARY" | grep -qiE -- "$LOCATION" || { echo "d8-root-run-claims: cannot run — LOCATION no longer matches its canary" >&2; exit 2; }
printf '%s' "$CANARY" | grep -qiE -- "$OUTCOME"  || { echo "d8-root-run-claims: cannot run — OUTCOME no longer matches its canary"  >&2; exit 2; }

# The homes, and the heading that owns the claim inside each. `PART1-ONLY` is
# the Makefile: a home whose one claim is nothing part 2 matches, so it owns no
# exempt range and every line of it is read. A home is exempt from part 1
# outright — requiring a home to cite itself is meaningless — and exempt from
# part 2 only within the range that heading owns.
#
# These are read by awk rather than grep, so a literal paren is bracketed and
# never backslashed: awk takes `\(` as an escape it does not know, warns, and
# drops the backslash, turning the literal into a group — so `Backend \(Go\)`
# silently stops matching `### Backend (Go)`. Measured, from the guard below
# firing. `[(]` is the same character to both dialects.
#
# EVERY ROW BELOW OWES A FIXTURE, IN BOTH DIRECTIONS, AND SAYS WHICH ONE.
# A row is the strongest suppression this checker has — it exempts a whole
# heading range from part 2 — which makes it the widest-blast-radius edit in
# the file, and it is guarded by nothing while the phrase lists are guarded in
# both directions. Measured at 43ee942: a row was added and the recall corpus
# still printed `37/37 all passed`. What each row owes, in
# `scripts/gates/tests/d8-root-run-claims-recall.sh`:
#
#   docs/Architecture.md
#       FIXTURE 1, both directions — a sentence inside the D8 section states
#       the model and must NOT flag, while a dependent section of the same
#       file must.
#   .claude/skills/local-gates/SKILL.md
#       FIXTURE 1 site 8 — the dependent `The law` section still flags — and
#       FIXTURE 2, where its repair must not, plus FIXTURE 4 `renamed` for the
#       heading that no longer resolves.
#   Makefile
#       FIXTURE 5, both directions — mk-pre and mk-post, each verbatim, each
#       asserting reach rather than only a verdict.
#
# A NEW ROW BRINGS ITS OWN PAIR BEFORE IT IS ADDED HERE. The next one is named
# in #118 DoD 2: PR #67 lands `docs/DEPLOYMENT.md`, whose `:198` restates this
# very model. Nothing mechanically enforces the obligation yet — a corpus check
# enumerating these rows and failing on one carrying no fixture is `FW-23`, and
# is deliberately not built here.
home_pattern() { # home_pattern <path>; prints the heading regex, or returns 1
  case "$1" in
    docs/Architecture.md)                printf '%s' '^#+[[:space:]]+D8([[:space:]]|$)' ;;
    .claude/skills/local-gates/SKILL.md) printf '%s' '^#+[[:space:]]+Backend [(]Go[)]' ;;
    Makefile)                            printf '%s' 'PART1-ONLY' ;;
    *) return 1 ;;
  esac
}

# The section runs from its heading to the line before the next heading of the
# same or a higher level, or to end of file.
home_range() { # home_range <path> <heading-regex>; prints "START END", or returns 1
  awk -v pat="$2" '
    start == 0 && $0 ~ pat { if (match($0, /^#+/)) { start = NR; lvl = RLENGTH; next } }
    start != 0 && end == 0 && match($0, /^#+/) && RLENGTH <= lvl { end = NR - 1 }
    END { if (start == 0) exit 1; if (end == 0) end = NR; print start, end }
  ' "$1"
}

corpus=0    # files the sweep opened
surface=0   # files carrying both a toolchain word and a directory word
scanned=0   # files that actually reached part 2
part1=0     # files stating the model and citing no home
part2=0     # sentences restating the retracted outcome outside a home section
exempt=0    # sentences suppressed as retraction records, each printed above

# `-co --exclude-standard` so a NEW, not-yet-committed document is checked too.
# The fourteenth statement this check exists to stop arrived in a file that did
# not exist on main, so a tracked-files-only sweep would have missed the one
# case it was written for.
#
# THE PATHSPEC IS `*.md` AND `Makefile`, AND THAT IS A LIMIT. All thirteen
# sites were prose or the Makefile, so this is the corpus the class actually
# occupied — but a fourteenth could land in a comment in `scripts/**/*.sh`,
# and this sweep would not see it. Adding `*.sh` is not free: it would put the
# recall corpus, which holds all nine retracted sentences as fixtures, and
# this header, which describes the model in order to check it, inside the
# sweep. The extension is owed work with an exemption to design, not a one-word
# change to the line below.
#
# `-z` with a NUL-delimited read, `--deduplicate`, and `--` before the operands.
# The previous `for f in $(git ls-files ...)` word-split on spaces, glob-expanded
# `*`, took a C-quoted non-ASCII path as several paths, and listed an unmerged
# path once per stage — four ways to report clean over a file it never opened,
# on a check whose own design note says it exists for the file that does not
# exist yet, which is the file most likely to be named by a human. The safe read
# loop is the one at `scripts/loop/diff-domains.sh:35`.
while IFS= read -r -d '' f; do
  corpus=$((corpus + 1))

  # A CORPUS FILE THAT IS NOT TEXT IS `cannot run`, NEVER PART OF A VERDICT.
  # The invariant this restores: a file counted in `part 2 read N` had its
  # lines read. Both ways that was false were measured on this script as it
  # stood at 43ee942, by construction rather than by reading.
  #
  #   An .md holding a real restatement and ONE NUL BYTE printed
  #   `clean · corpus 1 · surface 1 · part 2 read 1 file(s) · part 2: 0
  #   flagged`, exit 0, nothing on stderr. grep answers `Binary file X
  #   matches` with no line number; the guard in part 2 correctly refused that
  #   answer and the file was then discarded in silence while still counted as
  #   read — a clean summary with no error at all, which is strictly worse than
  #   the shape that guard's own comment says it refuses.
  #
  #   The UTF-16LE form of the same sentence — what PowerShell 5.1's `>` writes
  #   on this Windows-first repository — never reached that guard. Its bytes are
  #   `g\0o\0 \0v\0e\0t\0`, which matches no alternative in any list here, so it
  #   was dropped at the surface filter below, uncounted and unnamed, and the
  #   run printed clean over it.
  #
  # ONE TEST COVERS BOTH, because UTF-16 of ASCII prose is NUL-bearing by
  # construction, and it has to sit HERE, ahead of the surface filter, or the
  # second case is gone before it is reached.
  #
  # `read -d ''` is the same construct this file already reads `git ls-files -z`
  # with, below: the delimiter is NUL, so it returns 0 only when it found one
  # before EOF. It costs no fork, which matters at 98 files — measured on this
  # tree, the checker runs 7.7s with it against 6.0s without, where the obvious
  # `tr -d '\000' | cmp -s - "$f"` cost 10.3s. It also sidesteps that form's
  # trap: `cmp` needs `--` before the operand, because the recall corpus carries
  # a fixture named `-i.md` and `cmp - -i.md` reads it as `--ignore-initial`.
  if IFS= read -r -d '' _ < "$f"; then
    printf 'd8-root-run-claims: cannot run — %s holds a NUL byte, so its lines cannot be read as text; a UTF-16 file, which is what PowerShell 5.1 writes with `>`, reads exactly like this. Re-save it as UTF-8.\n' "$f" >&2
    exit 2
  fi

  grep -qiE -- "$TOOLCHAIN" "$f" || continue
  grep -qiE -- "$DIRECTORY" "$f" || continue
  surface=$((surface + 1))

  hstart=0; hend=-1
  if pat="$(home_pattern "$f")"; then
    if [ "$pat" = 'PART1-ONLY' ]; then
      :
    elif ! range="$(home_range "$f" "$pat")"; then
      printf 'd8-root-run-claims: cannot run — the home section of %s no longer resolves; the heading it names was renamed or removed\n' "$f" >&2
      exit 2
    else
      hstart="${range%% *}"; hend="${range##* }"
    fi
  else
    # Part 1 applies to non-home files only.
    if ! grep -qE -- "$CITES_HOME" "$f"; then
      echo "part 1 · states the root-run model and cites no home: $f"
      grep -niE -- "$DIRECTORY" "$f" | head -3 | sed 's/^/    /'
      part1=$((part1 + 1))
      continue
    fi
  fi

  scanned=$((scanned + 1))

  # Part 2. Candidate lines are pre-filtered on the outcome list, then split
  # into sentences; a sentence must carry BOTH an outcome and a location to be
  # a restatement. awk is the splitter and nothing else — the matching stays in
  # grep, so one regex dialect decides every verdict in this file.
  while IFS= read -r numbered; do
    n="${numbered%%:*}"; text="${numbered#*:}"
    # `-a` on the pre-filter below is what makes the invariant hold on grep's
    # behaviour rather than depend on it: with it, lines come back numbered
    # whatever grep decides about the file's bytes. GNU grep 3.0 on this host
    # calls a file binary only on a NUL, which the corpus guard above already
    # refuses — but that policy is a property of the grep version and the
    # locale, not of this file, so this guard stays as the last resort.
    # It now EXITS rather than `continue`s. Continuing was the defect: it took
    # grep's `Binary file X matches`, which carries no line number, dropped the
    # file, and left it counted in `part 2 read N` — an error on stderr and a
    # clean summary on stdout is the shape the old comment here said it
    # refused, and silence with a clean summary is worse than that shape.
    case "$n" in ''|*[!0-9]*)
      printf 'd8-root-run-claims: cannot run — grep returned a line without a number for %s, so this sweep cannot read it: %s\n' "$f" "$numbered" >&2
      exit 2 ;;
    esac
    [ "$n" -ge "$hstart" ] && [ "$n" -le "$hend" ] && continue
    while IFS= read -r s; do
      printf '%s' "$s" | grep -qiE -- "$OUTCOME"  || continue
      printf '%s' "$s" | grep -qiE -- "$LOCATION" || continue
      if printf '%s' "$s" | grep -qiE -- "$RETRACTS"; then
        echo "exempt · $f:$n records the retraction rather than restating it — read it; it is not a pass:"
        printf '    %s\n' "$s"
        exempt=$((exempt + 1))
        continue
      fi
      echo "part 2 · restates the retracted root-run outcome: $f:$n"
      printf '    %s\n' "$s"
      part2=$((part2 + 1))
    done < <(printf '%s\n' "$text" | awk '{ n = split($0, a, /[.!?] +/); for (i = 1; i <= n; i++) print a[i] }')
  done < <(grep -naiE -- "$OUTCOME" "$f")
done < <(git ls-files -z -co --exclude-standard --deduplicate -- '*.md' 'Makefile')

# The other half of the same rule: "the check ran and found nothing" and "the
# check did not run" must not print the same thing. An empty corpus, an empty
# surface, or a part 2 that no file reached is exit 2 — a clean verdict from a
# sweep that opened nothing is the exact defect class this file exists to find.
if [ "$corpus" -eq 0 ]; then
  echo "d8-root-run-claims: cannot run — the corpus is empty; no .md file and no Makefile was listed" >&2
  exit 2
fi
if [ "$surface" -eq 0 ]; then
  echo "d8-root-run-claims: cannot run — $corpus file(s) opened, none carried both a toolchain word and a directory word" >&2
  exit 2
fi
if [ "$scanned" -eq 0 ] && [ "$part1" -eq 0 ]; then
  echo "d8-root-run-claims: cannot run — $surface file(s) on the surface, none reached part 2" >&2
  exit 2
fi

fail=0
[ "$part1" -eq 0 ] && [ "$part2" -eq 0 ] || fail=1

if [ "$fail" -eq 0 ]; then
  printf 'd8-root-run-claims: clean · corpus %s file(s) · surface %s · part 2 read %s file(s) · part 1: 0 · part 2: 0 flagged, %s exempted\n' \
    "$corpus" "$surface" "$scanned" "$exempt"
else
  printf 'd8-root-run-claims: FAIL · corpus %s file(s) · surface %s · part 2 read %s file(s) · part 1: %s · part 2: %s flagged, %s exempted\n' \
    "$corpus" "$surface" "$scanned" "$part1" "$part2" "$exempt"
  echo "    the homes are Architecture.md § D8, local-gates § Backend (Go), and the Makefile — cite one; do not restate it"
fi

exit "$fail"
