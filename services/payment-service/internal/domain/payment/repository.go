package payment

import "context"

type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Payment, error)

	UpdateCompleted(ctx context.Context, id string, transactionID string, paidAt string) error
	UpdateFailed(ctx context.Context, id string, failureCode string, failureReason string) error

	CreateAttempt(ctx context.Context, attempt *PaymentAttempt) error
}
