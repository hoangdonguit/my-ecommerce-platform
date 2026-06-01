# Order Service Kafka Tracing Summary

## Purpose

Extend order-service OpenTelemetry tracing from HTTP-only tracing to Kafka publish and consume tracing.

## Implemented

- Added Kafka headers carrier for OpenTelemetry propagation.
- Injected TraceContext into Kafka headers when publishing order.created.
- Extracted TraceContext from Kafka headers when consuming saga events.
- Added span for Kafka publish: kafka publish order.created.
- Added span for Kafka consume: kafka consume saga event.
- Updated order-service image with Kafka tracing support.

## Runtime Image

- hoangdonguit/order-service:otel-kafka-orderservice-20260602003109

## Verification

A new order was created through web-gateway.

Result:

- HTTP status: 201 Created
- Order ID: 2a45bdf6-8d17-40d2-bbb2-7fa6e3b7a834
- Final saga event: payment.completed
- Final order status: COMPLETED
- Bad pods: none

OpenTelemetry Collector metrics:

Before:

- accepted spans: 233
- sent spans: 233
- refused spans: 0
- failed export spans: 0

After:

- accepted spans: 241
- sent spans: 241
- refused spans: 0
- failed export spans: 0

Delta:

- accepted spans increased by 8
- sent spans increased by 8

## Conclusion

Kafka tracing for order-service is working at runtime. The order-service can now create tracing spans for Kafka publishing and saga-event consuming, and the OpenTelemetry Collector successfully exports those spans to Tempo.
