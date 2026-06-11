# Phase 6.5 - Controlled Chaos Suite Summary

## 1. Purpose

This document records the Phase 6.5 controlled chaos proof.

The goal was to verify that the platform can tolerate selected controlled failures while retaining service availability, GitOps health, observability, and recovery behavior.

## 2. Evidence Location

Commit-safe artifacts:

- tests/chaos/results/phase6-controlled-chaos-suite-20260611221749

Raw local-only logs:

- .local-notes/chaos/phase6-controlled-chaos-suite-20260611221749

Raw full logs and port-forward logs are kept under .local-notes and are not committed.

## 3. Git Context

Current relevant commit before this evidence document:

- d574848 test: harden controlled chaos experiments

The controlled chaos manifests were hardened before execution:

- tests/chaos/experiments/pod-kill-payment.yaml
- tests/chaos/experiments/cpu-stress-inventory.yaml
- tests/chaos/experiments/network-delay-kafka.yaml

## 4. Scenarios Executed

### 4.1 Payment API one-shot pod kill

Experiment:

- Kind: PodChaos
- Name: payment-api-one-shot-pod-kill
- Namespace: default
- Target: app=payment-api
- Action: pod-kill
- Mode: one

Observed result:

- Existing payment-api pod was killed.
- Deployment created a replacement pod.
- payment-api rollout completed successfully.
- HTTP /api/health stayed 200 during the observation window.
- Chaos resource was deleted after the scenario.

Verdict: PASS.

### 4.2 Inventory API CPU stress

Experiment:

- Kind: StressChaos
- Name: inventory-api-cpu-stress
- Namespace: default
- Target: app=inventory-api
- Stressor: CPU load 100, workers 1
- Duration: 5 minutes

Observed result:

- Target inventory-api pod reached about 505m CPU during stress.
- HPA scaled inventory-api from 1 pod to 7 pods during the stress window.
- HTTP /api/health stayed 200 throughout the observation window.
- Chaos resource was deleted after the scenario.
- inventory-api rollout remained healthy.

Verdict: PASS.

### 4.3 Order service to Kafka network delay

Experiment:

- Kind: NetworkChaos
- Name: order-service-kafka-delay
- Namespace: default
- Target: app=order-service
- External target: kafka-svc.kafka.svc.cluster.local
- Delay: latency 1000ms, jitter 200ms
- Duration: 5 minutes

During the network delay, a small k6 baseline was executed:

- Script: tests/k6/baseline-e2e-5rps.js
- Rate: 5 RPS
- Duration: 60 seconds
- Iterations: 301
- Accepted orders: 301
- HTTP failed rate: NA
- Unexpected error rate: NA
- p95 latency: 74.181084 ms

Observed result:

- k6 5 RPS during Kafka delay passed thresholds.
- HTTP /api/health stayed 200 throughout the observation window.
- Temporary Kafka lag was observed in order-service-saga-monitor during the delay.
- After deleting the NetworkChaos resource, consumer lag returned to 0.
- order-service rollout remained healthy.

Verdict: PASS WITH EXPECTED TEMPORARY LAG.

## 5. Final Post-Chaos State

Final checks after cleanup:

- No remaining Chaos Mesh experiments.
- All ArgoCD applications were Synced and Healthy.
- Default application workloads were Running.
- PostgreSQL and PgBouncer were Running.
- Loki, Alloy, Tempo, and OTel Collector were Running.
- HTTP /api/health returned 200.
- Dashboard root returned 200.
- No active ECommerce alerts were observed.
- Loki continued to return recent application logs from the default namespace.

## 6. Important Observation

The CPU stress scenario demonstrated autoscaling behavior: inventory-api scaled out under CPU pressure while the public health endpoint remained available.

The Kafka delay scenario demonstrated temporary asynchronous backlog behavior. This is expected in an event-driven Saga system when communication to Kafka is delayed. The key evidence is that the system recovered after the chaos was removed and consumer lag returned to 0.

## 7. Limitations

This test does not prove full disaster recovery.

Not tested in this suite:

- PostgreSQL failure
- Kafka broker crash
- Node-level outage
- Persistent storage failure
- Long-duration network partition
- Multi-service simultaneous chaos

These are intentionally left for future production-hardening work.

## 8. Verdict

Result: PASS.

The platform passed the controlled chaos suite for stateless pod kill, CPU stress with autoscaling, and Kafka network delay with temporary backlog recovery.

The result is suitable for an advanced cloud-native lab and pre-production prototype report.
