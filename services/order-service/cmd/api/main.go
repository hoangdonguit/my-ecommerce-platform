package main

import (
	"log"

	"github.com/gin-gonic/gin"

	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/messaging"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/persistence"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/interfaces/http"
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

	orderRepo := persistence.NewOrderRepository(pool)
	orderPublisher := messaging.NewOrderPublisher(cfg.KafkaBroker, cfg.OrderCreatedTopic)
	defer orderPublisher.Close()

	orderService := orderapp.NewService(orderRepo, orderPublisher)
	orderHandler := httpapi.NewOrderHandler(orderService)

	router := httpapi.SetupRouter(orderHandler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
