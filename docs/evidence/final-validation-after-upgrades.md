# Final Validation After Architecture Upgrades

## Thời điểm kiểm tra

Ngày 20/05/2026, sau khi hoàn tất các nâng cấp chính của hệ thống.

## Các nâng cấp đã hoàn tất

- Route service database qua PgBouncer.
- Chuyển API key Web Gateway sang Kubernetes Secret.
- Bổ sung smoke test Saga end-to-end.
- Scale Kafka consumer theo lag bằng KEDA.
- Tăng Kafka partitions và consumer parallelism cho payment stage.
- Bổ sung MongoDB infrastructure.
- Bổ sung read-model-service theo mô hình CQRS.
- Expose MongoDB read model qua Web Gateway.
- Bổ sung tab MongoDB Read Model trên Dashboard.
- Bổ sung Redis Cache-Aside ở Web Gateway cho read API.
- Bổ sung PostgreSQL backup/restore runbook.

## Kết quả smoke test sau nâng cấp

Smoke test tạo đơn thành công qua Web Gateway.

Kết quả các tầng:

- orders.status = COMPLETED
- outbox.status = PUBLISHED
- inventory_reservations.status = RESERVED
- payments.status = COMPLETED
- notifications.status = SENT

Điều này chứng minh MongoDB CQRS và Redis cache không làm hỏng luồng Saga chính.

## Kiểm tra service health

Web Gateway `/api/health/services` trả về OK cho:

- order_service
- inventory_service
- payment_service
- notification_service
- read_model_service

## Kiểm tra MongoDB Read Model

API `/api/read-model/orders?limit=3` trả về đơn mới nhất vừa được tạo từ smoke test.

Điều này chứng minh luồng:

Kafka payment.completed
→ read-model-service
→ MongoDB
→ Web Gateway
→ Dashboard/API

đang hoạt động.

## Kiểm tra Redis Cache

API `/api/read-model/orders?limit=100` trả về:

- Lần 1: X-Cache: MISS
- Lần 2: X-Cache: HIT

Điều này chứng minh Redis Cache-Aside ở Web Gateway đã hoạt động.

## Kiểm tra PostgreSQL backup/restore

Backup thành công các database:

- order_db
- inventory_db
- payment_db
- notification_db

Restore check thành công với `order_db.dump`, dữ liệu restore gồm:

- orders: 20701 rows
- order_items: 20701 rows
- outbox: 20701 rows

Restore check sử dụng database tạm và không ảnh hưởng database thật.

## Kết luận

Sau các nâng cấp, hệ thống vẫn chạy ổn định, Saga end-to-end hoạt động đúng, read model MongoDB hoạt động, Redis cache hoạt động, và PostgreSQL có quy trình backup/restore cơ bản.
