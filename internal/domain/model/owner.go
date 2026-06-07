package model

import "time"

type Owner struct {
	UserID int64 `json:"userId"`
}

type BusinessConnection struct {
	ID          string    `json:"id"`
	Owner       Owner     `json:"owner"`
	UserChatID  int64     `json:"userChatId"`
	IsEnabled   bool      `json:"isEnabled"`
	CanReply    bool      `json:"canReply"`
	ConnectedAt time.Time `json:"connectedAt"`
}
