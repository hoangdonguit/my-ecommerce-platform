# Phase 6.1 - ArgoCD Metrics Scrape Evidence

## 1. Purpose

This document records the proof that Prometheus can scrape ArgoCD application-controller metrics through a dedicated ServiceMonitor.

This closes the first advanced metric gap found in the Phase 6.1 advanced metric availability audit.

## 2. Change Added

A new ServiceMonitor was added:

- File: `k8s/monitoring/argocd-application-controller-servicemonitor.yaml`
- Resource: `ServiceMonitor monitoring/argocd-application-controller`
- Target namespace: `argocd`
- Target service: `argocd-metrics`
- Target port: `metrics`
- Prometheus selector label: `release=kube-prometheus-stack`

## 3. Why This Was Needed

Before this change, ArgoCD exposed metrics services in the `argocd` namespace, but Prometheus did not have ArgoCD application metrics such as `argocd_app_info`.

The previous metric audit showed:

- `argocd_app_info`: no series
- `argocd_app_health_status`: no series
- `argocd_app_sync_status`: no series

After adding the ServiceMonitor, Prometheus successfully discovered and scraped the `argocd-metrics` target.

## 4. Verification Result

Prometheus target health:

- Job: `argocd-metrics`
- Health: `up`
- Scrape URL: `http://10.42.2.15:8082/metrics`
- Last error: empty

Prometheus query result:

- Query: `argocd_app_info`
- Series count: `8`

Observed applications:

- `analytics-layer`: Healthy / Synced
- `cdc-layer`: Healthy / Synced
- `ecommerce-infrastructure`: Healthy / Synced
- `ecommerce-platform`: Healthy / Synced
- `security-layer`: Healthy / Synced
- `infrastructure-layer`: Healthy / Synced
- `monitoring-addons`: Healthy / Synced
- `observability-layer`: Healthy / Synced

Alert expression dry check:

- `argocd_app_info{sync_status!="Synced"} == 1`: 0 series
- `argocd_app_info{health_status!="Healthy"} == 1`: 0 series

This is expected because all ArgoCD applications were healthy and synced at verification time.

## 5. Important Metric Note

This ArgoCD setup does not expose separate metrics named:

- `argocd_app_health_status`
- `argocd_app_sync_status`

Instead, health and sync states are labels on `argocd_app_info`.

Correct alert expressions should use:

- `argocd_app_info{sync_status!="Synced"} == 1`
- `argocd_app_info{health_status!="Healthy"} == 1`

## 6. Verdict

Result: PASS.

Prometheus can now scrape ArgoCD application metrics.

The next step is to add official ArgoCD alert rules for:

- ArgoCD application OutOfSync
- ArgoCD application not Healthy
