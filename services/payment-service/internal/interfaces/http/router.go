package http

import "github.com/gin-gonic/gin"

func SetupRouter(handler *PaymentHandler) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.Health)
		api.GET("/payments/order/:orderId", handler.GetPaymentByOrderID)
		api.GET("/payments/:id", handler.GetPaymentByID)
	}

	return router
}
