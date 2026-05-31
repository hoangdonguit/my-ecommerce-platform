# Sanity E2E 10 RPS After Warm Capacity

## Run

- Run ID: sanity-e2e-10rps-warmcapacity-20260531225832
- Gateway URL: http://100.65.255.2:30517
- Scenario: create COD orders through Web Gateway
- Change included:
  - web-gateway warm replicas: 2
  - order-service warm replicas: 2
  - order-service KEDA minReplicaCount: 2

## K6 Result

- Total requests: 600
- Accepted orders: 600
- Failed requests: 0
- HTTP error rate: 0%
- Unexpected error rate: 0%
- Throughput: about 10 requests/second
- Average latency: 182.47ms
- Median latency: 126.13ms
- p90 latency: 430.26ms
- p95 latency: 623.86ms
- Max latency: 1317.73ms
- Dropped iterations: 0

## Drain Result

The Saga pipeline drained successfully.

- ORDER_PENDING: 77 -> 75 -> 6 -> 0
- ORDER_OUTBOX_OPEN: 0
- INVENTORY_OUTBOX_OPEN: 0
- PAYMENT_OUTBOX_OPEN: 0
- Final marker: SANITY_10RPS_WARMCAPACITY_DRAIN_OK

Note: the drain command was started slightly after K6 completed, so the drain timing is approximate. The final correctness result is valid.

## Conclusion

The 10 RPS sanity benchmark passed after warm capacity tuning.

Compared with the previous 10 RPS run, p95 latency improved significantly and HTTP failures dropped to 0%.
