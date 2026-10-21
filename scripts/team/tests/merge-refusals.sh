#!/usr/bin/env bash
# merge-refusals.sh — proves scripts/team/merge.sh refuses what #229 says it refuses.
# gh is a stub, and the tokens in the environment are sentinel strings, never real ones.
#
# Every case starts from one fixture set in which every check holds, and breaks one
# thing; the positive controls prove that set merges, without which "always refuse"
# would pass every refusal below.
#   1. a PR that is not open
#   2. an author other than Caisesiume
#   3. a moved head
#   4. the label awaiting-human; 5. the label owner-merge
#   6. no TheReviewNinja APPROVED review at the head: none, changes requested, approved
#      at another SHA, approved by another login, an approval superseded or dismissed
#   7. `gates` or `migrations` not completed/success at the head: missing, failed, in
#      progress, green only from another app, green superseded by a later red run
#   8. a safety path without a passing safety verdict at the head
#   9. a default account other than Caisesiume (cannot run)
#  10. --apply runs exactly one merge, pinned to the head, as the default account;
#      a dry run and a refusal run none; no token reaches gh or the output
#  11. merge.sh never expands GH_JUDGE_TOKEN and never asks for --admin, --auto,
#      --squash or --rebase
#
# Hermetic: the stub answers what `gh --jq` would have printed (standalone jq is not
# installed, so a stub cannot run the filter); fixtures are files; nothing touches the
# network or the repository. The jq filters were run live, read-only, against PR #228.
# Usage: scripts/team/tests/merge-refusals.sh    Exit: 0 all pass · 1 any failure

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$DIR/../merge.sh"
TMP="$(mktemp -d)" || TMP=''
[ -n "$TMP" ] && [ -d "$TMP" ] || { echo 'FAIL  mktemp -d gave no directory'; exit 1; }
trap 'rm -rf "$TMP"' EXIT

FIX="$TMP/fix"; CALLS="$TMP/calls.log"; mkdir -p "$FIX" "$TMP/bin"
SECRET="sentinel-$$-not-a-token"
HEAD=1111111111111111111111111111111111111111
OLD=2222222222222222222222222222222222222222
T=$'\t'

cat > "$TMP/bin/gh" <<'STUB'
#!/bin/sh
# Logs every call and the NAMES of any token variables it can see, never a value.
seen=''
for v in GH_TOKEN GITHUB_TOKEN GH_JUDGE_TOKEN; do printenv "$v" >/dev/null 2>&1 && seen="$seen $v"; done
printf '%s | tokens:%s\n' "$*" "$seen" >> "$CALLS"
case "$*" in
  "api user "*)                              cat "$FIX/user" ;;
  "pr view "*"state,mergeCommit"*)           cat "$FIX/after" ;;
  "pr view "*)                               cat "$FIX/pr" ;;
  "api repos/"*"/pulls/"*"/reviews "*)       cat "$FIX/reviews" ;;
  "api repos/"*"/commits/"*"/check-runs "*)  cat "$FIX/checks" ;;
  "api repos/"*"/pulls/"*"/files "*)         cat "$FIX/files" ;;
  "api repos/"*"/issues/"*"/comments "*)     cat "$FIX/comments" ;;
  "pr merge "*) printf 'MERGE %s\n' "$*" >> "$CALLS"; printf 'MERGED%sabc\n' "$(printf '\t')" > "$FIX/after" ;;
  *) printf 'WRITE %s\n' "$*" >> "$CALLS" ;;
esac
exit 0
STUB
chmod +x "$TMP/bin/gh"

pr()       { printf '%s%s%s%s%s%s%s\n' "${3:-OPEN}" "$T" "$1" "$T" "${4:-Caisesiume}" "$T" "${2:-}" > "$FIX/pr"; }
lines()    { f="$1"; shift; : > "$FIX/$f"; while [ "$#" -gt 0 ]; do printf '%s\n' "$1" >> "$FIX/$f"; shift; done; }
review()   { printf '%s%s%s%s%s%s%s%s%s' "$1" "$T" "$2" "$T" "$3" "$T" "$4" "$T" "$5"; }   # login sha at id state
check()    { printf '%s%s%s%s%s%s%s%s%s' "$1" "$T" "$2" "$T" "$3" "$T" "$4" "$T" "$5"; }    # name app id status conclusion
verdict()  { printf '%s%s%s%s<!-- team:safety v1 pr=9 sha=%s result=%s -->' "$1" "$T" "$2" "$T" "$3" "$4"; }
APPROVED_AT_HEAD="$(review TheReviewNinja "$HEAD" 2026-10-20T10:00:00Z 50 APPROVED)"
GATES_OK="$(check gates github-actions 100 completed success)"
MIGR_OK="$(check migrations github-actions 101 completed success)"
green() {   # every check holds
  printf 'Caisesiume\n' > "$FIX/user"; printf 'OPEN%s\n' "$T" > "$FIX/after"
  pr "$HEAD"; lines reviews "$APPROVED_AT_HEAD"; lines checks "$GATES_OK" "$MIGR_OK"
  lines files service/cmd/turfgps/main.go README.md; lines comments
}

fails=0; OUT=''; RC=0
run() { # under sentinel tokens; a dry run unless the case passes --apply
  : > "$CALLS"
  OUT="$(GH="$TMP/bin/gh" CALLS="$CALLS" FIX="$FIX" GH_JUDGE_TOKEN="$SECRET" \
         GH_TOKEN="$SECRET-gh" GITHUB_TOKEN="$SECRET-github" bash "$SCRIPT" --pr 9 --sha "$HEAD" "$@" 2>&1)"
  RC=$?
}
ok()   { printf 'PASS  %s\n' "$1"; }
bad()  { printf 'FAIL  %s — rc %s: %s\n' "$1" "$RC" "$(printf '%s' "$OUT" | tr '\n' '|' | cut -c1-300)"; fails=$((fails + 1)); }
# expect <label> <rc> <substring> — and, every time, no merge, no write, no token seen
expect() {
  if [ "$RC" != "$2" ]; then bad "$1"; return; fi
  case "$OUT" in *"$3"*) ;; *) bad "$1"; return ;; esac
  if grep -qE '^(MERGE|WRITE) ' "$CALLS"; then bad "$1 (gh was asked to write: $(grep -m1 -E '^(MERGE|WRITE) ' "$CALLS"))"; return; fi
  if grep -qE 'tokens:.+' "$CALLS"; then bad "$1 (a token reached gh: $(grep -m1 -oE 'tokens:.+' "$CALLS"))"; return; fi
  case "$OUT" in *"$SECRET"*) bad "$1 (a token value reached the output)"; return ;; esac
  ok "$1"
}
refused() { expect "$1" 1 "$2"; }

# 0 — the positive controls
green; run;                      expect "all checks hold: a dry run would merge" 0 "would run: gh pr merge 9 -R Caisesiume/TurfGPS --merge --match-head-commit $HEAD"
green; lines reviews "$(review TheReviewNinja "$HEAD" 2026-10-20T09:00:00Z 40 CHANGES_REQUESTED)" "$APPROVED_AT_HEAD"
run;                             expect "changes requested, then approved at the head: would merge" 0 "would run:"
green; lines checks "$(check gates github-actions 90 completed failure)" "$GATES_OK" "$MIGR_OK"
run;                             expect "a red gates run re-run green: would merge" 0 "would run:"
green; lines files docs/SPECIFICATION.md; lines comments "$(verdict TheReviewNinja 2026-10-20T10:00:00Z "$HEAD" pass)"
run;                             expect "a safety path with a passing verdict at the head: would merge" 0 "safety verdict: pass"

# 1 — not open
green; pr "$HEAD" '' MERGED; run; refused "a merged PR is refused" "is MERGED, not open"
green; pr "$HEAD" '' CLOSED; run; refused "a closed PR is refused" "is CLOSED, not open"
# 2 — an untrusted author
green; pr "$HEAD" '' OPEN someone-else;  run; refused "a PR by another login is refused" "by 'someone-else', not Caisesiume"
green; pr "$HEAD" '' OPEN TheReviewNinja; run; refused "a PR by the judge's login is refused" "by 'TheReviewNinja', not Caisesiume"
# 3 — a moved head
green; pr "$OLD"; run;           refused "a moved head is refused" "the head moved"
# 4, 5 — reserved for the Owner
green; pr "$HEAD" 'risk:low,awaiting-human'; run; refused "awaiting-human is refused" "labelled awaiting-human"
green; pr "$HEAD" 'owner-merge,risk:high';   run; refused "owner-merge is refused" "labelled owner-merge"

# 6 — no approval at the head
green; lines reviews; run;       refused "no review at all is refused" "(latest review there: none)"
green; lines reviews "$(review TheReviewNinja "$HEAD" 2026-10-20T10:00:00Z 50 CHANGES_REQUESTED)"
run;                             refused "changes requested at the head is refused" "(latest review there: CHANGES_REQUESTED)"
green; lines reviews "$(review TheReviewNinja "$OLD" 2026-10-20T10:00:00Z 50 APPROVED)"
run;                             refused "an approval at another SHA is refused" "(latest review there: none)"
green; lines reviews "$(review Caisesiume "$HEAD" 2026-10-20T10:00:00Z 50 APPROVED)"
run;                             refused "an approval by another login is refused" "(latest review there: none)"
green; lines reviews "$APPROVED_AT_HEAD" "$(review TheReviewNinja "$HEAD" 2026-10-20T11:00:00Z 60 COMMENTED)"
run;                             refused "an approval superseded by a later comment review is refused" "(latest review there: COMMENTED)"
green; lines reviews "$(review TheReviewNinja "$HEAD" 2026-10-20T10:00:00Z 50 DISMISSED)"
run;                             refused "a dismissed approval is refused" "(latest review there: DISMISSED)"

# 7 — the required checks
green; lines checks "$MIGR_OK";  run; refused "a missing gates run is refused" "check 'gates' at $HEAD is missing"
green; lines checks "$GATES_OK"; run; refused "a missing migrations run is refused" "check 'migrations' at $HEAD is missing"
green; lines checks "$GATES_OK" "$(check migrations github-actions 101 completed failure)"
run;                             refused "a failed migrations run is refused" "check 'migrations' at $HEAD is completed/failure"
green; lines checks "$(check gates github-actions 100 in_progress '')" "$MIGR_OK"
run;                             refused "a gates run in progress is refused" "check 'gates' at $HEAD is in_progress/none"
green; lines checks "$(check gates some-other-app 100 completed success)" "$MIGR_OK"
run;                             refused "a green gates run from another app is refused" "check 'gates' at $HEAD is missing"
green; lines checks "$GATES_OK" "$(check gates github-actions 110 completed failure)" "$MIGR_OK"
run;                             refused "a green gates run superseded by a red re-run is refused" "check 'gates' at $HEAD is completed/failure"

# 8 — a safety path without a passing verdict at the head
green; lines files service/cmd/turfgps/main.go docs/SPECIFICATION.md
run;                             refused "a safety path with no verdict is refused" "no passing safety verdict"
lines comments "$(verdict TheReviewNinja 2026-10-20T10:00:00Z "$HEAD" block)"; run; refused "a safety path with a block is refused" "(latest: block)"
lines comments "$(verdict TheReviewNinja 2026-10-20T10:00:00Z "$OLD" pass)";   run; refused "a safety path with a pass at another SHA is refused" "(latest: none)"
lines comments "$(verdict Caisesiume 2026-10-20T10:00:00Z "$HEAD" pass)";      run; refused "a safety path with a pass by another login is refused" "(latest: none)"
lines comments "$(verdict TheReviewNinja 2026-10-20T10:00:00Z "$HEAD" pass)" \
               "$(verdict TheReviewNinja 2026-10-20T11:00:00Z "$HEAD" block)"; run; refused "a pass superseded by a later block is refused" "(latest: block)"
lines files service/x.go docs/Requirements/access-classification.md; lines comments
run;                             refused "a renamed-away safety file still counts" "access-classification.md"

# 9 — the default account
green; printf 'TheReviewNinja\n' > "$FIX/user"; run --apply
expect "a default account other than Caisesiume cannot run" 2 "the default gh account is 'TheReviewNinja'"

# 10 — the live path
green; run --apply
merges="$(grep -c '^MERGE ' "$CALLS")"
if [ "$RC" = 0 ] && [ "$merges" = 1 ] \
   && grep -qxF "MERGE pr merge 9 -R Caisesiume/TurfGPS --merge --match-head-commit $HEAD" "$CALLS" \
   && ! grep -qE 'tokens:.+' "$CALLS" && ! grep -q '^WRITE ' "$CALLS"; then
  case "$OUT" in *"merged: #9 at $HEAD as Caisesiume"*) ok "--apply runs one merge, pinned to the head, with no token" ;; *) bad "--apply reports the merge" ;; esac
else bad "--apply runs one merge, pinned to the head, with no token ($merges merge calls; $(grep -m1 -oE 'tokens:.+' "$CALLS"))"; fi
green; pr "$HEAD" owner-merge; run --apply; refused "--apply on a refused PR runs no merge" "labelled owner-merge"

# 11 — static: no judge token, no bypass
if grep -nE '[$][{]?GH_JUDGE_TOKEN' "$SCRIPT" | grep -q .; then
  printf 'FAIL  merge.sh expands GH_JUDGE_TOKEN\n'; fails=$((fails + 1))
else ok "merge.sh never expands GH_JUDGE_TOKEN"; fi
if grep -vE '^[[:space:]]*#' "$SCRIPT" | grep -nE -- '--(admin|auto|squash|rebase)\b' | grep -q .; then
  printf 'FAIL  merge.sh asks for a bypass or another merge method\n'; fails=$((fails + 1))
else ok "merge.sh asks for no --admin, --auto, --squash or --rebase"; fi

[ "$fails" -eq 0 ] || { printf '\n%s test(s) failed\n' "$fails"; exit 1; }
printf '\nall passed\n'
exit 0
