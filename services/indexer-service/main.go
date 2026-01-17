package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Kyei-Ernest/libsystem/services/indexer-service/worker"
	"github.com/Kyei-Ernest/libsystem/shared/elasticsearch"
	"github.com/Kyei-Ernest/libsystem/shared/kafka"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	log.Println("Indexer Service Starting...")

	// Configuration
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9093"), ",")
	kafkaTopic := getEnv("KAFKA_TOPIC", "document.uploaded") // Consuming uploaded events
	esAddress := getEnv("ELASTICSEARCH_URL", "http://localhost:9200")

	// Initialize Elasticsearch Client
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esAddress},
	})
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Verify ES Connection
	info, err := esClient.Info().Do(context.Background())
	if err != nil {
		log.Printf("Warning: Could not connect to Elasticsearch: %v", err)
	} else {
		log.Printf("Connected to Elasticsearch: %v", info.Version)
	}

	// Initialize Kafka Consumer
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: kafkaBrokers,
		Topic:   kafkaTopic,
		GroupID: "indexer-service-group",
	})
	defer consumer.Close()

	// Initialize MinIO Client
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin123")
	minioUseSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Initialize DLQ Producer
	dlqTopic := kafkaTopic + "-dlq"
	dlqProducer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: kafkaBrokers,
		Topic:   dlqTopic,
	})
	defer dlqProducer.Close()

	// Initialize Processor
	processor := worker.NewProcessor(esClient, minioClient, dlqProducer, "documents", dlqTopic)

	log.Printf("Listening for events on topic %s...", kafkaTopic)

	// Initialize Deletion Consumer
	deletionTopic := "document.deleted"
	deleteConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: kafkaBrokers,
		Topic:   deletionTopic,
		GroupID: "indexer-service-deletion-group",
	})
	defer deleteConsumer.Close()

	// Graceful Shutdown Support
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-stop
		log.Println("Shutting down...")
		cancel()
	}()

	// Deletion Consumption Loop (Background)
	go func() {
		log.Printf("Listening for deletion events on topic %s...", deletionTopic)
		for {
			msg, err := deleteConsumer.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error reading deletion message: %v", err)
				continue
			}

			log.Printf("Processing deletion for %s", string(msg.Value))
			if err := processor.Delete(ctx, msg.Value); err != nil {
				log.Printf("Failed to process deletion: %v", err)
			}
		}
	}()

	// Main consumption loop (Uploads)
	log.Println("Starting to consume upload messages...")

	for {
		msg, err := consumer.ReadMessage(ctx) // Use the context for graceful shutdown
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, exit loop
				log.Println("Context cancelled, stopping message consumption.")
				break
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		log.Printf("Processing message from topic %s, partition %d, offset %d",
			msg.Topic, msg.Partition, msg.Offset)

		// Process with built-in retry logic from processor
		processingErr := processor.Process(ctx, msg.Value)

		if processingErr != nil {
			log.Printf("Failed to process message: %v", processingErr)
		} else {
			log.Printf("Successfully processed message")
		}
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
