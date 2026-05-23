package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
)

type Config struct {
	BotToken string
}

type Bot struct {
	bot    *bot.Bot
	logger *slog.Logger
}

func NewTelegramBot(cfg *Config, logger *slog.Logger) (*Bot, error) {
	b, err := bot.New(cfg.BotToken, bot.WithInitialOffset(-1), bot.WithDebug())
	if err != nil {
		return nil, fmt.Errorf("new telegram bot: %w", err)
	}
	return &Bot{
		bot:    b,
		logger: logger.With("component", "telegram-gateways"),
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("starting telegram bot webhook processor")
	b.bot.StartWebhook(ctx)
	return nil

}

func (b *Bot) Stop() {
	b.logger.Info("telegram gateways stopped")
}
