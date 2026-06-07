# Phase 6.1 - Alerting Foundation Evidence

## 1. Purpose

This document records the first alerting foundation proof for the `my-ecommerce-platform` project.

The goal of this step is to prove that the system now has a working PrometheusRule-based alerting foundation after Phase 5 capacity and observability hardening.

This evidence does not claim that the system already has complete enterprise alerting. It only confirms that the first Kubernetes-level alert rules were created, accepted by the Prometheus Operator, and loaded by Prometheus.

## 2. Files Added

The following files were added:

- `docs/observability/phase6-alerting-plan.md`
- `docs/observability/phase6-alerting-foundation-evidence.md`
- `k8s/monitoring/ecommerce-platform-alert-rules.yaml`

## 3. PrometheusRule Object

The custom PrometheusRule object was created in the `monitoring` namespace:

- Name: `ecommerce-platform-alert-rules`
- Namespace: `monitoring`
- Label: `release=kube-prometheus-stack`

The label is required because the current Prometheus instance selects rule objects using:

- `ruleSelector.matchLabels.release = kube-prometheus-stack`

## 4. Operator Validation

The PrometheusRule object was accepted by the Prometheus Operator.

Observed annotation:

- `prometheus-operator-validated: "true"`

This means the operator validated the custom rule object instead of rejecting it.

## 5. Rules Loaded by Prometheus

Prometheus successfully loaded the custom rule group:

- `ecommerce-platform.phase6.foundation`

Prometheus also exposed the custom rule file under its generated rulefiles directory.

The following alert rules were visible from the Prometheus rules API:

- `ECommercePodPendingTooLong`
- `ECommercePodCrashLooping`
- `ECommerceContainerRestarting`
- `ECommerceDeploymentUnavailable`
- `ECommerceNodePressure`
- `ECommerceTempoUnavailable`

This confirms that the rules were not only applied to Kubernetes, but also loaded into Prometheus runtime.

## 6. Initial Alert Scope

The first alerting foundation covers:

- Pods stuck in Pending.
- Containers in CrashLoopBackOff.
- Non-Istio application container restarts.
- Deployments with unavailable replicas.
- Kubernetes node pressure conditions.
- Tempo deployment availability.

The first rule set intentionally uses common Kubernetes and kube-state-metrics metrics only.

## 7. Out of Scope

The following alerts were intentionally not added in this step:

- Kafka consumer lag alerts.
- PgBouncer pool saturation alerts.
- PostgreSQL exporter alerts.
- ArgoCD application health alerts.
- Business SLO alerts.
- Alertmanager external notification routing.

Reason: these require exporter and scrape validation first. Adding rules for metrics that do not exist in Prometheus would create YAML-only alerting without real operational value.

## 8. Verification Commands Used

Key verification commands:

- `kubectl apply --dry-run=server -f k8s/monitoring/ecommerce-platform-alert-rules.yaml`
- `kubectl apply -f k8s/monitoring/ecommerce-platform-alert-rules.yaml`
- `kubectl -n monitoring get prometheusrule ecommerce-platform-alert-rules -o yaml`
- `kubectl -n monitoring get prometheus kube-prometheus-stack-prometheus -o yaml`
- `curl http://127.0.0.1:19090/api/v1/rules`

## 9. Result

Result: PASS.

The alerting foundation is now active at the Prometheus rule level.

The next steps should be:

1. Commit the alerting plan, evidence document, and PrometheusRule manifest.
2. Let ArgoCD reconcile the monitoring layer from Git.
3. Verify that the monitoring application remains Synced and Healthy.
4. Later, validate additional exporters before adding Kafka, PgBouncer, PostgreSQL, and ArgoCD alerts.
