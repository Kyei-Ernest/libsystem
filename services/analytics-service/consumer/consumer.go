package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/analytics-service/models"
	"github.com/Kyei-Ernest/libsystem/services/analytics-service/repository"
	"github.com/Kyei-Ernest/libsystem/shared/kafka"
	"github.com/google/uuid"
)

type AnalyticsConsumer struct {
	consumer *kafka.Consumer
	repo     repository.AnalyticsRepository
}

func NewAnalyticsConsumer(brokers []string, groupID string, repo repository.AnalyticsRepository) *AnalyticsConsumer {
	// We need to subscribe to multiple topics. The shared Consumer structure might only support one topic per instance
	// or accept a list. Let's check shared consumer.
	// If shared consumer only supports one topic, we might need multiple instances or modify shared consumer.
	// For now, let's assume we create one consumer per topic or list of topics.
	// Actually, kafka-go reader can take a list of topics (GroupTopics).

	// Assuming shared/kafka/consumer.go supports simple config.
	// We will initialize it in Start() or here.

	return &AnalyticsConsumer{
		// consumers created in Start
		repo: repo,
	}
}

// Note: Using the shared consumer helper might be restrictive if it only allows one topic.
// Let's implement the consumption loop here using the shared consumer as a base or helper.

func (c *AnalyticsConsumer) Start(ctx context.Context, brokers []string) error {
	// We want to listen to document.viewed and document.downloaded
	// Simple approach: One consumer for each topic to avoid complexity with shared lib

	go c.consumeTopic(ctx, brokers, "document.viewed")
	go c.consumeTopic(ctx, brokers, "document.downloaded")

	return nil
}

func (c *AnalyticsConsumer) consumeTopic(ctx context.Context, brokers []string, topic string) {
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "analytics-service",
	})

	// Create handler
	handler := func(msg []byte) error {
		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			return err
		}

		event := &models.AnalyticsEvent{
			EventType: models.EventType(topic), // document.viewed or document.downloaded
			CreatedAt: time.Now(),
		}

		// Parse fields
		if idStr, ok := payload["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				event.DocumentID = id
			}
		}

		if uidStr, ok := payload["user_id"].(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				event.UserID = &uid
			}
		}

		if tsStr, ok := payload["occurred_at"].(string); ok {
			if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
				event.OccurredAt = ts
			} else {
				event.OccurredAt = time.Now()
			}
		} else {
			event.OccurredAt = time.Now()
		}

		// Store other fields as metadata
		event.Metadata = payload

		// Save to DB
		if err := c.repo.Create(event); err != nil {
			log.Printf("Failed to save event: %v", err)
			return err
		}

		return nil
	}

	log.Printf("Starting consumer for topic: %s", topic)

	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("Consumer error for %s: %v", topic, err)
			if ctx.Err() != nil {
				return
			}
			time.Sleep(time.Second) // Backoff on error
			continue
		}

		if err := handler(msg.Value); err != nil {
			log.Printf("Handler error for %s: %v", topic, err)
		}
	}
}
