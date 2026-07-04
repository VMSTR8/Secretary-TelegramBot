package inbound

import (
	"context"
	"log/slog"
	"noirbot/internal/usecase/businessconn"
	"noirbot/internal/usecase/businessmsg"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UpdateRouter struct {
	connUC *businessconn.Usecase
	msgUC  *businessmsg.Usecase
	mapper *UpdateMapper
	log    *slog.Logger
}

func NewUpdateRouter(
	connUC *businessconn.Usecase,
	msgUC *businessmsg.Usecase,
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
			r.log.WarnContext(ctx, "businessconn failed", "err", err)
		}
	case update.BusinessMessage != nil:
		msg, ok := r.mapper.ToIncomingMessage(update.BusinessMessage)
		if !ok {
			r.log.WarnContext(ctx, "failed to map business_message", "update_id", update.ID)

			return
		}

		if err := r.msgUC.Execute(ctx, msg); err != nil {
			r.log.ErrorContext(ctx, "businessmsg failed",
				"err", err,
				"guest_id", msg.GuestID,
				"conn_id", msg.BusinessConnectionID,
			)
		}
	default:
		r.log.DebugContext(ctx, "unhandled update type", "update_id", update.ID)
	}
}

func (r *UpdateRouter) AsHandlerFunc() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		r.Handle(ctx, b, update)
	}
}
