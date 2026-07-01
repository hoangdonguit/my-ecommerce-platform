# 02 - Order Flow

## Main command path

```text
Client / k6
-> web-gateway
-> order-service
-> PostgreSQL + Transactional Outbox
-> Kafka order.created
-> inventory-consumer
-> payment-consumer
-> notification-consumer
-> order status COMPLETED / FAILED / CANCELLED
```

## Important patterns

### Saga Choreography

The system does not use one distributed ACID transaction across all services. Each service commits its local transaction and emits domain events for the next step.

### Transactional Outbox

Order and event data are persisted together before publishing to Kafka. This reduces the risk of committing business data without publishing the corresponding event.

### Idempotency

The order path includes idempotency handling so duplicated client requests or retries do not create duplicated logical orders.

### Kafka lag

HTTP success does not mean the full asynchronous workflow is complete. Kafka consumer lag and order status must be checked to confirm that the workflow has converged.

### CDC and ClickHouse

CDC, Debezium, Kafka Connect, Dynamic Redis Filter, and ClickHouse are read/analytics side paths. They do not directly make PENDING orders become COMPLETED.
