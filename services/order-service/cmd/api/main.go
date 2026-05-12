package main

import (
	"context"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/messaging"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/persistence"
	httpapi "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/interfaces/http"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/worker" // IMPORT WORKER Ở ĐÂY
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
	log.Println("postgres connected successfully")

	// 2. Kết nối Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("WARNING: failed to connect redis, idempotency will fallback to postgres: %v", err)
	} else {
		log.Println("redis connected successfully")
	}
	defer rdb.Close()

	// 3. Khởi tạo các Dependencies
	orderRepo := persistence.NewOrderRepository(pool)
	orderPublisher := messaging.NewOrderPublisher(cfg.KafkaBroker, cfg.OrderCreatedTopic)
	defer orderPublisher.Close()

	// --- KÍCH HOẠT KAFKA CONSUMER TỰ ĐỘNG BẮT LỖI SAGA ---
	kafkaBrokers := strings.Split(cfg.KafkaBroker, ",")
	orderConsumer := messaging.NewOrderConsumer(kafkaBrokers, orderRepo)
	go orderConsumer.Start(context.Background())
	// -----------------------------------------------------

	// --- KÍCH HOẠT OUTBOX WORKER ---
	outboxWorker := worker.NewOutboxWorker(pool, orderPublisher)
	go outboxWorker.Start(context.Background())
	// -------------------------------

	// TRUYỀN REDIS VÀO SERVICE Ở ĐÂY
	orderService := orderapp.NewService(orderRepo, orderPublisher, rdb)
	orderHandler := httpapi.NewOrderHandler(orderService, rdb)

	router := httpapi.SetupRouter(orderHandler)

	log.Printf("%s starting on port %s", cfg.AppName, cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}