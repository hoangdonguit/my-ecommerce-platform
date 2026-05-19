-- 002_add_outbox_and_indexes.sql
-- Purpose:
-- - Make Outbox Pattern reproducible from migrations.
-- - Add composite indexes for common order queries and outbox worker scan.
-- - Safe to run multiple times.

CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  event_type VARCHAR(100) NOT NULL,
  aggregate_id VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Query pattern: list/filter orders by status and time.
CREATE INDEX IF NOT EXISTS idx_orders_status_created_at
ON orders(status, created_at);

-- Query pattern: list user's order history by user and time.
CREATE INDEX IF NOT EXISTS idx_orders_user_created_at
ON orders(user_id, created_at);

-- Query pattern: Outbox Worker scans pending events ordered by time.
-- Existing live DB may already have idx_outbox_status_created, so reuse this name.
CREATE INDEX IF NOT EXISTS idx_outbox_status_created
ON outbox(status, created_at)
WHERE status = 'PENDING';
