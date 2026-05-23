package model

import "time"

type IncomingMessage struct {
	BusinessConnectionID string
	OwnerID              int64
	GuestID              int64
	Text                 string
	ReceivedAt           time.Time
}

type ReplyDraft struct {
	BusinessConnectionID string
	GuestID              int64
	Text                 string
}
