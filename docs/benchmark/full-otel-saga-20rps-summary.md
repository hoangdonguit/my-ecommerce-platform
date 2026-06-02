# Full OTel Saga 20 RPS Benchmark Summary

## Run

- Run ID: `full-otel-saga-20rps-20260603025939`
- Scenario: k6 20 RPS, 60 seconds, create COD orders through Web Gateway
- Runtime:
  - Full OTel base tag: `otel-full-saga-20260603005334`
  - Inventory hotfix tag: `otel-full-saga-inventory-hotfix-20260603013722`
  - Payment hotfix tag: `otel-full-saga-payment-hotfix-20260603015846`

## K6 Result

- Total requests: `1201`
- Request rate: `19.94407111669858`
- Accepted orders: `1201`
- Rejected orders: `0`
- HTTP failed rate: `0`
- Unexpected error rate: `0`
- Average latency: `105.00958322731029 ms`
- Median latency: `95.138395 ms`
- p90 latency: `119.752303 ms`
- p95 latency: `145.089741 ms`
- Max latency: `656.282001 ms`
- Dropped iterations: `0`

## Drain Result

- Drain status: `DRAIN_OK`
- Final DB state for this run: `1201 COMPLETED orders`
- Order outbox: `PUBLISHED only`
- Inventory outbox: `PUBLISHED only`
- Payment outbox: `PUBLISHED only`
- Kafka lag after benchmark: `0` for main Saga consumer groups

## Post-benchmark Runtime State

- ArgoCD state after verification: `Synced / Healthy`
- Runtime pods: `Running`, restart `0` for target services
- `order-service` temporarily scaled from `2` to `3` replicas during/after the benchmark and then returned to `2` replicas.
- This is expected autoscaling behavior from the KEDA/HPA setup, not a GitOps drift or service failure.
- Strict post-benchmark error grep did not find real runtime errors.

## Conclusion

The full OpenTelemetry Saga runtime passed the 20 RPS benchmark.

Compared with the earlier 20 RPS hot-path stress run, this result has `0` HTTP failures, `0` unexpected errors, much lower p95 latency, and complete Saga/outbox drain.

## Evidence files

- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/api-key-state-real-20rps.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/api-key-state.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/db-after.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/db-secret-state.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/drain-after.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/full-otel-saga-20rps-preflight-1rps-3s.js`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/full-otel-saga-20rps.js`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/gateway-health-before-real-20rps.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/gateway-health-before.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/k6-finished-at.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/k6-important.json`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/k6-output.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/k6-summary.json`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/kafka-lag-after.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/kafka-lag-before.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/post-benchmark-state.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/preflight-k6-output.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/preflight-k6-summary.json`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/run-started-at.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/runtime-after.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/runtime-before.txt`
- `docs/benchmark/runs/full-otel-saga-20rps-20260603025939/service-errors-after.txt`
