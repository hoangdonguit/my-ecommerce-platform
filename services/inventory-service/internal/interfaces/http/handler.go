package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	inventoryapp "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/app/inventory"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/shared/errs"
)

type InventoryHandler struct {
	service *inventoryapp.Service
}

func NewInventoryHandler(service *inventoryapp.Service) *InventoryHandler {
	return &InventoryHandler{
		service: service,
	}
}

func (h *InventoryHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "inventory-service is running",
		Data: gin.H{
			"service": "inventory-service",
		},
	})
}

func (h *InventoryHandler) GetInventory(c *gin.Context) {
	productID := c.Param("productId")

	inv, err := h.service.GetInventory(c.Request.Context(), productID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Inventory fetched successfully",
		Data:    inventoryapp.ToInventoryResponse(inv),
	})
}

func (h *InventoryHandler) GetReservationByOrderID(c *gin.Context) {
	orderID := c.Param("orderId")

	reservation, err := h.service.GetReservationByOrderID(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Reservation fetched successfully",
		Data:    inventoryapp.ToReservationResponse(reservation),
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
