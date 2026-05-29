CREATE TABLE IF NOT EXISTS inventory_outbox_events (
  id UUID PRIMARY KEY,
  aggregate_id UUID NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  topic VARCHAR(100) NOT NULL,
  message_key VARCHAR(255) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT,
  next_attempt_at TIMESTAMP NOT NULL DEFAULT NOW(),
  published_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_inventory_outbox_aggregate_event UNIQUE (aggregate_id, event_type)
);

CREATE INDEX IF NOT EXISTS idx_inventory_outbox_status_next_attempt
ON inventory_outbox_events(status, next_attempt_at);

CREATE INDEX IF NOT EXISTS idx_inventory_outbox_aggregate_id
ON inventory_outbox_events(aggregate_id);
