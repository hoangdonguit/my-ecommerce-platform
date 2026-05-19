# E-commerce Microservices Order Processing Platform

Đồ án chuyên ngành NT114: Thiết kế và triển khai nền tảng xử lý đơn hàng thương mại điện tử theo kiến trúc Microservices Cloud-Native sử dụng Kubernetes và GitOps.

## 1. Thông tin đồ án

- Môn học: NT114 - Đồ án chuyên ngành
- Đề tài: Thiết kế và triển khai nền tảng xử lý đơn hàng thương mại điện tử theo kiến trúc Microservices Cloud-Native sử dụng Kubernetes và GitOps
- Tên tiếng Anh: Design and Implementation of a Cloud-Native Microservices E-commerce Order Processing Platform using Kubernetes and GitOps
- GVHD: ThS. Lê Anh Tuấn

## 2. Thành viên thực hiện

| STT | Họ và tên | MSSV |
|---:|---|---|
| 1 | Hoàng Xuân Đồng | 23520297 |
| 2 | Đỗ Thái Hậu | 23520450 |

## 3. Mục tiêu hệ thống

Hệ thống tập trung vào bài toán xử lý đơn hàng thương mại điện tử trong môi trường phân tán. Trọng tâm không chỉ là xây dựng API CRUD, mà còn là thiết kế một nền tảng có khả năng chịu tải, xử lý bất đồng bộ, đảm bảo tính nhất quán dữ liệu và có khả năng quan sát khi vận hành.

Các mục tiêu chính:

- Xử lý giao dịch phân tán bằng Saga Choreography.
- Tách tải xử lý bằng Apache Kafka.
- Đảm bảo idempotency khi client retry request.
- Tự động mở rộng consumer theo Kafka lag bằng KEDA.
- Tối ưu kết nối PostgreSQL bằng PgBouncer.
- Tách tải đọc bằng MongoDB CQRS Read Model.
- Giảm tải đọc lặp lại bằng Redis Cache-Aside.
- Bổ sung retry và Dead Letter Queue để tránh consumer bị kẹt bởi message lỗi.
- Kiểm thử hiệu năng bằng k6.
- Quan sát hệ thống bằng Grafana, Prometheus và Istio dashboard.
- Đồng bộ triển khai theo GitOps bằng ArgoCD.
- Kiểm tra chất lượng cơ bản bằng GitHub Actions CI.

## 4. Kiến trúc tổng quan

Luồng xử lý chính của hệ thống:

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

Luồng CQRS Read Model:

    Kafka payment.completed
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
    Dashboard MongoDB Read Model

## 5. Thành phần chính

| Nhóm | Thành phần | Vai trò |
|---|---|---|
| API Gateway | Web Gateway | Điểm vào API, xác thực X-API-Key, cache read API |
| Microservice | Order Service | Tạo đơn, lưu order, ghi outbox, theo dõi Saga |
| Microservice | Inventory Service | Quản lý tồn kho, giữ hàng, release stock |
| Microservice | Payment Service | Xử lý thanh toán giả lập/COD |
| Microservice | Notification Service | Tạo thông báo sau thanh toán |
| Read Model | read-model-service | Consume event và ghi dữ liệu đọc vào MongoDB |
| Dashboard | ecommerce-dashboard | UI quan sát đơn hàng, lỗi và MongoDB read model |
| Database | PostgreSQL | Source of truth cho dữ liệu giao dịch |
| Pooling | PgBouncer | Giảm connection trực tiếp tới PostgreSQL |
| Messaging | Apache Kafka | Event streaming cho Saga |
| Cache | Redis | Idempotency và read cache |
| Document DB | MongoDB | CQRS read model cho dashboard |
| Autoscaling | KEDA | Scale API/consumer, đặc biệt theo Kafka lag |
| GitOps | ArgoCD | Đồng bộ manifest từ GitHub về Kubernetes |
| Observability | Prometheus/Grafana/Istio | Theo dõi pod, service, latency, success rate |
| Testing | k6 | Smoke/load/stress/spike/soak/idempotency tests |
| CI | GitHub Actions | Test/build/check cơ bản khi push code |

## 6. Các cơ chế đảm bảo độ tin cậy

### 6.1. Database per Service

Mỗi service có database riêng:

- order_db
- inventory_db
- payment_db
- notification_db

Cách này giúp giảm coupling giữa các service và phù hợp với kiến trúc microservices.

### 6.2. Transactional Outbox

Order Service không publish event trực tiếp ngay sau khi tạo đơn. Thay vào đó, thao tác tạo đơn và ghi event outbox được thực hiện trong cùng transaction.

Luồng xử lý:

    Create order + insert outbox event
            |
            v
    Outbox Worker quét event PENDING
            |
            v
    Publish Kafka
            |
            v
    Mark event PUBLISHED

Mục tiêu là giảm rủi ro đơn hàng đã ghi vào PostgreSQL nhưng event không được gửi sang Kafka.

### 6.3. Saga Choreography

Luồng thành công:

    order.created
        -> inventory.reserved
        -> payment.completed
        -> notification sent
        -> order COMPLETED

Luồng lỗi hoặc bù trừ:

    inventory.failed / payment.failed
        -> order FAILED hoặc CANCELLED
        -> release stock nếu cần

### 6.4. Redis Idempotency

Redis được dùng để hỗ trợ idempotency key. Khi client retry request tạo đơn, hệ thống có thể tránh tạo trùng đơn hàng.

### 6.5. PgBouncer

Các service kết nối PostgreSQL thông qua PgBouncer thay vì mở quá nhiều connection trực tiếp tới PostgreSQL. Cơ chế này giúp ổn định hơn khi scale nhiều pod hoặc chạy benchmark.

### 6.6. KEDA Kafka Lag Autoscaling

Consumer không chỉ scale theo CPU mà scale theo Kafka consumer lag. Khi backlog tăng, KEDA tăng số lượng replica consumer để xử lý song song tốt hơn.

### 6.7. MongoDB CQRS Read Model

PostgreSQL vẫn là source of truth. MongoDB chỉ đóng vai trò read model phục vụ dashboard và truy vấn đọc nhanh.

Luồng đã triển khai:

    Kafka payment.completed
        -> read-model-service
        -> MongoDB order_read_models
        -> Web Gateway
        -> Dashboard MongoDB Read Model

### 6.8. Redis Cache-Aside

Web Gateway cache một số API đọc:

- GET /api/read-model/orders
- GET /api/read-model/orders/:id
- GET /api/inventories

TTL hiện tại: 5 giây.

Kết quả kiểm tra kỳ vọng:

- Request đầu: X-Cache: MISS
- Request sau: X-Cache: HIT

### 6.9. Kafka Retry / DLQ

Đã bổ sung retry và Dead Letter Queue mức nhẹ cho các consumer quan trọng.

| Consumer | Input topic | DLQ topic |
|---|---|---|
| payment-consumer | inventory.reserved | inventory.reserved.dlq |
| notification-consumer | payment.completed | payment.completed.dlq |
| notification-consumer | payment.failed | payment.failed.dlq |

Cơ chế:

    FetchMessage
        -> xử lý nghiệp vụ
        -> lỗi thì retry tối đa 3 lần
        -> vẫn lỗi thì publish sang .dlq
        -> commit offset để consumer không bị kẹt

Trong kịch bản thành công, DLQ rỗng là đúng kỳ vọng.

### 6.10. PostgreSQL Backup / Restore

Đã bổ sung script backup và restore-check:

    ./scripts/backup/postgres-backup.sh
    ./scripts/backup/postgres-restore-check.sh backups/postgres/<timestamp>/order_db.dump

Restore check tạo database tạm, restore file dump, kiểm tra bảng và số dòng, sau đó xóa database tạm. Script không ghi đè database thật.

## 7. Kubernetes / GitOps

Hệ thống được triển khai trên Kubernetes với các nhóm manifest chính:

    k8s/
      kafka/
      mongodb/
      services/
      ...

ArgoCD quản lý các application chính:

- ecommerce-infrastructure
- ecommerce-platform
- infrastructure-layer

Trạng thái kiểm tra gần nhất:

- ArgoCD: Synced / Healthy
- default namespace pods: Running
- db namespace pods: Running
- kafka namespace pods: Running

## 8. Dashboard

Dashboard hiện có các vùng chính:

- Cửa hàng Demo
- Tổng quan hệ thống
- Tra cứu đơn hàng
- Giám sát đơn lỗi
- MongoDB Read Model

Tab MongoDB Read Model dùng để chứng minh luồng:

    Kafka event
        -> read-model-service
        -> MongoDB
        -> Web Gateway
        -> Dashboard

## 9. GitHub Actions CI

Đã bổ sung workflow CI cơ bản.

CI hiện kiểm tra:

- Secret/IP cũ ở mức cơ bản.
- Bash syntax.
- k6 JavaScript syntax.
- Go test/build cho các service.
- Dashboard npm build.

Workflow đã chạy thành công trên GitHub Actions.

## 10. Cấu trúc thư mục chính

    .
    ├── docs/
    │   ├── benchmark/
    │   ├── evidence/
    │   ├── report/
    │   └── runbook/
    ├── k8s/
    │   ├── kafka/
    │   ├── mongodb/
    │   ├── services/
    │   └── ...
    ├── scripts/
    │   └── backup/
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

## 11. Yêu cầu môi trường

Máy chạy cần có:

- Docker
- kubectl
- k6
- jq
- Node.js/npm
- Go
- Quyền truy cập Kubernetes cluster
- Quyền push GitHub/Docker Hub nếu build image

## 12. Lệnh kiểm tra nhanh

### 12.1. Kiểm tra Git / ArgoCD / Pods

    cd /home/xuandong/Doanchuyennganh/my-ecommerce-platform

    git status --short
    git log --oneline --decorate -5

    kubectl -n argocd get applications -o wide

    kubectl get pods -n default
    kubectl get pods -n db
    kubectl get pods -n kafka

### 12.2. Nạp API key từ Kubernetes Secret

    export API_KEY="$(
      kubectl -n default get secret ecommerce-runtime-secrets        
       -o jsonpath='{.data.WEB_GATEWAY_API_KEY}' | base64 -d
    )"

### 12.3. Smoke test Saga end-to-end

    ./tests/smoke/saga-success.sh

Kỳ vọng:

- orders.status = COMPLETED
- outbox.status = PUBLISHED
- inventory_reservations.status = RESERVED
- payments.status = COMPLETED
- notifications.status = SENT
- SMOKE TEST PASSED

### 12.4. Reset môi trường benchmark

Lệnh này xóa dữ liệu test, reset Redis, reset Kafka topics và restart service. Chỉ chạy khi muốn benchmark sạch.

    CONFIRM_RESET=YES ./tests/k6/reset.sh

### 12.5. Chạy load test

    k6 run tests/k6/load-test.js

### 12.6. Theo dõi HPA/pods khi test

    watch -n 2 'kubectl get hpa,pods -n default'

### 12.7. Theo dõi Kafka lag

    watch -n 5 'kubectl -n kafka exec kafka-0 -- kafka-consumer-groups.sh --bootstrap-server kafka.kafka.svc.cluster.local:9092 --describe --group payment-service-group | head -30'

### 12.8. Tổng hợp trạng thái DB sau benchmark

    kubectl -n db exec postgresql-0 -- bash -lc "
    export PGPASSWORD="\$(cat /opt/bitnami/postgresql/secrets/postgres-password)"

    echo '--- orders by status ---'
    psql -U postgres -d order_db -c "
    SELECT status, COUNT(*) FROM orders GROUP BY status ORDER BY status;
    "

    echo '--- outbox by status ---'
    psql -U postgres -d order_db -c "
    SELECT status, COUNT(*) FROM outbox GROUP BY status ORDER BY status;
    "

    echo '--- payments by status ---'
    psql -U postgres -d payment_db -c "
    SELECT status, COUNT(*) FROM payments GROUP BY status ORDER BY status;
    "

    echo '--- notifications by status ---'
    psql -U postgres -d notification_db -c "
    SELECT status, COUNT(*) FROM notifications GROUP BY status ORDER BY status;
    "

    echo '--- inventory reservations by status ---'
    psql -U postgres -d inventory_db -c "
    SELECT status, COUNT(*) FROM inventory_reservations GROUP BY status ORDER BY status;
    "
    "

### 12.9. Kiểm tra MongoDB Read Model qua Gateway

    curl -sS --max-time 10       
      -H "X-API-Key: $API_KEY"       
      "http://100.65.255.2:32193/api/read-model/orders?limit=3" | jq .

### 12.10. Kiểm tra Redis cache

    curl -sS -D - -o /dev/null       
      -H "X-API-Key: $API_KEY"       
      "http://100.65.255.2:32193/api/read-model/orders?limit=100" | grep -i "x-cache\|http"

    curl -sS -D - -o /dev/null       
      -H "X-API-Key: $API_KEY"       
      "http://100.65.255.2:32193/api/read-model/orders?limit=100" | grep -i "x-cache\|http"

Kỳ vọng:

- Lần 1: X-Cache: MISS
- Lần 2: X-Cache: HIT

### 12.11. Kiểm tra DLQ

    for topic in inventory.reserved.dlq payment.completed.dlq payment.failed.dlq; do
      echo "----- DLQ topic: $topic -----"
      kubectl -n kafka exec kafka-0 -- kafka-console-consumer.sh         
        --bootstrap-server localhost:9092         
        --topic "$topic"         
        --from-beginning         
        --timeout-ms 3000         
        --max-messages 3 || true
    done

Nếu hiện `Processed a total of 0 messages` thì nghĩa là DLQ đang rỗng, đúng kỳ vọng trong kịch bản thành công.

## 13. Benchmark

Các kịch bản hiện có:

- tests/k6/smoke-test.js
- tests/k6/load-test.js
- tests/k6/stress-test.js
- tests/k6/stress-test-multi.js
- tests/k6/spike-test.js
- tests/k6/soak-test.js
- tests/k6/idempotency-test.js

Quy trình benchmark khuyến nghị:

1. Kiểm tra pod, ArgoCD và Git.
2. Chạy smoke test.
3. Nếu cần dữ liệu sạch, chạy: CONFIRM_RESET=YES ./tests/k6/reset.sh
4. Mở Grafana và watch HPA/Kafka lag.
5. Chạy load test.
6. Tổng hợp DB status.
7. Kiểm tra MongoDB read model, Redis cache và DLQ.
8. Chụp ảnh minh chứng.

## 14. Tài liệu bổ sung

Một số tài liệu trong repo:

- docs/benchmark/payment-throughput-tuning.md
- docs/benchmark/redis-read-cache.md
- docs/evidence/final-validation-after-upgrades.md
- docs/evidence/dlq-retry-validation.md
- docs/report/progress-after-12-05-upgrades.md
- docs/runbook/kafka-dlq-retry.md
- docs/runbook/postgres-backup-restore.md

## 15. Trạng thái hiện tại

- GitHub Actions CI: Passed.
- ArgoCD: Synced / Healthy.
- Smoke test: Passed trước khi reset benchmark.
- Benchmark environment: đã reset sạch, sẵn sàng chạy k6 load/stress test tiếp theo.

Việc còn lại gần nhất:

- Chạy load test sau reset.
- Tổng hợp bảng benchmark.
- Chờ máy Hậu online nếu muốn join cluster và tách App/Data node.
- Hoàn thiện báo cáo chính thức.
