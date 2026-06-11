# Phase 6.1 - ArgoCD Application Alert Rules Evidence

## 1. Purpose

This document records the proof that ArgoCD application alert rules were added to the platform PrometheusRule and loaded by Prometheus.

This extends the Phase 6.1 alerting foundation with GitOps-aware alerting.

## 2. Change Added

The existing PrometheusRule was updated:

- File: k8s/monitoring/ecommerce-platform-alert-rules.yaml
- Resource: PrometheusRule monitoring/ecommerce-platform-alert-rules
- Group: ecommerce-platform.phase6.foundation

Two ArgoCD alert rules were added:

- ECommerceArgoCDAppOutOfSync
- ECommerceArgoCDAppNotHealthy

## 3. Metric Source

The alerts use the ArgoCD application-controller metric:

- argocd_app_info

This metric is available after adding:

- ServiceMonitor monitoring/argocd-application-controller

The metric exposes application state through labels, not through separate gauge metrics.

Important labels:

- name
- namespace
- health_status
- sync_status
- repo
- dest_namespace
- dest_server

## 4. Alert Expressions

OutOfSync alert expression:

    argocd_app_info{
      namespace="argocd",
      sync_status!="Synced"
    } == 1

Not Healthy alert expression:

    argocd_app_info{
      namespace="argocd",
      health_status!="Healthy"
    } == 1

## 5. Verification Result

GitOps revision after the change:

- Commit: e07510b
- Application: monitoring-addons
- Sync status: Synced
- Health status: Healthy
- Source path: k8s/monitoring

Live PrometheusRule verification:

- ECommerceArgoCDAppOutOfSync: present
- ECommerceArgoCDAppNotHealthy: present
- prometheus-operator-validated: true

Prometheus API loaded rule names:

- ECommerceArgoCDAppNotHealthy
- ECommerceArgoCDAppOutOfSync
- ECommerceContainerRestarting
- ECommerceDeploymentUnavailable
- ECommerceNodePressure
- ECommercePodCrashLooping
- ECommercePodPendingTooLong
- ECommerceTempoUnavailable

Current expression result:

- argocd_app_info{sync_status!="Synced"} == 1: 0 series
- argocd_app_info{health_status!="Healthy"} == 1: 0 series

This is expected because all ArgoCD applications were Healthy and Synced at verification time.

## 6. Current Covered ArgoCD Failure Modes

These alert rules cover:

- Application drift from Git
- Failed or incomplete synchronization
- Degraded ArgoCD application health
- Failed rollout reflected through ArgoCD health
- GitOps-managed stack degradation

## 7. Verdict

Result: PASS.

The platform now has GitOps-aware alerting through ArgoCD application metrics.

Phase 6.1 alerting now includes:

- Kubernetes scheduling/runtime/availability alerts
- Node pressure alert
- Tempo availability alert
- ArgoCD OutOfSync alert
- ArgoCD NotHealthy alert
