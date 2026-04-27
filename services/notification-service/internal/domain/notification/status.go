package notification

const (
	StatusPending = "PENDING"
	StatusSent    = "SENT"
	StatusFailed  = "FAILED"
)

const (
	ChannelEmail = "EMAIL"
	ChannelSMS   = "SMS"
	ChannelInApp = "IN_APP"
)

const (
	EventPaymentCompleted = "payment.completed"
	EventPaymentFailed    = "payment.failed"
)
