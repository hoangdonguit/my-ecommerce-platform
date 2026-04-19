package orderapp

type OrderCreatedEvent struct {
	EventType      string                  `json:"event_type"`
	OrderID        string                  `json:"order_id"`
	UserID         string                  `json:"user_id"`
	Status         string                  `json:"status"`
	Currency       string                  `json:"currency"`
	PaymentMethod  string                  `json:"payment_method"`
	TotalAmount    float64                 `json:"total_amount"`
	IdempotencyKey string                  `json:"idempotency_key"`
	Items          []OrderCreatedEventItem `json:"items"`
}

type OrderCreatedEventItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}
