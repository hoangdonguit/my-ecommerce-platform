# Phase 5 VM Disk Cleanup Evidence

## Scope

This document records the VM disk cleanup after Phase 4 capacity benchmarking and stress-finding.

The cleanup was performed because long-running OpenStack VMs and repeated benchmark/stress tests accumulated container images, exited containers, completed debug pods, journal logs, temporary files, and service logs.

## Access Path

SSH was performed through Tailscale IPs:

- vm1-gateway: 100.65.255.2
- vm2-mesh: 100.71.13.10
- vm3-gitops: 100.126.17.65

The OpenStack private LAN IPs `192.168.100.*` are not the main SSH access path from the operator machine.

## Cleanup Actions

Safe cleanup only:

- journal vacuum
- apt cache cleanup
- unused container image prune
- old `/tmp` and `/var/tmp` cleanup
- completed/debug pod cleanup
- exited container cleanup
- second image prune
- ClickHouse text log truncate

The cleanup did not delete:

- Kubernetes state
- PostgreSQL data
- Kafka data
- ClickHouse data
- MongoDB data
- PVC/local-path data
- active container images
- active container snapshots

## Result Summary

Initial cleanup result:

- vm1-gateway root disk: 81% -> 63%
- vm2-mesh root disk: 74% -> 55%
- vm3-gitops root disk: 84% -> 77%

After cleanup round 2 and final verification:

- vm1-gateway root disk: 63%
- vm2-mesh root disk: 53%
- vm3-gitops root disk: 75%

ClickHouse log cleanup:

- `/var/log/clickhouse-server`: 396M -> 4.0K

## Final Disk State

vm1-gateway:

    root disk: 63%
    /var/lib/rancher/k3s: 6.2G
    /var/log: 153M
    /var/lib/kubelet: 62M
    journal: 112M

vm2-mesh:

    root disk: 53%
    /var/lib/rancher/k3s: 4.9G
    /var/log: 84M
    /var/lib/kubelet: 1.9M
    journal: 56M

vm3-gitops:

    root disk: 75%
    /var/lib/rancher/k3s: 11G
    /var/log: 520M
    /var/lib/kubelet: 501M
    journal: 240M

## Cluster State After Cleanup

After cleanup:

- Git state was clean.
- ArgoCD applications were Synced / Healthy.
- No abnormal non-Running pod remained.
- Kafka lag returned `NO_NONZERO_NUMERIC_LAG`.
- Nodes remained Ready.

## CronJob Observation

The `order-status-reconciler` and `payment-status-reconciler` CronJobs run every minute.

Current observed settings:

- `concurrencyPolicy: Forbid`
- `successfulJobsHistoryLimit: 3`
- `failedJobsHistoryLimit: 5`
- image: `postgres:16-alpine`
- nodeSelector: `vm3-gitops`

Completed reconciler pods are expected because these CronJobs run periodically.

## Remaining Risks

- vm3-gitops still has the highest disk usage.
- `/var/lib/rancher/k3s` on vm3-gitops remains around 11G.
- Old scheduling events still show `Insufficient cpu` and node affinity/selector constraints.
- Disk cleanup does not fix the main 60/70 RPS bottleneck.
- The next capacity hardening task should focus on node placement, resource requests/limits, and KEDA scaling behavior.

## Verdict

PASS WITH FOLLOW-UP.

The cleanup was effective and safe. The previous image filesystem pressure on vm3-gitops was reduced from the 84-85% range to about 75%.

The remaining capacity bottleneck is no longer primarily disk cleanup. The next issue to address is scheduling/capacity placement under HPA/KEDA scale-up.
