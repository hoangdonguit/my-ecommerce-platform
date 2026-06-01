package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/app"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/client"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/config"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/interfaces/http"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/observability"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	tracingShutdown, err := observability.InitTracing(context.Background(), observability.TracingConfig{
		Enabled:     cfg.OTelEnabled,
		ServiceName: cfg.OTelServiceName,
		Environment: cfg.OTelEnvironment,
		Endpoint:    cfg.OTelEndpoint,
	})
	if err != nil {
		log.Fatalf("failed to initialize tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tracingShutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracing: %v", err)
		}
	}()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	orderClient := client.NewOrderClient(cfg.OrderServiceURL)
	inventoryClient := client.NewInventoryClient(cfg.InventoryServiceURL)
	paymentClient := client.NewPaymentClient(cfg.PaymentServiceURL)
	notificationClient := client.NewNotificationClient(cfg.NotificationServiceURL)
	readModelClient := client.NewReadModelClient(cfg.ReadModelServiceURL)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("WARNING: failed to connect redis cache: %v", err)
	} else {
		log.Println("redis cache connected successfully")
	}

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
		readModelClient,
		redisClient,
		sagaService,
	)

	router := httpapi.SetupRouter(handler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)
	log.Printf("order service url: %s", cfg.OrderServiceURL)
	log.Printf("inventory service url: %s", cfg.InventoryServiceURL)
	log.Printf("payment service url: %s", cfg.PaymentServiceURL)
	log.Printf("notification service url: %s", cfg.NotificationServiceURL)
	log.Printf("read model service url: %s", cfg.ReadModelServiceURL)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
