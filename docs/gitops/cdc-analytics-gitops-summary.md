# CDC and Analytics GitOps Management Summary

## Problem

Kafka Connect and ClickHouse manifests were stored in the repository, but they were not managed by the existing ArgoCD ecommerce-platform application.

The ecommerce-platform application only syncs:

- k8s/services

Therefore changes under these folders were not applied automatically:

- k8s/cdc
- k8s/analytics

## Fix

Two dedicated ArgoCD applications are added:

- cdc-layer
  - path: k8s/cdc
  - manages Kafka Connect / Debezium CDC resources

- analytics-layer
  - path: k8s/analytics
  - manages ClickHouse analytics resources

## Reason

This keeps the GitOps structure clear:

- ecommerce-platform: application services
- cdc-layer: Kafka Connect / Debezium
- analytics-layer: ClickHouse OLAP stack

## Safety

Prune is disabled initially to avoid deleting existing manually-created resources during adoption.

Self-heal is enabled so future drift can be corrected by ArgoCD.
