package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/model"
)

type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetByProductID(ctx context.Context, productID string) (*model.Inventory, error) {
	query := `
  SELECT product_id, available_quantity, updated_at
  FROM inventories
  WHERE product_id = $1
 `

	var inv model.Inventory
	err := r.db.QueryRow(ctx, query, productID).Scan(
		&inv.ProductID,
		&inv.AvailableQuantity,
		&inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

func (r *InventoryRepository) FindReservationByOrderID(ctx context.Context, orderID string) (*model.InventoryReservation, error) {
	query := `
  SELECT id, order_id, product_id, reserved_quantity, status, created_at
  FROM inventory_reservations
  WHERE order_id = $1
 `

	var res model.InventoryReservation
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&res.ID,
		&res.OrderID,
		&res.ProductID,
		&res.ReservedQuantity,
		&res.Status,
		&res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *InventoryRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *InventoryRepository) GetInventoryForUpdate(ctx context.Context, tx pgx.Tx, productID string) (*model.Inventory, error) {
	query := `
  SELECT product_id, available_quantity, updated_at
  FROM inventories
  WHERE product_id = $1
  FOR UPDATE
 `

	var inv model.Inventory
	err := tx.QueryRow(ctx, query, productID).Scan(
		&inv.ProductID,
		&inv.AvailableQuantity,
		&inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

func (r *InventoryRepository) UpdateAvailableQuantity(ctx context.Context, tx pgx.Tx, productID string, quantity int) error {
	query := `
  UPDATE inventories
  SET available_quantity = $1, updated_at = NOW()
  WHERE product_id = $2
 `

	_, err := tx.Exec(ctx, query, quantity, productID)
	return err
}

func (r *InventoryRepository) CreateReservation(ctx context.Context, tx pgx.Tx, reservation model.InventoryReservation) error {
	query := `
  INSERT INTO inventory_reservations (id, order_id, product_id, reserved_quantity, status)
  VALUES ($1, $2, $3, $4, $5)
 `

	_, err := tx.Exec(ctx, query,
		reservation.ID,
		reservation.OrderID,
		reservation.ProductID,
		reservation.ReservedQuantity,
		reservation.Status,
	)

	return err
}
