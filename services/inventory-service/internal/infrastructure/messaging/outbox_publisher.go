package messaging

import (
	"context"
	"time"

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

func (p *OutboxPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
