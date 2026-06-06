# Phase 4 Post-Hardening Benchmark - 20 RPS

## Scope

This document records the Phase 4 post-hardening 20 RPS benchmark for `my-ecommerce-platform`.

The purpose is to verify that the system can handle a controlled E2E order workload after Phase 4 hardening/proof work.

This benchmark was executed after:

- Redis password fallback hardening
- PgBouncer runtime stats audit
- ClickHouse CDC freshness proof
- Istio runtime AuthZ/mTLS proof
- Post-hardening smoke checkpoint
- Post-hardening 5 RPS benchmark
- Post-hardening 10 RPS benchmark

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-20rps-20260606154630

Test script:

    tests/k6/baseline-e2e-20rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 20 requests/second
    duration: 60 seconds
    preAllocatedVUs: 40
    maxVUs: 120

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

    checks_total: 1201
    checks_succeeded: 1201 / 1201
    checks_failed: 0 / 1201

Custom metrics:

    accepted_orders: 1201
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 1201
    http_req_failed: 0.00%
    http_req_duration avg: 33.69ms
    http_req_duration min: 18.27ms
    http_req_duration med: 26.32ms
    http_req_duration max: 577.64ms
    http_req_duration p90: 45.74ms
    http_req_duration p95: 76.22ms

Execution:

    iterations: 1201
    effective rate: 19.975881/s
    interrupted iterations: 0

Note:

    The result has 1201 iterations instead of exactly 1200 because k6 constant-arrival-rate scheduling can produce a small timing difference around the 60-second boundary.

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, the ClickHouse CDC consumer group had transient lag:

    group: clickhouse-orders-flat-cdc-v2
    topic: cdc_flat.order_db.public.orders
    lag range observed: 3 to 19 across partitions

After a 60-second cooldown, Kafka lag was checked again and returned:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH OBSERVATION

The transient ClickHouse CDC lag drained to zero after cooldown.

## ClickHouse Freshness After Cooldown

ClickHouse CDC table after cooldown:

    table: analytics.orders_flat_cdc_events
    rows: 27049
    max_ingested_at: 2026-06-06 08:47:39.657
    max_source_ts_ms: 1780735654416

Newest rows included PENDING and COMPLETED order events from the benchmark window.

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

PgBouncer clients were visible through `SHOW CLIENTS`.

One `order_db` client was active during inspection and had a small `wait_us` value, but `cl_waiting` and `maxwait` remained zero. This is treated as an in-flight inspection/runtime observation, not pool backlog.

PgBouncer verdict:

    PASS

## Pod and Resource State

Main workloads remained Running after benchmark.

Observed resource usage examples:

    order-service pods: 83m / 55m / 88m CPU, about 47-51Mi memory
    web-gateway pods: 61m / 24m CPU, about 49-50Mi memory
    postgresql: 423m CPU, 287Mi memory
    pgbouncer: 78m CPU, 4Mi memory
    kafka: 177m CPU, 1355Mi memory

Runtime observation:

    order-service had 3 running pods during the post-benchmark check.

No main workload restart anomaly was observed after the run.

## Verdict

PASS WITH OBSERVATION.

Pass:

- 1201 accepted orders.
- 0 failed checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 76.22ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Main workloads remained Running.
- ClickHouse had fresh order events after cooldown.
- Kafka lag drained to zero after cooldown.

Observation:

- ClickHouse CDC consumer had transient lag immediately after the run.
- The lag range was 3 to 19 across partitions.
- The lag drained after 60 seconds.
- order-service had 3 running pods during the post-benchmark check.

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

The 20 RPS run remained stable and did not show API-level degradation compared with the 5 RPS and 10 RPS runs.

## Notes

This is a controlled post-hardening benchmark, not the maximum capacity proof.

The immediate ClickHouse CDC lag should be monitored again at higher RPS levels.

## Next Action

Before increasing to 30 RPS or 50 RPS, inspect scaling status and resource configuration:

1. Check HPA/KEDA state.
2. Check order-service replica behavior.
3. Check consumer lag/drain behavior.
4. Confirm PgBouncer pool state remains healthy.

## Post-Benchmark Scaling and Resource Check

A scaling/resource inventory was collected after the 20 RPS benchmark.

Raw local evidence:

    .local-notes/benchmark/phase4-post-hardening-20rps-20260606154630/scale-resource-inventory-after-20rps.txt

HPA/KEDA state:

    order-service:
      scaler: keda-hpa-order-service-scaler
      trigger: CPU
      target: 25%
      minPods: 2
      maxPods: 8
      current steady replicas after cooldown: 2

    inventory-api:
      trigger: CPU
      minPods: 1
      maxPods: 8

    payment-api:
      trigger: CPU
      minPods: 1
      maxPods: 8

    notification-api:
      trigger: CPU
      minPods: 1
      maxPods: 8

    inventory-consumer:
      trigger: Kafka
      minPods: 1
      maxPods: 16

    payment-consumer:
      trigger: Kafka
      minPods: 1
      maxPods: 16

    notification-consumer:
      trigger: Kafka
      minPods: 1
      maxPods: 16

Scaling event observed:

    order-service scaled from 2 to 3 replicas during/after the 20 RPS load
    reason: CPU resource utilization above target

Scale-down event observed:

    order-service scaled from 3 back to 2 replicas after cooldown
    reason: all metrics below target

Resource state after cooldown:

    order-service pods: about 14m CPU each, about 48-49Mi memory
    web-gateway pods: about 5m CPU each, about 47Mi memory
    postgresql: about 19m CPU, 281Mi memory
    pgbouncer: about 8m CPU, 5Mi memory
    kafka: about 37m CPU, 1355Mi memory

Scaling/resource verdict:

    PASS

The autoscaling behavior after the 20 RPS benchmark was expected. The temporary third order-service pod was created because CPU utilization exceeded the target and was later removed after the load dropped.

This confirms that it is safe to proceed to a 30 RPS benchmark before attempting 50 RPS.
