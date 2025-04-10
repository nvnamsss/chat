package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/services"
)

// MessageController handles HTTP requests related to messages
type MessageController struct {
	messageService services.MessageService
	chatService    services.ChatService
}

// NewMessageController creates a new message controller
func NewMessageController(messageService services.MessageService, chatService services.ChatService) *MessageController {
	return &MessageController{
		messageService: messageService,
		chatService:    chatService,
	}
}

// RegisterRoutes registers the controller routes with the router
func (c *MessageController) RegisterRoutes(router *gin.RouterGroup) {
	messages := router.Group("/messages")
	{
		messages.POST("", c.SendMessage)
		messages.GET("", c.ListMessages)
		messages.GET("/:id", c.GetMessage)
		messages.PUT("/:id", c.UpdateMessage)
		messages.DELETE("/:id", c.DeleteMessage)
	}
}

// SendMessage handles sending a new message to a chat and getting a response from the LLM
func (c *MessageController) SendMessage(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse chat ID from query parameter
	chatIDStr := ctx.Query("chatId")
	if chatIDStr == "" {
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Missing chat ID"))
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid chat ID", "chatID", chatIDStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid chat ID"))
		return
	}

	// Parse request
	var req dtos.MessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorw("Failed to parse send message request", "error", err)
		respondError(ctx, errors.Wrap(err, errors.ErrInvalidRequest, "Invalid request format"))
		return
	}

	// Send message
	message, err := c.messageService.SendMessage(ctx.Request.Context(), chatID, userID, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, message)
}

// GetMessage handles getting a single message by ID
func (c *MessageController) GetMessage(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse message ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid message ID", "id", idStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid message ID"))
		return
	}

	// Get message
	message, err := c.messageService.GetMessage(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}

	// Verify the user has access to this chat
	chat, err := c.chatService.GetChat(ctx.Request.Context(), message.ChatID)
	if err != nil {
		respondError(ctx, err)
		return
	}

	if chat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this message"))
		return
	}

	ctx.JSON(http.StatusOK, message)
}

// ListMessages handles listing all messages for a chat
func (c *MessageController) ListMessages(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse request parameters
	var req dtos.ListMessagesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Errorw("Failed to parse list messages request", "error", err)
		respondError(ctx, errors.Wrap(err, errors.ErrInvalidRequest, "Invalid request format"))
		return
	}

	// Verify the user has access to this chat
	chat, err := c.chatService.GetChat(ctx.Request.Context(), req.ChatID)
	if err != nil {
		respondError(ctx, err)
		return
	}

	if chat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this chat"))
		return
	}

	// Get messages
	messages, err := c.messageService.ListMessages(ctx.Request.Context(), &req)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// UpdateMessage handles updating a message
func (c *MessageController) UpdateMessage(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse message ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid message ID", "id", idStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid message ID"))
		return
	}

	// Get the message first to check ownership
	existingMessage, err := c.messageService.GetMessage(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}

	// Verify the user has access to this chat
	chat, err := c.chatService.GetChat(ctx.Request.Context(), existingMessage.ChatID)
	if err != nil {
		respondError(ctx, err)
		return
	}

	if chat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this message"))
		return
	}

	// Parse request
	var req dtos.MessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorw("Failed to parse update message request", "error", err)
		respondError(ctx, errors.Wrap(err, errors.ErrInvalidRequest, "Invalid request format"))
		return
	}

	// Update message
	message, err := c.messageService.UpdateMessage(ctx.Request.Context(), id, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, message)
}

// DeleteMessage handles deleting a message
func (c *MessageController) DeleteMessage(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse message ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid message ID", "id", idStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid message ID"))
		return
	}

	// Get the message first to check ownership
	existingMessage, err := c.messageService.GetMessage(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}

	// Verify the user has access to this chat
	chat, err := c.chatService.GetChat(ctx.Request.Context(), existingMessage.ChatID)
	if err != nil {
		respondError(ctx, err)
		return
	}

	if chat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this message"))
		return
	}

	// Delete message
	if err := c.messageService.DeleteMessage(ctx.Request.Context(), id); err != nil {
		respondError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
