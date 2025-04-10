package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
	"github.com/nvnamsss/chat/src/repositories"
)

// messageService implements the MessageService interface
type messageService struct {
	messageRepo repositories.MessageRepository
	chatRepo    repositories.ChatRepository
	llmAdapter  adapters.LLMAdapter
	kafka       KafkaProducer
}

// NewMessageService creates a new message service
func NewMessageService(
	messageRepo repositories.MessageRepository,
	chatRepo repositories.ChatRepository,
	llmAdapter adapters.LLMAdapter,
	kafka KafkaProducer,
) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
		llmAdapter:  llmAdapter,
		kafka:       kafka,
	}
}

// SendMessage sends a new user message to a chat and gets LLM response
func (s *messageService) SendMessage(ctx context.Context, chatID int64, userID string, req *dtos.MessageRequest) (*dtos.MessageResponse, error) {
	log := logger.Context(ctx)
	log.Infow("Processing new message", "chatID", chatID, "userID", userID)

	// Verify chat exists
	chat, err := s.chatRepo.Get(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// Verify the user owns the chat
	if chat.UserID != userID {
		return nil, errors.New(errors.ErrForbidden, "User does not have access to this chat")
	}

	// Create user message
	userMessage := &models.Message{
		ChatID:  chatID,
		UserID:  &userID,
		Role:    "user",
		Content: req.Content,
	}

	// Save user message to database
	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, err
	}

	// Publish message event
	userMsgEvent := &dtos.KafkaMessage[dtos.MessagePayload]{
		ID:        uuid.New().String(),
		Event:     models.EventMessageCreated,
		Timestamp: time.Now().Unix(),
		Payload: dtos.MessagePayload{
			MessageID: userMessage.ID,
			ChatID:    userMessage.ChatID,
			UserID:    userMessage.UserID,
			Role:      userMessage.Role,
			Content:   userMessage.Content,
		},
	}

	if err := s.kafka.PublishMessageEvent(ctx, userMsgEvent); err != nil {
		log.Errorw("Failed to publish user message event", "error", err, "messageID", userMessage.ID)
		// Continue despite error
	}

	// Get chat history for context
	messages, _, err := s.messageRepo.GetByChatID(ctx, chatID, 20, 0)
	if err != nil {
		return nil, err
	}

	// Prepare LLM request with context
	llmMessages := make([]dtos.LLMMessage, 0, len(messages)+1)

	// Add previous messages as context (limit to a reasonable number)
	for _, msg := range messages {
		llmMessages = append(llmMessages, dtos.LLMMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add the new message
	llmMessages = append(llmMessages, dtos.LLMMessage{
		Role:    userMessage.Role,
		Content: userMessage.Content,
	})

	// Create LLM request
	llmRequest := &dtos.LLMRequest{
		Messages: llmMessages,
	}

	// Get LLM response
	llmResponse, err := s.llmAdapter.GenerateResponse(ctx, llmRequest)
	if err != nil {
		log.Errorw("LLM request failed", "error", err)
		return nil, errors.Wrap(err, errors.ErrLLMService, "Failed to get response from LLM service")
	}

	// Create assistant message
	assistantMessage := &models.Message{
		ChatID:  chatID,
		Role:    "assistant",
		Content: llmResponse.Message.Content,
	}

	// Save assistant message to database
	if err := s.messageRepo.Create(ctx, assistantMessage); err != nil {
		return nil, err
	}

	// Publish assistant message event
	assistantMsgEvent := &dtos.KafkaMessage[dtos.MessagePayload]{
		ID:        uuid.New().String(),
		Event:     models.EventMessageCreated,
		Timestamp: time.Now().Unix(),
		Payload: dtos.MessagePayload{
			MessageID: assistantMessage.ID,
			ChatID:    assistantMessage.ChatID,
			Role:      assistantMessage.Role,
			Content:   assistantMessage.Content,
		},
	}

	if err := s.kafka.PublishMessageEvent(ctx, assistantMsgEvent); err != nil {
		log.Errorw("Failed to publish assistant message event", "error", err, "messageID", assistantMessage.ID)
		// Continue despite error
	}

	// Return the user's message
	return &dtos.MessageResponse{
		ID:        userMessage.ID,
		ChatID:    userMessage.ChatID,
		UserID:    userMessage.UserID,
		Role:      userMessage.Role,
		Content:   userMessage.Content,
		CreatedAt: userMessage.CreatedAt,
		UpdatedAt: userMessage.UpdatedAt,
	}, nil
}

// GetMessage retrieves a message by ID
func (s *messageService) GetMessage(ctx context.Context, id int64) (*dtos.MessageResponse, error) {
	log := logger.Context(ctx)
	log.Debugw("Getting message", "id", id)

	message, err := s.messageRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dtos.MessageResponse{
		ID:        message.ID,
		ChatID:    message.ChatID,
		UserID:    message.UserID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
	}, nil
}

// ListMessages lists all messages for a chat
func (s *messageService) ListMessages(ctx context.Context, req *dtos.ListMessagesRequest) (*dtos.ListMessagesResponse, error) {
	log := logger.Context(ctx)
	log.Debugw("Listing messages", "chatID", req.ChatID, "limit", req.Limit, "offset", req.Offset)

	if req.Limit <= 0 {
		req.Limit = 50
	}

	messages, total, err := s.messageRepo.GetByChatID(ctx, req.ChatID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	messageResponses := make([]dtos.MessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = dtos.MessageResponse{
			ID:        message.ID,
			ChatID:    message.ChatID,
			UserID:    message.UserID,
			Role:      message.Role,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
			UpdatedAt: message.UpdatedAt,
		}
	}

	return &dtos.ListMessagesResponse{
		Messages: messageResponses,
		Total:    total,
	}, nil
}

// UpdateMessage updates a message
func (s *messageService) UpdateMessage(ctx context.Context, id int64, req *dtos.MessageRequest) (*dtos.MessageResponse, error) {
	log := logger.Context(ctx)
	log.Infow("Updating message", "id", id)

	// Get existing message
	message, err := s.messageRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only allow updating user messages, not assistant messages
	if message.Role != "user" {
		return nil, errors.New(errors.ErrForbidden, "Can only update user messages")
	}

	// Update message
	message.Content = req.Content

	// Save to database
	if err := s.messageRepo.Update(ctx, message); err != nil {
		return nil, err
	}

	// Publish event
	event := &dtos.KafkaMessage[dtos.MessagePayload]{
		ID:        uuid.New().String(),
		Event:     models.EventMessageUpdated,
		Timestamp: time.Now().Unix(),
		Payload: dtos.MessagePayload{
			MessageID: message.ID,
			ChatID:    message.ChatID,
			UserID:    message.UserID,
			Role:      message.Role,
			Content:   message.Content,
		},
	}

	if err := s.kafka.PublishMessageEvent(ctx, event); err != nil {
		log.Errorw("Failed to publish message updated event", "error", err, "messageID", message.ID)
	}

	return &dtos.MessageResponse{
		ID:        message.ID,
		ChatID:    message.ChatID,
		UserID:    message.UserID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
	}, nil
}

// DeleteMessage deletes a message
func (s *messageService) DeleteMessage(ctx context.Context, id int64) error {
	log := logger.Context(ctx)
	log.Infow("Deleting message", "id", id)

	return s.messageRepo.Delete(ctx, id)
}
