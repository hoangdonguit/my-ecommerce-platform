# Full OpenTelemetry saga runtime summary

## Purpose

This checkpoint deploys full OpenTelemetry trace propagation across the Saga flow.

## Runtime image tags

- Full OTel tag: `otel-full-saga-20260603005334`
- Inventory hotfix tag: `otel-full-saga-inventory-hotfix-20260603013722`
- Payment hotfix tag: `otel-full-saga-payment-hotfix-20260603015846`

## Latest smoke result

- Smoke order ID: `25c883a7-1bab-471a-9431-ae7094037801`
- Order completed: `1`
- Trace context persistence:
  - order outbox has traceparent: `1`
  - inventory outbox has traceparent: `1`
  - payment trace_headers has traceparent: `1`
  - payment outbox has traceparent: `1`

## Evidence files

- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/run-info.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/runtime-image-state.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/argocd-state.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/gateway-health.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/smoke-output.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/order-id.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/db-trace-headers-check.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/validation.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/kafka-lag-after.txt`
- `docs/opentelemetry/runs/otel-full-saga-runtime-final-20260603020222/service-logs-after.txt`
