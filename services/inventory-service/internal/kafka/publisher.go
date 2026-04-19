package kafka

import (
	"context"
	"encoding/json"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Publisher struct {
	reservedWriter *kafkago.Writer
	failedWriter   *kafkago.Writer
}

func NewPublisher(broker, reservedTopic, failedTopic string) *Publisher {
	return &Publisher{
		reservedWriter: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Topic:        reservedTopic,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
			Async:        false,
		},
		failedWriter: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Topic:        failedTopic,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
			Async:        false,
		},
	}
}

func (p *Publisher) PublishReserved(ctx context.Context, key string, event any) error {
	return p.publish(ctx, p.reservedWriter, key, event)
}

func (p *Publisher) PublishFailed(ctx context.Context, key string, event any) error {
	return p.publish(ctx, p.failedWriter, key, event)
}

func (p *Publisher) publish(ctx context.Context, writer *kafkago.Writer, key string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafkago.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	return writer.WriteMessages(ctx, msg)
}
