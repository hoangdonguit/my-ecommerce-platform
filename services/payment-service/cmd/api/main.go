package main

import (
	"log"

	"github.com/gin-gonic/gin"

	paymentapp "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/app/payment"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/persistence"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/interfaces/http"
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

	repo := persistence.NewPaymentRepository(pool)

	gateway := paymentapp.NewSimulatedPaymentGateway()

	// Payment terminal events are emitted via payment_outbox_events.
	service := paymentapp.NewService(repo, nil, gateway)
	handler := httpapi.NewPaymentHandler(service)

	router := httpapi.SetupRouter(handler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
