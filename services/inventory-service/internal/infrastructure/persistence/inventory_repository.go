package persistence

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	domaininventory "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/domain/inventory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetInventoryByProductID(ctx context.Context, productID string) (*domaininventory.Inventory, error) {
	query := `
		SELECT product_id, sku, on_hand_quantity, reserved_quantity, available_quantity, updated_at
		FROM inventories
		WHERE product_id = $1
	`

	var inv domaininventory.Inventory
	err := r.db.QueryRow(ctx, query, productID).Scan(
		&inv.ProductID,
		&inv.SKU,
		&inv.OnHandQuantity,
		&inv.ReservedQuantity,
		&inv.AvailableQuantity,
		&inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

func (r *InventoryRepository) FindReservationByOrderID(ctx context.Context, orderID string) (*domaininventory.InventoryReservation, error) {
	query := `
		SELECT id, order_id, user_id, status, reason, created_at, updated_at
		FROM inventory_reservations
		WHERE order_id = $1
	`

	var reservation domaininventory.InventoryReservation
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&reservation.ID,
		&reservation.OrderID,
		&reservation.UserID,
		&reservation.Status,
		&reservation.Reason,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	items, err := r.findReservationItems(ctx, reservation.ID)
	if err != nil {
		return nil, err
	}

	reservation.Items = items
	return &reservation, nil
}

func (r *InventoryRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *InventoryRepository) GetInventoriesForUpdate(ctx context.Context, tx pgx.Tx, productIDs []string) (map[string]*domaininventory.Inventory, error) {
	if len(productIDs) == 0 {
		return map[string]*domaininventory.Inventory{}, nil
	}

	ids := uniqueStrings(productIDs)
	sort.Strings(ids)

	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))

	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT product_id, sku, on_hand_quantity, reserved_quantity, available_quantity, updated_at
		FROM inventories
		WHERE product_id IN (%s)
		FOR UPDATE
	`, strings.Join(placeholders, ","))

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*domaininventory.Inventory)

	for rows.Next() {
		var inv domaininventory.Inventory
		err := rows.Scan(
			&inv.ProductID,
			&inv.SKU,
			&inv.OnHandQuantity,
			&inv.ReservedQuantity,
			&inv.AvailableQuantity,
			&inv.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		invCopy := inv
		result[inv.ProductID] = &invCopy
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (r *InventoryRepository) UpdateInventoryQuantities(ctx context.Context, tx pgx.Tx, productID string, onHandQuantity, reservedQuantity, availableQuantity int) error {
	query := `
		UPDATE inventories
		SET
			on_hand_quantity = $2,
			reserved_quantity = $3,
			available_quantity = $4,
			updated_at = NOW()
		WHERE product_id = $1
	`

	cmd, err := tx.Exec(ctx, query, productID, onHandQuantity, reservedQuantity, availableQuantity)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *InventoryRepository) CreateReservation(ctx context.Context, tx pgx.Tx, reservation *domaininventory.InventoryReservation) error {
	query := `
		INSERT INTO inventory_reservations (
			id, order_id, user_id, status, reason, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tx.Exec(
		ctx,
		query,
		reservation.ID,
		reservation.OrderID,
		reservation.UserID,
		reservation.Status,
		reservation.Reason,
		reservation.CreatedAt,
		reservation.UpdatedAt,
	)

	return err
}

func (r *InventoryRepository) CreateReservationItems(ctx context.Context, tx pgx.Tx, items []domaininventory.InventoryReservationItem) error {
	query := `
		INSERT INTO inventory_reservation_items (
			id, reservation_id, product_id, requested_quantity, reserved_quantity, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, item := range items {
		_, err := tx.Exec(
			ctx,
			query,
			item.ID,
			item.ReservationID,
			item.ProductID,
			item.RequestedQuantity,
			item.ReservedQuantity,
			item.Status,
			item.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *InventoryRepository) findReservationItems(ctx context.Context, reservationID string) ([]domaininventory.InventoryReservationItem, error) {
	query := `
		SELECT id, reservation_id, product_id, requested_quantity, reserved_quantity, status, created_at
		FROM inventory_reservation_items
		WHERE reservation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domaininventory.InventoryReservationItem, 0)

	for rows.Next() {
		var item domaininventory.InventoryReservationItem
		err := rows.Scan(
			&item.ID,
			&item.ReservationID,
			&item.ProductID,
			&item.RequestedQuantity,
			&item.ReservedQuantity,
			&item.Status,
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

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))

	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}
