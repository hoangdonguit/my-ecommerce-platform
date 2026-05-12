package main

import (
	"context"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	inventoryapp "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/app/inventory"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/messaging"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/persistence"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/interfaces/http"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. Kết nối PostgreSQL
	pool, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Postgres connected successfully")

	// 2. Khởi tạo Repository & Publisher
	repo := persistence.NewInventoryRepository(pool)
	publisher := messaging.NewInventoryPublisher(
		cfg.KafkaBroker,
		cfg.InventoryReservedTopic,
		cfg.InventoryFailedTopic,
	)
	defer publisher.Close()

	// 3. Khởi tạo Service
	service := inventoryapp.NewService(repo, publisher)

	// 4. Setup Kafka Consumer (Chạy ngầm TRƯỚC KHI Server Run)
	kafkaBrokers := strings.Split(cfg.KafkaBroker, ",")
	invConsumer := messaging.NewInventoryConsumer(kafkaBrokers, service)
	go func() {
		log.Println("🚀 Inventory Consumer is starting...")
		invConsumer.Start(context.Background())
	}()

	// 5. Setup Router & Handler
	handler := httpapi.NewInventoryHandler(service)
	router := httpapi.SetupRouter(handler)

	log.Printf("📡 %s starting on port %s", cfg.AppName, cfg.AppPort)

	// 6. Chạy Server
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}