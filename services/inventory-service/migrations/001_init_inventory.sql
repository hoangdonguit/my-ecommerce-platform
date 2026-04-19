CREATE TABLE IF NOT EXISTS inventories (
  product_id VARCHAR(100) PRIMARY KEY,
  sku VARCHAR(100) UNIQUE,
  on_hand_quantity INT NOT NULL DEFAULT 0 CHECK (on_hand_quantity >= 0),
  reserved_quantity INT NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
  available_quantity INT NOT NULL DEFAULT 0 CHECK (available_quantity >= 0),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_reservations (
  id UUID PRIMARY KEY,
  order_id UUID NOT NULL UNIQUE,
  user_id VARCHAR(100) NOT NULL,
  status VARCHAR(50) NOT NULL,
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_reservation_items (
  id UUID PRIMARY KEY,
  reservation_id UUID NOT NULL REFERENCES inventory_reservations(id) ON DELETE CASCADE,
  product_id VARCHAR(100) NOT NULL,
  requested_quantity INT NOT NULL CHECK (requested_quantity > 0),
  reserved_quantity INT NOT NULL CHECK (reserved_quantity >= 0),
  status VARCHAR(50) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_reservations_order_id
ON inventory_reservations(order_id);

CREATE INDEX IF NOT EXISTS idx_inventory_reservations_status
ON inventory_reservations(status);

CREATE INDEX IF NOT EXISTS idx_inventory_reservation_items_reservation_id
ON inventory_reservation_items(reservation_id);

CREATE INDEX IF NOT EXISTS idx_inventory_reservation_items_product_id
ON inventory_reservation_items(product_id);