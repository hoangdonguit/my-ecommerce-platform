package persistence

import (
	"context"
	"errors"
	"time"

	domainpayment "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/domain/payment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domainpayment.Payment) error {
	query := `
        INSERT INTO payments (
            id, order_id, user_id, amount, currency, payment_method,
            status, failure_code, failure_reason, transaction_id,
            idempotency_key, paid_at, created_at, updated_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
    `

	_, err := r.db.Exec(
		ctx,
		query,
		p.ID,
		p.OrderID,
		p.UserID,
		p.Amount,
		p.Currency,
		p.PaymentMethod,
		p.Status,
		p.FailureCode,
		p.FailureReason,
		p.TransactionID,
		p.IdempotencyKey,
		p.PaidAt,
		p.CreatedAt,
		p.UpdatedAt,
	)

	return err
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*domainpayment.Payment, error) {
	return r.findOne(ctx, `
        SELECT id, order_id, user_id, amount, currency, payment_method,
               status, failure_code, failure_reason, transaction_id,
               idempotency_key, paid_at, created_at, updated_at
        FROM payments
        WHERE id = $1
    `, id)
}

func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*domainpayment.Payment, error) {
	return r.findOne(ctx, `
        SELECT id, order_id, user_id, amount, currency, payment_method,
               status, failure_code, failure_reason, transaction_id,
               idempotency_key, paid_at, created_at, updated_at
        FROM payments
        WHERE order_id = $1
    `, orderID)
}

func (r *PaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domainpayment.Payment, error) {
	return r.findOne(ctx, `
        SELECT id, order_id, user_id, amount, currency, payment_method,
               status, failure_code, failure_reason, transaction_id,
               idempotency_key, paid_at, created_at, updated_at
        FROM payments
        WHERE idempotency_key = $1
    `, key)
}

func (r *PaymentRepository) findOne(ctx context.Context, query string, arg string) (*domainpayment.Payment, error) {
	var p domainpayment.Payment

	err := r.db.QueryRow(ctx, query, arg).Scan(
		&p.ID,
		&p.OrderID,
		&p.UserID,
		&p.Amount,
		&p.Currency,
		&p.PaymentMethod,
		&p.Status,
		&p.FailureCode,
		&p.FailureReason,
		&p.TransactionID,
		&p.IdempotencyKey,
		&p.PaidAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PaymentRepository) UpdateCompleted(ctx context.Context, id string, transactionID string, paidAt string) error {
	parsedPaidAt, err := time.Parse(time.RFC3339, paidAt)
	if err != nil {
		return err
	}

	query := `
        UPDATE payments
        SET status = $2,
            transaction_id = $3,
            paid_at = $4,
            failure_code = '',
            failure_reason = '',
            updated_at = NOW()
        WHERE id = $1
    `

	cmd, err := r.db.Exec(ctx, query, id, domainpayment.StatusCompleted, transactionID, parsedPaidAt)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PaymentRepository) UpdateFailed(ctx context.Context, id string, failureCode string, failureReason string) error {
	query := `
        UPDATE payments
        SET status = $2,
            failure_code = $3,
            failure_reason = $4,
            updated_at = NOW()
        WHERE id = $1
    `

	cmd, err := r.db.Exec(ctx, query, id, domainpayment.StatusFailed, failureCode, failureReason)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PaymentRepository) CreateAttempt(ctx context.Context, attempt *domainpayment.PaymentAttempt) error {
	query := `
        INSERT INTO payment_attempts (
            id, payment_id, order_id, status, gateway_transaction_id,
            failure_code, failure_reason, raw_response, created_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `

	_, err := r.db.Exec(
		ctx,
		query,
		attempt.ID,
		attempt.PaymentID,
		attempt.OrderID,
		attempt.Status,
		attempt.GatewayTransactionID,
		attempt.FailureCode,
		attempt.FailureReason,
		attempt.RawResponse,
		attempt.CreatedAt,
	)

	return err
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
