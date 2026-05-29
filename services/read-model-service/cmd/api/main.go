package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	kafkago "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	AppPort         string
	KafkaBroker     string
	KafkaTopic      string
	KafkaGroupID    string
	MongoURI        string
	MongoHost       string
	MongoUsername   string
	MongoPassword   string
	MongoAuthSource string
	MongoDatabase   string
	MongoCollection string
}

type PaymentCompletedEvent struct {
	EventType     string  `json:"event_type"`
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	PaymentID     string  `json:"payment_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id"`
	PaidAt        string  `json:"paid_at"`
}

type OrderReadModel struct {
	OrderID       string     `bson:"order_id" json:"order_id"`
	UserID        string     `bson:"user_id" json:"user_id"`
	SagaStatus    string     `bson:"saga_status" json:"saga_status"`
	PaymentStatus string     `bson:"payment_status" json:"payment_status"`
	PaymentID     string     `bson:"payment_id" json:"payment_id"`
	Amount        float64    `bson:"amount" json:"amount"`
	Currency      string     `bson:"currency" json:"currency"`
	PaymentMethod string     `bson:"payment_method" json:"payment_method"`
	TransactionID string     `bson:"transaction_id" json:"transaction_id"`
	SourceEvent   string     `bson:"source_event" json:"source_event"`
	PaidAt        *time.Time `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
	CreatedAt     time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `bson:"updated_at" json:"updated_at"`
}

type App struct {
	cfg        Config
	mongo      *mongo.Client
	collection *mongo.Collection
}

func main() {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, err := connectMongo(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect mongodb: %v", err)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
	}()

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	app := &App{
		cfg:        cfg,
		mongo:      mongoClient,
		collection: collection,
	}

	if err := app.ensureIndexes(ctx); err != nil {
		log.Fatalf("failed to ensure indexes: %v", err)
	}

	go app.runConsumer(ctx)

	router := app.setupRouter()
	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		log.Printf("read-model-service listening on :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
}

func loadConfig() Config {
	cfg := Config{
		AppPort:         getEnv("APP_PORT", "8085"),
		KafkaBroker:     getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:      getEnv("KAFKA_TOPIC_PAYMENT_COMPLETED", "payment.completed"),
		KafkaGroupID:    getEnv("KAFKA_GROUP_ID", "read-model-service-group"),
		MongoURI:        strings.TrimSpace(os.Getenv("MONGODB_URI")),
		MongoHost:       getEnv("MONGODB_HOST", "mongodb.default.svc.cluster.local:27017"),
		MongoUsername:   getEnv("MONGODB_USERNAME", "root"),
		MongoPassword:   strings.TrimSpace(os.Getenv("MONGODB_PASSWORD")),
		MongoAuthSource: getEnv("MONGODB_AUTH_SOURCE", "admin"),
		MongoDatabase:   getEnv("MONGODB_DATABASE", "ecommerce_read"),
		MongoCollection: getEnv("MONGODB_COLLECTION", "order_read_models"),
	}

	if cfg.MongoURI == "" && cfg.MongoPassword == "" {
		log.Fatal("MONGODB_PASSWORD is required when MONGODB_URI is not set")
	}

	return cfg
}

func connectMongo(ctx context.Context, cfg Config) (*mongo.Client, error) {
	mongoURI := cfg.MongoURI
	if mongoURI == "" {
		mongoURI = fmt.Sprintf(
			"mongodb://%s:%s@%s/%s?authSource=%s",
			url.QueryEscape(cfg.MongoUsername),
			url.QueryEscape(cfg.MongoPassword),
			cfg.MongoHost,
			cfg.MongoDatabase,
			url.QueryEscape(cfg.MongoAuthSource),
		)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, err
	}

	log.Println("mongodb connected successfully")
	return client, nil
}

func (a *App) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "order_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("ux_order_read_models_order_id"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_order_read_models_user_updated"),
		},
		{
			Keys:    bson.D{{Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_order_read_models_updated_at"),
		},
	}

	indexCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := a.collection.Indexes().CreateMany(indexCtx, indexes)
	return err
}

func (a *App) runConsumer(ctx context.Context) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  []string{a.cfg.KafkaBroker},
		Topic:    a.cfg.KafkaTopic,
		GroupID:  a.cfg.KafkaGroupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Printf("read-model consumer listening topic=%s group=%s", a.cfg.KafkaTopic, a.cfg.KafkaGroupID)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("failed to fetch kafka message: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var event PaymentCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal payment.completed: %v", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		if err := a.upsertPaymentCompleted(ctx, event); err != nil {
			log.Printf("failed to upsert read model order_id=%s err=%v", event.OrderID, err)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit kafka message order_id=%s err=%v", event.OrderID, err)
			continue
		}

		log.Printf("upserted order read model order_id=%s payment_id=%s", event.OrderID, event.PaymentID)
	}
}

func (a *App) upsertPaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error {
	event.OrderID = strings.TrimSpace(event.OrderID)
	if event.OrderID == "" {
		return errors.New("order_id is required")
	}

	now := time.Now().UTC()
	paidAt := parseTimePtr(event.PaidAt)

	update := bson.M{
		"$set": bson.M{
			"order_id":       event.OrderID,
			"user_id":        event.UserID,
			"saga_status":    "COMPLETED",
			"payment_status": strings.ToUpper(defaultString(event.Status, "COMPLETED")),
			"payment_id":     event.PaymentID,
			"amount":         event.Amount,
			"currency":       strings.ToUpper(defaultString(event.Currency, "VND")),
			"payment_method": strings.ToUpper(event.PaymentMethod),
			"transaction_id": event.TransactionID,
			"source_event":   defaultString(event.EventType, "payment.completed"),
			"paid_at":        paidAt,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := a.collection.UpdateOne(
		opCtx,
		bson.M{"order_id": event.OrderID},
		update,
		options.Update().SetUpsert(true),
	)

	return err
}

func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.GET("/health", a.health)
		api.GET("/read-model/orders", a.listOrders)
		api.GET("/read-model/orders/:orderId", a.getOrder)
	}

	return router
}

func (a *App) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := a.mongo.Ping(ctx, readpref.Primary()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "read-model-service is not healthy",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "read-model-service is running",
		"data": gin.H{
			"service":    "read-model-service",
			"mongo_db":   a.cfg.MongoDatabase,
			"collection": a.cfg.MongoCollection,
		},
	})
}

func (a *App) listOrders(c *gin.Context) {
	page := clampInt(queryInt(c, "page", 1), 1, 100000)
	limit := clampInt(queryInt(c, "limit", 50), 1, 1000)
	userID := strings.TrimSpace(c.Query("user_id"))

	filter := bson.M{}
	if userID != "" {
		filter["user_id"] = userID
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cursor, err := a.collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to query read model",
			"error":   err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	var orders []OrderReadModel
	if err := cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to decode read model",
			"error":   err.Error(),
		})
		return
	}

	if orders == nil {
		orders = []OrderReadModel{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "read model orders fetched successfully",
		"data":    orders,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
		},
	})
}

func (a *App) getOrder(c *gin.Context) {
	orderID := strings.TrimSpace(c.Param("orderId"))
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "orderId is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var order OrderReadModel
	err := a.collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "read model order not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to query read model",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "read model order fetched successfully",
		"data":    order,
	})
}

func parseTimePtr(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}

	return &t
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func queryInt(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
