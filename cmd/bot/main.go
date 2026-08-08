package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/Asylann/Socrates/internal/ai"
	"github.com/Asylann/Socrates/internal/config"
	"github.com/Asylann/Socrates/internal/ratelimit"
	"github.com/Asylann/Socrates/internal/service"
	"github.com/Asylann/Socrates/internal/storage"
	"github.com/Asylann/Socrates/internal/telegram"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Could not load .env file")
	}

	config := config.LoadFromEnv()

	// Initializing Telegram bot
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = false

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start interacting with the bot"},
		{Command: "help", Description: "Get help about using the bot"},
		{Command: "clear_history", Description: "Clear your conversation history"},
	}
	_, err = bot.Request(tgbotapi.NewSetMyCommands(commands..))
	if err != nil {
		log.Fatalf("Failed to set bot commands: %v", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opts)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	cancel()

	aiClient := ai.NewOpenRouterClient(config.OpenRouterToken, config.OpenRouterModel)

	chatStorage := storage.NewRedisChat(redisClient)

	chatService := service.NewChatService(aiClient, chatStorage, config.SystemPrompt)

	rateLimiter := ratelimit.NewManager(15 * time.Minute)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	rootCtx, rootCancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	handler := telegram.NewHandler(bot, chatService, rateLimiter, &wg)

	// updated polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running. Go to @" + bot.Self.UserName + " on Telegram to interact.")
	go func() {
		for {
			select {
			case update := <-utes:
				handler.HandleMessage(rootCtx, update)
			case <-rootCtx.Done():
				log.Println("Stopping message processing...")
				return
			}
		}
	}()

	<-sigChan
	log.Println("\nShutdown signal received. Gracefully shutting down...")

pdates
	bot.StopReceivingUpdates()
	log.Println("Telegram updates stopped")

	// Wait for in-flight handlers with timeout
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		log.Println("All handlers completed")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for handlers to complete")
	}
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis client: %v", err)
	}
	log.Println("Redis client closed")
	log.Println("Bot shutdown complete")
}
