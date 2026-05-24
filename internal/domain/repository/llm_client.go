package repository

import "context"

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type LLMClient interface {
	Generate(ctx context.Context, systemPrompt, userText string) (string, error)
}
