package repository

import (
	"context"
	"noirbot/internal/domain/model"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type BusinessSender interface {
	Send(ctx context.Context, draft model.ReplyDraft) error
}
