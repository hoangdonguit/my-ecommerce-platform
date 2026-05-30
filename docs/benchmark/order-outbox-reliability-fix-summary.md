# Order Outbox Reliability Fix Summary

## Problem

During the 20 RPS benchmark after Payment Outbox, the system ended with stuck orders:

- Some orders stayed in PENDING.
- Order outbox rows were already marked PUBLISHED.
- Kafka consumer lag was 0.
- Inventory reservations did not exist for those orders.
- Payments did not exist for those orders.

This means the system had no remaining backlog to process, but the affected orders were still stuck.

## Root Cause

The old Order Outbox worker only had two states:

- PENDING
- PUBLISHED

The worker selected PENDING rows, published a Kafka batch, and marked the whole batch as PUBLISHED.

If some order.created events were not effectively processed downstream, those orders had no retry path because the outbox row was already marked PUBLISHED.

## Fix

The order outbox schema was extended with reliability fields:

- attempts
- last_error
- next_attempt_at
- published_at
- updated_at

The Order Outbox worker was changed to:

- use PROCESSING state before publishing
- retry PENDING / FAILED events
- retry stale PROCESSING events
- use Kafka RequiredAcks=RequireAll
- mark publish failure as FAILED with backoff
- requeue stale PUBLISHED order.created events when the corresponding order is still PENDING after a threshold

## Deployment

New image:

- hoangdonguit/order-service:order-outbox-reliability-20260530202151

Git revision:

- 8109494

## Recovery Result

The new worker detected and requeued the stuck order.created events:

- Requeued stale published orders: 36
- Republished order.created batch: 36

After the pipeline replayed those events:

- Orders COMPLETED: 9232
- Orders FAILED: 4
- Orders PENDING: 0
- Order outbox open rows: 0
- Inventory outbox open rows: 0
- Payment outbox open rows: 0

Final result:

- STUCK_ORDER_RECOVERY_OK

## Conclusion

The Order Outbox reliability fix solved the stuck PENDING order issue.

The system now has a self-healing path for the case where an order.created event was marked PUBLISHED but the order still remained PENDING without downstream inventory/payment state.
