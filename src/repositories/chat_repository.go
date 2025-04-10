package repositories

import (
	"context"

	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/models"
)

// ChatRepository defines the interface for chat data access
type ChatRepository interface {
	// Create creates a new chat
	Create(ctx context.Context, chat *models.Chat) error

	// Get retrieves a chat by ID
	Get(ctx context.Context, id int64) (*models.Chat, error)

	// GetByUserID retrieves all chats for a user
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Chat, int64, error)

	// Search searches chats by title
	Search(ctx context.Context, req *dtos.SearchChatsRequest, userID string) ([]*models.Chat, int64, error)

	// Update updates a chat
	Update(ctx context.Context, chat *models.Chat) error

	// Delete deletes a chat
	Delete(ctx context.Context, id int64) error
}
