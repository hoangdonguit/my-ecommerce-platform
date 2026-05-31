# PostgreSQL Logical Replication Plan for CDC

## Purpose

Debezium PostgreSQL connector requires logical decoding from PostgreSQL WAL.

The current PostgreSQL setting before CDC is:

- wal_level = replica

CDC requires:

- wal_level = logical

## Change

PostgreSQL is managed by Helm using the Bitnami PostgreSQL chart.

A small override values file is added:

- k8s/db/postgresql-cdc-values.yaml

The override enables:

- wal_level = logical
- max_replication_slots = 10
- max_wal_senders = 16
- max_slot_wal_keep_size = 2GB

It also increases PostgreSQL resources because CDC adds WAL and replication slot overhead.

## Expected Impact

Changing wal_level requires a PostgreSQL restart.

During the restart, application services may temporarily fail database requests, so no benchmark should run during this operation.

## Verification

After Helm upgrade, verify:

- SHOW wal_level; returns logical
- PostgreSQL pod is Ready
- Restart count is acceptable
- Saga tables and outbox tables remain clean
