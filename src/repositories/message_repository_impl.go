package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
	"gorm.io/gorm"
)

// messageRepository implements the MessageRepository interface
type messageRepository struct {
	db adapters.DBAdapter
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db adapters.DBAdapter) MessageRepository {
	return &messageRepository{db: db}
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	log := logger.Context(ctx)
	now := time.Now()
	message.CreatedAt = now
	message.UpdatedAt = now

	result := r.db.GetDB().WithContext(ctx).Create(message)
	if result.Error != nil {
		log.Errorw("Failed to create message", "error", result.Error)
		return errors.Wrap(result.Error, errors.ErrInternal, "Failed to create message")
	}

	return nil
}

// Get retrieves a message by ID
func (r *messageRepository) Get(ctx context.Context, id int64) (*models.Message, error) {
	log := logger.Context(ctx)
	var message models.Message

	result := r.db.GetDB().WithContext(ctx).First(&message, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Debugw("Message not found", "id", id)
			return nil, errors.New(errors.ErrNotFound, "Message not found")
		}
		log.Errorw("Failed to get message", "error", result.Error, "id", id)
		return nil, errors.Wrap(result.Error, errors.ErrInternal, "Failed to get message")
	}

	return &message, nil
}

// GetByChatID retrieves all messages for a chat
func (r *messageRepository) GetByChatID(ctx context.Context, chatID int64, limit, offset int) ([]*models.Message, int64, error) {
	log := logger.Context(ctx)
	var messages []*models.Message
	var total int64

	db := r.db.GetDB().WithContext(ctx)

	// Get total count
	if err := db.Model(&models.Message{}).Where("chat_id = ?", chatID).Count(&total).Error; err != nil {
		log.Errorw("Failed to count messages", "error", err, "chatID", chatID)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to count messages")
	}

	// Get messages with pagination
	if err := db.Where("chat_id = ?", chatID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		log.Errorw("Failed to get messages", "error", err, "chatID", chatID)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to get messages")
	}

	return messages, total, nil
}

// Update updates a message
func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	log := logger.Context(ctx)
	message.UpdatedAt = time.Now()

	result := r.db.GetDB().WithContext(ctx).Model(message).Updates(map[string]interface{}{
		"content":    message.Content,
		"updated_at": message.UpdatedAt,
	})

	if result.Error != nil {
		log.Errorw("Failed to update message", "error", result.Error, "id", message.ID)
		return errors.Wrap(result.Error, errors.ErrInternal, "Failed to update message")
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Message with ID %d not found", message.ID))
	}

	return nil
}

// Delete deletes a message
func (r *messageRepository) Delete(ctx context.Context, id int64) error {
	log := logger.Context(ctx)

	result := r.db.GetDB().WithContext(ctx).Delete(&models.Message{}, id)
	if result.Error != nil {
		log.Errorw("Failed to delete message", "error", result.Error, "id", id)
		return errors.Wrap(result.Error, errors.ErrInternal, "Failed to delete message")
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Message with ID %d not found", id))
	}

	return nil
}
