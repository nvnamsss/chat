package dtos

import (
	"time"
)

// ChatRequest represents a request to create a new chat
type ChatRequest struct {
	Title string `json:"title" binding:"required"`
}

// ChatResponse represents a chat in API responses
type ChatResponse struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListChatsResponse represents a list of chats in API responses
type ListChatsResponse struct {
	Chats []ChatResponse `json:"chats"`
	Total int64          `json:"total"`
}

// SearchChatsRequest represents a request to search chats
type SearchChatsRequest struct {
	Query  string `form:"query"`
	Limit  int    `form:"limit,default=10"`
	Offset int    `form:"offset,default=0"`
}

// KafkaMessage is a generic structure for Kafka messages with a typed payload
type KafkaMessage[T any] struct {
	ID        string `json:"id"`
	Event     string `json:"event"`
	Timestamp int64  `json:"timestamp"`
	Payload   T      `json:"payload"`
}

// ChatPayload represents the payload for chat-related Kafka messages
type ChatPayload struct {
	ChatID int64  `json:"chatId"`
	UserID string `json:"userId"`
	Title  string `json:"title"`
}
