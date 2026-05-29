# Saga Cluster Stable Checkpoint

## 1. Thời điểm checkpoint

- Ngày kiểm tra: 2026-05-29
- Git revision ổn định: 3a0c2ba
- Trạng thái GitOps:
  - ecommerce-infrastructure: Synced / Healthy
  - ecommerce-platform: Synced / Healthy
  - infrastructure-layer: Synced / Healthy

## 2. Mô hình cluster hiện tại

Cluster K3s chạy trên 3 VM OpenStack qua private IP.

| Node | Vai trò | Ghi chú |
|---|---|---|
| vm1-gateway | control-plane, etcd | web-gateway, ecommerce-dashboard, Kafka/Kafdrop |
| vm2-mesh | control-plane, etcd | order, inventory, payment, notification, read-model services |
| vm3-gitops | control-plane, etcd | Redis, MongoDB, ArgoCD/GitOps |

Dashboard chính:

- http://100.65.255.2:30517

Web Gateway NodePort:

- http://100.65.255.2:32193

## 3. Kết quả kiểm tra sau khi mở lại hệ thống

### 3.1 Cluster / GitOps

Kết quả kiểm tra:

- 3 node K3s đều Ready.
- ArgoCD các app đều Synced / Healthy.
- Không phát hiện pod lỗi ở các trạng thái:
  - Pending
  - CrashLoopBackOff
  - Error
  - ImagePullBackOff
  - Unknown
  - Terminating

### 3.2 API / Dashboard

Dashboard proxy qua Nginx hoạt động đúng.

Kết quả:

- GET http://100.65.255.2:30517/api/inventories
- HTTP 200 OK
- Dashboard hiển thị được danh sách đơn hàng.
- Đơn smoke test mới hiển thị COMPLETED trên dashboard.

## 4. Smoke test E2E

Tạo đơn smoke test qua dashboard proxy:

- POST http://100.65.255.2:30517/api/orders

Kết quả:

| Trường | Giá trị |
|---|---|
| HTTP status | 201 Created |
| Order ID | 83d34a7a-84be-4bc2-8fa8-f9297da5264c |
| Initial status | PENDING |
| Final status | COMPLETED |

Luồng xử lý đã đi đủ:

1. Dashboard / Store gửi request tạo đơn.
2. Web Gateway forward request vào Order Service.
3. Order Service tạo order PENDING và publish event order.created.
4. Inventory Consumer xử lý order.created.
5. Inventory Service tạo reservation và publish inventory.reserved hoặc inventory.failed.
6. Payment Consumer xử lý inventory.reserved.
7. Payment Service tạo payment COMPLETED hoặc FAILED.
8. Order Saga Monitor cập nhật order COMPLETED hoặc FAILED.
9. Read Model Service ghi dữ liệu sang MongoDB.
10. Notification Consumer xử lý payment event và ghi notification.

## 5. Trạng thái dữ liệu sau smoke test

### 5.1 PostgreSQL

Order DB:

| Status | Count |
|---|---:|
| COMPLETED | 14 |
| FAILED | 4 |

Inventory DB:

| Status | Count |
|---|---:|
| RESERVED | 14 |
| FAILED | 2 |
| ROLLBACKED | 2 |

Payment DB:

| Status | Count |
|---|---:|
| COMPLETED | 14 |
| FAILED | 2 |

Notification DB:

| Status | Count |
|---|---:|
| SENT | 17 |

### 5.2 MongoDB Read Model

| Collection | Count |
|---|---:|
| order_read_models | 14 |

### 5.3 Kafka Lag

Sau smoke test, không phát hiện consumer lag lớn hơn 0 ở các group chính:

- inventory-service-group
- payment-service-group
- notification-service-group
- order-service-saga-monitor
- read-model-service-group

## 6. Các lỗi đã khắc phục

### 6.1 Dashboard production gọi API lỗi

Nguyên nhân:

- Dashboard production ban đầu gọi API chưa ổn định.
- Có lúc thiếu hoặc lệch WEB_GATEWAY_API_KEY sau khi rotate secret.
- Dev mode và cluster mode dùng cấu hình khác nhau.

Cách khắc phục:

- Đưa dashboard vào Kubernetes.
- Expose dashboard qua NodePort 30517.
- Dùng Nginx reverse proxy /api sang web-gateway.default.svc.cluster.local:8090.
- API key được lấy từ Kubernetes Secret runtime, không nhúng trực tiếp vào JS bundle.
- Strip Origin và Referer khi proxy request để tránh lỗi browser-origin.

### 6.2 Order Service trỏ Redis sai host

Nguyên nhân:

- Order Service fallback sang Redis host cũ: redis-master.cache.svc.cluster.local.

Cách khắc phục:

- Cấu hình REDIS_ADDR từ ecommerce-runtime-config.
- Runtime Redis đúng: redis.default.svc.cluster.local:6379.

### 6.3 Inventory Consumer xử lý Kafka chưa an toàn

Nguyên nhân:

- Consumer cần xử lý business thành công trước khi commit Kafka offset.

Cách khắc phục:

- Sửa Inventory Consumer theo flow:
  1. FetchMessage
  2. HandleOrderCreated
  3. CommitMessages

### 6.4 Payment Consumer không xử lý inventory.reserved ban đầu

Nguyên nhân:

- Có thời điểm payment-service-group join group nhưng chưa được assign partition.
- Payment Consumer không đọc được inventory.reserved.

Cách khắc phục:

- Rebuild và redeploy payment-service.
- Payment Consumer sau đó nhận partition assignment và xử lý inventory.reserved thành công.

### 6.5 Notification Consumer chưa xử lý payment event ban đầu

Nguyên nhân:

- Consumer group notification đang rebalancing hoặc chưa assign partition.
- notification_db ban đầu chưa có dữ liệu.

Cách khắc phục:

- Restart Notification Consumer.
- Consumer xử lý backlog payment.completed và payment.failed.
- notification_db có SENT = 17.

## 7. Kết luận checkpoint

Hệ thống hiện đã ổn định cho luồng Saga chính:

1. Create Order
2. Inventory Reservation
3. Payment Processing
4. Order Status Update
5. MongoDB Read Model Update
6. Notification Sent

Checkpoint này đủ điều kiện để chuyển sang benchmark nhẹ, sau đó mới chạy spike test, stress test và chaos test.
