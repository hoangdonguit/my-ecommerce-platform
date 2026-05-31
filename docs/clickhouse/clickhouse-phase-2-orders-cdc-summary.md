# ClickHouse Phase 2 Orders CDC Ingestion Summary

## Purpose

This phase proves that order changes captured by Debezium CDC can be ingested into ClickHouse for analytics.

## Source Flow

Web Gateway
-> PostgreSQL order_db.public.orders
-> Debezium PostgreSQL connector
-> Kafka flat CDC topic
-> ClickHouse Kafka Engine table
-> ClickHouse Materialized View
-> MergeTree analytics table

## Kafka Topic

- Source topic: cdc_flat.order_db.public.orders

## ClickHouse Tables

- analytics.orders_flat_cdc_queue
  - Engine: Kafka
  - Reads from cdc_flat.order_db.public.orders

- analytics.mv_orders_flat_cdc_events
  - Engine: MaterializedView
  - Parses flat Debezium JSON events

- analytics.orders_flat_cdc_events
  - Engine: MergeTree
  - Stores analytics-ready CDC events

## Live Proof

A new order was created through the Web Gateway.

ClickHouse received two realtime CDC events for the same order:

1. Insert event
   - status = PENDING
   - op = c

2. Update event
   - status = COMPLETED
   - op = u

This proves that ClickHouse can consume order lifecycle changes from Kafka CDC.

## Aggregate Result

The aggregate query showed:

- COMPLETED / r: snapshot records from the initial Debezium snapshot
- FAILED / r: snapshot records from existing failed orders
- PENDING / c: realtime insert events
- COMPLETED / u: realtime update events

## Conclusion

ClickHouse is now integrated as an OLAP analytics store.

PostgreSQL remains the OLTP source of truth, while ClickHouse receives CDC events for reporting, analytics, and future dashboard queries.
