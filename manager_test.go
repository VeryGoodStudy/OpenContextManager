package ocm

import (
	"errors"
	"sync"
	"testing"
)

// mockStorage is a simple in-memory storage for testing the Manager.
type mockStorage struct {
	mu   sync.RWMutex
	data map[string]*Context
}

func newMockStorage() *mockStorage {
	return &mockStorage{data: make(map[string]*Context)}
}

func (s *mockStorage) Save(ctx *Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[ctx.ID] = ctx
	return nil
}

func (s *mockStorage) Load(id string) (*Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ctx, ok := s.data[id]
	if !ok {
		return nil, ErrContextNotFound
	}
	return ctx, nil
}

func (s *mockStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrContextNotFound
	}
	delete(s.data, id)
	return nil
}

func (s *mockStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	return ids, nil
}

func TestManager_Create(t *testing.T) {
	m := NewManager(newMockStorage())
	ctx, err := m.Create("ctx-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ctx.ID != "ctx-1" {
		t.Errorf("ID = %q, want ctx-1", ctx.ID)
	}
}

func TestManager_AppendAndGet(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")

	err := m.Append("ctx-1", NewMessage(RoleUser, "hello"))
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	ctx, err := m.Get("ctx-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(ctx.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(ctx.Messages))
	}
	if ctx.Messages[0].Content != "hello" {
		t.Errorf("Content = %q, want hello", ctx.Messages[0].Content)
	}
}

func TestManager_GetWithStrategy(t *testing.T) {
	s := NewSlidingWindowStrategy(2)
	m := NewManager(newMockStorage(), WithStrategy(s))
	_, _ = m.Create("ctx-1")

	_ = m.Append("ctx-1", NewMessage(RoleUser, "a"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "b"))
	_ = m.Append("ctx-1", NewMessage(RoleUser, "c"))

	// Get should apply strategy (keep last 2)
	ctx, err := m.Get("ctx-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(ctx.Messages) != 2 {
		t.Fatalf("Messages len = %d, want 2", len(ctx.Messages))
	}

	// GetRaw should return all messages
	raw, err := m.GetRaw("ctx-1")
	if err != nil {
		t.Fatalf("GetRaw: %v", err)
	}
	if len(raw.Messages) != 3 {
		t.Fatalf("Raw Messages len = %d, want 3", len(raw.Messages))
	}
}

func TestManager_Delete(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")

	if err := m.Delete("ctx-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := m.Get("ctx-1")
	if !errors.Is(err, ErrContextNotFound) {
		t.Errorf("Get after delete: %v, want ErrContextNotFound", err)
	}
}

func TestManager_List(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("a")
	_, _ = m.Create("b")

	ids, err := m.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("List len = %d, want 2", len(ids))
	}
}

func TestManager_SetMetadata(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")

	err := m.SetMetadata("ctx-1", "model", "gpt-4")
	if err != nil {
		t.Fatalf("SetMetadata: %v", err)
	}

	ctx, err := m.Get("ctx-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ctx.Metadata["model"] != "gpt-4" {
		t.Errorf("Metadata[model] = %v, want gpt-4", ctx.Metadata["model"])
	}
}

func TestManager_AppendNotFound(t *testing.T) {
	m := NewManager(newMockStorage())
	err := m.Append("nonexistent", NewMessage(RoleUser, "hello"))
	if !errors.Is(err, ErrContextNotFound) {
		t.Errorf("Append error = %v, want ErrContextNotFound", err)
	}
}

func TestManager_SummarizeFirstNRounds(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")

	_ = m.Append("ctx-1", NewMessage(RoleSystem, "You are helpful."))
	_ = m.Append("ctx-1", NewMessage(RoleUser, "q1"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a1"))
	_ = m.Append("ctx-1", NewMessage(RoleUser, "q2"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a2"))
	_ = m.Append("ctx-1", NewMessage(RoleUser, "q3"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a3"))

	summarizer := func(msgs []Message) (string, error) {
		return "summary of rounds", nil
	}

	err := m.SummarizeFirstNRounds("ctx-1", 2, summarizer)
	if err != nil {
		t.Fatalf("SummarizeFirstNRounds: %v", err)
	}

	ctx, err := m.GetRaw("ctx-1")
	if err != nil {
		t.Fatalf("GetRaw: %v", err)
	}

	// Expect: [system, summary, q3, a3]
	if len(ctx.Messages) != 4 {
		t.Fatalf("Messages len = %d, want 4", len(ctx.Messages))
	}
	if ctx.Messages[0].Role != RoleSystem {
		t.Errorf("Messages[0].Role = %q, want system", ctx.Messages[0].Role)
	}
	if ctx.Messages[1].Role != RoleSummary {
		t.Errorf("Messages[1].Role = %q, want summary", ctx.Messages[1].Role)
	}
	if ctx.Messages[1].Content != "summary of rounds" {
		t.Errorf("Messages[1].Content = %q, want 'summary of rounds'", ctx.Messages[1].Content)
	}
	if ctx.Messages[2].Content != "q3" {
		t.Errorf("Messages[2].Content = %q, want q3", ctx.Messages[2].Content)
	}
	if ctx.Messages[3].Content != "a3" {
		t.Errorf("Messages[3].Content = %q, want a3", ctx.Messages[3].Content)
	}
}

func TestManager_SummarizeFirstNRounds_NoSystem(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")

	_ = m.Append("ctx-1", NewMessage(RoleUser, "q1"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a1"))
	_ = m.Append("ctx-1", NewMessage(RoleUser, "q2"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a2"))

	err := m.SummarizeFirstNRounds("ctx-1", 1, func(msgs []Message) (string, error) {
		return "summary", nil
	})
	if err != nil {
		t.Fatalf("SummarizeFirstNRounds: %v", err)
	}

	ctx, _ := m.GetRaw("ctx-1")
	// Expect: [summary, q2, a2]
	if len(ctx.Messages) != 3 {
		t.Fatalf("Messages len = %d, want 3", len(ctx.Messages))
	}
	if ctx.Messages[0].Role != RoleSummary {
		t.Errorf("Messages[0].Role = %q, want summary", ctx.Messages[0].Role)
	}
	if ctx.Messages[1].Content != "q2" {
		t.Errorf("Messages[1].Content = %q, want q2", ctx.Messages[1].Content)
	}
}

func TestManager_SummarizeFirstNRounds_NZero(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")
	_ = m.Append("ctx-1", NewMessage(RoleUser, "q1"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a1"))

	err := m.SummarizeFirstNRounds("ctx-1", 0, func(msgs []Message) (string, error) {
		return "summary", nil
	})
	if err != nil {
		t.Fatalf("SummarizeFirstNRounds: %v", err)
	}

	ctx, _ := m.GetRaw("ctx-1")
	if len(ctx.Messages) != 2 {
		t.Fatalf("Messages len = %d, want 2 (unchanged)", len(ctx.Messages))
	}
}

func TestManager_SummarizeFirstNRounds_FewerRoundsThanN(t *testing.T) {
	m := NewManager(newMockStorage())
	_, _ = m.Create("ctx-1")
	_ = m.Append("ctx-1", NewMessage(RoleUser, "q1"))
	_ = m.Append("ctx-1", NewMessage(RoleAssistant, "a1"))

	// Ask for 5 rounds but only 1 exists — should summarize the 1 complete round.
	err := m.SummarizeFirstNRounds("ctx-1", 5, func(msgs []Message) (string, error) {
		return "summary", nil
	})
	if err != nil {
		t.Fatalf("SummarizeFirstNRounds: %v", err)
	}

	ctx, _ := m.GetRaw("ctx-1")
	// Expect: [summary]
	if len(ctx.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(ctx.Messages))
	}
	if ctx.Messages[0].Role != RoleSummary {
		t.Errorf("Messages[0].Role = %q, want summary", ctx.Messages[0].Role)
	}
}

func TestManager_SummarizeFirstNRounds_NotFound(t *testing.T) {
	m := NewManager(newMockStorage())
	err := m.SummarizeFirstNRounds("nonexistent", 1, func(msgs []Message) (string, error) {
		return "summary", nil
	})
	if !errors.Is(err, ErrContextNotFound) {
		t.Errorf("error = %v, want ErrContextNotFound", err)
	}
}
