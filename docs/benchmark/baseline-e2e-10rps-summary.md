# Baseline E2E 10 RPS Benchmark

## Thời điểm

- Run ID: `baseline-e2e-10rps-20260529213706`
- Gateway URL: `http://100.65.255.2:30517`
- Kịch bản: tạo đơn COD qua Web Gateway, sau đó kiểm tra toàn bộ Saga drain.

## Kết quả K6

- Tổng request: 600
- Accepted orders: 600
- HTTP error rate: 0%
- Unexpected error rate: 0%
- Throughput: khoảng 10 requests/second
- p95 latency: 166.85ms
- Max latency: 1.15s

## Trạng thái DB trước benchmark

Orders:

- COMPLETED: 2364
- FAILED: 4

Payments:

- COMPLETED: 2364
- FAILED: 2

## Trạng thái DB sau benchmark và drain

Orders:

- COMPLETED: 2964
- FAILED: 4
- PENDING: 0

Payments:

- COMPLETED: 2964
- FAILED: 2
- PROCESSING: 0

Inventory:

- RESERVED: 2964
- FAILED: 2
- ROLLBACKED: 2

Kafka consumer groups:

- inventory-service-group: lag 0
- payment-service-group: lag 0
- notification-service-group: lag 0
- order-service-saga-monitor: lag 0
- read-model-service-group: lag 0

## Kết luận

Benchmark `baseline-e2e-10rps` đạt PASS cho core Saga flow.

Hệ thống xử lý được 600 đơn trong 60 giây ở mức 10 RPS, không có lỗi HTTP, không còn order PENDING, không còn payment PROCESSING, và Kafka lag về 0 sau giai đoạn drain.

## Ghi chú

Lệnh drain được chạy hơi trễ sau khi k6 kết thúc, nên benchmark này xác nhận trạng thái eventual consistency cuối cùng, nhưng không dùng để kết luận chính xác thời gian drain.
