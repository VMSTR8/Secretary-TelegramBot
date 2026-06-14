package redis

import (
	"noirbot/internal/domain/model"
	"time"
)

type businessConnectionDTO struct {
	ID          string    `json:"id"`
	Owner       ownerDTO  `json:"owner"`
	UserChatID  int64     `json:"userChatId"`
	IsEnabled   bool      `json:"isEnabled"`
	CanReply    bool      `json:"canReply"`
	ConnectedAt time.Time `json:"connectedAt"`
}
type ownerDTO struct {
	UserID int64 `json:"userId"`
}

func toBusinessConnectionDTO(c model.BusinessConnection) businessConnectionDTO {
	return businessConnectionDTO{
		ID:          c.ID,
		Owner:       ownerDTO{UserID: c.Owner.UserID},
		UserChatID:  c.UserChatID,
		IsEnabled:   c.IsEnabled,
		CanReply:    c.CanReply,
		ConnectedAt: c.ConnectedAt,
	}
}

func fromBusinessConnectionDTO(dto businessConnectionDTO) model.BusinessConnection {
	return model.BusinessConnection{
		ID:          dto.ID,
		Owner:       model.Owner{UserID: dto.Owner.UserID},
		UserChatID:  dto.UserChatID,
		IsEnabled:   dto.IsEnabled,
		CanReply:    dto.CanReply,
		ConnectedAt: dto.ConnectedAt,
	}
}
