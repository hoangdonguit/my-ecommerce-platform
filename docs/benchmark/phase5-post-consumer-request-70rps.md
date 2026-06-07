# Phase 5 Post-Consumer-Request Benchmark - 70 RPS

## Scope

This document records the 70 RPS benchmark after reducing consumer resource requests.

The goal is to verify whether reducing consumer requests from `200m/256Mi` to `50m/128Mi` reduces the previous consumer FailedScheduling issue.

## Related Changes

Consumer request tuning commit:

    ef3b552 capacity: reduce consumer resource requests

Evidence commit:

    30281c9 docs: record phase5 consumer request tuning

Consumers changed:

- inventory-consumer
- payment-consumer
- notification-consumer

Requests changed from:

    cpu: 200m
    memory: 256Mi

to:

    cpu: 50m
    memory: 128Mi

Limits were not reduced:

    cpu: 500m
    memory: 512Mi

## Environment

Runtime URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase5-post-consumer-request-70rps-20260607130022

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

Live consumer resources:

    inventory-consumer:
      requests: 50m / 128Mi
      limits: 500m / 512Mi

    payment-consumer:
      requests: 50m / 128Mi
      limits: 500m / 512Mi

    notification-consumer:
      requests: 50m / 128Mi
      limits: 500m / 512Mi

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
    http_req_duration avg: 66.25ms
    http_req_duration min: 21.14ms
    http_req_duration med: 50.72ms
    http_req_duration max: 1.81s
    http_req_duration p90: 98.82ms
    http_req_duration p95: 122.72ms

Execution:

    iterations: 4201
    effective rate: 69.813124/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after benchmark, Kafka had visible transient lag and several consumer groups were rebalancing:

    inventory-service-group: rebalancing
    order-service-saga-monitor: rebalancing
    payment-service-group: rebalancing

Visible immediate lag:

    inventory-service-group / order.created:
      lag range: about 352 to 416

    order-service-saga-monitor / payment.completed:
      visible lag up to 59

    payment-service-group / inventory.reserved:
      visible lag up to 174

After cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH WARNING

Kafka eventually drained, but immediate lag and rebalancing were worse than desired.

## HPA / KEDA Behavior

Immediately after benchmark:

    order-service:
      CPU: 59% / 25%
      replicas: 8

    inventory-consumer:
      metric: 160 / 20
      KEDA active

    payment-consumer:
      metric: 320 / 20
      KEDA active

After cooldown:

    Kafka lag: NO_NONZERO_NUMERIC_LAG
    no abnormal non-Running pod remained

However, many normal `Killing` events appeared because KEDA/HPA scaled consumers up and then down repeatedly.

## Scheduling Result

No clear FailedScheduling / Insufficient CPU error was observed in the final filtered event output.

This suggests the consumer request reduction successfully reduced the previous scheduling pressure.

Scheduling verdict:

    PASS

## Order Completion Time

Final order status:

    COMPLETED: 4201

Completion-time query:

    orders: 4201
    min_complete_seconds: 1.840813
    avg_complete_seconds: 72.2477536374672697
    max_complete_seconds: 132.000088

## Comparison

Previous Phase 4 70 RPS:

    avg_complete_seconds: about 76.25s
    max_complete_seconds: about 150.68s
    had strong FailedScheduling / Pending pressure

Phase 5 post-placement 70 RPS:

    avg_complete_seconds: about 53.76s
    max_complete_seconds: about 115.04s
    scheduling improved, but warning still existed

Current Phase 5 post-consumer-request 70 RPS:

    avg_complete_seconds: about 72.25s
    max_complete_seconds: about 132.00s
    no clear FailedScheduling in final events
    but consumer rebalancing/churn increased

Interpretation:

    Reducing consumer requests helped scheduling.
    However, it did not improve end-to-end async completion time.
    The new dominant issue is KEDA/Kafka consumer rebalancing and scale churn.

## Verdict

PASS WITH WARNING / REGRESSION OBSERVATION.

Pass:

- 4201 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 122.72ms, below the 1500ms threshold.
- Kafka lag drained to zero after cooldown.
- All 4201 orders reached COMPLETED.
- No clear FailedScheduling / Insufficient CPU error appeared in final filtered events.

Warning:

- Immediate Kafka lag was high.
- Several consumer groups were rebalancing.
- Average completion time worsened compared with the previous post-placement 70 RPS run.
- KEDA appears too aggressive for the current cluster and workload behavior.

## Capacity Interpretation

The consumer request reduction should be kept because it reduces scheduling pressure and did not break smoke/restart behavior.

However, 70 RPS is still not a clean safe rated capacity.

Current interpretation:

    60 RPS: stable candidate
    70 RPS: burst / stress candidate with KEDA rebalancing warning
    80 RPS: should not be tested yet

## Next Action

Do not run 80 RPS.

Next tuning target:

1. Keep consumer requests at 50m / 128Mi.
2. Keep consumer limits at 500m / 512Mi.
3. Tune KEDA to reduce aggressive scale-out and rebalancing.
4. Rerun 70 RPS after KEDA tuning.
