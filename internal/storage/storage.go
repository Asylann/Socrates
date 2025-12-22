package storage

import "context"

type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// ChatStorage defines interface for storing and retrieving conversation history.
type ChatStorage interface {
	GetMessages(ctx context.Context, userID int64, limit int) ([]Message, error)

	AppendMessage(ctx context.Context, userID int64, message Message) error

	ClearMessages(ctx context.Context, userID int64) error
}
