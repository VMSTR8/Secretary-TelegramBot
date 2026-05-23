package pkg

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
	Greetings     []string `envconfig:"GREETINGS" default:"привет,прив,здоров,хай,ку"`
	AllowedOwners []int64  `envconfig:"ALLOWED_OWNERS"`
}

type TelegramConfig struct {
	BotToken      string `envconfig:"BOT_TOKEN" required:"true"`
	WebhookSecret string `envconfig:"WEBHOOK_SECRET"`
}

type HTTPConfig struct {
	Addr            *string       `envconfig:"HTTP_ADDR" default:":8080"`
	ReadTimeout     time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	WriteTimeout    time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	ShutdownTimeout time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"5s"`
}

type DeepSeekConfig struct {
	BaseURL string        `envconfig:"DEEPSEEK_BASE_URL" default:"https://api.deepseek.com/v1"`
	APIKey  string        `envconfig:"DEEPSEEK_API_KEY" required:"true"`
	Model   string        `envconfig:"DEEPSEEK_MODEL" default:"deepseek-chat"`
	Timeout time.Duration `envconfig:"DEEPSEEK_TIMEOUT" default:"30s"`
}

type BotConfig struct {
	SystemPromt string `envconfig:"BOT_SYSTEM_PROMT" required:"true"`
}

type FloodConfig struct {
	WindowDuration time.Duration `envconfig:"FLOOD_WINDOW" default:"60s"`
	MaxLen         int           `envconfig:"FLOOD_MAX_LEN" default:"20"`
	Threshold      int           `envconfig:"FLOOD_THRESHOLD" default:"5"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}
