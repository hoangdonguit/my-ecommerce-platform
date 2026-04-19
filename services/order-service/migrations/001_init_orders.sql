CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY,
  user_id VARCHAR(100) NOT NULL,
  status VARCHAR(50) NOT NULL,
  currency VARCHAR(10) NOT NULL,
  payment_method VARCHAR(50) NOT NULL,
  shipping_address TEXT NOT NULL,
  note TEXT,
  total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
  id UUID PRIMARY KEY,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id VARCHAR(100) NOT NULL,
  quantity INT NOT NULL CHECK (quantity > 0),
  unit_price NUMERIC(12,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);