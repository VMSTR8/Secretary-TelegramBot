package config

import "errors"

var ErrInvalidFloodTTL = errors.New("config: FLOOD_REDIS_TTL must be >= FLOOD_WINDOW")
