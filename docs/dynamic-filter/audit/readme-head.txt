# kafka-connect-dynamic-filter

<p align="center">
  <img src="./assets/logo.png" alt="kafka-connect-dynamic-filter logo" width="200"/>
</p>

[![CI](https://github.com/caobahuong/kafka-connect-dynamic-filter/actions/workflows/ci.yml/badge.svg)](https://github.com/caobahuong/kafka-connect-dynamic-filter/actions/workflows/ci.yml)
[![Maven Central](https://img.shields.io/maven-central/v/io.github.caobahuong/kafka-connect-dynamic-filter-core?label=Maven%20Central)](https://central.sonatype.com/artifact/io.github.caobahuong/kafka-connect-dynamic-filter-core)
[![Java](https://img.shields.io/badge/Java-17%2B-orange?logo=openjdk)](https://adoptium.net/)
[![Kafka Connect](https://img.shields.io/badge/Kafka_Connect-3.x-blue?logo=apachekafka)](https://kafka.apache.org/documentation/#connect)
[![License](https://img.shields.io/badge/license-Apache_2.0-blue)](LICENSE.txt)

🇺🇸 English · [🇻🇳 Tiếng Việt](./README.vi.md) · [🇨🇳 中文](./README.zh.md) · [🇯🇵 日本語](./README.ja.md)

---

**A Kafka Connect SMT for filtering Debezium CDC records with dynamic conditions** update rules at any time without restarting the connector.

> This document covers features and installation. For a step-by-step deployment walkthrough, see: [Deploying kafka-connect-dynamic-filter](https://github.com/caobahuong/kafka-connect-dynamic-filter/tree/main/blogs/en)

---

## Why This Library?

Debezium's built-in `Filter` SMT requires static Groovy/JS scripts hard-coded into the connector config. Changing a filter means editing the config, redeploying, and restarting the connector disrupting the pipeline each time.

For example: suppose your pipeline currently streams only records where `ref_id` is 1, 2, or 3. Adding values 4 and 5 with Debezium Filter requires editing the config, redeploying, restarting the connector, and tolerating pipeline disruption in the meantime.

With this library, you update the rule in Redis, a Kafka topic, or a JSON file. The connector picks it up on the very next record no config changes, no restart, no downtime.

| | Debezium Filter (built-in) | kafka-connect-dynamic-filter |
|---|---|---|
| Rule update | ❌ Connector restart required | ✅ Takes effect on next record |
| Rule language | Groovy / JS hardcoded | Dynamic JSON |
| Rule source | Config file | Redis, Kafka topic, File, or custom |
| Compound conditions | Write code | AND / OR / nested via JSON |

---

## Modules

> The project ships a `core` library for building your own loader. <br>
> The modules below are ready-made implementations provided by the author.

| Module | Where rules are stored |
|---|---|
| [**core**](./kafka-connect-dynamic-filter-core/) | Base library implement your own rule source |
| [**redis**](./kafka-connect-dynamic-filter-redis/) | Redis key, with Keyspace Notifications + polling |
| [**kafka**](./kafka-connect-dynamic-filter-kafka/) | Kafka topic |
| [**file**](./kafka-connect-dynamic-filter-file/) | JSON file, auto-reloads on change |

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Rule Syntax](#rule-syntax)
- [Rule Sources](#rule-sources)
- [How Records Are Filtered](#how-records-are-filtered)
- [Metrics and JMX](#metrics-and-jmx)
- [Debug & Troubleshooting](#debug--troubleshooting)
- [Configuration Reference](#configuration-reference)
- [Testing](#testing)
- [Contributing](#contributing)

---

## Installation

### Step 1: Download the JAR and copy to the plugin directory

Each module ships as a **self-contained fat JAR** core classes and all runtime dependencies are bundled inside. You only need one file.

Download the latest JAR from [Releases](https://github.com/caobahuong/kafka-connect-dynamic-filter/releases), then:

```bash
PLUGIN_DIR=$KAFKA_CONNECT_PLUGINS_DIR/kafka-connect-dynamic-filter
mkdir -p $PLUGIN_DIR

# Pick exactly one module
cp kafka-connect-dynamic-filter-redis-*.jar  $PLUGIN_DIR/
# or: kafka-connect-dynamic-filter-kafka-*.jar
# or: kafka-connect-dynamic-filter-file-*.jar
```

> **Not sure where your plugin directory is?** Run this to find it quickly:
> ```bash
> find /opt /usr /etc -iname "*kafka*" 2>/dev/null | grep plugin
> ```
> It's usually at `/opt/kafka-connect/plugins/`

After copying, restart Kafka Connect workers:

```bash
# systemd
systemctl restart kafka-connect

# or Confluent Platform
confluent local services connect restart
```

> **Using Docker?**
>
> Copy the JAR into the running container and restart:
> ```bash
> docker cp kafka-connect-dynamic-filter-redis-*.jar <container_name>:/opt/kafka-connect/plugins/
> docker restart <container_name>
> ```
> This approach loses the JAR on container recreation. Mount a volume in `docker-compose.yml` for persistence:
> ```yaml
> services:
>   kafka-connect:
>     volumes:
>       - ./plugins:/opt/kafka-connect/plugins
> ```
> Copy the JAR to `./plugins/` on the host, then run `docker compose restart kafka-connect`.

### Step 2: Declare the SMT in the connector config

**Via the Kafka Connect REST API:**
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

Example:
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
# Verify the connector received the config
curl http://localhost:8083/connectors/my-debezium-connector/status
```

| Module | `type` value |
|---|---|
| Redis | `io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter` |
| Kafka | `io.kafkaconnect.dynamicfilter.kafka.KafkaDynamicFilter` |
| File | `io.kafkaconnect.dynamicfilter.file.FileDynamicFilter` |
| Core (custom) | `io.kafkaconnect.dynamicfilter.DynamicListFilter` |

> See full per-module config in [Configuration Reference](#configuration-reference).

### Use core as a library (Maven / Gradle)

If you want to implement your own rule source via the Java API:

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

### Build from source

```bash
git clone https://github.com/caobahuong/kafka-connect-dynamic-filter.git
cd kafka-connect-dynamic-filter
./mvnw clean package -DskipTests
```

---

## Quick Start

Assumes you have already copied the JAR and restarted workers (Step 1 above). Concrete examples for each module below.

### Redis

Add to your connector config:

```properties
transforms=dynamicFilter
transforms.dynamicFilter.type=io.kafkaconnect.dynamicfilter.redis.RedisDynamicFilter
transforms.dynamicFilter.filter.id=company-filter
transforms.dynamicFilter.field.name=com_id
transforms.dynamicFilter.redis.uri=redis://localhost:6379
transforms.dynamicFilter.redis.key=filter:companies
transforms.dynamicFilter.empty.list.behavior=pass_all
```

Control the filter from anywhere, at any time no restart needed:

```bash
# Allow only companies 1, 2, 3
redis-cli SET filter:companies '[1, 2, 3]'

# Add company 4 takes effect on the next record
redis-cli SET filter:companies '[1, 2, 3, 4]'
