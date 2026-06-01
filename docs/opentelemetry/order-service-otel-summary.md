# Order Service OpenTelemetry Summary

## Purpose

Instrument order-service with OpenTelemetry HTTP tracing so traces can continue from web-gateway into the order processing service.

## Implemented

- OpenTelemetry SDK initialization in order-service main.
- OTLP gRPC export to OpenTelemetry Collector.
- Runtime OTEL_* configuration in order-service config.
- Gin inbound request tracing using otelgin.
- Kubernetes deployment updated with OTEL_* environment variables.
- order-service image updated to an OpenTelemetry-enabled image.

## Runtime Configuration

- OTEL_ENABLED=true
- OTEL_SERVICE_NAME=order-service
- OTEL_ENVIRONMENT=development
- OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector.observability.svc.cluster.local:4317

## Deployment Result

order-service was deployed successfully through ArgoCD.

Observed runtime state:

- ecommerce-platform: Synced / Healthy
- order-service replicas: 2/2 ready
- image: hoangdonguit/order-service:otel-orderservice-20260601232328
- postgres connected successfully
- redis connected successfully
- order outbox worker started
- saga monitor started
- bad pods: none

## Verification Result

A test order was created through web-gateway after order-service tracing was enabled.

Request result:

- HTTP status: 201 Created
- order-service handled POST /api/v1/orders successfully
- web-gateway handled POST /api/orders successfully

OpenTelemetry Collector metrics showed:

- accepted spans: 5
- sent spans to Tempo: 5
- refused spans: 0
- failed export spans: 0

Tempo /ready returned HTTP 200 OK.

## Conclusion

HTTP tracing now works across the main hot path:

Client -> web-gateway -> order-service

The next phase is to instrument Kafka publish and consume paths so the order saga can be traced beyond HTTP.
