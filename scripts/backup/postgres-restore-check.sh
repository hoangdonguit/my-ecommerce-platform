#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-db}"
POSTGRES_POD="${POSTGRES_POD:-postgresql-0}"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <dump-file>"
  echo "Example: $0 backups/postgres/20260520013000/order_db.dump"
  exit 1
fi

DUMP_FILE="$1"

if [ ! -s "$DUMP_FILE" ]; then
  echo "ERROR: dump file not found or empty: $DUMP_FILE"
  exit 1
fi

BASE_NAME="$(basename "$DUMP_FILE" .dump)"
RESTORE_DB="restore_check_${BASE_NAME}_$(date +%Y%m%d%H%M%S)"

echo "===== POSTGRES RESTORE CHECK ====="
echo "dump_file=$DUMP_FILE"
echo "restore_db=$RESTORE_DB"

kubectl -n "$NAMESPACE" exec "$POSTGRES_POD" -- bash -lc "
  export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
  dropdb -U postgres --if-exists '$RESTORE_DB'
  createdb -U postgres '$RESTORE_DB'
"

cat "$DUMP_FILE" | kubectl -n "$NAMESPACE" exec -i "$POSTGRES_POD" -- bash -lc "
  export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
  pg_restore -U postgres -d '$RESTORE_DB' --no-owner --role=postgres
"

kubectl -n "$NAMESPACE" exec "$POSTGRES_POD" -- bash -lc "
  export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"

  echo
  echo '--- restored tables ---'
  psql -U postgres -d '$RESTORE_DB' -c '\dt'

  echo
  echo '--- restored row counts ---'
  psql -U postgres -d '$RESTORE_DB' -c \"
    SELECT schemaname, relname, n_live_tup
    FROM pg_stat_user_tables
    ORDER BY relname;
  \"

  dropdb -U postgres '$RESTORE_DB'
"

echo
echo "Restore check completed successfully."
