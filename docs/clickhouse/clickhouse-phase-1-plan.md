# ClickHouse Phase 1 Plan

## Purpose

Add ClickHouse as an OLAP analytics store.

PostgreSQL remains the OLTP source of truth for transactional order/payment/inventory data.

ClickHouse is used for analytical queries, reporting, and future dashboard workloads.

## Architecture

PostgreSQL
-> Debezium CDC
-> Kafka topic cdc.order_db.public.orders
-> ClickHouse Kafka Engine / consumer
-> Analytics tables

## Scope

This phase deploys ClickHouse only.

Data ingestion from Kafka CDC will be implemented in the next phase after the ClickHouse pod is verified.

## Deployment

- Namespace: analytics
- StatefulSet: clickhouse
- Service: clickhouse
- HTTP port: 8123
- Native port: 9000
- PVC: 10Gi local-path
- Node: vm2-mesh
