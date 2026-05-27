package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Telegram      TelegramConfig
	HTTP          HTTPConfig
	DeepSeek      DeepSeekConfig
	Bot           BotConfig
	Flood         FloodConfig
	Greetings     []string `default:"привет,прив,здоров,хай,ку"                      envconfig:"GREETINGS"`
	AllowedOwners []int64  `envconfig:"ALLOWED_OWNERS"`
}

type TelegramConfig struct {
	BotToken      string `envconfig:"BOT_TOKEN"      required:"true"`
	WebhookSecret string `envconfig:"WEBHOOK_SECRET"`
}

type HTTPConfig struct {
	Addr            *string       `default:":8080" envconfig:"HTTP_ADDR"`
	ReadTimeout     time.Duration `default:"10s"   envconfig:"HTTP_READ_TIMEOUT"`
	WriteTimeout    time.Duration `default:"10s"   envconfig:"HTTP_WRITE_TIMEOUT"`
	ShutdownTimeout time.Duration `default:"5s"    envconfig:"HTTP_SHUTDOWN_TIMEOUT"`
}

type DeepSeekConfig struct {
	BaseURL string        `default:"https://api.deepseek.com/v1" envconfig:"DEEPSEEK_BASE_URL"`
	APIKey  string        `envconfig:"DEEPSEEK_API_KEY"          required:"true"`
	Model   string        `default:"deepseek-chat"               envconfig:"DEEPSEEK_MODEL"`
	Timeout time.Duration `default:"30s"                         envconfig:"DEEPSEEK_TIMEOUT"`
}

type BotConfig struct {
	SystemPrompt string `envconfig:"BOT_SYSTEM_PROMPT" required:"true"`
}

type FloodConfig struct {
	WindowDuration time.Duration `default:"60s" envconfig:"FLOOD_WINDOW"`
	MaxLen         int           `default:"20"  envconfig:"FLOOD_MAX_LEN"`
	Threshold      int           `default:"5"   envconfig:"FLOOD_THRESHOLD"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}
