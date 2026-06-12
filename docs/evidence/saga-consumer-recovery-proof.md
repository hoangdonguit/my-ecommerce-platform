# Saga Consumer Recovery Proof

## 1. Mục tiêu

Tài liệu này ghi nhận quá trình kiểm tra và phục hồi runtime Saga sau khi hệ thống bị sự cố restart/reset.

Mục tiêu:

- Xác minh các consumer Kafka đã rejoin group.
- Kiểm tra lại luồng Saga sau khi CDC/Debezium đã phục hồi.
- Chạy light smoke test để xác nhận order có thể đi từ `PENDING` sang `COMPLETED`.
- Lưu evidence phục vụ báo cáo và vận hành.

## 2. Bối cảnh

Sau khi phục hồi CDC/Debezium và Dynamic Redis Filter, light smoke đầu tiên tạo được order nhưng Saga bị kẹt ở trạng thái `PENDING`.

Triệu chứng ban đầu:

- Order được tạo thành công.
- Order outbox đã `PUBLISHED`.
- Kafka topic `order.created` có message.
- Inventory reservation chưa tạo.
- Payment chưa tạo.
- Notification chưa tạo.
- Order vẫn `PENDING`.

Điều này cho thấy lỗi không còn nằm ở CDC, mà nằm ở runtime consumer/Saga pipeline.

## 3. Kiểm tra trước phục hồi

Các workload chính đều tồn tại trong namespace `default`:

- `order-service`
- `inventory-consumer`
- `payment-consumer`
- `notification-consumer`
- `read-model-service`
- `redis`
- `mongodb`
- `web-gateway`

Các consumer group thật của hệ thống:

- `inventory-service-group`
- `inventory-rollback-group-v2`
- `payment-service-group`
- `notification-service-group`
- `order-service-saga-monitor`
- `read-model-service-group`

Trước khi phục hồi, một số group có tồn tại nhưng chưa thể hiện assignment/offset rõ ràng ở các consumer chính. Vì vậy nhóm thực hiện rollout restart các thành phần Saga runtime để ép consumer rejoin Kafka group.

## 4. Thao tác phục hồi

Nhóm thực hiện rollout restart các deployment sau:

- `inventory-consumer`
- `payment-consumer`
- `notification-consumer`
- `order-service`
- `read-model-service`

Kết quả rollout:

- `inventory-consumer` successfully rolled out
- `payment-consumer` successfully rolled out
- `notification-consumer` successfully rolled out
- `order-service` successfully rolled out
- `read-model-service` successfully rolled out

Sau rollout, các pod consumer mới đều ở trạng thái `Running`.

## 5. Consumer group sau phục hồi

Sau khi chờ consumer rejoin, các group đã có assignment rõ:

### inventory-service-group

Consumer lắng nghe topic:

- `order.created`

Có partition được consume với lag 0, ví dụ:

- partition 2: current offset = 33, log end offset = 33, lag = 0
- partition 5: current offset = 1, log end offset = 1, lag = 0

### payment-service-group

Consumer lắng nghe topic:

- `inventory.reserved`

Có partition được consume với lag 0, ví dụ:

- partition 2: current offset = 1, log end offset = 1, lag = 0
- partition 5: current offset = 1, log end offset = 1, lag = 0

### notification-service-group

Consumer lắng nghe topic:

- `payment.completed`
- `payment.failed`

Có partition được consume với lag 0.

### order-service-saga-monitor

Order service lắng nghe:

- `inventory.failed`
- `payment.failed`
- `payment.completed`

Có partition `payment.completed` được consume với lag 0.

### read-model-service-group

Read model service lắng nghe:

- `payment.completed`

Có partition được consume với lag 0.

## 6. Light smoke test sau phục hồi

Smoke test tạo order mới:

- Order ID: `110b0b08-0364-48d0-ac31-63be2d60f4fa`
- User ID: `smoke-user-002`
- Product ID: `prod-123`
- Payment method: `COD`
- Total amount: `24000000 VND`

Kết quả tạo order:

- HTTP code: `201`
- Order ban đầu: `PENDING`

Sau khi chờ Saga:

- `order.status=PENDING elapsed=0s`
- `order.status=COMPLETED elapsed=6s`

Smoke test kết luận:

    SMOKE TEST PASSED

## 7. Database evidence

Sau smoke test, database ghi nhận đầy đủ các trạng thái kỳ vọng.

### Order

- `status = COMPLETED`
- `total_amount = 24000000.00`
- `payment_method = COD`
- `idempotency_key = smoke-price-20260612150419`

### Order outbox

- `event_type = order.created`
- `status = PUBLISHED`

### Inventory reservation

- `status = RESERVED`

### Payment

- `status = COMPLETED`
- `payment_method = COD`
- `transaction_id = cod_d4102e01-1725-40d6-bb17-6caa1f8825d6`

### Notification

- `event_type = payment.completed`
- `channel = IN_APP`
- `title = Thanh toán thành công`
- `status = SENT`

## 8. Service log evidence

### Inventory consumer

Inventory consumer đã nhận và xử lý event:

    fetched order.created
    received order.created order_id=110b0b08-0364-48d0-ac31-63be2d60f4fa
    processed and committed order.created

Inventory outbox sau đó publish event:

    inventory outbox batch published count=1

### Payment consumer

Payment consumer đã nhận và xử lý event:

    received inventory.reserved order_id=110b0b08-0364-48d0-ac31-63be2d60f4fa
    processed inventory.reserved successfully

Payment outbox sau đó publish event:

    payment outbox batch published count=1

### Notification consumer

Notification consumer đã nhận và xử lý event:

    received payment.completed order_id=110b0b08-0364-48d0-ac31-63be2d60f4fa
    processed payment.completed successfully

Log gửi notification:

    NOTIFICATION SENT

### Order service saga monitor

Order service đã nhận event hoàn tất thanh toán:

    received saga event event_type=payment.completed
    processed and committed saga event
    status=COMPLETED

### Read model service

Read model service đã cập nhật projection:

    upserted order read model order_id=110b0b08-0364-48d0-ac31-63be2d60f4fa

## 9. CDC status confirm

Sau khi phục hồi Saga, CDC vẫn hoạt động bình thường.

Kafka topics CDC tồn tại:

- `cdc.order_db.public.orders`
- `cdc_dynamic.order_db.public.orders`
- `__debezium-heartbeat.cdc.order_db`

Connector status:

- `order-db-orders-connector`: `RUNNING`, task `0` `RUNNING`
- `order-db-orders-dynamic-filter-connector`: `RUNNING`, task `0` `RUNNING`

## 10. Evidence directory

Raw evidence được lưu tại:

    docs/evidence/runs/saga-consumer-recovery-20260612150220

Các file đáng chú ý:

- `05-saga-smoke.txt`
- `post-smoke-lag-inventory-service-group.txt`
- `post-smoke-lag-payment-service-group.txt`
- `post-smoke-lag-notification-service-group.txt`
- `post-smoke-lag-order-service-saga-monitor.txt`
- `post-smoke-lag-read-model-service-group.txt`
- `log-inventory-consumer.txt`
- `log-payment-consumer.txt`
- `log-notification-consumer.txt`
- `log-order-service.txt`
- `log-read-model-service.txt`
- `07-cdc-normal-status.json`
- `08-cdc-dynamic-status.json`

## 11. Kết luận

Sau khi restart các Saga runtime components, consumer group đã rejoin Kafka thành công và light smoke test đã pass.

Luồng nghiệp vụ đã được xác nhận hoạt động lại đầy đủ:

    POST /orders
    -> order.created
    -> inventory RESERVED
    -> inventory.reserved
    -> payment COMPLETED
    -> payment.completed
    -> notification SENT
    -> order COMPLETED
    -> read model upserted

Kết luận vận hành:

- CDC/Debezium/Dynamic Redis Filter đã được phục hồi và GitOps hóa bằng CronJob bootstrap connector.
- Saga runtime đã phục hồi sau rollout restart consumer/order/read-model.
- Hệ thống hiện đã qua light smoke test hậu sự cố.
- Nếu lỗi tương tự tái diễn, ưu tiên kiểm tra consumer group assignment và thực hiện rollout restart các consumer trước khi kết luận lỗi logic nghiệp vụ.
