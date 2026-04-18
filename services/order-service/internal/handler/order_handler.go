package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/model"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req model.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	idemKey := c.GetHeader("X-Idempotency-Key")

	order, duplicated, err := h.service.CreateOrder(c.Request.Context(), req, idemKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	statusCode := http.StatusCreated
	if duplicated {
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, gin.H{
		"message":    "order processed",
		"duplicated": duplicated,
		"data":       order,
	})
}
