package config

import "errors"

var (
	ErrInvalidFloodTTL      = errors.New("config: FLOOD_REDIS_TTL must be >= FLOOD_WINDOW")
	ErrInvalidShortVoiceCfg = errors.New("config: SHORT_VOICE_MAX_DURATION and SHORT_VOICE_RESPONSE_WINDOW must be > 0")
	ErrInvalidLongVoiceCfg  = errors.New("config: LONG_VOICE_MAX_DURATION must be > 0 and > SHORT_VOICE_MAX_DURATION")
	ErrVoiceDurationOverlap = errors.New("config: SHORT_VOICE_MAX_DURATION must be < LONG_VOICE_MAX_DURATION")
)
