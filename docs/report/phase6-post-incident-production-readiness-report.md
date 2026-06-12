# Phase 6 Post-Incident Production Readiness Report

## 1. Executive Summary

This report summarizes the post-incident stabilization work performed on the `my-ecommerce-platform` system after runtime issues were found in the CDC layer and the Saga consumer pipeline.

The system is now stabilized.

The main outcomes are:

- Kafka Connect CDC connectors were restored.
- The Dynamic Redis Filter CDC connector was restored.
- Kafka Connect connector registration was GitOps-hardened using a Kubernetes CronJob.
- A controlled connector-loss recovery proof was completed successfully.
- Saga consumers were recovered by restarting runtime components.
- The end-to-end Saga smoke test passed after recovery.
- Raw evidence, runbooks, and stabilization checkpoints were committed to the repository.

Current final state:

- ArgoCD `cdc-layer`: `Synced` / `Healthy`
- Kafka Connect connectors: present and `RUNNING`
- Debezium normal connector task: `RUNNING`
- Debezium Dynamic Redis Filter connector task: `RUNNING`
- Saga workloads: `Running`
- CDC topics: present
- End-to-end order Saga: verified through smoke test

## 2. Scope

This stabilization report covers the following areas:

- Kafka Connect runtime connector recovery
- Debezium CDC recovery
- Dynamic Redis Filter recovery
- PostgreSQL replication slot handling
- GitOps hardening for connector registration
- Saga consumer recovery
- End-to-end Saga smoke evidence
- ArgoCD synchronization verification
- Remaining operational risks and next hardening recommendations

This report does not introduce new load testing or benchmark runs. The purpose of this phase is stabilization and operational hardening after the incident.

## 3. System Context

The platform is an event-driven microservices system deployed on Kubernetes.

Core runtime components include:

- `order-service`
- `inventory-api`
- `inventory-consumer`
- `payment-api`
- `payment-consumer`
- `notification-api`
- `notification-consumer`
- `read-model-service`
- PostgreSQL
- PgBouncer
- Kafka
- Kafka Connect with Debezium
- Redis
- MongoDB
- ClickHouse
- Prometheus / Grafana
- Tempo / OpenTelemetry
- ArgoCD

The main business flow is based on Saga choreography:

    POST /orders
    -> order.created
    -> inventory reservation
    -> inventory.reserved
    -> payment processing
    -> payment.completed
    -> notification
    -> order status update
    -> read model projection

## 4. Incident 1: Kafka Connect Lost Runtime Connectors

### 4.1 Symptom

Kafka Connect was running, but the Connect REST API returned an empty connector list.

The expected connectors were missing:

- `order-db-orders-connector`
- `order-db-orders-dynamic-filter-connector`

This meant that the Connect worker was alive, but the connector registration state was lost.

### 4.2 Impact

The CDC runtime pipeline was unavailable.

The affected capabilities were:

- PostgreSQL `orders` CDC to Kafka
- Dynamic Redis Filter CDC stream
- CDC topic generation
- CDC evidence continuity after restart/reset

Before recovery, CDC data topics were missing or inactive:

- `cdc.order_db.public.orders`
- `cdc_dynamic.order_db.public.orders`

### 4.3 Recovery

The normal Debezium connector was restored:

- Connector: `order-db-orders-connector`
- Database: `order_db`
- Table: `public.orders`
- Topic prefix: `cdc.order_db`
- Replication slot: `dbz_order_slot`
- Publication: `dbz_order_publication`

The Dynamic Redis Filter connector was restored:

- Connector: `order-db-orders-dynamic-filter-connector`
- Topic prefix: `cdc_dynamic.order_db`
- Filter implementation: `io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter`
- Redis key: `filter:order-status`
- Filter field: `status`
- Active Redis rule: `["PENDING", "COMPLETED"]`

Final connector state:

    order-db-orders-connector: RUNNING / task 0 RUNNING
    order-db-orders-dynamic-filter-connector: RUNNING / task 0 RUNNING

## 5. Incident 2: Dynamic Connector Stale Replication Slot

### 5.1 Symptom

The Dynamic Redis Filter connector initially failed to start its task because PostgreSQL still had a stale logical replication slot.

Observed issue:

    Cannot obtain valid replication slot 'dbz_order_dynamic_filter_slot'

### 5.2 Recovery Method

The PostgreSQL replication slot state was inspected first.

The slot was dropped only after confirming that it was inactive:

    active = false

After the inactive slot was removed, the Dynamic Redis Filter connector was recreated and successfully started.

### 5.3 Final State

The dynamic connector returned to:

    RUNNING / task 0 RUNNING

The connector log confirmed that the Redis rule was being loaded:

    Filter rule updated for 'order-status-filter': status IN [COMPLETED, PENDING]

### 5.4 Important Safety Decision

The system does not automatically drop PostgreSQL replication slots.

This is intentional because dropping replication slots can lose CDC position or interrupt a valid CDC stream. Replication slot cleanup remains a controlled runbook operation.

## 6. GitOps Hardening for CDC Connector Bootstrap

### 6.1 Added GitOps Resource

A new manifest was added:

    k8s/cdc/kafka-connect-connector-bootstrap.yaml

This manifest creates:

- ServiceAccount: `kafka-connect-connector-bootstrap`
- RBAC allowing read access to PostgreSQL Secret in namespace `db`
- ConfigMap containing connector definitions without real passwords
- CronJob: `kafka-connect-connector-bootstrap`

### 6.2 CronJob Behavior

The CronJob runs every 10 minutes:

    */10 * * * *

It applies connector configuration through Kafka Connect REST API:

    PUT /connectors/{name}/config

This makes the operation idempotent:

- If the connector already exists, it updates the connector config.
- If the connector is missing, it recreates the connector.
- The PostgreSQL password is injected from Kubernetes Secret at runtime.
- No database password is committed to Git.

### 6.3 Recovery Proof

A controlled recovery proof was executed.

Before deletion:

    [
      "order-db-orders-dynamic-filter-connector",
      "order-db-orders-connector"
    ]

After deleting both connectors:

    []

After running a manual Job from the CronJob:

    APPLIED connector=order-db-orders-connector status=201
    APPLIED connector=order-db-orders-dynamic-filter-connector status=201
    CONNECTORS=["order-db-orders-dynamic-filter-connector", "order-db-orders-connector"]

After startup:

    order-db-orders-connector: RUNNING / task 0 RUNNING
    order-db-orders-dynamic-filter-connector: RUNNING / task 0 RUNNING

This proves that lost Kafka Connect connector registration can now be recovered through a GitOps-managed bootstrap resource.

## 7. Incident 3: Saga Consumers Stuck After order.created

### 7.1 Symptom

After the CDC layer was restored, the first Saga smoke test still failed.

Observed behavior:

- Order creation returned HTTP `201`.
- Order remained `PENDING`.
- `order.created` outbox event was `PUBLISHED`.
- Kafka topic `order.created` contained messages.
- No inventory reservation was created.
- No payment was created.
- No notification was created.

This proved that CDC was no longer the blocking issue. The remaining issue was in the Kafka consumer / Saga runtime layer.

### 7.2 Consumer Groups

The real consumer groups were confirmed:

- `inventory-service-group`
- `inventory-rollback-group-v2`
- `payment-service-group`
- `notification-service-group`
- `order-service-saga-monitor`
- `read-model-service-group`

### 7.3 Recovery

The following deployments were restarted:

- `inventory-consumer`
- `payment-consumer`
- `notification-consumer`
- `order-service`
- `read-model-service`

After restart, the consumers rejoined Kafka groups and group assignment became visible again.

## 8. End-to-End Saga Smoke Evidence

### 8.1 Smoke Input

A new order was created through the public gateway.

Smoke parameters:

- Gateway: `http://100.65.255.2:30517`
- Product ID: `prod-123`
- User ID: `smoke-user-002`
- Payment method: `COD`
- Total amount: `24000000 VND`
- Order ID: `110b0b08-0364-48d0-ac31-63be2d60f4fa`

### 8.2 Smoke Result

The order successfully moved from `PENDING` to `COMPLETED`.

Observed output:

    order.status=PENDING elapsed=0s
    order.status=COMPLETED elapsed=6s
    SMOKE TEST PASSED

### 8.3 Database Evidence

The database confirmed the expected final state:

- `orders.status = COMPLETED`
- `order_outbox.status = PUBLISHED`
- `inventory_reservations.status = RESERVED`
- `payments.status = COMPLETED`
- `notifications.status = SENT`

### 8.4 Service Log Evidence

Inventory consumer:

    received order.created
    processed and committed order.created

Payment consumer:

    received inventory.reserved
    processed inventory.reserved successfully

Notification consumer:

    received payment.completed
    processed payment.completed successfully

Order service Saga monitor:

    received saga event event_type=payment.completed
    processed and committed saga event
    status=COMPLETED

Read model service:

    upserted order read model

## 9. Final Live Verification

The final live check confirmed the following:

### 9.1 Git State

Latest commit:

    5feff34 docs(ops): add post-incident stabilization checkpoint

Git working tree:

    clean

### 9.2 ArgoCD State

Relevant ArgoCD applications:

- `cdc-layer`: `Synced` / `Healthy`
- `ecommerce-platform`: `Synced` / `Healthy`
- `monitoring-addons`: `Synced` / `Healthy`
- `observability-layer`: `Synced` / `Healthy`
- `analytics-layer`: `Synced` / `Healthy`

The `cdc-layer` application points to:

- Repository: `https://github.com/hoangdonguit/my-ecommerce-platform`
- Path: `k8s/cdc`
- Target revision: `main`
- Synced revision: `5feff3420fa209d07fba31e621b8f7e4d2b4da85`

### 9.3 ArgoCD Tracking

The CDC bootstrap resources are tracked by ArgoCD:

- ServiceAccount tracking:
  `cdc-layer:/ServiceAccount:cdc/kafka-connect-connector-bootstrap`
- ConfigMap tracking:
  `cdc-layer:/ConfigMap:cdc/kafka-connect-connector-bootstrap-config`
- CronJob tracking:
  `cdc-layer:batch/CronJob:cdc/kafka-connect-connector-bootstrap`

### 9.4 Kafka Connect State

Connector list:

    [
      "order-db-orders-dynamic-filter-connector",
      "order-db-orders-connector"
    ]

Connector status:

    order-db-orders-connector: RUNNING / task 0 RUNNING
    order-db-orders-dynamic-filter-connector: RUNNING / task 0 RUNNING

### 9.5 Saga Workload State

Saga runtime workloads are running:

- `inventory-consumer`: `2/2`
- `payment-consumer`: `1/1`
- `notification-consumer`: `1/1`
- `order-service`: `2/2`
- `read-model-service`: `1/1`

### 9.6 Kafka Topics

Required topics are present:

- `order.created`
- `inventory.reserved`
- `inventory.failed`
- `payment.completed`
- `payment.failed`
- `order.cancelled`
- `cdc.order_db.public.orders`
- `cdc_dynamic.order_db.public.orders`
- `__debezium-heartbeat.cdc.order_db`
- Kafka Connect internal topics:
  - `debezium_connect_configs`
  - `debezium_connect_offsets`
  - `debezium_connect_statuses`

## 10. Evidence Artifacts

CDC recovery proof:

    docs/evidence/cdc-gitops-bootstrap-recovery-proof.md

Saga recovery proof:

    docs/evidence/saga-consumer-recovery-proof.md

Saga recovery runbook:

    docs/runbook/saga-consumer-recovery-runbook.md

Post-incident checkpoint:

    docs/evidence/post-incident-stabilization-checkpoint-20260612.md

Raw evidence directories:

    docs/evidence/runs/saga-pending-debug-20260612142909
    docs/evidence/runs/saga-consumer-recovery-20260612150220

## 11. Commit Timeline

CDC connector bootstrap:

    5d54f04 ops(cdc): gitops bootstrap kafka connect connectors

CDC recovery proof:

    453c5a3 docs(cdc): record connector bootstrap recovery proof

Saga recovery proof and runbook:

    05f8b6e docs(saga): record consumer recovery proof and runbook

Raw evidence artifacts:

    fd1998f docs(evidence): add raw saga recovery run artifacts

Post-incident checkpoint:

    5feff34 docs(ops): add post-incident stabilization checkpoint

## 12. Current Production Readiness Assessment

The system is ready for the next hardening step.

Current readiness status:

- Application workloads are running.
- CDC connectors are running.
- Dynamic Redis Filter is running.
- Kafka topics are present.
- Connector bootstrap is GitOps-managed.
- Saga flow passed the light smoke test.
- Evidence and runbooks are committed.
- ArgoCD confirms the CDC layer is synced and healthy.

## 13. Remaining Risks

### 13.1 PostgreSQL Replication Slot Staleness

The system still requires manual runbook handling if a replication slot becomes stale.

Reason:

- Automatic slot deletion is unsafe.
- Slot deletion can lose CDC position.
- Slot deletion must only happen after confirming `active = false`.

### 13.2 Business-Level Saga Stuck Detection

The system currently has evidence and a runbook for Saga recovery, but stronger alerting should be added in a future phase.

Recommended alerts:

- Orders stuck in `PENDING` longer than a threshold.
- Outbox events repeatedly requeued.
- Kafka consumer lag above threshold.
- Kafka Connect connector not `RUNNING`.
- Dynamic connector task missing.
- CDC topic inactivity after expected business events.

### 13.3 No Heavy Retest After Stabilization

No additional heavy benchmark was executed after stabilization. This is intentional because the incident was runtime/operational, not a capacity regression investigation.

The previous capacity and production readiness evidence remains valid, and this phase only adds stabilization evidence after the CDC/Saga incident.

## 14. Next Recommended Step

The next recommended work is not another load test. The next step should be alert hardening:

1. Add Prometheus alerts for Kafka Connect connector health.
2. Add alert for failed CDC bootstrap CronJob.
3. Add alert for orders stuck in `PENDING`.
4. Add alert for repeated outbox requeue.
5. Add alert for Kafka consumer lag or missing assignment.
6. Add a short evidence document proving the alerts are loaded by Prometheus.

## 15. Conclusion

The post-incident stabilization work is complete.

The CDC layer has been restored and GitOps-hardened. The Dynamic Redis Filter connector has been restored and verified. The Saga runtime has been recovered and verified through an end-to-end smoke test. ArgoCD confirms that the CDC layer is synced and healthy at the latest repository revision.

The system is stable enough to proceed to the next hardening phase, which should focus on alerting for the exact operational failure modes discovered during this incident.
