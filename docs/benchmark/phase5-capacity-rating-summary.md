# Phase 5 Capacity Rating Summary

## Scope

This document summarizes the Phase 5 capacity-rating result for `my-ecommerce-platform`.

The goal is to answer the teacher's capacity planning question:

1. What load level can the system handle under the current fixed lab resources?
2. What bottlenecks were found and fixed?
3. What load level is currently safe to claim under the current environment?

This file summarizes important milestones only. Small sanity checks are not documented separately.

## Current Fixed-Resource Environment

The current benchmark environment is the existing lab Kubernetes cluster on OpenStack VMs.

Runtime URL:

    http://100.65.255.2:30517

Current scaling model:

- order-service: HPA by CPU
- inventory-consumer: KEDA by Kafka lag
- payment-consumer: KEDA by Kafka lag
- notification-consumer: KEDA by Kafka lag

Current consumer resource profile:

    requests.cpu: 50m
    requests.memory: 128Mi
    limits.cpu: 500m
    limits.memory: 512Mi

Important GitOps fix:

- ArgoCD ignores `/spec/replicas` for autoscaled workloads.
- This prevents ArgoCD from fighting HPA/KEDA over live replica counts.

## Important Phase 5 Fixes

### 1. Node Placement Fix

Earlier, some stateless workloads were too restricted by placement rules, causing scheduling pressure under high load.

Placement was relaxed so stateless services could run on eligible worker nodes instead of being pinned too narrowly.

### 2. Consumer Resource Request Tuning

Consumer requests were reduced from:

    cpu: 200m
    memory: 256Mi

to:

    cpu: 50m
    memory: 128Mi

Limits were intentionally kept unchanged:

    cpu: 500m
    memory: 512Mi

This reduced scheduling pressure without reducing runtime headroom.

### 3. KEDA Consumer Tuning

General consumer KEDA tuning:

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

- Inventory is the first Saga consumer stage after `order.created`.
- Earlier 70 RPS results showed inventory lag as the first-stage bottleneck.
- Keeping inventory-consumer warm at 2 replicas significantly improved end-to-end completion time.

### 4. ArgoCD Replica Drift Fix

When inventory-consumer minReplicaCount was raised to 2, ArgoCD initially fought KEDA/HPA because Git had `replicas: 1`.

Fix:

- Add `ignoreDifferences` for `/spec/replicas` on autoscaled Deployments.
- Keep `RespectIgnoreDifferences=true`.

Affected workloads:

- order-service
- inventory-api
- inventory-consumer
- payment-api
- payment-consumer
- notification-api
- notification-consumer

### 5. Tempo Observability Fix

After high-load tests, Tempo showed OOMKilled / CrashLoopBackOff.

Cause:

- Tempo local WAL/traces and compaction were too heavy for the previous resource profile.

Previous Tempo resources:

    requests.cpu: 100m
    requests.memory: 512Mi
    limits.cpu: 500m
    limits.memory: 1Gi

Updated Tempo resources:

    requests.cpu: 200m
    requests.memory: 1Gi
    limits.cpu: 1 core
    limits.memory: 2Gi

Post-fix verification:

    Tempo pod: 1/1 Running
    restart count: 0
    /ready: HTTP 200 OK, body = ready
    observability-layer: Synced / Healthy

## 70 RPS Confirmed Result

70 RPS was run twice after inventory-specific KEDA tuning.

### 70 RPS Run 1

Result:

    accepted_orders: 4200
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: 228.37ms
    orders COMPLETED: 4200
    avg_complete_seconds: about 16.72s
    max_complete_seconds: about 32.78s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

### 70 RPS Repeat Run

Result:

    accepted_orders: 4201
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: 136.93ms
    orders COMPLETED: 4201
    avg_complete_seconds: about 15.95s
    max_complete_seconds: about 30.73s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

### 70 RPS Verdict

PASS CONFIRMED.

70 RPS is stable under the current fixed lab resources.

## 80 RPS Confirmed Result

80 RPS was also run twice after tuning.

### 80 RPS Run 1

Result:

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

Result:

    accepted_orders: 4801
    checks_succeeded: 4801 / 4801
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration p95: about 348.13ms
    orders COMPLETED: 4801
    avg_complete_seconds: about 28.18s
    max_complete_seconds: about 48.18s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

### 80 RPS Verdict

PASS CONFIRMED WITH OBSERVATION.

80 RPS passed twice for the business path:

- 0% HTTP error rate.
- All accepted orders reached COMPLETED.
- Kafka lag drained after cooldown.
- No persistent FailedScheduling / Insufficient CPU / OOM / BackOff in the business path.
- ArgoCD returned to Synced / Healthy after the Tempo fix.

Observation:

- Tempo required resource tuning after high-load tracing.
- Transient Kafka lag and autoscaling are expected immediately after high-load runs.

## Operational Observations

### Kafka Lag

Transient Kafka lag is expected immediately after high-load runs because the system uses asynchronous Saga processing.

The important pass condition is:

    Kafka lag must drain to NO_NONZERO_NUMERIC_LAG after cooldown.

This condition passed for final 70 RPS and 80 RPS runs.

### Autoscaling

During high load, HPA/KEDA scaled several workloads up:

- order-service
- inventory-consumer
- payment-consumer
- notification-consumer

Scale-down events appear as normal Kubernetes `Killing` events after the benchmark. These are expected and are not treated as failures when pods return to a healthy baseline.

### Observability

Tempo initially failed after high-load trace volume, then was fixed by increasing resources.

After the fix:

    observability-layer: Synced / Healthy
    tempo: 1/1 Running
    /ready: 200 OK

## Current Capacity Statement

Conservative statement:

    Under the current fixed lab resources, the system has confirmed stable behavior at 70 RPS through two repeated runs.

Updated stronger statement:

    Under the current fixed lab resources and after Phase 5 tuning, the system also passed 80 RPS twice with 0% HTTP error rate, all accepted orders completed, and Kafka lag drained after cooldown.

Recommended wording for report:

    The current demonstrated fixed-resource capacity is 80 RPS in the lab environment, with 70 RPS as the conservative safe rating and 80 RPS as the confirmed high-load rating after tuning. Further testing above 80 RPS should be treated as stress-to-break testing, not normal capacity validation.

## Remaining Limitations

The current rating is based on the lab OpenStack cluster and current workload profile.

It should not be generalized to unlimited production scale because true upper bounds depend on:

- PostgreSQL write throughput
- PgBouncer pool settings
- Kafka partitions and broker capacity
- Consumer processing throughput
- HPA/KEDA reaction time
- Node CPU/memory capacity
- Observability overhead
- Network and storage performance

## Next Actions

Do not continue normal capacity benchmarking immediately after 80 RPS.

Recommended next phase:

1. Write Phase 5 final checkpoint.
2. Add alerting rules for Kafka lag, HPA pending pods, Tempo OOM, PostgreSQL pressure, and disk pressure.
3. Add centralized logging with Loki/Promtail or Grafana Alloy.
4. Write runbooks for high Kafka lag, pending pods, DB pressure, and observability failure.
5. Optionally run 90 RPS only as explicit stress-to-break testing.
