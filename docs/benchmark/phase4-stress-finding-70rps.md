# Phase 4 Stress Finding - 70 RPS

## Scope

This document records the Phase 4 70 RPS stress-finding run for `my-ecommerce-platform`.

This run is not classified as a stable capacity benchmark. It is used to identify the next bottleneck after the 60 RPS capacity-warning point.

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-stress-finding-70rps-20260606182928

Test script:

    tests/k6/baseline-e2e-70rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 70 requests/second
    duration: 60 seconds
    preAllocatedVUs: 140
    maxVUs: 420

Thresholds:

    http_req_failed < 1%
    unexpected_error_rate < 1%
    http_req_duration p95 < 1500ms

## k6 Result

Total result:

    checks_total: 4201
    checks_succeeded: 4201 / 4201
    checks_failed: 0 / 4201

Custom metrics:

    accepted_orders: 4201
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 4201
    http_req_failed: 0.00%
    http_req_duration avg: 49.37ms
    http_req_duration min: 18.12ms
    http_req_duration med: 31.39ms
    http_req_duration max: 567.33ms
    http_req_duration p90: 93.85ms
    http_req_duration p95: 152.17ms

Execution:

    iterations: 4201
    effective rate: 69.867507/s
    interrupted iterations: 0

k6 verdict:

    API PASS

The HTTP/API hot path still accepted orders successfully at 70 RPS.

## Kafka Lag

Immediately after the stress run, Kafka had clear transient lag.

Observed immediate lag:

    clickhouse-orders-flat-cdc-v2 / cdc_flat.order_db.public.orders:
      lag range: 34 to 48

    inventory-service-group / order.created:
      lag range: 103 to 125 in visible partitions

    payment-service-group / inventory.reserved:
      lag range: 11 to 166

Also observed:

    payment-service-group is rebalancing

After 420 seconds cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    EVENTUAL DRAIN PASS WITH STRESS WARNING

Kafka eventually drained, but the immediate lag was higher and involved core Saga consumer groups.

## PgBouncer State

PgBouncer after stress:

    inventory_db cl_active = 3
    notification_db cl_active = 2
    order_db cl_active = 28
    payment_db cl_active = 5

Pool waiting:

    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

PgBouncer verdict:

    PASS

PgBouncer did not show pool backlog. The 70 RPS bottleneck is not PgBouncer waiting.

## Scaling and Scheduling Behavior

Immediately after stress:

    order-service HPA:
      observed CPU: 51% / 25%
      desired replicas: 6

    inventory-consumer:
      deployment: READY 1 / REPLICAS 4

    payment-consumer:
      deployment: READY 2 / REPLICAS 4

Pending pods immediately after stress:

    inventory-consumer pending pods
    order-service pending pods
    payment-consumer pending pods

After 420 seconds cooldown:

    Kafka lag: NO_NONZERO_NUMERIC_LAG

However, HPA/deployment state still showed delayed scale pressure:

    notification-consumer:
      HPA metric: 2500m/20
      deployment: READY 2 / REPLICAS 8

    payment-consumer:
      HPA metric: 2500m/20
      deployment: READY 4 / REPLICAS 8

    order-service:
      deployment: READY 2 / REPLICAS 2

Pods not Running/Completed after cooldown included:

    notification-consumer pending pods
    payment-consumer pending pods

Scheduling events repeatedly showed:

    0/3 nodes are available:
      1 Insufficient cpu
      2 node(s) didn't match Pod's node affinity/selector

The cluster also had image filesystem warnings on vm3-gitops:

    Insufficient free disk space on the node's image filesystem
    85% of 19.2 GiB used

Scaling/scheduling verdict:

    OPERATIONAL STRESS CONFIRMED

The system accepted and eventually completed all orders, but Kubernetes could not schedule all desired scaled pods under current fixed resources and placement constraints.

## Resource Snapshot

Immediately after stress, node usage included:

    vm1-gateway: 1553m CPU, 5604Mi memory
    vm2-mesh: 1683m CPU, 4869Mi memory
    vm3-gitops: 2106m CPU, 6585Mi memory

Selected pod usage:

    postgresql: 696m CPU, 491Mi memory
    pgbouncer: 159m CPU, 5Mi memory
    kafka: 223m CPU, 1414Mi memory
    clickhouse: 155m CPU, 831Mi memory
    kafka-connect-debezium: 55m CPU, 510Mi memory

## Order Completion Time

Order status after cooldown:

    COMPLETED: 4201

Completion-time query:

    orders: 4201
    min_complete_seconds: 1.257510
    avg_complete_seconds: 76.2476001018805046
    max_complete_seconds: 150.682006

Interpretation:

    All accepted orders eventually completed.
    However, async completion time increased sharply compared with lower RPS runs.

## Verdict

API PASS, EVENTUAL COMPLETION PASS, OPERATIONAL STRESS / FIXED-RESOURCE LIMIT CONFIRMED.

Pass:

- 4201 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 152.17ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Kafka lag eventually drained to zero.
- All 4201 orders reached COMPLETED.

Stress findings:

- Core consumer groups had significant immediate Kafka lag.
- payment-service-group entered rebalancing.
- Average completion time increased to about 76.25 seconds.
- Maximum completion time reached about 150.68 seconds.
- Multiple desired scaled pods remained Pending.
- FailedScheduling was caused by Insufficient CPU and node affinity/selector constraints.
- vm3-gitops image filesystem pressure remained an operational risk.

## Capacity Interpretation

70 RPS is not a stable capacity rating for the current fixed-resource cluster.

Current interpretation:

    40 RPS = safe rated capacity
    50 RPS = burst capacity with warning
    60 RPS = overload/capacity-warning zone
    70 RPS = stress-finding / fixed-resource limit confirmation

This run confirms that the API hot path can accept 70 RPS, but the full system cannot be described as stable at 70 RPS because scheduling and async drain pressure are too high.

## Next Action

Do not continue to 80 RPS as a normal benchmark.

Before any higher stress test:

1. Fix or document node affinity/selector placement constraints.
2. Clean or expand vm3-gitops image filesystem.
3. Review KEDA scale targets for consumer workloads.
4. Consider reducing requests for lightweight consumers or increasing node CPU capacity.
5. Write a capacity-rating methodology document for the report.
