package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName           string
	AppPort           string
	AppEnv            string
	DBURL             string
	KafkaBroker       string
	OrderCreatedTopic string
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
