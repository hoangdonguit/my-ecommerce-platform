# Phase 6.2 - Logging Gap Inspection

## 1. Purpose

This document records the initial Phase 6.2 logging inspection result.

The goal was to verify whether the platform already had a centralized logging stack before adding new logging components.

## 2. Current Cluster State

At inspection time, all ArgoCD applications were healthy and synced:

- analytics-layer: Synced / Healthy
- cdc-layer: Synced / Healthy
- ecommerce-infrastructure: Synced / Healthy
- ecommerce-platform: Synced / Healthy
- infrastructure-layer: Synced / Healthy
- monitoring-addons: Synced / Healthy
- observability-layer: Synced / Healthy
- security-layer: Synced / Healthy

No non-running or failed pods were found in the quick bad-pod check.

## 3. Existing Observability Components

The cluster currently has:

- Prometheus
- Grafana
- kube-state-metrics
- node-exporter
- OpenTelemetry Collector
- Tempo

These components cover metrics and traces.

## 4. Missing Logging Components

The inspection did not find any centralized log collection stack.

Missing components:

- Loki
- Promtail
- Grafana Agent / Alloy for log collection
- Fluent Bit
- Fluentd
- Dedicated logging namespace or logging workload

No Loki service was found in the cluster.

## 5. Repository State

The repository currently contains observability and monitoring files for:

- ArgoCD ServiceMonitor
- PrometheusRule alert rules
- Istio monitor
- Istio dashboards
- OTel Collector
- Tempo

The repository does not currently contain Loki, Promtail, Alloy, Fluent Bit, or Fluentd manifests.

## 6. Grafana Datasource State

Grafana currently has these datasources:

- Alertmanager
- Prometheus

Grafana does not currently have a Loki datasource.

## 7. Decision

Phase 6.2 should add a minimal centralized logging stack.

Recommended next implementation:

- Loki in the observability namespace
- Promtail as a DaemonSet on all nodes
- Grafana Loki datasource ConfigMap in the monitoring namespace
- Evidence document proving log ingestion works
- Optional dashboard or saved LogQL examples after ingestion is verified

## 8. Scope Control

This phase should not start chaos testing or heavy benchmark scenarios yet.

Logging should be added before running failure-injection tests so that incidents can be investigated from logs, metrics, traces, and alerts together.

## 9. Verdict

Result: GAP CONFIRMED.

The platform has metrics and traces, but centralized logging is not yet implemented.
