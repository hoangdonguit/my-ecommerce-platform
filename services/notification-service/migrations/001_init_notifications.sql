CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY,
  user_id VARCHAR(100) NOT NULL,
  order_id UUID NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  channel VARCHAR(50) NOT NULL,
  recipient VARCHAR(255),
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  status VARCHAR(50) NOT NULL,
  failure_reason TEXT,
  sent_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_order_id ON notifications(order_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);