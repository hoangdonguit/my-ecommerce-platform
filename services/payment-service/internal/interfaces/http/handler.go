package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	paymentapp "github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/app/payment"
	"github.com/hoangdonguit/my-ecommerce-platform/payment-service/internal/shared/errs"
)

type PaymentHandler struct {
	service *paymentapp.Service
}

func NewPaymentHandler(service *paymentapp.Service) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "payment-service is running",
		Data: gin.H{
			"service": "payment-service",
		},
	})
}

func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderID := c.Param("orderId")

	payment, err := h.service.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment fetched successfully",
		Data:    paymentapp.ToPaymentResponse(payment),
	})
}

func (h *PaymentHandler) GetPaymentByID(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.service.GetPaymentByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment fetched successfully",
		Data:    paymentapp.ToPaymentResponse(payment),
	})
}

func handleError(c *gin.Context, err error) {
	switch {
	case errs.IsCode(err, "BAD_REQUEST"):
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
			Error:   "bad request",
		})
	case errs.IsCode(err, "NOT_FOUND"):
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
			Error:   "not found",
		})
	case errs.IsCode(err, "CONFLICT"):
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Message: err.Error(),
			Error:   "conflict",
		})
	default:
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
			Error:   err.Error(),
		})
	}
}
