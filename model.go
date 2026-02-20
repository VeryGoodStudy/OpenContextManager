package ocm

import "time"

// Role represents the role of a message sender.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single message in a conversation context.
type Message struct {
	Role      Role                   `json:"role"`
	Content   string                 `json:"content"`
	Name      string                 `json:"name,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// Context represents a conversation context containing a sequence of messages.
type Context struct {
	ID        string                 `json:"id"`
	Messages  []Message              `json:"messages"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewMessage creates a new Message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// WithName sets the name field on a Message.
func (m Message) WithName(name string) Message {
	m.Name = name
	return m
}

// WithMetadata sets a metadata key-value pair on a Message.
func (m Message) WithMetadata(key string, value interface{}) Message {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
	return m
}

// NewContext creates a new Context with the given ID.
func NewContext(id string) *Context {
	now := time.Now()
	return &Context{
		ID:        id,
		Messages:  make([]Message, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage appends a message to the context and updates the timestamp.
func (c *Context) AddMessage(msg Message) {
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// SetMetadata sets a metadata key-value pair on the Context.
func (c *Context) SetMetadata(key string, value interface{}) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]interface{})
	}
	c.Metadata[key] = value
	c.UpdatedAt = time.Now()
}

// MessageCount returns the number of messages in the context.
func (c *Context) MessageCount() int {
	return len(c.Messages)
}
