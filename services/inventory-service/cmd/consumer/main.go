package main

import (
	"context"
	"encoding/json"
	"log"

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

	service := inventoryapp.NewService(repo, publisher)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.OrderCreatedTopic,
		GroupID: cfg.KafkaGroupID,
	})
	defer reader.Close()

	log.Printf(
		"inventory consumer listening topic=%s group=%s",
		cfg.OrderCreatedTopic,
		cfg.KafkaGroupID,
	)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("failed to read message: %v", err)
			continue
		}

		var event inventoryapp.OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal order.created event: %v", err)
			continue
		}

		log.Printf("received order.created order_id=%s items=%d", event.OrderID, len(event.Items))

		if err := service.HandleOrderCreated(context.Background(), event); err != nil {
			log.Printf("failed to handle order.created order_id=%s err=%v", event.OrderID, err)
			continue
		}

		log.Printf("processed order.created successfully order_id=%s", event.OrderID)
	}
}
