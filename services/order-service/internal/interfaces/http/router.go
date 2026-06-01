package http

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRouter(orderHandler *OrderHandler) *gin.Engine {
	router := gin.Default()

	router.Use(otelgin.Middleware("order-service"))

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, APIResponse{
				Success: true,
				Message: "order-service is running",
				Data: gin.H{
					"service": "order-service",
				},
			})
		})

		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders", orderHandler.ListOrders)
		api.GET("/orders/:id", orderHandler.GetOrderByID)
		api.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)
	}

	return router
}
