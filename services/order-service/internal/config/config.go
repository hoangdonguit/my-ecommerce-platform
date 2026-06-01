package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName           string
	AppPort           string
	AppEnv            string
	DBURL             string
	KafkaBroker       string
	OrderCreatedTopic string
	RedisAddr         string
	RedisPassword     string

	OTelEnabled     bool
	OTelServiceName string
	OTelEnvironment string
	OTelEndpoint    string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppName:           getEnv("APP_NAME", "order-service"),
		AppPort:           getEnv("APP_PORT", "8081"),
		AppEnv:            getEnv("APP_ENV", "development"),
		DBURL:             getEnv("DB_URL", ""),
		KafkaBroker:       getEnv("KAFKA_BROKER", "localhost:9092"),
		OrderCreatedTopic: getEnv("KAFKA_TOPIC_ORDER_CREATED", "order.created"),
		RedisAddr:         getEnv("REDIS_ADDR", "redis-master.cache.svc.cluster.local:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", "redissecret"), // Mật khẩu mặc định của cụm K8s

		OTelEnabled:     getEnvBool("OTEL_ENABLED", false),
		OTelServiceName: getEnv("OTEL_SERVICE_NAME", "order-service"),
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
	if cfg.DBURL == "" {
		log.Fatal("DB_URL is required")
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
