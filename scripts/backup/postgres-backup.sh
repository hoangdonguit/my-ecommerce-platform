#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-db}"
POSTGRES_POD="${POSTGRES_POD:-postgresql-0}"
BACKUP_DIR="${BACKUP_DIR:-backups/postgres}"
TIMESTAMP="$(date +%Y%m%d%H%M%S)"

DATABASES=(
  order_db
  inventory_db
  payment_db
  notification_db
)

mkdir -p "$BACKUP_DIR/$TIMESTAMP"

echo "===== POSTGRES BACKUP ====="
echo "namespace=$NAMESPACE"
echo "pod=$POSTGRES_POD"
echo "backup_dir=$BACKUP_DIR/$TIMESTAMP"

for DB in "${DATABASES[@]}"; do
  OUT_FILE="$BACKUP_DIR/$TIMESTAMP/${DB}.dump"

  echo
  echo "Backing up $DB -> $OUT_FILE"

  kubectl -n "$NAMESPACE" exec "$POSTGRES_POD" -- bash -lc "
    export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
    pg_dump -U postgres -d '$DB' -Fc
  " > "$OUT_FILE"

  if [ ! -s "$OUT_FILE" ]; then
    echo "ERROR: backup file is empty: $OUT_FILE"
    exit 1
  fi

  ls -lh "$OUT_FILE"
done

echo
echo "Backup completed: $BACKUP_DIR/$TIMESTAMP"
