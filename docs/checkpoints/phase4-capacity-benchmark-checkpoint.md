# Phase 4 Capacity Benchmark Checkpoint

## Purpose

This checkpoint records the current system state after Phase 4 hardening and capacity benchmarking.

It is intended to help future report writing and future ChatGPT sessions understand what has been completed.

## Current GitOps State

At the end of the 70 RPS stress-finding phase:

    git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy

Current applications include:

    analytics-layer
    cdc-layer
    ecommerce-infrastructure
    ecommerce-platform
    infrastructure-layer
    monitoring-addons
    observability-layer
    security-layer

## Current Runtime State

After the 70 RPS stress-finding run, the cluster eventually returned to baseline:

    Kafka lag: NO_NONZERO_NUMERIC_LAG
    no pods remained Pending
    order-service returned to 2 replicas
    inventory-consumer returned to 1 replica
    payment-consumer returned to 1 replica
    notification-consumer returned to 1 replica

However, the run still confirmed stress pressure during high load.

## Completed Phase 4 Proofs

The following important proof documents were created:

- Grafana no-data time sync investigation
- payment outbox runtime proof
- Redis password fallback hardening
- PgBouncer runtime stats audit
- ClickHouse CDC freshness proof
- Istio runtime AuthZ/mTLS proof
- post-hardening smoke check
- post-hardening 5 RPS benchmark
- post-hardening 10 RPS benchmark
- post-hardening 20 RPS benchmark
- post-hardening 30 RPS benchmark
- post-hardening 40 RPS benchmark
- post-hardening 50 RPS benchmark
- post-hardening 60 RPS benchmark
- 70 RPS stress finding
- capacity rating methodology

## Capacity Rating

Current fixed-resource interpretation:

    Safe rated capacity:
        40 RPS, about 2400 orders/minute

    Burst capacity:
        50 RPS, about 3000 orders/minute

    Overload warning zone:
        60 RPS, about 3600 orders/minute

    Stress-finding zone:
        70 RPS, about 4200 orders/minute

## Main Finding

The API hot path is not the first observed bottleneck.

The system can accept order requests up to 70 RPS in k6 tests with 0% HTTP failure.

However, full-system capacity must include asynchronous drain and Kubernetes stability.

At higher load, the main observed limits are:

- Kafka transient lag
- consumer group rebalancing
- increased order completion time
- HPA/KEDA scale pressure
- Pending pods
- FailedScheduling due to Insufficient CPU and node affinity/selector constraints
- image filesystem pressure on vm3-gitops

## Known Risks

Known risks after Phase 4:

1. vm3-gitops image filesystem pressure around 85%.
2. node affinity/selector limits pod placement.
3. consumer scaling can request more pods than the current cluster can schedule.
4. completion time increases significantly from 50 RPS onward.
5. current safe capacity should not be overstated as 60 or 70 RPS.

## Recommended Next Phase

Recommended next phase: Capacity hardening and observability completion.

Main tasks:

1. Fix or document node affinity/selector constraints.
2. Clean or expand vm3-gitops disk/image filesystem.
3. Review consumer CPU/memory requests.
4. Review KEDA lag target and cooldown behavior.
5. Add centralized logging with Loki.
6. Add alert rules for Kafka lag, Pending pods, PgBouncer waiting, disk pressure, and high completion time.
7. Rerun 60 and 70 RPS after fixes.
8. Only then consider 80 RPS stress-to-break.

