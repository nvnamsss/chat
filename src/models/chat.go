package models

import (
	"time"
)

// Chat represents a single chat session
type Chat struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	UserID    string    `gorm:"column:user_id;not null;index"`
	Title     string    `gorm:"column:title;not null;check:title <> ''"`
	Messages  []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName specifies the table name for Chat
func (Chat) TableName() string {
	return "chats"
}

// Message represents a single message in a chat
type Message struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	ChatID    int64     `gorm:"column:chat_id;not null;index"`
	Chat      Chat      `gorm:"foreignKey:ChatID"`
	UserID    *string   `gorm:"column:user_id"`       // Can be null for LLM responses
	Role      string    `gorm:"column:role;not null"` // "user" or "assistant"
	Content   string    `gorm:"column:content;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "messages"
}

// Event types for Kafka messages
const (
	EventChatCreated    = "chat.created"
	EventChatUpdated    = "chat.updated"
	EventMessageCreated = "message.created"
	EventMessageUpdated = "message.updated"
)
