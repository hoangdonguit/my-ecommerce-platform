package payment

import "time"

const (
	OutboxStatusPending    = "PENDING"
	OutboxStatusProcessing = "PROCESSING"
	OutboxStatusPublished  = "PUBLISHED"
	OutboxStatusFailed     = "FAILED"
)

type OutboxEvent struct {
	ID            string
	AggregateID   string
	EventType     string
	Topic         string
	MessageKey    string
	Payload       []byte
	Headers       []byte
	Status        string
	Attempts      int
	LastError     string
	NextAttemptAt time.Time
	PublishedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
