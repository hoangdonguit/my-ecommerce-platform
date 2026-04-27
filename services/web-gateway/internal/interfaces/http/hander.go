package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/app"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/client"
)

type Handler struct {
	orders        *client.OrderClient
	inventory     *client.InventoryClient
	payment       *client.PaymentClient
	notifications *client.NotificationClient
	saga          *app.SagaService
}

func NewHandler(
	orders *client.OrderClient,
	inventory *client.InventoryClient,
	payment *client.PaymentClient,
	notifications *client.NotificationClient,
	saga *app.SagaService,
) *Handler {
	return &Handler{
		orders:        orders,
		inventory:     inventory,
		payment:       payment,
		notifications: notifications,
		saga:          saga,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "web-gateway is running",
		Data: gin.H{
			"service": "web-gateway",
		},
	})
}

func (h *Handler) ServicesHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	result := gin.H{}

	result["order_service"] = h.checkService(ctx, h.orders.Health)
	result["inventory_service"] = h.checkService(ctx, h.inventory.Health)
	result["payment_service"] = h.checkService(ctx, h.payment.Health)
	result["notification_service"] = h.checkService(ctx, h.notifications.Health)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "services health checked",
		Data:    result,
	})
}

func (h *Handler) checkService(ctx context.Context, fn func(context.Context) error) gin.H {
	if err := fn(ctx); err != nil {
		return gin.H{
			"ok":    false,
			"error": err.Error(),
		}
	}

	return gin.H{
		"ok": true,
	}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req client.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}
	if idempotencyKey == "" {
		idempotencyKey = "web-" + uuid.NewString()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	order, err := h.orders.CreateOrder(ctx, req, idempotencyKey)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Order created successfully",
		Data: gin.H{
			"order":           order,
			"idempotency_key": idempotencyKey,
			"saga_url":        fmt.Sprintf("/api/orders/%s/saga", order.ID),
		},
	})
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "user_id is required",
			Error:   "bad request",
		})
		return
	}

	page := parseIntQuery(c, "page", 1)
	limit := parseIntQuery(c, "limit", 10)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	orders, meta, err := h.orders.ListOrders(ctx, userID, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Orders fetched successfully",
		Data:    orders,
		Meta:    meta,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := h.orders.GetOrder(ctx, orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order fetched successfully",
		Data:    order,
	})
}

func (h *Handler) GetOrderSaga(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	detail, err := h.saga.GetSaga(ctx, orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order saga fetched successfully",
		Data:    detail,
	})
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func handleError(c *gin.Context, err error) {
	var downstreamErr *client.DownstreamError
	if errors.As(err, &downstreamErr) {
		statusCode := downstreamErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadGateway
		}

		message := downstreamErr.Message
		if message == "" {
			message = "downstream service error"
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: message,
			Error: gin.H{
				"downstream_url": downstreamErr.URL,
				"status_code":    downstreamErr.StatusCode,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
		Error:   err.Error(),
	})
}
