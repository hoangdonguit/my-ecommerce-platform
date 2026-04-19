package model

type OrderCreatedEvent struct {
	EventType      string `json:"event_type"`
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	ProductID      string `json:"product_id"`
	Quantity       int    `json:"quantity"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotency_key"`
}

type InventoryReservedEvent struct {
	EventType string `json:"event_type"`
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

type InventoryFailedEvent struct {
	EventType string `json:"event_type"`
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Reason    string `json:"reason"`
	Status    string `json:"status"`
}
