package messaging

import (
	"context"
	"encoding/json"
	"time"

	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/observability"
	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type OrderPublisher struct {
	writer *kafkago.Writer
	topic  string
}

func NewOrderPublisher(broker string, topic string) *OrderPublisher {
	return &OrderPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Topic:        topic,
			Balancer:     &kafkago.Hash{}, // Hash by key → same order → same partition
			RequiredAcks: kafkago.RequireAll,
			Async:        false, // Giữ sync để đảm bảo delivery
			BatchSize:    500,   // Gửi tối đa 500 msg/lần
			BatchTimeout: 5 * time.Millisecond,
		},
		topic: topic,
	}
}

// Gửi từng message (giữ lại để tương thích)
func (p *OrderPublisher) PublishOrderCreated(ctx context.Context, event orderapp.OrderCreatedEvent) error {
	return p.PublishOrderCreatedBatch(ctx, []orderapp.OrderCreatedEvent{event})
}

// Gửi BATCH - dùng WriteMessages 1 lần cho toàn bộ
func (p *OrderPublisher) PublishOrderCreatedBatch(ctx context.Context, events []orderapp.OrderCreatedEvent) error {
	ctx, span := otel.Tracer("order-service").Start(ctx, "kafka publish order.created")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", p.topic),
		attribute.Int("messaging.batch.message_count", len(events)),
	)

	msgs := make([]kafkago.Message, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			continue
		}
		headers := make([]kafkago.Header, 0)
		otel.GetTextMapPropagator().Inject(ctx, observability.NewKafkaHeadersCarrier(&headers))

		msgs = append(msgs, kafkago.Message{
			Key:     []byte(event.OrderID),
			Value:   payload,
			Time:    time.Now(),
			Headers: headers,
		})
	}
	if len(msgs) == 0 {
		return nil
	}

	if err := p.writer.WriteMessages(ctx, msgs...); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *OrderPublisher) Close() error {
	return p.writer.Close()
}
