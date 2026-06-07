# Phase 5 Post-Placement Benchmark - 70 RPS

## Scope

This document records the 70 RPS benchmark after the Phase 5 node placement fix.

The goal is to compare the new result against the previous Phase 4 70 RPS stress-finding run.

## Related Fix

Node placement fix:

    0586c4c capacity: relax stateless service node placement

The scalable stateless workloads were changed from hard `nodeSelector: vm2-mesh` to node affinity allowing:

    vm1-gateway
    vm2-mesh

Stateful/data/GitOps workloads were intentionally not moved to vm1/vm2.

## Environment

Runtime URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase5-post-placement-70rps-20260607005545

Test script:

    tests/k6/baseline-e2e-70rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 70 requests/second
    duration: 60 seconds
    preAllocatedVUs: 140
    maxVUs: 420

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
    http_req_duration avg: 52.57ms
    http_req_duration min: 17.58ms
    http_req_duration med: 36.11ms
    http_req_duration max: 812.73ms
    http_req_duration p90: 104.25ms
    http_req_duration p95: 144.11ms

Execution:

    iterations: 4201
    effective rate: 69.853571/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after benchmark, Kafka had visible transient lag.

ClickHouse CDC lag:

    cdc_flat.order_db.public.orders:
      lag range: 22 to 40

Inventory consumer lag:

    inventory-service-group / order.created:
      visible lag range: 128 to 302

Payment consumer lag:

    payment-service-group / inventory.reserved:
      visible lag range: 3 to 106

Also observed:

    payment-service-group is rebalancing

After cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH WARNING

Kafka eventually drained, but core Saga consumer lag was visible immediately after the benchmark.

## PgBouncer State

PgBouncer after benchmark:

    inventory_db cl_waiting = 0
    notification_db cl_waiting = 0
    order_db cl_waiting = 0
    payment_db cl_waiting = 0
    maxwait = 0

PgBouncer verdict:

    PASS

PgBouncer was not the bottleneck in this run.

## HPA / KEDA Behavior

Immediately after benchmark:

    order-service:
      CPU: 65% / 25%
      replicas: 7

    inventory-consumer:
      metric: 160 / 20
      later scaled up to 8

    payment-consumer:
      metric: 320 / 20
      later scaled up to 16

    notification-consumer:
      later scaled up to 8

After cooldown:

    order-service returned to 2 replicas
    inventory-consumer returned to 1 replica
    payment-consumer returned to 1 replica
    notification-consumer returned to 1 replica
    Kafka lag returned to zero
    no abnormal non-Running pod remained

## Scheduling Warning

Despite the improved placement, FailedScheduling still appeared for consumer scale-out pods.

Observed pattern:

    0/3 nodes are available:
      1 node(s) didn't match Pod's node affinity/selector
      2 Insufficient cpu

Interpretation:

    vm3-gitops is intentionally excluded from stateless service placement.
    vm1-gateway and vm2-mesh are the only eligible nodes.
    At 70 RPS, KEDA requested more consumer replicas than vm1/vm2 could schedule under current CPU requests and existing workloads.

This means the previous hard vm2-only placement bottleneck was improved, but the cluster still hits CPU scheduling limits under aggressive consumer scale-out.

## Order Completion Time

Final order status:

    COMPLETED: 4201

Completion-time query:

    orders: 4201
    min_complete_seconds: 1.417204
    avg_complete_seconds: 53.7632751485360628
    max_complete_seconds: 115.044520

## Comparison With Previous 70 RPS Run

Previous Phase 4 70 RPS:

    accepted_orders: 4201
    avg_complete_seconds: about 76.25s
    max_complete_seconds: about 150.68s
    had strong scheduling/resource pressure

Current Phase 5 post-placement 70 RPS:

    accepted_orders: 4201
    avg_complete_seconds: about 53.76s
    max_complete_seconds: about 115.04s
    Kafka drained to zero
    all orders completed
    order-service returned to baseline
    consumer FailedScheduling still appeared

Interpretation:

    Node placement fix improved 70 RPS behavior significantly.
    Completion time improved.
    The system recovered cleanly after cooldown.
    However, consumer scale-out still exceeds schedulable CPU capacity on eligible nodes.

## Verdict

PASS WITH WARNING.

Pass:

- 4201 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 144.11ms, below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Kafka lag drained to zero after cooldown.
- All 4201 orders reached COMPLETED.
- HPA/KEDA eventually returned workloads to baseline.

Warning:

- Core Saga consumer lag appeared immediately after benchmark.
- payment-service-group was rebalancing.
- KEDA scaled consumers aggressively.
- consumer pods still hit FailedScheduling due to insufficient CPU on eligible nodes.
- 70 RPS should not yet be considered a clean safe capacity rating.

## Capacity Interpretation

After the node placement fix:

    60 RPS improved to PASS WITH OBSERVATION.
    70 RPS improved from stress-finding to PASS WITH WARNING / burst capacity candidate.

However, safe rated capacity should not be raised to 70 RPS yet.

Current interpretation:

    40 RPS: conservative safe capacity
    50 RPS: safe/burst candidate
    60 RPS: stable candidate after placement fix
    70 RPS: burst capacity with scheduling warning

## Next Action

Do not run 80 RPS yet.

Next tuning target:

1. Review consumer CPU/memory requests.
2. Review KEDA lagThreshold and maxReplica behavior.
3. Consider lowering maxReplica or increasing lagThreshold to avoid excessive scale-out.
4. Consider whether lightweight consumers can use smaller CPU requests.
5. Rerun 70 RPS after consumer/KEDA tuning.
