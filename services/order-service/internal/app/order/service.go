package orderapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/persistence"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/shared/errs"
)

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error
}

type Service struct {
	repo      domainorder.Repository
	publisher EventPublisher
	rdb       *redis.Client // Thêm Redis client
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

	// --- 1. CHECK REDIS CACHE TRƯỚC (NẾU REDIS SỐNG) ---
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

	// --- 2. CHECK POSTGRES (FALLBACK) ---
	existing, err := s.repo.FindByIdempotencyKey(ctx, idemKey)
	if err == nil {
		// Cache lại vào Redis trước khi trả về (nếu redis sống)
		if s.rdb != nil {
			orderData, _ := json.Marshal(existing)
			s.rdb.Set(ctx, redisKey, orderData, 24*time.Hour)
		}
		return existing, true, nil
	}
	if err != nil && !persistence.IsNotFound(err) {
		return nil, false, errs.WrapInternal(err, "failed to check idempotency key")
	}

	// --- 3. TẠO ORDER MỚI ---
	now := time.Now()
	orderID := uuid.NewString()

	items := make([]domainorder.OrderItem, 0, len(req.Items))
	totalAmount := 0.0

	for _, reqItem := range req.Items {
		unitPrice := 100000.0 // Giá giả lập
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

	if err := s.repo.Create(ctx, ord); err != nil {
		return nil, false, errs.WrapInternal(err, "failed to create order")
	}

	created, err := s.repo.FindByID(ctx, ord.ID)
	if err != nil {
		return nil, false, errs.WrapInternal(err, "failed to reload created order")
	}

	// --- 4. CACHE KẾT QUẢ MỚI VÀO REDIS ---
	if s.rdb != nil {
		newOrderData, _ := json.Marshal(created)
		s.rdb.Set(ctx, redisKey, newOrderData, 24*time.Hour)
	}

	// --- 5. BẮN EVENT LÊN KAFKA ---
	event := buildOrderCreatedEvent(created)
	if s.publisher != nil {
		if err := s.publisher.PublishOrderCreated(ctx, event); err != nil {
			return nil, false, errs.WrapInternal(err, "failed to publish order.created event")
		}
	}

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