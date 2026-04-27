package main

import (
	"context"
	"encoding/json"
	"log"

	paymentapp "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/app/payment"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/messaging"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/persistence"
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

	repo := persistence.NewPaymentRepository(pool)

	publisher := messaging.NewPaymentPublisher(
		cfg.KafkaBroker,
		cfg.PaymentCompletedTopic,
		cfg.PaymentFailedTopic,
	)
	defer publisher.Close()

	gateway := paymentapp.NewSimulatedPaymentGateway()
	service := paymentapp.NewService(repo, publisher, gateway)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.InventoryReservedTopic,
		GroupID: cfg.KafkaGroupID,
	})
	defer reader.Close()

	log.Printf(
		"payment consumer listening topic=%s group=%s",
		cfg.InventoryReservedTopic,
		cfg.KafkaGroupID,
	)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("failed to read message: %v", err)
			continue
		}

		var event paymentapp.InventoryReservedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal inventory.reserved event: %v", err)
			continue
		}

		log.Printf("received inventory.reserved order_id=%s amount=%.2f method=%s",
			event.OrderID,
			event.TotalAmount,
			event.PaymentMethod,
		)

		if err := service.HandleInventoryReserved(context.Background(), event); err != nil {
			log.Printf("failed to handle inventory.reserved order_id=%s err=%v", event.OrderID, err)
			continue
		}

		log.Printf("processed inventory.reserved successfully order_id=%s", event.OrderID)
	}
}
