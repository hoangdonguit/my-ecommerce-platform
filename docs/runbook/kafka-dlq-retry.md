# Kafka DLQ / Retry Runbook

## Mục tiêu

Hệ thống sử dụng Kafka để xử lý Saga bất đồng bộ giữa các service. Khi consumer xử lý message lỗi, nếu không có retry hoặc DLQ, message lỗi có thể làm consumer bị kẹt hoặc gây khó khăn khi điều tra nguyên nhân.

Vì vậy hệ thống bổ sung cơ chế lightweight retry và Dead Letter Queue cho các consumer quan trọng.

## Phạm vi triển khai

Các consumer đã bổ sung retry và DLQ:

- payment-consumer
  - Input topic: inventory.reserved
  - DLQ topic: inventory.reserved.dlq

- notification-consumer
  - Input topic: payment.completed
  - DLQ topic: payment.completed.dlq
  - Input topic: payment.failed
  - DLQ topic: payment.failed.dlq

## Cơ chế xử lý

Luồng xử lý mới:

1. Consumer đọc message bằng FetchMessage.
2. Xử lý nghiệp vụ.
3. Nếu lỗi, retry tối đa 3 lần.
4. Nếu vẫn lỗi, publish message gốc sang topic .dlq.
5. Commit offset sau khi xử lý thành công hoặc sau khi đã đưa message lỗi vào DLQ.

## Ý nghĩa

Cơ chế này giúp:

- Tránh consumer bị kẹt vĩnh viễn ở một message lỗi.
- Giữ lại payload lỗi để debug.
- Tăng độ tin cậy cho Saga bất đồng bộ.
- Giúp hệ thống tiếp tục xử lý các message sau.

## Kiểm tra consumer group lag

Chạy lệnh:

    for g in payment-service-group notification-service-group; do
      echo "----- GROUP: $g -----"
      kubectl -n kafka exec kafka-0 -- kafka-consumer-groups.sh \
        --bootstrap-server kafka.kafka.svc.cluster.local:9092 \
        --describe --group "$g" || true
    done

Kỳ vọng bình thường:

- payment-service-group lag = 0
- notification-service-group lag = 0 với topic payment.completed
- payment.failed có thể hiện dấu `-` nếu chưa có message, đây là bình thường

## Kiểm tra DLQ topic

Chạy lệnh:

    for topic in inventory.reserved.dlq payment.completed.dlq payment.failed.dlq; do
      echo "----- DLQ topic: $topic -----"
      kubectl -n kafka exec kafka-0 -- kafka-console-consumer.sh \
        --bootstrap-server localhost:9092 \
        --topic "$topic" \
        --from-beginning \
        --timeout-ms 3000 \
        --max-messages 3 || true
    done

Nếu không có message lỗi, kafka-console-consumer có thể báo TimeoutException và `Processed a total of 0 messages`. Trường hợp này là bình thường, nghĩa là DLQ đang rỗng.

## Kết quả kiểm tra sau triển khai

Smoke test Saga end-to-end sau khi bổ sung DLQ/retry:

- orders.status = COMPLETED
- outbox.status = PUBLISHED
- inventory_reservations.status = RESERVED
- payments.status = COMPLETED
- notifications.status = SENT

DLQ topic không có message lỗi trong kịch bản thành công.

## Kết luận

DLQ/retry mức nhẹ giúp hệ thống an toàn hơn trước lỗi xử lý message, nhưng không làm thay đổi luồng nghiệp vụ chính khi hệ thống hoạt động bình thường.
