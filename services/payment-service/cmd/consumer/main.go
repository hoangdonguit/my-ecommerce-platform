package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	paymentapp "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/app/payment"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/messaging"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/persistence"
	kafkago "github.com/segmentio/kafka-go"
)

const maxProcessAttempts = 3

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

	repo := persistence.NewPaymentRepository(pool)

	eventPublisher := messaging.NewPaymentPublisher(
		cfg.KafkaBroker,
		cfg.PaymentCompletedTopic,
		cfg.PaymentFailedTopic,
	)
	defer eventPublisher.Close()

	outboxPublisher := messaging.NewPaymentOutboxPublisher(cfg.KafkaBroker)
	defer outboxPublisher.Close()

	go startPaymentOutboxPublisherLoop(ctx, repo, outboxPublisher)

	gateway := paymentapp.NewSimulatedPaymentGateway()

	// Payment terminal events are now published by payment_outbox_events.
	// Keep direct publisher disabled in the business service to avoid duplicate events.
	service := paymentapp.NewService(repo, nil, gateway)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.InventoryReservedTopic,
		GroupID: cfg.KafkaGroupID,
	})
	defer reader.Close()

	dlqTopic := cfg.InventoryReservedTopic + ".dlq"
	dlqWriter := newDLQWriter(cfg.KafkaBroker, dlqTopic)
	defer dlqWriter.Close()

	log.Printf(
		"payment consumer listening topic=%s group=%s dlq=%s",
		cfg.InventoryReservedTopic,
		cfg.KafkaGroupID,
		dlqTopic,
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

		if err := processPaymentMessage(ctx, service, msg); err != nil {
			log.Printf("payment consumer failed topic=%s partition=%d offset=%d err=%v",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err,
			)

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

func startPaymentOutboxPublisherLoop(ctx context.Context, repo *persistence.PaymentRepository, publisher *messaging.PaymentOutboxPublisher) {
	batchSize := getEnvInt("PAYMENT_OUTBOX_BATCH_SIZE", 200)
	idleSleep := time.Duration(getEnvInt("PAYMENT_OUTBOX_IDLE_SLEEP_MS", 200)) * time.Millisecond

	log.Printf("payment outbox publisher started batch_size=%d idle_sleep=%s mode=batch", batchSize, idleSleep)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		events, err := repo.FetchPendingOutboxEvents(ctx, batchSize)
		if err != nil {
			log.Printf("payment outbox fetch failed err=%v", err)
			time.Sleep(idleSleep)
			continue
		}

		if len(events) == 0 {
			time.Sleep(idleSleep)
			continue
		}

		ids := make([]string, 0, len(events))
		for _, event := range events {
			ids = append(ids, event.ID)
		}

		publishCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err = publisher.PublishBatch(publishCtx, events)
		cancel()

		if err != nil {
			log.Printf("payment outbox batch publish failed count=%d err=%v", len(events), err)

			if markErr := repo.MarkOutboxEventsFailed(ctx, ids, err.Error()); markErr != nil {
				log.Printf("payment outbox batch mark failed failed count=%d err=%v", len(ids), markErr)
			}

			time.Sleep(idleSleep)
			continue
		}

		if err := repo.MarkOutboxEventsPublished(ctx, ids); err != nil {
			log.Printf("payment outbox batch mark published failed count=%d err=%v", len(ids), err)
			time.Sleep(idleSleep)
			continue
		}

		log.Printf(
			"payment outbox batch published count=%d first_aggregate_id=%s last_aggregate_id=%s",
			len(events),
			events[0].AggregateID,
			events[len(events)-1].AggregateID,
		)
	}
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func processPaymentMessage(ctx context.Context, service *paymentapp.Service, msg kafkago.Message) error {
	var event paymentapp.InventoryReservedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal inventory.reserved event: %w", err)
	}

	log.Printf("received inventory.reserved order_id=%s amount=%.2f method=%s",
		event.OrderID,
		event.TotalAmount,
		event.PaymentMethod,
	)

	var lastErr error
	for attempt := 1; attempt <= maxProcessAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		err := service.HandleInventoryReserved(attemptCtx, event)
		cancel()

		if err == nil {
			log.Printf("processed inventory.reserved successfully order_id=%s attempt=%d",
				event.OrderID,
				attempt,
			)
			return nil
		}

		lastErr = err
		log.Printf("retry inventory.reserved order_id=%s attempt=%d/%d err=%v",
			event.OrderID,
			attempt,
			maxProcessAttempts,
			err,
		)

		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	}

	return fmt.Errorf("handle inventory.reserved order_id=%s after %d attempts: %w",
		event.OrderID,
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
