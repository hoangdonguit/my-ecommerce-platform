FROM maven:3.9.9-eclipse-temurin-17 AS builder

WORKDIR /src

RUN git clone --depth 1 https://github.com/caobahuong/kafka-connect-dynamic-filter.git .

RUN mvn -q -pl kafka-connect-dynamic-filter-redis -am package -DskipTests

FROM quay.io/debezium/connect:2.7

USER root

RUN mkdir -p /kafka/connect/kafka-connect-dynamic-filter

COPY --from=builder /src/kafka-connect-dynamic-filter-redis/target/kafka-connect-dynamic-filter-redis-*.jar /kafka/connect/kafka-connect-dynamic-filter/

USER kafka
