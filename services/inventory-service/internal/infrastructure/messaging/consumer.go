package messaging

import (
	"context"
	"encoding/json"
	"log"

	inventoryapp "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/app/inventory"
	"github.com/segmentio/kafka-go"
)

type InventoryConsumer struct {
	reader  *kafka.Reader
	service *inventoryapp.Service
}

func NewInventoryConsumer(brokers []string, service *inventoryapp.Service) *InventoryConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     "inventory-rollback-group-v2", // Đổi group để reset offset
		GroupTopics: []string{"payment.failed", "order.cancelled"},
	})
	return &InventoryConsumer{
		reader:  r,
		service: service,
	}
}

func (c *InventoryConsumer) Start(ctx context.Context) {
	log.Println("🔥 Inventory Consumer: Đang trực chiến Rollback...")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			continue
		}

		// Đọc linh hoạt cả snake_case và camelCase
		var event struct {
			OrderID  string `json:"order_id"`
			OrderID2 string `json:"orderId"`
		}

		if err := json.Unmarshal(m.Value, &event); err == nil {
			oid := event.OrderID
			if oid == "" { oid = event.OrderID2 }
			
			if oid != "" {
				log.Printf("⚠️ Bắt được lệnh Rollback cho đơn [%s]", oid)
				c.service.RollbackInventory(ctx, oid)
			}
		}
	}
}