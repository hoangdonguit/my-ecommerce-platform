# CDC-1 Orders Connector Summary

## Purpose

This phase proves that PostgreSQL changes can be captured by Debezium and published to Kafka through Kafka Connect.

## Components

- PostgreSQL logical replication
- Kafka Connect
- Debezium PostgreSQL connector
- Kafka CDC topic

## Target Table

- Database: order_db
- Schema: public
- Table: orders

## PostgreSQL Publication

- Publication: dbz_order_publication
- Table: public.orders

## Debezium Connector

- Connector name: order-db-orders-connector
- Slot name: dbz_order_slot
- Topic prefix: cdc.order_db
- CDC topic: cdc.order_db.public.orders
- Snapshot mode: initial

## Result

The connector was created successfully and reached RUNNING state.

The Kafka topic was created:

- cdc.order_db.public.orders

The PostgreSQL replication slot was active:

- dbz_order_slot
- plugin: pgoutput
- database: order_db
- active: true

## Live CDC Proof

A new order was created through the Web Gateway.

Debezium captured two realtime non-snapshot events:

1. Insert event:
   - op = c
   - snapshot = false
   - status = PENDING

2. Update event:
   - op = u
   - snapshot = false
   - status = COMPLETED

This proves the flow:

Web Gateway -> PostgreSQL -> Debezium -> Kafka CDC topic.

## Notes

The existing Saga / Transactional Outbox pipeline was not replaced.

Debezium CDC is currently used as an additional data-change stream for audit, analytics, and future ClickHouse integration.
