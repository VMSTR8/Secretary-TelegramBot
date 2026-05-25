package inbound

import (
	"context"
	"log/slog"
	"noirbot/internal/usecase/handle_business_connection"
	"noirbot/internal/usecase/handle_business_message"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UpdateRouter struct {
	connUC *handle_business_connection.Usecase
	msgUC  *handle_business_message.Usecase
	mapper *UpdateMapper
	log    *slog.Logger
}

func NewUpdateRouter(
	connUC *handle_business_connection.Usecase,
	msgUC *handle_business_message.Usecase,
	m *UpdateMapper,
	log *slog.Logger,
) *UpdateRouter {
	return &UpdateRouter{
		connUC: connUC,
		msgUC:  msgUC,
		mapper: m,
		log:    log.With("component", "update_router"),
	}
}
func (r *UpdateRouter) Handle(ctx context.Context, _ *bot.Bot, update *models.Update) {
	switch {
	case update.BusinessConnection != nil:
		conn := r.mapper.ToBusinessConnection(update.BusinessConnection)
		if err := r.connUC.Execute(ctx, conn); err != nil {
			r.log.Warn("handle_business_connection failed", "err", err)
		}
	case update.BusinessMessage != nil:
		msg, ok := r.mapper.ToIncomingMessage(update.BusinessMessage)
		if !ok {
			r.log.Warn("failed to map business_message", "update_id", update.ID)
			return
		}
		if err := r.msgUC.Execute(ctx, msg); err != nil {
			r.log.Error("handle_business_message_failed",
				"err", err,
				"guest_id", msg.GuestID,
				"conn_id", msg.BusinessConnectionID,
			)
		}
	default:
		r.log.Debug("unhandled update type", "update_id", update.ID)
	}
}

func (r *UpdateRouter) AsHandlerFunc() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		r.Handle(ctx, b, update)
	}
}
