# Istio internal API authorization summary

## Purpose

This checkpoint hardens internal API access using Istio security policies.

Only `web-gateway` is allowed to call internal API workloads labelled:

- `security-tier=internal-api`

Protected internal APIs:

- `order-service`
- `inventory-api`
- `payment-api`
- `notification-api`
- `read-model-service`

## Git commits

- `a2b0858 security: remove committed frontend API key`
- `9f52eb7 security: restrict internal APIs with istio policies`

## Runtime policy

- `PeerAuthentication/internal-api-mtls-strict`: mTLS STRICT for internal API workloads.
- `AuthorizationPolicy/allow-web-gateway-to-internal-apis`: allows only principal `cluster.local/ns/default/sa/web-gateway-sa`.

## Verification

Evidence run:

- `docs/security/runs/istio-internal-api-authz-20260603172630/runtime-state.txt`
- `docs/security/runs/istio-internal-api-authz-20260603172630/injected-curl-pod.txt`
- `docs/security/runs/istio-internal-api-authz-20260603172630/negative-non-gateway-deny.txt`
- `docs/security/runs/istio-internal-api-authz-20260603172630/positive-smoke.txt`
- `docs/security/runs/istio-internal-api-authz-20260603172630/security-authz-smoke-summary.json`
- `docs/security/runs/istio-internal-api-authz-20260603172630/istio-traffic-snapshot.txt`

Expected result:

- Normal user path through dashboard/web-gateway passes.
- Direct call from a non-gateway workload is denied or blocked by Istio mTLS/RBAC.
- ArgoCD apps remain Synced/Healthy.
