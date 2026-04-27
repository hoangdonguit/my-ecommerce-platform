package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/client"
)

type State struct {
	Exists bool   `json:"exists"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	Data   any    `json:"data,omitempty"`
}

type SagaStep struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

type SagaDetail struct {
	Order         *client.OrderResponse `json:"order"`
	Inventory     State                 `json:"inventory"`
	Payment       State                 `json:"payment"`
	Notifications State                 `json:"notifications"`
	Timeline      []SagaStep            `json:"timeline"`
	SagaStatus    string                `json:"saga_status"`
	Warnings      []string              `json:"warnings,omitempty"`
}

type SagaService struct {
	Orders        *client.OrderClient
	Inventory     *client.InventoryClient
	Payment       *client.PaymentClient
	Notifications *client.NotificationClient
}

func NewSagaService(
	orders *client.OrderClient,
	inventory *client.InventoryClient,
	payment *client.PaymentClient,
	notifications *client.NotificationClient,
) *SagaService {
	return &SagaService{
		Orders:        orders,
		Inventory:     inventory,
		Payment:       payment,
		Notifications: notifications,
	}
}

func (s *SagaService) GetSaga(ctx context.Context, orderID string) (*SagaDetail, error) {
	order, err := s.Orders.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	warnings := make([]string, 0)

	inventoryState := State{
		Exists: false,
		Status: "PENDING",
		Reason: "waiting for inventory reservation",
	}

	reservation, err := s.Inventory.GetReservationByOrderID(ctx, orderID)
	if err == nil && reservation != nil {
		inventoryState = State{
			Exists: true,
			Status: strings.ToUpper(reservation.Status),
			Reason: reservation.Reason,
			Data:   reservation,
		}
	} else if err != nil && client.IsNotFound(err) {
		inventoryState = State{
			Exists: false,
			Status: "PENDING",
			Reason: "inventory service has not processed this order yet",
		}
	} else if err != nil {
		warnings = append(warnings, fmt.Sprintf("inventory service unavailable: %v", err))
		inventoryState = State{
			Exists: false,
			Status: "UNKNOWN",
			Reason: "inventory service unavailable",
		}
	}

	paymentState := State{
		Exists: false,
		Status: "WAITING",
		Reason: "waiting for inventory to reserve stock",
	}

	payment, err := s.Payment.GetPaymentByOrderID(ctx, orderID)
	if err == nil && payment != nil {
		paymentState = State{
			Exists: true,
			Status: strings.ToUpper(payment.Status),
			Reason: payment.FailureReason,
			Data:   payment,
		}
	} else if err != nil && client.IsNotFound(err) {
		reason := "payment service has not processed this order yet"
		status := "PENDING"

		if !isInventorySuccess(inventoryState.Status) {
			status = "WAITING"
			reason = "waiting for inventory reservation"
		}

		paymentState = State{
			Exists: false,
			Status: status,
			Reason: reason,
		}
	} else if err != nil {
		warnings = append(warnings, fmt.Sprintf("payment service unavailable: %v", err))
		paymentState = State{
			Exists: false,
			Status: "UNKNOWN",
			Reason: "payment service unavailable",
		}
	}

	notificationState := State{
		Exists: false,
		Status: "WAITING",
		Reason: "waiting for terminal payment event",
	}

	notifications, err := s.Notifications.ListByOrderID(ctx, orderID)
	if err == nil {
		if len(notifications) > 0 {
			status := "PENDING"
			reason := ""

			for _, item := range notifications {
				if strings.EqualFold(item.Status, "SENT") {
					status = "SENT"
					break
				}
				if strings.EqualFold(item.Status, "FAILED") {
					status = "FAILED"
					reason = item.FailureReason
				}
			}

			notificationState = State{
				Exists: true,
				Status: status,
				Reason: reason,
				Data:   notifications,
			}
		} else {
			notificationState = State{
				Exists: false,
				Status: "PENDING",
				Reason: "no notification found for this order yet",
				Data:   []client.NotificationResponse{},
			}
		}
	} else if err != nil && client.IsNotFound(err) {
		notificationState = State{
			Exists: false,
			Status: "PENDING",
			Reason: "no notification found for this order yet",
			Data:   []client.NotificationResponse{},
		}
	} else if err != nil {
		warnings = append(warnings, fmt.Sprintf("notification service unavailable: %v", err))
		notificationState = State{
			Exists: false,
			Status: "UNKNOWN",
			Reason: "notification service unavailable",
		}
	}

	timeline := buildTimeline(order, inventoryState, paymentState, notificationState)
	sagaStatus := deriveSagaStatus(inventoryState, paymentState, notificationState)

	return &SagaDetail{
		Order:         order,
		Inventory:     inventoryState,
		Payment:       paymentState,
		Notifications: notificationState,
		Timeline:      timeline,
		SagaStatus:    sagaStatus,
		Warnings:      warnings,
	}, nil
}

func buildTimeline(order *client.OrderResponse, inventory State, payment State, notifications State) []SagaStep {
	steps := make([]SagaStep, 0, 4)

	steps = append(steps, SagaStep{
		Key:         "order_created",
		Label:       "Order Created",
		Status:      "success",
		Description: fmt.Sprintf("Order %s was created with status %s", order.ID, order.Status),
	})

	if isInventorySuccess(inventory.Status) {
		steps = append(steps, SagaStep{
			Key:         "inventory_reserved",
			Label:       "Inventory Reserved",
			Status:      "success",
			Description: "Stock has been reserved successfully",
		})
	} else if isInventoryFailed(inventory.Status) {
		steps = append(steps, SagaStep{
			Key:         "inventory_failed",
			Label:       "Inventory Failed",
			Status:      "failed",
			Description: inventory.Reason,
		})
	} else {
		steps = append(steps, SagaStep{
			Key:         "inventory_pending",
			Label:       "Inventory Processing",
			Status:      "pending",
			Description: inventory.Reason,
		})
	}

	if isPaymentSuccess(payment.Status) {
		steps = append(steps, SagaStep{
			Key:         "payment_completed",
			Label:       "Payment Completed",
			Status:      "success",
			Description: "Payment was captured successfully",
		})
	} else if isPaymentFailed(payment.Status) {
		steps = append(steps, SagaStep{
			Key:         "payment_failed",
			Label:       "Payment Failed",
			Status:      "failed",
			Description: payment.Reason,
		})
	} else {
		status := "pending"
		label := "Payment Processing"
		if strings.EqualFold(payment.Status, "WAITING") {
			status = "waiting"
			label = "Payment Waiting"
		}

		steps = append(steps, SagaStep{
			Key:         "payment_pending",
			Label:       label,
			Status:      status,
			Description: payment.Reason,
		})
	}

	if strings.EqualFold(notifications.Status, "SENT") {
		steps = append(steps, SagaStep{
			Key:         "notification_sent",
			Label:       "Notification Sent",
			Status:      "success",
			Description: "User notification has been sent",
		})
	} else if strings.EqualFold(notifications.Status, "FAILED") {
		steps = append(steps, SagaStep{
			Key:         "notification_failed",
			Label:       "Notification Failed",
			Status:      "failed",
			Description: notifications.Reason,
		})
	} else {
		status := "pending"
		if !isPaymentSuccess(payment.Status) && !isPaymentFailed(payment.Status) {
			status = "waiting"
		}

		steps = append(steps, SagaStep{
			Key:         "notification_pending",
			Label:       "Notification Waiting",
			Status:      status,
			Description: notifications.Reason,
		})
	}

	return steps
}

func deriveSagaStatus(inventory State, payment State, notifications State) string {
	if isInventoryFailed(inventory.Status) {
		return "INVENTORY_FAILED"
	}

	if isPaymentFailed(payment.Status) {
		if strings.EqualFold(notifications.Status, "SENT") {
			return "FAILED_NOTIFIED"
		}
		return "PAYMENT_FAILED"
	}

	if isPaymentSuccess(payment.Status) {
		if strings.EqualFold(notifications.Status, "SENT") {
			return "COMPLETED"
		}
		return "PAYMENT_COMPLETED"
	}

	return "PROCESSING"
}

func isInventorySuccess(status string) bool {
	status = strings.ToUpper(status)
	return status == "RESERVED" || status == "CONFIRMED"
}

func isInventoryFailed(status string) bool {
	return strings.EqualFold(status, "FAILED")
}

func isPaymentSuccess(status string) bool {
	return strings.EqualFold(status, "COMPLETED")
}

func isPaymentFailed(status string) bool {
	return strings.EqualFold(status, "FAILED")
}
