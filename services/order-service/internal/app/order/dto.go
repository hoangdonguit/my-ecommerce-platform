package orderapp

import "time"

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	UserID          string                   `json:"user_id" binding:"required"`
	Items           []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
	Currency        string                   `json:"currency" binding:"required"`
	PaymentMethod   string                   `json:"payment_method" binding:"required"`
	ShippingAddress string                   `json:"shipping_address" binding:"required"`
	Note            string                   `json:"note"`
}

type OrderItemResponse struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderResponse struct {
	ID              string              `json:"id"`
	UserID          string              `json:"user_id"`
	Status          string              `json:"status"`
	Currency        string              `json:"currency"`
	PaymentMethod   string              `json:"payment_method"`
	ShippingAddress string              `json:"shipping_address"`
	Note            string              `json:"note,omitempty"`
	Items           []OrderItemResponse `json:"items"`
	TotalAmount     float64             `json:"total_amount"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}
