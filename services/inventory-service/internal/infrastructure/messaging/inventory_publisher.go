package messaging

import (
	"context"
	"encoding/json"
	"time"

	inventoryapp "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/app/inventory"
	kafkago "github.com/segmentio/kafka-go"
)

type InventoryPublisher struct {
	reservedWriter *kafkago.Writer
	failedWriter   *kafkago.Writer
}

func NewInventoryPublisher(broker, reservedTopic, failedTopic string) *InventoryPublisher {
	return &InventoryPublisher{
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

func (p *InventoryPublisher) PublishReserved(ctx context.Context, event inventoryapp.InventoryReservedEvent) error {
	return p.publish(ctx, p.reservedWriter, event.OrderID, event)
}

func (p *InventoryPublisher) PublishFailed(ctx context.Context, event inventoryapp.InventoryFailedEvent) error {
	return p.publish(ctx, p.failedWriter, event.OrderID, event)
}

func (p *InventoryPublisher) publish(ctx context.Context, writer *kafkago.Writer, key string, event any) error {
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

func (p *InventoryPublisher) Close() error {
	var firstErr error

	if p.reservedWriter != nil {
		if err := p.reservedWriter.Close(); err != nil && firstErr == nil {
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
