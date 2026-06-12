# Post-Incident Stabilization Checkpoint - 2026-06-12

## 1. Purpose

This checkpoint summarizes the post-incident stabilization work performed after the Kafka Connect, CDC, and Saga runtime issues.

The goals of this document are to:

- Record what failed.
- Record how the system was recovered.
- Record the evidence and operational proof.
- Identify which recovery paths are now GitOps-managed.
- Identify which risks still require manual runbooks or future alerting.
- Provide a clean handoff document for later project phases.

## 2. Incident Summary

Two independent runtime problems were observed.

First, Kafka Connect was still running, but the registered connectors were missing from the Connect REST API. This broke the CDC pipeline even though the Kafka Connect worker itself was healthy.

Second, after CDC was recovered, the Saga flow was still stuck after `order.created`. Orders were created successfully and the outbox event was published, but downstream consumers did not continue the workflow until the Saga runtime components were restarted.

## 3. CDC / Kafka Connect Issue

### 3.1 Symptom

Kafka Connect returned an empty connector list:

    []

This meant that Kafka Connect was alive, but the runtime connector registration was lost.

The missing connectors were:

- `order-db-orders-connector`
- `order-db-orders-dynamic-filter-connector`

The CDC data topics were also missing before a new CDC event was generated:

- `cdc.order_db.public.orders`
- `cdc_dynamic.order_db.public.orders`

### 3.2 Impact

The CDC pipeline was not available at runtime.

The following capabilities were affected:

- PostgreSQL-to-Kafka CDC for `public.orders`
- Dynamic Redis Filter CDC stream
- Kafka CDC topic generation
- CDC evidence continuity after restart/reset

### 3.3 Runtime Recovery

The normal Debezium connector was restored:

- Connector: `order-db-orders-connector`
- Database: `order_db`
- Table: `public.orders`
- Topic prefix: `cdc.order_db`
- Replication slot: `dbz_order_slot`
- Publication: `dbz_order_publication`

The Dynamic Redis Filter connector was also restored:

- Connector: `order-db-orders-dynamic-filter-connector`
- Topic prefix: `cdc_dynamic.order_db`
- Filter implementation: `io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter`
- Redis key: `filter:order-status`
- Filter field: `status`
- Active Redis rule: `["PENDING", "COMPLETED"]`

Final runtime status:

    order-db-orders-connector: RUNNING / task 0 RUNNING
    order-db-orders-dynamic-filter-connector: RUNNING / task 0 RUNNING

CDC topics appeared again after a new order event was created:

- `cdc.order_db.public.orders`
- `cdc_dynamic.order_db.public.orders`
- `__debezium-heartbeat.cdc.order_db`

## 4. Dynamic Connector Replication Slot Issue

### 4.1 Symptom

The Dynamic Redis Filter connector initially failed to start its task due to a stale PostgreSQL replication slot.

The observed error was:

    Cannot obtain valid replication slot 'dbz_order_dynamic_filter_slot'

### 4.2 Recovery Approach

The replication slot was inspected through PostgreSQL system metadata.

The slot was only dropped after confirming that it was inactive:

    active = false

After the inactive stale slot was removed, the Dynamic Redis Filter connector was recreated and its task started successfully.

### 4.3 Result

The dynamic connector returned to:

    RUNNING / task 0 RUNNING

The connector log confirmed that the Redis rule was loaded repeatedly:

    Filter rule updated for 'order-status-filter': status IN [COMPLETED, PENDING]

### 4.4 Operational Limit

The bootstrap automation intentionally does not auto-drop PostgreSQL replication slots.

Reason:

- Dropping a replication slot can lose CDC position.
- Dropping an active slot can interrupt a valid CDC stream.
- Slot recovery must remain a controlled runbook action.

## 5. GitOps Hardening for Kafka Connect Connectors

### 5.1 Added Manifest

The following manifest was added:

    k8s/cdc/kafka-connect-connector-bootstrap.yaml

It creates:

- ServiceAccount: `kafka-connect-connector-bootstrap`
- RBAC to read the PostgreSQL Secret from namespace `db`
- ConfigMap containing connector definitions
- CronJob: `kafka-connect-connector-bootstrap`

### 5.2 Schedule

The CronJob runs every 10 minutes:

    */10 * * * *

### 5.3 Behavior

The CronJob uses the Kafka Connect REST API:

    PUT /connectors/{name}/config

This makes the bootstrap process idempotent:

- If the connector exists, its config is updated.
- If the connector is missing, it is recreated.
- PostgreSQL password is injected at runtime from Kubernetes Secret.
- No database password is committed to Git.

### 5.4 Chaos Recovery Proof

A controlled recovery proof was executed.

Initial connectors:

    [
      "order-db-orders-dynamic-filter-connector",
      "order-db-orders-connector"
    ]

Both connectors were deleted from Kafka Connect runtime.

Connector list after deletion:

    []

A manual Job was created from the CronJob.

Bootstrap Job result:

    APPLIED connector=order-db-orders-connector status=201
    APPLIED connector=order-db-orders-dynamic-filter-connector status=201
    CONNECTORS=["order-db-orders-dynamic-filter-connector", "order-db-orders-connector"]

After waiting for startup, both connectors returned to:

    RUNNING / task 0 RUNNING

This proves that lost connector runtime registration can now be recovered through GitOps-managed bootstrap resources.

## 6. Saga Runtime Issue

### 6.1 Symptom

After CDC was restored, the first light Saga smoke test still failed.

Observed behavior:

- Order creation returned HTTP `201`.
- Order status remained `PENDING`.
- Order outbox event `order.created` was `PUBLISHED`.
- Kafka topic `order.created` contained messages.
- No inventory reservation was created.
- No payment was created.
- No notification was created.

This indicated that CDC was no longer the blocking issue. The remaining problem was in the Kafka consumer / Saga runtime layer.

### 6.2 Consumer Groups

The real consumer groups were confirmed as:

- `inventory-service-group`
- `inventory-rollback-group-v2`
- `payment-service-group`
- `notification-service-group`
- `order-service-saga-monitor`
- `read-model-service-group`

### 6.3 Runtime Recovery

The following deployments were restarted:

- `inventory-consumer`
- `payment-consumer`
- `notification-consumer`
- `order-service`
- `read-model-service`

After the rollout restart, the consumers rejoined Kafka groups and group assignment became visible again.

## 7. Saga Smoke Test Evidence

### 7.1 Test Input

A new order was created through the gateway.

Smoke test parameters:

- Gateway: `http://100.65.255.2:30517`
- Product ID: `prod-123`
- User ID: `smoke-user-002`
- Payment method: `COD`
- Total amount: `24000000 VND`
- Order ID: `110b0b08-0364-48d0-ac31-63be2d60f4fa`

### 7.2 Result

The order moved from `PENDING` to `COMPLETED`.

Observed smoke output:

    order.status=PENDING elapsed=0s
    order.status=COMPLETED elapsed=6s
    SMOKE TEST PASSED

### 7.3 Database Evidence

The database confirmed the expected final state:

- `orders.status = COMPLETED`
- `order_outbox.status = PUBLISHED`
- `inventory_reservations.status = RESERVED`
- `payments.status = COMPLETED`
- `notifications.status = SENT`

### 7.4 Service Log Evidence

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

## 8. Evidence Artifacts

CDC GitOps recovery proof:

    docs/evidence/cdc-gitops-bootstrap-recovery-proof.md

Saga consumer recovery proof:

    docs/evidence/saga-consumer-recovery-proof.md

Saga recovery runbook:

    docs/runbook/saga-consumer-recovery-runbook.md

Raw evidence directories:

    docs/evidence/runs/saga-pending-debug-20260612142909
    docs/evidence/runs/saga-consumer-recovery-20260612150220

Raw evidence commit:

    fd1998f docs(evidence): add raw saga recovery run artifacts

## 9. Commit Timeline

CDC connector bootstrap manifest:

    5d54f04 ops(cdc): gitops bootstrap kafka connect connectors

CDC recovery proof:

    453c5a3 docs(cdc): record connector bootstrap recovery proof

Saga proof and runbook:

    05f8b6e docs(saga): record consumer recovery proof and runbook

Raw Saga recovery artifacts:

    fd1998f docs(evidence): add raw saga recovery run artifacts

## 10. Current Stabilized State

The expected stabilized state after recovery is:

- Kafka Connect worker is running.
- Normal Debezium connector is running.
- Dynamic Redis Filter connector is running.
- CDC topics are present.
- Connector bootstrap CronJob exists and has completed successfully.
- Saga consumers have rejoined Kafka groups.
- Light Saga smoke test passed.
- Order lifecycle works end-to-end from `PENDING` to `COMPLETED`.
- Inventory, payment, notification, order status update, and read model projection are all confirmed by logs and database output.

## 11. Remaining Risks

### 11.1 Stale PostgreSQL Replication Slot

The bootstrap CronJob does not auto-drop replication slots.

Mitigation:

- Use the CDC runbook.
- Check `pg_replication_slots`.
- Only drop a slot when `active = false`.

### 11.2 Saga Consumer Group Stuck After Severe Runtime Reset

If Saga gets stuck again after a severe Kafka/runtime reset:

- Check consumer group assignment.
- Check topic offsets.
- Restart Saga runtime components if consumers do not rejoin.
- Run the light Saga smoke test again.

### 11.3 Missing Alert for Business-Level Stuck Orders

The system now has technical recovery evidence, but a future phase should add alerts for:

- Orders stuck in `PENDING` longer than a threshold.
- Outbox events repeatedly requeued.
- Kafka consumer lag above threshold.
- Kafka Connect connector not `RUNNING`.
- Dynamic connector task missing.

## 12. Conclusion

The system has been stabilized after the incident.

CDC is now hardened through a GitOps-managed Kafka Connect connector bootstrap CronJob. The Dynamic Redis Filter connector was restored and verified. The Saga runtime was recovered by restarting the consumer and orchestration components, after which the end-to-end order flow passed a light smoke test.

The remaining risks are documented and should be converted into alerts or automated remediation in a later hardening phase.
