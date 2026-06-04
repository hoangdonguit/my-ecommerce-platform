# Payment Outbox Runtime Proof

## Scope

This document records the Phase 4 runtime proof for the payment outbox flow in `my-ecommerce-platform`.

The goal is not to claim that the payment-service no longer uses Kafka.
The correct claim is:

- `payment-api` does not directly publish terminal payment events in the request path.
- Terminal payment events are stored in `payment_outbox_events`.
- `payment-consumer` / payment outbox publisher publishes those events to Kafka.
- Runtime E2E confirms the flow completes without outbox backlog.

No API key, password, token, private key, or raw secret value is stored in this document.

## Test Environment

- Repo path: `~/Doanchuyennganh/my-ecommerce-platform`
- Default runtime URL: `http://100.65.255.2:30517`
- Test path: dashboard NodePort -> web-gateway -> internal services
- API key source: Kubernetes Secret `ecommerce-runtime-secrets`, key `WEB_GATEWAY_API_KEY`
- API key value: hidden, not printed
- API key length observed: 64

## Preflight

Gateway health:

    GET http://100.65.255.2:30517/api/health
    HTTP 200
    Response: web-gateway is running

Services health:

    GET http://100.65.255.2:30517/api/health/services
    HTTP 200
    inventory_service: ok=true
    notification_service: ok=true
    order_service: ok=true
    payment_service: ok=true
    read_model_service: ok=true

Inventory check:

    product_id: prod-123
    on_hand_quantity: 1000000
    reserved_quantity: 18517
    available_quantity: 981483

## Baseline Before E2E

Before the E2E run, `payment_outbox_events` had no pending backlog:

    status    count
    PUBLISHED 11356

No `PENDING`, `PROCESSING`, or `FAILED` status was observed.

## E2E Smoke Run

Run tag:

    phase4-payment-outbox-proof-20260605022838

Created order:

    order_id: e39b1d3c-dd3b-4057-b49f-e28146566d0e
    user_id: phase4-payment-outbox-proof-20260605022838-user
    idempotency_key: smoke-price-20260605022838
    product_id: prod-123
    quantity: 1
    payment_method: COD
    total_amount: 24000000

Create order result:

    HTTP 201
    initial order status: PENDING

Saga wait result:

    order.status=PENDING at elapsed=0s
    order.status=COMPLETED at elapsed=6s

Database result:

    orders.status: COMPLETED
    order outbox.status: PUBLISHED
    inventory_reservations.status: RESERVED
    payments.status: COMPLETED
    notifications.status: SENT

Smoke verdict:

    SMOKE TEST PASSED

## Payment Outbox After E2E

After the E2E run, `payment_outbox_events` increased by one event:

    status    count
    PUBLISHED 11357

No `PENDING`, `PROCESSING`, or `FAILED` status was observed.

Newest payment outbox event:

    id: 05854aed-9070-482a-ba02-b34732cc0207
    aggregate_id: 9f0ce99a-d534-4043-b097-6df9f6725d52
    event_type: payment.completed
    topic: payment.completed
    status: PUBLISHED
    attempts: 1
    created_at: 2026-06-04 19:28:40.049973
    updated_at: 2026-06-04 19:28:40.214562
    published_at: 2026-06-04 19:28:40.214562

This event corresponds to the E2E payment row:

    payment_id: 9f0ce99a-d534-4043-b097-6df9f6725d52
    order_id: e39b1d3c-dd3b-4057-b49f-e28146566d0e
    payment.status: COMPLETED

## Kafka Lag After E2E

Kafka consumer group lag after the E2E run was observed as 0 for the important groups involved in the flow:

    payment-service-group: lag 0 on inventory.reserved
    order-service-saga-monitor: lag 0 on payment.completed / payment.failed / inventory.failed
    inventory-service-group: lag 0 on order.created
    notification-service-group: lag 0 on payment.completed
    read-model-service-group: lag 0 on payment.completed
    clickhouse-orders-flat-cdc-v2: lag 0 on cdc_flat.order_db.public.orders

Some partitions with log-end-offset 0 show `-` instead of numeric lag. These represent empty partitions and are not treated as backlog.

## Payment Consumer Log

The payment consumer logged the expected outbox publish event after the E2E run:

    2026/06/04 19:28:40 payment outbox batch published count=1 first_aggregate_id=9f0ce99a-d534-4043-b097-6df9f6725d52 last_aggregate_id=9f0ce99a-d534-4043-b097-6df9f6725d52

## Verdict

PASS.

The runtime proof confirms that a new E2E order completed successfully and produced a new `payment.completed` event through `payment_outbox_events`.

Evidence summary:

- Gateway and downstream services were healthy.
- A new order was created through `http://100.65.255.2:30517`.
- Saga completed successfully.
- Payment row became `COMPLETED`.
- New `payment_outbox_events` row was created and marked `PUBLISHED`.
- No payment outbox backlog remained.
- Kafka consumer lag was 0 for the relevant groups.
- Payment consumer log confirmed one outbox batch was published.

Correct final wording:

    payment-api does not directly publish terminal payment events in the request path;
    terminal payment events are persisted in payment_outbox_events and published by payment-consumer/outbox publisher.

Incorrect wording to avoid:

    payment-service does not use Kafka anymore.
