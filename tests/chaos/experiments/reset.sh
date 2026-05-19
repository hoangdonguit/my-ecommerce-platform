#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# SCRIPT: SYSTEM ENVIRONMENT RESET & COOLDOWN
# MỤC ĐÍCH: Dọn dẹp dữ liệu test và khôi phục trạng thái hệ thống trước benchmark.
#
# CẢNH BÁO:
# - Script này có lệnh destructive: TRUNCATE DB, FLUSHALL Redis, delete Kafka topic.
# - Chỉ chạy khi thật sự muốn reset môi trường test.
# - Cần xác nhận bằng: CONFIRM_RESET=YES ./reset.sh
# ==============================================================================

if [ "${CONFIRM_RESET:-}" != "YES" ]; then
  echo "[ABORT] Script này sẽ xóa dữ liệu test."
  echo "        Nếu chắc chắn muốn chạy, dùng:"
  echo "        CONFIRM_RESET=YES $0"
  exit 1
fi

echo "[INFO] Bắt đầu tiến trình dọn dẹp và khôi phục hệ thống..."

echo "[0/5] Đọc PostgreSQL password từ Kubernetes Secret..."
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

echo "[1/5] Ngắt kết nối Consumer để giải phóng khóa trên Kafka..."
kubectl scale deployment inventory-consumer --replicas=0 -n default
sleep 5

echo "[2/5] Dọn dẹp PostgreSQL Database..."
psql_exec "order_db" "TRUNCATE TABLE orders CASCADE; TRUNCATE TABLE outbox CASCADE;"
psql_exec "inventory_db" "TRUNCATE TABLE inventory_reservations CASCADE; TRUNCATE TABLE inventory_reservation_items CASCADE; UPDATE inventories SET on_hand_quantity = 1000000, available_quantity = 1000000, reserved_quantity = 0;"

echo "[3/5] Dọn dẹp Redis..."
REDIS_POD="$(kubectl get pods -n default -l app=redis -o jsonpath='{.items[0].metadata.name}')"

if [ -z "$REDIS_POD" ]; then
  echo "[ERROR] Không tìm thấy Redis pod bằng label app=redis"
  exit 1
fi

kubectl exec -i "$REDIS_POD" -n default -- redis-cli FLUSHALL

echo "[4/5] Khởi tạo lại Kafka Topic: order.created..."
kubectl exec -i kafka-0 -n kafka -- \
  kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic order.created || true

echo "      -> Chờ Kafka xóa topic hoàn toàn..."
sleep 15

kubectl exec -i kafka-0 -n kafka -- \
  kafka-topics.sh --bootstrap-server localhost:9092 --create --if-not-exists --topic order.created --partitions 8 --replication-factor 1

echo "[5/5] Khởi động lại services và chờ pre-warming..."
kubectl rollout restart deployment/order-service -n default
kubectl scale deployment inventory-consumer --replicas=1 -n default

echo "      -> Đang chờ Order Service đạt trạng thái Available..."
kubectl wait --for=condition=available --timeout=90s deployment/order-service -n default
sleep 5

unset POSTGRES_PASSWORD

echo "=============================================================================="
echo "[SUCCESS] Hệ thống đã được reset. Sẵn sàng chạy K6/Chaos Test."
echo "=============================================================================="
