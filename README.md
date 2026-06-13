# my-ecommerce-platform

A cloud-native, event-driven e-commerce order processing platform built for **NT114 - Specialized Project**.

This project demonstrates how an e-commerce order can be processed across multiple distributed services using Kubernetes, GitOps, Apache Kafka, Saga Choreography, Transactional Outbox, Redis, PostgreSQL, MongoDB, ClickHouse, CDC, autoscaling, observability, security hardening, performance testing, and chaos testing.

This is not a simple CRUD API demo. The system is designed as a production-oriented academic prototype that focuses on distributed transaction consistency, asynchronous processing, infrastructure automation, operational visibility, and capacity evaluation.

---

## 1. Project Information

| Item             | Description                                                                                                                              |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Course           | NT114 - Specialized Project                                                                                                              |
| Vietnamese title | Thiết kế và triển khai nền tảng xử lý đơn hàng thương mại điện tử theo kiến trúc Microservices Cloud-Native sử dụng Kubernetes và GitOps |
| English title    | Design and Implementation of a Cloud-Native Microservices E-commerce Order Processing Platform using Kubernetes and GitOps               |
| Supervisor       | MSc. Lê Anh Tuấn                                                                                                                         |
| Repository       | `https://github.com/hoangdonguit/my-ecommerce-platform.git`                                                                              |

## 2. Team Members

| No. | Full name       | Student ID |
| --: | --------------- | ---------- |
|   1 | Hoàng Xuân Đồng | 23520297   |
|   2 | Đỗ Thái Hậu     | 23520450   |

---

## 3. Project Goals

The goal of this project is to build a distributed e-commerce order processing platform that reflects real cloud-native system design problems.

Instead of processing an order in a single monolithic transaction, the system breaks the workflow into multiple independent services and asynchronous events. This makes the system easier to scale, easier to observe, and more resilient to partial failures.

Main goals:

* Implement distributed order processing using **Saga Choreography**.
* Use **Apache Kafka** as the event backbone between services.
* Use the **Transactional Outbox** pattern to reduce database-message dual-write risk.
* Use **Redis-based idempotency** to reduce duplicate order creation.
* Use a **Redis atomic stock gate** for flash-sale-like high concurrency scenarios.
* Use **PostgreSQL database-per-service** as the transactional source of truth.
* Use **PgBouncer** to reduce PostgreSQL connection pressure.
* Use **MongoDB** for CQRS read models.
* Use **ClickHouse** for analytics.
* Use **Debezium and Kafka Connect** for CDC from PostgreSQL.
* Use a custom **Dynamic Redis Filter** for filtered CDC streaming.
* Use **KEDA and HPA** for autoscaling.
* Use **ArgoCD** for GitOps deployment.
* Use **Istio mTLS and AuthorizationPolicy** for internal service security.
* Use **Prometheus, Grafana, Loki/Alloy, OpenTelemetry, and Tempo** for observability.
* Use **k6 and Chaos Mesh** for performance and resilience validation.

---

## 4. System Overview

The platform is organized into several layers.

| Layer               | Components                                                                      | Purpose                                                                            |
| ------------------- | ------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Access/UI           | `ecommerce-dashboard`, `web-gateway`                                            | Provides user/demo entry points and routes requests to internal services           |
| Business Services   | `order-service`, `inventory-service`, `payment-service`, `notification-service` | Implements the order processing workflow                                           |
| Read Model          | `read-model-service`, MongoDB                                                   | Provides CQRS read-side data for dashboard queries                                 |
| Messaging           | Apache Kafka                                                                    | Connects services through asynchronous domain events                               |
| Transactional Data  | PostgreSQL, PgBouncer                                                           | Stores source-of-truth data and manages DB connections                             |
| Cache/Runtime State | Redis                                                                           | Supports idempotency, cache-aside, flash-sale stock gate, and Dynamic Filter rules |
| Analytics           | ClickHouse                                                                      | Stores CDC-derived analytical data                                                 |
| CDC                 | Kafka Connect, Debezium, Dynamic Redis Filter                                   | Streams PostgreSQL changes into Kafka and analytics topics                         |
| Platform            | Kubernetes/K3s, Docker, ArgoCD, KEDA, Istio                                     | Runs, deploys, scales, and secures workloads                                       |
| Observability       | Prometheus, Grafana, Loki, Alloy, OpenTelemetry, Tempo                          | Collects metrics, logs, traces, dashboards, and alerts                             |
| Testing/Ops         | k6, Chaos Mesh, smoke scripts, runbooks                                         | Validates correctness, performance, resilience, and recovery procedures            |

---

## 5. High-Level Architecture

```text
Client / Dashboard / k6
        |
        v
+---------------------+
| ecommerce-dashboard |
+---------------------+
        |
        v
+---------------------+
|     web-gateway     |
| API key + read cache|
+---------------------+
        |
        v
Istio mTLS + AuthorizationPolicy
        |
        +-------------------+-------------------+-------------------+
        |                   |                   |                   |
        v                   v                   v                   v
+---------------+   +---------------+   +---------------+   +-------------------+
| order-service |   | inventory-api |   |  payment-api  |   | notification-api  |
+---------------+   +---------------+   +---------------+   +-------------------+
        |
        v
PostgreSQL order_db + Transactional Outbox
        |
        v
Apache Kafka
        |
        +-------------------+-------------------+-------------------+
        |                   |                   |
        v                   v                   v
inventory-consumer   payment-consumer    notification-consumer
        |                   |                   |
        v                   v                   v
inventory_db       payment_db        notification_db
```

Side paths:

```text
Kafka payment.completed
        |
        v
read-model-service
        |
        v
MongoDB ecommerce_read.order_read_models
        |
        v
Dashboard read-model page
```

```text
PostgreSQL order_db.public.orders
        |
        v
Debezium PostgreSQL Connector
        |
        v
Kafka CDC topics
        |
        v
ClickHouse analytics
```

---

## 6. Order Processing Flow

### 6.1. Successful Saga Flow

```text
Client / k6 / Dashboard
    -> web-gateway
    -> order-service
    -> order_db.orders + order_db.order_items + order_db.outbox
    -> order outbox worker
    -> Kafka topic: order.created
    -> inventory-consumer
    -> inventory_db.inventory_reservations
    -> inventory_db.inventory_outbox_events
    -> inventory outbox publisher
    -> Kafka topic: inventory.reserved
    -> payment-consumer
    -> payment_db.payments + payment_db.payment_attempts
    -> payment_db.payment_outbox_events
    -> payment outbox publisher
    -> Kafka topic: payment.completed
    -> fan-out:
        1. order-service saga monitor updates order status to COMPLETED
        2. read-model-service upserts MongoDB read model
        3. notification-consumer creates notification
```

### 6.2. Failure and Compensation Flow

```text
inventory.failed / payment.failed
    -> order-service saga monitor
    -> order status becomes FAILED or CANCELLED
    -> inventory rollback or stock release is triggered when needed
    -> notification-consumer may create failure notification
```

### 6.3. MongoDB Read Model Behavior

MongoDB is not the final linear step after PostgreSQL order status update. It is an independent CQRS branch that consumes `payment.completed`.

```text
Kafka topic: payment.completed
    ├── order-service saga monitor -> PostgreSQL orders.status = COMPLETED
    ├── read-model-service -> MongoDB order_read_models.saga_status = COMPLETED
    └── notification-consumer -> notification_db.notifications
```

Because these consumers run independently, the dashboard may show MongoDB read model data earlier than the order tab refreshes from PostgreSQL. This is expected in an eventually consistent architecture.

---

## 7. Microservices

### 7.1. web-gateway

The Web Gateway is the main API entry point.

Responsibilities:

* receives dashboard and test requests;
* validates `X-API-Key`;
* routes requests to internal services;
* applies selected Redis cache-aside logic;
* exposes unified API paths for order lookup, read model queries, inventory queries, and Saga tracing.

Typical API groups:

```text
/api/orders
/api/orders/:id
/api/orders/:id/saga
/api/read-model/orders
/api/inventories
/api/health
```

### 7.2. order-service

The Order Service owns order creation and Saga state.

Responsibilities:

* validates order requests;
* supports idempotency;
* writes `orders` and `order_items`;
* writes `order.created` into the order outbox in the same database transaction;
* runs an outbox worker that publishes events to Kafka;
* consumes terminal Saga events such as `payment.completed`, `payment.failed`, and `inventory.failed`;
* updates final order status.

Key design point:

```text
Order data and order.created event are persisted atomically.
Kafka publishing happens later through the outbox worker.
```

This reduces the risk of saving an order without publishing its corresponding event.

### 7.3. inventory-service

The Inventory Service is split into API and consumer workloads.

Responsibilities:

* manages product stock;
* consumes `order.created`;
* reserves inventory for order items;
* writes reservation records;
* writes resulting events into `inventory_outbox_events`;
* publishes `inventory.reserved` or `inventory.failed`;
* supports rollback/release behavior.

### 7.4. payment-service

The Payment Service simulates COD/payment processing.

Responsibilities:

* consumes `inventory.reserved`;
* creates payment records;
* records payment attempts;
* writes terminal payment events into `payment_outbox_events`;
* publishes `payment.completed` or `payment.failed` through the payment outbox publisher.

The service avoids direct Kafka publishing from the application service. Payment terminal events are published through the outbox path.

### 7.5. notification-service

The Notification Service reacts to terminal payment events.

Responsibilities:

* consumes `payment.completed` and `payment.failed`;
* creates user-facing notifications;
* marks notifications as sent;
* retries failed message handling;
* uses DLQ topics after repeated failures.

### 7.6. read-model-service

The Read Model Service implements the CQRS read side.

Responsibilities:

* consumes `payment.completed`;
* transforms event data into dashboard-friendly read models;
* upserts documents into MongoDB;
* supports fast read-side queries.

### 7.7. ecommerce-dashboard

The dashboard provides the demonstration and operational UI.

Main pages/features:

* storefront/demo order creation;
* order lookup;
* failed order monitoring;
* Saga tracing;
* MongoDB read model view;
* system overview.

---

## 8. Data Layer

### 8.1. PostgreSQL

PostgreSQL is the transactional source of truth.

Databases:

| Database          | Owner                | Main Data                                  |
| ----------------- | -------------------- | ------------------------------------------ |
| `order_db`        | order-service        | orders, order items, order outbox          |
| `inventory_db`    | inventory-service    | inventory, reservations, inventory outbox  |
| `payment_db`      | payment-service      | payments, payment attempts, payment outbox |
| `notification_db` | notification-service | notifications                              |

PostgreSQL is used because it provides transactional guarantees, reliable source-of-truth storage, and CDC support through WAL/replication.

### 8.2. PgBouncer

PgBouncer is a lightweight PostgreSQL connection pooler.

It solves a practical operational problem: many services, consumers, and background jobs can create too many direct PostgreSQL connections. PgBouncer pools and reuses connections, reducing connection churn and improving stability under benchmark load.

### 8.3. Redis

Redis is used for several runtime purposes:

1. idempotency;
2. cache-aside for selected read APIs;
3. flash-sale atomic stock gate;
4. Dynamic Filter rule storage.

Important Dynamic Filter key:

```text
filter:order-status
```

Example rule:

```json
["PENDING", "COMPLETED"]
```

### 8.4. MongoDB

MongoDB stores CQRS read models.

Database and collection:

```text
ecommerce_read.order_read_models
```

MongoDB is used because read models are document-oriented and optimized for dashboard queries. PostgreSQL remains the transactional source of truth.

### 8.5. ClickHouse

ClickHouse is used for analytics.

Purpose:

* stores flattened CDC events;
* supports analytical queries;
* separates OLAP workloads from PostgreSQL OLTP workloads;
* demonstrates an analytics pipeline built from CDC events.

---

## 9. Kafka, Saga, and Transactional Outbox

### 9.1. Kafka Topics

Main Saga topics:

```text
order.created
inventory.reserved
inventory.failed
payment.completed
payment.failed
order.cancelled
```

CDC topics:

```text
cdc.order_db.public.orders
cdc_dynamic.order_db.public.orders
```

### 9.2. Consumer Groups

Main consumer groups:

```text
inventory-service-group
inventory-rollback-group-v2
payment-service-group
notification-service-group
order-service-saga-monitor
read-model-service-group
```

### 9.3. Transactional Outbox

The system uses three main outbox layers:

| Service           | Outbox Table                           | Published Events                         |
| ----------------- | -------------------------------------- | ---------------------------------------- |
| order-service     | `order_db.outbox`                      | `order.created`                          |
| inventory-service | `inventory_db.inventory_outbox_events` | `inventory.reserved`, `inventory.failed` |
| payment-service   | `payment_db.payment_outbox_events`     | `payment.completed`, `payment.failed`    |

The outbox pattern solves the database-message dual-write problem.

Without outbox:

```text
write database
publish Kafka event
```

If the database write succeeds but Kafka publishing fails, the system can become inconsistent.

With outbox:

```text
write business data + write outbox event in one transaction
outbox worker publishes event later
mark event as PUBLISHED
```

### 9.4. DLQ and Retry

The system includes retry and DLQ behavior for important Kafka consumers.

| Consumer              | Input Topic          | DLQ Topic                |
| --------------------- | -------------------- | ------------------------ |
| payment-consumer      | `inventory.reserved` | `inventory.reserved.dlq` |
| notification-consumer | `payment.completed`  | `payment.completed.dlq`  |
| notification-consumer | `payment.failed`     | `payment.failed.dlq`     |

Processing model:

```text
FetchMessage
    -> process business logic
    -> retry on transient error
    -> publish to DLQ if retries fail
    -> commit offset to avoid blocking the consumer group
```

---

## 10. CDC, Kafka Connect, Debezium, and Dynamic Redis Filter

### 10.1. Debezium CDC

Debezium captures changes from PostgreSQL and streams them to Kafka.

Normal connector:

```text
order-db-orders-connector
PostgreSQL order_db.public.orders
    -> cdc.order_db.public.orders
```

CDC is a side path. It supports analytics and downstream streaming. It is not required for an order to become `COMPLETED`.

### 10.2. Kafka Connect

Kafka Connect runs Debezium connectors and manages connector tasks.

The project includes a GitOps-managed connector bootstrap mechanism:

```text
CronJob kafka-connect-connector-bootstrap
    -> reads connector templates
    -> injects database password from Kubernetes Secret at runtime
    -> applies connector configs through Kafka Connect REST API
```

The bootstrap job avoids printing raw secrets.

### 10.3. Dynamic Redis Filter

The Dynamic Redis Filter is a custom Kafka Connect SMT that filters CDC events based on rules stored in Redis.

Dynamic connector:

```text
order-db-orders-dynamic-filter-connector
PostgreSQL order_db.public.orders
    -> Debezium
    -> RedisDynamicFilter SMT
    -> cdc_dynamic.order_db.public.orders
```

Important configuration:

| Setting                                           | Meaning                                      |
| ------------------------------------------------- | -------------------------------------------- |
| `transforms.dynamicFilter.type`                   | Custom Redis Dynamic Filter class            |
| `transforms.dynamicFilter.field.name`             | Field used for filtering, currently `status` |
| `transforms.dynamicFilter.redis.uri`              | Redis service URI                            |
| `transforms.dynamicFilter.redis.key`              | Redis key containing allowed values          |
| `transforms.dynamicFilter.redis.poll.interval.ms` | Poll interval for refreshing rules           |
| `transforms.dynamicFilter.empty.list.behavior`    | Behavior when the rule list is empty         |

Why it matters:

* filter rules can be changed through Redis;
* Kafka Connect does not need to be restarted;
* downstream CDC topics can focus on relevant order statuses;
* the project demonstrates a custom streaming extension beyond basic Debezium usage.

---

## 11. Kubernetes, Docker, and GitOps

### 11.1. Docker

Each service is containerized.

Typical Go service image pattern:

1. build Go binary in a builder stage;
2. copy the binary into a smaller runtime image;
3. run the service inside Kubernetes.

The dashboard is built with Node/Vite and served through Nginx.

Kafka Connect uses a custom image that includes Debezium and the Dynamic Redis Filter plugin.

### 11.2. Kubernetes/K3s

The system runs on Kubernetes/K3s.

Kubernetes resources used:

* Deployments;
* StatefulSets;
* Services;
* ConfigMaps;
* Secrets;
* ServiceAccounts;
* RBAC;
* CronJobs;
* Jobs;
* Probes;
* resource requests and limits;
* HPA/KEDA autoscaling objects.

### 11.3. ArgoCD GitOps

ArgoCD synchronizes Kubernetes manifests from GitHub to the cluster.

Main ArgoCD applications:

```text
analytics-layer
cdc-layer
ecommerce-infrastructure
ecommerce-platform
infrastructure-layer
monitoring-addons
observability-layer
security-layer
```

Expected state:

```text
Synced / Healthy
```

Benefits:

* Git becomes the desired state;
* deployment history is auditable;
* runtime drift can be detected;
* infrastructure can be reconstructed more reliably.

---

## 12. Autoscaling with KEDA and HPA

HPA scales workloads based on resource metrics such as CPU usage.

KEDA scales Kafka consumers based on Kafka lag. This is important because CPU usage alone may not reflect message backlog. A consumer may have low CPU but still be behind in message processing.

| Workload              | Scaling Signal                                      |
| --------------------- | --------------------------------------------------- |
| API services          | CPU-based HPA                                       |
| inventory-consumer    | Kafka lag on `order.created`                        |
| payment-consumer      | Kafka lag on `inventory.reserved`                   |
| notification-consumer | Kafka lag on `payment.completed` / `payment.failed` |

The inventory consumer is especially important because it is the first asynchronous stage after `order.created`. If it cannot drain the backlog quickly enough, orders remain pending longer.

---

## 13. Security and Networking

Security layers:

* Kubernetes Secrets for runtime values;
* dedicated ServiceAccounts for workloads;
* RBAC for limited permissions;
* Istio mTLS for encrypted service-to-service traffic;
* AuthorizationPolicy to restrict internal API access;
* NetworkPolicy for selected namespace/service traffic;
* restricted admin exposure.

Important idea:

```text
Public clients should go through dashboard/gateway.
Internal APIs should not be directly callable by arbitrary pods.
```

The project does not claim complete Zero Trust. It implements a Zero-Trust-oriented foundation suitable for an advanced academic prototype.

---

## 14. Observability and Operations

### 14.1. Prometheus and Grafana

Prometheus collects metrics. Grafana visualizes the system.

Used for:

* pod status;
* deployment availability;
* node health;
* resource usage;
* benchmark observation;
* alert analysis.

### 14.2. Loki and Alloy

Loki stores logs. Alloy collects Kubernetes/service logs.

Purpose:

* centralize logs from multiple pods;
* debug Kafka consumers and Saga failures;
* reduce dependence on manual `kubectl logs`.

### 14.3. OpenTelemetry and Tempo

OpenTelemetry instruments services and propagates trace context. Tempo stores distributed traces.

Why it matters:

* HTTP requests can be traced across services;
* Kafka headers can carry trace context across asynchronous boundaries;
* outbox workers and consumers can be connected to a logical trace;
* distributed Saga debugging becomes easier.

### 14.4. Alerting

Foundation alerts:

```text
PodPendingTooLong
CrashLoopBackOff
DeploymentUnavailable
NodePressure
TempoUnavailable
ArgoCDAppOutOfSync
ArgoCDAppNotHealthy
```

Incident-driven alerts:

```text
KafkaConnectDeploymentUnavailable
KafkaConnectConnectorHealthProbeFailed
KafkaConnectBootstrapJobFailed
KafkaConnectCronJobSuspended
SagaPendingOrderProbeFailed
SagaRuntimeDeploymentUnavailable
SagaRuntimePodRestarting
```

---

## 15. Testing, Benchmarking, and Chaos Engineering

### 15.1. Test Categories

| Category         | Purpose                                          |
| ---------------- | ------------------------------------------------ |
| Smoke test       | Confirms the end-to-end Saga works after changes |
| Baseline k6 test | Measures stable capacity under controlled RPS    |
| Load test        | Simulates normal and elevated traffic            |
| Spike test       | Simulates sudden traffic bursts                  |
| Stress test      | Pushes the system into bottleneck conditions     |
| Soak test        | Checks long-running stability                    |
| Idempotency test | Confirms duplicate request protection            |
| Flash-sale test  | Validates Redis atomic stock gate                |
| Chaos test       | Validates resilience under controlled failures   |

### 15.2. Capacity Interpretation

In the current lab environment, the system has confirmed stable baseline capacity up to **80 RPS** for the order processing flow.

Interpretation:

* 80 RPS is the last confirmed stable baseline rating for the current fixed-resource cluster.
* It is not a claim that the architecture can never go beyond 80 RPS.
* Heavier spike/stress scenarios showed increased latency, asynchronous backlog, and database connection pressure.
* Therefore, 80 RPS is used as the conservative fixed-resource rating.

### 15.3. Observed Bottlenecks

Under heavier tests, the main observed pressure points were:

1. Kafka backlog at the inventory consumer stage;
2. PostgreSQL connection pressure in some background jobs;
3. asynchronous Saga completion delay under high burst load;
4. dependency on Kafka partitioning and consumer parallelism;
5. inventory hot-key/hot-row pressure during flash-sale traffic.

### 15.4. Chaos Engineering

Chaos experiments validate resilience, not maximum throughput.

Typical experiments:

* killing a payment API pod;
* adding CPU stress to inventory API;
* injecting network delay between order-service and Kafka.

Purpose:

* verify Kubernetes self-healing;
* observe service behavior during partial failure;
* validate alerting and monitoring;
* confirm recovery after cleanup.

---

## 16. Repository Structure

```text
.
├── docs/
│   ├── benchmark/
│   ├── cdc/
│   ├── chaos/
│   ├── checklists/
│   ├── clickhouse/
│   ├── dynamic-filter/
│   ├── evidence/
│   ├── observability/
│   ├── opentelemetry/
│   ├── report/
│   ├── runbook/
│   └── security/
├── docker/
│   └── kafka-connect-dynamic-filter/
├── k8s/
│   ├── analytics/
│   ├── argocd/
│   ├── cdc/
│   ├── db/
│   ├── istio/
│   ├── kafka/
│   ├── mongodb/
│   ├── monitoring/
│   ├── observability/
│   ├── redis/
│   ├── security/
│   └── services/
├── scripts/
│   ├── backup/
│   └── flash-sale/
├── services/
│   ├── ecommerce-dashboard/
│   ├── inventory-service/
│   ├── notification-service/
│   ├── order-service/
│   ├── payment-service/
│   ├── read-model-service/
│   └── web-gateway/
└── tests/
    ├── chaos/
    ├── k6/
    └── smoke/
```

---

## 17. Common Operations

### 17.1. Check Git and Runtime Status

```bash
git status --short
git log --oneline --decorate -10

kubectl -n argocd get applications.argoproj.io \
  -o custom-columns=NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision

kubectl get pods -A | grep -Ev 'Running|Completed|STATUS' || true
```

### 17.2. Load API Key Safely

```bash
export API_KEY="$(
  kubectl -n default get secret ecommerce-runtime-secrets \
    -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
)"

echo "API_KEY loaded; length=${#API_KEY}"
```

### 17.3. Run Saga Smoke Test

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"

./tests/smoke/saga-success.sh
```

Expected result:

```text
orders.status = COMPLETED
outbox.status = PUBLISHED
inventory_reservations.status = RESERVED
payments.status = COMPLETED
notifications.status = SENT
SMOKE TEST PASSED
```

### 17.4. Reset Benchmark Environment

Use only when a clean benchmark environment is required.

```bash
CONFIRM_RESET=YES ./tests/k6/reset.sh
```

### 17.5. Run a Baseline k6 Test

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"

RUN_ID="baseline-80rps-$(date +%Y%m%d%H%M%S)"

API_KEY="$API_KEY" \
GATEWAY_URL="$GATEWAY_URL" \
RUN_ID="$RUN_ID" \
k6 run tests/k6/baseline-e2e-80rps.js
```

### 17.6. Check Kafka Consumer Lag

```bash
kubectl -n kafka exec kafka-0 -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group inventory-service-group
```

### 17.7. Check Kafka Connect Connectors

```bash
kubectl -n cdc exec deploy/kafka-connect-debezium -- \
  curl -sS http://localhost:8083/connectors | jq

kubectl -n cdc exec deploy/kafka-connect-debezium -- \
  curl -sS http://localhost:8083/connectors/order-db-orders-connector/status | jq

kubectl -n cdc exec deploy/kafka-connect-debezium -- \
  curl -sS http://localhost:8083/connectors/order-db-orders-dynamic-filter-connector/status | jq
```

### 17.8. Check Redis Dynamic Filter Rule

```bash
REDIS_POD="$(kubectl -n default get pod -o name | grep -i redis | head -1 | sed 's#pod/##')"

kubectl -n default exec "$REDIS_POD" -- \
  redis-cli GET filter:order-status
```

---

## 18. Evidence and Documentation Map

| Area                    | Important Files                                                     |
| ----------------------- | ------------------------------------------------------------------- |
| Benchmark               | `docs/benchmark/`                                                   |
| Final k6 suite          | `docs/benchmark/phase6-final-k6-suite-summary.md`                   |
| Raw k6 artifacts        | `docs/benchmark/k6-final-artifacts/`                                |
| CDC                     | `docs/cdc/`, `docs/evidence/cdc-gitops-bootstrap-recovery-proof.md` |
| Dynamic Filter          | `docs/dynamic-filter/`                                              |
| ClickHouse              | `docs/clickhouse/`                                                  |
| Saga recovery           | `docs/evidence/saga-consumer-recovery-proof.md`                     |
| Phase 7 alert hardening | `docs/evidence/phase7-incident-driven-alert-hardening-evidence.md`  |
| Chaos                   | `docs/chaos/phase6-controlled-chaos-suite-summary.md`               |
| Backup/restore          | `docs/runbook/postgres-backup-restore.md`                           |
| Runbooks                | `docs/runbook/`                                                     |
| Security                | `docs/security/`                                                    |
| Final report            | `docs/report/`                                                      |

---

## 19. Limitations and Future Work

This system is a production-oriented academic prototype, not a fully managed production platform.

Current limitations:

* cluster is limited to a lab/VM environment;
* PostgreSQL is still a single-primary transactional source;
* payment gateway is simulated/COD;
* notification is in-app/log based rather than real email/SMS delivery;
* Dynamic Filter currently focuses on the `orders` table and `status` field;
* Kafka consumer lag alerting can be improved with a dedicated Kafka exporter;
* secret management is based on Kubernetes Secrets rather than Vault or a cloud secret manager;
* database sharding and multi-region deployment are not implemented.

Future improvements:

* increase Kafka partitions and tune consumer parallelism;
* move all background database jobs through PgBouncer;
* add Kafka exporter for richer lag metrics;
* introduce database partitioning or sharding;
* integrate a real payment gateway;
* add external notification providers;
* add policy-as-code scanning;
* add SBOM/image vulnerability scanning;
* improve disaster recovery with automated backup verification;
* add canary/blue-green deployment with Argo Rollouts or Istio traffic shifting.

---

## 20. References

* Kubernetes Documentation: https://kubernetes.io/docs/
* Docker Documentation: https://docs.docker.com/
* Apache Kafka Documentation: https://kafka.apache.org/documentation/
* PostgreSQL Documentation: https://www.postgresql.org/docs/
* PgBouncer Documentation: https://www.pgbouncer.org/
* Redis Documentation: https://redis.io/docs/latest/
* MongoDB Documentation: https://www.mongodb.com/docs/
* ClickHouse Documentation: https://clickhouse.com/docs/
* Debezium Documentation: https://debezium.io/documentation/
* Kafka Connect Documentation: https://kafka.apache.org/documentation/#connect
* ArgoCD Documentation: https://argo-cd.readthedocs.io/
* KEDA Documentation: https://keda.sh/docs/
* Istio Documentation: https://istio.io/latest/docs/
* Prometheus Documentation: https://prometheus.io/docs/
* Grafana Documentation: https://grafana.com/docs/
* Grafana Loki Documentation: https://grafana.com/docs/loki/latest/
* Grafana Alloy Documentation: https://grafana.com/docs/alloy/latest/
* OpenTelemetry Documentation: https://opentelemetry.io/docs/
* Grafana Tempo Documentation: https://grafana.com/docs/tempo/latest/
* k6 Documentation: https://grafana.com/docs/k6/latest/
* Chaos Mesh Documentation: https://chaos-mesh.org/docs/
