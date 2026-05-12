package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxWorker struct {
	db        *pgxpool.Pool
	publisher orderapp.EventPublisher
}

func NewOutboxWorker(db *pgxpool.Pool, publisher orderapp.EventPublisher) *OutboxWorker {
	return &OutboxWorker{db: db, publisher: publisher}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	log.Println("🚀 Outbox Worker Started: Đang quét sự kiện ngầm...")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		log.Printf("[OutboxWorker-Error] Lỗi mở Transaction: %v", err)
		return
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, payload
		FROM outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 100
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		log.Printf("[OutboxWorker-Error] Lỗi truy vấn bảng outbox: %v", err)
		return
	}

	type record struct {
		ID      string
		Payload string
	}
	var records []record
	for rows.Next() {
		var rec record
		if err := rows.Scan(&rec.ID, &rec.Payload); err != nil {
			log.Printf("[OutboxWorker-Error] Scan: %v", err)
			continue
		}
		records = append(records, rec)
	}
	rows.Close()

	if len(records) == 0 {
		return
	}

	log.Printf("[OutboxWorker] 👉 Đã tóm được %d event PENDING. Đang ném lên Kafka...", len(records))

	// Build batch events
	var events []orderapp.OrderCreatedEvent
	var ids []string
	for _, rec := range records {
		var event orderapp.OrderCreatedEvent
		if err := json.Unmarshal([]byte(rec.Payload), &event); err != nil {
			log.Printf("[OutboxWorker-Error] Parse JSON %s: %v", rec.ID, err)
			continue
		}
		events = append(events, event)
		ids = append(ids, rec.ID)
	}

	// Gửi BATCH 1 lần thay vì loop
	if err := w.publisher.PublishOrderCreatedBatch(ctx, events); err != nil {
		log.Printf("[OutboxWorker-Error] Lỗi gửi batch Kafka: %v", err)
		return
	}

	log.Printf("[OutboxWorker] ✅ Đã gửi thành công %d event. Cập nhật trạng thái...", len(ids))
	_, err = tx.Exec(ctx, `UPDATE outbox SET status = 'PUBLISHED' WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		log.Printf("[OutboxWorker-Error] Lỗi update PUBLISHED: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[OutboxWorker-Error] Lỗi Commit: %v", err)
	}
}
