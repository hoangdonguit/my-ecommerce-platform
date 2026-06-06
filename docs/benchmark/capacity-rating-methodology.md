# Capacity Rating Methodology

## Scope

This document defines the capacity rating method for `my-ecommerce-platform`.

The goal is to answer the teacher's question:

1. What load level can the current fixed-resource system handle?
2. How should the team reason about the system limit when physical resources are no longer the first bottleneck?

This document is based on the Phase 4 post-hardening benchmark series:

- 5 RPS
- 10 RPS
- 20 RPS
- 30 RPS
- 40 RPS
- 50 RPS
- 60 RPS
- 70 RPS stress finding

## Why API Success Alone Is Not Enough

The system uses an asynchronous Saga flow.

A successful HTTP response from `POST /api/orders` only proves that the API hot path accepted the order.

A full capacity rating must also check whether the backend can drain the asynchronous workflow:

    order created
    -> outbox published
    -> Kafka event delivered
    -> inventory reserved
    -> payment completed
    -> order updated to COMPLETED
    -> notification/read model/analytics updated

Therefore, capacity is not judged only by k6 HTTP results.

## Capacity Pass Criteria

A benchmark is considered stable PASS only when all of the following are true:

- HTTP error rate < 1%
- unexpected error rate < 1%
- p95 latency < 1500ms
- accepted orders eventually become COMPLETED
- Kafka lag drains to zero after cooldown
- PgBouncer has no pool waiting backlog
- no abnormal pod crash/restart
- no pod remains Pending after cooldown
- autoscaling returns to a stable state
- completion time remains acceptable for the benchmark scope

## Capacity Verdict Levels

### PASS

The workload is stable.

Expected signs:

- k6 passes
- Kafka drains
- all accepted orders complete
- no Pending pod after cooldown
- HPA/KEDA stabilizes
- PgBouncer waiting remains zero

### PASS WITH WARNING

The workload completes, but the system shows signs of pressure.

Examples:

- transient Kafka lag increases
- consumer rebalancing appears
- autoscaling takes longer to stabilize
- completion time increases noticeably
- temporary operational warning appears but clears after stabilization

### OPERATIONAL CAPACITY WARNING

The API and business flow may still complete, but the infrastructure is no longer stable enough to call the load a safe rating.

Examples:

- pods remain Pending after cooldown
- FailedScheduling appears
- HPA/KEDA wants more pods than the cluster can schedule
- completion time increases sharply
- node disk/image filesystem pressure appears

### STRESS FINDING

The run is no longer used to prove stable capacity.

It is used to identify bottlenecks and breaking behavior.

## Current Fixed-Resource Cluster Rating

Based on Phase 4 benchmark evidence:

| Load | Orders/min approx | Result | Interpretation |
|---|---:|---|---|
| 5 RPS | 300 | PASS | Stable |
| 10 RPS | 600 | PASS | Stable |
| 20 RPS | 1200 | PASS | Stable |
| 30 RPS | 1800 | PASS WITH OBSERVATION | Stable API, transient async lag |
| 40 RPS | 2400 | PASS WITH OBSERVATION | Cleanest safe high-load point |
| 50 RPS | 3000 | PASS WITH WARNING | Burst capacity candidate |
| 60 RPS | 3600 | OPERATIONAL CAPACITY WARNING | Fixed-resource pressure appears |
| 70 RPS | 4200 | STRESS FINDING | Confirms scheduling/resource limit |

## Current Recommended Rating

For the current fixed-resource cluster:

    Safe rated capacity: 40 RPS, about 2400 orders/minute

    Burst capacity: 50 RPS, about 3000 orders/minute

    Overload warning zone: 60 RPS, about 3600 orders/minute

    Stress-finding zone: 70 RPS, about 4200 orders/minute

40 RPS is the safest value to report as the current stable capacity because it satisfies both API and operational stability expectations.

50 RPS can be reported as burst capacity because it completes successfully but introduces warning signs.

60 RPS and 70 RPS should not be reported as stable capacity for the current fixed-resource cluster.

## Observed Bottlenecks

The current bottleneck is not the API hot path.

The API accepted traffic successfully up to 70 RPS.

The current bottleneck is also not PgBouncer pool waiting, because PgBouncer showed:

    cl_waiting = 0
    maxwait = 0

The main observed bottleneck is the Kubernetes scheduling/autoscaling layer under higher load:

- HPA/KEDA requested more pods
- some desired pods could not be scheduled
- FailedScheduling appeared
- reasons included Insufficient CPU and node affinity/selector constraints
- vm3-gitops image filesystem pressure appeared

The async backend also showed increasing pressure:

- Kafka lag increased with RPS
- consumer groups rebalanced
- completion time increased as RPS increased

## Current Software/Architecture Limit Reasoning

The fixed-resource benchmark identifies the first practical limit of the current deployment.

To estimate a higher software-side limit, the next step is to remove the current infrastructure bottleneck and then retest.

The likely next bottlenecks to measure are:

1. Consumer throughput per replica
2. Kafka partition and consumer group throughput
3. PostgreSQL write throughput
4. PgBouncer pool saturation
5. Redis hot-path behavior
6. Outbox worker batch/interval limits
7. HPA/KEDA reaction time
8. Network and node I/O

The system should not claim an unlimited capacity even if CPU/RAM are increased.

Even with more physical resources, the software architecture will still have limits from Kafka partitions, database write throughput, consumer processing rate, and Saga completion time.

## How To Prove Higher Capacity Later

To test beyond the current fixed-resource rating:

1. Fix or relax node affinity/selector constraints.
2. Clean or expand vm3-gitops image filesystem.
3. Review CPU/memory requests for lightweight consumers.
4. Review KEDA targets and cooldown behavior.
5. Rerun 60 RPS and 70 RPS.
6. If 70 RPS becomes stable, run 80 RPS as stress-to-break.
7. Record new safe/burst/overload ratings.

## Report-Friendly Summary

With the current cluster resources, the system is rated safely at about 40 RPS, equivalent to about 2400 order requests per minute.

The system can burst to 50 RPS, equivalent to about 3000 order requests per minute, but with observable operational warnings.

At 60 RPS and above, the API still accepts requests and all orders eventually complete, but the Kubernetes scheduling/autoscaling layer shows fixed-resource pressure. Therefore, 60 RPS and 70 RPS are not considered stable capacity ratings for the current resource configuration.

