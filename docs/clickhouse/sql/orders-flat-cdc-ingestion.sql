CREATE DATABASE IF NOT EXISTS analytics;

DROP VIEW IF EXISTS analytics.mv_orders_flat_cdc_events;
DROP TABLE IF EXISTS analytics.orders_flat_cdc_queue;
DROP TABLE IF EXISTS analytics.orders_flat_cdc_events;

CREATE TABLE analytics.orders_flat_cdc_events
(
    ingested_at DateTime64(3) DEFAULT now64(3),

    kafka_topic String,
    kafka_partition Int32,
    kafka_offset UInt64,

    order_id String,
    user_id String,
    status LowCardinality(String),
    currency LowCardinality(String),
    payment_method LowCardinality(String),
    shipping_address String,
    note String,
    total_amount Decimal(18, 2),
    idempotency_key String,

    created_at_us Int64,
    updated_at_us Int64,

    op LowCardinality(String),
    source_ts_ms Int64,
    source_lsn UInt64,
    source_tx_id UInt64,
    source_table String,
    source_db String
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(ingested_at)
ORDER BY (order_id, source_ts_ms, kafka_offset);

CREATE TABLE analytics.orders_flat_cdc_queue
(
    raw_message String
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'kafka-svc.kafka.svc.cluster.local:9092',
    kafka_topic_list = 'cdc_flat.order_db.public.orders',
    kafka_group_name = 'clickhouse-orders-flat-cdc-v2',
    kafka_format = 'JSONAsString',
    kafka_num_consumers = 1,
    kafka_skip_broken_messages = 100;

CREATE MATERIALIZED VIEW analytics.mv_orders_flat_cdc_events
TO analytics.orders_flat_cdc_events
AS
SELECT
    now64(3) AS ingested_at,

    _topic AS kafka_topic,
    _partition AS kafka_partition,
    _offset AS kafka_offset,

    JSONExtractString(raw_message, 'id') AS order_id,
    JSONExtractString(raw_message, 'user_id') AS user_id,
    JSONExtractString(raw_message, 'status') AS status,
    JSONExtractString(raw_message, 'currency') AS currency,
    JSONExtractString(raw_message, 'payment_method') AS payment_method,
    JSONExtractString(raw_message, 'shipping_address') AS shipping_address,
    JSONExtractString(raw_message, 'note') AS note,
    toDecimal64OrZero(JSONExtractString(raw_message, 'total_amount'), 2) AS total_amount,
    JSONExtractString(raw_message, 'idempotency_key') AS idempotency_key,

    JSONExtractInt(raw_message, 'created_at') AS created_at_us,
    JSONExtractInt(raw_message, 'updated_at') AS updated_at_us,

    JSONExtractString(raw_message, '__op') AS op,
    JSONExtractInt(raw_message, '__source_ts_ms') AS source_ts_ms,
    toUInt64OrZero(toString(JSONExtractInt(raw_message, '__source_lsn'))) AS source_lsn,
    toUInt64OrZero(toString(JSONExtractInt(raw_message, '__source_txId'))) AS source_tx_id,
    JSONExtractString(raw_message, '__table') AS source_table,
    JSONExtractString(raw_message, '__db') AS source_db
FROM analytics.orders_flat_cdc_queue
WHERE JSONExtractString(raw_message, 'id') != '';
