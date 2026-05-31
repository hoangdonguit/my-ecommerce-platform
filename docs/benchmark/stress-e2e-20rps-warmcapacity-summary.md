# E2E 20 RPS Benchmark After Warm Capacity

## Run

- Run ID: stress-e2e-20rps-warmcapacity-20260531231206
- Gateway URL: http://100.65.255.2:30517
- Scenario: create COD orders through Web Gateway
- Change included:
  - web-gateway warm replicas: 2
  - order-service warm replicas: 2
  - order-service KEDA minReplicaCount: 2

## K6 Result

- Total requests: 1200
- Accepted orders: 1200
- Failed requests: 0
- HTTP error rate: 0%
- Unexpected error rate: 0%
- Throughput: about 20 requests/second
- Average latency: 373.02ms
- Median latency: 232.33ms
- p90 latency: 924.45ms
- p95 latency: 1226.50ms
- Max latency: 2018.01ms
- Dropped iterations: 0

## Drain Result

The Saga pipeline drained successfully.

- ORDER_PENDING: 1122 -> 1097 -> 993 -> 933 -> 704 -> 500 -> 325 -> 305 -> 182 -> 45 -> 0
- ORDER_OUTBOX_OPEN: 0
- INVENTORY_OUTBOX_OPEN: 0
- PAYMENT_OUTBOX_OPEN: 0
- Final marker: STRESS_20RPS_WARMCAPACITY_DRAIN_OK

During drain check 2, PostgreSQL temporarily rejected connections and reported recovery/not accepting connections. The pod did not restart, had Restart Count 0, and became available again in the next checks. This is recorded as a transient infrastructure warning, not a Saga failure.

## Final DB State

- Orders COMPLETED: 14268
- Orders FAILED: 4
- Order outbox: PUBLISHED only
- Inventory outbox: PUBLISHED only
- Payment outbox: PUBLISHED only

## Conclusion

The 20 RPS benchmark passed after warm capacity tuning.

Compared with the previous 20 RPS stress run, HTTP failures dropped to 0%, p95 latency improved below the 1.5s target, and the Saga pipeline eventually drained all pending work to completion.

The remaining infrastructure note is that PostgreSQL briefly rejected connections during drain, so future optimization can focus on PostgreSQL/PgBouncer resource tuning.
