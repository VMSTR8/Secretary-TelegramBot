package service_test

import (
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/service"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShortVoiceDetector_Detect(t *testing.T) {
	const maxDuration = 10 * time.Second

	detector := service.NewShortVoiceDetector(service.ShortVoiceDetectorConfig{
		MaxDuration: maxDuration,
	})

	tests := []struct {
		name     string
		msg      model.IncomingMessage
		wantKind model.TriggerKind
	}{
		{
			name: "voice короче порога — триггер short_voice",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: 5 * time.Second,
			},
			wantKind: model.TriggerKindShortVoice,
		},
		{
			name: "voice ровно на пороге — триггер short_voice",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: maxDuration,
			},
			wantKind: model.TriggerKindShortVoice,
		},
		{
			name: "voice длиннее порога — без триггера",
			msg: model.IncomingMessage{
				Kind:          model.MessageKindVoice,
				VoiceDuration: 30 * time.Second,
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
