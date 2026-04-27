package paymentapp

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"
)

type GatewayResult struct {
	Success       bool
	TransactionID string
	FailureCode   string
	FailureReason string
	RawResponse   string
}

type PaymentGateway interface {
	Charge(orderID string, userID string, amount float64, currency string, method string) GatewayResult
}

type SimulatedPaymentGateway struct{}

func NewSimulatedPaymentGateway() *SimulatedPaymentGateway {
	return &SimulatedPaymentGateway{}
}

func (g *SimulatedPaymentGateway) Charge(orderID string, userID string, amount float64, currency string, method string) GatewayResult {
	method = strings.ToUpper(strings.TrimSpace(method))
	currency = strings.ToUpper(strings.TrimSpace(currency))

	if amount <= 0 {
		return GatewayResult{
			Success:       false,
			FailureCode:   domainpayment.FailureInvalidAmount,
			FailureReason: "payment amount must be greater than 0",
			RawResponse:   "invalid amount",
		}
	}

	if currency == "" {
		currency = "VND"
	}

	switch method {
	case "COD":
		return GatewayResult{
			Success:       true,
			TransactionID: fmt.Sprintf("cod_%s", uuid.NewString()),
			RawResponse:   "cash on delivery accepted",
		}

	case "CARD":
		return simulateCardPayment(orderID, userID, amount)

	case "MOMO", "ZALOPAY", "BANK_TRANSFER":
		return GatewayResult{
			Success:       true,
			TransactionID: fmt.Sprintf("%s_%s", strings.ToLower(method), uuid.NewString()),
			RawResponse:   "wallet or bank payment accepted",
		}

	default:
		return GatewayResult{
			Success:       false,
			FailureCode:   domainpayment.FailureUnsupportedMethod,
			FailureReason: "unsupported payment method",
			RawResponse:   "unsupported payment method",
		}
	}
}

func simulateCardPayment(orderID string, userID string, amount float64) GatewayResult {
	lowerUser := strings.ToLower(userID)

	if strings.Contains(lowerUser, "blocked") {
		return GatewayResult{
			Success:       false,
			FailureCode:   domainpayment.FailureInsufficientFunds,
			FailureReason: "card rejected for this user",
			RawResponse:   "card rejected",
		}
	}

	if amount > 50000000 {
		return GatewayResult{
			Success:       false,
			FailureCode:   domainpayment.FailureInsufficientFunds,
			FailureReason: "insufficient funds",
			RawResponse:   "insufficient funds",
		}
	}

	if strings.HasSuffix(orderID, "999") {
		return GatewayResult{
			Success:       false,
			FailureCode:   domainpayment.FailureGatewayTimeout,
			FailureReason: "payment gateway timeout",
			RawResponse:   "gateway timeout",
		}
	}

	time.Sleep(200 * time.Millisecond)

	return GatewayResult{
		Success:       true,
		TransactionID: fmt.Sprintf("card_%s", uuid.NewString()),
		RawResponse:   "card payment captured",
	}
}
