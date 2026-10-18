#!/usr/bin/env bash
# migrate-verify.sh — apply a directory's migrations to a throwaway PostGIS
# database, run each verify file part by part, run each rollback twice, and
# print one typed result (#153).
#
# It exists so the migrations' assertions are executed rather than argued. It
# runs in CI, in the `migrations` job of `.github/workflows/gates.yml`, against
# that job's PostGIS service container, which is discarded with the job. Nothing
# durable is provisioned and no migration is applied anywhere that outlives the
# run: `Architecture.md` gates applying a migration to any lasting database on
# explicit human authorisation, and #153 authorises an ephemeral instance only.
#
# EPHEMERAL BY CONSTRUCTION. It refuses a PGHOST that is not loopback, and it
# never touches an existing database: it creates its own, empty, from template0,
# runs everything there, and drops it on exit.
#
# WHAT RUNS, for DIR:
#   1. every `[0-9]*.sql` that is not `*.verify.sql` or `*.down.sql`, in byte
#      order, each followed by its `NNNN_name.verify.sql`. An apply that fails
#      stops the sequence; what it stops is reported `not run`, never left out.
#   2. each verify file PART BY PART. A part opens at a line beginning
#      `-- === Part <id>` and runs to the next; everything above the first part
#      is a preamble run ahead of each part; a file with no part line is one
#      part, `whole`. Each part runs in its own session under ON_ERROR_STOP, so
#      a failure in one part does not hide the result of the next. The part's
#      file keeps the original's name and line numbers — other parts' lines are
#      blanked, not removed — so an error names the line in the real file.
#   3. each applied migration's `NNNN_name.down.sql`, in reverse order, TWICE.
#      A rollback written with `IF EXISTS` exits 0 whether it dropped everything
#      or nothing, so its exit status is observed rather than assumed: each run
#      reports its exit status beside the relations it actually removed.
#   4. a residue check: any relation present after the rollbacks that was not
#      present before the first apply.
#
# A PART'S RESULT:
#   pass       the part ran to its end and raised nothing
#   fail       the part raised an error; its first ERROR line is quoted
#   unreached  the part ran to its end but said, in a NOTICE, WARNING or INFO
#              containing "not a pass", that it could not reach its condition
#              — an empty table where a loaded copy is needed. It is not a pass.
# A part whose output carries a query plan is listed under `explain:`, and its
# output is kept as a `.explain` file in the logs directory, so the workflow can
# show EXPLAIN (ANALYZE, BUFFERS) output in a group of its own.
#
# Usage:  scripts/db/migrate-verify.sh DIR
#         connection from libpq's environment: PGHOST (loopback only), PGPORT,
#         PGUSER, PGPASSWORD; PGDATABASE, default postgres, is only where the
#         scratch database is created from.
# Exit:   0 pass, or `none` — DIR holds no migration · 1 something failed
#         3 incomplete — nothing failed, but a part was unreached or a verify or
#           rollback file is absent · 2 cannot run
# Each step's output is kept in the `logs:` directory the result names.

set -uo pipefail
export LC_ALL=C

die() { echo "migrate-verify: cannot run — $*" >&2; exit 2; }

[ "$#" -eq 1 ] || die 'usage: migrate-verify.sh DIR'
DIR="$1"
[ -d "$DIR" ] || die "no such directory: $DIR"

ROOT="$(cd "$(dirname "$0")/../.." 2>/dev/null && pwd)"
HEAD="$(git -C "$ROOT" rev-parse HEAD 2>/dev/null || echo unknown)"

migs=()
while IFS= read -r f; do migs+=("$f"); done < <(
  find "$DIR" -maxdepth 1 -type f -name '[0-9]*.sql' ! -name '*.verify.sql' ! -name '*.down.sql' | sort)

if [ "${#migs[@]}" -eq 0 ]; then
  printf 'migrate-verify:\n  dir: %s\n  head: %s\n  migrations: none — nothing applied, verified or rolled back\n  result: none\n' "$DIR" "$HEAD"
  exit 0
fi

command -v psql >/dev/null 2>&1 || die 'no psql on PATH'
case "${PGHOST:-}" in
  127.0.0.1|localhost|::1) ;;
  *) die "PGHOST is '${PGHOST:-}', not loopback; this applies and rolls back migrations and runs only against a throwaway local instance" ;;
esac

LOG="$(mktemp -d 2>/dev/null)" && [ -d "$LOG" ] || die 'mktemp -d failed'
MAINT="${PGDATABASE:-postgres}"
DB="mv_$(date +%s)_$$"
export PGCONNECT_TIMEOUT="${PGCONNECT_TIMEOUT:-10}"
PSQL=(psql -X -v ON_ERROR_STOP=1)

"${PSQL[@]}" -q -d "$MAINT" -c "CREATE DATABASE \"$DB\" TEMPLATE template0" > "$LOG/000-create-database.log" 2>&1 \
  || die "cannot create the scratch database: $(head -n 1 "$LOG/000-create-database.log")"
cleanup() { "${PSQL[@]}" -q -d "$MAINT" -c "DROP DATABASE IF EXISTS \"$DB\" WITH (FORCE)" >/dev/null 2>&1; }
trap cleanup EXIT

server="$("${PSQL[@]}" -At -d "$DB" -c 'SHOW server_version' 2>/dev/null | head -n 1)"
postgis="$("${PSQL[@]}" -At -d "$DB" -c "SELECT coalesce(default_version, 'absent') FROM pg_available_extensions WHERE name = 'postgis' UNION ALL SELECT 'absent' LIMIT 1" 2>/dev/null | head -n 1)"

# The relations a migration could have made: every pg_class entry outside the
# system schemas, except an extension's own and the indexes on them, and TOAST.
SNAP_SQL="SELECT format('%s.%s', n.nspname, c.relname)
  FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_index i ON i.indexrelid = c.oid
 WHERE c.relkind <> 't'
   AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
   AND n.nspname !~ '^pg_(toast_)?temp_'
   AND NOT EXISTS (SELECT 1 FROM pg_depend d
                    WHERE d.classid = 'pg_class'::regclass AND d.deptype = 'e'
                      AND d.objid IN (c.oid, coalesce(i.indrelid, 0::oid)))
"
# Sorted here, not in SQL: comm needs the byte order LC_ALL=C gives, and the
# server's collation is not that.
snap() { "${PSQL[@]}" -At -d "$DB" -c "$SNAP_SQL" 2>> "$LOG/snapshots.err" | sort > "$1"; }

seq=0
# step <label> <file> — runs one file in its own session, queries echoed. Sets
# rc, and plan=1 when the output holds a query plan; leaves $out (stdout) and
# $err (stderr) beside a .log holding both.
step() {
  seq=$((seq + 1))
  local stem; stem="$LOG/$(printf '%03d' "$seq")-$1"
  out="$stem.out"; err="$stem.err"
  rc=0; "${PSQL[@]}" -e -d "$DB" -f "$2" > "$out" 2> "$err" || rc=$?
  { echo "== $2 (exit $rc)"; echo '-- messages'; cat "$err"; echo '-- output'; cat "$out"; } > "$stem.log"
  plan=0; if grep -q 'QUERY PLAN' "$out"; then plan=1; cp "$out" "$stem.explain"; fi
  return 0
}
# A quoted, single-line YAML scalar: the line with psql's path prefix cut back to file:line.
q() { printf '"%s"' "$(printf '%s' "$1" | tr -d '\r' | sed -E 's#^psql:([^ ]*/)?##; s/\\/\\\\/g; s/"/\\"/g' | cut -c1-300)"; }
first_error() { grep -m1 -E 'ERROR:|FATAL:' "$err" || head -n 1 "$err"; }

failed=0; incomplete=0; explain=(); applied=()
report="$LOG/report.yml"; : > "$report"
emit() { printf '%s\n' "$*" >> "$report"; }

snap "$LOG/before.snap"
stopped=0
for m in "${migs[@]}"; do
  base="$(basename "$m" .sql)"; v="$DIR/$base.verify.sql"
  emit "    - file: $base.sql"
  if [ "$stopped" -eq 1 ]; then
    emit '      apply: not run — an earlier apply failed'
    emit '      verify: not run'; incomplete=1; continue
  fi
  step "$base-apply" "$m"
  if [ "$rc" -ne 0 ]; then
    emit "      apply: {result: fail, exit: $rc, error: $(q "$(first_error)")}"
    emit '      verify: not run — the apply failed'
    failed=1; stopped=1; continue
  fi
  emit '      apply: pass'; applied+=("$m")
  if [ ! -f "$v" ]; then emit '      verify: absent'; incomplete=1; continue; fi
  emit '      verify:'
  pdir="$LOG/parts-$base"; mkdir -p "$pdir"
  # "<id> <first line> <last line> <preamble length>" for each part, in order.
  while read -r id s e pre; do
    mkdir -p "$pdir/$id"; pf="$pdir/$id/$(basename "$v")"
    awk -v pre="$pre" -v s="$s" -v e="$e" '{ print (NR <= pre || (NR >= s && NR <= e)) ? $0 : "" }' "$v" > "$pf"
    step "$base-verify-$id" "$pf"
    if [ "$rc" -ne 0 ]; then
      emit "        - {part: $id, result: fail, exit: $rc, error: $(q "$(first_error)")}"; failed=1
    elif said="$(grep -m1 -iE '(NOTICE|WARNING|INFO):.*not a pass' "$err")"; then
      emit "        - {part: $id, result: unreached, says: $(q "$said")}"; incomplete=1
    else
      emit "        - {part: $id, result: pass}"
    fi
    [ "$plan" -eq 1 ] && explain+=("$base.verify.sql part $id")
  done < <(awk '
    { sub(/\r$/, "") }
    /^-- === Part [A-Za-z0-9]+/ { match($0, /Part [A-Za-z0-9]+/); n++; id[n] = substr($0, RSTART + 5, RLENGTH - 5); start[n] = NR }
    END {
      if (n == 0) { print "whole 1 " NR " 0"; exit }
      for (k = 1; k <= n; k++) print id[k], start[k], (k < n ? start[k + 1] - 1 : NR), start[1] - 1
    }' "$v")
done

emit '  rollback:'
if [ "${#applied[@]}" -eq 0 ]; then emit '    - none applied, so none run'; fi
for (( i=${#applied[@]}-1; i>=0; i-- )); do
  base="$(basename "${applied[$i]}" .sql)"; d="$DIR/$base.down.sql"
  if [ ! -f "$d" ]; then emit "    - {file: $base.down.sql, result: absent}"; incomplete=1; continue; fi
  verdict=pass; runs=''; counts=(); codes=()
  for run in 1 2; do
    snap "$LOG/pre.snap"; step "$base-down-run$run" "$d"; snap "$LOG/post.snap"
    gone="$(comm -23 "$LOG/pre.snap" "$LOG/post.snap" | paste -sd, - | sed 's/,/, /g')"
    n="$(comm -23 "$LOG/pre.snap" "$LOG/post.snap" | grep -c .)"
    counts+=("$n"); codes+=("$rc")
    [ "$rc" -eq 0 ] || verdict=fail
    runs="$runs, run$run: {exit: $rc, dropped: [$gone]}"
  done
  seen="exit ${codes[0]} then ${codes[1]}"
  if [ "${codes[0]}" = "${codes[1]}" ] && [ "${counts[0]}" != "${counts[1]}" ]; then
    seen="exit ${codes[0]} on both runs, with ${counts[0]} and then ${counts[1]} relation(s) dropped: the exit status does not tell them apart"
  fi
  [ "$verdict" = fail ] && failed=1
  emit "    - {file: $base.down.sql, result: $verdict$runs, observed: $(q "$seen")}"
done

snap "$LOG/after.snap"
residue="$(comm -13 "$LOG/before.snap" "$LOG/after.snap" | paste -sd, - | sed 's/,/, /g')"
if [ -s "$LOG/snapshots.err" ]; then residue_line="unmeasured — a snapshot query failed: $(q "$(head -n 1 "$LOG/snapshots.err")")"; failed=1
elif [ -n "$residue" ]; then residue_line="[$residue]"; failed=1
else residue_line='none'; fi

result=pass; rc_all=0
[ "$incomplete" -eq 1 ] && { result=incomplete; rc_all=3; }
[ "$failed" -eq 1 ] && { result=fail; rc_all=1; }
ex='none'; [ "${#explain[@]}" -gt 0 ] && ex="[$(printf '%s, ' "${explain[@]}" | sed 's/, $//')]"

printf 'migrate-verify:\n'
printf '  dir: %s\n' "$DIR"
printf '  head: %s\n' "$HEAD"
printf '  server: {postgresql: %s, postgis_available: %s}\n' "${server:-unknown}" "${postgis:-unknown}"
printf '  migrations:\n'
cat "$report"
printf '  residue: %s\n' "$residue_line"
printf '  explain: %s\n' "$ex"
printf '  result: %s\n' "$result"
printf '  logs: %s\n' "$LOG"
exit "$rc_all"
