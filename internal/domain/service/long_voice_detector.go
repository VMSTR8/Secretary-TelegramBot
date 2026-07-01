package service

import (
	"fmt"
	"noirbot/internal/domain/model"
	"time"
)

type LongVoiceDetectorConfig struct {
	MinDuration time.Duration
	MaxDuration time.Duration
}

type LongVoiceDetector struct {
	cfg LongVoiceDetectorConfig
}

func NewLongVoiceDetector(cfg LongVoiceDetectorConfig) *LongVoiceDetector {
	return &LongVoiceDetector{
		cfg: cfg,
	}
}

func (d *LongVoiceDetector) Detect(msg model.IncomingMessage) model.TriggerDecision {
	switch {
	case msg.Kind != model.MessageKindVoice:
		return model.TriggerDecision{Kind: model.TriggerKindNone}
	case msg.VoiceDuration <= d.cfg.MinDuration:
		return model.TriggerDecision{Kind: model.TriggerKindNone}
	case msg.VoiceDuration > d.cfg.MaxDuration:
		return model.TriggerDecision{Kind: model.TriggerKindNone}
	default:
		return model.TriggerDecision{
			Kind:   model.TriggerKindLongVoice,
			Reason: fmt.Sprintf("voice %s > %s <= %s", msg.VoiceDuration.String(), d.cfg.MinDuration.String(), d.cfg.MaxDuration.String()),
		}
	}
}
