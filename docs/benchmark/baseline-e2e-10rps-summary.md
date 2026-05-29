# Baseline E2E 10 RPS Benchmark With Immediate Drain

## Thời điểm

- Run ID: `baseline-e2e-10rps-redrain-20260529215304`
- Gateway URL: `http://100.65.255.2:30517`
- Kịch bản: tạo đơn COD qua Web Gateway ở mức 10 RPS trong 60 giây, sau đó chạy drain check ngay lập tức.

## Kết quả K6

- Tổng request: 601
- Accepted orders: 601
- HTTP error rate: 0%
- Unexpected error rate: 0%
- Throughput: khoảng 10 requests/second
- p95 latency: 181.91ms
- Max latency: 388.03ms

## Trạng thái DB trước benchmark

Orders:

- COMPLETED: 2964
- FAILED: 4

Payments:

- COMPLETED: 2964
- FAILED: 2

## Trạng thái DB sau benchmark và drain

Orders:

- COMPLETED: 3565
- FAILED: 4
- PENDING: 0

Payments:

- COMPLETED: 3565
- FAILED: 2
- PROCESSING: 0

Inventory:

- RESERVED: 3565
- FAILED: 2
- ROLLBACKED: 2

Kafka consumer groups:

- inventory-service-group: lag 0
- payment-service-group: lag 0
- notification-service-group: lag 0
- order-service-saga-monitor: lag 0
- read-model-service-group: lag 0

## Drain timeline

- K6 finished at: `2026-05-29T21:54:08+07:00`
- Drain started at: `2026-05-29T21:54:11+07:00`
- Drain finished at: `2026-05-29T21:57:10+07:00`
- Estimated drain time after K6: khoảng 3 phút 2 giây

Trong quá trình drain, hệ thống có backlog tạm thời:

- Check 1: PENDING còn 560, Kafka lag còn 8 dòng
- Check 3: payment PROCESSING tạm thời còn 2
- Check 6: PENDING = 0, PROCESSING = 0, Kafka lag = 0

## Kết luận

Benchmark `baseline-e2e-10rps-redrain` đạt PASS cho core Saga flow.

Hệ thống xử lý được 601 đơn trong 60 giây ở mức 10 RPS, không có lỗi HTTP, không còn order PENDING, không còn payment PROCESSING, và Kafka lag về 0 sau khoảng 3 phút drain.

Kết quả này phù hợp với mô hình event-driven Saga: request được nhận nhanh, các consumer xử lý bất đồng bộ, và hệ thống đạt eventual consistency sau giai đoạn drain.
