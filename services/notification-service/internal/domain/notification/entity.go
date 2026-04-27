package notification

import "time"

type Notification struct {
	ID            string
	UserID        string
	OrderID       string
	EventType     string
	Channel       string
	Recipient     string
	Title         string
	Message       string
	Status        string
	FailureReason string
	SentAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
