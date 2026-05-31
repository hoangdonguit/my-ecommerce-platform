# CDC-1 Kafka Connect + Debezium Plan

## Purpose

Deploy Kafka Connect with Debezium PostgreSQL connector to capture PostgreSQL changes and publish them to Kafka topics.

## Scope

This phase only deploys Kafka Connect infrastructure.

It does not replace the existing Saga / Transactional Outbox pipeline.

## Target

Initial CDC target:

- order_db.public.orders

## Architecture

PostgreSQL WAL
-> Debezium PostgreSQL connector
-> Kafka Connect
-> Kafka CDC topics

## Safety

- PostgreSQL is already configured with wal_level=logical.
- Kafka Connect runs as a separate deployment in namespace cdc.
- Existing business services remain unchanged.
