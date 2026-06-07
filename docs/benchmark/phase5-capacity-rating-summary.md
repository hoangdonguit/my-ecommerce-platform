# Phase 5 Capacity Rating Summary

## Scope

This document summarizes the Phase 5 capacity-rating result for `my-ecommerce-platform`.

The goal is to answer the teacher's capacity planning question:

1. What load level can the system handle under the current fixed lab resources?
2. What bottlenecks or warnings remain before claiming a higher capacity rating?

This file intentionally summarizes important milestones only. Small sanity checks are not documented separately.

## Current Fixed-Resource Environment

The current benchmark environment is the existing lab Kubernetes cluster on OpenStack VMs.

Runtime URL:

    http://100.65.255.2:30517

Current scaling model:

- order-service: HPA by CPU
- inventory-consumer: KEDA by Kafka lag
- payment-consumer: KEDA by Kafka lag
- notification-consumer: KEDA by Kafka lag
- inventory/payment/notification consumers use reduced resource requests:
  - requests.cpu: 50m
  - requests.memory: 128Mi
  - limits.cpu: 500m
  - limits.memory: 512Mi

Important GitOps fix:

- ArgoCD ignores `/spec/replicas` for autoscaled workloads.
- This prevents ArgoCD from fighting HPA/KEDA over live replica counts.

## Important Phase 5 Fixes Before Final Capacity Runs

### Node Placement Fix

Earlier, some stateless workloads were too restricted by placement rules, causing scheduling pressure under high load.

The placement was relaxed so stateless services could run on eligible worker nodes instead of being pinned too narrowly.

### Consumer Resource Request Tuning

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

### KEDA Tuning

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
- Earlier 70 RPS results showed inventory lag as a first-stage bottleneck.
- Keeping inventory-consumer warm at 2 replicas significantly improved end-to-end completion time.

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

70 RPS is currently the strongest fixed-resource rated capacity candidate because:

- It passed twice after the final tuning.
- It had 0% HTTP error rate.
- All accepted orders reached COMPLETED.
- Kafka lag drained after cooldown.
- No FailedScheduling, Insufficient CPU, OOM, BackOff, or ImagePull failure was observed.
- ArgoCD remained Synced / Healthy.

## 80 RPS Upper-Bound Run

80 RPS was run once as an upper-bound / stress-to-break candidate.

Result:

    accepted_orders: 4801
    checks_succeeded: 4801 / 4801
    http_req_failed: 0.00%
    unexpected_error_rate: 0.00%
    http_req_duration avg: 63.88ms
    http_req_duration p95: 143.05ms
    http_req_duration max: 621.62ms
    orders COMPLETED: 4801
    avg_complete_seconds: about 23.13s
    max_complete_seconds: about 48.06s
    Kafka after cooldown: NO_NONZERO_NUMERIC_LAG

80 RPS showed transient Kafka lag and autoscaling activity, but the system drained successfully after cooldown.

### 80 RPS Verdict

PASS WITH OBSERVATION / UPPER-BOUND PASS CANDIDATE.

80 RPS passed once, but it should not yet replace 70 RPS as the official fixed-resource capacity rating until it is repeated successfully.

## Operational Observations

### Kafka Lag

Transient Kafka lag is expected immediately after high-load runs because the system uses asynchronous Saga processing.

The important pass condition is:

    Kafka lag must drain to NO_NONZERO_NUMERIC_LAG after cooldown.

This condition passed for the final 70 RPS and 80 RPS runs.

### Autoscaling

During high load, HPA/KEDA scaled several workloads up:

- order-service
- inventory-consumer
- payment-consumer
- notification-consumer

Scale-down events appear as normal Kubernetes `Killing` events after the benchmark. These are expected and are not treated as failures when pods return to a healthy baseline.

### Observability Warning

Tempo in the observability layer showed a temporary OOMKilled / BackOff warning during later checks, but the business/data path remained healthy and ArgoCD eventually returned to Healthy.

This should be tracked as an observability-layer tuning item, not as a business-path benchmark failure.

## Current Capacity Statement

Current conservative statement:

    Under the current fixed lab resources, the system has confirmed stable behavior at 70 RPS through two repeated runs.

Extended observation:

    The system has also passed one 80 RPS upper-bound run, but 80 RPS requires one repeat run before it can be promoted to an official fixed-resource capacity rating.

## Decision Rule

If repeat 80 RPS also passes with:

- 0% HTTP error rate
- all orders COMPLETED
- Kafka lag drains after cooldown
- no FailedScheduling / Insufficient CPU / OOM / BackOff
- ArgoCD returns to Synced / Healthy

then 80 RPS can be promoted to the next fixed-resource capacity candidate.

If repeat 80 RPS fails or shows persistent backlog/resource problems, keep the official fixed-resource rating at 70 RPS and classify 80 RPS as the upper-bound stress boundary.

## Next Action

Run one repeat 80 RPS test.

Do not proceed to 90 RPS until 80 RPS repeat is analyzed and documented.
