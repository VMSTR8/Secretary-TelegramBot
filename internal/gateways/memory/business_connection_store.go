package memory

import (
	"context"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"sync"
)

var _ repository.BusinessConnectionStore = (*BusinessConnectionStore)(nil)

type BusinessConnectionStore struct {
	mu    sync.RWMutex
	conns map[string]model.BusinessConnection
}

func NewBusinessConnectionStore() *BusinessConnectionStore {
	return &BusinessConnectionStore{
		conns: make(map[string]model.BusinessConnection),
	}
}

func (s *BusinessConnectionStore) Get(_ context.Context, connectionID string) (model.BusinessConnection, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, ok := s.conns[connectionID]

	return conn, ok, nil
}

func (s *BusinessConnectionStore) Put(_ context.Context, conn model.BusinessConnection) error {
	if conn.ID == "" {
		return ErrEmptyConnectionID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.conns[conn.ID] = conn

	return nil
}

func (s *BusinessConnectionStore) Delete(_ context.Context, connectionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, connectionID)

	return nil
}
