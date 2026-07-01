package service_test

import (
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/service"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLongVoiceDetector_Detect(t *testing.T) {
	const (
		minDuration = 10 * time.Second
		maxDuration = 600 * time.Second
	)

	detector := service.NewLongVoiceDetector(service.LongVoiceDetectorConfig{
		MinDuration: minDuration,
		MaxDuration: maxDuration,
	})

	tests := []struct {
		name     string
		msg      model.IncomingMessage
		wantKind model.TriggerKind
	}{
		{
			name: "voice в диапазоне — триггер long_voice",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: 30 * time.Second,
			},
			wantKind: model.TriggerKindLongVoice,
		},
		{
			name: "voice ровно на верхней границе — триггер long_voice",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: maxDuration,
			},
			wantKind: model.TriggerKindLongVoice,
		},
		{
			name: "voice на нижней границе short — без триггера",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: minDuration,
			},
			wantKind: model.TriggerKindNone,
		},
		{
			name: "voice короче min — без триггера",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: 5 * time.Second,
			},
			wantKind: model.TriggerKindNone,
		},
		{
			name: "voice длиннее max — без триггера",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: maxDuration + time.Second,
			},
			wantKind: model.TriggerKindNone,
		},
		{
			name: "текстовое сообщение — без триггера",
			msg: model.IncomingMessage{
				Kind: model.MessageKindText,
				Text: "привет",
			},
			wantKind: model.TriggerKindNone,
		},
		{
			name:     "пустой Kind — без триггера",
			msg:      model.IncomingMessage{},
			wantKind: model.TriggerKindNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.Detect(tt.msg)
			require.Equal(t, tt.wantKind, got.Kind)
		})
	}
}
