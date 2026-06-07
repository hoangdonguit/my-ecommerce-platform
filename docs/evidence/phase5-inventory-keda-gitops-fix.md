# Phase 5 Inventory KEDA and GitOps Replica Drift Fix Evidence

## Scope

This document records the Phase 5 fix for the inventory-consumer KEDA tuning and ArgoCD replica drift issue.

The goal was to tune the first Saga consumer stage and prevent ArgoCD from fighting KEDA/HPA over Deployment replicas.

## Problem

After setting:

    inventory-consumer-scaler minReplicaCount: 2

the cluster showed unstable ArgoCD sync behavior.

Root cause:

    ArgoCD was still comparing Deployment.spec.replicas for autoscaled workloads.
    KEDA/HPA wanted inventory-consumer replicas = 2.
    Git desired state still had inventory-consumer replicas = 1.
    Therefore, ecommerce-platform kept switching between Synced and OutOfSync.

An intermediate patch also produced an invalid YAML indentation in the ArgoCD Application manifest. This was fixed before proceeding.

## GitOps Fix

The ArgoCD Application manifest was repaired and expanded with ignoreDifferences for autoscaled Deployment replicas.

Commit:

    65d327c fix: repair ecommerce argocd ignore differences

The following workloads now ignore `/spec/replicas` drift:

- order-service
- inventory-api
- inventory-consumer
- payment-api
- payment-consumer
- notification-api
- notification-consumer

The Application keeps:

    RespectIgnoreDifferences=true

## Inventory KEDA Tuning

Inventory consumer was tuned separately because the previous 70 RPS run showed inventory-service-group/order.created as the first-stage bottleneck.

Commit:

    cb6d1cf capacity: tune inventory consumer keda scaling

Inventory scaler values:

    minReplicaCount: 2
    maxReplicaCount: 8
    cooldownPeriod: 180
    lagThreshold: "40"
    activationLagThreshold: "10"

Payment and notification KEDA settings were left unchanged:

    minReplicaCount: 1
    maxReplicaCount: 8
    cooldownPeriod: 180
    lagThreshold: "80"
    activationLagThreshold: "20"

## Runtime Stability Verification

After applying the fixed ArgoCD Application spec:

    ecommerce-platform: Synced / Healthy
    revision: 65d327c
    inventory-consumer: 2 / 2
    payment-consumer: 1 / 1
    notification-consumer: 1 / 1

The state remained stable across repeated checks.

## Smoke Test

Smoke run:

    inventory-keda-tuning-smoke-20260607210530

Result:

    HTTP_CODE=201
    order.status=PENDING at elapsed=2s
    order.status=COMPLETED at elapsed=7s

Database result:

    orders.status: COMPLETED
    outbox.status: PUBLISHED
    inventory_reservations.status: RESERVED
    payments.status: COMPLETED
    notifications.status: SENT

Kafka lag after smoke:

    NO_NONZERO_NUMERIC_LAG

Final state:

    all ArgoCD applications: Synced / Healthy
    no abnormal non-Running pod
    git status: clean
    origin/main...HEAD: 0 0

## Verdict

PASS.

The GitOps replica drift was fixed, inventory-consumer remained stable at 2 replicas, and the full Saga smoke flow passed.

## Next Action

Rerun 70 RPS after inventory-specific KEDA tuning.

Key metrics to compare:

- immediate inventory-service-group lag
- consumer group rebalancing
- order completion avg/max time
- FailedScheduling / Insufficient CPU events
- Kafka lag after cooldown
