# Phase 4 Post-Hardening Benchmark - 50 RPS

## Scope

This document records the Phase 4 post-hardening 50 RPS benchmark for `my-ecommerce-platform`.

The purpose is to verify the current system capacity under a controlled E2E order workload after Phase 4 hardening and previous 5/10/20/30/40 RPS benchmarks.

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-50rps-20260606172858

Test script:

    tests/k6/baseline-e2e-50rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 50 requests/second
    duration: 60 seconds
    preAllocatedVUs: 100
    maxVUs: 300

Thresholds:

    http_req_failed < 1%
    unexpected_error_rate < 1%
    http_req_duration p95 < 1500ms

## k6 Result

Total result:

    checks_total: 3001
    checks_succeeded: 3001 / 3001
    checks_failed: 0 / 3001

Custom metrics:

    accepted_orders: 3001
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 3001
    http_req_failed: 0.00%
    http_req_duration avg: 39.77ms
    http_req_duration min: 17.28ms
    http_req_duration med: 28.33ms
    http_req_duration max: 819.82ms
    http_req_duration p90: 54.06ms
    http_req_duration p95: 72.47ms

Execution:

    iterations: 3001
    effective rate: 49.903406/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, Kafka had transient lag and one consumer group was rebalancing.

Observed immediate warning:

    inventory-service-group is rebalancing

Observed immediate lag:

    clickhouse-orders-flat-cdc-v2 / cdc_flat.order_db.public.orders:
      lag range: 1 to 6

    inventory-service-group / order.created:
      lag range: 34 to 130

    notification-service-group / payment.completed:
      lag: 1

    order-service-saga-monitor / payment.completed:
      lag range: 2 to 4

    read-model-service-group / payment.completed:
      lag: 1

After 180 seconds cooldown:

    NO_NONZERO_NUMERIC_LAG

After extra 300 seconds stabilization:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH WARNING

The transient lag drained to zero, but inventory-service-group lag and rebalancing show that the async backend is starting to experience noticeable pressure at 50 RPS.

## PgBouncer State

PgBouncer after benchmark:

    order_db cl_active = 28
    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

PgBouncer verdict:

    PASS

No PgBouncer pool backlog was observed.

## ClickHouse Freshness

After cooldown:

    table: analytics.orders_flat_cdc_events
    rows: 41461
    max_ingested_at: 2026-06-06 10:31:24.849
    max_source_ts_ms: 1780741883372

Newest rows included COMPLETED order events from the benchmark window.

ClickHouse verdict:

    PASS

## Scaling Behavior

Immediately after benchmark:

    order-service:
      HPA observed CPU: 58% / 25%
      desired replicas: 5
      deployment: READY 4 / REPLICAS 5

    inventory-consumer:
      HPA metric: 160/20
      desired replicas: 1 initially, then scaled up
      deployment snapshot: READY 1 / REPLICAS 4

After 180 seconds cooldown:

    order-service:
      HPA observed CPU: 4% / 25%
      deployment: READY 4 / REPLICAS 5

    inventory-consumer:
      HPA metric: 2500m/20
      deployment: READY 2 / REPLICAS 8

After extra 300 seconds stabilization:

    order-service:
      HPA observed CPU: 5% / 25%
      deployment: READY 2 / REPLICAS 2

    inventory-consumer:
      HPA metric: 20/20
      deployment: READY 1 / REPLICAS 1

Scaling verdict:

    PASS WITH WARNING

Autoscaling eventually stabilized back to the normal baseline. However, inventory-consumer showed scaling churn and delayed stabilization at 50 RPS.

## Pod and Runtime Warnings

No pod remained Pending/Failed after extra stabilization.

All main workloads returned to Running state.

Runtime warnings observed during the test window:

    inventory-consumer temporary ErrImagePull / ImagePullBackOff
    Docker Hub DNS lookup issue
    vm3-gitops image filesystem warning: 85% of 19.2GiB used

These warnings did not cause the 50 RPS benchmark to fail, but they are operational risks for higher-load tests.

## Order Completion Time

Completion-time query for the 50 RPS run:

    orders: 3001
    min_complete_seconds: 1.406672
    avg_complete_seconds: 36.3669021116294568
    max_complete_seconds: 99.095063

Final order status summary:

    COMPLETED: 3001

Interpretation:

    All accepted orders eventually completed.
    Completion time increased compared with 40 RPS.
    The async backend still drained successfully, but backlog duration became more visible.

## Verdict

PASS WITH WARNING.

Pass:

- 3001 accepted orders.
- 0 failed checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 72.47ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Kafka lag drained to zero after cooldown/stabilization.
- ClickHouse had fresh completed order events.
- All 3001 benchmark orders reached COMPLETED.
- Workloads stabilized after extra cooldown.

Warning:

- inventory-service-group showed rebalancing and order.created lag up to 130 immediately after benchmark.
- inventory-consumer scaling was delayed/churned before stabilizing.
- average completion time increased to about 36.36 seconds.
- max completion time reached about 99.09 seconds.
- temporary image pull/DNS issues occurred.
- vm3-gitops image filesystem reached 85% usage.

## Capacity Interpretation

At this point:

    40 RPS can be treated as a cleaner safe capacity baseline.
    50 RPS can be treated as a passing burst-capacity candidate with operational warnings.

50 RPS is not yet proven to be the maximum capacity of the system.

Further capacity testing should continue with 60 RPS, while monitoring:

- inventory-service-group lag
- inventory-consumer scaling behavior
- order completion time
- node disk/image filesystem pressure
- PgBouncer cl_waiting/maxwait
- PostgreSQL CPU and write pressure

## Next Action

Proceed to 60 RPS only after documenting this result.

If 60 RPS passes but warnings increase, continue to classify capacity as:

    safe rated capacity
    burst capacity
    breaking point
