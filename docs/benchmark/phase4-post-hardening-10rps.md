# Phase 4 Post-Hardening Benchmark - 10 RPS

## Scope

This document records the Phase 4 post-hardening 10 RPS benchmark for `my-ecommerce-platform`.

The purpose is to verify that the system can handle a controlled E2E order workload after the Phase 4 hardening/proof work.

This benchmark was executed after:

- Redis password fallback hardening
- PgBouncer runtime stats audit
- ClickHouse CDC freshness proof
- Istio runtime AuthZ/mTLS proof
- Post-hardening smoke checkpoint
- Post-hardening 5 RPS benchmark

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-10rps-20260606153130

Test script:

    tests/k6/baseline-e2e-10rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 10 requests/second
    duration: 60 seconds
    preAllocatedVUs: 20
    maxVUs: 60

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

    checks_total: 600
    checks_succeeded: 600 / 600
    checks_failed: 0 / 600

Custom metrics:

    accepted_orders: 600
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 600
    http_req_failed: 0.00%
    http_req_duration avg: 33.59ms
    http_req_duration min: 19.2ms
    http_req_duration med: 25.91ms
    http_req_duration max: 512.67ms
    http_req_duration p90: 36.36ms
    http_req_duration p95: 74.03ms

Execution:

    iterations: 600
    effective rate: 9.995325/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, the ClickHouse CDC consumer group had transient lag:

    group: clickhouse-orders-flat-cdc-v2
    topic: cdc_flat.order_db.public.orders
    lag range observed: 1 to 9 across partitions

After a 60-second cooldown, Kafka lag was checked again and returned:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH OBSERVATION

The transient ClickHouse CDC lag drained to zero after cooldown.

## ClickHouse Freshness After Cooldown

ClickHouse CDC table after cooldown:

    table: analytics.orders_flat_cdc_events
    rows: 24645
    max_ingested_at: 2026-06-06 08:32:39.199
    max_source_ts_ms: 1780734753740

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

PgBouncer verdict:

    PASS

## Pod and Resource State

Main workloads remained Running after benchmark.

Observed resource usage examples:

    order-service pods: 46m / 62m CPU, about 47-49Mi memory
    web-gateway pods: 4m / 38m CPU, about 48-50Mi memory
    postgresql: 165m CPU, 251Mi memory
    pgbouncer: 45m CPU, 3Mi memory
    kafka: 98m CPU, 1351Mi memory

No main workload restart anomaly was observed after the run.

## Verdict

PASS WITH OBSERVATION.

Pass:

- 600 accepted orders.
- 0 failed checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 74.03ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Main workloads remained Running.
- ClickHouse had fresh order events after cooldown.
- Kafka lag drained to zero after cooldown.

Observation:

- ClickHouse CDC consumer had small transient lag immediately after the run.
- The lag drained after 60 seconds.
- This should be monitored again at higher RPS.

## Comparison With 5 RPS

Previous 5 RPS result:

    accepted_orders: 300
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 80.74ms

Current 10 RPS result:

    accepted_orders: 600
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    p95 latency: 74.03ms

The 10 RPS run remained stable and did not show API-level degradation compared with the 5 RPS run.

## Notes

This is a controlled post-hardening baseline benchmark.

It does not prove the maximum capacity of the system.

The immediate post-test ClickHouse CDC lag should be monitored again at 20 RPS and above.

## Next Action

Proceed to a 20 RPS post-hardening benchmark using the same evidence pattern:

1. Precheck Git/GitOps/pods.
2. Run k6 benchmark.
3. Check Kafka lag immediately and after cooldown if needed.
4. Check PgBouncer pools.
5. Check ClickHouse freshness.
6. Check pod restarts and resource usage.
