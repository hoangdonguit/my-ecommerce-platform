package persistence

import (
	"context"

	domaininventory "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/domain/inventory"
	"github.com/jackc/pgx/v5"
)

func (r *InventoryRepository) CreateOutboxEvent(ctx context.Context, tx pgx.Tx, event *domaininventory.OutboxEvent) error {
	query := `
		INSERT INTO inventory_outbox_events (
			id, aggregate_id, event_type, topic, message_key, payload,
			status, attempts, next_attempt_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, 'PENDING', 0, NOW(), NOW(), NOW())
		ON CONFLICT (aggregate_id, event_type)
		DO UPDATE SET
			topic = EXCLUDED.topic,
			message_key = EXCLUDED.message_key,
			payload = EXCLUDED.payload,
			status = CASE
				WHEN inventory_outbox_events.status = 'PUBLISHED'
				THEN inventory_outbox_events.status
				ELSE 'PENDING'
			END,
			next_attempt_at = CASE
				WHEN inventory_outbox_events.status = 'PUBLISHED'
				THEN inventory_outbox_events.next_attempt_at
				ELSE NOW()
			END,
			updated_at = NOW()
	`

	_, err := tx.Exec(
		ctx,
		query,
		event.ID,
		event.AggregateID,
		event.EventType,
		event.Topic,
		event.MessageKey,
		event.Payload,
	)
	return err
}

func (r *InventoryRepository) FetchPendingOutboxEvents(ctx context.Context, limit int) ([]domaininventory.OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		UPDATE inventory_outbox_events
		SET
			status = 'PROCESSING',
			attempts = attempts + 1,
			updated_at = NOW()
		WHERE id IN (
			SELECT id
			FROM inventory_outbox_events
			WHERE status IN ('PENDING', 'FAILED')
			  AND next_attempt_at <= NOW()
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

	events := make([]domaininventory.OutboxEvent, 0)

	for rows.Next() {
		var event domaininventory.OutboxEvent
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

func (r *InventoryRepository) MarkOutboxEventPublished(ctx context.Context, id string) error {
	query := `
		UPDATE inventory_outbox_events
		SET
			status = 'PUBLISHED',
			published_at = NOW(),
			updated_at = NOW(),
			last_error = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *InventoryRepository) MarkOutboxEventFailed(ctx context.Context, id string, lastError string) error {
	query := `
		UPDATE inventory_outbox_events
		SET
			status = 'FAILED',
			last_error = $2,
			next_attempt_at = NOW() + (LEAST(attempts, 12) * INTERVAL '10 seconds'),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, lastError)
	return err
}
