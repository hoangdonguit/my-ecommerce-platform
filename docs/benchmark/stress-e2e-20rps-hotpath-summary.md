# Stress E2E 20 RPS After Hot Path Fix

## Run

- Run ID: stress-e2e-20rps-hotpath-20260531023221
- Gateway URL: http://100.65.255.2:30517
- Scenario: create COD orders through Web Gateway
- Purpose: stress test after reliability and hot path fixes

## K6 Result

- Total requests: 1184
- Accepted orders: 1111
- Failed requests: 73
- HTTP error rate: 6.16%
- Unexpected error rate: 6.16%
- Throughput: about 15.85 requests/second
- Average latency: 1433.20ms
- Median latency: 838.10ms
- p90 latency: 2478.80ms
- p95 latency: 2887.20ms
- Max latency: 15149.42ms
- Dropped iterations: 17

## Threshold Result

This run did not pass the strict HTTP benchmark thresholds:

- p95 latency target: below 1500ms
- actual p95 latency: 2887.20ms
- HTTP error target: below 1%
- actual HTTP error rate: 6.16%

## Drain Result

The Saga pipeline eventually drained successfully.

- ORDER_PENDING: 988 -> 784 -> 654 -> 573 -> 465 -> 333 -> 206 -> 94 -> 19 -> 0
- ORDER_OUTBOX_OPEN: 0
- INVENTORY_OUTBOX_OPEN: 0
- PAYMENT_OUTBOX_OPEN: 0
- Final marker: STRESS_20RPS_HOTPATH_DRAIN_OK

## Final DB State

- Orders COMPLETED: 12468
- Orders FAILED: 4
- Order outbox: PUBLISHED only
- Inventory outbox: PUBLISHED only
- Payment outbox: PUBLISHED only

## Conclusion

The 20 RPS run is a stress-limit finding, not a pass benchmark.

Compared with the earlier failed 20 RPS run, the hot path fix significantly reduced HTTP failures. However, the synchronous HTTP layer still does not meet the strict latency and error-rate targets at 20 RPS.

The important correctness result is that the Saga pipeline eventually recovered and drained all pending orders to completion.
