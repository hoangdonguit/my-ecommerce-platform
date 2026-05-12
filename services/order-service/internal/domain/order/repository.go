package order

import "context"

type Repository interface {
	Create(ctx context.Context, order *Order, outboxEventType string, outboxPayload []byte) error
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Order, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]Order, int, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}