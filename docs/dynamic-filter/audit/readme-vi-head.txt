# kafka-connect-dynamic-filter

<p align="center">
  <img src="./assets/logo.png" alt="kafka-connect-dynamic-filter logo" width="200"/>
</p>

[![Maven Central](https://img.shields.io/maven-central/v/io.github.caobahuong/kafka-connect-dynamic-filter-core?label=Maven%20Central)](https://central.sonatype.com/artifact/io.github.caobahuong/kafka-connect-dynamic-filter-core)
[![Java](https://img.shields.io/badge/Java-17%2B-orange?logo=openjdk)](https://adoptium.net/)
[![Kafka Connect](https://img.shields.io/badge/Kafka_Connect-3.x-blue?logo=apachekafka)](https://kafka.apache.org/documentation/#connect)
[![License](https://img.shields.io/badge/license-Apache_2.0-blue)](LICENSE.txt)

[🇺🇸 English](./README.md) · 🇻🇳 Tiếng Việt · [🇨🇳 中文](./README.zh.md) · [🇯🇵 日本語](./README.ja.md)

---

**Một SMT cho Kafka Connect giúp lọc bản ghi CDC từ Debezium theo điều kiện động** thay đổi rule bất cứ lúc nào mà không cần restart connector.

> Tài liệu này thiên về giới thiệu tính năng và cách cài đặt. Nếu bạn muốn đọc dạng bài blog có triển khai step-by-step, xem tại: [Hướng dẫn triển khai kafka-connect-dynamic-filter](https://github.com/caobahuong/kafka-connect-dynamic-filter/tree/main/blogs/vi)

---

## Tại Sao Dùng Thư Viện Này?

SMT `Filter` tích hợp sẵn của Debezium yêu cầu script Groovy/JS tĩnh viết cứng vào config muốn thay đổi phải restart connector, gây ra sự bất tiện và thiếu linh động trong việc filter.

Ví dụ: Trước đây chỉ có các bản ghi có `ref_id` là 1, 2, 3 được bắn lên Kafka topic. Nay cần thêm `ref_id` 4, 5 với Debezium Filter bạn phải sửa config, deploy lại, restart connector và chịu một khoảng thời gian pipeline bị gián đoạn.

Với thư viện này, bạn chỉ cần cập nhật rule trên Redis, Kafka topic, hoặc file JSON connector tự nhận ngay ở bản ghi tiếp theo, không cần đụng vào config, không restart, không downtime.

| | Debezium Filter (tích hợp sẵn) | kafka-connect-dynamic-filter |
|---|---|---|
| Cập nhật rule | ❌ Phải restart connector | ✅ Hiệu lực ngay bản ghi tiếp theo |
| Ngôn ngữ rule | Groovy / JS hardcode | JSON động |
| Nguồn rule | Config file | Redis, Kafka topic, File, hoặc tự custom |
| Điều kiện phức hợp | Viết code | AND / OR / lồng nhau qua JSON |

---

## Các Module

> Dự án cung cấp thư viện `core` để bạn tự custom. <br>
> Các module dưới đây là implement phổ biến do tác giả cung cấp sẵn.

| Module | Nơi lưu quy tắc |
|---|---|
| [**core**](./kafka-connect-dynamic-filter-core/) | Thư viện nền tự implement nguồn rule |
| [**redis**](./kafka-connect-dynamic-filter-redis/) | Lấy rule từ Redis key, hỗ trợ Keyspace Notifications + polling |
| [**kafka**](./kafka-connect-dynamic-filter-kafka/) | Lấy rule từ Kafka topic |
| [**file**](./kafka-connect-dynamic-filter-file/) | Lấy rule từ file JSON, tự reload khi file thay đổi |

---

## Mục Lục

- [Cài Đặt](#cài-đặt)
- [Hướng Dẫn Nhanh](#hướng-dẫn-nhanh)
- [Cú Pháp Rule](#cú-pháp-rule)
- [Nguồn Rule](#nguồn-rule)
- [Cách Bản Ghi Được Lọc](#cách-bản-ghi-được-lọc)
- [Metrics và JMX](#metrics-và-jmx)
- [Debug & Troubleshooting](#debug--troubleshooting)
- [Cấu Hình Chi Tiết](#cấu-hình-chi-tiết)
- [Kiểm Thử](#kiểm-thử)
- [Đóng Góp](#đóng-góp)

---

## Cài Đặt

### Bước 1: Tải JAR và copy vào thư mục plugin

Mỗi module là một **fat JAR** độc lập core và toàn bộ dependency đã được đóng gói sẵn bên trong, không cần cài thêm gì.

Tải JAR mới nhất từ trang [Releases](https://github.com/caobahuong/kafka-connect-dynamic-filter/releases), sau đó:

```bash
PLUGIN_DIR=$KAFKA_CONNECT_PLUGINS_DIR/kafka-connect-dynamic-filter
mkdir -p $PLUGIN_DIR

# Chọn đúng một module
cp kafka-connect-dynamic-filter-redis-*.jar  $PLUGIN_DIR/
# hoặc: kafka-connect-dynamic-filter-kafka-*.jar
# hoặc: kafka-connect-dynamic-filter-file-*.jar
```

> **Chưa biết plugin dir nằm ở đâu?** Chạy lệnh sau để tìm nhanh:
> ```bash
> find /opt /usr /etc -iname "*kafka*" 2>/dev/null | grep plugin
> ```
> Thường sẽ nằm ở `/opt/kafka-connect/plugins/`

Sau khi copy xong, khởi động lại Kafka Connect workers:

```bash
# systemd
systemctl restart kafka-connect

# hoặc nếu dùng Confluent Platform
confluent local services connect restart
```

> **Nếu dùng Docker**
>
> Copy JAR vào container đang chạy rồi restart:
> ```bash
> docker cp kafka-connect-dynamic-filter-redis-*.jar <container_name>:/opt/kafka-connect/plugins/
> docker restart <container_name>
> ```
> Tuy nhiên cách trên sẽ mất JAR sau mỗi lần recreate container. Nên mount volume trong `docker-compose.yml` để JAR tồn tại lâu dài:
> ```yaml
> services:
>   kafka-connect:
>     volumes:
>       - ./plugins:/opt/kafka-connect/plugins
> ```
> Sau đó copy JAR vào thư mục `./plugins/` trên host và `docker compose restart kafka-connect` là xong.

### Bước 2 Khai báo SMT trong config connector
**Thông qua Kafka Connect REST API:
```bash
curl -X PUT http://localhost:8083/connectors/my-debezium-connector/config \
  -H "Content-Type: application/json" \
  -d '{
    ... other config 
    "connector.class": "io.debezium.connector.mysql.MySqlConnector",
    "transforms": "dynamicFilter",
    "transforms.dynamicFilter.type": "io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter",
    "transforms.dynamicFilter.filter.id": "company-filter",
    "transforms.dynamicFilter.field.name": "com_id",
    "transforms.dynamicFilter.redis.uri": "redis://localhost:6379",
    "transforms.dynamicFilter.redis.key": "filter:companies"
    ... 
  }'
```

Ví dụ:
```bash
curl -X PUT http://localhost:8083/connectors/my-debezium-connector/config \
  -H "Content-Type: application/json" \
  -d '{
    ... other config 
    "transforms": "filterComId",
    "transforms.filterComId.type": "io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter",
    "transforms.filterComId.filter.id": "company-filter",
    "transforms.filterComId.field.name": "com_id",
    "transforms.filterComId.empty.list.behavior": "pass_all",
    "transforms.filterComId.redis.uri": "redis://<IP_ADDRESS>:6379/1",
    "transforms.filterComId.redis.key": "REDIS_CONFIG_RULE_KEY"
    ... 
  }'
```

```bash
# Kiểm tra connector đã nhận config chưa
curl http://localhost:8083/connectors/my-debezium-connector/status
```

| Module | Giá trị `type` |
|---|---|
| Redis | `io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter` |
| Kafka | `io.kafkaconnect.dynamicfilter.kafka.KafkaDynamicFilter` |
| File | `io.kafkaconnect.dynamicfilter.file.FileDynamicFilter` |
| Core (custom) | `io.kafkaconnect.dynamicfilter.DynamicListFilter` |

> Xem config đầy đủ từng module ở phần [Cấu Hình Chi Tiết](#cấu-hình-chi-tiết).

### Dùng core như thư viện (Maven / Gradle)

Nếu bạn tự implement nguồn rule thông qua Java API:

```xml
<dependency>
  <groupId>io.kafkaconnect</groupId>
  <artifactId>kafka-connect-dynamic-filter-core</artifactId>
  <version>0.1.0</version>
</dependency>
```

```groovy
implementation 'io.kafkaconnect:kafka-connect-dynamic-filter-core:0.1.0'
```

### Build từ source

```bash
git clone https://github.com/caobahuong/kafka-connect-dynamic-filter.git
cd kafka-connect-dynamic-filter
./mvnw clean package -DskipTests
```

---

## Hướng Dẫn Nhanh

Giả sử bạn đã copy JAR và restart workers (Bước 1 ở trên). Dưới đây là ví dụ cụ thể cho từng module.

### Redis

Thêm vào config connector:

```properties
transforms=dynamicFilter
transforms.dynamicFilter.type=io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter
transforms.dynamicFilter.filter.id=company-filter
transforms.dynamicFilter.field.name=com_id
transforms.dynamicFilter.redis.uri=redis://localhost:6379
transforms.dynamicFilter.redis.key=filter:companies
transforms.dynamicFilter.empty.list.behavior=pass_all
```

Điều khiển filter từ bất kỳ đâu, bất kỳ lúc nào không restart:

```bash
# Chỉ cho qua công ty 1, 2, 3
redis-cli SET filter:companies '[1, 2, 3]'

# Thêm công ty 4 có hiệu lực ngay bản ghi tiếp theo
redis-cli SET filter:companies '[1, 2, 3, 4]'

# Nhiều điều kiện: công ty được phép + đang hoạt động + chưa xoá
redis-cli SET filter:companies '{
  "op": "AND",
  "conditions": [
    {"field": "com_id",     "op": "IN",      "values": [1, 2, 3]},
    {"field": "status",     "op": "EQ",      "values": ["ACTIVE"]},
    {"field": "deleted_at", "op": "IS_NULL"}
  ]
}'

# Cho tất cả qua
redis-cli DEL filter:companies
```

> **Khuyến nghị:** Nên bật `Keyspace Notifications` để filter nhận rule mới **tức thì** thay vì chờ hết chu kỳ polling (mặc định 5 giây).
>
> Bật trên Redis server:
> ```bash
> redis-cli CONFIG SET notify-keyspace-events KEA
> ```
> Bật trong connector config:
> ```properties
> transforms.dynamicFilter.redis.keyspace.notifications.enabled=true
> transforms.dynamicFilter.redis.keyspace.db.index=0
> ```
> Xem thêm chi tiết tại phần [Nguồn Rule → Redis](#redis-1).

### Kafka topic

Thêm vào config connector:

```properties
transforms=dynamicFilter
transforms.dynamicFilter.type=io.kafkaconnect.dynamicfilter.kafka.KafkaDynamicFilter
transforms.dynamicFilter.filter.id=company-filter
transforms.dynamicFilter.field.name=com_id
transforms.dynamicFilter.rule.bootstrap.servers=localhost:9092
transforms.dynamicFilter.rule.topic=filter-rules
```

Publish message JSON lên topic `filter-rules` để cập nhật rule. Publish tombstone (null value) để xoá rule.
