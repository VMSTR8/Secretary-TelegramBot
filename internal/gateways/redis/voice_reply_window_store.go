package redis

import (
	"context"
	"fmt"
	"noirbot/internal/domain/repository"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ repository.VoiceReplyWindowStore = (*VoiceReplyWindowStore)(nil)

const voiceReplyKeyPrefix = "vrw:"

type VoiceReplyWindowStore struct {
	client *redis.Client
}

func NewVoiceReplyWindowStore(client *redis.Client) *VoiceReplyWindowStore {
	return &VoiceReplyWindowStore{
		client: client,
	}
}

func (s *VoiceReplyWindowStore) TryEnter(
	ctx context.Context,
	connectionID string,
	guestID int64,
	ttl time.Duration,
) (bool, error) {
	key := voiceReplyKey(connectionID, guestID)

	ok, err := s.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("voice reply window setnx: %w", err)
	}

	return ok, nil
}

func voiceReplyKey(connectionID string, guestID int64) string {
	return voiceReplyKeyPrefix + connectionID + ":" + strconv.FormatInt(guestID, 10)
}

func (s *VoiceReplyWindowStore) Release(ctx context.Context, connectionID string, guestID int64) error {
	key := voiceReplyKey(connectionID, guestID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("voice reply window release: %w", err)
	}

	return nil
}
