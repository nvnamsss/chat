package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
	"github.com/nvnamsss/chat/src/repositories"
)

// chatService implements the ChatService interface
type chatService struct {
	chatRepo repositories.ChatRepository
	kafka    KafkaProducer
}

// NewChatService creates a new chat service
func NewChatService(chatRepo repositories.ChatRepository, kafka KafkaProducer) ChatService {
	return &chatService{
		chatRepo: chatRepo,
		kafka:    kafka,
	}
}

// CreateChat creates a new chat for a user
func (s *chatService) CreateChat(ctx context.Context, userID string, req *dtos.ChatRequest) (*dtos.ChatResponse, error) {
	log := logger.Context(ctx)
	log.Infow("Creating new chat", "userID", userID, "title", req.Title)

	// Create chat entity
	chat := &models.Chat{
		UserID: userID,
		Title:  req.Title,
	}

	// Save to database
	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	// Publish event
	event := &dtos.KafkaMessage[dtos.ChatPayload]{
		ID:        uuid.New().String(),
		Event:     models.EventChatCreated,
		Timestamp: time.Now().Unix(),
		Payload: dtos.ChatPayload{
			ChatID: chat.ID,
			UserID: chat.UserID,
			Title:  chat.Title,
		},
	}

	if err := s.kafka.PublishChatEvent(ctx, event); err != nil {
		// Just log the error but don't fail the request
		log.Errorw("Failed to publish chat created event", "error", err, "chatID", chat.ID)
	}

	// Convert to response DTO
	return &dtos.ChatResponse{
		ID:        chat.ID,
		UserID:    chat.UserID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}, nil
}

// GetChat retrieves a chat by ID
func (s *chatService) GetChat(ctx context.Context, id int64) (*dtos.ChatResponse, error) {
	log := logger.Context(ctx)
	log.Debugw("Getting chat", "id", id)

	chat, err := s.chatRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dtos.ChatResponse{
		ID:        chat.ID,
		UserID:    chat.UserID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}, nil
}

// ListChats lists all chats for a user
func (s *chatService) ListChats(ctx context.Context, userID string, limit, offset int) (*dtos.ListChatsResponse, error) {
	log := logger.Context(ctx)
	log.Debugw("Listing chats", "userID", userID, "limit", limit, "offset", offset)

	if limit <= 0 {
		limit = 10
	}

	chats, total, err := s.chatRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	chatResponses := make([]dtos.ChatResponse, len(chats))
	for i, chat := range chats {
		chatResponses[i] = dtos.ChatResponse{
			ID:        chat.ID,
			UserID:    chat.UserID,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
		}
	}

	return &dtos.ListChatsResponse{
		Chats: chatResponses,
		Total: total,
	}, nil
}

// SearchChats searches chats by title for a user
func (s *chatService) SearchChats(ctx context.Context, userID string, req *dtos.SearchChatsRequest) (*dtos.ListChatsResponse, error) {
	log := logger.Context(ctx)
	log.Debugw("Searching chats", "userID", userID, "query", req.Query, "limit", req.Limit, "offset", req.Offset)

	if req.Limit <= 0 {
		req.Limit = 10
	}

	chats, total, err := s.chatRepo.Search(ctx, req, userID)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	chatResponses := make([]dtos.ChatResponse, len(chats))
	for i, chat := range chats {
		chatResponses[i] = dtos.ChatResponse{
			ID:        chat.ID,
			UserID:    chat.UserID,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
		}
	}

	return &dtos.ListChatsResponse{
		Chats: chatResponses,
		Total: total,
	}, nil
}

// UpdateChat updates a chat
func (s *chatService) UpdateChat(ctx context.Context, id int64, req *dtos.ChatRequest) (*dtos.ChatResponse, error) {
	log := logger.Context(ctx)
	log.Infow("Updating chat", "id", id, "title", req.Title)

	// Get existing chat
	chat, err := s.chatRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update chat
	chat.Title = req.Title

	// Save to database
	if err := s.chatRepo.Update(ctx, chat); err != nil {
		return nil, err
	}

	// Publish event
	event := &dtos.KafkaMessage[dtos.ChatPayload]{
		ID:        uuid.New().String(),
		Event:     models.EventChatUpdated,
		Timestamp: time.Now().Unix(),
		Payload: dtos.ChatPayload{
			ChatID: chat.ID,
			UserID: chat.UserID,
			Title:  chat.Title,
		},
	}

	if err := s.kafka.PublishChatEvent(ctx, event); err != nil {
		// Just log the error but don't fail the request
		log.Errorw("Failed to publish chat updated event", "error", err, "chatID", chat.ID)
	}

	return &dtos.ChatResponse{
		ID:        chat.ID,
		UserID:    chat.UserID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}, nil
}

// DeleteChat deletes a chat
func (s *chatService) DeleteChat(ctx context.Context, id int64) error {
	log := logger.Context(ctx)
	log.Infow("Deleting chat", "id", id)

	return s.chatRepo.Delete(ctx, id)
}
