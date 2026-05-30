# Final E2E 10 RPS Benchmark After Hot Path Fix

## Run

- Run ID: final-e2e-10rps-hotpath-20260531011858
- Gateway URL: http://100.65.255.2:30517
- Scenario: create COD orders through Web Gateway
- Fixes included:
  - Payment Transactional Outbox
  - Order Outbox Reliability and stale published requeue
  - Payment Missing Reconciler
  - Web Gateway CreateOrder timeout increased
  - Removed Redis KEYS invalidation from Order Create hot path

## K6 Result

- Total requests: 601
- Accepted orders: 599
- Failed requests: 2
- HTTP error rate: 0.33%
- Unexpected error rate: 0.33%
- Throughput: about 10 requests/second
- Average latency: 321.69ms
- Median latency: 256.17ms
- p90 latency: 614.77ms
- p95 latency: 1058.63ms
- Max latency: 1684.77ms
- Dropped iterations: 0

## Drain Result

The Saga pipeline drained successfully.

- ORDER_PENDING: 376 -> 202 -> 148 -> 32 -> 0
- ORDER_OUTBOX_OPEN: 0
- INVENTORY_OUTBOX_OPEN: 0
- PAYMENT_OUTBOX_OPEN: 0
- Final marker: FINAL_10RPS_HOTPATH_DRAIN_OK

## Conclusion

The 10 RPS benchmark passed after the hot path fix.

The system accepted almost all requests, stayed below the 1.5s p95 latency target, and eventually drained all pending Saga states to completion.

The remaining 2 HTTP failures should be tracked when running higher-load stress tests, but they are below the 1% threshold for this benchmark.
