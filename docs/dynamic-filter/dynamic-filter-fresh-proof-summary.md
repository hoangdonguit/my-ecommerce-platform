# Dynamic Filter fresh proof summary

## Purpose

This checkpoint validates the Redis-driven Kafka Connect Dynamic Filter for order CDC events.

The runtime rule is:

```json
["PENDING"]
```

With this rule, the dynamic CDC topic should keep only orders whose `status` is `PENDING`.

## Runtime components

- Kafka Connect image: `hoangdonguit/kafka-connect-debezium:dynamic-filter-redis-20260601022704`
- Connector: `order-db-orders-dynamic-filter-connector`
- Redis key: `filter:order-status`
- Dynamic topic: `cdc_dynamic.order_db.public.orders`
- Flat comparison topic: `cdc_flat.order_db.public.orders`

## Fresh proof result

- Order ID: `c511889f-98dd-427e-a087-d35150dbad7a`
- Idempotency key: `dynamic-filter-proof-20260602230529`
- Redis rule: `["PENDING"]`
- Dynamic topic contains PENDING: `1`
- Dynamic topic contains COMPLETED: `0`
- Flat topic contains COMPLETED: `1`

Result:

```text
VALIDATION_PASSED: dynamic topic contains PENDING and blocks COMPLETED
```

## Evidence files

- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/run-info.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/gateway-health.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/redis-rule.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/connector-status-before.json`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/create-order-request.json`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/create-order-http-code.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/create-order-response.json`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/create-order-response.pretty.json`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/order-id.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/saga-wait.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/dynamic-topic-matches.jsonl`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/flat-topic-comparison.jsonl`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/db-order-status.txt`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/connector-status-after.json`
- `docs/dynamic-filter/runs/dynamic-filter-fresh-proof-20260602230529/validation.txt`
