# Grafana No Data Investigation - Time Sync Root Cause

## Context

At the beginning of Phase 4, Grafana dashboards showed `No data` / temporary `NetworkError`.

Initial checks showed that the application and Prometheus were not down:

- Grafana pod was Running after recovery.
- Prometheus pod was Running.
- Prometheus still returned Kubernetes and Istio metrics.
- Direct Prometheus queries returned workload CPU and memory data.

## Root Cause

The main root cause was time drift on the operating machine used to access Grafana.

Before the fix, the local machine time was not synchronized correctly. As a result, Grafana relative time ranges such as `now-15m` and `now-30m` queried a time window that did not match the actual Prometheus metric timestamps.

Because of this, dashboards appeared as `No data` even though Prometheus still had valid data.

## Evidence

Direct Prometheus/Grafana datasource queries showed that metrics existed, but dashboard range queries using the wrong `now` window returned no series.

After checking the machine time, NTP was found inactive:

    System clock synchronized: no
    NTP service: inactive

## Fix

The machine timezone and NTP synchronization were corrected:

    sudo timedatectl set-timezone Asia/Ho_Chi_Minh
    sudo timedatectl set-ntp true
    sudo systemctl restart systemd-timesyncd

After the fix:

    System clock synchronized: yes
    NTP service: active
    Time zone: Asia/Ho_Chi_Minh (+07, +0700)

The built-in Kubernetes Grafana dashboard started displaying data again.

## Grafana Memory Note

During the investigation, Grafana was also observed to have been OOMKilled under the previous memory limit.

Grafana memory was increased through Helm:

    requests.memory = 512Mi
    limits.memory   = 1Gi

After the change:

    Grafana pod: Running
    Restart count: 0
    Observed memory usage: around 420Mi

This memory change is kept for now to reduce the risk of Grafana OOM during Phase 4 benchmark evidence collection.

## Temporary Dashboard Cleanup

A temporary Phase 4 dashboard was created during diagnosis, but it was removed after confirming that the original Kubernetes dashboard worked again once time synchronization was fixed.

The system continues to use the existing Grafana dashboards:

- Kubernetes / Compute Resources / Namespace Workloads
- Istio Service Dashboard
- Other existing monitoring dashboards

## Current Status

- Time synchronization is fixed.
- Built-in Kubernetes dashboard displays data.
- Istio dashboard is available.
- Temporary Phase 4 dashboard was removed.
- Git working tree was returned to clean state before writing this document.

## Notes for Phase 4

Before running benchmark or collecting observability evidence, always verify:

    date
    date -u
    timedatectl
    kubectl -n monitoring get pod

Expected time state:

    Time zone: Asia/Ho_Chi_Minh (+07, +0700)
    System clock synchronized: yes
    NTP service: active

This avoids false `No data` results in Grafana and prevents misleading benchmark evidence.
