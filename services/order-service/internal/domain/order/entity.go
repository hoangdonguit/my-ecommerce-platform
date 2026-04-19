package order

import "time"

type Order struct {
	ID              string
	UserID          string
	Status          string
	Currency        string
	PaymentMethod   string
	ShippingAddress string
	Note            string
	TotalAmount     float64
	IdempotencyKey  string
	Items           []OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	UnitPrice float64
	CreatedAt time.Time
}
