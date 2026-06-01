# OpenTelemetry Audit Summary

## Current Monitoring Stack

The cluster already has a monitoring namespace with:

- Grafana
- Prometheus
- kube-state-metrics
- node-exporter

Grafana is exposed through NodePort 31000.

## Missing Tracing Components

The cluster does not currently have:

- OpenTelemetry Collector
- Tempo
- Jaeger

## Direction

Add a dedicated observability-layer for tracing.

Planned components:

- OpenTelemetry Collector for receiving OTLP traces from services
- Tempo as the trace backend
- Grafana datasource integration for trace visualization

## Next Step

Create GitOps-managed manifests under:

- k8s/observability

Then add an ArgoCD application:

- observability-layer
