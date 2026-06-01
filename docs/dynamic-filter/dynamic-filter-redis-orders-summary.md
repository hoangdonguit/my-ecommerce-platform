# Dynamic Filter Redis Orders Summary

## Purpose

This phase proves that Kafka Connect can dynamically filter Debezium CDC records using an external Redis rule without restarting the connector.

## Components

- Kafka Connect
- Debezium PostgreSQL connector
- kafka-connect-dynamic-filter Redis SMT
- Redis rule store
- Kafka dynamic CDC topic

## Connector

- Connector name: order-db-orders-dynamic-filter-connector
- Slot name: dbz_order_dynamic_filter_slot
- Topic prefix: cdc_dynamic.order_db
- Output topic: cdc_dynamic.order_db.public.orders
- Filter field: status
- Redis key: filter:order-status
- Redis URI: redis://redis.default.svc.cluster.local:6379

## Rule 1: COMPLETED

Redis rule: ["COMPLETED"]

Result:

- PENDING insert events were filtered out.
- COMPLETED update event passed through.

Observed event:

- status = COMPLETED
- __op = u

## Rule 2: PENDING

Redis rule: ["PENDING"]

Result:

- PENDING insert events passed through.
- COMPLETED update events were filtered out.

Observed events:

- status = PENDING
- __op = c

## Conclusion

Dynamic filtering works.

The filter rule was changed in Redis while the Kafka Connect connector remained RUNNING. No connector restart was required for changing the filter rule.
