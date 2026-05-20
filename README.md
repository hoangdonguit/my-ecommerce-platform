# E-commerce Microservices Order Processing Platform

A cloud-native e-commerce order processing platform designed and implemented as a course project for **NT114 - Specialized Project**.

The system focuses on distributed order processing, event-driven communication, data consistency, autoscaling, observability, and GitOps-based deployment on Kubernetes.

> This project is not just a CRUD API demo. It demonstrates how an e-commerce order goes through multiple distributed steps such as order creation, inventory reservation, payment processing, notification delivery, read-model projection, and operational monitoring.

---

## 1. Project Information

| Item | Description |
|---|---|
| Course | NT114 - Specialized Project |
| Vietnamese title | Thiết kế và triển khai nền tảng xử lý đơn hàng thương mại điện tử theo kiến trúc Microservices Cloud-Native sử dụng Kubernetes và GitOps |
| English title | Design and Implementation of a Cloud-Native Microservices E-commerce Order Processing Platform using Kubernetes and GitOps |
| Supervisor | MSc. Lê Anh Tuấn |

## 2. Team Members

| No. | Full name | Student ID |
|---:|---|---|
| 1 | Hoàng Xuân Đồng | 23520297 |
| 2 | Đỗ Thái Hậu | 23520450 |

---

## 3. Project Goals

The project simulates a distributed e-commerce order processing platform. Instead of simply storing orders in a database, each order is processed through multiple services and asynchronous events.

Main goals:

- Implement distributed transaction handling with **Saga Choreography**.
- Use **Apache Kafka** for asynchronous event-driven communication.
- Prevent event loss with the **Transactional Outbox** pattern.
- Prevent duplicate orders with **Redis-based Idempotency**.
- Support high-concurrency flash sale scenarios with a **Redis Atomic Stock Gate**.
- Autoscale consumers based on **Kafka lag using KEDA**.
- Optimize PostgreSQL connections with **PgBouncer**.
- Improve read performance with **MongoDB CQRS Read Model**.
- Reduce repeated read load with **Redis Cache-Aside**.
- Add lightweight **Retry and Dead Letter Queue** handling for Kafka consumers.
- Validate the system using **k6 smoke, load, stress, spike, soak, idempotency, and flash sale tests**.
- Observe the system using **Prometheus, Grafana, and Istio dashboards**.
- Manage Kubernetes deployment using **GitOps with ArgoCD**.
- Add basic CI validation using **GitHub Actions**.

---

## 4. High-Level Architecture

### 4.1. Main Order Processing Flow

```text
Client / Dashboard / k6
        |
        v
Web Gateway
        |
        v
Order Service
        |
        v
PostgreSQL order_db
        |
        v
Transactional Outbox
        |
        v
Kafka topic: order.created
        |
        v
Inventory Consumer
        |
        v
Kafka topic: inventory.reserved / inventory.failed
        |
        v
Payment Consumer
        |
        v
Kafka topic: payment.completed / payment.failed
        |
        +-----------------------------+
        |                             |
        v                             v
Order Saga Monitor              Notification Consumer
        |                             |
        v                             v
Update order status             notification_db
```

### 4.2. CQRS Read Model Flow

```text
Kafka topic: payment.completed
        |
        v
read-model-service
        |
        v
MongoDB order_read_models
        |
        v
Web Gateway
        |
        v
Dashboard MongoDB Read Model Page
```

### 4.3. Flash Sale Flow

```text
Client / k6 Flash Sale Test
        |
        v
Web Gateway
        |
        v
Order Service
        |
        v
Redis Atomic Stock Gate
        |
        +-----------------------------+
        |                             |
        v                             v
Accept order                  Reject when sold out
        |
        v
Saga flow through Kafka
```

The Flash Sale mode uses Redis as a fast stock gate before creating an order. This prevents excessive requests from overloading PostgreSQL and Kafka when the product stock has already been exhausted.

---

## 5. Main Components

| Category | Component | Responsibility |
|---|---|---|
| API Gateway | Web Gateway | API entry point, X-API-Key validation, read API caching |
| Microservice | Order Service | Order creation, outbox writing, Saga status tracking |
| Microservice | Inventory Service | Product stock management, reservation, stock release |
| Microservice | Payment Service | Simulated COD/payment processing |
| Microservice | Notification Service | Notification creation after payment events |
| Read Model | read-model-service | Consumes Kafka events and writes MongoDB read models |
| Dashboard | ecommerce-dashboard | UI for order lookup, failed order monitoring, and read model viewing |
| Database | PostgreSQL | Source of truth for transactional data |
| Connection Pooling | PgBouncer | Reduces direct PostgreSQL connection pressure |
| Messaging | Apache Kafka | Event streaming backbone for Saga communication |
| Cache | Redis | Idempotency, read cache, and flash sale stock gate |
| Document Database | MongoDB | CQRS read model storage |
| Autoscaling | KEDA | Scales API and consumers, especially based on Kafka lag |
| Service Mesh | Istio | Gateway, traffic management, and request observability |
| GitOps | ArgoCD | Synchronizes Kubernetes manifests from GitHub |
| Observability | Prometheus / Grafana / Istio | Monitors pods, services, latency, success rate, and traffic |
| Testing | k6 | Smoke, load, stress, spike, soak, idempotency, and flash sale tests |
| CI | GitHub Actions | Basic secret scan, syntax check, build, and test workflow |

---

## 6. Reliability Mechanisms

### 6.1. Database per Service

Each service owns its own database:

- `order_db`
- `inventory_db`
- `payment_db`
- `notification_db`

This reduces coupling between services and follows the microservices database ownership principle.

### 6.2. Transactional Outbox

Order Service does not directly publish Kafka events immediately after creating an order. Instead, it stores the order and outbox event in the same database transaction.

```text
Create order + insert outbox event
        |
        v
Outbox worker scans PENDING events
        |
        v
Publish event to Kafka
        |
        v
Mark outbox event as PUBLISHED
```

This reduces the risk of saving an order without publishing its corresponding event.

### 6.3. Saga Choreography

Successful flow:

```text
order.created
    -> inventory.reserved
    -> payment.completed
    -> notification sent
    -> order COMPLETED
```

Failure or compensation flow:

```text
inventory.failed / payment.failed
    -> order FAILED or CANCELLED
    -> release stock when needed
```

### 6.4. Redis Idempotency

Redis is used to support idempotency keys. If the client retries the same order request, the system can avoid creating duplicate orders.

### 6.5. Flash Sale Stock Gate

The Flash Sale mode uses Redis to protect the system during high-concurrency product ordering.

Characteristics:

- Stock is initialized before the test.
- Each order request checks the Redis atomic stock gate.
- Only requests within available stock are accepted.
- Sold-out requests are rejected early.
- PostgreSQL and Kafka only receive accepted orders.

Related scripts:

```bash
./scripts/flash-sale/init-stock.sh
k6 run tests/k6/flash-sale-test.js
k6 run tests/k6/flash-sale-spike-test.js
```

### 6.6. PgBouncer

Services connect to PostgreSQL through PgBouncer instead of opening too many direct database connections. This improves stability when many pods are running or when benchmark tests are executed.

### 6.7. KEDA Kafka Lag Autoscaling

Kafka consumers can scale based on consumer lag instead of only CPU usage. When backlog increases, KEDA increases the number of consumer replicas to process messages in parallel.

### 6.8. MongoDB CQRS Read Model

PostgreSQL remains the source of truth. MongoDB is used as a read model for faster dashboard queries.

```text
Kafka payment.completed
    -> read-model-service
    -> MongoDB order_read_models
    -> Web Gateway
    -> Dashboard MongoDB Read Model
```

### 6.9. Redis Cache-Aside

Web Gateway caches selected read APIs:

- `GET /api/read-model/orders`
- `GET /api/read-model/orders/:id`
- `GET /api/inventories`

Current TTL: 5 seconds.

Expected behavior:

- First request: `X-Cache: MISS`
- Repeated request: `X-Cache: HIT`

### 6.10. Kafka Retry and DLQ

Lightweight retry and Dead Letter Queue handling has been added to important Kafka consumers.

| Consumer | Input topic | DLQ topic |
|---|---|---|
| payment-consumer | inventory.reserved | inventory.reserved.dlq |
| notification-consumer | payment.completed | payment.completed.dlq |
| notification-consumer | payment.failed | payment.failed.dlq |

Processing logic:

```text
FetchMessage
    -> process business logic
    -> retry up to 3 times on error
    -> publish to .dlq if it still fails
    -> commit offset to avoid blocking the consumer group
```

In a successful scenario, DLQ topics are expected to be empty.

### 6.11. PostgreSQL Backup and Restore Check

The repository includes backup and restore-check scripts:

```bash
./scripts/backup/postgres-backup.sh
./scripts/backup/postgres-restore-check.sh backups/postgres/<timestamp>/order_db.dump
```

The restore check creates a temporary database, restores a dump file, verifies tables and row counts, and then removes the temporary database. It does not overwrite production data.

---

## 7. Kubernetes and GitOps

The system is deployed on Kubernetes.

Main manifest groups:

```text
k8s/
  db/
  istio/
  kafka/
  mongodb/
  monitoring/
  redis/
  services/
```

ArgoCD manages the following applications:

- `ecommerce-infrastructure`
- `ecommerce-platform`
- `infrastructure-layer`

Expected state:

- ArgoCD: `Synced / Healthy`
- default namespace pods: `Running`
- db namespace pods: `Running`
- kafka namespace pods: `Running`

---

## 8. Dashboard

The dashboard includes:

- Demo Store
- System Overview
- Order Lookup
- Failed Order Monitoring
- MongoDB Read Model

The MongoDB Read Model page demonstrates the flow:

```text
Kafka event
    -> read-model-service
    -> MongoDB
    -> Web Gateway
    -> Dashboard
```

---

## 9. GitHub Actions CI

The CI workflow currently checks:

- Basic old secret and cluster IP scan.
- Bash syntax.
- k6 JavaScript syntax.
- Go test/build for services.
- Dashboard npm build.

The workflow is expected to run on push and pull request events.

---

## 10. Main Directory Structure

```text
.
├── docs/
│   ├── benchmark/
│   ├── evidence/
│   ├── report/
│   ├── runbook/
│   └── security/
├── k8s/
│   ├── db/
│   ├── istio/
│   ├── kafka/
│   ├── mongodb/
│   ├── monitoring/
│   ├── redis/
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

## 11. Environment Requirements

The machine running this project should have:

- Docker
- kubectl
- k6
- jq
- Node.js and npm
- Go
- Access to a Kubernetes cluster
- GitHub and container registry access when building or pushing images

---

## 12. Environment Variables

### 12.1. Gateway URL

Do not hardcode real cluster IPs in source code. Use an environment variable when running tests or API checks.

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"
```

For local development:

```bash
export GATEWAY_URL="http://localhost:8090"
```

### 12.2. API Key

The API key is stored in a Kubernetes Secret.

```bash
export API_KEY="$(
  kubectl -n default get secret ecommerce-runtime-secrets \
    -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
)"
```

### 12.3. Dashboard Environment File

Create a local `.env` file from the example:

```bash
cd services/ecommerce-dashboard
cp .env.example .env
```

The `.env` file is for local development only and must not be committed.

---

## 13. Quick Commands

### 13.1. Check Git, ArgoCD, and Pods

```bash
cd /home/xuandong/Doanchuyennganh/my-ecommerce-platform

git status --short
git log --oneline --decorate -5

kubectl -n argocd get applications -o wide

kubectl get pods -n default
kubectl get pods -n db
kubectl get pods -n kafka
```

### 13.2. Run Saga Smoke Test

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"
export API_KEY="$(
  kubectl -n default get secret ecommerce-runtime-secrets \
    -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
)"

./tests/smoke/saga-success.sh
```

Expected result:

- `orders.status = COMPLETED`
- `outbox.status = PUBLISHED`
- `inventory_reservations.status = RESERVED`
- `payments.status = COMPLETED`
- `notifications.status = SENT`
- `SMOKE TEST PASSED`

### 13.3. Reset Benchmark Environment

This command clears benchmark data, resets Redis, resets Kafka topics, clears MongoDB read model data, and restarts services. Use it only when a clean benchmark environment is needed.

```bash
CONFIRM_RESET=YES ./tests/k6/reset.sh
```

### 13.4. Run Load Test

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"
export API_KEY="$(
  kubectl -n default get secret ecommerce-runtime-secrets \
    -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
)"

RUN_ID="$(date +%Y%m%d%H%M%S)"
mkdir -p docs/benchmark/runs

K6_NO_COLOR=1 k6 run --quiet tests/k6/load-test.js \
  | tee "docs/benchmark/runs/load-after-upgrade-${RUN_ID}.log"
```

### 13.5. Run Flash Sale Test

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"
export API_KEY="$(
  kubectl -n default get secret ecommerce-runtime-secrets \
    -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
)"

./scripts/flash-sale/init-stock.sh

RUN_ID="$(date +%Y%m%d%H%M%S)"
mkdir -p docs/benchmark/runs

K6_NO_COLOR=1 k6 run --quiet tests/k6/flash-sale-test.js \
  | tee "docs/benchmark/runs/flash-sale-${RUN_ID}.log"
```

### 13.6. Run Flash Sale Spike Test

```bash
export GATEWAY_URL="http://<GATEWAY_HOST>:<NODE_PORT>"
export API_KEY="$(
  kubectl -n default get secret ecommerce-runtime-secrets \
    -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
)"

./scripts/flash-sale/init-stock.sh

RUN_ID="$(date +%Y%m%d%H%M%S)"
mkdir -p docs/benchmark/runs

K6_NO_COLOR=1 k6 run --quiet tests/k6/flash-sale-spike-test.js \
  | tee "docs/benchmark/runs/flash-sale-spike-${RUN_ID}.log"
```

### 13.7. Monitor HPA and Pods

```bash
watch -n 2 'kubectl get hpa,pods -n default'
```

### 13.8. Monitor Kafka Consumer Lag

```bash
watch -n 5 'kubectl -n kafka exec kafka-0 -- kafka-consumer-groups.sh --bootstrap-server kafka.kafka.svc.cluster.local:9092 --describe --group payment-service-group | head -30'
```

### 13.9. Summarize Database Status After Benchmark

```bash
kubectl -n db exec postgresql-0 -- bash -lc '
export PGPASSWORD="$(cat /opt/bitnami/postgresql/secrets/postgres-password)"

echo "--- orders by status ---"
psql -U postgres -d order_db -c "
SELECT status, COUNT(*) FROM orders GROUP BY status ORDER BY status;
"

echo "--- outbox by status ---"
psql -U postgres -d order_db -c "
SELECT status, COUNT(*) FROM outbox GROUP BY status ORDER BY status;
"

echo "--- payments by status ---"
psql -U postgres -d payment_db -c "
SELECT status, COUNT(*) FROM payments GROUP BY status ORDER BY status;
"

echo "--- notifications by status ---"
psql -U postgres -d notification_db -c "
SELECT status, COUNT(*) FROM notifications GROUP BY status ORDER BY status;
"

echo "--- inventory reservations by status ---"
psql -U postgres -d inventory_db -c "
SELECT status, COUNT(*) FROM inventory_reservations GROUP BY status ORDER BY status;
"
'
```

### 13.10. Check MongoDB Read Model Through Gateway

```bash
curl -sS --max-time 10 \
  -H "X-API-Key: $API_KEY" \
  "${GATEWAY_URL}/api/read-model/orders?limit=3" | jq .
```

### 13.11. Check Redis Read Cache

```bash
curl -sS -D - -o /dev/null \
  -H "X-API-Key: $API_KEY" \
  "${GATEWAY_URL}/api/read-model/orders?limit=100" | grep -i "x-cache\|http"

curl -sS -D - -o /dev/null \
  -H "X-API-Key: $API_KEY" \
  "${GATEWAY_URL}/api/read-model/orders?limit=100" | grep -i "x-cache\|http"
```

Expected result:

- First request: `X-Cache: MISS`
- Second request: `X-Cache: HIT`

### 13.12. Check DLQ Topics

```bash
for topic in inventory.reserved.dlq payment.completed.dlq payment.failed.dlq; do
  echo "----- DLQ topic: $topic -----"
  kubectl -n kafka exec kafka-0 -- kafka-console-consumer.sh \
    --bootstrap-server localhost:9092 \
    --topic "$topic" \
    --from-beginning \
    --timeout-ms 3000 \
    --max-messages 3 || true
done
```

If the output contains `Processed a total of 0 messages`, the DLQ topic is empty, which is expected in a successful test scenario.

---

## 14. Benchmark Scenarios

| Script | Purpose |
|---|---|
| `tests/k6/smoke-test.js` | Basic API validation |
| `tests/k6/load-test.js` | Stable medium load |
| `tests/k6/stress-test.js` | Find system bottlenecks |
| `tests/k6/stress-test-multi.js` | Stress multiple endpoints |
| `tests/k6/spike-test.js` | Sudden traffic increase |
| `tests/k6/soak-test.js` | Long-running stability test |
| `tests/k6/idempotency-test.js` | Duplicate order prevention test |
| `tests/k6/flash-sale-test.js` | Flash sale stock gate validation |
| `tests/k6/flash-sale-spike-test.js` | High-concurrency flash sale test |

Recommended benchmark flow:

1. Check Git, ArgoCD, and pods.
2. Run smoke test.
3. Reset the environment using `CONFIRM_RESET=YES ./tests/k6/reset.sh`.
4. Open Grafana and monitor HPA/Kafka lag.
5. Run load, stress, spike, or flash sale test.
6. Summarize database status.
7. Check MongoDB read model, Redis cache, and DLQ.
8. Capture screenshots as evidence.
9. Record results in the before/after benchmark comparison table.

---

## 15. Security and Configuration Practices

Current practices:

- Do not commit `.env` files.
- Do not commit raw k6 result dumps in `tests/k6/results/`.
- Do not hardcode real cluster IPs in README or k6 scripts.
- Runtime secrets are injected through Kubernetes Secrets.
- Example manifest files only use placeholders.
- GitHub Actions includes a basic scan for old secrets and cluster IPs.

Planned improvements:

- Add detailed NetworkPolicy per namespace/service.
- Review and improve RBAC and ServiceAccounts.
- Add policy-as-code scanning using Checkov or Terrascan.
- Standardize environment management with Kustomize and/or Helm.
- Add canary or blue-green deployment workflow with Istio or Argo Rollouts.
- Evaluate OpenTelemetry, Vault, and Debezium as future enhancements.

---

## 16. Additional Documentation

Useful documents in this repository:

- `docs/benchmark/payment-throughput-tuning.md`
- `docs/benchmark/redis-read-cache.md`
- `docs/evidence/final-validation-after-upgrades.md`
- `docs/evidence/dlq-retry-validation.md`
- `docs/report/progress-after-12-05-upgrades.md`
- `docs/runbook/kafka-dlq-retry.md`
- `docs/runbook/postgres-backup-restore.md`

---

## 17. Current Status

Completed:

- Saga Choreography for distributed order processing.
- Transactional Outbox.
- Redis Idempotency.
- PgBouncer connection pooling.
- KEDA autoscaling for API and Kafka consumers.
- MongoDB CQRS Read Model.
- Redis Cache-Aside for read APIs.
- Kafka Retry and DLQ.
- PostgreSQL backup and restore-check scripts.
- Flash Sale Stock Gate.
- Dashboard for order lookup, failed order monitoring, and MongoDB read model.
- GitOps deployment with ArgoCD.
- Basic GitHub Actions CI.

Next tasks:

- Re-run benchmark after README/config cleanup.
- Build a before/after benchmark comparison table.
- Add detailed NetworkPolicy and RBAC manifests.
- Install and run Checkov or Terrascan for manifest scanning.
- Consider Istio Canary, OpenTelemetry, Vault, and Debezium as optional advanced extensions.
- Complete the final project report.
