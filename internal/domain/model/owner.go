package model

import "time"

type Owner struct {
	UserID int64
}

type BusinessConnection struct {
	ID          string
	Owner       Owner
	UserChatID  int64
	IsEnabled   bool
	CanReply    bool
	ConnectedAt time.Time
}
