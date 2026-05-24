package inbound

import (
	"noirbot/internal/domain/model"
	"time"

	tgmodels "github.com/go-telegram/bot/models"
)

type UpdateMapper struct{}

func NewUpdateMapper() *UpdateMapper {
	return &UpdateMapper{}
}

func (m *UpdateMapper) ToIncomingMessage(src *tgmodels.Message) (model.IncomingMessage, bool) {
	if src == nil || src.From == nil || src.Text == "" {
		return model.IncomingMessage{}, false
	}
	return model.IncomingMessage{
		BusinessConnectionID: src.BusinessConnectionID,
		GuestID:              src.From.ID,
		Text:                 src.Text,
		ReceivedAt:           time.Unix(int64(src.Date), 0),
	}, true
}

func (m *UpdateMapper) ToBusinessConnection(src *tgmodels.BusinessConnection) model.BusinessConnection {
	conn := model.BusinessConnection{
		ID:          src.ID,
		Owner:       model.Owner{UserID: src.User.ID},
		UserChatID:  src.UserChatID,
		IsEnabled:   src.IsEnabled,
		ConnectedAt: time.Unix(src.Date, 0),
	}

	if src.Rights != nil {
		conn.CanReply = src.Rights.CanReply
	}
	return conn
}
