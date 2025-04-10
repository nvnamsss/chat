package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
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

	query := `
		INSERT INTO chats (user_id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.GetDB().QueryRowContext(
		ctx,
		query,
		chat.UserID,
		chat.Title,
		chat.CreatedAt,
		chat.UpdatedAt,
	).Scan(&chat.ID)

	if err != nil {
		log.Errorw("Failed to create chat", "error", err)
		return errors.Wrap(err, errors.ErrInternal, "Failed to create chat")
	}

	return nil
}

// Get retrieves a chat by ID
func (r *chatRepository) Get(ctx context.Context, id int64) (*models.Chat, error) {
	log := logger.Context(ctx)
	var chat models.Chat

	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM chats
		WHERE id = $1
	`

	err := r.db.GetDB().GetContext(ctx, &chat, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debugw("Chat not found", "id", id)
			return nil, errors.New(errors.ErrNotFound, "Chat not found")
		}
		log.Errorw("Failed to get chat", "error", err, "id", id)
		return nil, errors.Wrap(err, errors.ErrInternal, "Failed to get chat")
	}

	return &chat, nil
}

// GetByUserID retrieves all chats for a user
func (r *chatRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Chat, int64, error) {
	log := logger.Context(ctx)
	var chats []*models.Chat
	var total int64

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM chats
		WHERE user_id = $1
	`
	err := r.db.GetDB().GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		log.Errorw("Failed to count chats", "error", err, "userID", userID)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to count chats")
	}

	// Get chats with pagination
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM chats
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`
	err = r.db.GetDB().SelectContext(ctx, &chats, query, userID, limit, offset)
	if err != nil {
		log.Errorw("Failed to get chats", "error", err, "userID", userID)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to get chats")
	}

	return chats, total, nil
}

// Search searches chats by title
func (r *chatRepository) Search(ctx context.Context, req *dtos.SearchChatsRequest, userID string) ([]*models.Chat, int64, error) {
	log := logger.Context(ctx)
	var chats []*models.Chat
	var total int64

	// Prepare search term for LIKE query
	searchTerm := "%" + req.Query + "%"

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM chats
		WHERE user_id = $1 AND title ILIKE $2
	`
	err := r.db.GetDB().GetContext(ctx, &total, countQuery, userID, searchTerm)
	if err != nil {
		log.Errorw("Failed to count chats in search", "error", err, "query", req.Query)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to search chats")
	}

	// Get chats matching search term with pagination
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM chats
		WHERE user_id = $1 AND title ILIKE $2
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`
	err = r.db.GetDB().SelectContext(ctx, &chats, query, userID, searchTerm, req.Limit, req.Offset)
	if err != nil {
		log.Errorw("Failed to search chats", "error", err, "query", req.Query)
		return nil, 0, errors.Wrap(err, errors.ErrInternal, "Failed to search chats")
	}

	return chats, total, nil
}

// Update updates a chat
func (r *chatRepository) Update(ctx context.Context, chat *models.Chat) error {
	log := logger.Context(ctx)
	chat.UpdatedAt = time.Now()

	query := `
		UPDATE chats
		SET title = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.GetDB().ExecContext(ctx, query, chat.Title, chat.UpdatedAt, chat.ID)
	if err != nil {
		log.Errorw("Failed to update chat", "error", err, "id", chat.ID)
		return errors.Wrap(err, errors.ErrInternal, "Failed to update chat")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Chat with ID %d not found", chat.ID))
	}

	return nil
}

// Delete deletes a chat
func (r *chatRepository) Delete(ctx context.Context, id int64) error {
	log := logger.Context(ctx)

	// Start a transaction to delete the chat and all its messages
	tx, err := r.db.GetDB().BeginTxx(ctx, nil)
	if err != nil {
		log.Errorw("Failed to begin transaction", "error", err)
		return errors.Wrap(err, errors.ErrInternal, "Failed to begin transaction")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Errorw("Failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	// Delete all messages for the chat
	_, err = tx.ExecContext(ctx, "DELETE FROM messages WHERE chat_id = $1", id)
	if err != nil {
		log.Errorw("Failed to delete messages for chat", "error", err, "chatID", id)
		return errors.Wrap(err, errors.ErrInternal, "Failed to delete messages")
	}

	// Delete the chat
	result, err := tx.ExecContext(ctx, "DELETE FROM chats WHERE id = $1", id)
	if err != nil {
		log.Errorw("Failed to delete chat", "error", err, "id", id)
		return errors.Wrap(err, errors.ErrInternal, "Failed to delete chat")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Chat with ID %d not found", id))
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Errorw("Failed to commit transaction", "error", err)
		return errors.Wrap(err, errors.ErrInternal, "Failed to commit transaction")
	}

	return nil
}
