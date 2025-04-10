package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
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

	query := `
		INSERT INTO messages (chat_id, user_id, role, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.GetDB().QueryRowContext(
		ctx,
		query,
		message.ChatID,
		message.UserID,
		message.Role,
		message.Content,
		message.CreatedAt,
		message.UpdatedAt,
	).Scan(&message.ID)

	if err != nil {
		log.Errorw("Failed to create message", "error", err)
		return errors.Wrap(err, errors.ErrInternal, "Failed to create message")
	}

	return nil
}

// Get retrieves a message by ID
func (r *messageRepository) Get(ctx context.Context, id int64) (*models.Message, error) {
	log := logger.Context(ctx)
	var message models.Message

	query := `
		SELECT id, chat_id, user_id, role, content, created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	err := r.db.GetDB().GetContext(ctx, &message, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debugw("Message not found", "id", id)
			return nil, errors.New(errors.ErrNotFound, "Message not found")
		}
		log.Errorw("Failed to get message", "error", err, "id", id)
		return nil, errors.Wrap(err, errors.ErrInternal, "Failed to get message")
	}

	return &message, nil
}

// GetByChatID retrieves all messages for a chat
func (r *messageRepository) GetByChatID(ctx context.Context, chatID int64, limit, offset int) ([]*models.Message, int64, error) {
	log := logger.Context(ctx)
	var messages []*models.Message
	var total int64

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM messages
		WHERE chat_id = $1
	`
	err := r.db.GetDB().GetContext(ctx, &total, countQuery, chatID)
	if err != nil {
		log.Errorw("Failed to count messages", "error", err, "chatID", chatID)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to count messages")
	}

	// Get messages with pagination
	query := `
		SELECT id, chat_id, user_id, role, content, created_at, updated_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	err = r.db.GetDB().SelectContext(ctx, &messages, query, chatID, limit, offset)
	if err != nil {
		log.Errorw("Failed to get messages", "error", err, "chatID", chatID)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to get messages")
	}

	return messages, total, nil
}

// Update updates a message
func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	log := logger.Context(ctx)
	message.UpdatedAt = time.Now()

	query := `
		UPDATE messages
		SET content = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.GetDB().ExecContext(ctx, query, message.Content, message.UpdatedAt, message.ID)
	if err != nil {
		log.Errorw("Failed to update message", "error", err, "id", message.ID)
		return errors.Wrap(err, errors.ErrInternal, "Failed to update message")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Message with ID %d not found", message.ID))
	}

	return nil
}

// Delete deletes a message
func (r *messageRepository) Delete(ctx context.Context, id int64) error {
	log := logger.Context(ctx)

	query := "DELETE FROM messages WHERE id = $1"
	result, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		log.Errorw("Failed to delete message", "error", err, "id", id)
		return errors.Wrap(err, errors.ErrInternal, "Failed to delete message")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Message with ID %d not found", id))
	}

	return nil
}
