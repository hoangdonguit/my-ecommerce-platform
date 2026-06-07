# Phase 5 Repeat Inventory-KEDA Benchmark - 70 RPS

## Scope

This document records the repeat 70 RPS benchmark after inventory-specific KEDA tuning.

The goal is to confirm whether the previous strong 70 RPS result was stable and repeatable.

## Environment

Runtime URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase5-repeat-inventory-keda-70rps-20260607212420

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
    Kafka lag before test: NO_NONZERO_NUMERIC_LAG

Baseline replicas:

    inventory-consumer: 2
    payment-consumer: 1
    notification-consumer: 1
    order-service: 2

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
    http_req_duration avg: 55.63ms
    http_req_duration min: 17.59ms
    http_req_duration med: 40.01ms
    http_req_duration max: 581.28ms
    http_req_duration p90: 101.31ms
    http_req_duration p95: 136.93ms

Execution:

    iterations: 4201
    effective rate: 69.840024/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after benchmark:

    transient lag appeared
    inventory-service-group showed a short rebalancing warning
    ClickHouse CDC had small lag
    Saga consumer groups had temporary lag

After cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH TRANSIENT LAG OBSERVATION

## HPA / KEDA Behavior

During the run, autoscaling occurred:

    inventory-consumer scaled up
    payment-consumer scaled up
    notification-consumer scaled up
    order-service scaled up

No FailedScheduling, Insufficient CPU, OOM, BackOff, or ImagePull failure was observed.

Scale-down Killing events after the test were normal HPA/KEDA scale-down behavior.

## Order Completion Time

Final order status:

    COMPLETED: 4201

Completion-time query:

    orders: 4201
    min_complete_seconds: 1.380870
    avg_complete_seconds: 15.9501629878600333
    max_complete_seconds: 30.734379

## Comparison With Previous 70 RPS Run

Previous post-inventory-KEDA 70 RPS:

    accepted_orders: 4200
    avg_complete_seconds: about 16.72s
    max_complete_seconds: about 32.78s

Repeat post-inventory-KEDA 70 RPS:

    accepted_orders: 4201
    avg_complete_seconds: about 15.95s
    max_complete_seconds: about 30.73s

Interpretation:

    70 RPS is repeatable after inventory-specific KEDA tuning.
    The second run confirmed very similar and slightly better async completion time.
    No scheduling failure was observed.

## Verdict

PASS CONFIRMED.

Pass:

- 4201 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 136.93ms, far below the 1500ms threshold.
- Kafka lag drained to zero after cooldown.
- All 4201 orders reached COMPLETED.
- Average completion time was about 15.95s.
- No FailedScheduling / Insufficient CPU / OOM / BackOff was observed.
- ArgoCD remained Synced / Healthy.

Observation:

- Transient Kafka lag and autoscaling still appear immediately after the benchmark.
- This is acceptable for the current asynchronous Saga design because the lag drains after cooldown.

## Capacity Interpretation

After two successful 70 RPS runs, 70 RPS is the strongest current fixed-resource rated capacity candidate.

80 RPS can be tested next as an upper-bound / stress-to-break test, not yet as a safe rated capacity.
