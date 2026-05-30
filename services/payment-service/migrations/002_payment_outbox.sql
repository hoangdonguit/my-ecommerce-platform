CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS payment_outbox_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
  CONSTRAINT uq_payment_outbox_aggregate_event UNIQUE (aggregate_id, event_type)
);

CREATE INDEX IF NOT EXISTS idx_payment_outbox_status_next_attempt
ON payment_outbox_events(status, next_attempt_at);

CREATE INDEX IF NOT EXISTS idx_payment_outbox_aggregate_id
ON payment_outbox_events(aggregate_id);

CREATE OR REPLACE FUNCTION enqueue_payment_outbox_event()
RETURNS TRIGGER AS $$
DECLARE
  v_event_type TEXT;
  v_topic TEXT;
  v_payload JSONB;
BEGIN
  IF NEW.status NOT IN ('COMPLETED', 'FAILED') THEN
    RETURN NEW;
  END IF;

  IF TG_OP = 'UPDATE' AND OLD.status = NEW.status THEN
    RETURN NEW;
  END IF;

  IF NEW.status = 'COMPLETED' THEN
    v_event_type := 'payment.completed';
    v_topic := 'payment.completed';

    v_payload := jsonb_build_object(
      'event_type', 'payment.completed',
      'order_id', NEW.order_id::text,
      'user_id', NEW.user_id,
      'payment_id', NEW.id::text,
      'amount', NEW.amount,
      'currency', NEW.currency,
      'payment_method', NEW.payment_method,
      'status', NEW.status,
      'transaction_id', COALESCE(NEW.transaction_id, ''),
      'paid_at', COALESCE(to_char(NEW.paid_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')
    );
  ELSE
    v_event_type := 'payment.failed';
    v_topic := 'payment.failed';

    v_payload := jsonb_build_object(
      'event_type', 'payment.failed',
      'order_id', NEW.order_id::text,
      'user_id', NEW.user_id,
      'payment_id', NEW.id::text,
      'amount', NEW.amount,
      'currency', NEW.currency,
      'payment_method', NEW.payment_method,
      'status', NEW.status,
      'failure_code', COALESCE(NEW.failure_code, ''),
      'reason', COALESCE(NEW.failure_reason, ''),
      'failure_reason', COALESCE(NEW.failure_reason, '')
    );
  END IF;

  INSERT INTO payment_outbox_events (
    aggregate_id,
    event_type,
    topic,
    message_key,
    payload,
    status,
    attempts,
    next_attempt_at,
    created_at,
    updated_at
  )
  VALUES (
    NEW.id,
    v_event_type,
    v_topic,
    NEW.order_id::text,
    v_payload,
    'PENDING',
    0,
    NOW(),
    NOW(),
    NOW()
  )
  ON CONFLICT (aggregate_id, event_type)
  DO UPDATE SET
    topic = EXCLUDED.topic,
    message_key = EXCLUDED.message_key,
    payload = EXCLUDED.payload,
    status = CASE
      WHEN payment_outbox_events.status = 'PUBLISHED'
      THEN payment_outbox_events.status
      ELSE 'PENDING'
    END,
    next_attempt_at = CASE
      WHEN payment_outbox_events.status = 'PUBLISHED'
      THEN payment_outbox_events.next_attempt_at
      ELSE NOW()
    END,
    updated_at = NOW();

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_payments_terminal_outbox ON payments;

CREATE TRIGGER trg_payments_terminal_outbox
AFTER INSERT OR UPDATE OF status ON payments
FOR EACH ROW
EXECUTE FUNCTION enqueue_payment_outbox_event();
