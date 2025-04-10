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

// ChatController handles HTTP requests related to chats
type ChatController struct {
	chatService services.ChatService
}

// NewChatController creates a new chat controller
func NewChatController(chatService services.ChatService) *ChatController {
	return &ChatController{
		chatService: chatService,
	}
}

// RegisterRoutes registers the controller routes with the router
func (c *ChatController) RegisterRoutes(router *gin.RouterGroup) {
	chats := router.Group("/chats")
	{
		chats.POST("", c.CreateChat)
		chats.GET("", c.ListChats)
		chats.GET("/search", c.SearchChats)
		chats.GET("/:id", c.GetChat)
		chats.PUT("/:id", c.UpdateChat)
		chats.DELETE("/:id", c.DeleteChat)
	}
}

// CreateChat handles the creation of a new chat
func (c *ChatController) CreateChat(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse request
	var req dtos.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorw("Failed to parse create chat request", "error", err)
		respondError(ctx, errors.Wrap(err, errors.ErrInvalidRequest, "Invalid request format"))
		return
	}

	// Create chat
	chat, err := c.chatService.CreateChat(ctx.Request.Context(), userID, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, chat)
}

// GetChat handles getting a single chat by ID
func (c *ChatController) GetChat(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse chat ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid chat ID", "id", idStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid chat ID"))
		return
	}

	// Get chat
	chat, err := c.chatService.GetChat(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}

	// Verify the user owns the chat
	if chat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this chat"))
		return
	}

	ctx.JSON(http.StatusOK, chat)
}

// ListChats handles listing chats for the authenticated user
func (c *ChatController) ListChats(ctx *gin.Context) {
	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse pagination parameters
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get chats
	response, err := c.chatService.ListChats(ctx.Request.Context(), userID, limit, offset)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// SearchChats handles searching chats by title
func (c *ChatController) SearchChats(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse search parameters
	var req dtos.SearchChatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Errorw("Failed to parse search request", "error", err)
		respondError(ctx, errors.Wrap(err, errors.ErrInvalidRequest, "Invalid search parameters"))
		return
	}

	// Search chats
	response, err := c.chatService.SearchChats(ctx.Request.Context(), userID, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateChat handles updating a chat
func (c *ChatController) UpdateChat(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse chat ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid chat ID", "id", idStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid chat ID"))
		return
	}

	// Get existing chat to verify ownership
	existingChat, err := c.chatService.GetChat(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}

	// Verify the user owns the chat
	if existingChat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this chat"))
		return
	}

	// Parse request
	var req dtos.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorw("Failed to parse update chat request", "error", err)
		respondError(ctx, errors.Wrap(err, errors.ErrInvalidRequest, "Invalid request format"))
		return
	}

	// Update chat
	chat, err := c.chatService.UpdateChat(ctx.Request.Context(), id, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, chat)
}

// DeleteChat handles deleting a chat
func (c *ChatController) DeleteChat(ctx *gin.Context) {
	log := logger.Context(ctx.Request.Context())

	// Get user ID from JWT token
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		respondError(ctx, errors.New(errors.ErrUnauthorized, "User not authenticated"))
		return
	}

	// Parse chat ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorw("Invalid chat ID", "id", idStr, "error", err)
		respondError(ctx, errors.New(errors.ErrInvalidRequest, "Invalid chat ID"))
		return
	}

	// Get existing chat to verify ownership
	existingChat, err := c.chatService.GetChat(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}

	// Verify the user owns the chat
	if existingChat.UserID != userID {
		respondError(ctx, errors.New(errors.ErrForbidden, "User does not have access to this chat"))
		return
	}

	// Delete chat
	if err := c.chatService.DeleteChat(ctx.Request.Context(), id); err != nil {
		respondError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// getUserIDFromContext extracts the user ID from the JWT token in the context
func getUserIDFromContext(ctx *gin.Context) string {
	// In a real application, this would be set by the auth middleware
	// based on the JWT token
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		return ""
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		return ""
	}

	return userID
}
