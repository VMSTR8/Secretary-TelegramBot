package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Telegram      TelegramConfig
	HTTP          HTTPConfig
	Redis         RedisConfig
	DeepSeek      DeepSeekConfig
	Bot           BotConfig
	Flood         FloodConfig
	Greetings     []string `default:"привет,прив,здоров,хай,ку" envconfig:"GREETINGS"`
	ShortVoice    ShortVoiceConfig
	AllowedOwners []int64 `envconfig:"ALLOWED_OWNERS"`
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

type RedisConfig struct {
	Addr                  string        `default:"localhost:6379" envconfig:"REDIS_ADDR"`
	Password              string        `default:""               envconfig:"REDIS_PASSWORD"`
	DB                    int           `default:"0"              envconfig:"REDIS_DB"`
	DialTimeout           time.Duration `default:"5s"             envconfig:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout           time.Duration `default:"3s"             envconfig:"REDIS_READ_TIMEOUT"`
	WriteTimeout          time.Duration `default:"3s"             envconfig:"REDIS_WRITE_TIMEOUT"`
	PoolSize              int           `default:"20"             envconfig:"REDIS_POOL_SIZE"`
	BusinessConnectionTTL time.Duration `default:"604800s"        envconfig:"REDIS_BUSINESS_TTL"`
}

type DeepSeekConfig struct {
	BaseURL string        `default:"https://api.deepseek.com/v1" envconfig:"DEEPSEEK_BASE_URL"`
	APIKey  string        `envconfig:"DEEPSEEK_API_KEY"          required:"true"`
	Model   string        `default:"deepseek-chat"               envconfig:"DEEPSEEK_MODEL"`
	Timeout time.Duration `default:"30s"                         envconfig:"DEEPSEEK_TIMEOUT"`
}

type BotConfig struct {
	SystemPrompt     string `envconfig:"BOT_SYSTEM_PROMPT"      required:"true"`
	ShortVoicePrompt string `envconfig:"BOT_SHORT_VOICE_PROMPT" required:"true"`
}

type FloodConfig struct {
	WindowDuration time.Duration `default:"60s"  envconfig:"FLOOD_WINDOW"`
	MaxLen         int           `default:"20"   envconfig:"FLOOD_MAX_LEN"`
	Threshold      int           `default:"5"    envconfig:"FLOOD_THRESHOLD"`
	RedisTTL       time.Duration `default:"120s" envconfig:"FLOOD_REDIS_TTL"`
}

type ShortVoiceConfig struct {
	MaxDuration    time.Duration `default:"10s" envconfig:"SHORT_VOICE_MAX_DURATION"`
	ResponseWindow time.Duration `default:"60s" envconfig:"SHORT_VOICE_RESPONSE_WINDOW"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := cfg.validateFlood(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	if err := cfg.validateShortVoice(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validateFlood() error {
	if c.Flood.RedisTTL < c.Flood.WindowDuration {
		return fmt.Errorf("%w: ttl=%s, window=%s",
			ErrInvalidFloodTTL,
			c.Flood.RedisTTL,
			c.Flood.WindowDuration,
		)
	}

	return nil
}

func (c *Config) validateShortVoice() error {
	if c.ShortVoice.ResponseWindow <= 0 || c.ShortVoice.MaxDuration <= 0 {
		return fmt.Errorf("%w: response_window=%d, max_duration=%d",
			ErrInvalidShortVoiceCfg,
			c.ShortVoice.ResponseWindow,
			c.ShortVoice.MaxDuration,
		)
	}

	return nil
}
