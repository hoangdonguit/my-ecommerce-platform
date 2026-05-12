package messaging

import (
	"context"
	"encoding/json"
	"time"

	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	kafkago "github.com/segmentio/kafka-go"
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
			Balancer:     &kafkago.Hash{},   // Hash by key → same order → same partition
			RequiredAcks: kafkago.RequireOne,
			Async:        false,             // Giữ sync để đảm bảo delivery
			BatchSize:    500,               // Gửi tối đa 500 msg/lần
			BatchTimeout: 10 * time.Millisecond,
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
	msgs := make([]kafkago.Message, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			continue
		}
		msgs = append(msgs, kafkago.Message{
			Key:   []byte(event.OrderID),
			Value: payload,
			Time:  time.Now(),
		})
	}
	if len(msgs) == 0 {
		return nil
	}
	return p.writer.WriteMessages(ctx, msgs...)
}

func (p *OrderPublisher) Close() error {
	return p.writer.Close()
}
