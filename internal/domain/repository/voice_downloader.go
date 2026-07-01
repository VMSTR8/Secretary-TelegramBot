package repository

import (
	"context"
	"io"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type VoiceDownloader interface {
	Download(ctx context.Context, fileID string) (io.ReadCloser, error)
}
