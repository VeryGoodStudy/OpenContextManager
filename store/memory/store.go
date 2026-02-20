// Package memory provides an in-memory implementation of the ocm.Storage interface.
package memory

import (
	"sync"

	ocm "github.com/VeryGoodStudy/OpenContextManager"
)

// Store is a thread-safe in-memory storage backend for contexts.
type Store struct {
	mu   sync.RWMutex
	data map[string]*ocm.Context
}

// New creates a new in-memory Store.
func New() *Store {
	return &Store{
		data: make(map[string]*ocm.Context),
	}
}

// Save persists a context in memory.
func (s *Store) Save(ctx *ocm.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[ctx.ID] = ctx
	return nil
}

// Load retrieves a context by ID from memory.
func (s *Store) Load(id string) (*ocm.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ctx, ok := s.data[id]
	if !ok {
		return nil, ocm.ErrContextNotFound
	}
	return ctx, nil
}

// Delete removes a context by ID from memory.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ocm.ErrContextNotFound
	}
	delete(s.data, id)
	return nil
}

// List returns all context IDs stored in memory.
func (s *Store) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	return ids, nil
}
