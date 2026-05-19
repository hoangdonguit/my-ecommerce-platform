#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# SCRIPT: SYSTEM ENVIRONMENT RESET & COOLDOWN
# MỤC ĐÍCH: Dọn dẹp dữ liệu test và khôi phục trạng thái hệ thống trước benchmark.
#
# CẢNH BÁO:
# - Script này có lệnh destructive: TRUNCATE DB, FLUSHALL Redis, delete Kafka topics.
# - Chỉ chạy trong môi trường test/demo.
# - Cần xác nhận bằng: CONFIRM_RESET=YES ./tests/k6/reset.sh
# ==============================================================================

if [ "${CONFIRM_RESET:-}" != "YES" ]; then
  echo "[ABORT] Script này sẽ xóa dữ liệu test."
  echo "        Nếu chắc chắn muốn chạy, dùng:"
  echo "        CONFIRM_RESET=YES $0"
  exit 1
fi

KAFKA_PARTITIONS="${KAFKA_PARTITIONS:-16}"

echo "[INFO] Bắt đầu reset môi trường benchmark..."
echo "[INFO] Kafka topic partitions: ${KAFKA_PARTITIONS}"

echo "[0/7] Đọc PostgreSQL password từ Kubernetes Secret..."
POSTGRES_PASSWORD="$(kubectl -n db get secret postgresql -o jsonpath='{.data.postgres-password}' | base64 -d)"

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

echo "[1/7] Scale down các consumer để tránh xử lý event trong lúc reset..."
kubectl -n default scale deployment inventory-consumer --replicas=0 || true
kubectl -n default scale deployment payment-consumer --replicas=0 || true
kubectl -n default scale deployment notification-consumer --replicas=0 || true

sleep 8

echo "[2/7] Dọn PostgreSQL databases..."
psql_exec "order_db" "
TRUNCATE TABLE orders CASCADE;
TRUNCATE TABLE outbox CASCADE;
"

psql_exec "inventory_db" "
TRUNCATE TABLE inventory_reservations CASCADE;
TRUNCATE TABLE inventory_reservation_items CASCADE;
UPDATE inventories
SET on_hand_quantity = 1000000,
    available_quantity = 1000000,
    reserved_quantity = 0;
"

psql_exec "payment_db" "
TRUNCATE TABLE payments CASCADE;
"

psql_exec "notification_db" "
TRUNCATE TABLE notifications CASCADE;
"

echo "[3/7] Dọn Redis..."
REDIS_POD="$(kubectl -n default get pods -l app=redis -o jsonpath='{.items[0].metadata.name}')"

if [ -z "$REDIS_POD" ]; then
  echo "[ERROR] Không tìm thấy Redis pod bằng label app=redis"
  exit 1
fi

kubectl -n default exec -i "$REDIS_POD" -- redis-cli FLUSHALL

echo "[4/7] Reset Kafka topics của Saga..."
KAFKA_TOPICS=(
  "order.created"
  "inventory.reserved"
  "inventory.failed"
  "payment.completed"
  "payment.failed"
)

for topic in "${KAFKA_TOPICS[@]}"; do
  echo "      -> Delete topic: $topic"
  kubectl -n kafka exec -i kafka-0 -- \
    kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic "$topic" || true
done

echo "      -> Chờ Kafka xóa topic..."
sleep 20

for topic in "${KAFKA_TOPICS[@]}"; do
  echo "      -> Create topic: $topic"
  kubectl -n kafka exec -i kafka-0 -- \
    kafka-topics.sh --bootstrap-server localhost:9092 \
    --create --if-not-exists \
    --topic "$topic" \
    --partitions "${KAFKA_PARTITIONS}" \
    --replication-factor 1
done

echo "[5/7] Restart services để consumer group join lại sạch..."
kubectl -n default rollout restart deployment/order-service
kubectl -n default rollout restart deployment/inventory-api
kubectl -n default rollout restart deployment/payment-api
kubectl -n default rollout restart deployment/notification-api
kubectl -n default rollout restart deployment/web-gateway

kubectl -n default scale deployment inventory-consumer --replicas=1
kubectl -n default scale deployment payment-consumer --replicas=1
kubectl -n default scale deployment notification-consumer --replicas=1

kubectl -n default rollout restart deployment/inventory-consumer
kubectl -n default rollout restart deployment/payment-consumer
kubectl -n default rollout restart deployment/notification-consumer

echo "[6/7] Chờ deployments Available..."
kubectl -n default rollout status deployment/order-service --timeout=180s
kubectl -n default rollout status deployment/inventory-api --timeout=180s
kubectl -n default rollout status deployment/payment-api --timeout=180s
kubectl -n default rollout status deployment/notification-api --timeout=180s
kubectl -n default rollout status deployment/web-gateway --timeout=180s
kubectl -n default rollout status deployment/inventory-consumer --timeout=180s
kubectl -n default rollout status deployment/payment-consumer --timeout=180s
kubectl -n default rollout status deployment/notification-consumer --timeout=180s

echo "[7/7] Kiểm tra nhanh consumer groups..."
sleep 15

for g in inventory-service-group payment-service-group notification-service-group order-service-saga-monitor; do
  echo
  echo "----- GROUP: $g -----"
  kubectl -n kafka exec kafka-0 -- kafka-consumer-groups.sh \
    --bootstrap-server kafka.kafka.svc.cluster.local:9092 \
    --describe --group "$g" --members --verbose || true
done

unset POSTGRES_PASSWORD

echo "=============================================================================="
echo "[SUCCESS] Hệ thống đã được reset sạch. Sẵn sàng chạy K6/Chaos Test."
echo "=============================================================================="
