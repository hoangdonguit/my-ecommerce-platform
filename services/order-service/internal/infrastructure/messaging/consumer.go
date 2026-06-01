package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/observability"
	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type OrderConsumer struct {
	reader *kafkago.Reader
	repo   domainorder.Repository
}

type sagaEvent struct {
	EventType string `json:"event_type"`
	OrderID   string `json:"order_id"`
}

func NewOrderConsumer(brokers []string, repo domainorder.Repository) *OrderConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        brokers,
		GroupID:        "order-service-saga-monitor",
		GroupTopics:    []string{"inventory.failed", "payment.failed", "payment.completed"},
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	return &OrderConsumer{
		reader: reader,
		repo:   repo,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) {
	log.Println("Order saga monitor started: listening inventory.failed, payment.failed, payment.completed")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("order saga monitor stopped: %v", ctx.Err())
				return
			}

			log.Printf("failed to fetch saga message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		c.processMessage(ctx, msg)
	}
}

func (c *OrderConsumer) processMessage(ctx context.Context, msg kafkago.Message) {
	headers := msg.Headers
	messageCtx := otel.GetTextMapPropagator().Extract(
		ctx,
		observability.NewKafkaHeadersCarrier(&headers),
	)

	messageCtx, span := otel.Tracer("order-service").Start(messageCtx, "kafka consume saga event")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", msg.Topic),
		attribute.String("messaging.operation", "process"),
		attribute.Int("messaging.kafka.partition", msg.Partition),
		attribute.Int64("messaging.kafka.offset", msg.Offset),
	)

	var event sagaEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		span.RecordError(err)

		log.Printf(
			"failed to unmarshal saga message topic=%s partition=%d offset=%d err=%v payload=%s",
			msg.Topic,
			msg.Partition,
			msg.Offset,
			err,
			string(msg.Value),
		)

		if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
			span.RecordError(commitErr)
			log.Printf("failed to commit malformed saga message offset=%d err=%v", msg.Offset, commitErr)
		}

		return
	}

	span.SetAttributes(
		attribute.String("order.id", event.OrderID),
		attribute.String("saga.event_type", event.EventType),
	)

	if event.OrderID == "" || event.EventType == "" {
		log.Printf(
			"invalid saga message topic=%s partition=%d offset=%d event_type=%q order_id=%q payload=%s",
			msg.Topic,
			msg.Partition,
			msg.Offset,
			event.EventType,
			event.OrderID,
			string(msg.Value),
		)

		if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
			span.RecordError(commitErr)
			log.Printf("failed to commit invalid saga message offset=%d err=%v", msg.Offset, commitErr)
		}

		return
	}

	targetStatus := ""
	switch event.EventType {
	case "payment.completed":
		targetStatus = domainorder.StatusCompleted
	case "inventory.failed", "payment.failed":
		targetStatus = domainorder.StatusFailed
	default:
		log.Printf(
			"skip unsupported saga event event_type=%s order_id=%s topic=%s offset=%d",
			event.EventType,
			event.OrderID,
			msg.Topic,
			msg.Offset,
		)

		if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
			span.RecordError(commitErr)
			log.Printf("failed to commit skipped saga message order_id=%s offset=%d err=%v", event.OrderID, msg.Offset, commitErr)
		}

		return
	}

	span.SetAttributes(attribute.String("order.target_status", targetStatus))

	log.Printf(
		"received saga event event_type=%s order_id=%s target_status=%s topic=%s partition=%d offset=%d",
		event.EventType,
		event.OrderID,
		targetStatus,
		msg.Topic,
		msg.Partition,
		msg.Offset,
	)

	if err := c.repo.UpdateStatus(messageCtx, event.OrderID, targetStatus); err != nil {
		span.RecordError(err)

		log.Printf(
			"failed to update order status order_id=%s target_status=%s err=%v",
			event.OrderID,
			targetStatus,
			err,
		)

		time.Sleep(1 * time.Second)
		return
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		span.RecordError(err)

		log.Printf(
			"failed to commit saga message order_id=%s target_status=%s offset=%d err=%v",
			event.OrderID,
			targetStatus,
			msg.Offset,
			err,
		)

		return
	}

	span.SetAttributes(attribute.String("order.status", targetStatus))

	log.Printf(
		"processed and committed saga event order_id=%s status=%s offset=%d",
		event.OrderID,
		targetStatus,
		msg.Offset,
	)
}
