package payment

const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
	StatusCancelled  = "CANCELLED"
	StatusRefunded   = "REFUNDED"
)

const (
	FailureInsufficientFunds = "INSUFFICIENT_FUNDS"
	FailureInvalidAmount     = "INVALID_AMOUNT"
	FailureGatewayTimeout    = "GATEWAY_TIMEOUT"
	FailureUnsupportedMethod = "UNSUPPORTED_PAYMENT_METHOD"
	FailureInternalError     = "INTERNAL_ERROR"
)
