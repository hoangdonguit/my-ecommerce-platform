package main

import (
	"context"
	"encoding/json"
	"log"

	notificationapp "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/app/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/domain/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/persistence"
	kafkago "github.com/segmentio/kafka-go"
)

type BaseEvent struct {
	EventType string `json:"event_type"`
}

func main() {
	cfg := config.Load()

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

	log.Printf(
		"notification consumer listening topics=%s,%s group=%s",
		cfg.PaymentCompletedTopic,
		cfg.PaymentFailedTopic,
		cfg.KafkaGroupID,
	)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("failed to read message: %v", err)
			continue
		}

		var base BaseEvent
		if err := json.Unmarshal(msg.Value, &base); err != nil {
			log.Printf("failed to unmarshal base event: %v", err)
			continue
		}

		switch base.EventType {
		case notification.EventPaymentCompleted:
			var event notificationapp.PaymentCompletedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("failed to unmarshal payment.completed event: %v", err)
				continue
			}

			log.Printf("received payment.completed order_id=%s user_id=%s", event.OrderID, event.UserID)

			if err := service.HandlePaymentCompleted(context.Background(), event); err != nil {
				log.Printf("failed to handle payment.completed order_id=%s err=%v", event.OrderID, err)
				continue
			}

			log.Printf("processed payment.completed successfully order_id=%s", event.OrderID)

		case notification.EventPaymentFailed:
			var event notificationapp.PaymentFailedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("failed to unmarshal payment.failed event: %v", err)
				continue
			}

			log.Printf("received payment.failed order_id=%s user_id=%s", event.OrderID, event.UserID)

			if err := service.HandlePaymentFailed(context.Background(), event); err != nil {
				log.Printf("failed to handle payment.failed order_id=%s err=%v", event.OrderID, err)
				continue
			}

			log.Printf("processed payment.failed successfully order_id=%s", event.OrderID)

		default:
			log.Printf("ignored unknown event_type=%s", base.EventType)
		}
	}
}
