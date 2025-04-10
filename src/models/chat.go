package models

import (
	"time"
)

// Chat represents a single chat session
type Chat struct {
	ID        int64     `db:"id"`
	UserID    string    `db:"user_id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Message represents a single message in a chat
type Message struct {
	ID        int64     `db:"id"`
	ChatID    int64     `db:"chat_id"`
	UserID    *string   `db:"user_id"` // Can be null for LLM responses
	Role      string    `db:"role"`    // "user" or "assistant"
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Event types for Kafka messages
const (
	EventChatCreated    = "chat.created"
	EventChatUpdated    = "chat.updated"
	EventMessageCreated = "message.created"
	EventMessageUpdated = "message.updated"
)
