# CDC-0 Readiness Audit Summary

## Purpose

This audit checks whether the current PostgreSQL + Kafka platform is ready for Debezium CDC.

## Result

The platform is not ready for Debezium yet because PostgreSQL is currently configured with:

- wal_level = replica

Debezium PostgreSQL connector requires logical decoding, so PostgreSQL must be changed to:

- wal_level = logical

## Existing Good Conditions

- max_replication_slots = 10
- max_wal_senders = 16
- PostgreSQL listen_addresses = *
- No existing replication slots
- Business tables have primary keys
- Kafka is running
- Existing business Kafka topics already use 8 partitions
- No existing Kafka Connect or Debezium deployment was found

## Existing Databases Checked

- order_db
- inventory_db
- payment_db
- notification_db

## Conclusion

CDC can continue only after enabling logical replication on PostgreSQL and restarting PostgreSQL in a controlled way.

The next step is to inspect the PostgreSQL Kubernetes manifests and patch the PostgreSQL configuration to enable wal_level=logical.
