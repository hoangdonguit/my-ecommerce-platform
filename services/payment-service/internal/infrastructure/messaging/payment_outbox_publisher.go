package messaging

import (
	"context"
	"time"

	domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"
	kafkago "github.com/segmentio/kafka-go"
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

func (p *PaymentOutboxPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
