# Istio Runtime AuthZ and mTLS Proof

## Scope

This document records the Phase 4 runtime proof for Istio sidecar injection, internal API AuthorizationPolicy enforcement, and STRICT mTLS-targeted workload protection.

The proof focuses on internal API workloads labeled:

    security-tier=internal-api

These workloads include:

    order-service
    inventory-api
    payment-api
    notification-api
    read-model-service

This document does not claim that every workload in the cluster is part of the mesh.

## Runtime Resources

Istio control plane:

    istiod: Running

Istio security resources in namespace `default`:

    PeerAuthentication/internal-api-mtls-strict
    mode: STRICT
    selector:
      security-tier=internal-api

    AuthorizationPolicy/allow-web-gateway-to-internal-apis
    action: ALLOW
    selector:
      security-tier=internal-api
    allowed principal:
      cluster.local/ns/default/sa/web-gateway-sa

DestinationRule:

    order-service-cb
    host: order-service.default.svc.cluster.local

## Injection Verification

A canary pod with explicit injection annotation was created.

Canary result:

    READY: 2/2
    initContainers: istio-init, istio-proxy
    containers: curl
    tlsMode: istio
    istio-proxy image: registry.istio.io/release/proxyv2:1.30.0
    TRUST_DOMAIN: cluster.local
    CA_ADDR: istiod.istio-system.svc:15012

This confirmed that the Istio sidecar injector works.

Note: in this Istio setup, `istio-proxy` appears as a native sidecar under `initContainers`, not under normal app `containers`.

## Existing Workload Sidecar State

The following internal API pods had Istio native sidecars:

    inventory-api:
      initContainers=istio-init,istio-proxy
      tlsMode=istio

    notification-api:
      initContainers=istio-init,istio-proxy
      tlsMode=istio

    order-service:
      initContainers=istio-init,istio-proxy
      tlsMode=istio

    payment-api:
      initContainers=istio-init,istio-proxy
      tlsMode=istio

    read-model-service:
      initContainers=istio-init,istio-proxy
      tlsMode=istio

The web-gateway pods also had Istio native sidecars:

    web-gateway:
      initContainers=istio-init,istio-proxy
      tlsMode=istio

Non-internal infrastructure workloads such as Redis, MongoDB, and consumers were not part of this internal API AuthZ proof.

## Positive Test

A temporary test pod was created with:

    serviceAccountName: web-gateway-sa
    sidecar.istio.io/inject: "true"

The pod had:

    initContainers=istio-init,istio-proxy
    containers=curl
    tlsMode=istio

Expected result:

    web-gateway-sa should be allowed to call internal APIs.

Actual result:

    order-service health:
      HTTP_CODE=200

    inventory-api health:
      HTTP_CODE=200

    payment-api health:
      HTTP_CODE=200

    notification-api health:
      HTTP_CODE=200

    read-model-service health:
      HTTP_CODE=200

Positive test verdict:

    PASS

## Negative Test

A temporary test pod was created with:

    serviceAccountName: default
    sidecar.istio.io/inject: "true"

The pod had:

    initContainers=istio-init,istio-proxy
    containers=curl
    tlsMode=istio

Expected result:

    default serviceAccount should be denied by Istio AuthorizationPolicy.

Actual result:

    order-service health:
      HTTP_CODE=403
      RBAC: access denied

    inventory-api health:
      HTTP_CODE=403
      RBAC: access denied

    payment-api health:
      HTTP_CODE=403
      RBAC: access denied

    notification-api health:
      HTTP_CODE=403
      RBAC: access denied

    read-model-service health:
      HTTP_CODE=403
      RBAC: access denied

Negative test verdict:

    PASS

## Cleanup

Temporary test pods were deleted after the proof:

    istio-authz-allowed
    istio-authz-denied
    istio-injection-canary

## Verdict

PASS.

The runtime proof confirms:

- Istio sidecar injection works.
- Internal API pods are injected with Istio native sidecars.
- Internal API pods have `security.istio.io/tlsMode=istio`.
- `web-gateway-sa` is allowed to call internal APIs.
- `default` serviceAccount is denied with HTTP 403 and `RBAC: access denied`.
- The AuthorizationPolicy is enforced at runtime.

## Notes

This proof validates runtime authorization for the internal API boundary.

It does not claim:

- Every pod in the cluster is meshed.
- Redis, MongoDB, Kafka, and consumer workloads are protected by this exact AuthorizationPolicy.
- The whole cluster is fully zero-trust production-ready.
- Rate limiting, JWT/OIDC, or centralized logging are complete.

## Follow-up

Recommended next steps:

1. Query Prometheus Istio metrics during later benchmarks to confirm `connection_security_policy="mutual_tls"` traffic volume.
2. Re-run this positive/negative AuthZ proof after major service mesh or workload changes.
3. Add rate limiting at the gateway or mesh layer.
4. Add a runbook for diagnosing unexpected Istio 403 errors.
