package main

import (
	"context"
	"fmt"
	"log/slog"
	"noirbot/internal/domain/repository"
	"noirbot/internal/domain/service"
	"noirbot/internal/gateways/deepseek"
	httpgw "noirbot/internal/gateways/http"
	"noirbot/internal/gateways/memory"
	redisstore "noirbot/internal/gateways/redis"
	"noirbot/internal/gateways/telegram/inbound"
	"noirbot/internal/gateways/telegram/outbound"
	"noirbot/internal/usecase/handle_business_connection"
	"noirbot/internal/usecase/handle_business_message"
	"noirbot/pkg/config"
	"os"

	"github.com/go-telegram/bot"
	"github.com/redis/go-redis/v9"
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

			newRedisClient,

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
			bindRedisLifecycle,
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

func bindRedisLifecycle(cfg *config.Config, lc fx.Lifecycle, r *redis.Client) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Redis.DialTimeout)
			defer cancel()

			if err := r.Ping(pingCtx).Err(); err != nil {
				return fmt.Errorf("ping redis: %w", err)
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			return r.Close()
		},
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

func newBusinessConnectionStore(r *redis.Client, cfg *config.Config) repository.BusinessConnectionStore {
	return redisstore.NewBusinessConnectionStore(r, cfg.Redis.BusinessConnectionTTL)
}

func newMessageWindowStore(r *redis.Client, cfg *config.Config) repository.MessageWindowStore {
	return redisstore.NewMessageWindowStore(r, cfg.Flood.WindowDuration, cfg.Flood.RedisTTL)
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

func newRedisClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
	})

	return rdb
}
