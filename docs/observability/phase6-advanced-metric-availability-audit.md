# Phase 6.1 - Advanced Metric Availability Audit

## 1. Purpose

This document records the Prometheus metric availability audit for advanced alerting in the `my-ecommerce-platform` project.

The goal is to avoid creating YAML-only alert rules for metrics that are not currently scraped or observable by Prometheus.

This audit was performed after the Phase 6.1 alerting foundation was completed and GitOps reconciliation was confirmed.

## 2. Audit Context

- Audit time: `2026-06-07 18:49:00 UTC`
- Local time: `2026-06-08 01:49:00 +07`
- Git HEAD at audit time: `ccc55f8`
- Prometheus access method: local `kubectl port-forward` to `svc/kube-prometheus-stack-prometheus` in namespace `monitoring`
- Raw query output: stored under `.local-notes/audit/phase6-advanced-metric-audit-20260608014853`
- Raw query output is intentionally not committed.

## 3. Important Interpretation Rule

A query returning zero time series does not always mean the metric is invalid.

There are two cases:

1. Sparse state metrics:
   - Some Kubernetes state metrics may only appear when a specific state exists.
   - Example: a CrashLoopBackOff or waiting-reason metric may have no active series when the cluster is healthy.
   - These metrics can still be useful if paired with restart-based alerts.

2. Exporter-dependent metrics:
   - Metrics such as ArgoCD, Kafka, PostgreSQL, and PgBouncer metrics require exporters or ServiceMonitor/PodMonitor scraping.
   - If all common metrics in one subsystem return zero series, that subsystem should be treated as not yet available for Prometheus alerting.

Therefore, this audit separates Kubernetes foundation metrics from advanced exporter-dependent metrics.

## 4. Metric Availability Result

| Category | Metric | Purpose | Query Result | Series Count | Decision |
|---|---|---|---:|---:|---|
| Kubernetes | `kube_pod_status_phase` | Core pod phase metric for Pending alert | YES | 290 | Safe for foundation alerts |
| Kubernetes | `kube_pod_container_status_waiting_reason` | Waiting reason metric for CrashLoopBackOff alert | NO | 0 | Keep with caution; likely sparse when no pod is waiting |
| Kubernetes | `kube_pod_container_status_restarts_total` | Core restart counter for restart alert | YES | 61 | Safe for foundation alerts |
| Kubernetes | `kube_deployment_status_replicas_unavailable` | Deployment availability metric | YES | 36 | Safe for foundation alerts |
| Kubernetes | `kube_node_status_condition` | Node pressure condition metric | YES | 45 | Safe for foundation alerts |
| cAdvisor | `container_memory_working_set_bytes` | Container memory usage metric | YES | 168 | Safe for resource dashboards or future alerts |
| cAdvisor | `container_cpu_usage_seconds_total` | Container CPU usage metric | YES | 168 | Safe for resource dashboards or future alerts |
| ArgoCD | `argocd_app_info` | ArgoCD application info metric | NO | 0 | Do not write official ArgoCD alerts yet |
| ArgoCD | `argocd_app_health_status` | ArgoCD application health status metric | NO | 0 | Do not write official ArgoCD alerts yet |
| ArgoCD | `argocd_app_sync_status` | ArgoCD application sync status metric | NO | 0 | Do not write official ArgoCD alerts yet |
| Kafka | `kafka_consumergroup_lag` | Kafka consumer group lag metric | NO | 0 | Do not write official Kafka lag alerts yet |
| Kafka | `kafka_consumer_lag` | Kafka consumer lag metric | NO | 0 | Do not write official Kafka lag alerts yet |
| Kafka | `kafka_exporter_kafka_consumergroup_lag` | Kafka exporter consumer group lag metric | NO | 0 | Do not write official Kafka lag alerts yet |
| Kafka | `kafka_consumergroup_current_offset` | Kafka consumer group current offset metric | NO | 0 | Do not write official Kafka offset alerts yet |
| Kafka | `kafka_topic_partition_current_offset` | Kafka topic partition current offset metric | NO | 0 | Do not write official Kafka offset alerts yet |
| PostgreSQL | `pg_up` | PostgreSQL exporter up metric | NO | 0 | PostgreSQL exporter is not verified yet |
| PostgreSQL | `pg_stat_activity_count` | PostgreSQL active connection metric | NO | 0 | Do not write official PostgreSQL alerts yet |
| PostgreSQL | `pg_stat_database_numbackends` | PostgreSQL database backend count metric | NO | 0 | Do not write official PostgreSQL alerts yet |
| PgBouncer | `pgbouncer_up` | PgBouncer exporter up metric | NO | 0 | PgBouncer exporter is not verified yet |
| PgBouncer | `pgbouncer_pools_cl_active` | PgBouncer active client pool metric | NO | 0 | Do not write official PgBouncer pool alerts yet |
| PgBouncer | `pgbouncer_pools_cl_waiting` | PgBouncer waiting client pool metric | NO | 0 | Do not write official PgBouncer pool alerts yet |

## 5. Alerting Decision

The current Phase 6.1 foundation alerts remain valid because they rely mainly on Kubernetes and cAdvisor metrics that are already visible in Prometheus.

The CrashLoopBackOff alert should be kept, but it should be interpreted together with the restart alert:

- `ECommercePodCrashLooping` catches explicit CrashLoopBackOff state when the series exists.
- `ECommerceContainerRestarting` catches container restarts even when waiting-reason series are not currently visible.

Advanced alerts must not be added yet for these areas:

- ArgoCD application health and sync status.
- Kafka consumer lag.
- PostgreSQL connection and database health.
- PgBouncer pool saturation.

Reason: their Prometheus metrics are not currently available.

## 6. Recommended Next Steps

Recommended order:

1. Keep the existing Phase 6.1 foundation rule set.
2. Commit this metric availability audit as evidence.
3. Inspect ArgoCD metrics exposure first, because ArgoCD already exists and may only need ServiceMonitor or scrape configuration.
4. After that, add Kafka exporter or JMX exporter before writing Kafka lag alerts.
5. Add PostgreSQL exporter and PgBouncer exporter before writing DB and pool saturation alerts.
6. Re-run this audit after each exporter or scrape change.
7. Only then add official advanced PrometheusRule alerts.

## 7. Verdict

Result: PASS for metric audit.

The audit confirms that the foundation alerting layer is correctly limited to currently available Kubernetes-level metrics.

The audit also confirms that advanced alerts for ArgoCD, Kafka, PostgreSQL, and PgBouncer must wait until their metrics are exposed and scraped by Prometheus.
