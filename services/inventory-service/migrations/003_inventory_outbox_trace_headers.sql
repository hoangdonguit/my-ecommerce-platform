-- Add trace context headers for OpenTelemetry propagation through inventory outbox.
-- Backward compatible: existing rows use an empty JSON object.

ALTER TABLE inventory_outbox_events
ADD COLUMN IF NOT EXISTS headers JSONB NOT NULL DEFAULT '{}'::jsonb;
