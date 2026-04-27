package payment

import "time"

type Payment struct {
	ID             string
	OrderID        string
	UserID         string
	Amount         float64
	Currency       string
	PaymentMethod  string
	Status         string
	FailureCode    string
	FailureReason  string
	TransactionID  string
	IdempotencyKey string
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PaymentAttempt struct {
	ID                   string
	PaymentID            string
	OrderID              string
	Status               string
	GatewayTransactionID string
	FailureCode          string
	FailureReason        string
	RawResponse          string
	CreatedAt            time.Time
}
