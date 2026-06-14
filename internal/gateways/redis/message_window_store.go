package redis

import (
	"context"
	"fmt"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const floodKeyPrefix = "flood:"

var _ repository.MessageWindowStore = (*MessageWindowStore)(nil)

type MessageWindowStore struct {
	client    *redis.Client
	window    time.Duration
	ttl       time.Duration
	memberSeq atomic.Uint64
}

func NewMessageWindowStore(client *redis.Client, window, ttl time.Duration) *MessageWindowStore {
	return &MessageWindowStore{
		client: client,
		window: window,
		ttl:    ttl,
	}
}

func (s *MessageWindowStore) Append(ctx context.Context, ownerID, guestID int64, _ model.IncomingMessage) error {
	key := floodKey(ownerID, guestID)
	now := time.Now()
	score := float64(now.UnixNano())
	cutoff := strconv.FormatInt(now.Add(-s.window).UnixNano(), 10)

	pipe := s.client.TxPipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: s.newMember(now)})
	pipe.ZRemRangeByScore(ctx, key, "-inf", cutoff)
	pipe.Expire(ctx, key, s.ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis flood append owner=%d guest=%d: %w", ownerID, guestID, err)
	}

	return nil
}

func (s *MessageWindowStore) CountSince(ctx context.Context, ownerID, guestID int64, since time.Time) (int, error) {
	key := floodKey(ownerID, guestID)
	minScore := "(" + strconv.FormatInt(since.UnixNano(), 10)

	count, err := s.client.ZCount(ctx, key, minScore, "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("redis flood count owner=%d guest=%d: %w", ownerID, guestID, err)
	}

	return int(count), nil
}

func (s *MessageWindowStore) newMember(now time.Time) string {
	seq := s.memberSeq.Add(1)

	return strconv.FormatInt(now.UnixNano(), 10) + ":" + strconv.FormatUint(seq, 10)
}

func floodKey(ownerID, guestID int64) string {
	return floodKeyPrefix + strconv.FormatInt(ownerID, 10) + ":" + strconv.FormatInt(guestID, 10)
}
