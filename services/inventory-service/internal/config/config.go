package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort                string
	DBURL                  string
	KafkaBroker            string
	OrderCreatedTopic      string
	InventoryReservedTopic string
	InventoryFailedTopic   string
	KafkaGroupID           string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppPort:                getEnv("APP_PORT", "8082"),
		DBURL:                  getEnv("DB_URL", ""),
		KafkaBroker:            getEnv("KAFKA_BROKER", "localhost:9092"),
		OrderCreatedTopic:      getEnv("KAFKA_TOPIC_ORDER_CREATED", "order.created"),
		InventoryReservedTopic: getEnv("KAFKA_TOPIC_INVENTORY_RESERVED", "inventory.reserved"),
		InventoryFailedTopic:   getEnv("KAFKA_TOPIC_INVENTORY_FAILED", "inventory.failed"),
		KafkaGroupID:           getEnv("KAFKA_GROUP_ID", "inventory-service-group"),
	}

	if cfg.DBURL == "" {
		log.Fatal("DB_URL is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
