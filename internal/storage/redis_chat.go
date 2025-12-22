package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisChat struct {
	client *redis.Client
}

func NewRedisChat(client *redis.Client) *RedisChat {
	return &RedisChat{client: client}
}

func (r *RedisChat) GetMessages(ctx context.Context, userID int64, limit int) ([]Message, error) {
	key := fmt.Sprintf("conv:%d", userID)

	// Get all messages from the list (keeping order)
	raw, err := r.client.LRange(ctx, key, -int64(limit), -1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	messages := make([]Message, 0, len(raw))
	for _, item := range raw {
		var msg Message
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// AppendMessage adds a message to user's conversation history.
func (r *RedisChat) AppendMessage(ctx context.Context, userID int64, message Message) error {
	key := fmt.Sprintf("conv:%d", userID)

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := r.client.RPush(ctx, key, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to append message: %w", err)
	}

	if err := r.client.LTrim(ctx, key, -20, -1).Err(); err != nil {
		return fmt.Errorf("failed to trim messages: %w", err)
	}

	return nil
}

// ClearMessages removes all messages for a user.
func (r *RedisChat) ClearMessages(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("conv:%d", userID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear messages: %w", err)
	}

	return nil
}
