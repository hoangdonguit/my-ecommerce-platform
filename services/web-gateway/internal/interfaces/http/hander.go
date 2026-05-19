package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/app"
	"github.com/hoangdonguit/my-ecommerce-platform/web-gateway/internal/client"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	orders        *client.OrderClient
	inventory     *client.InventoryClient
	payment       *client.PaymentClient
	notifications *client.NotificationClient
	readModels    *client.ReadModelClient
	cache         *redis.Client
	saga          *app.SagaService
}

func NewHandler(
	orders *client.OrderClient,
	inventory *client.InventoryClient,
	payment *client.PaymentClient,
	notifications *client.NotificationClient,
	readModels *client.ReadModelClient,
	cache *redis.Client,
	saga *app.SagaService,
) *Handler {
	return &Handler{
		orders:        orders,
		inventory:     inventory,
		payment:       payment,
		notifications: notifications,
		readModels:    readModels,
		cache:         cache,
		saga:          saga,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "web-gateway is running",
		Data: gin.H{
			"service": "web-gateway",
		},
	})
}

func (h *Handler) ServicesHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	result := gin.H{}
	result["order_service"] = h.checkService(ctx, h.orders.Health)
	result["inventory_service"] = h.checkService(ctx, h.inventory.Health)
	result["payment_service"] = h.checkService(ctx, h.payment.Health)
	result["notification_service"] = h.checkService(ctx, h.notifications.Health)
	result["read_model_service"] = h.checkService(ctx, h.readModels.Health)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "services health checked",
		Data:    result,
	})
}

func (h *Handler) checkService(ctx context.Context, fn func(context.Context) error) gin.H {
	if err := fn(ctx); err != nil {
		return gin.H{"ok": false, "error": err.Error()}
	}
	return gin.H{"ok": true}
}

// === HÀM PROXY KÉO TOÀN BỘ SẢN PHẨM ===
func (h *Handler) ListInventories(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cacheKey := "gateway:cache:inventories:all"
	if cachedData, ok := h.getCache(ctx, cacheKey); ok {
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json; charset=utf-8", cachedData)
		return
	}

	// Gọi trực tiếp đến inventory-service nội bộ K8s
	req, err := http.NewRequestWithContext(ctx, "GET", "http://inventory-api.default.svc.cluster.local:8082/api/v1/inventories", nil)
	if err != nil {
		handleError(c, err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		handleError(c, err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		handleError(c, err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		if raw, err := json.Marshal(result); err == nil {
			h.setCache(ctx, cacheKey, raw, 5*time.Second)
		}
	}

	c.Header("X-Cache", "MISS")
	c.JSON(resp.StatusCode, result)
}

// ======================================

func (h *Handler) CreateOrder(c *gin.Context) {
	var req client.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}
	if idempotencyKey == "" {
		idempotencyKey = "web-" + uuid.NewString()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := h.orders.CreateOrder(ctx, req, idempotencyKey)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Order created successfully",
		Data: gin.H{
			"order":           order,
			"idempotency_key": idempotencyKey,
			"saga_url":        fmt.Sprintf("/api/orders/%s/saga", order.ID),
		},
	})
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "user_id is required",
			Error:   "bad request",
		})
		return
	}

	page := parseIntQuery(c, "page", 1)
	limit := parseIntQuery(c, "limit", 10)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	orders, meta, err := h.orders.ListOrders(ctx, userID, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Orders fetched successfully",
		Data:    orders,
		Meta:    meta,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := h.orders.GetOrder(ctx, orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order fetched successfully",
		Data:    order,
	})
}

func (h *Handler) GetOrderSaga(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.saga.GetSaga(ctx, orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order saga fetched successfully",
		Data:    detail,
	})
}

func (h *Handler) ListReadModelOrders(c *gin.Context) {
	userID := c.Query("user_id")
	page := parseIntQuery(c, "page", 1)
	limit := parseIntQuery(c, "limit", 20)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cacheKey := fmt.Sprintf("gateway:cache:read-model:orders:user:%s:page:%d:limit:%d", userID, page, limit)
	if cachedData, ok := h.getCache(ctx, cacheKey); ok {
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json; charset=utf-8", cachedData)
		return
	}

	orders, meta, err := h.readModels.ListOrders(ctx, userID, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Read model orders fetched successfully",
		Data:    orders,
		Meta:    meta,
	}

	if raw, err := json.Marshal(response); err == nil {
		h.setCache(ctx, cacheKey, raw, 5*time.Second)
	}

	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetReadModelOrder(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cacheKey := fmt.Sprintf("gateway:cache:read-model:order:%s", orderID)
	if cachedData, ok := h.getCache(ctx, cacheKey); ok {
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json; charset=utf-8", cachedData)
		return
	}

	order, err := h.readModels.GetOrder(ctx, orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Read model order fetched successfully",
		Data:    order,
	}

	if raw, err := json.Marshal(response); err == nil {
		h.setCache(ctx, cacheKey, raw, 5*time.Second)
	}

	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

func (h *Handler) getCache(ctx context.Context, key string) ([]byte, bool) {
	if h.cache == nil {
		return nil, false
	}

	data, err := h.cache.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	return data, true
}

func (h *Handler) setCache(ctx context.Context, key string, value []byte, ttl time.Duration) {
	if h.cache == nil {
		return
	}

	if err := h.cache.Set(ctx, key, value, ttl).Err(); err != nil {
		fmt.Printf("failed to set gateway cache key=%s err=%v\n", key, err)
	}
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func handleError(c *gin.Context, err error) {
	var downstreamErr *client.DownstreamError
	if errors.As(err, &downstreamErr) {
		statusCode := downstreamErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadGateway
		}

		message := downstreamErr.Message
		if message == "" {
			message = "downstream service error"
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: message,
			Error: gin.H{
				"downstream_url": downstreamErr.URL,
				"status_code":    downstreamErr.StatusCode,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
		Error:   err.Error(),
	})
}
