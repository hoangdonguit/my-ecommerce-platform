# Redis Password Fallback Hardening

## Scope

This document records the security hardening for removing a hardcoded Redis password fallback from `order-service`.

The goal is to remove source-level hardcoded fallback secret text while preserving the current runtime behavior.

## Finding

Before the fix, `order-service` had this fallback value in source code:

    RedisPassword: getEnv("REDIS_PASSWORD", "redissecret")

This was considered a hardcoded fallback secret / code smell.

## Runtime Context Before Fix

Runtime checks showed:

    REDIS_ADDR_PRESENT=yes
    REDIS_PASSWORD_PRESENT=

Redis deployment currently does not configure `requirepass`.

Redis no-auth check:

    redis-cli ping
    PONG

Therefore, the current Redis runtime is no-auth, and `order-service` does not require `REDIS_PASSWORD` to be set.

## Fix

The fallback was changed to an empty string:

    RedisPassword: getEnv("REDIS_PASSWORD", "")

This aligns `order-service` with the existing `web-gateway` Redis password pattern.

## Build and Deployment

New image:

    hoangdonguit/order-service:redis-fallback-hardening-20260606092113

Docker build and push completed successfully.

Git commit:

    ec1241a security: remove order redis password fallback

ArgoCD state after sync:

    ecommerce-platform Synced / Healthy
    revision ec1241ae995c975ff46a6aa24d5ee629c1cb18e8

Runtime image after sync:

    hoangdonguit/order-service:redis-fallback-hardening-20260606092113

Runtime pods:

    order-service-8549fffd6b-b4p2s Running 2/2 restart=0
    order-service-8549fffd6b-zw2m2 Running 2/2 restart=0

Runtime env after sync:

    REDIS_ADDR_PRESENT=yes
    REDIS_PASSWORD_PRESENT=

## Runtime Smoke Test

Smoke run:

    redis-fallback-hardening-runtime-smoke-20260606094830

Test URL:

    http://100.65.255.2:30517

Created order:

    order_id: 366519ec-6687-4563-8786-5dc1382caa25
    user_id: redis-fallback-hardening-runtime-smoke-20260606094830-user
    idempotency_key: smoke-price-20260606094830
    payment_method: COD
    total_amount: 24000000

Saga result:

    order.status=PENDING at elapsed=1s
    order.status=COMPLETED at elapsed=6s

Database result:

    orders.status: COMPLETED
    order outbox.status: PUBLISHED
    inventory_reservations.status: RESERVED
    payments.status: COMPLETED
    notifications.status: SENT

Smoke verdict:

    SMOKE TEST PASSED

## Verdict

PASS.

The hardcoded Redis password fallback was removed from `order-service`, the new image was deployed by ArgoCD, and the runtime Saga smoke test passed on the new image.

## Notes

This change does not enable Redis authentication.

Future Redis authentication hardening should be handled as a separate task:

1. Enable Redis auth.
2. Store Redis password in Kubernetes Secret.
3. Inject `REDIS_PASSWORD` into services that need Redis.
4. Run smoke/E2E and security evidence again.
