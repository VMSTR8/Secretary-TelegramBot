package memory

import (
	"context"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"sync"
	"time"
)

var _ repository.MessageWindowStore = (*MessageWindowStore)(nil)

type windowKey struct {
	ownerID int64
	guestID int64
}

type MessageWindowStore struct {
	mu      sync.Mutex
	windows map[windowKey][]time.Time
}

func NewMessageWindowStore() *MessageWindowStore {
	return &MessageWindowStore{
		windows: make(map[windowKey][]time.Time),
	}
}

func (s *MessageWindowStore) Append(_ context.Context, ownerID, guestID int64, _ model.IncomingMessage) error {
	key := windowKey{ownerID: ownerID, guestID: guestID}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.windows[key] = append(s.windows[key], time.Now())

	return nil
}

func (s *MessageWindowStore) CountSince(_ context.Context, ownerID, guestID int64, since time.Time) (int, error) {
	key := windowKey{ownerID: ownerID, guestID: guestID}

	s.mu.Lock()
	defer s.mu.Unlock()

	timestamps := s.windows[key]

	fresh := timestamps[:0]
	for _, t := range timestamps {
		if t.After(since) {
			fresh = append(fresh, t)
		}
	}

	s.windows[key] = fresh

	return len(fresh), nil
}
