package messaging

import (
	"context"
	"encoding/json"
	"time"

	paymentapp "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/app/payment"
	kafkago "github.com/segmentio/kafka-go"
)

type PaymentPublisher struct {
	completedWriter *kafkago.Writer
	failedWriter    *kafkago.Writer
}

func NewPaymentPublisher(broker string, completedTopic string, failedTopic string) *PaymentPublisher {
	return &PaymentPublisher{
		completedWriter: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Topic:        completedTopic,
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

func (p *PaymentPublisher) PublishCompleted(ctx context.Context, event paymentapp.PaymentCompletedEvent) error {
	return p.publish(ctx, p.completedWriter, event.OrderID, event)
}

func (p *PaymentPublisher) PublishFailed(ctx context.Context, event paymentapp.PaymentFailedEvent) error {
	return p.publish(ctx, p.failedWriter, event.OrderID, event)
}

func (p *PaymentPublisher) publish(ctx context.Context, writer *kafkago.Writer, key string, event any) error {
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

func (p *PaymentPublisher) Close() error {
	var firstErr error

	if p.completedWriter != nil {
		if err := p.completedWriter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if p.failedWriter != nil {
		if err := p.failedWriter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
