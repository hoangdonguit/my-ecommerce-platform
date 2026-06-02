-- Add trace context headers for OpenTelemetry propagation through order outbox.
-- Backward compatible: existing rows use an empty JSON object.

ALTER TABLE outbox
ADD COLUMN IF NOT EXISTS headers JSONB NOT NULL DEFAULT '{}'::jsonb;
