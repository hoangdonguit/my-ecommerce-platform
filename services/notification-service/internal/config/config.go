package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string
	DBURL   string

	KafkaBroker  string
	KafkaGroupID string

	PaymentCompletedTopic string
	PaymentFailedTopic    string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppName: getEnv("APP_NAME", "notification-service"),
		AppPort: getEnv("APP_PORT", "8084"),
		AppEnv:  getEnv("APP_ENV", "development"),
		DBURL:   getEnv("DB_URL", ""),

		KafkaBroker:  getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "notification-service-group"),

		PaymentCompletedTopic: getEnv("KAFKA_TOPIC_PAYMENT_COMPLETED", "payment.completed"),
		PaymentFailedTopic:    getEnv("KAFKA_TOPIC_PAYMENT_FAILED", "payment.failed"),
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
