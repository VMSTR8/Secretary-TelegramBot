package main

import (
	"fmt"
	"log/slog"
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
