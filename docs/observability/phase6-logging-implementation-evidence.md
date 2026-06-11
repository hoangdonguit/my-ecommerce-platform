# Phase 6.2 - Centralized Logging Implementation Evidence

## 1. Purpose

This document records the implementation and verification of centralized logging for the platform.

Phase 6.2 adds log collection and log querying capability to complement the existing metrics, traces, and alerting stack.

## 2. Components Added

The following components were added:

- Loki
- Grafana Alloy
- Grafana Loki datasource

## 3. Manifest Files Added

Observability layer:

- k8s/observability/loki-config.yaml
- k8s/observability/loki.yaml
- k8s/observability/alloy-rbac.yaml
- k8s/observability/alloy-config.yaml
- k8s/observability/alloy.yaml

Monitoring layer:

- k8s/monitoring/loki-datasource-configmap.yaml

## 4. Deployment Model

Loki is deployed in the observability namespace as a single-replica StatefulSet.

Storage:

- StorageClass: local-path
- PVC: loki-data-loki-0
- Size: 5Gi
- Access mode: ReadWriteOnce

Alloy is deployed in the observability namespace as a single-replica Deployment.

Alloy collects Kubernetes pod logs through the Kubernetes API and forwards them to Loki.

## 5. Loki Readiness Verification

Loki readiness endpoint:

- Endpoint: /ready
- Result: HTTP 200 OK
- Body: ready

## 6. Loki Ingestion Verification

Loki labels endpoint returned these labels:

- app
- app_kubernetes_io_name
- container
- instance
- job
- namespace
- pod
- service_name

Loki series query result:

- Query window: recent 15 minutes
- Series count: 44

Observed namespaces included:

- default
- db
- argocd
- observability
- monitoring
- kube-system

## 7. Log Query Verification

Query against default namespace:

- Query: {namespace="default"}
- Stream count: 8
- Sample services observed:
  - mongodb
  - notification-api
  - order-service
  - payment-api

Sample application log type observed:

- GIN health check logs from backend services
- MongoDB WiredTiger checkpoint logs

Query against Alloy self logs:

- Query: {namespace="observability", pod=~"alloy-.+"}
- Stream count: 1
- Sample log: opened log stream events from loki.source.kubernetes

## 8. Loki Runtime Metrics Verification

Loki metrics showed ingestion activity:

- loki_distributor_lines_received_total: 967043
- loki_ingester_memory_streams: 57
- loki_distributor_bytes_received_total: present

This confirms that logs were received by Loki and stored as active streams.

## 9. Grafana Datasource Verification

Grafana datasource API returned:

- Alertmanager
- Loki
- Prometheus

Loki datasource:

- Name: Loki
- Type: loki
- UID: loki
- URL: http://loki.observability.svc.cluster.local:3100
- Health: OK
- Health message: Data source successfully connected.

## 10. Pod and Storage Verification

Observed runtime state:

- pod/alloy: Running
- pod/loki-0: Running
- service/alloy: ClusterIP
- service/loki: ClusterIP
- persistentvolumeclaim/loki-data-loki-0: Bound

## 11. Scope and Limitations

This implementation is suitable for the current advanced lab and pre-production prototype scope.

Current limitations:

- Loki is deployed as a single-replica instance.
- Loki uses local-path storage.
- Retention is configured for short-term lab evidence.
- This is not a production-grade horizontally scalable Loki deployment.
- Production deployment should use object storage, stronger retention policy, and HA architecture.

## 12. Verdict

Result: PASS.

The platform now has centralized logging through Loki and Grafana Alloy.

Phase 6 observability now includes:

- Metrics through Prometheus
- Dashboards through Grafana
- Traces through OpenTelemetry Collector and Tempo
- Alerts through PrometheusRule
- Logs through Loki and Alloy
