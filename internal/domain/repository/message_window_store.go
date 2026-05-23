package repository

import (
	"context"
	"noirbot/internal/domain/model"
	"time"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type MessageWindowStore interface {
	Append(ctx context.Context, ownerID, guestID int64, msg model.IncomingMessage) error
	CountSince(ctx context.Context, ownerID, guestID int64, since time.Time) (int, error)
}
