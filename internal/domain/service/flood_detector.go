package service

import (
	"context"
	"fmt"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"time"
)

type FloodDetectorConfig struct {
	WindowDuration time.Duration
	MaxLen         int
	Threshold      int
}

type FloodDetector struct {
	cfg   FloodDetectorConfig
	store repository.MessageWindowStore
}

func NewFloodDetector(cfg FloodDetectorConfig, store repository.MessageWindowStore) *FloodDetector {
	return &FloodDetector{
		cfg:   cfg,
		store: store,
	}
}

func (d *FloodDetector) Detect(ctx context.Context, msg model.IncomingMessage) (model.TriggerDecision, error) {
	if len([]rune(msg.Text)) > d.cfg.MaxLen {
		return model.TriggerDecision{Kind: model.TriggerKindNone}, nil
	}

	if err := d.store.Append(ctx, msg.OwnerID, msg.GuestID, msg); err != nil {
		return model.TriggerDecision{}, fmt.Errorf("flood detector append: %w", err)
	}

	since := time.Now().Add(-d.cfg.WindowDuration)
	count, err := d.store.CountSince(ctx, msg.OwnerID, msg.GuestID, since)
	if err != nil {
		return model.TriggerDecision{}, fmt.Errorf("flood detector count: %w", err)
	}

	if count >= d.cfg.Threshold {
		return model.TriggerDecision{
			Kind:   model.TriggerKindFlood,
			Reason: fmt.Sprintf("%d messages in %s", count, d.cfg.WindowDuration),
		}, nil
	}

	return model.TriggerDecision{Kind: model.TriggerKindNone}, nil
}
