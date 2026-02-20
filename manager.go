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

// Summarizer is a function that generates a summary string from a slice of messages.
type Summarizer func(messages []Message) (string, error)

// SummarizeFirstNRounds summarizes the first n conversation rounds of the context
// identified by id, then removes those rounds and inserts the summary in their place.
// A "round" consists of a user message immediately followed by an assistant message.
// A leading system message is always preserved.
// The summary is inserted as a message with RoleSummary after any leading system message.
// If fewer than n complete rounds exist, all complete rounds found are summarized.
func (m *Manager) SummarizeFirstNRounds(id string, n int, summarizer Summarizer) error {
	if n <= 0 {
		return nil
	}
	ctx, err := m.storage.Load(id)
	if err != nil {
		return err
	}

	msgs := ctx.Messages
	startIdx := 0
	if len(msgs) > 0 && msgs[0].Role == RoleSystem {
		startIdx = 1
	}

	// Find the end index of the first n complete rounds.
	roundEnd := startIdx
	roundsFound := 0
	i := startIdx
	for i+1 < len(msgs) && roundsFound < n {
		if msgs[i].Role == RoleUser && msgs[i+1].Role == RoleAssistant {
			roundEnd = i + 2
			roundsFound++
			i += 2
		} else {
			i++
		}
	}

	if roundsFound == 0 {
		return nil
	}

	summary, err := summarizer(msgs[startIdx:roundEnd])
	if err != nil {
		return err
	}

	summaryMsg := NewMessage(RoleSummary, summary)

	newMsgs := make([]Message, 0, startIdx+1+len(msgs)-roundEnd)
	newMsgs = append(newMsgs, msgs[:startIdx]...)
	newMsgs = append(newMsgs, summaryMsg)
	newMsgs = append(newMsgs, msgs[roundEnd:]...)

	ctx.ReplaceMessages(newMsgs)
	return m.storage.Save(ctx)
}
