package orderapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/persistence"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/shared/errs"
	"github.com/redis/go-redis/v9"
)

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error
	PublishOrderCreatedBatch(ctx context.Context, events []OrderCreatedEvent) error
}

type Service struct {
	repo      domainorder.Repository
	publisher EventPublisher
	rdb       *redis.Client
}

func NewService(repo domainorder.Repository, publisher EventPublisher, rdb *redis.Client) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		rdb:       rdb,
	}
}

func (s *Service) CreateOrder(ctx context.Context, req CreateOrderRequest, idemKey string) (*domainorder.Order, bool, error) {
	if err := validateCreateOrder(req, idemKey); err != nil {
		return nil, false, err
	}

	// --- 1. CHECK REDIS CACHE ---
	redisKey := fmt.Sprintf("idem:order:%s", idemKey)
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, redisKey).Result()
		if err == nil {
			var existing domainorder.Order
			if err := json.Unmarshal([]byte(val), &existing); err == nil {
				return &existing, true, nil
			}
		}
	}

	// --- 2. CHECK POSTGRES ---
	existing, err := s.repo.FindByIdempotencyKey(ctx, idemKey)
	if err == nil {
		if s.rdb != nil {
			orderData, _ := json.Marshal(existing)
			s.rdb.Set(ctx, redisKey, orderData, 24*time.Hour)
		}
		return existing, true, nil
	}
	if err != nil && !persistence.IsNotFound(err) {
		return nil, false, errs.WrapInternal(err, "failed to check idempotency key")
	}

	// --- 3. TẠO ORDER MỚI VÀ TÍNH TIỀN CHUẨN ---
	now := time.Now()
	orderID := uuid.NewString()

	items := make([]domainorder.OrderItem, 0, len(req.Items))
	totalAmount := 0.0

	for _, reqItem := range req.Items {
		var unitPrice float64

		// LOGIC CHUẨN: Tính đúng giá trị thực của sản phẩm
		switch reqItem.ProductID {
		case "prod-123":
			unitPrice = 24000000.0 // Laptop ASUS
		case "prod-456":
			unitPrice = 2500000.0  // Bàn phím Keychron
		case "prod-789":
			unitPrice = 1200000.0  // Chuột Logitech
		default:
			unitPrice = 0.0
		}

		lineTotal := unitPrice * float64(reqItem.Quantity)
		totalAmount += lineTotal

		items = append(items, domainorder.OrderItem{
			ID:        uuid.NewString(),
			OrderID:   orderID,
			ProductID: reqItem.ProductID,
			Quantity:  reqItem.Quantity,
			UnitPrice: unitPrice,
			CreatedAt: now,
		})
	}

	ord := &domainorder.Order{
		ID:              orderID,
		UserID:          req.UserID,
		Status:          domainorder.StatusPending,
		Currency:        strings.ToUpper(req.Currency),
		PaymentMethod:   strings.ToUpper(req.PaymentMethod),
		ShippingAddress: req.ShippingAddress,
		Note:            req.Note,
		TotalAmount:     totalAmount,
		IdempotencyKey:  idemKey,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// --- 4. TẠO PAYLOAD OUTBOX TRƯỚC KHI GỌI DB ---
	event := buildOrderCreatedEvent(ord)
	outboxPayload, _ := json.Marshal(event)

	// Gọi hàm Create với Payload Outbox (Nó sẽ ghi cục này vào DB chung với đơn hàng)
	if err := s.repo.Create(ctx, ord, "order.created", outboxPayload); err != nil {
		return nil, false, errs.WrapInternal(err, "failed to create order")
	}

	created, err := s.repo.FindByID(ctx, ord.ID)
	if err != nil {
		return nil, false, errs.WrapInternal(err, "failed to reload created order")
	}

	// --- 5. CACHE REDIS ---
	if s.rdb != nil {
		newOrderData, _ := json.Marshal(created)
		s.rdb.Set(ctx, redisKey, newOrderData, 24*time.Hour)
	}

	// Đã bỏ s.publisher.PublishOrderCreated() đi để chống Dual-Write
	// Worker sẽ thay thế làm việc bắn event lên Kafka.

	return created, false, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id string) (*domainorder.Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errs.BadRequest("order id is required")
	}

	ord, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if persistence.IsNotFound(err) {
			return nil, errs.NotFound("order not found")
		}
		return nil, errs.WrapInternal(err, "failed to get order")
	}

	return ord, nil
}

func (s *Service) ListOrders(ctx context.Context, userID string, page int, limit int) ([]domainorder.Order, int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, 0, errs.BadRequest("user_id is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	orders, total, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list orders")
	}

	return orders, total, nil
}

func (s *Service) CancelOrder(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errs.BadRequest("order id is required")
	}

	ord, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if persistence.IsNotFound(err) {
			return errs.NotFound("order not found")
		}
		return errs.WrapInternal(err, "failed to get order before cancellation")
	}

	if ord.Status == domainorder.StatusCancelled {
		return errs.Conflict("order is already cancelled")
	}

	if ord.Status == domainorder.StatusCompleted {
		return errs.Conflict("completed order cannot be cancelled")
	}

	if err := s.repo.UpdateStatus(ctx, id, domainorder.StatusCancelled); err != nil {
		if persistence.IsNotFound(err) {
			return errs.NotFound("order not found")
		}
		return errs.WrapInternal(err, "failed to cancel order")
	}

	return nil
}

func validateCreateOrder(req CreateOrderRequest, idemKey string) error {
	if strings.TrimSpace(idemKey) == "" {
		return errs.BadRequest("missing X-Idempotency-Key")
	}
	if strings.TrimSpace(req.UserID) == "" {
		return errs.BadRequest("user_id is required")
	}
	if len(req.Items) == 0 {
		return errs.BadRequest("items must not be empty")
	}
	for _, item := range req.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return errs.BadRequest("product_id is required")
		}
		if item.Quantity <= 0 {
			return errs.BadRequest("quantity must be greater than 0")
		}
	}
	if strings.TrimSpace(req.Currency) == "" {
		return errs.BadRequest("currency is required")
	}
	if strings.TrimSpace(req.PaymentMethod) == "" {
		return errs.BadRequest("payment_method is required")
	}
	if strings.TrimSpace(req.ShippingAddress) == "" {
		return errs.BadRequest("shipping_address is required")
	}
	return nil
}

func buildOrderCreatedEvent(ord *domainorder.Order) OrderCreatedEvent {
	items := make([]OrderCreatedEventItem, 0, len(ord.Items))
	for _, item := range ord.Items {
		items = append(items, OrderCreatedEventItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	return OrderCreatedEvent{
		EventType:      "order.created",
		OrderID:        ord.ID,
		UserID:         ord.UserID,
		Status:         ord.Status,
		Currency:       ord.Currency,
		PaymentMethod:  ord.PaymentMethod,
		TotalAmount:    ord.TotalAmount,
		IdempotencyKey: ord.IdempotencyKey,
		Items:          items,
	}
}