# Phase 7 Incident-Driven Alert Hardening Report

## 1. Executive Summary

Phase 7 adds incident-driven alerting and health probes to the `my-ecommerce-platform` system.

This phase was created after the post-incident stabilization work in Phase 6. The goal was to convert the observed failure modes into operational detection mechanisms.

The implemented hardening covers:

- Kafka Connect connector health.
- Kafka Connect connector bootstrap failure.
- Kafka Connect recovery CronJob suspension.
- Saga pending order detection.
- Saga runtime deployment availability.
- Saga runtime pod restarts.

Phase 7 did not run a heavy benchmark. This phase focuses on operational detection and post-incident hardening.

## 2. Background

During the previous incident, the system experienced two main problems.

First, Kafka Connect was running but its connector registration was lost. This broke CDC at runtime even though the Connect worker itself was healthy.

Second, after CDC recovery, the Saga flow became stuck after `order.created`. Orders were created successfully, but the downstream Saga pipeline did not continue until the consumer runtime components were restarted.

Phase 7 addresses these issues by adding probes and Prometheus alerts around the exact failure points.

## 3. Implemented Components

### 3.1 Kafka Connect Connector Health Probe

A new CronJob was added:

    kafka-connect-connector-health-probe

Namespace:

    cdc

Schedule:

    */5 * * * *

The probe checks:

- Kafka Connect REST API availability.
- Required connector existence.
- Connector state.
- Connector task state.

Required connectors:

- `order-db-orders-connector`
- `order-db-orders-dynamic-filter-connector`

The probe fails if:

- A connector is missing.
- A connector is not `RUNNING`.
- A connector has no task.
- Any connector task is not `RUNNING`.

### 3.2 Saga Pending Order Health Probe

A new CronJob was added:

    saga-pending-order-health-probe

Namespace:

    db

Schedule:

    */5 * * * *

The probe checks whether any order has been stuck in `PENDING` for more than 10 minutes.

Healthy result:

    stale_pending_orders=0

This directly targets the Saga failure mode observed after the CDC incident.

### 3.3 Incident-Driven Prometheus Rules

A new PrometheusRule was added:

    phase7-incident-driven-alert-rules

Namespace:

    monitoring

It defines seven alert rules:

- `KafkaConnectDeploymentUnavailable`
- `KafkaConnectConnectorHealthProbeFailed`
- `KafkaConnectBootstrapJobFailed`
- `KafkaConnectCronJobSuspended`
- `SagaPendingOrderProbeFailed`
- `SagaRuntimeDeploymentUnavailable`
- `SagaRuntimePodRestarting`

## 4. Verification Result

The manifests passed client-side dry run and were applied successfully.

Created resources:

    cronjob.batch/kafka-connect-connector-health-probe
    cronjob.batch/saga-pending-order-health-probe
    prometheusrule.monitoring.coreos.com/phase7-incident-driven-alert-rules

### 4.1 Kafka Connect Probe Verification

Manual Job:

    kafka-connect-connector-health-probe-manual-20260612160509

Result:

    connectors=["order-db-orders-dynamic-filter-connector", "order-db-orders-connector"]
    order-db-orders-connector: connector=RUNNING tasks=['RUNNING']
    order-db-orders-dynamic-filter-connector: connector=RUNNING tasks=['RUNNING']
    OK all required Kafka Connect connectors are RUNNING

### 4.2 Saga Pending Order Probe Verification

Manual Job:

    saga-pending-order-health-probe-manual-20260612160514

Result:

    stale_pending_orders=0

### 4.3 PrometheusRule Verification

The live PrometheusRule contains:

    KafkaConnectBootstrapJobFailed
    KafkaConnectConnectorHealthProbeFailed
    KafkaConnectCronJobSuspended
    KafkaConnectDeploymentUnavailable
    SagaPendingOrderProbeFailed
    SagaRuntimeDeploymentUnavailable
    SagaRuntimePodRestarting

## 5. Git Evidence

Implementation commit:

    383a0ab ops(alerts): add incident-driven health probes and alerts

Evidence files:

    docs/evidence/runs/phase7-connect-probe-kafka-connect-connector-health-probe-manual-20260612160509.txt
    docs/evidence/runs/phase7-saga-pending-probe-saga-pending-order-health-probe-manual-20260612160514.txt

Phase 7 evidence document:

    docs/evidence/phase7-incident-driven-alert-hardening-evidence.md

Phase 7 report:

    docs/report/phase7-incident-driven-alert-hardening-report.md

## 6. Production Readiness Impact

Phase 7 improves operational readiness by adding detection around previously observed incident patterns.

Before Phase 7, the system had recovery proof and runbooks. After Phase 7, the system also has alerting hooks and scheduled probes for those risks.

Improved areas:

- Faster detection of missing or broken Kafka Connect connectors.
- Faster detection of failed connector bootstrap.
- Faster detection of stuck Saga orders.
- Faster detection of Saga runtime deployment issues.
- Better operational evidence for production-readiness reporting.

## 7. Remaining Limitations

Phase 7 adds alert rules and probes, but it does not fully automate remediation.

Remaining limitations:

- PostgreSQL replication slot cleanup is still manual by design.
- Saga runtime restart is still runbook-based.
- Kafka consumer lag alerting can be improved further if Kafka exporter metrics are added.
- Business-level metrics would be cleaner if services expose native Prometheus counters for pending orders and outbox requeue count.

## 8. Recommended Future Improvements

Recommended future improvements:

1. Add service-native Prometheus metrics for order status counts.
2. Add outbox requeue counters.
3. Add Kafka exporter for direct consumer lag metrics.
4. Add Grafana panels for CDC/Saga incident indicators.
5. Add Alertmanager routing and notification channels.
6. Add a controlled alert firing proof in a non-production maintenance window.

## 9. Conclusion

Phase 7 is complete.

The system now includes incident-driven health probes and alert rules for the CDC and Saga failure modes discovered during post-incident stabilization.

The probes passed manual verification and the alert rules were loaded into Kubernetes as a PrometheusRule. This completes the operational hardening step after the CDC/Saga incident.
