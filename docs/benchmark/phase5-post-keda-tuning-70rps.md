# Phase 5 Post-KEDA-Tuning Benchmark - 70 RPS

## Scope

This document records the 70 RPS benchmark after KEDA consumer tuning.

## Related Change

KEDA tuning commit:

    f6a1af5 capacity: tune consumer keda scaling

Evidence commit:

    63d5069 docs: record phase5 keda consumer tuning

KEDA values after tuning:

    cooldownPeriod: 180
    maxReplicaCount: 8
    lagThreshold: "80"
    activationLagThreshold: "20"

## Environment

Runtime URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase5-post-keda-tuning-70rps-20260607134150

Test script:

    tests/k6/baseline-e2e-70rps.js

Test profile:

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

Live KEDA config showed:

    inventory-consumer-scaler:
      maxReplicaCount: 8
      cooldownPeriod: 180
      lagThreshold: "80"
      activationLagThreshold: "20"

    payment-consumer-scaler:
      maxReplicaCount: 8
      cooldownPeriod: 180
      lagThreshold: "80"
      activationLagThreshold: "20"

    notification-consumer-scaler:
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
    http_req_duration avg: 91.71ms
    http_req_duration min: 24.57ms
    http_req_duration med: 60.33ms
    http_req_duration max: 1.83s
    http_req_duration p90: 191.64ms
    http_req_duration p95: 280.01ms

Execution:

    iterations: 4200
    effective rate: 69.749855/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag Immediately After Benchmark

Observed:

    inventory-service-group was rebalancing.

Visible lag:

    inventory-service-group / order.created:
      lag range: about 255 to 460

    clickhouse-orders-flat-cdc-v2:
      lag range: about 17 to 36

Payment and notification lag were not the main visible bottleneck in this run.

## HPA / KEDA Behavior

Immediately after benchmark:

    order-service:
      CPU: 98% / 25%
      replicas: 8

    inventory-consumer:
      metric: 640 / 80
      replicas initially still 1
      later scaled to 4 and 8

    payment-consumer:
      metric: 80 / 80
      later scaled to 4

    notification-consumer:
      metric: 80 / 80
      later scaled to 2

No FailedScheduling / Insufficient CPU error was observed.

## Cooldown Result

After cooldown:

    Kafka lag: NO_NONZERO_NUMERIC_LAG

Deployments after cooldown:

    inventory-consumer: 1
    payment-consumer: 1
    notification-consumer: 1
    order-service: 3

No abnormal non-Running pod remained.

## Order Completion Time

Final order status:

    COMPLETED: 4200

Completion-time query:

    orders: 4200
    min_complete_seconds: 1.320108
    avg_complete_seconds: 68.4575224742857143
    max_complete_seconds: 152.352565

## Comparison

Phase 5 post-placement 70 RPS:

    avg_complete_seconds: about 53.76s
    max_complete_seconds: about 115.04s

Phase 5 post-consumer-request 70 RPS:

    avg_complete_seconds: about 72.25s
    max_complete_seconds: about 132.00s

Current post-KEDA-tuning 70 RPS:

    avg_complete_seconds: about 68.46s
    max_complete_seconds: about 152.35s

Interpretation:

    KEDA tuning reduced the previous scheduling risk and kept the system recoverable.
    However, it did not improve end-to-end async completion time.
    Inventory consumer became the visible first-stage bottleneck after tuning.

## Verdict

PASS WITH WARNING / MIXED RESULT.

Pass:

- 4200 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 280.01ms, below the 1500ms threshold.
- Kafka lag drained to zero after cooldown.
- All 4200 orders reached COMPLETED.
- No clear FailedScheduling / Insufficient CPU error was observed.

Warning:

- Inventory consumer group was rebalancing immediately after benchmark.
- Inventory lag on `order.created` reached about 255-460.
- Average completion time was worse than the best post-placement 70 RPS run.
- Max completion time reached about 152.35s.

## Capacity Interpretation

70 RPS remains a burst/stress candidate, not a clean safe capacity rating.

Current interpretation:

    60 RPS: stable candidate
    70 RPS: burst candidate with inventory/KEDA warning
    80 RPS: should not be tested yet

## Next Action

Do not run 80 RPS.

Tune inventory-consumer separately because it is now the visible first-stage bottleneck:

    minReplicaCount: 1 -> 2
    lagThreshold: "80" -> "40"
    activationLagThreshold: "20" -> "10"

Keep payment and notification KEDA settings unchanged for now.
