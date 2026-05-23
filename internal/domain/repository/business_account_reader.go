package repository

import (
	"context"
	"noirbot/internal/domain/model"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type BusinessAccountReader interface {
	GetConnection(ctx context.Context, connectionID string) (model.BusinessConnection, error)
}
