package model

import "time"

type MessageKind string

const (
	MessageKindText  MessageKind = "text"
	MessageKindVoice MessageKind = "voice"
)

type IncomingMessage struct {
	BusinessConnectionID string
	OwnerID              int64
	GuestID              int64
	Kind                 MessageKind
	Text                 string
	VoiceDuration        time.Duration
	ReceivedAt           time.Time
}

type ReplyDraft struct {
	BusinessConnectionID string
	GuestID              int64
	Text                 string
}
