package memory

import (
	"context"
	"noirbot/internal/domain/repository"
)

var _ repository.OwnerWhitelist = (*OwnerWhitelist)(nil)

type OwnerWhitelist struct {
	allowed map[int64]struct{}
}

func NewOwnerWhitelist(ids []int64) *OwnerWhitelist {
	allowed := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}
	return &OwnerWhitelist{allowed: allowed}
}

func (w *OwnerWhitelist) IsAllowed(_ context.Context, ownerID int64) (bool, error) {
	if len(w.allowed) == 0 {
		return true, nil
	}
	_, ok := w.allowed[ownerID]
	return ok, nil
}
