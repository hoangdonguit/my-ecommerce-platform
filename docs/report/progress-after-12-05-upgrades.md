# Tổng hợp tiến độ sau báo cáo ngày 12/05/2026

## Mục tiêu

File này tổng hợp các hạng mục đã hoàn thành sau báo cáo tiến độ ngày 12/05/2026, đối chiếu với kế hoạch nâng cấp kiến trúc của đồ án.

## Trạng thái tổng quát

Sau các lần nâng cấp, hệ thống đã chuyển từ mức microservices cơ bản sang nền tảng xử lý đơn hàng cloud-native có:

- PostgreSQL làm database giao dịch chính.
- PgBouncer làm connection pooler cho PostgreSQL.
- Kafka làm event streaming cho Saga Choreography.
- KEDA autoscale consumer theo Kafka lag.
- MongoDB làm CQRS read model.
- Redis làm idempotency và read cache.
- DLQ/retry cho Kafka consumer quan trọng.
- Dashboard quan sát đơn hàng, lỗi và MongoDB read model.
- PostgreSQL backup/restore runbook.
- GitHub Actions CI cơ bản.
- ArgoCD GitOps đồng bộ ở trạng thái Synced/Healthy.

## Các hạng mục đã hoàn thành

### 1. Hoàn thiện nền tảng triển khai

- Đã route kết nối database qua PgBouncer.
- Đã bổ sung smoke test Saga end-to-end.
- Đã chuẩn hóa image tag theo từng lần build.
- Đã chuyển Web Gateway API key sang Kubernetes Secret.
- Đã loại bỏ hard-code API key khỏi dashboard và k6 scripts.
- ArgoCD đã đồng bộ các application ở trạng thái Synced/Healthy.

### 2. Xử lý bất đồng bộ và autoscaling

- Đã chuyển KEDA consumer scaler từ CPU-based sang Kafka lag-based.
- Đã tăng Kafka partitions và consumer parallelism cho payment stage.
- Đã benchmark trước/sau khi tăng payment consumer throughput.
- Đã bổ sung lightweight retry và Dead Letter Queue cho:
  - payment-consumer
  - notification-consumer

### 3. MongoDB CQRS Read Model

- Đã triển khai MongoDB trên Kubernetes.
- Đã xây dựng read-model-service.
- Đã consume event payment.completed để ghi dữ liệu đọc vào MongoDB.
- Đã expose read model qua Web Gateway.
- Đã bổ sung tab MongoDB Read Model trên dashboard.

Luồng đã chứng minh:

    Kafka payment.completed
        -> read-model-service
        -> MongoDB
        -> Web Gateway
        -> Dashboard

### 4. Redis Cache-Aside

- Đã bổ sung Redis cache ở Web Gateway cho:
  - GET /api/read-model/orders
  - GET /api/read-model/orders/:id
  - GET /api/inventories
- Đã kiểm tra header:
  - Lần 1: X-Cache: MISS
  - Lần 2: X-Cache: HIT
- TTL cache hiện dùng 5 giây để cân bằng hiệu năng và độ mới dữ liệu.

### 5. PostgreSQL Backup / Restore

- Đã bổ sung script backup PostgreSQL cho:
  - order_db
  - inventory_db
  - payment_db
  - notification_db
- Đã bổ sung restore-check vào database tạm.
- Đã kiểm tra restore thành công với order_db.
- Backup file thật được đưa vào .gitignore, không commit dump lên GitHub.

### 6. DevOps / CI

- Đã bổ sung GitHub Actions CI cơ bản.
- CI hiện kiểm tra:
  - Secret/IP cũ ở mức cơ bản.
  - Bash syntax.
  - k6 JavaScript syntax.
  - Go test/build các service.
  - Dashboard npm build.
- Workflow CI đã chạy thành công trên GitHub Actions.

## Kết quả kiểm tra cuối

Smoke test sau các nâng cấp đã pass:

- orders.status = COMPLETED
- outbox.status = PUBLISHED
- inventory_reservations.status = RESERVED
- payments.status = COMPLETED
- notifications.status = SENT

Redis cache đã hoạt động:

- X-Cache: MISS ở request đầu.
- X-Cache: HIT ở request sau.

Kafka DLQ trong kịch bản thành công đang rỗng, đây là đúng kỳ vọng.

## Các hạng mục còn lại

### Nên làm nếu còn thời gian

- Nâng CI bằng Gitleaks hoặc Trivy.
- Chuẩn hóa bảng benchmark smoke/load/stress test.
- Bổ sung branch protection trên GitHub.
- Chuẩn hóa Dockerfile theo non-root user.
- Bổ sung security headers cho dashboard hoặc Web Gateway.

### Có thể đưa vào hướng phát triển

- Flash Sale Mode bằng Redis Atomic Stock Gate.
- NetworkPolicy/RBAC chi tiết.
- OpenTelemetry tracing.
- Debezium CDC.
- Canary deployment bằng Istio.
- External Secrets hoặc Vault.
- Join máy Hậu vào cluster và tách App/Data node.

## Kết luận

Các nâng cấp sau ngày 12/05 đã giải quyết phần lớn góp ý chính: chứng minh xử lý bất đồng bộ bằng Kafka, scale consumer theo lag, tách tải đọc bằng MongoDB CQRS, giảm tải đọc bằng Redis cache, có backup/restore, có DLQ/retry và có CI kiểm tra tự động.

Hệ thống hiện đã đủ nền tảng để đưa vào báo cáo chính thức và demo kỹ thuật.
