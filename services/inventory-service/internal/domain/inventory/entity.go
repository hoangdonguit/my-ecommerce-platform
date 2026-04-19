package inventory

import "time"

type Inventory struct {
	ProductID         string
	SKU               string
	OnHandQuantity    int
	ReservedQuantity  int
	AvailableQuantity int
	UpdatedAt         time.Time
}

type InventoryReservation struct {
	ID        string
	OrderID   string
	UserID    string
	Status    string
	Reason    string
	Items     []InventoryReservationItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type InventoryReservationItem struct {
	ID                string
	ReservationID     string
	ProductID         string
	RequestedQuantity int
	ReservedQuantity  int
	Status            string
	CreatedAt         time.Time
}
