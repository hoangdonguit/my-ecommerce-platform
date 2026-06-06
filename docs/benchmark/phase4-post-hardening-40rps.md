# Phase 4 Post-Hardening Benchmark - 40 RPS

## Scope

This document records the Phase 4 post-hardening 40 RPS benchmark for `my-ecommerce-platform`.

The purpose is to verify that the system can handle a controlled E2E order workload after Phase 4 hardening/proof work.

This benchmark was executed after:

- Redis password fallback hardening
- PgBouncer runtime stats audit
- ClickHouse CDC freshness proof
- Istio runtime AuthZ/mTLS proof
- Post-hardening smoke checkpoint
- Post-hardening 5 RPS benchmark
- Post-hardening 10 RPS benchmark
- Post-hardening 20 RPS benchmark
- Post-hardening 30 RPS benchmark

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-40rps-20260606165625

Test script:

    tests/k6/baseline-e2e-40rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 40 requests/second
    duration: 60 seconds
    preAllocatedVUs: 80
    maxVUs: 240

Thresholds:

    http_req_failed < 1%
    unexpected_error_rate < 1%
    http_req_duration p95 < 1500ms

## Precheck

Git and GitOps state before benchmark:

    git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy

Gateway and service health passed before the benchmark.

## k6 Result

Total result:

    checks_total: 2401
    checks_succeeded: 2401 / 2401
    checks_failed: 0 / 2401

Custom metrics:

    accepted_orders: 2401
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 2401
    http_req_failed: 0.00%
    http_req_duration avg: 31.62ms
    http_req_duration min: 17.53ms
    http_req_duration med: 26.41ms
    http_req_duration max: 442.05ms
    http_req_duration p90: 42.52ms
    http_req_duration p95: 56.43ms

Execution:

    iterations: 2401
    effective rate: 39.934318/s
    interrupted iterations: 0

Note:

    The result has 2401 iterations instead of exactly 2400 because k6 constant-arrival-rate scheduling can produce a small timing difference around the 60-second boundary.

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, Kafka had transient lag in multiple areas.

ClickHouse CDC lag:

    group: clickhouse-orders-flat-cdc-v2
    topic: cdc_flat.order_db.public.orders
    lag range observed: 34 to 57 across partitions

Other small transient lag was also observed:

    notification-service-group / payment.completed: lag 1 to 4
    order-service-saga-monitor / payment.completed: lag 1 to 3
    read-model-service-group / payment.completed: lag 1

After a 90-second cooldown, Kafka lag was checked again and returned:

    NO_NONZERO_NUMERIC_LAG

After an extra 180-second stabilization check, Kafka lag still returned:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH OBSERVATION

The transient lag drained to zero after cooldown.

## ClickHouse Freshness After Cooldown

ClickHouse CDC table after cooldown:

    table: analytics.orders_flat_cdc_events
    rows: 35457
    max_ingested_at: 2026-06-06 09:58:08.871
    max_source_ts_ms: 1780739886859

Newest rows after cooldown included COMPLETED order events from the benchmark window.

ClickHouse verdict:

    PASS

## PgBouncer State After Benchmark

PgBouncer pools observed:

    inventory_db
    notification_db
    order_db
    payment_db
    pgbouncer

Pool state:

    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

Observed order_db pool activity:

    order_db cl_active = 22
    order_db cl_waiting = 0
    order_db maxwait = 0

PgBouncer clients were visible through `SHOW CLIENTS`.

PgBouncer verdict:

    PASS

## Pod and Resource State

Main workloads remained Running after benchmark.

Observed resource usage immediately after benchmark:

    order-service pods: 50m / 153m / 171m CPU on visible pods, about 47-54Mi memory
    web-gateway pods: 121m / 44m CPU, about 52-53Mi memory
    postgresql: 531m CPU, 394Mi memory
    pgbouncer: 123m CPU, 5Mi memory
    kafka: 211m CPU, 1363Mi memory

No main workload restart anomaly was observed after the run.

## Scaling State

Immediately after the benchmark:

    order-service HPA:
      target: cpu 25%
      observed: cpu 54% / 25%
      minPods: 2
      maxPods: 8
      replicas: 4

    order-service deployment:
      READY: 4
      REPLICAS: 4
      AVAILABLE: 4

After the 90-second cooldown:

    order-service HPA:
      target: cpu 25%
      observed: cpu 5% / 25%
      replicas: 6

    order-service deployment:
      READY: 4
      REPLICAS: 6
      AVAILABLE: 4

This looked like a delayed HPA scaling decision, so an extra 180-second stabilization check was executed.

After the extra stabilization check:

    order-service HPA:
      target: cpu 25%
      observed: cpu 5% / 25%
      replicas: 2

    order-service deployment:
      READY: 2
      REPLICAS: 2
      AVAILABLE: 2

Scaling verdict:

    PASS WITH OBSERVATION

The order-service scale-up was expected under load. The delayed desired replica count after the first cooldown later stabilized back to the minimum replica count.

## Order Completion Time

A completion-time query was executed for the 40 RPS run.

Result:

    orders: 2401
    min_complete_seconds: 1.425625
    avg_complete_seconds: 21.2022167784256560
    max_complete_seconds: 39.993729

Interpretation:

    Orders did not all complete instantly.
    Some orders completed quickly, but the average completion time was about 21.2 seconds and the slowest completed order took almost 40 seconds.

This explains why the dashboard can appear to show all orders as COMPLETED when opened after the benchmark has already drained.

## Why PENDING Becomes COMPLETED Quickly In This Benchmark

This benchmark uses the internal COD/simulated payment path to focus on backend Saga performance.

The benchmark flow is:

    POST /api/orders
    -> create order as PENDING
    -> publish order.created through outbox/Kafka
    -> inventory consumer reserves stock
    -> payment consumer completes COD/internal payment
    -> payment.completed event is published
    -> order-service saga monitor updates order to COMPLETED
    -> notification/read-model/ClickHouse update asynchronously

Because this benchmark does not call a real external payment gateway, there is no user payment delay, banking callback delay, fraud check, shipping process, or warehouse process.

Therefore, fast transition from PENDING to COMPLETED is expected for this backend benchmark.

This result is valid for measuring:

- order API throughput
- outbox reliability
- Kafka/Saga processing
- inventory reservation
- internal COD/simulated payment completion
- order status update
- notification/read-model/ClickHouse update
- PgBouncer/DB/autoscaling behavior

This result should not be described as real-world payment gateway latency.

## Verdict

PASS WITH OBSERVATION.

Pass:

- 2401 accepted orders.
- 0 failed checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 56.43ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Main workloads remained Running.
- Kafka lag drained to zero after cooldown.
- ClickHouse had fresh COMPLETED order events after cooldown.
- order-service scaled up under load and later stabilized back to 2 replicas.
- All 2401 benchmark orders reached COMPLETED.

Observation:

- ClickHouse CDC transient lag was higher than previous benchmark levels.
- Small transient lag also appeared in notification/order-saga/read-model groups.
- Average order completion time was about 21.2 seconds, with max nearly 40 seconds.
- This benchmark uses COD/internal simulated payment, not real external payment gateway payment.

## Comparison With Previous Post-Hardening Benchmarks

5 RPS result:

    accepted_orders: 300
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 80.74ms

10 RPS result:

    accepted_orders: 600
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 74.03ms

20 RPS result:

    accepted_orders: 1201
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 76.22ms

30 RPS result:

    accepted_orders: 1801
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 49.86ms

40 RPS result:

    accepted_orders: 2401
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 56.43ms

The 40 RPS run remained stable at the API level and stayed far below the latency threshold.

## Notes

This is a controlled post-hardening benchmark, not the maximum capacity proof.

Kafka lag should be monitored carefully at the next higher load level because transient lag is increasing with RPS.

## Next Action

Proceed cautiously to a 50 RPS benchmark.

For 50 RPS, keep the same evidence pattern:

1. Precheck Git/GitOps/pods.
2. Run k6 benchmark.
3. Check Kafka lag immediately.
4. Check PgBouncer pools.
5. Check ClickHouse freshness.
6. Check HPA/KEDA and resource usage.
7. Wait cooldown.
8. Re-check Kafka lag and scaling stabilization.
9. Record order completion time.
