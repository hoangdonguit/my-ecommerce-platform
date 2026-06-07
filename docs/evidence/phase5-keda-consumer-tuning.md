# Phase 5 KEDA Consumer Tuning Evidence

## Scope

This document records the Phase 5 KEDA tuning for Saga consumer autoscaling.

The goal is to reduce excessive consumer scale-out and Kafka consumer group rebalancing observed during the 70 RPS benchmark after consumer request reduction.

## Background

Before this tuning, the consumer resources had already been reduced to:

    requests.cpu: 50m
    requests.memory: 128Mi
    limits.cpu: 500m
    limits.memory: 512Mi

That reduced scheduling pressure, but the 70 RPS benchmark still showed high transient Kafka lag and consumer group rebalancing.

## Change

The following KEDA ScaledObjects were tuned:

- inventory-consumer-scaler
- payment-consumer-scaler
- notification-consumer-scaler

Changed values:

    cooldownPeriod: 60 -> 180
    maxReplicaCount: 16 -> 8
    lagThreshold: "20" -> "80"
    activationLagThreshold: "1" -> "20"

Polling interval remained:

    pollingInterval: 15

## Git Commit

    f6a1af5 capacity: tune consumer keda scaling

## Runtime Verification

After ArgoCD sync:

    ecommerce-platform: Synced / Healthy
    revision: f6a1af5
    git status: clean
    origin/main...HEAD: 0 0

Live ScaledObjects showed:

    maxReplicaCount: 8
    cooldownPeriod: 180
    lagThreshold: "80"
    activationLagThreshold: "20"

HPA also reflected the new KEDA target:

    inventory-consumer-scaler:
      target: 80/80
      maxPods: 8

    payment-consumer-scaler:
      target: 80/80
      maxPods: 8

    notification-consumer-scaler:
      target: 80/80 and 80/80
      maxPods: 8

## Smoke Test

Smoke run:

    keda-tuning-smoke-20260607133447

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

Pod state after smoke:

    no abnormal non-Running pod remained

## Verdict

PASS.

KEDA tuning was applied successfully, live HPA reflected the new max/target settings, and the full Saga smoke flow still passed.

## Next Action

Rerun 70 RPS after KEDA tuning.

The target is to compare against the previous 70 RPS runs:

- Phase 5 post-placement 70 RPS
- Phase 5 post-consumer-request 70 RPS

Key things to watch:

- HTTP/k6 success
- immediate Kafka lag
- consumer group rebalancing
- KEDA scale behavior
- order completion time
- FailedScheduling / Insufficient CPU events
