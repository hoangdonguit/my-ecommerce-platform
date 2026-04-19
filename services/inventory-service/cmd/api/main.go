package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/db"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/handler"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/kafka"
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
	h := handler.NewInventoryHandler(svc)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/inventory/:productId", h.GetInventory)

	log.Printf("inventory-service api running on :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("cannot run server: %v", err)
	}
}
