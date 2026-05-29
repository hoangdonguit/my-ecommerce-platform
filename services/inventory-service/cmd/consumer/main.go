package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	inventoryapp "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/app/inventory"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/messaging"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/persistence"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()

	pool, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer pool.Close()

	log.Println("postgres connected successfully")

	repo := persistence.NewInventoryRepository(pool)

	publisher := messaging.NewInventoryPublisher(
		cfg.KafkaBroker,
		cfg.InventoryReservedTopic,
		cfg.InventoryFailedTopic,
	)
	defer publisher.Close()

	outboxPublisher := messaging.NewOutboxPublisher(cfg.KafkaBroker)
	defer outboxPublisher.Close()

	service := inventoryapp.NewService(repo, publisher)

	ctx := context.Background()

	go startOutboxPublisherLoop(ctx, repo, outboxPublisher)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        []string{cfg.KafkaBroker},
		Topic:          cfg.OrderCreatedTopic,
		GroupID:        cfg.KafkaGroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer reader.Close()

	log.Printf(
		"inventory consumer listening topic=%s group=%s broker=%s",
		cfg.OrderCreatedTopic,
		cfg.KafkaGroupID,
		cfg.KafkaBroker,
	)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("failed to fetch order.created message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf(
			"fetched order.created partition=%d offset=%d key=%s bytes=%d",
			msg.Partition,
			msg.Offset,
			string(msg.Key),
			len(msg.Value),
		)

		var event inventoryapp.OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf(
				"failed to unmarshal order.created partition=%d offset=%d err=%v payload=%s",
				msg.Partition,
				msg.Offset,
				err,
				string(msg.Value),
			)

			if commitErr := reader.CommitMessages(ctx, msg); commitErr != nil {
				log.Printf("failed to commit malformed message offset=%d err=%v", msg.Offset, commitErr)
			}
			continue
		}

		log.Printf("received order.created order_id=%s user_id=%s items=%d", event.OrderID, event.UserID, len(event.Items))

		if err := service.HandleOrderCreated(ctx, event); err != nil {
			log.Printf("failed to handle order.created order_id=%s err=%v", event.OrderID, err)
			time.Sleep(1 * time.Second)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit order.created order_id=%s offset=%d err=%v", event.OrderID, msg.Offset, err)
			continue
		}

		log.Printf("processed and committed order.created order_id=%s offset=%d", event.OrderID, msg.Offset)
	}
}

func startOutboxPublisherLoop(ctx context.Context, repo *persistence.InventoryRepository, publisher *messaging.OutboxPublisher) {
	batchSize := getEnvInt("OUTBOX_BATCH_SIZE", 100)
	idleSleep := time.Duration(getEnvInt("OUTBOX_IDLE_SLEEP_MS", 1000)) * time.Millisecond

	log.Printf("inventory outbox publisher started batch_size=%d idle_sleep=%s", batchSize, idleSleep)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		events, err := repo.FetchPendingOutboxEvents(ctx, batchSize)
		if err != nil {
			log.Printf("inventory outbox fetch failed err=%v", err)
			time.Sleep(idleSleep)
			continue
		}

		if len(events) == 0 {
			time.Sleep(idleSleep)
			continue
		}

		for _, event := range events {
			publishCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			err := publisher.PublishRaw(publishCtx, event.Topic, event.MessageKey, event.Payload)
			cancel()

			if err != nil {
				log.Printf(
					"inventory outbox publish failed id=%s aggregate_id=%s event_type=%s topic=%s attempts=%d err=%v",
					event.ID,
					event.AggregateID,
					event.EventType,
					event.Topic,
					event.Attempts,
					err,
				)
				if markErr := repo.MarkOutboxEventFailed(ctx, event.ID, err.Error()); markErr != nil {
					log.Printf("inventory outbox mark failed failed id=%s err=%v", event.ID, markErr)
				}
				continue
			}

			if err := repo.MarkOutboxEventPublished(ctx, event.ID); err != nil {
				log.Printf("inventory outbox mark published failed id=%s err=%v", event.ID, err)
				continue
			}

			log.Printf(
				"inventory outbox published id=%s aggregate_id=%s event_type=%s topic=%s attempts=%d",
				event.ID,
				event.AggregateID,
				event.EventType,
				event.Topic,
				event.Attempts,
			)
		}
	}
}

func getEnvInt(key string, fallback int) int {
	value := strconv.Itoa(fallback)
	if raw := getenv(key); raw != "" {
		value = raw
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getenv(key string) string {
	return os.Getenv(key)
}
