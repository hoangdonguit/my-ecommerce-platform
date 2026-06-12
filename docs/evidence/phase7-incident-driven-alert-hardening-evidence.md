# Phase 7 Incident-Driven Alert Hardening Evidence

## 1. Purpose

This document records the Phase 7 incident-driven alert hardening work.

Phase 7 was created after the CDC and Saga runtime incident. Instead of adding generic alerts, this phase adds health probes and Prometheus alerts for the exact failure modes observed during the incident.

## 2. Scope

Phase 7 covers:

- Kafka Connect connector health.
- Kafka Connect connector bootstrap job health.
- Saga pending order detection.
- Saga runtime availability.
- Saga runtime pod restarts.
- Evidence from manual probe jobs.

No heavy benchmark was executed in this phase. The purpose is operational hardening, not capacity testing.

## 3. Added Kubernetes Resources

### 3.1 Kafka Connect Connector Health Probe

Manifest:

    k8s/cdc/kafka-connect-connector-health-probe.yaml

Resource:

    CronJob/cdc/kafka-connect-connector-health-probe

Schedule:

    */5 * * * *

Purpose:

- Query Kafka Connect REST API.
- Confirm both required connectors exist.
- Confirm each connector is `RUNNING`.
- Confirm connector tasks exist.
- Confirm each task is `RUNNING`.

Required connectors:

- `order-db-orders-connector`
- `order-db-orders-dynamic-filter-connector`

### 3.2 Saga Pending Order Health Probe

Manifest:

    k8s/monitoring/saga-pending-order-health-probe.yaml

Resource:

    CronJob/db/saga-pending-order-health-probe

Schedule:

    */5 * * * *

Purpose:

- Query PostgreSQL `order_db`.
- Count orders stuck in `PENDING` longer than 10 minutes.
- Fail the Job if stale pending orders exist.

Probe query:

    select count(*)
    from orders
    where status = 'PENDING'
      and created_at < now() - interval '10 minutes';

Expected healthy result:

    stale_pending_orders=0

### 3.3 Incident-Driven PrometheusRule

Manifest:

    k8s/monitoring/phase7-incident-driven-alert-rules.yaml

Resource:

    PrometheusRule/monitoring/phase7-incident-driven-alert-rules

## 4. Alert Rules Added

The PrometheusRule contains seven alert rules:

- `KafkaConnectDeploymentUnavailable`
- `KafkaConnectConnectorHealthProbeFailed`
- `KafkaConnectBootstrapJobFailed`
- `KafkaConnectCronJobSuspended`
- `SagaPendingOrderProbeFailed`
- `SagaRuntimeDeploymentUnavailable`
- `SagaRuntimePodRestarting`

## 5. Manual Probe Evidence

### 5.1 Kafka Connect Connector Health Probe

Manual Job:

    kafka-connect-connector-health-probe-manual-20260612160509

Result:

    connectors=["order-db-orders-dynamic-filter-connector", "order-db-orders-connector"]
    order-db-orders-connector: connector=RUNNING tasks=['RUNNING']
    order-db-orders-dynamic-filter-connector: connector=RUNNING tasks=['RUNNING']
    OK all required Kafka Connect connectors are RUNNING

Evidence file:

    docs/evidence/runs/phase7-connect-probe-kafka-connect-connector-health-probe-manual-20260612160509.txt

### 5.2 Saga Pending Order Health Probe

Manual Job:

    saga-pending-order-health-probe-manual-20260612160514

Result:

    stale_pending_orders=0

Evidence file:

    docs/evidence/runs/phase7-saga-pending-probe-saga-pending-order-health-probe-manual-20260612160514.txt

## 6. Live Resource Verification

The following live resources were created successfully:

    CronJob/cdc/kafka-connect-connector-health-probe
    CronJob/db/saga-pending-order-health-probe
    PrometheusRule/monitoring/phase7-incident-driven-alert-rules

The PrometheusRule contains the expected alert names:

    KafkaConnectBootstrapJobFailed
    KafkaConnectConnectorHealthProbeFailed
    KafkaConnectCronJobSuspended
    KafkaConnectDeploymentUnavailable
    SagaPendingOrderProbeFailed
    SagaRuntimeDeploymentUnavailable
    SagaRuntimePodRestarting

## 7. Commit

Phase 7 implementation commit:

    383a0ab ops(alerts): add incident-driven health probes and alerts

## 8. Conclusion

Phase 7 added incident-driven alert hardening for CDC and Saga failure modes discovered during the post-incident stabilization work.

The added probes passed manual verification:

- Kafka Connect required connectors are present and running.
- Saga has zero stale pending orders.
- The PrometheusRule contains all expected alert rules.

The system is now better prepared to detect recurrence of the exact runtime failures observed in the incident.
