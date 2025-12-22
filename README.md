# Socrates AI - Telegram Bot

A production-ready Telegram bot powered by OpenRouter AI with Redis-backed conversation memory and per-user rate limiting.

## Features

- **AI-Powered Responses**: Uses OpenRouter API with support for multiple LLM models
- **Conversation Memory**: Redis-backed per-user chat context (last 20 messages)
- **Rate Limiting**: In-memory token-bucket rate limiter per user (5 messages/minute)
- **Graceful Shutdown**: Proper handling of SIGINT/SIGTERM with inflight request completion
- **Testable Architecture**: Interface-based design with unit tests and Redis integration tests
- **Clean Structure**: Modular project organization with separation of concerns

## Project Structure

```
cmd/
 └── bot/
     └── main.go                 # Application entry point with setup & graceful shutdown

internal/
 ├── telegram/
 │    └── handler.go            # Telegram message handler & routing
 ├── service/
 │    ├── chat.go              # Chat service with business logic
 │    └── chat_test.go         # Unit tests with mocks
 ├── storage/
 │    ├── storage.go           # ChatStorage interface
 │    ├── redis_chat.go        # Redis implementation
 │    └── redis_chat_test.go   # Redis integration tests
 ├── ai/
 │    └── openrouter.go        # OpenRouter client wrapper
 └── config/
      └── config.go            # Configuration management

go.mod                           # Go module definition with dependencies
.env.example                    # Environment variables template
```

## Prerequisites

- Go 1.24+
- Redis 7.0+ (or Docker)
- Telegram Bot Token (get from @BotFather)
- OpenRouter API Key

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/Asylann/Socrates.git
cd Socrates
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```env
TELEGRAM_BOT_TOKEN=your_token_here
OPENROUTER_API_KEY=your_key_here
OPENROUTER_MODEL=openai/gpt-3.5-turbo
REDIS_URL=redis://localhost:6379/0
```

### 4. Start Redis (using Docker)

```bash
docker run -d -p 6379:6379 redis:7-alpine
```

Or locally if Redis is installed:

```bash
redis-server
```

## Running the Bot

### Development

```bash
go run ./cmd/bot/main.go
```

### Build

```bash
go build -o socrates-bot ./cmd/bot
./socrates-bot
```

## Testing

### Run all tests

```bash
go test ./...
```

### Run tests with coverage

```bash
go test -cover ./...
```

### Run specific test file

```bash
go test -v ./internal/service/...
```

### Unit Tests (no external dependencies)

```bash
go test ./internal/service
```

### Integration Tests (requires Redis)

```bash
go test ./internal/storage
```

## Architecture

### Component Overview

#### Telegram Handler
- Receives messages from Telegram API
- Enforces rate limiting before processing
- Delegates to ChatService
- Manages response sending

#### ChatService
- Orchestrates the conversation flow
- Loads conversation history from Redis
- Appends user message
- Calls AI client
- Stores AI response
- Implements business logic

#### Storage Layer
- `ChatStorage` interface: abstraction for conversation persistence
- `RedisChat`: Redis implementation storing messages as JSON in lists
  - Key format: `conv:{userID}`
  - Maintains last 20 messages per user
  - Automatic list trimming on append

#### AI Client
- Wraps OpenRouter API client
- Converts `storage.Message` to OpenRouter format
- Handles API errors gracefully
- Returns structured responses

#### Rate Limiter
- Token-bucket algorithm per user
- Configurable capacity and refill rate
- Automatic cleanup of idle buckets

### Message Flow

```
User Message (Telegram)
        ↓
    Handler receives update
        ↓
    Rate limiter check
        ↓
    ChatService.ProcessMessage()
        ↓
    Load conversation history (Redis)
        ↓
    Append user message to history
        ↓
    Call AI Client with full context
        ↓
    Append AI response to history (Redis)
        ↓
    Send response to Telegram
```

## Configuration Options

### Rate Limiter Configuration

Edit in `cmd/bot/main.go`:

```go
bucket := h.rateLimiter.GetBucket(userID, 5, 5.0/60.0) // 5 tokens per 60 seconds
//                                    ↑   ↑----- capacity (tokens)
//                                    └------ refill per second
```

Current: 5 messages per 60 seconds

## Available OpenRouter Models

See [OpenRouter Models](https://openrouter.ai/docs/models) for full list.


## Troubleshooting

### Bot doesn't respond

1. Check bot token is correct
2. Verify Redis is running: `redis-cli ping`
3. Check logs for errors

## Performance Considerations

- **Concurrent Users**: Unlimited (in-memory rate limiter, distributed by Redis)
- **Message History**: Last 20 per user (automatic trimming)
- **Redis Memory**: ~100 bytes per message × 20 × users
- **Timeout**: 5 seconds for graceful shutdown

## Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `github.com/wojtess/openrouter-api-go` - OpenRouter API client
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/joho/godotenv` - Environment loading
- `github.com/stretchr/testify` - Testing framework
- `github.com/testcontainers/testcontainers-go` - Docker containers for tests

## Contributing

Contributions welcome!

**Built with Go, Redis, and OpenRouter AI** 🚀
