package inbound_test

import (
	"noirbot/internal/domain/model"
	"noirbot/internal/gateways/telegram/inbound"
	"testing"
	"time"

	tgmodels "github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/require"
)

func TestUpdateMapper_ToIncomingMessage(t *testing.T) {
	mapper := inbound.NewUpdateMapper()
	from := &tgmodels.User{ID: 999}

	tests := []struct {
		name     string
		src      *tgmodels.Message
		wantOK   bool
		wantKind model.MessageKind
		wantText string
		wantDur  time.Duration
	}{
		{
			name: "text сообщение → MessageKindText",
			src: &tgmodels.Message{
				BusinessConnectionID: "conn-1",
				From:                 from,
				Text:                 "привет",
			},
			wantOK:   true,
			wantKind: model.MessageKindText,
			wantText: "привет",
		},
		{
			name: "voice сообщение → MessageKindVoice + длительность",
			src: &tgmodels.Message{
				BusinessConnectionID: "conn-1",
				From:                 from,
				Voice:                &tgmodels.Voice{Duration: 7},
			},
			wantOK:   true,
			wantKind: model.MessageKindVoice,
			wantDur:  7 * time.Second,
		},
		{
			name: "audio file (не voice) → пропускаем как неподдерживаемый тип",
			src: &tgmodels.Message{
				BusinessConnectionID: "conn-1",
				From:                 from,
				Audio:                &tgmodels.Audio{Duration: 5},
			},
			wantOK: false,
		},
		{
			name:   "ни text, ни voice → пропускаем",
			src:    &tgmodels.Message{From: from},
			wantOK: false,
		},
		{
			name:   "From == nil → пропускаем",
			src:    &tgmodels.Message{Text: "x"},
			wantOK: false,
		},
		{
			name:   "src == nil → пропускаем",
			src:    nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := mapper.ToIncomingMessage(tt.src)

			require.Equal(t, tt.wantOK, ok)

			if !tt.wantOK {
				return
			}

			require.Equal(t, tt.wantKind, got.Kind)
			require.Equal(t, tt.wantText, got.Text)
			require.Equal(t, tt.wantDur, got.VoiceDuration)
		})
	}
}
