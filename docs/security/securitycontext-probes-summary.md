# SecurityContext and health probes summary

## Purpose

This checkpoint hardens API workloads by running app containers as non-root users and adding Kubernetes health probes.

## Commit

- `18d4cf0 security: harden API workloads with health probes`

## Hardened workloads

- `web-gateway`
- `order-service`
- `inventory-api`
- `payment-api`
- `notification-api`
- `read-model-service`

## Security settings

Each API workload now uses:

- `runAsNonRoot: true`
- `runAsUser: 65532`
- `runAsGroup: 65532`
- `allowPrivilegeEscalation: false`
- `capabilities.drop: ["ALL"]`
- `seccompProfile: RuntimeDefault`

## Health probes

- `web-gateway`: `/api/health`
- Internal APIs: `/api/v1/health`

Each target API workload has:

- `startupProbe`
- `readinessProbe`
- `livenessProbe`

## Evidence

- `docs/security/runs/securitycontext-probes-20260603180937/runtime-state.txt`
- `docs/security/runs/securitycontext-probes-20260603180937/live-securitycontext-probes.txt`
- `docs/security/runs/securitycontext-probes-20260603180937/non-root-user-check.txt`
- `docs/security/runs/securitycontext-probes-20260603180937/smoke-output.txt`
- `docs/security/runs/securitycontext-probes-20260603180937/smoke-summary.json`
