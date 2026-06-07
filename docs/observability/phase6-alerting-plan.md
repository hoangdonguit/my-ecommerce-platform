# Phase 6.1 - Alerting Plan

## 1. Objective

Phase 6.1 adds the first operational alerting layer for the `my-ecommerce-platform` system after Phase 5 completed capacity validation and observability hardening.

The goal is not to add every possible enterprise alert at once. The goal is to create a safe, verifiable alerting foundation based only on metrics that are already expected to exist in the current Prometheus stack.

Main objectives:

- Detect pods stuck in Pending or scheduling-related failure states.
- Detect CrashLoopBackOff and abnormal container restarts.
- Detect deployments with unavailable replicas.
- Detect Kubernetes node pressure conditions.
- Add a dedicated Tempo availability alert because Tempo previously had OOM issues.
- Establish a clean foundation for future Kafka, PgBouncer, PostgreSQL, ArgoCD, SLO, and notification routing alerts.

## 2. Design Principles

Do not create alerts from metrics that are not currently available in Prometheus.

The first rule set only uses common Kubernetes, kube-state-metrics, and cAdvisor metrics:

- kube_pod_status_phase
- kube_pod_container_status_waiting_reason
- kube_pod_container_status_restarts_total
- kube_deployment_status_replicas_unavailable
- kube_node_status_condition

The following alert groups are intentionally excluded from the first implementation:

- Kafka consumer lag alerts
- PgBouncer pool saturation alerts
- PostgreSQL exporter alerts
- ArgoCD application health alerts
- Business SLO alerts

Reason: these alerts require exporter and scrape validation first. If Prometheus has no time series for those metrics, the rules would only be formal YAML and would not provide real operational value.

## 3. PrometheusRule Manifest

The first custom rule manifest is:

- k8s/monitoring/ecommerce-platform-alert-rules.yaml

The rule must include this label:

- release: kube-prometheus-stack

The current Prometheus instance selects PrometheusRule resources by this label.

## 4. Initial Alerts

### ECommercePodPendingTooLong

Triggers when a pod in an important namespace stays Pending for more than 5 minutes.

Possible causes:

- Insufficient CPU or memory.
- Overly strict node placement.
- PVC or storage issues.
- Taint and toleration mismatch.
- Image pull or scheduling problems.

### ECommercePodCrashLooping

Triggers when a container stays in CrashLoopBackOff for at least 5 minutes.

Possible causes:

- Application crash.
- Invalid configuration.
- Missing secret or environment variable.
- Dependency failure.
- Repeated probe failure.

### ECommerceContainerRestarting

Triggers when a non-Istio application container restarts within the last 10 minutes.

Possible causes:

- OOMKilled.
- Application panic.
- Probe-based restart.
- Manual rollout.
- Temporary dependency failure.

### ECommerceDeploymentUnavailable

Triggers when a deployment in an important namespace has unavailable replicas for more than 5 minutes.

Possible causes:

- Failed rollout.
- Pods cannot be scheduled.
- Pods cannot become Ready.
- Image pull failure.
- Resource pressure.

### ECommerceNodePressure

Triggers when a Kubernetes node reports DiskPressure, MemoryPressure, PIDPressure, or NetworkUnavailable for more than 5 minutes.

Possible impact:

- New pods may not be scheduled.
- Existing workloads may become unstable.
- Kafka, PostgreSQL, Tempo, ArgoCD, and benchmark reliability may be affected.

### ECommerceTempoUnavailable

Triggers when the Tempo deployment has unavailable replicas for more than 3 minutes.

Reason:

Tempo previously had OOM issues during high-load phases. This alert makes Tempo availability visible as part of the observability layer.

## 5. Out of Scope for This Step

This step does not configure Alertmanager notification receivers such as email, Telegram, Slack, or webhook routing.

This step does not add Kafka, PgBouncer, PostgreSQL, or ArgoCD alerts until their metrics are verified in Prometheus.

This step does not run chaos, spike, stress, or soak tests to trigger alerts. Trigger validation should be done only after the rule is loaded successfully.

## 6. Verification

After applying the manifest, verify the rule object:

- kubectl -n monitoring get prometheusrule ecommerce-platform-alert-rules

Then verify that Prometheus has loaded the rules:

- port-forward Prometheus
- query the Prometheus rules API
- confirm that ECommerce* rules are visible

The alerting foundation is considered PASS when Prometheus can see the ECommerce* rules.
