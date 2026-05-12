package messaging

import (
	"context"
	"encoding/json"
	"log"

	domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"
	"github.com/segmentio/kafka-go"
)

type OrderConsumer struct {
	reader *kafka.Reader
	repo   domainorder.Repository
}

func NewOrderConsumer(brokers []string, repo domainorder.Repository) *OrderConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     "order-service-saga-monitor",
		// Đã bổ sung payment.completed vào danh sách hóng chuyện
		GroupTopics: []string{"inventory.failed", "payment.failed", "payment.completed"},
	})
	return &OrderConsumer{
		reader: r,
		repo:   repo,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) {
	log.Println("🔥 Order Consumer Started: Đang hóng kết quả Saga từ Kafka...")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Consumer error: %v\n", err)
			if ctx.Err() != nil {
				return
			}
			continue
		}

		// Hỗ trợ cả 2 định dạng JSON (Snake case và Camel case) để không trượt event
		var event struct {
			OrderID  string `json:"order_id"`
			OrderID2 string `json:"orderId"`
		}

		if err := json.Unmarshal(m.Value, &event); err == nil {
			oid := event.OrderID
			if oid == "" {
				oid = event.OrderID2
			}

			if oid != "" {
				var targetStatus string

				// Xác định trạng thái chuẩn xác dựa vào tên Topic
				switch m.Topic {
				case "payment.completed":
					targetStatus = "COMPLETED"
				case "inventory.failed", "payment.failed":
					targetStatus = "FAILED"
				default:
					continue
				}

				log.Printf("⚠️ Bắt được sự kiện từ Topic [%s]. Chuyển Order [%s] sang %s...\n", m.Topic, oid, targetStatus)

				// ÉP Database chuyển trạng thái Order
				err := c.repo.UpdateStatus(ctx, oid, targetStatus)
				if err != nil {
					log.Printf("❌ Cập nhật DB thất bại cho Order [%s]: %v\n", oid, err)
				} else {
					log.Printf("✅ Đã tự động cập nhật Order [%s] thành %s!\n", oid, targetStatus)
				}
			}
		}
	}
}
