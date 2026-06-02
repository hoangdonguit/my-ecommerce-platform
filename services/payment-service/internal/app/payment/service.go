package paymentapp

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/infrastructure/persistence"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/shared/errs"
)

type Service struct {
	repo    domainpayment.Repository
	gateway PaymentGateway
}

func NewService(repo domainpayment.Repository, gateway PaymentGateway) *Service {
	return &Service{
		repo:    repo,
		gateway: gateway,
	}
}

func (s *Service) GetPaymentByOrderID(ctx context.Context, orderID string) (*domainpayment.Payment, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, errs.BadRequest("order_id is required")
	}

	payment, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if persistence.IsNotFound(err) {
			return nil, errs.NotFound("payment not found")
		}
		return nil, errs.WrapInternal(err, "failed to get payment")
	}

	return payment, nil
}

func (s *Service) GetPaymentByID(ctx context.Context, id string) (*domainpayment.Payment, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errs.BadRequest("payment id is required")
	}

	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if persistence.IsNotFound(err) {
			return nil, errs.NotFound("payment not found")
		}
		return nil, errs.WrapInternal(err, "failed to get payment")
	}

	return payment, nil
}

func (s *Service) HandleInventoryReserved(ctx context.Context, event InventoryReservedEvent) error {
	if err := validateInventoryReservedEvent(event); err != nil {
		return err
	}

	existing, err := s.repo.FindByOrderID(ctx, event.OrderID)
	if err == nil && existing != nil {
		return s.handleExistingPayment(ctx, existing)
	}
	if err != nil && !persistence.IsNotFound(err) {
		return errs.WrapInternal(err, "failed to check existing payment")
	}

	now := time.Now()
	payment := &domainpayment.Payment{
		ID:             uuid.NewString(),
		OrderID:        event.OrderID,
		UserID:         event.UserID,
		Amount:         event.TotalAmount,
		Currency:       strings.ToUpper(event.Currency),
		PaymentMethod:  strings.ToUpper(event.PaymentMethod),
		Status:         domainpayment.StatusProcessing,
		FailureCode:    "",
		FailureReason:  "",
		TransactionID:  "",
		IdempotencyKey: "payment:" + event.OrderID,
		PaidAt:         nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return errs.WrapInternal(err, "failed to create payment")
	}

	return s.finalizePayment(ctx, payment)
}

func (s *Service) handleExistingPayment(ctx context.Context, payment *domainpayment.Payment) error {
	switch payment.Status {
	case domainpayment.StatusCompleted:
		return nil
	case domainpayment.StatusFailed:
		return nil
	case domainpayment.StatusProcessing:
		return s.finalizePayment(ctx, payment)
	default:
		return nil
	}
}

func (s *Service) finalizePayment(ctx context.Context, payment *domainpayment.Payment) error {
	result := s.gateway.Charge(
		payment.OrderID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.PaymentMethod,
	)

	attemptStatus := domainpayment.StatusFailed
	if result.Success {
		attemptStatus = domainpayment.StatusCompleted
	}

	attempt := &domainpayment.PaymentAttempt{
		ID:                   uuid.NewString(),
		PaymentID:            payment.ID,
		OrderID:              payment.OrderID,
		Status:               attemptStatus,
		GatewayTransactionID: result.TransactionID,
		FailureCode:          result.FailureCode,
		FailureReason:        result.FailureReason,
		RawResponse:          result.RawResponse,
		CreatedAt:            time.Now(),
	}

	if err := s.repo.CreateAttempt(ctx, attempt); err != nil {
		return errs.WrapInternal(err, "failed to create payment attempt")
	}

	if result.Success {
		paidAt := time.Now().UTC().Format(time.RFC3339)

		if err := s.repo.UpdateCompleted(ctx, payment.ID, result.TransactionID, paidAt); err != nil {
			return errs.WrapInternal(err, "failed to update payment completed")
		}

		return nil
	}

	if err := s.repo.UpdateFailed(ctx, payment.ID, result.FailureCode, result.FailureReason); err != nil {
		return errs.WrapInternal(err, "failed to update payment failed")
	}

	return nil
}

func validateInventoryReservedEvent(event InventoryReservedEvent) error {
	if strings.TrimSpace(event.OrderID) == "" {
		return errs.BadRequest("order_id is required")
	}

	if strings.TrimSpace(event.UserID) == "" {
		return errs.BadRequest("user_id is required")
	}

	if event.TotalAmount <= 0 {
		return errs.BadRequest("total_amount must be greater than 0")
	}

	if strings.TrimSpace(event.Currency) == "" {
		return errs.BadRequest("currency is required")
	}

	if strings.TrimSpace(event.PaymentMethod) == "" {
		return errs.BadRequest("payment_method is required")
	}

	return nil
}
