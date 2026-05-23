package repository

import "context"

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type LLMClient interface {
	Generate(ctx context.Context, systemPromt, userText string) (string, error)
}
