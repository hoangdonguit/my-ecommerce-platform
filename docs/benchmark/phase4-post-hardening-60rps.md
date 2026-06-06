# Phase 4 Post-Hardening Benchmark - 60 RPS

## Scope

This document records the Phase 4 post-hardening 60 RPS benchmark for `my-ecommerce-platform`.

The purpose is to evaluate whether the system remains stable beyond the previous 50 RPS burst-capacity point.

This run is important for capacity planning because it shows that the API path still passes, while the Kubernetes scheduling/autoscaling layer starts showing fixed-resource pressure.

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local benchmark folder:

    .local-notes/benchmark/phase4-post-hardening-60rps-20260606175014

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
    http_req_duration avg: 41.60ms
    http_req_duration min: 17.06ms
    http_req_duration med: 31.26ms
    http_req_duration max: 307.57ms
    http_req_duration p90: 75.60ms
    http_req_duration p95: 96.39ms

Execution:

    iterations: 3600
    effective rate: 59.894797/s
    interrupted iterations: 0

k6 verdict:

    PASS

## Kafka Lag

Immediately after the benchmark, Kafka had transient lag across multiple groups.

Observed immediate lag included:

    clickhouse-orders-flat-cdc-v2 / cdc_flat.order_db.public.orders:
      lag range: 17 to 30

    inventory-service-group / order.created:
      lag range: 50 to 69

    notification-service-group / payment.completed:
      lag range: 2 to 6

    order-service-saga-monitor / payment.completed:
      lag: 1

    payment-service-group / inventory.reserved:
      lag range: 1 to 2

    read-model-service-group / payment.completed:
      lag range: 2 to 4

After cooldown:

    NO_NONZERO_NUMERIC_LAG

Kafka verdict:

    PASS WITH OBSERVATION

Kafka backlog drained successfully after cooldown.

## PgBouncer State

PgBouncer after benchmark:

    order_db cl_active = 30
    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

PgBouncer verdict:

    PASS

No PgBouncer pool waiting backlog was observed.

## Scaling and Scheduling Behavior

Immediately after benchmark:

    order-service HPA:
      observed CPU: 72% / 25%
      desired replicas: 6

    order-service deployment:
      READY: 4
      REPLICAS: 6

    inventory-consumer HPA:
      metric: 160/20
      desired replicas initially remained 1, then later scaled

After cooldown:

    inventory-consumer HPA:
      metric: 2500m/20
      desired replicas: 8
      deployment: READY 1 / REPLICAS 8

    payment-consumer HPA:
      metric: 2500m/20
      desired replicas: 8
      deployment: READY 2 / REPLICAS 8

    order-service:
      deployment: READY 4 / REPLICAS 6

At this point, multiple pods were Pending.

Extra stabilization diagnosis later showed:

    Pods not Running/Completed: none
    Kafka lag: NO_NONZERO_NUMERIC_LAG
    order status: COMPLETED = 3600
    order-service returned to 2 replicas
    inventory-consumer returned to 1 replica
    payment-consumer returned to 1 replica

However, recent events showed many FailedScheduling warnings during the 60 RPS window:

    0/3 nodes are available:
      1 Insufficient cpu
      2 node(s) didn't match Pod's node affinity/selector

The cluster also reported image filesystem warnings on vm3-gitops:

    Insufficient free disk space on the node's image filesystem
    85% of 19.2 GiB used

Scaling/scheduling verdict:

    CAPACITY WARNING

The system eventually stabilized, but 60 RPS exposed fixed-resource pressure and scheduling constraints.

## Node and Resource Context

Cluster nodes:

    vm1-gateway: 4 CPU, about 8 Gi memory
    vm2-mesh: 4 CPU, about 8 Gi memory
    vm3-gitops: 4 CPU, about 8 Gi memory

Observed node usage during diagnosis:

    vm1-gateway: 273m CPU, 5469Mi memory
    vm2-mesh: 300m CPU, 4633Mi memory
    vm3-gitops: 582m CPU, 6208Mi memory

Default namespace service requests/limits include:

    order-service:
      requests.cpu=200m
      requests.memory=256Mi
      limits.cpu=500m
      limits.memory=512Mi

    inventory-consumer:
      requests.cpu=200m
      requests.memory=256Mi
      limits.cpu=500m
      limits.memory=512Mi

    payment-consumer:
      requests.cpu=200m
      requests.memory=256Mi
      limits.cpu=500m
      limits.memory=512Mi

The scheduling warning indicates that the autoscaler requested more pods than could be scheduled under current node placement and CPU request constraints.

## Order Completion Time

Completion-time query for the 60 RPS run:

    orders: 3600
    min_complete_seconds: 1.431411
    avg_complete_seconds: 45.2118565030555556
    max_complete_seconds: 126.495995

Final order status summary:

    COMPLETED: 3600

Interpretation:

    All accepted orders eventually completed.
    However, average and maximum completion time increased compared with 50 RPS.
    This confirms that async backend drain time is becoming more visible as RPS increases.

## Verdict

API PASS, EVENTUAL COMPLETION PASS, OPERATIONAL CAPACITY WARNING.

Pass:

- 3600 accepted orders.
- 0 failed k6 checks.
- 0% HTTP failure rate.
- 0% unexpected error rate.
- p95 latency was 96.39ms, far below the 1500ms threshold.
- PgBouncer had no pool waiting backlog.
- Kafka lag drained to zero.
- All 3600 benchmark orders reached COMPLETED.
- The system eventually returned to a stable baseline.

Warning:

- Multiple pods became Pending after the run.
- FailedScheduling occurred due to Insufficient CPU and node affinity/selector constraints.
- inventory-consumer and payment-consumer attempted to scale to 8 replicas.
- order-service attempted to scale to 6 replicas.
- Completion time increased to avg 45.21s and max 126.49s.
- vm3-gitops image filesystem reached 85% usage and triggered ImageGC/FreeDiskSpace warnings.

## Capacity Interpretation

With the current fixed-resource cluster:

    40 RPS should be treated as the cleaner safe rated capacity.
    50 RPS should be treated as burst capacity with warnings.
    60 RPS should be treated as overload/capacity-warning zone.

60 RPS is not a clean stable capacity rating, even though the API path passed.

The reason is that capacity rating must include async drain, pod scheduling, autoscaling stability, Kafka lag recovery, and order completion time, not only HTTP success.

## Next Action

Do not run 70 RPS as a normal capacity benchmark.

If 70 RPS is tested, it should be explicitly classified as stress-to-break / bottleneck-finding, not as a stable baseline.

Before further stress testing, recommended checks/fixes are:

1. Review node affinity/selector placement rules.
2. Check why extra consumer/order pods can only schedule on a limited node pool.
3. Clean vm3-gitops image filesystem or increase disk.
4. Consider whether consumer HPA/KEDA scale targets are too aggressive.
5. Record final capacity rating methodology for the report.
