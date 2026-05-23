package main

type Config struct {
	BotToken     string `envconfig:"BOT_TOKEN" required:"true"`
	AllowedUsers string `envconfig:"ALLOWED_USERS"`
}
