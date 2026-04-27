package notificationapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domainnotification "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/domain/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/infrastructure/persistence"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/shared/errs"
)

type Service struct {
	repo domainnotification.Repository
}

func NewService(repo domainnotification.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HandlePaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error {
	if strings.TrimSpace(event.OrderID) == "" {
		return errs.BadRequest("order_id is required")
	}

	if strings.TrimSpace(event.UserID) == "" {
		return errs.BadRequest("user_id is required")
	}

	title := "Thanh toán thành công"
	message := fmt.Sprintf(
		"Đơn hàng %s đã thanh toán thành công %.0f %s bằng %s.",
		event.OrderID,
		event.Amount,
		strings.ToUpper(event.Currency),
		strings.ToUpper(event.PaymentMethod),
	)

	return s.createAndSend(ctx, event.UserID, event.OrderID, event.EventType, title, message)
}

func (s *Service) HandlePaymentFailed(ctx context.Context, event PaymentFailedEvent) error {
	if strings.TrimSpace(event.OrderID) == "" {
		return errs.BadRequest("order_id is required")
	}

	if strings.TrimSpace(event.UserID) == "" {
		return errs.BadRequest("user_id is required")
	}

	reason := strings.TrimSpace(event.Reason)
	if reason == "" {
		reason = event.FailureCode
	}
	if reason == "" {
		reason = "unknown payment failure"
	}

	title := "Thanh toán thất bại"
	message := fmt.Sprintf(
		"Đơn hàng %s thanh toán thất bại. Lý do: %s.",
		event.OrderID,
		reason,
	)

	return s.createAndSend(ctx, event.UserID, event.OrderID, event.EventType, title, message)
}

func (s *Service) createAndSend(ctx context.Context, userID string, orderID string, eventType string, title string, message string) error {
	channel := domainnotification.ChannelInApp

	existing, err := s.repo.FindByOrderIDAndEventType(ctx, orderID, eventType, channel)
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && !persistence.IsNotFound(err) {
		return errs.WrapInternal(err, "failed to check existing notification")
	}

	now := time.Now()

	notification := &domainnotification.Notification{
		ID:            uuid.NewString(),
		UserID:        userID,
		OrderID:       orderID,
		EventType:     eventType,
		Channel:       channel,
		Recipient:     userID,
		Title:         title,
		Message:       message,
		Status:        domainnotification.StatusPending,
		FailureReason: "",
		SentAt:        nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return errs.WrapInternal(err, "failed to create notification")
	}

	sentAt := time.Now().UTC().Format(time.RFC3339)

	if err := s.repo.UpdateSent(ctx, notification.ID, sentAt); err != nil {
		return errs.WrapInternal(err, "failed to update notification sent")
	}

	fmt.Printf("[NOTIFICATION SENT] user=%s order=%s event=%s title=%s message=%s\n",
		userID,
		orderID,
		eventType,
		title,
		message,
	)

	return nil
}

func (s *Service) GetNotificationByID(ctx context.Context, id string) (*domainnotification.Notification, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errs.BadRequest("notification id is required")
	}

	notification, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if persistence.IsNotFound(err) {
			return nil, errs.NotFound("notification not found")
		}
		return nil, errs.WrapInternal(err, "failed to get notification")
	}

	return notification, nil
}

func (s *Service) ListByUserID(ctx context.Context, userID string, page int, limit int) ([]domainnotification.Notification, int, error) {
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

	items, total, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list notifications")
	}

	return items, total, nil
}

func (s *Service) ListByOrderID(ctx context.Context, orderID string) ([]domainnotification.Notification, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, errs.BadRequest("order_id is required")
	}

	items, err := s.repo.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to list notifications by order")
	}

	return items, nil
}
