package services

import (
	"context"

	"github.com/nvnamsss/chat/src/dtos"
)

// ChatService defines the interface for chat operations
type ChatService interface {
	// CreateChat creates a new chat for a user
	CreateChat(ctx context.Context, userID string, req *dtos.ChatRequest) (*dtos.ChatResponse, error)

	// GetChat retrieves a chat by ID
	GetChat(ctx context.Context, id int64) (*dtos.ChatResponse, error)

	// ListChats lists all chats for a user
	ListChats(ctx context.Context, userID string, limit, offset int) (*dtos.ListChatsResponse, error)

	// SearchChats searches chats by title for a user
	SearchChats(ctx context.Context, userID string, req *dtos.SearchChatsRequest) (*dtos.ListChatsResponse, error)

	// UpdateChat updates a chat
	UpdateChat(ctx context.Context, id int64, req *dtos.ChatRequest) (*dtos.ChatResponse, error)

	// DeleteChat deletes a chat
	DeleteChat(ctx context.Context, id int64) error
}
