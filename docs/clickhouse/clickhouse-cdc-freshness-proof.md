# ClickHouse CDC Freshness Proof

## Scope

This document records the Phase 4 freshness proof for the CDC pipeline from PostgreSQL orders to ClickHouse analytics storage.

The goal is to prove that a newly created order appears in ClickHouse through the Debezium/Kafka Connect CDC pipeline.

This proof does not claim that ClickHouse is the source of truth. PostgreSQL remains the transactional source of truth.

## Pipeline

The verified pipeline is:

    PostgreSQL order_db.orders
    -> Debezium / Kafka Connect
    -> Kafka topic cdc_flat.order_db.public.orders
    -> ClickHouse analytics.orders_flat_cdc_events

## Precheck

Git and GitOps state before the proof:

    git status: clean
    origin/main...HEAD: 0 0
    ArgoCD applications: Synced / Healthy

CDC runtime state:

    kafka-connect-debezium pod: Running
    service kafka-connect-debezium: ClusterIP 8083

Kafka Connect connector status:

    connector: order-db-orders-connector
    connector state: RUNNING
    task 0 state: RUNNING
    type: source

ClickHouse runtime state:

    clickhouse-0 pod: Running
    service clickhouse: ClusterIP, ports 8123 and 9000

## ClickHouse Tables

ClickHouse databases included:

    analytics
    default
    system

Analytics tables included:

    mv_orders_flat_cdc_events
    orders_flat_cdc_events
    orders_flat_cdc_queue

Target table:

    analytics.orders_flat_cdc_events

Important schema fields:

    ingested_at DateTime64(3)
    kafka_topic String
    kafka_partition Int32
    kafka_offset UInt64
    order_id String
    user_id String
    status LowCardinality(String)
    currency LowCardinality(String)
    payment_method LowCardinality(String)
    total_amount Decimal(18, 2)
    idempotency_key String
    op LowCardinality(String)
    source_ts_ms Int64
    source_table String
    source_db String

## Freshness Result

ClickHouse row count and latest timestamps:

    rows: 22843
    max_ingested_at: 2026-06-06 05:42:11.190
    max_source_ts_ms: 1780724528526

The newest ClickHouse rows included the PgBouncer smoke order:

    order_id: 5131a5a5-ae6f-454d-9985-b86fc9f3e751

Rows observed in ClickHouse:

    ingested_at: 2026-06-06 05:42:11.190
    topic: cdc_flat.order_db.public.orders
    partition: 4
    offset: 2840
    order_id: 5131a5a5-ae6f-454d-9985-b86fc9f3e751
    status: PENDING
    op: c
    source_ts_ms: 1780724526516

    ingested_at: 2026-06-06 05:42:11.190
    topic: cdc_flat.order_db.public.orders
    partition: 4
    offset: 2841
    order_id: 5131a5a5-ae6f-454d-9985-b86fc9f3e751
    status: COMPLETED
    op: u
    source_ts_ms: 1780724528526

## Correlation With Smoke Test

The same order was created by the PgBouncer stats smoke test:

    smoke run: pgbouncer-stats-smoke-20260606124205
    order_id: 5131a5a5-ae6f-454d-9985-b86fc9f3e751
    order status in transactional DB: COMPLETED
    order outbox status: PUBLISHED
    inventory reservation status: RESERVED
    payment status: COMPLETED
    notification status: SENT

ClickHouse captured both the create event and the update event:

    PENDING   op=c
    COMPLETED op=u

This proves that a newly created order was propagated from PostgreSQL through CDC into ClickHouse.

## Verdict

PASS.

Evidence summary:

- Kafka Connect connector `order-db-orders-connector` was RUNNING.
- Connector task 0 was RUNNING.
- ClickHouse analytics tables existed.
- `analytics.orders_flat_cdc_events` had fresh data.
- The new smoke order appeared in ClickHouse.
- Both create and update CDC events were visible.
- The latest `ingested_at` matched the smoke test time window.

## Notes

This proof validates CDC freshness into the ClickHouse event table.

It does not prove:

- ClickHouse is the transactional source of truth.
- A current-state read model table is fully correct.
- All analytics queries are optimized.
- Production-grade retention, compaction, or replay policy is complete.

## Follow-up

Recommended next steps:

1. Add a current-state query or materialized view proof if the report needs a read-model snapshot.
2. Compare PostgreSQL order count with ClickHouse CDC event count by time window.
3. Re-check ClickHouse freshness during higher RPS benchmark.
4. Add runbook steps for Kafka Connect connector failure and ClickHouse ingestion lag.
