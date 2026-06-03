# Baseline 20 RPS Post-security Benchmark Summary

## Purpose

This benchmark validates the main order creation flow after the security hardening phase at a 20 RPS baseline workload.

The run was executed after two important stabilization steps:

- Tempo memory limit was raised to avoid OOMKilled during trace ingestion.
- ArgoCD was configured to ignore /spec/replicas for Deployment/default/order-service, so KEDA/HPA can control replica count without causing GitOps drift.

## Commit policy

Commit-safe summary:

- docs/benchmark/baseline-20rps-post-security-summary.md

Local-only raw evidence:

- .local-notes/benchmark/post-security-20rps-20260603212623/

The local evidence directory is ignored by .gitignore and must not be committed.

## Run metadata

| Item | Value |
|---|---|
| Run ID | post-security-20rps-20260603212623 |
| Git revision before final run | d79f7ec |
| Public entrypoint | http://100.65.255.2:30517 |
| k6 script | tests/k6/baseline-e2e-20rps.js |
| Scenario | constant-arrival-rate |
| Target rate | 20 requests/second |
| Duration | 60 seconds |
| Pre-allocated VUs | 40 |
| Product ID | prod-123 |
| Request endpoint | POST /api/orders |

## Pre-run stabilization

### Tempo memory

Tempo had previously reached OOMKilled with the old 512Mi memory limit. Before benchmark execution, Tempo was updated to:

| Item | Value |
|---|---|
| Tempo memory request | 512Mi |
| Tempo memory limit | 1Gi |

### ArgoCD and KEDA/HPA replica drift

During preflight, ecommerce-platform became OutOfSync because KEDA/HPA scaled Deployment/default/order-service away from the Git manifest replica count.

The fix was persisted through:

| File | Purpose |
|---|---|
| k8s/argocd/ecommerce-platform-application.yaml | Manage ecommerce-platform Application in Git |
| ignoreDifferences /spec/replicas | Let KEDA/HPA manage order-service replica count |
| RespectIgnoreDifferences=true | Prevent ArgoCD self-heal from fighting autoscaling |

After this fix, ecommerce-platform returned to Synced / Healthy before the 20 RPS run.

## k6 result

| Metric | Value |
|---|---|
| HTTP requests | 1200 |
| Actual request rate | 19.949 req/s |
| Iterations | 1200 |
| Accepted orders | 1200 |
| Rejected orders | Not present / none recorded |
| Dropped iterations | Not present / none recorded |
| Checks succeeded | 1200 / 1200 |
| Checks failed | 0 |
| HTTP failure rate | 0.00% |
| Unexpected error rate | 0.00% |
| HTTP duration avg | 144.00 ms |
| HTTP duration median | 120.03 ms |
| HTTP duration p90 | 198.36 ms |
| HTTP duration p95 | 263.72 ms |
| HTTP duration max | 1.25 s |

Threshold result:

| Threshold | Result |
|---|---|
| http_req_duration p95 < 1500ms | PASS |
| http_req_failed < 1% | PASS |
| unexpected_error_rate < 1% | PASS |

## Comparison with 10 RPS baseline

| Metric | 10 RPS | 20 RPS | Observation |
|---|---:|---:|---|
| HTTP requests | 601 | 1200 | Approximately doubled |
| Actual request rate | 9.9789 req/s | 19.949 req/s | Target achieved |
| Accepted orders | 601 | 1200 | Approximately doubled |
| HTTP failure rate | 0.00% | 0.00% | Stable |
| Unexpected error rate | 0.00% | 0.00% | Stable |
| Avg latency | 133.06 ms | 144.00 ms | Slight increase |
| p90 latency | 182.19 ms | 198.36 ms | Slight increase |
| p95 latency | 206.84 ms | 263.72 ms | Increased but still low |
| Max latency | 895.18 ms | 1.25 s | Increased but within threshold |

## Immediate post-run health

Immediately after the 20 RPS run:

| Component | Result |
|---|---|
| ArgoCD ecommerce-platform | Synced / Progressing |
| Main app pods | Running / Ready except one transient order-service Pending pod |
| order-service | replicas=5, ready=4, updated=5, available=4 |
| Tempo | Running, restart 0 |
| Tempo memory | 666Mi |
| OTel Collector memory | 62Mi |
| Kafka consumer lag | No nonzero numeric lag |
| PgBouncer | Running, 0 restart |
| PostgreSQL | Running, 0 restart |
| Git working tree | Clean |

The temporary Progressing state was caused by KEDA/HPA scaling order-service after load.

## Post-run resource snapshot

Node usage immediately after the run:

| Node | CPU | Memory |
|---|---|---|
| vm1-gateway | 677m / 16% | 5876Mi / 74% |
| vm2-mesh | 583m / 14% | 4597Mi / 57% |
| vm3-gitops | 737m / 18% | 6174Mi / 77% |

Important pod usage immediately after the run:

| Component | CPU | Memory |
|---|---|---|
| ecommerce-dashboard | 6m | 43Mi |
| web-gateway | 5-6m each | 49-52Mi |
| order-service | 20-23m each | 48-53Mi |
| inventory-api | 6m | 43Mi |
| payment-api | 6m | 42Mi |
| notification-api | 5m | 41Mi |
| read-model-service | 6m | 48Mi |
| PostgreSQL | 65m | 337Mi |
| PgBouncer | 21m | 5Mi |
| Kafka | 39m | 1341Mi |
| Kafka Connect / Debezium | 24m | 514Mi |
| OTel Collector | 5m | 62Mi |
| Tempo | 6m | 666Mi |

## Cooldown observation

After a 180-second cooldown:

| Component | Result |
|---|---|
| ArgoCD applications | All Synced / Healthy |
| order-service | replicas=2, ready=2, updated=2, available=2 |
| Pending order-service pods | None |
| HPA order-service | cpu 6% / 25%, minPods 2, maxPods 8, replicas 2 |
| Tempo | Running, restart 0, LAST_REASON none |
| Tempo memory | 84Mi |
| OTel Collector memory | 43Mi |
| Node CPU/RAM | Returned to normal baseline range |
| Non-running pods | Completed jobs only |

Cooldown node usage:

| Node | CPU | Memory |
|---|---|---|
| vm1-gateway | 373m / 9% | 5278Mi / 66% |
| vm2-mesh | 516m / 12% | 4546Mi / 57% |
| vm3-gitops | 555m / 13% | 6125Mi / 77% |

## Kafka and database result

Kafka consumer lag after the run:

| Area | Result |
|---|---|
| Kafka nonzero numeric lag | None observed |
| inventory-service-group | lag 0 |
| payment-service-group | lag 0 |
| order-service-saga-monitor | lag 0 |
| notification-service-group | lag 0 |
| read-model-service-group | lag 0 |
| clickhouse-orders-flat-cdc-v2 | lag 0 |

Database components after the run:

| Component | Result |
|---|---|
| PgBouncer | Running, 0 restart |
| PostgreSQL | Running, 0 restart |

## Notes

- The run created 1200 accepted orders in 60 seconds.
- No HTTP failures were observed.
- No unexpected application errors were observed.
- No Kafka backlog remained after the run.
- PostgreSQL and PgBouncer remained healthy.
- KEDA/HPA temporarily scaled order-service up to 5 replicas after the run.
- One order-service pod was briefly Pending during scale-up, but this resolved after cooldown.
- Tempo peaked around 666Mi after 20 RPS and dropped to 84Mi after cooldown without restart or OOMKilled.

## Final verdict

POST-SECURITY 20 RPS BASELINE: PASS WITH TRANSIENT AUTOSCALE NOTE

The platform handled the 20 RPS post-security baseline successfully with 0% HTTP failure rate, 0% unexpected error rate, p95 latency around 263.72ms, Kafka lag 0, and healthy database components.

The only operational note is that KEDA/HPA temporarily scaled order-service and caused a short Progressing/Pending state immediately after the run. After cooldown, ArgoCD returned to Synced / Healthy, order-service returned to 2 ready replicas, and Tempo memory dropped back to a normal range without restart.
