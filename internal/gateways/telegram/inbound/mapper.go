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
	if src == nil || src.From == nil {
		return model.IncomingMessage{}, false
	}

	base := model.IncomingMessage{
		BusinessConnectionID: src.BusinessConnectionID,
		GuestID:              src.From.ID,
		ReceivedAt:           time.Now().UTC(),
	}

	switch {
	case src.Text != "":
		base.Kind = model.MessageKindText
		base.Text = src.Text

		return base, true
	case src.Voice != nil:
		base.Kind = model.MessageKindVoice
		base.VoiceDuration = time.Duration(src.Voice.Duration) * time.Second

		return base, true
	default:
		return model.IncomingMessage{}, false
	}
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
