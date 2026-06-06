# PgBouncer Runtime Stats Audit

## Scope

This document records the Phase 4 PgBouncer runtime stats audit for `my-ecommerce-platform`.

The goal is to prove that PgBouncer runtime pool statistics can be inspected and that a smoke E2E order can complete without visible pool waiting/backlog.

This document does not claim that PgBouncer is fully production-hardened.

## Background

Before this audit, PgBouncer admin/stat commands failed with:

    FATAL: not allowed

The root cause was that the PgBouncer user existed in `userlist.txt`, but was not listed in `admin_users` or `stats_users`.

## Configuration Change

PgBouncer ConfigMap was updated with:

    admin_users = postgres
    stats_users = postgres

A rollout annotation was added to the PgBouncer Deployment because the config file is mounted through `subPath`, and the existing pod did not reload the updated ConfigMap automatically.

Runtime config after rollout included:

    auth_type = plain
    auth_file = /etc/pgbouncer/userlist.txt
    admin_users = postgres
    stats_users = postgres
    pool_mode = transaction
    max_client_conn = 5000
    default_pool_size = 80

PgBouncer version:

    PgBouncer 1.25.1

## Baseline Stats Before Smoke

Baseline `SHOW POOLS` returned runtime data for:

    inventory_db
    order_db
    payment_db
    pgbouncer

Baseline pool state:

    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

Baseline `SHOW CLIENTS` showed idle clients for inventory, order, payment, and pgbouncer admin connection.

Baseline `SHOW STATS` returned transaction/query counters for inventory, order, payment, and pgbouncer databases.

## Smoke Test

Smoke run:

    pgbouncer-stats-smoke-20260606124205

Test URL:

    http://100.65.255.2:30517

Created order:

    order_id: 5131a5a5-ae6f-454d-9985-b86fc9f3e751
    user_id: pgbouncer-stats-smoke-20260606124205-user
    idempotency_key: smoke-price-20260606124205
    payment_method: COD
    total_amount: 24000000

Saga result:

    order.status=PENDING at elapsed=1s
    order.status=COMPLETED at elapsed=7s

Database result:

    orders.status: COMPLETED
    order outbox.status: PUBLISHED
    inventory_reservations.status: RESERVED
    payments.status: COMPLETED
    notifications.status: SENT

Smoke verdict:

    SMOKE TEST PASSED

## Stats After Smoke

After smoke, `SHOW POOLS` returned runtime data for:

    inventory_db
    notification_db
    order_db
    payment_db
    pgbouncer

After-smoke pool state:

    cl_waiting = 0
    maxwait = 0
    pool_mode = transaction

After-smoke `SHOW CLIENTS` showed 7 client connections and all had:

    wait = 0
    wait_us = 0

After-smoke `SHOW STATS` returned counters for:

    inventory_db
    notification_db
    order_db
    payment_db
    pgbouncer

This confirms that PgBouncer runtime stats are observable before and after a real E2E smoke flow.

## Kafka Lag After Smoke

Kafka consumer groups after the smoke run had no non-zero numeric lag for the important Saga/read-side groups.

Empty partitions may show `-` instead of a numeric lag. These are not treated as backlog.

## Verdict

PASS WITH WARNING.

Pass:

- PgBouncer runtime admin/stat access is available.
- `SHOW POOLS`, `SHOW CLIENTS`, and `SHOW STATS` work.
- A real E2E order completed successfully after enabling stats access.
- PgBouncer showed no pool waiting after the smoke flow.
- Kafka lag after the smoke flow had no non-zero numeric backlog.

Warning:

- `auth_type=plain` remains a lab-oriented setting.
- Runtime DB URLs currently use `sslmode=disable`.
- `max_client_conn=5000` and `default_pool_size=80` should be reviewed under higher load.
- This audit is not a full production-grade PgBouncer security hardening proof.

## Follow-up

Recommended next steps:

1. Re-check PgBouncer stats during 10/20/50 RPS benchmark.
2. Review whether `default_pool_size=80` is too large for the lab cluster.
3. Consider stronger PgBouncer auth mode if compatible with the PostgreSQL/user setup.
4. Consider TLS/SSL hardening for DB connections in a future production-oriented phase.
