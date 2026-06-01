# ClickHouse Memory Tuning Summary

## Problem

After ClickHouse CDC ingestion was enabled, a simple aggregate query over the analytics table failed with:

- MEMORY_LIMIT_EXCEEDED
- maximum memory around 614 MiB

The ClickHouse pod was healthy, but the configured memory limit was too small for analytics queries.

## Fix

Increase ClickHouse resources and memory ratio:

- max_server_memory_usage_to_ram_ratio: 0.6 -> 0.75
- requests: 250m / 512Mi -> 500m / 1Gi
- limits: 1000m / 1Gi -> 1500m / 2Gi

## Expected Result

ClickHouse aggregate queries should run successfully again while keeping PostgreSQL as the OLTP source of truth and ClickHouse as the OLAP analytics store.
