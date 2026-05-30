# Advanced Architecture Baseline Before CDC / OTel / Payment Outbox

## Thời điểm

- Audit ID: advanced-audit-before-20260530164429
- Git revision: 976825a282e32a2dbccee7469c5cd7f3d14ebc7a
- Baseline run: baseline-e2e-20rps-batchoutbox-20260530111246
- Gateway URL: http://100.65.255.2:30517

## Mục tiêu baseline

Baseline này được ghi lại trước khi triển khai các nâng cấp nâng cao:

- Payment Transactional Outbox
- Order/Payment pipeline tuning
- Debezium CDC
- Dynamic Kafka Connect filter plugin
- OpenTelemetry tracing

Mục tiêu là có số liệu before/after để chứng minh vì sao hệ thống cần được nâng cấp.

## Kết quả 20 RPS hiện tại

Kịch bản:

- Constant arrival rate: 20 RPS
- Duration: 60s
- Endpoint: POST /api/orders
- Payment method: COD
- Product: prod-123

Kết quả k6:

- Accepted orders: 1195
- Dropped iterations: 5
- HTTP failed rate: 0%
- Unexpected error rate: 0%
- Average latency: 304.63ms
- p95 latency: 1889.16ms
- Max latency: 3440.33ms

Nhận xét:

- Hệ thống nhận request tương đối ổn, không có HTTP error.
- Tuy nhiên p95 latency vượt mốc kỳ vọng 1500ms.
- Có 5 dropped iterations, chứng tỏ pipeline HTTP/order path còn nghẽn khi chạy 20 RPS.

## Trạng thái consistency sau drain

Drain cuối cùng đạt:

- PENDING_COUNT = 0
- PROCESSING_COUNT = 0
- OUTBOX_OPEN = 0
- LAG_NONZERO = 0
- FINAL_20RPS_BATCH_OUTBOX_E2E_DRAIN_OK

DB sau benchmark:

Orders:

- COMPLETED: 7161
- FAILED: 4

Payments:

- COMPLETED: 7161
- FAILED: 2

Inventory:

- RESERVED: 7161
- FAILED: 2
- ROLLBACKED: 2

Inventory outbox:

- PUBLISHED: 2429

Order outbox:

- PUBLISHED: 7165

Kết luận:

- Sau khi tối ưu inventory outbox batch publishing, hệ thống không còn mất event inventory.reserved.
- Core Saga đạt eventual consistency cuối cùng.
- Bottleneck còn lại nằm ở latency/order-payment pipeline và khả năng xử lý payment event/order completion.

## KEDA / Kafka baseline

Runtime hiện có:

- KEDA operator đang chạy.
- API scaler dùng CPU.
- Consumer scaler dùng Kafka lag.
- Consumer lag threshold: 20.
- Kafka topics chính có 8 partitions.
- Consumer max replica hiện cấu hình tới 16.

Nhận xét:

- KEDA đã tồn tại, không phải phần chưa có.
- Cần verify trong benchmark xem KEDA có scale consumer lên khi lag tăng hay không.
- Vì topic có 8 partitions, scale consumer vượt quá 8 replicas cho cùng một consumer group có thể không tăng parallelism thực tế.

## CDC readiness baseline

Hiện trạng PostgreSQL:

- wal_level = replica
- max_replication_slots = 10
- max_wal_senders = 16
- Chưa có replication slot.
- Chưa có publication.
- Chưa có Kafka Connect/Debezium pod runtime.

Nhận xét:

- PostgreSQL chưa sẵn sàng cho Debezium CDC.
- Cần chuyển wal_level sang logical, tạo publication/slot hoặc để Debezium quản lý theo connector config.
- Cần triển khai Kafka Connect + Debezium connector song song, không thay thế ngay luồng Saga chính.

## Vấn đề kỹ thuật được phát hiện

### 1. Payment dual-write risk

Payment Service hiện có rủi ro:

- Update payment DB sang trạng thái terminal.
- Sau đó publish payment.completed hoặc payment.failed trực tiếp sang Kafka.

Nếu DB update thành công nhưng Kafka publish lỗi, event có thể bị mất. Đây là lý do cần bổ sung Payment Transactional Outbox.

### 2. Order/payment pipeline còn nghẽn

Benchmark 20 RPS cho thấy:

- p95 latency cao.
- dropped iterations xuất hiện.
- drain cần nhiều vòng mới sạch pending/lag.

Cần tối ưu order/payment pipeline và quan sát KEDA trong lúc benchmark.

### 3. Debezium chưa sẵn sàng

Chưa có Kafka Connect/Debezium runtime. PostgreSQL vẫn ở wal_level=replica.

## Thứ tự nâng cấp tiếp theo

1. Bổ sung Payment Transactional Outbox.
2. Tối ưu order/payment pipeline.
3. Chạy lại 20 RPS để có after benchmark.
4. Triển khai Debezium CDC song song.
5. Bổ sung Dynamic Filter plugin nếu CDC ổn.
6. Bổ sung OpenTelemetry tracing.
7. Tổng hợp before/after và cập nhật báo cáo.
