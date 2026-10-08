#!/usr/bin/env bash
# size-ratchet.sh — has an agent definition or a skill grown past its baseline? (#207)
#
# #205 step 1.2: these files may shrink freely and grow only through a visible
# edit to the baseline beside this script, so growth is reviewed like any
# other change.
#
# THE MEASURED SET is every `.claude/agents/*.md` and `.claude/skills/*/SKILL.md`
# under the root this script resolves from its own path, never from the
# caller's directory. It is a glob and not `git ls-files`, so a file nobody
# added is measured too.
#
# BYTES, NOT COUNTING CARRIAGE RETURNS. Under plain `wc -c` a CRLF checkout
# (core.autocrlf=true, the reference host) reads one byte per line more than
# an LF checkout of the same commit, so a baseline taken on either would fail
# or flatter the other. Without the CRs the count is the committed blob's size
# for any file git stores with LF endings — every file in the set at #207,
# checked one by one against `git ls-tree -l` — however it was checked out.
#
# THE BASELINE is `size-ratchet.baseline`, beside this script:
#
#     <bytes> <path> <reason>
#
# one row per measured file, `<path>` relative to the root, `<reason>` the rest
# of the line and never empty; a line starting `#` is a comment. Growth raises
# the row's number and rewrites its reason, so one diff line carries both. A
# shrink passes with no edit and lowers nothing: lowering the row is how a cut
# keeps its gain.
#
# TREE AND BASELINE NAME THE SAME FILES. A file with no row is growth from
# nothing. A row with no file is a measurement this run did not make, and
# passing it would let a broken glob report half a tree clean. Both fail.
#
# WHAT IT CANNOT SEE:
#   - whether a reason is new. A growth edit keeping the old reason passes;
#     the diff shows it, and that half is the reviewer's.
#   - growth moved out of the set — into a skill's supporting file, a
#     document a definition cites, or a dotfile, which a shell glob skips.
#     Only the two globs above are measured.
#   - the commit. It measures the working tree, so an uncommitted edit counts;
#     a commit is measured by checking it out.
#   - a path holding whitespace, which no row can name. That file fails as
#     having no row: closed, not open.
#
# Usage: scripts/gates/size-ratchet.sh        (`make size-ratchet`)
# Exit:  0  every file is within its row, and every row has its file
#        1  a file is over its row or has none, or a row has no file; each is
#           printed with its numbers
#        2  cannot run — given an argument; the baseline is unreadable, holds a
#           line that does not parse, names a path twice or holds no row; no
#           file matched; or a file could not be read or counted, or counted
#           zero bytes. Outranks 1, and every line found is still printed.
#           Never reported as clean.

set -uo pipefail

DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || DIR=''
ROOT=''
[ -z "$DIR" ] || ROOT="$(cd "$DIR/../.." 2>/dev/null && pwd)" || ROOT=''
BASE="$DIR/size-ratchet.baseline"

measured=0; total=0; base_total=0; over=0; norow=0; stale=0; cannot=0

# The summary prints on every path, and the clean line only when something was
# measured and nothing failed or went unmeasured: a run that measured nothing
# must not print what a clean run prints.
finish() {
  printf 'size-ratchet · %s files · %s of %s bytes · %s over · %s no row · %s stale row\n' \
    "$measured" "$total" "$base_total" "$over" "$norow" "$stale"
  [ "$cannot" -eq 0 ] && [ "$measured" -gt 0 ] || exit 2
  [ $(( over + norow + stale )) -eq 0 ] || exit 1
  printf 'clean · every file is within its row, and every row has its file\n'
  exit 0
}

if [ "$#" -ne 0 ]; then
  echo "size-ratchet: cannot run — it takes no arguments; it measures the set its header names, never a list the caller hands it" >&2
  cannot=1; finish
fi
if [ -z "$ROOT" ]; then
  echo "size-ratchet: cannot run — this script could not resolve its own path, so neither the tree nor the baseline can be located" >&2
  cannot=1; finish
fi
if [ ! -f "$BASE" ] || [ ! -r "$BASE" ]; then
  printf 'size-ratchet: cannot run — no readable baseline at %s\n' "$BASE" >&2
  cannot=1; finish
fi

# Every baseline line is a row, a comment or blank, and anything else is named
# rather than skipped: a dropped row is a file measured against nothing. CRs go
# first because this file is checked out CRLF on the reference host too.
ROWTEXT="$(LC_ALL=C tr -d '\r' < "$BASE" | LC_ALL=C awk '
  /^[ \t]*$/ || /^[ \t]*#/ { next }
  /^[ \t]*[0-9]+[ \t]+[^ \t]+[ \t]+[^ \t]/ {
    if ($2 in seen) { printf "- line %d names %s a second time\n", NR, $2; next }
    seen[$2] = 1; printf "+ B\t%s\t%s\n", $1, $2; next
  }
  { printf "- line %d does not parse: %s\n", NR, $0 }')" || {
  printf 'size-ratchet: cannot run — the baseline at %s could not be read\n' "$BASE" >&2
  cannot=1; finish; }

BAD="$(printf '%s\n' "$ROWTEXT" | sed -n 's/^- //p')"
ROWS="$(printf '%s\n' "$ROWTEXT" | sed -n 's/^+ //p')"
if [ -n "$BAD" ]; then
  printf 'size-ratchet: cannot run — the baseline at %s does not parse:\n' "$BASE" >&2
  printf '%s\n' "$BAD" | sed 's/^/size-ratchet:   /' >&2
  cannot=1; finish
fi
if [ -z "$ROWS" ]; then
  printf 'size-ratchet: cannot run — the baseline at %s holds no row\n' "$BASE" >&2
  cannot=1; finish
fi

shopt -s nullglob
set -- "$ROOT"/.claude/agents/*.md "$ROOT"/.claude/skills/*/SKILL.md
shopt -u nullglob
if [ "$#" -eq 0 ]; then
  printf 'size-ratchet: cannot run — nothing under %s matches the measured set, and zero files is not a clean set\n' "$ROOT" >&2
  cannot=1; finish
fi

# One tagged stream for the comparison: M <bytes> <path> for a measurement, U
# <path> for a file that could not be measured — already reported here, and
# kept so that its row is not reported a second time as stale.
STREAM=''
for f in "$@"; do
  rel="${f#"$ROOT"/}"
  n=''; why='unreadable'
  if [ -f "$f" ] && [ -r "$f" ]; then
    why='its count could not be taken'
    n="$(LC_ALL=C tr -d '\r' < "$f" | LC_ALL=C wc -c)" || n=''
    n="${n//[!0-9]/}"
  fi
  if [ -z "$n" ]; then :
  elif [ "$n" -eq 0 ]; then why='zero bytes, which this gate will not vouch for'
  else printf -v line 'M\t%s\t%s\n' "$n" "$rel"; STREAM+="$line"; continue
  fi
  printf 'cannot measure · %s · %s\n' "$why" "$rel"
  cannot=1; printf -v line 'U\t%s\n' "$rel"; STREAM+="$line"
done

RESULT="$( { printf '%s\n' "$ROWS"; printf '%s' "$STREAM"; } | LC_ALL=C awk -F'\t' '
  $1 == "B" { base[$3] = $2 + 0; order[++nb] = $3; next }
  $1 == "U" { seen[$2] = 1; next }
  $1 == "M" {
    seen[$3] = 1; m++; t += $2
    if (!($3 in base))          { printf "no row · %s · %d bytes\n", $3, $2; nr++ }
    else if ($2 + 0 > base[$3]) { printf "over · %s · %d > %d (+%d)\n", $3, $2, base[$3], $2 - base[$3]; ov++ }
    next
  }
  END {
    for (i = 1; i <= nb; i++) {
      p = order[i]; bt += base[p]
      if (!(p in seen)) { printf "stale row · %s · %d bytes, and no such file is measured\n", p, base[p]; st++ }
    }
    printf "= %d %d %d %d %d %d\n", m, t, bt, ov, nr, st
  }')" || RESULT=''

STATS="$(printf '%s\n' "$RESULT" | sed -n 's/^= //p')"
printf '%s\n' "$RESULT" | sed -e '/^= /d' -e '/^$/d'
if ! [[ "$STATS" =~ ^[0-9]+\ [0-9]+\ [0-9]+\ [0-9]+\ [0-9]+\ [0-9]+$ ]]; then
  echo "size-ratchet: cannot run — the comparison produced no result" >&2
  cannot=1; finish
fi
read -r measured total base_total over norow stale <<< "$STATS"
finish
