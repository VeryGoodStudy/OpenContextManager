package ocm

import "errors"

// Common errors returned by storage implementations.
var (
	ErrContextNotFound = errors.New("context not found")
)

// Storage defines the interface for context data persistence.
// Implementations of this interface decouple the business logic from the
// underlying storage mechanism, allowing different backends (memory, file,
// database, Redis, etc.) to be used interchangeably.
type Storage interface {
	// Save persists a context. If the context already exists, it is overwritten.
	Save(ctx *Context) error

	// Load retrieves a context by its ID.
	// Returns ErrContextNotFound if the context does not exist.
	Load(id string) (*Context, error)

	// Delete removes a context by its ID.
	// Returns ErrContextNotFound if the context does not exist.
	Delete(id string) error

	// List returns all context IDs currently stored.
	List() ([]string, error)
}
