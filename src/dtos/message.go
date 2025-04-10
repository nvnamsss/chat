package dtos

import (
	"time"
)

// MessageRequest represents a request to create a new message
type MessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chatId"`
	UserID    *string   `json:"userId,omitempty"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListMessagesResponse represents a list of messages in API responses
type ListMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
}

// ListMessagesRequest represents a request to list messages in a chat
type ListMessagesRequest struct {
	ChatID int64 `form:"chatId" binding:"required"`
	Limit  int   `form:"limit,default=50"`
	Offset int   `form:"offset,default=0"`
}

// MessagePayload represents the payload for message-related Kafka messages
type MessagePayload struct {
	MessageID int64   `json:"messageId"`
	ChatID    int64   `json:"chatId"`
	UserID    *string `json:"userId,omitempty"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
}

// LLMRequest represents a request to the LLM vendor service
type LLMRequest struct {
	Messages  []LLMMessage `json:"messages"`
	Model     string       `json:"model,omitempty"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

// LLMMessage represents a single message in an LLM request
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents a response from the LLM vendor service
type LLMResponse struct {
	Message  LLMMessage `json:"message"`
	Usage    LLMUsage   `json:"usage"`
	Model    string     `json:"model"`
	Finished bool       `json:"finished"`
}

// LLMUsage represents token usage information from the LLM vendor
type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
