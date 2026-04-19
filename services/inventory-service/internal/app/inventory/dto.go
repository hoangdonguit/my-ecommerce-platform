package inventoryapp

import "time"

type InventoryResponse struct {
	ProductID         string    `json:"product_id"`
	SKU               string    `json:"sku"`
	OnHandQuantity    int       `json:"on_hand_quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ReservationItemResponse struct {
	ProductID         string `json:"product_id"`
	RequestedQuantity int    `json:"requested_quantity"`
	ReservedQuantity  int    `json:"reserved_quantity"`
	Status            string `json:"status"`
}

type ReservationResponse struct {
	ID        string                    `json:"id"`
	OrderID   string                    `json:"order_id"`
	UserID    string                    `json:"user_id"`
	Status    string                    `json:"status"`
	Reason    string                    `json:"reason,omitempty"`
	Items     []ReservationItemResponse `json:"items"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}
