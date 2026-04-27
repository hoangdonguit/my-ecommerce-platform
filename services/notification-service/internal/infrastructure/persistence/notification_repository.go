package persistence

import (
	"context"
	"errors"
	"time"

	domainnotification "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/domain/notification"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domainnotification.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, order_id, event_type, channel, recipient,
			title, message, status, failure_reason, sent_at,
			created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		n.ID,
		n.UserID,
		n.OrderID,
		n.EventType,
		n.Channel,
		n.Recipient,
		n.Title,
		n.Message,
		n.Status,
		n.FailureReason,
		n.SentAt,
		n.CreatedAt,
		n.UpdatedAt,
	)

	return err
}

func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*domainnotification.Notification, error) {
	return r.findOne(ctx, `
		SELECT id, user_id, order_id, event_type, channel, recipient,
		       title, message, status, COALESCE(failure_reason, ''), sent_at,
		       created_at, updated_at
		FROM notifications
		WHERE id = $1
	`, id)
}

func (r *NotificationRepository) FindByOrderIDAndEventType(ctx context.Context, orderID string, eventType string, channel string) (*domainnotification.Notification, error) {
	var n domainnotification.Notification

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, order_id, event_type, channel, recipient,
		       title, message, status, COALESCE(failure_reason, ''), sent_at,
		       created_at, updated_at
		FROM notifications
		WHERE order_id = $1 AND event_type = $2 AND channel = $3
	`, orderID, eventType, channel).Scan(
		&n.ID,
		&n.UserID,
		&n.OrderID,
		&n.EventType,
		&n.Channel,
		&n.Recipient,
		&n.Title,
		&n.Message,
		&n.Status,
		&n.FailureReason,
		&n.SentAt,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (r *NotificationRepository) ListByUserID(ctx context.Context, userID string, limit int, offset int) ([]domainnotification.Notification, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1
	`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, order_id, event_type, channel, recipient,
		       title, message, status, COALESCE(failure_reason, ''), sent_at,
		       created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domainnotification.Notification, 0)

	for rows.Next() {
		var n domainnotification.Notification

		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.OrderID,
			&n.EventType,
			&n.Channel,
			&n.Recipient,
			&n.Title,
			&n.Message,
			&n.Status,
			&n.FailureReason,
			&n.SentAt,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		items = append(items, n)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return items, total, nil
}

func (r *NotificationRepository) ListByOrderID(ctx context.Context, orderID string) ([]domainnotification.Notification, error) {
	query := `
		SELECT id, user_id, order_id, event_type, channel, recipient,
		       title, message, status, COALESCE(failure_reason, ''), sent_at,
		       created_at, updated_at
		FROM notifications
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domainnotification.Notification, 0)

	for rows.Next() {
		var n domainnotification.Notification

		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.OrderID,
			&n.EventType,
			&n.Channel,
			&n.Recipient,
			&n.Title,
			&n.Message,
			&n.Status,
			&n.FailureReason,
			&n.SentAt,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, n)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}

func (r *NotificationRepository) UpdateSent(ctx context.Context, id string, sentAt string) error {
	parsedSentAt, err := time.Parse(time.RFC3339, sentAt)
	if err != nil {
		return err
	}

	query := `
		UPDATE notifications
		SET status = $2,
		    sent_at = $3,
		    failure_reason = '',
		    updated_at = NOW()
		WHERE id = $1
	`

	cmd, err := r.db.Exec(ctx, query, id, domainnotification.StatusSent, parsedSentAt)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *NotificationRepository) UpdateFailed(ctx context.Context, id string, reason string) error {
	query := `
		UPDATE notifications
		SET status = $2,
		    failure_reason = $3,
		    updated_at = NOW()
		WHERE id = $1
	`

	cmd, err := r.db.Exec(ctx, query, id, domainnotification.StatusFailed, reason)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *NotificationRepository) findOne(ctx context.Context, query string, arg string) (*domainnotification.Notification, error) {
	var n domainnotification.Notification

	err := r.db.QueryRow(ctx, query, arg).Scan(
		&n.ID,
		&n.UserID,
		&n.OrderID,
		&n.EventType,
		&n.Channel,
		&n.Recipient,
		&n.Title,
		&n.Message,
		&n.Status,
		&n.FailureReason,
		&n.SentAt,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
