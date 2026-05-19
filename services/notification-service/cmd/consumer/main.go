package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	notificationapp "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/app/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/domain/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/persistence"
	kafkago "github.com/segmentio/kafka-go"
)

const maxProcessAttempts = 3

type BaseEvent struct {
	EventType string `json:"event_type"`
}

type DLQPayload struct {
	OriginalTopic     string            `json:"original_topic"`
	OriginalPartition int               `json:"original_partition"`
	OriginalOffset    int64             `json:"original_offset"`
	OriginalKey       string            `json:"original_key"`
	Error             string            `json:"error"`
	FailedAt          string            `json:"failed_at"`
	Payload           string            `json:"payload"`
	Headers           map[string]string `json:"headers,omitempty"`
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer pool.Close()

	log.Println("postgres connected successfully")

	repo := persistence.NewNotificationRepository(pool)
	service := notificationapp.NewService(repo)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		GroupID: cfg.KafkaGroupID,
		GroupTopics: []string{
			cfg.PaymentCompletedTopic,
			cfg.PaymentFailedTopic,
		},
	})
	defer reader.Close()

	completedDLQTopic := cfg.PaymentCompletedTopic + ".dlq"
	failedDLQTopic := cfg.PaymentFailedTopic + ".dlq"

	completedDLQWriter := newDLQWriter(cfg.KafkaBroker, completedDLQTopic)
	defer completedDLQWriter.Close()

	failedDLQWriter := newDLQWriter(cfg.KafkaBroker, failedDLQTopic)
	defer failedDLQWriter.Close()

	log.Printf(
		"notification consumer listening topics=%s,%s group=%s dlq=%s,%s",
		cfg.PaymentCompletedTopic,
		cfg.PaymentFailedTopic,
		cfg.KafkaGroupID,
		completedDLQTopic,
		failedDLQTopic,
	)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("failed to fetch message: %v", err)
			if ctx.Err() != nil {
				return
			}
			continue
		}

		if err := processNotificationMessage(ctx, service, msg); err != nil {
			log.Printf("notification consumer failed topic=%s partition=%d offset=%d err=%v",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err,
			)

			dlqWriter := completedDLQWriter
			dlqTopic := completedDLQTopic
			if msg.Topic == cfg.PaymentFailedTopic {
				dlqWriter = failedDLQWriter
				dlqTopic = failedDLQTopic
			}

			if dlqErr := publishDLQ(ctx, dlqWriter, msg, err); dlqErr != nil {
				log.Printf("failed to publish dlq topic=%s partition=%d offset=%d err=%v",
					msg.Topic,
					msg.Partition,
					msg.Offset,
					dlqErr,
				)
				continue
			}

			log.Printf("sent message to dlq=%s original_topic=%s offset=%d",
				dlqTopic,
				msg.Topic,
				msg.Offset,
			)
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit message topic=%s partition=%d offset=%d err=%v",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err,
			)
			continue
		}
	}
}

func processNotificationMessage(ctx context.Context, service *notificationapp.Service, msg kafkago.Message) error {
	var base BaseEvent
	if err := json.Unmarshal(msg.Value, &base); err != nil {
		return fmt.Errorf("unmarshal base event: %w", err)
	}

	switch base.EventType {
	case notification.EventPaymentCompleted:
		var event notificationapp.PaymentCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("unmarshal payment.completed event: %w", err)
		}

		log.Printf("received payment.completed order_id=%s user_id=%s", event.OrderID, event.UserID)

		return retry(ctx, "payment.completed", event.OrderID, func(attemptCtx context.Context) error {
			return service.HandlePaymentCompleted(attemptCtx, event)
		})

	case notification.EventPaymentFailed:
		var event notificationapp.PaymentFailedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("unmarshal payment.failed event: %w", err)
		}

		log.Printf("received payment.failed order_id=%s user_id=%s", event.OrderID, event.UserID)

		return retry(ctx, "payment.failed", event.OrderID, func(attemptCtx context.Context) error {
			return service.HandlePaymentFailed(attemptCtx, event)
		})

	default:
		return fmt.Errorf("unknown event_type=%s", base.EventType)
	}
}

func retry(ctx context.Context, eventType string, orderID string, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 1; attempt <= maxProcessAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		err := fn(attemptCtx)
		cancel()

		if err == nil {
			log.Printf("processed %s successfully order_id=%s attempt=%d",
				eventType,
				orderID,
				attempt,
			)
			return nil
		}

		lastErr = err
		log.Printf("retry %s order_id=%s attempt=%d/%d err=%v",
			eventType,
			orderID,
			attempt,
			maxProcessAttempts,
			err,
		)

		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	}

	return fmt.Errorf("handle %s order_id=%s after %d attempts: %w",
		eventType,
		orderID,
		maxProcessAttempts,
		lastErr,
	)
}

func newDLQWriter(broker string, topic string) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:         kafkago.TCP(broker),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
	}
}

func publishDLQ(ctx context.Context, writer *kafkago.Writer, msg kafkago.Message, cause error) error {
	payload := DLQPayload{
		OriginalTopic:     msg.Topic,
		OriginalPartition: msg.Partition,
		OriginalOffset:    msg.Offset,
		OriginalKey:       string(msg.Key),
		Error:             cause.Error(),
		FailedAt:          time.Now().UTC().Format(time.RFC3339),
		Payload:           string(msg.Value),
		Headers:           headersToMap(msg.Headers),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return writer.WriteMessages(ctx, kafkago.Message{
		Key:   msg.Key,
		Value: raw,
		Time:  time.Now(),
		Headers: []kafkago.Header{
			{Key: "x-original-topic", Value: []byte(msg.Topic)},
			{Key: "x-error", Value: []byte(cause.Error())},
		},
	})
}

func headersToMap(headers []kafkago.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string]string, len(headers))
	for _, h := range headers {
		result[h.Key] = string(h.Value)
	}

	return result
}
