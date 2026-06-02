# Payment outbox-only deployment summary

## Purpose

This checkpoint deploys the payment service in outbox-only mode. Payment terminal events are no longer published directly by the business service. Instead, `payment.completed` / `payment.failed` are emitted through `payment_outbox_events`.

## Commit and image

- Code commit: `f84e156 refactor(payment): use outbox-only event publishing`
- Image: `hoangdonguit/payment-service:outbox-only-20260602012451`
- Deployments:
  - `payment-api`
  - `payment-consumer`

## Smoke test result

Smoke test was executed through web-gateway port-forward because NodePort access from the local machine timed out, while in-cluster service DNS was healthy.

- Gateway URL used: `http://localhost:8090`
- Smoke order ID: `b515d689-cfc8-4055-a1b8-ea603f8d64be`
- Result:
  - Order status: `COMPLETED`
  - Order outbox status: `PUBLISHED`
  - Inventory reservation status: `RESERVED`
  - Payment status: `COMPLETED`
  - Notification status: `SENT`

## Runtime notes

- Kafka consumer lag was checked after the smoke test.
- Payment outbox records were checked directly in `payment_db.payment_outbox_events`.
- Secrets were read from Kubernetes Secret and were not printed.

## Evidence files

- `docs/benchmark/runs/payment-outbox-only-20260602221530/git-image-state.txt`
- `docs/benchmark/runs/payment-outbox-only-20260602221530/kafka-lag-after.txt`
- `docs/benchmark/runs/payment-outbox-only-20260602221530/db-smoke-result.txt`
- `docs/benchmark/runs/payment-outbox-only-20260602221530/order-id.txt`
