package handle_business_connection

import (
	"context"
	"fmt"
	"log/slog"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
)

type Usecase struct {
	store repository.BusinessConnectionStore
	log   *slog.Logger
}

func New(store repository.BusinessConnectionStore, log *slog.Logger) *Usecase {
	return &Usecase{
		store: store,
		log:   log.With("usecase", "handle_business_connection"),
	}
}

func (uc *Usecase) Execute(ctx context.Context, conn model.BusinessConnection) error {
	if !conn.IsEnabled {
		if err := uc.store.Delete(ctx, conn.ID); err != nil {
			return fmt.Errorf("delete business connection: %w", err)
		}
		uc.log.InfoContext(ctx, "business connection removed",
			slog.String("connection_id", conn.ID),
			slog.Int64("owner_id", conn.Owner.UserID),
		)
		return nil
	}

	if err := uc.store.Put(ctx, conn); err != nil {
		return fmt.Errorf("put business connection: %w", err)
	}
	uc.log.InfoContext(ctx, "business connection saved",
		slog.String("connection_id", conn.ID),
		slog.Int64("owner_id", conn.Owner.UserID),
		slog.Bool("can_reply", conn.CanReply),
	)
	return nil
}
