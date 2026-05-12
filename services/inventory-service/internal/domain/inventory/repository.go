package inventory

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	GetInventoryByProductID(ctx context.Context, productID string) (*Inventory, error)
	FindReservationByOrderID(ctx context.Context, orderID string) (*InventoryReservation, error)

	BeginTx(ctx context.Context) (pgx.Tx, error)

	GetInventoriesForUpdate(ctx context.Context, tx pgx.Tx, productIDs []string) (map[string]*Inventory, error)
	UpdateInventoryQuantities(ctx context.Context, tx pgx.Tx, productID string, onHandQuantity, reservedQuantity, availableQuantity int) error

	CreateReservation(ctx context.Context, tx pgx.Tx, reservation *InventoryReservation) error
	CreateReservationItems(ctx context.Context, tx pgx.Tx, items []InventoryReservationItem) error

	ListAllInventories(ctx context.Context) ([]Inventory, error)
	UpdateReservationStatus(ctx context.Context, tx pgx.Tx, reservationID string, status string) error
	AtomicReserveInventory(ctx context.Context, tx pgx.Tx, productID string, quantity int) error
}
