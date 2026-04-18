package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/model"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/repository"
)

type Publisher interface {
	PublishOrderCreated(ctx context.Context, event model.OrderCreatedEvent) error
}

type OrderService struct {
	repo      *repository.OrderRepository
	publisher Publisher
}

func NewOrderService(repo *repository.OrderRepository, publisher Publisher) *OrderService {
	return &OrderService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest, idemKey string) (*model.Order, bool, error) {
	if req.Quantity <= 0 {
		return nil, false, errors.New("quantity must be greater than 0")
	}

	if idemKey == "" {
		return nil, false, errors.New("missing X-Idempotency-Key")
	}

	existing, err := s.repo.FindByIdempotencyKey(ctx, idemKey)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	order := model.Order{
		ID:             uuid.NewString(),
		UserID:         req.UserID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		Status:         "PENDING",
		IdempotencyKey: idemKey,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, false, err
	}

	event := model.OrderCreatedEvent{
		EventType:      "order.created",
		OrderID:        order.ID,
		UserID:         order.UserID,
		ProductID:      order.ProductID,
		Quantity:       order.Quantity,
		Status:         order.Status,
		IdempotencyKey: order.IdempotencyKey,
	}

	if err := s.publisher.PublishOrderCreated(ctx, event); err != nil {
		return nil, false, err
	}

	return &order, false, nil
}
