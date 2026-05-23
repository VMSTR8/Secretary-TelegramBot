package main

import (
	"log/slog"
	"noirbot/internal/domain/repository"
	"noirbot/internal/domain/service"
	"noirbot/internal/gateways/memory"
	"os"

	"go.uber.org/fx"

	"noirbot/pkg/config"
)

func main() {
	app := fx.New(
		fx.Provide(
			newLogger,
			config.Load,
			newGreetingDetector,
			newFloodDetector,
			newOwnerWhitelist,
			newBusinessConnectionStore,
			newMessageWindowStore,
		),
		fx.Invoke(
			func(log *slog.Logger) {
				log.Info("bot starting...")
			},
		),
	)

	app.Run()
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func newGreetingDetector(cfg *config.Config) *service.GreetingDetector {
	return service.NewGreetingDetector(cfg.Greetings)
}

func newFloodDetector(cfg *config.Config, store repository.MessageWindowStore) *service.FloodDetector {
	return service.NewFloodDetector(service.FloodDetectorConfig{
		WindowDuration: cfg.Flood.WindowDuration,
		MaxLen:         cfg.Flood.MaxLen,
		Threshold:      cfg.Flood.Threshold,
	}, store)
}

func newOwnerWhitelist(cfg *config.Config) repository.OwnerWhitelist {
	return memory.NewOwnerWhitelist(cfg.AllowedOwners)
}

func newBusinessConnectionStore() repository.BusinessConnectionStore {
	return memory.NewBusinessConnectionStore()
}

func newMessageWindowStore() repository.MessageWindowStore {
	return memory.NewMessageWindowStore()
}
