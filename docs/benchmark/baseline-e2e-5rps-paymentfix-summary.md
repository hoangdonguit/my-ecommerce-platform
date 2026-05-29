# Baseline E2E 5 RPS Benchmark After Payment Fix

## Thời điểm

- Run ID: `baseline-e2e-5rps-paymentfix-20260529200108`
- Git revision: `051e8a4`
- Gateway URL: `http://100.65.255.2:30517`
- Kịch bản: tạo đơn COD qua Web Gateway, sau đó kiểm tra toàn bộ Saga drain.

## Kết quả K6

- Tổng request: 301
- Accepted orders: 301
- HTTP error rate: 0%
- Unexpected error rate: 0%
- Throughput: khoảng 5 requests/second
- p95 latency: 130.52ms
- Max latency: 624.79ms

## Trạng thái DB trước benchmark

Orders:

- COMPLETED: 2063
- FAILED: 4

Payments:

- COMPLETED: 2063
- FAILED: 2

## Trạng thái DB sau benchmark và drain

Orders:

- COMPLETED: 2364
- FAILED: 4
- PENDING: 0

Payments:

- COMPLETED: 2364
- FAILED: 2
- PROCESSING: 0

Inventory:

- RESERVED: 2364
- FAILED: 2
- ROLLBACKED: 2

Kafka consumer groups:

- inventory-service-group: lag 0
- payment-service-group: lag 0
- notification-service-group: lag 0
- order-service-saga-monitor: lag 0
- read-model-service-group: lag 0

## Kết luận

Benchmark `baseline-e2e-5rps` sau khi sửa `payment-service` đạt PASS cho core Saga flow.

Hệ thống xử lý được 301 đơn trong 60 giây ở mức 5 RPS, không có lỗi HTTP, không còn order PENDING, không còn payment PROCESSING, và Kafka lag về 0 sau giai đoạn drain.

Payment fix chính:

- Nếu payment đã tồn tại ở trạng thái PROCESSING, payment-service sẽ resume/finalize lại thay vì return nil.
- Payment Status Reconciler và Order Status Reconciler vẫn được giữ như cơ chế self-healing để đảm bảo eventual consistency.

## Ghi chú

Trong lúc drain, script có lỗi nhỏ do dùng `break` bên trong pipeline với `tee`, làm vòng lặp không tự dừng dù đã đạt `FINAL_E2E_DRAIN_OK`. Lỗi này không ảnh hưởng kết quả benchmark.
