package repository

import (
	"context"
	"io"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type Transcriber interface {
	Transcribe(ctx context.Context, reader io.ReadCloser) (string, error)
}
