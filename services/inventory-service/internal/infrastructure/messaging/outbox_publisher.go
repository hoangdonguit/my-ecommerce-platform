package messaging

import (
	"context"
	"time"

	domaininventory "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/domain/inventory"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/observability"
	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type OutboxPublisher struct {
	writer *kafkago.Writer
}

func NewOutboxPublisher(broker string) *OutboxPublisher {
	return &OutboxPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
			Async:        false,
		},
	}
}

func (p *OutboxPublisher) PublishRaw(ctx context.Context, topic string, key string, payload []byte) error {
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	})
}

func (p *OutboxPublisher) PublishBatch(ctx context.Context, events []domaininventory.OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}

	ctx, span := otel.Tracer("inventory-service").Start(ctx, "kafka publish inventory outbox batch")
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

func (p *OutboxPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
