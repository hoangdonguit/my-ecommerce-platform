package notificationapp

type PaymentCompletedEvent struct {
	EventType     string  `json:"event_type"`
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	PaymentID     string  `json:"payment_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id"`
	PaidAt        string  `json:"paid_at"`
}

type PaymentFailedEvent struct {
	EventType     string  `json:"event_type"`
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	PaymentID     string  `json:"payment_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	FailureCode   string  `json:"failure_code"`
	Reason        string  `json:"reason"`
}
