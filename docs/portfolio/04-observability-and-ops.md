# 04 - Observability and Operations

## Metrics

Prometheus and Grafana are used to observe:

- pod status
- deployment availability
- node health
- CPU and memory usage
- HTTP latency and success rate
- benchmark behavior
- alert conditions

## Logs

Loki and Alloy centralize logs from Kubernetes workloads. Logs are useful for debugging Kafka consumers, Saga failures, retry behavior, and service errors.

## Traces

OpenTelemetry and Tempo provide distributed tracing. They help follow requests across service boundaries and asynchronous processing paths.

## Kafka lag

Kafka lag is checked after benchmark and chaos tests. Lag draining back to 0 is used as evidence that asynchronous consumers have caught up.

## Database pressure

PostgreSQL and PgBouncer are monitored for connection pressure. During the 60-minute soak test, PostgreSQL used about 16/100 connections.

## Operations gap

The system has alerting and runbook foundations, but production-grade alerting still needs stronger Kafka exporter metrics, deeper SLO alerts, and automated incident workflows.
