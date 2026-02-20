// Package file provides a file-based implementation of the ocm.Storage interface.
package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	ocm "github.com/VeryGoodStudy/OpenContextManager"
)

const fileSuffix = ".ctx.json"

// Store is a file-based storage backend that persists each context as a
// separate JSON file in a configurable directory.
type Store struct {
	mu  sync.RWMutex
	dir string
}

// New creates a new file-based Store. The directory is created if it does
// not exist.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id+fileSuffix)
}

// Save persists a context as a JSON file.
func (s *Store) Save(ctx *ocm.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(ctx.ID), data, 0o600)
}

// Load reads a context from a JSON file.
func (s *Store) Load(id string) (*ocm.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ocm.ErrContextNotFound
		}
		return nil, err
	}

	var ctx ocm.Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

// Delete removes a context's JSON file.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return ocm.ErrContextNotFound
		}
		return err
	}
	return nil
}

// List returns all context IDs by scanning the storage directory.
func (s *Store) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, fileSuffix) {
			ids = append(ids, strings.TrimSuffix(name, fileSuffix))
		}
	}
	return ids, nil
}
