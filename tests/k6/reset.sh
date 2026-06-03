#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# SCRIPT: SYSTEM ENVIRONMENT RESET & COOLDOWN
# MỤC ĐÍCH:
#   Dọn dữ liệu test và đưa hệ thống về trạng thái sạch trước benchmark/stress/chaos.
#
# CẢNH BÁO:
#   - Destructive: TRUNCATE DB, FLUSHALL Redis, delete/recreate Kafka topics.
#   - Chỉ chạy trong môi trường test/demo.
#   - Cần xác nhận bằng:
#       CONFIRM_RESET=YES ./tests/k6/reset.sh
# ==============================================================================

if [ "${CONFIRM_RESET:-}" != "YES" ]; then
  echo "[ABORT] Script này sẽ xóa dữ liệu test."
  echo "        Nếu chắc chắn muốn chạy, dùng:"
  echo "        CONFIRM_RESET=YES $0"
  exit 1
fi

KAFKA_PARTITIONS="${KAFKA_PARTITIONS:-16}"
KAFKA_BOOTSTRAP="${KAFKA_BOOTSTRAP:-kafka-svc.kafka.svc.cluster.local:9092}"

KAFKA_TOPICS=(
  "order.created"
  "order.cancelled"
  "inventory.reserved"
  "inventory.failed"
  "payment.completed"
  "payment.failed"
  "inventory.reserved.dlq"
  "payment.completed.dlq"
  "payment.failed.dlq"
  "payment.dlq"
)

SCALED_OBJECTS=(
  "inventory-consumer-scaler"
  "payment-consumer-scaler"
  "notification-consumer-scaler"
)

CONSUMER_DEPLOYS=(
  "inventory-consumer"
  "payment-consumer"
  "notification-consumer"
)

API_DEPLOYS=(
  "order-service"
  "inventory-api"
  "payment-api"
  "notification-api"
  "web-gateway"
  "read-model-service"
)

echo "[INFO] Bắt đầu reset môi trường benchmark..."
echo "[INFO] Kafka bootstrap: ${KAFKA_BOOTSTRAP}"
echo "[INFO] Kafka topic partitions: ${KAFKA_PARTITIONS}"

echo "[0/9] Đọc PostgreSQL password từ Kubernetes Secret..."
POSTGRES_PASSWORD="$(kubectl -n db get secret postgresql -o jsonpath='{.data.postgres-password}' | base64 -d 2>/dev/null || true)"

if [ -z "$POSTGRES_PASSWORD" ]; then
  echo "[ERROR] Không đọc được postgres-password từ secret db/postgresql"
  exit 1
fi

psql_exec() {
  local db="$1"
  local sql="$2"

  kubectl exec -i postgresql-0 -n db -- \
    env PGPASSWORD="$POSTGRES_PASSWORD" \
    psql -U postgres -d "$db" -v ON_ERROR_STOP=1 -c "$sql"
}

kafka_exec() {
  kubectl -n kafka exec -i kafka-0 -- sh -lc "$*"
}

echo "[1/9] Pause KEDA ScaledObjects cho consumer trong lúc reset..."
for so in "${SCALED_OBJECTS[@]}"; do
  kubectl -n default annotate scaledobject "$so" autoscaling.keda.sh/paused-replicas="0" --overwrite || true
done

echo "[2/9] Scale down consumers để tránh xử lý event trong lúc reset..."
for d in "${CONSUMER_DEPLOYS[@]}"; do
  kubectl -n default scale deployment "$d" --replicas=0 || true
done

sleep 10

echo "[3/9] Dọn PostgreSQL databases..."

psql_exec "order_db" "
TRUNCATE TABLE order_items CASCADE;
TRUNCATE TABLE orders CASCADE;
TRUNCATE TABLE outbox CASCADE;
"

psql_exec "inventory_db" "
TRUNCATE TABLE inventory_reservation_items CASCADE;
TRUNCATE TABLE inventory_reservations CASCADE;
TRUNCATE TABLE inventory_outbox_events CASCADE;
UPDATE inventories
SET on_hand_quantity = 1000000,
    available_quantity = 1000000,
    reserved_quantity = 0;
"

psql_exec "payment_db" "
TRUNCATE TABLE payment_attempts CASCADE;
TRUNCATE TABLE payment_outbox_events CASCADE;
TRUNCATE TABLE payments CASCADE;
"

psql_exec "notification_db" "
TRUNCATE TABLE notifications CASCADE;
"

echo "[4/9] Dọn Redis..."
REDIS_POD="$(kubectl -n default get pods -l app=redis -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"

if [ -n "$REDIS_POD" ]; then
  kubectl -n default exec -i "$REDIS_POD" -- redis-cli FLUSHALL
else
  echo "      -> Redis pod not found, skip."
fi

echo "[4.5/9] Dọn MongoDB read model nếu có..."
if kubectl -n default get deploy mongodb >/dev/null 2>&1; then
  kubectl -n default exec deploy/mongodb -- bash -lc '
    mongosh \
      -u "$MONGO_INITDB_ROOT_USERNAME" \
      -p "$MONGO_INITDB_ROOT_PASSWORD" \
      --authenticationDatabase admin \
      ecommerce_read \
      --eval "
        db.order_read_models.deleteMany({});
        print(\"order_read_models count = \" + db.order_read_models.countDocuments());
      "
  '
else
  echo "      -> MongoDB deployment not found, skip."
fi

echo "[5/9] Reset Kafka topics của Saga..."
for topic in "${KAFKA_TOPICS[@]}"; do
  echo "      -> Delete topic: $topic"
  kafka_exec '
    TOPICS=/opt/bitnami/kafka/bin/kafka-topics.sh
    [ -x "$TOPICS" ] || TOPICS=kafka-topics.sh
    "$TOPICS" --bootstrap-server "'"$KAFKA_BOOTSTRAP"'" --delete --topic "'"$topic"'" || true
  '
done

echo "      -> Chờ Kafka xóa topic..."
sleep 20

for topic in "${KAFKA_TOPICS[@]}"; do
  echo "      -> Create topic: $topic"
  kafka_exec '
    TOPICS=/opt/bitnami/kafka/bin/kafka-topics.sh
    [ -x "$TOPICS" ] || TOPICS=kafka-topics.sh
    "$TOPICS" --bootstrap-server "'"$KAFKA_BOOTSTRAP"'" \
      --create --if-not-exists \
      --topic "'"$topic"'" \
      --partitions "'"$KAFKA_PARTITIONS"'" \
      --replication-factor 1
  '
done

echo "[6/9] Restart APIs và consumers để join lại sạch..."
for d in "${API_DEPLOYS[@]}"; do
  kubectl -n default rollout restart deployment "$d" || true
done

for d in "${CONSUMER_DEPLOYS[@]}"; do
  kubectl -n default scale deployment "$d" --replicas=1 || true
  kubectl -n default rollout restart deployment "$d" || true
done

echo "[7/9] Unpause KEDA ScaledObjects..."
for so in "${SCALED_OBJECTS[@]}"; do
  kubectl -n default annotate scaledobject "$so" autoscaling.keda.sh/paused-replicas- || true
done

echo "[8/9] Chờ deployments Available..."
for d in "${API_DEPLOYS[@]}" "${CONSUMER_DEPLOYS[@]}"; do
  kubectl -n default rollout status deployment "$d" --timeout=240s || true
done

echo "[9/9] Kiểm tra nhanh consumer groups và lag..."
sleep 15

for g in inventory-service-group payment-service-group notification-service-group order-service-saga-monitor read-model-service-group; do
  echo
  echo "----- GROUP: $g -----"
  kubectl -n kafka exec kafka-0 -- sh -lc '
    GROUPS=/opt/bitnami/kafka/bin/kafka-consumer-groups.sh
    [ -x "$GROUPS" ] || GROUPS=kafka-consumer-groups.sh
    "$GROUPS" --bootstrap-server "'"$KAFKA_BOOTSTRAP"'" \
      --describe --group "'"$g"'" --members --verbose || true
  '
done

unset POSTGRES_PASSWORD

echo "=============================================================================="
echo "[SUCCESS] Hệ thống đã được reset sạch. Sẵn sàng chạy K6/Chaos Test."
echo "=============================================================================="
