package config

import (
	"os"
	"time"
)

type Config struct {
	TelegramToken       string
	OpenRouterToken     string
	OpenRouterModel     string
	RedisURL            string
	RateLimitCapacity   int
	RateLimitRefillRate float64
	RateLimitCleanupTTL time.Duration
	ShutdownTimeout     time.Duration
}

func LoadFromEnv() *Config {
	return &Config{
		TelegramToken:       os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenRouterToken:     os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel:     getEnvOrDefault("OPENROUTER_MODEL", "openai/gpt-3.5-turbo"),
		RedisURL:            getEnvOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		RateLimitCapacity:   5,
		RateLimitRefillRate: 5.0 / 60.0, // 5 messages per 60 seconds
		RateLimitCleanupTTL: 15 * time.Minute,
		ShutdownTimeout:     5 * time.Second,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
