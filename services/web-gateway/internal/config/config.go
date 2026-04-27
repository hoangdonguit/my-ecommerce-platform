package config

import (
	"log"
	"os"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string

	OrderServiceURL        string
	InventoryServiceURL    string
	PaymentServiceURL      string
	NotificationServiceURL string
}

func Load() Config {
	cfg := Config{
		AppName: getEnv("APP_NAME", "web-gateway"),
		AppPort: getEnv("APP_PORT", "8090"),
		AppEnv:  getEnv("APP_ENV", "development"),

		OrderServiceURL:        getEnv("ORDER_SERVICE_URL", "http://localhost:8081/api/v1"),
		InventoryServiceURL:    getEnv("INVENTORY_SERVICE_URL", "http://localhost:8082/api/v1"),
		PaymentServiceURL:      getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083/api/v1"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8084/api/v1"),
	}

	validate(cfg)
	return cfg
}

func validate(cfg Config) {
	if cfg.AppPort == "" {
		log.Fatal("APP_PORT is required")
	}
	if cfg.OrderServiceURL == "" {
		log.Fatal("ORDER_SERVICE_URL is required")
	}
	if cfg.InventoryServiceURL == "" {
		log.Fatal("INVENTORY_SERVICE_URL is required")
	}
	if cfg.PaymentServiceURL == "" {
		log.Fatal("PAYMENT_SERVICE_URL is required")
	}
	if cfg.NotificationServiceURL == "" {
		log.Fatal("NOTIFICATION_SERVICE_URL is required")
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
