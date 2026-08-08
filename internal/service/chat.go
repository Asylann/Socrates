package service

import (
	"context"
	"fmt"

	"github.com/Asylann/Socrates/internal/i"
	"github.com/Asylann/Socrates/internal/storage"
)

// ChatService handles conversation logic.
type ChatService struct {
	aiClient     i.AIClient
	storage      storage.ChatStorage
	systemPrompt string
}

func NewChatService(aiClient ai.AIClient, storage storage.ChatStorage, systemPrompt string) *ChatService {
	return &ChatService{
		aiClient:     aiClient,
		storage:      storage,
		systemPrompt: systemPrompt,
	}
}

// ProcessMessage processes a user message and returns AI response.
func (cs *ChatService) ProcessMessage(ctx context.Context, userID int64, userMessage string) (string, error) {
	// last 20 messages saved>
	messages, err := cs.storage.GetMessages(ctx, userID, 20)
	if err != nil {
		return "", fmt.Errorf("failed to load messages: %w", err)
	}

	userMsg := storage.Message{
		Role:    "user",
		Content: userMessage,
	}
	messages = append(messages, userMsg)
	if cs.systemPrompt != "" {
		messages = append([]storage.Message{{Role: "system", Content: cs.systemPrompt}}, messages...)
	}

	if err := cs.storage.AppendMessage(ctx, userID, userMsg); err != nil {
		return "", fmt.Errorf("failed to store user message: %w", err)
	}
	aiResponse, err := cs.aiClient.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("ai chat failed: %w", err)
	}

	assistantMsg := storage.Message{
		Role:    "assistant",
		Content: aiResponse,
	}
	if err := cs.storage.AppendMessage(ctx, userID, assistantMsg); err != nil {
		return "", fmt.Errorf("failed to store assistant message: %w", err)
	}

	return aiResponse, nil
}

// returns the conversation history for a user.
func (cs *ChatService) GetConversationHistory(ctx context.Context, userID int64) ([]storage.Message, error) {
	messages, err := cs.storage.GetMessages(ctx, userID, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	return messages, nil
}

// clears the conversation history for a user.
func (cs *ChatService) ClearConversation(ctx context.Context, userID int64) error {
	if err := cs.storage.ClearMessages(ctx, userID); err != nil {
		return fmt.Errorf("failed to clear messages: %w", err)
	}
	return nil
}
