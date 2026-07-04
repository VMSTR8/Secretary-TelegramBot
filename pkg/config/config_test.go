package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig_validateLongVoice(t *testing.T) {
	tests := []struct {
		name     string
		shortMax time.Duration
		longMax  time.Duration
		wantErr  error
	}{
		{
			name:     "short < long — валидно",
			shortMax: 10 * time.Second,
			longMax:  600 * time.Second,
			wantErr:  nil,
		},
		{
			name:     "long max = 0 — ErrInvalidLongVoiceCfg",
			shortMax: 10 * time.Second,
			longMax:  0,
			wantErr:  ErrInvalidLongVoiceCfg,
		},
		{
			name:     "long max < 0 — ErrInvalidLongVoiceCfg",
			shortMax: 10 * time.Second,
			longMax:  -1 * time.Second,
			wantErr:  ErrInvalidLongVoiceCfg,
		},
		{
			name:     "short == long — ErrVoiceDurationOverlap",
			shortMax: 600 * time.Second,
			longMax:  600 * time.Second,
			wantErr:  ErrVoiceDurationOverlap,
		},
		{
			name:     "short > long — ErrVoiceDurationOverlap",
			shortMax: 700 * time.Second,
			longMax:  600 * time.Second,
			wantErr:  ErrVoiceDurationOverlap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			c.ShortVoice.MaxDuration = tt.shortMax
			c.LongVoice.MaxDuration = tt.longMax

			err := c.validateLongVoice()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_validateShortVoice(t *testing.T) {
	tests := []struct {
		name    string
		window  time.Duration
		maxDur  time.Duration
		wantErr bool
	}{
		{
			name:    "оба > 0 — валидно",
			window:  60 * time.Second,
			maxDur:  10 * time.Second,
			wantErr: false,
		},
		{
			name:    "window = 0 — ошибка",
			window:  0,
			maxDur:  10 * time.Second,
			wantErr: true,
		},
		{
			name:    "max = 0 — ошибка",
			window:  60 * time.Second,
			maxDur:  0,
			wantErr: true,
		},
		{
			name:    "window < 0 — ошибка",
			window:  -1 * time.Second,
			maxDur:  10 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			c.ShortVoice.ResponseWindow = tt.window
			c.ShortVoice.MaxDuration = tt.maxDur

			err := c.validateShortVoice()

			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidShortVoiceCfg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_validateFlood(t *testing.T) {
	tests := []struct {
		name    string
		ttl     time.Duration
		window  time.Duration
		wantErr bool
	}{
		{
			name:    "ttl > window — валидно",
			ttl:     120 * time.Second,
			window:  60 * time.Second,
			wantErr: false,
		},
		{
			name:    "ttl == window — валидно",
			ttl:     60 * time.Second,
			window:  60 * time.Second,
			wantErr: false,
		},
		{
			name:    "ttl < window — ошибка",
			ttl:     30 * time.Second,
			window:  60 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			c.Flood.RedisTTL = tt.ttl
			c.Flood.WindowDuration = tt.window

			err := c.validateFlood()

			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidFloodTTL)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
