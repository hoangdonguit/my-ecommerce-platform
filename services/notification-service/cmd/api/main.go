package main

import (
	"log"

	"github.com/gin-gonic/gin"

	notificationapp "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/app/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/persistence"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/interfaces/http"
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

	repo := persistence.NewNotificationRepository(pool)
	service := notificationapp.NewService(repo)
	handler := httpapi.NewNotificationHandler(service)

	router := httpapi.SetupRouter(handler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
