# Phase 5 Capacity and Observability Checkpoint

## Scope

This checkpoint closes Phase 5 of `my-ecommerce-platform`.

Phase 5 focused on:

- answering the fixed-resource capacity question,
- tuning high-load bottlenecks,
- stabilizing KEDA/HPA behavior,
- fixing GitOps replica drift,
- fixing Tempo observability instability after high-load benchmarks.

This checkpoint should be used as the starting context for Phase 6.

## Final Runtime State

Final verified state:

    Git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy
    abnormal pods: none

Verified ArgoCD applications:

- analytics-layer: Synced / Healthy
- cdc-layer: Synced / Healthy
- ecommerce-infrastructure: Synced / Healthy
- ecommerce-platform: Synced / Healthy
- infrastructure-layer: Synced / Healthy
- monitoring-addons: Synced / Healthy
- observability-layer: Synced / Healthy
- security-layer: Synced / Healthy

## Main Phase 5 Result

The system has demonstrated successful high-load processing under the current fixed lab resources.

Final capacity interpretation:

    70 RPS:
      confirmed stable through two repeated runs.

    80 RPS:
      confirmed through two repeated high-load runs after tuning.

Recommended report wording:

    Under the current fixed lab resources, the system has confirmed stable behavior at 70 RPS and has also demonstrated successful 80 RPS high-load operation after Phase 5 tuning. For conservative capacity planning, 70 RPS can be presented as the safe rating, while 80 RPS can be presented as the confirmed high-load rating in the lab environment.

Do not generalize this result to unlimited production capacity. The result only applies to the current lab cluster, current workload profile, current database state, and current scaling configuration.

## Capacity Evidence Summary

### 70 RPS Run 1

    accepted_orders: 4200
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: about 228.37ms
    orders COMPLETED: 4200
    avg_complete_seconds: about 16.72s
    max_complete_seconds: about 32.78s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

### 70 RPS Repeat Run

    accepted_orders: 4201
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: about 136.93ms
    orders COMPLETED: 4201
    avg_complete_seconds: about 15.95s
    max_complete_seconds: about 30.73s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

### 80 RPS Run 1

    accepted_orders: 4801
    checks_succeeded: 4801 / 4801
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: about 143.05ms
    orders COMPLETED: 4801
    avg_complete_seconds: about 23.13s
    max_complete_seconds: about 48.06s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

### 80 RPS Repeat Run

    accepted_orders: 4801
    checks_succeeded: 4801 / 4801
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: about 348.13ms
    orders COMPLETED: 4801
    avg_complete_seconds: about 28.18s
    max_complete_seconds: about 48.18s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

## Important Fixes Completed in Phase 5

### 1. Node Placement Relaxation

Problem:

    Some stateless services were too constrained by node placement rules.
    Under high load, this contributed to scheduling pressure.

Fix:

    Relaxed stateless service placement so workloads can be scheduled on eligible worker nodes.

Effect:

    Reduced FailedScheduling / Insufficient CPU pressure during high-load tests.

### 2. Consumer Resource Request Tuning

Problem:

    Consumer pods requested more CPU/memory than their observed runtime usage.
    KEDA scale-out consumed schedulable resources quickly.

Previous consumer requests:

    cpu: 200m
    memory: 256Mi

Updated consumer requests:

    cpu: 50m
    memory: 128Mi

Limits were kept unchanged:

    cpu: 500m
    memory: 512Mi

Effect:

    Reduced scheduling pressure while preserving runtime burst headroom.

### 3. KEDA Consumer Tuning

General KEDA tuning:

    cooldownPeriod: 180
    maxReplicaCount: 8
    lagThreshold: "80"
    activationLagThreshold: "20"

Inventory-specific tuning:

    minReplicaCount: 2
    maxReplicaCount: 8
    cooldownPeriod: 180
    lagThreshold: "40"
    activationLagThreshold: "10"

Reason:

    inventory-consumer is the first Saga consumer stage after order.created.
    Earlier 70 RPS runs showed inventory lag as the first-stage bottleneck.

Effect:

    End-to-end completion time improved sharply.
    70 RPS average completion improved to about 16s.
    80 RPS average completion stayed under about 30s.

### 4. ArgoCD Replica Drift Fix

Problem:

    After inventory-consumer minReplicaCount was raised to 2, ArgoCD and KEDA/HPA fought over Deployment.spec.replicas.

Fix:

    Added ignoreDifferences for /spec/replicas on autoscaled Deployments.
    Kept RespectIgnoreDifferences=true.

Affected workloads:

- order-service
- inventory-api
- inventory-consumer
- payment-api
- payment-consumer
- notification-api
- notification-consumer

Effect:

    ecommerce-platform stopped flipping between Synced and OutOfSync.
    KEDA/HPA can control live replicas without GitOps drift noise.

### 5. Tempo Observability Fix

Problem:

    After high-load tests, Tempo entered OOMKilled / CrashLoopBackOff.
    observability-layer became Progressing / OutOfSync during the failed resource patch attempts.

Root cause:

    Tempo resource limit was too low for high-load trace volume and local WAL/compaction.
    During the fix, two invalid manifests were accidentally committed:
      - request memory higher than limit,
      - invalid quantity value 2Gi1Gi.

Final correct Tempo resources:

    requests.cpu: 200m
    requests.memory: 1Gi
    limits.cpu: 1 core
    limits.memory: 2Gi

Final verification:

    Tempo pod: 1/1 Running
    restart count: 0
    /ready: HTTP 200 OK, body = ready
    observability-layer: Synced / Healthy

## Final Post-Fix Smoke

After Tempo fix:

    Saga smoke: PASS
    HTTP_CODE: 201
    order status: PENDING -> COMPLETED in about 7s
    outbox: PUBLISHED
    inventory reservation: RESERVED
    payment: COMPLETED
    notification: SENT
    Kafka lag before smoke: NO_NONZERO_NUMERIC_LAG
    Kafka lag after smoke: NO_NONZERO_NUMERIC_LAG

## Current Architecture Notes

The business path remains:

    client / test traffic
      -> web-gateway / Istio ingress
      -> order-service
      -> PostgreSQL order_db + outbox
      -> Kafka order.created
      -> inventory-consumer
      -> inventory_db
      -> Kafka inventory.reserved
      -> payment-consumer
      -> payment_db
      -> Kafka payment.completed
      -> order-service saga monitor
      -> order COMPLETED
      -> notification-consumer
      -> notification_db

CDC / Debezium / ClickHouse are not the reason PENDING becomes COMPLETED faster.

Their role is mainly:

- read-side analytics,
- reporting,
- CDC event capture,
- ClickHouse analytical queries,
- read-model / observability support.

The speed of PENDING -> COMPLETED mainly depends on:

- Kafka partitions and lag,
- consumer throughput,
- KEDA scaling behavior,
- order-service saga monitor,
- PostgreSQL write performance,
- PgBouncer / DB connection behavior,
- pod scheduling availability.

## Current Capacity Statement for Report

Suggested wording:

    With the current fixed OpenStack/Kubernetes lab resources, the system was tested using staircase load benchmarks. After tuning placement, consumer resources, KEDA scaling, GitOps replica drift, and Tempo observability resources, the system successfully completed two 70 RPS runs and two 80 RPS runs. In all final runs, HTTP error rate was 0%, all accepted orders reached COMPLETED, and Kafka lag drained to zero after cooldown.

Conservative statement:

    70 RPS is the safe fixed-resource rating.

Stronger demonstrated statement:

    80 RPS is the confirmed high-load rating in the current lab environment after Phase 5 tuning.

Caution:

    Testing beyond 80 RPS should be treated as stress-to-break testing, not normal capacity validation.

## Remaining Limitations

The current rating is not a theoretical maximum.

True upper bounds may still be limited by:

- PostgreSQL write throughput
- PgBouncer pool settings
- Kafka broker throughput
- Kafka partition count
- consumer processing throughput
- HPA/KEDA reaction delay
- node CPU and memory capacity
- disk pressure and image garbage collection
- observability overhead
- network and storage performance

## Documentation State

Important Phase 5 documents:

- docs/benchmark/phase5-capacity-rating-summary.md
- docs/benchmark/phase5-post-inventory-keda-70rps.md
- docs/benchmark/phase5-repeat-inventory-keda-70rps.md
- docs/evidence/phase5-inventory-keda-gitops-fix.md
- docs/evidence/phase5-keda-consumer-tuning.md
- docs/evidence/phase5-consumer-resource-request-tuning.md

This checkpoint is the closing document for Phase 5.

## Phase 6 Roadmap

Phase 6 should focus on production-readiness and operations, not more normal capacity benchmarking.

Recommended order:

### 1. Alerting

Add alert rules for:

- Kafka lag not draining
- consumer group rebalancing / high lag
- HPA maxed out
- pending pods / FailedScheduling
- PostgreSQL pressure
- PgBouncer connection pressure
- node disk pressure
- Tempo OOM / CrashLoopBackOff
- ArgoCD OutOfSync / Degraded apps

### 2. Centralized Logging

Add Loki + Promtail or Grafana Alloy.

Goal:

    Collect logs from services, consumers, Kafka-related workloads, PostgreSQL/PgBouncer, and observability components into Grafana.

### 3. Runbooks

Write incident runbooks for:

- high Kafka lag
- pod Pending / FailedScheduling
- PostgreSQL pressure
- PgBouncer connection exhaustion
- Tempo OOM
- ArgoCD OutOfSync
- node disk pressure
- failed or delayed Saga completion

### 4. Backup and Restore Validation

Check backup scripts and validate restore flow.

Focus:

- PostgreSQL backup
- restore check
- evidence document for backup/restore
- whether analytics/read-model data needs separate backup or can be rebuilt

### 5. Chaos Testing

Run chaos tests separately from capacity benchmarks.

Existing chaos experiment candidates:

- pod-kill-payment
- network-delay-kafka
- cpu-stress-inventory

Chaos should be run with low or moderate background traffic, not during normal capacity rating.

### 6. Scenario K6 Tests

K6 scenario tests should be used after alerting/logging/runbooks are ready.

Candidate tests:

- idempotency
- flash-sale
- flash-sale-spike
- spike
- stress
- soak

These are not needed to close Phase 5.
