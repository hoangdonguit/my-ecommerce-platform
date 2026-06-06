# Phase 5 Node Placement Fix Evidence

## Scope

This document records the Phase 5 node placement fix for stateless/scalable services in `my-ecommerce-platform`.

The goal is to reduce the scheduling bottleneck observed during 60/70 RPS capacity tests, where HPA/KEDA attempted to scale pods but Kubernetes reported:

    0/3 nodes are available:
      1 Insufficient cpu
      2 node(s) didn't match Pod's node affinity/selector

## Problem

Before this fix, many stateless API/consumer workloads were pinned to `vm2-mesh` through hard `nodeSelector`.

Affected workloads included:

- order-service
- inventory-api
- inventory-consumer
- payment-api
- payment-consumer
- notification-api
- notification-consumer
- read-model-service

This meant scale-out pods could only schedule on `vm2-mesh`.

When vm2-mesh ran out of schedulable CPU, extra pods became Pending even though vm1-gateway still had capacity.

## Change

The hard `nodeSelector` for the scalable/stateless service group was replaced with required node affinity allowing:

    vm1-gateway
    vm2-mesh

The fix was applied to:

- k8s/services/order-service.yaml
- k8s/services/inventory-service.yaml
- k8s/services/payment-service.yaml
- k8s/services/notification-service.yaml
- k8s/services/read-model-service.yaml

Git commit:

    0586c4c capacity: relax stateless service node placement

Stateful/data/GitOps components were not moved.

The fix intentionally did not move PostgreSQL, Kafka, ClickHouse, MongoDB, Redis, PgBouncer, ArgoCD, or monitoring workloads.

## Runtime Verification

After ArgoCD sync:

    ecommerce-platform: Synced / Healthy
    revision: 0586c4c
    git status: clean
    origin/main...HEAD: 0 0

Live deployment placement showed:

    nodeSelector: empty
    nodeAffinity: kubernetes.io/hostname In [vm1-gateway, vm2-mesh]

for:

- order-service
- inventory-api
- inventory-consumer
- payment-api
- payment-consumer
- notification-api
- notification-consumer
- read-model-service

## Pod Distribution After Fix

Runtime placement after rollout showed pods distributed across vm1-gateway and vm2-mesh.

Examples:

    order-service:
      vm1-gateway
      vm2-mesh

    inventory-api:
      vm1-gateway

    inventory-consumer:
      vm1-gateway

    payment-api:
      vm2-mesh

    payment-consumer:
      vm1-gateway

    notification-api:
      vm2-mesh

    notification-consumer:
      vm1-gateway

    read-model-service:
      vm1-gateway

This confirms that the scalable stateless services are no longer pinned only to vm2-mesh.

## Smoke Test

Smoke run:

    node-placement-smoke-20260607003740

Created order:

    order_id: 6a775922-a4ea-4713-90ee-1ade7080f615
    payment_method: COD
    total_amount: 24000000

Smoke result:

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

Smoke verdict:

    PASS

## Verdict

PASS.

The node placement fix was applied successfully and the system still passed the full Saga smoke flow.

This fix addresses one of the main bottlenecks observed in the 60/70 RPS tests: scale-out pods were previously blocked by hard node placement on vm2-mesh.

## Next Action

Rerun 60 RPS after the placement fix.

The goal is to verify whether the previous 60 RPS Pending pod / FailedScheduling behavior is reduced or eliminated.

