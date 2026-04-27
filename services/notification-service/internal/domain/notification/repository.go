package notification

import "context"

type Repository interface {
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	FindByOrderIDAndEventType(ctx context.Context, orderID string, eventType string, channel string) (*Notification, error)
	ListByUserID(ctx context.Context, userID string, limit int, offset int) ([]Notification, int, error)
	ListByOrderID(ctx context.Context, orderID string) ([]Notification, error)
	UpdateSent(ctx context.Context, id string, sentAt string) error
	UpdateFailed(ctx context.Context, id string, reason string) error
}
