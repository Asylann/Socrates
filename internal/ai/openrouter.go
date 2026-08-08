package ai

import (
	"context"
	"fmt"

	openrouter "github.com/wojtess/openrouter-api-go"

	"github.com/Asylann/Socrates/internal/storage"
)

// OpenRouterClient that implements interface AIClient
type OpenRouterClient struct {
	or    openrouter.OpenRouterClient
	model string
}

func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	return &OpenRouterClient{
		or:    openrouter.NewOpenRouterClient(apiKey),
		model: model,
	}
}

// Chat sends messages to OpenRouter and returns the response.
func (c *OpenRouterClient) Chat(ctx context.Context, messages []storage.Message) (string, error) {
	msgRequests := make([]openrouter.MessageRequest, 0, len(messages))
	for _, m := range messages {
		var role openrouter.MessageRole
		switch m.Role {
		case "user":
			role = openrouter.RoleUser
		case "assistant":
			role = openrouter.RoleAssistant
		case "system":
			role = openrouter.RoleSystem
		default:
			role = openrouter.Roleuser
		}

		msgRequests = append(msgRequests, openrouter.MessageRequest{
			Role:    role,
			Content: openrouter.TextContent(m.Content),
		})
	}

	// Call OpenRouter API
	request := openrouter.Request{
		Model:    c.model,
		Messages: msgRequests,
	}

	resp, err := c.or.FetchChatCompletions(request)
	if err != nil {
		return "", fmt.Errorf("openrouter api call failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from openrouter")
	}
	if resp.Choices[0].Message == nil {
		return "", fmt.Errorf("empty message in response")
	}

	return resp.Choices[0].Message.Content, nil
}
