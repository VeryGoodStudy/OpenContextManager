package ocm

// Option is a functional option for configuring the Manager.
type Option func(*Manager)

// WithStrategy sets the context management strategy for the Manager.
func WithStrategy(strategy Strategy) Option {
	return func(m *Manager) {
		m.strategy = strategy
	}
}

// Manager provides the business logic for managing LLM conversation contexts.
// It uses a Storage interface for persistence and an optional Strategy for
// context processing, ensuring the business layer is fully decoupled from
// the data storage layer.
type Manager struct {
	storage  Storage
	strategy Strategy
}

// NewManager creates a new Manager with the given storage backend and options.
func NewManager(storage Storage, opts ...Option) *Manager {
	m := &Manager{
		storage: storage,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Create creates a new context with the given ID and persists it.
func (m *Manager) Create(id string) (*Context, error) {
	ctx := NewContext(id)
	if err := m.storage.Save(ctx); err != nil {
		return nil, err
	}
	return ctx, nil
}

// Get retrieves a context by ID. If a strategy is configured, the strategy
// is applied to produce a processed view of the context.
func (m *Manager) Get(id string) (*Context, error) {
	ctx, err := m.storage.Load(id)
	if err != nil {
		return nil, err
	}
	if m.strategy != nil {
		return m.strategy.Apply(ctx), nil
	}
	return ctx, nil
}

// GetRaw retrieves a context by ID without applying any strategy.
func (m *Manager) GetRaw(id string) (*Context, error) {
	return m.storage.Load(id)
}

// Append adds a message to the context identified by the given ID and
// persists the updated context.
func (m *Manager) Append(id string, msg Message) error {
	ctx, err := m.storage.Load(id)
	if err != nil {
		return err
	}
	ctx.AddMessage(msg)
	return m.storage.Save(ctx)
}

// Delete removes a context by its ID from storage.
func (m *Manager) Delete(id string) error {
	return m.storage.Delete(id)
}

// List returns all stored context IDs.
func (m *Manager) List() ([]string, error) {
	return m.storage.List()
}

// SetMetadata sets a metadata key-value pair on the context and persists it.
func (m *Manager) SetMetadata(id, key string, value interface{}) error {
	ctx, err := m.storage.Load(id)
	if err != nil {
		return err
	}
	ctx.SetMetadata(key, value)
	return m.storage.Save(ctx)
}
