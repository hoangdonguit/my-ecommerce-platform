package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/model"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*model.Order, error) {
	query := `
  SELECT id, user_id, product_id, quantity, status, idempotency_key, created_at
  FROM orders
  WHERE idempotency_key = $1
 `

	var order model.Order
	err := r.db.QueryRow(ctx, query, key).Scan(
		&order.ID,
		&order.UserID,
		&order.ProductID,
		&order.Quantity,
		&order.Status,
		&order.IdempotencyKey,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) Create(ctx context.Context, order model.Order) error {
	query := `
  INSERT INTO orders (id, user_id, product_id, quantity, status, idempotency_key)
  VALUES ($1, $2, $3, $4, $5, $6)
 `

	_, err := r.db.Exec(ctx, query,
		order.ID,
		order.UserID,
		order.ProductID,
		order.Quantity,
		order.Status,
		order.IdempotencyKey,
	)

	return err
}
