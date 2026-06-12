# Saga Consumer Recovery Runbook

## Mục tiêu

Runbook này dùng khi order bị kẹt ở trạng thái `PENDING` dù order outbox đã `PUBLISHED` và topic `order.created` có message.

## Triệu chứng

Các dấu hiệu thường gặp:

- Order tạo thành công nhưng không chuyển sang `COMPLETED`.
- `order_outbox` có event `order.created` trạng thái `PUBLISHED`.
- Topic `order.created` có message.
- Không có inventory reservation mới.
- Không có payment mới.
- Không có notification mới.

## Kiểm tra nhanh

Kiểm tra workloads:

    kubectl -n default get deploy,pod -o wide

Kiểm tra consumer groups:

    kubectl -n kafka exec kafka-0 -- \
    kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list | sort

Các group chính cần có:

- `inventory-service-group`
- `payment-service-group`
- `notification-service-group`
- `order-service-saga-monitor`
- `read-model-service-group`

Kiểm tra lag từng group:

    for g in inventory-service-group payment-service-group notification-service-group order-service-saga-monitor read-model-service-group; do
      echo "----- $g -----"
      kubectl -n kafka exec kafka-0 -- \
      kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group "$g"
    done

## Phục hồi nhẹ

Nếu consumer group không có assignment rõ hoặc Saga không đi tiếp, restart các runtime components:

    kubectl -n default rollout restart deploy/inventory-consumer
    kubectl -n default rollout restart deploy/payment-consumer
    kubectl -n default rollout restart deploy/notification-consumer
    kubectl -n default rollout restart deploy/order-service
    kubectl -n default rollout restart deploy/read-model-service

Chờ rollout:

    kubectl -n default rollout status deploy/inventory-consumer --timeout=180s
    kubectl -n default rollout status deploy/payment-consumer --timeout=180s
    kubectl -n default rollout status deploy/notification-consumer --timeout=180s
    kubectl -n default rollout status deploy/order-service --timeout=180s
    kubectl -n default rollout status deploy/read-model-service --timeout=180s

## Kiểm tra sau phục hồi

Chờ consumer rejoin:

    sleep 45

Kiểm tra consumer lag:

    for g in inventory-service-group payment-service-group notification-service-group order-service-saga-monitor read-model-service-group; do
      echo "----- $g -----"
      kubectl -n kafka exec kafka-0 -- \
      kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group "$g"
    done

Chạy smoke test:

    export API_KEY="<load-from-secret>"
    GATEWAY_URL="http://100.65.255.2:30517" bash tests/smoke/saga-success.sh

Kỳ vọng:

- `orders.status = COMPLETED`
- `inventory_reservations.status = RESERVED`
- `payments.status = COMPLETED`
- `notifications.status = SENT`

## Lưu ý

Không reset Kafka hoặc truncate database nếu chỉ cần phục hồi consumer group. Ưu tiên rollout restart runtime components trước.

Nếu smoke vẫn fail, đọc log theo thứ tự:

1. `inventory-consumer`
2. `payment-consumer`
3. `notification-consumer`
4. `order-service`
5. `read-model-service`
