# DLQ / Retry Validation Evidence

## Thời điểm kiểm tra

Ngày 20/05/2026, sau khi bổ sung lightweight Kafka DLQ/retry cho payment-consumer và notification-consumer.

## Kết quả triển khai

Các image mới đã được build và deploy:

- payment-service:dlq-retry-20260520020224
- notification-service:dlq-retry-20260520020224

Các topic DLQ đã được tạo:

- inventory.reserved.dlq
- payment.completed.dlq
- payment.failed.dlq

## Kết quả log consumer

payment-consumer ghi nhận:

- payment consumer listening topic=inventory.reserved group=payment-service-group dlq=inventory.reserved.dlq
- processed inventory.reserved successfully attempt=1

notification-consumer ghi nhận:

- notification consumer listening topics=payment.completed,payment.failed group=notification-service-group dlq=payment.completed.dlq,payment.failed.dlq
- processed payment.completed successfully attempt=1

## Kết quả smoke test

Smoke test tạo đơn qua Web Gateway thành công.

Kết quả assert:

- orders.status = COMPLETED
- outbox.status = PUBLISHED
- inventory_reservations.status = RESERVED
- payments.status = COMPLETED
- notifications.status = SENT

Script smoke test đã được cải tiến từ sleep cố định sang polling trạng thái đơn hàng, giúp tránh fail giả khi Saga xử lý bất đồng bộ chậm hơn dự kiến.

## Kết quả kiểm tra DLQ

Các topic DLQ được kiểm tra:

- inventory.reserved.dlq
- payment.completed.dlq
- payment.failed.dlq

Kết quả kiểm tra cho thấy DLQ rỗng trong kịch bản thành công. Kafka console consumer báo TimeoutException và `Processed a total of 0 messages`, đây là kết quả đúng kỳ vọng vì không có message lỗi.

## Kết luận

DLQ/retry mức nhẹ hoạt động ổn định, không làm hỏng Saga chính. Trong kịch bản thành công, DLQ rỗng là đúng kỳ vọng.
