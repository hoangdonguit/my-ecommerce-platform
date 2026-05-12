package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	orderapp "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/app/order"
	"github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/shared/errs"
	"github.com/redis/go-redis/v9"
)

// Thêm Redis client vào struct
type OrderHandler struct {
	service *orderapp.Service
	redis   *redis.Client
}

// Cập nhật hàm khởi tạo để nhận Redis client
func NewOrderHandler(service *orderapp.Service, redisClient *redis.Client) *OrderHandler {
	return &OrderHandler{
		service: service,
		redis:   redisClient,
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

	// --- LOGIC XÓA CACHE (INVALIDATION) ---
	// Xóa toàn bộ cache liên quan đến danh sách order trên Dashboard để tránh dữ liệu thiu
	keys, _ := h.redis.Keys(c.Request.Context(), "dashboard:orders:*").Result()
	if len(keys) > 0 {
		h.redis.Del(c.Request.Context(), keys...)
	}
	// --------------------------------------

	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    orderapp.ToOrderResponse(order),
	})
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.Query("user_id")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Message: "Invalid page"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Message: "Invalid limit"})
		return
	}

	// --- LOGIC ĐỌC CACHE (CACHE-ASIDE) ---
	// 1. Tạo Key duy nhất (Dựa trên user_id, page, limit)
	cacheKey := fmt.Sprintf("dashboard:orders:user:%s:page:%s:limit:%s", userID, pageStr, limitStr)

	// 2. Kiểm tra Redis trước
	cachedData, err := h.redis.Get(c.Request.Context(), cacheKey).Bytes()
	if err == nil {
		// HIT CACHE: Parse dữ liệu và trả về ngay lập tức
		var cachedResponse APIResponse
		if json.Unmarshal(cachedData, &cachedResponse) == nil {
			c.JSON(http.StatusOK, cachedResponse)
			return
		}
	}
	// --------------------------------------

	// 3. MISS CACHE: Đọc từ Database
	orders, total, err := h.service.ListOrders(c.Request.Context(), userID, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Orders fetched successfully",
		Data:    orderapp.ToOrderResponses(orders),
		Meta: gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	}

	// --- LOGIC GHI CACHE ---
	// 4. Lưu kết quả vào Redis với TTL = 5 giây
	if responseData, jsonErr := json.Marshal(response); jsonErr == nil {
		h.redis.SetNX(c.Request.Context(), cacheKey, responseData, 5*time.Second)
	}
	// -----------------------

	c.JSON(http.StatusOK, response)
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

	// Cập nhật trạng thái thì cũng nên xóa cache Dashboard
	keys, _ := h.redis.Keys(c.Request.Context(), "dashboard:orders:*").Result()
	if len(keys) > 0 {
		h.redis.Del(c.Request.Context(), keys...)
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
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Message: err.Error(), Error: "bad request"})
	case errs.IsCode(err, "NOT_FOUND"):
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Message: err.Error(), Error: "not found"})
	case errs.IsCode(err, "CONFLICT"):
		c.JSON(http.StatusConflict, APIResponse{Success: false, Message: err.Error(), Error: "conflict"})
	default:
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Message: "Internal server error", Error: err.Error()})
	}
}