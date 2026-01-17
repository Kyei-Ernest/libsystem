package worker

import (
	"context"
	"fmt"
	"time"
)

// sendToDLQ sends a failed message to the dead letter queue
func (p *Processor) sendToDLQ(message []byte, processingErr error) error {
	if p.producer == nil {
		return fmt.Errorf("DLQ producer not configured")
	}

	// Create DLQ message with error details
	dlqEvent := map[string]interface{}{
		"original_message": string(message),
		"error":            processingErr.Error(),
		"failed_at":        time.Now().Format(time.RFC3339),
		"retry_count":      "max_exceeded",
	}

	// Send to DLQ with nil key and the DLQ message as value
	return p.producer.Publish(context.Background(), nil, dlqEvent)
}
