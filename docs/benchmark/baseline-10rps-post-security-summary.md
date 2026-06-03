# Baseline 10 RPS Post-security Benchmark Summary

## Purpose

This benchmark validates the main order creation flow after the security hardening phase and after the Tempo memory limit adjustment.

The test confirms that the platform can sustain a 10 RPS baseline workload through the public demo entrypoint while keeping API errors, Kafka lag, database health, and observability stability under control.

## Commit policy

Commit-safe summary:

- docs/benchmark/baseline-10rps-post-security-summary.md

Local-only raw evidence:

- .local-notes/benchmark/post-security-10rps-20260603200030/

The local evidence directory is ignored by .gitignore and must not be committed.

## Run metadata

| Item | Value |
|---|---|
| Run ID | post-security-10rps-20260603200030 |
| Git revision before benchmark | c8355b6 |
| Public entrypoint | http://100.65.255.2:30517 |
| k6 script | tests/k6/baseline-e2e-10rps.js |
| Scenario | constant-arrival-rate |
| Target rate | 10 requests/second |
| Duration | 60 seconds |
| Pre-allocated VUs | 20 |
| Max VUs | 60 |
| Product ID | prod-123 |
| Request endpoint | POST /api/orders |

## Important precondition

Before this 10 RPS run, Tempo was found to be unstable with the old memory limit:

| Item | Old value |
|---|---|
| Tempo memory request | 256Mi |
| Tempo memory limit | 512Mi |
| Last failure | OOMKilled |
| Exit code | 137 |

The manifest was updated and synced by ArgoCD before the benchmark:

| Item | New value |
|---|---|
| Tempo memory request | 512Mi |
| Tempo memory limit | 1Gi |
| Commit | c8355b6 observability: raise tempo memory limit |

After rollout, the new Tempo pod started with restart count 0 and remained Running.

## k6 result

| Metric | Value |
|---|---|
| HTTP requests | 601 |
| Actual request rate | 9.9789 req/s |
| Iterations | 601 |
| Accepted orders | 601 |
| Rejected orders | Not present / none recorded |
| Dropped iterations | Not present / none recorded |
| Checks succeeded | 601 / 601 |
| Checks failed | 0 |
| HTTP failure rate | 0.00% |
| Unexpected error rate | 0.00% |
| HTTP duration avg | 133.06 ms |
| HTTP duration median | 115.98 ms |
| HTTP duration p90 | 182.19 ms |
| HTTP duration p95 | 206.84 ms |
| HTTP duration max | 895.18 ms |

Threshold result:

| Threshold | Result |
|---|---|
| http_req_duration p95 < 1500ms | PASS |
| http_req_failed < 1% | PASS |
| unexpected_error_rate < 1% | PASS |

## Post-run application health

| Component | Result |
|---|---|
| Main application pods | Running / Ready |
| Application pod restarts | 0 |
| PostgreSQL | Running, 0 restart |
| PgBouncer | Running, 0 restart |
| Kafka | Running |
| Kafka consumer lag | 0 for main groups |
| OTel Collector | Running |
| Tempo | Running, 0 restart after memory patch |
| Git working tree | Clean |

Main Kafka groups observed with lag 0:

- inventory-service-group
- payment-service-group
- order-service-saga-monitor
- notification-service-group
- read-model-service-group
- clickhouse-orders-flat-cdc-v2

## Resource snapshot after 10 RPS

Node usage immediately after the run:

| Node | CPU | Memory |
|---|---|---|
| vm1-gateway | 561m / 14% | 5859Mi / 73% |
| vm2-mesh | 770m / 19% | 4638Mi / 58% |
| vm3-gitops | 826m / 20% | 6157Mi / 77% |

Important pod usage immediately after the run:

| Component | CPU | Memory |
|---|---|---|
| ecommerce-dashboard | 5m | 43Mi |
| web-gateway | 5m each | 45-52Mi |
| order-service | 19-30m each | 41-51Mi |
| inventory-api | 5m | 43Mi |
| payment-api | 4m | 42Mi |
| notification-api | 5m | 41Mi |
| read-model-service | 6m | 45Mi |
| PostgreSQL | 106m | 294Mi |
| PgBouncer | 19m | 5Mi |
| Kafka | 44m | 1339Mi |
| Kafka Connect / Debezium | 24m | 514Mi |
| OTel Collector | 3m | 49Mi |
| Tempo | 6m | 629Mi |

## Cooldown observation

After a 180-second cooldown:

| Component | CPU | Memory |
|---|---|---|
| OTel Collector | 2m | 40Mi |
| Tempo | 5m | 80Mi |

Tempo remained Running with restart count 0 and no new OOMKilled event.

This shows that Tempo memory increased during/after trace ingestion, then dropped after flush/compaction. The 1Gi limit is appropriate for continuing to the next benchmark stage.

## Notes

- The run created 601 accepted orders in 60 seconds.
- No HTTP failures were observed.
- No unexpected application errors were observed.
- No Kafka backlog remained after the run.
- PostgreSQL and PgBouncer remained healthy.
- Tempo memory must still be watched in the 20 RPS benchmark because it peaked around 629Mi after the 10 RPS run.

## Final verdict

POST-SECURITY 10 RPS BASELINE: PASS

The platform handled the 10 RPS post-security baseline successfully with 0% HTTP failure rate, 0% unexpected error rate, p95 latency around 206.84ms, Kafka lag 0, and healthy database components.

The only operational note is observability memory behavior: Tempo required a memory limit increase before benchmark and should be monitored carefully in the 20 RPS run.
