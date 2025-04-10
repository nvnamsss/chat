package services

import (
	"context"

	"github.com/nvnamsss/chat/src/dtos"
)

// KafkaProducer defines the interface for publishing events to Kafka
type KafkaProducer interface {
	// PublishChatEvent publishes a chat event to Kafka
	PublishChatEvent(ctx context.Context, message *dtos.KafkaMessage[dtos.ChatPayload]) error

	// PublishMessageEvent publishes a message event to Kafka
	PublishMessageEvent(ctx context.Context, message *dtos.KafkaMessage[dtos.MessagePayload]) error
}
