package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	notificationapp "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/app/notification"
	"github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/shared/errs"
)

type NotificationHandler struct {
	service *notificationapp.Service
}

func NewNotificationHandler(service *notificationapp.Service) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "notification-service is running",
		Data: gin.H{
			"service": "notification-service",
		},
	})
}

func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	id := c.Param("id")

	notification, err := h.service.GetNotificationByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Notification fetched successfully",
		Data:    notificationapp.ToNotificationResponse(notification),
	})
}

func (h *NotificationHandler) ListNotificationsByUserID(c *gin.Context) {
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

	items, total, err := h.service.ListByUserID(c.Request.Context(), userID, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Notifications fetched successfully",
		Data:    notificationapp.ToNotificationResponses(items),
		Meta: gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *NotificationHandler) ListNotificationsByOrderID(c *gin.Context) {
	orderID := c.Param("orderId")

	items, err := h.service.ListByOrderID(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Notifications fetched successfully",
		Data:    notificationapp.ToNotificationResponses(items),
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
