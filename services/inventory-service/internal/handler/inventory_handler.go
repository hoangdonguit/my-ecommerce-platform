package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/service"
)

type InventoryHandler struct {
	service *service.InventoryService
}

func NewInventoryHandler(service *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) GetInventory(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing productId",
		})
		return
	}

	inv, err := h.service.GetInventory(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "inventory not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": inv,
	})
}
