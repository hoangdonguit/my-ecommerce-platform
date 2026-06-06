# Phase 4 Post-Hardening Smoke Check

## Scope

This document records the Phase 4 post-hardening smoke checkpoint after the following proof/hardening items:

- Redis password fallback hardening
- PgBouncer runtime stats audit
- ClickHouse CDC freshness proof
- Istio runtime AuthZ/mTLS proof
- Grafana/NTP investigation
- Payment outbox runtime proof

The purpose is to confirm that the system still completes the main E2E Saga flow after these changes.

## Environment

Runtime test URL:

    http://100.65.255.2:30517

Raw local evidence:

    .local-notes/audit/phase4-post-hardening-smoke-20260606134501/post-hardening-smoke-check.txt

Git/GitOps state before smoke:

    git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy

Applications checked:

    analytics-layer
    cdc-layer
    ecommerce-infrastructure
    ecommerce-platform
    infrastructure-layer
    monitoring-addons
    observability-layer
    security-layer

All were Synced / Healthy.

## Health Check

Gateway health:

    HTTP/1.1 200 OK
    web-gateway is running

Services health:

    HTTP/1.1 200 OK

Service health result:

    inventory_service: ok=true
    notification_service: ok=true
    order_service: ok=true
    payment_service: ok=true
    read_model_service: ok=true

## Smoke Run

Smoke tag:

    post-hardening-smoke-20260606134502

Created order:

    order_id: 3a78df81-a29b-4315-b2cf-1bf9c0b90d0e
    user_id: post-hardening-smoke-20260606134502-user
    idempotency_key: smoke-price-20260606134502
    payment_method: COD
    total_amount: 24000000

Saga transition:

    order.status=PENDING at elapsed=1s
    order.status=COMPLETED at elapsed=6s

Database result:

    orders.status: COMPLETED
    order outbox.status: PUBLISHED
    inventory_reservations.status: RESERVED
    payments.status: COMPLETED
    notifications.status: SENT

Smoke verdict:

    SMOKE TEST PASSED

## Kafka Lag After Smoke

Kafka consumer group check returned:

    NO_NONZERO_NUMERIC_LAG

This means no non-zero numeric lag was detected for the checked consumer groups after the smoke run.

## PgBouncer State After Smoke

PgBouncer pools checked after the smoke:

    inventory_db
    notification_db
    order_db
    payment_db
    pgbouncer

Pool state:

    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

PgBouncer clients were visible through `SHOW CLIENTS`.

One `order_db` client was active during the stats query, but there was no client waiting backlog:

    wait = 0
    cl_waiting = 0
    maxwait = 0

## Pod State After Smoke

Post-smoke pod check showed the main workloads running:

    clickhouse
    kafka-connect-debezium
    pgbouncer
    postgresql
    ecommerce-dashboard
    inventory-api
    inventory-consumer
    mongodb
    notification-api
    notification-consumer
    order-service
    payment-api
    payment-consumer
    read-model-service
    redis
    web-gateway
    istiod
    kafka
    grafana
    prometheus
    otel-collector
    tempo

Short-lived reconciler jobs were observed as `Succeeded`.

## Verdict

PASS.

The system completed the main E2E order Saga after Phase 4 hardening/proof work.

Evidence summary:

- Gateway health passed.
- Services health passed.
- Order creation returned HTTP 201.
- Saga completed successfully.
- Transactional database states matched the expected result.
- Kafka had no non-zero numeric lag after smoke.
- PgBouncer showed no waiting pool backlog.
- Main pods remained Running.
- Git remained clean and aligned with origin.

## Notes

This is a smoke checkpoint, not a load benchmark.

It proves the main flow still works after hardening, but it does not replace 5/10/20/50 RPS benchmark evidence.

## Next Action

Proceed to a small post-hardening benchmark, starting with 5 RPS or 10 RPS before increasing load.
