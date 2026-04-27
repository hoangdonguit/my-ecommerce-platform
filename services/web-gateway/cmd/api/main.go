package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/app"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/client"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/config"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/interfaces/http"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	orderClient := client.NewOrderClient(cfg.OrderServiceURL)
	inventoryClient := client.NewInventoryClient(cfg.InventoryServiceURL)
	paymentClient := client.NewPaymentClient(cfg.PaymentServiceURL)
	notificationClient := client.NewNotificationClient(cfg.NotificationServiceURL)

	sagaService := app.NewSagaService(
		orderClient,
		inventoryClient,
		paymentClient,
		notificationClient,
	)

	handler := httpapi.NewHandler(
		orderClient,
		inventoryClient,
		paymentClient,
		notificationClient,
		sagaService,
	)

	router := httpapi.SetupRouter(handler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)
	log.Printf("order service url: %s", cfg.OrderServiceURL)
	log.Printf("inventory service url: %s", cfg.InventoryServiceURL)
	log.Printf("payment service url: %s", cfg.PaymentServiceURL)
	log.Printf("notification service url: %s", cfg.NotificationServiceURL)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
