# Observability Layer Deploy Summary

## Purpose

Deploy the tracing backend foundation for OpenTelemetry-based distributed tracing.

## Components

- OpenTelemetry Collector
- Grafana Tempo

## Namespace

- observability

## Endpoints

OpenTelemetry Collector:

- OTLP gRPC: otel-collector.observability.svc.cluster.local:4317
- OTLP HTTP: otel-collector.observability.svc.cluster.local:4318
- Metrics: otel-collector.observability.svc.cluster.local:8888

Tempo:

- HTTP: tempo.observability.svc.cluster.local:3200
- OTLP gRPC: tempo.observability.svc.cluster.local:4317
- OTLP HTTP: tempo.observability.svc.cluster.local:4318

## Verification Result

- observability-layer is Synced and Healthy in ArgoCD.
- tempo pod is Running with 0 restarts.
- otel-collector pod is Running with 0 restarts.
- Tempo /ready returns HTTP 200 OK.
- OpenTelemetry Collector /metrics returns HTTP 200 OK.
- No bad pods were found.

## Notes

Tempo was changed from grafana/tempo:3.0.0 to grafana/tempo:2.8.2 to reduce noisy backend scheduler logs in this lab environment.

The current Tempo storage uses emptyDir/local storage, which is acceptable for course demo and testing. Trace data may be lost if the Tempo pod is recreated.
