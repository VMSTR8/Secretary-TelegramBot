package outbound

import (
	"context"
	"fmt"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"time"

	"github.com/go-telegram/bot"
)

var _ repository.BusinessAccountReader = (*AccountReader)(nil)

type AccountReader struct {
	bot *bot.Bot
}

func NewAccountReader(b *bot.Bot) *AccountReader {
	return &AccountReader{
		bot: b,
	}
}

func (r *AccountReader) GetConnection(ctx context.Context, connectionID string) (model.BusinessConnection, error) {
	conn, err := r.bot.GetBusinessConnection(ctx, &bot.GetBusinessConnectionParams{
		BusinessConnectionID: connectionID,
	})
	if err != nil {
		return model.BusinessConnection{}, fmt.Errorf("get business connection: %w", err)
	}

	return model.BusinessConnection{
		ID:          conn.ID,
		Owner:       model.Owner{UserID: conn.User.ID},
		UserChatID:  conn.UserChatID,
		IsEnabled:   conn.IsEnabled,
		CanReply:    conn.Rights.CanReply,
		ConnectedAt: time.Unix(conn.Date, 0),
	}, nil
}
