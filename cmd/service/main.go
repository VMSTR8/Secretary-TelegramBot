package main

import (
	"fmt"
	"log/slog"
	"noirbot/internal/domain/repository"
	"noirbot/internal/domain/service"
	"noirbot/internal/gateways/deepseek"
	"noirbot/internal/gateways/memory"
	"noirbot/internal/gateways/telegram/inbound"
	"noirbot/internal/gateways/telegram/outbound"
	"noirbot/internal/usecase/handle_business_connection"
	"noirbot/internal/usecase/handle_business_message"
	"noirbot/pkg/config"
	"os"

	httpgw "noirbot/internal/gateways/http"

	"github.com/go-telegram/bot"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			newLogger,
			config.Load,
			newBot,

			newGreetingDetector,
			newFloodDetector,
			newShortVoiceDetector,

			newOwnerWhitelist,
			newBusinessConnectionStore,
			newMessageWindowStore,

			newBusinessSender,
			newBusinessAccountReader,

			newDeepseekConfig,
			newLLMClient,

			newHandleBusinessMessageConfig,
			handle_business_connection.New,
			handle_business_message.New,

			inbound.NewLazyHandler,
			inbound.NewUpdateMapper,
			inbound.NewUpdateRouter,
			inbound.NewWebhookHandler,

			httpgw.NewEngine,
			httpgw.New,
		),
		fx.Invoke(
			wireLazyHandler,
			httpgw.RegisterRoutes,
			bindHTTPServerLifecycle,
		),
	)

	app.Run()
}

func wireLazyHandler(lazy *inbound.LazyHandler, router *inbound.UpdateRouter) {
	lazy.Set(router.AsHandlerFunc())
}

func bindHTTPServerLifecycle(lc fx.Lifecycle, s *httpgw.Server) {
	lc.Append(fx.Hook{
		OnStart: s.Start,
		OnStop:  s.Stop,
	})
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func newBot(cfg *config.Config) (*bot.Bot, error) {
	b, err := bot.New(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return b, nil
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

func newShortVoiceDetector(cfg *config.Config) *service.ShortVoiceDetector {
	return service.NewShortVoiceDetector(service.ShortVoiceDetectorConfig{
		MaxDuration: cfg.ShortVoice.MaxDuration,
	})
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

func newBusinessSender(b *bot.Bot) repository.BusinessSender {
	return outbound.NewSender(b)
}

func newBusinessAccountReader(b *bot.Bot) repository.BusinessAccountReader {
	return outbound.NewAccountReader(b)
}

func newDeepseekConfig(cfg *config.Config) deepseek.Config {
	return deepseek.Config{
		BaseURL: cfg.DeepSeek.BaseURL,
		APIKey:  cfg.DeepSeek.APIKey,
		Model:   cfg.DeepSeek.Model,
		Timeout: cfg.DeepSeek.Timeout,
	}
}

func newLLMClient(c deepseek.Config) repository.LLMClient {
	return deepseek.NewClient(c)
}

func newHandleBusinessMessageConfig(cfg *config.Config) handle_business_message.Config {
	return handle_business_message.Config{
		SystemPrompt:     cfg.Bot.SystemPrompt,
		ShortVoicePrompt: cfg.Bot.ShortVoicePrompt,
	}
}
