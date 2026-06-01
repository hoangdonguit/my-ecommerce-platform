package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string

	OrderServiceURL        string
	InventoryServiceURL    string
	PaymentServiceURL      string
	NotificationServiceURL string
	ReadModelServiceURL    string
	RedisAddr              string
	RedisPassword          string

	OTelEnabled     bool
	OTelServiceName string
	OTelEnvironment string
	OTelEndpoint    string
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
		ReadModelServiceURL:    getEnv("READ_MODEL_SERVICE_URL", "http://localhost:8085/api/v1"),
		RedisAddr:              getEnv("REDIS_ADDR", "redis.default.svc.cluster.local:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),

		OTelEnabled:     getEnvBool("OTEL_ENABLED", false),
		OTelServiceName: getEnv("OTEL_SERVICE_NAME", "web-gateway"),
		OTelEnvironment: getEnv("OTEL_ENVIRONMENT", getEnv("APP_ENV", "development")),
		OTelEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector.observability.svc.cluster.local:4317"),
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
	if cfg.ReadModelServiceURL == "" {
		log.Fatal("READ_MODEL_SERVICE_URL is required")
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
