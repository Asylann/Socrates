package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Asylann/Socrates/internal/config"
	"github.com/Asylann/Socrates/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIClient mocks the AIClient interface
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) Chat(ctx context.Context, messages []storage.Message) (string, error) {
	args := m.Called(ctx, messages)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

// MockChatStorage mocks the ChatStorage interface
type MockChatStorage struct {
	mock.Mock
}

func (m *MockChatStorage) GetMessages(ctx context.Context, userID int64, limit int) ([]storage.Message, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.Message), args.Error(1)
}

func (m *MockChatStorage) AppendMessage(ctx context.Context, userID int64, message storage.Message) error {
	args := m.Called(ctx, userID, message)
	return args.Error(0)
}

func (m *MockChatStorage) ClearMessages(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestProcessMessage_Success(t *testing.T) {
	ctx := context.Background()
	userID := int64(12345)
	userMessage := "Hello, how are you?"

	// Setup mocks
	mockStorage := new(MockChatStorage)
	mockAI := new(MockAIClient)

	// GetMessages should return empty list initially
	mockStorage.On("GetMessages", ctx, userID, 20).Return([]storage.Message{}, nil)

	// AppendMessage should succeed for user message
	mockStorage.On("AppendMessage", ctx, userID, storage.Message{
		Role:    "user",
		Content: userMessage,
	}).Return(nil)

	// Chat should return a response
	expectedResponse := "I'm doing well, thank you!"
	mockAI.On("Chat", ctx, []storage.Message{
		{Role: "system", Content: config.DefaultSystemPrompt()},
		{Role: "user", Content: userMessage},
	}).Return(expectedResponse, nil)

	// AppendMessage should succeed for assistant message
	mockStorage.On("AppendMessage", ctx, userID, storage.Message{
		Role:    "assistant",
		Content: expectedResponse,
	}).Return(nil)

	// Create service and test
	service := NewChatService(mockAI, mockStorage, config.DefaultSystemPrompt())
	response, err := service.ProcessMessage(ctx, userID, userMessage)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockStorage.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

func TestProcessMessage_AIError(t *testing.T) {
	ctx := context.Background()
	userID := int64(12345)
	userMessage := "Hello"

	mockStorage := new(MockChatStorage)
	mockAI := new(MockAIClient)

	mockStorage.On("GetMessages", ctx, userID, 20).Return([]storage.Message{}, nil)
	mockStorage.On("AppendMessage", ctx, userID, storage.Message{
		Role:    "user",
		Content: userMessage,
	}).Return(nil)

	// AI call fails
	mockAI.On("Chat", ctx, mock.MatchedBy(func(msgs []storage.Message) bool {
		return len(msgs) == 2 && msgs[0].Role == "system" && msgs[1].Role == "user"
	})).Return("", errors.New("ai service error"))

	service := NewChatService(mockAI, mockStorage, config.DefaultSystemPrompt())
	_, err := service.ProcessMessage(ctx, userID, userMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ai chat failed")
}

func TestProcessMessage_WithHistory(t *testing.T) {
	ctx := context.Background()
	userID := int64(12345)
	userMessage := "What was I asking?"

	mockStorage := new(MockChatStorage)
	mockAI := new(MockAIClient)

	// Simulate previous conversation
	previousMessages := []storage.Message{
		{Role: "user", Content: "What is Go?"},
		{Role: "assistant", Content: "Go is a compiled language."},
	}

	mockStorage.On("GetMessages", ctx, userID, 20).Return(previousMessages, nil)
	mockStorage.On("AppendMessage", ctx, userID, storage.Message{
		Role:    "user",
		Content: userMessage,
	}).Return(nil)

	expectedResponse := "You were asking about Go programming language."
	mockAI.On("Chat", ctx, []storage.Message{
		{Role: "system", Content: config.DefaultSystemPrompt()},
		{Role: "user", Content: "What is Go?"},
		{Role: "assistant", Content: "Go is a compiled language."},
		{Role: "user", Content: userMessage},
	}).Return(expectedResponse, nil)

	mockStorage.On("AppendMessage", ctx, userID, storage.Message{
		Role:    "assistant",
		Content: expectedResponse,
	}).Return(nil)

	service := NewChatService(mockAI, mockStorage, config.DefaultSystemPrompt())
	response, err := service.ProcessMessage(ctx, userID, userMessage)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestGetConversationHistory_Success(t *testing.T) {
	ctx := context.Background()
	userID := int64(12345)

	mockStorage := new(MockChatStorage)

	expectedMessages := []storage.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	mockStorage.On("GetMessages", ctx, userID, 20).Return(expectedMessages, nil)

	service := NewChatService(nil, mockStorage, config.DefaultSystemPrompt())
	messages, err := service.GetConversationHistory(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedMessages, messages)
}

func TestClearConversation_Success(t *testing.T) {
	ctx := context.Background()
	userID := int64(12345)

	mockStorage := new(MockChatStorage)
	mockStorage.On("ClearMessages", ctx, userID).Return(nil)

	service := NewChatService(nil, mockStorage, config.DefaultSystemPrompt())
	err := service.ClearConversation(ctx, userID)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestClearConversation_Error(t *testing.T) {
	ctx := context.Background()
	userID := int64(12345)

	mockStorage := new(MockChatStorage)
	mockStorage.On("ClearMessages", ctx, userID).Return(errors.New("redis error"))

	service := NewChatService(nil, mockStorage, config.DefaultSystemPrompt())
	err := service.ClearConversation(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clear messages")
}
