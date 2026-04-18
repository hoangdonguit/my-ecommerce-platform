package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/db"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/handler"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/kafka"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/repository"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/service"
)

func main() {
	cfg := config.Load()

	conn, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("cannot connect db: %v", err)
	}
	defer conn.Close()

	orderRepo := repository.NewOrderRepository(conn)
	publisher := kafka.NewOrderPublisher(cfg.KafkaBroker, cfg.OrderCreatedTopic)
	orderSvc := service.NewOrderService(orderRepo, publisher)
	orderHandler := handler.NewOrderHandler(orderSvc)

	router := gin.Default()
	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("order-service running on :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("cannot run server: %v", err)
	}
}
