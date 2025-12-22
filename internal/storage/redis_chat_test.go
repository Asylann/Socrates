package storage

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get container port: %v", err)
	}

	redisURL := "redis://" + host + ":" + port.Port()
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	cleanup := func() {
		client.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		container.Terminate(ctx)
	}

	return client, cleanup
}

func TestRedisChat_AppendAndGetMessages(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	storage := NewRedisChat(client)
	userID := int64(123)

	// Add first message
	msg1 := Message{Role: "user", Content: "Hello"}
	err := storage.AppendMessage(ctx, userID, msg1)
	require.NoError(t, err)

	// Add second message
	msg2 := Message{Role: "assistant", Content: "Hi there!"}
	err = storage.AppendMessage(ctx, userID, msg2)
	require.NoError(t, err)

	// Retrieve messages
	messages, err := storage.GetMessages(ctx, userID, 20)
	require.NoError(t, err)

	assert.Len(t, messages, 2)
	assert.Equal(t, msg1.Role, messages[0].Role)
	assert.Equal(t, msg1.Content, messages[0].Content)
	assert.Equal(t, msg2.Role, messages[1].Role)
	assert.Equal(t, msg2.Content, messages[1].Content)
}

func TestRedisChat_KeepsLast20Messages(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	storage := NewRedisChat(client)
	userID := int64(456)

	// Add 30 messages
	for i := 0; i < 30; i++ {
		msg := Message{
			Role:    "user",
			Content: "Message " + string(rune(i)),
		}
		err := storage.AppendMessage(ctx, userID, msg)
		require.NoError(t, err)
	}

	// Should only keep last 20
	messages, err := storage.GetMessages(ctx, userID, 20)
	require.NoError(t, err)

	assert.Len(t, messages, 20)
	assert.Contains(t, messages[0].Content, "Message")
}

func TestRedisChat_ClearMessages(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	storage := NewRedisChat(client)
	userID := int64(789)

	// Add messages
	msg := Message{Role: "user", Content: "Test"}
	err := storage.AppendMessage(ctx, userID, msg)
	require.NoError(t, err)

	// Clear messages
	err = storage.ClearMessages(ctx, userID)
	require.NoError(t, err)

	// Verify cleared
	messages, err := storage.GetMessages(ctx, userID, 20)
	require.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestRedisChat_EmptyHistory(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	storage := NewRedisChat(client)
	userID := int64(999)

	// Get messages for non-existent user
	messages, err := storage.GetMessages(ctx, userID, 20)
	require.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestRedisChat_MultipleUsers(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	storage := NewRedisChat(client)

	user1ID := int64(111)
	user2ID := int64(222)

	// Add messages for user 1
	msg1 := Message{Role: "user", Content: "User 1 message"}
	err := storage.AppendMessage(ctx, user1ID, msg1)
	require.NoError(t, err)

	// Add messages for user 2
	msg2 := Message{Role: "user", Content: "User 2 message"}
	err = storage.AppendMessage(ctx, user2ID, msg2)
	require.NoError(t, err)

	// Verify user 1
	messages1, err := storage.GetMessages(ctx, user1ID, 20)
	require.NoError(t, err)
	assert.Len(t, messages1, 1)
	assert.Equal(t, "User 1 message", messages1[0].Content)

	// Verify user 2
	messages2, err := storage.GetMessages(ctx, user2ID, 20)
	require.NoError(t, err)
	assert.Len(t, messages2, 1)
	assert.Equal(t, "User 2 message", messages2[0].Content)
}
