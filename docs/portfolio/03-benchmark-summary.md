# 03 - Benchmark Summary

| Scenario | Result | Meaning |
| --- | --- | --- |
| Smoke test | End-to-end Saga passed | System is ready for deeper tests |
| Baseline | 180 RPS stable, p95 about 472.51 ms, 0% HTTP error | Highest confirmed fixed-resource stable rating |
| 200 RPS baseline | Near-threshold/degradation boundary | Not used as the stable rating |
| Idempotency | 200 HTTP requests mapped to 100 logical orders | Duplicate request protection works |
| Flash-sale | Stock 100 accepted exactly 100, rest rejected by 409 | Redis atomic stock gate prevents oversell |
| Soak 60 min | 93,954 requests, 0% HTTP error, p95 about 170.22 ms | Stable under sustained mixed workload |
| Stress | Latency and Kafka backlog increased | Controlled degradation beyond stable range |
| Chaos - order-service pod deletion | Recovery around 22 seconds, about 0.01% error | Kubernetes self-healing works under controlled failure |

## Interpretation

The project separates stable capacity, near-threshold behavior, and controlled overload. It does not claim production readiness. The results are valid for the current fixed-resource K3s lab environment.
