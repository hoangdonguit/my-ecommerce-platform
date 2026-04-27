package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DownstreamError struct {
	StatusCode int
	URL        string
	Message    string
	Body       string
}

func (e *DownstreamError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("downstream error status=%d url=%s message=%s", e.StatusCode, e.URL, e.Message)
	}
	return fmt.Sprintf("downstream error status=%d url=%s body=%s", e.StatusCode, e.URL, e.Body)
}

func IsNotFound(err error) bool {
	var downstreamErr *DownstreamError
	if errors.As(err, &downstreamErr) {
		return downstreamErr.StatusCode == http.StatusNotFound
	}
	return false
}

type Envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	Meta    json.RawMessage `json:"meta,omitempty"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method string, path string, headers map[string]string, body any, out any, metaOut any) error {
	fullURL := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return &DownstreamError{
			StatusCode: 0,
			URL:        fullURL,
			Message:    err.Error(),
		}
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var envelope Envelope
	_ = json.Unmarshal(rawBody, &envelope)

	if resp.StatusCode >= 400 {
		message := envelope.Message
		if message == "" {
			message = string(envelope.Error)
		}

		return &DownstreamError{
			StatusCode: resp.StatusCode,
			URL:        fullURL,
			Message:    message,
			Body:       string(rawBody),
		}
	}

	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &envelope); err != nil {
			return fmt.Errorf("failed to parse downstream envelope from %s: %w", fullURL, err)
		}
	}

	if !envelope.Success {
		return &DownstreamError{
			StatusCode: resp.StatusCode,
			URL:        fullURL,
			Message:    envelope.Message,
			Body:       string(rawBody),
		}
	}

	if out != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("failed to parse data from %s: %w", fullURL, err)
		}
	}

	if metaOut != nil && len(envelope.Meta) > 0 && string(envelope.Meta) != "null" {
		if err := json.Unmarshal(envelope.Meta, metaOut); err != nil {
			return fmt.Errorf("failed to parse meta from %s: %w", fullURL, err)
		}
	}

	return nil
}

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderRequest struct {
	UserID          string                   `json:"user_id"`
	Items           []CreateOrderItemRequest `json:"items"`
	Currency        string                   `json:"currency"`
	PaymentMethod   string                   `json:"payment_method"`
	ShippingAddress string                   `json:"shipping_address"`
	Note            string                   `json:"note,omitempty"`
	IdempotencyKey  string                   `json:"idempotency_key,omitempty"`
}

type OrderItemResponse struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderResponse struct {
	ID              string              `json:"id"`
	UserID          string              `json:"user_id"`
	Status          string              `json:"status"`
	Currency        string              `json:"currency"`
	PaymentMethod   string              `json:"payment_method"`
	ShippingAddress string              `json:"shipping_address"`
	Note            string              `json:"note,omitempty"`
	Items           []OrderItemResponse `json:"items"`
	TotalAmount     float64             `json:"total_amount"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

type ListMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type ReservationItemResponse struct {
	ProductID         string `json:"product_id"`
	RequestedQuantity int    `json:"requested_quantity"`
	ReservedQuantity  int    `json:"reserved_quantity"`
	Status            string `json:"status"`
}

type ReservationResponse struct {
	ID        string                    `json:"id"`
	OrderID   string                    `json:"order_id"`
	UserID    string                    `json:"user_id"`
	Status    string                    `json:"status"`
	Reason    string                    `json:"reason,omitempty"`
	Items     []ReservationItemResponse `json:"items"`
	CreatedAt string                    `json:"created_at"`
	UpdatedAt string                    `json:"updated_at"`
}

type PaymentResponse struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	FailureCode   string  `json:"failure_code,omitempty"`
	FailureReason string  `json:"failure_reason,omitempty"`
	TransactionID string  `json:"transaction_id,omitempty"`
	PaidAt        string  `json:"paid_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type NotificationResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	OrderID       string `json:"order_id"`
	EventType     string `json:"event_type"`
	Channel       string `json:"channel"`
	Recipient     string `json:"recipient,omitempty"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason,omitempty"`
	SentAt        string `json:"sent_at,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type OrderClient struct {
	client *Client
}

func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{client: NewClient(baseURL)}
}

func (c *OrderClient) Health(ctx context.Context) error {
	return c.client.do(ctx, http.MethodGet, "/health", nil, nil, nil, nil)
}

func (c *OrderClient) CreateOrder(ctx context.Context, req CreateOrderRequest, idempotencyKey string) (*OrderResponse, error) {
	downstreamReq := req
	downstreamReq.IdempotencyKey = ""

	var order OrderResponse
	err := c.client.do(
		ctx,
		http.MethodPost,
		"/orders",
		map[string]string{
			"X-Idempotency-Key": idempotencyKey,
		},
		downstreamReq,
		&order,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*OrderResponse, error) {
	var order OrderResponse
	err := c.client.do(ctx, http.MethodGet, "/orders/"+url.PathEscape(orderID), nil, nil, &order, nil)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (c *OrderClient) ListOrders(ctx context.Context, userID string, page int, limit int) ([]OrderResponse, *ListMeta, error) {
	path := fmt.Sprintf(
		"/orders?user_id=%s&page=%d&limit=%d",
		url.QueryEscape(userID),
		page,
		limit,
	)

	var orders []OrderResponse
	var meta ListMeta

	err := c.client.do(ctx, http.MethodGet, path, nil, nil, &orders, &meta)
	if err != nil {
		return nil, nil, err
	}

	return orders, &meta, nil
}

type InventoryClient struct {
	client *Client
}

func NewInventoryClient(baseURL string) *InventoryClient {
	return &InventoryClient{client: NewClient(baseURL)}
}

func (c *InventoryClient) Health(ctx context.Context) error {
	return c.client.do(ctx, http.MethodGet, "/health", nil, nil, nil, nil)
}

func (c *InventoryClient) GetReservationByOrderID(ctx context.Context, orderID string) (*ReservationResponse, error) {
	var reservation ReservationResponse
	err := c.client.do(ctx, http.MethodGet, "/reservations/"+url.PathEscape(orderID), nil, nil, &reservation, nil)
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

type PaymentClient struct {
	client *Client
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{client: NewClient(baseURL)}
}

func (c *PaymentClient) Health(ctx context.Context) error {
	return c.client.do(ctx, http.MethodGet, "/health", nil, nil, nil, nil)
}

func (c *PaymentClient) GetPaymentByOrderID(ctx context.Context, orderID string) (*PaymentResponse, error) {
	var payment PaymentResponse
	err := c.client.do(ctx, http.MethodGet, "/payments/order/"+url.PathEscape(orderID), nil, nil, &payment, nil)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

type NotificationClient struct {
	client *Client
}

func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{client: NewClient(baseURL)}
}

func (c *NotificationClient) Health(ctx context.Context) error {
	return c.client.do(ctx, http.MethodGet, "/health", nil, nil, nil, nil)
}

func (c *NotificationClient) ListByOrderID(ctx context.Context, orderID string) ([]NotificationResponse, error) {
	var notifications []NotificationResponse
	err := c.client.do(ctx, http.MethodGet, "/notifications/order/"+url.PathEscape(orderID), nil, nil, &notifications, nil)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}
