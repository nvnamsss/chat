package services

import (
	"context"

	"github.com/nvnamsss/chat/src/dtos"
)

// MessageService defines the interface for message operations
type MessageService interface {
	// SendMessage sends a new user message to a chat and gets LLM response
	SendMessage(ctx context.Context, chatID int64, userID string, req *dtos.MessageRequest) (*dtos.MessageResponse, error)

	// GetMessage retrieves a message by ID
	GetMessage(ctx context.Context, id int64) (*dtos.MessageResponse, error)

	// ListMessages lists all messages for a chat
	ListMessages(ctx context.Context, req *dtos.ListMessagesRequest) (*dtos.ListMessagesResponse, error)

	// UpdateMessage updates a message
	UpdateMessage(ctx context.Context, id int64, req *dtos.MessageRequest) (*dtos.MessageResponse, error)

	// DeleteMessage deletes a message
	DeleteMessage(ctx context.Context, id int64) error
}
