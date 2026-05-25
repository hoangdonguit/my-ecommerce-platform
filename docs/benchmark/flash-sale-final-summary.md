# Flash Sale Final Benchmark Summary

## Scenario

- Script: `tests/k6/flash-sale-test.js`
- Product ID: `prod-123`
- Flash sale stock: `100`
- Max virtual users: `200`
- Goal: verify Redis Atomic Stock Gate prevents overselling under concurrent order requests.

## k6 Result

| Metric | Result |
|---|---:|
| Total HTTP requests | 11,039 |
| Accepted orders | 100 |
| Rejected orders | 10,928 |
| Unexpected errors | 11 |
| Accepted rate | 0.90% |
| Rejected rate | 98.99% |
| Unexpected error rate | 0.09% |
| Average request duration | 798.91ms |
| Median request duration | 660.18ms |
| p90 request duration | 1.41s |
| p95 request duration | 1.83s |
| Max request duration | 6.46s |

## Database Verification

| Component | Status | Count |
|---|---|---:|
| orders | COMPLETED | 100 |
| outbox | PUBLISHED | 100 |
| payments | COMPLETED | 100 |
| notifications | SENT | 100 |
| inventory_reservations | RESERVED | 100 |

## Kafka Verification

All checked consumer groups had `LAG = 0` after the scenario completed.

Checked groups:

- `inventory-service-group`
- `payment-service-group`
- `notification-service-group`
- `order-service-saga-monitor`
- `read-model-service-group`

## Redis Verification

| Key | Final value |
|---|---:|
| `flashsale:stock:prod-123` | 0 |

## Interpretation

The flash sale scenario passed the main correctness requirement.

With flash sale stock set to 100, the system accepted exactly 100 orders and rejected the remaining concurrent requests using HTTP 409. PostgreSQL, Kafka Saga, Inventory, Payment, and Notification all processed exactly 100 valid orders. This confirms that the Redis Atomic Stock Gate successfully prevented overselling and protected the downstream database/event pipeline from unnecessary load.

There were 11 transient gateway timeout responses during the spike, corresponding to 0.09% unexpected errors. This did not affect the stock correctness or Saga consistency. The result should be documented as a minor runtime timeout under burst load, while the core flash sale correctness requirement passed.
