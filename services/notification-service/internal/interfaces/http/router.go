package http

import "github.com/gin-gonic/gin"

func SetupRouter(handler *NotificationHandler) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.Health)
		api.GET("/notifications", handler.ListNotificationsByUserID)
		api.GET("/notifications/:id", handler.GetNotificationByID)
		api.GET("/notifications/order/:orderId", handler.ListNotificationsByOrderID)
	}

	return router
}
