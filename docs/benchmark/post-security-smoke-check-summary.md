# Post-security Smoke Check Summary

## Purpose

This smoke check verifies that the platform still works correctly after the security hardening phase, before running post-security benchmark workloads.

This is not a stress benchmark. It is a safety check to confirm that the security changes did not accidentally break the main demo path, admin access policy, service health, Kafka processing, or database connectivity.

## Commit policy

Commit-safe summary:

- docs/benchmark/post-security-smoke-check-summary.md

Local-only raw evidence:

- .local-notes/benchmark/post-security-smoke-20260603194243/

The local evidence directory is intentionally ignored by .gitignore and must not be committed.

## Operating endpoints checked

| Area | Expected state | Result |
|---|---|---|
| Public demo / k6 entrypoint | http://100.65.255.2:30517 reachable | PASS |
| Direct web-gateway NodePort | 32193 closed | PASS |
| web-gateway Kubernetes Service | ClusterIP only | PASS |
| ArgoCD admin NodePort | 30080 and 30443 protected by allowlist | PASS |
| Grafana admin NodePort | 31000 protected by allowlist | PASS |
| Chaos Dashboard NodePort | 31333 and 30930 protected by allowlist | PASS |
| Kafdrop NodePort | 30090 protected by allowlist | PASS |

## Smoke checklist result

| Check | Result | Notes |
|---|---|---|
| Git working tree before smoke | PASS | Clean |
| ArgoCD applications | PASS | All apps Synced / Healthy |
| Kubernetes nodes | PASS | All nodes Ready |
| Main application pods | PASS | Running / Ready |
| Public dashboard path | PASS | 30517 returned HTTP 200 |
| Direct web-gateway NodePort | PASS | 32193 refused / closed |
| Admin access from allowlisted workstation | PASS | Admin NodePorts reachable |
| Admin access from non-allowlisted VM | PASS | Admin NodePorts blocked / timeout |
| k6 smoke test | PASS | 100% checks succeeded |
| Kafka consumer lag | PASS | Main groups have lag 0 |
| PostgreSQL | PASS | Running, 0 restart |
| PgBouncer | PASS | Running, 0 restart |
| Git working tree after local evidence | PASS | Clean; .local-notes/ ignored |
| Tempo | WARN | Previous OOMKilled, restart count 3, no new restart after smoke |

## k6 smoke result

Test script:

- tests/k6/smoke-test.js

Runtime configuration:

| Item | Value |
|---|---|
| GATEWAY_URL | http://100.65.255.2:30517 |
| RUN_ID | post-security-smoke-20260603194243 |
| VUs | 1 |
| Duration | 2 minutes |

Main results:

| Metric | Value |
|---|---|
| Checks total | 276 |
| Checks succeeded | 276 / 276 |
| Checks failed | 0 |
| HTTP requests | 276 |
| Iterations | 92 |
| HTTP failure rate | 0.00% |
| HTTP duration avg | 103.4 ms |
| HTTP duration median | 94.04 ms |
| HTTP duration p90 | 127.63 ms |
| HTTP duration p95 | 153.56 ms |
| HTTP duration max | 1.07 s |

Passed checks:

- health 200
- services health 200
- order created 2xx

## Post-smoke infrastructure snapshot

Node usage after smoke:

| Node | CPU | Memory |
|---|---|---|
| vm1-gateway | 465m / 11% | 5242Mi / 66% |
| vm2-mesh | 604m / 15% | 4526Mi / 57% |
| vm3-gitops | 779m / 19% | 6152Mi / 77% |

Important pod usage after smoke:

| Component | CPU | Memory |
|---|---|---|
| ecommerce-dashboard | 6m | 42Mi |
| web-gateway | 6m each | 44-45Mi |
| order-service | 20m each | 47-49Mi |
| inventory-api | 7m | 42Mi |
| payment-api | 5m | 42Mi |
| notification-api | 6m | 41Mi |
| read-model-service | 6m | 45Mi |
| PostgreSQL | 38m | 262Mi |
| PgBouncer | 13m | 5Mi |
| Kafka | 44m | 1336Mi |
| Kafka Connect / Debezium | 23m | 514Mi |
| OTel Collector | 2m | 40Mi |
| Tempo | 6m | 63Mi |

## Kafka and database result

Main Kafka consumer groups were checked after the smoke run.

Important groups observed with lag 0:

- inventory-service-group
- payment-service-group
- order-service-saga-monitor
- notification-service-group
- read-model-service-group
- clickhouse-orders-flat-cdc-v2

Database components:

| Component | Result |
|---|---|
| PgBouncer | Running, 0 restart |
| PostgreSQL | Running, 0 restart |

## Warning: Tempo memory

Tempo had a previous OOMKilled event:

| Item | Value |
|---|---|
| Restart count | 3 |
| Last reason | OOMKilled |
| Current state | Running / Ready |

No new restart was observed after the post-security smoke test.

This does not block the smoke check, but it must be watched during the next 10 RPS and 20 RPS benchmark runs, especially together with OTel trace volume.

## Final verdict

FINAL POST-SECURITY SMOKE CHECK: PASS WITH WARNING

The platform remains functional after security hardening. The public demo path still works, the direct web-gateway NodePort remains closed, admin NodePorts are protected by allowlist, k6 smoke passes with 0% HTTP failure rate, Kafka consumers are not lagging, and PostgreSQL/PgBouncer remain healthy.

The only warning is Tempo memory stability because of an earlier OOMKilled restart. The next benchmark phase should proceed carefully with monitoring for Tempo/OTel memory, Kafka lag, DB/PgBouncer, pod restarts, CPU/RAM, p95/p99 latency, and Saga completion.
