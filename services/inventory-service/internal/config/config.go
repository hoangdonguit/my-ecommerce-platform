package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName                string
	AppPort                string
	AppEnv                 string
	DBURL                  string
	KafkaBroker            string
	KafkaGroupID           string
	OrderCreatedTopic      string
	InventoryReservedTopic string
	InventoryFailedTopic   string

	OTelEnabled              bool
	OTelServiceName          string
	OTelEnvironment          string
	OTelExporterOTLPEndpoint string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppName:                getEnv("APP_NAME", "inventory-service"),
		AppPort:                getEnv("APP_PORT", "8082"),
		AppEnv:                 getEnv("APP_ENV", "development"),
		DBURL:                  getEnv("DB_URL", ""),
		KafkaBroker:            getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaGroupID:           getEnv("KAFKA_GROUP_ID", "inventory-service-group"),
		OrderCreatedTopic:      getEnv("KAFKA_TOPIC_ORDER_CREATED", "order.created"),
		InventoryReservedTopic: getEnv("KAFKA_TOPIC_INVENTORY_RESERVED", "inventory.reserved"),
		InventoryFailedTopic:   getEnv("KAFKA_TOPIC_INVENTORY_FAILED", "inventory.failed"),

		OTelEnabled:              getEnvBool("OTEL_ENABLED", false),
		OTelServiceName:          getEnv("OTEL_SERVICE_NAME", "inventory-service"),
		OTelEnvironment:          getEnv("OTEL_ENVIRONMENT", getEnv("APP_ENV", "development")),
		OTelExporterOTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
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

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
