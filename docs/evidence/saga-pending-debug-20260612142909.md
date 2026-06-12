# Saga Pending Investigation

## Context

Sau khi phục hồi CDC/Debezium và Dynamic Redis Filter, light smoke tạo được order mới nhưng Saga không đi tiếp sau `order.created`.

## Symptom

- Order được tạo thành công.
- Order outbox đã `PUBLISHED`.
- Kafka topic `order.created` có message.
- Inventory reservation chưa xuất hiện.
- Payment và notification chưa xuất hiện.
- Order vẫn ở trạng thái `PENDING`.

## Evidence directory

`docs/evidence/runs/saga-pending-debug-20260612142909`

## Next analysis target

Ưu tiên kiểm tra `inventory-service-group`, deployment env của `inventory-consumer`, log runtime của `inventory-consumer`, và offset của topic `order.created`.
