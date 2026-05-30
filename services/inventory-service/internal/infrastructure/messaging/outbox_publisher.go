package messaging

import (
	"context"
	"time"

	domaininventory "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/domain/inventory"
	kafkago "github.com/segmentio/kafka-go"
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

	messages := make([]kafkago.Message, 0, len(events))
	now := time.Now()

	for _, event := range events {
		messages = append(messages, kafkago.Message{
			Topic: event.Topic,
			Key:   []byte(event.MessageKey),
			Value: event.Payload,
			Time:  now,
		})
	}

	return p.writer.WriteMessages(ctx, messages...)
}

func (p *OutboxPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
