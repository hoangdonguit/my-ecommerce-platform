# Phase 5 Post-Inventory-KEDA Benchmark - 70 RPS

## Scope

This document records the 70 RPS benchmark after inventory-specific KEDA tuning and GitOps replica drift fix.

## Related Changes

Inventory KEDA tuning:

    cb6d1cf capacity: tune inventory consumer keda scaling

GitOps replica drift fix:

    65d327c fix: repair ecommerce argocd ignore differences

Evidence:

    2406b2d docs: record phase5 inventory keda gitops fix

## Environment

Runtime URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase5-post-inventory-keda-70rps-20260607210657

Test script:

    tests/k6/baseline-e2e-70rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 70 requests/second
    duration: 60 seconds
    preAllocatedVUs: 140
    maxVUs: 420

## Precheck

Before benchmark:

    git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy
    no abnormal non-Running pod

Baseline consumer replicas:

    inventory-consumer: 2 / 2
    payment-consumer: 1 / 1
    notification-consumer: 1 / 1

Inventory KEDA:

    minReplicaCount: 2
    maxReplicaCount: 8
    cooldownPeriod: 180
    lagThreshold: "40"
    activationLagThreshold: "10"

Payment and notification KEDA remained:

    minReplicaCount: 1
    maxReplicaCount: 8
    cooldownPeriod: 180
    lagThreshold: "80"
    activationLagThreshold: "20"

## k6 Result

Total result:

    checks_total: 4200
    checks_succeeded: 4200 / 4200
    checks_failed: 0 / 4200

Custom metrics:

    accepted_orders: 4200
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 4200
    http_req_failed: 0.00%
    http_req_duration avg: 79.98ms
    http_req_duration min: 19.9ms
    http_req_duration med: 52.03ms
    http_req_duration max: 985.69ms
    http_req_duration p90: 156.87ms
    http_req_duration p95: 228.37ms

Execution:

    iterations: 4200
    effective rate: 69.672402/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after benchmark, transient lag existed in:

- ClickHouse CDC group
- inventory-service-group / order.created
- payment-service-group / inventory.reserved
- notification-service-group / payment.completed
- small order-service-saga-monitor lag

After cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH TRANSIENT LAG OBSERVATION

## HPA / KEDA Behavior

Immediately after benchmark:

    order-service: scaled to 7, later 8
    inventory-consumer: scaled to 8
    payment-consumer: scaled to 4, later 8
    notification-consumer: scaled to 8

After cooldown:

    order-service: 2
    inventory-consumer: 2
    payment-consumer: 1
    notification-consumer: 1

No FailedScheduling / Insufficient CPU / OOM / BackOff was observed.

Scale-down `Killing` events were normal HPA/KEDA scale-down events.

## Resource Snapshot

Immediately after benchmark:

    vm1-gateway CPU: 2769m, memory: 5587Mi
    vm2-mesh CPU: 1481m, memory: 4552Mi
    vm3-gitops CPU: 2284m, memory: 5801Mi

Selected consumer usage was low per pod, generally tens of millicores CPU and single/tens of Mi memory.

## Order Completion Time

Final order status:

    COMPLETED: 4200

Completion-time query:

    orders: 4200
    min_complete_seconds: 1.535773
    avg_complete_seconds: 16.7152847702380952
    max_complete_seconds: 32.779477

## Comparison

Previous Phase 4 70 RPS:

    avg_complete_seconds: about 76.25s
    max_complete_seconds: about 150.68s
    had scheduling stress

Phase 5 post-placement 70 RPS:

    avg_complete_seconds: about 53.76s
    max_complete_seconds: about 115.04s

Phase 5 post-consumer-request 70 RPS:

    avg_complete_seconds: about 72.25s
    max_complete_seconds: about 132.00s

Phase 5 post-KEDA-tuning 70 RPS:

    avg_complete_seconds: about 68.46s
    max_complete_seconds: about 152.35s

Current Phase 5 post-inventory-KEDA 70 RPS:

    avg_complete_seconds: about 16.72s
    max_complete_seconds: about 32.78s

Interpretation:

    Inventory-specific KEDA tuning significantly improved the first Saga stage.
    End-to-end async completion time improved sharply.
    The system recovered to baseline after cooldown.
    No scheduling failure was observed.

## Verdict

PASS WITH OBSERVATION.

Pass:

- 4200 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 228.37ms, below the 1500ms threshold.
- Kafka lag drained to zero after cooldown.
- All 4200 orders reached COMPLETED.
- Average completion time improved strongly to about 16.72s.
- No FailedScheduling / Insufficient CPU / OOM / BackOff was observed.
- ArgoCD remained Synced / Healthy.

Observation:

- Transient Kafka lag still appears immediately after the benchmark.
- KEDA/HPA scales several workloads to high replica counts during the run.
- 70 RPS should be repeated once more before being declared the official fixed-resource safe rating.

## Capacity Interpretation

After inventory-specific KEDA tuning:

    60 RPS: stable/pass
    70 RPS: strong stable candidate / pass with observation
    80 RPS: not needed yet

Current recommended capacity statement:

    Under the current fixed lab resources, the system has demonstrated a successful 70 RPS run with 0% HTTP error rate, all orders completed, Kafka lag drained after cooldown, and no scheduling failure. 70 RPS is therefore the current strongest candidate for fixed-resource rated capacity, pending one repeat run for confirmation.

## Next Action

1. Run one repeat 70 RPS confirmation test, or
2. Stop load testing here and write a Phase 5 capacity checkpoint.

Do not run 80 RPS unless the goal is explicitly stress-to-break.
