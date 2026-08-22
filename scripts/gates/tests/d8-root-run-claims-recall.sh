#!/usr/bin/env bash
# d8-root-run-claims-recall.sh — the recall of d8-root-run-claims.sh, as a command. (#118)
#
# #118 DoD 4 asks for a check that FAILS on a fourteenth restatement. A check
# can only be held to that if its recall is a number someone other than its
# author can recompute, so this corpus exists to make it one. It is the whole
# of the evidence behind the recall rows in that script's header — and since
# cycle 3 that is enforced rather than said: the last block of this file reads
# those rows back and fails when one of them is not the number this run just
# measured. "If they ever disagree, this file is the one that ran" used to be
# an instruction to a reader, which is what a second statement of a number
# always degrades into. It is now a check.
#
# THE CORPUS IS THE HISTORY. Every sentence in the `pre` fixture is quoted
# verbatim from the pre-repair tree at 6fbf7de, and every sentence in the
# `post` fixture from the repaired tree at 71d629f, each at the path it lives
# at. FIXTURE 5 does the same for the Makefile across its own repair,
# c8088fa. `git show 6fbf7de:.claude/agents/go-quality-critic.md` recovers the
# paragraph any one of them was cut from. They were not written for this test
# and they are not paraphrases of the class — they ARE the class.
#
# YES, THAT PUTS NINE COPIES OF THE RETRACTED MODEL BACK IN THE TREE, and no,
# it is not the defect recurring. Each one is a quotation under test: it is
# heredoc input written to a throwaway directory, attributed to the commit it
# came from, and asserted to be WRONG. A test for a bug contains the bug. The
# checker sweeps `*.md` and the `Makefile` and does not read this file, which
# is the boundary that lets the corpus hold them; `scripts/**/*.sh` staying
# outside that sweep is recorded as a limit in the checker's own header.
#
# BOTH DIRECTIONS ARE ASSERTED, and the second matters as much as the first: a
# check that flags the eight restatements is worthless if it also flags the
# eight repairs, because the repairs are the remedy prescribed by
# `local-gates § Documentation gates` gate 2. `pre` must go red. `post` must
# go green.
#
# THE HELD-OUT SET IS PRINTED, NOT ASSERTED. Recall over `pre` is in-sample by
# construction — the vocabulary was drawn from those eight. So the corpus also
# carries restatements invented for it, in wordings taken from none of the
# thirteen, and reports how many of those are caught without failing on the
# number. That number is the honest one, it is expected to be low, and it is
# printed on every run so it cannot be quietly tuned away. Do not add a phrase
# to the checker in order to move it: move a held-out wording into `pre` only
# when the tree actually acquires it.
#
# THE HEADER CHECK AT THE END DOES NOT CHANGE THAT, and the distinction is the
# whole reason it is written the way it is. It asserts that the checker's
# header REPEATS this number, never that the number holds any value. Adding a
# phrase to move it still only moves it, in both places; what is no longer
# possible is moving it in one.
#
# Hermetic: throwaway git repositories under mktemp, fixture files as the only
# input, no network, no repository state, nothing written outside TMP.
# Usage: scripts/gates/tests/d8-root-run-claims-recall.sh    Exit: 0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$DIR/../d8-root-run-claims.sh"
# `mktemp -d` is guarded because this file's own promise is that nothing is
# written outside TMP, and only `set -u` is in force — no `set -e`. Unguarded, a
# failing mktemp leaves TMP empty and every `mkdir -p "$TMP/$1"` below becomes
# `mkdir -p /pre` at the filesystem root, while the EXIT trap becomes
# `rm -rf ""`. The trap is armed only once TMP is known good.
TMP="$(mktemp -d)" || TMP=''
[ -n "$TMP" ] && [ -d "$TMP" ] || {
  printf 'FAIL  mktemp -d gave no usable directory; refusing to run rather than write outside one\n'; exit 1; }
trap 'rm -rf "$TMP"' EXIT

[ -f "$SCRIPT" ] || { printf 'FAIL  the checker is not at %s\n' "$SCRIPT"; exit 1; }

fails=0
OUT=''; RC=0

mkrepo() { mkdir -p "$TMP/$1"; git -C "$TMP/$1" init -q; }

# A FAILED `cd` MUST NOT ARRIVE AS THE CHECKER'S OWN rc. `cd` returns 1, and 1
# is exactly what the `pre` and `edge` fixtures assert — so a fixture directory
# that could not be entered used to read as the expected result, on the two
# assertions that carry the most weight in this file. rc 99 is outside the
# checker's range (0, 1, 2) and OUT says which fixture.
run() { # run <fixture> [checker]
  OUT="$(cd "$TMP/$1" 2>/dev/null || { printf 'could not enter the fixture directory %s\n' "$TMP/$1"; exit 99; }
         bash "${2:-$SCRIPT}" 2>&1)"
  RC=$?
}

pass() { printf 'PASS  %s\n' "$1"; }
bad()  { printf 'FAIL  %s — %s\n' "$1" "$2"; fails=$((fails + 1)); }

check_rc() { # check_rc <label> <expected-rc>
  [ "$RC" = "$2" ] && pass "$1" && return 0
  bad "$1" "expected rc $2, got rc $RC: $(printf '%s' "$OUT" | tr '\n' '|' | cut -c1-160)"
}
check_has() { # check_has <label> <substring>
  case "$OUT" in *"$2"*) pass "$1" ;; *) bad "$1" "output does not contain: $2" ;; esac
}
check_lacks() { # check_lacks <label> <substring>
  case "$OUT" in *"$2"*) bad "$1" "output must not contain: $2" ;; *) pass "$1" ;; esac
}

# ---------------------------------------------------------------------------
# THE NINE HISTORICAL SITES REGISTER THEMSELVES WHERE THEY ARE ASSERTED, so the
# in-sample recall rows in the checker's header can be checked against what
# this run measured instead of being kept by hand. They were kept by hand until
# cycle 3, and it drifted exactly as a second statement does: measured at
# 43ee942, adding one OUTCOME alternative moved the held-out count printed at
# the bottom of this file to `1 of 4` while the header row still read
# `0 OF 4 CAUGHT`, and all 37 assertions of the day passed.
#
# THIS NINE IS: the 8 dependent sites at 6fbf7de plus `Makefile:16-23` at
# c8088fa^. It is NOT the nine FIXTURE 1 asserts a flag COUNT over, which swaps
# the Makefile for the synthetic `Architecture.md` dependent-section case. Two
# disjoint nines; each is named where it is asserted.
#
# Part 1's verdict is per FILE, not per sentence, so a site counts as caught by
# part 1 when part 1 named the file it lives in.
sites_total=0; p1_caught=0; p2_caught=0
site_seen() { # site_seen <file>
  sites_total=$((sites_total + 1))
  case "$OUT" in *"cites no home: $1"*) p1_caught=$((p1_caught + 1)) ;; esac
}
site_flags() { # site_flags <label> <file> <substring> — part 2 MUST flag this site
  site_seen "$2"
  case "$OUT" in
    *"$3"*) p2_caught=$((p2_caught + 1)); pass "$1" ;;
    *)      bad "$1" "output does not contain: $3" ;;
  esac
}
site_misses() { # site_misses <label> <file> <substring> — part 2 does NOT flag it
  site_seen "$2"
  case "$OUT" in
    *"$3"*) p2_caught=$((p2_caught + 1))
            bad "$1" "the measured miss is no longer a miss — intended or not, the header's in-sample row moves with it" ;;
    *)      pass "$1" ;;
  esac
}

# The same nine in repaired form, counted the same way and printing nothing:
# that no repair is flagged is asserted collectively per fixture, and these
# calls exist only so the header's third row is a measured number too.
repaired_total=0; repaired_flagged=0
repaired() { # repaired <file>
  repaired_total=$((repaired_total + 1))
  case "$OUT" in *"restates the retracted root-run outcome: $1"*) repaired_flagged=$((repaired_flagged + 1)) ;; esac
}

# The checker's header is read back and checked for AGREEMENT with the numbers
# above — never for a value. A row that disagrees is a failure whichever of the
# two moved; the fix is to make the header state what was measured, never to
# make the measurement state what the header wants.
header_states() { # header_states <label> <row-key> <expected-text>
  local row
  row="$(grep -iF -m1 -- "$2" "$SCRIPT")"
  if [ -z "$row" ]; then bad "$1" "the checker has no header row carrying: $2"; return; fi
  if printf '%s' "$row" | grep -qiF -- "$3"; then pass "$1"
  else bad "$1" "this run measured '$3'; the header row reads:$(printf '%s' "$row" | sed 's/^#//')"; fi
}

# ---------------------------------------------------------------------------
# FIXTURE 1 — `pre`: the eight sites as they stood at 6fbf7de. Must go red.
#
# Each file carries the toolchain word and the directory word its real
# counterpart carried, because those are what put a file on the surface at all,
# and each cites a home, because all thirteen sites did — that is the whole
# reason part 1 caught none of them.
# ---------------------------------------------------------------------------
mkrepo pre
mkdir -p "$TMP/pre/.claude/agents" "$TMP/pre/.claude/skills/local-gates" \
         "$TMP/pre/.claude/skills/agent-handoffs" "$TMP/pre/docs"

# Sites 1 and 2 — go-quality-critic.md:41 and :48. Site 1 is the one the
# advisory phrase list missed: its location cue is "from the wrong place", two
# sentences away from the "repo root" that puts the file on the surface.
cat > "$TMP/pre/.claude/agents/go-quality-critic.md" <<'MD'
Format, vet, lint, and build are the author's gates, and their commands live in `local-gates § Backend (Go)`.

```powershell
cd "$(git rev-parse --show-toplevel)/service"   # the module, not the repo root
staticcheck ./...
```

**Confirm them; do not retype them.** The PR body must carry the directory they ran in — a report without one is not evidence that anything was compiled, and re-running the list yourself from the wrong place reproduces the same vacuous pass one level further down.

Its `cd` is load-bearing for the same reason every Go command's is: the module is not at the repository root, per `Architecture.md § D8`, so run from there it inspects zero files and reports zero faults, which is indistinguishable from clean.
MD

# Site 3 — linus-quality-critic.md:98.
cat > "$TMP/pre/.claude/agents/linus-quality-critic.md" <<'MD'
Then run the one check the gate cannot run for you, per `local-gates § Backend (Go)`:

```powershell
go test ./... -run <relevant> -count=1
```

The `cd` matters here for the same reason it matters there: run from the repository root, `-run` selects from no packages and passes.
MD

# Site 4 — linus-security-critic.md:104.
cat > "$TMP/pre/.claude/agents/linus-security-critic.md" <<'MD'
The gates are in `local-gates § Backend (Go)`; these two are this bench's own instruments.

**The two directories differ, and that is not a slip.** `govulncheck` is a Go tool: it resolves against the module and, run from the root, finds no packages and reports no vulnerabilities — the same silent pass the Go gate has, arriving in a security review, which is where it costs most.
MD

# Site 5 — progressive-results-specialist.md:59.
cat > "$TMP/pre/.claude/agents/progressive-results-specialist.md" <<'MD'
**4 — Gates.** Run the code gates across the whole diff — `local-gates § Backend (Go)`. Take the commands and their working directories from the skill; a Go gate run from the repository root passes against nothing.
MD

# Site 6 — validation-agent.md:45.
cat > "$TMP/pre/.claude/agents/validation-agent.md" <<'MD'
Run the backend gates — go vet, lint, tests, build — per `local-gates § Backend (Go)`.

**This step used to `cd` to the repository root, and that was the exact wrong place.** The Go module lives in `service/`, per `Architecture.md § D8`, so from the root every one of these commands resolves against nothing, exits zero, and prints what a clean tree prints.
MD

# Site 7 — agent-handoffs/SKILL.md:306.
cat > "$TMP/pre/.claude/skills/agent-handoffs/SKILL.md" <<'MD'
- **The gates that could pass having read nothing.** `Architecture.md § D8` puts the Go module in `service/`, and gofmt is the quiet one. The gate block carried no working directory, so from the repository root all five commands resolve against nothing, exit zero, and print exactly what a clean module prints — and the prescribed report line was character-for-character what that vacuous run produces.
MD

# Site 8 — local-gates/SKILL.md:97, and the reason the home exemption is per
# SECTION. This file is the home of `local-gates § Backend (Go)`; law 1 below
# it was a DEPENDENT of `Architecture.md § D8` living in the same file. A
# per-file exemption saw a home and skipped the file whole, which is how this
# site stayed invisible. The sentence inside the home section states the model
# and must NOT be flagged.
cat > "$TMP/pre/.claude/skills/local-gates/SKILL.md" <<'MD'
## Code gates

### Backend (Go)

Run `go vet ./...` from `service/`. Measured from the repository root, four of the five are loud and gofmt exits zero having read the tree by accident of layout.

### Frontend (Vite + React)

Every command runs from `web/`.

## The law

1. **The code line names its directory** for exactly the same reason: five passes over an empty root are character-for-character the five passes over a clean module, so a report that omits where it ran is not evidence that anything was compiled.
MD

# Not one of the eight: the same section rule asserted in both directions on
# the other markdown home. Inside `Architecture.md § D8` the model is stated
# because that is D8's job; the paragraph under a later heading is a dependent
# and must be flagged.
cat > "$TMP/pre/docs/Architecture.md" <<'MD'
### D8 — The Go service in `service/`, as one peer directory among several

**What it costs.** A go build run from the repository root passes against nothing, which is this section's own claim to make.

## Runtime topology

The gates run from the repository root, where go vet resolves against nothing and exits zero.
MD

run pre
check_rc   "pre · the eight restatements fail the build"                    1
site_flags "pre · site 1  go-quality-critic:41  (the phrase list's miss)"   .claude/agents/go-quality-critic.md              "vacuous pass one level further down"
site_flags "pre · site 2  go-quality-critic:48"                             .claude/agents/go-quality-critic.md              "inspects zero files"
site_flags "pre · site 3  linus-quality-critic:98"                          .claude/agents/linus-quality-critic.md           "selects from no packages"
site_flags "pre · site 4  linus-security-critic:104"                        .claude/agents/linus-security-critic.md          "reports no vulnerabilities"
site_flags "pre · site 5  progressive-results-specialist:59"                .claude/agents/progressive-results-specialist.md "passes against nothing"
site_flags "pre · site 6  validation-agent:45"                              .claude/agents/validation-agent.md               "prints what a clean tree prints"
site_flags "pre · site 7  agent-handoffs:306"                               .claude/skills/agent-handoffs/SKILL.md           "print exactly what a clean module prints"
site_flags "pre · site 8  local-gates:97  (home file, dependent section)"   .claude/skills/local-gates/SKILL.md              "character-for-character the five passes"
check_has  "pre · a dependent section of a home file is still checked"      "docs/Architecture.md"
check_lacks "pre · a home section states the model and is not flagged"      "which is this section's own claim to make"
# THIS FIXTURE'S NINE IS NOT THE HEADER'S NINE, and the two sit adjacent enough
# to be read as one. Here it is the 8 dependent sites above plus the synthetic
# `Architecture.md` dependent-section case immediately below them — and it
# EXCLUDES the Makefile, which has its own fixture in FIXTURE 5. All nine flag,
# which is why the count asserted is 9 rather than the header's 8. The header's
# nine swaps the synthetic case for `Makefile:16-23`, the one part 2 misses;
# that is the whole of the difference between `9 flagged` here and `8 OF 9` there.
check_has  "pre · part 1 catches none of this fixture's nine; part 2 flags all nine" \
                                                                           "part 1: 0 · part 2: 9 flagged"

# ---------------------------------------------------------------------------
# FIXTURE 2 — `post`: the same eight sites as repaired at 71d629f, plus the
# retraction record. Must go green, and must say out loud what it suppressed.
# ---------------------------------------------------------------------------
mkrepo post
mkdir -p "$TMP/post/.claude/agents" "$TMP/post/.claude/skills/agent-handoffs" \
         "$TMP/post/.claude/skills/local-gates"

cat > "$TMP/post/.claude/agents/go-quality-critic.md" <<'MD'
```powershell
cd "$(git rev-parse --show-toplevel)/service"   # the module, not the repo root
staticcheck ./...
```

The PR body must carry the directory they ran in — a report without one is not evidence that anything was compiled, and re-running the list yourself from the wrong place measures the wrong tree one level further down, per `Architecture.md § D8`.

Its `cd` is load-bearing for the same reason every Go command's is: the module is not at the repository root, per `Architecture.md § D8`, so run from there it inspects none of this repository's code.
MD

cat > "$TMP/post/.claude/agents/linus-quality-critic.md" <<'MD'
```powershell
go test ./... -run <relevant> -count=1
```

The `cd` matters here for the same reason it matters there: the module is not at the repository root, per `Architecture.md § D8`, so from there `-run` selects from none of this repository's packages — the directory, not the pattern, decides which tests were available to be selected at all.
MD

cat > "$TMP/post/.claude/agents/linus-security-critic.md" <<'MD'
`govulncheck` is a Go tool: it resolves against the module, and there is none at the repository root, per `Architecture.md § D8`, so a run from there scans none of this service's dependencies. **Report the directory with the result.**
MD

cat > "$TMP/post/.claude/agents/progressive-results-specialist.md" <<'MD'
**4 — Gates.** Run the code gates — `local-gates § Backend (Go)`. Take the commands and their working directories from the skill; the directory is what decides which tree a Go gate measured, per `Architecture.md § D8`.
MD

cat > "$TMP/post/.claude/agents/validation-agent.md" <<'MD'
Run the backend gates — go vet, lint, tests, build — per `local-gates § Backend (Go)`.

The Go module lives in `service/`, per `Architecture.md § D8`, so the directory these commands ran in is what decides which tree they measured, and the root is not this service — `local-gates § Backend (Go)` records what each gate reports from there.
MD

cat > "$TMP/post/.claude/skills/agent-handoffs/SKILL.md" <<'MD'
- **The gates whose report could not say where they ran.** `Architecture.md § D8` puts the Go module in `service/`, and gofmt is the quiet one. The gate block carried no working directory, and neither did the report line it prescribed — so a reported result was identical whether the commands had been run over the module or over a root that holds none, and no reader could tell which. **The mechanism first recorded here — all five commands passing vacuously over an empty root — was itself measured false and retracted on 22 August 2026; `Architecture.md § D8` carries the retraction and `local-gates § Backend (Go)` the measurement.**
MD

# SITE 8's MUST-NOT-FLAG HALF, quoted from `local-gates/SKILL.md:97` at
# 71d629f. It was asserted in the `pre` direction only until cycle 3, and it is
# the site that motivates the subtlest rule in the checker: the home exemption
# is per SECTION, so `The law` below is read while `Backend (Go)` above it is
# not. Its repair therefore has to pass through part 2 rather than around it,
# and nothing said so. The `Backend (Go)` heading is not padding — without it
# the home section does not resolve and the checker exits 2.
cat > "$TMP/post/.claude/skills/local-gates/SKILL.md" <<'MD'
### Backend (Go)

**Run the backend gates through the Makefile, from the repository root**, where the Makefile is.

## The law

1. **The code line names its directory** for exactly the same reason: the directory a Go command ran in is what decides which tree it measured, per `Architecture.md § D8`, and `§ Backend (Go)` above records what each of the five reports from the wrong one — so a report that omits where it ran is not evidence that anything was compiled.
MD

run post
check_rc    "post · the eight repaired dependents pass the build"         0
check_lacks "post · no repair is flagged as a restatement"                "part 2 · restates"
check_has   "post · the retraction record is exempted, and printed"       "exempt · "
check_has   "post · the suppression is counted where a reader sees it"    "part 2: 0 flagged, 1 exempted"
# Reach, for the same reason FIXTURE 5 asserts it: a clean summary names no
# file, so the read count is the only line proving site 8's repair went THROUGH
# part 2 rather than around it as a home file. This fixture read 5 before the
# repair above was added and reads 6 with it — measured, not assumed. (The
# seventh corpus file, `validation-agent.md`, carries no file-level directory
# word in its repaired form and never reaches the surface; that is why 6 and
# not 7.)
check_has   "post · site 8's repair was read, not skipped as a home file" "part 2 read 6 file(s)"

# The eight repaired dependents, registered for the header's third row; the
# ninth repaired site is the Makefile's, in FIXTURE 5. Sites 1 and 2 share a
# file and are two sites, so that file is registered twice.
repaired .claude/agents/go-quality-critic.md
repaired .claude/agents/go-quality-critic.md
repaired .claude/agents/linus-quality-critic.md
repaired .claude/agents/linus-security-critic.md
repaired .claude/agents/progressive-results-specialist.md
repaired .claude/agents/validation-agent.md
repaired .claude/skills/agent-handoffs/SKILL.md
repaired .claude/skills/local-gates/SKILL.md

# ---------------------------------------------------------------------------
# FIXTURE 3 — `edge`: the sweep's own failure modes, each of which reported
# clean over something it never looked at.
# ---------------------------------------------------------------------------
mkrepo edge
mkdir -p "$TMP/edge/docs"

# A path with a space and a path beginning with a dash. Both were invisible to
# the previous `for f in $(git ls-files ...)`, and the design note says this
# check exists for the file that does not exist yet — the file a human names.
cat > "$TMP/edge/docs/gate notes.md" <<'MD'
Per `Architecture.md § D8`, go vet run from the repository root passes against nothing.
MD
cat > "$TMP/edge/-i.md" <<'MD'
Per `Architecture.md § D8`, go build from the repository root exits zero having read nothing.
MD

# The retraction exemption is per SENTENCE. A paragraph may record the
# retraction in one sentence and restate the model in another; the restatement
# is still a restatement.
cat > "$TMP/edge/docs/mixed.md" <<'MD'
Per `Architecture.md § D8`, go vet run from the repository root finds no packages and reports success. That claim was measured false and retracted on 22 August 2026.
MD

# The Makefile is a home of the directory encoding and of nothing part 2
# matches, so every line of it is read. This case asserted the opposite until
# cycle 2 — and asserted it over a comment saying "from this root", which is in
# no phrase list, so the file never reached the surface and the exemption it
# claimed to prove was never exercised. A vacuous assertion about the one file
# the sweep skipped whole.
cat > "$TMP/edge/Makefile" <<'MD'
# Run from the repository root, go vet resolves against nothing and exits zero.
gates:
	cd service && go vet ./...
MD

run edge
check_rc   "edge · a path with a space is opened, not skipped"        1
check_has  "edge · ... and named"                                     "docs/gate notes.md"
check_has  "edge · a path beginning with a dash is opened"            "-i.md"
check_has  "edge · a restatement beside a retraction still flags"     "finds no packages and reports success"
check_has  "edge · a restatement in the Makefile is read, not skipped" "restates the retracted root-run outcome: Makefile:1"

# ---------------------------------------------------------------------------
# FIXTURE 4 — the instrument's own failure modes. A checker that cannot run
# must never print the sentence a clean tree prints.
# ---------------------------------------------------------------------------
mkrepo empty
run empty
check_rc    "empty · an empty corpus is 'cannot run', not 'clean'"    2
check_lacks "empty · ... and never says clean"                        "clean ·"

mkrepo nosurface
printf 'This document mentions no toolchain and no directory.\n' > "$TMP/nosurface/docs.md"
run nosurface
check_rc    "no surface · nothing reached a part, so it is exit 2"    2
check_lacks "no surface · ... and never says clean"                   "clean ·"

# A CORPUS FILE THAT IS NOT TEXT — both shapes, because at 43ee942 they failed
# in two different places, and each fixture below is one of them. What each
# printed then, and why one guard reaches both, is argued once at the guard
# itself, in the checker's sweep loop; what a reader will find there is that a
# NUL-bearing file was counted in `part 2 read N` without its lines ever being
# read. These two pin it.
#
# The first is that file: a REAL restatement plus one NUL byte, which grep
# answers with `Binary file X matches` and no line number.
mkrepo nulbyte
printf 'Per `Architecture.md \302\247 D8`, go vet run from the repository root passes against nothing.\n\0' \
  > "$TMP/nulbyte/doc.md"
run nulbyte
check_rc    "not text · a NUL byte in a corpus file is 'cannot run'"  2
check_has   "not text · ... and the file is named, with the cause"    "holds a NUL byte"
check_lacks "not text · ... and it never reports a clean tree"        "clean ·"

# The second is the UTF-16LE form of the same sentence — what PowerShell 5.1's
# `>` produces on this Windows-first repository — which failed EARLIER and more
# quietly, at the surface filter rather than in part 2. It is written byte for
# byte rather than through iconv, which this corpus does not otherwise depend
# on. `ok.md` is here so the drop under test cannot be mistaken for the
# empty-surface exit 2: without a file that does reach the surface, this
# fixture would go green for the wrong reason.
mkrepo utf16
printf 'Per `Architecture.md \302\247 D8`, go vet resolves against the directory it is run from; the repository root is not this service.\n' \
  > "$TMP/utf16/ok.md"
: > "$TMP/utf16/psh.md"
u16='go vet run from the repository root passes against nothing.'
u16i=0
while [ "$u16i" -lt "${#u16}" ]; do
  printf '%s\0' "${u16:$u16i:1}" >> "$TMP/utf16/psh.md"
  u16i=$((u16i + 1))
done
run utf16
check_rc    "not text · a UTF-16 file is 'cannot run', not a silent drop"  2
check_has   "not text · ... and it is named rather than skipped"          "psh.md"
check_has   "not text · ... and the message says what wrote it"           "PowerShell"
check_lacks "not text · ... and it never reports a clean tree"            "clean ·"

# A control byte inside one alternative: it compiles, matches nothing, and
# disables that alternative in silence. This file has shipped that byte once.
BS="$(printf '\b')"
sed "s/'vacuous'/'vac${BS}uous'/" "$SCRIPT" > "$TMP/mutant-byte.sh"
run post "$TMP/mutant-byte.sh"
check_rc    "mutation · a control byte in a phrase is 'cannot run'"   2
check_has   "mutation · ... and the broken alternative is named"      "carries a control byte"
check_lacks "mutation · ... and it never reports a clean tree"        "clean ·"

# An alternative that does not compile: an unbalanced paren used to print to
# stderr and then report the tree clean and exit 0.
sed "s/'vacuous'/'vacuous('/" "$SCRIPT" > "$TMP/mutant-paren.sh"
run post "$TMP/mutant-paren.sh"
check_rc    "mutation · an uncompilable phrase is 'cannot run'"       2
check_has   "mutation · ... and the broken alternative is named"      "does not compile"
check_lacks "mutation · ... and it never reports a clean tree"        "clean ·"

# A home whose home section was renamed: the check can no longer tell the home
# from its dependents, so it refuses rather than guessing in either direction.
mkrepo renamed
mkdir -p "$TMP/renamed/.claude/skills/local-gates"
cat > "$TMP/renamed/.claude/skills/local-gates/SKILL.md" <<'MD'
### The Go gates

Run `go vet ./...` from `service/`, never from the repository root.
MD
run renamed
check_rc   "renamed · an unlocatable home section is 'cannot run'"    2
check_has  "renamed · ... and says which file"                        "local-gates/SKILL.md"

# ---------------------------------------------------------------------------
# FIXTURE 5 — the Makefile, both directions, both verbatim.
#
# The two texts below are why the exemption this file used to assert mattered:
# the thirteenth restatement of the model was at `Makefile:16` until `c8088fa`
# repaired it, and the anchored citation that replaced it is at `Makefile:16`
# today. A check exempting the file sees neither. What the Makefile is and is
# not a home OF is argued once, in the checker's header.
#
# Reach is the thing under test here, so `part 2 read 1 file(s)` is asserted in
# both directions: it is the only line proving the sweep opened the Makefile
# rather than passing over it. Each fixture is the header block verbatim plus
# the `vet` recipe, which is what carries the toolchain word.
# ---------------------------------------------------------------------------

# The repair, quoted from `Makefile:11-22` at b651862. It cites the home and
# says in a clause what a reader will find there — the remedy
# `local-gates § Documentation gates` gate 2 prescribes. Extending reach to
# this file must not condemn it, and this is the assertion that says so.
mkrepo mk-post
cat > "$TMP/mk-post/Makefile" <<'MD'
# `Architecture.md § D8` puts the Go module in `service/` and leaves no module
# at the repository root. Every Go command therefore resolves against the
# directory it is run from, and that directory — not the tree — is what decides
# what a gate measured.
#
# What each of the five reports from that wrong directory is NOT restated here.
# `local-gates § Backend (Go)` is its one home and carries it per gate, with
# the host and the date the measurement was taken on — including which gates
# are loud without being discriminating, which is the half a recipe can get
# wrong. A paragraph of it here would be one more statement of the model #118
# exists to collapse, in this file's own words, going stale on the next
# measurement with no diff here to notice.
vet:
	cd service && go vet ./...
MD
run mk-post
check_rc   "mk-post · the anchored citation is not condemned"   0
check_has  "mk-post · ... and part 2 opened the Makefile to say so" "part 2 read 1 file(s)"
repaired Makefile   # the ninth repaired site

# The thirteenth site is `Makefile:16-23` at c8088fa^ — the per-gate
# restatement, and the lines the checker header names as the one it misses.
# The specimen below quotes `Makefile:11-23`, and the extra paragraph is not
# padding: `Makefile:11-15` is the module-path premise that
# `Architecture.md § D8` makes the Makefile's own legitimate content — the
# claim behind its `PART1-ONLY` entry — and "repository root" in it is the
# ONLY file-level directory word either paragraph carries. Quoting 16-23
# alone was measured: the sweep exits 2, "none carried both a toolchain word
# and a directory word", and never opens the file. That would turn what is
# asserted here from an in-reach miss into an out-of-reach one, so the
# specimen keeps the premise and the site stays 16-23.
#
# IT IS NOT CAUGHT, and the miss is asserted rather than left unsaid: the
# alternative is a corpus going green while the one historical restatement in
# the Makefile's own format walks through it. Why it is missed, and what the
# only variant catching it costs on the repaired tree, are measured and argued
# once — in the checker's own header, under THE ONE IN-SAMPLE SITE IT MISSES.
# That header names TWO blockers, each alone sufficient, and this comment does
# not repeat which: a second copy of the cause here is what went stale last
# cycle, when the header's single stated cause was relayed as established
# rather than measured. Read it there.
#
# If a later change catches this text, this line fails. That is intended: the
# recall numbers in that header move with it, and they should not move quietly.
mkrepo mk-pre
cat > "$TMP/mk-pre/Makefile" <<'MD'
# `Architecture.md § D8` puts the Go module in `service/` and leaves no module
# at the repository root. Every Go command therefore resolves against the
# directory it is run from, and that directory — not the tree — is what decides
# what a gate measured.
#
# Measured from this root, four of the five are loud: vet, lint and build each
# fail to resolve a module, and test fails on the absent-cgo host reason that is
# identical from either directory, so on this host it discriminates nothing.
# Only gofmt exits 0, and its output is identical to a clean run — not because
# it read nothing, but because it walks the filesystem instead of resolving a
# module, and every Go file in this repository sits inside the module
# directory. It reads the same tree by accident of layout, and reads more the
# moment a .go file lands outside it.
vet:
	cd service && go vet ./...
MD
run mk-pre
check_rc    "mk-pre · MEASURED MISS - the thirteenth site does not flag"  0
site_misses "mk-pre · ... and the miss is the ninth site of the header's nine" \
            Makefile "restates the retracted root-run outcome: Makefile"
check_has   "mk-pre · ... and it is a miss inside reach, not outside it"  "part 2 read 1 file(s)"

printf '\nthe thirteenth site (Makefile, c8088fa^) is in reach and NOT caught.\nFIXTURE 5 carries why; the checker header carries what the only variant\ncatching it costs.\n'

# ---------------------------------------------------------------------------
# HELD OUT — reported, never asserted. See the header: these are wordings
# invented for this corpus, drawn from none of the thirteen sites. The number
# below is what this check is worth against a restatement nobody has written
# yet, and it is expected to be low. Do not close the gap by adding these
# phrases to the checker; that would only make this number lie.
# ---------------------------------------------------------------------------
mkrepo heldout
held_caught=0
held_total=0
held_missed=''
try_held() { # try_held <label> <sentence>
  held_total=$((held_total + 1))
  rm -f "$TMP/heldout/h.md"
  printf 'Per `Architecture.md § D8`, note the following. %s\n' "$2" > "$TMP/heldout/h.md"
  run heldout
  if [ "$RC" = 1 ]; then
    held_caught=$((held_caught + 1))
  else
    held_missed="$held_missed
    missed: $1"
  fi
}

try_held "compiles nothing / still exits 0" \
  'Run from the top of the repository, go build ./... compiles nothing and still exits 0.'
try_held "green build over an empty tree" \
  'A CI job that invokes the Go toolchain from the wrong directory will report a green build over an empty tree.'
try_held "quietly succeeds having examined nothing" \
  'Because there is no module at the repository root, go vet ./... there quietly succeeds having examined nothing.'
try_held "sees a directory, not a module" \
  'From the repository root the go command sees a directory rather than a module, and its silence means nothing.'

printf '\nheld-out wordings caught: %s of %s%s\n' "$held_caught" "$held_total" "$held_missed"
printf 'in-sample recall over the nine historical sites is asserted above, the\nninth as a miss; this number is not asserted.\n'

# ---------------------------------------------------------------------------
# THE CHECKER'S HEADER, CHECKED AGAINST THIS RUN — for AGREEMENT, never for a
# value. Those four rows were a hand-kept second statement of everything above,
# which is the gate-2 duplication shape #118 exists to collapse, sitting inside
# the instrument built to enforce it. Measured: one added OUTCOME alternative
# moved `held-out wordings caught` to 1 of 4 while the header row still read
# `0 OF 4 CAUGHT`, all 37 assertions passed and nothing failed.
#
# WHAT IS ASSERTED IS ONLY THAT THE ROW REPEATS THE NUMBER, so nothing here
# makes held-out recall a target: it stays printed rather than asserted at a
# value, and adding a phrase to the checker to move it still only makes the
# number move. What is now impossible is moving it and leaving the header
# saying otherwise.
#
# A failure here is fixed by making the header state what was measured. It is
# never fixed by making the measurement state what the header wants.
# ---------------------------------------------------------------------------
header_states "header · part 1's in-sample row states this run's count" \
  'part 1, exit status'                        "$p1_caught of $sites_total sites"
header_states "header · part 2's in-sample row states this run's count" \
  'part 2, exit status'                        "$p2_caught of $sites_total sites"
header_states "header · the repaired row states this run's denominator" \
  'repaired sites, which must NOT flag'        "the $repaired_total repaired sites"
header_states "header · the repaired row states this run's count" \
  'repaired sites, which must NOT flag'        "$repaired_flagged of $repaired_total flagged"
header_states "header · the held-out row states the number just printed" \
  'held-out wordings, invented for the corpus' "$held_caught of $held_total caught"

[ "$fails" -eq 0 ] || { printf '\n%s check(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
