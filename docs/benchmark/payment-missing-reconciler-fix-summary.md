# Payment Missing Reconciler Fix Summary

## Problem

During the 20 RPS benchmark after Order Outbox Reliability, some orders stayed in PENDING.

The stuck orders had this state:

- Order status: PENDING
- Inventory reservation: RESERVED
- Inventory outbox: PUBLISHED
- Payment row: missing
- Payment outbox row: missing
- Kafka lag: 0

This means the pipeline had no remaining backlog, but the payment stage was never created for those orders.

## Root Cause

The payment consumer can fail after inventory.reserved is published. If the message is sent to DLQ or committed after retry failure, the main Kafka topic no longer has lag.

The existing payment-status-reconciler only handled stale COD PROCESSING payments. It did not handle the case where inventory was already RESERVED but the payment row was missing.

## Fix

The payment-status-reconciler was extended with a second phase:

- Find stale PENDING COD orders.
- Check inventory_db for RESERVED reservation.
- Check payment_db to confirm payment does not exist.
- Insert a COMPLETED payment row with a reconcile transaction id.
- Let the existing payment trigger create payment_outbox_events.
- Let the payment outbox publisher emit payment.completed.
- Let order-service update the order to COMPLETED.

## Deployment

Git revision:

- e9a21bc

CronJob:

- payment-status-reconciler
- namespace: db
- schedule: */1 * * * *

## Verification

Manual job result:

- No stale COD PROCESSING payments found.
- No stale PENDING orders need missing-payment repair.

Final state:

- Orders COMPLETED: 10419
- Orders FAILED: 4
- Orders PENDING: 0
- Payments COMPLETED: 10419
- Payments FAILED: 2
- Payment outbox open rows: 0

## Conclusion

The system now has a self-healing path for this case:

Inventory RESERVED but Payment row missing.

This complements the previous Order Outbox Reliability fix.
