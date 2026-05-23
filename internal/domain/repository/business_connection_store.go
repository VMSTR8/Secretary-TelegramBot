package repository

import (
	"context"
	"noirbot/internal/domain/model"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type BusinessConnectionStore interface {
	Get(ctx context.Context, connectionID string) (model.BusinessConnection, bool, error)
	Put(ctx context.Context, conn model.BusinessConnection) error
	Delete(ctx context.Context, connectionID string) error
}
