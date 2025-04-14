package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
	"gorm.io/gorm"
)

// chatRepository implements the ChatRepository interface
type chatRepository struct {
	db adapters.DBAdapter
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db adapters.DBAdapter) ChatRepository {
	return &chatRepository{db: db}
}

// Create creates a new chat
func (r *chatRepository) Create(ctx context.Context, chat *models.Chat) error {
	log := logger.Context(ctx)
	now := time.Now()
	chat.CreatedAt = now
	chat.UpdatedAt = now

	result := r.db.GetDB().WithContext(ctx).Create(chat)
	if result.Error != nil {
		log.Errorw("Failed to create chat", "error", result.Error)
		return errors.Wrap(result.Error, errors.ErrInternal, "Failed to create chat")
	}

	return nil
}

// Get retrieves a chat by ID
func (r *chatRepository) Get(ctx context.Context, id int64) (*models.Chat, error) {
	log := logger.Context(ctx)
	var chat models.Chat

	result := r.db.GetDB().WithContext(ctx).First(&chat, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Debugw("Chat not found", "id", id)
			return nil, errors.New(errors.ErrNotFound, "Chat not found")
		}
		log.Errorw("Failed to get chat", "error", result.Error, "id", id)
		return nil, errors.Wrap(result.Error, errors.ErrInternal, "Failed to get chat")
	}

	return &chat, nil
}

// GetByUserID retrieves all chats for a user
func (r *chatRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Chat, int64, error) {
	log := logger.Context(ctx)
	var chats []*models.Chat
	var total int64

	// Get total count
	result := r.db.GetDB().WithContext(ctx).Model(&models.Chat{}).Where("user_id = ?", userID).Count(&total)
	if result.Error != nil {
		log.Errorw("Failed to count chats", "error", result.Error, "userID", userID)
		return nil, 0, errors.Wrap(result.Error, errors.ErrInternal, "Failed to count chats")
	}

	// Get chats with pagination
	result = r.db.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&chats)

	if result.Error != nil {
		log.Errorw("Failed to get chats", "error", result.Error, "userID", userID)
		return nil, 0, errors.Wrap(result.Error, errors.ErrInternal, "Failed to get chats")
	}

	return chats, total, nil
}

// Search searches chats by title
func (r *chatRepository) Search(ctx context.Context, req *dtos.SearchChatsRequest, userID string) ([]*models.Chat, int64, error) {
	log := logger.Context(ctx)
	var chats []*models.Chat
	var total int64

	db := r.db.GetDB().WithContext(ctx)
	query := db.Model(&models.Chat{}).Where("user_id = ?", userID)

	if req.Query != "" {
		query = query.Where("title ILIKE ?", "%"+req.Query+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		log.Errorw("Failed to count chats in search", "error", err, "query", req.Query)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to search chats")
	}

	// Get chats with pagination
	if err := query.Order("updated_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&chats).Error; err != nil {
		log.Errorw("Failed to search chats", "error", err, "query", req.Query)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to search chats")
	}

	return chats, total, nil
}

// Update updates a chat
func (r *chatRepository) Update(ctx context.Context, chat *models.Chat) error {
	log := logger.Context(ctx)
	chat.UpdatedAt = time.Now()

	result := r.db.GetDB().WithContext(ctx).Model(chat).Updates(map[string]interface{}{
		"title":      chat.Title,
		"updated_at": chat.UpdatedAt,
	})

	if result.Error != nil {
		log.Errorw("Failed to update chat", "error", result.Error, "id", chat.ID)
		return errors.Wrap(result.Error, errors.ErrInternal, "Failed to update chat")
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Chat with ID %d not found", chat.ID))
	}

	return nil
}

// Delete deletes a chat
func (r *chatRepository) Delete(ctx context.Context, id int64) error {
	log := logger.Context(ctx)

	// Start a transaction
	tx := r.db.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		log.Errorw("Failed to begin transaction", "error", tx.Error)
		return errors.Wrap(tx.Error, errors.ErrInternal, "Failed to begin transaction")
	}

	// Rollback transaction on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete the chat (messages will be deleted automatically due to ON DELETE CASCADE)
	result := tx.Delete(&models.Chat{}, id)
	if result.Error != nil {
		tx.Rollback()
		log.Errorw("Failed to delete chat", "error", result.Error, "id", id)
		return errors.Wrap(result.Error, errors.ErrInternal, "Failed to delete chat")
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Chat with ID %d not found", id))
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Errorw("Failed to commit transaction", "error", err)
		return errors.Wrap(err, errors.ErrInternal, "Failed to commit transaction")
	}

	return nil
}
