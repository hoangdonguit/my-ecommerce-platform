# Dynamic Filter Kafka Connect Image Build Summary

## Purpose

Build a custom Kafka Connect image based on Debezium Connect 2.7 with kafka-connect-dynamic-filter Redis SMT installed.

## Image

hoangdonguit/kafka-connect-debezium:dynamic-filter-redis-20260601022704

## Result

- Docker build: success
- Local image inspect: success
- Docker push: success
- Remote manifest inspect: success

## Included Plugin

- kafka-connect-dynamic-filter-redis
- Main SMT class: io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter

## Notes

Raw docker build/push logs are not committed to avoid noisy repository history.
