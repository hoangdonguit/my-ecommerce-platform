# Load Test After Upgrade Summary

## Test scenario

- Script: `tests/k6/load-test.js`
- Max virtual users: 50
- Total created orders: 20,989
- Architecture: Web Gateway -> Order Service -> PostgreSQL Outbox -> Kafka -> Inventory -> Payment -> Notification -> MongoDB Read Model

## k6 result

| Metric | Result |
|---|---:|
| HTTP request success rate | 100.00% |
| HTTP failed rate | 0.00% |
| Total orders created | 20,989 |
| Average request duration | 52.8ms |
| Median request duration | 47.23ms |
| p90 request duration | 69.85ms |
| p95 request duration | 82.74ms |
| Max request duration | 933.62ms |

## Database result right after load

| Table / Service | Status | Count |
|---|---|---:|
| orders | COMPLETED | 11,804 |
| orders | PENDING | 9,185 |
| outbox | PUBLISHED | 20,989 |
| payments | COMPLETED | 11,822 |
| notifications | SENT | 11,807 |
| inventory_reservations | RESERVED | 20,989 |

## Database result after Saga drain

| Table / Service | Status | Count |
|---|---|---:|
| orders | COMPLETED | 20,989 |
| payments | COMPLETED | 20,989 |
| notifications | SENT | 20,989 |
| completed_payments source count | COMPLETED | 20,989 |

## Kafka lag after drain

All main consumer groups were drained after waiting for the Saga pipeline to finish.

| Consumer group | Result |
|---|---|
| inventory-service-group | Lag cleared |
| payment-service-group | Lag cleared |
| notification-service-group | Lag cleared |
| order-service-saga-monitor | Lag cleared |
| read-model-service-group | Lag cleared |

## MongoDB Read Model

| Metric | Count |
|---|---:|
| PostgreSQL completed payments | 20,989 |
| MongoDB order_read_models | 20,987 |
| Missing read-model records | 2 |
| Read-model sync ratio | 99.9905% |

Missing order IDs:

- `5d9ce54b-f643-44de-989e-9c626cdd6b39`
- `99f3d923-701b-480e-81cc-46b37e14bad9`

MongoDB Read Model is an asynchronous CQRS projection built from `payment.completed` events. PostgreSQL remains the transactional source of truth. The load test confirms that all orders, payments, and notifications completed successfully in PostgreSQL. The read model reached 20,987 out of 20,989 records, which indicates that the projection layer is almost fully synchronized but still needs a reconciliation or backfill mechanism to guarantee 100% projection consistency after large benchmark runs or consumer restarts.

## Interpretation

The upgraded system successfully accepted all load-test order creation requests with zero HTTP failures and low p95 latency. The order ingestion path, Web Gateway, Order Service, PgBouncer, PostgreSQL, and Outbox remained stable under load.

Because the system uses an asynchronous Saga architecture, not all orders became `COMPLETED` immediately after the load test ended. Inventory reservation completed for all orders first, while payment processing continued to drain the backlog. After additional drain time, all 20,989 orders reached `COMPLETED`, all 20,989 payments were `COMPLETED`, and all 20,989 notifications were `SENT`.

This demonstrates successful high-throughput ingestion with eventual consistency across the Saga pipeline. The small MongoDB Read Model difference should be documented as a projection consistency limitation and can be addressed by adding a read-model reconciliation/backfill job.
