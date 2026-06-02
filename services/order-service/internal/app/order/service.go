package orderapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/infrastructure/persistence"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/observability"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/shared/errs"
	"github.com/redis/go-redis/v9"
)

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error
	PublishOrderCreatedBatch(ctx context.Context, events []OrderCreatedEvent) error
	PublishOrderCreatedBatchWithHeaders(ctx context.Context, events []OrderCreatedEvent, headersByOrderID map[string]map[string]string) error
}

type Service struct {
	repo              domainorder.Repository
	publisher         EventPublisher
	rdb               *redis.Client
	flashSaleEnabled  bool
	flashSaleProducts map[string]bool
}

func NewService(repo domainorder.Repository, publisher EventPublisher, rdb *redis.Client) *Service {
	return &Service{
		repo:              repo,
		publisher:         publisher,
		rdb:               rdb,
		flashSaleEnabled:  parseBoolEnv(os.Getenv("ENABLE_FLASH_SALE_GATE")),
		flashSaleProducts: parseFlashSaleProducts(os.Getenv("FLASH_SALE_PRODUCTS")),
	}
}

const flashSaleMissingStock = int64(-999999999)

const flashSaleReserveScript = `
local stock = redis.call("GET", KEYS[1])
if not stock then
  return -999999999
end

stock = tonumber(stock)
local qty = tonumber(ARGV[1])

if stock < qty then
  return -1
end

return redis.call("DECRBY", KEYS[1], qty)
`

type flashSaleReservation struct {
	ProductID string
	Quantity  int
	Key       string
}

func parseBoolEnv(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func parseFlashSaleProducts(value string) map[string]bool {
	result := make(map[string]bool)

	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result[item] = true
		}
	}

	return result
}

func (s *Service) isFlashSaleProduct(productID string) bool {
	if !s.flashSaleEnabled {
		return false
	}

	if s.flashSaleProducts["*"] {
		return true
	}

	return s.flashSaleProducts[productID]
}

func redisResultToInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case string:
		var parsed int64
		_, err := fmt.Sscan(v, &parsed)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func (s *Service) reserveFlashSaleStock(ctx context.Context, req CreateOrderRequest) ([]flashSaleReservation, error) {
	if !s.flashSaleEnabled || s.rdb == nil {
		return nil, nil
	}

	quantityByProduct := make(map[string]int)

	for _, item := range req.Items {
		if s.isFlashSaleProduct(item.ProductID) {
			quantityByProduct[item.ProductID] += item.Quantity
		}
	}

	if len(quantityByProduct) == 0 {
		return nil, nil
	}

	reservations := make([]flashSaleReservation, 0, len(quantityByProduct))

	for productID, quantity := range quantityByProduct {
		key := fmt.Sprintf("flashsale:stock:%s", productID)

		value, err := s.rdb.Eval(ctx, flashSaleReserveScript, []string{key}, quantity).Result()
		if err != nil {
			s.releaseFlashSaleStock(ctx, reservations)
			return nil, errs.WrapInternal(err, "failed to reserve flash sale stock")
		}

		remaining, ok := redisResultToInt64(value)
		if !ok {
			s.releaseFlashSaleStock(ctx, reservations)
			return nil, errs.Internal("invalid flash sale stock result")
		}

		if remaining == flashSaleMissingStock {
			s.releaseFlashSaleStock(ctx, reservations)
			return nil, errs.Conflict(fmt.Sprintf("flash sale stock is not initialized for product %s", productID))
		}

		if remaining < 0 {
			s.releaseFlashSaleStock(ctx, reservations)
			return nil, errs.Conflict(fmt.Sprintf("flash sale product %s is sold out", productID))
		}

		reservations = append(reservations, flashSaleReservation{
			ProductID: productID,
			Quantity:  quantity,
			Key:       key,
		})
	}

	return reservations, nil
}

func (s *Service) releaseFlashSaleStock(ctx context.Context, reservations []flashSaleReservation) {
	if s.rdb == nil || len(reservations) == 0 {
		return
	}

	for _, reservation := range reservations {
		_ = s.rdb.IncrBy(ctx, reservation.Key, int64(reservation.Quantity)).Err()
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

	// --- 3. FLASH SALE ATOMIC STOCK GATE ---
	flashSaleReservations, err := s.reserveFlashSaleStock(ctx, req)
	if err != nil {
		return nil, false, err
	}

	// --- 4. TẠO ORDER MỚI VÀ TÍNH TIỀN CHUẨN ---
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
			unitPrice = 2500000.0 // Bàn phím Keychron
		case "prod-789":
			unitPrice = 1200000.0 // Chuột Logitech
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
	outboxHeaders := observability.MarshalTraceHeaders(observability.InjectTraceHeaders(ctx))

	// Gọi hàm Create với Payload + trace headers Outbox (ghi chung transaction với đơn hàng)
	if err := s.repo.Create(ctx, ord, "order.created", outboxPayload, outboxHeaders); err != nil {
		s.releaseFlashSaleStock(ctx, flashSaleReservations)
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
	if limit > 1000 {
		limit = 1000
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
