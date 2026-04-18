package model

import "time"

type CreateOrderRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

type Order struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ProductID      string    `json:"product_id"`
	Quantity       int       `json:"quantity"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrderCreatedEvent struct {
	EventType      string `json:"event_type"`
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	ProductID      string `json:"product_id"`
	Quantity       int    `json:"quantity"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotency_key"`
}
