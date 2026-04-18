package kafka

import (
	"context"
	"encoding/json"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/model"
)

type OrderPublisher struct {
	writer *kafkago.Writer
	topic  string
}

func NewOrderPublisher(broker, topic string) *OrderPublisher {
	return &OrderPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(broker),
			Topic:        topic,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
			Async:        false,
		},
		topic: topic,
	}
}

func (p *OrderPublisher) PublishOrderCreated(ctx context.Context, event model.OrderCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafkago.Message{
		Key:   []byte(event.OrderID),
		Value: payload,
		Time:  time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}
