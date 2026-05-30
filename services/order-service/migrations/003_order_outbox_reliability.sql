ALTER TABLE outbox
  ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_error TEXT,
  ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD COLUMN IF NOT EXISTS published_at TIMESTAMP NULL,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();

UPDATE outbox
SET
  updated_at = created_at
WHERE updated_at IS NULL;

UPDATE outbox
SET
  published_at = created_at
WHERE status = 'PUBLISHED'
  AND published_at IS NULL;

UPDATE outbox
SET
  next_attempt_at = NOW()
WHERE next_attempt_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_status_next_attempt
ON outbox(status, next_attempt_at);

CREATE INDEX IF NOT EXISTS idx_outbox_aggregate_id
ON outbox(aggregate_id);

CREATE INDEX IF NOT EXISTS idx_outbox_published_pending_order
ON outbox(status, published_at, aggregate_id)
WHERE event_type = 'order.created';
