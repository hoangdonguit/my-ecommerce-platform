package persistence

import (
	"context"

	domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"
)

func (r *PaymentRepository) FetchPendingOutboxEvents(ctx context.Context, limit int) ([]domainpayment.OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		UPDATE payment_outbox_events
		SET
			status = 'PROCESSING',
			attempts = attempts + 1,
			updated_at = NOW()
		WHERE id IN (
			SELECT id
			FROM payment_outbox_events
			WHERE (
				(status IN ('PENDING', 'FAILED') AND next_attempt_at <= NOW())
				OR
				(status = 'PROCESSING' AND updated_at < NOW() - INTERVAL '60 seconds')
			)
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING
			id,
			aggregate_id::text,
			event_type,
			topic,
			message_key,
			payload,
			status,
			attempts,
			COALESCE(last_error, ''),
			next_attempt_at,
			published_at,
			created_at,
			updated_at
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domainpayment.OutboxEvent, 0)

	for rows.Next() {
		var event domainpayment.OutboxEvent

		if err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.EventType,
			&event.Topic,
			&event.MessageKey,
			&event.Payload,
			&event.Status,
			&event.Attempts,
			&event.LastError,
			&event.NextAttemptAt,
			&event.PublishedAt,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return events, nil
}

func (r *PaymentRepository) MarkOutboxEventsPublished(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE payment_outbox_events
		SET
			status = 'PUBLISHED',
			published_at = NOW(),
			updated_at = NOW(),
			last_error = NULL
		WHERE id = ANY($1::uuid[])
	`

	_, err := r.db.Exec(ctx, query, ids)
	return err
}

func (r *PaymentRepository) MarkOutboxEventsFailed(ctx context.Context, ids []string, lastError string) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE payment_outbox_events
		SET
			status = 'FAILED',
			last_error = $2,
			next_attempt_at = NOW() + (LEAST(attempts, 12) * INTERVAL '10 seconds'),
			updated_at = NOW()
		WHERE id = ANY($1::uuid[])
	`

	_, err := r.db.Exec(ctx, query, ids, lastError)
	return err
}
