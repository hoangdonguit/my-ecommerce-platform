# OpenTelemetry trace headers schema summary

## Purpose

This checkpoint adds backward-compatible schema support for OpenTelemetry trace context propagation through the transactional outbox flow.

## Schema changes

- `order_db.outbox.headers JSONB NOT NULL DEFAULT '{}'::jsonb`
- `inventory_db.inventory_outbox_events.headers JSONB NOT NULL DEFAULT '{}'::jsonb`
- `payment_db.payments.trace_headers JSONB NOT NULL DEFAULT '{}'::jsonb`
- `payment_db.payment_outbox_events.headers JSONB NOT NULL DEFAULT '{}'::jsonb`

The payment outbox trigger function `enqueue_payment_outbox_event()` was updated to copy `payments.trace_headers` into `payment_outbox_events.headers`.

## Runtime smoke result

- Smoke order ID: `3fb8cdf9-f97e-48a1-87a8-78c99649ca3b`
- Result: Saga completed successfully after applying the schema migration.
- Current code has not yet started writing real trace headers, so new header columns are expected to contain `{}` until the next code checkpoint.

## Evidence files

- `docs/opentelemetry/runs/otel-trace-headers-schema-20260602232803/run-info.txt`
- `docs/opentelemetry/runs/otel-trace-headers-schema-20260602232803/gateway-health.txt`
- `docs/opentelemetry/runs/otel-trace-headers-schema-20260602232803/smoke-output.txt`
- `docs/opentelemetry/runs/otel-trace-headers-schema-20260602232803/order-id.txt`
- `docs/opentelemetry/runs/otel-trace-headers-schema-20260602232803/db-headers-check.txt`
- `docs/opentelemetry/runs/otel-trace-headers-schema-20260602232803/kafka-lag-after.txt`
