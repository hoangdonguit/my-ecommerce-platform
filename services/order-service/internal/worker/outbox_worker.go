package worker

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxWorker struct {
	db                         *pgxpool.Pool
	publisher                  orderapp.EventPublisher
	batchSize                  int
	interval                   time.Duration
	staleProcessingAfter       int
	requeueStalePublished      bool
	requeuePublishedAfter      int
	lastPublishedRequeueAt     time.Time
	publishedRequeueCheckEvery time.Duration
}

func NewOutboxWorker(db *pgxpool.Pool, publisher orderapp.EventPublisher) *OutboxWorker {
	return &OutboxWorker{
		db:                         db,
		publisher:                  publisher,
		batchSize:                  getEnvInt("ORDER_OUTBOX_BATCH_SIZE", 200),
		interval:                   time.Duration(getEnvInt("ORDER_OUTBOX_INTERVAL_MS", 200)) * time.Millisecond,
		staleProcessingAfter:       getEnvInt("ORDER_OUTBOX_STALE_PROCESSING_SECONDS", 60),
		requeueStalePublished:      getEnvBool("ORDER_OUTBOX_REQUEUE_STALE_PUBLISHED", true),
		requeuePublishedAfter:      getEnvInt("ORDER_OUTBOX_REQUEUE_AFTER_SECONDS", 120),
		publishedRequeueCheckEvery: time.Duration(getEnvInt("ORDER_OUTBOX_REQUEUE_CHECK_MS", 5000)) * time.Millisecond,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	log.Printf(
		"order outbox worker started batch_size=%d interval=%s stale_processing_after=%ds requeue_published=%v requeue_after=%ds",
		w.batchSize,
		w.interval,
		w.staleProcessingAfter,
		w.requeueStalePublished,
		w.requeuePublishedAfter,
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.requeueStalePublishedOrders(ctx)
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		log.Printf("[OrderOutboxWorker] begin transaction failed err=%v", err)
		return
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		UPDATE outbox
		SET
			status = 'PROCESSING',
			attempts = attempts + 1,
			updated_at = NOW()
		WHERE id IN (
			SELECT id
			FROM outbox
			WHERE (
				(status IN ('PENDING', 'FAILED') AND next_attempt_at <= NOW())
				OR
				(status = 'PROCESSING' AND updated_at < NOW() - ($2::int * INTERVAL '1 second'))
			)
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id::text, payload::text
	`, w.batchSize, w.staleProcessingAfter)
	if err != nil {
		log.Printf("[OrderOutboxWorker] fetch events failed err=%v", err)
		return
	}

	type record struct {
		ID      string
		Payload string
	}

	records := make([]record, 0, w.batchSize)

	for rows.Next() {
		var rec record
		if err := rows.Scan(&rec.ID, &rec.Payload); err != nil {
			log.Printf("[OrderOutboxWorker] scan failed err=%v", err)
			continue
		}
		records = append(records, rec)
	}

	rows.Close()

	if rows.Err() != nil {
		log.Printf("[OrderOutboxWorker] rows error err=%v", rows.Err())
		return
	}

	if len(records) == 0 {
		return
	}

	events := make([]orderapp.OrderCreatedEvent, 0, len(records))
	ids := make([]string, 0, len(records))
	parseFailedIDs := make([]string, 0)

	for _, rec := range records {
		var event orderapp.OrderCreatedEvent
		if err := json.Unmarshal([]byte(rec.Payload), &event); err != nil {
			log.Printf("[OrderOutboxWorker] parse json failed id=%s err=%v", rec.ID, err)
			parseFailedIDs = append(parseFailedIDs, rec.ID)
			continue
		}

		events = append(events, event)
		ids = append(ids, rec.ID)
	}

	if len(parseFailedIDs) > 0 {
		if err := markFailedTx(ctx, tx, parseFailedIDs, "failed to parse outbox payload"); err != nil {
			log.Printf("[OrderOutboxWorker] mark parse failed failed err=%v", err)
			return
		}
	}

	if len(events) == 0 {
		if err := tx.Commit(ctx); err != nil {
			log.Printf("[OrderOutboxWorker] commit parse-only failed err=%v", err)
		}
		return
	}

	publishCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err = w.publisher.PublishOrderCreatedBatch(publishCtx, events)
	cancel()

	if err != nil {
		log.Printf("[OrderOutboxWorker] publish batch failed count=%d err=%v", len(events), err)

		if markErr := markFailedTx(ctx, tx, ids, err.Error()); markErr != nil {
			log.Printf("[OrderOutboxWorker] mark publish failed failed err=%v", markErr)
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			log.Printf("[OrderOutboxWorker] commit failed-state failed err=%v", commitErr)
		}
		return
	}

	_, err = tx.Exec(ctx, `
		UPDATE outbox
		SET
			status = 'PUBLISHED',
			published_at = NOW(),
			updated_at = NOW(),
			last_error = NULL
		WHERE id = ANY($1::uuid[])
	`, ids)
	if err != nil {
		log.Printf("[OrderOutboxWorker] mark published failed count=%d err=%v", len(ids), err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[OrderOutboxWorker] commit failed count=%d err=%v", len(ids), err)
		return
	}

	log.Printf(
		"[OrderOutboxWorker] batch published count=%d first_order_id=%s last_order_id=%s",
		len(events),
		events[0].OrderID,
		events[len(events)-1].OrderID,
	)
}

func markFailedTx(ctx context.Context, tx pgx.Tx, ids []string, lastError string) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := tx.Exec(ctx, `
		UPDATE outbox
		SET
			status = 'FAILED',
			last_error = $2,
			next_attempt_at = NOW() + (LEAST(attempts, 12) * INTERVAL '10 seconds'),
			updated_at = NOW()
		WHERE id = ANY($1::uuid[])
	`, ids, lastError)

	return err
}

func (w *OutboxWorker) requeueStalePublishedOrders(ctx context.Context) {
	if !w.requeueStalePublished {
		return
	}

	if !w.lastPublishedRequeueAt.IsZero() && time.Since(w.lastPublishedRequeueAt) < w.publishedRequeueCheckEvery {
		return
	}

	w.lastPublishedRequeueAt = time.Now()

	rows, err := w.db.Query(ctx, `
		UPDATE outbox ob
		SET
			status = 'PENDING',
			next_attempt_at = NOW(),
			last_error = 'requeued because order is still PENDING after published',
			updated_at = NOW()
		FROM orders o
		WHERE ob.aggregate_id = o.id::text
		  AND ob.event_type = 'order.created'
		  AND ob.status = 'PUBLISHED'
		  AND o.status = 'PENDING'
		  AND ob.published_at < NOW() - ($1::int * INTERVAL '1 second')
		RETURNING ob.aggregate_id
	`, w.requeuePublishedAfter)
	if err != nil {
		log.Printf("[OrderOutboxWorker] requeue stale published failed err=%v", err)
		return
	}
	defer rows.Close()

	count := 0
	first := ""
	last := ""

	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			continue
		}

		if count == 0 {
			first = orderID
		}
		last = orderID
		count++
	}

	if count > 0 {
		log.Printf("[OrderOutboxWorker] requeued stale published orders count=%d first_order_id=%s last_order_id=%s", count, first, last)
	}
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func getEnvBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	switch raw {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
