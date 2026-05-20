#!/usr/bin/env bash
set -euo pipefail

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8090}"
PRODUCT_ID="${PRODUCT_ID:-prod-123}"
USER_ID="${USER_ID:-smoke-user-002}"

if [ -z "${API_KEY:-}" ]; then
  echo "ERROR: Missing API_KEY. Run: export API_KEY='your-api-key'"
  exit 1
fi

IDEM_KEY="smoke-price-$(date +%Y%m%d%H%M%S)"
CREATE_BODY="/tmp/create-order-${IDEM_KEY}.json"

echo "===== SAGA SUCCESS SMOKE TEST ====="
echo "GATEWAY_URL=$GATEWAY_URL"
echo "PRODUCT_ID=$PRODUCT_ID"
echo "USER_ID=$USER_ID"
echo "IDEM_KEY=$IDEM_KEY"

echo
echo "===== CHECK INVENTORY ====="
curl -sS --max-time 10 \
  -H "X-API-Key: $API_KEY" \
  "$GATEWAY_URL/api/inventories" \
| jq ".data[] | select(.product_id==\"$PRODUCT_ID\")"

echo
echo "===== CREATE ORDER ====="
HTTP_CODE=$(
  curl -sS --max-time 20 \
    -o "$CREATE_BODY" \
    -w "%{http_code}" \
    -X POST "$GATEWAY_URL/api/orders" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $API_KEY" \
    -H "X-Idempotency-Key: $IDEM_KEY" \
    -d "{
      \"user_id\": \"$USER_ID\",
      \"items\": [
        {
          \"product_id\": \"$PRODUCT_ID\",
          \"quantity\": 1
        }
      ],
      \"currency\": \"VND\",
      \"payment_method\": \"COD\",
      \"shipping_address\": \"UIT Thu Duc\",
      \"note\": \"saga success smoke test\"
    }"
)

echo "HTTP_CODE=$HTTP_CODE"
cat "$CREATE_BODY" | jq .

ORDER_ID=$(jq -r '.data.order.id // empty' "$CREATE_BODY")

if [ -z "$ORDER_ID" ] || [ "$ORDER_ID" = "null" ]; then
  echo "ERROR: Cannot parse ORDER_ID from response"
  exit 1
fi

echo "ORDER_ID=$ORDER_ID"

echo
echo "===== WAIT SAGA ====="
SAGA_WAIT_TIMEOUT="${SAGA_WAIT_TIMEOUT:-180}"
SAGA_WAIT_INTERVAL="${SAGA_WAIT_INTERVAL:-5}"
SAGA_DEADLINE=$((SECONDS + SAGA_WAIT_TIMEOUT))
ORDER_STATUS=""

while [ "$SECONDS" -lt "$SAGA_DEADLINE" ]; do
  ORDER_STATUS="$(
    kubectl -n db exec postgresql-0 -- bash -lc "
      export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
      psql -U postgres -d order_db -t -A -c \"SELECT status FROM orders WHERE id = '$ORDER_ID';\"
    " | tr -d '[:space:]'
  )"

  echo "order.status=$ORDER_STATUS elapsed=${SECONDS}s"

  if [ "$ORDER_STATUS" = "COMPLETED" ] || [ "$ORDER_STATUS" = "FAILED" ]; then
    break
  fi

  sleep "$SAGA_WAIT_INTERVAL"
done

echo
echo "===== CHECK DATABASE RESULT ====="
kubectl -n db exec postgresql-0 -- bash -lc "
export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"

echo '--- order ---'
psql -U postgres -d order_db -x -c \"
SELECT id, user_id, status, total_amount, payment_method, idempotency_key, created_at, updated_at
FROM orders
WHERE id = '$ORDER_ID';
\"

echo '--- outbox ---'
psql -U postgres -d order_db -x -c \"
SELECT event_type, aggregate_id, status, created_at
FROM outbox
WHERE aggregate_id = '$ORDER_ID';
\"

echo '--- inventory reservation ---'
psql -U postgres -d inventory_db -x -c \"
SELECT id, order_id, status, reason, created_at, updated_at
FROM inventory_reservations
WHERE order_id = '$ORDER_ID';
\"

echo '--- payment ---'
psql -U postgres -d payment_db -x -c \"
SELECT id, order_id, amount, currency, payment_method, status, transaction_id, created_at, updated_at
FROM payments
WHERE order_id = '$ORDER_ID';
\"

echo '--- notification ---'
psql -U postgres -d notification_db -x -c \"
SELECT id, user_id, order_id, event_type, channel, title, status, sent_at, created_at
FROM notifications
WHERE order_id = '$ORDER_ID';
\"
"

echo
echo "===== ASSERT EXPECTED RESULT ====="

ORDER_STATUS="$(
  kubectl -n db exec postgresql-0 -- bash -lc "
    export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
    psql -U postgres -d order_db -t -A -c \"SELECT status FROM orders WHERE id = '$ORDER_ID';\"
  " | tr -d '[:space:]'
)"

OUTBOX_STATUS="$(
  kubectl -n db exec postgresql-0 -- bash -lc "
    export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
    psql -U postgres -d order_db -t -A -c \"SELECT status FROM outbox WHERE aggregate_id = '$ORDER_ID' LIMIT 1;\"
  " | tr -d '[:space:]'
)"

INVENTORY_STATUS="$(
  kubectl -n db exec postgresql-0 -- bash -lc "
    export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
    psql -U postgres -d inventory_db -t -A -c \"SELECT status FROM inventory_reservations WHERE order_id = '$ORDER_ID' LIMIT 1;\"
  " | tr -d '[:space:]'
)"

PAYMENT_STATUS="$(
  kubectl -n db exec postgresql-0 -- bash -lc "
    export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
    psql -U postgres -d payment_db -t -A -c \"SELECT status FROM payments WHERE order_id = '$ORDER_ID' LIMIT 1;\"
  " | tr -d '[:space:]'
)"

NOTIFICATION_STATUS="$(
  kubectl -n db exec postgresql-0 -- bash -lc "
    export PGPASSWORD=\"\$(cat /opt/bitnami/postgresql/secrets/postgres-password)\"
    psql -U postgres -d notification_db -t -A -c \"SELECT status FROM notifications WHERE order_id = '$ORDER_ID' LIMIT 1;\"
  " | tr -d '[:space:]'
)"

echo "orders.status=$ORDER_STATUS"
echo "outbox.status=$OUTBOX_STATUS"
echo "inventory_reservations.status=$INVENTORY_STATUS"
echo "payments.status=$PAYMENT_STATUS"
echo "notifications.status=$NOTIFICATION_STATUS"

FAILED=0

[ "$ORDER_STATUS" = "COMPLETED" ] || FAILED=1
[ "$OUTBOX_STATUS" = "PUBLISHED" ] || FAILED=1
[ "$INVENTORY_STATUS" = "RESERVED" ] || FAILED=1
[ "$PAYMENT_STATUS" = "COMPLETED" ] || FAILED=1
[ "$NOTIFICATION_STATUS" = "SENT" ] || FAILED=1

if [ "$FAILED" -ne 0 ]; then
  echo "===== SMOKE TEST FAILED ====="
  exit 1
fi

echo "===== SMOKE TEST PASSED ====="
