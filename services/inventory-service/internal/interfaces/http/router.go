package http

import "github.com/gin-gonic/gin"

func SetupRouter(handler *InventoryHandler) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.Health)
		api.GET("/inventories/:productId", handler.GetInventory)
		api.GET("/reservations/:orderId", handler.GetReservationByOrderID)
		api.GET("/inventories", handler.GetInventoriesList) // Thêm dòng này để mở cổng lấy toàn bộ kho hàng
	}

	return router
}
