package inbound

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type LazyHandler struct {
	fn  atomic.Pointer[bot.HandlerFunc]
	log *slog.Logger
}

func NewLazyHandler(l *slog.Logger) *LazyHandler {
	return &LazyHandler{
		log: l.With("component", "lazy_handler"),
	}
}

func (h *LazyHandler) Set(fn bot.HandlerFunc) {
	h.fn.Store(&fn)
}

func (h *LazyHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	fn := h.fn.Load()
	if fn == nil {
		h.log.WarnContext(ctx, "lazy handler called before Set — update dropped")

		return
	}

	(*fn)(ctx, b, update)
}
