package telegram

import (
	"context"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Asylann/Socrates/internal/ratelimit"
	"github.com/Asylann/Socrates/internal/service"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	chatService *service.ChatService
	rateLimiter *ratelimit.Manager
	wg          *sync.WaitGroup
}

func NewHandler(bot *tgbotapi.BotAPI, chatService *service.ChatService, rateLimiter *ratelimit.Manager, wg *sync.WaitGroup) *Handler {
	return &Handler{
		bot:         bot,
		chatService: chatService,
		rateLimiter: rateLimiter,
		wg:          wg,
	}
}

// HandleMessage processes a incoming Telegram message.
func (h *Handler) HandleMessage(ctx context.Context, update tgbotapi.Update) {
	h.wg.Add(1)
	defer h.wg.Done()

	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	messageText := update.Message.Text

	log.Printf("[%s] %s", update.Message.From.UserName, messageText)

	switch update.Message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(chatID, "Welcome to Socrates AI Bot! Send me a message to start chatting.")
		h.bot.Send(msg)
		return
	case "help":
		helpText := "This is Socrates AI Bot. You can chat with me by sending messages. Use /start to begin."
		msg := tgbotapi.NewMessage(chatID, helpText)
		h.bot.Send(msg)
		return
	case "clear_history":
		err := h.chatService.ClearConversation(ctx, userID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Error clearing conversation history.")
			h.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(chatID, "Conversation history cleared.")
		h.bot.Send(msg)
		return
	}

	// 5 messages per minute = 5/60 tokens per second
	bucket := h.rateLimiter.GetBucket(userID, 5, 5.0/60.0)
	if !bucket.Allow(1) {
		msg := tgbotapi.NewMessage(chatID, "Rate limit exceeded. Please wait a few seconds.")
		h.bot.Send(msg)
		return
	}

	typingAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	h.bot.Send(typingAction)

	response, err := h.chatService.ProcessMessage(ctx, userID, messageText)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Error processing your message. Please try again.")
		h.bot.Send(msg)
		return
	}

	// Send response
	replyMsg := tgbotapi.NewMessage(chatID, response)
	replyMsg.ReplyToMessageID = update.Message.MessageID
	_, err = h.bot.Send(replyMsg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
