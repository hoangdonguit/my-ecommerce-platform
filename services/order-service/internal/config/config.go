package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort           string
	DBURL             string
	KafkaBroker       string
	OrderCreatedTopic string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppPort:           getEnv("APP_PORT", "8081"),
		DBURL:             getEnv("DB_URL", ""),
		KafkaBroker:       getEnv("KAFKA_BROKER", "localhost:9092"),
		OrderCreatedTopic: getEnv("KAFKA_TOPIC_ORDER_CREATED", "order.created"),
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
