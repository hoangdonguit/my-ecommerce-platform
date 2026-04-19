package persistence

import (
	"context"
	"errors"

	domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, ord *domainorder.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders (
			id, user_id, status, currency, payment_method,
			shipping_address, note, total_amount, idempotency_key,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = tx.Exec(
		ctx,
		orderQuery,
		ord.ID,
		ord.UserID,
		ord.Status,
		ord.Currency,
		ord.PaymentMethod,
		ord.ShippingAddress,
		ord.Note,
		ord.TotalAmount,
		ord.IdempotencyKey,
		ord.CreatedAt,
		ord.UpdatedAt,
	)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO order_items (
			id, order_id, product_id, quantity, unit_price, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, item := range ord.Items {
		_, err = tx.Exec(
			ctx,
			itemQuery,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domainorder.Order, error) {
	orderQuery := `
		SELECT
			id, user_id, status, currency, payment_method,
			shipping_address, note, total_amount, idempotency_key,
			created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var ord domainorder.Order

	err := r.db.QueryRow(ctx, orderQuery, id).Scan(
		&ord.ID,
		&ord.UserID,
		&ord.Status,
		&ord.Currency,
		&ord.PaymentMethod,
		&ord.ShippingAddress,
		&ord.Note,
		&ord.TotalAmount,
		&ord.IdempotencyKey,
		&ord.CreatedAt,
		&ord.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	items, err := r.findItemsByOrderID(ctx, ord.ID)
	if err != nil {
		return nil, err
	}

	ord.Items = items
	return &ord, nil
}

func (r *OrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domainorder.Order, error) {
	query := `
		SELECT id
		FROM orders
		WHERE idempotency_key = $1
	`

	var orderID string
	err := r.db.QueryRow(ctx, query, key).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, orderID)
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domainorder.Order, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE user_id = $1
	`

	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id, user_id, status, currency, payment_method,
			shipping_address, note, total_amount, idempotency_key,
			created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]domainorder.Order, 0)

	for rows.Next() {
		var ord domainorder.Order

		err := rows.Scan(
			&ord.ID,
			&ord.UserID,
			&ord.Status,
			&ord.Currency,
			&ord.PaymentMethod,
			&ord.ShippingAddress,
			&ord.Note,
			&ord.TotalAmount,
			&ord.IdempotencyKey,
			&ord.CreatedAt,
			&ord.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		items, err := r.findItemsByOrderID(ctx, ord.ID)
		if err != nil {
			return nil, 0, err
		}

		ord.Items = items
		orders = append(orders, ord)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *OrderRepository) findItemsByOrderID(ctx context.Context, orderID string) ([]domainorder.OrderItem, error) {
	query := `
		SELECT
			id, order_id, product_id, quantity, unit_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domainorder.OrderItem, 0)

	for rows.Next() {
		var item domainorder.OrderItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
