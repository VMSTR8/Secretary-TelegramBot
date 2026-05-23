package repository

import "context"

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type OwnerWhitelist interface {
	IsAllowed(ctx context.Context, ownerID int64) (bool, error)
}
