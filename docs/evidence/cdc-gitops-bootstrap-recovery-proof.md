# CDC GitOps Bootstrap Recovery Proof

## 1. Mục tiêu

Tài liệu này ghi nhận quá trình phục hồi và GitOps hóa cơ chế bootstrap Kafka Connect connectors cho Debezium CDC và Dynamic Redis Filter.

Mục tiêu chính:

- Khôi phục Debezium CDC sau khi Kafka Connect mất connector runtime registration.
- Khôi phục Dynamic Redis Filter connector.
- Đưa connector bootstrap vào GitOps để tránh phụ thuộc thao tác thủ công.
- Ghi lại bằng chứng kiểm chứng phục hồi sau khi xóa connector runtime.

## 2. Bối cảnh lỗi

Sau một lần restart/sự cố cụm, Kafka Connect worker vẫn chạy nhưng danh sách connector bị rỗng.

Trạng thái lỗi ban đầu:

    curl http://localhost:8083/connectors
    []

Tác động:

- Luồng CDC không còn connector runtime.
- Topic CDC data không còn xuất hiện trong Kafka.
- Dynamic Redis Filter không còn hoạt động ở runtime.
- Các proof cũ vẫn tồn tại trong repository, nhưng runtime hiện tại đã mất connector registration.

Các connector cần phục hồi:

- order-db-orders-connector
- order-db-orders-dynamic-filter-connector

## 3. Phục hồi runtime CDC

Nhóm đã khôi phục normal Debezium connector:

- Connector: order-db-orders-connector
- Topic prefix: cdc.order_db
- Database: order_db
- Table: public.orders
- Slot: dbz_order_slot
- Publication: dbz_order_publication

Kết quả sau phục hồi:

    order-db-orders-connector: RUNNING
    task 0: RUNNING

## 4. Phục hồi Dynamic Redis Filter

Nhóm đã khôi phục Dynamic Filter connector:

- Connector: order-db-orders-dynamic-filter-connector
- Topic prefix: cdc_dynamic.order_db
- Database: order_db
- Table: public.orders
- Slot: dbz_order_dynamic_filter_slot
- Dynamic filter class: io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter
- Filter field: status
- Redis key: filter:order-status
- Redis rule hiện tại: ["PENDING", "COMPLETED"]

Ban đầu dynamic connector bị kẹt ở replication slot cũ:

    Cannot obtain valid replication slot 'dbz_order_dynamic_filter_slot'

Nhóm đã kiểm tra PostgreSQL replication slot và chỉ drop slot khi slot inactive.

Kết quả sau khi xử lý stale slot:

    order-db-orders-dynamic-filter-connector: RUNNING
    task 0: RUNNING

Log xác nhận Dynamic Filter hoạt động:

    Filter rule updated for 'order-status-filter': status IN [COMPLETED, PENDING]
    WorkerSourceTask{id=order-db-orders-dynamic-filter-connector-0} Source task finished initialization and start

## 5. Kafka topics sau khi tạo order event

Sau khi chạy light smoke để tạo order mới, Kafka xuất hiện lại các topic CDC data:

- cdc.order_db.public.orders
- cdc_dynamic.order_db.public.orders
- __debezium-heartbeat.cdc.order_db

Điều này xác nhận Debezium đã capture event từ PostgreSQL và publish ra Kafka topic CDC.

## 6. GitOps hardening

Đã bổ sung manifest:

    k8s/cdc/kafka-connect-connector-bootstrap.yaml

Manifest này tạo các tài nguyên:

- ServiceAccount: kafka-connect-connector-bootstrap
- Role trong namespace db: chỉ được get Secret postgresql
- RoleBinding: bind ServiceAccount cdc sang Role ở db
- ConfigMap: chứa connector definitions, không hardcode password thật
- CronJob: kafka-connect-connector-bootstrap
- Schedule: */10 * * * *

CronJob dùng cơ chế idempotent:

    PUT /connectors/{name}/config

Ý nghĩa:

- Nếu connector còn tồn tại, Job update lại config.
- Nếu connector bị mất, Job tạo lại connector.
- Password PostgreSQL được inject runtime từ Kubernetes Secret.
- Không commit password thật vào Git.

## 7. Recovery proof

Nhóm đã kiểm chứng bằng chaos test nhẹ.

Trạng thái trước test:

    [
      "order-db-orders-dynamic-filter-connector",
      "order-db-orders-connector"
    ]

Sau đó xóa cả hai connector runtime.

Trạng thái sau khi xóa:

    []

Chạy manual Job từ CronJob:

    APPLIED connector=order-db-orders-connector status=201
    APPLIED connector=order-db-orders-dynamic-filter-connector status=201
    CONNECTORS=["order-db-orders-dynamic-filter-connector", "order-db-orders-connector"]

Sau khi chờ connector khởi động:

    order-db-orders-connector: RUNNING / task 0 RUNNING
    order-db-orders-dynamic-filter-connector: RUNNING / task 0 RUNNING

Log xác nhận:

    WorkerSourceTask{id=order-db-orders-connector-0} Source task finished initialization and start
    WorkerSourceTask{id=order-db-orders-dynamic-filter-connector-0} Source task finished initialization and start
    Filter rule updated for 'order-status-filter': status IN [COMPLETED, PENDING]

## 8. Commit

Commit đã push lên GitHub:

    5d54f04 ops(cdc): gitops bootstrap kafka connect connectors

## 9. Giới hạn còn lại

CronJob không tự động drop PostgreSQL replication slot. Đây là chủ ý an toàn.

Lý do:

- Drop replication slot có thể làm mất vị trí CDC.
- Nếu slot còn active thì không được drop.
- Nếu slot stale/inactive thì cần xác minh bằng pg_replication_slots trước.

Runbook xử lý slot stale:

1. Kiểm tra pg_replication_slots.
2. Xác nhận slot dbz_order_dynamic_filter_slot có active = false.
3. Xóa dynamic connector đang lỗi.
4. Drop slot inactive.
5. Recreate connector bằng bootstrap Job.
6. Kiểm tra connector task trở về RUNNING.

## 10. Kết luận

CDC/Debezium và Dynamic Redis Filter đã được phục hồi và GitOps hóa ở mức connector bootstrap.

Rủi ro mất connector registration sau restart đã được giảm đáng kể nhờ CronJob idempotent chạy định kỳ. Trường hợp replication slot stale vẫn cần runbook vận hành riêng để xử lý an toàn.
