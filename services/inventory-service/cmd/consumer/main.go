package main

import (
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/db"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/kafka"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/model"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/repository"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/service"
)

func main() {
	cfg := config.Load()

	conn, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("cannot connect db: %v", err)
	}
	defer conn.Close()

	repo := repository.NewInventoryRepository(conn)
	publisher := kafka.NewPublisher(cfg.KafkaBroker, cfg.InventoryReservedTopic, cfg.InventoryFailedTopic)
	svc := service.NewInventoryService(repo, publisher)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.OrderCreatedTopic,
		GroupID: cfg.KafkaGroupID,
	})

	defer reader.Close()

	log.Printf("inventory-service consumer listening topic=%s group=%s", cfg.OrderCreatedTopic, cfg.KafkaGroupID)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("read message error: %v", err)
			continue
		}

		var event model.OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		log.Printf("received order.created: %+v", event)

		if err := svc.HandleOrderCreated(context.Background(), event); err != nil {
			log.Printf("handle event error: %v", err)
			continue
		}

		log.Printf("processed order.created successfully, order_id=%s", event.OrderID)
	}
}
