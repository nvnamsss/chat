package repositories

import (
	"context"

	"github.com/nvnamsss/chat/src/models"
)

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, message *models.Message) error

	// Get retrieves a message by ID
	Get(ctx context.Context, id int64) (*models.Message, error)

	// GetByChatID retrieves all messages for a chat
	GetByChatID(ctx context.Context, chatID int64, limit, offset int) ([]*models.Message, int64, error)

	// Update updates a message
	Update(ctx context.Context, message *models.Message) error

	// Delete deletes a message
	Delete(ctx context.Context, id int64) error
}
