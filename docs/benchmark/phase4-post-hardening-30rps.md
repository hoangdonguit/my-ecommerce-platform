# Phase 4 Post-Hardening Benchmark - 30 RPS

## Scope

This document records the Phase 4 post-hardening 30 RPS benchmark for `my-ecommerce-platform`.

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
- 20 RPS scaling/resource observation

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-30rps-20260606162435

Test script:

    tests/k6/baseline-e2e-30rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 30 requests/second
    duration: 60 seconds
    preAllocatedVUs: 60
    maxVUs: 180

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

    checks_total: 1801
    checks_succeeded: 1801 / 1801
    checks_failed: 0 / 1801

Custom metrics:

    accepted_orders: 1801
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 1801
    http_req_failed: 0.00%
    http_req_duration avg: 32.04ms
    http_req_duration min: 18.49ms
    http_req_duration med: 26.73ms
    http_req_duration max: 696.16ms
    http_req_duration p90: 40.25ms
    http_req_duration p95: 49.86ms

Execution:

    iterations: 1801
    effective rate: 29.947567/s
    interrupted iterations: 0

Note:

    The result has 1801 iterations instead of exactly 1800 because k6 constant-arrival-rate scheduling can produce a small timing difference around the 60-second boundary.

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, Kafka had transient lag in two areas.

ClickHouse CDC lag:

    group: clickhouse-orders-flat-cdc-v2
    topic: cdc_flat.order_db.public.orders
    lag range observed: 11 to 22 across partitions

Payment consumer lag:

    group: payment-service-group
    topic: inventory.reserved
    lag range observed: 2 to 6 across partitions

After a 60-second cooldown, Kafka lag was checked again and returned:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH OBSERVATION

The transient lag drained to zero after cooldown.

## ClickHouse Freshness After Cooldown

ClickHouse CDC table after cooldown:

    table: analytics.orders_flat_cdc_events
    rows: 30653
    max_ingested_at: 2026-06-06 09:25:58.070
    max_source_ts_ms: 1780737951926

Newest rows included COMPLETED order events from the benchmark window.

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

    order_db cl_active = 20
    order_db cl_waiting = 0
    order_db maxwait = 0

PgBouncer clients were visible through `SHOW CLIENTS`.

PgBouncer verdict:

    PASS

## Pod and Resource State

Main workloads remained Running after benchmark.

Observed resource usage examples:

    order-service pods: 115m / 74m / 95m / 129m CPU, about 45-53Mi memory
    web-gateway pods: 49m / 60m CPU, about 52Mi memory
    postgresql: 506m CPU, 340Mi memory
    pgbouncer: 106m CPU, 5Mi memory
    kafka: 201m CPU, 1358Mi memory

No main workload restart anomaly was observed after the run.

## Scaling State

HPA/KEDA state after benchmark:

    order-service HPA:
      target: cpu 25%
      observed: cpu 45% / 25%
      minPods: 2
      maxPods: 8
      replicas: 4

    order-service deployment:
      READY: 4
      REPLICAS: 4
      AVAILABLE: 4

Other API deployments stayed at their normal replica counts:

    inventory-api: 1
    payment-api: 1
    notification-api: 1
    read-model-service: 1
    web-gateway: 2

Scaling verdict:

    PASS WITH OBSERVATION

The order-service scale-up to 4 replicas was expected because CPU exceeded the HPA target.

## Verdict

PASS WITH OBSERVATION.

Pass:

- 1801 accepted orders.
- 0 failed checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 49.86ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Main workloads remained Running.
- ClickHouse had fresh order events after cooldown.
- Kafka lag drained to zero after cooldown.

Observation:

- ClickHouse CDC consumer had transient lag immediately after the run.
- Payment consumer also had small transient lag immediately after the run.
- Both drained after 60 seconds.
- order-service scaled up to 4 replicas because CPU exceeded the HPA target.

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

The 30 RPS run remained stable at the API level and stayed far below the latency threshold.

## Notes

This is a controlled post-hardening benchmark, not the maximum capacity proof.

The immediate Kafka lag should be monitored carefully at the next higher load level because 30 RPS introduced transient lag in both ClickHouse CDC and payment-service-group.

## Next Action

Before attempting 50 RPS, run one of the following:

1. A 40 RPS intermediate benchmark, or
2. A short scaling/lag inventory after 30 RPS cooldown.

Jumping directly from 30 RPS to 50 RPS is possible, but 40 RPS is safer because Kafka transient lag started to involve the payment consumer at 30 RPS.
