#!/bin/bash
# ==============================================================================
# SCRIPT: SYSTEM ENVIRONMENT RESET & COOLDOWN
# MỤC ĐÍCH: Dọn dẹp toàn bộ dữ liệu (DB, Cache, Message Queue) và khởi tạo lại
# trạng thái hệ thống (Pre-warming) trước khi thực hiện Stress/Load Test.
# ==============================================================================

echo "[INFO] Bắt đầu tiến trình dọn dẹp và khôi phục hệ thống..."

# ------------------------------------------------------------------------------
echo "[1/5] Ngắt kết nối Consumer để giải phóng khóa (Lock) trên Kafka..."
# ------------------------------------------------------------------------------
# NOTE: Bắt buộc scale consumer về 0 trước khi xóa Topic để tránh lỗi TopicExistsException
kubectl scale deployment inventory-consumer --replicas=0 -n default
sleep 5

# ------------------------------------------------------------------------------
echo "[2/5] Dọn dẹp PostgreSQL Database (Order & Inventory)..."
# ------------------------------------------------------------------------------
# 2.1. Dọn dẹp Order Database (Xóa sạch Đơn hàng & Sự kiện Outbox)
kubectl exec -it postgresql-0 -n db -- env PGPASSWORD=securepassword psql -U postgres -d order_db -c "TRUNCATE TABLE orders CASCADE; TRUNCATE TABLE outbox CASCADE;"

# 2.2. Dọn dẹp Inventory Database (Xóa lịch sử giữ chỗ & Đặt lại 1 triệu tồn kho)
kubectl exec -it postgresql-0 -n db -- env PGPASSWORD=securepassword psql -U postgres -d inventory_db -c "TRUNCATE TABLE inventory_reservations CASCADE; TRUNCATE TABLE inventory_reservation_items CASCADE; UPDATE inventories SET on_hand_quantity = 1000000, available_quantity = 1000000, reserved_quantity = 0;"

# ------------------------------------------------------------------------------
echo "[3/5] Dọn dẹp Redis (Xóa Cache Dashboard & Idempotency Keys)..."
# ------------------------------------------------------------------------------
# Xác định chính xác Pod Redis và thực thi lệnh FLUSHALL
REDIS_POD=$(kubectl get pods -n default | grep redis | awk '{print $1}' | head -n 1)
kubectl exec -it $REDIS_POD -n default -- redis-cli FLUSHALL

# ------------------------------------------------------------------------------
echo "[4/5] Khởi tạo lại Message Queue (Kafka Topic: order.created)..."
# ------------------------------------------------------------------------------
kubectl exec -it kafka-0 -n kafka -- kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic order.created
echo "      -> Đang chờ Broker xóa hoàn toàn các segment files vật lý (15 giây)..."
sleep 15
kubectl exec -it kafka-0 -n kafka -- kafka-topics.sh --bootstrap-server localhost:9092 --create --topic order.created --partitions 8 --replication-factor 1

# ------------------------------------------------------------------------------
echo "[5/5] Khởi động lại các Services (Giải phóng RAM) và chờ Pre-warming..."
# ------------------------------------------------------------------------------
# 5.1. Restart Order Service để chống rò rỉ bộ nhớ (Memory Leak) từ đợt test trước
kubectl rollout restart deployment/order-service -n default

# 5.2. Kích hoạt lại Inventory Consumer với 1 Replica chuẩn
kubectl scale deployment inventory-consumer --replicas=1 -n default

# 5.3. Chờ đợi Readiness Probe xác nhận hệ thống sẵn sàng nhận traffic
echo "      -> Đang chờ Order Service đạt trạng thái 'Available'..."
kubectl wait --for=condition=available --timeout=90s deployment/order-service -n default
sleep 5

echo "=============================================================================="
echo "[SUCCESS] Hệ thống đã được khôi phục về trạng thái gốc. Sẵn sàng chạy K6 Test!"
echo "=============================================================================="