package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/config"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/db"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/persistence"
)

func main() {
	cfg := config.Load()

	pool, err := db.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()

	repo := persistence.NewOrderRepository(pool)

	order, err := repo.FindByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err != nil {
		log.Fatalf("failed to find order: %v", err)
	}

	fmt.Printf("ORDER: %+v\n", order)
	fmt.Printf("ITEMS: %+v\n", order.Items)
}
