 package main


import (

    "context"

    "encoding/json"

    "log"

    "sync"


    inventoryapp "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/app/inventory"

    "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/config"

    "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/db"

    "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/messaging"

    "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/persistence"

    kafkago "github.com/segmentio/kafka-go"

)


const workerCount = 1 // 20 goroutines xử lý song song


func main() {

    cfg := config.Load()

    pool, err := db.NewPostgres(cfg.DBURL)

    if err != nil {

        log.Fatalf("failed to connect postgres: %v", err)

    }

    defer pool.Close()

    log.Println("postgres connected successfully")


    repo := persistence.NewInventoryRepository(pool)

    publisher := messaging.NewInventoryPublisher(

        cfg.KafkaBroker,

        cfg.InventoryReservedTopic,

        cfg.InventoryFailedTopic,

    )

    defer publisher.Close()


    service := inventoryapp.NewService(repo, publisher)


    reader := kafkago.NewReader(kafkago.ReaderConfig{

        Brokers:        []string{cfg.KafkaBroker},

        Topic:          cfg.OrderCreatedTopic,

        GroupID:        cfg.KafkaGroupID,

        MinBytes:       10e3, // 10KB

        MaxBytes:       10e6, // 10MB - fetch nhiều hơn mỗi lần

        CommitInterval: 500,  // ms - commit theo batch

    })

    defer reader.Close()


    log.Printf("inventory consumer listening topic=%s group=%s workers=%d",

        cfg.OrderCreatedTopic, cfg.KafkaGroupID, workerCount)


    // Channel để distribute messages cho workers

    msgChan := make(chan kafkago.Message, workerCount*20)


    // Start worker pool

    var wg sync.WaitGroup

    for i := 0; i < workerCount; i++ {

        wg.Add(1)

        go func(workerID int) {

            defer wg.Done()

            for msg := range msgChan {

                var event inventoryapp.OrderCreatedEvent

                if err := json.Unmarshal(msg.Value, &event); err != nil {

                    log.Printf("[Worker%d] failed to unmarshal: %v", workerID, err)

                    continue

                }

                log.Printf("[Worker%d] received order.created order_id=%s items=%d",

                    workerID, event.OrderID, len(event.Items))

                if err := service.HandleOrderCreated(context.Background(), event); err != nil {

                    log.Printf("[Worker%d] failed to handle order_id=%s err=%v",

                        workerID, event.OrderID, err)

                    continue

                }

                log.Printf("[Worker%d] processed order.created successfully order_id=%s",

                    workerID, event.OrderID)

            }

        }(i)

    }


    // Main loop: đọc message và đẩy vào channel

    for {

        msg, err := reader.ReadMessage(context.Background())

        if err != nil {

            log.Printf("failed to read message: %v", err)

            continue

        }

        msgChan <- msg

    }

} 