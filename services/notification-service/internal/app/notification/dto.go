package notificationapp

import "time"

type NotificationResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	OrderID       string     `json:"order_id"`
	EventType     string     `json:"event_type"`
	Channel       string     `json:"channel"`
	Recipient     string     `json:"recipient,omitempty"`
	Title         string     `json:"title"`
	Message       string     `json:"message"`
	Status        string     `json:"status"`
	FailureReason string     `json:"failure_reason,omitempty"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
