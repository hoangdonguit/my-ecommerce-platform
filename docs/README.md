# Documentation Convention

This folder stores commit-safe documentation and evidence for the `my-ecommerce-platform` project.

Do not commit raw secrets, API keys, passwords, private keys, `.env`, or local-only notes.

## Folder Meaning

### benchmark

Use for load test, stress test, spike test, k6 results, RPS/latency/error-rate summaries, and performance bottleneck analysis.

Raw benchmark artifacts should be stored under:

    docs/benchmark/runs/<run-id>/

Benchmark summary files should be stored directly under:

    docs/benchmark/

### evidence

Use for runtime proof or validation of a specific technical claim.

Examples:

- payment outbox runtime proof
- DLQ/retry validation
- final validation after upgrades

Raw proof artifacts that belong to a specific subsystem may stay under that subsystem's `runs/` folder.

### security

Use for security and hardening documents:

- secret handling
- admin NodePort allowlist
- Istio mTLS/AuthZ
- NetworkPolicy
- securityContext/probes
- repository hardcode scan

Raw security proof artifacts should be stored under:

    docs/security/runs/<run-id>/

### opentelemetry

Use for observability documents:

- Grafana
- Prometheus
- OpenTelemetry Collector
- Tempo
- traces
- dashboards
- metric/tracing incidents

### runbook

Use for operational recovery guides and troubleshooting procedures.

Examples:

- Kafka lag runbook
- DLQ/retry runbook
- PostgreSQL backup/restore runbook
- Grafana No Data runbook

### report

Use for report-facing summaries written for the school project report.

These files should be easier to read than raw evidence files.

### checkpoints

Use for point-in-time system snapshots.

A checkpoint should record:

- timestamp
- git commit
- ArgoCD state
- runtime URL
- pod state
- important proof already completed
- known risks
- next steps

### gitops

Use for ArgoCD/GitOps management notes.

### cdc

Use for CDC / Debezium / Kafka Connect / PostgreSQL logical replication documents.

Raw CDC proof artifacts should be stored under:

    docs/cdc/runs/<run-id>/

### clickhouse

Use for ClickHouse analytics/read-side documents.

Raw ClickHouse proof artifacts should be stored under:

    docs/clickhouse/runs/<run-id>/

### dynamic-filter

Use for kafka-connect-dynamic-filter plugin research, build, audit, and runtime proof.

Raw dynamic-filter proof artifacts should be stored under:

    docs/dynamic-filter/runs/<run-id>/

## Benchmark and Evidence Rule

For every important runtime change, create evidence before moving to the next major step.

Important changes include:

- payment/outbox changes
- Kafka consumer changes
- DB/PgBouncer changes
- security/Istio/NetworkPolicy changes
- observability changes
- KEDA/HPA resource changes
- benchmark threshold changes
- chaos/backup/restore tests

Recommended flow:

1. Pre-check

    - git status
    - git revision
    - ArgoCD health
    - pod state
    - Kafka lag
    - DB/outbox pending state

2. Run the smallest valid smoke/E2E test first.

3. Run benchmark only after smoke/E2E passes.

4. Collect post-check evidence.

    - k6 result
    - Kafka lag after
    - DB/outbox after
    - pod restarts/errors after
    - CPU/memory if relevant

5. Write a summary document.

Benchmark summary files should include:

    # Title
    ## Scope
    ## Environment
    ## Command
    ## Result
    ## Runtime Evidence
    ## Bottleneck / Observation
    ## Verdict
    ## Next Action

Use clear verdicts:

- PASS
- PASS WITH WARNING
- FAIL
- NEED VERIFY

## Placement Rule

Do not move files automatically based only on keyword search.

A raw artifact should usually stay with its run folder, even if it contains keywords from another domain.

For example:

- `docs/benchmark/runs/.../otel-check.txt` can stay in benchmark if it was collected for a benchmark run.
- `docs/security/runs/.../positive-smoke.txt` can stay in security if it supports an AuthZ proof.
- `docs/cdc/runs/.../replication-slot.txt` can stay in CDC if it supports a CDC proof.

Move only when the document's main purpose is clearly wrong for its folder.
