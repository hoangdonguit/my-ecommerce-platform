package model

import "time"

type Inventory struct {
	ProductID         string    `json:"product_id"`
	AvailableQuantity int       `json:"available_quantity"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type InventoryReservation struct {
	ID               string    `json:"id"`
	OrderID          string    `json:"order_id"`
	ProductID        string    `json:"product_id"`
	ReservedQuantity int       `json:"reserved_quantity"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}
