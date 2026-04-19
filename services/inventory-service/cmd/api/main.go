package main

import (
	"log"

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
	handler := httpapi.NewInventoryHandler(service)

	router := httpapi.SetupRouter(handler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
