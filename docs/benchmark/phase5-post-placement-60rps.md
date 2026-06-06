# Phase 5 Post-Placement Benchmark - 60 RPS

## Scope

This document records the 60 RPS benchmark after the Phase 5 node placement fix.

The goal is to verify whether relaxing stateless service placement from only `vm2-mesh` to `vm1-gateway` + `vm2-mesh` reduces the Pending pod / FailedScheduling issue observed in the earlier Phase 4 60 RPS run.

## Related Fix

Node placement fix commit:

    0586c4c capacity: relax stateless service node placement

Evidence commit:

    8f2cfb0 docs: record phase5 node placement fix

The fix replaced hard `nodeSelector: vm2-mesh` with required node affinity allowing:

    vm1-gateway
    vm2-mesh

for stateless/scalable services.

## Environment

Runtime URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase5-post-placement-60rps-20260607004050

Test script:

    tests/k6/baseline-e2e-60rps.js

Test profile:

    executor: constant-arrival-rate
    rate: 60 requests/second
    duration: 60 seconds
    preAllocatedVUs: 120
    maxVUs: 360

Thresholds:

    http_req_failed < 1%
    unexpected_error_rate < 1%
    http_req_duration p95 < 1500ms

## Precheck

Before benchmark:

    git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy
    no abnormal non-Running pod

Runtime placement before benchmark confirmed that stateless services were distributed across vm1-gateway and vm2-mesh.

Examples:

    order-service:
      vm1-gateway
      vm2-mesh

    inventory-api:
      vm1-gateway

    inventory-consumer:
      vm1-gateway

    notification-api:
      vm2-mesh

    notification-consumer:
      vm1-gateway

    payment-api:
      vm2-mesh

    payment-consumer:
      vm1-gateway

    read-model-service:
      vm1-gateway

## k6 Result

Total result:

    checks_total: 3600
    checks_succeeded: 3600 / 3600
    checks_failed: 0 / 3600

Custom metrics:

    accepted_orders: 3600
    unexpected_error_rate: 0.00%

HTTP metrics:

    http_reqs: 3600
    http_req_failed: 0.00%
    http_req_duration avg: 40.05ms
    http_req_duration min: 16.09ms
    http_req_duration med: 27.99ms
    http_req_duration max: 683.08ms
    http_req_duration p90: 58.29ms
    http_req_duration p95: 125.45ms

Execution:

    iterations: 3600
    effective rate: 59.893122/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after benchmark, the visible non-zero lag was limited to ClickHouse CDC:

    clickhouse-orders-flat-cdc-v2:
      lag range: 3 to 12

No visible non-zero lag was printed for the core Saga groups in the filtered output.

After cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS

## HPA / Deployment Behavior

Immediately after benchmark:

    order-service HPA:
      CPU: 36% / 25%
      replicas: 6

    order-service deployment:
      READY: 6
      REPLICAS: 6
      AVAILABLE: 6

No Pending pod was observed.

After cooldown:

    order-service deployment:
      READY: 5
      REPLICAS: 5
      AVAILABLE: 5

No Pending pod was observed after cooldown.

Observation:

    order-service had not yet returned to the baseline 2 replicas after the 300s cooldown window.
    This is not a benchmark failure, but should be monitored as HPA/KEDA scale-down behavior.

## Resource Snapshot

Immediately after benchmark:

    vm1-gateway:
      CPU: 1995m
      memory: 5662Mi

    vm2-mesh:
      CPU: 634m
      memory: 4273Mi

    vm3-gitops:
      CPU: 1242m
      memory: 5472Mi

The node placement fix allowed order-service scale-out pods to run across vm1-gateway and vm2-mesh.

## Order Completion Time

Final order status:

    COMPLETED: 3600

Completion-time query:

    orders: 3600
    min_complete_seconds: 1.501084
    avg_complete_seconds: 46.2734865977777778
    max_complete_seconds: 89.584396

## Comparison With Previous 60 RPS Run

Previous Phase 4 60 RPS result:

    orders: 3600
    avg_complete_seconds: 45.2118565030555556
    max_complete_seconds: 126.495995
    had Pending pods / FailedScheduling

Current Phase 5 post-placement 60 RPS result:

    orders: 3600
    avg_complete_seconds: 46.2734865977777778
    max_complete_seconds: 89.584396
    no Pending pods
    no FailedScheduling events in final output

Interpretation:

    The node placement fix did not significantly reduce average completion time.
    However, it successfully removed the main operational bottleneck: scale-out pod scheduling failure.
    Max completion time improved from about 126.50s to about 89.58s.

## Verdict

PASS WITH OBSERVATION.

Pass:

- 3600 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 125.45ms, below the 1500ms threshold.
- Kafka lag drained to zero after cooldown.
- All 3600 orders reached COMPLETED.
- No Pending pod remained.
- No FailedScheduling event was observed in the final benchmark output.
- Node placement fix successfully allowed scale-out pods to schedule.

Observation:

- order-service remained above baseline replica count after 300s cooldown.
- Average completion time remained similar to the previous 60 RPS run.
- Further tuning may require reviewing HPA scale-down behavior, KEDA targets, and backend async completion time.

## Capacity Interpretation

After the node placement fix, 60 RPS improved from:

    OPERATIONAL CAPACITY WARNING

to:

    PASS WITH OBSERVATION

This suggests the current fixed-resource cluster can now handle 60 RPS better than before.

However, 60 RPS should not yet replace 40 RPS as the safe rated capacity until a follow-up stabilization check and possibly a repeated 60 RPS run confirm the result.

## Next Action

1. Run an extra stabilization check to confirm order-service returns to baseline.
2. If stable, rerun 70 RPS after placement fix.
3. Compare 70 RPS post-placement with the old 70 RPS stress result.
