package ai

import (
	"context"

	"github.com/Asylann/Socrates/internal/storage"
)

// Interface for AI services.
type AIClient interface {
	Chat(ctx context.Context, messages []storage.Message) (string, error)
}
