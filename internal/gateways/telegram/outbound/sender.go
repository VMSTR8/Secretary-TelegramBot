package outbound

import (
	"context"
	"fmt"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"

	tgmodel "github.com/go-telegram/bot/models"

	"github.com/go-telegram/bot"
)

var _ repository.BusinessSender = (*Sender)(nil)

type Sender struct {
	bot *bot.Bot
}

func NewSender(b *bot.Bot) *Sender {
	return &Sender{
		bot: b,
	}
}

func (s *Sender) Send(ctx context.Context, draft model.ReplyDraft) error {
	_, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
		BusinessConnectionID: draft.BusinessConnectionID,
		ChatID:               draft.GuestID,
		Text:                 draft.Text,
		ParseMode:            tgmodel.ParseModeHTML,
	})
	if err != nil {
		return fmt.Errorf("telegram send message: %w", err)
	}

	return nil
}
