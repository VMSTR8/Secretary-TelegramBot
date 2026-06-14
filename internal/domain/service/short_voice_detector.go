package service

import (
	"noirbot/internal/domain/model"
	"time"
)

type ShortVoiceDetectorConfig struct {
	MaxDuration time.Duration
}

type ShortVoiceDetector struct {
	cfg ShortVoiceDetectorConfig
}

func NewShortVoiceDetector(cfg ShortVoiceDetectorConfig) *ShortVoiceDetector {
	return &ShortVoiceDetector{
		cfg: cfg,
	}
}

func (d *ShortVoiceDetector) Detect(msg model.IncomingMessage) model.TriggerDecision {
	if msg.Kind != model.MessageKindVoice {
		return model.TriggerDecision{Kind: model.TriggerKindNone}
	}

	if msg.VoiceDuration > d.cfg.MaxDuration {
		return model.TriggerDecision{Kind: model.TriggerKindNone}
	}

	return model.TriggerDecision{
		Kind:   model.TriggerKindShortVoice,
		Reason: "voice " + msg.VoiceDuration.String() + " <= " + d.cfg.MaxDuration.String(),
	}
}
