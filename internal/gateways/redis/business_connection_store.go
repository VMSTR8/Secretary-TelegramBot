package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

const bcKeyPrefix = "bc:"

var _ repository.BusinessConnectionStore = (*BusinessConnectionStore)(nil)

type BusinessConnectionStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewBusinessConnectionStore(client *redis.Client, ttl time.Duration) *BusinessConnectionStore {
	return &BusinessConnectionStore{
		client: client,
		ttl:    ttl,
	}
}

func (s *BusinessConnectionStore) Get(
	ctx context.Context,
	connectionID string,
) (model.BusinessConnection, bool, error) {
	key := bcKeyPrefix + connectionID

	data, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.BusinessConnection{}, false, nil
	}

	if err != nil {
		return model.BusinessConnection{}, false, fmt.Errorf("redis get business_connection %s: %w", connectionID, err)
	}

	var dto businessConnectionDTO
	if unMshErr := json.Unmarshal([]byte(data), &dto); unMshErr != nil {
		return model.BusinessConnection{}, false, fmt.Errorf("redis unmarshal business_connection %s: %w", connectionID, unMshErr)
	}

	return fromBusinessConnectionDTO(dto), true, nil
}

func (s *BusinessConnectionStore) Put(ctx context.Context, conn model.BusinessConnection) error {
	if conn.ID == "" {
		return ErrEmptyConnectionID
	}

	key := bcKeyPrefix + conn.ID

	data, err := json.Marshal(toBusinessConnectionDTO(conn))
	if err != nil {
		return fmt.Errorf("redis marshal business_connection %s: %w", conn.ID, err)
	}

	if setErr := s.client.Set(ctx, key, data, s.ttl).Err(); setErr != nil {
		return fmt.Errorf("redis set business_connection %s: %w", conn.ID, setErr)
	}

	return nil
}

func (s *BusinessConnectionStore) Delete(ctx context.Context, connectionID string) error {
	key := bcKeyPrefix + connectionID

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del business_connection %s: %w", connectionID, err)
	}

	return nil
}
