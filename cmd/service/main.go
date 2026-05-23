package main

import (
	"context"
	"fmt"
	"log/slog"
	"noirbot/internal/gateways/telegram"
	"os"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	app := fx.New(
		fx.Provide(
			newConfig,
			newBot,
		),
		fx.Invoke(
			func(cfg *Config) {
				slog.Info("config loaded")
			},
		),
	)

	app.Run()
}

func newConfig() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("could not process env vars: %w", err)
	}
	return cfg, nil
}

func newBot(lc fx.Lifecycle, cfg *Config, logger *slog.Logger) (*telegram.Bot, error) {
	b, err := telegram.NewTelegramBot(&telegram.Config{
		BotToken: cfg.BotToken}, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create telegram bot: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := b.Start(ctx); err != nil {
					slog.Error("telegram gateways error", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
	})

	return b, nil
}
