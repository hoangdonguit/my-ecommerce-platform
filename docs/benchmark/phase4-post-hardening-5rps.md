# Phase 4 Post-Hardening Benchmark - 5 RPS

## Scope

This document records the Phase 4 post-hardening 5 RPS benchmark for `my-ecommerce-platform`.

The purpose is to verify that the system can still handle a small controlled E2E order workload after the Phase 4 hardening/proof work.

This benchmark was executed after:

- Redis password fallback hardening
- PgBouncer runtime stats audit
- ClickHouse CDC freshness proof
- Istio runtime AuthZ/mTLS proof
- Post-hardening smoke checkpoint

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-5rps-20260606143022

Test script:

    tests/k6/baseline-e2e-5rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 5 requests/second
    duration: 60 seconds
    preAllocatedVUs: 10
    maxVUs: 30

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

    checks_total: 300
    checks_succeeded: 300 / 300
    checks_failed: 0 / 300

Custom metrics:

    accepted_orders: 300
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 300
    http_req_failed: 0.00%
    http_req_duration avg: 36.08ms
    http_req_duration min: 21.33ms
    http_req_duration med: 27.96ms
    http_req_duration max: 289.8ms
    http_req_duration p90: 43.27ms
    http_req_duration p95: 80.74ms

Execution:

    iterations: 300
    effective rate: 4.997187/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, the ClickHouse CDC consumer group had a small transient lag:

    clickhouse-orders-flat-cdc-v2
    topic: cdc_flat.order_db.public.orders
    lag: 1 on several partitions

After a 60-second cooldown, Kafka lag was checked again and returned:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH OBSERVATION

The transient ClickHouse CDC lag drained to zero after cooldown.

## ClickHouse Freshness After Cooldown

ClickHouse CDC table after cooldown:

    table: analytics.orders_flat_cdc_events
    rows: 23445
    max_ingested_at: 2026-06-06 07:31:32.106
    max_source_ts_ms: 1780731086477

Newest rows included completed order events from the benchmark window.

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

PgBouncer verdict:

    PASS

## Pod and Resource State

Main workloads remained Running after benchmark.

Observed resource usage examples:

    order-service pods: 28m / 42m CPU, about 48-49Mi memory
    web-gateway pods: 5m / 23m CPU, about 46-49Mi memory
    postgresql: 113m CPU, 220Mi memory
    pgbouncer: 27m CPU, 3Mi memory
    kafka: 71m CPU, 1349Mi memory

No main workload restart anomaly was observed after the run.

## Verdict

PASS.

The 5 RPS post-hardening benchmark passed.

Evidence summary:

- 300 accepted orders.
- 0 failed checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 80.74ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Kafka transient ClickHouse CDC lag drained to zero after cooldown.
- ClickHouse had fresh completed order events.
- Main workloads remained Running.

## Notes

This is a small post-hardening baseline benchmark.

It does not prove the maximum capacity of the system.

The immediate post-test ClickHouse CDC lag should be monitored again at higher RPS levels, because a small transient lag appeared right after the 5 RPS run and then drained after cooldown.

## Next Action

Proceed to a 10 RPS post-hardening benchmark using the same evidence pattern:

1. Precheck Git/GitOps/pods.
2. Run k6 benchmark.
3. Check Kafka lag immediately and after cooldown if needed.
4. Check PgBouncer pools.
5. Check ClickHouse freshness.
6. Check pod restarts and resource usage.
