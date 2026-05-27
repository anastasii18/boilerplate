package telegram

import (
	"context"
	"platform/pkg/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type TelegramClient interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type client struct {
	bot *bot.Bot
}

// NewClient создает новый клиент для Telegram Bot API
func NewClient(bot *bot.Bot) *client {
	return &client{
		bot: bot,
	}
}

// SendMessage отправляет сообщение в указанный чат
func (c *client) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := c.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdownV1,
	})

	if err != nil {
		logger.Error(ctx, "SendMessage", zap.Error(err))
		return err
	}

	return nil
}
