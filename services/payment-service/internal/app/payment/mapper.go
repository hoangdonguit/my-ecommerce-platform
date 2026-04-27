package paymentapp

import domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"

func ToPaymentResponse(p *domainpayment.Payment) PaymentResponse {
	return PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		UserID:        p.UserID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		PaymentMethod: p.PaymentMethod,
		Status:        p.Status,
		FailureCode:   p.FailureCode,
		FailureReason: p.FailureReason,
		TransactionID: p.TransactionID,
		PaidAt:        p.PaidAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
