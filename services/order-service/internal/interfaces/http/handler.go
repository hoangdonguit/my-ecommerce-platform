package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/shared/errs"
)

type OrderHandler struct {
	service *orderapp.Service
}

func NewOrderHandler(service *orderapp.Service) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req orderapp.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	idemKey := c.GetHeader("X-Idempotency-Key")

	order, duplicated, err := h.service.CreateOrder(c.Request.Context(), req, idemKey)
	if err != nil {
		handleError(c, err)
		return
	}

	statusCode := http.StatusCreated
	message := "Order created successfully"

	if duplicated {
		statusCode = http.StatusOK
		message = "Order already exists for this idempotency key"
	}

	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    orderapp.ToOrderResponse(order),
	})
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.Query("user_id")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid page parameter",
			Error:   "page must be a number",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid limit parameter",
			Error:   "limit must be a number",
		})
		return
	}

	orders, total, err := h.service.ListOrders(c.Request.Context(), userID, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Orders fetched successfully",
		Data:    orderapp.ToOrderResponses(orders),
		Meta: gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order fetched successfully",
		Data:    orderapp.ToOrderResponse(order),
	})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.CancelOrder(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order cancelled successfully",
		Data: gin.H{
			"id":     id,
			"status": "CANCELLED",
		},
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
