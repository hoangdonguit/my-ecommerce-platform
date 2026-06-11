# Phase 6 - Observability and Alert Response Runbook

## 1. Purpose

This runbook describes how to investigate the Phase 6 observability alerts and logging stack.

It is intended for the current advanced lab and pre-production prototype scope.

The platform currently has:

- Metrics through Prometheus
- Dashboards through Grafana
- Alerts through PrometheusRule
- Traces through OpenTelemetry Collector and Tempo
- Logs through Loki and Grafana Alloy
- GitOps state through ArgoCD

## 2. General First Checks

Always start with a safe read-only check.

    cd ~/Doanchuyennganh/my-ecommerce-platform

    git status --short
    git rev-list --left-right --count origin/main...HEAD

    kubectl get applications.argoproj.io -n argocd --sort-by=.metadata.name
    kubectl get pods -A | grep -Ev 'Running|Completed|STATUS' || true

Expected healthy baseline:

- Git status is clean
- Ahead/behind is 0 0
- ArgoCD applications are Synced and Healthy
- No unexpected bad pods

## 3. Check Active Prometheus Alerts

Use port-forward to Prometheus.

    kubectl -n monitoring port-forward svc/kube-prometheus-stack-prometheus 19090:9090

In another terminal:

    curl -s http://127.0.0.1:19090/api/v1/alerts \
      | python3 -m json.tool

Quick alert-name view:

    curl -s http://127.0.0.1:19090/api/v1/alerts \
      | grep -o 'ECommerce[A-Za-z0-9]*' \
      | sort -u

## 4. ECommercePodPendingTooLong

Meaning:

A pod has stayed Pending for more than 5 minutes.

Common causes:

- Not enough CPU or memory
- Node pressure
- PVC cannot bind
- Image pull issue
- Node selector or affinity problem
- Taints or tolerations mismatch

Commands:

    kubectl get pods -A | grep Pending || true
    kubectl describe pod -n <namespace> <pod>
    kubectl get events -A --sort-by=.lastTimestamp | tail -80
    kubectl describe node <node-name>
    kubectl get pvc -A

Logs are usually not available for a Pending pod because the container has not started.

## 5. ECommercePodCrashLooping

Meaning:

A container has been in CrashLoopBackOff for at least 5 minutes.

Commands:

    kubectl get pods -A | grep -E 'CrashLoopBackOff|Error' || true
    kubectl describe pod -n <namespace> <pod>
    kubectl logs -n <namespace> <pod> -c <container> --previous --tail=120
    kubectl logs -n <namespace> <pod> -c <container> --tail=120

Check Loki logs:

    kubectl -n observability port-forward svc/loki 13100:3100

Then query:

    curl -sG http://127.0.0.1:13100/loki/api/v1/query_range \
      --data-urlencode 'query={namespace="<namespace>", pod="<pod>"}' \
      --data-urlencode 'limit=50'

Look for:

- panic
- connection refused
- timeout
- authentication failure
- missing environment variable
- database migration error
- OOMKilled

## 6. ECommerceContainerRestarting

Meaning:

A container restarted within the last 10 minutes.

Commands:

    kubectl get pods -A
    kubectl describe pod -n <namespace> <pod>
    kubectl logs -n <namespace> <pod> -c <container> --previous --tail=120

Check restart reason:

    kubectl get pod -n <namespace> <pod> \
      -o jsonpath='{range .status.containerStatuses[*]}{.name}{" restartCount="}{.restartCount}{" lastReason="}{.lastState.terminated.reason}{" lastExitCode="}{.lastState.terminated.exitCode}{"\n"}{end}'

If the container is an Istio sidecar and the application is healthy, confirm whether the restart was caused by rollout or mesh update.

## 7. ECommerceDeploymentUnavailable

Meaning:

A deployment has unavailable replicas for more than 5 minutes.

Commands:

    kubectl -n <namespace> get deploy <deployment>
    kubectl -n <namespace> rollout status deploy/<deployment>
    kubectl -n <namespace> describe deploy <deployment>
    kubectl -n <namespace> get rs,pods -l app=<app-label>
    kubectl get events -n <namespace> --sort-by=.lastTimestamp | tail -80

Check:

- Readiness probe failure
- Image pull failure
- Pending pod
- CrashLoopBackOff
- Insufficient resources
- ConfigMap or Secret mismatch

## 8. ECommerceNodePressure

Meaning:

A node has reported DiskPressure, MemoryPressure, PIDPressure, or NetworkUnavailable.

Commands:

    kubectl get nodes
    kubectl describe node <node-name>
    kubectl top nodes
    kubectl top pods -A --sort-by=memory
    kubectl top pods -A --sort-by=cpu

Check disk usage on the node if shell access is available:

    df -h
    sudo du -xh /var/lib/rancher/k3s 2>/dev/null | sort -h | tail -30
    sudo du -xh /var/lib/kubelet 2>/dev/null | sort -h | tail -30

Immediate safe actions:

- Stop non-essential benchmark or chaos test
- Check large logs or images
- Avoid deleting PVC data blindly
- Do not reset the cluster unless explicitly planned

## 9. ECommerceTempoUnavailable

Meaning:

Tempo has unavailable replicas for more than 3 minutes.

Commands:

    kubectl -n observability get pods,deploy,svc | grep tempo
    kubectl -n observability describe deploy tempo
    kubectl -n observability describe pod -l app=tempo
    kubectl -n observability logs deploy/tempo --tail=120

Check readiness:

    kubectl -n observability port-forward svc/tempo 13200:3200

Then:

    curl -i http://127.0.0.1:13200/ready

Known historical issue:

Tempo had previous memory pressure or OOM risk. Check memory requests and limits before increasing traffic.

## 10. ECommerceArgoCDAppOutOfSync

Meaning:

An ArgoCD application has drifted from Git.

Commands:

    kubectl -n argocd get application <app-name>
    kubectl -n argocd describe application <app-name>

Check app status:

    kubectl -n argocd get application <app-name> \
      -o jsonpath='sync={.status.sync.status} health={.status.health.status} revision={.status.sync.revision} path={.spec.source.path}{"\n"}'

Check resource tree:

    kubectl -n argocd get application <app-name> \
      -o jsonpath='{range .status.resources[*]}{.kind} {.namespace}/{.name} status={.status} health={.health.status}{"\n"}{end}'

Safe correction flow:

- Confirm Git HEAD and origin/main
- Inspect local diff
- Commit and push the intended change
- Hard refresh the app
- Let ArgoCD reconcile

Hard refresh:

    kubectl -n argocd annotate application <app-name> \
      argocd.argoproj.io/refresh=hard \
      --overwrite

## 11. ECommerceArgoCDAppNotHealthy

Meaning:

An ArgoCD application has one or more unhealthy resources.

Commands:

    kubectl -n argocd get application <app-name>
    kubectl -n argocd describe application <app-name>

Then inspect the unhealthy resource from the resource tree.

Common mappings:

- monitoring-addons: PrometheusRule, ServiceMonitor, Grafana ConfigMap
- observability-layer: Tempo, OTel Collector, Loki, Alloy
- ecommerce-platform: application services, HPA/KEDA, gateway
- infrastructure-layer: Kafka and infrastructure services
- cdc-layer: Debezium and CDC services
- analytics-layer: ClickHouse and analytics services

## 12. Loki and Alloy Logging Checks

Check runtime:

    kubectl -n observability get pods,svc,pvc | grep -E 'loki|alloy|NAME'
    kubectl -n observability logs deploy/alloy --tail=120
    kubectl -n observability logs statefulset/loki --tail=120

Check Loki readiness:

    kubectl -n observability port-forward svc/loki 13100:3100

Then:

    curl -i http://127.0.0.1:13100/ready
    curl -s http://127.0.0.1:13100/loki/api/v1/labels | python3 -m json.tool

Query logs from default namespace:

    START_NS="$(python3 -c 'import time; print(int((time.time()-900)*1000000000))')"
    END_NS="$(python3 -c 'import time; print(int(time.time()*1000000000))')"

    curl -sG http://127.0.0.1:13100/loki/api/v1/query_range \
      --data-urlencode 'query={namespace="default"}' \
      --data-urlencode 'limit=20' \
      --data-urlencode "start=$START_NS" \
      --data-urlencode "end=$END_NS" \
      | python3 -m json.tool

Useful LogQL examples:

    {namespace="default"}
    {namespace="default", app="order-service"}
    {namespace="default", app="payment-api"}
    {namespace="db"}
    {namespace="kafka"}
    {namespace="observability", app="alloy"}
    {namespace="observability", app="loki"}

## 13. Grafana Datasource Checks

Port-forward Grafana:

    kubectl -n monitoring port-forward svc/kube-prometheus-stack-grafana 13000:80

Get admin password:

    GRAFANA_PASS="$(kubectl -n monitoring get secret kube-prometheus-stack-grafana -o jsonpath='{.data.admin-password}' | base64 -d)"

List datasources:

    curl -s -u "admin:$GRAFANA_PASS" http://127.0.0.1:13000/api/datasources \
      | python3 -m json.tool

Check Loki datasource health:

    curl -s -u "admin:$GRAFANA_PASS" http://127.0.0.1:13000/api/datasources/uid/loki/health \
      | python3 -m json.tool

Expected result:

- status: OK
- message: Data source successfully connected.

## 14. Evidence Policy

Keep commit-safe evidence in docs.

Do not commit:

- raw secret values
- kubeconfig files
- .env files
- private keys
- raw logs containing secrets
- .local-notes content

Raw inspection logs should stay under .local-notes and remain untracked.

## 15. Escalation Order

Use this order when debugging incidents:

1. ArgoCD sync and health
2. Kubernetes pod and deployment state
3. Prometheus metrics and alerts
4. Loki logs
5. Tempo traces
6. Database and Kafka state
7. Recent Git changes
8. Recent benchmark or chaos activity

## 16. Verdict

This runbook is the first operational response document for Phase 6.

It connects the alerting, logging, tracing, metrics, and GitOps layers into a practical incident investigation flow.
