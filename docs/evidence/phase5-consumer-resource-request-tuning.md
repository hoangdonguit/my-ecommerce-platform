# Phase 5 Consumer Resource Request Tuning Evidence

## Scope

This document records the Phase 5 consumer resource request tuning for `my-ecommerce-platform`.

The goal is to reduce scheduling pressure observed during 70 RPS benchmarks without lowering runtime limits.

## Background

During the Phase 5 post-placement 70 RPS benchmark, the system completed all accepted orders and drained Kafka lag after cooldown, but consumer scale-out still produced scheduling warnings:

    0/3 nodes are available:
      1 node(s) didn't match Pod's node affinity/selector
      2 Insufficient cpu

The consumer workloads were light in actual CPU and memory usage, but each pod still requested:

    cpu: 200m
    memory: 256Mi

When KEDA scaled consumers aggressively, these requests consumed schedulable CPU quickly.

## Change

The following consumers were updated:

- inventory-consumer
- payment-consumer
- notification-consumer

Requests were changed from:

    cpu: 200m
    memory: 256Mi

to:

    cpu: 50m
    memory: 128Mi

Limits were intentionally not reduced:

    cpu: 500m
    memory: 512Mi

This preserves runtime headroom during startup, restart, and traffic spikes.

## Git Commit

    ef3b552 capacity: reduce consumer resource requests

## Runtime Verification

After ArgoCD sync, live deployment resources showed:

    inventory-consumer:
      requests.cpu: 50m
      requests.memory: 128Mi
      limits.cpu: 500m
      limits.memory: 512Mi

    payment-consumer:
      requests.cpu: 50m
      requests.memory: 128Mi
      limits.cpu: 500m
      limits.memory: 512Mi

    notification-consumer:
      requests.cpu: 50m
      requests.memory: 128Mi
      limits.cpu: 500m
      limits.memory: 512Mi

ArgoCD applications were Synced / Healthy.

Git state was clean:

    origin/main...HEAD: 0 0

## Smoke Test

Smoke run:

    consumer-request-smoke-20260607111946

Result:

    HTTP_CODE=201
    order.status=PENDING at elapsed=1s
    order.status=COMPLETED at elapsed=6s

Database result:

    orders.status: COMPLETED
    outbox.status: PUBLISHED
    inventory_reservations.status: RESERVED
    payments.status: COMPLETED
    notifications.status: SENT

Kafka lag after smoke:

    NO_NONZERO_NUMERIC_LAG

Smoke verdict:

    PASS

## Restart Simulation

The following deployments were rollout restarted:

- inventory-consumer
- payment-consumer
- notification-consumer

Result:

    inventory-consumer: successfully rolled out
    payment-consumer: successfully rolled out
    notification-consumer: successfully rolled out

New pods were Running with:

    restartCount: 0

No FailedScheduling, OOMKilled, BackOff, ImagePull, Readiness, or Liveness failure was observed.

The observed `Killing` events were normal rollout events for stopping old pods.

## Node Allocated Resources After Tuning

After request reduction:

    vm1-gateway:
      cpu requests: 2700m / 4000m = 67%
      memory requests: 5778Mi = 72%

    vm2-mesh:
      cpu requests: 1875m / 4000m = 46%
      memory requests: 3116Mi = 39%

    vm3-gitops:
      cpu requests: 1525m / 4000m = 38%
      memory requests: 3584Mi = 45%

## Verdict

PASS.

The consumer request reduction was applied successfully and did not break the Saga smoke flow or consumer restart behavior.

Limits were preserved, so runtime headroom remains available.

## Next Action

Rerun 70 RPS after consumer request tuning.

The goal is to verify whether the previous consumer FailedScheduling warning is reduced or removed.
