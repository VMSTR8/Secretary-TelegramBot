package repository

import (
	"context"
	"time"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

type VoiceReplyWindowStore interface {
	// TryEnter atomically opens a reply window if it is not already open.
	// Returns true if the window was just opened (the caller may reply),
	// or false if the window is already open (the caller must skip).
	TryEnter(ctx context.Context, connectionID string, guestID int64, ttl time.Duration) (bool, error)

	// Release drops the reservation if reply failed (DEL).
	Release(ctx context.Context, connectionID string, guestID int64) error
}
