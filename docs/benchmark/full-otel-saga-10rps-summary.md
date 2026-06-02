# Full OTel Saga 10 RPS Benchmark Summary

## Run

- Run ID: `full-otel-saga-10rps-20260603022306`
- Scenario: k6 10 RPS, 60 seconds, create COD orders through Web Gateway
- Runtime:
  - Full OTel base tag: `otel-full-saga-20260603005334`
  - Inventory hotfix tag: `otel-full-saga-inventory-hotfix-20260603013722`
  - Payment hotfix tag: `otel-full-saga-payment-hotfix-20260603015846`

## K6 Result

- Total requests: `600`
- Request rate: `9.98366993588484`
- Accepted orders: `600`
- Rejected orders: `0`
- HTTP failed rate: `0`
- Unexpected error rate: `0`
- Average latency: `103.93881688833329 ms`
- Median latency: `95.0980625 ms`
- p90 latency: `106.9196407 ms`
- p95 latency: `115.31035684999996 ms`
- Max latency: `891.718644 ms`
- Dropped iterations: `0`

## Drain Result

- Drain status: `DRAIN_OK`
- Final DB state for this run: `600 COMPLETED orders`
- Kafka lag after benchmark: `0` for the main Saga consumer groups
- Runtime after benchmark: ArgoCD `Synced / Healthy`, all target pods `Running`

## Conclusion

The full OpenTelemetry Saga runtime passed the 10 RPS benchmark.

The system accepted all 600 requests, produced no HTTP or unexpected errors, kept p95 latency far below the 1500ms threshold, and drained all Saga/outbox work successfully.

## Evidence files

- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/run-started-at.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/runtime-before.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/gateway-health-before.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/db-before.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/kafka-lag-before.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/full-otel-saga-10rps.js`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/k6-output.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/k6-summary.json`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/k6-important.json`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/drain-after.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/db-after.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/kafka-lag-after.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/runtime-after.txt`
- `docs/benchmark/runs/full-otel-saga-10rps-20260603022306/service-errors-after.txt`
