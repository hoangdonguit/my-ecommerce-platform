-- Add trace context headers for OpenTelemetry propagation through payment outbox.
-- Backward compatible: existing rows use an empty JSON object.

ALTER TABLE payments
ADD COLUMN IF NOT EXISTS trace_headers JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE payment_outbox_events
ADD COLUMN IF NOT EXISTS headers JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE OR REPLACE FUNCTION public.enqueue_payment_outbox_event()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
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
    headers,
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
    COALESCE(NEW.trace_headers, '{}'::jsonb),
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
    headers = EXCLUDED.headers,
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
$function$;
