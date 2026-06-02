package messaging

import (
	"context"
	"time"

	domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/observability"
	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type PaymentOutboxPublisher struct {
	writer *kafkago.Writer
}

func NewPaymentOutboxPublisher(broker string) *PaymentOutboxPublisher {
	return &PaymentOutboxPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
			Async:        false,
			BatchSize:    500,
			BatchTimeout: 5 * time.Millisecond,
		},
	}
}

func (p *PaymentOutboxPublisher) PublishBatch(ctx context.Context, events []domainpayment.OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}

	ctx, span := otel.Tracer("payment-service").Start(ctx, "kafka publish payment outbox batch")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.Int("messaging.batch.message_count", len(events)),
	)

	messages := make([]kafkago.Message, 0, len(events))
	now := time.Now()

	for _, event := range events {
		headers := observability.KafkaHeadersFromJSON(event.Headers)
		if len(headers) == 0 {
			headers = make([]kafkago.Header, 0)
			otel.GetTextMapPropagator().Inject(ctx, observability.NewKafkaHeadersCarrier(&headers))
		}

		messages = append(messages, kafkago.Message{
			Topic:   event.Topic,
			Key:     []byte(event.MessageKey),
			Value:   event.Payload,
			Time:    now,
			Headers: headers,
		})
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *PaymentOutboxPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
