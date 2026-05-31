# Dynamic Filter Audit Summary

## Purpose

Evaluate kafka-connect-dynamic-filter for adding dynamic CDC filtering to the existing Debezium/Kafka Connect pipeline.

## Current System

Kafka Connect currently runs two Debezium PostgreSQL connectors:

- order-db-orders-connector
- order-db-orders-flat-connector

The existing Kafka Connect image does not include kafka-connect-dynamic-filter yet.

## Plugin Finding

The plugin provides a Kafka Connect SMT for filtering Debezium CDC records with dynamic rules.

It supports multiple rule sources:

- Redis
- Kafka topic
- JSON file
- Custom source through the core module

## Important Classes

- io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter
- io.kafkaconnect.dynamicfilter.kafka.KafkaDynamicFilter
- io.kafkaconnect.dynamicfilter.file.FileDynamicFilter
- io.kafkaconnect.dynamicfilter.DynamicListFilter

## Selected Direction

Use RedisDynamicFilter first.

Reasons:

- Redis already exists in the current platform.
- Rules can be changed with redis-cli without restarting Kafka Connect.
- It is easy to demonstrate dynamic rule changes during presentation.
- Existing CDC connectors will not be modified.

## Planned Demo

Create a new independent Debezium connector:

- order-db-orders-dynamic-filter-connector

Use a new topic prefix:

- cdc_dynamic.order_db

Filter field:

- status

Redis key:

- filter:order-status

Demo rule changes:

- ["COMPLETED"] passes only completed order events
- ["PENDING"] passes only pending order events

This proves dynamic filtering without restarting the connector.
