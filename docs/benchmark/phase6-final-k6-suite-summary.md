# Phase 6 - Final K6 Suite Summary

## 1. Purpose

This document records the final k6 test suite after Phase 6 alerting, Loki logging, runbook, PostgreSQL backup/restore proof, and controlled chaos proof.

The suite includes functional, capacity, flash-sale, spike, stress, and soak scenarios.

## 2. Evidence Locations

Commit-safe artifacts:

- docs/benchmark/k6-final-artifacts/phase6-final-k6-suite-20260611233952

Ignored raw result directory:

- tests/k6/results/phase6-final-k6-suite-20260611233952

The raw result directory is intentionally ignored by .gitignore. The commit-safe extracted evidence is stored under docs/benchmark.

## Grafana Screenshot Note

Grafana screenshots are not used as primary evidence for this final benchmark run.

The reason is that the final k6 suite was executed as a continuous sequence of many scenarios. Functional, baseline, flash-sale, spike, stress, and soak tests ran close to each other, so Grafana time-series panels overlap multiple phases in the same dashboard window. A static screenshot would therefore be difficult to attribute accurately to one specific test.

The primary benchmark evidence is instead based on k6 summary exports, extracted k6 output blocks, Kubernetes post-checks, Prometheus alert checks, Loki quick checks, and Kafka lag snapshots.

Reference note:

- docs/benchmark/grafana-screenshots/phase6-final-k6-suite-20260611233952/README.md

## 3. High-Level Result

Baseline capacity tests from 5 RPS to 80 RPS completed with exit code 0.

The highest baseline p95 was observed at 80 RPS and remained far below the 1500 ms SLO threshold.

Heavy tests such as flash-sale-spike, spike-test, stress-test-multi, and stress-test also completed with exit code 0, but latency increased to the multi-second range. These scenarios should be interpreted as degradation and limit tests.

## 4. Detailed Metrics

See:

- docs/benchmark/k6-final-artifacts/phase6-final-k6-suite-20260611233952/detailed-k6-metrics.md
- docs/benchmark/k6-final-artifacts/phase6-final-k6-suite-20260611233952/detailed-k6-metrics.tsv
- docs/benchmark/k6-final-artifacts/phase6-final-k6-suite-20260611233952/k6-output-scenario-and-threshold-extracts.md
- docs/benchmark/k6-final-artifacts/phase6-final-k6-suite-20260611233952/k6-result-interpretation.md

## 5. Important Post-Suite Findings

### 5.1 Kafka backlog

After the full suite, inventory-service-group still had large Kafka lag on order.created partitions.

This indicates that the API layer could accept a large number of requests faster than the asynchronous inventory consumer tier could drain them.

This is an expected bottleneck under stress/spike testing and should be documented as a production-hardening item.

### 5.2 PostgreSQL direct connection pressure

Some order/payment reconciler CronJob pods failed with PostgreSQL:

FATAL: sorry, too many clients already

This indicates direct PostgreSQL connection pressure during/after the heavy suite.

Recommended hardening:

- move reconciler jobs through PgBouncer instead of postgresql:5432
- set CronJob concurrencyPolicy: Forbid
- reduce or tune batch size
- add failedJobsHistoryLimit and successfulJobsHistoryLimit
- add alert/runbook for PostgreSQL connection saturation

### 5.3 Runtime health

Postcheck still showed:

- ArgoCD applications Synced/Healthy
- HTTP /api/health returned 200
- no active ECommerce alert text was returned by the alert query
- application workloads remained Running
- inventory-consumer scaled to max replicas under Kafka lag

## 6. Verdict

Result: PASS WITH OBSERVED ASYNC BACKLOG AND DB CONNECTION PRESSURE.

The system demonstrated strong synchronous API capacity up to at least 80 RPS and survived all heavy k6 scenarios, but the final suite exposed two important bottlenecks:

- Kafka inventory consumer backlog under heavy order ingestion
- PostgreSQL connection saturation affecting reconciler CronJobs

These findings should be included in the report as realistic stress-test observations and future production-hardening work.
